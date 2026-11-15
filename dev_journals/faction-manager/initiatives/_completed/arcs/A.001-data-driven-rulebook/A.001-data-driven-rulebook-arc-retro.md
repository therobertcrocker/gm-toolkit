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

No. Each part of the design holds, and the three models (goals, abilities,
effects) are the right breakdown. The better shapes were in the approach.

### The shape of the approach

- **Engine first, then content.** The arc built every SWN special as data
  (`A.001.1` to `A.001.8`), then authored H.A.W.K on top (`A.001.5`). The
  alternative is content first: author H.A.W.K and build only the primitives it
  needs, with SWN catching up. That would have tested the arc's claim from the
  start, and found the drift costs that never cost anything and the asset types
  held in Go (`A.001.5` Efforts 2 and 4) months sooner. It would have left SWN,
  the only rulebook with a rules text to check against, half-built for most of the
  arc. Engine first still holds, because SWN was the reference that proved each
  primitive right; the gap is that the second rulebook joined at the end. A thin
  H.A.W.K slice (a few tags and assets on a small map) run from `A.001.3` on would
  have kept the claim under test, for the cost of a stand-in rulebook to maintain.
- **Sliced by holder and surface.** The initiatives split along who holds a rule
  and how it fires: the A side, the S side, tags, goals. `A.001.7`'s R-1 showed the
  holder is not a real seam: an effect is blind to whether an asset or a tag holds
  it, and scope decides the holder. Slicing by vocabulary (one effect model, with
  assets, tags and later goals as holders from the start) would have spared the
  holder retrofit and AP-9's widening. It was not visible at Arc-Discovery; it came
  from reading 22 registrars. The surface axis was the only one in sight, and it
  cost one retrofit, not a rebuild.
- **Plans that wrote the code.** `A.001.7` planned about 17,500 lines for 3,000
  lines of Go; `A.001.8`'s just-in-time effort plans cut the ratio to a third, and
  its Efforts 2 and 3 reported no anchor drift. A plan that fixes decisions and seams, and lets execution write
  the code, is the better shape. The arc found it inside itself.

### The shape of the design

- **Two models for abilities and effects.** An ability is steps of target → gate → effect
  (ten effects, `domain.Contest`). An effect entry is trigger → filters → gates →
  effect (about 40 registrars, `domain.EffectContest`). Both name their effect,
  carry a block named for it, and roll their contest through the same
  `dice.Engine.Gate`; the shelter and `relocate` sit on both sides. The
  alternative is one model where "the GM spends the action" is a trigger and
  picking a target is a filter that asks. It would give one data shape, one loader
  and one contest type, and the `A` and `S` flags would go away on their own
  (`R.035`). It would cost an action trigger that prompts in order and halts the
  chain on a failed gate, which the hook path does not do, and a rewrite of every
  ability entry. The two shapes hold (Robert's call at this close): merging them
  would either push the ability's ordered, asking, halting chain into every hook,
  or flatten both into something too loose to check. The shared parts already
  meet where they should, in `dice` and `relocate`; the flags are the only
  repetition worth removing (`R.035`).
- **Goals as a third vocabulary.** Goals have their own events, measures and XP
  tables, split from effect triggers on purpose (`A.001.8` D-18: a goal's `kill`
  counts any cause, an effect's `asset_destroyed` only combat). The alternative,
  goals as effects that move progress, would blur those meanings. The split holds:
  a goal measures across batches, an effect acts on one. The cost is a third list
  for a rulebook author to learn.
- **Copy at compose, not reference.** A campaign holds a full copy of its rulebook,
  so a GM's game never changes under them (persistence Key Decisions). This close
  paid the price: one rulebook change meant hand edits in four campaign copies,
  and `hawk-claude` had drifted from its source unseen. The alternative, a campaign
  that reads the source with its own overrides, would end the drift and give up the
  self-contained campaign. The copy holds for a GM's real game; what is missing is
  a sign that a copy is older than its source.
- **A closed vocabulary in Go.** Each effect is a Go registrar that data names; a
  rulebook cannot write new logic. The alternative, conditions and formulas in the
  data, would let a rulebook add rules with no Go, at the cost of a small language
  to design, check and test. The closed set holds: an unknown name fails at load,
  every effect has its tests, and H.A.W.K's dozen new effects took about 1,600
  lines. WWN is the next test of whether the set grows or settles.

Timing, briefly: an asset count at Arc-Plan (AP-7), the question "what does sharing
buy once tags are data?" (AP-8), and page bodies updated per effort (`D.010`)
would each have landed a known turn sooner. None changed the shape.

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
| Rework | 3 | 30 fix commits across the eight branches (about one commit in four), a fix pass at every effort close; this close found 10 seams, all fixed, one a do-nothing prompt in play and the rest drift between initiatives |
| Fit | 4 | The arc set the data-driven rule and every initiative follows it; the architecture front door still frames the engine as SWN-only, and `D.010` grew at each close |
| Simplicity | 4 | One holder-blind effect model serves assets and tags, one table-driven goal engine replaced twelve handlers, one `relocate` primitive serves every instant move; the vocabulary is about 40 registrars, the problem's real size |
| **Total** | **18 / 25** | Mean 3.6: a full delivery on a spine that never reversed, held back by an arc-plan that grew by three initiatives and docs that trailed every close |

**Verdict:** The engine runs any rulebook from data and H.A.W.K is playable, on
cross-cutting decisions that held for all eight initiatives; the cost was a
slicing that bent five times and architecture pages that lagged the code at
every close.

**What surprised me:** the second rulebook's sharpest test was of the first: SWN's
drift costs, tags and goal flow were broken or inert in ways only H.A.W.K and the
data-driven rewrite made visible.

## Follow-ups

| ID | Item | Trigger | Detail |
|---|---|---|---|
| `R.035` | Retire the `A` and `S` flags; name `P` | -- | `A` repeats the `[ability]` block (the `A.001` close made the loader hold them together) and gates Use Asset Ability (`use_asset_ability.go`); derive "has an ability" from the block. `S` is read by no code. `P` means "ask the government to buy here and to land here" (`NeedsPurchasePermission`, `NeedsArrivalPermission` in `domain/asset.go`), beside `purchase_permission` for buying only; make it a named field. Touches every asset in both rulebooks and the campaign copies, and reverses ACD-1's "the flag still picks the harness" |
