package testharness

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/config"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/tag"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

const (
	DefSecurityPersonnel = "F1-001"
	DefHeavyDropAssets   = "F2-001"
)

type Harness struct {
	Engine       *engine.Engine
	FactionState *state.FactionState
	Cfg          *config.Config
	Collector    *ScriptedCollector
	Observer     *RecordingObserver
}

// stubSpatialMap accepts any location ID, returning TL 5 and distance 1.
// Used by the test harness so the world index is always populated without
// requiring real spatial data files.
type stubSpatialMap struct{}

func (s *stubSpatialMap) Location(id string) (spatial.Location, bool) {
	return &stubLocation{id: id}, true
}

func (s *stubSpatialMap) Distance(_, _ string, _ int) (int, error) {
	return 1, nil
}

type stubLocation struct{ id string }

func (l *stubLocation) ID() string      { return l.id }
func (l *stubLocation) Name() string    { return l.id }
func (l *stubLocation) TechLevel() int  { return 5 }
func (l *stubLocation) Population() int { return 0 }

func NewHarness(t *testing.T, dataDir string) *Harness {
	t.Helper()
	dir := t.TempDir()
	cfg := &config.Config{
		FactionDataDir: dataDir,
		StatePath:      filepath.Join(dir, "state.toml"),
		HistoryPath:    filepath.Join(dir, "history.jsonl"),
	}
	eng, err := engine.New(cfg)
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	eng.World = world.NewWithMap(&stubSpatialMap{})
	actions.RegisterDefaultActions(eng)

	return &Harness{
		Engine:       eng,
		FactionState: &state.FactionState{CampaignID: "test", Factions: make(map[string]*domain.Faction)},
		Cfg:          cfg,
		Collector:    &ScriptedCollector{},
		Observer:     &RecordingObserver{},
	}
}

func (h *Harness) RegisterTags() {
	tag.RegisterDefaultTags(h.Engine, h.FactionState)
}

func (h *Harness) AddFaction(id, homeworld string, force, cunning, wealth int) *domain.Faction {
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
	faction.Assets = map[string]*domain.Asset{
		id + "-asset-1": {
			ID:           id + "-asset-1",
			DefinitionID: DefSecurityPersonnel,
			OwnerID:      id,
			Location:     homeworld,
			CurrentHP:    3,
			Ready:        true,
			Maintained:   true,
		},
	}
	h.FactionState.Factions[id] = faction
	return faction
}

func ReadHistory(t *testing.T, path string) []domain.EventRecord {
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

func CountKind(kinds []string, want string) int {
	count := 0
	for _, k := range kinds {
		if k == want {
			count++
		}
	}
	return count
}

func AssertKinds(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("observer kinds:\n got  %v\n want %v", got, want)
	}
}

func FindMutationByCause(records []domain.EventRecord, cause string) (domain.MutationRecord, bool) {
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

func FindMutationType(records []domain.EventRecord, mutType string) (domain.MutationRecord, bool) {
	for _, rec := range records {
		for _, m := range rec.Mutations {
			if m.Type == mutType {
				return m, true
			}
		}
	}
	return domain.MutationRecord{}, false
}

func FindMutationByTypeAndCause(records []domain.EventRecord, mutType, cause string) (domain.MutationRecord, bool) {
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

func AddBase(faction *domain.Faction, world string, hp int) *domain.Base {
	base := &domain.Base{
		ID:          fmt.Sprintf("%s-base-%s-1", faction.ID, world),
		OwnerID:     faction.ID,
		Location:    world,
		CurrentHP:   hp,
		MaxHP:       hp,
		Ready:       true,
		IsHomeworld: false,
	}
	faction.Bases = append(faction.Bases, base)
	return base
}

func AddAssetOnWorld(faction *domain.Faction, world string) *domain.Asset {
	asset := &domain.Asset{
		ID:           fmt.Sprintf("%s-%s-extra", faction.ID, world),
		DefinitionID: DefSecurityPersonnel,
		OwnerID:      faction.ID,
		Location:     world,
		CurrentHP:    3,
		Ready:        true,
		Maintained:   true,
	}
	faction.Assets[asset.ID] = asset
	return asset
}

func CheckStep(t *testing.T, description string, ok bool, detail string) {
	t.Helper()
	if ok {
		t.Logf("  ✓ %s", description)
	} else {
		t.Errorf("  ✗ %s: %s", description, detail)
	}
}
