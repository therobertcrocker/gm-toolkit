package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
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
	registry  *hooks.Registry
	index     *world.Index
	effects   *effect.EffectsEngine
	attackers []*domain.Asset
	mutations []domain.Mutation
}

func NewAttack(collector action.Collector, roller domain.Roller, registry *hooks.Registry, index *world.Index, effects *effect.EffectsEngine) *AttackAction {
	return &AttackAction{collector: collector, roller: roller, registry: registry, index: index, effects: effects}
}

func (attack *AttackAction) Name() string { return "Attack" }

func (attack *AttackAction) Validate(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool {
	for _, attacker := range eligibleAttackers(faction, factionState) {
		if attackerHasTarget(attacker, rulebook, attack.registry, attack.index) {
			return true
		}
	}
	return false
}

func (attack *AttackAction) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	var candidates []*domain.Asset
	for _, attacker := range eligibleAttackers(faction, factionState) {
		if attackerHasTarget(attacker, rulebook, attack.registry, attack.index) {
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

		defenders := liveDefendersAtHex(attacker, attackerDef, assetHPTracker, rulebook, attack.registry, attack.index)
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

		attack.mutations = append(attack.mutations, attack.fireMatchup(domain.TriggerOwnAttack, faction, attacker, attackerDef, defenderFaction, defender, factionState).Mutations...)

		attackResult := dispatch.RollWithHooks(
			hooks.RollContext{
				Phase:         hooks.PhaseAttack,
				Actor:         faction,
				Opponent:      defenderFaction,
				Attribute:     string(attackerDef.Attack.AttackerStat),
				Asset:         attacker,
				OpponentAsset: defender,
				World:         attacker.Location.WorldID,
				Action:        attack.Name(),
			},
			domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(faction, attackerDef.Attack.AttackerStat)},
			attack.registry, attack.collector, attack.roller, faction, factionState, rulebook,
		)
		attackRoll := attackResult.Sum

		defResult := dispatch.RollWithHooks(
			hooks.RollContext{
				Phase:         hooks.PhaseDefense,
				Actor:         defenderFaction,
				Opponent:      faction,
				Attribute:     string(attackerDef.Attack.DefenderStat),
				Asset:         defender,
				OpponentAsset: attacker,
				World:         attacker.Location.WorldID,
				Action:        attack.Name(),
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
		attackerWins := attackRoll > defenseRoll || (attackRoll == defenseRoll && tieOutcome != hooks.TieDefenderWins)
		defenderWins := defenseRoll > attackRoll || (defenseRoll == attackRoll && tieOutcome != hooks.TieAttackerWins)

		// Attack damage: attacker wins on tie (TieStandard/TieAttackerWins) or strictly greater.
		if attackerWins {
			if attackerDef.Attack.Damage == nil {
				// deliberate: a special hit skips the redirect prompt, because it has no damage to send to a Base (D-15).
				attack.applyAssetDamage(defenderFaction.ID, faction.ID, defender, 0, assetHPTracker, attacker, false, "attack")
			} else {
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
							domain.BaseHPDelta{FactionID: defenderFaction.ID, BaseID: base.ID, Delta: -damage, Cause: "attack", CausedByFactionID: faction.ID, CausedByAssetID: attacker.ID},
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
						attack.applyAssetDamage(defenderFaction.ID, faction.ID, defender, damage, assetHPTracker, attacker, false, "attack")
					}
				} else {
					attack.applyAssetDamage(defenderFaction.ID, faction.ID, defender, damage, assetHPTracker, attacker, false, "attack")
				}
			}
		}

		// deliberate: a disabled defender rolls its defense but deals no counter and fires no own_defense_win entry, because the book makes it useless (Effort 3 Decision 2).
		defenderActive := defender.Active()

		// Counterattack damage: defender wins on tie (TieStandard/TieDefenderWins) or strictly greater.
		if defenderWins && defenderActive {
			if defenderDef.Counter != nil && dispatch.ResolveAttackEligibility(attack.registry, hooks.AttackEligibilityQuery{
				Attacker:    defender,
				AttackerDef: defenderDef,
				Defender:    attacker,
				DefenderDef: attackerDef,
				Side:        hooks.SideCounter,
			}) {
				counterDamage := defenderDef.Counter.Roll(attack.roller)
				attack.applyAssetDamage(faction.ID, defenderFaction.ID, attacker, counterDamage, assetHPTracker, defender, true, "attack")
			}
		}

		// deliberate: own_defense_win needs an outright win, not a tie, because on a tie the attack also landed and the defense did not succeed (Decision 13).
		if defenderWins && !attackerWins && defenderActive {
			outcome := attack.fireMatchup(domain.TriggerOwnDefenseWin, defenderFaction, defender, defenderDef, faction, attacker, factionState)
			attack.mutations = append(attack.mutations, outcome.Mutations...)
			// deliberate: no prompt when the attacker has no damage and no counter, because its self-attack could do nothing (Decision Record 9).
			canHurtItself := attackerDef.Attack.Damage != nil || attackerDef.Counter != nil
			if outcome.Counterstrike && canHurtItself && attacker.CurrentHP+assetHPTracker[attacker.ID] > 0 {
				confirmed, err := attack.collector.ConfirmCounterstrike(defenderFaction, defender, attacker)
				if err != nil {
					return fmt.Errorf("attack: counterstrike prompt: %w", err)
				}
				if confirmed {
					attack.counterstrike(faction, attacker, attackerDef, defenderFaction, defender, assetHPTracker, factionState, rulebook)
				}
			}
		}
	}
	return nil
}

