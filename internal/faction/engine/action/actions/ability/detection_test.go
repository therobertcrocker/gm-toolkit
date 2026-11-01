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

// tieToAttacker is a Cat 5 hook that gives every tie to the attacker (Fanatical's
// lose_ties on the defender).
type tieToAttacker struct{}

func (tieToAttacker) ResolveTie(hooks.RollContext, *state.FactionState) hooks.TieOutcome {
	return hooks.TieAttackerWins
}

func TestPassesGate_AnAttackContestAsksTheTieResolvers(t *testing.T) {
	phases := &phaseRecorder{}
	cases := []struct {
		name       string
		roll       domain.ContestRoll
		wantPass   bool
		wantPhases []hooks.RollPhase
	}{
		{"attack: the tie goes to the attacker", domain.ContestRollAttack, true, []hooks.RollPhase{hooks.PhaseAttack, hooks.PhaseDefense}},
		{"test: the tie stays with the defender", domain.ContestRollTest, false, []hooks.RollPhase{hooks.PhaseFactionTest, hooks.PhaseContested}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			phases.seen = nil
			informers := &domain.Asset{ID: "inf", Location: domain.Location{WorldID: "Krylos"}}
			alpha := &domain.Faction{ID: "alpha", Cunning: 2, Assets: map[string]*domain.Asset{"inf": informers}}
			beta := &domain.Faction{ID: "beta", Cunning: 2}
			factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": beta}}
			registry := hooks.NewRegistry()
			registry.RegisterRollModifier(hooks.GlobalScope(), "phases", phases)
			registry.RegisterTieResolver(hooks.FactionScope("beta"), "lose_ties", tieToAttacker{})
			step := domain.AbilityStep{
				Effect:  domain.EffectRevealStealth,
				Target:  domain.TargetFaction,
				Contest: &domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning, Roll: tc.roll},
			}
			target := domain.AbilityTarget{Kind: domain.TargetFaction, Faction: beta}

			passed, err := passesGate(step, alpha, informers, target, registry, nil, &seqRoller{values: []int{5, 5}}, factionState, nil)
			if err != nil {
				t.Fatalf("passesGate: %v", err)
			}
			if passed != tc.wantPass {
				t.Errorf("passed = %v, want %v", passed, tc.wantPass)
			}
			if !reflect.DeepEqual(phases.seen, tc.wantPhases) {
				t.Errorf("phases = %v, want %v", phases.seen, tc.wantPhases)
			}
		})
	}
}

// phaseRecorder is a Cat 1 hook that records each roll's phase and offers nothing.
type phaseRecorder struct {
	seen []hooks.RollPhase
}

func (recorder *phaseRecorder) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	recorder.seen = append(recorder.seen, ctx.Phase)
	return nil
}
