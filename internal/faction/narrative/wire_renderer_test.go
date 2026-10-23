package narrative

import (
	"flag"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/narrative/digest"
)

var updateGolden = flag.Bool("update", false, "regenerate golden files")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func attackDigest() digest.CycleDigest {
	attacker := digest.FactionRef{ID: "faction-a", Name: "Iron Hegemony"}
	defender := digest.FactionRef{ID: "faction-b", Name: "Free Worlds League"}
	return digest.CycleDigest{
		Cycle: 1,
		ActiveFactions: []digest.FactionBeat{
			{Faction: attacker},
			{Faction: defender},
		},
		Cross: []digest.CrossEvent{
			{
				Kind:                   digest.CrossAttack,
				Attacker:               attacker,
				Defender:               defender,
				AttackerAsset:          digest.AssetMove{AssetID: "a1", AssetName: "Shock Infantry"},
				DefenderAsset:          digest.AssetMove{AssetID: "b1", AssetName: "Security Forces"},
				DamageToDefender:       3,
				DamageToAttacker:       2,
				DefenderAssetDestroyed: true,
			},
		},
		Headline: digest.Headline{
			Kind:    digest.HeadlineMajorAttack,
			Subject: attacker,
		},
	}
}

func goalDigest() digest.CycleDigest {
	faction := digest.FactionRef{ID: "faction-a", Name: "Iron Hegemony"}
	return digest.CycleDigest{
		Cycle: 2,
		ActiveFactions: []digest.FactionBeat{
			{
				Faction:  faction,
				XPGained: 4,
				GoalEvents: []digest.GoalEvent{
					{
						Kind:      digest.GoalCompleted,
						GoalID:    "goal-1",
						GoalName:  "Expand Influence",
						XPAwarded: 4,
					},
				},
			},
		},
		Headline: digest.Headline{
			Kind:    digest.HeadlineGoalCompleted,
			Subject: faction,
		},
	}
}

func quietDigest() digest.CycleDigest {
	return digest.CycleDigest{
		Cycle: 3,
		QuietFactions: []digest.FactionRef{
			{ID: "faction-a", Name: "Iron Hegemony"},
			{ID: "faction-b", Name: "Free Worlds League"},
		},
		Headline: digest.Headline{Kind: digest.HeadlineQuiet},
	}
}

func TestWireRendererDeterminism(t *testing.T) {
	r := NewWireRenderer()
	d := attackDigest()

	out1, err := r.Render(d, 1)
	if err != nil {
		t.Fatal(err)
	}
	out2, err := r.Render(d, 1)
	if err != nil {
		t.Fatal(err)
	}
	if out1 != out2 {
		t.Error("same seed produced different output")
	}
}

func TestWireRendererVariation(t *testing.T) {
	r := NewWireRenderer()
	d := attackDigest()

	out1, err := r.Render(d, 1)
	if err != nil {
		t.Fatal(err)
	}
	out2, err := r.Render(d, 2)
	if err != nil {
		t.Fatal(err)
	}
	if out1 == out2 {
		t.Error("different seeds produced identical output")
	}
}

func TestWireRendererGolden(t *testing.T) {
	r := NewWireRenderer()
	cases := []struct {
		name string
		d    digest.CycleDigest
	}{
		{"attack", attackDigest()},
		{"goal_completed", goalDigest()},
		{"quiet", quietDigest()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := r.Render(tc.d, 42)
			if err != nil {
				t.Fatal(err)
			}
			goldenPath := filepath.Join("testdata", tc.name+".golden.md")
			if *updateGolden {
				if err := os.MkdirAll("testdata", 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden %s: %v (run with -update to generate)", goldenPath, err)
			}
			if string(want) != got {
				t.Errorf("golden mismatch for %s\nwant:\n%s\ngot:\n%s", tc.name, want, got)
			}
		})
	}
}

func TestRenderCrossEvent_NoDamageWithCondition(t *testing.T) {
	crossEvent := digest.CrossEvent{
		Kind:             digest.CrossAttack,
		Attacker:         digest.FactionRef{ID: "faction-a", Name: "Iron Hegemony"},
		Defender:         digest.FactionRef{ID: "faction-b", Name: "Free Worlds League"},
		DamageToDefender: 0,
		Conditions:       []domain.ConditionKind{domain.ConditionNoAbility},
	}

	for seed := range int64(len(crossNoDamageTemplates)) * 4 {
		got := renderCrossEvent(crossEvent, rand.New(rand.NewSource(seed)))
		if !strings.Contains(got, "no damage") {
			t.Errorf("seed %d: %q does not use a no-damage template", seed, got)
		}
		wantClause := " Free Worlds League cannot use asset abilities there until Iron Hegemony's next turn."
		if !strings.HasSuffix(got, wantClause) {
			t.Errorf("seed %d: %q does not end with the no_ability clause", seed, got)
		}
	}
}

func TestRenderCrossEvent_AttackWithInterception(t *testing.T) {
	crossEvent := digest.CrossEvent{
		Kind:             digest.CrossAttack,
		Attacker:         digest.FactionRef{ID: "faction-a", Name: "Iron Hegemony"},
		Defender:         digest.FactionRef{ID: "faction-b", Name: "Free Worlds League"},
		DamageToDefender: 5,
		Interceptions:    []domain.SpecialEffectType{domain.SpecialShelterAsset},
	}

	for seed := range int64(len(crossAttackTemplates)) * 4 {
		got := renderCrossEvent(crossEvent, rand.New(rand.NewSource(seed)))
		if !strings.HasSuffix(got, " A stricken asset went to ground in its boltholes.") {
			t.Errorf("seed %d: %q does not end with the shelter_asset clause", seed, got)
		}
	}
}

