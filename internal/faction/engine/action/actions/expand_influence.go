package actions

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ExpandInfluence places a new Base of Influence on a world or reinforces an
// existing one. New bases trigger a contested roll; rivals that tie or beat the
// roll may make a free attack against the new base.
type ExpandInfluence struct {
	collector   action.Collector
	roller      domain.Roller
	index       *world.Index
	worldEngine *world.WorldEngine
	order       action.ExpandInfluenceOrder
	mutations   []domain.Mutation
}

func NewExpandInfluence(collector action.Collector, roller domain.Roller, index *world.Index, worldEngine *world.WorldEngine) *ExpandInfluence {
	return &ExpandInfluence{collector: collector, roller: roller, index: index, worldEngine: worldEngine}
}

func (ei *ExpandInfluence) Name() string { return "Expand Influence" }

func (ei *ExpandInfluence) Validate(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) bool {
	if faction.Coin < 1 {
		return false
	}
	return len(eligibleNewBaseWorlds(faction, rulebook, ei.worldEngine)) > 0 ||
		len(damagedNonHomeworldBases(faction)) > 0 ||
		len(growableNonHomeworldBases(faction)) > 0
}

func (ei *ExpandInfluence) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	eligible := eligibleNewBaseWorlds(faction, rulebook, ei.worldEngine)
	order, err := ei.collector.SelectExpandInfluenceOrder(faction, factionState, eligible)
	if err != nil {
		return fmt.Errorf("expand influence: %w", err)
	}
	ei.order = order
	return nil
}

func (ei *ExpandInfluence) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	switch ei.order.Mode {
	case action.ExpandModeNew:
		return ei.resolveNewBase(faction, factionState, rulebook)
	case action.ExpandModeReinforce:
		return ei.resolveReinforce(faction)
	default:
		return fmt.Errorf("expand influence: unknown mode %q", ei.order.Mode)
	}
}

func (ei *ExpandInfluence) Output() ([]domain.Mutation, error) {
	return ei.mutations, nil
}

func (ei *ExpandInfluence) resolveNewBase(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	cost := ei.order.HPAmount
	if faction.Coin < cost {
		return fmt.Errorf("expand influence: insufficient Coin: need %d, have %d", cost, faction.Coin)
	}

	newBase := domain.Base{
		ID:          newBaseID(faction, ei.order.World),
		OwnerID:     faction.ID,
		Location:    ei.order.World,
		CurrentHP:   ei.order.HPAmount,
		MaxHP:       ei.order.HPAmount,
		Ready:       false,
		IsHomeworld: false,
	}

	// CoinDelta emitted first: Coin is spent even if the base is immediately destroyed.
	ei.mutations = append(ei.mutations,
		domain.CoinDelta{FactionID: faction.ID, Delta: -cost, Cause: "expand", CausedByFactionID: faction.ID},
		domain.BaseAdded{FactionID: faction.ID, Base: newBase, Cause: "expand", CausedByFactionID: faction.ID},
	)

	// Contested roll: faction rolls 1d10 + Cunning; each rival on the world rolls the same.
	// Rivals that tie or beat the faction's roll may make a free attack against the new base.
	factionRoll := ei.roller.Roll(10) + faction.Cunning
	baseHPTracker := 0

	rivals := rivalsOnWorld(factionState, faction.ID, ei.order.World, ei.index)
	sort.Slice(rivals, func(i, j int) bool { return rivals[i].Name < rivals[j].Name })

	for _, rival := range rivals {
		if newBase.CurrentHP+baseHPTracker <= 0 {
			break
		}
		rivalRoll := ei.roller.Roll(10) + rival.Cunning
		if rivalRoll < factionRoll {
			continue
		}
		confirmed, err := ei.collector.ConfirmRivalFreeAttack(rival, rivalRoll, factionRoll)
		if err != nil {
			return fmt.Errorf("expand influence: rival free attack prompt: %w", err)
		}
		if !confirmed {
			continue
		}
		sub := &baseAttack{collector: ei.collector, roller: ei.roller}
		if err := sub.resolve(rival, &newBase, faction, &ei.mutations, &baseHPTracker, rulebook); err != nil {
			return fmt.Errorf("expand influence: rival free attack: %w", err)
		}
	}
	return nil
}

