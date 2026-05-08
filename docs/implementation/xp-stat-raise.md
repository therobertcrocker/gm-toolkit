# Implementation Plan: XP Stat Raise

## Overview

Factions earn XP by completing goals. At the beginning of each turn, before action selection, they may spend accumulated XP to raise a Force, Cunning, or Wealth rating by one. Raising a stat increases MaxHP and unlocks higher-tier assets. XP is consumed on spend.

This feature inserts a new **stat raise phase** into the `RunFactionTurn` orchestrator pipeline, between the existing bookkeeping checkpoint and action selection. It requires two new mutation types, a new `InputCollector` method, a new `TurnObserver` event, and integration tests that assert the full spending contract.

**Source rule:** SWN p.217, "Raising Faction Stats." XP cost = HP value of the new rating (same table as HP values). Rating cap is 8.

---

## Turn Pipeline After This Feature

```
Goal Lock Check
  → Bookkeeping (income + maintenance)
  → AwaitCheckpoint(bookkeeping)
  → [NEW] Stat Raise phase (elective; skipped if no stat is raiseable)
  → Action Selection
  → Action Resolution + Goal Progress + Reactor Dispatch
  → Apply + Record + Save
  → AwaitCheckpoint(action_result)
  → finishFactionTurn
```

No new checkpoint constant is added. The `SelectStatRaise` call itself is the pause point for interactive collectors. `ScriptedCollector` and AI collectors return immediately.

---

## Files Changed Per Phase

### Phase 1 — Domain: New Mutations + Apply Cases

**`internal/faction/domain/mutation.go`** — add two new types:

```go
// XPSpent decrements a faction's XP by Amount. Emitted by the stat raise phase.
type XPSpent struct {
    FactionID string `json:"faction_id"`
    Amount    int    `json:"amount"`
    Cause     string `json:"cause"`
}

func (m XPSpent) Type() string { return "xp_spent" }

// StatRaised increments one of a faction's attribute ratings by one and
// recalculates MaxHP. Emitted alongside XPSpent by the stat raise phase.
type StatRaised struct {
    FactionID string          `json:"faction_id"`
    Stat      domain.FactionStat `json:"stat"`
    OldRating int             `json:"old_rating"`
    NewRating int             `json:"new_rating"`
    Cause     string          `json:"cause"`
}

func (m StatRaised) Type() string { return "stat_raised" }
```

**`internal/faction/engine/mutation/mutation.go`** — add two switch cases to `Apply`:

```go
case domain.XPSpent:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        faction.XP -= v.Amount
    }
case domain.StatRaised:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        switch v.Stat {
        case domain.StatForce:
            faction.Force = v.NewRating
        case domain.StatCunning:
            faction.Cunning = v.NewRating
        case domain.StatWealth:
            faction.Wealth = v.NewRating
        }
        faction.MaxHP = domain.CalcMaxHP(faction)
    }
```

`StatRaised` sets the rating to `NewRating` (not `+= 1`) so Apply is idempotent if replayed.

---

### Phase 2 — Engine: Interface, Observer, Phase Implementation

**`internal/faction/engine/collector.go`** — add method to `InputCollector`:

```go
// SelectStatRaise is called at the start of each faction turn if the faction
// has enough XP to raise at least one attribute. eligible lists the stats that
// are affordable and below the cap. Return nil to skip the raise.
SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error)
```

**`internal/faction/engine/observer.go`** — add method to `TurnObserver`:

```go
// OnStatRaiseApplied fires after the stat raise phase resolves. raised is nil
// if the faction skipped; mutations is empty in that case.
OnStatRaiseApplied(faction *domain.Faction, raised *domain.FactionStat, mutations []domain.Mutation)
```

**`internal/faction/engine/testharness/observer.go`** — add `OnStatRaiseApplied` stub to `RecordingObserver`.

**`internal/faction/engine/testharness/collector.go`** — add to `ScriptedCollector`:

