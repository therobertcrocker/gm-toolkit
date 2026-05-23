package effect_test

import (
	"log/slog"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// TestEffectsEngine_RegistersTransportReactor verifies that ApplyAll registers
// exactly one MutationReactor at FactionScope for a faction whose assets include
// a transport-capable asset, and zero for a faction with no transport assets.
func TestEffectsEngine_RegistersTransportReactor(t *testing.T) {
	transportDefID := "test-transport"
	nonTransportDefID := "test-soldier"

	transportDef := &domain.AssetDefinition{
		ID:    transportDefID,
		Speed: 2,
		Transport: &domain.TransportProfile{
			MaxHex:     3,
			CoinCost:   1,
			CargoTypes: []domain.AssetType{domain.TypeSpecialForces},
			MaxCargo:   1,
		},
	}
	nonTransportDef := &domain.AssetDefinition{
		ID:   nonTransportDefID,
		Type: domain.TypeSpecialForces,
	}

	rb := &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			transportDefID:    transportDef,
			nonTransportDefID: nonTransportDef,
		},
	}

	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"alpha": {
				ID: "alpha",
				Assets: map[string]*domain.Asset{
					"transport-1": {ID: "transport-1", DefinitionID: transportDefID, OwnerID: "alpha"},
				},
			},
			"beta": {
				ID: "beta",
				Assets: map[string]*domain.Asset{
					"soldier-1": {ID: "soldier-1", DefinitionID: nonTransportDefID, OwnerID: "beta"},
				},
			},
		},
	}

	e := effect.New(slog.New(slog.DiscardHandler))
	e.Register(effects.NewTransportHandler(transportDef))

	registry := hooks.NewRegistry()
	e.ApplyAll(factionState, rb, registry, slog.New(slog.DiscardHandler))

	wantSource := "transport:" + transportDefID

	// Faction alpha has a transport asset — one reactor should be registered.
	alphaReactors := registry.MutationReactorsFor("alpha", "")
	if len(alphaReactors) != 1 {
		t.Fatalf("alpha: got %d MutationReactors, want 1", len(alphaReactors))
	}
	if got := alphaReactors[0].Source; got != wantSource {
		t.Errorf("alpha reactor source = %q, want %q", got, wantSource)
	}

	// Faction beta has no transport asset — no reactor should be registered.
	betaReactors := registry.MutationReactorsFor("beta", "")
	if len(betaReactors) != 0 {
		t.Errorf("beta: got %d MutationReactors, want 0", len(betaReactors))
	}
}
