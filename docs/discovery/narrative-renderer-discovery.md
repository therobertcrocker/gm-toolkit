# Narrative Renderer — Discovery & Planning

The Narrative Renderer reads a campaign's `history.jsonl` for a given cycle and emits a human-readable narrative of what happened during that turn. The output is in-fiction prose — descriptive, voiced as a wire-service news dispatch — not a literal translation of mechanical effects ("Faction A lost 2 Coin"). It is the final feature before v1.0.0.

The renderer is post-hoc: it runs as a Cobra command after a cycle is complete, reads from history (the source of truth), and writes a markdown file to the campaign's `narratives/` directory. Re-running on the same cycle produces a new file with an incremented suffix; this is a feature, not a bug — variation on re-run is intentional.


## Responsibilities

- Read `history.jsonl` and select all records for a target cycle
- Resolve mutation IDs (faction, asset definition, goal, tag) to display names via faction state and rulebook
- Group causally related mutations into higher-level events (Buy, Sell, Refit, Attack, Bribe, Repair, Expand, Goal events)
- Pair cross-faction events (attacks, ability strikes) into single structs that carry both attacker and defender perspectives
- Render the resulting `CycleDigest` as wire-service-styled markdown with a cycle headline, lede paragraph, per-faction sections, and an aggregated tail for inactive factions
- Support reproducibility via `--seed` and quick previews via `--stdout`

<br/>
<br/>

# Architecture

## Hybrid: Digest Layer + Pluggable Renderer

The renderer is split into two layers:

1. **Digest layer (always deterministic):** reads history + faction state + rulebook → produces a structured `CycleDigest` with all IDs resolved to names, mutations causally grouped, and cross-faction events paired. Pure transform; no prose.
2. **Renderer layer (pluggable):** consumes the digest and produces prose. v1 ships only a deterministic wire-service implementation. The interface is shaped to accept a future LLM-backed renderer without rearchitecture.

The digest layer is reusable by any renderer — it does the *facts* work. The renderer does the *voice* work. This separation is the whole point: the v1 deterministic renderer and the post-v1 LLM renderer differ only in how they convert the same digest to prose.

<br/>

## Output Structure

Each cycle produces one markdown file in three parts:

- **Headline** — one line, top of file, `##` heading. Selected from cycle events by priority (see Decisions Log).
- **Lede paragraph** — 2–3 sentences, cycle-wide context.
- **Per-faction sections** — one `###` section per active faction, declarative news prose. Cross-faction events (esp. Attack) are owned by the actor's section; the defender's section does not echo (YAGNI for v1).
- **Quiet tail** — final paragraph aggregating factions that took a turn but had no narrative-worthy mutations: "No movements reported from X, Y, Z."

<br/>
<br/>

# Voice

Wire-service news dispatch with light editorializing. Reuters in space.

- Terse, present-tense, dispassionate
- Faction names as actors ("The Voltari struck Akaris")
- Light flavor verbs allowed ("struck", "fell back", "pressed") — no purple prose
- Numbers in prose match numbers in the digest
- Markdown headers (`##`, `###`); declarative sentences; no bullet lists

The voice is hardcoded in template phrasing pools for v1. Variant rotation via a seeded RNG provides re-run variation without configurability surface. Voice-as-config is reserved for the post-v1 LLM mode (see below).

<br/>
<br/>

# Post-v1 LLM Mode

A second renderer implementation is planned for after v1.0.0:

- **System prompt drives voice.** A TOML config (e.g. `narrative-voice.toml`) defines the system prompt — voice register, structural constraints, grounding rules. Different campaigns can ship different voices without code changes.
- **Digest is the grounding payload.** The same `CycleDigest` the deterministic renderer consumes is serialized as JSON and handed to the model. All names, numbers, and pairings are pre-resolved — the model is responsible for prose only, not for facts.
- **Hallucination is constrained at the input layer.** The model can't invent factions or worlds because it only sees the ones in the digest. Numerical drift is the residual risk; the system prompt enforces "numbers in prose must match the digest."
- **Cost and latency:** pennies per cycle on Haiku/Sonnet; 2–10 seconds per render. Acceptable for a CLI command.
- **Offline fallback:** deterministic renderer remains available and always works.

The architectural seam — `Renderer` interface — exists in v1 specifically so the LLM mode can drop in as a second implementation rather than a rewrite.

<br/>
<br/>

# Notes