```go
SelectStatRaiseFn func(*domain.Faction, []domain.FactionStat) (*domain.FactionStat, error)

func (c *ScriptedCollector) SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error) {
    if c.SelectStatRaiseFn != nil {
        return c.SelectStatRaiseFn(faction, eligible)
    }
    return nil, nil  // default: skip raise
}
```

**`internal/faction/engine/action/actions/mocks/`** — regenerate `MockInputCollector`:

```
go generate ./internal/faction/engine/...
```

**`internal/faction/engine/stat_raise.go`** — new file in the `engine` package. Write three package-level functions (no `Engine` receiver needed):

```go
package engine

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

func eligibleStatRaises(faction *domain.Faction) []domain.FactionStat {
    type entry struct {
        stat   domain.FactionStat
        rating int
    }
    all := []entry{
        // fill in all three stats
    }
    eligible := make([]domain.FactionStat, 0, 3)
    for _, e := range all {
        // eligible if: e.rating < 8 AND faction.XP >= domain.HPValueForRating(e.rating+1)
    }
    return eligible
}

func buildStatRaiseMutations(faction *domain.Faction, stat domain.FactionStat) []domain.Mutation {
    var oldRating int
    switch stat {
    // resolve oldRating from faction fields
    }
    cost := // HPValueForRating of the next rating
    return []domain.Mutation{
        // XPSpent first, then StatRaised
        // NewRating = oldRating+1 (not a delta), Cause = "stat_raise" on both
    }
}

func prepareStatRaise(faction *domain.Faction, collector InputCollector) (*domain.FactionStat, []domain.Mutation, error) {
    eligible := eligibleStatRaises(faction)
    if len(eligible) == 0 {
        return nil, nil, nil // skip: don't call the collector
    }
    stat, err := // collector.SelectStatRaise(...)
    if err != nil { /* propagate */ }
    if stat == nil {
        return nil, nil, nil // player declined
    }
    return stat, buildStatRaiseMutations(faction, *stat), nil
}
```

**`internal/faction/engine/orchestrator.go`** — insert the stat raise phase between these two existing blocks (after bookkeeping checkpoint, before action selection):

```go
// --- existing, end of Phase 2 ---
if err := collector.AwaitCheckpoint(CheckpointBookkeeping); err != nil {
    observer.OnError(faction, err)
    return false, err
}

// TODO: insert Phase 2b here

// --- existing, start of Phase 3 ---
available := e.Action.AvailableActions(faction, factionState, e.Rulebook, collector)
```

Your Phase 2b block should:
1. Call `prepareStatRaise` — handle error via `observer.OnError` + return
2. If mutations came back (len > 0): call `e.applyAndRecord`, handle error
3. Fire `observer.OnStatRaiseApplied` only inside that same `len > 0` branch — not on skip

---

### Phase 3 — Integration Tests

Add `stat_raise_test.go` in `internal/faction/engine/testharness/integration/`.

**Harness quick-reference** (from `harness_test.go` / `fixtures_test.go`):
- `h := newHarness(t)` — engine + temp state + empty factions map + blank ScriptedCollector + RecordingObserver
- `alpha := h.addFaction("alpha", "Tartarus", force, cunning, wealth)` — returns `*domain.Faction`; faction has one Security Personnel asset
- Mutate `alpha` fields directly after `addFaction` to set XP, tags, etc.
- `h.engine.Turn.Start(h.factionState)` — must be called before `RunCycle`
- `h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer)` — runs one full cycle
- `readHistory(t, h.cfg.HistoryPath)` — returns `[]domain.EventRecord`
- `findMutationType(records, "xp_spent")` — returns `(MutationRecord, bool)`
- `checkStep(t, "description", condition, "detail on fail")` — labeled pass/fail assertion

---

**`TestRunCycle_StatRaise_XPSpent`** — the happy path

