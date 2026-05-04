package engine_test

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
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const (
	testDataDir = "../data"

	// Asset definitions used across tests. SecurityPersonnel is a Force asset
	// with cost 2, HP 3, 1d4 counter, and a Force-vs-Force / 1d3+1 attack.
	defSecurityPersonnel = "F1-001"
)

// --- harness ---

type harness struct {
	engine       *engine.Engine
	factionState *state.FactionState
	cfg          *config.Config
	collector    *testharness.ScriptedCollector
	observer     *testharness.RecordingObserver
}

// newHarness builds an Engine wired to the real rulebook, a fresh tmpdir for
// state + history paths, and a default ScriptedCollector + RecordingObserver.
// Caller mutates factionState and collector overrides before invoking RunCycle.
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

// addFaction appends a faction with one SecurityPersonnel asset on its homeworld.
func (h *harness) addFaction(id, homeworld string, force, cunning, wealth int) *domain.Faction {
	f := &domain.Faction{
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
	f.Assets = []*domain.Asset{{
		ID:           id + "-asset-1",
		DefinitionID: defSecurityPersonnel,
		OwnerID:      id,
		Location:     homeworld,
		CurrentHP:    3,
		Ready:        true,
		Maintained:   true,
	}}
	h.factionState.Factions[id] = f
	return f
}

// readHistory parses the history.jsonl file into a slice of EventRecords.
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

// fixedRoller returns values from a pre-set sequence in order; panics if the
// sequence is exhausted (so tests fail loudly instead of falling back to an
// implicit zero value).
type fixedRoller struct {
	values []int
	index  int
}

func (r *fixedRoller) Roll(_ int) int {
	if r.index >= len(r.values) {
		panic("fixedRoller exhausted")
	}
	v := r.values[r.index]
	r.index++
	return v
}

// assertKinds compares observer kinds to a wanted sequence.
func assertKinds(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("observer kinds:\n got  %v\n want %v", got, want)
	}
}

// findMutationByCause returns the first MutationRecord whose payload's "cause"
// field matches. Cause lives in each mutation's JSON payload, not on the
// MutationRecord wrapper, so we peek into Payload.
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

// --- test 1: two-faction full cycle ---

