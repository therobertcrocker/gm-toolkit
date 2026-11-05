package adapter

import (
	"errors"
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
)

func newRollCardAdapter(t *testing.T) (*Adapter, *testharness.Harness) {
	t.Helper()
	harness := testharness.NewHarness(t)
	return New(harness.Engine, harness.FactionState, harness.Paths, slog.New(slog.DiscardHandler)), harness
}

func rollContext(harness *testharness.Harness) hooks.RollContext {
	return hooks.RollContext{
		Phase:     domain.RollPhaseAttack,
		Actor:     harness.AddFaction("alpha", "Andama", 3, 2, 1),
		Attribute: domain.StatForce,
	}
}

func TestPreRollPayload_HeadlineAndDetail(t *testing.T) {
	adapter, harness := newRollCardAdapter(t)
	alpha := harness.AddFaction("alpha", "Andama", 3, 2, 1)
	beta := harness.AddFaction("beta", "Krylos", 1, 2, 3)
	securityName := harness.Engine.Rulebook.Assets[testharness.DefSecurityPersonnel].Name

	payload := adapter.preRollPayload(hooks.PreRollCard{
		Context: hooks.RollContext{
			Phase:         domain.RollPhaseDefense,
			Actor:         alpha,
			Opponent:      beta,
			Attribute:     domain.StatCunning,
			Asset:         alpha.Assets["alpha-asset-1"],
			OpponentAsset: beta.Assets["beta-asset-1"],
			World:         "Andama",
			Action:        "Attack",
		},
		Base: domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: 2},
	})

	if want := "alpha rolls Cunning to defend"; payload.Headline != want {
		t.Errorf("Headline = %q, want %q", payload.Headline, want)
	}
	want := securityName + " · Attack · on Andama · against beta's " + securityName
	if payload.Detail != want {
		t.Errorf("Detail = %q, want %q", payload.Detail, want)
	}
}

func TestPayloads_Versus(t *testing.T) {
	adapter, harness := newRollCardAdapter(t)
	ctx := rollContext(harness)
	base := domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: 3}
	result := hooks.RollResult{Dice: []int{7}, Modifier: 3, Sum: 10}

	if got := adapter.preRollPayload(hooks.PreRollCard{Context: ctx, Base: base}).Dice; got != "1d10 +3" {
		t.Errorf("before card with no total: Dice = %q", got)
	}
	if got := adapter.postRollPayload(hooks.PostRollCard{Context: ctx, Result: result}).Dice; got != "[7] +3 = 10" {
		t.Errorf("after card with no total: Dice = %q", got)
	}

	total := 9
	ctx.OpponentTotal = &total
	if got := adapter.preRollPayload(hooks.PreRollCard{Context: ctx, Base: base}).Dice; got != "1d10 +3 vs 9" {
		t.Errorf("before card with a total: Dice = %q", got)
	}
	if got := adapter.postRollPayload(hooks.PostRollCard{Context: ctx, Result: result}).Dice; got != "[7] +3 = 10 vs 9" {
		t.Errorf("after card with a total: Dice = %q", got)
	}
}

func TestPreRollPayload_OfferRows(t *testing.T) {
	adapter, harness := newRollCardAdapter(t)
	ctx := rollContext(harness)
	harness.AddFaction("gamma-id", "Krylos", 1, 1, 1).Name = "Gamma Cartel"

	payload := adapter.preRollPayload(hooks.PreRollCard{
		Context: ctx,
		Offers:  []hooks.ModifierOffer{{Description: "Add 1d10", Source: "Warlike", FactionID: "gamma-id"}},
	})

	want := []string{"Add 1d10 — Warlike, Gamma Cartel"}
	if !slices.Equal(payload.Offers, want) {
		t.Errorf("Offers = %q, want %q", payload.Offers, want)
	}
}

func TestPreRollPayload_StruckRows(t *testing.T) {
	adapter, harness := newRollCardAdapter(t)

	payload := adapter.preRollPayload(hooks.PreRollCard{
		Context: rollContext(harness),
		Struck: []hooks.StruckOffer{
			{Offer: hooks.ModifierOffer{Source: "Warlike"}, StruckBy: "Blackmail"},
			{Offer: hooks.ModifierOffer{Source: "Warlike"}},
		},
	})

	want := []string{"Warlike: struck by Blackmail", "Warlike: used this turn"}
	if !slices.Equal(payload.Struck, want) {
		t.Errorf("Struck = %q, want %q", payload.Struck, want)
	}
}

