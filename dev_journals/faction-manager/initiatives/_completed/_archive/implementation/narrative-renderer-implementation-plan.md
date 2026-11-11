# Narrative Renderer — Implementation Plan

A phased build guide for the `feature/narrative-renderer` branch. Read the discovery doc first ([narrative-renderer-discovery.md](../discovery/narrative-renderer-discovery.md)), then the key files listed per phase.

**Branch:** `feature/narrative-renderer`

<br/>

## How to use this doc

For per-phase work, **scope your reads to the relevant sections** rather than loading the full plan. Each phase only needs a subset. Run `grep -n "^##" docs/initiatives/implementation/narrative-renderer-implementation-plan.md` to get current line numbers, then `Read` with `offset` + `limit`.

| Section | Read for | Notes |
|---------|----------|-------|
| Design Decisions | All phases | Always read first |
| Verified `Cause` Set | Phase 2 | Drives the grouping switch |
| Package Layout | Phases 1, 3, 4 | File paths and roles |
| Domain Types | Phases 1, 2 | Full struct definitions |
| Mutation → Digest Mapping | Phase 2 | Per-`Cause` grouping rules |
| Renderer Interface | Phase 3 | Interface + wire impl shape |
| Cobra Command | Phase 4 | Command body + path resolution |
| Phase 1 — Digest Skeleton | Phase 1 | Scope + commit message |
| Phase 2 — Digest Construction | Phase 2 | Scope + commit message |
| Phase 3 — Wire Renderer | Phase 3 | Scope + commit message |
| Phase 4 — Cobra Command and Fixture | Phase 4 | Scope + commit message |
| Phase 5 — Docs and Smoke | Phase 5 | Scope + commit message |
| Verification | All phases | Smoke commands live here |
| Risks | All phases | Read at start of each phase |

<br/>

## Design Decisions (read before any phase)

These were ratified in the discovery session and must be followed precisely:

| # | Decision |
|---|----------|
| 1 | Hybrid architecture: deterministic `digest` layer + pluggable `Renderer` interface |
| 2 | v1 ships deterministic wire-service renderer only; LLM renderer is post-v1 |
| 3 | New Cobra command `narrate <cycle>` with `--out`, `--seed`, `--stdout`, `--campaign` |
| 4 | Default output path `campaigns/<name>/narratives/cycle-<NNN>.md`; rerun increments to `-002`, `-003` (3-digit padding) |
| 5 | `--out <path>` bypasses increment and overwrites |
| 6 | Output structure: cycle headline + lede + per-faction sections + aggregated quiet tail |
| 7 | Cross-faction events owned by actor's section; defender section does not echo |
| 8 | Goal events flow as ordinary narrative beats — no special milestone struct |
| 9 | Headline priority: Attack > GoalCompleted > FactionDestroyed > HomeworldShift > GoalAbandoned > Quiet; tie-break by impact then alphabetical |
| 10 | `AssetMaintainedFlag` mutations dropped from narration |
| 11 | Faction destruction detected by scanning history for `FactionHPDelta` to 0 |
| 12 | Live state acceptable for ID→name resolution; `resolveFactionName` falls back to ID |
| 13 | Variant rotation via seeded RNG; default seed is `time.Now().UnixNano()`; reported in stdout for reproducibility |

<br/>

## Verified `Cause` Set

Driven by `grep "Cause:" internal/faction/engine/`:

`{abandon_goal, ability, attack, bookkeeping, bribe, buy, expand, goal_completed, occupation_failed, player_choice, refit, repair, sell}`

Causal grouping in `Build` keys off these literals. If a new action adds a new `Cause`, the grouping switch must be extended.

<br/>

## Package Layout