func TestRunCycle_TwoFactionsBothPickSellAsset(t *testing.T) {
	h := newHarness(t)
	h.addFaction("alpha", "Tartarus", 4, 3, 2)
	h.addFaction("beta", "Hadrian", 4, 3, 2)

	h.collector.SelectActionFn = func(_ *domain.Faction, available []action.Action) (action.Action, error) {
		for _, action := range available {
			if action.Name() == "Sell Asset" {
				return action, nil
			}
		}
		t.Fatalf("Sell Asset not in available actions")
		return nil, nil
	}
	h.collector.SelectAssetFn = func(assets []*domain.Asset, _ *loader.Rulebook) (*domain.Asset, error) {
		return assets[0], nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	wantKinds := []string{
		"TurnStarted", "GoalLockApplied", "BookkeepingApplied", "ActionSelected", "ActionResolved", "TurnCompleted",
		"TurnStarted", "GoalLockApplied", "BookkeepingApplied", "ActionSelected", "ActionResolved", "TurnCompleted",
		"CycleCompleted",
	}
	assertKinds(t, h.observer.Kinds(), wantKinds)

	// History: each faction emits 2 EventRecords — one for bookkeeping,
	// one for the Sell Asset resolution. No goal-lock events fire (LockNone).
	records := readHistory(t, h.cfg.HistoryPath)
	if len(records) != 4 {
		t.Fatalf("history records: got %d, want 4", len(records))
	}

	wantPerFaction := map[string]int{"alpha": 0, "beta": 0}
	sellRecords := 0
	for _, rec := range records {
		if _, known := wantPerFaction[rec.FactionID]; !known {
			t.Errorf("history: unexpected faction id %q", rec.FactionID)
			continue
		}
		wantPerFaction[rec.FactionID]++
		for _, m := range rec.Mutations {
			var fields struct {
				Cause string `json:"cause"`
			}
			if err := json.Unmarshal(m.Payload, &fields); err == nil && fields.Cause == "sell" {
				sellRecords++
			}
		}
	}
	if wantPerFaction["alpha"] != 2 || wantPerFaction["beta"] != 2 {
		t.Errorf("history record counts: got %v, want each faction = 2", wantPerFaction)
	}
	// Each Sell Asset emits AssetRemoved + CoinDelta (both with cause "sell"),
	// so two factions × two mutations = 4.
	if sellRecords != 4 {
		t.Errorf("sell mutations in history: got %d, want 4", sellRecords)
	}

	// State on disk reflects the final cycle state.
	if _, err := os.Stat(h.cfg.StatePath); err != nil {
		t.Errorf("state file missing: %v", err)
	}
}

// --- test 2: LockSkip — Change Homeworld in transit ---

func TestRunCycle_LockSkip_ChangeHomeworld(t *testing.T) {
	h := newHarness(t)
	locked := h.addFaction("alpha", "Tartarus", 4, 3, 2)
	locked.ActiveGoal = &domain.ActiveGoal{
		GoalID:         "G-012",
		TargetWorld:    "NewHome",
		TurnsRemaining: 2, // ticks to 1; goal is NOT yet complete this cycle
	}
	h.addFaction("beta", "Hadrian", 4, 3, 2)

	// Beta will pick Sell Asset; alpha is locked-skip and shouldn't reach SelectAction.
	h.collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID == "alpha" {
			t.Fatalf("SelectAction called for locked faction alpha")
		}
		for _, action := range available {
			if action.Name() == "Sell Asset" {
				return action, nil
			}
		}
		t.Fatalf("Sell Asset not in available actions")
		return nil, nil
	}
	h.collector.SelectAssetFn = func(assets []*domain.Asset, _ *loader.Rulebook) (*domain.Asset, error) {
		return assets[0], nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	// Both factions fire TurnStarted + GoalLockApplied + TurnCompleted; only
	// the unlocked faction reaches Bookkeeping/Action. CycleCompleted closes.
	// Order alpha-vs-beta first depends on the random rotation; assert by
	// presence and by per-faction structure.
	kinds := h.observer.Kinds()
	if got, want := countKind(kinds, "TurnStarted"), 2; got != want {
		t.Errorf("TurnStarted count: got %d, want %d", got, want)
	}
	if got, want := countKind(kinds, "GoalLockApplied"), 2; got != want {
		t.Errorf("GoalLockApplied count: got %d, want %d", got, want)
	}
	if got, want := countKind(kinds, "BookkeepingApplied"), 1; got != want {
		t.Errorf("BookkeepingApplied count: got %d, want %d (locked faction skips bookkeeping)", got, want)
	}
	if got, want := countKind(kinds, "ActionSelected"), 1; got != want {
		t.Errorf("ActionSelected count: got %d, want %d", got, want)
	}
	if got, want := countKind(kinds, "TurnCompleted"), 2; got != want {
		t.Errorf("TurnCompleted count: got %d, want %d", got, want)
	}
	if got, want := countKind(kinds, "CycleCompleted"), 1; got != want {
		t.Errorf("CycleCompleted count: got %d, want %d", got, want)
	}

	// GoalLockApplied for alpha must carry LockSkip + a single GoalTurnsTick.
	var alphaLock testharness.GoalLockPayload
	for _, ev := range h.observer.Events {
		if ev.Kind == "GoalLockApplied" && ev.Faction != nil && ev.Faction.ID == "alpha" {
			alphaLock = ev.Payload.(testharness.GoalLockPayload)
		}
	}
	if alphaLock.Lock.Type != goal.LockSkip {
		t.Errorf("alpha lock type: got %v, want LockSkip", alphaLock.Lock.Type)
	}
	if len(alphaLock.Mutations) != 1 {
		t.Fatalf("alpha lock mutations: got %d, want 1 (GoalTurnsTick)", len(alphaLock.Mutations))
	}

	// History must contain a goal-lock event for alpha with cause
	// "change_homeworld_transit", proving the orchestrator persisted the
	// lock mutations through applyAndRecord.
	records := readHistory(t, h.cfg.HistoryPath)
	rec, ok := findMutationByCause(records, "change_homeworld_transit")
	if !ok {
		t.Fatalf("history missing change_homeworld_transit mutation; records=%+v", records)
	}
	if rec.Type != "goal_turns_tick" {
		t.Errorf("change_homeworld_transit mutation type: got %s, want goal_turns_tick", rec.Type)
	}

	// Locked faction's TurnsRemaining ticked from 2 to 1.
	if got := h.factionState.Factions["alpha"].ActiveGoal.TurnsRemaining; got != 1 {
		t.Errorf("alpha TurnsRemaining: got %d, want 1", got)
	}
}

// --- test 3: LockRestrictActions — Planetary Seizure phase 1 ---

func TestRunCycle_LockRestrictActions_PlanetarySeizurePhase1(t *testing.T) {
	h := newHarness(t)
	attacker := h.addFaction("alpha", "Tartarus", 4, 3, 2)
	target := h.addFaction("beta", "Tartarus", 2, 2, 2) // co-located on Tartarus
	attacker.ActiveGoal = &domain.ActiveGoal{
		GoalID:          "G-004",
		ProcessPhase:    1,
		TargetFactionID: target.ID,
		TargetWorld:     "Tartarus",
		TurnsRemaining:  3,
	}

	// Use a deterministic roller so the attack succeeds and damage kills the
	// defender: 10 (attack) + 1 (defense) + 3 (damage from 1d3+1, rolled value 3 → damage 4).
	h.engine.Rand = &fixedRoller{values: []int{10, 1, 3}}

	// Capture what SelectAction sees for alpha. The lock should have already
	// filtered the available list down to Attack only.
	var alphaAvailable []string
	h.collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID == "alpha" {
			alphaAvailable = nil
			for _, action := range available {
				alphaAvailable = append(alphaAvailable, action.Name())
			}
			for _, action := range available {
				if action.Name() == "Attack" {
					return action, nil
				}
			}
			t.Fatalf("Attack not in alpha's available actions: %v", alphaAvailable)
		}
		// beta picks no action — keep the test focused on alpha's lock.
		return nil, nil
	}
	h.collector.SelectAttackersFn = func(eligible []*domain.Asset, _ *loader.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}
	h.collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *loader.Rulebook) (*domain.Asset, error) {
		return eligible[0], nil
	}
	h.collector.ConfirmRedirectToBaseFn = func(_ *domain.Faction, _ *domain.Base, _ int) (bool, error) {
		return false, nil
	}

	if err := h.engine.Turn.Start(h.factionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if len(alphaAvailable) != 1 || alphaAvailable[0] != "Attack" {
		t.Fatalf("alpha available actions under LockRestrictActions: got %v, want [Attack]", alphaAvailable)
	}

	// Observer fired ActionResolved at least once (alpha attacked beta).
	resolvedCount := 0
	for _, ev := range h.observer.Events {
		if ev.Kind == "ActionResolved" {
			resolvedCount++
		}
	}
	if resolvedCount != 1 {
		t.Errorf("ActionResolved count: got %d, want 1 (alpha attacks beta)", resolvedCount)
	}

	// History includes the attack mutation event(s).
	records := readHistory(t, h.cfg.HistoryPath)
	if _, ok := findMutationByCause(records, "attack"); !ok {
		t.Fatalf("history missing attack mutation; records=%+v", records)
	}

	// Beta's lone asset was destroyed.
	if got := len(h.factionState.Factions["beta"].Assets); got != 0 {
		t.Errorf("beta assets after attack: got %d, want 0", got)
	}
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
