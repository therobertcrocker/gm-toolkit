# H.A.W.K Asset Authoring — Retro

## In Plain Words

- **H.A.W.K is playable.** Robert opens the `hawk` campaign in the TUI and plays a
  turn on the Core Sector map. Eight factions hold H.A.W.K tags and H.A.W.K assets,
  and pick from the shared `_core` goals. No SWN record is left in `rulebooks/hawk/`.
- **H.A.W.K has a written canon** at `campaigns/hawk/canon/`: the campaign, the
  setting, a glossary, places, factions, and an entry for every tag and asset. It
  came from two lore interviews, and each later pass added to it.
- **The rulebook:** 16 tags and 84 assets (29 Force, 28 Cunning, 27 Wealth), with
  eight H.A.W.K asset types such as Aethership, Front and Workforce. The map traces
  the canon's grid image: 277 hexes, 19 regions, 16 places and 17 Drift lanes.
- **Travel works differently, in every rulebook.** A Drift jump now costs movement
  points by the asset's drift rating, and points left over carry to the next turn.
  Ships move in short, mid and long speed classes. The TUI says where an asset in
  transit is. Aether Burners can run hot: faster moves that burn the hull.
- **New rules any rulebook can use:** asset types declared in `settings.toml`; tag
  upkeep that lapses, removes or damages when unpaid; aiming a die at a faction by
  its tag; a drift rating bonus; Coin by count, per hit or once a turn; teleports
  that ignore permission and lockdowns; a sacrifice limited to some types; a
  hiding-place ability; and a public teleport network that rivals ride for a fare.
- **Not done here:** the seed is a test board, to be regenerated for real play
  after the arc closes, and the numbers wait for play to tune them.

## 1 — Did we do what we set out to do?

Yes, and more. The discovery's In Plain Words promised a canon, H.A.W.K tags and
assets mapped onto the data-driven engine, and a real campaign with a live turn at
the end of each pass. All six efforts landed in the order D-4 set, and each of
Efforts 3 to 6 ended with a played turn.

- **Landed as designed:** the campaign scaffolded before the canon (D-6); the canon
  as topic files with ID headings (D-5, overview Decision 3); `HAWK-` IDs (D-7);
  a sort table per pass (D-8); the pass shape, lore-gate → reflavor → sort → build →
  data (D-9); the tracked seed state and the resets (D-10); tag–asset links written
  by the second pass (D-11; Arcano-Engineered Warbeasts to Aether-Tech Evangelists).
- **Bigger than the discovery expected:** every pass needed engine work. The
  discovery thought the tag pass "may need no new primitive"; it built six (Effort 2
  Decisions 3 and 6 to 10). Effort 4 rebuilt travel, because measuring the map showed
  that `drift_costs` and drift ratings only picked a route and the path's cost was
  thrown away (Effort 4 Decision 18). That is a fix to the engine, not H.A.W.K data.
