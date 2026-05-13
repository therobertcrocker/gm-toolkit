package tags

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const FanaticalTagID = "T-005"

// FanaticalRollResultHook implements hooks.RollResultHook for the Fanatical tag.
// Automatically rerolls any die showing a 1.
type FanaticalRollResultHook struct{}

func (hook *FanaticalRollResultHook) OnRollResult(_ hooks.RollContext, result hooks.RollResult, _ *state.FactionState, _ *rulebook.Rulebook) hooks.RerollDirective {
	var indices []int
	for i, die := range result.Dice {
		if die == 1 {
			indices = append(indices, i)
		}
	}
	return hooks.RerollDirective{
		Source:   "tag:Fanatical",
		Indices:  indices,
		Elective: false,
	}
}

// FanaticalTieResolver implements hooks.TieResolver for the Fanatical tag.
// Fanatical factions always lose ties during attacks.
type FanaticalTieResolver struct {
	FactionID string
}

func (resolver *FanaticalTieResolver) ResolveTie(ctx hooks.RollContext, _ *state.FactionState) hooks.TieOutcome {
	if ctx.Actor.ID == resolver.FactionID {
		return hooks.TieDefenderWins
	}
	return hooks.TieAttackerWins
}

type FanaticalHandler struct{}

func (FanaticalHandler) TagID() string { return FanaticalTagID }

func (FanaticalHandler) Apply(faction *domain.Faction, hookRegistry *hooks.Registry) {
	hookRegistry.RegisterRollResultHook(
		hooks.FactionScope(faction.ID),
		"fanatical-roll-result",
		&FanaticalRollResultHook{},
	)
	hookRegistry.RegisterTieResolver(
		hooks.FactionScope(faction.ID),
		"fanatical-tie-resolver",
		&FanaticalTieResolver{FactionID: faction.ID},
	)
}
