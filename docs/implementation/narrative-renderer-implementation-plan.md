# Narrative Renderer — Implementation Plan

A phased build guide for the `feature/narrative-renderer` branch. Read the discovery doc first ([narrative-renderer-discovery.md](../discovery/narrative-renderer-discovery.md)), then the key files listed per phase.

**Branch:** `feature/narrative-renderer`

<br/>

## How to use this doc

For per-phase work, **scope your reads to the relevant sections** rather than loading the full plan. Each phase only needs a subset. Run `grep -n "^##" docs/implementation/narrative-renderer-implementation-plan.md` to get current line numbers, then `Read` with `offset` + `limit`.

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

**Files:** `internal/faction/narrative/digest/digest.go`, `internal/faction/narrative/history_reader.go`, `internal/faction/narrative/digest/digest_test.go`.

**Scope:** all type definitions; `LoadCycle`; `digest.Build` stub returning a `CycleDigest` with `Cycle` set and the rest empty. Tests cover `LoadCycle` (file not found, multiple cycles, single cycle filter) and the empty-build error case ("no history records for cycle N").

**Commit:** `feat: phase 1 — narrative digest types and history reader`

<br/>

### Phase 2 — Digest Construction

Fill in `digest.Build` with full mutation→digest mapping, cross-faction pairing, ID resolution, and headline selection.

**Files:** `internal/faction/narrative/digest/digest.go`, expanded `digest_test.go`.

**Scope:** per-`Cause` grouping for all mutation types; cross-faction pairing via `CausedByFactionID`; `resolveFactionName`/`resolveAssetName`/`resolveGoalName`/`resolveTagName` helpers with ID fallback; headline selection per priority (decision #9); faction destruction detection by scanning history for `FactionHPDelta` to 0. Tests cover one path per row in the mutation→digest mapping table.

**Commit:** `feat: phase 2 — digest construction with full mutation mapping`

<br/>

### Phase 3 — Wire Renderer

Add the renderer interface and deterministic implementation.

**Files:** `internal/faction/narrative/renderer.go`, `internal/faction/narrative/wire_renderer.go`, `internal/faction/narrative/wire_templates.go`, `internal/faction/narrative/wire_renderer_test.go`, `internal/faction/narrative/testdata/*.golden.md`.

**Scope:** `Renderer` interface; full template set (headline, lede, faction section, quiet tail, per-event renderers); variant pools; golden-file tests with hand-built `CycleDigest` literals; same-seed determinism test; different-seed variation test.

**Commit:** `feat: phase 3 — wire-service deterministic renderer`

<br/>

### Phase 4 — Cobra Command and Fixture

Wire the command into the binary; add the demo fixture campaign.

**Files:** `cmd/faction-manager/commands/narrate.go`, `cmd/faction-manager/commands/narrate/cmd.go`, `cmd/faction-manager/commands/app.go` (one-line registration), `campaigns/narrative-fixture/history.jsonl`, `campaigns/narrative-fixture/faction_state.toml`.

**Scope:** command body with all flags wired; `resolveOutputPath` with increment-on-rerun and `--out` bypass; fixture campaign exercising buy, sell, refit, attack (with destruction + counter), bribe, expand, repair, goal completion, abandon, homeworld in one cycle.

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

### Manual smoke

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
