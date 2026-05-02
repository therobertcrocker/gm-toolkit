package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// LockType classifies the constraint that a faction's active goal imposes on
// its current turn.
type LockType int

const (
	LockNone            LockType = iota // normal flow
	LockSkip                            // skip faction entirely (Change Homeworld in-progress)
	LockRestrictActions                 // limit available actions (Seize Planet combat phase)
)

// GoalLock describes how the active goal constrains this faction's turn.
type GoalLock struct {
	Type           LockType
	AllowedActions []string // non-nil only when Type == LockRestrictActions
}

// GoalEngine evaluates and advances faction goal state.
type GoalEngine struct{}

func NewGoalEngine() *GoalEngine { return &GoalEngine{} }

// CheckLock evaluates the faction's active goal and returns the appropriate
// lock state. For LockSkip (Change Homeworld) it decrements TurnsRemaining
// and emits completion mutations when the countdown reaches zero. For Planetary
// Seizure Phase 1 (occupation) it also handles countdown and completion.
// Must be called before bookkeeping and action selection each turn.
func (ge *GoalEngine) CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) (GoalLock, []domain.Mutation) {
	if faction.ActiveGoal == nil {
		return GoalLock{Type: LockNone}, nil
	}
	switch faction.ActiveGoal.GoalID {
	case "G-012":
		return checkLockChangeHomeworld(faction)
	case "G-004":
		return checkLockPlanetarySeizure(faction, factionState, rulebook)
	}
	return GoalLock{Type: LockNone}, nil
}

func checkLockChangeHomeworld(faction *domain.Faction) (GoalLock, []domain.Mutation) {
	newTurns := faction.ActiveGoal.TurnsRemaining - 1
	tick := domain.GoalTurnsTick{
		FactionID: faction.ID,
		GoalID:    faction.ActiveGoal.GoalID,
		Cause:     "change_homeworld_transit",
	}
	if newTurns == 0 {
		return GoalLock{Type: LockSkip}, []domain.Mutation{
			tick,
			domain.HomeworldChanged{
				FactionID: faction.ID,
				FromWorld: faction.Homeworld,
				ToWorld:   faction.ActiveGoal.TargetWorld,
				Cause:     "goal_completed",
			},
			domain.GoalCompleted{
				FactionID: faction.ID,
				GoalID:    faction.ActiveGoal.GoalID,
				XPAwarded: 0,
			},
		}
	}
	return GoalLock{Type: LockSkip}, []domain.Mutation{tick}
}

func checkLockPlanetarySeizure(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) (GoalLock, []domain.Mutation) {
	goal := faction.ActiveGoal
	if goal.ProcessPhase == 0 {
		return GoalLock{Type: LockNone}, nil
	}
	if goal.ProcessPhase == 1 {
		return GoalLock{Type: LockRestrictActions, AllowedActions: []string{"Attack"}}, nil
	}
	// Phase 2: occupation.
	if !factionHasUnstealthedAssetOn(faction, goal.TargetWorld) {
		return GoalLock{Type: LockNone}, []domain.Mutation{
			domain.GoalAbandoned{
				FactionID: faction.ID,
				GoalID:    goal.GoalID,
				Cause:     "occupation_failed",
			},
		}
	}
	tick := domain.GoalTurnsTick{
		FactionID: faction.ID,
		GoalID:    goal.GoalID,
		Cause:     "planetary_seizure_occupation",
	}
	newTurns := goal.TurnsRemaining - 1
	if newTurns == 0 {
		xp := calcPlanetarySeizureXP(goal.TargetFactionID, factionState)
		var pgTag domain.Tag
		if t, ok := rulebook.Tags["T-011"]; ok {
			pgTag = *t
		}
		return GoalLock{Type: LockNone}, []domain.Mutation{
			tick,
			domain.TagAdded{FactionID: faction.ID, Tag: pgTag, Cause: "goal_completed"},
			domain.GoalCompleted{FactionID: faction.ID, GoalID: goal.GoalID, XPAwarded: xp},
			domain.XPAwarded{FactionID: faction.ID, Amount: xp, Cause: "goal_completed"},
		}
	}
	return GoalLock{Type: LockNone}, []domain.Mutation{tick}
}

