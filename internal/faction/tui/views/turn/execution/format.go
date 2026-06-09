package execution

import (
	"fmt"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// renderEvent turns one observer event into a rendered stream block: a header
// line followed by indented, player-meaningful mutation lines. Faction-named
// events render their name; phase events (whose payloads carry no faction)
// render a phase phrase only.
func renderEvent(msg adapter.ObserverEventMsg) string {
	header, muts := eventHeader(msg)
	if len(muts) == 0 {
		return header
	}
	var b strings.Builder
	b.WriteString(header)
	for _, mut := range muts {
		if line, include := formatMutation(mut); include {
			b.WriteString("\n" + styles.Dim.Render("    "+line))
		}
	}
	return b.String()
}

// eventHeader returns the styled header line and the mutations to indent under
// it (nil for events that carry none).
func eventHeader(msg adapter.ObserverEventMsg) (string, []domain.Mutation) {
	switch msg.Kind {
	case adapter.EvtCycleStarted:
		p := msg.Payload.(adapter.CycleStartedPayload)
		return styles.Strong.Render(fmt.Sprintf("Cycle %d started", p.CycleNumber)), nil
	case adapter.EvtFactionTurnStarted:
		return factionHeader(msg.Payload.(*domain.Faction).Name, "turn started"), nil
	case adapter.EvtFactionSkipped:
		return factionHeader(msg.Payload.(*domain.Faction).Name, "skipped"), nil
	case adapter.EvtFactionTurnCompleted:
		return factionHeader(msg.Payload.(*domain.Faction).Name, "turn completed"), nil
	case adapter.EvtStatRaiseSkipped:
		return factionHeader(msg.Payload.(*domain.Faction).Name, "stat raise skipped"), nil
	case adapter.EvtGoalLockApplied:
		return phaseHeader("goal-lock applied"), msg.Payload.(adapter.GoalLockAppliedPayload).Mutations
	case adapter.EvtBookkeepingApplied:
		return phaseHeader("bookkeeping applied"), msg.Payload.(adapter.BookkeepingAppliedPayload).Mutations
	case adapter.EvtStatRaiseApplied:
		return phaseHeader("stat raise applied"), msg.Payload.(adapter.StatRaiseAppliedPayload).Mutations
	case adapter.EvtMovementTicked:
		return phaseHeader("movement ticked"), msg.Payload.(adapter.MovementTickedPayload).Mutations
	case adapter.EvtMovementResolved:
		return phaseHeader("movement resolved"), msg.Payload.(adapter.MovementResolvedPayload).Mutations
	case adapter.EvtActionSelected:
		p := msg.Payload.(adapter.ActionSelectedPayload)
		name := "(none)"
		if p.Selected != nil {
			name = p.Selected.Name()
		}
		return phaseHeader("action selected: " + name), nil
	case adapter.EvtActionResolved:
		p := msg.Payload.(adapter.ActionResolvedPayload)
		name := "(none)"
		if p.Selected != nil {
			name = p.Selected.Name()
		}
		return phaseHeader("action resolved: " + name), p.Mutations
	case adapter.EvtCycleCompleted:
		p := msg.Payload.(adapter.CycleCompletedPayload)
		return styles.Strong.Render(fmt.Sprintf("Cycle %d completed", p.CycleNumber)), nil
	case adapter.EvtIndexSkipped:
		p := msg.Payload.(adapter.IndexSkippedPayload)
		return phaseHeader("index skipped: " + strings.Join(p.Skipped, ", ")), nil
	case adapter.EvtError:
		return styles.Danger.Render("error: " + msg.Payload.(adapter.ErrorPayload).Err.Error()), nil
	}
	return "", nil
}

func factionHeader(name, phrase string) string {
	return styles.Strong.Render(name) + styles.Subtle.Render("  "+phrase)
}

func phaseHeader(phrase string) string {
	return styles.Subtle.Render(phrase)
}

// formatMutation renders one mutation as a single indented line, returning
// include=false for bookkeeping noise the stream omits. AssetAdded surfaces the
// asset's DefinitionID — domain.Asset carries no display name (that lives on the
// rulebook's AssetDefinition, and this formatter has no rulebook).
func formatMutation(m domain.Mutation) (string, bool) {
	switch mut := m.(type) {
	case domain.FactionHPDelta:
		return fmt.Sprintf("HP %+d", mut.Delta), true
	case domain.AssetHPDelta:
		return fmt.Sprintf("asset %s HP %+d", mut.AssetID, mut.Delta), true
	case domain.BaseHPDelta:
		return fmt.Sprintf("base %s HP %+d", mut.BaseID, mut.Delta), true
	case domain.BaseHealed:
		return fmt.Sprintf("base %s healed %+d", mut.BaseID, mut.Delta), true
	case domain.BaseExpanded:
		return fmt.Sprintf("base %s expanded %+d", mut.BaseID, mut.Delta), true
	case domain.CoinDelta:
		return fmt.Sprintf("Coin %+d", mut.Delta), true
	case domain.AssetAdded:
		return fmt.Sprintf("+asset %s", mut.Asset.DefinitionID), true
	case domain.AssetRemoved:
		return fmt.Sprintf("-asset %s", mut.AssetID), true
	case domain.BaseAdded:
		return fmt.Sprintf("+base %s", mut.Base.ID), true
	case domain.BaseDestroyed:
		return fmt.Sprintf("base %s destroyed", mut.BaseID), true
	case domain.StatRaised:
		return fmt.Sprintf("%s %d→%d", mut.Stat, mut.OldRating, mut.NewRating), true
	case domain.GoalCompleted:
		return fmt.Sprintf("goal completed (+%d XP)", mut.XPAwarded), true
	case domain.GoalAbandoned:
		return "goal abandoned", true
	case domain.HomeworldChanged:
		return fmt.Sprintf("homeworld → %s", mut.ToWorld.WorldID), true
	case domain.MovementOrderIssued:
		return "move issued", true
	case domain.MovementOrderRevised:
		return "move revised", true
	case domain.MovementOrderCancelled:
		return "move cancelled", true
	case domain.MovementOrderCompleted:
		return "move completed", true
	default:
		// Folded as bookkeeping noise or redundant with a surfaced line:
		// GoalTurnsTick, GoalProgressed, GoalInitiated, GoalPhaseAdvanced,
		// AssetMaintainedFlag, AssetStealthApplied, AssetStealthCleared,
		// AssetMoved, XPSpent, XPAwarded, MovementOrderProgressed,
		// InfluenceDelta, TagAdded.
		return "", false
	}
}
