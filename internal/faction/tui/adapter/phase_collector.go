package adapter

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
)

type phaseCollector struct{ adapter *Adapter }

func (a *Adapter) Phase() engine.PhaseCollector { return &phaseCollector{adapter: a} }

// ask sends a CollectorAskMsg to the UI thread and blocks until the reply arrives.
func (p *phaseCollector) ask(kind AskKind, faction *domain.Faction, payload any) (any, error) {
	reply := make(chan any, 1)
	p.adapter.askCh <- CollectorAskMsg{Kind: kind, Faction: faction, Payload: payload, Reply: reply}
	received := <-reply
	if err, ok := received.(error); ok {
		return nil, err
	}
	return received, nil
}

func (p *phaseCollector) AwaitCheckpoint(phase string) error {
	if !p.adapter.perFactionCadence.Load() && phase != engine.CheckpointCycleSummary {
		return nil
	}
	_, err := p.ask(AskAwaitCheckpoint, nil, AwaitCheckpointPayload{Phase: phase})
	return err
}

func (p *phaseCollector) SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error) {
	raw, err := p.ask(AskSelectAction, faction, SelectActionPayload{Available: available})
	if err != nil {
		return nil, err
	}
	chosen, ok := raw.(action.Action)
	if !ok {
		return nil, fmt.Errorf("phase_collector.SelectAction: unexpected reply type %T", raw)
	}
	return chosen, nil
}

func (p *phaseCollector) SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error) {
	raw, err := p.ask(AskSelectStatRaise, faction, SelectStatRaisePayload{Eligible: eligible})
	if err != nil {
		return nil, err
	}
	chosen, ok := raw.(*domain.FactionStat)
	if !ok {
		return nil, fmt.Errorf("phase_collector.SelectStatRaise: unexpected reply type %T", raw)
	}
	return chosen, nil
}

func (p *phaseCollector) SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
	raw, err := p.ask(AskSelectMovementDecisions, faction, SelectMovementDecisionsPayload{Eligible: eligible})
	if err != nil {
		return nil, err
	}
	chosen, ok := raw.([]world.MovementDecision)
	if !ok {
		return nil, fmt.Errorf("phase_collector.SelectMovementDecisions: unexpected reply type %T", raw)
	}
	return chosen, nil
}

func (p *phaseCollector) SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error) {
	raw, err := p.ask(AskSelectTransportCargo, nil, SelectTransportCargoPayload{Transport: transport, EligibleCargo: eligibleCargo, Profile: profile})
	if err != nil {
		return nil, err
	}
	chosen, ok := raw.([]*domain.Asset)
	if !ok {
		return nil, fmt.Errorf("phase_collector.SelectTransportCargo: unexpected reply type %T", raw)
	}
	return chosen, nil
}

var _ engine.PhaseCollector = (*phaseCollector)(nil)
