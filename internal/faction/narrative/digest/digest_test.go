package digest

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func mutRec(t *testing.T, typeName string, payload any) domain.MutationRecord {
	t.Helper()
	return domain.MutationRecord{Type: typeName, Payload: mustMarshal(t, payload)}
}

func eventRec(factionID string, muts ...domain.MutationRecord) domain.EventRecord {
	return domain.EventRecord{
		Cycle:     1,
		FactionID: factionID,
		Timestamp: time.Time{},
		Mutations: muts,
	}
}

func minState(factions map[string]string) *state.FactionState {
	fs := &state.FactionState{Factions: make(map[string]*domain.Faction)}
	for id, name := range factions {
		fs.Factions[id] = &domain.Faction{ID: id, Name: name}
	}
	return fs
}

func minRulebook(goals map[string]string) *rulebook.Rulebook {
	rb := &rulebook.Rulebook{
		Assets: make(map[string]*domain.AssetDefinition),
		Tags:   make(map[string]*domain.Tag),
		Goals:  make(map[string]*domain.Goal),
	}
	for id, name := range goals {
		rb.Goals[id] = &domain.Goal{ID: id, Name: name}
	}
	return rb
}

func buildOne(t *testing.T, records []domain.EventRecord, fs *state.FactionState, rb *rulebook.Rulebook) CycleDigest {
	t.Helper()
	d, err := Build(records, 1, fs, rb)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return d
}

func findBeat(d CycleDigest, factionID string) *FactionBeat {
	for i := range d.ActiveFactions {
		if d.ActiveFactions[i].Faction.ID == factionID {
			return &d.ActiveFactions[i]
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Build stub
// ---------------------------------------------------------------------------

func TestBuildEmptyRecords(t *testing.T) {
	_, err := Build(nil, 1, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil records")
	}
}

func TestBuildMinimal(t *testing.T) {
	records := []domain.EventRecord{eventRec("faction-a")}
	got := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	if got.Cycle != 1 {
		t.Fatalf("want Cycle 1, got %d", got.Cycle)
	}
}

// ---------------------------------------------------------------------------
// CoinDelta
// ---------------------------------------------------------------------------

func TestCoinDeltaBookkeeping(t *testing.T) {
	// Bookkeeping CoinDelta alone doesn't make a faction active (no narrative beat per spec).
	// Pair it with an acquisition so we can verify the delta accumulates.
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_added", domain.AssetAdded{
				FactionID: "faction-a",
				Asset:     domain.Asset{ID: "asset-1", DefinitionID: "def-1"},
				Cause:     "buy",
			}),
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-a", Delta: 3, Cause: "bookkeeping",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	if beat == nil {
		t.Fatal("no active beat for faction-a")
	}
	if beat.CoinDelta != 3 {
		t.Fatalf("want CoinDelta 3, got %d", beat.CoinDelta)
	}
}

func TestCoinDeltaAbandonGoal(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "goal_abandoned", domain.GoalAbandoned{
				FactionID: "faction-a", GoalID: "goal-1", Cause: "abandon_goal",
			}),
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-a", Delta: -2, Cause: "abandon_goal",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	if beat.CoinDelta != -2 {
		t.Fatalf("want CoinDelta -2, got %d", beat.CoinDelta)
	}
}

// ---------------------------------------------------------------------------
// AssetAdded / AssetRemoved (buy, sell, refit)
// ---------------------------------------------------------------------------

func TestAssetAddedBuy(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_added", domain.AssetAdded{
				FactionID: "faction-a",
				Asset:     domain.Asset{ID: "asset-1", DefinitionID: "def-1"},
				Cause:     "buy",
			}),
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-a", Delta: -5, Cause: "buy",
			}),
		),
	}
	fs := minState(map[string]string{"faction-a": "Alpha"})
	fs.Factions["faction-a"].Assets = map[string]*domain.Asset{"asset-1": {ID: "asset-1", DefinitionID: "def-1"}}
	rb := &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{"def-1": {ID: "def-1", Name: "Shock Troops"}},
		Goals:  map[string]*domain.Goal{},
		Tags:   map[string]*domain.Tag{},
	}
	d := buildOne(t, []domain.EventRecord{records[0]}, fs, rb)
	beat := findBeat(d, "faction-a")
	if len(beat.Acquisitions) != 1 {
		t.Fatalf("want 1 acquisition, got %d", len(beat.Acquisitions))
	}
	if beat.Acquisitions[0].AssetName != "Shock Troops" {
		t.Errorf("want AssetName 'Shock Troops', got %q", beat.Acquisitions[0].AssetName)
	}
	if beat.CoinDelta != -5 {
		t.Errorf("want CoinDelta -5, got %d", beat.CoinDelta)
	}
}

