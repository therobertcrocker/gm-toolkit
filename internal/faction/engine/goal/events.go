package goal

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// matchKill reads the removed asset from the state before the batch, where it
// still exists. Any cause counts when the holder caused it (D-14).
func matchKill(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		removed, ok := mutation.(domain.AssetRemoved)
		if !ok || !sideMatches(filters.Owner, removed.FactionID, ctx.faction.ID) || !sideMatches(filters.CausedBy, removed.CausedByFactionID, ctx.faction.ID) {
			continue
		}
		owner := ctx.factionState.Factions[removed.FactionID]
		if owner == nil || owner.Assets[removed.AssetID] == nil {
			continue
		}
		asset := owner.Assets[removed.AssetID]
		definition := ctx.rulebook.Assets[asset.DefinitionID]
		if definition == nil {
			continue
		}
		if len(filters.Category) > 0 && !slices.Contains(filters.Category, definition.Category) {
			continue
		}
		if filters.RatedAbove != "" && definition.MinRating <= ctx.faction.StatScore(filters.RatedAbove) {
			continue
		}
		matches = append(matches, match{amount: 1, world: asset.Location.WorldID})
	}
	return matches
}

func matchDamage(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		var ownerID, causedByID string
		var delta int
		// deliberate: FactionHPDelta is not damage here, because every Base hit emits one beside its BaseHPDelta and would count twice (Decision 1).
		switch hit := mutation.(type) {
		case domain.AssetHPDelta:
			ownerID, causedByID, delta = hit.FactionID, hit.CausedByFactionID, hit.Delta
		case domain.BaseHPDelta:
			ownerID, causedByID, delta = hit.FactionID, hit.CausedByFactionID, hit.Delta
		default:
			continue
		}
		if delta >= 0 || !sideMatches(filters.Owner, ownerID, ctx.faction.ID) || !sideMatches(filters.CausedBy, causedByID, ctx.faction.ID) {
			continue
		}
		matches = append(matches, match{amount: -delta})
	}
	return matches
}

func matchBasePlanted(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		added, ok := mutation.(domain.BaseAdded)
		if !ok || !sideMatches(filters.CausedBy, added.CausedByFactionID, ctx.faction.ID) {
			continue
		}
		world := added.Base.Location.WorldID
		if filters.NewWorld && ctx.faction.HasBaseOn(world) {
			continue
		}
		matches = append(matches, match{amount: 1, world: world})
	}
	return matches
}

func matchInfluenceBought(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		bought, ok := mutation.(domain.InfluenceDelta)
		if !ok || bought.Delta <= 0 || !sideMatches(filters.CausedBy, bought.CausedByFactionID, ctx.faction.ID) {
			continue
		}
		matches = append(matches, match{amount: bought.Delta})
	}
	return matches
}

func matchGovernmentTaken(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		changed, ok := mutation.(domain.WorldGovernmentChanged)
		if !ok || !sideMatches(filters.To, changed.ToFactionID, ctx.faction.ID) {
			continue
		}
		matches = append(matches, match{amount: 1, world: changed.WorldID, mutation: changed})
	}
	return matches
}

// matchFactionDestroyed matches the one HP loss that takes a faction to 0 or
// less, measured against its HP before the batch plus the batch's earlier losses.
// caused_by reads that loss, so a third party's last blow does not match (D-15).
func matchFactionDestroyed(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	lost := make(map[string]int)
	for _, mutation := range ctx.batch {
		delta, ok := mutation.(domain.FactionHPDelta)
		if !ok {
			continue
		}
		target := ctx.factionState.Factions[delta.FactionID]
		if target == nil {
			continue
		}
		before := target.CurrentHP + lost[delta.FactionID]
		lost[delta.FactionID] += delta.Delta
		if before <= 0 || before+delta.Delta > 0 {
			continue
		}
		if filters.Target == targetStartInput && delta.FactionID != ctx.faction.ActiveGoal.TargetFactionID {
			continue
		}
		if !sideMatches(filters.CausedBy, delta.CausedByFactionID, ctx.faction.ID) {
			continue
		}
		matches = append(matches, match{amount: 1, mutation: delta})
	}
	return matches
}

func matchActionTaken(ctx evalContext, filters domain.GoalFilters) []match {
	var matches []match
	for _, mutation := range ctx.batch {
		taken, ok := mutation.(domain.ActionTaken)
		if !ok || !sideMatches(filters.Faction, taken.FactionID, ctx.faction.ID) {
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
