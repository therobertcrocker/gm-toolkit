package dice

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// contextRecorder is a Cat 1 hook that offers nothing and records each roll's context.
type contextRecorder struct{ contexts []hooks.RollContext }

func (recorder *contextRecorder) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	recorder.contexts = append(recorder.contexts, ctx)
	return nil
}

// countingTieResolver answers with a fixed outcome and records each ask.
type countingTieResolver struct {
	outcome  hooks.TieOutcome
	contexts []hooks.RollContext
}

func (resolver *countingTieResolver) ResolveTie(ctx hooks.RollContext, _ *state.FactionState) hooks.TieOutcome {
	resolver.contexts = append(resolver.contexts, ctx)
	return resolver.outcome
}

func recordedEngine(values ...int) (Engine, *contextRecorder) {
	recorder := &contextRecorder{}
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "recorder", recorder)
	return Engine{Registry: registry, Roller: &dicetest.Sequence{Values: values}}, recorder
}

func contestFactions() (*domain.Faction, *domain.Faction) {
	return &domain.Faction{ID: "alpha", Force: 3, Cunning: 2}, &domain.Faction{ID: "beta", Force: 1, Cunning: 4}
}

func TestStat_TotalsTheDieAndStatAndFillsTheContext(t *testing.T) {
	alpha, beta := contestFactions()
	alphaAsset, betaAsset := &domain.Asset{ID: "a-1"}, &domain.Asset{ID: "b-1"}
	engine, recorder := recordedEngine(6)

	result := engine.Stat(
		domain.RollPhaseAttack,
		Side{Faction: alpha, Asset: alphaAsset, Stat: domain.StatForce, Detects: true},
		Side{Faction: beta, Asset: betaAsset, Stat: domain.StatCunning},
		"Krylos", "Attack",
	)

	if !slices.Equal(result.Dice, []int{6}) || result.Modifier != 3 || result.Sum != 9 {
		t.Errorf("result = %+v, want one die of 6, modifier 3, sum 9", result)
	}
	want := hooks.RollContext{
		Phase:             domain.RollPhaseAttack,
		Actor:             alpha,
		Opponent:          beta,
		Attribute:         domain.StatForce,
		OpponentAttribute: domain.StatCunning,
		Asset:             alphaAsset,
		OpponentAsset:     betaAsset,
		World:             "Krylos",
		Action:            "Attack",
		Detection:         true,
	}
	if len(recorder.contexts) != 1 || recorder.contexts[0] != want {
		t.Errorf("contexts = %+v, want only %+v", recorder.contexts, want)
	}
}

func TestStat_ZeroOpponentLeavesTheOpponentEmpty(t *testing.T) {
	alpha, _ := contestFactions()
	engine, recorder := recordedEngine(6)

	engine.Stat(domain.RollPhaseFactionTest, Side{Faction: alpha, Stat: domain.StatCunning}, Side{}, "Krylos", "Expand Influence")

	got := recorder.contexts[0]
	if got.Opponent != nil || got.OpponentAttribute != "" || got.OpponentAsset != nil {
		t.Errorf("context = %+v, want no opponent, opponent stat or opponent asset", got)
	}
}