- **Live state is acceptable for ID→name resolution.** Faction and asset names are stable; live state is a sufficient source. The exception is faction destruction: if a destroyed faction is removed from state, name resolution falls back to the faction ID. Faction destruction itself is detected by scanning history for `FactionHPDelta` taking `CurrentHP` to 0 — not by checking state.
- **`AssetMaintainedFlag` is dropped from narration.** Bookkeeping noise; not newsworthy.
- **Cross-faction pairing is single-record.** Attack and ability strikes write attacker-side and defender-side mutations into one `EventRecord`; pairing logic uses `CausedByFactionID` within that record.
- **Variation per re-run is intentional.** Re-running with a fresh seed produces different phrasing. `--seed` enables reproducibility when the GM finds a phrasing they like.
- **Depends on existing history emission** — `internal/faction/engine/history_engine.go` is the source of truth; no changes required.

<br/>
<br/>

# Decisions Log

### feature/narrative-renderer

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Hybrid architecture: deterministic digest layer + pluggable renderer interface | Digest is reusable by any renderer; separates facts from voice; lets us ship deterministic v1 and add LLM later without rearchitecture |
| 2 | v1.0.0 ships deterministic mode only; LLM mode is post-v1 | Deterministic gets us to v1 with no external dependencies; LLM is additive, not blocking |
| 3 | Cobra command `narrate <cycle>` with flags `--out`, `--seed`, `--stdout`, `--campaign` | Regenerable; matches existing command pattern; reproducibility via seed |
| 4 | Output to `campaigns/<name>/narratives/cycle-<NNN>.md`; rerun increments suffix `-002`, `-003` | Preserves history; user prunes manually if it gets noisy; lexical sort matches numeric sort |
| 5 | `--out <path>` bypasses increment logic and overwrites | Explicit path = explicit override; the user is asserting intent |
| 6 | Voice hardcoded in templates: wire-service / Reuters with light editorializing | YAGNI — no real present-day need for voice configurability in deterministic mode; templates are the voice |
| 7 | Voice-config-via-TOML deferred to LLM mode | The config knob only meaningfully exists in LLM mode where the config *is* the system prompt; deterministic templates aren't parameterizable in any useful way |
| 8 | Structure: cycle headline + lede + per-faction sections + aggregated quiet tail | Maps cleanly onto wire-service news layout (front page + section dispatches); easy to scan per faction; cross-faction events have a natural owner |
| 9 | Cross-faction events owned by actor's section in prose; defender section does not echo | YAGNI — discovery noted echo is "OK but not required"; revisit in Phase 5 if prose feels lopsided |
| 10 | Goal events flow as ordinary narrative beats — no special "milestone" struct | Treating them specially would force the digest to make narrative judgments; let the renderer lean into them via headline priority instead |
| 11 | Headline priority: Attack > Goal completion > Faction destruction > Homeworld shift > Goal abandonment > Quiet | War news leads the front page; long-term strategic shifts take inside columns; matches a hard-edged setting |
| 12 | Headline tie-break: highest impact (most damage / highest XP) → alphabetical by faction name | Deterministic resolution; documented in `digest.go` so it's auditable |
| 13 | `AssetMaintainedFlag` mutations are dropped from narration | Bookkeeping noise; not newsworthy |
| 14 | Variation on re-run is a feature, not a bug | Seeded RNG provides reproducibility when wanted; default seed is `time.Now().UnixNano()` |
| 15 | Faction destruction detected by scanning history for `FactionHPDelta` to 0 — not by reading state | Live state may have removed destroyed factions; history is the source of truth |
| 16 | Add `campaigns/narrative-fixture/` covering one cycle that exercises every digest path | Doubles as manual demo and snapshot-test source; existing `goal-test-1` is too sparse |

<br/>

### Notable alternatives rejected

| Decision # | Alternative | Why rejected |
|------------|-------------|--------------|
| 1 | Pure deterministic renderer (no digest abstraction) | Locks in templates; no upgrade path to LLM without rewriting; couples voice and facts |
| 1 | Pure LLM renderer (no digest layer) | Hallucination risk on names/numbers; no offline path; ties v1 to an external dependency |
| 6 | TOML-configurable voice in deterministic mode | Templates aren't parameterizable in any meaningful way without rewriting them; the knob only earns its keep in LLM mode |
| 8 | Single cycle-wide narrative (no per-faction sections) | Scanning is hard — "what did MY faction do" requires reading the whole thing |
| 8 | Per-faction sections only (no cycle headline/lede) | Cross-faction events become awkward; no front-page framing |
| 11 | Goal completion as top headline | Discussed; rejected because faction-level violence (attack) is more newsworthy in a wire-service voice |

<br/>
<br/>

# Open Questions

| # | Question | Relevant Feature |
|---|----------|------------------|
| 1 | Should the defender's section echo a brief mention of cross-faction events when the attack was significant (e.g. asset destroyed)? | Per-faction sections |
| 2 | If the post-v1 LLM mode lands, does the deterministic renderer stay or get retired? | LLM mode |
| 3 | Is a global "campaign-wide" narrative (multi-cycle digest) ever wanted, or are single-cycle dispatches sufficient? | Future scope |
