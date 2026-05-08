# GM Toolkit — Design Review

**Date:** 2026-05-08  
**Scope:** Full codebase, all packages under `cmd/` and `internal/`  
**Question answered:** *For what this tool is actually doing, is this the best way to do it?*

This is a design critique, not a correctness audit. The goal is to assess whether the architectural choices were right for the end result — a CLI tool for one GM to run SWN faction turns — not to confirm that the implemented design works correctly (it does).

---

## What the Tool Actually Does

Before critiquing, pin the actual user-facing behavior:

1. A GM runs a faction turn: bookkeeping (income, maintenance), optional stat raise, one action per faction.
2. The tool enforces the SWN rules for each action, rolls dice, and writes state to disk.
3. After a cycle, the GM can render a narrative summary from the history log.
4. State can be paused mid-turn and resumed.

That's it. The tool is a **rules-enforcing bookkeeping assistant** that produces a **human-readable log**. It runs locally, handles maybe 10 factions and 50 assets, and is used by one person at a time.

---

## What Was Right

### The `cmd/` vs. `internal/` split

Clean and load-bearing. The engine knows nothing about terminals, Cobra, or `huh`. Every interactive decision routes through `InputCollector`; every output routes through `TurnObserver`. The result: the engine is headless, testable in isolation, and the planned TUI slot-replacement is a real thing rather than architectural fiction.

This is the best structural decision in the codebase. It would have been easy to let `huh` calls bleed into action resolution — keeping that boundary hard took discipline and pays off concretely in the test suite.

### Mutations for history recording

The decision to express state changes as typed `Mutation` values rather than direct writes has a real, concrete payoff: the JSONL history log is produced for free, and `digest.Build` can reconstruct the narrative of a cycle without replaying the game state. The mutation log is the right granularity for the narrative layer.

The `Cause` and `CausedByFactionID` attribution fields on every mutation are essential — the goal progress tracker would be meaningless without them, and adding them retroactively would have been painful.

### The hooks system

The five-category hook system is well-designed for the problem. Looking at the full tag and asset catalogs, all five categories are load-bearing:

