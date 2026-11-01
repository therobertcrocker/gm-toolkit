package domain_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestEffectFiltersUsed(t *testing.T) {
	stealthy := false
	cases := []struct {
		name    string
		filters domain.EffectFilters
		want    []domain.EffectFilterName
	}{
		{
			name:    "owner and scope only",
			filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
			want:    nil,
		},
		{
			name: "every optional filter, stealthy false still counts",
			filters: domain.EffectFilters{
				Owner:        domain.OwnerRival,
				Scope:        domain.ScopeWorld,
				Attribute:    domain.StatCunning,
				Phases:       []domain.EffectRollPhase{domain.RollPhaseDefense},
				Action:       "Expand Influence",
				AssetTypes:   []domain.AssetType{domain.TypeStarship},
				Categories:   []domain.FactionStat{domain.StatForce},
				Stealthy:     &stealthy,
				MinTechLevel: 4,
				CausedBy:     domain.OwnerOwn,
			},
			want: []domain.EffectFilterName{
				domain.FilterAttribute, domain.FilterPhase, domain.FilterAction,
				domain.FilterAssetType, domain.FilterCategory, domain.FilterStealthy,
				domain.FilterMinTechLevel, domain.FilterCausedBy,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.filters.Used(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Used() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEffectGatesUsed(t *testing.T) {
	cases := []struct {
		name  string
		gates domain.EffectGates
		want  []domain.EffectGateName
	}{
		{name: "no gates", gates: domain.EffectGates{}, want: nil},
		{
			name: "every gate",
			gates: domain.EffectGates{
				OncePer: domain.OncePerTargetFaction,
				NoStack: true,
				Contest: &domain.EffectContest{Roller: domain.ContestRollerSelf},
			},
			want: []domain.EffectGateName{domain.GateOncePer, domain.GateNoStack, domain.GateContest},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.gates.Used(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Used() = %v, want %v", got, tc.want)
			}
		})
	}
}
