# Data-Driven Rulebook (H.A.W.K) — Arc Retro

## In Plain Words

- **The engine runs more than one rulebook.** SWN and H.A.W.K both load, compose
  and play. A rulebook is a folder of TOML: a `rulebook.toml` manifest, a
  `settings.toml`, its tags, its assets and its map. Goals come from a shared
  `rulebooks/_core/` that every rulebook includes.
- **Behavior comes from data, not from IDs.** Every asset ability, passive effect,
  tag and goal states what it does in its own entry. No Go code looks up an asset,
  tag or goal by its ID. A new rule that the existing effects can say needs no Go.
- **H.A.W.K is playable.** Robert opens the `hawk` campaign in the TUI and plays a
  turn on the Core Sector map: 16 tags, 84 assets, eight factions, and a written
  canon at `campaigns/hawk/canon/`.
- **SWN got more complete on the way.** All 20 SWN tags and every SWN asset special
  now work in play, where four tags and one ability did before. Travel costs
  movement points in every rulebook.
- **Taken out, for now:** the Seize Planet and Change Homeworld actions (back with
  `F.040`), and Psychic Academy's data entry (waits for `R.025`).

## 1 — Did we do what we set out to do?

Yes, with a wider boundary than the arc-plan first drew. The concept's vision had
three parts: every `A`- and `S`-flag special resolves from data, goals and tags are
campaign-agnostic, and H.A.W.K ships as a full playable rulebook. All three hold.

- **Landed as the concept framed it:** effect-keyed dispatch and the SWN re-prefix
  (`A.001.1`); the A-side composition pipeline and the one `relocate` primitive
  (`A.001.2`); the S-side trigger/filter/gate/effect model (`A.001.3`); the shared
  `_core` home (`A.001.4`); H.A.W.K authored with a living canon (`A.001.5`).
- **Grown past the concept:** the arc started at five candidates and shipped eight.
  `A.001.6` split off the S side once the inventory showed 33 unbuilt assets
  (AP-7). `A.001.7` and `A.001.8` made tags and goals data, where the concept only
  asked to neutralize their text (AP-8). `A.001.5` shipped a campaign, a map and a
  travel engine fix, not only a rulebook (AP-11).
- **Reshaped:** "goals and tags share one source home" (ACD-6, ACD-7) became goals
  only. Tags are setting identity, and once tags are data no H.A.W.K tag needs to
  share a record with SWN. The canon moved from `docs/campaigns/hawk/` to
  `campaigns/hawk/canon/` (AP-10).
- **Removed or deferred:** Seize Planet and Change Homeworld left the engine with
  the goal process they rode on (`F.040`); Psychic Academy's entry waits for the
  roll card (`R.025`); the H.A.W.K seed is a test board, and the numbers wait for
  play.
- **At this close:** the review's ten findings were all fixed. The biggest:
  Dreadnought and Capital Fleet carried the `A` flag with no ability, so the
  action menu offered a prompt that did nothing. Both lost the flag; Capital
  Fleet's "move as an action" is its Speed, as `A.001.2` Decision 5 read it,
  which reverses `A.001.6` Decision Record 7 (Robert's call). The loader now
  holds the `A` flag and an `[ability]` block together and refuses a step with no
  effect, and the do-nothing "confirm applied" ask is gone end to end. The rest:
  the catalog's Key Decision now states ACD-1, `hooks/doc.go` lists the real Cat 4
  family, the never-registered `MaintenanceCostModifier` is deleted, the TUI
  ability targets carry the holder, a tag's prose key is `rules_text`, and the
  tracker carries the stale-page list (`D.010`) and the dead H.A.W.K triggers
  (`F.040`).

## 2 — Is there a better way we didn't see?

The spine holds. Three things showed up later than they had to.

- **The spine.** ACD-1 to ACD-5 survived eight initiatives without a reversal. The
  one design that beat the arc's own plan was `A.001.7`'s "effects are blind to
  their holder" (R-1): one `effect.Holder` let tags reuse the asset registrars, and
  goals then reused the filters. The arc-discovery could not have seen it; it came
  from reading 22 registrars.
- **Inventory before slicing.** The arc-plan sized the S side by generalizing
  `[transport]`, and the `A.001.3` Discovery had to re-slice it after counting the
  data (AP-7). The same arc-plan session could have counted the 43 `S`-flag assets.
  It would have cost an hour and saved an amendment, not an initiative: the split
  was right either way.
- **Shared tags.** ACD-6 assumed shared goals and tags. `E.002` showed that data-driven
  tags remove the reason to share them, and `A.001.4` shrank to goals. Had the
  arc-discovery asked "what does sharing buy once tags are data?", `A.001.7` and
  `A.001.8` would have been in the first arc-plan. The cost was one exploration and
  one amendment; nothing was built twice.
- **The second rulebook found the first rulebook's bugs.** SWN's drift costs never
  cost anything, and only H.A.W.K's map showed it (`A.001.5` Effort 4). That is the
  arc's thesis working as intended, but it argues for an earlier, smaller non-SWN
  test: a throwaway rulebook in `A.001.4` would have found it months sooner.
- **Architecture pages lagged every close.** `A.001.7`, `A.001.8` and `A.001.5` each
  graduated Key Decisions and left page bodies stale, so `D.010` grew at each close.
  Updating each page body in the effort that changed it would have kept the wiki true
  for the next Discovery, at a doc edit per effort.
- **Plan weight came down.** `A.001.7` wrote about 17,500 lines of plan for 3,000
  lines of Go; `A.001.8`'s just-in-time effort plans cut that to a third. The arc
  learned this one inside itself.

## 3 — Does it fit the project?

Yes. The arc is where "data-driven over bespoke" became the project's rule, so it
sets the fit more than it follows it.

- **Matches the arc's own decisions:** ACD-1 (the mechanic picks the engine:
  Marketers and Seditionists went S-side, AP-6; the H.A.W.K shelter lives on both
  sides); ACD-2 and ACD-3 (both surfaces keyed on the effect, not the ID); ACD-4
  (one `relocate` primitive under the ability, the movement grant, teleports and
  the public mover); ACD-5 (every block names its effect; `[transport]` stays as the
  hybrid allows, `R.019` folds it in).
