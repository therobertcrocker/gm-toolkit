package goal

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// matchKill reads the removed asset from the state before the batch, where it
// still exists. Any cause counts when the holder caused it (D-14).
func matchKill(evaluation evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range evaluation.batch {
		removed, ok := mutation.(domain.AssetRemoved)
		if !ok || !sideMatches(filters.Owner, removed.FactionID, evaluation.faction.ID) || !sideMatches(filters.CausedBy, removed.CausedByFactionID, evaluation.faction.ID) {
			continue
		}
		owner := evaluation.factionState.Factions[removed.FactionID]
		if owner == nil || owner.Assets[removed.AssetID] == nil {
			continue
		}
		asset := owner.Assets[removed.AssetID]
		definition := evaluation.rulebook.Assets[asset.DefinitionID]
		if definition == nil {
			continue
		}
		if len(filters.Category) > 0 && !slices.Contains(filters.Category, definition.Category) {
			continue
		}
		if filters.RatedAbove != "" && definition.MinRating <= evaluation.faction.StatScore(filters.RatedAbove) {
			continue
		}
		matches = append(matches, match{amount: 1, world: asset.Location.WorldID})
	}
	return matches
}

// matchDamage counts the hit points a hit takes off: the damage, capped at the HP
// its asset or Base had left, read from the state before the batch plus the
// batch's earlier deltas.
func matchDamage(evaluation evalContext, filters domain.GoalFilters) []match {
	var matches []match
	remaining := make(map[string]int)
	for _, mutation := range evaluation.batch {
		var ownerID, causedByID, hitKey string
		var delta, hpBefore int
		var known bool
		// deliberate: FactionHPDelta is not damage here, because every Base hit emits one beside its BaseHPDelta and would count twice (Decision 1).
		switch hit := mutation.(type) {
		case domain.AssetHPDelta:
			ownerID, causedByID, delta = hit.FactionID, hit.CausedByFactionID, hit.Delta
			hitKey = "asset/" + hit.AssetID
			hpBefore, known = assetHP(evaluation.factionState, hit.FactionID, hit.AssetID)
		case domain.BaseHPDelta:
			ownerID, causedByID, delta = hit.FactionID, hit.CausedByFactionID, hit.Delta
			hitKey = "base/" + hit.BaseID
			hpBefore, known = baseHP(evaluation.factionState, hit.FactionID, hit.BaseID)
		default:
			continue
		}
		hp, seen := remaining[hitKey]
		if !seen {
			hp = hpBefore
		}
		remaining[hitKey] = hp + delta
		if delta >= 0 || !sideMatches(filters.Owner, ownerID, evaluation.faction.ID) || !sideMatches(filters.CausedBy, causedByID, evaluation.faction.ID) {
			continue
		}
		taken := -delta
		// deliberate: overkill does not count, not the damage rolled, because the goal asks for hit points of damage inflicted (Robert's rules call at the A.001.8 close).
		if known {
			taken = min(taken, max(hp, 0))
		}
		if taken > 0 {
			matches = append(matches, match{amount: taken})
		}
	}
	return matches
}

func assetHP(factionState *state.FactionState, factionID, assetID string) (int, bool) {
	owner := factionState.Factions[factionID]
	if owner == nil || owner.Assets[assetID] == nil {
		return 0, false
	}
	return owner.Assets[assetID].CurrentHP, true
}

func baseHP(factionState *state.FactionState, factionID, baseID string) (int, bool) {
	owner := factionState.Factions[factionID]
	if owner == nil {
		return 0, false
	}
	for _, base := range owner.Bases {
		if base.ID == baseID {
			return base.CurrentHP, true
		}
	}
	return 0, false
}

func matchBasePlanted(evaluation evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range evaluation.batch {
		added, ok := mutation.(domain.BaseAdded)
		if !ok || !sideMatches(filters.CausedBy, added.CausedByFactionID, evaluation.faction.ID) {
			continue
		}
		world := added.Base.Location.WorldID
		if filters.NewWorld && evaluation.faction.HasBaseOn(world) {
			continue
		}
		matches = append(matches, match{amount: 1, world: world})
	}
	return matches
}

func matchInfluenceBought(evaluation evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range evaluation.batch {
		bought, ok := mutation.(domain.InfluenceDelta)
		if !ok || bought.Delta <= 0 || !sideMatches(filters.CausedBy, bought.CausedByFactionID, evaluation.faction.ID) {
			continue
		}
		matches = append(matches, match{amount: bought.Delta})
	}
	return matches
}

func matchGovernmentTaken(evaluation evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range evaluation.batch {
		changed, ok := mutation.(domain.WorldGovernmentChanged)
		if !ok || !sideMatches(filters.To, changed.ToFactionID, evaluation.faction.ID) {
			continue
		}
		matches = append(matches, match{amount: 1, world: changed.WorldID, mutation: changed})
	}
	return matches
}

// matchFactionDestroyed matches the one HP loss that takes a faction to 0 or
// less, measured against its HP before the batch plus the batch's earlier losses.
// caused_by reads that loss, so a third party's last blow does not match (D-15).
func matchFactionDestroyed(evaluation evalContext, filters domain.GoalFilters) []match {
	var matches []match
	lost := make(map[string]int)
	for _, mutation := range evaluation.batch {
		delta, ok := mutation.(domain.FactionHPDelta)
		if !ok {
			continue
		}
		target := evaluation.factionState.Factions[delta.FactionID]
		if target == nil {
			continue
		}
		before := target.CurrentHP + lost[delta.FactionID]
		lost[delta.FactionID] += delta.Delta
		if before <= 0 || before+delta.Delta > 0 {
			continue
		}
		if filters.Target == targetStartInput && delta.FactionID != evaluation.faction.ActiveGoal.TargetFactionID {
			continue
		}
		if !sideMatches(filters.CausedBy, delta.CausedByFactionID, evaluation.faction.ID) {
			continue
		}
		matches = append(matches, match{amount: 1, mutation: delta})
	}
	return matches
}

func matchActionTaken(evaluation evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range evaluation.batch {
		taken, ok := mutation.(domain.ActionTaken)
		if !ok || !sideMatches(filters.Faction, taken.FactionID, evaluation.faction.ID) {
			continue
		}
		if len(filters.Action) > 0 && !slices.Contains(filters.Action, taken.Action) {
			continue
		}
		if slices.Contains(filters.ActionNot, taken.Action) {
			continue
		}
		matches = append(matches, match{amount: 1})
	}
	return matches
}
