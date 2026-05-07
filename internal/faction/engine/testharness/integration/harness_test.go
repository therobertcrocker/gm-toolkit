package integration

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/config"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/tag"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const (
	testDataDir          = "../../../data"
	defSecurityPersonnel = "F1-001"
	defHeavyDropAssets   = "F2-001" // Force 2, cost 4, TL4
)

type harness struct {
	engine       *engine.Engine
	factionState *state.FactionState
	cfg          *config.Config
	collector    *testharness.ScriptedCollector
	observer     *testharness.RecordingObserver
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	eng, err := engine.New(testDataDir)
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	actions.RegisterDefaultActions(eng)

	dir := t.TempDir()
	return &harness{
		engine:       eng,
		factionState: &state.FactionState{CampaignID: "test", Factions: make(map[string]*domain.Faction)},
		cfg: &config.Config{
			StatePath:   filepath.Join(dir, "state.toml"),
			HistoryPath: filepath.Join(dir, "history.jsonl"),
		},
		collector: &testharness.ScriptedCollector{},
		observer:  &testharness.RecordingObserver{},
	}
}

// registerTags wires tag hooks for all factions currently in h.factionState.
// Call after all addFaction calls so the state is populated before walking.
func (h *harness) registerTags() {
	tag.RegisterDefaultTags(h.engine, h.factionState)
}

// addFaction appends a faction with one Security Personnel asset on its homeworld.
func (h *harness) addFaction(id, homeworld string, force, cunning, wealth int) *domain.Faction {
	faction := &domain.Faction{
		ID:        id,
		Name:      id,
		Scale:     domain.ScaleMinor,
		Force:     force,
		Cunning:   cunning,
		Wealth:    wealth,
		Homeworld: homeworld,
		MaxHP:     20,
		CurrentHP: 20,
		Coin:      0,
	}
	faction.Assets = []*domain.Asset{{
		ID:           id + "-asset-1",
		DefinitionID: defSecurityPersonnel,
		OwnerID:      id,
		Location:     homeworld,
		CurrentHP:    3,
		Ready:        true,
		Maintained:   true,
	}}
	h.factionState.Factions[id] = faction
	return faction
}

func readHistory(t *testing.T, path string) []domain.EventRecord {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening history: %v", err)
	}
	defer file.Close()

	var records []domain.EventRecord
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var rec domain.EventRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			t.Fatalf("parsing history line %q: %v", scanner.Text(), err)
		}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanning history: %v", err)
	}
	return records
}

func countKind(kinds []string, want string) int {
	count := 0
	for _, k := range kinds {
		if k == want {
			count++
		}
	}
	return count
}

func assertKinds(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("observer kinds:\n got  %v\n want %v", got, want)
	}
}

// findMutationByCause returns the first MutationRecord whose JSON payload
// contains a "cause" field matching the given value.
func findMutationByCause(records []domain.EventRecord, cause string) (domain.MutationRecord, bool) {
	for _, rec := range records {
		for _, m := range rec.Mutations {
			var fields struct {
				Cause string `json:"cause"`
			}
			if err := json.Unmarshal(m.Payload, &fields); err != nil {
				continue
			}
			if fields.Cause == cause {
				return m, true
			}
		}
	}
	return domain.MutationRecord{}, false
}

// findMutationType returns the first MutationRecord with the given type discriminator.
func findMutationType(records []domain.EventRecord, mutType string) (domain.MutationRecord, bool) {
	for _, rec := range records {
		for _, m := range rec.Mutations {
			if m.Type == mutType {
				return m, true
			}
		}
	}
	return domain.MutationRecord{}, false
}

// findMutationByTypeAndCause returns the first MutationRecord matching both type and cause.
func findMutationByTypeAndCause(records []domain.EventRecord, mutType, cause string) (domain.MutationRecord, bool) {
	for _, rec := range records {
		for _, m := range rec.Mutations {
			if m.Type != mutType {
				continue
			}
			var fields struct {
				Cause string `json:"cause"`
			}
			if err := json.Unmarshal(m.Payload, &fields); err != nil {
				continue
			}
			if fields.Cause == cause {
				return m, true
			}
		}
	}
	return domain.MutationRecord{}, false
}
