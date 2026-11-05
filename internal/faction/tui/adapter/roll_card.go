package adapter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

var rollPhaseWords = map[domain.RollPhase]string{
	domain.RollPhaseAttack:      "to attack",
	domain.RollPhaseDefense:     "to defend",
	domain.RollPhaseFactionTest: "to test",
	domain.RollPhaseContested:   "to resist",
}

// preRollPayload words the before card. Safe on the engine goroutine: it is parked
// on this ask, and the payload holds only strings.
func (a *Adapter) preRollPayload(card hooks.PreRollCard) PreRollCardPayload {
	offers := make([]string, 0, len(card.Offers))
	for _, offer := range card.Offers {
		offers = append(offers, fmt.Sprintf("%s — %s, %s", offer.Description, offer.Source, a.factionName(offer.FactionID)))
	}
	struck := make([]string, 0, len(card.Struck))
	for _, row := range card.Struck {
		struck = append(struck, struckLine(row))
	}
	return PreRollCardPayload{
		Headline: rollHeadline(card.Context),
		Detail:   a.rollDetail(card.Context),
		Dice:     fmt.Sprintf("%dd%d %+d%s", card.Base.NumDice, card.Base.Sides, card.Base.Modifier, versus(card.Context)),
		Offers:   offers,
		Struck:   struck,
	}
}

// postRollPayload words the after card.
func (a *Adapter) postRollPayload(card hooks.PostRollCard) PostRollCardPayload {
	rerolls := make([]string, 0, len(card.Rerolls))
	for _, directive := range card.Rerolls {
		rerolls = append(rerolls, fmt.Sprintf("Reroll the %s — %s, %s", diceAt(card.Result.Dice, directive.Indices), directive.Source, a.factionName(directive.FactionID)))
	}
	return PostRollCardPayload{
		Headline: rollHeadline(card.Context),
		Detail:   a.rollDetail(card.Context),
		Dice:     settledDice(card),
		Applied:  appliedNotes(card.Applied),
		Rerolls:  rerolls,
	}
}

// rollLine is a settled roll's feed line.
func rollLine(card hooks.PostRollCard) string {
	line := fmt.Sprintf("%s: %s", rollHeadline(card.Context), settledDice(card))
	if notes := appliedNotes(card.Applied); len(notes) > 0 {
		line += " (" + strings.Join(notes, "; ") + ")"
	}
	return line
}

func (a *Adapter) emitRollLine(card hooks.PostRollCard) {
	a.emit(EvtRollSettled, RollSettledPayload{Line: rollLine(card)})
}

func rollHeadline(ctx hooks.RollContext) string {
	return fmt.Sprintf("%s rolls %s %s", ctx.Actor.Name, ctx.Attribute, rollPhaseWords[ctx.Phase])
}

func (a *Adapter) rollDetail(ctx hooks.RollContext) string {
	var parts []string
	if ctx.Asset != nil {
		parts = append(parts, a.assetName(ctx.Asset))
	}
	if ctx.Action != "" {
		parts = append(parts, ctx.Action)
	}
	if ctx.World != "" {
		parts = append(parts, "on "+a.worldName(ctx.World))
	}
	if ctx.Opponent != nil {
		opponent := "against " + ctx.Opponent.Name
		if ctx.OpponentAsset != nil {
			opponent += "'s " + a.assetName(ctx.OpponentAsset)
		}
		parts = append(parts, opponent)
	}
	return strings.Join(parts, " · ")
}

func settledDice(card hooks.PostRollCard) string {
	return fmt.Sprintf("%v %+d = %d%s", card.Result.Dice, card.Result.Modifier, card.Result.Sum, versus(card.Context))
}

func versus(ctx hooks.RollContext) string {
	if ctx.OpponentTotal == nil {
		return ""
	}
	return fmt.Sprintf(" vs %d", *ctx.OpponentTotal)
}

func struckLine(row hooks.StruckOffer) string {
	if row.StruckBy == "" {
		return row.Offer.Source + ": used this turn"
	}
	return fmt.Sprintf("%s: struck by %s", row.Offer.Source, row.StruckBy)
}

func appliedNotes(applied []hooks.AppliedReroll) []string {
	notes := make([]string, 0, len(applied))
	for _, reroll := range applied {
		notes = append(notes, fmt.Sprintf("%s rerolled a %s", reroll.Source, joinDice(reroll.Was)))
	}
	return notes
}

func diceAt(dice, indices []int) string {
	values := make([]int, 0, len(indices))
	for _, index := range indices {
		if index >= 0 && index < len(dice) {
			values = append(values, dice[index])
		}
	}
	return joinDice(values)
}

func joinDice(values []int) string {
	words := make([]string, 0, len(values))
	for _, value := range values {
		words = append(words, strconv.Itoa(value))
	}
	return strings.Join(words, " and ")
}

func (a *Adapter) factionName(factionID string) string {
	if faction := a.factionState.Factions[factionID]; faction != nil {
		return faction.Name
	}
	return factionID
}

func (a *Adapter) assetName(asset *domain.Asset) string {
	if def := a.engine.Rulebook.Assets[asset.DefinitionID]; def != nil {
		return def.Name
	}
	return asset.DefinitionID
}

func (a *Adapter) worldName(worldID string) string {
	if location, ok := a.engine.World.Location(worldID); ok {
		return location.Name()
	}
	return worldID
}