func TestContest_RollsTheAttackerFirstWithItsPhases(t *testing.T) {
	cases := []struct {
		name         string
		kind         domain.ContestRoll
		attackPhase  domain.RollPhase
		defensePhase domain.RollPhase
	}{
		{"an attack", domain.ContestRollAttack, domain.RollPhaseAttack, domain.RollPhaseDefense},
		{"a test", domain.ContestRollTest, domain.RollPhaseFactionTest, domain.RollPhaseContested},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			alpha, beta := contestFactions()
			alphaAsset, betaAsset := &domain.Asset{ID: "a-1"}, &domain.Asset{ID: "b-1"}
			engine, recorder := recordedEngine(7, 2)

			outcome := engine.Contest(
				testCase.kind,
				Side{Faction: alpha, Asset: alphaAsset, Stat: domain.StatForce},
				Side{Faction: beta, Asset: betaAsset, Stat: domain.StatCunning},
				"Krylos", "Attack",
			)

			if outcome.Attack.Sum != 10 || outcome.Defense.Sum != 6 {
				t.Errorf("sums = %d and %d, want 10 from the first die and 6 from the second", outcome.Attack.Sum, outcome.Defense.Sum)
			}
			wantAttack := hooks.RollContext{Phase: testCase.attackPhase, Actor: alpha, Opponent: beta, Attribute: domain.StatForce, OpponentAttribute: domain.StatCunning, Asset: alphaAsset, OpponentAsset: betaAsset, World: "Krylos", Action: "Attack"}
			wantDefense := hooks.RollContext{Phase: testCase.defensePhase, Actor: beta, Opponent: alpha, Attribute: domain.StatCunning, OpponentAttribute: domain.StatForce, Asset: betaAsset, OpponentAsset: alphaAsset, World: "Krylos", Action: "Attack", OpponentTotal: recorder.contexts[1].OpponentTotal}
			if !slices.Equal(recorder.contexts, []hooks.RollContext{wantAttack, wantDefense}) {
				t.Errorf("contexts = %+v, want %+v then %+v", recorder.contexts, wantAttack, wantDefense)
			}
		})
	}
}

func TestContest_SameFactionDropsTheOpponent(t *testing.T) {
	alpha, _ := contestFactions()
	asset := &domain.Asset{ID: "a-1"}
	engine, recorder := recordedEngine(5, 5)

	engine.Contest(
		domain.ContestRollAttack,
		Side{Faction: alpha, Asset: asset, Stat: domain.StatForce},
		Side{Faction: alpha, Asset: asset, Stat: domain.StatCunning},
		"Krylos", "Attack",
	)

	for index, wantStats := range [][2]domain.FactionStat{{domain.StatForce, domain.StatCunning}, {domain.StatCunning, domain.StatForce}} {
		got := recorder.contexts[index]
		if got.Opponent != nil || got.OpponentAsset != nil {
			t.Errorf("roll %d: context = %+v, want no opponent and no opponent asset", index, got)
		}
		if got.Attribute != wantStats[0] || got.OpponentAttribute != wantStats[1] {
			t.Errorf("roll %d: stats = %s against %s, want %s against %s", index, got.Attribute, got.OpponentAttribute, wantStats[0], wantStats[1])
		}
	}
}

func TestContest_DefenderSeesTheAttackTotal(t *testing.T) {
	alpha, beta := contestFactions()
	engine, recorder := recordedEngine(7, 2)

	outcome := engine.Contest(
		domain.ContestRollAttack,
		Side{Faction: alpha, Stat: domain.StatForce},
		Side{Faction: beta, Stat: domain.StatCunning},
		"Krylos", "Attack",
	)

	if recorder.contexts[0].OpponentTotal != nil {
		t.Errorf("attack roll: OpponentTotal = %d, want nil", *recorder.contexts[0].OpponentTotal)
	}
	got := recorder.contexts[1].OpponentTotal
	if got == nil || *got != outcome.Attack.Sum {
		t.Errorf("defense roll: OpponentTotal = %v, want the attack's sum of %d", got, outcome.Attack.Sum)
	}
}

func TestContest_SameFactionStillPassesTheAttackTotal(t *testing.T) {
	alpha, _ := contestFactions()
	engine, recorder := recordedEngine(7, 2)

	outcome := engine.Contest(
		domain.ContestRollTest,
		Side{Faction: alpha, Stat: domain.StatForce},
		Side{Faction: alpha, Stat: domain.StatCunning},
		"Krylos", "Attack",
	)

	defense := recorder.contexts[1]
	if defense.Opponent != nil {
		t.Errorf("defense roll: Opponent = %+v, want nil", defense.Opponent)
	}
	if defense.OpponentTotal == nil || *defense.OpponentTotal != outcome.Attack.Sum {
		t.Errorf("defense roll: OpponentTotal = %v, want the attack's sum of %d", defense.OpponentTotal, outcome.Attack.Sum)
	}
}