| Path | Role |
|------|------|
| `internal/faction/narrative/digest/digest.go` | `CycleDigest` types + pure `Build` function |
| `internal/faction/narrative/digest/digest_test.go` | Unit tests over `[]EventRecord → CycleDigest` |
| `internal/faction/narrative/renderer.go` | `Renderer` interface + `NewWireRenderer` factory |
| `internal/faction/narrative/wire_renderer.go` | Deterministic wire-service implementation |
| `internal/faction/narrative/wire_templates.go` | Phrasing pools (variant rotation) |
| `internal/faction/narrative/history_reader.go` | `LoadCycle(historyPath, cycle)` |
| `internal/faction/narrative/wire_renderer_test.go` | Golden-file render tests |
| `internal/faction/narrative/testdata/*.golden.md` | Snapshot fixtures |
| `cmd/faction-manager/commands/narrate.go` | Registration shim (matches `review.go` pattern) |
| `cmd/faction-manager/commands/narrate/cmd.go` | Command body — flags, path resolution, file write |
| `campaigns/narrative-fixture/faction_state.toml` | Demo + test fixture |
| `campaigns/narrative-fixture/history.jsonl` | One cycle exercising every digest path |

<br/>

## Domain Types

### CycleDigest

```go
package digest

type CycleDigest struct {
    Cycle          int
    ActiveFactions []FactionBeat
    QuietFactions  []FactionRef
    Cross          []CrossEvent
    Headline       Headline
}

type FactionRef struct {
    ID   string
    Name string
}

type FactionBeat struct {
    Faction      FactionRef
    Goal         *GoalRef
    CoinDelta    int
    HPDelta      int
    XPGained     int
    Acquisitions []AssetMove
    Losses       []AssetMove
    Movements    []AssetMove
    Bribes       []BribeEvent
    Repairs      []RepairEvent
    Expansions   []ExpansionEvent
    GoalEvents   []GoalEvent
    StealthOps   []StealthEvent
    Notes        []string
}

type GoalRef struct {
    ID   string
    Name string
}

type AssetMove struct {
    AssetID      string
    AssetName    string
    DefinitionID string
    From         string
    To           string
    Cause        string
}

type BribeEvent struct {
    BaseID   string
    Location string
    Coin     int
}

type RepairEvent struct {
    Target    AssetMove
    HPGained  int
    Coin      int
    IsFaction bool
}

type ExpansionEvent struct {
    BaseID   string
    Location string
    NewBase  bool
    HPDelta  int
    Coin     int
}

type GoalEventKind int

const (
    GoalCompleted GoalEventKind = iota
    GoalAbandoned
    GoalHomeworldShift
    GoalTagGained
)

type GoalEvent struct {
    Kind      GoalEventKind
    GoalID    string
    GoalName  string
    XPAwarded int
    FromWorld string
    ToWorld   string
    TagName   string
}

type StealthEvent struct {
    AssetID   string
    AssetName string
    Applied   bool
}

type CrossKind int

const (
    CrossAttack CrossKind = iota
    CrossAbilityStrike
)

type CrossEvent struct {
    Kind                   CrossKind
    Attacker               FactionRef
    Defender               FactionRef
    AttackerAsset          AssetMove
    DefenderAsset          AssetMove
    DamageToDefender       int
    DamageToAttacker       int
    DefenderAssetDestroyed bool
    AttackerAssetDestroyed bool
    BaseHit                *BaseHitDetail
    CoinDrained            int
    StealthBroken          bool
}

type BaseHitDetail struct {
    BaseID    string
    Location  string
    Damage    int
    Destroyed bool
}

type HeadlineKind int

const (
    HeadlineMajorAttack HeadlineKind = iota
    HeadlineGoalCompleted
    HeadlineFactionDestroyed
    HeadlineHomeworldShift
    HeadlineGoalAbandoned
    HeadlineQuiet
)

type Headline struct {
    Subject FactionRef
    Kind    HeadlineKind
    Detail  string
}
```

<br/>

## Mutation → Digest Mapping