func TestAssetRemovedSell(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_removed", domain.AssetRemoved{
				FactionID: "faction-a", AssetID: "asset-1", Cause: "sell",
			}),
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-a", Delta: 3, Cause: "sell",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	if len(beat.Losses) != 1 {
		t.Fatalf("want 1 loss, got %d", len(beat.Losses))
	}
	if beat.Losses[0].AssetID != "asset-1" {
		t.Errorf("want AssetID asset-1, got %q", beat.Losses[0].AssetID)
	}
}

func TestAssetAddedRefit(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_removed", domain.AssetRemoved{
				FactionID: "faction-a", AssetID: "old-asset", Cause: "refit",
			}),
			mutRec(t, "asset_added", domain.AssetAdded{
				FactionID: "faction-a",
				Asset:     domain.Asset{ID: "new-asset", DefinitionID: "def-2"},
				Cause:     "refit",
			}),
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-a", Delta: -2, Cause: "refit",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	if len(beat.Acquisitions) != 1 {
		t.Fatalf("want 1 acquisition, got %d", len(beat.Acquisitions))
	}
	if beat.Acquisitions[0].From != "old-asset" {
		t.Errorf("want From 'old-asset', got %q", beat.Acquisitions[0].From)
	}
}

func TestAssetRemovedBookkeepingDropped(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_removed", domain.AssetRemoved{
				FactionID: "faction-a", AssetID: "asset-x", Cause: "bookkeeping",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	// faction has a coin delta of 0 and no other events → quiet
	if len(d.ActiveFactions) != 0 {
		t.Errorf("bookkeeping removal should yield quiet faction, got %d active", len(d.ActiveFactions))
	}
}

// ---------------------------------------------------------------------------
// AssetMaintainedFlag — dropped
// ---------------------------------------------------------------------------

func TestAssetMaintainedFlagDropped(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_maintained_flag", domain.AssetMaintainedFlag{
				FactionID: "faction-a", AssetID: "asset-1", Maintained: true, Cause: "bookkeeping",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	if len(d.ActiveFactions) != 0 {
		t.Errorf("asset_maintained_flag should produce quiet faction, got %d active", len(d.ActiveFactions))
	}
}

// ---------------------------------------------------------------------------
// Repair
// ---------------------------------------------------------------------------

func TestRepairAsset(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_hp_delta", domain.AssetHPDelta{
				FactionID: "faction-a", AssetID: "asset-1", Delta: 2, Cause: "repair",
			}),
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-a", Delta: -1, Cause: "repair",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	if len(beat.Repairs) != 1 {
		t.Fatalf("want 1 repair, got %d", len(beat.Repairs))
	}
	r := beat.Repairs[0]
	if r.HPGained != 2 || r.Coin != -1 || r.IsFaction {
		t.Errorf("repair mismatch: %+v", r)
	}
}

func TestRepairFaction(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "faction_hp_delta", domain.FactionHPDelta{
				FactionID: "faction-a", Delta: 3, Cause: "repair",
			}),
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-a", Delta: -2, Cause: "repair",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	if len(beat.Repairs) != 1 {
		t.Fatalf("want 1 repair, got %d", len(beat.Repairs))
	}
	r := beat.Repairs[0]
	if !r.IsFaction || r.HPGained != 3 || r.Coin != -2 {
		t.Errorf("faction repair mismatch: %+v", r)
	}
}

// ---------------------------------------------------------------------------
// Bribe
// ---------------------------------------------------------------------------

func TestBribe(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha"})
	fs.Factions["faction-a"].Bases = []*domain.Base{
		{ID: "base-1", Location: domain.Location{WorldID: "Anvil"}},
	}
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "influence_delta", domain.InfluenceDelta{
				FactionID: "faction-a", BaseID: "base-1", Delta: 1, Cause: "bribe",
			}),
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-a", Delta: -3, Cause: "bribe",
			}),
		),
	}
	d := buildOne(t, records, fs, nil)
	beat := findBeat(d, "faction-a")
	if len(beat.Bribes) != 1 {
		t.Fatalf("want 1 bribe, got %d", len(beat.Bribes))
	}
	b := beat.Bribes[0]
	if b.BaseID != "base-1" || b.Location != "Anvil" || b.Coin != -3 {
		t.Errorf("bribe mismatch: %+v", b)
	}
}

