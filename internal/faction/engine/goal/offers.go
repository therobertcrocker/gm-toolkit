package goal

import (
	"fmt"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// Offers lists the goals faction may pick, in goal ID order: each goal whose
// gates are open (D-11, D-12), with the candidates for the start inputs it
// declares (D-3). A goal with an input that has no candidate is left out.
func (goalEngine *GoalEngine) Offers(faction *domain.Faction, factionState *state.FactionState) []domain.GoalOffer {
	var offers []domain.GoalOffer
	for _, goalID := range slices.Sorted(maps.Keys(goalEngine.goals)) {
		goal := goalEngine.goals[goalID].goal
		if !goalEngine.gatesOpen(goal, faction) {
			continue
		}
		offer := domain.GoalOffer{Goal: goal}
		if hasInput(goal, inputTargetFaction) {
			offer.Targets = livingRivals(faction.ID, factionState)
			if len(offer.Targets) == 0 {
				continue
			}
		}
		offers = append(offers, offer)
	}
	return offers
}

// gatesOpen reports whether goal's start gates let faction pick it: every action
// it needs is registered, and its rating, when it has a rating gate, is above the
// faction's record. A goal the faction never completed has no record, so its
// rating gate is open.
func (goalEngine *GoalEngine) gatesOpen(goal *domain.Goal, faction *domain.Faction) bool {
	if goal.Start == nil {
		return true
	}
	for _, action := range goal.Start.NeedsActions {
		if !slices.Contains(goalEngine.actionNames, action) {
			return false
		}
	}
	if gate := goal.Start.Gate; gate != nil {
		stat := gate.RatingAboveLastCompletion
		recorded, completed := faction.CompletionRatings[goal.ID]
		if completed && faction.StatScore(stat) <= recorded[stat] {
			return false
		}
	}
	return true
}

func livingRivals(factionID string, factionState *state.FactionState) []*domain.Faction {
	var rivals []*domain.Faction
	for _, rivalID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		if rival := factionState.Factions[rivalID]; isLivingRival(rival, factionID) {
			rivals = append(rivals, rival)
		}
	}
	return rivals
}

// Pick builds the mutation that gives faction the goal it chose. The choice must
// name a goal in offers, with a target from that offer when the goal takes one and
// none when it does not; anything else is a collector bug. A goal whose progress
// leaves out the assets held stealthed at the start records them here (D-10).
func (goalEngine *GoalEngine) Pick(faction *domain.Faction, offers []domain.GoalOffer, choice domain.GoalChoice) (domain.GoalPicked, error) {
	index := slices.IndexFunc(offers, func(offer domain.GoalOffer) bool { return offer.Goal.ID == choice.GoalID })
	if index < 0 {
		return domain.GoalPicked{}, fmt.Errorf("goal %q is not on offer to %s", choice.GoalID, faction.ID)
	}
	offer := offers[index]
	targetOffered := slices.ContainsFunc(offer.Targets, func(target *domain.Faction) bool { return target.ID == choice.TargetFactionID })
	switch {
	case len(offer.Targets) == 0 && choice.TargetFactionID != "":
		return domain.GoalPicked{}, fmt.Errorf("goal %q takes no target faction, got %q", choice.GoalID, choice.TargetFactionID)
	case len(offer.Targets) > 0 && !targetOffered:
		return domain.GoalPicked{}, fmt.Errorf("goal %q needs a target faction on offer, got %q", choice.GoalID, choice.TargetFactionID)
	}
	picked := domain.GoalPicked{FactionID: faction.ID, GoalID: offer.Goal.ID, TargetFactionID: choice.TargetFactionID}
	if progress := offer.Goal.Progress; progress != nil && progress.Filters.ExcludeHeldAtStart {
		picked.StealthedAtStart = stealthedAssetIDs(faction)
	}
	return picked, nil
}

func stealthedAssetIDs(faction *domain.Faction) []string {
	var assetIDs []string
	for _, asset := range domain.SortedAssets(faction) {
		if asset.Stealthy {
			assetIDs = append(assetIDs, asset.ID)
		}
	}
	return assetIDs
}
