package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// AttackAction targets rival assets with one or more of the faction's own
// assets. All attackers are committed up front; the matchup sequence runs to
// completion. Each attacker may attack once; a defender may defend multiple
// times. Only known (non-stealthy) rival assets may be targeted.
type AttackAction struct {
	collector action.Collector
	roller    domain.Roller
	attackers []*domain.Asset
	mutations []domain.Mutation
}

func NewAttack(collector action.Collector, roller domain.Roller) *AttackAction {
	return &AttackAction{collector: collector, roller: roller}
}

func (attack *AttackAction) Name() string { return "Attack" }

func (attack *AttackAction) Validate(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool {
	for _, attacker := range eligibleAttackers(faction) {
		if attackerHasTarget(attacker, faction, factionState, rulebook) {
			return true
		}
	}
	return false
}

func (attack *AttackAction) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	var candidates []*domain.Asset
	for _, attacker := range eligibleAttackers(faction) {
		if attackerHasTarget(attacker, faction, factionState, rulebook) {
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

		defenders := liveDefenders(factionState, faction.ID, attacker.Location, assetHPTracker)
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

		// Both rolls use the attacking asset's Attack profile — it defines which
		// stats are tested on each side.
		attackRoll := attack.roller.Roll(10) + statScore(faction, attackerDef.Attack.AttackerStat)
		defenseRoll := attack.roller.Roll(10) + statScore(defenderFaction, attackerDef.Attack.DefenderStat)

		// Attack damage: attacker wins on tie or strictly greater.
		if attackRoll >= defenseRoll {
			damage := attackerDef.Attack.Damage.Roll(attack.roller)
			base := factionBaseOnWorld(defenderFaction, attacker.Location)

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

		// Counterattack damage: defender wins (strictly greater) or tie.
		if defenseRoll >= attackRoll {
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
func attackerHasTarget(attacker *domain.Asset, faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool {
	def, ok := rulebook.Assets[attacker.DefinitionID]
	if !ok || def.Attack == nil {
		return false
	}
	return len(eligibleDefenders(factionState, faction.ID, attacker.Location)) > 0
}

// liveDefenders filters eligibleDefenders to exclude assets whose effective HP
// has been driven to 0 by earlier matchups in the current Resolve pass.
func liveDefenders(factionState *state.FactionState, attackerFactionID, world string, assetHPTracker map[string]int) []*domain.Asset {
	defenders := eligibleDefenders(factionState, attackerFactionID, world)
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
	for _, faction := range factionState.Factions {
		for _, factionAsset := range faction.Assets {
			if factionAsset.ID == asset.ID {
				return faction
			}
		}
	}
	return nil
}
