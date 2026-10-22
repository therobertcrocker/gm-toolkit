package goals

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestInsideEnemyTerritory_CountsRivalGround(t *testing.T) {
	assetWorlds := map[string]string{
		"governed":         "Tartarus",
		"rival-governed":   "Krylos",
		"independent":      "Hadrian",
		"rival-homeworld":  "Anchorage",
		"seized-homeworld": "Vesta",
		"lost-homeworld":   "Brennus",
		"in-transit":       "",
	}
	faction := &domain.Faction{
		ID:         "f1",
		Cunning:    5,
		Homeworld:  domain.Location{WorldID: "Brennus"},
		ActiveGoal: &domain.ActiveGoal{GoalID: "G-009"},
		Assets:     map[string]*domain.Asset{},
	}
	var mutations []domain.Mutation
	for assetID, worldID := range assetWorlds {
		faction.Assets[assetID] = &domain.Asset{ID: assetID, Location: domain.Location{WorldID: worldID}}
		mutations = append(mutations, domain.AssetStealthApplied{FactionID: "f1", AssetID: assetID})
	}
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"f1": faction,
			"f2": {ID: "f2", Homeworld: domain.Location{WorldID: "Anchorage"}},
			"f3": {ID: "f3", Homeworld: domain.Location{WorldID: "Vesta"}},
		},
		Governments: map[string]string{"Tartarus": "f1", "Krylos": "f2", "Brennus": "f2", "Vesta": "f1"},
	}

	got := InsideEnemyTerritory{}.UpdateProgress(faction, mutations, factionState, nil, nil)

	want := []domain.Mutation{domain.GoalProgressed{FactionID: "f1", GoalID: "G-009", Delta: 3, Cause: "inside_enemy_territory"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mutations = %+v, want %+v (Krylos, Anchorage and Brennus count)", got, want)
	}
}