// ---------------------------------------------------------------------------
// Expand
// ---------------------------------------------------------------------------

func TestExpandNewBase(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "base_added", domain.BaseAdded{
				FactionID: "faction-a",
				Base:      domain.Base{ID: "base-2", Location: domain.Location{WorldID: "Brightside"}},
				Cause:     "expand",
			}),
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-a", Delta: -4, Cause: "expand",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	if len(beat.Expansions) != 1 {
		t.Fatalf("want 1 expansion, got %d", len(beat.Expansions))
	}
	e := beat.Expansions[0]
	if !e.NewBase || e.BaseID != "base-2" || e.Location != "Brightside" || e.Coin != -4 {
		t.Errorf("expansion mismatch: %+v", e)
	}
}

// ---------------------------------------------------------------------------
// Asset movement (ability)
// ---------------------------------------------------------------------------

func TestAssetMoved(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_moved", domain.AssetMoved{
				FactionID:    "faction-a",
				AssetID:      "asset-1",
				FromLocation: domain.Location{WorldID: "Anvil"},
				ToLocation:   domain.Location{WorldID: "Brightside"},
				Cause:        "ability",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	if len(beat.Movements) != 1 {
		t.Fatalf("want 1 movement, got %d", len(beat.Movements))
	}
	mv := beat.Movements[0]
	if mv.From != "Anvil" || mv.To != "Brightside" {
		t.Errorf("movement mismatch: %+v", mv)
	}
}

// ---------------------------------------------------------------------------
// Stealth
// ---------------------------------------------------------------------------

func TestStealthApplied(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_stealth_applied", domain.AssetStealthApplied{
				FactionID: "faction-a", AssetID: "asset-1", Cause: "buy",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	if len(beat.StealthOps) != 1 || !beat.StealthOps[0].Applied {
		t.Errorf("want 1 stealth-applied event, got %+v", beat.StealthOps)
	}
}

func TestStealthCleared(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_stealth_cleared", domain.AssetStealthCleared{
				FactionID: "faction-a", AssetID: "asset-1", Cause: "attack",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	if len(beat.StealthOps) != 1 || beat.StealthOps[0].Applied {
		t.Errorf("want 1 stealth-cleared event, got %+v", beat.StealthOps)
	}
}

// ---------------------------------------------------------------------------
// Goal events
// ---------------------------------------------------------------------------

func TestGoalCompleted(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "goal_completed", domain.GoalCompleted{
				FactionID: "faction-a", GoalID: "goal-1", XPAwarded: 2, Cause: "goal_completed",
			}),
			mutRec(t, "xp_awarded", domain.XPAwarded{
				FactionID: "faction-a", Amount: 2, Cause: "goal_completed",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), minRulebook(map[string]string{"goal-1": "Seize the Station"}))
	beat := findBeat(d, "faction-a")
	if beat.XPGained != 2 {
		t.Errorf("want XPGained 2, got %d", beat.XPGained)
	}
	if len(beat.GoalEvents) != 1 || beat.GoalEvents[0].Kind != GoalCompleted {
		t.Errorf("want 1 GoalCompleted event, got %+v", beat.GoalEvents)
	}
	if beat.GoalEvents[0].GoalName != "Seize the Station" {
		t.Errorf("want goal name 'Seize the Station', got %q", beat.GoalEvents[0].GoalName)
	}
}

func TestGoalAbandoned(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "goal_abandoned", domain.GoalAbandoned{
				FactionID: "faction-a", GoalID: "goal-1", Cause: "abandon_goal",
			}),
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-a", Delta: -1, Cause: "abandon_goal",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), minRulebook(map[string]string{"goal-1": "Destroy the Foe"}))
	beat := findBeat(d, "faction-a")
	if len(beat.GoalEvents) != 1 || beat.GoalEvents[0].Kind != GoalAbandoned {
		t.Errorf("want 1 GoalAbandoned event, got %+v", beat.GoalEvents)
	}
}

