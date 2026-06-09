package overlay

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
)

// defaultFor's reply for each kind must be the exact dynamic type the phase
// collector asserts on (phase_collector.go), or the cycle errors at the
// assertion. AwaitCheckpoint's reply is discarded (any non-error value is
// legal); SelectAction's skip is an untyped nil.
func TestDefaultForMatchesCollectorAssertions(t *testing.T) {
	if _, ok := defaultFor(adapter.AskAwaitCheckpoint).(bool); !ok {
		t.Errorf("AskAwaitCheckpoint: want bool ack, got %T", defaultFor(adapter.AskAwaitCheckpoint))
	}
	if v := defaultFor(adapter.AskSelectAction); v != nil {
		t.Errorf("AskSelectAction: want untyped nil skip, got %T", v)
	}
	if _, ok := defaultFor(adapter.AskSelectStatRaise).(*domain.FactionStat); !ok {
		t.Errorf("AskSelectStatRaise: want *domain.FactionStat, got %T", defaultFor(adapter.AskSelectStatRaise))
	}
	if _, ok := defaultFor(adapter.AskSelectMovementDecisions).([]world.MovementDecision); !ok {
		t.Errorf("AskSelectMovementDecisions: want []world.MovementDecision, got %T", defaultFor(adapter.AskSelectMovementDecisions))
	}
	if _, ok := defaultFor(adapter.AskSelectTransportCargo).([]*domain.Asset); !ok {
		t.Errorf("AskSelectTransportCargo: want []*domain.Asset, got %T", defaultFor(adapter.AskSelectTransportCargo))
	}
}