// counterstrike turns the attacker on itself (D-18): its attack roll against its
// own defense. A hit deals its attack damage; a miss deals its counter damage, if
// it has one. The cause is not "attack", so no own_attack_hit fires on it.
func (attack *AttackAction) counterstrike(faction *domain.Faction, attacker *domain.Asset, attackerDef *domain.AssetDefinition, defenderFaction *domain.Faction, source *domain.Asset, tracker map[string]int, factionState *state.FactionState, rulebook *rulebook.Rulebook) {
	worldID := attacker.Location.WorldID
	// deliberate: Opponent nil, not faction, because Actor == Opponent would ask every asset hook on the world twice (overview Decision 9).
	attackRoll := dispatch.RollWithHooks(
		hooks.RollContext{Phase: hooks.PhaseAttack, Actor: faction, Attribute: string(attackerDef.Attack.AttackerStat), Asset: attacker, World: worldID, Action: attack.Name()},
		domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(faction, attackerDef.Attack.AttackerStat)},
		attack.registry, attack.collector, attack.roller, faction, factionState, rulebook,
	).Sum
	defenseRoll := dispatch.RollWithHooks(
		hooks.RollContext{Phase: hooks.PhaseDefense, Actor: faction, Attribute: string(attackerDef.Attack.DefenderStat), Asset: attacker, World: worldID, Action: attack.Name()},
		domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(faction, attackerDef.Attack.DefenderStat)},
		attack.registry, attack.collector, attack.roller, faction, factionState, rulebook,
	).Sum
	tieOutcome := dispatch.ResolveTie(attack.registry, hooks.RollContext{Phase: hooks.PhaseAttack, Actor: faction, Asset: attacker, World: worldID}, factionState)

	damage := attackerDef.Counter
	if attackRoll > defenseRoll || (attackRoll == defenseRoll && tieOutcome != hooks.TieDefenderWins) {
		damage = attackerDef.Attack.Damage
	}
	if damage == nil {
		return
	}
	attack.applyAssetDamage(faction.ID, defenderFaction.ID, attacker, damage.Roll(attack.roller), tracker, source, false, string(domain.SpecialCounterstrike))
}

func (attack *AttackAction) Output() ([]domain.Mutation, error) {
	return attack.mutations, nil
}

// fireMatchup raises the matchup route for the declaring asset. A nil effects
// engine (action tests) raises nothing.
func (attack *AttackAction) fireMatchup(trigger domain.EffectTrigger, faction *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, opponent *domain.Faction, opponentAsset *domain.Asset, factionState *state.FactionState) effect.MatchupOutcome {
	if attack.effects == nil {
		return effect.MatchupOutcome{}
	}
	return attack.effects.FireMatchup(trigger, effect.MatchupContext{
		Faction:       faction,
		Asset:         asset,
		Def:           def,
		Opponent:      opponent,
		OpponentAsset: opponentAsset,
		FactionState:  factionState,
	})
}

// applyAssetDamage emits AssetHPDelta and, when lethal, AssetRemoved inline.
// factionID is the owner of the damaged asset; causedByFactionID is the attacker.
// Updates the local tracker so subsequent re-checks see the correct effective HP.
func (attack *AttackAction) applyAssetDamage(factionID, causedByFactionID string, asset *domain.Asset, damage int, tracker map[string]int, causedByAsset *domain.Asset, counter bool, cause string) {
	causedByAssetID := ""
	if causedByAsset != nil {
		causedByAssetID = causedByAsset.ID
	}
	attack.mutations = append(attack.mutations, domain.AssetHPDelta{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Delta:             -damage,
		Cause:             cause,
		CausedByFactionID: causedByFactionID,
		CausedByAssetID:   causedByAssetID,
		Counter:           counter,
	})
	tracker[asset.ID] -= damage
	if asset.CurrentHP+tracker[asset.ID] <= 0 {
		attack.mutations = append(attack.mutations, domain.AssetRemoved{
			FactionID:         factionID,
			AssetID:           asset.ID,
			Cause:             cause,
			CausedByFactionID: causedByFactionID,
		})
	}
}

// attackerHasTarget returns true if the attacker has a valid Attack profile and
// at least one eligible defender on its world.
func attackerHasTarget(attacker *domain.Asset, rulebook *rulebook.Rulebook, registry *hooks.Registry, index *world.Index) bool {
	def, ok := rulebook.Assets[attacker.DefinitionID]
	if !ok || def.Attack == nil {
		return false
	}
	return len(eligibleDefendersAtHex(attacker, def, rulebook, registry, index)) > 0
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
func liveDefendersAtHex(attacker *domain.Asset, attackerDef *domain.AssetDefinition, assetHPTracker map[string]int, rulebook *rulebook.Rulebook, registry *hooks.Registry, index *world.Index) []*domain.Asset {
	return filterLiveDefenders(eligibleDefendersAtHex(attacker, attackerDef, rulebook, registry, index), assetHPTracker)
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
