# Data-Driven Tags — Retro

## In Plain Words

- **Tags run from rulebook data now, as assets do.** A tag in `tags.toml` carries the
  same `[[effects]]` list as an asset, and the engine loads, checks, registers and
  fires it the same way. The old tag engine and its four Go handlers are gone.
- **All 20 SWN tags work in play,** where four did before. The one gap is Psychic
  Academy's data entry, which waits for the roll card (`R.025`); its engine part is
  built.
- **Factions hold tag IDs, not copies,** and the IDs carry the rulebook prefix
  (`SWN-T-020`). A rulebook change now reaches a tag a faction already holds. The
  government role is a `government = true` flag in the data.
- **New rules the tags needed:** a `settings.toml` per setting (drift costs moved in),
  a population floor per asset type (Starships need 500,000 people) with a waiver,
  tags that grant an action (Teach Planetary Population), an asset only a tag holder
  can buy (Gengineered Slaves), and a toll for landing on a Pirates world.
- **Rolls reach further.** The three bare contests (Tripwire Cells, Panopticon
  Matrix, Lobbyists) roll through the hook path, and a roll asks the opponent's tags
  too. Informers and Seductress roll as attacks, as the book words them.
- **Scavengers pays only for its own fights** (`B.042`), and counts a lost Base.
- **Refit obeys the world's purchase rules** (tech level and population), as a
  purchase does.

## 1 — Did we do what we set out to do?

Yes. The discovery's In Plain Words promised all 20 SWN tags in data, the four Go
handlers gone, and new engine parts for the tags that needed them. All of that
landed.

- **Landed as designed:** tag IDs held by factions (D-20); `[[effects]]` on the tag
  record; `engine/tag` deleted (D-14); the four handlers as data, with `B.042`'s owner
  filter (D-2); the government flag; `settings.toml` (D-10); Colonists, Technical
  Expertise and the waiver (D-8, D-9); Pirates (D-6); Deep Rooted's loss (D-19);
  Exchange Consulate's XP roll (D-18); the `R.023` contests (D-16); Teach Planetary
  Population (D-11); Gengineered Slaves and `requires_tag` (D-13).
- **Deferred, as planned:** Psychic Academy's `[[effects]]` entry (D-7) waits for
  `R.025`. Contests that ask the GM wait for a follow-on (overview Decision 2).
- **Shipped with a known defect:** the `rival_roll` registrar confirms an elective
  reroll through the rival's collector, not the Academy owner's (Effort 3 Decision
  Record 3). Nothing plays wrong today, because no data entry uses it.
