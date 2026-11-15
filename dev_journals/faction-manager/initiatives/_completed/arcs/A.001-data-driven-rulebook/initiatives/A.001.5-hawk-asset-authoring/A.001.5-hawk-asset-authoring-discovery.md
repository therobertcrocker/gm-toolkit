# H.A.W.K Asset Authoring — Discovery

## In Plain Words

- **What** — Write H.A.W.K's own faction rulebook: its tags and its assets (Force, Cunning, Wealth), each with H.A.W.K names and fiction, and each mapped onto the engine's data-driven mechanics. First, interview Robert to seed a written H.A.W.K canon. Then set up a real H.A.W.K campaign with its own map and starting factions, so every pass ends with a live turn.
- **Why** — The engine runs every asset, tag and goal from rulebook data, but it has only ever run SWN. H.A.W.K is the campaign the tool exists to serve, and it has no rulebook folder, no campaign and no written canon. Until it has them, the arc's claim (the engine holds a rulebook that is not its reference) is untested.
- **How it fits** — The last initiative of arc `A.001`. It uses what `A.001.2`, `A.001.3`, `A.001.6` and `A.001.7` built (the ability and effect registries, tags as data) and the shared `_core` goals from `A.001.4` and `A.001.8`. Closing it reaches "H.A.W.K playable": Robert opens the H.A.W.K campaign in the TUI and runs a turn on its map. Then the arc close opens.
- **This doc** — A short Discovery (AP-10). It settles the process shape: six efforts, the canon's layout, the campaign's setup, the `HAWK-` ID scheme, and how each pass records its sort calls. It holds no fiction; the lore interview is Effort 1, in its own session.

## Problem

The engine is multi-rulebook in shape but not in proof. Every rulebook folder in the repo is SWN (`rulebooks/swn/`) or shared (`rulebooks/_core/`), and every campaign is a sandbox. A playable H.A.W.K needs:

- **A rulebook folder** — `rulebooks/hawk/` with `assets/<category>_assets.toml`, `tags.toml`, `settings.toml` and a `rulebook.toml` that includes `../_core` for goals. None of it exists.
- **A campaign** — `campaigns/hawk/` with a map and starting factions. None of it exists.
- **A canon** — the setting, its factions, its places, its terminology and the fiction behind each element. Today H.A.W.K exists only in Robert's head and in `F.041`'s encounter work. Authoring without a canon means inventing setting facts during a mechanical pass, then contradicting them a pass later.
- **A repeatable pass** — 73 SWN assets (Force 25, Cunning 24, Wealth 24) and 20 tags each need a H.A.W.K reflavor and a mechanical home. Some will reuse a built primitive as pure data; some will need a near-miss parameterized or a new primitive built. Without a recorded sort, a later pass cannot see what an earlier one decided, and the arc close cannot audit what was built.

## Architecture Context

Pages read: `docs/architecture/architecture-overview.md`, `docs/architecture/engine/ability-catalog.md`.

Binding context:

- **Extensible static data** (overview design principle) — rules content lives in the campaign's rulebook TOML, which the GM owns; changing it needs no Go. H.A.W.K content is data first; Go appears only where the sort finds a missing primitive.
- **The flag picks the engine, the effect picks the primitive** (ability-catalog Key Decision; ACD-1) — the H.A.W.K `A`/`S`/`P` flag chooses the harness; the mechanic chooses the engine and hook category.
- **Reuse the latent domain vocabulary** (ability-catalog Key Decision) — a reflavored asset reaches for a built primitive before a new one.
- **The surface is visible in the data** (ability-catalog Key Decision; ACD-5) — H.A.W.K entries use the same `[ability]` / `effects` shapes as SWN.
- **One effect, one primitive, many harnesses** (ACD-4) — a new H.A.W.K mechanic that matches an SWN one on another surface reuses its primitive.
- **The catalog is binding mechanical input** (ACD-8) — the catalog's decomposition binds. The page now calls itself a frozen design artifact and sends as-built readers to the registrar tables (`effect/registrar.go`, `effects/registrars.go`, ability dispatch); it has no note for `A.001.7`'s tag effects. So the sort reads today's primitives from the registrar tables (D-8).
- **`_core` holds goals only; tags stay per rulebook** (AP-8) — H.A.W.K writes its own `tags.toml` in `A.001.7`'s data form and shares no tag record with SWN.
- **A rulebook `settings.toml` holds per-setting rule values** (AP-9) — H.A.W.K needs its own; `drift_costs` is required.
- **The canon is a living doc at `campaigns/hawk/canon/`, plain Markdown** (ACD-9 as amended by AP-10) — seeded by the interview, grown by authoring, ready for the Codex to take over later.