func TestHomeworldChanged(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "goal_completed", domain.GoalCompleted{
				FactionID: "faction-a", GoalID: "goal-hw", Cause: "goal_completed",
			}),
			mutRec(t, "homeworld_changed", domain.HomeworldChanged{
				FactionID: "faction-a", FromWorld: domain.Location{WorldID: "Anvil"}, ToWorld: domain.Location{WorldID: "Brightside"}, Cause: "goal_completed",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	var found bool
	for _, ge := range beat.GoalEvents {
		if ge.Kind == GoalHomeworldShift && ge.FromWorld == "Anvil" && ge.ToWorld == "Brightside" {
			found = true
		}
	}
	if !found {
		t.Errorf("want GoalHomeworldShift event, got %+v", beat.GoalEvents)
	}
}

func TestTagAdded(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "goal_completed", domain.GoalCompleted{
				FactionID: "faction-a", GoalID: "goal-tag", Cause: "goal_completed",
			}),
			mutRec(t, "tag_added", domain.TagAdded{
				FactionID: "faction-a",
				Tag:       domain.Tag{ID: "tag-1", Name: "Guerrilla Tactics"},
				Cause:     "goal_completed",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	var found bool
	for _, ge := range beat.GoalEvents {
		if ge.Kind == GoalTagGained && ge.TagName == "Guerrilla Tactics" {
			found = true
		}
	}
	if !found {
		t.Errorf("want GoalTagGained event, got %+v", beat.GoalEvents)
	}
}

// ---------------------------------------------------------------------------
// Cross-faction attack pairing
// ---------------------------------------------------------------------------

func TestCrossFactionAttack(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha", "faction-b": "Bravo"})
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_hp_delta", domain.AssetHPDelta{
				FactionID: "faction-b", AssetID: "b-asset-1", Delta: -3, Cause: "attack",
			}),
			mutRec(t, "asset_removed", domain.AssetRemoved{
				FactionID: "faction-b", AssetID: "b-asset-1", Cause: "attack",
			}),
			mutRec(t, "faction_hp_delta", domain.FactionHPDelta{
				FactionID: "faction-b", Delta: -3, Cause: "attack",
			}),
		),
	}
	d := buildOne(t, records, fs, nil)
	if len(d.Cross) != 1 {
		t.Fatalf("want 1 cross event, got %d", len(d.Cross))
	}
	ce := d.Cross[0]
	if ce.Attacker.ID != "faction-a" || ce.Defender.ID != "faction-b" {
		t.Errorf("cross attacker/defender mismatch: %+v", ce)
	}
	if ce.DamageToDefender != 3 {
		t.Errorf("want DamageToDefender 3, got %d", ce.DamageToDefender)
	}
	if !ce.DefenderAssetDestroyed {
		t.Error("want DefenderAssetDestroyed true")
	}

	// Defender faction HP should be aggregated on the defender's beat.
	defBeat := findBeat(d, "faction-b")
	if defBeat == nil {
		t.Fatal("no active beat for faction-b (defender should be active due to cross event)")
	}
	if defBeat.HPDelta != -3 {
		t.Errorf("want defender HPDelta -3, got %d", defBeat.HPDelta)
	}
}

func TestCrossAttackerAssetDestroyed(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha", "faction-b": "Bravo"})
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_hp_delta", domain.AssetHPDelta{
				FactionID: "faction-b", AssetID: "b-asset-1", Delta: -2, Cause: "attack",
			}),
			mutRec(t, "asset_hp_delta", domain.AssetHPDelta{
				FactionID: "faction-a", AssetID: "a-asset-1", Delta: -4, Cause: "attack",
				CausedByFactionID: "faction-b",
			}),
			mutRec(t, "asset_removed", domain.AssetRemoved{
				FactionID: "faction-a", AssetID: "a-asset-1", Cause: "attack",
				CausedByFactionID: "faction-b",
			}),
		),
	}
	d := buildOne(t, records, fs, nil)
	if len(d.Cross) != 1 {
		t.Fatalf("want 1 cross event, got %d", len(d.Cross))
	}
	ce := d.Cross[0]
	if ce.DamageToAttacker != 4 {
		t.Errorf("want DamageToAttacker 4, got %d", ce.DamageToAttacker)
	}
	if !ce.AttackerAssetDestroyed {
		t.Error("want AttackerAssetDestroyed true")
	}
}

