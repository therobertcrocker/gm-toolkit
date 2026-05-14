package tags

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const WarlikeTagID = "T-020"

// WarlikeRollModifier implements hooks.RollModifier for the Warlike tag.
// Once per turn, adds +1d10 to Force attack rolls, keeping the highest die.
type WarlikeRollModifier struct{}

func (modifier *WarlikeRollModifier) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	if ctx.Phase != hooks.PhaseAttack || ctx.Attribute != string(domain.StatForce) {
		return nil
	}
	return []hooks.ModifierOffer{{
		Source:      "tag:Warlike",
		Description: "roll an extra d10, keep highest",
		BudgetKey:   "tag:Warlike",
		// keep-highest: extra die trimmed post-roll via RollState.SetKeepHighest
		// keep-highest: SetKeepHighest(1) assumes a 1-die base pool; if the base
		// pool ever grows, this trim count must be updated alongside it.
		Apply: func(rollState *hooks.RollState) {
			rollState.AddDie(10)
			rollState.SetKeepHighest(1)
		},
	}}
}

type WarlikeHandler struct{}

func (WarlikeHandler) TagID() string { return WarlikeTagID }

func (WarlikeHandler) Apply(faction *domain.Faction, hookRegistry *hooks.Registry) {
	hookRegistry.RegisterRollModifier(
		hooks.FactionScope(faction.ID),
		"warlike",
		&WarlikeRollModifier{},
	)
}
