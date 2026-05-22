package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// AttackAction targets rival assets with one or more of the faction's own
// assets. All attackers are committed up front; the matchup sequence runs to
// completion. Each attacker may attack once; a defender may defend multiple
// times. Only known (non-stealthy) rival assets may be targeted.
type AttackAction struct {
	collector action.Collector
	roller    domain.Roller
	registry  *hooks.Registry
	index     *world.Index
	attackers []*domain.Asset
	mutations []domain.Mutation
}

func NewAttack(collector action.Collector, roller domain.Roller, registry *hooks.Registry, index *world.Index) *AttackAction {
	return &AttackAction{collector: collector, roller: roller, registry: registry, index: index}
}

func (attack *AttackAction) Name() string { return "Attack" }

func (attack *AttackAction) Validate(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool {
	for _, attacker := range eligibleAttackers(faction) {
		if attackerHasTarget(attacker, faction, rulebook, attack.index) {
			return true
		}
	}
	return false
}

func (attack *AttackAction) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	var candidates []*domain.Asset
	for _, attacker := range eligibleAttackers(faction) {
		if attackerHasTarget(attacker, faction, rulebook, attack.index) {
			candidates = append(candidates, attacker)
		}
	}

	selected, err := attack.collector.SelectAttackers(candidates, rulebook)
	if err != nil {
		return fmt.Errorf("attack: %w", err)
	}
	attack.attackers = selected
	return nil
}

func (attack *AttackAction) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	// Local HP trackers reflect damage dealt this sequence before mutations are
	// applied to state, so mid-sequence re-checks see accurate effective HP.
	assetHPTracker := make(map[string]int)
	baseHPTracker := make(map[string]int)
	// Stealth-cleared tracks assets already cleared this sequence; defenders can
	// face multiple matchups and we emit the mutation only once.
	stealthCleared := make(map[string]bool)

	for _, attacker := range attack.attackers {
		attackerDef, ok := rulebook.Assets[attacker.DefinitionID]
		if !ok || attackerDef.Attack == nil {
			continue
		}

		// Re-check: skip if destroyed earlier in this sequence.
		if attacker.CurrentHP+assetHPTracker[attacker.ID] <= 0 {
			continue
		}

		defenders := liveDefendersAtHex(faction.ID, attacker.Location.RegionHex, assetHPTracker, attack.index)
		if len(defenders) == 0 {
			continue
		}

		defender, err := attack.collector.SelectDefender(attacker, defenders, rulebook)
		if err != nil {
			return fmt.Errorf("attack: selecting defender: %w", err)
		}

		defenderDef, ok := rulebook.Assets[defender.DefinitionID]
		if !ok {
			return fmt.Errorf("attack: defender definition not found: %s", defender.DefinitionID)
		}

		defenderFaction := ownerFaction(factionState, defender)
		if defenderFaction == nil {
			return fmt.Errorf("attack: owner faction not found for defender %s", defender.ID)
		}

		// Stealth loss is emitted before damage so the EventRecord timeline is consistent.
		if attacker.Stealthy && !stealthCleared[attacker.ID] {
			attack.mutations = append(attack.mutations, domain.AssetStealthCleared{
				FactionID:         faction.ID,
				AssetID:           attacker.ID,
				Cause:             "attack",
				CausedByFactionID: faction.ID,
			})
			stealthCleared[attacker.ID] = true
		}
		if defender.Stealthy && !stealthCleared[defender.ID] {
			attack.mutations = append(attack.mutations, domain.AssetStealthCleared{
				FactionID:         defenderFaction.ID,
				AssetID:           defender.ID,
				Cause:             "attack",
				CausedByFactionID: faction.ID,
			})
			stealthCleared[defender.ID] = true
		}

		attackResult := dispatch.RollWithHooks(
			hooks.RollContext{
				Phase:     hooks.PhaseAttack,
				Actor:     faction,
				Opponent:  defenderFaction,
				Attribute: string(attackerDef.Attack.AttackerStat),
				Asset:     attacker,
				World:     attacker.Location.WorldID,
			},
			domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(faction, attackerDef.Attack.AttackerStat)},
			attack.registry, attack.collector, attack.roller, faction, factionState, rulebook,
		)
		attackRoll := attackResult.Sum

		defResult := dispatch.RollWithHooks(
			hooks.RollContext{
				Phase:     hooks.PhaseDefense,
				Actor:     defenderFaction,
				Opponent:  faction,
				Attribute: string(attackerDef.Attack.DefenderStat),
				Asset:     defender,
				World:     attacker.Location.WorldID,
			},
			domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(defenderFaction, attackerDef.Attack.DefenderStat)},
			attack.registry, attack.collector, attack.roller, defenderFaction, factionState, rulebook,
		)
		defenseRoll := defResult.Sum

		tieCtx := hooks.RollContext{
			Phase:    hooks.PhaseAttack,
			Actor:    faction,
			Opponent: defenderFaction,
			Asset:    attacker,
			World:    attacker.Location.WorldID,
		}
		tieOutcome := dispatch.ResolveTie(attack.registry, tieCtx, factionState)

		// Attack damage: attacker wins on tie (TieStandard/TieAttackerWins) or strictly greater.
		if attackRoll > defenseRoll || (attackRoll == defenseRoll && tieOutcome != hooks.TieDefenderWins) {
			damage := attackerDef.Attack.Damage.Roll(attack.roller)
			base := factionBaseOnWorld(defenderFaction, attacker.Location.WorldID)

			// Redirect prompt offered only when the defender has a live Base on this world.
			if base != nil && base.CurrentHP+baseHPTracker[base.ID] > 0 {
				redirect, err := attack.collector.ConfirmRedirectToBase(defenderFaction, base, damage)
				if err != nil {
					return fmt.Errorf("attack: redirect prompt: %w", err)
				}
				if redirect {
					attack.mutations = append(attack.mutations,
						domain.BaseHPDelta{FactionID: defenderFaction.ID, BaseID: base.ID, Delta: -damage, Cause: "attack", CausedByFactionID: faction.ID},
						// SWN: damage to a Base is also dealt directly to faction HP.
						domain.FactionHPDelta{FactionID: defenderFaction.ID, Delta: -damage, Cause: "attack", CausedByFactionID: faction.ID},
					)
					baseHPTracker[base.ID] -= damage
					if base.CurrentHP+baseHPTracker[base.ID] <= 0 {
						attack.mutations = append(attack.mutations, domain.BaseDestroyed{
							FactionID:         defenderFaction.ID,
							BaseID:            base.ID,
							Cause:             "attack",
							CausedByFactionID: faction.ID,
						})
					}
				} else {
					attack.applyAssetDamage(defenderFaction.ID, faction.ID, defender, damage, assetHPTracker)
				}
			} else {
				attack.applyAssetDamage(defenderFaction.ID, faction.ID, defender, damage, assetHPTracker)
			}
		}

		// Counterattack damage: defender wins on tie (TieStandard/TieDefenderWins) or strictly greater.
		if defenseRoll > attackRoll || (defenseRoll == attackRoll && tieOutcome != hooks.TieAttackerWins) {
			if defenderDef.Counter != nil {
				counterDamage := defenderDef.Counter.Roll(attack.roller)
				attack.applyAssetDamage(faction.ID, defenderFaction.ID, attacker, counterDamage, assetHPTracker)
			}
		}
	}
	return nil
}

