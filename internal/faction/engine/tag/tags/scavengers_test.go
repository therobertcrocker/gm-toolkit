package tags_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/tag/tags"
)

func TestScavengersReactor_EmitsCoinDeltaOnKill(t *testing.T) {
	reactor := &tags.ScavengersReactor{FactionID: "f1"}
	mutations := []domain.Mutation{
		domain.AssetRemoved{FactionID: "f2", AssetID: "a1", Cause: "attack", CausedByFactionID: "f1"},
	}

	result := reactor.OnMutations(mutations, nil, nil)

	if len(result) != 1 {
		t.Fatalf("got %d mutations, want 1", len(result))
	}
	coin, ok := result[0].(domain.CoinDelta)
	if !ok {
		t.Fatalf("result[0] type = %T, want CoinDelta", result[0])
	}
	if coin.FactionID != "f1" || coin.Delta != 1 {
		t.Errorf("CoinDelta = {%s, %d}, want {f1, 1}", coin.FactionID, coin.Delta)
	}
}

// Scavengers fires even when the destroyed asset belongs to the owning faction.
func TestScavengersReactor_OwnAssetDestroyed(t *testing.T) {
	reactor := &tags.ScavengersReactor{FactionID: "f1"}
	mutations := []domain.Mutation{
		domain.AssetRemoved{FactionID: "f1", AssetID: "a1", Cause: "attack", CausedByFactionID: "f2"},
	}

	result := reactor.OnMutations(mutations, nil, nil)
	if len(result) != 1 {
		t.Fatalf("got %d mutations, want 1 (own-asset kill should still trigger Scavengers)", len(result))
	}
}

func TestScavengersReactor_IgnoresNonCombatRemoval(t *testing.T) {
	reactor := &tags.ScavengersReactor{FactionID: "f1"}
	mutations := []domain.Mutation{
		domain.AssetRemoved{FactionID: "f1", AssetID: "a1", Cause: "sell", CausedByFactionID: "f1"},
	}

	result := reactor.OnMutations(mutations, nil, nil)
	if len(result) != 0 {
		t.Errorf("got %d mutations, want 0 (sell removal should not trigger Scavengers)", len(result))
	}
}
