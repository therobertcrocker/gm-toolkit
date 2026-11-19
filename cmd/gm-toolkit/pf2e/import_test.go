package pf2e

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/foundry"
)

func TestPrintReport(t *testing.T) {
	report := &foundry.Report{
		Source:   "actors/kas.json",
		Creature: "kas",
		Kind:     creature.KindPC,
		Folder:   "party/kas",
		Skipped: []foundry.Skip{
			{Item: "Keen Eyes", What: "FlatModifier perception", Reason: foundry.ReasonConditional},
			{Item: "Bravery", What: "AdjustDegreeOfSuccess", Reason: foundry.ReasonLaterRule},
		},
		Unused: map[string]int{"Note": 6, "GrantItem": 4},
	}
	shown := &creature.Creature{
		Name:    "Kas Fringebramble",
		Kind:    creature.KindPC,
		Level:   14,
		Defense: creature.Defense{HP: 230, AC: 36},
		Layers:  []string{"import.toml"},
	}
	var out bytes.Buffer
	printReport(&out, report, shown, nil, false)
	for _, want := range []string{
		"actors/kas.json  wrote party/kas/import.toml",
		"Kas Fringebramble  pc 14",
		"Skipped, " + foundry.ReasonConditional + ":\n  Keen Eyes: FlatModifier perception",
		"Skipped, " + foundry.ReasonLaterRule + ":\n  Bravery: AdjustDegreeOfSuccess",
		"Not used by the sim: GrantItem ×4, Note ×6",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}

	out.Reset()
	printReport(&out, report, nil, errors.New("bad layer"), false)
	if !strings.Contains(out.String(), "Load failed: bad layer") {
		t.Errorf("output lacks the load error:\n%s", out.String())
	}

	out.Reset()
	printReport(&out, &foundry.Report{Source: "items/swear-fidelity.json"}, nil, nil, false)
	if !strings.Contains(out.String(), "items/swear-fidelity.json  not imported") {
		t.Errorf("output lacks the loose item line:\n%s", out.String())
	}
}