| Mutation | `Cause` | Maps to |
|----------|---------|---------|
| `CoinDelta` | `bookkeeping` | `FactionBeat.CoinDelta` (income aggregation) |
| `CoinDelta` | `buy`, `refit`, `sell` | Embedded in `Acquisitions`/`Losses`; aggregated into `FactionBeat.CoinDelta` |
| `CoinDelta` | `bribe` | Paired with `InfluenceDelta` → `BribeEvent` |
| `CoinDelta` | `repair` | Embedded in `RepairEvent` |
| `CoinDelta` | `expand` | Embedded in `ExpansionEvent` |
| `CoinDelta` | `abandon_goal` | Aggregated into `FactionBeat.CoinDelta` (income forfeit) |
| `AssetAdded` | `buy`, `refit` | `Acquisitions` |
| `AssetRemoved` | `sell` | `Losses` |
| `AssetRemoved` | `refit` | Encoded as `From` on the paired `AssetAdded` in `Acquisitions` |
| `AssetRemoved` | `attack` | `CrossEvent.DefenderAsset`, `DefenderAssetDestroyed=true` |
| `AssetRemoved` | `bookkeeping` | Dropped (asset attrition is noise) |
| `AssetHPDelta` | `repair` | `RepairEvent` |
| `AssetHPDelta` | `attack` | `CrossEvent.DamageToDefender` or `DamageToAttacker` based on `FactionID` vs actor |
| `AssetMoved` | `ability` | `Movements` |
| `AssetStealthApplied` | `buy` | `StealthOps{Applied: true}` |
| `AssetStealthCleared` | `attack`, `ability` | `StealthOps{Applied: false}`; ability case also sets `CrossEvent.StealthBroken` |
| `AssetMaintainedFlag` | any | Dropped |
| `FactionHPDelta` | `attack` | Aggregated into `FactionBeat.HPDelta`; also echoed in `CrossEvent.BaseHit.Damage` if from base redirect |
| `FactionHPDelta` | `repair` | `RepairEvent{IsFaction: true}` |
| `BaseHPDelta` | `attack` | `CrossEvent.BaseHit` |
| `BaseHPDelta` | `expand` | `ExpansionEvent` |
| `BaseDestroyed` | `attack` | `CrossEvent.BaseHit{Destroyed: true}` |
| `BaseAdded` | `expand` | `ExpansionEvent{NewBase: true}` |
| `BaseHealed` | `expand` | `ExpansionEvent` |
| `BaseExpanded` | `expand` | `ExpansionEvent` |
| `GoalAbandoned` | `abandon_goal` | `GoalEvent{Kind: GoalAbandoned}` |
| `GoalCompleted` | `goal_completed` | `GoalEvent{Kind: GoalCompleted}` |
| `XPAwarded` | `goal_completed` | Folded into matching `GoalCompleted` event; aggregated into `FactionBeat.XPGained` |
| `HomeworldChanged` | `goal_completed` | `GoalEvent{Kind: GoalHomeworldShift}` |
| `TagAdded` | `goal_completed` | `GoalEvent{Kind: GoalTagGained}` |
| `InfluenceDelta` | `bribe` | `BribeEvent` |

Cross-faction events use single-record pairing: any mutation in the actor's `EventRecord` with `FactionID != EventRecord.FactionID` is the defender side. Pair these into one `CrossEvent` per attacker/defender combination.

<br/>

## Renderer Interface

```go
package narrative

import "github.com/therobertcrocker/gm-toolkit/internal/faction/narrative/digest"

type Renderer interface {
    Render(cycleDigest digest.CycleDigest, seed int64) (string, error)
}

func NewWireRenderer() Renderer
```

The deterministic implementation:

```go
type wireRenderer struct{}

func (renderer *wireRenderer) Render(cycleDigest digest.CycleDigest, seed int64) (string, error) {
    rng := rand.New(rand.NewSource(seed))
    var output strings.Builder

    output.WriteString(renderHeadline(cycleDigest, rng))
    output.WriteString("\n\n")
    output.WriteString(renderLede(cycleDigest, rng))
    output.WriteString("\n\n")
    for _, beat := range cycleDigest.ActiveFactions {
        output.WriteString(renderFactionSection(beat, cycleDigest.Cross, rng))
        output.WriteString("\n")
    }
    if len(cycleDigest.QuietFactions) > 0 {
        output.WriteString(renderQuietTail(cycleDigest.QuietFactions, rng))
    }
    return output.String(), nil
}
```