func (ei *ExpandInfluence) resolveReinforce(faction *domain.Faction) error {
	base := findBase(faction, ei.order.BaseID)
	if base == nil {
		return fmt.Errorf("expand influence: base %q not found", ei.order.BaseID)
	}
	switch ei.order.SubMode {
	case action.ReinforceHeal:
		maxHP := base.EffectiveMaxHP(faction)
		amount := min(ei.order.HPAmount, maxHP-base.CurrentHP)
		if faction.Coin < amount {
			return fmt.Errorf("expand influence: insufficient Coin: need %d, have %d", amount, faction.Coin)
		}
		ei.mutations = append(ei.mutations,
			domain.BaseHealed{FactionID: faction.ID, BaseID: base.ID, Delta: amount, Cause: "expand", CausedByFactionID: faction.ID},
			domain.CoinDelta{FactionID: faction.ID, Delta: -amount, Cause: "expand", CausedByFactionID: faction.ID},
		)
	case action.ReinforceMax:
		amount := min(ei.order.HPAmount, faction.MaxHP-base.MaxHP)
		if faction.Coin < amount {
			return fmt.Errorf("expand influence: insufficient Coin: need %d, have %d", amount, faction.Coin)
		}
		ei.mutations = append(ei.mutations,
			domain.BaseExpanded{FactionID: faction.ID, BaseID: base.ID, Delta: amount, Cause: "expand", CausedByFactionID: faction.ID},
			domain.CoinDelta{FactionID: faction.ID, Delta: -amount, Cause: "expand", CausedByFactionID: faction.ID},
		)
	default:
		return fmt.Errorf("expand influence: unknown reinforce sub-mode %q", ei.order.SubMode)
	}
	return nil
}

// baseAttack resolves a rival's free attack against a newly placed Base of
// Influence. Not registered with the action engine — only invoked from
// ExpandInfluence.Resolve. Bases do not counterattack.
type baseAttack struct {
	collector action.Collector
	roller    domain.Roller
}

// resolve runs the attack sequence for one rival against the target base.
// baseHPTracker accumulates damage across rival attacks so re-checks see the
// correct effective HP before mutations are applied to state.
func (ba *baseAttack) resolve(rival *domain.Faction, target *domain.Base, ownerFaction *domain.Faction, mutations *[]domain.Mutation, baseHPTracker *int, rulebook *rulebook.Rulebook) error {
	eligible := rivalAssetsOnWorld(rival, target.Location)
	if len(eligible) == 0 {
		return nil
	}

	attackers, err := ba.collector.SelectBaseAttackers(rival, eligible, rulebook)
	if err != nil {
		return fmt.Errorf("base attack: selecting attackers: %w", err)
	}

	for _, attacker := range attackers {
		if target.CurrentHP+*baseHPTracker <= 0 {
			break
		}
		attackerDef, ok := rulebook.Assets[attacker.DefinitionID]
		if !ok || attackerDef.Attack == nil {
			continue
		}
		attackRoll := ba.roller.Roll(10) + statScore(rival, attackerDef.Attack.AttackerStat)
		defenseRoll := ba.roller.Roll(10) + statScore(ownerFaction, domain.StatCunning)

		if attackRoll >= defenseRoll {
			damage := attackerDef.Attack.Damage.Roll(ba.roller)
			*mutations = append(*mutations,
				// SWN: damage to a Base is also dealt directly to faction HP.
				domain.BaseHPDelta{FactionID: ownerFaction.ID, BaseID: target.ID, Delta: -damage, Cause: "expand", CausedByFactionID: rival.ID},
				domain.FactionHPDelta{FactionID: ownerFaction.ID, Delta: -damage, Cause: "expand", CausedByFactionID: rival.ID},
			)
			*baseHPTracker -= damage
			if target.CurrentHP+*baseHPTracker <= 0 {
				*mutations = append(*mutations, domain.BaseDestroyed{FactionID: ownerFaction.ID, BaseID: target.ID, Cause: "expand", CausedByFactionID: rival.ID})
			}
		}
	}
	return nil
}