- **Added beyond the plans:** asset types as rulebook data (Effort 2 Decision 3);
  running hot (Effort 4 Decision 24); a held jump keeps its cost (Effort 4 Decision
  Record 6); the engine decides when an order runs hot (Decision Record 7); the door
  shelter's release rules (Effort 5 Decision Records 4 and 5); Carrier Corps, a
  Force carrier (Effort 6 Decision 7); a hidden public grant serves only its owner
  (Effort 6 Decision Record 6). At this close: an occupied door shelter cannot move
  or ride, healing never frees an asset from one, and a shelter ability needs its
  loss entry at load (Effort 5 Decision Records 6 to 8); a hot ship does not burn
  while it waits at an edge (Robert's rules call; Effort 4 Decision Record 10); the
  "In the Drift" label needs a lane jump next (Decision Record 11); the counting
  `coin_gain` params are checked at load (Effort 6 Decision Record 10); and
  `hawk-claude` is the registry's active campaign (Effort 3 Decision Record 5).
- **Cut or deferred:** people in the canon (overview Decision Record 17); the
  Material Plane map (`F.045`, `F.046`); fighting in transit (`F.052`); an embargo
  on the public mover (`F.054`); teleport range by lanes (`F.055`); the seed for real
  play (Effort 3 Decision 15); number tuning (every pass).
- **Shipped with known faults, filed:** the goal pick cuts descriptions (`B.048`);
  the movement form loses its title and its Hold option (`B.050`, `B.051`); a
  self-jump's empty pick spends the turn (`F.024`); feed lines name no owner
  (`B.053`). None comes from H.A.W.K data; H.A.W.K's bigger map and new rows showed
  them.

## 2 — Is there a better way we didn't see?

The pass shape holds. Two findings came one effort later than they had to.

- **The pass shape.** A Plan session per pass that drafts the fiction, sorts each
  row and signs off, then a build phase before the data, kept every primitive in the
  effort that needed it. A single up-front design across all four passes could not
  have worked, because each pass designed from the canon that Effort 1 wrote and
  from what the last live turn showed. D-9's departure from `efforts.md` was right.
- **Travel should have been measured in Effort 3.** Effort 3 drew the map and set
  `drift_costs = [6, 5, 4, 3, 2]` on the claim that "a crossing always costs more
  than a step". Nothing in the engine made that true, and Effort 4's Plan session
  found it by measuring trips on the map. A trip-length check in Effort 3, with the
  throwaway programs Effort 4 wrote (`scratch/pacecheck/` and the others), would have
  moved the travel engine into the effort that owned the board and kept Effort 4 to
  Force.
- **Teleport range was set on a measure that ignores lanes.** Effort 4 picked Hubs
  8 and Gate 14 by counting hexes with lanes free, and Effort 6's live turn found
  the Transit Authority reaching 12 of 16 places at 12 hexes (Effort 6 Decision
  Record 5). `F.055` now holds Robert's model. The same measure was in front of
  Effort 4; a reach count per teleport, as Effort 6 did, would have shown it.
- **Plan size.** About 12,300 lines of plan for about 1,600 lines of production Go,
  but the plans also carry about 5,500 lines of literal TOML and the canon drafts.
  Effort 2's 3,500 lines for 16 tags and six builds was the heaviest; later passes
  were leaner per row. Fine for a content initiative, where the data is the work.
- **Stand-ins of SWN types** (Effort 4 Decision 11) cost one cancelled live turn
  (Effort 4 Decision Record 5). Adding H.A.W.K types to the throwaway copies would
  have been edits to data due for deletion; the choice still holds.

## 3 — Does it fit the project?

Yes. It is the data-driven rule applied to a second rulebook, and it found the
places where the engine still assumed SWN. Two Key Decisions now read wrong until
this close graduates the new ones.

- **Matches extensible static data:** every H.A.W.K rule is TOML. The asset type
  names left Go for `settings.toml`, so no Go names a type, and the loader checks
  every name the data uses. This is the effect-mutation decision "a rule value or
  reference lives in the entry, never in engine code".
- **Matches the arc:** ACD-1 (the mechanic picks the engine: the shelter is an
  ability on the A side and a condition on the S side); ACD-4 (teleports reuse the
  one `relocate` primitive with two new switches; the door shelter reuses the
  `sheltered` condition); ACD-5 (every new block names its effect); AP-8 (H.A.W.K's
  own tags, `_core` goals only); AP-9 (`settings.toml` holds per-setting values).
- **Matches mutations and history:** `TagLapsed` and the fare's `CoinDelta` are
  mutations with apply arms, digest lines and feed lines (Effort 2 Decision Record
  3; Effort 6 Decision Record 7).
- **Diverges, on purpose:** hooks' "a grant never leaks to another faction" now has
  a public grant that walks every faction's buckets (Effort 6 Decision 10).
  World-movement's "only Deep Strike Landers skip lockdown" now has relocates that
  skip it too (Effort 4 Decision 7). Drift jumps cost points in SWN as well, so the
  sandbox moves slower (Effort 4 Decision 18). The discovery's "no `internal/spatial`
  changes" gained two read-only helpers (Effort 4 Decision 26).
- **Diverges in process:** D-9 departs from `efforts.md`'s single Discovery; overview
  Decision 1 gave Effort 3 its own Plan session; Claude played Effort 5's live turn
  on `hawk-claude` at Robert's request (Effort 5 Decision Record 1). Each is recorded.

## 4 — What does it change about the project?

- **The engine has run a non-reference rulebook end to end.** The arc's claim is
  tested: a rulebook with its own types, tags, assets, settings and map loads,
  composes and plays with no SWN record in it. WWN is next in line.
- **A rulebook must declare `asset_types`.** Any new rulebook, and any campaign
  copy made before Effort 2, needs the list, or it fails to load.
- **Tags can cost Coin.** Upkeep runs in bookkeeping after income and before asset
  maintenance, and a lapsed tag goes dark but keeps charging. A new tag effect must
  respect the lapse, which the holder check already does for hooks.
- **Travel is a point budget.** Any work on movement (fighting in transit,
  multi-map, lane-count range) starts from banked points, in-the-Drift positions
  and the hold rules of Effort 4 Decision Record 6.
- **Grants can be public.** A movement grant can serve rivals, so code that reads
  `GrantedMovementAbilities` must expect a source owned by another faction.
- **Campaign data is under test.** `TestHawkSeed_MatchesRulebookAndMap` checks every
  seeded definition against the H.A.W.K rulebook, so a rulebook edit that strands a
  seeded asset fails the gate.
- **The canon is the Codex's first content.** Plain Markdown with stable IDs and
  `GM-only` callouts, ready for the Codex seed to pick up.
- **Docs out of date until this close:** world-movement (point costs, banking,
  in-transit labels, holds, running hot, relocate switches), hooks (the drift-rating
  query, the upkeep trigger, `opponent_tag`, public grants), turn-pipeline (upkeep
  in bookkeeping), effect-mutation (`TagLapsed`), and persistence (`asset_types`).
  `D.010` already holds the page-body lag from `A.001.7`.

## 5 — Rating

| Metric | Score | Evidence |
|---|---|---|
| Delivered | 5 | All six efforts, a played turn after each from Effort 3, and a travel engine fix the discovery did not plan; the faults found are older UI bugs, filed |
| Plan held | 3 | Many reversals, most from live turns by design: Effort 3's drift costs, overview Decision 4 for Effort 5, Effort 4 Decision 17, Effort 6's range, and four review reversals of commit tasks |
| Rework | 3 | A fix pass at every effort close (about 1,000 lines over six); this close found 10 and fixed 9, among them a real bug in shipped data (an occupied Demiplane could move or ride as cargo, left behind by Effort 6's Front cargo) and a hot ship burning while parked |
| Fit | 4 | Data-driven and mutation rules hold throughout; two Key Decisions reversed on purpose and stale until this close; the travel change reaches SWN |
| Simplicity | 4 | About 1,600 lines of Go for a dozen new effects and switches across 100 entries; running hot and the door shelter each needed a review pass to reach their simpler shape |
| **Total** | **19 / 25** | Mean 3.8: full delivery and a sound fit, held back by a plan that bent at almost every effort and review rework around the door shelter |

**Verdict:** Reached "H.A.W.K playable" with a full canon, 100 rulebook entries and
a real travel engine; the live turn after every pass caught what reviews alone
would have missed, at the cost of a plan that bent at almost every effort.

**What surprised me:** the second rulebook's biggest finding was not about
H.A.W.K at all: SWN's Drift costs had never cost anything.

## Follow-ups

| ID | Item | Trigger | Detail |
|---|---|---|---|
| -- (not picked) | Regenerate the H.A.W.K seed for real play | after the `A.001` arc close | The seed is a test board of hand-picked starting assets and two extras (Effort 3 Decision 15, Effort 5 Decision 17, Effort 6 Decision 18); two ships sit below the ship floor (Effort 3 Decision Record 2). Build the real opening position from the canon |
| `B.054` | Tag lapse not read by three rules | when a lapsing tag meets one | `opponent_tag` does not check a lapse; `turn_start` fires before upkeep, so it reads last turn's lapse; `CanPurchase` ignores a lapse for `requires_tag` (a rules call). No H.A.W.K data hits them today (Effort 2 Decision Record 4) |
| `B.055` | Upkeep misses a tag gained this cycle | -- | Upkeep reads hooks registered at cycle start, so a tag gained earlier in the same cycle is not charged until the next one. Reading held tags in bookkeeping closes it for upkeep (Effort 2 Decision Record 4) |
| `R.032` | Fare label read from the grant, not the definition | when a definition has two relocate sources | `relocateVia` (`internal/faction/tui/adapter/action_collector.go`) rebuilds the fare from the definition, so a definition with two relocate sources could label one fare and charge another (Effort 6 Decision Record 8) |
| `R.033` | One resolver for speed and drift rating | -- | `AssetDriftRatingResolver` repeats `AssetSpeedResolver` (Effort 2 Decision Record 4) |
| `R.034` | The shelter release asks through a payment | -- | The turn-start release (`turn.OfferShelterReleases`) asks `ConfirmPayment` with Amount 0 and `PaymentReleaseShelter`, so the TUI special-cases its banner and confirm text, and every collector (scripted, smoke, mock) must know a 0-Coin payment is a release. Give it its own confirm ask (`A.001.5` close review, not picked) |
| `C.003` | Sacrifice text says "false front" | -- | The TUI label and the narrative line "A false front fell in place of the asset it covered" (`internal/faction/narrative/wire_templates.go`) show for Unpaid Interns and Glamoured Doubles too; word it from the asset (Effort 5 Out of Scope) |
| `F.058` | Upkeep line on the bookkeeping card | when play asks | Tag upkeep reaches history and the feed but has no line of its own on the bookkeeping card (Effort 2 Out of Scope) |
| `B.056` | One rule for a missing asset definition | -- | The effects engine skips an unknown definition and the movement phase fails the turn (`effect/effect.go`, `world/movement.go`); pick one rule (discovery Out of Scope) |
| `D.010` (amend) | Page bodies for H.A.W.K's engine work | with `D.010` | This close graduates Key Decisions only; the page bodies of world-movement, hooks, turn-pipeline, effect-mutation and persistence still describe the old travel tick, no upkeep and fixed asset types |
