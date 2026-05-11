# Decisions Log

A record of key decisions made during development, grouped by feature branch.

## Index

| Section | Decisions | Topic |
|---------|-----------|-------|
| [Scaffolding & Foundation](#scaffolding--foundation) | 1–18 | Language, architecture, storage, domain model |
| [feature/faction-create-wizard](#featurefaction-create-wizard) | 19–23 | App struct, command packages, wizard, env config |
| [chore/internal-code-review](#choreinternal-code-review) | 24–28 | FactionScale, HP table, TOML tags, asset loader |
| [feature/faction-list](#featurefaction-list) | 29–30 | List display, binary/data layout |
| [feature/turn-engine-scaffolding](#featureturn-engine-scaffolding) | 31–39 | Engine structure, bookkeeping, turn phases |
| [feature/turn-command](#featureturn-command) | 40–45 | Mutation interface, state writes, ANSI styling |
| [feature/action-engine](#featureaction-engine) | 46–53 | Action interface, factories, collector, registry |
| [feature/history-engine](#featurehistory-engine) | 54–58 | MutationRecord, per-faction granularity, history writes |
| [chore/code-review-2](#chorecode-review-2) | 59–63 | Paths package, FactionStat, mutation naming |
| [simple-actions](#simple-actions) | 64–70 | Factions map, new mutations, asset IDs, InputCollector |
| [attack-action](#attack-action) | 71–76 | Base type, Roller interface, redirect mutations |
| [feature/tui](#featuretui) | 77–83 | Bubbletea, state machine, TUICollector, goroutine bridge |
| [feature/tui-qol](#featuretui-qol) | 84–89 | Post-hoc narrative, log scope, cycle summary |
| [feature/expand-influence](#featureexpand-influence) | 90–95 | baseAttack sub-mechanic, goroutine bridge extension |
| [feature/use-asset-ability](#featureuse-asset-ability) | 96–109 | AbilityEngine, step handlers, world list, stealths |
| [feature/goal-engine](#featuregoal-engine) | 110–123 | ActiveGoal, GoalLock, XP timing, CheckLock |
| [feature/narrative-renderer](#featurenarrative-renderer) | 124–134 | Digest layer, Renderer interface, wire-service v1 |
| [feature/core-engine-orchestrator](#featurecore-engine-orchestrator) | 135–148 | Orchestrator, InputCollector, sub-engine packages |
| [feature/testing-suite-phase1](#featuretesting-suite-phase1) | 149–151 | gomock, package renames |
| [feature/testing-suite-phases2-3](#featuretesting-suite-phases2-3) | 152–157 | Integration test structure, harness |
| [feature/event-hooks](#featureevent-hooks) | 158–188 | Hook registry, dispatch, tags, stat raises |
| [chore/test-infra-split](#choretest-infra-split) | 189–190 | testharness promotion, scenarios split |
| [refactor/faction-assets-map](#refactorfaction-assets-map) | 191–194 | Assets map conversion |
| [refactor/persistence-write-cadence](#refactorpersistence-write-cadence) | 195–196 | History vs. state write separation |
| [feat/spatial-effort-2](#featspatial-effort-2) | 197–210 | Wiring the spatial layer into the faction engine: config ownership, world sub-engine, index-based scan replacements, TL/P-flag enforcement |

<br />

### Scaffolding & Foundation

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Go as the implementation language | Familiar to primary developer; compiles to a single binary; strong stdlib |
| 2 | CLI-first architecture | Keeps UI and business logic cleanly separated; frontend can be added later |
| 3 | TOML for static data, JSON lines for history | TOML is human-readable and hand-editable; JSON lines is easy to append programmatically and render into narrative |
| 4 | No database | Data volume is small; access patterns are simple; a database would be overkill |
| 5 | Actions as first-class logic, not data | Actions have complex conditional logic that can't live in TOML; metadata could be data but resolution belongs in code |
| 6 | Review Mode and Turn Mode as separate flows | Clean separation of read-only browsing from stateful turn execution |
| 7 | Turn pause/resume support | A turn is always completed in one sitting but can be saved mid-execution and resumed later |
| 8 | Domain types in `internal/faction/domain` | Separates pure data types from logic; clean import path |
| 9 | `Rulebook` as single static data object, no `DataLoader` interface | Interface was premature abstraction; loader is internal plumbing, not a public contract |
| 10 | `Rulebook` passed explicitly, not as a global | Idiomatic Go; avoids hidden dependencies; easier to test |
| 11 | Dice notation parsed at load time, stored as `DiceRoll` struct | Keeps TOML human-friendly; structured data makes AI and resolution engine cleaner |
| 12 | Assets split by category into `*_assets.toml` files, merged at load time | Easier to hand-edit; loader globs automatically so new files need no code changes |
| 13 | Special-effect-only attacks (no dice damage) have attack section omitted | Engine handles these by asset ID; description captures the mechanic; avoids inventing a parallel data structure before the engine is designed |
| 14 | `Asset.Definition *AssetDefinition` tagged `toml:"-"` | Runtime link to static data; excluded from serialization; resolved at engine startup |
| 15 | Campaign scoped by `CampaignID`; state path is `campaigns/<id>/faction_state.toml` | Keeps multiple campaigns isolated; hooks up cleanly when a campaign manager is built later |
| 16 | Conventional commits + semantic versioning | Consistent history; clear versioning baseline at v0.1.0 |
| 17 | Dev journal updated on every branch merge | Keeps design decisions and progress in sync with the codebase |
| 18 | Assets no longer store AssetDefinitions, only DefinitionID | Avoids circular references and serialization issues; when we need the definition, we can look it up from the Rulebook using the ID; simplifies the data model |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #2 | TUI or web app as primary interface | Couples UI to domain logic; CLI-first keeps them separable |
| #3 | SQLite for state storage | Data is hand-editable TOML; a database adds complexity without benefit at this data volume |
| #4 | Single monolithic state file for all campaigns | Campaign scoping isolates data; prevents one broken campaign from affecting others |
| #9 | `DataLoader` interface for the Rulebook | Interface was premature abstraction; loader is internal plumbing, not a public contract |

<br />

### feature/faction-create-wizard

| # | Decision | Rationale |
|---|----------|-----------|
| 19 | `App` struct owns the engine; commands wired in `Execute()` | Explicit dependency injection; no globals; engine initialized once before command tree is built |
| 20 | Command packages split into `faction/`, `review/`, `turn/` subdirectories | Readability and scalability; each mode has its own package with clear boundaries |
| 21 | Wizard sub-steps live in `faction/wizard/` package; receive only the data they need | Keeps create wizard orchestration clean; decouples wizard steps from the full Rulebook |
| 22 | Domain helpers (`CalcMaxHP`, `RatingsFromScale`, `AssetCountsFromScale`) moved to `internal/faction/domain` | Game logic belongs in the domain layer, not the command layer |
| 23 | `FACTION_DATA_DIR` environment variable controls data path at runtime | Dev/distribution separation; data files stay in source tree during development; binary resolves path at runtime via env var, falls back to path relative to executable |

<br />

### chore/internal-code-review

| # | Decision | Rationale |
|---|----------|-----------|
| 24 | `FactionScale` exported; `ScaleFromString` removed | `huh.NewSelect[domain.FactionScale]()` uses constants directly — no string conversion needed; eliminates a class of invalid-value bugs |
| 25 | `HPValueForRating` converted from map to switch | Avoids a heap allocation on every call; switch is clearer and idiomatic for a fixed value table |
| 26 | TOML tags added to `Faction`, `Asset`, `Tag`, `Goal` | Documents the serialization contract explicitly; snake_case keys are conventional TOML; tags are load-bearing for `FactionState` round-trips |
| 27 | Duplicate asset ID detection added to loader | Silent overwrites when merging `*_assets.toml` files would lose data with no error; loader now returns an error on collision |
| 28 | Commands currently own state mutation and path resolution | Noted as design debt: `newCreateCmd` appends directly to state and computes the state path — both should move into the engine as it grows |

<br />

### feature/faction-list

| # | Decision | Rationale |
|---|----------|-----------|
| 29 | `faction list` shows one summary line per faction (name, scale, HP, goal) | Quick orientation for the GM; richer per-faction detail belongs in Review Mode |
| 30 | Binary run from `cmd/faction-manager/`; `campaigns/` is a sibling of `bin/` | Keeps data out of the binary directory; clean separation between executable and campaign files |

<br />

### feature/turn-engine-scaffolding

| # | Decision | Rationale |
|---|----------|-----------|
| 31 | `Engine` restructured as core orchestrator composing sub-engines | Maps directly to the discovery doc design; keeps each sub-engine focused and testable; dependencies stay explicit |
| 32 | `TurnEngine` is a plain struct with no back-reference to `Engine` | No dependency needed now; if Rulebook access is required later, pass `*loader.Rulebook` directly rather than the whole engine |
| 33 | Faction order is a rotation, not a shuffle | Rules specify "roll a die no smaller than the number of factions; proceed in order" — a starting index rotation satisfies this without inventing a shuffle |
| 34 | `ApplyBookkeeping` is idempotent via `BookkeepingApplied` flag | Ensures income and maintenance are never double-applied if a paused turn is resumed after bookkeeping was already run |
| 35 | `Advance` returning `true` is the seam for Mutation and History engines | Turn completion is a single, clean signal point; future engines plug in here without touching `TurnEngine` |
| 36 | `maintenanceCost` returns 0 until `AssetDefinition` carries structured cost data | Maintenance costs exist in asset description text only; deferred until the field is modelled and resolved via Rulebook |
| 37 | Mutation and History writes commit together at the end of each faction's turn | Keeps state and history always in sync; pause/resume is correct because each faction's commit lands before the next faction's bookkeeping runs — deferring to Cycle end would require re-playing all prior factions on resume |
| 38 | `Asset.Maintained` doubles as the consecutive-miss tracker for the two-turn loss rule | First missed payment sets `Maintained = false`; second consecutive miss destroys the asset — no extra field needed |
| 39 | `TurnPhase` is an int/iota enum replacing `BookkeepingApplied bool` | Three phases needed (Bookkeeping, Action, Complete) for correct pause/resume across action resolution; a bool cannot represent the Action phase |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #31 | All methods on a single `Engine` struct | Would work today but makes sub-engine dependencies implicit as the codebase grows |
| #33 | Full shuffle instead of rotation | Rules specify a fixed list order with a random start, not random ordering each turn |
| #37 | Defer all writes to end of full Cycle | If interrupted mid-Cycle, all prior factions would need to re-act on resume; per-faction commits are the correct granularity given pause/resume is first-class |

<br />

### feature/turn-command

| # | Decision | Rationale |
|---|----------|-----------|
| 40 | `Mutation` interface defined in `domain` with `Type()` and `Describe()` | `Type()` is the stable history serialization discriminator; `Describe()` is the narrative renderer contract; domain owns the contract, engine owns the logic |
| 41 | `MutationEngine` is the sole writer to campaign state; `BookkeepingResult.RecordedMutations` bridges calculation to history recording | Centralizes all state writes; bookkeeping logic stays pure and testable; mutations are applied inside `ApplyBookkeeping` and returned under `RecordedMutations` for history recording only |
| 42 | `applyMaintenance` receives a pointer-to-slice and a `startCoin` (simulated post-income balance) | Income is prepended by the caller before maintenance runs so the running balance is correct; pointer-to-slice avoids a messy return tuple |
| 43 | `godotenv` + `config.go` for environment variable management | Dev path lives in `.env`, never hardcoded; `LoadConfig()` is the single point of failure with a helpful error; follows the project's existing env-var pattern |
| 44 | ANSI escape codes for turn wizard terminal styling; `lipgloss` deferred to TUI layer | `lipgloss` is a layout library for Bubbletea TUI components, not sequential CLI output; ANSI codes are sufficient and add no dependency |
| 45 | Pre-merge senior engineer code review added as a standard practice | Catches stale comments, misleading docs, missing tests, and alignment issues before they land on `main`; cost is low, benefit is high |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #41 | Apply mutations inline inside `applyMaintenance` | Would scatter state writes across bookkeeping logic; breaks the "MutationEngine is sole writer" invariant |
| #44 | Use `lipgloss` for terminal styling | Designed for TUI layout components (panels, borders, grids); wrong abstraction for sequential text output |

<br />

### feature/action-engine

| # | Decision | Rationale |
|---|----------|-----------|
| 46 | `Action` interface lives in `engine`, not `domain` | Depends on `loader.Rulebook` and `state.FactionState`; domain cannot import either |
| 47 | `Action` interface has four methods: Validate, Inputs, Resolve, Output | Matches discovery doc design; Validate gates selection, Inputs collects data, Resolve executes logic, Output produces mutations — clean separation of concerns |
| 48 | Actions are stateful structs; Inputs stores collected data, Resolve reads it | Keeps the interface uniform across all actions; no generic input bag or type assertions needed |
| 49 | `ActionEngine` uses factories (`func() Action`) rather than registered instances | Ensures a fresh zero-value struct per faction turn; prevents stale state from a previous faction's action phase carrying over |
| 50 | `InputCollector` interface injected into actions via factory; `GMCollector` lives in `cmd/forms` | Keeps `huh` out of the engine; AI agent plugs in by implementing the same interface with goal-driven logic; Inputs method itself never changes between GM and AI modes |
| 51 | Action selection `huh` prompt lives in the wizard, not the `ActionEngine` | `ActionEngine` is pure orchestration; UI concerns belong in the command layer — same principle as bookkeeping display living in the wizard |
| 52 | `No Action` hardcoded as a wizard-level option, not a registered action | It is a GM UI affordance, not a game mechanic; keeping it out of the action registry avoids polluting AI action selection |
| 53 | Turn command receives `*engine.Engine` rather than individual sub-engines | Turn wizard needs TurnEngine, ActionEngine, MutationEngine, and Rulebook — passing individual engines grew to the point where passing the full orchestrator is the cleaner call |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #49 | Register action instances directly | Single instance shared across all factions; leftover state from a prior faction's Inputs would persist if resolution failed mid-way |
| #50 | `huh` calls directly inside action `Inputs` methods | Couples engine to a UI library; AI agent would require a different concrete action type rather than a different collector |
| #53 | Pass individual sub-engines to turn command | Started with TurnEngine only; grew to TurnEngine + ActionEngine + MutationEngine + Rulebook — at that point the full engine is the right boundary |

<br />

### feature/history-engine

| # | Decision | Rationale |
|---|----------|-----------|
| 54 | `Describe()` removed from `Mutation` interface | Description is a renderer concern; prose on the mutation type locks the renderer to pre-baked strings and mixes display logic into the domain |
| 55 | `MutationRecord` stores a `json.RawMessage` payload alongside the type discriminator | Preserves full structured mutation data for querying; avoids custom marshalers; sidesteps interface serialisation issues |
| 56 | History records at per-faction granularity, not per-Cycle | Consistent with per-faction state commits; a per-Cycle record would require buffering history until Cycle end while state already commits per-faction, creating a sync gap on interrupted Cycles |
| 57 | `HistoryEngine` is a separate engine, not folded into `MutationEngine.Apply` | Single responsibility; history engine can grow independently without touching the mutation path |
| 58 | Index-based `huh.Select` for action selection | Interface equality is unreliable in huh's option matching; integer indices are unambiguous |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #56 | Per-Cycle event record (original discovery doc design) | State commits per-faction for pause/resume correctness; deferring history to Cycle end would create a sync gap on interrupted Cycles |
| #58 | `engine.Action` directly as huh option value | huh's option matching behaved unexpectedly with interface values; integer indices are unambiguous |

<br />

### chore/code-review-2

| # | Decision | Rationale |
|---|----------|-----------|
| 59 | `cmd/faction-manager/paths` package with `paths.New(campaignID)` as canonical campaign path resolver | Three command files each had an inline `filepath.Join` for the same paths; single source of truth eliminates the inline history path derivation in the wizard and makes future path changes a one-line edit |
| 60 | `domain.FactionStat` typed throughout faction creation wizard | Removes untyped string literals and comparisons from `create.go` and `select_assets.go`; compiler enforces valid values; consistent with how `loader` converts TOML strings to typed constants at the boundary |
| 61 | `BookkeepingResult.Mutations` renamed to `RecordedMutations` | Mutations are already applied inside `ApplyBookkeeping`; the old name implied the caller should apply them, creating a double-apply risk; `RecordedMutations` makes the recording-only purpose explicit |
| 62 | `FactionStat` moved from `asset.go` to `faction.go` | Used across faction creation, loader, and asset definitions — it is a domain-wide type, not an asset-specific concern |
| 63 | Concrete actions moved to `engine/actions` sub-package | `ActionEngine` holds only the `Action` interface and factory registry and never references concrete types; `engine/actions` imports `engine` for the interface contract with no circular import; all future actions have a clear, consistent home before the list grows |

<br />

### simple-actions

| # | Decision | Rationale |
|---|----------|-----------|
| 64 | `FactionState.Factions` refactored from `[]*Faction` to `map[string]*Faction` | Actions and mutations reference factions by ID; O(1) lookup replaces O(n) scan; TOML shape changes from `[[factions]]` to `[factions.<id>]` (accepted as pre-1.0 break) |
| 65 | New mutation types: `AssetAdded`, `FactionHPDelta`, `AssetHPDelta` | Required to express the new actions' state changes without bypassing `MutationEngine`; HP deltas carry a signed delta rather than an absolute new value so repair and damage share a type |
| 66 | Turn-start re-ready flips `Ready` directly, not via a mutation | Deterministic turn-start housekeeping; fully derivable from the cycle counter; no history replay feature exists that would need the event recorded; chose YAGNI over the "MutationEngine is sole writer" invariant |
| 67 | Asset ID suffix is per-`(faction, definition)`, monotonic among live assets | Prefix scan picks `max(N) + 1`; per-def scope because `defID` is already in the ID; no persisted counter across sells because JSONL history is self-contained and nothing in the current system is confused by a reused ID after a sell |
| 68 | `InputCollector` interface stays wide (one method per action) | Seven methods projected once all actions land; narrowing (per-action ISP, primitives, request/response) will be reassessed when pain is concrete |
| 69 | Refit replacement filter: same category, different definition, attribute ≥ `MinRating`, faction `Coin ≥ max(0, newCost - oldCost)` | Matches BuyAsset's pre-filter approach; avoids dead-end menu options; old-asset list is also pre-filtered to assets with ≥1 valid replacement |
| 70 | Refitted asset starts at full HP per new definition | Rules say "replace with new asset" — inheriting old HP would be a special case with no rule support |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #66 | Add `AssetReadyFlag` mutation and record turn-start re-readies to history | Pure housekeeping — adds noise to `history.jsonl` for events fully determined by the cycle counter |
| #67 | `Faction.NextAssetCounter map[defID]int` persisted counter for monotonic IDs across sells | Would solve ID reuse after sell+buy, but no present-day consequence: state holds only live assets, history is append-only self-contained events, no cross-references; speculative fix for a problem that doesn't exist yet |

<br />

### attack-action

| # | Decision | Rationale |
|---|----------|-----------|
| 71 | `Base` domain type added; `Faction.Bases` replaces a bare location string | Bases carry HP and location and are the redirect target for overflow damage; a bare string cannot represent that; first-class type keeps the domain model consistent with the game rules |
| 72 | Homeworld Base seeded at faction create time | Homeworld is always a Base of Influence per SWN rules; seeding at create avoids special-casing it throughout the attack and redirect paths |
| 73 | `Roller` interface + `RandRoller` production impl; `DiceRoll.Roll(roller)` takes a `Roller` | Injects the random source; deterministic fake roller enables table-driven unit tests without global state or random seeds |
| 74 | Attack resolution uses up-front attacker commit then per-matchup defender re-check | Matches SWN rules — all attackers are declared before resolution; re-check each matchup in sequence because a prior matchup may have destroyed the defender's asset before the current one resolves |
| 75 | Redirect to homeworld emits `FactionHPDelta`; redirect to non-homeworld Base emits `BaseHPDelta` | SWN rules treat homeworld Base destruction as direct faction HP loss; non-homeworld Bases are independent entities — different mutation types preserve this distinction in history |
| 76 | `ConfirmRedirectToBase` lives on the `InputCollector` interface | Redirect is a game-mechanic prompt that occurs mid-resolution; keeping it in the interface ensures a future AI agent handles it through the same contract without changing the action engine |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #73 | `math/rand` global with a fixed seed | Global state makes parallel tests unreliable; a fixed seed hard-codes test assumptions into production code |
| #76 | Redirect decision hard-coded as "always redirect" or resolved outside `InputCollector` | Bypasses the source-agnostic contract; would require a parallel mechanism when AI decision-making lands |

<br />

### feature/tui

| # | Decision | Rationale |
|---|----------|-----------|
| 77 | Bubbletea TUI replaces the `huh` wizard; `wizard.go` and `GMCollector` deleted | TUI is the exclusive Turn Mode interface; retaining the wizard as a dead code path would create two diverging UIs; deletion is cleaner than deprecation |
| 78 | Nine-state state machine owns all TUI transitions | Each state owns its own rendering and key handling; explicit states prevent ad-hoc flag proliferation; transitions are the only place side effects (history writes, snapshot capture) fire |
| 79 | `TUICollector` bridges pre-collected TUI inputs to `InputCollector` | Engine never sees the TUI; collector methods return pre-filled values; consistent with the source-agnostic design — the engine cannot distinguish `TUICollector` from a future `AICollector` |
| 80 | Goroutine/channel bridge for `ConfirmRedirectToBase` | Redirect must happen post-roll (after the attack hits), not pre-collected; resolution runs in a goroutine, sends `AttackRedirectMsg` to the BubbleTea event loop, and blocks on a `responseCh chan bool` until the user answers — the standard BubbleTea pattern for mid-computation interactivity |
| 81 | Nil collectors in action factory registrations | Registered factories are called only for `Validate` (action selection menu); `Validate` never calls the collector; nil avoids a now-deleted `GMCollector` reference while the collector-at-Run-time injection refactor is deferred |
| 82 | `lipgloss.NewStyle().Width(n)` for cycle summary column padding | `fmt.Sprintf("%-12s", ...)` measures bytes, not visible characters; ANSI escape codes from lipgloss color styles inflate byte length and break alignment; `lipgloss.Width` is ANSI-aware |
| 83 | Per-faction HP/Coin snapshot taken at the skip/bookkeeping state transition | Snapshot captures the faction's state before any mutations apply; correct start value for the cycle summary delta without replaying history |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #80 | Pre-collect redirect answer before running attack resolution | Redirect answer depends on the attack roll — the GM can only decide whether to redirect after seeing that the attack hit; pre-collection is mechanically incorrect |
| #81 | Inject a real collector at factory registration time | No concrete `InputCollector` to inject after `GMCollector` was deleted; deferred to a future refactor where the collector is injected at action Run time rather than construction time |

<br />

### feature/tui-qol

| # | Decision | Rationale |
|---|----------|-----------|
| 84 | Post-hoc narrative from mutations + pre-mutation faction state, no engine interface changes | Narrative strings (names, amounts) are derivable from mutations + the pre-Apply faction snapshot; avoids adding `Log(string)` to `InputCollector`, which would ripple across all implementations including future AI collectors |
| 85 | `narrateAction` and `narrateAttack` called before `Mutation.Apply` | Asset lookups (name from ID via the faction's Assets slice) must happen before assets are removed; calling before Apply ensures the pre-mutation state is still intact |
| 86 | `attackCollector *TUICollector` stored on `TurnModel` across the goroutine boundary | Attacker/defender pairs needed for per-matchup narrative are held by the `TUICollector` created in `startAttackResolution`; storing it on the model bridges the goroutine lifetime to `handleAttackCompleted` where narration fires |
| 87 | Play-by-play log resets per faction, not per cycle | The log is scoped to one faction's action; a cumulative cycle log would grow into an unreadable wall of text and conflate different factions' events |
| 88 | `logSummary` takes the first log line as the cycle summary Result column | First line captures the most significant event (action outcome); multi-line result strings would break the table layout; last column is unconstrained so long first lines are not truncated |
| 89 | Base-redirect attribution in `narrateAttack` is global (faction-level), not per-matchup | True per-matchup attribution would require indexing mutations against attacker order, which is not encoded in the mutation list; the global heuristic is correct for all common cases (one base per defending faction per attack) |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #84 | Add `Log(message string)` to `InputCollector`; narrate from inside `Resolve()` | Would require all three existing `InputCollector` implementations (TUI, test mock, future AI) to implement `Log`; narrative is a renderer concern, not a resolution concern — keeping it in the TUI layer is the right separation |
| #87 | Accumulate log across full cycle | Grows into a wall of text; individual faction turns are the natural scope for a play-by-play; the cycle summary table already covers the full cycle at a glance |

<br />

### feature/expand-influence

| # | Decision | Rationale |
|---|----------|-----------|
| 90 | `baseAttack` is an unexported sub-struct within `expand_influence.go`, not a registered action | Free rival attacks are a sub-mechanic of Expand Influence, not an independent GM choice; registering it would surface it in the action selection menu and couple the two actions inappropriately |
| 91 | `baseHPTracker *int` threaded by pointer across rival attacks | Mutations are accumulated but not yet applied during `resolveNewBase`; the tracker allows each successive rival — and each attacker within a rival's sequence — to see the correct effective base HP before any mutations are applied to state |
| 92 | Goroutine/channel bridge extended with `ConfirmRivalFreeAttack` (bool) and `SelectBaseAttackers` (`[]*domain.Asset`) | Both prompts occur mid-`Resolve` after rolls that cannot be pre-collected; same pattern as `ConfirmRedirectToBase` in Attack — resolution goroutine sends a typed message to the BubbleTea event loop and blocks on a response channel until the GM answers |
| 93 | `ReinforceMax` for the increase-max-HP sub-mode (rejected `ReinforceExpand`, `ReinforceGrow`) | `ReinforceExpand` clashed semantically with the outer `ExpandMode` type; `ReinforceGrow` was ambiguous; `ReinforceMax` is explicit about what the sub-mode changes |
| 94 | Rivals for the contested roll sorted by name for deterministic order | `FactionState.Factions` is a map; iterating in random order would make the sequence of mid-resolution prompts unpredictable across runs |
| 95 | Input model helpers duplicated from `engine/actions` into the `inputs` package | Cross-package coupling for six small filter functions would require a new shared package with no other consumers; YAGNI — duplication is acknowledged with an inline comment and the functions remain independent |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| #90 | Register `baseAttack` as a standalone action or reuse the existing `Attack` action | Standalone registration would expose it in the action menu; reusing `Attack` would require the normal attack flow to carry Expand Influence context — both couple unrelated mechanics |
| #92 | Pre-collect `SelectBaseAttackers` before starting resolution | Which rivals win the contested roll (and therefore which eligible attacker lists are needed) is unknown until the dice are rolled in `Resolve`; pre-collection is mechanically impossible |

<br />

### feature/use-asset-ability

| # | Decision | Rationale |
|---|----------|-----------|
| 96 | Movement does not enforce hex distance — GM confirms, engine updates `Asset.Location` only | Distances are GM-world-specific and would require a per-campaign world graph that doesn't exist; GM adjudication is the correct boundary for spatial constraints |
| 97 | Assets with `Ability == nil` fall through to GM adjudication: `ConfirmAbilityApplied` shows the asset description and blocks for GM confirmation | Several abilities (Pretech Logistics, Tripwire Cells, Seditionists) are too bespoke to encode as TOML step data; the fallback keeps them playable before they receive custom handlers |
| 98 | `AbilityEngine` has a step handler registry keyed by `AbilityStepType` plus a custom handler override map keyed by asset definition ID | Two-layer dispatch — most abilities compose standard steps; bespoke abilities bypass the registry entirely via the custom handler map; the two paths are independent and neither intrudes on the other |
| 99 | World list for `SelectMoveDestination` is derived from `FactionState` at resolve time; `"Astral Sea"` appended last as a hardcoded option | No static world list exists; all known worlds are implied by where assets and bases currently are; Astral Sea is a rules-canonical transit space with no permanent placements and must always be available |
| 100 | Movement Coin cost is emitted as a `CoinDelta` mutation (negative) when `step.CoinCost > 0` | Coin cost is part of using the ability, not a separate action; encoding it as a mutation keeps the mutation log complete and consistent with every other Coin change in the system |
| 101 | Faction test uses the existing `Roller` interface — same dice-rolling contract as Attack | Same interface; same deterministic-roller test pattern; no new abstraction needed when the existing one fits exactly |
| 102 | `reveal_stealth` emits one `AssetStealthCleared` per stealthed asset on the acting asset's world belonging to the target faction | SWN rules say all stealthy assets on the world are revealed; one mutation per asset is the correct granularity for history replay and for the play-by-play narrator |
| 103 | `coin_steal` emits paired `CoinDelta` mutations: negative on target, positive on acting faction | Drain destroys Coin; steal transfers it — paired mutations correctly model the transfer without needing a new transfer mutation type |
| 104 | Faction test tie goes to the defender — ability effect does not apply | Same tie-goes-to-defender rule as Attack; consistent with the established precedent |
| 105 | `UseAssetAbility` selects all assets up-front (ordered list), then resolves each in sequence — same committed-upfront pattern as Attack | SWN rules require declaring all intended assets before resolution; committed-upfront enforces this and prevents re-selecting after seeing earlier results |
| 106 | `SelectAbilityAssets` returns an ordered list; order determines resolution sequence | SWN specifies "each form of asset must be fully used before the next type is triggered"; ordering is meaningful, not decorative — the input model shows `[1]`, `[2]` badges alongside selected items |
| 107 | Nil guard added after `SelectFactionTestTarget` in `factionTestStepHandler` | Nil return is a valid collector signal (cancelled or no candidates); without the guard, calling `Roll` on a nil faction would panic at runtime |
| 108 | Nil guard added before `step.EffectDice.Roll` in `applyAbilityEffect` for `coin_drain` and `coin_steal` | Effect dice are required for these effects but are not validated at load time; a missing TOML field would reach here as nil and panic |
| 109 | Faction test candidates sorted by name at end of `factionTestCandidates` | `FactionState.Factions` is a map; without sorting, the sequence of `SelectFactionTestTarget` prompts would differ between runs — same fix rationale as rivals in `feature/expand-influence` |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| 97 | Block until custom handlers are written for all bespoke assets | Blocks the entire action until every edge case is handled; GM fallback is the right interim path for abilities that have no step encoding |
| 98 | Single flat handler map keyed by step type only | Custom overrides per definition ID cannot be expressed with a single registry; bespoke abilities would require inventing a synthetic step type per asset |
| 105 | Collect assets one at a time as each resolves | SWN rules require up-front declaration; allowing per-step selection gives the GM information about prior results before committing later assets |

<br />

### feature/goal-engine

| # | Decision | Rationale |
|---|----------|-----------|
| 110 | `Faction.Goal *Goal` replaced by `Faction.ActiveGoal *ActiveGoal` | `Goal` was a reference to static data; `ActiveGoal` carries live progress state (Progress, ProcessPhase, TurnsRemaining, target fields) — the two concepts are fundamentally different and deserve separate types |
| 111 | Goal behavior lives in code as per-goal-type handlers; TOML holds metadata only (name, description, difficulty label) | Goal resolution logic has conditional branching on mutation types, faction stats, and world state — none of which can be expressed in data; metadata stays in TOML so GMs can rename goals or adjust descriptions without touching code |
| 112 | Change Homeworld is a goal, not a registered action; Difficulty 0 added to `goals.toml` | Change Homeworld spans multiple turns and requires lock state; modeling it as an action would require the action system to manage multi-turn state, which is the Goal Engine's job |
| 113 | Seize Planet (action) initiates the Planetary Seizure (goal) process; Goal Engine manages subsequent turns | The action creates the goal state; the engine steps it through combat → occupation → completion over subsequent turns; same pattern as Expand Influence (action initiates, Goal Engine observes) |
| 114 | Abandon Goal is a registered action that deducts income and clears `ActiveGoal` | Abandonment is a GM-initiated game-turn decision with mechanical cost (income forfeited); registering it as an action exposes it in the action menu at the right point in the turn flow |
| 115 | `GoalEngine.UpdateProgress` is called by the TUI after `ActionEngine.Run` and before `MutationEngine.Apply` | Goal progress inspection must happen before destructive mutations fire (Destroy the Foe XP reads live target HP); doing it in the TUI rather than inside `ActionEngine.Run` keeps the engine layer free of Goal dependencies |
| 116 | All mutations gain `CausedByFactionID string` and `Cause string` attribution fields | Goal handlers must distinguish "rival Force kill caused by acting faction" from maintenance losses or own-asset losses; attribution fields carry this provenance through the mutation list without requiring separate event types |
| 117 | `AssetStealthApplied` is emitted during Buy Asset when a Stealth-type Cunning asset is purchased | The goal handler for Inside Enemy Territory needs a discrete event to count stealths that post-date goal adoption; emitting at purchase time (not at use time) matches the SWN rule |
| 118 | `GoalEngine.CheckLock(faction, factionState) GoalLock` return value (Option B) | Option A (post-Action inspection) couldn't express Change Homeworld (no action is taken); Option B (pre-turn return value) gives the TUI a clean routing signal before any action selection runs |
| 119 | Each goal type has its own `CalculateXP` formula; XP for variable-difficulty goals calculated at completion time before destructive mutations apply | XP formulas reference live state (target faction stats for Destroy the Foe, rival presence for Expand Influence); calculating before Apply ensures the data hasn't been destroyed by that turn's mutations |
| 120 | `Base.Influence int` is a new field; Bribe is the only action that modifies it; no mechanical effect beyond Wealth of Worlds progress | Influence exists to satisfy the Wealth of Worlds goal's "Coin spent on bribes" requirement; it has no other game mechanic, so a dedicated counter field is correct — not a general-purpose currency modifier |
| 121 | Inside Enemy Territory tracks only stealths applied after the goal was adopted; stealths at adoption time do not count | SWN rule: "units already stealthed when this goal is adopted do not count"; `AssetStealthApplied` mutations only fire for new stealths, so pre-adoption stealths naturally do not contribute |
| 122 | Change Homeworld completion fires inside `CheckLock` when `TurnsRemaining` reaches 0, not inside `UpdateProgress` | There is no action on the completion turn — `CheckLock` is the only Goal Engine call that runs for a skipped faction; `UpdateProgress` requires an action's mutation list as input |
| 123 | Destroy the Foe XP calculated before the target faction is removed from state | XP formula reads target stats; if computed after `FactionHPDelta` or faction removal mutations apply, the target faction may no longer exist in state |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| 115 | Call `UpdateProgress` inside `ActionEngine.Run` | Would require `ActionEngine` to depend on `GoalEngine`; creates a circular-style coupling between two engines that should be peers; TUI orchestration is the right layer for cross-engine sequencing |
| 118 | Option A — inspect mutations post-action to determine lock | Cannot express Change Homeworld (skip the entire turn — there is no action to inspect) |

<br />

### feature/narrative-renderer

| # | Decision | Rationale |
|---|----------|-----------|
| 124 | Hybrid architecture: deterministic `digest` layer + pluggable `Renderer` interface | `digest.Build` extracts all meaningful events into a structured `CycleDigest` without formatting concerns; the `Renderer` interface then expresses those events as prose; the LLM renderer (post-v1) will operate on the same structured input without touching history parsing logic |
| 125 | v1 ships wire-service renderer only; LLM renderer is post-v1 | Building LLM integration before the rendering interface is validated would lock in design decisions prematurely; deterministic output proves the interface contract before a second implementation is written |
| 126 | Default output path `campaigns/<id>/narratives/cycle-NNN.md`; reruns increment to `-002`, `-003` (3-digit zero-padding) | Increments preserve prior runs for comparison without overwriting; 3-digit padding ensures lexical sort matches numeric order |
| 127 | Output structure: cycle headline + lede paragraph + per-faction `###` sections + quiet tail | Mirrors how a GM describes a cycle: headline captures the most important event, lede sets context, per-faction detail follows, quiet factions are acknowledged without crowding out active ones |
| 128 | Cross-faction events owned by the actor's section; defender section does not echo | Avoids duplication — narrating both sides produces the same engagement twice with different phrasing; the attacker's section gives the full outcome |
| 129 | Goal events flow as ordinary narrative beats within `FactionBeat` — no separate milestone struct | Goal completion is already captured in `GoalEvent` within `FactionBeat`; a parallel milestone type would duplicate data and add indirection with no rendering benefit |
| 130 | Headline priority: Attack > GoalCompleted > FactionDestroyed > HomeworldShift > GoalAbandoned > Quiet; tie-break by impact then alphabetical | Ordered by reader impact — violent events drive the most immediate GM/player decisions; quiet cycles should not displace any meaningful passive event |
| 131 | `AssetMaintainedFlag` mutations dropped from narration | Maintenance flags are bookkeeping artifacts emitted for every asset on every turn; including them buries meaningful events in noise |
| 132 | Faction destruction detected by comparing history actor IDs against live `factionState.Factions` | Simpler than scanning for `FactionHPDelta`-to-zero: a faction ID that appears in the cycle's history records but is absent from live state was destroyed before state was saved |
| 133 | Live state acceptable for ID→name resolution; resolvers fall back to the ID | History records store IDs, not names; live state is the source of truth for the current name; fallback to ID prevents nil panics when a faction has been destroyed and is absent from state |
| 134 | Variant rotation via seeded `math/rand` RNG; default seed is `time.Now().UnixNano()`; seed printed to stdout after each run | Deterministic output is required for golden-file tests and for GMs to share reproducible narratives; printing the seed allows any run to be reproduced exactly with `--seed` |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| 124 | Monolithic renderer that reads JSONL and produces prose directly | Would mix history parsing with rendering logic; the LLM renderer would have no structured input to work from — it would need to re-parse history itself |
| 128 | Echo the attack in both attacker and defender sections | Creates duplicate prose that reads as padding; cross-faction events are single-perspective by convention |

<br />

### feature/core-engine-orchestrator

| # | Decision | Rationale |
|---|----------|-----------|
| 135 | Acknowledgement gating lives on `InputCollector` as `AwaitCheckpoint(phase string) error` | Collector owns turn pacing; observers stay strictly fire-and-forget; manual implementations block on user input; test/AI implementations return nil immediately |
| 136 | Action selection lives on `InputCollector` as `SelectAction(faction, available) (Action, error)` | Engine computes `AvailableActions` itself; the observer is not in the decision loop; the collector interface is already the "caller decides" contract |
| 137 | `EventHook` ships as interface + documented dispatch site only — no registry, no dispatcher | The seam is needed now to lock the mutation-apply order; the dispatcher has no consumer until the Tag Engine ships; YAGNI |
| 138 | Hook recursion bounded at depth 5; on cap trip the engine logs and stops | Surfaces content bugs rather than silently absorbing runaway loops |
| 139 | Observer and collector are passed per-call to `RunFactionTurn` / `RunCycle`, not stored on `Engine` | Engine stays a long-lived stateless toolbox; the same instance can serve a manual run and a headless test without re-construction |
| 140 | Runtime config flows through a new `internal/faction/config` package holding a `Config` struct | Establishes the config pattern before future runtime knobs (dry-run, log level, AI settings) join; keeps CLI path derivation in `paths/` and engine path injection in `Config` — **superseded by #204** |
| 141 | `Turn.ApplyBookkeeping` refactored to return `(BookkeepingResult, []domain.Mutation, error)` without applying mutations | Removes the asymmetry where one sub-engine wrote state and the others didn't; makes mutation flow uniform — the orchestrator owns every `Apply + Record + Save` call |
| 142 | `ActionFactory` changes from `func() Action` to `func(InputCollector) Action`; `AvailableActions` takes a collector | Without this, `SelectAction` returning a ready `Action` is broken — prior factories produced nil-collector stubs valid only for `Validate`; the TUI's switch-on-Name reconstruction logic is eliminated |
| 143 | `Engine` gains a `Rand domain.Roller` field initialized to `engine.NewRandRoller()` in `New` | Action factories that need a roller capture `e.Rand` in the closure; single injection point enables deterministic headless tests via `eng.Rand = &fixedRoller{...}` |
| 144 | `RegisterDefaultActions(*Engine)` helper lives in `internal/faction/engine/actions` | Prior registration site (`commands/turn.go`) was deleted in Phase 1; placing the helper next to the action structs keeps engine and actions free of import cycles; test harnesses and future TUI use a single call |
| 145 | Sub-engines extracted to their own packages: `ability`, `action`, `goal`, `history`, `mutation`, `turn` | Engine package was a single flat directory mixing six sub-engines and the orchestrator; named packages enforce the sub-engine boundary at the import level and make dependencies explicit |
| 146 | `engine.NewWithRulebook(*loader.Rulebook)` added alongside `New(dataDir string)` | `engine.New` does disk I/O, blocking truly hermetic engine tests; injecting an already-loaded rulebook removes the test dependency on fixture files |
| 147 | `Checkpoint*` constants renamed from `Phase*` to `Checkpoint*` | `domain.Phase*` (turn-cursor state) and the original `engine.Phase*` (checkpoint gate names) were adjacent and easy to confuse; `Checkpoint*` unambiguously names the `AwaitCheckpoint` call sites |
| 148 | `TurnEngine` takes a `domain.Roller`; `buildFactionOrder` uses it | `buildFactionOrder` had called `rand.IntN` directly, bypassing `e.Rand`; faction ordering was non-deterministic in headless tests even when `eng.Rand` was swapped to a fixed roller |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| 137 | Ship `EventHook` with a dispatcher and `RegisterHook` now | No consumer until the Tag Engine lands; building the dispatcher first would require test coverage for a code path with no callers — pure speculative work |
| 139 | Store observer + collector on `Engine` at construction time | Engine would need to be reconstructed for every test scenario that uses different observer/collector configs; per-call injection is idiomatic Go and keeps the engine stateless |
| 141 | Leave `ApplyBookkeeping` applying mutations internally | Breaks uniform mutation flow — one sub-engine was writing state, all others returned mutations for the caller to apply; asymmetry makes the orchestrator's write path non-obvious |

<br />

### feature/testing-suite-phase1

| # | Decision | Rationale |
|---|----------|-----------|
| 149 | `go.uber.org/mock/gomock` + generated `MockCollector` in `engine/actions/mocks/` | The `InputCollector` interface has 14 methods; hand-written fake structs require every method to be declared per test file even when only 1–2 are exercised. gomock generates the mock once; each test declares only the expectations it cares about, and any unexpected call fails the test automatically |
| 150 | moved `internal/engine/actions/` into `internal/faction/engine/action/`| Actions are an implementation of the action interface, and belong within that sub-engine.
| 151 | `internal/faction/loader` package renamed to `internal/faction/rulebook` | The package name `loader` described the mechanism (loading files); `rulebook` describes what it produces — static game data. Consistent with the `Rulebook` type name already used throughout the codebase |

<br />

### feature/testing-suite-phases2-3

| # | Decision | Rationale |
|---|----------|-----------|
| 152 | `integration_test/` as a standalone Go package for the integration test suite | Build tags exclude tests from normal runs (not wanted here); a flat file in `engine/` mixes concerns; a separate package gives each file a clear job, enforces the public-API-only contract, and is discovered automatically by `go test ./...` |
| 153 | Three-file split inside `integration_test/`: `harness_test.go` / `fixtures_test.go` / `scenarios_test.go` | Infrastructure (engine wiring, history readers, mutation finders) lives separately from fixture builders and scenario-specific helpers, which in turn live separately from the test scenarios themselves; each file has a single clear job |
| 154 | `orchestrator_test.go` rewritten as `package engine` white-box unit test covering only `filterAllowedActions` | The existing three `TestRunCycle_*` tests were integration tests and moved to `integration_test/`; the only genuinely unit-testable surface of `core_orchestrator.go` is the unexported `filterAllowedActions` function — white-box access requires `package engine` |
| 155 | `FixedRoller` promoted to `testharness` package | Any test package that needs deterministic dice can import testharness; keeping it local to `orchestrator_test.go` (as `fixedRoller`) would require duplication in `integration_test/` |
| 156 | `checkStep(t, description, ok, detail)` helper for labeled assertion output | `t.Logf("  ✓ …")` / `t.Errorf("  ✗ …")` pattern surfaces a per-assertion confirmation log with `go test -v` — readable report without external tooling or test framework dependencies |
| 157 | `TestRunCycle_Bookkeeping_AssetDestroyedSecondMiss` skipped — not implemented | `maintenanceCost` returns 0 (stub — see decisions log #36 and deferred items); the two-miss destruction path cannot fire until structured cost data is added to `AssetDefinition` |

<br />

### feature/event-hooks

| # | Decision | Rationale |
|---|----------|-----------|
| 158 | `Registry` uses `map[Scope][]Registered*` (one map per category) | Preserves registration order within each scope bucket; scope-keyed map is O(1) lookup; composing the result slice (global + faction + asset) makes ordering explicit and testable |
| 159 | Lookup ordering: global first, then faction-scoped, then asset-scoped | Global hooks are framework-level and should fire before per-faction or per-asset overrides; deterministic ordering across all consumers without needing registration timestamps |
| 160 | `RollState` included in Phase 1 `types.go` alongside `ModifierOffer` | `ModifierOffer.Apply` requires a concrete `*RollState` parameter; defining it alongside the offer type avoids a forward reference or a dummy `any` signature that would require breaking changes later |
| 161 | `HookBudgets` cleared to `nil` (not zeroed per key) in `ApplyBookkeeping` | "nil" matches the "once per turn" rule wording exactly — budgets reset wholesale each turn and re-accumulate as hooks fire; zeroing individual keys would require registered hook descriptors to be consulted during bookkeeping, coupling the engine to registered hook state before wiring exists |
| 162 | `eventhooks.Collector` embedded into `engine.InputCollector`; mock regenerated from `engine.InputCollector` | `SelectModifiers` and `ConfirmReroll` are hook-dispatch concerns, not action-resolution concerns; embedding keeps the collector interface composable; regenerating the mock from the full interface means tests always use a mock that satisfies the real production type |
| 163 | `MockCollector` renamed to `MockInputCollector` in `mocks/` after regeneration | Mock type name reflects source interface name (`InputCollector`); `go:generate` directive on `core_input_collector.go` keeps the generation command co-located with the interface definition |
| 164 | `ScriptedCollector` defaults: `SelectModifiers` returns all offers; `ConfirmReroll` returns true | Integration tests exercise hooks with no GM input; accepting all offers and confirming all rerolls is the maximal-coverage default; tests that need selective behavior override via `SelectModifiersFn`/`ConfirmRerollFn` |
| 165 | `dispatchMutationReactors` passes the full accumulated `combined` slice to each depth's reactors, not only the new mutations | Reactors are responsible for idempotency; the plan spec says `combined`; stateful one-shot stubs in tests avoid double-fire without changing the dispatch contract |
| 166 | Reactor depth cap fires at `depth >= 5` (not `> 5`), bounding total rounds to 5 (depths 0–4) | Depth 5 is the cap trip; 5 rounds is the plan spec; the bound is clear and testable |
| 167 | `EventHook` interface deleted; `MutationReactor` in `eventhooks/interfaces.go` is the sole Cat 3 type | `EventHook` was always a placeholder; `MutationReactor` has the same shape and lives in the right package; removing the alias eliminates confusion |
| 168 | `ResolveTie` dispatch uses first-registration-wins unconditionally (not first-non-Standard) | First-registration-wins is a predictable contract; iterating until the first non-Standard result would silently skip Standard-returning resolvers, making the system harder to reason about; if a resolver is registered it must be authoritative |
| 169 | Registry threaded to `maintenanceCost` via `ApplyBookkeeping` parameter, not stored on `TurnEngine` | Keeps `TurnEngine` stateless with respect to hooks; registry is owned by `Engine` and passed where needed; no constructor signature change required for `TurnEngine.New` |
| 170 | nil-safe guards added to all Cat 4+5 dispatch helpers | Test call sites pass nil registry; nil guard returns base value without panicking; production path always provides a real registry via `Engine.Hooks` |
| 171 | `tag/` package uses a `handlers` map keyed by tag ID (`map[string]func(*engine.Engine, *domain.Faction)`); each tag file owns its `TagID` constant and a `Register` func | Mirrors the `action/actions/` pattern: `tag.go` is a pure lookup table (one line per tag), ID and registration logic live next to the implementation; unknown tag IDs are silently skipped with no special-case code |
| 172 | `RegisterDefaultTags` takes `*state.FactionState` in addition to `*engine.Engine` | Tag hooks are faction-scoped and must be registered per owning faction; walking state at startup is the only way to know which factions own which tags; actions need no such walk because they are unconditionally registered |
| 173 | `ScavengersReactor` matches on `AssetRemoved{Cause: "attack"}` (no `AssetDestroyed` type exists) | The domain represents combat kills as `AssetRemoved` with `Cause: "attack"`; non-combat removals (sell, refit, bookkeeping) carry different cause strings and must not trigger Scavengers |
| 174 | `harness.registerTags()` helper called per-scenario after `addFaction`, not inside `newHarness` | `newHarness` creates an empty faction state; tags are faction-scoped and must be registered after factions are populated; calling `RegisterDefaultTags` on an empty state would silently register nothing |
| 175 | Keep-highest semantics implemented via `RollState.SetKeepHighest` trim step, not via `ModifierOffer` extension or a paired `RollResultHook` | A trim on `RollState` keeps all dice-pool mutations in one place — `Apply` accumulates, `RollWithHooks` fires once; a paired `RollResultHook` would cross a category boundary (Cat 1 offer registering a Cat 2 effect), complicating the dispatch contract with no benefit |
| 176 | `action.Collector` embeds `hooks.Collector`; attack roll sites pass the action collector directly to `RollWithHooks` | `engine.InputCollector` already embeds both; embedding in `action.Collector` makes the hierarchy consistent and avoids a separate `hooks.Collector` field on every action struct; `ScriptedCollector` and `MockInputCollector` both already satisfy `hooks.Collector`, so no implementation changes were needed |
| 177 | `dispatch.ResolveTie` probes `ctx.Actor.ID` first, then `ctx.Opponent.ID` if no attacker resolver is found | A faction-scoped `TieResolver` must fire when that faction is both attacking and defending; the dispatch layer handles role detection at lookup time; the resolver then reads `ctx.Actor.ID` vs. `resolver.FactionID` to return the correct `TieOutcome` for each role |
| 178 | Preceptor Archive chosen over Pirates as Cat 4 proof-of-life tag | Preceptor exercises the `AssetCostModifier` → `ResolveAssetCost` → `buy_asset.go` path already plumbed in Phase 3c; Pirates would require extending the movement-cost path which was not migrated; cleaner integration test (single faction, buy action, no movement graph needed) |
| 179 | `max(newCost, 0)` floor in `PreceptorArchiveCostModifier.ModifyAssetCost` (not a `baseCost > 0` guard) | Both approaches produce the same result for cost-0 assets; `max` makes the non-negative invariant explicit at the return site rather than buried in a branch condition; floor is expressed once regardless of how many discount modifiers chain together |
| 180 | `*rulebook.Rulebook` stored on `TurnEngine` (unlike the hooks registry, which is passed per-call) | Rulebook is static data that never changes between turns; storing it avoids threading it through every call site; the registry is turn/session-scoped and owned by `Engine`, so passing it per-call remains correct |
| 181 | Over-cap surcharge computed via two-pass in `applyMaintenance`: count-by-category first, then apply surplus as +1 per excess asset in the main loop | Pre-counting is required because each asset's effective cost must be known before the affordability check; a single-pass approach would require look-ahead or deferred correction; the surcharge map tracks remaining charges and decrements in iteration order, which is arbitrary but total-correct |
| 182 | `SelectStatRaise` returns `*domain.FactionStat`; nil = skip | Idiomatic Go optional; avoids sentinel string values; same "nil = nothing chosen" pattern as `SelectAction` returning nil for No Action |
| 183 | Stat raise phase fires after `CheckpointBookkeeping`, before action selection | Player sees their post-bookkeeping XP balance before deciding; consistent with PDF rule "at the beginning of each turn" meaning before the action |
| 184 | No new checkpoint for stat raise | The `SelectStatRaise` call itself is the interactive pause; adding a checkpoint would double-pause the turn for a phase that only fires when the faction has enough XP |
| 185 | `StatRaised.Apply` sets rating to `NewRating` (not `+= 1`) | Idempotent under replay; the old rating is carried in the mutation for history readability |
| 186 | Eligibility computed in the engine, passed to collector as `eligible []domain.FactionStat` | Consistent with `AvailableActions` pattern; collector shows only valid options without needing to re-derive the cost table |
| 187 | Skip phase entirely (don't call collector) when `eligibleStatRaises` is empty | Avoids blocking interactive TUIs on a no-op phase; `IneligibleNoPrompt` integration test asserts this |
| 188 | `SellAsset.Inputs` returns error when `SelectAsset` returns nil | Prevents nil dereference in `Resolve`; nil from collector means user cancelled; a clean error is better than a panic |

<br />

### chore/test-infra-split

| # | Decision | Rationale |
|---|----------|-----------|
| 189 | Test helpers promoted from `integration/*_test.go` to `testharness/harness.go` (regular exported file) | `_test.go` helpers are only visible within their own package; an exported file in `testharness/` lets any future test package import `NewHarness`, `ReadHistory`, `CheckStep`, etc. without duplication |
| 190 | `integration/` renamed to `scenarios/`; `scenarios_test.go` split into `full_cycle_test.go`, `goal_lock_test.go`, `actions_test.go`, `hooks_test.go`, `tags_test.go`, `stat_raise_test.go` | Single 905-line file made it hard to find specific scenarios; grouping by mechanic (goal locks, actions, hooks, tags) matches how tests are reasoned about and extended |

**Notable alternatives rejected:**

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| 158 | Flat global slice with per-entry scope field (filter on lookup) | Preserves global registration order across scopes but makes scope-filtered lookup O(n) across the whole registry; buckets keep lookup fast with no ambiguity about ordering semantics |
| 161 | Zero out only budget keys declared by registered hooks | Requires registry to be held on engine during bookkeeping (Phase 3a wiring); cleaner to defer that dependency; nil-on-clear is safe since hooks re-register budgets from zero each turn |
| 165 | Feed only new mutations to each subsequent depth (trigger-batch approach) | Cleaner integration test design, but conflicts with the plan's `combined` spec and forces awkward trigger-type matching in reactors; reactor idempotency is the correct boundary |
| 169 | Store registry on `TurnEngine` (add to constructor) | Would work, but adds coupling to `TurnEngine` for something it doesn't own; parameter threading is more explicit and avoids changing the `turn.New` signature |
| 175 | Paired `RollResultHook` registered by the offer's `Apply` to drop the lowest die | Crosses a category boundary; also requires the Cat 2 hook to receive the pre-trim dice state, which may include the extra die that hasn't been "chosen" to keep yet — ordering semantics become ambiguous |
| 176 | Add a separate `hooks.Collector` parameter to `NewAttack` and each action factory | Produces two parallel collector fields in every action struct; the factory closure would need to cast or double-pass the same `InputCollector` — mechanical overhead with no conceptual gain |
| 177 | Register `FanaticalTieResolver` at global scope; check both factions inline | Global scope would fire for all ties, requiring the resolver to check whether either faction is Fanatical; faction-scoped registration is more precise and consistent with every other hook in the system |

<br />

### refactor/faction-assets-map

| # | Decision | Rationale |
|---|----------|-----------|
| 191 | `Faction.Assets` converted from `[]*Asset` to `map[string]*Asset` | Matches the precedent of `FactionState.Factions` (Decision 64); pre-1.0 breaking change is acceptable; O(1) lookups in mutation engine remove 5+ linear scans per mutation batch |
| 192 | Canonical sort order in `applyMaintenance` is ascending asset ID | Map iteration is unordered; maintenance charging is coin-budget-sensitive (first asset wins when coin runs out); sorting by ID makes the result deterministic and reproducible across runs |
| 193 | `action.Collector.SelectAsset` keeps `[]*domain.Asset` parameter; map→slice conversion is internal to `SellAsset.Inputs` | The collector interface should not know about the internal storage shape; `domain.SortedAssets` produces a consistent ordered slice without exposing the map |
| 194 | `ownerFaction` in `attack.go` simplified to `factionState.Factions[asset.OwnerID]` | `Asset.OwnerID` already encodes the owner; the previous O(n × factions) scan was a holdover from when the field may not have been trusted; direct lookup is correct and faster |

### refactor/persistence-write-cadence

| # | Decision | Rationale |
|---|----------|-----------|
| 195 | `applyAndRecord` no longer calls `state.Save`; callers are responsible for saving at phase-gate checkpoints | History (JSONL) must be written on every mutation batch — it is an audit log. State (TOML) is a resume checkpoint — it only needs to reflect the boundary between meaningful phases. Merging them in one function obscured the distinction and produced 3–4 redundant TOML rewrites per faction turn |
| 196 | `state.Save` is called explicitly at two points: after `AwaitCheckpoint(CheckpointBookkeeping)` and after `AwaitCheckpoint(CheckpointActionResult)` | These are the only two boundaries where the turn's progress is meaningful for resume. Goal-lock tick mutations and stat-raise mutations between these gates are captured by the next save; `TurnState.Phase` encodes enough phase position to make earlier saves redundant |

### feat/spatial-effort-2

| # | Decision | Rationale |
|---|----------|-----------|
| 197 | `internal/faction/config.Config` is the canonical owner of all runtime paths (`FactionDataDir`, `SpatialDataDir`, `StatePath`, `HistoryPath`, `NarrativesPath`); `cmd/faction-manager/paths` package retired | The tool is headless — no layered CLI state, no campaign ID derivation needed. All paths flow in as env vars, live in the internal config, and are threaded straight through. Two config structs (CLI + internal) for the same data was arbitrary confusion. Supersedes #140 and #59 |
| 198 | `engine.New()` takes `*config.Config` instead of bare `dataDir string`; `LoadConfig()` returns `*config.Config` directly | `FactionDataDir` is a construction-time path like `StatePath` — no reason for it to travel as a raw string while the others live in a struct. Single entry point, single type, no implicit conventions |
| 199 | All five paths are required env vars (`FACTION_DATA_DIR`, `SPATIAL_DATA_DIR`, `FACTION_STATE_PATH`, `FACTION_HISTORY_PATH`, `FACTION_NARRATIVES_PATH`); `--campaign` flag retired | Campaign identity is now fully encoded in the paths themselves — the GM sets them per session. Removing the derived-path pattern eliminates the indirection and makes the tool easier to wire into scripts and automation |
| 200 | `drift_costs.toml` is its own file for now; if more small or ungroupable static data accumulates, consolidate into a single `constants.toml` | One file per concern is fine at small scale; proliferating single-value TOML files is noise — a shared constants file is the consolidation point if the pattern repeats |
| 201 | `world` sub-engine package (`internal/faction/engine/world`) wraps `HybridMap` + per-turn `Index`; `Engine.World` field replaces raw `spatialMap`/`spatialIndex` fields | Follows the established sub-engine pattern; consolidates index building, distance queries, and fragment lookups under one accessor; makes the spatial concern visible at the `Engine` boundary in the same way as `Turn`, `Action`, and `Goal` |
| 202 | Sub-engine named `world`, not `spatial` | `internal/faction/engine/spatial` would require an import alias inside its own package to import `internal/spatial` (same package name collision); `world` maps more directly to the game domain and avoids the alias entirely |
| 203 | `world.NewWithMap(spatial.SpatialMap)` alongside `world.New(dataDir)` — holds the interface, not `*HybridMap` | Mirrors `engine.NewWithRulebook`; any `SpatialMap` implementation (hex, graph, stub) can be injected in tests without touching disk |
| 204 | O(n) fallback paths removed from all index-based scan functions; `*world.Index` is always required — nil panics rather than falls back | Keeping both paths invites the slow path to silently remain exercised; a panic surfaces misconfiguration immediately and is the correct signal that the harness needs a spatial map |
| 205 | `indexFromState` test helper in `action/actions/test_helpers_test.go` builds `*world.Index` directly from `*state.FactionState` | Action-layer unit tests own their spatial setup; building the index from the same state the test constructs is minimal, exact, and requires no disk or stub SpatialMap at this layer |
| 206 | `world.Index` fields and all `fragmentID` parameters renamed to `AssetsByLocation`, `BasesByLocation`, `locationID` throughout the faction engine; `worlds.toml` replaces `fragments.toml` in the spatial package | "Fragment" is specific to the `astral_sea` hybrid map model; the engine's spatial layer treats locations as opaque string keys, so the naming should be agnostic to the underlying map type |
| 207 | `testharness.NewHarness` wires a `stubSpatialMap` into `eng.World` via `world.NewWithMap`; the stub accepts any location ID with TL 5 and distance 1 | Per #203, the harness needs a spatial map — not nil. The stub lets all scenario tests run without real spatial files, preserves the panic-on-nil invariant (#204), and gives TL 5 (permissive) so existing tests don't filter out any test assets |
| 208 | `SelectBuyOrder(purchasablePerWorld map[string][]*domain.AssetDefinition)` replaces `SelectBuyOrder(worlds, purchasable)`; `SelectExpandInfluenceOrder` gains `eligibleNewBaseWorlds []string` | TL enforcement in BuyAsset requires the world to be selected before definitions are filtered; passing a per-world map makes the dependency explicit and lets the collector show only valid defs for the chosen world. ExpandInfluence's eligible world list is pre-filtered by TL before the collector sees it |
| 209 | `purchasableDefinitions` uses `dispatch.ResolveWorldTechLevel` to apply tag-based TL modifiers before filtering; raw `loc.TechLevel()` is used in `eligibleNewBaseWorlds` (ExpandInfluence has no hook registry) | Consistent with the existing `ResolveAssetCost` pattern for BuyAsset; ExpandInfluence registry can be added when a tag actually needs to modify world TL for base placement |
| 210 | `factionHasPlanetaryGovernment(factionID, bases)` checks for any faction-owned base on the target world; no base `Type` field or faction tag check required at this stage | `Base` has no `Type` field in the domain; the simplest enforceable interpretation of the P-flag rule is "faction has an administrative presence (base) on this world." Refine if tag T-011 or base types are added |
