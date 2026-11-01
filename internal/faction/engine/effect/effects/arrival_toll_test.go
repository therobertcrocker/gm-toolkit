package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

func TestArrivalToll_TagHolderChargesRivalsOnItsBaseWorlds(t *testing.T) {
	pirates := &domain.Tag{ID: "SWN-T-010", Name: "Pirates"}
	bravo := &domain.Faction{ID: "bravo", Tags: []string{pirates.ID},
		Bases: []*domain.Base{{ID: "bravo-krylos", OwnerID: "bravo", Location: domain.Location{WorldID: "Krylos"}}}}
	alpha := &domain.Faction{ID: "alpha"}
	factions := map[string]*domain.Faction{"alpha": alpha, "bravo": bravo}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialArrivalToll].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: bravo, Tag: pirates},
		Entry: domain.EffectDefinition{
			Trigger:     domain.TriggerArrivalQuery,
			Effect:      domain.SpecialArrivalToll,
			Filters:     domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeBase},
			ArrivalToll: &domain.ArrivalTollParams{Amount: 1},
		},
		Registry: registry,
	})

	if got, want := dispatch.ArrivalTolls(registry, factions, alpha, "Krylos"), []hooks.ArrivalToll{{FactionID: "bravo", Amount: 1}}; !reflect.DeepEqual(got, want) {
		t.Errorf("tolls on the Base world = %+v, want %+v", got, want)
	}
	if got := dispatch.ArrivalTolls(registry, factions, alpha, "Hadrian"); len(got) != 0 {
		t.Errorf("tolls with no Base = %+v, want none", got)
	}
	if got := dispatch.ArrivalTolls(registry, factions, bravo, "Krylos"); len(got) != 0 {
		t.Errorf("tolls on the holder's own asset = %+v, want none", got)
	}
	bravo.Tags = nil
	if got := dispatch.ArrivalTolls(registry, factions, alpha, "Krylos"); len(got) != 0 {
		t.Errorf("tolls after the tag is dropped = %+v, want none", got)
	}
}
