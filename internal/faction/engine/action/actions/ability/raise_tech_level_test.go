package ability

import (
	"log/slog"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

var teachStep = domain.AbilityStep{
	Effect:         domain.EffectRaiseTechLevel,
	Target:         domain.TargetWorld,
	Roll:           &domain.RollTable{Die: domain.DiceRoll{NumDice: 1, Sides: 12}, Outcomes: []domain.AbilityOutcome{{From: 12, To: 12}}},
	RaiseTechLevel: &domain.RaiseTechLevelParams{CoinCost: 2, TechLevel: 4},
}

func TestRaiseTechLevel(t *testing.T) {
	cases := []struct {
		name string
		coin int
		roll int
		want []domain.Mutation
	}{
		{"a 12 pays and records", 2, 12, []domain.Mutation{
			domain.CoinDelta{FactionID: "f1", Delta: -2, Cause: "ability", CausedByFactionID: "f1"},
			domain.WorldTechLevelRaised{FactionID: "f1", WorldID: "near", TechLevel: 4, Cause: "ability", CausedByFactionID: "f1"},
		}},
		{"a miss pays only", 2, 11, []domain.Mutation{
			domain.CoinDelta{FactionID: "f1", Delta: -2, Cause: "ability", CausedByFactionID: "f1"},
		}},
		{"too poor to pay does nothing", 1, 12, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faction := &domain.Faction{ID: "f1", Coin: tc.coin}
			got, err := raiseTechLevel(StepContext{
				Faction:    faction,
				Step:       teachStep,
				Target:     domain.AbilityTarget{Kind: domain.TargetWorld, World: "near", Faction: faction},
				RollResult: tc.roll,
			})
			if err != nil {
				t.Fatalf("raiseTechLevel: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestRaiseTechLevelWorlds(t *testing.T) {
	engine := world.NewWithMap(&relocRouter{
		worlds:     map[string]spatial.RegionHex{"origin": relocHex("r1", 0), "near": relocHex("r1", 1), "mid": relocHex("r1", 2), "far": relocHex("r1", 5)},
		techLevels: map[string]int{"origin": 2, "near": 3, "mid": 4, "far": 1},
	}, slog.New(slog.DiscardHandler))
	base := func(world string) *domain.Base { return &domain.Base{Location: domain.Location{WorldID: world}} }
	faction := func(coin int) *domain.Faction {
		return &domain.Faction{
			ID: "f1", Coin: coin,
			Bases:           []*domain.Base{base("near"), base("origin"), base("mid"), base("far")},
			WorldTechLevels: map[string]int{"far": 4},
		}
	}

	if got, want := raiseTechLevelWorlds(teachStep, faction(2), engine), []string{"near", "origin"}; !slices.Equal(got, want) {
		t.Errorf("worlds = %v, want %v (mid is TL4 already, far is recorded)", got, want)
	}
	if got := raiseTechLevelWorlds(teachStep, faction(1), engine); len(got) != 0 {
		t.Errorf("worlds for a faction that cannot pay = %v, want none", got)
	}
	if got := raiseTechLevelWorlds(teachStep, faction(2), nil); len(got) != 0 {
		t.Errorf("worlds with no world engine = %v, want none", got)
	}

	tag := &domain.Tag{ID: "T-013", Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{teachStep}}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction(1)}}
	if HasTargets(factionState.Factions["f1"], domain.AbilityHolder{Tag: tag}, tag.Ability, factionState, engine) {
		t.Error("HasTargets = true for a faction that cannot pay, want false")
	}
}