- **Added beyond the plan:** Informers and Seductress roll as attacks (Effort 5); the
  Expand Influence free attack asks the tie resolvers (Effort 1 Decision Record 5); a
  world's population is required in the map source (Effort 4 review); the goal line
  in the digest shows Exchange Consulate's bonus XP (Effort 3 Decision Record 4).
  At this close: Refit asks `CanPurchase`, which closes its older missing TL check;
  Scavengers counts self-damage deaths and Bases (Robert's rules call).
- **Reshaped on the way:** effects became blind to their holder (R-1), wider than the
  discovery's "each registrar names its holders"; Mercenary Group became a speed
  floor, not a self-move (R-4); the Starship rule became a population floor per asset
  type (R-6); the discovery's four efforts became five (R-5).

## 2 — Is there a better way we didn't see?

The code design holds. The cost sat in the planning.

- **The design.** R-1 (effects blind to their holder, scope decides the holder) was
  itself the better way, found in the Effort 1 Plan session before any code. It
  turned "teach each registrar about tags" into one `effect.Holder` and a ~60-line
  `holder.go`. Only `grant_movement` and `ignore_coin_loss` still refuse a tag. A
  separate tag engine with its own registrars, the shape the discovery started from,
  would have doubled the vocabulary. Nothing on the branch argues for going back.
- **The plans.** The five effort plans run to about 17,500 lines, for about 3,000
  lines of production Go. They specced literal code, and each execution still had
  to re-ground and log anchor drift (Effort 5 Decision Record, Commit 5). A plan
  that fixes decisions and file seams, and lets execution write the code, would have
  cost a fraction of the planning time. What it would have lost: the decisions that
  surfaced while writing literal code (R-1, R-4 and R-6 all came out of Plan
  sessions reading the code closely). So the depth paid for itself in decisions,
  not in the code listings.
- **The slicing.** The discovery's Effort 3 held about 14 pieces of work and had to
  split (R-5). A size check at the overview plan, against the phase budget, would
  have caught it a session sooner. Small cost either way.
- **The architecture pages.** Five efforts changed binding Key Decisions (asset
  scope, the opponent's faction bucket, the `T-011` tag), and no page changed on the
  branch. The pages now say the opposite of the code in two places until this close
  graduates them. Graduating at each effort close would have kept them true, at the
  price of a doc edit per effort.

## 3 — Does it fit the project?

Mostly yes. The divergences are deliberate and recorded, but not yet in the
architecture pages.

- **Matches effect-mutation:** an effect is data keyed on its name and the registrar
  picks the hook category; a tag never names a category. A known effect with no
  registrar fails loud at engine construction, for a tag entry as for an asset. The
  rule layer registers and never dispatches.
- **Matches the data-driven rule** (Robert's standing call, overview Decision 7):
  each new effect takes its numbers and references from its entry. `population_floor`
  names no asset type in Go (R-6); `xp_roll` reads its die, threshold and goal from
  the entry; the arrival toll reads its amount from the entry, not copying
  `lock_out`'s hard-coded 1d4.
- **Matches persistence:** the strict tag loader takes the new keys; references (goal
  IDs, `requires_tag`) are checked at load; saves moved by a re-seed, with no
  migration code (D-4).
- **Diverges from hooks, on purpose:** "every S effect registers at asset scope" is
  now "scope decides the holder; a tag registers at faction scope". "The roll
  dispatcher never reads the opponent's faction bucket" is reversed: `rollScopes`
  reads the opponent's faction bucket and every bystander's, and each hook checks its
  owner relation itself. Both pages are stale until step 8 of this close.
- **Diverges from persistence, on purpose:** the government record's Key Decision
  names "the `T-011` tag"; the role is now the `government` flag, and Colonists can
  write a record entry at creation (`homeworld_government`).
- **Arc decisions:** ACD-3 and ACD-5 hold: the S side is a profile that registers
  hooks, and a tag's surface is visible in its TOML (`[[effects]]`, `[ability]`).
  AP-9 overrode the arc plan's "punt tags that grant actions or assets", recorded in
  the arc plan.
- **Concurrent work on the branch:** six devcontainer and shell commits and
  `cmd/scripts/hawk_finale_sim.py` (a H.A.W.K boss-fight simulator) ride on this
  branch. They are not tags work, but work running at the same time needed them, so
  they stay. The Effort 1 and 2 plan files predate the `<ID>.E<n>-` naming and keep
  their names.

## 4 — What does it change about the project?

- **A tag needs no Go when existing effects can say its rule.** `A.001.5`'s H.A.W.K
  tag pass is unblocked.
- **The faction holder is ready for goals.** `A.001.8` can reuse `effect.Holder` and
  the roll filters (`goal`, `homeworld`, `opponent_attribute`).
- **Every roll hook must check who it serves.** A roll now reads every faction's
  faction bucket, so a faction-scope roll hook that skips `rollMatches` would fire on
  other factions' rolls. New roll registrars inherit that duty.
- **Contests ask no one, for now.** Under the nil-collector rule, the three contests
  take free dice and forced rerolls only. Machiavellian's once-per-turn die and Book
  of Secrets' reroll do not reach them until the follow-on lands.
- **GM data got stricter.** A world needs `population` in the map source; every
  ability contest needs `roll`; the rulebook needs `settings.toml` in place of
  `drift_costs.toml`. Older campaigns need a re-compose.
- **Per-setting rule values have a home.** `settings.toml` is where the next setting
  constant goes, not a new file.
- **Docs out of date until this close:** hooks (asset scope, the opponent's bucket,
  the nil-collector roll), effect-mutation (tag engine, the faction holder),
  persistence (`T-011`, drift costs, tag IDs), narrative-digest (tag names from the
  rulebook).

## 5 — Rating

| Metric | Score | Evidence |
|---|---|---|
| Delivered | 4 | All 20 tags in data as promised; `rival_roll` shipped with a known wrong-player defect |
| Plan held | 3 | Six overview reversals (R-1 to R-6) plus Effort 5's Decision 4; all made before or at review, none mid-commit |
| Rework | 3 | A fix pass at every effort close, none large; this close's review found 10, of which 8 were fixed: one real rule bypass (Refit), one silent misconfiguration (asset roll scopes) |
| Fit | 4 | Follows the effect model and Decision 7; two hooks Key Decisions reversed on purpose but left stale on the pages |
| Simplicity | 4 | One `effect.Holder` and one `holder.go` carry the whole tag side; 41 registrar files is the vocabulary's real size |
| **Total** | **18 / 25** | Mean 3.6: above "as expected", held back by plan reversals and rework |

**Verdict:** Shipped what it promised, on a design simpler than the one it started
from; the cost was in plan length and in architecture pages left stale until the
close.

**What surprised me:** the biggest design win (effects blind to their holder) came
from a Plan session reading 22 registrars, not from the discovery.

## Follow-ups

| ID | Item | Trigger | Detail |
|---|---|---|---|
| `F.037` | Contests that ask | after `R.025` | Tripwire Cells, Panopticon Matrix and Lobbyists roll with a nil collector (`A.001.7` overview Decision 2), so limited dice and elective rerolls never reach them; give them the roll card. Decide who sees it for a roll that fires after a batch or on another faction's turn. Lobbyists first: its call site already prompts |
| `R.027` | Registrar `faction`-scope sweep | -- | Audit every registrar for a `faction` scope that a tag would need; land each with the tag that uses it (`A.001.7` overview R-3). `grant_movement` and `ignore_coin_loss` still refuse a tag holder |
| `B.043` | `rival_roll` asks the wrong player | with `R.025` | `RollWithHooks` confirms an elective reroll through the roller's collector, so Psychic Academy's reroll on a rival is confirmed by the rival; a once-per-turn gate would spend on the first rival roll. Settle before the Academy's entry lands (`A.001.7` Effort 3 Decision Record 3) |
| `B.044` | World TL missing from map source reads as 0 | -- | The spatial builder now refuses a world with no `population` (`A.001.7` Effort 4 review), but a missing tech level still reads as 0, the gap that population had ("TL keeps its old gap"); check what it blocks and refuse it the same way |
| `D.010` | Update architecture pages for data-driven tags | -- | The `A.001.7` close graduates Key Decisions only; the page bodies, diagrams and Dependencies still describe the tag engine, Go tag handlers, `drift_costs.toml` and the `T-011` constant. Refresh effect-mutation, hooks, persistence, goals and orchestrator against the source, and check narrative-digest, world-movement, actions and ability-catalog for tag holders, `settings.toml`, the toll, the speed floor and tag abilities |
| `B.045` | Refused relocate still counts its toll | -- | Relocate's total check adds every order's tolls before the government is asked, so a faction can be refused for Coin on an order the government would refuse anyway; `totalCost` does the same. Left as is in `A.001.7` Effort 4 |
