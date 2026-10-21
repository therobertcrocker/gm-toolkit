package turn

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestExpireConditions(t *testing.T) {
	endsAtF1 := domain.Condition{Kind: domain.ConditionNoAbility, SourceAssetID: "src", ExpiresAtTurnOf: "f1"}
	endsAtF3 := domain.Condition{Kind: domain.ConditionNoAbility, SourceAssetID: "src2", ExpiresAtTurnOf: "f3"}
	endsElsewhere := domain.Condition{Kind: domain.ConditionNoAbility, SourceAssetID: "src3"}
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"f1": {ID: "f1", Assets: map[string]*domain.Asset{
				"f1-a": {ID: "f1-a", Conditions: []domain.Condition{endsElsewhere}},
			}},
			"f2": {ID: "f2", Assets: map[string]*domain.Asset{
				"f2-a": {ID: "f2-a", Conditions: []domain.Condition{endsAtF1, endsAtF3}},
			}},
			"f3": {ID: "f3", Assets: map[string]*domain.Asset{
				"f3-a": {ID: "f3-a", Conditions: []domain.Condition{endsAtF1}},
			}},
		},
	}

	mutations := ExpireConditions(factionState, "f1")

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2: %+v", len(mutations), mutations)
	}
	wantHolders := []string{"f2", "f3"}
	for i, mutation := range mutations {
		cleared, ok := mutation.(domain.AssetConditionCleared)
		if !ok {
			t.Fatalf("mutations[%d] is %T, want AssetConditionCleared", i, mutation)
		}
		if cleared.FactionID != wantHolders[i] || cleared.Condition != endsAtF1 {
			t.Errorf("mutations[%d] = %+v, want a clear of %+v on %s", i, cleared, endsAtF1, wantHolders[i])
		}
		if cleared.Cause != "condition_expired" || cleared.CausedByFactionID != "f1" {
			t.Errorf("mutations[%d] cause = %q by %q, want condition_expired by f1", i, cleared.Cause, cleared.CausedByFactionID)
		}
	}
}