func TestPostRollPayload_RerollRows(t *testing.T) {
	adapter, harness := newRollCardAdapter(t)
	ctx := rollContext(harness)
	harness.AddFaction("gamma-id", "Krylos", 1, 1, 1).Name = "gamma"

	payload := adapter.postRollPayload(hooks.PostRollCard{
		Context: ctx,
		Result:  hooks.RollResult{Dice: []int{7}, Modifier: 3, Sum: 10},
		Rerolls: []hooks.RerollDirective{{Source: "Psychic Academy (forced)", Indices: []int{0}, FactionID: "gamma-id", Elective: true}},
	})

	want := []string{"Reroll the 7 — Psychic Academy (forced), gamma"}
	if !slices.Equal(payload.Rerolls, want) {
		t.Errorf("Rerolls = %q, want %q", payload.Rerolls, want)
	}
}

func TestAppliedNote_OnTheCardAndInTheFeedLine(t *testing.T) {
	adapter, harness := newRollCardAdapter(t)
	card := hooks.PostRollCard{
		Context: rollContext(harness),
		Result:  hooks.RollResult{Dice: []int{7}, Modifier: 3, Sum: 10},
		Applied: []hooks.AppliedReroll{{Source: "Fanatical", Was: []int{1}}},
	}

	if want := []string{"Fanatical rerolled a 1"}; !slices.Equal(adapter.postRollPayload(card).Applied, want) {
		t.Errorf("Applied = %q, want %q", adapter.postRollPayload(card).Applied, want)
	}
	if want := "alpha rolls Force to attack: [7] +3 = 10 (Fanatical rerolled a 1)"; rollLine(card) != want {
		t.Errorf("rollLine = %q, want %q", rollLine(card), want)
	}
}

func TestPayloads_UnknownNamesFallBackToIDs(t *testing.T) {
	adapter, harness := newRollCardAdapter(t)
	ctx := rollContext(harness)
	ctx.Asset = &domain.Asset{ID: "ghost", DefinitionID: "NO-SUCH-DEF"}

	payload := adapter.preRollPayload(hooks.PreRollCard{
		Context: ctx,
		Offers:  []hooks.ModifierOffer{{Description: "Add 1d10", Source: "Warlike", FactionID: "no-such-faction"}},
	})

	if payload.Detail != "NO-SUCH-DEF" {
		t.Errorf("Detail = %q, want the definition ID", payload.Detail)
	}
	if !strings.HasSuffix(payload.Offers[0], ", no-such-faction") {
		t.Errorf("Offers[0] = %q, want the faction ID", payload.Offers[0])
	}
}

func TestPayloads_HoldOnlyPlainValues(t *testing.T) {
	for _, payload := range []any{PreRollCardPayload{}, PostRollCardPayload{}} {
		payloadType := reflect.TypeOf(payload)
		for index := range payloadType.NumField() {
			field := payloadType.Field(index)
			isString := field.Type.Kind() == reflect.String
			isStringSlice := field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String
			if !isString && !isStringSlice {
				t.Errorf("%s.%s is %s, want a string or a string slice (D-12)", payloadType.Name(), field.Name, field.Type)
			}
		}
	}
}

func TestSelectReroll_NoRerollsLeavesOneFeedLineAndDoesNotAsk(t *testing.T) {
	adapter, harness := newRollCardAdapter(t)
	card := hooks.PostRollCard{Context: rollContext(harness), Result: hooks.RollResult{Dice: []int{7}, Modifier: 3, Sum: 10}}

	done := make(chan *hooks.RerollDirective, 1)
	go func() { done <- adapter.Action().SelectReroll(card) }()

	select {
	case msg := <-adapter.askCh:
		t.Fatalf("asked with no rerolls on offer: kind %d", msg.Kind)
	case choice := <-done:
		if choice != nil {
			t.Errorf("SelectReroll = %+v, want nil", choice)
		}
	}
	lines := drainRollLines(adapter)
	if want := []string{"alpha rolls Force to attack: [7] +3 = 10"}; !slices.Equal(lines, want) {
		t.Errorf("feed lines = %q, want %q", lines, want)
	}
}

func TestSelectModifiers_ReturnsOnlyTheTakenOffers(t *testing.T) {
	adapter, harness := newRollCardAdapter(t)
	offers := []hooks.ModifierOffer{{Source: "first"}, {Source: "second"}}
	answerNextAsk(t, adapter, AskPreRollCard, []int{1})

	chosen := adapter.Action().SelectModifiers(hooks.PreRollCard{Context: rollContext(harness), Offers: offers})

	if len(chosen) != 1 || chosen[0].Source != "second" {
		t.Errorf("SelectModifiers = %+v, want only the second offer", chosen)
	}
}