- **Cat 1 (RollModifier):** Ten or more tags (Warlike, Machiavellian, Plutocratic, Deep Rooted, Theocratic, Savage, Imperialists, Exchange Consulate, Perimeter Agency, Eugenics Cult) all add +1d10 keep-highest to specific rolls under specific conditions. The budget key system is needed here — each of these is "once per turn."
- **Cat 2 (RollResultHook):** Fanatical rerolls dice showing 1; Psychic Academy forces a rival to reroll any die; Book of Secrets (asset) does both. These are genuinely post-roll effects that cannot be Cat 1.
- **Cat 3 (MutationReactor):** Scavengers (coin per kill), False Front (intercept a destruction), Tripwire Cells (attack on stealth mutation), Boltholes (intercept destruction), Cracked Comms (self-attack on defend success), Treachery (asset switches sides on kill), Franchise (coin drain on successful attack), Panopticon Matrix (turn-start stealth test). This is a large, genuinely reactive class.
- **Cat 4 (Rule Modifiers):** Preceptor Archive (cost reduction), Technical Expertise (world tech level floor), Mercenary Group (movement grant for all assets), Pirates (cost for other factions' movement), Local Investments (extra cost to buy on the world). All four modifier interfaces have named consumers.
- **Cat 5 (TieResolver):** Fanatical always loses ties during attacks.

The recursive Cat 3 dispatch is also justified. Tripwire Cells is a Cat 3 reactor that fires on `AssetStealthApplied` mutations — it can produce `AssetRemoved` mutations, which in turn trigger the Scavengers reactor. That's a real chaining scenario, not a hypothetical.

The complexity is real, but proportionate. A two-hook simplification would have been under-building for this game.

### Turn resumability design

`TurnState` persisting the faction order, current index, and phase to disk is exactly the right granularity. A crash mid-turn resumes from the right faction's bookkeeping phase. For a GM tool where "mid-turn" can mean "close the laptop," this matters.

### Test harness architecture

`testharness.Harness` + `ScriptedCollector` + `RecordingObserver` is clean infrastructure. The scenario tests in `testharness/scenarios/` cover the full turn pipeline with explicit scripting of every GM decision. The `FixedRoller` injection makes dice deterministic. This is the right test structure for a system where most interesting behavior is at the integration level.

### TOML for state, JSONL for history

TOML for state is correct. The file is human-readable and hand-editable — that's not a nice-to-have for a GM tool, it's essential. The GM needs to be able to fix a bad state without the tool's help.

JSONL for history is correct. Append-only, machine-parseable, never rewritten.

---

## What Was Wrong, or at Minimum Questionable

### 1. The `Action` interface has a temporal coupling problem

```go
type Action interface {
    Validate(faction, state, rulebook) bool
    Inputs(faction, state, rulebook) error    // stores data on struct
    Resolve(faction, state, rulebook) error   // reads data Inputs stored
    Output() ([]Mutation, error)
}
```

`Resolve` depends on `Inputs` having been called first. There's no type-level guarantee of this. `ActionEngine.Run` enforces the order, but if you call `Resolve` before `Inputs` — as could happen in a test or future caller — you get zero results or a panic, not a compile error.

The `Output() ([]Mutation, error)` return is also a design smell: every implementation returns `nil` for the error. The method signature promises something it never delivers.

A simpler signature that makes the data flow explicit:

```go
type Action interface {
    Name() string
    Validate(faction, state, rulebook) bool
    Run(faction, state, rulebook, collector) ([]Mutation, error)
}
```

`Run` combines `Inputs` + `Resolve` + `Output` into one method. The struct is created fresh per turn (already the case). The data collected mid-resolution stays local to the method. No temporal coupling; no unused error return.

The reason the current design exists (Decision 47, 48) is "clean separation of concerns." In practice, `Inputs` and `Resolve` are always called together in sequence by `ActionEngine.Run`. They're one operation with a seam inside it that serves no current purpose.

### 2. `UseAssetAbility` has a hidden type assertion

In `register.go`:

```go
e.Action.Register(func(c action.Collector) action.Action {
    return NewUseAssetAbility(c.(engine.InputCollector), e.Rand, e.Ability)
})
```

`c.(engine.InputCollector)` is an unguarded type assertion. `action.Collector` is a subset of `engine.InputCollector`. If any caller ever passes an `action.Collector` that doesn't also implement `engine.InputCollector` — a test stub, for example — this panics at runtime.

This reveals a structural inconsistency: the action registration system uses `action.Collector` as the factory parameter, but `UseAssetAbility` requires the full `engine.InputCollector`. The clean fix is to change the factory signature to take `engine.InputCollector` directly, since every real implementation satisfies it anyway.

### 3. The goal engine is inconsistent with the action pattern

Actions use a handler registry:

```go
// register.go
e.Action.Register(func(c Collector) Action { return NewAttack(...) })
e.Action.Register(func(c Collector) Action { return NewBuyAsset(...) })
```

Goals use a direct switch on hardcoded IDs:

```go
// progress.go
switch actingFaction.ActiveGoal.GoalID {
case "G-001":
    return progressMilitaryConquest(...)
case "G-004":
    return progressPlanetarySeizure(...)
```

These are the same pattern — "given a typed entity, dispatch to the right handler" — implemented two different ways. The result is that adding a new goal requires touching `progress.go` and `goal.go` in two separate places, while adding a new action requires touching only `register.go`.

A goal handler registry would be consistent and reduce the surface area for new goal additions:

```go
type GoalHandlers struct {
    CheckLock      func(*Faction, *FactionState, *Rulebook) (GoalLock, []Mutation)
    UpdateProgress func(*Faction, []Mutation, *FactionState, *Rulebook) []Mutation
}
var goalHandlerRegistry = map[string]GoalHandlers{
    "G-001": {UpdateProgress: progressMilitaryConquest},
    "G-004": {CheckLock: checkLockPlanetarySeizure, UpdateProgress: progressPlanetarySeizure},
    ...
}
```

### 4. Name resolution at render time is fragile

`digest.Build` resolves faction and asset names from live `factionState` at render time:

```go
func resolveFactionName(factionID string, factionState *state.FactionState) string {
    if f, ok := factionState.Factions[factionID]; ok {
        return f.Name
    }
    return factionID  // fallback: the ID
}
```

A faction destroyed in cycle 3 is absent from live state. If the GM runs `narrate 3` after that faction is gone, its events will appear under the ID ("alpha") rather than the name ("The Galactic Syndicate"). Decision 133 calls this acceptable, but it's a real usability gap when running narrative on historical cycles.

The better design: embed the faction name in the `EventRecord` at write time. The name is known when the event is recorded. Reading it back is then deterministic regardless of current state.

```go
type EventRecord struct {
    Cycle      int
    FactionID  string
    FactionName string  // captured at write time
    ...
}
```

One field addition, no backwards-incompatible change to the rest of the format.

### 5. The data distribution story is unresolved

The static game data lives at `internal/faction/data/` in the source tree. At runtime, the binary reads it from `FACTION_DATA_DIR`:

```
FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data ./bin/faction-manager ...
```

This works for development. A GM who downloads the binary needs to know data files exist separately from the binary, find or create the directory, and set the environment variable. The architecture doc says this is intentional (different campaigns can carry different rulesets), which is a legitimate use case.

But the default case — a GM using the standard SWN rules without customization — should be zero-config. The default ruleset should ship embedded via `//go:embed`, with the external `FACTION_DATA_DIR` as an override:

```go
//go:embed data/*.toml
var defaultData embed.FS
```

The data is small. The "custom campaign rules" use case is preserved via the env var. The tool becomes distributable as a single binary for the common case.

### 6. `state.Save` is called on every intermediate pipeline step

`applyAndRecord` is called multiple times during a single faction's turn — for goal lock mutations, bookkeeping mutations, and action mutations separately. Each call opens and rewrites the full TOML state file. For the data volumes involved (small), this is not a performance problem.

But the intent is blurry. The state file exists as a **resumability checkpoint** — if the process dies, the turn restarts from the last committed point. The history file benefits from being written incrementally (a partial record is useful; you can see that bookkeeping ran). The state file doesn't: writing it after the goal-lock phase and again after bookkeeping doesn't increase the quality of resumability versus writing it once after bookkeeping.

Clarifying the intent would make this cleaner: history writes every step (for auditability); state saves once per phase gate (for resumability). The current code conflates the two by tying both to every `applyAndRecord` call.

### 7. The `MutationEngine` sole-writer invariant leaks at turn start

`readyAllAssets` in `turn.go`:

```go
func readyAllAssets(factionState *state.FactionState) {
    for _, faction := range factionState.Factions {
        for _, asset := range faction.Assets {
            asset.Ready = true
        }
    }
}
```

This writes directly to `FactionState` without going through `MutationEngine`. Decision 66 acknowledges this and justifies it correctly — there's no history replay feature that would need ready-flag events recorded. That reasoning holds.

The problem is that "MutationEngine is the sole writer" is stated as a design invariant in the architecture doc, but it isn't one in practice. New contributors working from the doc will assume it's an enforced rule; they'll find the exception by accident. The documentation should state plainly that turn-start housekeeping (ready flags, phase reset) bypasses MutationEngine by design — MutationEngine covers game-event state changes, not internal plumbing.

---

## Summary Table

| Area | Verdict | Note |
|------|---------|------|
| `cmd/` vs. `internal/` separation | Right | Best structural decision in the codebase |
| Mutation pattern for history | Right | Genuine payoff in narrative layer |
| Hooks system (5 categories, scoped registry) | Right | All 5 categories have multiple real consumers in the full tag/asset set |
| Turn resumability design | Right | Correct granularity for the use case |
| Test harness + ScriptedCollector | Right | Good integration test infrastructure |
| TOML state / JSONL history | Right | Correct format choices |
| `Action` interface (4-method split) | Questionable | `Inputs`+`Resolve` temporal coupling adds fragility with no present-day benefit |
| `UseAssetAbility` type assertion | Wrong | Latent panic; reveals interface hierarchy inconsistency |
| Goal engine (direct switch) | Inconsistent | Same dispatch pattern as actions; should use same registry approach |
| Name resolution at render time | Fragile | Destroyed faction names break; should embed at write time |
| Data distribution | Unresolved | Binary should embed default data; external override for custom campaigns |
| `state.Save` frequency | Minor | Intent is blurry; history and state checkpointing serve different purposes |
| `readyAllAssets` bypassing MutationEngine | Acknowledged | Fine in practice; architecture doc overstates the invariant's actual scope |

---

## The Core Question, Answered

For what this tool actually does — a local, single-user, rule-enforcing bookkeeping assistant for a GM running SWN faction turns — the foundational choices are correct. The `cmd/internal` split, the mutation/history pattern, the `InputCollector`/`TurnObserver` interfaces, the hooks system, and the turn resumability design are all well-matched to the problem.

The main remaining issues are structural inconsistencies rather than fundamental misjudgments: the action interface's temporal coupling, the goal engine using a different dispatch pattern than actions, the type assertion in `UseAssetAbility`, and name resolution at render time. None of these block the tool from working, but each is a place where a new contributor can be surprised or where a future change can introduce a quiet bug.

The one genuinely unresolved question is distribution: the tool isn't yet shippable as a standalone binary, and the path to making it so (embedding the default data) is straightforward.