No architecture page covers the rulebook folder layout or the campaign folder as a whole; the closest are the ability catalog and `persistence.md`. `D.010` already notes that several engine pages lag the data-driven tags.

## Design Summary

Settled by the arc-plan and AP-10 before this session:

- **Effort 1 — lore seed.** An interview with Robert on H.A.W.K's setting, factions, terminology and the fiction behind its elements. Output: the seeded canon at `campaigns/hawk/canon/`. Its own session; no registry dependency.
- **Passes** — one file at a time. Each pass runs one loop:

  ```
  lore-gate → reflavor → mechanical sort → review
  ```

  - **Lore-gate** — confirm the batch's creative direction against the canon.
  - **Reflavor** — name, description and fiction per asset or tag; new fiction folds back into the canon.
  - **Mechanical sort** — per entry: reuse a primitive (pure data), parameterize a near-miss, or build a missing primitive.
  - **Review** — before commit.

  Output per asset pass: `rulebooks/hawk/assets/<category>_assets.toml` with `HAWK-` IDs, any new primitives, and canon updates. Rolls use `R.028`'s roll data shapes (`E.004`). The tag pass writes `rulebooks/hawk/tags.toml` in `A.001.7`'s data form.
- **ID seed** — `HAWK-` prefix, first mirroring SWN's tier suffix (`HAWK-F1-001`), drifting as the set diverges.

This session widened and reordered that. The initiative now also ships the H.A.W.K campaign, its map and its starting factions (D-1), set up early so each pass ends with a live turn (D-2). The tag pass moves ahead of the asset passes (D-3). The result is six efforts:

```
E1 lore → E2 tags → E3 campaign setup → E4 Force → E5 Cunning → E6 Wealth
```

## Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| D-1 | "H.A.W.K playable" means Robert opens a H.A.W.K campaign in the TUI and runs a turn on a H.A.W.K map; the initiative ships the campaign folder at `campaigns/hawk/` and its map, not only the rulebook | Robert's call; a rulebook that only loads would leave the arc's claim half-tested. Adds to the arc-plan boundary, which names assets and tags only |
| D-2 | The campaign, its map and its starting factions are set up right after the lore seed, before the asset passes; each pass then ends with a live turn in the TUI | Robert's call; a bad mechanical sort shows up in the pass that made it, not after all of them. The map and factions are fiction, so the lore interview supplies their facts and the setup turns them into data |
| D-3 | The tag pass moves up, between Effort 1 and the campaign setup | Robert's call. Factions get their tags at setup, and the loader checks that each `requires_tag` names a known tag (`internal/faction/rulebook/rulebook.go`), so H.A.W.K tags come first; no faction or gate is written twice. Amends the arc-plan order, which put tags last |
| D-4 | Six efforts, each its own squash boundary: E1 lore interview; E2 tag pass; E3 campaign setup; E4 Force, E5 Cunning, E6 Wealth | Follows from D-1 to D-3. Passes vary in size (the tag pass may need no new primitive, Force may need several), so one effort each keeps every review readable |
| D-5 | The canon is a few topic files, one `##` entry per entity with a stable ID, not one file per entity; the TOML `description` summarizes the canon entry and is never the only copy of an element's fiction | One interview should not seed dozens of tiny files. Stable-ID entries keep a later Codex import or a split one to one, without building the Codex shape now (the Codex seed is awareness, not authorization) |
| D-6 | E1 opens with `campaign create hawk --path campaigns` (no rules), then writes the canon into the new folder; E2 adds rules with `campaign add-rules` | `campaign create` refuses an existing target (`cmd/gm-toolkit/campaign/create.go`) and `Scaffold` a non-empty one (`internal/campaigns/scaffold.go`), so the campaign must exist before the canon does. The scaffold builds the folder tree, so nobody copies `Paths()` by hand. The registered campaign cannot run until E3, which is harmless while it is not set active. Rejected: E3 writes `campaign.toml` by hand and runs `campaign register` |
| D-7 | IDs: assets `HAWK-<F/C/W><min_rating>-<nnn>`, tags `HAWK-T-<nnn>`, table keys the PascalCase H.A.W.K name; numbered in authoring order within a tier; a cut number is retired, never reused; a committed ID never changes, even if `min_rating` does; the link to an SWN ancestor lives in the sort record, not the ID | The loader keys on `id` (`internal/faction/rulebook/rulebook.go`) and never parses the tier, which comes from `min_rating`, so drift costs no code. Saves, `requires_tag` and canon entries pin IDs, so stability wins. This is the arc seed's "mirror, then drift" made concrete. Rejected: IDs that keep SWN's slot numbers |
| D-8 | Each pass records its sort calls as a sort table in its effort plan, one row per asset or tag: HAWK ID, name, SWN ancestor, surface, call (reuse / parameterize / build), primitive, notes. The pass's Plan session reads the live vocabulary from the registrar tables as part of its re-grounding | The sort is what the pass plans to build, so it is the plan's core; it moves to `_completed/` with the plan for the arc close, and the plan's Decision Record catches changes made during execution. Rejected: a standing `rulebooks/hawk/SORT.md` (a dev log in the GM-facing folder, and a second copy) and TOML comments (unreadable in one go, and copied into every campaign) |
| D-9 | Each pass (E2, E4, E5, E6) runs: a Plan session (lore-gate, reflavor, sort table; fiction drafted in the plan; Robert's sign-off is the loop's review step), then Execution P1 for build rows only (primitive code and tests, green build, no data), then P2 (the TOML file replaces its stand-in, canon entries written, compose into `campaigns/hawk`, a live turn), then the effort close. E1 and E3 are fully specified in the overview plan and run straight from it | A primitive need appears only in its pass's sort, so it stays in that effort, ahead of the data that relies on it. **Departs from `efforts.md`** ("the single Discovery owns the whole design surface"): pass design needs the canon E1 writes, so it cannot happen here. The arc-plan already says A.001.5 does not map cleanly onto Discovery and Plan |
| D-10 | H.A.W.K's starting factions are a tracked seed state, `campaigns/hawk/state/state.seed.toml`, written in E3 from the canon (factions, stats, homeworlds, governments, starting assets). Each later pass swaps its category's starting assets in the seed and resets with `cmd/scripts/reset-campaign.sh hawk` before its live turn | A played-on save keeps stand-in asset IDs after a pass replaces them, and the engine handles a missing definition unevenly: the effects engine skips it (`effect/effect.go`), the movement phase fails the turn (`world/movement.go`). Only the seed is tracked, so it is also the one versioned record of H.A.W.K's opening position. Rejected: factions made in the TUI wizard, which leaves nothing tracked |
| D-11 | A tag–asset link is written by whichever pass creates its second half. E2 removes `requires_tag` from the stand-in copy of Gengineered Slaves, and a H.A.W.K tag that needs an `asset_definition` filter on a H.A.W.K asset leaves that effect out. The asset pass that creates the asset (E4–E6) adds both ends, and its sort table notes that it reopens `tags.toml` | `checkReferences` (`internal/faction/rulebook/rulebook.go`) refuses both directions at load. The SWN data has one link of each kind, Gengineered Slaves `requires_tag = "SWN-T-003"` and Eugenics Cult's `asset_definition = "SWN-F1-004"`, so E2's stand-ins would fail to load as first written. Found after sign-off by a side review. Rejected: pointing the stand-in at a H.A.W.K tag (edits throwaway data to fake a link) |

## Effort Breakdown