Variant rotation lives in `wire_templates.go` as package-level slices; helper `pick(rng, pool)` selects deterministically.

<br/>

## Cobra Command

`cmd/faction-manager/commands/narrate.go` is a one-line shim that registers the command. The body lives in `cmd/faction-manager/commands/narrate/cmd.go`:

- Flag registration (`--campaign` required; `--out`, `--seed`, `--stdout` optional)
- Default seed: `time.Now().UnixNano()` when `--seed` not set; print seed to stdout after writing
- Load faction state via `state.Load(campaignPaths.State)`
- Load history via `narrative.LoadCycle(campaignPaths.History, cycleNumber)`
- Build digest via `digest.Build(events, cycleNumber, factionState, rulebook)`
- Render via `narrative.NewWireRenderer().Render(cycleDigest, seed)`
- Write to `--out` (verbatim) or `resolveOutputPath(campaignID, cycleNumber)` (with increment)
- `--stdout` skips file write entirely

`resolveOutputPath`: if `cycle-NNN.md` doesn't exist, return it. Otherwise scan `cycle-NNN-002.md`, `-003.md`, … up to `-999`; return first non-existent. Both numbers zero-padded so lexical sort matches numeric sort.

<br/>

## Phasing

Five phases. Commit at the end of each per CLAUDE.md cadence.

### Phase 1 — Digest Skeleton

Land types and history reader; no rendering yet.

#### Key files to read

| File | Why |
|------|-----|
| `internal/faction/domain/event.go` | `EventRecord` and `MutationRecord` — the types `LoadCycle` reads and returns |
| `cmd/faction-manager/paths/paths.go` | `Paths.History` is the path `LoadCycle` receives |

#### Tasks

**1. Create `internal/faction/narrative/digest/digest.go`**

All type definitions from the Domain Types section above. No behavior yet.

**2. Create `internal/faction/narrative/history_reader.go`**

```go
package narrative

// LoadCycle reads history.jsonl at historyPath and returns all EventRecords
// whose Cycle field matches cycleNumber. Returns an error if no records match.
func LoadCycle(historyPath string, cycleNumber int) ([]domain.EventRecord, error)
```

Implementation: open the file; scan line by line with `bufio.Scanner`; `json.Unmarshal` each line into `domain.EventRecord`; collect those where `record.Cycle == cycleNumber`. Return `fmt.Errorf("no history records for cycle %d", cycleNumber)` if the slice is empty.

**3. Add `digest.Build` stub to `digest.go`**

```go
func Build(
    records []domain.EventRecord,
    cycleNumber int,
    factionState *state.FactionState,
    rulebook *loader.Rulebook,
) (CycleDigest, error)
```

Stub body: return `CycleDigest{}, fmt.Errorf("no history records for cycle %d", cycleNumber)` when `len(records) == 0`; otherwise return `CycleDigest{Cycle: cycleNumber}, nil`.

**4. Create `internal/faction/narrative/digest/digest_test.go`**

Table-driven tests for `LoadCycle`:
- file not found → wrapped error containing the path
- JSONL with two cycles → only records matching the requested cycle returned
- JSONL with one cycle, matching → records returned
- JSONL with one cycle, non-matching → error "no history records for cycle N"

One test for the stub: `Build(nil, 1, ...)` → error; `Build(oneRecord, 1, ...)` → `CycleDigest{Cycle: 1}`.

**Commit:** `feat: phase 1 — narrative digest types and history reader`

<br/>

### Phase 2 — Digest Construction

Fill in `digest.Build` with full mutation→digest mapping, cross-faction pairing, ID resolution, and headline selection.

