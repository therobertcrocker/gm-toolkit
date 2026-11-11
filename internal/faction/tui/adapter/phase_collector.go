package adapter

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
)

type phaseCollector struct{ adapter *Adapter }

func (a *Adapter) Phase() engine.PhaseCollector { return &phaseCollector{adapter: a} }

// ask sends a CollectorAskMsg to the UI thread and blocks until the reply
// arrives. Hoisted onto *Adapter so both the phase and action collectors share
// the one unbuffered askCh (Discovery Decision 4).
func (a *Adapter) ask(kind AskKind, faction *domain.Faction, payload any) (any, error) {
	reply := make(chan any, 1)
	a.askCh <- CollectorAskMsg{Kind: kind, Faction: faction, Payload: payload, Reply: reply}
	received := <-reply
	if err, ok := received.(error); ok {
		return nil, err
	}
	return received, nil
}

// replyAs reads an ask's reply as T. An untyped nil is a legal empty answer (a skip, or
// a decline), and any other type is a code bug that must not pass as a zero value.
func replyAs[T any](method string, raw any) (T, error) {
	var empty T
	if raw == nil {
		return empty, nil
	}
	reply, ok := raw.(T)
	if !ok {
		return empty, fmt.Errorf("%s: unexpected reply type %T", method, raw)
	}
	return reply, nil
}

func (p *phaseCollector) AwaitCheckpoint(phase string) error {
	if !p.adapter.perFactionCadence.Load() && phase != engine.CheckpointCycleSummary {
		return nil
	}
	_, err := p.adapter.ask(AskAwaitCheckpoint, nil, AwaitCheckpointPayload{Phase: phase})
	return err
}

func (p *phaseCollector) SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error) {
	raw, err := p.adapter.ask(AskSelectAction, faction, SelectActionPayload{Available: available})
	if err != nil {
		return nil, err
	}
	return replyAs[action.Action]("SelectAction", raw)
}

func (p *phaseCollector) SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error) {
	raw, err := p.adapter.ask(AskSelectStatRaise, faction, SelectStatRaisePayload{Eligible: eligible})
	if err != nil {
		return nil, err
	}
	return replyAs[*domain.FactionStat]("SelectStatRaise", raw)
}

func (p *phaseCollector) SelectGoal(faction *domain.Faction, offers []domain.GoalOffer) (*domain.GoalChoice, error) {
	raw, err := p.adapter.ask(AskSelectGoal, faction, SelectGoalPayload{Offers: offers})
	if err != nil {
		return nil, err
	}
	return replyAs[*domain.GoalChoice]("SelectGoal", raw)
}

func (p *phaseCollector) SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
	raw, err := p.adapter.ask(AskSelectMovementDecisions, faction, SelectMovementDecisionsPayload{Eligible: eligible})
	if err != nil {
		return nil, err
	}
	return replyAs[[]world.MovementDecision]("SelectMovementDecisions", raw)
}

func (p *phaseCollector) SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error) {
	raw, err := p.adapter.ask(AskSelectTransportCargo, nil, SelectTransportCargoPayload{Transport: transport, EligibleCargo: eligibleCargo, Profile: profile})
	if err != nil {
		return nil, err
	}
	return replyAs[[]*domain.Asset]("SelectTransportCargo", raw)
}

func (p *phaseCollector) SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error) {
	return p.adapter.selectRelocateOrders(acting, candidates, candidateWorlds, maxAssets)
}

func (p *phaseCollector) ConfirmPayment(offer hooks.PaymentOffer) (bool, error) {
	return p.adapter.confirmPayment(offer)
}

func (p *phaseCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	return p.adapter.requestPermission(request)
}

func (p *phaseCollector) ConfirmPermissionContest(offer hooks.PermissionContestOffer) (bool, error) {
	return p.adapter.confirmPermissionContest(offer)
}

var _ engine.PhaseCollector = (*phaseCollector)(nil)