| Effort | Work | Output |
|--------|------|--------|
| E1 | Lore interview | The H.A.W.K campaign scaffolded first (D-6), then the seeded canon at `campaigns/hawk/canon/`, including the facts the map and the starting factions need |
| E2 | Tag pass | `rulebooks/hawk/` begins: `rulebook.toml` (includes `../_core`), the H.A.W.K `tags.toml`, and copies of SWN's three asset files (with `requires_tag` removed, D-11) and `settings.toml` as stand-ins, so the folder composes and loads; composed into `campaigns/hawk` |
| E3 | Campaign setup | The map, the starting factions as a seed state (D-10), and the real `settings.toml` values (the population floor depends on the map's populations); the first live turn |
| E4 | Force pass | `assets/force_assets.toml` replaces its stand-in; a live turn |
| E5 | Cunning pass | `assets/cunning_assets.toml` replaces its stand-in; a live turn |
| E6 | Wealth pass | `assets/wealth_assets.toml` replaces the last stand-in; "H.A.W.K playable" is reached |

## Canon Layout

```
campaigns/hawk/canon/
  README.md        index; how a pass folds fiction back in
  setting.md       premise, era, tone, tech
  glossary.md      H.A.W.K terms
  factions.md      one entry per faction
  places.md        sectors, regions, worlds (feeds the E3 map)
  people.md        characters, if the interview turns any up
  tags.md          tag fiction, one entry per HAWK- tag ID
  assets.md        asset fiction, one entry per HAWK- asset ID
```

The file names follow the Codex seed's entity kinds (factions, locations, characters, lore terms) where they overlap. `assets.md` splits by category only if it grows past easy reading.

## Campaign and Rulebook Mechanics

- **No spatial change** (D-1) — the map uses today's source shape, `spatial/source/layout.txt` plus `data.toml`, built by `gm-toolkit spatial build`, as `campaigns/sandbox/` does. The arc's out-of-arc line on `spatial/` covers the package, not a campaign's map data.
- **Stand-ins need no code** (D-2, D-4) — `ComposeRulebook` (`internal/campaigns/scaffold.go`) copies file by file, and `add-rules --replace` overwrites a file of the same name. So `rulebooks/hawk/` starts with copies of SWN's three asset files and `settings.toml`, and each later effort overwrites one. An include of `../swn` would not work: once `rulebooks/hawk/` supplies its own file, the compose refuses the duplicate (`ErrRulebookDuplicate`).
- **The `SWN-` prefix is not load-bearing** — no production Go reads it; only `testharness` pins SWN IDs. H.A.W.K assets and SWN stand-ins live side by side in one campaign without trouble.

## Out of Scope

- **New goals** — H.A.W.K composes the shared `_core` goals as they are (arc out-of-arc list).
- **`internal/spatial` changes** — the map is data in today's shape (D-1).
- **Codex work** — the canon is plain Markdown that the Codex can take over later; no Codex shape is built now (D-5).
- **`F.041`'s encounter data** — `campaigns/hawk/encounters/` belongs to `F.041` and its own branch.
- **A uniform missing-definition rule in the engine** — the effects engine skips an unknown definition and the movement phase fails the turn; D-10's seed resets work around it. Fixing it is a candidate Backlog row, not this initiative.
- **H.A.W.K scenario files** beyond the seed state — added later if play-testing asks for them.

## Follow-ons

Bookkeeping this discovery calls for, outside this doc:

- **Arc-plan amendment (AP-11)** — D-1 widens A.001.5's boundary to the campaign, map and starting factions; D-3 reorders its tag pass ahead of the asset passes; D-9 records that its passes design in their own Plan sessions.
- **Tracker write-up** — the A.001.5 row and write-up in `dev_journals/faction-manager/planned-work.md` still describe the old order and scope.
- **`B.021` count** — its row says "all 72 assets"; the SWN files hold 73, because `A.001.7` Effort 5 added Gengineered Slaves after the diff. The row gets a note that its audit does not cover that asset.

## Open Questions

None. The eleven questions raised in this session are resolved as D-1 to D-10; D-11 closes a load failure found after sign-off.