// UpdateProgress inspects the acting faction's mutation list for goal-relevant
// events, advances ActiveGoal.Progress, and returns supplemental mutations
// (GoalCompleted, XPAwarded, TagAdded) if the goal is now complete.
// Must be called before MutationEngine.Apply so destructive mutations have not
// yet fired (Destroy the Foe XP reads live target faction state).
func (ge *GoalEngine) UpdateProgress(
	actingFactionID string,
	mutations []domain.Mutation,
	factionState *state.FactionState,
	rulebook *loader.Rulebook,
) []domain.Mutation {
	actingFaction, ok := factionState.Factions[actingFactionID]
	if !ok || actingFaction.ActiveGoal == nil {
		return nil
	}
	switch actingFaction.ActiveGoal.GoalID {
	case "G-001":
		return progressMilitaryConquest(actingFaction, mutations, factionState, rulebook)
	case "G-002":
		return progressCommercialExpansion(actingFaction, mutations, factionState, rulebook)
	case "G-003":
		return progressIntelligenceCoup(actingFaction, mutations, factionState, rulebook)
	case "G-004":
		return progressPlanetarySeizure(actingFaction, mutations, factionState)
	case "G-005":
		return progressExpandInfluence(actingFaction, mutations, factionState)
	case "G-006":
		return progressBloodTheEnemy(actingFaction, mutations)
	case "G-007":
		return progressPeaceableKingdom(actingFaction, mutations)
	case "G-008":
		return progressDestroyTheFoe(actingFaction, mutations, factionState)
	case "G-009":
		return progressInsideEnemyTerritory(actingFaction, mutations, factionState)
	case "G-010":
		return progressInvincibleValor(actingFaction, mutations, factionState, rulebook)
	case "G-011":
		return progressWealthOfWorlds(actingFaction, mutations)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Goal handlers
// ---------------------------------------------------------------------------

func progressMilitaryConquest(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	kills := countAssetKillsByCategory(actingFaction.ID, domain.StatForce, mutations, factionState, rulebook)
	if kills == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + kills
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     kills,
		Cause:     "military_conquest",
	}
	if newProgress < actingFaction.Force {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, newProgress/2)...)
}

func progressCommercialExpansion(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	kills := countAssetKillsByCategory(actingFaction.ID, domain.StatWealth, mutations, factionState, rulebook)
	if kills == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + kills
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     kills,
		Cause:     "commercial_expansion",
	}
	if newProgress < actingFaction.Wealth {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, newProgress/2)...)
}

func progressIntelligenceCoup(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	kills := countAssetKillsByCategory(actingFaction.ID, domain.StatCunning, mutations, factionState, rulebook)
	if kills == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + kills
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     kills,
		Cause:     "intelligence_coup",
	}
	if newProgress < actingFaction.Cunning {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, newProgress/2)...)
}

// progressPlanetarySeizure handles the Phase 1 → Phase 2 transition. Phase 2
// (occupation countdown and completion) is managed by CheckLock each turn.
func progressPlanetarySeizure(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	goal := actingFaction.ActiveGoal
	if goal.ProcessPhase != 1 {
		return nil
	}
	// Collect asset IDs removed this turn.
	removedIDs := make(map[string]bool)
	for _, mutation := range mutations {
		if v, ok := mutation.(domain.AssetRemoved); ok {
			removedIDs[v.AssetID] = true
		}
	}
	// Check for surviving rival unstealthed assets on TargetWorld.
	for factionID, faction := range factionState.Factions {
		if factionID == actingFaction.ID {
			continue
		}
		for _, asset := range faction.Assets {
			if asset.Location == goal.TargetWorld && !asset.Stealthy && !removedIDs[asset.ID] {
				return nil
			}
		}
	}
	// No rivals remain — transition to occupation (3 turns).
	return []domain.Mutation{
		domain.GoalPhaseAdvanced{
			FactionID:      actingFaction.ID,
			GoalID:         goal.GoalID,
			ProcessPhase:   2,
			TurnsRemaining: 3,
			Cause:          "planetary_seizure_phase_advance",
		},
	}
}

func progressExpandInfluence(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	for _, mutation := range mutations {
		v, ok := mutation.(domain.BaseAdded)
		if !ok || v.CausedByFactionID != actingFaction.ID {
			continue
		}
		if factionHasBaseOn(actingFaction, v.Base.Location) {
			continue
		}
		xp := 1
		if worldHasRivalPresence(v.Base.Location, actingFaction.ID, factionState) {
			xp = 2
		}
		return completeGoal(actingFaction, xp)
	}
	return nil
}

