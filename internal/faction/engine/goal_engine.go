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
	faction.ActiveGoal.TurnsRemaining--
	if faction.ActiveGoal.TurnsRemaining == 0 {
		mutations := []domain.Mutation{
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
		faction.ActiveGoal = nil
		return GoalLock{Type: LockSkip}, mutations
	}
	return GoalLock{Type: LockSkip}, nil
}

func checkLockPlanetarySeizure(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) (GoalLock, []domain.Mutation) {
	goal := faction.ActiveGoal
	if goal.ProcessPhase == 0 {
		return GoalLock{Type: LockRestrictActions, AllowedActions: []string{"Attack"}}, nil
	}
	// Phase 1: occupation — faction must maintain at least one unstealthed asset on TargetWorld.
	if !factionHasUnstealthedAssetOn(faction, goal.TargetWorld) {
		goalID := goal.GoalID
		faction.ActiveGoal = nil
		return GoalLock{Type: LockNone}, []domain.Mutation{
			domain.GoalAbandoned{
				FactionID: faction.ID,
				GoalID:    goalID,
				Cause:     "occupation_failed",
			},
		}
	}
	goal.TurnsRemaining--
	if goal.TurnsRemaining == 0 {
		xp := calcPlanetarySeizureXP(goal.TargetFactionID, factionState)
		var pgTag domain.Tag
		if t, ok := rulebook.Tags["T-011"]; ok {
			pgTag = *t
		}
		mutations := []domain.Mutation{
			domain.TagAdded{
				FactionID: faction.ID,
				Tag:       pgTag,
				Cause:     "goal_completed",
			},
			domain.GoalCompleted{
				FactionID: faction.ID,
				GoalID:    goal.GoalID,
				XPAwarded: xp,
			},
			domain.XPAwarded{
				FactionID: faction.ID,
				Amount:    xp,
				Cause:     "goal_completed",
			},
		}
		faction.ActiveGoal = nil
		return GoalLock{Type: LockNone}, mutations
	}
	return GoalLock{Type: LockNone}, nil
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
	actingFaction.ActiveGoal.Progress += kills
	if actingFaction.ActiveGoal.Progress < actingFaction.Force {
		return nil
	}
	return completeGoal(actingFaction, actingFaction.ActiveGoal.Progress/2)
}

func progressCommercialExpansion(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	kills := countAssetKillsByCategory(actingFaction.ID, domain.StatWealth, mutations, factionState, rulebook)
	if kills == 0 {
		return nil
	}
	actingFaction.ActiveGoal.Progress += kills
	if actingFaction.ActiveGoal.Progress < actingFaction.Wealth {
		return nil
	}
	return completeGoal(actingFaction, actingFaction.ActiveGoal.Progress/2)
}

func progressIntelligenceCoup(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *loader.Rulebook) []domain.Mutation {
	kills := countAssetKillsByCategory(actingFaction.ID, domain.StatCunning, mutations, factionState, rulebook)
	if kills == 0 {
		return nil
	}
	actingFaction.ActiveGoal.Progress += kills
	if actingFaction.ActiveGoal.Progress < actingFaction.Cunning {
		return nil
	}
	return completeGoal(actingFaction, actingFaction.ActiveGoal.Progress/2)
}

// progressPlanetarySeizure handles the Phase 0 → Phase 1 transition. Phase 1
// (occupation countdown and completion) is managed by CheckLock each turn.
func progressPlanetarySeizure(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	goal := actingFaction.ActiveGoal
	if goal.ProcessPhase != 0 {
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
	goal.ProcessPhase = 1
	goal.TurnsRemaining = 3
	return nil
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
	actingFaction.ActiveGoal.Progress += damage
	threshold := actingFaction.Force + actingFaction.Cunning + actingFaction.Wealth
	if actingFaction.ActiveGoal.Progress < threshold {
		return nil
	}
	return completeGoal(actingFaction, 2)
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
		actingFaction.ActiveGoal.Progress = 0
		return nil
	}
	actingFaction.ActiveGoal.Progress++
	if actingFaction.ActiveGoal.Progress < 4 {
		return nil
	}
	return completeGoal(actingFaction, 1)
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
		actingFaction.ActiveGoal.Progress++
	}
	if actingFaction.ActiveGoal.Progress < actingFaction.Cunning {
		return nil
	}
	return completeGoal(actingFaction, 2)
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
	actingFaction.ActiveGoal.Progress += spent
	if actingFaction.ActiveGoal.Progress < 4*actingFaction.Wealth {
		return nil
	}
	return completeGoal(actingFaction, 2)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// completeGoal sets ActiveGoal to nil and returns the completion mutation pair.
func completeGoal(faction *domain.Faction, xp int) []domain.Mutation {
	goalID := faction.ActiveGoal.GoalID
	faction.ActiveGoal = nil
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
