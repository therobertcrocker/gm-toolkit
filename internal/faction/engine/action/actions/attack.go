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
	// The tracker reflects what this sequence did before its mutations apply, so
	// mid-sequence re-checks see each asset's HP and whether it left the fight.
	tracker := newSequenceTracker()
	baseHPTracker := make(map[string]int)
	// Stealth-cleared tracks assets already cleared this sequence; defenders can
	// face multiple matchups and we emit the mutation only once.
	stealthCleared := make(map[string]bool)

	for _, attacker := range attack.attackers {
		attackerDef, ok := rulebook.Assets[attacker.DefinitionID]
		if !ok || attackerDef.Attack == nil {
			continue
		}

		// Re-check: skip if destroyed or taken out of the fight earlier in this sequence.
		if !tracker.live(attacker) {
			continue
		}

		defenders := liveDefendersAtHex(attacker, attackerDef, tracker, rulebook, attack.registry, attack.index)
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

		attack.mutations = append(attack.mutations, attack.fireMatchup(domain.TriggerOwnAttack, faction, attacker, attackerDef, defenderFaction, defender, factionState, rulebook).Mutations...)

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
				if err := attack.applyAssetDamage(factionState, rulebook, defenderFaction.ID, faction.ID, defender, 0, tracker, attacker, false, "attack"); err != nil {
					return err
				}
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
					} else if err := attack.applyAssetDamage(factionState, rulebook, defenderFaction.ID, faction.ID, defender, damage, tracker, attacker, false, "attack"); err != nil {
						return err
					}
				} else if err := attack.applyAssetDamage(factionState, rulebook, defenderFaction.ID, faction.ID, defender, damage, tracker, attacker, false, "attack"); err != nil {
					return err
				}
			}
		}

		// deliberate: a disabled defender rolls its defense but deals no counter and fires no own_defense_win entry, because the book makes it useless (Effort 3 Decision 2).
		// deliberate: read state, not the tracker, because on a tie both sides hit at once: a defender this matchup killed or intercepted still counters (Effort 2 Decision 9).
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
				if err := attack.applyAssetDamage(factionState, rulebook, faction.ID, defenderFaction.ID, attacker, counterDamage, tracker, defender, true, "attack"); err != nil {
					return err
				}
			}
		}

		// deliberate: own_defense_win needs an outright win, not a tie, because on a tie the attack also landed and the defense did not succeed (Decision 13).
		if defenderWins && !attackerWins && defenderActive {
			outcome := attack.fireMatchup(domain.TriggerOwnDefenseWin, defenderFaction, defender, defenderDef, faction, attacker, factionState, rulebook)
			attack.mutations = append(attack.mutations, outcome.Mutations...)
			// deliberate: no prompt when the attacker has no damage and no counter, because its self-attack could do nothing (Decision Record 9).
			canHurtItself := attackerDef.Attack.Damage != nil || attackerDef.Counter != nil
			if outcome.Counterstrike && canHurtItself && tracker.live(attacker) {
				confirmed, err := attack.collector.ConfirmCounterstrike(defenderFaction, defender, attacker)
				if err != nil {
					return fmt.Errorf("attack: counterstrike prompt: %w", err)
				}
				if confirmed {
					if err := attack.counterstrike(faction, attacker, attackerDef, defenderFaction, defender, tracker, factionState, rulebook); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// counterstrike turns the attacker on itself (D-18): its attack roll against its
// own defense. A hit deals its attack damage; a miss deals its counter damage, if
// it has one. The cause is not "attack", so no own_attack_hit fires on it.
func (attack *AttackAction) counterstrike(faction *domain.Faction, attacker *domain.Asset, attackerDef *domain.AssetDefinition, defenderFaction *domain.Faction, source *domain.Asset, tracker *sequenceTracker, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
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
		return nil
	}
	return attack.applyAssetDamage(factionState, rulebook, faction.ID, defenderFaction.ID, attacker, damage.Roll(attack.roller), tracker, source, false, string(domain.SpecialCounterstrike))
}

func (attack *AttackAction) Output() ([]domain.Mutation, error) {
	return attack.mutations, nil
}

// fireMatchup raises the matchup route for the declaring asset. A nil effects
// engine (action tests) raises nothing.
func (attack *AttackAction) fireMatchup(trigger domain.EffectTrigger, faction *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, opponent *domain.Faction, opponentAsset *domain.Asset, factionState *state.FactionState, rulebook *rulebook.Rulebook) effect.MatchupOutcome {
	if attack.effects == nil {
		return effect.MatchupOutcome{}
	}
	return attack.effects.FireMatchup(trigger, effect.MatchupContext{
		Holder:        effect.Holder{Faction: faction, Asset: asset, Def: def},
		Asset:         asset,
		Opponent:      opponent,
		OpponentAsset: opponentAsset,
		FactionState:  factionState,
	}, rulebook)
}

// applyAssetDamage emits AssetHPDelta and, when the hit is lethal, the mutations
// of the Cat 6 interceptor that takes it, or else AssetRemoved. factionID is the
// owner of the damaged asset; causedByFactionID is the striker's. Every mutation
// goes through the tracker, so later re-checks see the outcome.
func (attack *AttackAction) applyAssetDamage(factionState *state.FactionState, rulebook *rulebook.Rulebook, factionID, causedByFactionID string, asset *domain.Asset, damage int, tracker *sequenceTracker, causedByAsset *domain.Asset, counter bool, cause string) error {
	causedByAssetID := ""
	if causedByAsset != nil {
		causedByAssetID = causedByAsset.ID
	}
	attack.emit(tracker, domain.AssetHPDelta{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Delta:             -damage,
		Cause:             cause,
		CausedByFactionID: causedByFactionID,
		CausedByAssetID:   causedByAssetID,
		Counter:           counter,
	})
	hpAfter := tracker.hpOf(asset)
	if hpAfter > 0 {
		return nil
	}
	intercepted, err := dispatch.InterceptLethal(attack.registry, hooks.LethalContext{
		Target:       asset,
		TargetDef:    rulebook.Assets[asset.DefinitionID],
		HPAfter:      hpAfter,
		Damage:       damage,
		Striker:      causedByAsset,
		FactionState: factionState,
		Batch:        attack.mutations,
	}, attack.collector)
	if err != nil {
		return fmt.Errorf("attack: lethal interception: %w", err)
	}
	if intercepted != nil {
		attack.emit(tracker, intercepted...)
		return nil
	}
	attack.emit(tracker, domain.AssetRemoved{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Cause:             cause,
		CausedByFactionID: causedByFactionID,
	})
	return nil
}

// emit appends mutations to the action's output and folds them into the tracker.
func (attack *AttackAction) emit(tracker *sequenceTracker, mutations ...domain.Mutation) {
	attack.mutations = append(attack.mutations, mutations...)
	tracker.record(mutations)
}

// sequenceTracker is what the Attack sequence has done to each asset before its
// batch applies: the HP change, and the assets that left the fight (removed, or
// moved to another faction).
type sequenceTracker struct {
	hp   map[string]int
	gone map[string]bool
}

func newSequenceTracker() *sequenceTracker {
	return &sequenceTracker{hp: make(map[string]int), gone: make(map[string]bool)}
}

// record folds mutations in: HP deltas add up, and a removal or an owner change
// takes the asset out of the rest of the sequence (Effort 2 Decision 3).
func (tracker *sequenceTracker) record(mutations []domain.Mutation) {
	for _, mutation := range mutations {
		switch m := mutation.(type) {
		case domain.AssetHPDelta:
			tracker.hp[m.AssetID] += m.Delta
		case domain.AssetRemoved:
			tracker.gone[m.AssetID] = true
		case domain.AssetOwnerChanged:
			tracker.gone[m.AssetID] = true
		}
	}
}

// hpOf is asset's HP once this sequence's mutations apply.
func (tracker *sequenceTracker) hpOf(asset *domain.Asset) int {
	return asset.CurrentHP + tracker.hp[asset.ID]
}

// live reports whether asset can still fight in this sequence.
func (tracker *sequenceTracker) live(asset *domain.Asset) bool {
	return !tracker.gone[asset.ID] && tracker.hpOf(asset) > 0
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

// liveDefendersOnWorld filters eligibleDefendersOnWorld to the assets still live
// in the current Resolve pass.
func liveDefendersOnWorld(attackerFactionID, worldID string, tracker *sequenceTracker, index *world.Index) []*domain.Asset {
	return filterLiveDefenders(eligibleDefendersOnWorld(attackerFactionID, worldID, index), tracker)
}

// liveDefendersAtHex filters eligibleDefendersAtHex to the assets still live in
// the current Resolve pass. Used when the attacker is mid-flight (empty WorldID).
func liveDefendersAtHex(attacker *domain.Asset, attackerDef *domain.AssetDefinition, tracker *sequenceTracker, rulebook *rulebook.Rulebook, registry *hooks.Registry, index *world.Index) []*domain.Asset {
	return filterLiveDefenders(eligibleDefendersAtHex(attacker, attackerDef, rulebook, registry, index), tracker)
}

func filterLiveDefenders(defenders []*domain.Asset, tracker *sequenceTracker) []*domain.Asset {
	live := make([]*domain.Asset, 0, len(defenders))
	for _, defender := range defenders {
		if tracker.live(defender) {
			live = append(live, defender)
		}
	}
	return live
}

// ownerFaction returns the faction that owns the given asset, or nil.
func ownerFaction(factionState *state.FactionState, asset *domain.Asset) *domain.Faction {
	return factionState.Factions[asset.OwnerID]
}