#### Key files to read

| File | Why |
|------|-----|
| `internal/faction/domain/mutation.go` | All mutation type strings (`"coin_delta"`, `"asset_removed"`, etc.) and their JSON field names — needed for the deserialization switch |
| `internal/faction/domain/faction.go` | `Faction.Name`, `Faction.Assets`, `Faction.Bases`, `Faction.Tags` — used by name resolution helpers |
| `internal/faction/loader/loader.go` | `Rulebook.Assets`, `Rulebook.Goals`, `Rulebook.Tags` — for definition lookups |
| `internal/faction/state/faction_state.go` | `FactionState.Factions map[string]*domain.Faction` — keyed by faction ID |

#### Tasks

**1. Name resolution helpers** (unexported, in `digest.go`)

```go
func resolveFactionName(factionID string, factionState *state.FactionState) string
// factionState.Factions[factionID].Name, or factionID if absent

func resolveAssetName(factionID, assetID string, factionState *state.FactionState) string
// scan factionState.Factions[factionID].Assets for matching ID; return Name or assetID

func resolveGoalName(goalID string, rulebook *loader.Rulebook) string
// rulebook.Goals[goalID].Name, or goalID if absent

func resolveTagName(tagID string, rulebook *loader.Rulebook) string
// rulebook.Tags[tagID].Name, or tagID if absent
```

**2. Deserialization pattern**

Each case in the grouping switch unmarshals `MutationRecord.Payload` into the concrete type:

```go
case "coin_delta":
    var m domain.CoinDelta
    if err := json.Unmarshal(mr.Payload, &m); err != nil {
        return CycleDigest{}, err
    }
    // map m to digest fields
```

`mr` is a `domain.MutationRecord` from `record.Mutations`. `record` is a `domain.EventRecord` from the input slice.

**3. Per-faction accumulation loop**

`Build` maintains a `beats map[string]*FactionBeat` keyed by faction ID. For each `domain.EventRecord`, create or fetch the beat for `record.FactionID`, then iterate `record.Mutations` and dispatch on `MutationRecord.Type` per the Mutation→Digest Mapping table.