func progressBloodTheEnemy(actingFaction *domain.Faction, mutations []domain.Mutation) []domain.Mutation {
	damage := 0
	for _, mutation := range mutations {
		switch v := mutation.(type) {
		case domain.AssetHPDelta:
			if v.CausedByFactionID == actingFaction.ID && v.FactionID != actingFaction.ID && v.Delta < 0 {
				damage += -v.Delta
			}
		case domain.BaseHPDelta:
			if v.CausedByFactionID == actingFaction.ID && v.FactionID != actingFaction.ID && v.Delta < 0 {
				damage += -v.Delta
			}
		}
	}
	if damage == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + damage
	threshold := actingFaction.Force + actingFaction.Cunning + actingFaction.Wealth
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     damage,
		Cause:     "blood_the_enemy",
	}
	if newProgress < threshold {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 2)...)
}

func progressPeaceableKingdom(actingFaction *domain.Faction, mutations []domain.Mutation) []domain.Mutation {
	attacked := false
	for _, mutation := range mutations {
		switch v := mutation.(type) {
		case domain.AssetHPDelta:
			if v.Cause == "attack" && v.CausedByFactionID == actingFaction.ID {
				attacked = true
			}
		case domain.BaseHPDelta:
			if v.Cause == "attack" && v.CausedByFactionID == actingFaction.ID {
				attacked = true
			}
		case domain.AssetStealthCleared:
			// Attacker stealth loss signals an attack was taken.
			if v.Cause == "attack" && v.CausedByFactionID == actingFaction.ID && v.FactionID == actingFaction.ID {
				attacked = true
			}
		}
	}
	if attacked {
		if actingFaction.ActiveGoal.Progress == 0 {
			return nil
		}
		return []domain.Mutation{domain.GoalProgressed{
			FactionID: actingFaction.ID,
			GoalID:    actingFaction.ActiveGoal.GoalID,
			Delta:     -actingFaction.ActiveGoal.Progress,
			Cause:     "peaceable_kingdom_reset",
		}}
	}
	newProgress := actingFaction.ActiveGoal.Progress + 1
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     1,
		Cause:     "peaceable_kingdom",
	}
	if newProgress < 4 {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 1)...)
}

func progressDestroyTheFoe(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	targetFaction, ok := factionState.Factions[actingFaction.ActiveGoal.TargetFactionID]
	if !ok {
		return nil
	}
	// Calculate effective HP after pending mutations (decision #14: before destructive mutations apply).
	effectiveHP := targetFaction.CurrentHP
	for _, mutation := range mutations {
		if v, ok := mutation.(domain.FactionHPDelta); ok && v.FactionID == targetFaction.ID {
			effectiveHP += v.Delta
		}
	}
	if effectiveHP > 0 {
		return nil
	}
	avg := (targetFaction.Force + targetFaction.Cunning + targetFaction.Wealth) / 3
	return completeGoal(actingFaction, 1+avg)
}

// progressInsideEnemyTerritory counts AssetStealthApplied events on worlds
// where a rival holds a Planetary Government tag. Approximation: a rival "holds
// PG on a world" when they have both the T-011 tag and a Base on that world.
func progressInsideEnemyTerritory(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	gained := 0
	for _, mutation := range mutations {
		v, ok := mutation.(domain.AssetStealthApplied)
		if !ok || v.FactionID != actingFaction.ID {
			continue
		}
		asset := findAsset(actingFaction, v.AssetID)
		if asset == nil {
			continue
		}
		if !rivalHasPlanetaryGovernmentOnWorld(actingFaction.ID, asset.Location, factionState) {
			continue
		}
		gained++
	}
	if gained == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + gained
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     gained,
		Cause:     "inside_enemy_territory",
	}
	if newProgress < actingFaction.Cunning {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 2)...)
}

func progressInvincibleValor(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	for _, mutation := range mutations {
		v, ok := mutation.(domain.AssetRemoved)
		if !ok || v.Cause != "attack" || v.CausedByFactionID != actingFaction.ID || v.FactionID == actingFaction.ID {
			continue
		}
		rivalFaction, ok := factionState.Factions[v.FactionID]
		if !ok {
			continue
		}
		asset := findAsset(rivalFaction, v.AssetID)
		if asset == nil {
			continue
		}
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			continue
		}
		if def.Category == domain.StatForce && def.MinRating > actingFaction.Force {
			return completeGoal(actingFaction, 2)
		}
	}
	return nil
}