- **Matches persistence:** GMs own all data; the binary embeds none; compose copies
  `_core` and the rulebook into the campaign, so the campaign stays self-contained.
  Loaders are strict, references are checked at load, and saves moved by re-seed
  with no migration code.
- **Matches effect-mutation:** the rule layer registers and never dispatches; goals
  emit, never mutate; every new record (`TagLapsed`, `GoalPicked`, `GoalFailed`,
  `ActionTaken`) has an apply arm and a digest line.
- **Narrowed on purpose, recorded:** ACD-6 and ACD-7 to goals (AP-8); ACD-9's path
  (AP-10); the `_core`-holds-goals-only line that `F.047` may revisit.
- **Diverges on purpose, recorded in the initiative retros:** hooks' "a grant never
  leaks" (public grants), the roll dispatcher now reads the opponent's bucket, and
  travel changed for SWN too.
- **Out of step:** the architecture front door. `architecture-overview.md` and
  `engine/overview.md` still call the engine "the SWN faction-turn system", and they
  and `orchestrator.md` count seven sub-engines. `A.001.7` deleted the tag engine;
  `engine.Engine` holds six (turn, effect, mutation, action, goal, world).

## 4 — What does it change about the project?

- **WWN is a data job, mostly.** A new rulebook is a folder: manifest, settings with
  `asset_types`, tags, assets, map. Go work is needed only for a rule the effect,
  ability and goal vocabularies cannot say.
- **The vocabulary is the new surface.** About 40 effect registrars, the ability
  pipeline's effects and the goal tables (`events.go`, `measures.go`, `xp.go`) are
  what a rulebook author reads. A new rule adds a vocabulary entry once; every
  rulebook then gets it as data. `ability-catalog.md` and `docs/contributing/engine.md`
  are where an author starts.
- **SWN is the reference, not the default.** No SWN ID is in Go, and H.A.W.K has no
  SWN record. A change that assumes SWN content now fails a test
  (`TestHawkSeed_MatchesRulebookAndMap`) or a load check.
- **Re-seed is the save policy.** Every ID change in the arc moved saves by
  re-composing and re-seeding. A campaign with real play history will need a
  different answer the first time an ID changes under it.
- **SWN plays with gaps until `F.040`.** No seizure, no homeworld change, and Deep
  Rooted has no trigger.
- **`R.004` is met and closed.** Its ask, a typed primitive registry in place of
  hard-coded dispatch, is what `A.001.1` to `A.001.3` built, and `A.001.8` did
  the goal half.
- **The SWN flags are now mostly repetition.** `A` repeats the `[ability]` block
  (the loader now holds them together), `S` is read by no code, and `P` is a real
  rule behind a letter, beside `purchase_permission`. They are SWN's notes for a
  person at the table, and the data now says what they hinted at.
- **Docs to fix:** `D.010` now lists every stale page body, the front door and
  the contributing guide included.
- **Process gap, unchanged:** the arc ran a content initiative (`A.001.5`) and a
  lore interview with no session type for either; AP-5 and A.001.5 D-9 worked
  around it. It stays a gestalt process item.

## 5 — Rating

| Metric | Score | Evidence |
|---|---|---|
| Delivered | 4 | All three parts of the concept's vision, plus tags and goals as data; two SWN actions and Psychic Academy's entry left the game until `F.040` and `R.025` |
| Plan held | 3 | ACD-1 to ACD-5 held across eight initiatives; the arc-plan took five amendments (AP-7 to AP-11) and grew from five initiatives to eight, each recorded before code |
| Rework | | |
| Fit | 4 | The arc set the data-driven rule and every initiative follows it; the architecture front door still frames the engine as SWN-only, and `D.010` grew at each close |
| Simplicity | 4 | One holder-blind effect model serves assets and tags, one table-driven goal engine replaced twelve handlers, one `relocate` primitive serves every instant move; the vocabulary is about 40 registrars, the problem's real size |

**Verdict:** [after Rework]

**What surprised me:** the second rulebook's sharpest test was of the first: SWN's
drift costs, tags and goal flow were broken or inert in ways only H.A.W.K and the
data-driven rewrite made visible.

## Follow-ups

| ID | Item | Trigger | Detail |
|---|---|---|---|
| `R.035` | Retire the `A` and `S` flags; name `P` | -- | `A` repeats the `[ability]` block (the `A.001` close made the loader hold them together) and gates Use Asset Ability (`use_asset_ability.go`); derive "has an ability" from the block. `S` is read by no code. `P` means "ask the government to buy here and to land here" (`NeedsPurchasePermission`, `NeedsArrivalPermission` in `domain/asset.go`), beside `purchase_permission` for buying only; make it a named field. Touches every asset in both rulebooks and the campaign copies, and reverses ACD-1's "the flag still picks the harness" |