func TestRenderCrossEvent_AbilityStrikeWithCondition(t *testing.T) {
	crossEvent := digest.CrossEvent{
		Kind:       digest.CrossAbilityStrike,
		Attacker:   digest.FactionRef{ID: "faction-a", Name: "Iron Hegemony"},
		Defender:   digest.FactionRef{ID: "faction-b", Name: "Free Worlds League"},
		Conditions: []domain.ConditionKind{domain.ConditionDisabled},
	}

	for seed := range int64(len(crossAbilityTemplates)) * 4 {
		got := renderCrossEvent(crossEvent, rand.New(rand.NewSource(seed)))
		wantClause := " Free Worlds League must pay to bring a disabled asset back into use."
		if !strings.HasSuffix(got, wantClause) {
			t.Errorf("seed %d: %q does not end with the disabled clause", seed, got)
		}
	}
}

func TestRenderLoss_LevyCoin(t *testing.T) {
	move := digest.AssetMove{AssetID: "b1", AssetName: "Party Machine", Cause: "levy_coin"}
	for seed := range int64(len(lossTemplates)) * 4 {
		got := renderLoss(move, "Free Worlds League", rand.New(rand.NewSource(seed)))
		if !strings.HasSuffix(got, " (forfeit to a monopoly)") {
			t.Errorf("seed %d: %q does not end with the monopoly forfeit", seed, got)
		}
	}
}

func TestRenderCrossEvent_Counterstruck(t *testing.T) {
	crossEvent := digest.CrossEvent{
		Kind:             digest.CrossAttack,
		Attacker:         digest.FactionRef{ID: "faction-a", Name: "Iron Hegemony"},
		Defender:         digest.FactionRef{ID: "faction-b", Name: "Free Worlds League"},
		DamageToAttacker: 5,
		Counterstruck:    true,
	}

	got := renderCrossEvent(crossEvent, rand.New(rand.NewSource(1)))
	wantClause := " Free Worlds League's countermeasures turned the attack back on Iron Hegemony."
	if !strings.HasSuffix(got, wantClause) {
		t.Errorf("%q does not end with the counterstruck clause", got)
	}
}

func TestRenderGoalEvent_GovernmentAndTagLoss(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	governed := renderGoalEvent(digest.GoalEvent{Kind: digest.GoalWorldGoverned, ToWorld: "Krylos"}, "Alpha", rng)
	if !strings.Contains(governed, "Alpha") || !strings.Contains(governed, "Krylos") {
		t.Errorf("governed line = %q, want the faction and the world", governed)
	}
	lost := renderGoalEvent(digest.GoalEvent{Kind: digest.GoalTagLost, TagName: "Planetary Government"}, "Bravo", rng)
	if !strings.Contains(lost, "Bravo") || !strings.Contains(lost, "Planetary Government") {
		t.Errorf("tag-lost line = %q, want the faction and the tag", lost)
	}
}

func TestRenderRefusalAndHold_Landing(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	refusal := renderRefusal(digest.RefusalEvent{AssetName: "Troops", World: "Krylos", DecidedBy: "Bravo", Landing: true}, "Alpha", rng)
	if !strings.Contains(refusal, "land Troops") {
		t.Errorf("refusal = %q, want the landing verb", refusal)
	}
	raise := renderRefusal(digest.RefusalEvent{AssetName: "Troops", World: "Krylos", DecidedBy: "Bravo"}, "Alpha", rng)
	if !strings.Contains(raise, "raise Troops") {
		t.Errorf("refusal = %q, want the purchase verb", raise)
	}
	hold := renderHold(digest.HoldEvent{AssetName: "Freighter", World: "Krylos", Reason: domain.HoldRefused}, "Alpha", rng)
	if !strings.Contains(hold, "Freighter") || !strings.Contains(hold, "Krylos") {
		t.Errorf("hold = %q, want the asset and the world", hold)
	}
}

func TestRenderRefusal_WithdrawnAndBarred(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	withdrawn := renderRefusal(digest.RefusalEvent{AssetName: "Troops", World: "Krylos", DecidedBy: "Bravo", Landing: true, Reason: domain.RefusalWithdrawn}, "Alpha", rng)
	if !strings.Contains(withdrawn, "Bravo") || !strings.Contains(withdrawn, "land Troops") {
		t.Errorf("withdrawn line = %q, want the Lobbyists' owner and the landing", withdrawn)
	}
	barred := renderRefusal(digest.RefusalEvent{AssetName: "Troops", World: "Krylos", DecidedBy: "Bravo", Reason: domain.RefusalBarred}, "Alpha", rng)
	if !strings.Contains(barred, "Bravo") || !strings.Contains(barred, "asking again") {
		t.Errorf("barred line = %q, want the Lobbyists' owner and the barred wording", barred)
	}
}

func TestRenderHold_LockdownReasons(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for _, reason := range []domain.HoldReason{domain.HoldLockdownPaid, domain.HoldLockdownUnpaid} {
		line := renderHold(digest.HoldEvent{AssetName: "Freighter", World: "Krylos", Reason: reason}, "Alpha", rng)
		if !strings.Contains(line, "lockdown") || !strings.Contains(line, "Krylos") {
			t.Errorf("%s: hold line = %q, want the lockdown and the world", reason, line)
		}
	}
}