func progressWealthOfWorlds(actingFaction *domain.Faction, mutations []domain.Mutation) []domain.Mutation {
	spent := 0
	for _, mutation := range mutations {
		if v, ok := mutation.(domain.InfluenceDelta); ok && v.CausedByFactionID == actingFaction.ID && v.Delta > 0 {
			spent += v.Delta
		}
	}
	if spent == 0 {
		return nil
	}
	newProgress := actingFaction.ActiveGoal.Progress + spent
	progressed := domain.GoalProgressed{
		FactionID: actingFaction.ID,
		GoalID:    actingFaction.ActiveGoal.GoalID,
		Delta:     spent,
		Cause:     "wealth_of_worlds",
	}
	if newProgress < 4*actingFaction.Wealth {
		return []domain.Mutation{progressed}
	}
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 2)...)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// completeGoal sets ActiveGoal to nil and returns the completion mutation pair.
func completeGoal(faction *domain.Faction, xp int) []domain.Mutation {
	goalID := faction.ActiveGoal.GoalID
	return []domain.Mutation{
		domain.GoalCompleted{
			FactionID: faction.ID,
			GoalID:    goalID,
			XPAwarded: xp,
			Cause:     "goal_completed",
		},
		domain.XPAwarded{
			FactionID: faction.ID,
			Amount:    xp,
			Cause:     "goal_completed",
		},
	}
}

// countAssetKillsByCategory counts rival AssetRemoved mutations caused by
// attack from actingFactionID where the removed asset's category matches.
func countAssetKillsByCategory(actingFactionID string, category domain.FactionStat, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) int {
	count := 0
	for _, mutation := range mutations {
		v, ok := mutation.(domain.AssetRemoved)
		if !ok || v.Cause != "attack" || v.CausedByFactionID != actingFactionID || v.FactionID == actingFactionID {
			continue
		}
		rivalFaction, ok := factionState.Factions[v.FactionID]
		if !ok {
			continue
		}
		asset := findAsset(rivalFaction, v.AssetID)
		if asset == nil {
			continue
		}
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			continue
		}
		if def.Category == category {
			count++
		}
	}
	return count
}

func findAsset(faction *domain.Faction, assetID string) *domain.Asset {
	for _, asset := range faction.Assets {
		if asset.ID == assetID {
			return asset
		}
	}
	return nil
}

func factionHasUnstealthedAssetOn(faction *domain.Faction, world string) bool {
	for _, asset := range faction.Assets {
		if asset.Location == world && !asset.Stealthy {
			return true
		}
	}
	return false
}

func factionHasBaseOn(faction *domain.Faction, world string) bool {
	for _, base := range faction.Bases {
		if base.Location == world {
			return true
		}
	}
	return false
}

func worldHasRivalPresence(world, actingFactionID string, factionState *state.FactionState) bool {
	for factionID, faction := range factionState.Factions {
		if factionID == actingFactionID {
			continue
		}
		for _, asset := range faction.Assets {
			if asset.Location == world {
				return true
			}
		}
		for _, base := range faction.Bases {
			if base.Location == world {
				return true
			}
		}
	}
	return false
}

// rivalHasPlanetaryGovernmentOnWorld returns true when any rival faction holds
// the T-011 Planetary Government tag and has a Base on world. This approximates
// "world controlled by a rival government" within the current domain model.
func rivalHasPlanetaryGovernmentOnWorld(actingFactionID, world string, factionState *state.FactionState) bool {
	for factionID, faction := range factionState.Factions {
		if factionID == actingFactionID {
			continue
		}
		if !factionHasPlanetaryGovernmentTag(faction) {
			continue
		}
		for _, base := range faction.Bases {
			if base.Location == world {
				return true
			}
		}
	}
	return false
}

func factionHasPlanetaryGovernmentTag(faction *domain.Faction) bool {
	for _, tag := range faction.Tags {
		if tag.ID == "T-011" {
			return true
		}
	}
	return false
}

func calcPlanetarySeizureXP(targetFactionID string, factionState *state.FactionState) int {
	if targetFactionID == "" {
		return 1
	}
	targetFaction, ok := factionState.Factions[targetFactionID]
	if !ok {
		return 1
	}
	avg := (targetFaction.Force + targetFaction.Cunning + targetFaction.Wealth) / 3
	xp := avg / 2
	if xp < 1 {
		return 1
	}
	return xp
}
