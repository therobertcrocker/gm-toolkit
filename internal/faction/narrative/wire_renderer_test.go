package narrative

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

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
		name   string
		d      digest.CycleDigest
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