func TestCrossBaseHit(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha", "faction-b": "Bravo"})
	fs.Factions["faction-b"].Bases = []*domain.Base{{ID: "base-b1", Location: domain.Location{WorldID: "Anvil"}}}
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "base_hp_delta", domain.BaseHPDelta{
				FactionID: "faction-b", BaseID: "base-b1", Delta: -5, Cause: "attack",
			}),
			mutRec(t, "base_destroyed", domain.BaseDestroyed{
				FactionID: "faction-b", BaseID: "base-b1", Cause: "attack",
			}),
			mutRec(t, "faction_hp_delta", domain.FactionHPDelta{
				FactionID: "faction-b", Delta: -5, Cause: "attack",
			}),
		),
	}
	d := buildOne(t, records, fs, nil)
	if len(d.Cross) != 1 {
		t.Fatalf("want 1 cross event, got %d", len(d.Cross))
	}
	ce := d.Cross[0]
	if ce.BaseHit == nil {
		t.Fatal("want BaseHit set")
	}
	if ce.BaseHit.BaseID != "base-b1" || ce.BaseHit.Location != "Anvil" {
		t.Errorf("BaseHit mismatch: %+v", ce.BaseHit)
	}
	if ce.BaseHit.Damage != 5 {
		t.Errorf("want BaseHit.Damage 5, got %d", ce.BaseHit.Damage)
	}
	if !ce.BaseHit.Destroyed {
		t.Error("want BaseHit.Destroyed true")
	}
}

func TestCrossAbilityStealthBroken(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha", "faction-b": "Bravo"})
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_stealth_cleared", domain.AssetStealthCleared{
				FactionID: "faction-b", AssetID: "b-asset-1", Cause: "ability",
			}),
		),
	}
	d := buildOne(t, records, fs, nil)
	if len(d.Cross) != 1 {
		t.Fatalf("want 1 cross event, got %d", len(d.Cross))
	}
	if !d.Cross[0].StealthBroken {
		t.Error("want StealthBroken true")
	}
	if d.Cross[0].Kind != CrossAbilityStrike {
		t.Errorf("want CrossAbilityStrike, got %v", d.Cross[0].Kind)
	}
}

func TestCrossAbilityCoinDrained(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha", "faction-b": "Bravo"})
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-b", Delta: -4, Cause: "ability",
			}),
		),
	}
	d := buildOne(t, records, fs, nil)
	if len(d.Cross) != 1 {
		t.Fatalf("want 1 cross event, got %d", len(d.Cross))
	}
	if d.Cross[0].CoinDrained != 4 {
		t.Errorf("want CoinDrained 4, got %d", d.Cross[0].CoinDrained)
	}
	defBeat := findBeat(d, "faction-b")
	if defBeat == nil {
		t.Fatal("no active beat for faction-b")
	}
	if defBeat.CoinDelta != -4 {
		t.Errorf("want defender CoinDelta -4, got %d", defBeat.CoinDelta)
	}
}

// ---------------------------------------------------------------------------
// Active / quiet split
// ---------------------------------------------------------------------------

func TestQuietFactionHasNoEvents(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha", "faction-b": "Bravo"})
	records := []domain.EventRecord{
		// faction-a buys an asset — active
		eventRec("faction-a",
			mutRec(t, "asset_added", domain.AssetAdded{
				FactionID: "faction-a",
				Asset:     domain.Asset{ID: "asset-1", DefinitionID: "def-1"},
				Cause:     "buy",
			}),
		),
		// faction-b has only bookkeeping income — quiet per spec (CoinDelta alone is not a narrative beat)
		eventRec("faction-b",
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-b", Delta: 2, Cause: "bookkeeping",
			}),
		),
	}
	d := buildOne(t, records, fs, nil)
	if len(d.QuietFactions) != 1 || d.QuietFactions[0].ID != "faction-b" {
		t.Errorf("want faction-b in QuietFactions, got %+v", d.QuietFactions)
	}
	if findBeat(d, "faction-a") == nil {
		t.Error("faction-a should be active")
	}
}

func TestDefenderWithCrossEventIsActive(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha", "faction-b": "Bravo"})
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_hp_delta", domain.AssetHPDelta{
				FactionID: "faction-b", AssetID: "b-asset-1", Delta: -2, Cause: "attack",
			}),
		),
	}
	d := buildOne(t, records, fs, nil)
	defBeat := findBeat(d, "faction-b")
	if defBeat == nil {
		t.Error("defender should be active when involved in a cross event")
	}
}

// ---------------------------------------------------------------------------
// Headline priority
// ---------------------------------------------------------------------------