func TestContest_NilDefenderAssetStaysNil(t *testing.T) {
	alpha, beta := contestFactions()
	engine, recorder := recordedEngine(5, 5)

	engine.Contest(
		domain.ContestRollAttack,
		Side{Faction: alpha, Asset: &domain.Asset{ID: "a-1"}, Stat: domain.StatForce},
		Side{Faction: beta, Stat: domain.StatCunning},
		"Krylos", "Attack",
	)

	if recorder.contexts[0].OpponentAsset != nil {
		t.Errorf("attack roll: OpponentAsset = %+v, want nil", recorder.contexts[0].OpponentAsset)
	}
	if recorder.contexts[1].Asset != nil {
		t.Errorf("defense roll: Asset = %+v, want nil", recorder.contexts[1].Asset)
	}
}

func TestOutcome_Readings(t *testing.T) {
	cases := []struct {
		name         string
		attack       int
		defense      int
		tie          hooks.TieOutcome
		attackLands  bool
		counterLands bool
		gatePasses   bool
	}{
		{"the attacker wins", 5, 3, hooks.TieStandard, true, false, true},
		{"the defender wins", 3, 5, hooks.TieStandard, false, true, false},
		{"a standard tie", 4, 4, hooks.TieStandard, true, true, false},
		{"a tie given to the attacker", 4, 4, hooks.TieAttackerWins, true, false, true},
		{"a tie given to the defender", 4, 4, hooks.TieDefenderWins, false, true, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			outcome := Outcome{Attack: hooks.RollResult{Sum: testCase.attack}, Defense: hooks.RollResult{Sum: testCase.defense}, tie: testCase.tie}
			if got := outcome.AttackLands(); got != testCase.attackLands {
				t.Errorf("AttackLands = %v, want %v", got, testCase.attackLands)
			}
			if got := outcome.CounterLands(); got != testCase.counterLands {
				t.Errorf("CounterLands = %v, want %v", got, testCase.counterLands)
			}
			if got := outcome.GatePasses(); got != testCase.gatePasses {
				t.Errorf("GatePasses = %v, want %v", got, testCase.gatePasses)
			}
		})
	}
}

func TestContest_AsksTheTieResolversOnlyOnAnAttacksTie(t *testing.T) {
	cases := []struct {
		name            string
		kind            domain.ContestRoll
		values          []int
		wantAsked       int
		wantAttackLands bool
	}{
		{"an attack's tie", domain.ContestRollAttack, []int{6, 5}, 1, false},
		{"a test's tie", domain.ContestRollTest, []int{6, 5}, 0, true},
		{"an attack with different sums", domain.ContestRollAttack, []int{6, 1}, 0, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			alpha, beta := contestFactions()
			asset := &domain.Asset{ID: "a-1"}
			resolver := &countingTieResolver{outcome: hooks.TieDefenderWins}
			registry := hooks.NewRegistry()
			registry.RegisterTieResolver(hooks.GlobalScope(), "resolver", resolver)
			engine := Engine{Registry: registry, Roller: &dicetest.Sequence{Values: testCase.values}}

			outcome := engine.Contest(
				testCase.kind,
				Side{Faction: alpha, Asset: asset, Stat: domain.StatForce},
				Side{Faction: beta, Stat: domain.StatCunning},
				"Krylos", "Attack",
			)

			if len(resolver.contexts) != testCase.wantAsked {
				t.Fatalf("resolver asked %d times, want %d", len(resolver.contexts), testCase.wantAsked)
			}
			if got := outcome.AttackLands(); got != testCase.wantAttackLands {
				t.Errorf("AttackLands = %v, want %v", got, testCase.wantAttackLands)
			}
			if testCase.wantAsked == 0 {
				return
			}
			want := hooks.RollContext{Phase: domain.RollPhaseAttack, Actor: alpha, Opponent: beta, Asset: asset, World: "Krylos", Action: "Attack"}
			if resolver.contexts[0] != want {
				t.Errorf("resolver context = %+v, want %+v", resolver.contexts[0], want)
			}
		})
	}
}