func TestSelectReroll_TakenLeavesNoFeedLine(t *testing.T) {
	adapter, harness := newRollCardAdapter(t)
	card := rerollCard(harness)
	answerNextAsk(t, adapter, AskPostRollCard, 0)

	choice := adapter.Action().SelectReroll(card)

	if choice == nil || choice.Source != "first" {
		t.Errorf("SelectReroll = %+v, want the first reroll", choice)
	}
	if lines := drainRollLines(adapter); len(lines) != 0 {
		t.Errorf("feed lines = %q, want none", lines)
	}
}

func TestSelectReroll_KeepLeavesOneFeedLine(t *testing.T) {
	adapter, harness := newRollCardAdapter(t)
	card := rerollCard(harness)
	answerNextAsk(t, adapter, AskPostRollCard, KeepRoll)

	if choice := adapter.Action().SelectReroll(card); choice != nil {
		t.Errorf("SelectReroll = %+v, want nil", choice)
	}
	if lines := drainRollLines(adapter); len(lines) != 1 {
		t.Errorf("feed lines = %q, want one", lines)
	}
}

func TestSelectModifiers_BadReplyDeclinesAndReports(t *testing.T) {
	replies := map[string]any{
		"wrong type":   "not a slice",
		"out of range": []int{5},
		"negative":     []int{-1},
		"an error":     errors.New("canceled"),
	}
	for name, reply := range replies {
		t.Run(name, func(t *testing.T) {
			adapter, harness := newRollCardAdapter(t)
			offers := []hooks.ModifierOffer{{Source: "first"}}
			answerNextAsk(t, adapter, AskPreRollCard, reply)

			if chosen := adapter.Action().SelectModifiers(hooks.PreRollCard{Context: rollContext(harness), Offers: offers}); chosen != nil {
				t.Errorf("SelectModifiers = %+v, want nil", chosen)
			}
			if errorCount, _ := drainEvents(adapter); errorCount != 1 {
				t.Errorf("EvtError count = %d, want 1", errorCount)
			}
		})
	}
}

func TestSelectReroll_BadReplyDeclinesReportsAndLeavesTheFeedLine(t *testing.T) {
	replies := map[string]any{
		"wrong type":   "not an int",
		"out of range": 5,
		"below keep":   KeepRoll - 1,
		"an error":     errors.New("canceled"),
	}
	for name, reply := range replies {
		t.Run(name, func(t *testing.T) {
			adapter, harness := newRollCardAdapter(t)
			card := rerollCard(harness)
			answerNextAsk(t, adapter, AskPostRollCard, reply)

			if choice := adapter.Action().SelectReroll(card); choice != nil {
				t.Errorf("SelectReroll = %+v, want nil", choice)
			}
			errorCount, lines := drainEvents(adapter)
			if errorCount != 1 {
				t.Errorf("EvtError count = %d, want 1", errorCount)
			}
			if len(lines) != 1 {
				t.Errorf("feed lines = %q, want one", lines)
			}
		})
	}
}

func rerollCard(harness *testharness.Harness) hooks.PostRollCard {
	return hooks.PostRollCard{
		Context: rollContext(harness),
		Result:  hooks.RollResult{Dice: []int{7}, Modifier: 3, Sum: 10},
		Rerolls: []hooks.RerollDirective{{Source: "first", Indices: []int{0}, Elective: true}},
	}
}

// answerNextAsk answers the next ask on AskPump, in the way the execution view does,
// after checking its kind.
func answerNextAsk(t *testing.T, adapter *Adapter, wantKind AskKind, answer any) {
	t.Helper()
	go func() {
		msg, ok := adapter.AskPump()().(CollectorAskMsg)
		if !ok {
			t.Error("AskPump did not return a CollectorAskMsg")
			return
		}
		if msg.Kind != wantKind {
			t.Errorf("ask kind = %d, want %d", msg.Kind, wantKind)
		}
		msg.Reply <- answer
	}()
}

// drainRollLines reads what is buffered on the event channel and returns the lines
// of its EvtRollSettled events.
func drainRollLines(adapter *Adapter) []string {
	_, lines := drainEvents(adapter)
	return lines
}

// drainEvents reads what is buffered on the event channel and returns how many
// EvtError events it held and the lines of its EvtRollSettled events.
func drainEvents(adapter *Adapter) (int, []string) {
	errorCount := 0
	var lines []string
	for {
		select {
		case msg := <-adapter.eventCh:
			switch msg.Kind {
			case EvtError:
				errorCount++
			case EvtRollSettled:
				lines = append(lines, msg.Payload.(RollSettledPayload).Line)
			}
		default:
			return errorCount, lines
		}
	}
}