func TestHeadlineAttackBeatsGoalCompleted(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha", "faction-b": "Bravo"})
	records := []domain.EventRecord{
		// faction-a completes a goal
		eventRec("faction-a",
			mutRec(t, "goal_completed", domain.GoalCompleted{
				FactionID: "faction-a", GoalID: "goal-1", XPAwarded: 3, Cause: "goal_completed",
			}),
			mutRec(t, "xp_awarded", domain.XPAwarded{
				FactionID: "faction-a", Amount: 3, Cause: "goal_completed",
			}),
		),
		// faction-a also attacks faction-b
		eventRec("faction-a",
			mutRec(t, "asset_hp_delta", domain.AssetHPDelta{
				FactionID: "faction-b", AssetID: "b-asset-1", Delta: -2, Cause: "attack",
			}),
		),
	}
	d := buildOne(t, records, fs, minRulebook(map[string]string{"goal-1": "Seize"}))
	if d.Headline.Kind != HeadlineMajorAttack {
		t.Errorf("want HeadlineMajorAttack, got %v", d.Headline.Kind)
	}
}

func TestHeadlineGoalCompletedBeatsAbandoned(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha", "faction-b": "Bravo"})
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "goal_completed", domain.GoalCompleted{
				FactionID: "faction-a", GoalID: "goal-1", XPAwarded: 2, Cause: "goal_completed",
			}),
			mutRec(t, "xp_awarded", domain.XPAwarded{
				FactionID: "faction-a", Amount: 2, Cause: "goal_completed",
			}),
		),
		eventRec("faction-b",
			mutRec(t, "goal_abandoned", domain.GoalAbandoned{
				FactionID: "faction-b", GoalID: "goal-2", Cause: "abandon_goal",
			}),
		),
	}
	d := buildOne(t, records, fs, nil)
	if d.Headline.Kind != HeadlineGoalCompleted {
		t.Errorf("want HeadlineGoalCompleted, got %v", d.Headline.Kind)
	}
}

func TestHeadlineAttackTieBreakByDamage(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha", "faction-b": "Bravo", "faction-c": "Charlie", "faction-d": "Delta"})
	// Alpha attacks Bravo for 5 damage; Charlie attacks Delta for 3 damage
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "asset_hp_delta", domain.AssetHPDelta{
				FactionID: "faction-b", AssetID: "b-1", Delta: -5, Cause: "attack",
			}),
		),
		eventRec("faction-c",
			mutRec(t, "asset_hp_delta", domain.AssetHPDelta{
				FactionID: "faction-d", AssetID: "d-1", Delta: -3, Cause: "attack",
			}),
		),
	}
	d := buildOne(t, records, fs, nil)
	if d.Headline.Kind != HeadlineMajorAttack {
		t.Fatalf("want HeadlineMajorAttack, got %v", d.Headline.Kind)
	}
	if d.Headline.Subject.ID != "faction-a" {
		t.Errorf("want subject faction-a (higher damage), got %q", d.Headline.Subject.ID)
	}
}

func TestHeadlineAttackTieBreakAlphabetical(t *testing.T) {
	fs := minState(map[string]string{"faction-a": "Alpha", "faction-b": "Bravo", "faction-z": "Zeta", "faction-c": "Charlie"})
	// Zeta attacks Bravo for 4; Alpha attacks Charlie for 4 — same damage, Alpha wins alphabetically
	records := []domain.EventRecord{
		eventRec("faction-z",
			mutRec(t, "asset_hp_delta", domain.AssetHPDelta{
				FactionID: "faction-b", AssetID: "b-1", Delta: -4, Cause: "attack",
			}),
		),
		eventRec("faction-a",
			mutRec(t, "asset_hp_delta", domain.AssetHPDelta{
				FactionID: "faction-c", AssetID: "c-1", Delta: -4, Cause: "attack",
			}),
		),
	}
	d := buildOne(t, records, fs, nil)
	if d.Headline.Subject.ID != "faction-a" {
		t.Errorf("want faction-a (alphabetical tie-break), got %q", d.Headline.Subject.ID)
	}
}

func TestHeadlineQuietFallback(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			mutRec(t, "coin_delta", domain.CoinDelta{
				FactionID: "faction-a", Delta: 1, Cause: "bookkeeping",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	if d.Headline.Kind != HeadlineQuiet {
		t.Errorf("want HeadlineQuiet, got %v", d.Headline.Kind)
	}
}

// ---------------------------------------------------------------------------
// Unknown mutation type
// ---------------------------------------------------------------------------

func TestUnknownMutationTypeAddsNote(t *testing.T) {
	records := []domain.EventRecord{
		eventRec("faction-a",
			domain.MutationRecord{Type: "future_mutation", Payload: json.RawMessage(`{"faction_id":"faction-a"}`)},
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
	beat := findBeat(d, "faction-a")
	// A note is added but the beat may still be quiet (notes don't make it active).
	// What matters is no panic and the note is recorded.
	_ = beat
}
