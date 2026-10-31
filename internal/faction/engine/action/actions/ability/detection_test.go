package ability

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// detectionRecorder is a Cat 1 hook that records whether each roll it sees is a
// detection roll, and offers nothing.
type detectionRecorder struct {
	detection []bool
}

func (recorder *detectionRecorder) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	recorder.detection = append(recorder.detection, ctx.Detection)
	return nil
}

func TestPassesGate_RevealMarksTheActorsRollAsDetection(t *testing.T) {
	cases := []struct {
		name   string
		effect domain.AbilityEffectType
		want   []bool
	}{
		{"reveal_stealth", domain.EffectRevealStealth, []bool{true, false}},
		{"another contest", domain.EffectDisableAsset, []bool{false, false}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			informers := &domain.Asset{ID: "inf", Location: domain.Location{WorldID: "Krylos"}}
			alpha := &domain.Faction{ID: "alpha", Cunning: 3, Assets: map[string]*domain.Asset{"inf": informers}}
			beta := &domain.Faction{ID: "beta", Cunning: 2}
			factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": beta}}
			recorder := &detectionRecorder{}
			registry := hooks.NewRegistry()
			registry.RegisterRollModifier(hooks.GlobalScope(), "recorder", recorder)
			step := domain.AbilityStep{
				Effect:  tc.effect,
				Target:  domain.TargetFaction,
				Contest: &domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
			}
			target := domain.AbilityTarget{Kind: domain.TargetFaction, Faction: beta}

			if _, err := passesGate(step, alpha, informers, target, registry, nil, &seqRoller{values: []int{5}}, factionState, nil); err != nil {
				t.Fatalf("passesGate: %v", err)
			}
			if !reflect.DeepEqual(recorder.detection, tc.want) {
				t.Errorf("Detection per roll = %v, want %v (attacker, then defender)", recorder.detection, tc.want)
			}
		})
	}
}