func TestContest_SameFactionTieNamesNoOpponent(t *testing.T) {
	alpha, _ := contestFactions()
	resolver := &countingTieResolver{}
	registry := hooks.NewRegistry()
	registry.RegisterTieResolver(hooks.GlobalScope(), "resolver", resolver)
	engine := Engine{Registry: registry, Roller: &dicetest.Sequence{Values: []int{5, 6}}}

	engine.Contest(
		domain.ContestRollAttack,
		Side{Faction: alpha, Stat: domain.StatForce},
		Side{Faction: alpha, Stat: domain.StatCunning},
		"Krylos", "Attack",
	)

	if len(resolver.contexts) != 1 || resolver.contexts[0].Opponent != nil {
		t.Errorf("resolver contexts = %+v, want one with no opponent", resolver.contexts)
	}
}

func gateContest() domain.Contest {
	return domain.Contest{AttackerStat: domain.StatForce, DefenderStat: domain.StatCunning, Roll: domain.ContestRollTest}
}

func TestGate_RollerSelf(t *testing.T) {
	cases := []struct {
		name   string
		values []int
		want   bool
	}{
		{"a win passes", []int{6, 1}, true},
		{"a tie fails", []int{6, 5}, false},
		{"a loss fails", []int{1, 6}, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			alpha, beta := contestFactions()
			engine, recorder := recordedEngine(testCase.values...)

			got := engine.Gate(gateContest(), domain.ContestRollerSelf, Side{Faction: alpha}, Side{Faction: beta}, "Krylos", "Use Asset Ability")

			if got != testCase.want {
				t.Errorf("Gate = %v, want %v", got, testCase.want)
			}
			first := recorder.contexts[0]
			if first.Actor != alpha || first.Attribute != domain.StatForce || first.OpponentAttribute != domain.StatCunning {
				t.Errorf("first roll = %+v, want alpha rolling Force against Cunning", first)
			}
		})
	}
}

func TestGate_RollerTarget(t *testing.T) {
	cases := []struct {
		name   string
		values []int
		want   bool
	}{
		{"the target's failed test passes", []int{2, 6}, true},
		{"a tie passes", []int{4, 3}, true},
		{"the target's won test fails", []int{6, 1}, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			alpha, beta := contestFactions()
			engine, recorder := recordedEngine(testCase.values...)

			got := engine.Gate(gateContest(), domain.ContestRollerTarget, Side{Faction: alpha}, Side{Faction: beta}, "Krylos", "")

			if got != testCase.want {
				t.Errorf("Gate = %v, want %v", got, testCase.want)
			}
			first, second := recorder.contexts[0], recorder.contexts[1]
			if first.Actor != beta || first.Attribute != domain.StatForce {
				t.Errorf("first roll = %+v, want the target rolling the contest's AttackerStat", first)
			}
			if second.Actor != alpha || second.Attribute != domain.StatCunning {
				t.Errorf("second roll = %+v, want the declaring side rolling the contest's DefenderStat", second)
			}
		})
	}
}

func TestGate_DetectsMarksOnlyTheDeclaringSide(t *testing.T) {
	cases := []struct {
		name          string
		roller        domain.ContestRoller
		wantDetection []bool
	}{
		{"roller self", domain.ContestRollerSelf, []bool{true, false}},
		{"roller target", domain.ContestRollerTarget, []bool{false, true}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			alpha, beta := contestFactions()
			engine, recorder := recordedEngine(5, 5)

			engine.Gate(gateContest(), testCase.roller, Side{Faction: alpha, Detects: true}, Side{Faction: beta}, "Krylos", "")

			got := []bool{recorder.contexts[0].Detection, recorder.contexts[1].Detection}
			if !slices.Equal(got, testCase.wantDetection) {
				t.Errorf("Detection by roll = %v, want %v", got, testCase.wantDetection)
			}
		})
	}
}