```go
func TestRunCycle_StatRaise_XPSpent(t *testing.T) {
    h := newHarness(t)
    alpha := h.addFaction("alpha", "Tartarus", 3, 2, 1)
    alpha.XP = 6 // exactly covers Force 3→4: HPValueForRating(4) = 6

    stat := domain.StatForce
    h.collector.SelectStatRaiseFn = func(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
        return &stat, nil
    }

    if err := h.engine.Turn.Start(h.factionState); err != nil {
        t.Fatalf("Turn.Start: %v", err)
    }
    if err := h.engine.RunCycle(h.factionState, h.cfg, h.collector, h.observer); err != nil {
        t.Fatalf("RunCycle: %v", err)
    }

    // assert: Force raised, XP consumed, MaxHP recalculated
    // assert: history contains "xp_spent" and "stat_raised"
    // unmarshal stat_raised payload and assert OldRating=3, NewRating=4
    //   hint: declare a local struct{ OldRating int `json:"old_rating"`; NewRating int `json:"new_rating"` }
    //         then json.Unmarshal(rec.Payload, &payload)
}
```

---

**`TestRunCycle_StatRaise_Skip`** — player has XP but declines

```go
func TestRunCycle_StatRaise_Skip(t *testing.T) {
    h := newHarness(t)
    alpha := h.addFaction("alpha", "Tartarus", 3, 2, 1)
    alpha.XP = 6

    h.collector.SelectStatRaiseFn = func(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
        return nil, nil // player declines
    }

    // start turn + run cycle

    // assert: Force still 3, XP still 6
    // assert: history does NOT contain "xp_spent"
}
```

---

**`TestRunCycle_StatRaise_IneligibleNoPrompt`** — collector must never fire when XP is too low

Key insight: cheapest possible raise is Wealth 1→2, costing `HPValueForRating(2) = 2`. Setting `alpha.XP = 1` guarantees no stat is eligible regardless of starting ratings.

```go
func TestRunCycle_StatRaise_IneligibleNoPrompt(t *testing.T) {
    h := newHarness(t)
    alpha := h.addFaction("alpha", "Tartarus", 3, 2, 1)
    alpha.XP = 1

    h.collector.SelectStatRaiseFn = func(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
        t.Fatal("SelectStatRaise called when no stats are eligible")
        return nil, nil
    }

    // start turn + run cycle

    // assert: Force still 3, XP still 1
}
```

---

**Existing scenarios:** No changes needed. `ScriptedCollector.SelectStatRaiseFn` defaults to nil → returns nil (skip). The new `"StatRaiseApplied"` observer event fires only when XP is actually spent, so existing `assertKinds` / `wantKinds` slices are not affected.

---

## Decisions to Ratify at Commit Time

Add to `decisions-log.md` under `feature/xp-stat-raise`:

| # | Decision | Rationale |
|---|----------|-----------|
| TBD | `SelectStatRaise` returns `*domain.FactionStat`; nil = skip | Idiomatic Go optional; avoids sentinel string values; same "nil = nothing chosen" pattern as `SelectAction` returning nil for No Action |
| TBD | Stat raise phase fires after `CheckpointBookkeeping`, before action selection | Player sees their post-bookkeeping XP balance before deciding; consistent with PDF rule "at the beginning of each turn" meaning before the action |
| TBD | No new checkpoint for stat raise | The `SelectStatRaise` call itself is the interactive pause; adding a checkpoint would double-pause the turn for a phase that only fires when the faction has enough XP |
| TBD | `StatRaised.Apply` sets rating to `NewRating` (not `+= 1`) | Idempotent under replay; the old rating is carried in the mutation for history readability |
| TBD | Eligibility computed in the engine, passed to collector as `eligible []domain.FactionStat` | Consistent with `AvailableActions` pattern; collector shows only valid options without needing to re-derive the cost table |
| TBD | Skip phase entirely when `eligibleStatRaises` is empty | Avoids calling the collector (and blocking interactive TUIs) for a no-op phase; IneligibleNoPrompt integration test asserts this |

---

## Docs Update

**`docs/rules/swn-faction-mechanics.md`** — update the turn sequence section to add stat raise as step 3 (before action):

```
1. Collect income
2. Pay maintenance
3. Spend XP to raise a stat (optional; skipped if no stat is raiseable or faction declines)
4. Take one action
```

This can be done in the same commit as Phase 1 or Phase 2.