// eligibleNewBaseWorlds returns worlds where the faction can place a new base,
// filtered by tech level: the world's TL must support the faction's highest-TL
// asset already operating there. Falls back to worldsForNewBase when worldEngine
// is nil (no spatial data loaded).
func eligibleNewBaseWorlds(faction *domain.Faction, rulebook *rulebook.Rulebook, worldEngine *world.WorldEngine) []string {
	candidates := worldsForNewBase(faction)
	if worldEngine == nil {
		return candidates
	}
	var result []string
	for _, worldID := range candidates {
		loc, ok := worldEngine.Location(worldID)
		if !ok {
			continue
		}
		maxAssetTL := 0
		if rulebook != nil {
			for _, asset := range faction.Assets {
				if asset.Location != worldID {
					continue
				}
				def, ok := rulebook.Assets[asset.DefinitionID]
				if ok && def.TechLevel > maxAssetTL {
					maxAssetTL = def.TechLevel
				}
			}
		}
		if loc.TechLevel() >= maxAssetTL {
			result = append(result, worldID)
		}
	}
	sort.Strings(result)
	return result
}

func worldsForNewBase(faction *domain.Faction) []string {
	worldsWithAssets := map[string]struct{}{}
	for _, asset := range faction.Assets {
		worldsWithAssets[asset.Location] = struct{}{}
	}
	worldsWithBase := map[string]struct{}{}
	for _, base := range faction.Bases {
		worldsWithBase[base.Location] = struct{}{}
	}
	var result []string
	for world := range worldsWithAssets {
		if _, exists := worldsWithBase[world]; !exists {
			result = append(result, world)
		}
	}
	sort.Strings(result)
	return result
}

func damagedNonHomeworldBases(faction *domain.Faction) []*domain.Base {
	var result []*domain.Base
	for _, base := range faction.Bases {
		if !base.IsHomeworld && base.CurrentHP < base.EffectiveMaxHP(faction) {
			result = append(result, base)
		}
	}
	return result
}

func growableNonHomeworldBases(faction *domain.Faction) []*domain.Base {
	var result []*domain.Base
	for _, base := range faction.Bases {
		if !base.IsHomeworld && base.MaxHP < faction.MaxHP {
			result = append(result, base)
		}
	}
	return result
}

func rivalsOnWorld(factionState *state.FactionState, factionID, locationID string, index *world.Index) []*domain.Faction {
	seen := make(map[string]struct{})
	var result []*domain.Faction
	for _, asset := range index.AssetsByLocation[locationID] {
		if asset.OwnerID == factionID {
			continue
		}
		if _, exists := seen[asset.OwnerID]; exists {
			continue
		}
		seen[asset.OwnerID] = struct{}{}
		if rival, exists := factionState.Factions[asset.OwnerID]; exists {
			result = append(result, rival)
		}
	}
	return result

}

func rivalAssetsOnWorld(rival *domain.Faction, world string) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range rival.Assets {
		if asset.Location == world && !asset.Stealthy && asset.Ready && asset.CurrentHP > 0 && asset.Maintained {
			result = append(result, asset)
		}
	}
	return result
}

func findBase(faction *domain.Faction, baseID string) *domain.Base {
	for _, base := range faction.Bases {
		if base.ID == baseID {
			return base
		}
	}
	return nil
}

// newBaseID generates a collision-free ID for a new non-homeworld Base.
// Follows the same monotonic-suffix pattern as asset IDs.
func newBaseID(faction *domain.Faction, world string) string {
	prefix := fmt.Sprintf("%s-base-%s-", faction.ID, world)
	highest := 0
	for _, base := range faction.Bases {
		if !strings.HasPrefix(base.ID, prefix) {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(base.ID, prefix))
		if err != nil {
			continue
		}
		if n > highest {
			highest = n
		}
	}
	return fmt.Sprintf("%s%d", prefix, highest+1)
}