func (attack *AttackAction) Output() ([]domain.Mutation, error) {
	return attack.mutations, nil
}

// applyAssetDamage emits AssetHPDelta and, when lethal, AssetRemoved inline.
// factionID is the owner of the damaged asset; causedByFactionID is the attacker.
// Updates the local tracker so subsequent re-checks see the correct effective HP.
func (attack *AttackAction) applyAssetDamage(factionID, causedByFactionID string, asset *domain.Asset, damage int, tracker map[string]int) {
	attack.mutations = append(attack.mutations, domain.AssetHPDelta{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Delta:             -damage,
		Cause:             "attack",
		CausedByFactionID: causedByFactionID,
	})
	tracker[asset.ID] -= damage
	if asset.CurrentHP+tracker[asset.ID] <= 0 {
		attack.mutations = append(attack.mutations, domain.AssetRemoved{
			FactionID:         factionID,
			AssetID:           asset.ID,
			Cause:             "attack",
			CausedByFactionID: causedByFactionID,
		})
	}
}

// attackerHasTarget returns true if the attacker has a valid Attack profile and
// at least one eligible defender on its world.
func attackerHasTarget(attacker *domain.Asset, faction *domain.Faction, rulebook *rulebook.Rulebook, index *world.Index) bool {
	def, ok := rulebook.Assets[attacker.DefinitionID]
	if !ok || def.Attack == nil {
		return false
	}
	return len(eligibleDefendersAtHex(faction.ID, attacker.Location.RegionHex, index)) > 0
}

// liveDefendersOnWorld filters eligibleDefendersOnWorld to exclude assets
// whose effective HP has been driven to 0 by earlier matchups in the current
// Resolve pass.
func liveDefendersOnWorld(attackerFactionID, worldID string, assetHPTracker map[string]int, index *world.Index) []*domain.Asset {
	return filterLiveDefenders(eligibleDefendersOnWorld(attackerFactionID, worldID, index), assetHPTracker)
}

// liveDefendersAtHex filters eligibleDefendersAtHex to exclude assets whose
// effective HP has been driven to 0 by earlier matchups in the current Resolve
// pass. Used when the attacker is mid-flight (empty WorldID).
func liveDefendersAtHex(attackerFactionID string, hex spatial.RegionHex, assetHPTracker map[string]int, index *world.Index) []*domain.Asset {
	return filterLiveDefenders(eligibleDefendersAtHex(attackerFactionID, hex, index), assetHPTracker)
}

func filterLiveDefenders(defenders []*domain.Asset, assetHPTracker map[string]int) []*domain.Asset {
	live := make([]*domain.Asset, 0, len(defenders))
	for _, defender := range defenders {
		if defender.CurrentHP+assetHPTracker[defender.ID] > 0 {
			live = append(live, defender)
		}
	}
	return live
}

// ownerFaction returns the faction that owns the given asset, or nil.
func ownerFaction(factionState *state.FactionState, asset *domain.Asset) *domain.Faction {
	return factionState.Factions[asset.OwnerID]
}