Unknown type: append `fmt.Sprintf("unknown mutation: %s", mr.Type)` to `beat.Notes` rather than returning an error (per Risk #1).

`AssetMaintainedFlag`: skip (drop entirely).

**4. Cross-faction pairing**

Each `domain.EventRecord` is owned by the acting faction (`record.FactionID` = actor). Within that record, any mutation where the mutation's `FactionID` field differs from `record.FactionID` is a defender-side mutation. Collect attacker + defender mutations for a given `(actor, target)` pair into one `CrossEvent`:

```go
// Within one EventRecord iteration:
// actor  = record.FactionID
// target = mutation's FactionID field (when it differs from record.FactionID)
```

For attack records: `AssetHPDelta` with `FactionID == record.FactionID` → `DamageToAttacker`; `AssetHPDelta` with `FactionID != record.FactionID` → `DamageToDefender`. `AssetRemoved` with `FactionID != record.FactionID` → `DefenderAsset`, `DefenderAssetDestroyed = true`. `AssetRemoved` with `FactionID == record.FactionID` → `AttackerAsset`, `AttackerAssetDestroyed = true`. `BaseHPDelta` / `BaseDestroyed` with `FactionID != record.FactionID` → `BaseHit`.

**5. Faction destruction detection**

After processing all records: for any faction ID that appears in at least one record but is absent from `factionState.Factions`, that faction was destroyed this cycle. Surface as `HeadlineFactionDestroyed` in headline selection and as a `GoalEvent{Kind: GoalCompleted}` beat entry for any faction whose `Destroy the Foe` goal targeted it (detected by `CausedByFactionID` on a `FactionHPDelta` with `delta` taking HP to 0 — check this by summing `FactionHPDelta.Delta` across the cycle for that faction).

**6. ActiveFactions vs QuietFactions split**

A faction beat is quiet if all of the following are empty: `Acquisitions`, `Losses`, `Movements`, `Bribes`, `Repairs`, `Expansions`, `GoalEvents`, `StealthOps`, and no `CrossEvent` has `Attacker.ID == faction.ID` or `Defender.ID == faction.ID`. Quiet factions go into `CycleDigest.QuietFactions` as `FactionRef` values; active factions go into `ActiveFactions`.

**7. Headline selection**

After all beats are built, walk the priority chain (decision #9) and pick the first matching kind:

1. `HeadlineMajorAttack` — any `CrossEvent` exists; tie-break: highest `DamageToDefender + DamageToAttacker`, then alphabetical by `Attacker.Name`
2. `HeadlineGoalCompleted` — any `GoalEvent{Kind: GoalCompleted}`; tie-break: highest `XPAwarded`, then alphabetical by faction name
3. `HeadlineFactionDestroyed` — destruction detected (step 5); tie-break: alphabetical by faction name
4. `HeadlineHomeworldShift` — any `GoalEvent{Kind: GoalHomeworldShift}`; tie-break: alphabetical
5. `HeadlineGoalAbandoned` — any `GoalEvent{Kind: GoalAbandoned}`; tie-break: alphabetical
6. `HeadlineQuiet` — fallback

**8. Expand `digest_test.go`**

One test per row in the Mutation→Digest Mapping table, using hand-built `[]domain.EventRecord` literals. Additional tests: cross-faction pairing with attacker and defender mutations in the same record; headline priority resolution (attack beats goal completion); tie-break (higher damage wins; equal damage → alphabetical).

**Commit:** `feat: phase 2 — digest construction with full mutation mapping`

<br/>

### Phase 3 — Wire Renderer

Add the renderer interface and deterministic implementation.

#### Key files to read

| File | Why |
|------|-----|
| `internal/faction/narrative/digest/digest.go` | All `CycleDigest` types — the input to every render function |

#### Tasks

**1. Create `internal/faction/narrative/renderer.go`**

`Renderer` interface and `NewWireRenderer()` factory as shown in the Renderer Interface section above.

**2. Create `internal/faction/narrative/wire_templates.go`**

Package-level `[]string` pools for each template slot. Minimum two variants per pool so variation tests can pass. Helper:

```go
func pick(rng *rand.Rand, pool []string) string {
    return pool[rng.Intn(len(pool))]
}
```

Template pools needed: `headlineAttackTemplates`, `headlineGoalCompletedTemplates`, `headlineFactionDestroyedTemplates`, `headlineHomeworldTemplates`, `headlineGoalAbandonedTemplates`, `headlineQuietTemplates`, `ledeAttackTemplates`, `ledeGoalTemplates`, `ledeQuietTemplates`, and per-event pools for acquisitions, losses, movements, bribes, repairs, expansions, goal events, stealth ops, cross-attack, cross-ability, quiet tail.

**3. Create `internal/faction/narrative/wire_renderer.go`**

Top-level `Render` method body as shown in the Renderer Interface section, plus these unexported functions:

```go
func renderHeadline(cycleDigest digest.CycleDigest, rng *rand.Rand) string
// picks from the pool matching cycleDigest.Headline.Kind;
// substitutes Subject.Name and Detail into the template

func renderLede(cycleDigest digest.CycleDigest, rng *rand.Rand) string
// 2–3 sentence paragraph; pool keyed on whether cycle had attacks, goal events, or was quiet

func renderFactionSection(beat digest.FactionBeat, crossEvents []digest.CrossEvent, rng *rand.Rand) string
// emits "### FactionName\n\n" then one prose sentence per event in order:
// goal events → acquisitions → losses → movements → bribes → repairs → expansions → stealth ops
// then any CrossEvent where CrossEvent.Attacker.ID == beat.Faction.ID

func renderQuietTail(quietFactions []digest.FactionRef, rng *rand.Rand) string
// picks a tail template and substitutes a comma-joined list of faction names
```

**4. Create `internal/faction/narrative/wire_renderer_test.go`**

Test cases:
- Same seed → byte-identical output: `render(digest, 1) == render(digest, 1)`
- Different seeds → not equal: `render(digest, 1) != render(digest, 2)`
- Golden-file tests: `render(attackDigest, 42)` matches `testdata/attack.golden.md`; `render(goalDigest, 42)` matches `testdata/goal_completed.golden.md`; `render(quietDigest, 42)` matches `testdata/quiet.golden.md`

Use `flag.Bool("update", ...)` in `TestMain` to regenerate golden files.

**5. Create `internal/faction/narrative/testdata/*.golden.md`**

Hand-write `attack.golden.md`, `goal_completed.golden.md`, `quiet.golden.md` using seed 42 output. After the renderer compiles, regenerate with `-update` and commit the result.

**Commit:** `feat: phase 3 — wire-service deterministic renderer`

<br/>

### Phase 4 — Cobra Command and Fixture

Wire the command into the binary; add the demo fixture campaign.

#### Key files to read

| File | Why |
|------|-----|
| `cmd/faction-manager/commands/turn/cmd.go` | Exact pattern for `--campaign` → `paths.New()` → `state.Load()` |
| `cmd/faction-manager/commands/app.go` | Where to add `root.AddCommand(a.narrateCmd())` |
| `cmd/faction-manager/paths/paths.go` | `Paths` struct — note it has no `Narratives` field; build that path directly from `campaignID` |

#### Tasks

**1. Create `cmd/faction-manager/commands/narrate.go`** — one-method shim on `App`:

```go
func (a *App) narrateCmd() *cobra.Command {
    return narrate.NewCmd(a.Engine.Rulebook)
}
```

**2. Edit `cmd/faction-manager/commands/app.go`**

Add one line in `Execute()` after the existing `root.AddCommand` calls:

```go
root.AddCommand(a.narrateCmd())
```

**3. Create `cmd/faction-manager/commands/narrate/cmd.go`**

Flags: `--campaign` (string, required), `--out` (string), `--seed` (int64), `--stdout` (bool).

Detect whether `--seed` was explicitly set using `cmd.Flags().Changed("seed")` inside `RunE`; if not set, assign `seed = time.Now().UnixNano()`.

Command body:
1. Parse `args[0]` as int; return error on parse failure
2. `p := paths.New(campaignID)`
3. `factionState, err := state.Load(p.State)`
4. `records, err := narrative.LoadCycle(p.History, cycleNumber)`
5. `cycleDigest, err := digest.Build(records, cycleNumber, factionState, rulebook)`
6. `output, err := narrative.NewWireRenderer().Render(cycleDigest, seed)`
7. If `--stdout`: `fmt.Print(output)`, print seed to stderr, return
8. If `--out` set: write to that path verbatim (create dirs, overwrite)
9. Otherwise: `dest, err = resolveOutputPath(campaignID, cycleNumber)`; create dirs; write

After a successful file write: `fmt.Printf("wrote %s (seed: %d)\n", dest, seed)`.

`resolveOutputPath(campaignID string, cycleNumber int) (string, error)`:
- Base: `filepath.Join(".", "campaigns", campaignID, "narratives")`
- Primary: `fmt.Sprintf("cycle-%03d.md", cycleNumber)` — return if absent
- Increments: `fmt.Sprintf("cycle-%03d-%03d.md", cycleNumber, i)` for i = 2..999 — return first absent
- If all 999 slots taken: return error

**4. Create `campaigns/narrative-fixture/faction_state.toml`**

Two factions minimum. Faction A: has Force assets, a base, an active goal. Faction B: has Force and Wealth assets, a base. Both factions must have IDs and names that match the `history.jsonl` records below.

**5. Create `campaigns/narrative-fixture/history.jsonl`**

One `domain.EventRecord` per event type, written as JSONL. Each line is a JSON object:

```json
{"cycle":1,"faction_id":"faction-a","timestamp":"2024-01-01T00:00:00Z","mutations":[...]}
```

The `mutations` array contains `{"type":"<type_string>","payload":{...}}` objects. Fields in `payload` must match the JSON tags on the corresponding mutation struct in `internal/faction/domain/mutation.go`.

Cover at minimum one record exercising: `buy` (AssetAdded + CoinDelta), `sell` (AssetRemoved + CoinDelta), `refit` (AssetAdded + AssetRemoved + CoinDelta), `attack` cross-event (AssetHPDelta on defender, AssetRemoved on defender, FactionHPDelta on defender — all with `caused_by_faction_id` set to attacker), `bribe` (CoinDelta + InfluenceDelta), `expand` (BaseAdded + CoinDelta), `repair` (AssetHPDelta + CoinDelta), `goal_completed` (GoalCompleted + XPAwarded), `abandon_goal` (GoalAbandoned + CoinDelta), `goal_completed` with `homeworld_changed`.

**Commit:** `feat: phase 4 — narrate command and demo fixture`

<br/>

### Phase 5 — Docs and Smoke

**Files:** `docs/dev-journal-factions.md`, `docs/decisions-log.md`, `docs/tracking/turn-engine-journal.md`.

**Scope:** dev-journal and tracking-journal updates per CLAUDE.md merge checklist; decisions-log entries for the decisions in the discovery doc; manual smoke against `goal-test-1` and the new fixture; reconcile the discovery doc's open questions if implementation surfaced answers (note: discovery doc itself is not updated retroactively — post-implementation answers go into the decisions log).

**Commit:** `docs: phase 5 — narrative renderer journals and decisions log`

<br/>

## Verification

### Unit (per phase)

- **Phase 1:** `LoadCycle` table-driven tests; empty-cycle error.
- **Phase 2:** one test per row in the mutation→digest mapping table; cross-faction pairing with attacker/defender mutations in the same record; headline priority resolution including tie-breaks.
- **Phase 3:** golden-file tests; determinism (same seed → byte-identical output); variation (different seeds → not equal).

### Integration

`LoadCycle` + `digest.Build` + `NewWireRenderer().Render` against `campaigns/narrative-fixture/`.

### Manual smoke (User will run these commands and verify outputs match expectations):

```
cd cmd/faction-manager
go build -o bin/faction-manager .
FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data \
  ./bin/faction-manager narrate 1 --campaign narrative-fixture --seed 1
# expect: campaigns/narrative-fixture/narratives/cycle-001.md exists

FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data \
  ./bin/faction-manager narrate 1 --campaign narrative-fixture --seed 1
# expect: cycle-001-002.md created, identical contents (same seed)

FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data \
  ./bin/faction-manager narrate 1 --campaign narrative-fixture --seed 2
# expect: cycle-001-003.md created, different phrasing

FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data \
  ./bin/faction-manager narrate 1 --campaign narrative-fixture --stdout
# expect: prints to stdout, writes nothing

FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data \
  ./bin/faction-manager narrate 1 --campaign narrative-fixture --out /tmp/test.md
# expect: writes /tmp/test.md, no increment

FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data \
  ./bin/faction-manager narrate 99 --campaign narrative-fixture
# expect: error "no history records for cycle 99", exit 1
```

<br/>

## Risks

| Risk | Mitigation |
|------|-----------|
| `Cause` strings drift if a new action is added during the branch | Phase 2 grouping switch is exhaustive; an unknown `Cause` should hit a default branch that surfaces a `Note` rather than panic |
| Faction destruction removes the faction from state | Detect destruction by scanning history; `resolveFactionName` falls back to ID |
| Cross-faction echo feels lopsided in prose | Phase 5 manual review; if real, add minimal echo paragraph in defender section as a follow-up |
| Golden files churn on every template tweak | Accept it — that's what golden files are for; regenerate via `-update` flag in the test |
