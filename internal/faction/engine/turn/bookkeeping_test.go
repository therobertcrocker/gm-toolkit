package turn

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestApplyBookkeeping_ClearsHookBudgets(t *testing.T) {
	te := newTurn()
	s := &state.FactionState{
		CampaignID: "test",
		Factions: map[string]*domain.Faction{
			"f1": {
				ID:      "f1",
				Force:   4,
				Cunning: 3,
				Wealth:  6,
				Coin:    10,
				HookBudgets: map[string]int{
					"tag:Warlike":   1,
					"tag:Fanatical": 2,
				},
			},
		},
	}
	if err := te.Start(s); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if _, _, err := te.ApplyBookkeeping(s, nil, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("ApplyBookkeeping: %v", err)
	}

	faction := s.Factions["f1"]
	if faction.HookBudgets != nil {
		t.Errorf("HookBudgets = %v after ApplyBookkeeping, want nil", faction.HookBudgets)
	}
}

type stubUpkeep struct{ upkeep hooks.TagUpkeep }

func (stub stubUpkeep) UpkeepDue(*domain.Faction) (hooks.TagUpkeep, bool) { return stub.upkeep, true }

// upkeepBookkeeping runs one bookkeeping for faction f1, which holds tag T-1 with
// upkeep params, and returns the mutations after the income line.
func upkeepBookkeeping(t *testing.T, engine *TurnEngine, faction *domain.Faction, params domain.CoinUpkeepParams) []domain.Mutation {
	t.Helper()
	s := &state.FactionState{CampaignID: "test", Factions: map[string]*domain.Faction{"f1": faction}}
	if err := engine.Start(s); err != nil {
		t.Fatalf("Start: %v", err)
	}
	registry := hooks.NewRegistry()
	registry.RegisterUpkeepCharger(hooks.FactionScope("f1"), "coin_upkeep:T-1", stubUpkeep{hooks.TagUpkeep{TagID: "T-1", Params: params}})
	_, mutations, err := engine.ApplyBookkeeping(s, registry, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("ApplyBookkeeping: %v", err)
	}
	return mutations[1:]
}

func TestApplyBookkeeping_Upkeep(t *testing.T) {
	lapse := domain.CoinUpkeepParams{Amount: 1, OnUnpaid: domain.UnpaidLapse}
	paid := domain.CoinDelta{FactionID: "f1", Delta: -1, Cause: "upkeep"}
	cases := []struct {
		name    string
		faction *domain.Faction
		params  domain.CoinUpkeepParams
		want    []domain.Mutation
	}{
		{"paid", &domain.Faction{ID: "f1", Coin: 2, Tags: []string{"T-1"}}, lapse, []domain.Mutation{paid}},
		{"paid from this turn's income", &domain.Faction{ID: "f1", Wealth: 2, Tags: []string{"T-1"}}, lapse, []domain.Mutation{paid}},
		{"paid, lifting a lapse", &domain.Faction{ID: "f1", Coin: 1, Tags: []string{"T-1"}, LapsedTags: []string{"T-1"}}, lapse, []domain.Mutation{
			paid, domain.TagLapsed{FactionID: "f1", TagID: "T-1", Lapsed: false, Cause: "upkeep", CausedByFactionID: "f1"},
		}},
		{"unpaid, lapses", &domain.Faction{ID: "f1", Tags: []string{"T-1"}}, lapse, []domain.Mutation{
			domain.TagLapsed{FactionID: "f1", TagID: "T-1", Lapsed: true, Cause: "upkeep", CausedByFactionID: "f1"},
		}},
		{"unpaid, already lapsed", &domain.Faction{ID: "f1", Tags: []string{"T-1"}, LapsedTags: []string{"T-1"}}, lapse, nil},
		{"unpaid, loses the tag", &domain.Faction{ID: "f1", Tags: []string{"T-1"}}, domain.CoinUpkeepParams{Amount: 1, OnUnpaid: domain.UnpaidLoseTag}, []domain.Mutation{
			domain.TagRemoved{FactionID: "f1", TagID: "T-1", Cause: "upkeep", CausedByFactionID: "f1"},
		}},
		{"unpaid, takes damage", &domain.Faction{ID: "f1", Tags: []string{"T-1"}}, domain.CoinUpkeepParams{Amount: 1, OnUnpaid: domain.UnpaidDamage, Damage: 2}, []domain.Mutation{
			domain.FactionHPDelta{FactionID: "f1", Delta: -2, Cause: "upkeep", CausedByFactionID: "f1"},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := upkeepBookkeeping(t, newTurn(), tc.faction, tc.params)
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestApplyBookkeeping_UpkeepIsPaidBeforeMaintenance(t *testing.T) {
	book := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"gear": {ID: "gear", Category: domain.StatForce, Maintenance: 1}}}
	faction := &domain.Faction{
		ID: "f1", Force: 1, Coin: 1, Tags: []string{"T-1"},
		Assets: map[string]*domain.Asset{"g-1": {ID: "g-1", DefinitionID: "gear", Maintained: true}},
	}
	got := upkeepBookkeeping(t, New(&randRoller{}, book, slog.New(slog.DiscardHandler)), faction, domain.CoinUpkeepParams{Amount: 1, OnUnpaid: domain.UnpaidLapse})
	want := []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: -1, Cause: "upkeep"},
		domain.AssetMaintainedFlag{FactionID: "f1", AssetID: "g-1", Maintained: false, Cause: "bookkeeping"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mutations = %+v, want the upkeep paid and the asset unmaintained", got)
	}
}
