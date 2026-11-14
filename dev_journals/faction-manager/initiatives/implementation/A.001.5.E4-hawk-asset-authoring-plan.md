# H.A.W.K Asset Authoring — Effort 4 Plan: Force Pass

## In Plain Words

- **What** — H.A.W.K's own Force assets: 28 armies, fleets and war-machines that replace the SWN copy in `rulebooks/hawk/assets/force_assets.toml`, written as corporate sci-fantasy and built on SWN's balance. Travel is retuned for the H.A.W.K map at the same time. A Drift jump costs real movement, ships move in short-, mid- and long-range speed classes, and Aether Burners can "run hot" to move faster at the cost of their hulls.
- **Why** — Force is the first asset category H.A.W.K owns. Measuring the real map showed that SWN's movement numbers barely work on it, and that Drift costs and ratings did nothing to travel time. Travel is central to how H.A.W.K plays, so it is fixed here, before more assets build on it.
- **How it fits** — Effort 4 of six in `A.001.5`, the last initiative of arc `A.001`. There are three phases: the travel engine, running hot, then the Force data and a live turn on the Effort 3 board. Efforts 5 and 6 build Cunning and Wealth on the same travel rules.

## Context / Goal

Design the Force pass (lore-gate, reflavor, sort; discovery D-9), retune travel for the H.A.W.K map, and specify the engine work and the data to the task.

- Discovery doc: [`A.001.5-hawk-asset-authoring-discovery.md`](../discovery/A.001.5-hawk-asset-authoring-discovery.md), D-7 to D-11
- Overview plan: [`A.001.5-hawk-asset-authoring-overview-plan.md`](A.001.5-hawk-asset-authoring-overview-plan.md), Effort 4, Pass shape, Live turn
- Effort 2 plan: [`A.001.5.E2-hawk-asset-authoring-plan.md`](A.001.5.E2-hawk-asset-authoring-plan.md), Decisions 3 and 9 (asset types; Drift-Runners)
- Effort 3 plan: [`A.001.5.E3-hawk-asset-authoring-plan.md`](A.001.5.E3-hawk-asset-authoring-plan.md), Decisions 9 and 15 (Drift costs; the seed is a test board)
- Canon: `campaigns/hawk/canon/`; Robert's Aethership references in gitignored `scratch/aethership_reference/`

## Decisions Ratified in Planning

1. **Mirror, then drift, with novelty built in** — Robert's call. SWN's eight tiers and its number bands stay, so the balance is borrowed, not rebuilt. Each row starts as a H.A.W.K concept and takes the numbers of its nearest SWN row. Rows may split, merge or be cut. About a third of the rows do something no SWN Force asset does, mostly by moving a built mechanic from another category onto Force. The result is 28 rows: 25 SWN slots (Gengineered Slaves recast) and 3 new ones (Unpaid Interns, Corporate Chaplains, Battle Choir). Nine rows are ★. Rejected: a fresh set built from the canon; a strict reflavor.
2. **Voice: corporate sci-fantasy, played straight** — Robert's brief. Corporate culture, magi-tech and high fantasy meet in every asset. The fiction is deadpan, and the humour comes from the combination, never from a punchline. It invents no new place or history facts. A name pairs one premise word with a plain role word (Corporate Militia, Unionized Partisans, Arcano-Engineered Warbeasts). A ship pairs a magi-tech or fantasy word with its hull type (Runebound Frigates). The final names are Robert's, set over five drafts in gitignored `scratch/A.001.5.E4-force-drafts.md`. Rejected along the way: "fiction term + noun" names, puns, and names that cite specific canon places or events.
3. **H.A.W.K asset types** — Robert's call. Security Personnel (SWN's Military Unit), Operative (Special Forces), Aethership (Starship), Stratagem (Tactic), Facility (kept), and a new Beast. `asset_types` lists H.A.W.K's six, then SWN's six for the Cunning and Wealth stand-ins until Efforts 5 and 6. Aetherships now exist on the board, so Drift-Runners' Aethership build floor comes alive.
4. **Drift ratings on Aetherships and transports** — Robert's call: Drift Haulers 2, Warded Carriers 2, Runebound Frigates 2, Hexcannon Gunships 3, Dreadnought 3. Every other Force asset is unrated, which counts as 1. With Decision 18 a rating now saves real movement on every Drift jump. Rejected: a rating on every asset; ratings scaled by tier.
5. **What an Aethership looks like goes into the canon** — from Robert's reference images. A long, angular hull on a spine, often in company livery, built around a glowing Aether-Core held in a ring or in cradling hull arms. Lit ward-domes cover open decks, energy fins flare from the hull, and the largest ships carry one spinal gun. No sails. It is written into `glossary.md`'s Aethership entry in Commit 8.
6. **War-beasts need Aether-Tech Evangelists** — Robert's call. Arcano-Engineered Warbeasts carries `requires_tag = "HAWK-T-011"`. The tag's rules text and its canon entry each gain a line, which writes both ends of the link (D-11). O.W.L. holds the tag, so beasts can appear on the test board. Rejected: no gate; Aether Refiners, which no seeded faction holds.
7. **The teleport rows use the instant `relocate` ability, and `relocate` learns two switches** — Robert's call. A transport moves its cargo on the transport's own movement order, which advances by the transport's speed (`internal/faction/engine/effect/effects/transport.go`), so a teleport circle with no speed would never arrive. Teleportation Hubs is SWN's Covert Shipping pattern, pure reuse. The Deep-Insertion Gate needs `ignores_permission` and `ignores_lockdown` on `RelocateParams`, named after the transport profile's fields: a parameterize row. Rejected: relocate with no new fields (the Gate loses its trait); SWN's transport with speed restored (the fiction says fixed installations).
8. **Warded Carriers is a Facility** — Robert's call. It keeps SWN's type, so it has no population floor. As a transport it keeps drift rating 2. Rejected: an Aethership.
9. **Aether-Tech Marines are plain Space Marines** — Robert's call. The draft's `capture_target` would have spent a 30-Coin unit on its first hit, and the final name no longer says takeover. Rejected: keeping the capture; moving it to a cheaper row.
10. **Arcano-Linked Batteries keep the name, and the canon holds** — Robert's call. Arcano-Tech needs a link to the Weave, so the fiction says the batteries work off the Material Plane only on Weave Anchor ground. This is flavour only, and the engine does not enforce it. The Radiant Throne's seeded battery at Solar Gate sits outside that fiction, which the test board accepts (Effort 3 Decision 15). Rejected: renaming to Aether-Linked; amending the canon.
11. **The stand-ins' type gap is accepted** — Robert's call. Seven Cunning and Wealth stand-ins name SWN's Force types in their rules or cargo lists: Seductress, Covert Shipping, Boltholes, Covert Transit Net, Medical Center, Freighter Contract and Shipping Combine. Five of them are in the seed. Until Efforts 5 and 6, they cannot carry, shelter, revive or be attacked by H.A.W.K Force units. Likewise, H.A.W.K transports name only H.A.W.K types, so they cannot carry stand-in assets of SWN types. Rejected: adding H.A.W.K types to the stand-in copies.
12. **Three phases** — Robert's call; amends overview Decision 6, which gave a pass one build phase. The travel engine (Decisions 7, 18, 19), running hot (Decision 24) and the data each get a session, in that order, so D-9's "build before data" holds. Rejected: running hot to the Backlog; travel as an effort of its own.
13. **A description is one fiction line plus the rules** — Robert's call. An asset has one text field, and the TUI shows it. The full fiction lives in `canon/assets.md` (D-5). The rules text says "place" where SWN says "world" or "planet", as the canon does.
14. **The seed swap is by ID suffix** — Every seeded Force stand-in has an heir in the same slot: `SWN-F<t>-<n>` becomes `HAWK-F<t>-<n>` for all ten seeded definitions, at the same tech level. The instance IDs change with them (overview, Live turn).
15. **Numbers for the new rows** — Unpaid Interns: SWN's False Front mechanic at tier 1 (hp 2, cost 1, no attack). Corporate Chaplains: Hardened Personnel's numbers at tier 3 (hp 4, cost 6, counter 1d4+1) with Medical Center's revive, which also covers Beasts. Battle Choir: tier 5 (hp 6, cost 12, counter 1d6) with Integral Protocols' extra d10 on its world's Force attacks. Shield Wardens' forced reroll is `once_per = "asset"`, so each unit forces one reroll a turn. Tuning waits for play.
16. **Wizard Advisors support the troops, not the spies** — Robert's call. The draft gave them Panopticon Matrix's stealth sweep, which read as Cunning's job. Now, once per turn, a Force defense by the faction on their place rolls an extra d10: they scry the attack before it lands. It pairs with Battle Choir, which does the same for attacks. Reuse: `bonus_die`, scope world, phase defense. Rejected: the stealth sweep; a world-wide forced reroll; a Force tech-level floor.
17. **Long-range travelling cargo goes to Wealth** — Robert's call. Travel on this map matters, so the option must exist. Force's travelling transports are short- and mid-range (Drift Haulers, Warded Carriers), and its far movers teleport (Decision 7). Effort 6 keeps Wealth's travelling transports, at long-range speed, with H.A.W.K cargo types and drift ratings. Recorded in the overview's Effort 6 scope.
18. **A Drift jump costs its rated movement points, in every rulebook** — Robert's call, and a fix. A move order walks its path one entry at a time, `speed` entries a turn, and a lane jump was one entry: `drift_costs` and drift ratings only picked the route, and the path's cost was thrown away (`internal/faction/engine/world/movement.go`). Now a turn gives an asset `speed` points; a hex step costs 1 and a lane jump costs `DriftCost(rating)`. This makes Effort 3 Decision 9 ("a crossing always costs more than a step"), Effort 2 Decision 9 (Drift-Runners) and Decision 4 true. SWN's sandbox slows too: its comment always said "hex-equivalent crossing cost", and its numbers can be retuned if play asks. Rejected: a per-rulebook switch.
19. **Points carry, and a ship mid-jump is in the Drift** — Robert's calls. Points not spent at the end of a turn are banked on the order and spent first next turn, so a trip takes its total cost ÷ speed, rounded up. A ship part-way through a jump stays on the hex at the near end of the lane, like any asset in transit: on no place, so it cannot be attacked or act. The TUI names where an asset in transit is: "In the Drift between Concordia and The Prime Material" mid-jump, "In Concordia, en route to Veil" otherwise. Today it shows a blank. Rejected: a jump that must start at the beginning of a turn.
20. **Three speed classes: short 10, mid 14, long 19** — Robert's targets, measured on the map (16 hex steps and 2 lane jumps for a median trip; 35 and 5 for the longest). Short: its own region in 1 turn (2 for Infernus, Radiance's Edge and The Maelstrom), a typical neighbouring region in 2. Mid: a typical neighbour in 1, the map in 4. Long: the whole map in 3 at most. Force movers by role: short are Drift Haulers and Aether-Tech Marines; mid are Warded Carriers, Runebound Frigates and Hexcannon Gunships; long is the Dreadnought. Efforts 5 and 6 sort their movers into the same classes. Rejected: speeds that meet "a neighbour in 1 turn" for small ships (about 15, too close to long-range); classes by tier.
21. **Drift costs `[10, 7, 5, 3, 2]`** — Robert's call; amends Effort 3 Decision 9's `[6, 5, 4, 3, 2]`. Front-loaded: an unrated ship finds the Drift punishing, and the first rating upgrades matter most. A jump still costs more than a step. Rejected: even steps `[10, 8, 6, 4, 2]`; back-loaded `[10, 9, 8, 5, 2]`; keeping E3's table.
22. **Transports drop `max_hex`** — The engine never reads a transport's `max_hex`. Only relocate reads `MaxHex` (`internal/faction/engine/relocate/relocate.go`). A transport carries its cargo to any place, and its speed sets the time. H.A.W.K's transports leave the key out, and their descriptions give their speed.
23. **Teleport ranges: Hubs 8, Gate 14** — Robert's call. Range counts real hexes, and a lane counts as zero. At 8, most places reach 2 to 9 others. At 14, the Gate reaches about half the map. Both stay below a mid-range ship's turn of travel. Rejected: Hubs 12 and Gate 20; Hubs 6 and Gate 10.
24. **Aether Burners run hot** — Robert's calls; replaces Effort 2's "every asset can move itself one hex" (speed floor 1). A move order may run hot, chosen per order in the movement form. Hot: the faction's Aetherships move 5 more points a turn, every other asset can move itself at speed 5, and each asset takes 1d4 damage every turn it moves hot. An asset that can move only by running hot always runs hot. It is built as data: a `hot` filter on the speed effects, a new `speed_bonus` effect beside `speed_floor`, and a new `hot_move` trigger for `self_damage`. Rejected: always on; speed floor 5 alone; a drift rating bonus (Drift-Runners' trick).
25. **Fighting in transit goes to the Backlog** — Robert's call. Robert expects a good part of faction interaction to happen in a region, not on a place: assets in transit that can reach each other should be able to fight. Attacks are place-based today, and the rule needs its own design (what "reachable" means, fleets against places in a region, mid-jump ships, turn order, targeting). Filed as a feature row in the plan commit. Rejected: building it in this initiative.
26. **Two read-only map helpers** — `spatial.Adjacent` and `RegionMap.RegionName`, for pricing a step and naming a region. Amends the discovery's Out of Scope ("no `internal/spatial` changes"): the map's shape and data do not change.

## Decision Record — Execution

1. **Commit 2: a transport test's made-up path is made contiguous** — `TestTransport_ReviseMidFlight` (`internal/faction/engine/testharness/scenarios/transport_lifecycle_test.go`) mocked a path to world-d that jumped from hex (2,0) to (5,0). Under the old tick, any path entry was one step. Under Decision 18, a non-neighbour step is a Drift jump, and the harness's `DriftCosts = [5]` is more than speed 2, so the order never arrived. The test's intent is the cargo manifest carried over a revise, not Drift travel, so the path now runs (2,0) → (2,1) → (2,2). Rejected: keeping the gap and asserting a banked jump; that would test Drift inside a revise test.
2. **Commit 4: a third speed stub takes the `hot` argument** — `internal/faction/engine/world/world_test.go` has its own `speedFloor` stub for `ModifyAssetSpeed`, which Task 8 does not list. It gains the `_ bool` parameter, like the two stubs Task 8 names.
3. **After Commit 6: a revise keeps its banked points** — Robert's call. `BuildMovementMutations`' Revise branch built a fresh `MovementOrder`, so a ship that had banked 3 of a 6-point jump lost those points when revised. Now the new order copies `Banked` from the old one, and the next tick spends the points on whatever the new path needs first. Banked points are turns already spent waiting, as Decision 19 says, so a revise does not take them back; the 1 Coin revise cost is unchanged. The run-hot choice is re-taken on a revise, as Commit 5 Task 2 specifies. `LocationLabel`'s fence-sign still holds where it is read: the tick runs before the movement decisions in a turn, so a revised order is ticked before its label is shown again. Rejected: keeping the points only when the new path starts with the same jump (more edge cases for no gain); dropping them (the old behaviour, which no one chose).
4. **Commit 8: Wizard Advisors and Battle Choir give their die once per turn per asset** — Robert's call; found before the live turn. Both descriptions say "once per turn", but Commit 7's data had no `gates` line, and `bonus_die` sets a budget only when `once_per` is set (`internal/faction/engine/effect/effects/bonus_die.go`), so the die came on every matching roll on the place. Both rows now carry `gates = { once_per = "asset" }`, like Shield Wardens: each copy gives one extra d10 a turn. This settles Decision 15's "Integral Protocols' extra d10" for Battle Choir: the die is Integral Protocols', the limit is the description's. Commit 7 Task 1 and the Sort Table now show the gate. Rejected: `once_per = "owner_faction"` (one die a turn however many copies); no limit, with "once per turn" dropped from the text.
5. **Commit 8: the live turn's findings** — Robert played turns 2 and 3 on the reset board. The travel engine matched the rules: Radiant Throne's hot Runebound Frigate spent 14 of 19 points and banked 5 before a 7-point jump, showing "In the Drift between …"; its hot Batteries banked 2 before a 10-point jump; the burn took 1 and 3 HP. Two findings, neither a fix here. Blackwater's Covert Transit Net was offered with no Special Forces asset left to move (its Hardened Personnel is now an Operative) and cancelled the turn: Decision 11's gap in play, which Robert accepts with no Backlog row, as Effort 5 replaces the stand-in. Informers' reveal-stealth faction pick lists every rival: filed as `F.053` with Robert's idea of revealing with no pick.

## Sort Table

Primitives are the registrar names in `internal/faction/engine/effect/effects/registrars.go`, the ability effects in `internal/faction/domain/asset.go`, or a transport profile. Surface: `—` plain, `effects` S-side, `ability` A-side, `transport`. ★ = no SWN Force asset does it.

| HAWK ID | Name | SWN ancestor | Surface | Call | Primitive | Notes |
|---------|------|--------------|---------|------|-----------|-------|
| HAWK-F1-001 | Corporate Security | Security Personnel | — | reuse | — | Security Personnel |
| HAWK-F1-002 | Unsanctioned Adventurers | Hitmen | — | reuse | — | Operative |
| HAWK-F1-003 | Reluctant Militia | Militia Unit | — | reuse | — | Security Personnel, P |
| HAWK-F1-004 | Arcano-Engineered Warbeasts | Gengineered Slaves | — | reuse | `requires_tag` | ★ Beast; needs `HAWK-T-011` (Decision 6) |
| HAWK-F1-005 | Unpaid Interns | False Front (Cunning) | effects | reuse | `sacrifice_self` | ★ new row (Decision 15) |
| HAWK-F2-001 | Drift Haulers | Heavy Drop Assets | transport | reuse | transport profile | short (10), drift 2, no `max_hex` |
| HAWK-F2-002 | Aetheric-Lance Corps | Elite Skirmishers | — | reuse | — | Security Personnel, P |
| HAWK-F2-003 | Shield Wardens | Hardened Personnel | effects | reuse | `reroll` on `enemy_roll_at_me` | ★ forced, `once_per = "asset"` |
| HAWK-F2-004 | Unionized Partisans | Guerrilla Populace | — | reuse | — | Security Personnel |
| HAWK-F3-001 | Overclocked Zealots | Zealots | effects | reuse | `self_damage` | Operative |
| HAWK-F3-002 | Runic Minefield | Cunning Trap | — | reuse | — | Stratagem |
| HAWK-F3-003 | Wizard Advisors | Counterintel Unit | effects | reuse | `bonus_die`, scope world | ★ Force defenses on its place, `once_per = "asset"` (Decision 16; Decision Record 4) |
| HAWK-F3-004 | Corporate Chaplains | Medical Center (Wealth) | effects | reuse | `restore_asset` | ★ new row (Decision 15) |
| HAWK-F4-001 | Warded Carriers | Beachhead Landers | transport | reuse | transport profile | Facility, mid (14), drift 2 (Decision 8) |
| HAWK-F4-002 | Teleportation Hubs | Extended Theater | ability | reuse | `relocate` | ★ A, no speed, range 8 (Decisions 7, 23) |
| HAWK-F4-003 | Runebound Frigates | Strike Fleet | — | reuse | — | Aethership, mid (14), drift 2 |
| HAWK-F4-004 | Arcane Strike Cadre | Postech Infantry | — | reuse | — | Security Personnel, P |
| HAWK-F5-001 | Hexcannon Gunships | Blockade Fleet | effects | reuse | `coin_steal` | Aethership, mid (14), drift 3 |
| HAWK-F5-002 | Bonded Armory | Pretech Logistics | ability | reuse | `purchase_enabler` | A |
| HAWK-F5-003 | Pact-Bound Assassins | Psychic Assassins | effects | reuse | `apply_stealth` | Operative |
| HAWK-F5-004 | Battle Choir | Integral Protocols' die | effects | reuse | `bonus_die`, scope world | ★ new row, `once_per = "asset"` (Decision 15; Decision Record 4) |
| HAWK-F6-001 | Chartered Mage-Knights | Pretech Infantry | effects | reuse | `counterstrike` | ★ as Cracked Comms; P |
| HAWK-F6-002 | Arcano-Linked Batteries | Planetary Defenses | effects | reuse | `restrict_attack` | Aethership attackers only (Decision 10) |
| HAWK-F6-003 | Forged Siege Engines | Gravtank Formation | — | reuse | — | Security Personnel, P |
| HAWK-F7-001 | Deep-Insertion Gate | Deep Strike Landers | ability | parameterize | `relocate` + `ignores_permission`, `ignores_lockdown` | ★ A, no speed, range 14; Commit 1 |
| HAWK-F7-002 | Abjurative Firewall | Integral Protocols | effects | reuse | `bonus_die`, `restrict_attack` | Facility |
| HAWK-F7-003 | Aether-Tech Marines | Space Marines | — | reuse | — | short (10), unrated; no P flag, so it never asks permission (Decision 9) |
| HAWK-F8-001 | Dreadnought | Capital Fleet | — | reuse | — | Aethership, long (19), drift 3, `purchase_permission` |

Engine rows outside the table: Drift jump cost and banked points (build, Commit 2), transit labels (build, Commit 3), the `hot` filter and `speed_bonus` (build, Commit 4), hot move orders (build, Commit 5), the `hot_move` burn (build, Commit 6). Aether Burners (`HAWK-T-003`) is rewritten in Commit 7.

## Shared Context

**Gate.** Every commit lands green on `go build ./...`, `go vet ./...` and `go test ./...` from the repo root. `TestNewWithRulebook_ShippedRulebook` (`internal/faction/engine/core_test.go`) composes and loads `rulebooks/swn` and `rulebooks/hawk` and builds the engine from each, so it checks every new effect, filter, type name and relocate block.

**Campaign commands** take the prefix `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns` and run after `go build -o gm-toolkit ./cmd/gm-toolkit` (overview Decision 5).

**Measuring travel.** The plan session measured the map with throwaway programs in gitignored `scratch/distcheck/`, `scratch/stepcheck/`, `scratch/pacecheck/` and `scratch/regioncheck/` (`go run ./scratch/pacecheck campaigns/hawk/spatial`). A later retune can rerun them.

## Out of Scope

- **Fighting in transit** — Decision 25; a Backlog row.
- **Retuning SWN's Drift costs or speeds** — Decision 18; only if sandbox play asks.
- **The Cunning and Wealth stand-ins' type lists and speeds** — Decision 11; Efforts 5 and 6 replace those files and sort their movers into the speed classes (Decision 20).
- **Engine enforcement of the Arcano-Linked Batteries' Weave fiction** — Decision 10.
- **Retuning SWN's combat numbers** — tuning waits for play (Decision 15).
- **Regenerating the seed for real play** — after the arc closes (Effort 3 Decision 15). E4 swaps IDs only.

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — The travel engine**
  - Commit 1 — `feat(E4): let relocate ignore permission and lockdown` (5 tasks)
  - Commit 2 — `feat(E4): make a Drift jump cost its rated movement points` (6 tasks)
  - Commit 3 — `feat(E4): say where an asset in transit is` (4 tasks)
- [x] **Phase 2 — Running hot**
  - Commit 4 — `feat(E4): add hot-only speed effects and speed_bonus` (8 tasks)
  - Commit 5 — `feat(E4): let a move order run hot` (6 tasks)
  - Commit 6 — `feat(E4): burn assets that move hot` (3 tasks)
- [x] **Phase 3 — The Force pass**
  - Commit 7 — `feat(E4): replace the Force stand-in with H.A.W.K's own assets` (4 tasks)
  - Commit 8 — `docs(E4): write the Force canon and play the live turn` (6 tasks)

### Phase 1 — The travel engine

**In plain words:** Teleports can land past permission and lockdowns, a Drift jump takes as many movement points as its cost says, points left over carry to the next turn, and the TUI says where a moving asset is.

#### Commit 1 — `feat(E4): let relocate ignore permission and lockdown`

##### Task 1 — `internal/faction/domain/asset.go`

`RelocateParams` gains the two switches.

**Find:**
```go
// RelocateParams configures a relocate step. Scope picks what moves (the acting asset vs
// other friendly assets); Range picks hex-radius vs region reach; MaxHex is the radius for
// RangeHexes (ignored for RangeRegion); CoinCost is charged per CoinCostMode; AssetType
// filters "other" candidates (empty = any), Categories narrows them by asset category
// (empty = any) and ExcludeTypes drops types; MaxAssets caps the count moved (0 = unbounded).
type RelocateParams struct {
	Scope        RelocateScope
	Range        RelocateRange
	MaxHex       int
	CoinCost     int
	CoinCostMode RelocateCostMode
	AssetType    AssetType
	Categories   []FactionStat
	ExcludeTypes []AssetType
	MaxAssets    int
}
```
**Replace with:**
```go
// RelocateParams configures a relocate step. Scope picks what moves (the acting asset vs
// other friendly assets); Range picks hex-radius vs region reach; MaxHex is the radius for
// RangeHexes (ignored for RangeRegion); CoinCost is charged per CoinCostMode; AssetType
// filters "other" candidates (empty = any), Categories narrows them by asset category
// (empty = any) and ExcludeTypes drops types; MaxAssets caps the count moved (0 = unbounded).
// IgnoresPermission lands each asset without asking the government, and IgnoresLockdown
// lands it past a Transport Lockdown, as a transport profile's fields do (Deep-Insertion Gate).
type RelocateParams struct {
	Scope             RelocateScope
	Range             RelocateRange
	MaxHex            int
	CoinCost          int
	CoinCostMode      RelocateCostMode
	AssetType         AssetType
	Categories        []FactionStat
	ExcludeTypes      []AssetType
	MaxAssets         int
	IgnoresPermission bool
	IgnoresLockdown   bool
}
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

The record reads the two keys, and `convertRelocate` carries them. Both the ability step and `grant_movement` go through `convertRelocate`, so both accept them.

**Find:**
```go
type relocateRecord struct {
	Scope        string   `toml:"scope"`
	Range        string   `toml:"range"`
	MaxHex       int      `toml:"max_hex"`
	CoinCost     int      `toml:"coin_cost"`
	CoinCostMode string   `toml:"coin_cost_mode"`
	AssetType    string   `toml:"asset_type"`
	Categories   []string `toml:"categories"`
	ExcludeTypes []string `toml:"exclude_types"`
	MaxAssets    int      `toml:"max_assets"`
}
```
**Replace with:**
```go
type relocateRecord struct {
	Scope             string   `toml:"scope"`
	Range             string   `toml:"range"`
	MaxHex            int      `toml:"max_hex"`
	CoinCost          int      `toml:"coin_cost"`
	CoinCostMode      string   `toml:"coin_cost_mode"`
	AssetType         string   `toml:"asset_type"`
	Categories        []string `toml:"categories"`
	ExcludeTypes      []string `toml:"exclude_types"`
	MaxAssets         int      `toml:"max_assets"`
	IgnoresPermission bool     `toml:"ignores_permission"`
	IgnoresLockdown   bool     `toml:"ignores_lockdown"`
}
```

**Find:**
```go
	params := &domain.RelocateParams{
		Scope:        scope,
		Range:        rng,
		MaxHex:       record.MaxHex,
		CoinCost:     record.CoinCost,
		CoinCostMode: costMode,
		MaxAssets:    record.MaxAssets,
	}
```
**Replace with:**
```go
	params := &domain.RelocateParams{
		Scope:             scope,
		Range:             rng,
		MaxHex:            record.MaxHex,
		CoinCost:          record.CoinCost,
		CoinCostMode:      costMode,
		MaxAssets:         record.MaxAssets,
		IgnoresPermission: record.IgnoresPermission,
		IgnoresLockdown:   record.IgnoresLockdown,
	}
```

##### Task 3 — `internal/faction/engine/relocate/relocate.go`

`Run` keeps locked worlds as destinations when the relocate ignores lockdowns, and `askLanding` asks no one when it ignores permission.

**Find:**
```go
	destinations := lockedOutRemoved(request, candidateWorlds)
	if len(destinations) == 0 {
```
**Replace with:**
```go
	destinations := candidateWorlds
	if !params.IgnoresLockdown {
		destinations = lockedOutRemoved(request, candidateWorlds)
	}
	if len(destinations) == 0 {
```

**Find:**
```go
// askLanding asks the destination's government whether order's asset may land
// there: a relocate asks at its one step (D-24). It returns the refusal, or nil.
func askLanding(request Request, order action.RelocateOrder) (*domain.PermissionRefused, error) {
	def := request.Rulebook.Assets[order.Asset.DefinitionID]
	if def == nil {
		return nil, nil
	}
```
**Replace with:**
```go
// askLanding asks the destination's government whether order's asset may land
// there: a relocate asks at its one step (D-24). A relocate that ignores
// permission asks no one. It returns the refusal, or nil.
func askLanding(request Request, order action.RelocateOrder) (*domain.PermissionRefused, error) {
	def := request.Rulebook.Assets[order.Asset.DefinitionID]
	if def == nil || request.Params.IgnoresPermission {
		return nil, nil
	}
```

##### Task 4 — `internal/faction/engine/relocate/relocate_test.go`

Append at the end of the file. The first test moves a P asset onto a governed world with no permission ask: the mock fails on any `RequestPermission` call. The second keeps a locked world as a destination and lands on it.

```go
func TestRun_IgnoresPermissionAsksNoOne(t *testing.T) {
	ctrl := gomock.NewController(t)
	faction, _, web, mercs, _, factionState := landingFixture()
	params := transitWebParams()
	params.IgnoresPermission = true
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: mercs, DestinationWorld: "mid"}}, nil)

	mutations, err := Run(Request{
		Faction: faction, Acting: web, Params: params,
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(), FactionState: factionState,
	})

	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if moved, ok := mutations[0].(domain.AssetMoved); !ok || moved.AssetID != "mp-1" || moved.ToLocation.WorldID != "mid" {
		t.Errorf("mutations[0] = %+v, want the mercenaries landed on mid without an ask", mutations[0])
	}
}

func TestRun_IgnoresLockdownLandsOnALockedWorld(t *testing.T) {
	ctrl := gomock.NewController(t)
	web := &domain.Asset{ID: "web", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	works := &domain.Asset{ID: "wf-1", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	faction := &domain.Faction{ID: "f1", Coin: 1, Assets: map[string]*domain.Asset{"web": web, "wf-1": works}}
	lockdown := &domain.Asset{ID: "lock", OwnerID: "f2", Location: domain.Location{WorldID: "mid"}, Ready: true, Maintained: true, CurrentHP: 10,
		Conditions: []domain.Condition{{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "f1", WorldID: "mid"}}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"f1": faction,
		"f2": {ID: "f2", Assets: map[string]*domain.Asset{"lock": lockdown}},
	}}
	params := transitWebParams()
	params.IgnoresLockdown = true
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ *domain.Asset, _ []*domain.Asset, worlds []string, _ int) ([]action.RelocateOrder, error) {
			if !slices.Contains(worlds, "mid") {
				t.Errorf("destinations = %v, want the locked world kept", worlds)
			}
			return []action.RelocateOrder{{Asset: works, DestinationWorld: "mid"}}, nil
		})

	mutations, err := Run(Request{
		Faction: faction, Acting: web, Params: params,
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(), FactionState: factionState,
	})

	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if moved, ok := mutations[0].(domain.AssetMoved); !ok || moved.AssetID != "wf-1" || moved.ToLocation.WorldID != "mid" {
		t.Errorf("mutations[0] = %+v, want wf-1 landed on the locked world", mutations[0])
	}
}
```

##### Task 5 — `internal/faction/rulebook/rulebook_test.go`

Append at the end of the file:

```go
func TestConvertRelocate_IgnoresPermissionAndLockdown(t *testing.T) {
	params, err := convertRelocate(&relocateRecord{Scope: "other", MaxHex: 3, CoinCost: 2, MaxAssets: 1, IgnoresPermission: true, IgnoresLockdown: true})
	if err != nil {
		t.Fatalf("convertRelocate: %v", err)
	}
	if !params.IgnoresPermission || !params.IgnoresLockdown {
		t.Errorf("got %+v, want IgnoresPermission and IgnoresLockdown set", *params)
	}
}
```

##### Commit message

```
feat(E4): let relocate ignore permission and lockdown

- RelocateParams: IgnoresPermission and IgnoresLockdown, read from
  ignores_permission and ignores_lockdown in a relocate block
- relocate.Run keeps locked worlds as destinations and asks no
  government when the switches are set
- tests: a P asset lands unasked; a locked world stays a destination
```

#### Commit 2 — `feat(E4): make a Drift jump cost its rated movement points`

##### Task 1 — `internal/spatial/region_map.go`

Two read-only helpers (Decision 26). Insert after `RegionOfHex`:

**Find:**
```go
func (regionMap *RegionMap) RegionOfHex(hex HexCoord) (string, bool) {
	id, ok := regionMap.hexIndex[hex]
	return id, ok
}
```
**Replace with:**
```go
func (regionMap *RegionMap) RegionOfHex(hex HexCoord) (string, bool) {
	id, ok := regionMap.hexIndex[hex]
	return id, ok
}

// RegionName returns a region's display name, or id when the map has no such
// region or the region has no name.
func (regionMap *RegionMap) RegionName(id string) string {
	if region, ok := regionMap.regions[id]; ok && region.Name != "" {
		return region.Name
	}
	return id
}

// Adjacent reports whether two hexes are neighbours on the grid. The two ends of a
// Drift lane are not.
func Adjacent(a, b HexCoord) bool {
	for _, delta := range hexNeighbors {
		if a.Q+delta.Q == b.Q && a.R+delta.R == b.R {
			return true
		}
	}
	return false
}
```

##### Task 2 — `internal/faction/domain/location.go`

The order banks unspent points (Decision 19).

**Find:**
```go
	DriftRating   int                 `toml:"drift_rating" json:"drift_rating"`
	CargoAssetIDs []string            `toml:"cargo_asset_ids,omitempty" json:"cargo_asset_ids,omitempty"`
```
**Replace with:**
```go
	DriftRating   int                 `toml:"drift_rating" json:"drift_rating"`
	CargoAssetIDs []string            `toml:"cargo_asset_ids,omitempty" json:"cargo_asset_ids,omitempty"`
	// Banked is the movement points carried toward the next step: a Drift jump
	// that costs more than a turn's speed is paid over several turns (A.001.5
	// Effort 4 Decision 19).
	Banked int `toml:"banked,omitempty" json:"banked,omitempty"`
```

##### Task 3 — `internal/faction/domain/mutation.go`

**Find:**
```go
type MovementOrderProgressed struct {
	FactionID         string            `json:"faction_id"`
	AssetID           string            `json:"asset_id"`
	NewStepIdx        int               `json:"new_step_idx"`
	RegionHex         spatial.RegionHex `json:"region_hex"`
```
**Replace with:**
```go
type MovementOrderProgressed struct {
	FactionID         string            `json:"faction_id"`
	AssetID           string            `json:"asset_id"`
	NewStepIdx        int               `json:"new_step_idx"`
	RegionHex         spatial.RegionHex `json:"region_hex"`
	Banked            int               `json:"banked,omitempty"`
```

##### Task 4 — `internal/faction/engine/mutation/mutation.go`

A step keeps the points it banked. A hold at the destination's edge starts its last step fresh.

**Find:**
```go
				if asset, ok := faction.Assets[v.AssetID]; ok && asset.CurrentOrder != nil {
					asset.CurrentOrder.StepIdx = v.NewStepIdx
					asset.Location = domain.Location{
						RegionHex: v.RegionHex,
					}
```
**Replace with:**
```go
				if asset, ok := faction.Assets[v.AssetID]; ok && asset.CurrentOrder != nil {
					asset.CurrentOrder.StepIdx = v.NewStepIdx
					asset.CurrentOrder.Banked = v.Banked
					asset.Location = domain.Location{
						RegionHex: v.RegionHex,
					}
```

**Find:**
```go
					asset.CurrentOrder.StepIdx = v.NewStepIdx
					asset.Location = domain.Location{RegionHex: v.RegionHex}
					if v.Reason == domain.HoldLockdownPaid {
```
**Replace with:**
```go
					asset.CurrentOrder.StepIdx = v.NewStepIdx
					asset.CurrentOrder.Banked = 0
					asset.Location = domain.Location{RegionHex: v.RegionHex}
					if v.Reason == domain.HoldLockdownPaid {
```

##### Task 5 — `internal/faction/engine/world/movement.go`

The ticker spends points (Decision 18).

**Find:**
```go
		newStepIdx := asset.CurrentOrder.StepIdx + speedOf(def)
		log.Debug("ticking movement order for asset", "asset_id", asset.ID, "from_hex", asset.Location.RegionHex, "to_hex", asset.CurrentOrder.Destination.RegionHex)
		if newStepIdx >= len(asset.CurrentOrder.Path)-1 {
```
**Replace with:**
```go
		order := asset.CurrentOrder
		points := order.Banked + speedOf(def)
		newStepIdx := order.StepIdx
		for newStepIdx < len(order.Path)-1 {
			cost := stepCost(order.Path[newStepIdx], order.Path[newStepIdx+1], rulebook, order.DriftRating)
			if cost > points {
				break
			}
			points -= cost
			newStepIdx++
		}
		log.Debug("ticking movement order for asset", "asset_id", asset.ID, "from_hex", asset.Location.RegionHex, "to_hex", order.Destination.RegionHex, "banked", points)
		if newStepIdx >= len(order.Path)-1 {
```

**Find:**
```go
			step := asset.CurrentOrder.Path[newStepIdx]
			mutations = append(mutations, domain.MovementOrderProgressed{
				FactionID:         asset.OwnerID,
				AssetID:           asset.ID,
				NewStepIdx:        newStepIdx,
				RegionHex:         step,
```
**Replace with:**
```go
			step := order.Path[newStepIdx]
			mutations = append(mutations, domain.MovementOrderProgressed{
				FactionID:         asset.OwnerID,
				AssetID:           asset.ID,
				NewStepIdx:        newStepIdx,
				RegionHex:         step,
				Banked:            points,
```

Then add after `TickMovementOrders` (before `// BuildMovementMutations issues, ...`):

```go
// stepCost is the movement points one step of a path takes: 1 for a neighbouring
// hex, the Drift cost of driftRating for a lane jump.
func stepCost(from, to spatial.RegionHex, rulebook *rulebook.Rulebook, driftRating int) int {
	if spatial.Adjacent(from.Coord, to.Coord) {
		return 1
	}
	return rulebook.DriftCost(driftRating)
}
```

The `TickMovementOrders` doc comment, first line: `// TickMovementOrders advances each open order of faction's assets by the asset's` becomes `// TickMovementOrders spends each open order's banked points plus the asset's`. The second line, `// speed, as registry's speed floors resolve it (Mercenary Group).`, becomes `// speed, as registry's speed floors resolve it (Mercenary Group): a hex step costs 1, a Drift jump its rated cost, and what is left is banked.`

##### Task 6 — tests

`internal/spatial/region_map_test.go`, append:

```go
func TestAdjacent(t *testing.T) {
	origin := HexCoord{Q: 2, R: 2}
	for _, delta := range hexNeighbors {
		if neighbour := (HexCoord{Q: origin.Q + delta.Q, R: origin.R + delta.R}); !Adjacent(origin, neighbour) {
			t.Errorf("Adjacent(%v, %v) = false, want true", origin, neighbour)
		}
	}
	for _, other := range []HexCoord{{Q: 2, R: 2}, {Q: 4, R: 2}, {Q: 3, R: 3}} {
		if Adjacent(origin, other) {
			t.Errorf("Adjacent(%v, %v) = true, want false", origin, other)
		}
	}
}

func TestRegionName(t *testing.T) {
	regionMap, err := newRegionMap(map[string]*Region{
		"c": {ID: "c", Name: "Concordia", Hexes: map[HexCoord]bool{{Q: 0, R: 0}: true}},
	}, nil)
	if err != nil {
		t.Fatalf("newRegionMap: %v", err)
	}
	if got := regionMap.RegionName("c"); got != "Concordia" {
		t.Errorf("RegionName(c) = %q, want Concordia", got)
	}
	if got := regionMap.RegionName("x"); got != "x" {
		t.Errorf("RegionName(x) = %q, want the ID for an unknown region", got)
	}
}
```

`internal/faction/engine/world/movement_test.go`, append. The path crosses one lane, from hex (0,0) in region `a` to (5,0) in region `b`, then one step to (6,0):

```go
func driftRulebook(speed int) *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets:     map[string]*domain.AssetDefinition{"ship": {ID: "ship", Speed: speed}},
		DriftCosts: []int{6, 5, 4, 3, 2},
	}
}

func laneOrder(driftRating, banked int) *domain.Asset {
	path := []spatial.RegionHex{
		{RegionID: "a", Coord: spatial.HexCoord{Q: 0}},
		{RegionID: "b", Coord: spatial.HexCoord{Q: 5}},
		{RegionID: "b", Coord: spatial.HexCoord{Q: 6}},
	}
	return &domain.Asset{ID: "s1", OwnerID: "f1", DefinitionID: "ship",
		CurrentOrder: &domain.MovementOrder{AssetID: "s1", Path: path, DriftRating: driftRating, Banked: banked}}
}

func tickOne(t *testing.T, asset *domain.Asset, rb *rulebook.Rulebook) domain.Mutation {
	t.Helper()
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{asset.ID: asset}}
	mutations, err := regionWorldEngine(t).TickMovementOrders(faction, rb, nil, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("TickMovementOrders: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("mutations = %+v, want one", mutations)
	}
	return mutations[0]
}

func TestTickMovementOrders_DriftJumpCostsItsRatedPoints(t *testing.T) {
	first := tickOne(t, laneOrder(1, 0), driftRulebook(3))
	if got, ok := first.(domain.MovementOrderProgressed); !ok || got.NewStepIdx != 0 || got.Banked != 3 {
		t.Errorf("turn 1 = %+v, want no step and 3 points banked toward the 6-point jump", first)
	}

	second := tickOne(t, laneOrder(1, 3), driftRulebook(3))
	if got, ok := second.(domain.MovementOrderProgressed); !ok || got.NewStepIdx != 1 || got.Banked != 0 {
		t.Errorf("turn 2 = %+v, want the jump made with the banked points and nothing left", second)
	}

	rated := tickOne(t, laneOrder(5, 0), driftRulebook(3))
	if _, ok := rated.(domain.MovementOrderCompleted); !ok {
		t.Errorf("rating 5 = %+v, want the 2-point jump and the last step in one turn", rated)
	}
}
```

`internal/faction/engine/mutation/mutation_test.go`, append:

```go
func TestMutationEngine_MovementBanksPointsAndAHoldClearsThem(t *testing.T) {
	s := newMutationTestState()
	asset := s.Factions["f1"].Assets["a1"]
	path := []spatial.RegionHex{{RegionID: "a", Coord: spatial.HexCoord{Q: 0}}, {RegionID: "b", Coord: spatial.HexCoord{Q: 5}}}
	asset.CurrentOrder = &domain.MovementOrder{AssetID: "a1", Path: path}
	me := New(slog.New(slog.DiscardHandler))

	if err := me.Apply(s, []domain.Mutation{
		domain.MovementOrderProgressed{FactionID: "f1", AssetID: "a1", NewStepIdx: 0, RegionHex: path[0], Banked: 4},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if asset.CurrentOrder.Banked != 4 {
		t.Errorf("Banked = %d, want 4", asset.CurrentOrder.Banked)
	}

	if err := me.Apply(s, []domain.Mutation{
		domain.MovementOrderHeld{FactionID: "f1", AssetID: "a1", WorldID: "Anchorage", NewStepIdx: 0, RegionHex: path[0], Reason: domain.HoldLockdownPaid},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if asset.CurrentOrder.Banked != 0 {
		t.Errorf("Banked after a hold = %d, want 0", asset.CurrentOrder.Banked)
	}
}
```

`internal/faction/engine/testharness/scenarios/transport_lifecycle_test.go`: `TestTransport_ReviseMidFlight`'s `pathToD` runs through neighbouring hexes (2,0), (2,1), (2,2) instead of jumping to (5,0) (Decision Record 1).

##### Commit message

```
feat(E4): make a Drift jump cost its rated movement points

- the movement tick spends banked points plus speed: a hex step
  costs 1, a lane jump its DriftCost; leftover points are banked
- MovementOrder.Banked and MovementOrderProgressed.Banked carry
  them; a lockdown hold clears them
- spatial: Adjacent and RegionMap.RegionName
- applies to every rulebook: drift_costs now costs time, as its
  settings comment always said
```

#### Commit 3 — `feat(E4): say where an asset in transit is`

##### Task 1 — `internal/faction/tui/views/turn/overlay/movement.go`

A shared label for an asset's whereabouts (Decision 19). Add after `worldNames`:

```go
// LocationLabel names where an asset is: its place or, in transit, the Drift lane
// it is crossing or the region it is crossing toward its destination.
func LocationLabel(asset *domain.Asset, spatialMap *spatial.RegionMap) string {
	if asset.Location.WorldID != "" {
		if world, ok := spatialMap.Location(asset.Location.WorldID); ok {
			return world.Name()
		}
		return asset.Location.WorldID
	}
	here := spatialMap.RegionName(asset.Location.RegionHex.RegionID)
	order := asset.CurrentOrder
	if order == nil {
		return "In " + here
	}
	// deliberate: banked points mean a Drift jump is under way, not a hex step, because a hex step costs 1 and the tick takes it while any point is left.
	if order.Banked > 0 && order.StepIdx+1 < len(order.Path) {
		return fmt.Sprintf("In the Drift between %s and %s", here, spatialMap.RegionName(order.Path[order.StepIdx+1].RegionID))
	}
	destination := order.Destination.WorldID
	if world, ok := spatialMap.Location(destination); ok {
		destination = world.Name()
	}
	return fmt.Sprintf("In %s, en route to %s", here, destination)
}
```

In `NewMovement`, the form's "at" line uses it:

**Find:**
```go
	worldOpts := worldOptions(spatialMap)
	names := worldNames(spatialMap)
```
**Replace with:**
```go
	worldOpts := worldOptions(spatialMap)
```

**Find:**
```go
		loc := names[asset.Location.WorldID]
		if loc == "" {
			loc = asset.Location.WorldID
		}
```
**Replace with:**
```go
		loc := LocationLabel(asset, spatialMap)
```

`worldNames` then has no caller, so it is deleted, and `LocationLabel` takes its place in the file.

##### Task 2 — `internal/faction/tui/views/turn/execution/execution.go`

In `resolveMovables`, the movable line uses the same label:

**Find:**
```go
		location := names[asset.Location.WorldID]
		if location == "" {
			location = asset.Location.WorldID
		}
```
**Replace with:**
```go
		location := overlay.LocationLabel(asset, spatialMap)
```

##### Task 3 — `internal/faction/tui/views/turn/overlay/movement_test.go`

Append:

```go
func TestLocationLabel(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"regions.toml": "[[region]]\nid = \"c\"\nname = \"Concordia\"\nhexes = [[0,0],[1,0]]\n\n" +
			"[[region]]\nid = \"p\"\nname = \"The Prime Material\"\nhexes = [[5,0],[6,0]]\n\n" +
			"[[warp]]\nfrom_region = \"c\"\nfrom_q = 1\nfrom_r = 0\nto_region = \"p\"\nto_q = 5\nto_r = 0\n",
		"worlds.toml": "[[world]]\nid = \"world-a\"\nname = \"Alpha\"\ntech_level = 4\nregion = \"c\"\nhex_q = 0\nhex_r = 0\n\n" +
			"[[world]]\nid = \"world-b\"\nname = \"Beta\"\ntech_level = 4\nregion = \"p\"\nhex_q = 6\nhex_r = 0\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	spatialMap, err := spatial.LoadRegionMap(dir)
	if err != nil {
		t.Fatalf("LoadRegionMap: %v", err)
	}
	path := []spatial.RegionHex{
		{RegionID: "c", Coord: spatial.HexCoord{Q: 0}}, {RegionID: "c", Coord: spatial.HexCoord{Q: 1}},
		{RegionID: "p", Coord: spatial.HexCoord{Q: 5}}, {RegionID: "p", Coord: spatial.HexCoord{Q: 6}},
	}
	moving := func(banked int) *domain.Asset {
		return &domain.Asset{ID: "s1", Location: domain.Location{RegionHex: path[1]},
			CurrentOrder: &domain.MovementOrder{Path: path, StepIdx: 1, Banked: banked, Destination: domain.Location{WorldID: "world-b"}}}
	}

	cases := []struct {
		name  string
		asset *domain.Asset
		want  string
	}{
		{"on a place", &domain.Asset{Location: domain.Location{WorldID: "world-a"}}, "Alpha"},
		{"mid-jump", moving(3), "In the Drift between Concordia and The Prime Material"},
		{"en route", moving(0), "In Concordia, en route to Beta"},
	}
	for _, tc := range cases {
		if got := LocationLabel(tc.asset, spatialMap); got != tc.want {
			t.Errorf("%s: LocationLabel = %q, want %q", tc.name, got, tc.want)
		}
	}
}
```

##### Task 4 — this plan

Tick Phase 1 in the Overview: `- [ ] **Phase 1 — The travel engine**` becomes `- [x] **Phase 1 — The travel engine**`.

##### Commit message

```
feat(E4): say where an asset in transit is

- overlay.LocationLabel: a place's name; "In the Drift between X
  and Y" while a jump is under way; "In X, en route to Y" otherwise
- the movement form and the execution view use it; an asset in
  transit showed a blank before
- E4 plan: Phase 1 ticked
```

### Phase 2 — Running hot

**In plain words:** Speed effects can apply only to orders run hot, a new effect adds speed to a type of asset, a move order can be marked hot in the movement form, and an asset moving hot takes damage each turn it moves.

#### Commit 4 — `feat(E4): add hot-only speed effects and speed_bonus`

##### Task 1 — `internal/faction/domain/effect.go`

The new effect, trigger and filter (Decision 24).

**Find:**
```go
	SpecialCoinUpkeep         SpecialEffectType = "coin_upkeep"
```
**Replace with:**
```go
	SpecialCoinUpkeep         SpecialEffectType = "coin_upkeep"
	SpecialSpeedBonus         SpecialEffectType = "speed_bonus"
```

**Find:**
```go
	TriggerUpkeep                 EffectTrigger = "upkeep"
```
**Replace with:**
```go
	TriggerUpkeep                 EffectTrigger = "upkeep"
	TriggerHotMove                EffectTrigger = "hot_move"
```

**Find:**
```go
	FilterOpponentTag       EffectFilterName = "opponent_tag"
```
**Replace with:**
```go
	FilterOpponentTag       EffectFilterName = "opponent_tag"
	FilterHot               EffectFilterName = "hot"
```

**Find:**
```go
	// OpponentTag names a tag the faction on the other side of the roll holds
	// (Godslayers' die against Soul-Bonded).
	OpponentTag string
}
```
**Replace with:**
```go
	// OpponentTag names a tag the faction on the other side of the roll holds
	// (Godslayers' die against Soul-Bonded).
	OpponentTag string
	// Hot limits a speed effect to move orders run hot (Aether Burners).
	Hot bool
}
```

**Find:**
```go
	if filters.OpponentTag != "" {
		present = append(present, FilterOpponentTag)
	}
	return present
}
```
**Replace with:**
```go
	if filters.OpponentTag != "" {
		present = append(present, FilterOpponentTag)
	}
	if filters.Hot {
		present = append(present, FilterHot)
	}
	return present
}
```

**Find:**
```go
// UpkeepUnpaid names what an unpaid upkeep costs the tag's holder.
```
**Replace with:**
```go
// SpeedBonusParams adds Amount to the speed of the holder's matching assets.
type SpeedBonusParams struct {
	Amount int
}

// UpkeepUnpaid names what an unpaid upkeep costs the tag's holder.
```

**Find:**
```go
	CoinUpkeep       *CoinUpkeepParams
}
```
**Replace with:**
```go
	CoinUpkeep       *CoinUpkeepParams
	SpeedBonus       *SpeedBonusParams
}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	CoinUpkeep       *coinUpkeepRecord       `toml:"coin_upkeep"`
}
```
**Replace with:**
```go
	CoinUpkeep       *coinUpkeepRecord       `toml:"coin_upkeep"`
	SpeedBonus       *speedBonusRecord       `toml:"speed_bonus"`
}
```

**Find:**
```go
	OpponentTag       string   `toml:"opponent_tag"`
}
```
**Replace with:**
```go
	OpponentTag       string   `toml:"opponent_tag"`
	Hot               bool     `toml:"hot"`
}
```

**Find:**
```go
type coinUpkeepRecord struct {
```
**Replace with:**
```go
type speedBonusRecord struct {
	Amount int `toml:"amount"`
}

type coinUpkeepRecord struct {
```

**Find:**
```go
		domain.TriggerArrivalQuery, domain.TriggerDriftRatingQuery, domain.TriggerUpkeep,
	}
```
**Replace with:**
```go
		domain.TriggerArrivalQuery, domain.TriggerDriftRatingQuery, domain.TriggerUpkeep,
		domain.TriggerHotMove,
	}
```

**Find:**
```go
		domain.SpecialDriftRatingBonus, domain.SpecialCoinUpkeep,
	}
```
**Replace with:**
```go
		domain.SpecialDriftRatingBonus, domain.SpecialCoinUpkeep, domain.SpecialSpeedBonus,
	}
```

**Find:**
```go
		OpponentTag: record.OpponentTag,
	}
```
**Replace with:**
```go
		OpponentTag: record.OpponentTag, Hot: record.Hot,
	}
```

**Find:**
```go
	if record.CoinUpkeep != nil {
		present = append(present, domain.SpecialCoinUpkeep)
	}
	return present
}
```
**Replace with:**
```go
	if record.CoinUpkeep != nil {
		present = append(present, domain.SpecialCoinUpkeep)
	}
	if record.SpeedBonus != nil {
		present = append(present, domain.SpecialSpeedBonus)
	}
	return present
}
```

**Find:**
```go
		definition.CoinUpkeep = &domain.CoinUpkeepParams{Amount: record.CoinUpkeep.Amount, OnUnpaid: onUnpaid, Damage: record.CoinUpkeep.Damage}
	}
	return nil
}
```
**Replace with:**
```go
		definition.CoinUpkeep = &domain.CoinUpkeepParams{Amount: record.CoinUpkeep.Amount, OnUnpaid: onUnpaid, Damage: record.CoinUpkeep.Damage}
	case domain.SpecialSpeedBonus:
		if record.SpeedBonus == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.SpeedBonus.Amount <= 0 {
			return fmt.Errorf("speed_bonus amount must be > 0, got %d", record.SpeedBonus.Amount)
		}
		definition.SpeedBonus = &domain.SpeedBonusParams{Amount: record.SpeedBonus.Amount}
	}
	return nil
}
```

##### Task 3 — `internal/faction/engine/hooks/rule_modifier.go`

The speed hook learns whether the order runs hot.

**Find:**
```go
// AssetSpeedModifier adjusts the speed, in hexes a turn, of an asset faction
// owns, read by the movement phase (Cat 4).
type AssetSpeedModifier interface {
	ModifyAssetSpeed(faction *domain.Faction, def *domain.AssetDefinition, speed int) int
}
```
**Replace with:**
```go
// AssetSpeedModifier adjusts the speed, in movement points a turn, of an asset
// faction owns, read by the movement phase (Cat 4). hot reports a move order run
// hot (Aether Burners).
type AssetSpeedModifier interface {
	ModifyAssetSpeed(faction *domain.Faction, def *domain.AssetDefinition, speed int, hot bool) int
}
```

##### Task 4 — `internal/faction/engine/hooks/dispatch/rules.go`

**Find:**
```go
func AssetSpeedResolver(registry *hooks.Registry, faction *domain.Faction) func(*domain.AssetDefinition) int {
	if registry == nil {
		return func(definition *domain.AssetDefinition) int { return definition.Speed }
	}
	modifiers := registry.AssetSpeedModifiersFor(faction.ID)
	for _, asset := range domain.SortedAssets(faction) {
		modifiers = append(modifiers, registry.AssetSpeedModifiersForAsset(faction.ID, asset.ID)...)
	}
	return func(definition *domain.AssetDefinition) int {
		speed := definition.Speed
		for _, registered := range modifiers {
			speed = registered.Hook.ModifyAssetSpeed(faction, definition, speed)
		}
		return speed
	}
}
```
**Replace with:**
```go
func AssetSpeedResolver(registry *hooks.Registry, faction *domain.Faction) func(definition *domain.AssetDefinition, hot bool) int {
	if registry == nil {
		return func(definition *domain.AssetDefinition, _ bool) int { return definition.Speed }
	}
	modifiers := registry.AssetSpeedModifiersFor(faction.ID)
	for _, asset := range domain.SortedAssets(faction) {
		modifiers = append(modifiers, registry.AssetSpeedModifiersForAsset(faction.ID, asset.ID)...)
	}
	return func(definition *domain.AssetDefinition, hot bool) int {
		speed := definition.Speed
		for _, registered := range modifiers {
			speed = registered.Hook.ModifyAssetSpeed(faction, definition, speed, hot)
		}
		return speed
	}
}
```

The doc comment's first line, `// AssetSpeedResolver returns the speed of an asset faction owns, by definition:`, becomes `// AssetSpeedResolver returns the speed of an asset faction owns, by definition and by whether its order runs hot:`.

##### Task 5 — `internal/faction/engine/effect/effects/speed_floor.go`

The floor accepts the `hot` filter.

**Find:**
```go
		domain.TriggerSpeedQuery: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
```
**Replace with:**
```go
		domain.TriggerSpeedQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterHot},
		},
```

**Find:**
```go
			speed:  ctx.Entry.SpeedFloor.Speed,
		},
```
**Replace with:**
```go
			speed:  ctx.Entry.SpeedFloor.Speed,
			hot:    ctx.Entry.Filters.Hot,
		},
```

**Find:**
```go
type speedFloorModifier struct {
	owner  *domain.Faction
	holder holderRef
	speed  int
}

func (modifier *speedFloorModifier) ModifyAssetSpeed(faction *domain.Faction, _ *domain.AssetDefinition, speed int) int {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || faction.ID != modifier.owner.ID {
		return speed
	}
	return max(speed, modifier.speed)
}
```
**Replace with:**
```go
type speedFloorModifier struct {
	owner  *domain.Faction
	holder holderRef
	speed  int
	hot    bool
}

func (modifier *speedFloorModifier) ModifyAssetSpeed(faction *domain.Faction, _ *domain.AssetDefinition, speed int, hot bool) int {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || faction.ID != modifier.owner.ID {
		return speed
	}
	if modifier.hot && !hot {
		return speed
	}
	return max(speed, modifier.speed)
}
```

The registrar's comment gains a sentence at its end: `The hot filter limits the floor to orders run hot (Aether Burners).`

##### Task 6 — `internal/faction/engine/effect/effects/speed_bonus.go` and `registrars.go`

`internal/faction/engine/effect/effects/speed_bonus.go` (new) — full contents:

```go
package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// speedBonusRegistrar reads owner and scope as speed_floor does: every asset the
// owner holds. asset_type narrows the bonus to those types, and hot to orders
// run hot (Aether Burners).
var speedBonusRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerSpeedQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterAssetType, domain.FilterHot},
		},
	},
	Register: registerSpeedBonus,
}

func registerSpeedBonus(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetSpeedModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialSpeedBonus),
		&speedBonusModifier{
			owner:  ctx.Holder.Faction,
			holder: refOf(ctx.Holder),
			amount: ctx.Entry.SpeedBonus.Amount,
			types:  ctx.Entry.Filters.AssetTypes,
			hot:    ctx.Entry.Filters.Hot,
		},
	)
}

// speedBonusModifier holds its owner, as speedFloorModifier does: a Cat 4 hook
// gets no FactionState.
type speedBonusModifier struct {
	owner  *domain.Faction
	holder holderRef
	amount int
	types  []domain.AssetType
	hot    bool
}

func (modifier *speedBonusModifier) ModifyAssetSpeed(faction *domain.Faction, def *domain.AssetDefinition, speed int, hot bool) int {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || faction.ID != modifier.owner.ID {
		return speed
	}
	if modifier.hot && !hot {
		return speed
	}
	if len(modifier.types) > 0 && !slices.Contains(modifier.types, def.Type) {
		return speed
	}
	return speed + modifier.amount
}
```

`internal/faction/engine/effect/effects/registrars.go`:

**Find:**
```go
		domain.SpecialCoinUpkeep:         coinUpkeepRegistrar,
```
**Replace with:**
```go
		domain.SpecialCoinUpkeep:         coinUpkeepRegistrar,
		domain.SpecialSpeedBonus:         speedBonusRegistrar,
```

##### Task 7 — the resolver's callers

`internal/faction/engine/orchestrator.go`, in `eligibleMovableAssets`: an asset that can move only hot is listed too.

**Find:**
```go
		if asset.CurrentOrder != nil || speedOf(def) > 0 {
```
**Replace with:**
```go
		if asset.CurrentOrder != nil || speedOf(def, false) > 0 || speedOf(def, true) > 0 {
```

`internal/faction/engine/world/movement.go`, in `TickMovementOrders` (Commit 5 swaps `false` for the order's flag):

**Find:**
```go
		points := order.Banked + speedOf(def)
```
**Replace with:**
```go
		points := order.Banked + speedOf(def, false)
```

##### Task 8 — tests

`internal/faction/engine/hooks/dispatch/rules_test.go`:

**Find:**
```go
func (stub stubAssetSpeedModifier) ModifyAssetSpeed(_ *domain.Faction, _ *domain.AssetDefinition, speed int) int {
```
**Replace with:**
```go
func (stub stubAssetSpeedModifier) ModifyAssetSpeed(_ *domain.Faction, _ *domain.AssetDefinition, speed int, _ bool) int {
```

In `TestAssetSpeedResolver`, the three calls `AssetSpeedResolver(nil, faction)(definition)` and `AssetSpeedResolver(registry, faction)(definition)` (twice) gain the argument `false`: `(definition, false)`.

`internal/faction/engine/orchestrator_test.go`:

**Find:**
```go
func (floor speedFloor) ModifyAssetSpeed(_ *domain.Faction, _ *domain.AssetDefinition, speed int) int {
```
**Replace with:**
```go
func (floor speedFloor) ModifyAssetSpeed(_ *domain.Faction, _ *domain.AssetDefinition, speed int, _ bool) int {
```

`internal/faction/engine/world/world_test.go`: its `speedFloor` stub's `ModifyAssetSpeed` gains `_ bool` the same way (Decision Record 2).

`internal/faction/engine/effect/effects/speed_floor_test.go`: the three resolver calls gain `false` — `(tc.definition)` becomes `(tc.definition, false)`, and `(facility)` becomes `(facility, false)` twice.

`internal/faction/engine/effect/effects/speed_bonus_test.go` (new) — full contents:

```go
package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

// TestSpeedBonus_RunningHot registers Aether Burners' two hot speed entries: +5
// for Aetherships, and a floor of 5 for everything.
func TestSpeedBonus_RunningHot(t *testing.T) {
	burners := &domain.Tag{ID: "HAWK-T-003", Name: "Aether Burners"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{burners.ID}}
	beta := &domain.Faction{ID: "beta"}
	registry := hooks.NewRegistry()
	holder := effect.Holder{Faction: alpha, Tag: burners}
	effects.Registrars()[domain.SpecialSpeedBonus].Register(effect.RegisterContext{
		Holder: holder, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger:    domain.TriggerSpeedQuery,
			Effect:     domain.SpecialSpeedBonus,
			Filters:    domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, AssetTypes: []domain.AssetType{"Aethership"}, Hot: true},
			SpeedBonus: &domain.SpeedBonusParams{Amount: 5},
		},
	})
	effects.Registrars()[domain.SpecialSpeedFloor].Register(effect.RegisterContext{
		Holder: holder, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger:    domain.TriggerSpeedQuery,
			Effect:     domain.SpecialSpeedFloor,
			Filters:    domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Hot: true},
			SpeedFloor: &domain.SpeedFloorParams{Speed: 5},
		},
	})
	frigate := &domain.AssetDefinition{ID: "frigate", Type: "Aethership", Speed: 14}
	troops := &domain.AssetDefinition{ID: "troops", Type: "Operative"}

	cases := []struct {
		name       string
		faction    *domain.Faction
		definition *domain.AssetDefinition
		hot        bool
		want       int
	}{
		{"an Aethership running hot gains 5", alpha, frigate, true, 19},
		{"an Aethership at normal speed gains nothing", alpha, frigate, false, 14},
		{"troops running hot move at the floor", alpha, troops, true, 5},
		{"troops at normal speed do not move", alpha, troops, false, 0},
		{"a rival's Aethership gains nothing", beta, frigate, true, 14},
	}
	for _, tc := range cases {
		if got := dispatch.AssetSpeedResolver(registry, tc.faction)(tc.definition, tc.hot); got != tc.want {
			t.Errorf("%s: speed = %d, want %d", tc.name, got, tc.want)
		}
	}

	alpha.LapsedTags = []string{burners.ID}
	if got := dispatch.AssetSpeedResolver(registry, alpha)(frigate, true); got != 14 {
		t.Errorf("speed while the tag is lapsed = %d, want 14", got)
	}
}
```

`internal/faction/rulebook/effects_test.go`, append:

```go
func TestConvertEffect_RunningHot(t *testing.T) {
	bonus, err := convertEffect(effectRecord{
		Trigger:    "speed_query",
		Effect:     "speed_bonus",
		Filters:    &effectFiltersRecord{Owner: "own", Scope: "faction", AssetType: []string{"Aethership"}, Hot: true},
		SpeedBonus: &speedBonusRecord{Amount: 5},
	})
	if err != nil {
		t.Fatalf("speed_bonus: %v", err)
	}
	if bonus.SpeedBonus == nil || bonus.SpeedBonus.Amount != 5 || !bonus.Filters.Hot {
		t.Errorf("speed_bonus = %+v, want amount 5 and the hot filter", bonus)
	}

	burn, err := convertEffect(effectRecord{
		Trigger:    "hot_move",
		Effect:     "self_damage",
		Filters:    &effectFiltersRecord{Owner: "own", Scope: "faction"},
		SelfDamage: &amountRecord{Amount: "1d4"},
	})
	if err != nil {
		t.Fatalf("hot_move self_damage: %v", err)
	}
	if burn.Trigger != domain.TriggerHotMove {
		t.Errorf("trigger = %q, want %q", burn.Trigger, domain.TriggerHotMove)
	}

	if _, err := convertEffect(effectRecord{
		Trigger: "speed_query", Effect: "speed_bonus",
		Filters: &effectFiltersRecord{Owner: "own", Scope: "faction"}, SpeedBonus: &speedBonusRecord{Amount: 0},
	}); err == nil {
		t.Error("speed_bonus amount 0: want an error")
	}
}
```
##### Commit message

```
feat(E4): add hot-only speed effects and speed_bonus

- the speed hook and AssetSpeedResolver take whether the order runs
  hot; a speed entry with filters.hot applies only then
- speed_bonus: adds an amount to the holder's assets, narrowed by
  asset_type and hot
- the hot_move trigger joins the vocabulary (Commit 6 uses it)
- an asset that can move only hot is offered in the movement phase
```

#### Commit 5 — `feat(E4): let a move order run hot`

##### Task 1 — `internal/faction/domain/location.go`

**Find:**
```go
	Banked int `toml:"banked,omitempty" json:"banked,omitempty"`
```
**Replace with:**
```go
	Banked int `toml:"banked,omitempty" json:"banked,omitempty"`
	// RunHot marks an order run hot: hot-only speed effects apply, and hot_move
	// effects fire each turn it moves (Aether Burners).
	RunHot bool `toml:"run_hot,omitempty" json:"run_hot,omitempty"`
```

##### Task 2 — `internal/faction/engine/world/movement.go`

The decision carries the choice. An asset that cannot move at all unless hot always runs hot (Decision 24). The ticker reads the order's flag.

**Find:**
```go
type MovementDecision struct {
	AssetID       string
	Kind          MovementDecisionKind
```
**Replace with:**
```go
type MovementDecision struct {
	AssetID       string
	Kind          MovementDecisionKind
	RunHot        bool
```

**Find:**
```go
		points := order.Banked + speedOf(def, false)
```
**Replace with:**
```go
		points := order.Banked + speedOf(def, order.RunHot)
```

**Find:**
```go
	driftRatingOf := dispatch.AssetDriftRatingResolver(registry, faction)
	for _, decision := range decisions {
```
**Replace with:**
```go
	driftRatingOf := dispatch.AssetDriftRatingResolver(registry, faction)
	speedOf := dispatch.AssetSpeedResolver(registry, faction)
	// runsHot is the decision's choice, or forced when the asset moves only hot.
	runsHot := func(decision MovementDecision, def *domain.AssetDefinition) bool {
		return decision.RunHot || speedOf(def, false) == 0
	}
	for _, decision := range decisions {
```

In the Issue branch:

**Find:**
```go
				StepIdx:       0,
				DriftRating:   driftRating,
				CargoAssetIDs: decision.CargoAssetIDs,
			}
```
**Replace with:**
```go
				StepIdx:       0,
				DriftRating:   driftRating,
				CargoAssetIDs: decision.CargoAssetIDs,
				RunHot:        runsHot(decision, def),
			}
```

In the Revise branch:

**Find:**
```go
				DriftRating:   driftRating,
				CargoAssetIDs: asset.CurrentOrder.CargoAssetIDs,
			}
```
**Replace with:**
```go
				DriftRating:   driftRating,
				CargoAssetIDs: asset.CurrentOrder.CargoAssetIDs,
				RunHot:        runsHot(decision, def),
			}
```

##### Task 3 — `internal/faction/rulebook/hot.go`

`internal/faction/rulebook/hot.go` (new) — full contents:

```go
package rulebook

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// RunsHot reports whether faction holds a live tag with a hot-only effect, so its
// move orders may run hot (Aether Burners). A lapsed tag does not count.
func (rulebook *Rulebook) RunsHot(faction *domain.Faction) bool {
	if faction == nil {
		return false
	}
	for _, tagID := range faction.Tags {
		tag, ok := rulebook.Tags[tagID]
		if !ok || slices.Contains(faction.LapsedTags, tagID) {
			continue
		}
		for _, entry := range tag.Effects {
			if entry.Filters.Hot {
				return true
			}
		}
	}
	return false
}
```

##### Task 4 — `internal/faction/tui/views/turn/overlay/movement.go`

Each row can run hot. The choice sits in the destination group, shown only to a faction that can run hot.

**Find:**
```go
type assetRow struct {
	assetID string
	kind    orderKind
	worldID string
}
```
**Replace with:**
```go
type assetRow struct {
	assetID string
	kind    orderKind
	worldID string
	hot     bool
}
```

**Find:**
```go
func NewMovement(eligible []*domain.Asset, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) Movement {
	data := &movementData{rows: make([]assetRow, len(eligible))}
	worldOpts := worldOptions(spatialMap)
```
**Replace with:**
```go
func NewMovement(faction *domain.Faction, eligible []*domain.Asset, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) Movement {
	data := &movementData{rows: make([]assetRow, len(eligible))}
	worldOpts := worldOptions(spatialMap)
	canRunHot := rb.RunsHot(faction)
```

**Find:**
```go
		destGroup := huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("%s → destination", assetName(asset, rb))).
				Options(worldOpts...).
				Value(&data.rows[idx].worldID),
		).WithHideFunc(func() bool {
```
**Replace with:**
```go
		destFields := []huh.Field{
			huh.NewSelect[string]().
				Title(fmt.Sprintf("%s → destination", assetName(asset, rb))).
				Options(worldOpts...).
				Value(&data.rows[idx].worldID),
		}
		if canRunHot {
			destFields = append(destFields, huh.NewConfirm().
				Title("Run hot?").
				Description("Faster, but the asset takes damage each turn it moves.").
				Value(&data.rows[idx].hot))
		}
		destGroup := huh.NewGroup(destFields...).WithHideFunc(func() bool {
```

**Find:**
```go
			out = append(out, world.MovementDecision{AssetID: r.assetID, Kind: kind, Destination: dest})
```
**Replace with:**
```go
			out = append(out, world.MovementDecision{AssetID: r.assetID, Kind: kind, Destination: dest, RunHot: r.hot})
```

##### Task 5 — the callers of `NewMovement`

`internal/faction/tui/views/turn/execution/execution.go`:

**Find:**
```go
		return overlay.NewMovement(msg.Payload.(adapter.SelectMovementDecisionsPayload).Eligible, m.rulebook, m.spatialMap)
```
**Replace with:**
```go
		return overlay.NewMovement(msg.Faction, msg.Payload.(adapter.SelectMovementDecisionsPayload).Eligible, m.rulebook, m.spatialMap)
```

`internal/faction/tui/views/turn/overlay/movement_test.go`: both `NewMovement(eligible, rb, spatialMap)` calls become `NewMovement(&domain.Faction{ID: "f1"}, eligible, rb, spatialMap)`.

##### Task 6 — tests

`internal/faction/engine/world/movement_test.go`, append. Revise is used because it needs no world on the map:

```go
func TestBuildMovementMutations_RunHot(t *testing.T) {
	rb := &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"ship":   {ID: "ship", Speed: 2},
			"troops": {ID: "troops"},
		},
		DriftCosts: []int{5},
	}
	order := func(id, def string) *domain.Asset {
		return &domain.Asset{ID: id, DefinitionID: def, OwnerID: "alpha",
			Location:     domain.Location{RegionHex: spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: 1}}},
			CurrentOrder: &domain.MovementOrder{AssetID: id, Path: makePath(5), StepIdx: 1}}
	}
	faction := &domain.Faction{ID: "alpha", Coin: 5, Assets: map[string]*domain.Asset{
		"cold": order("cold", "ship"), "hot": order("hot", "ship"), "forced": order("forced", "troops"),
	}}
	dest := &domain.Location{RegionHex: spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: 3}}}
	decisions := []world.MovementDecision{
		{AssetID: "cold", Kind: world.MovementDecisionRevise, Destination: dest},
		{AssetID: "hot", Kind: world.MovementDecisionRevise, Destination: dest, RunHot: true},
		{AssetID: "forced", Kind: world.MovementDecisionRevise, Destination: dest},
	}

	mutations, err := regionWorldEngine(t).BuildMovementMutations(decisions, faction, rb, nil, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("BuildMovementMutations: %v", err)
	}
	want := map[string]bool{"cold": false, "hot": true, "forced": true}
	for _, mutation := range mutations {
		if revised, ok := mutation.(domain.MovementOrderRevised); ok && revised.NewOrder.RunHot != want[revised.AssetID] {
			t.Errorf("%s: RunHot = %v, want %v", revised.AssetID, revised.NewOrder.RunHot, want[revised.AssetID])
		}
	}
}
```

`internal/faction/rulebook/rulebook_test.go`, append:

```go
func TestRunsHot(t *testing.T) {
	rb := &Rulebook{Tags: map[string]*domain.Tag{
		"burners": {ID: "burners", Effects: []domain.EffectDefinition{{Filters: domain.EffectFilters{Hot: true}}}},
		"plain":   {ID: "plain", Effects: []domain.EffectDefinition{{}}},
	}}
	cases := []struct {
		name    string
		faction *domain.Faction
		want    bool
	}{
		{"holds a hot tag", &domain.Faction{Tags: []string{"plain", "burners"}}, true},
		{"holds no hot tag", &domain.Faction{Tags: []string{"plain"}}, false},
		{"the hot tag is lapsed", &domain.Faction{Tags: []string{"burners"}, LapsedTags: []string{"burners"}}, false},
		{"no faction", nil, false},
	}
	for _, tc := range cases {
		if got := rb.RunsHot(tc.faction); got != tc.want {
			t.Errorf("%s: RunsHot = %v, want %v", tc.name, got, tc.want)
		}
	}
}
```

`internal/faction/tui/views/turn/overlay/movement_test.go`, append:

```go
func TestMovementDecisionsCarryRunHot(t *testing.T) {
	spatialMap := testRegionMap(t)
	rb := &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{"scout": {ID: "scout", Name: "Scout"}},
		Tags:   map[string]*domain.Tag{"burners": {ID: "burners", Effects: []domain.EffectDefinition{{Filters: domain.EffectFilters{Hot: true}}}}},
	}
	eligible := []*domain.Asset{{ID: "a-1", DefinitionID: "scout", Location: domain.Location{WorldID: "world-a"}}}

	o := NewMovement(&domain.Faction{ID: "f1", Tags: []string{"burners"}}, eligible, rb, spatialMap)
	o.data.rows[0].kind, o.data.rows[0].worldID, o.data.rows[0].hot = orderIssue, "world-b", true

	got := o.decisions()
	if len(got) != 1 || !got[0].RunHot {
		t.Errorf("decisions() = %+v, want one decision run hot", got)
	}
}
```

##### Commit message

```
feat(E4): let a move order run hot

- MovementDecision.RunHot and MovementOrder.RunHot; an asset that
  moves only hot always runs hot; the tick reads the order's flag
- Rulebook.RunsHot: a faction with a live hot-only tag effect
- the movement form asks "Run hot?" per order, for such a faction
```

#### Commit 6 — `feat(E4): burn assets that move hot`

##### Task 1 — `internal/faction/engine/effect/effects/self_damage.go`

`self_damage` accepts `hot_move` at scope `faction`. The lethal handling moves into a shared `damageSelf`, which the Zealots' reactor and the new burn both use, and which rolls only after the removed check, as before.

**Find:**
```go
var selfDamageRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnAttackHit:  selfDamageAccepts,
		domain.TriggerOwnCounterHit: selfDamageAccepts,
	},
	Register: registerSelfDamage,
}

func registerSelfDamage(ctx effect.RegisterContext) {
```
**Replace with:**
```go
var selfDamageRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnAttackHit:  selfDamageAccepts,
		domain.TriggerOwnCounterHit: selfDamageAccepts,
		domain.TriggerHotMove: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerSelfDamage,
}

func registerSelfDamage(ctx effect.RegisterContext) {
	if ctx.Entry.Trigger == domain.TriggerHotMove {
		ctx.Registry.RegisterMutationReactor(
			ctx.Holder.Scope(),
			ctx.Holder.HookID(domain.SpecialSelfDamage),
			&hotBurnReactor{
				holder:   refOf(ctx.Holder),
				amount:   ctx.Entry.SelfDamage.Amount,
				roller:   ctx.Roller,
				registry: ctx.Registry,
			},
		)
		return
	}
```

Replace the whole of `func (reactor *selfDamageReactor) OnMutations(...)`, from its signature line to its closing brace (it ends with `return extra` and `}` just above `// batchHP returns asset's HP`), with:

```go
func (reactor *selfDamageReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if len(hitTargets(mutations, reactor.holder.assetID, reactor.counter)) == 0 {
		return nil
	}
	asset, live := reactor.holder.live(factionState)
	if !live {
		return nil
	}
	return damageSelf(reactor.registry, factionState, mutations, reactor.holder.factionID, asset, reactor.def, reactor.amount, reactor.roller)
}

// hotBurnReactor is running hot's cost: each asset of its owner that a movement
// tick moves on an order run hot takes the entry's damage (Aether Burners).
type hotBurnReactor struct {
	holder   holderRef
	amount   domain.DiceRoll
	roller   dice.Roller
	registry *hooks.Registry
}

func (reactor *hotBurnReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	owner := factionState.Factions[reactor.holder.factionID]
	var extra []domain.Mutation
	for _, asset := range hotMovers(mutations, owner) {
		batch := append(slices.Clone(mutations), extra...)
		extra = append(extra, damageSelf(reactor.registry, factionState, batch, owner.ID, asset, rulebook.Assets[asset.DefinitionID], reactor.amount, reactor.roller)...)
	}
	return extra
}

// hotMovers returns owner's assets that a movement tick in mutations moved on an
// order run hot: a step taken, points banked toward a Drift jump, an arrival, or a
// hold at a locked place.
func hotMovers(mutations []domain.Mutation, owner *domain.Faction) []*domain.Asset {
	var movers []*domain.Asset
	for _, mutation := range mutations {
		var factionID, assetID string
		switch m := mutation.(type) {
		case domain.MovementOrderProgressed:
			factionID, assetID = m.FactionID, m.AssetID
		case domain.MovementOrderCompleted:
			factionID, assetID = m.FactionID, m.AssetID
		case domain.MovementOrderHeld:
			factionID, assetID = m.FactionID, m.AssetID
		}
		if factionID != owner.ID {
			continue
		}
		if asset, ok := owner.Assets[assetID]; ok && asset.CurrentOrder != nil && asset.CurrentOrder.RunHot {
			movers = append(movers, asset)
		}
	}
	return movers
}

// damageSelf is a hit an asset deals itself, rolled from amount: the HP loss and,
// when it is lethal, an automatic interceptor's save or the asset's removal.
// mutations is the batch the hit joins, read for damage already under way.
func damageSelf(registry *hooks.Registry, factionState *state.FactionState, mutations []domain.Mutation, factionID string, asset *domain.Asset, def *domain.AssetDefinition, amount domain.DiceRoll, roller dice.Roller) []domain.Mutation {
	effectiveHP, removed := batchHP(mutations, asset)
	if removed {
		return nil
	}
	damage := dice.Amount(amount, roller)
	extra := []domain.Mutation{domain.AssetHPDelta{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Delta:             -damage,
		Cause:             string(domain.SpecialSelfDamage),
		CausedByFactionID: factionID,
	}}
	if effectiveHP-damage > 0 {
		return extra
	}
	// deliberate: only automatic interceptors, not elective ones, because a reactor has no collector to ask (D-10).
	saved := dispatch.InterceptLethalAutomatic(registry, hooks.LethalContext{
		Target:       asset,
		TargetDef:    def,
		HPAfter:      effectiveHP - damage,
		Damage:       damage,
		FactionState: factionState,
		Batch:        append(slices.Clone(mutations), extra...),
	})
	if saved != nil {
		return append(extra, saved...)
	}
	return append(extra, domain.AssetRemoved{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Cause:             string(domain.SpecialSelfDamage),
		CausedByFactionID: factionID,
	})
}
```

##### Task 2 — `internal/faction/engine/effect/effects/self_damage_test.go`

Append. The existing Zealots tests keep `selfDamageReactor`'s behaviour pinned through the refactor.

```go
func TestSelfDamage_HotMoveBurnsEachHotMover(t *testing.T) {
	burners := &domain.Tag{ID: "HAWK-T-003", Name: "Aether Burners"}
	hot := &domain.Asset{ID: "frigate", OwnerID: "alpha", DefinitionID: "ship", CurrentHP: 8,
		CurrentOrder: &domain.MovementOrder{AssetID: "frigate", RunHot: true}}
	cold := &domain.Asset{ID: "hauler", OwnerID: "alpha", DefinitionID: "ship", CurrentHP: 8,
		CurrentOrder: &domain.MovementOrder{AssetID: "hauler"}}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{burners.ID}, Assets: map[string]*domain.Asset{hot.ID: hot, cold.ID: cold}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialSelfDamage].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: burners}, Registry: registry, Roller: dicetest.Constant{Value: 3},
		Entry: domain.EffectDefinition{
			Trigger:    domain.TriggerHotMove,
			Effect:     domain.SpecialSelfDamage,
			Filters:    domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
			SelfDamage: &domain.SelfDamageParams{Amount: domain.DiceRoll{NumDice: 1, Sides: 4}},
		},
	})
	reactors := registry.MutationReactorsForFaction("alpha")
	if len(reactors) != 1 {
		t.Fatalf("reactors = %d, want 1", len(reactors))
	}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"ship": {ID: "ship"}}}

	got := reactors[0].Hook.OnMutations([]domain.Mutation{
		domain.MovementOrderProgressed{FactionID: "alpha", AssetID: "frigate", Banked: 2},
		domain.MovementOrderProgressed{FactionID: "alpha", AssetID: "hauler", NewStepIdx: 1},
	}, factionState, rb)

	want := []domain.Mutation{domain.AssetHPDelta{FactionID: "alpha", AssetID: "frigate", Delta: -3, Cause: string(domain.SpecialSelfDamage), CausedByFactionID: "alpha"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("burn = %+v, want %+v (the hot frigate only)", got, want)
	}

	alpha.LapsedTags = []string{burners.ID}
	if got := reactors[0].Hook.OnMutations([]domain.Mutation{domain.MovementOrderProgressed{FactionID: "alpha", AssetID: "frigate"}}, factionState, rb); got != nil {
		t.Errorf("burn while the tag is lapsed = %+v, want none", got)
	}
}
```

Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"` to the file's imports.

##### Task 3 — this plan

Tick Phase 2 in the Overview: `- [ ] **Phase 2 — Running hot**` becomes `- [x] **Phase 2 — Running hot**`.

##### Commit message

```
feat(E4): burn assets that move hot

- self_damage accepts hot_move at scope faction: each of the
  holder's assets that a tick moves on a hot order takes the damage
- damageSelf: the lethal handling shared by the Zealots' reactor
  and the burn
- E4 plan: Phase 2 ticked
```

### Phase 3 — The Force pass

**In plain words:** H.A.W.K's 28 Force assets replace the SWN copy at the new speeds, the Drift costs and Aether Burners take their new values, the seed's Force units become H.A.W.K's, the canon gets their fiction, and Robert plays a turn with them.

#### Commit 7 — `feat(E4): replace the Force stand-in with H.A.W.K's own assets`

##### Task 1 — `rulebooks/hawk/assets/force_assets.toml`

Rewritten — full contents:

```toml
# Force Assets — H.A.W.K (A.001.5 Effort 4)

[assets.CorporateSecurity]
id = "HAWK-F1-001"
name = "Corporate Security"
category = "Force"
min_rating = 1
hp = 3
cost = 2
maintenance = 0
tech_level = 0
type = "Security Personnel"
flags = []
counter = "1d4"
description = "Guards in company colours with ward-batons and lanyards, posted at every company property."
attack = { attacker_stat = "Force", defender_stat = "Force", damage = "1d3+1" }

[assets.UnsanctionedAdventurers]
id = "HAWK-F1-002"
name = "Unsanctioned Adventurers"
category = "Force"
min_rating = 1
hp = 1
cost = 2
maintenance = 0
tech_level = 0
type = "Operative"
flags = []
counter = ""
description = "Adventurers who failed H.A.W.K's sanctioning and kept the gear, hired cheap and deniably against a rival's leadership."
attack = { attacker_stat = "Force", defender_stat = "Cunning", damage = "1d6" }

[assets.ReluctantMilitia]
id = "HAWK-F1-003"
name = "Reluctant Militia"
category = "Force"
min_rating = 1
hp = 4
cost = 4
maintenance = 0
tech_level = 3
type = "Security Personnel"
flags = ["P"]
counter = "1d4+1"
description = "The people of a company town, issued rifles by the company that owns the town; service is a clause in every lease."
attack = { attacker_stat = "Force", defender_stat = "Force", damage = "1d6" }

[assets.ArcanoEngineeredWarbeasts]
id = "HAWK-F1-004"
name = "Arcano-Engineered Warbeasts"
category = "Force"
min_rating = 1
hp = 6
cost = 2
maintenance = 0
tech_level = 4
type = "Beast"
flags = []
requires_tag = "HAWK-T-011"
counter = "1d4"
description = "Owlbears, drakes and stranger stock, bred in Arcano-Tech laboratories and leased under patent. Only a faction with the Aether-Tech Evangelists tag can buy them."
attack = { attacker_stat = "Force", defender_stat = "Force", damage = "1d6" }

[assets.UnpaidInterns]
id = "HAWK-F1-005"
name = "Unpaid Interns"
category = "Force"
min_rating = 1
hp = 2
cost = 1
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["S"]
counter = ""
description = "Apprentice mages and junior engineers on placement, who stand nearest whatever is most expensive. If another asset on this place suffers enough damage to destroy it, the faction can sacrifice the Unpaid Interns instead to nullify the killing blow."
effects = [
  { trigger = "lethal_damage", effect = "sacrifice_self", filters = { owner = "own", scope = "world" } },
]

[assets.DriftHaulers]
id = "HAWK-F2-001"
name = "Drift Haulers"
category = "Force"
min_rating = 2
hp = 6
cost = 4
maintenance = 0
tech_level = 4
speed = 10
drift_rating = 2
type = "Facility"
flags = ["S"]
counter = ""
description = "Short-hop haulers in company livery, built on a single Aetherstone ring. Moves at speed 10 (short-range). When it moves, it can carry one non-Aethership asset from its place with it, for 1 Coin."
transport = {
  coin_cost = 1,
  cargo_types = ["Facility", "Security Personnel", "Operative", "Stratagem", "Beast"],
  max_cargo = 1,
}

[assets.AethericLanceCorps]
id = "HAWK-F2-002"
name = "Aetheric-Lance Corps"
category = "Force"
min_rating = 2
hp = 5
cost = 5
maintenance = 0
tech_level = 4
type = "Security Personnel"
flags = ["P"]
counter = "1d4+1"
description = "Light cavalry on Aetherbikes with lance-rifles, who answer alarms across a company's holdings and raid supply lines between calls."
attack = { attacker_stat = "Force", defender_stat = "Force", damage = "2d4" }

[assets.ShieldWardens]
id = "HAWK-F2-003"
name = "Shield Wardens"
category = "Force"
min_rating = 2
hp = 4
cost = 4
maintenance = 0
tech_level = 3
type = "Operative"
flags = ["S"]
counter = "1d4+1"
description = "Site staff trained in the shield wall, carrying tower shields with ward-plates. Once per turn, when a rival attacks this asset, the owner may force the attacker to reroll their highest die."
effects = [
  {
    trigger = "enemy_roll_at_me",
    effect = "reroll",
    filters = { owner = "rival", scope = "self" },
    gates = { once_per = "asset" },
    reroll = { dice = 1, elective = true },
  },
]

[assets.UnionizedPartisans]
id = "HAWK-F2-004"
name = "Unionized Partisans"
category = "Force"
min_rating = 2
hp = 6
cost = 4
maintenance = 0
tech_level = 0
type = "Security Personnel"
flags = []
counter = ""
description = "Workers organised in a guild hall and armed from its stores, who stop the work rather than hold ground."
attack = { attacker_stat = "Force", defender_stat = "Cunning", damage = "1d4+1" }

[assets.OverclockedZealots]
id = "HAWK-F3-001"
name = "Overclocked Zealots"
category = "Force"
min_rating = 3
hp = 4
cost = 6
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["S"]
counter = "2d6"
description = "True believers fitted with Aether conduits that run raw power through their bodies. Overclocked Zealots take 1d4 damage every time they launch a successful attack or perform a counterattack."
attack = { attacker_stat = "Force", defender_stat = "Force", damage = "2d6" }
effects = [
  {
    trigger = "own_attack_hit",
    effect = "self_damage",
    filters = { owner = "own", scope = "self" },
    self_damage = { amount = "1d4" },
  },
  {
    trigger = "own_counter_hit",
    effect = "self_damage",
    filters = { owner = "own", scope = "self" },
    self_damage = { amount = "1d4" },
  },
]

[assets.RunicMinefield]
id = "HAWK-F3-002"
name = "Runic Minefield"
category = "Force"
min_rating = 3
hp = 2
cost = 5
maintenance = 0
tech_level = 0
type = "Stratagem"
flags = []
counter = "1d6+3"
description = "Runes laid into a landing field, a dock or a Drift approach, keyed to the company's staff badges, that trigger when a hostile unit crosses them."

[assets.WizardAdvisors]
id = "HAWK-F3-003"
name = "Wizard Advisors"
category = "Force"
min_rating = 3
hp = 4
cost = 6
maintenance = 0
tech_level = 4
type = "Operative"
flags = ["S"]
counter = "1d6"
description = "Wizards retained as advisors, who scry each attack on the company's holdings before it lands. Once per turn, a Force defense by the faction on this place rolls an additional d10."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "1d4+1" }
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "world", phase = ["defense"], attribute = "Force" },
    gates = { once_per = "asset" },
    bonus_die = { sides = 10, count = 1 },
  },
]

[assets.CorporateChaplains]
id = "HAWK-F3-004"
name = "Corporate Chaplains"
category = "Force"
min_rating = 3
hp = 4
cost = 6
maintenance = 0
tech_level = 3
type = "Operative"
flags = ["S"]
counter = "1d4+1"
description = "Soul-bound chaplains on the company payroll, with a field shrine and a claims desk. Once between turns, if a Security Personnel, Operative or Beast asset of the faction on this place is destroyed, the faction may pay half its purchase cost to restore it at 1 HP."
effects = [
  {
    trigger = "lethal_damage",
    effect = "restore_asset",
    filters = { owner = "own", scope = "world", asset_type = ["Security Personnel", "Operative", "Beast"] },
    gates = { once_per = "asset" },
  },
]

[assets.WardedCarriers]
id = "HAWK-F4-001"
name = "Warded Carriers"
category = "Force"
min_rating = 4
hp = 10
cost = 10
maintenance = 0
tech_level = 4
speed = 14
drift_rating = 2
type = "Facility"
flags = ["S"]
counter = ""
description = "Long-hulled carriers with a ward-dome over the drop deck, built to move a whole regiment. Moves at speed 14 (mid-range). When it moves, it can carry up to ten assets from its place with it, for 1 Coin."
transport = {
  coin_cost = 1,
  cargo_types = ["Facility", "Aethership", "Security Personnel", "Operative", "Stratagem", "Beast"],
  max_cargo = 10,
}

[assets.TeleportationHubs]
id = "HAWK-F4-002"
name = "Teleportation Hubs"
category = "Force"
min_rating = 4
hp = 10
cost = 10
maintenance = 0
tech_level = 4
type = "Facility"
flags = ["A"]
counter = ""
description = "Paired teleportation circles inlaid in depot floors and kept by a coven under service contract. As an action, any one non-Aethership asset within eight hexes may be moved instantly to any place within eight hexes for 1 Coin."
ability = {
  effect = "relocate",
  relocate = { scope = "other", max_hex = 8, coin_cost = 1, exclude_types = ["Aethership", "Starship"], max_assets = 1 },
}

[assets.RuneboundFrigates]
id = "HAWK-F4-003"
name = "Runebound Frigates"
category = "Force"
min_rating = 4
hp = 8
cost = 12
maintenance = 0
tech_level = 4
speed = 14
drift_rating = 2
type = "Aethership"
flags = []
counter = "1d8"
description = "Angular warships built around a ringed Aether-Core, their hulls etched with binding runes. Moves at speed 14 (mid-range)."
attack = { attacker_stat = "Force", defender_stat = "Force", damage = "2d6" }

[assets.ArcaneStrikeCadre]
id = "HAWK-F4-004"
name = "Arcane Strike Cadre"
category = "Force"
min_rating = 4
hp = 12
cost = 8
maintenance = 0
tech_level = 4
type = "Security Personnel"
flags = ["P"]
counter = "1d8"
description = "The standing force of a company's security arm: drilled riflemen in ward-cloth uniforms, with a combat mage in every squad."
attack = { attacker_stat = "Force", defender_stat = "Force", damage = "1d8" }

[assets.HexcannonGunships]
id = "HAWK-F5-001"
name = "Hexcannon Gunships"
category = "Force"
min_rating = 5
hp = 8
cost = 10
maintenance = 0
tech_level = 4
speed = 14
drift_rating = 3
type = "Aethership"
flags = ["S"]
counter = ""
description = "Collections gunships whose hexcannons curse a ship's engines rather than breach its hull. Moves at speed 14 (mid-range). When they successfully attack an enemy asset, they steal 1d4 Coin from the target faction, once per turn per faction, no matter how many Hexcannon Gunships attack."
attack = { attacker_stat = "Force", defender_stat = "Wealth", damage = "1d6" }
effects = [
  {
    trigger = "own_attack_hit",
    effect = "coin_steal",
    filters = { owner = "rival", scope = "world" },
    gates = { once_per = "target_faction" },
    coin_steal = { amount = "1d4" },
  },
]

[assets.BondedArmory]
id = "HAWK-F5-002"
name = "Bonded Armory"
category = "Force"
min_rating = 5
hp = 6
cost = 14
maintenance = 0
tech_level = 0
type = "Facility"
flags = ["A"]
counter = ""
description = "A bonded warehouse of salvage, surplus and quietly acquired arms. As an action, the owner may buy one Force asset on this place that requires up to tech level 5, at half again as many Coin as usual, rounded up. Only one asset can be bought this way per turn."
ability = { effect = "purchase_enabler", purchase = { category = "Force", max_tech_level = 5, cost_percent = 150 } }

[assets.PactBoundAssassins]
id = "HAWK-F5-003"
name = "Pact-Bound Assassins"
category = "Force"
min_rating = 5
hp = 4
cost = 12
maintenance = 0
tech_level = 4
type = "Operative"
flags = ["S"]
counter = ""
description = "Warlocks whose patron beyond the Material granted them the power to pass unseen. Pact-Bound Assassins start Stealthed when bought."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "2d6+2" }
effects = [
  { trigger = "on_purchase", effect = "apply_stealth", filters = { owner = "own", scope = "self" } },
]

[assets.BattleChoir]
id = "HAWK-F5-004"
name = "Battle Choir"
category = "Force"
min_rating = 5
hp = 6
cost = 12
maintenance = 0
tech_level = 4
type = "Operative"
flags = ["S"]
counter = "1d6"
description = "Battle-sorcerers wired into portable Aether-Cores, chanting in unison behind the line. Once per turn, a Force attack by the faction on this place rolls an additional d10."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "world", phase = ["attack"], attribute = "Force" },
    gates = { once_per = "asset" },
    bonus_die = { sides = 10, count = 1 },
  },
]

[assets.CharteredMageKnights]
id = "HAWK-F6-001"
name = "Chartered Mage-Knights"
category = "Force"
min_rating = 6
hp = 16
cost = 20
maintenance = 0
tech_level = 5
type = "Security Personnel"
flags = ["P", "S"]
counter = "2d8+2"
description = "An order of mage-knights chartered to a company, in rune-forged plate and company livery. If they successfully defend against an attack, the owner may make the attacking asset attack itself for normal damage or counterattack results."
attack = { attacker_stat = "Force", defender_stat = "Force", damage = "2d8" }
effects = [
  { trigger = "own_defense_win", effect = "counterstrike", filters = { owner = "own", scope = "self" } },
]

[assets.ArcanoLinkedBatteries]
id = "HAWK-F6-002"
name = "Arcano-Linked Batteries"
category = "Force"
min_rating = 6
hp = 20
cost = 18
maintenance = 0
tech_level = 4
type = "Facility"
flags = ["S"]
counter = "2d6+6"
description = "Aether cannon and ward-towers linked by Arcano-Tech relays across a Fragment's sky. They can only defend against attacks by Aethership assets."
effects = [
  {
    trigger = "attack_eligibility_query",
    effect = "restrict_attack",
    filters = { owner = "rival", scope = "self", asset_type = ["Aethership"] },
    restrict_attack = { side = "defense", allow = "only_listed" },
  },
]

[assets.ForgedSiegeEngines]
id = "HAWK-F6-003"
name = "Forged Siege Engines"
category = "Force"
min_rating = 6
hp = 14
cost = 25
maintenance = 0
tech_level = 4
type = "Security Personnel"
flags = ["P"]
counter = "1d10"
description = "Forged war-frames the size of houses, hired out on fixed-term contracts to break fortified lines."
attack = { attacker_stat = "Force", defender_stat = "Force", damage = "2d10+4" }

[assets.DeepInsertionGate]
id = "HAWK-F7-001"
name = "Deep-Insertion Gate"
category = "Force"
min_rating = 7
hp = 10
cost = 25
maintenance = 0
tech_level = 4
type = "Facility"
flags = ["A"]
counter = ""
description = "A long-range teleportation gate on a priority clearance that overrides customs, wards and lockdowns. As an action, any one non-Aethership asset within fourteen hexes may be moved instantly to any place within fourteen hexes for 2 Coin, even if the local government objects or a lockdown holds the place."
ability = {
  effect = "relocate",
  relocate = {
    scope = "other",
    max_hex = 14,
    coin_cost = 2,
    exclude_types = ["Aethership", "Starship"],
    max_assets = 1,
    ignores_permission = true,
    ignores_lockdown = true,
  },
}

[assets.AbjurativeFirewall]
id = "HAWK-F7-002"
name = "Abjurative Firewall"
category = "Force"
min_rating = 7
hp = 10
cost = 20
maintenance = 0
tech_level = 5
type = "Facility"
flags = ["S"]
counter = "2d8+2"
description = "A lattice of abjuration wards and scrying mirrors over a whole facility. It can defend only against attacks versus Cunning, but adds an additional die to the defender's roll."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "self", phase = ["defense"], attribute = "Cunning" },
    bonus_die = { sides = 10, count = 1 },
  },
  {
    trigger = "attack_eligibility_query",
    effect = "restrict_attack",
    filters = { owner = "rival", scope = "self", attribute = "Cunning" },
    restrict_attack = { side = "defense", allow = "only_listed" },
  },
]

[assets.AetherTechMarines]
id = "HAWK-F7-003"
name = "Aether-Tech Marines"
category = "Force"
min_rating = 7
hp = 16
cost = 30
maintenance = 0
tech_level = 4
speed = 10
type = "Security Personnel"
flags = []
counter = "2d8"
description = "Heavy boarding troops in sealed Aether-tech armour. Moves at speed 10 (short-range), and never needs the local government's permission to land."
attack = { attacker_stat = "Force", defender_stat = "Force", damage = "2d8+2" }

[assets.Dreadnought]
id = "HAWK-F8-001"
name = "Dreadnought"
category = "Force"
min_rating = 8
hp = 30
cost = 40
maintenance = 2
tech_level = 4
speed = 19
drift_rating = 3
type = "Aethership"
flags = ["A", "S"]
purchase_permission = true
counter = "3d8"
description = "A capital Aethership the size of a district, held aloft by a floating Aether-Core and built around a spinal Aether cannon. Moves at speed 19 (long-range). It costs an additional 2 Coin of maintenance each turn. The local government's permission is required to raise one, but not to move one in."
attack = { attacker_stat = "Force", defender_stat = "Force", damage = "3d10+4" }
```

##### Task 2 — `rulebooks/hawk/settings.toml`

The new Drift costs (Decision 21) and H.A.W.K's types (Decision 3).

**Find:**
```toml
# drift_costs[i] = hex-equivalent crossing cost for drift rating (i+1). The Drift
# is dangerous: a crossing always costs more than a step.
drift_costs = [6, 5, 4, 3, 2]

# Every asset type this rulebook's data may name: an asset's type, a transport's
# cargo, a relocate's filter, an effects filter and the floors below. SWN's types
# stay until Efforts 4 to 6 replace the stand-in assets.
asset_types = ["Military Unit", "Special Forces", "Facility", "Starship", "Tactic", "Logistics Facility", "Special", "Aethership"]
```
**Replace with:**
```toml
# drift_costs[i] = movement points a Drift jump costs at drift rating (i+1). The
# Drift is dangerous: a jump always costs more than a step. Front-loaded, so an
# unrated ship struggles and the first rating upgrades matter most.
drift_costs = [10, 7, 5, 3, 2]

# Every asset type this rulebook's data may name: an asset's type, a transport's
# cargo, a relocate's filter, an effects filter and the floors below. H.A.W.K's
# types come first; SWN's stay for the Cunning and Wealth stand-ins until
# Efforts 5 and 6 replace them.
asset_types = [
  "Security Personnel", "Operative", "Facility", "Aethership", "Stratagem", "Beast",
  "Military Unit", "Special Forces", "Starship", "Tactic", "Logistics Facility", "Special",
]
```

##### Task 3 — `rulebooks/hawk/tags.toml`

Aether-Tech Evangelists names the beasts it gates (Decision 6):

**Find:**
```toml
effect = "Every place with this faction's Base of Influence counts as at least tech level 4. Action: pay 2 Coin and roll 1d12; on a 12, the place's tech level becomes 4 for this faction."
```
**Replace with:**
```toml
effect = "Every place with this faction's Base of Influence counts as at least tech level 4. Action: pay 2 Coin and roll 1d12; on a 12, the place's tech level becomes 4 for this faction. Only this faction may buy Arcano-Engineered Warbeasts."
```

Aether Burners runs hot (Decision 24):

**Find:**
```toml
effect = "Once per turn, roll an extra d10 on a Force attack. Every asset can move itself one hex as an action. Every die that comes up 10 must be rerolled."
```
**Replace with:**
```toml
effect = "Once per turn, roll an extra d10 on a Force attack. A move order may run hot: the faction's Aetherships move 5 faster, every other asset can move itself at speed 5, and each asset running hot takes 1d4 damage every turn it moves. Every die that comes up 10 must be rerolled."
```

**Find:**
```toml
  {
    trigger = "speed_query",
    effect = "speed_floor",
    filters = { owner = "own", scope = "faction" },
    speed_floor = { speed = 1 },
  },
```
**Replace with:**
```toml
  {
    trigger = "speed_query",
    effect = "speed_bonus",
    filters = { owner = "own", scope = "faction", asset_type = ["Aethership"], hot = true },
    speed_bonus = { amount = 5 },
  },
  {
    trigger = "speed_query",
    effect = "speed_floor",
    filters = { owner = "own", scope = "faction", hot = true },
    speed_floor = { speed = 5 },
  },
  {
    trigger = "hot_move",
    effect = "self_damage",
    filters = { owner = "own", scope = "faction" },
    self_damage = { amount = "1d4" },
  },
```

##### Task 4 — `campaigns/hawk/state/state.seed.toml`

The seed's 17 Force units become H.A.W.K's (Decision 14). With the Edit tool and `replace_all`, replace every `SWN-F` with `HAWK-F`. The only `SWN-F` strings in the file are the 17 Force blocks' table header, `id` and `definition_id` lines (51 lines), so the swap covers ten definitions:

| Stand-in | Heir | Seeded |
|----------|------|--------|
| `SWN-F1-001` | `HAWK-F1-001` Corporate Security | 1 |
| `SWN-F2-002` | `HAWK-F2-002` Aetheric-Lance Corps | 1 |
| `SWN-F2-003` | `HAWK-F2-003` Shield Wardens | 2 |
| `SWN-F2-004` | `HAWK-F2-004` Unionized Partisans | 1 |
| `SWN-F3-001` | `HAWK-F3-001` Overclocked Zealots | 3 |
| `SWN-F3-003` | `HAWK-F3-003` Wizard Advisors | 3 |
| `SWN-F4-003` | `HAWK-F4-003` Runebound Frigates | 2 |
| `SWN-F5-003` | `HAWK-F5-003` Pact-Bound Assassins | 1 |
| `SWN-F6-002` | `HAWK-F6-002` Arcano-Linked Batteries | 1 |
| `SWN-F7-003` | `HAWK-F7-003` Aether-Tech Marines | 2 |

Then the header:

**Find:**
```toml
# Starting assets are SWN stand-ins until Efforts 4 to 6 swap in HAWK- assets.
```
**Replace with:**
```toml
# Force starting assets are H.A.W.K's own (Effort 4); Cunning and Wealth assets
# are SWN stand-ins until Efforts 5 and 6 swap in HAWK- assets.
```

Check: `grep -c "SWN-F" campaigns/hawk/state/state.seed.toml` prints `0`.

##### Commit message

```
feat(E4): replace the Force stand-in with H.A.W.K's own assets

- force_assets.toml: 28 H.A.W.K Force assets with HAWK-F IDs,
  H.A.W.K types, speed classes 10/14/19 and drift ratings
- settings.toml: drift_costs [10, 7, 5, 3, 2]; Security Personnel,
  Operative, Stratagem and Beast join asset_types
- tags.toml: Aether Burners runs hot (Aetherships +5, floor 5, 1d4
  burn); Aether-Tech Evangelists gates the warbeasts
- state.seed.toml: the 17 seeded Force units swap to their heirs
```

#### Commit 8 — `docs(E4): write the Force canon and play the live turn`

##### Task 1 — `campaigns/hawk/canon/assets.md`

The fiction of the 28 assets, from the signed-off drafts.

**Find:**
```markdown
The fiction behind each asset, one entry per `HAWK-` asset ID.
```
**Replace with:**
```markdown
The fiction behind each asset, one entry per `HAWK-` asset ID.

H.A.W.K's assets are corporate sci-fantasy: corporate culture, magi-tech and high fantasy at once. The rules for each asset are in `rulebooks/hawk/assets/`.

## Corporate Security — `HAWK-F1-001`

Guards posted to a warehouse, a dock or a refinery gate, in company colours, with a ward-baton and a lanyard. Each carries an Aetherstone charm that marks them as on shift to every lock and ward on site. They are the first face anyone meets at a company property.

## Unsanctioned Adventurers — `HAWK-F1-002`

Adventurers who failed H.A.W.K's sanctioning and kept the gear. They work cheap, they go straight for a rival's leadership, and they rarely survive the first room. Companies hire them because they are deniable, and because there is always another party.

## Reluctant Militia — `HAWK-F1-003`

The people of a company town, issued rifles by the company that owns the town, their homes and their debts. They drill on rest days in the refinery yard, and their muster roll is kept by the payroll office. Service is a clause in every lease, and the company enforces its clauses.

## Arcano-Engineered Warbeasts — `HAWK-F1-004`

Owlbears, drakes and stranger stock, bred in Arcano-Tech laboratories and grown to size in Aether vats, delivered with a serial number on the flank and a handler's manual. The patents are held by the Aether-tech houses, and only their licensees may field them. Every beast is leased, not sold, and is collected at the end of its term.

## Unpaid Interns — `HAWK-F1-005`

Apprentice mages and junior engineers on placement, working for experience and a letter of reference. They tend the Aether-Cores, recharge the wards under fire and stand nearest whatever is most expensive. When a ward fails, it is an intern who closes the breach.

## Drift Haulers — `HAWK-F2-001`

Short-hop haulers in company livery: a cargo spine clamped to a single Aetherstone ring, built to ride the near currents of the Drift. Each lifts one unit or one load at a time on a dispatch slip. The cargo is insured; the crew is not.

## Aetheric-Lance Corps — `HAWK-F2-002`

Light cavalry on Aetherbikes, each rider armed with a lance-rifle that fires a charged bolt or holds a blade of light at the tip. They answer alarms across a company's holdings and raid a rival's supply lines between calls. Response times are logged and reviewed each quarter.

## Shield Wardens — `HAWK-F2-003`

Clerks, dock crews and site staff trained in the shield wall, issued tower shields with ward-plates bolted to the face. The plates are recharged from the office Aether-Core at the end of each shift. Training is mandatory and annual.

## Unionized Partisans — `HAWK-F2-004`

The workers of a refinery or a foundry, organised in a guild hall and armed from its stores: mining charges, cutting wands, the old family blades. They do not hold ground. They stop the work, and they know where every fuse in the place is.

## Overclocked Zealots — `HAWK-F3-001`

True believers in the company's vision, fitted with Aether conduits that run raw power straight through their bodies with the limiters removed. Every strike they land burns them from the inside. They volunteered for the procedure and signed the waiver gladly.

## Runic Minefield — `HAWK-F3-002`

Runes laid into a landing field, a dock or a Drift approach, each charged from a buried Aetherstone and keyed to the company's staff badges. They lie dormant until a hostile unit crosses them, then trigger in sequence. The site plan marks where they are; only the security office holds the site plan.

## Wizard Advisors — `HAWK-F3-003`

Wizards trained to keep theoretical magic alive as the world turns to Arcano-Tech, now retained by companies as advisors. They scry the approaches to every company holding and brief the garrison each morning on where the next attack will land. The garrison is always ready for it, and the advisors invoice for the readiness.

## Corporate Chaplains — `HAWK-F3-004`

Soul-bound chaplains on the company payroll, with a field shrine and a claims desk. A fallen soldier can be called back across the line if the request is filed in time and half the cost is approved. The god who answers the request bills the soldier separately.

## Warded Carriers — `HAWK-F4-001`

Long-hulled carriers with a geodesic ward-dome over the drop deck, built to move a whole regiment with its staff and its records. The deployment order names the world; the carrier sets down, the dome opens, and the regiment walks off ready for work. Each head moved is charged to its department.

## Teleportation Hubs — `HAWK-F4-002`

Paired teleportation circles inlaid in the floors of depots across a company's holdings, maintained by a coven under service contract. The circles are built into the bedrock and do not move. A unit steps into one and out of another, with its transit logged at both ends. The circles are recalibrated every turn by an intern.

## Runebound Frigates — `HAWK-F4-003`

Angular warships built around a ringed Aether-Core, their hull plates etched with binding runes that hold the Core's power in check. Lance emitters run down the spine, and energy fins flare when they fire. They patrol a company's lanes and enforce its contracts on anyone who strays into them.

## Arcane Strike Cadre — `HAWK-F4-004`

The standing force of a company's security arm: drilled soldiers with Aether rifles and ward-cloth uniforms in company colours, each squad with a combat mage attached and a field battery behind them. Each cadre carries its own quartermaster and its own budget. They are the line every other unit falls back to.

## Hexcannon Gunships — `HAWK-F5-001`

Fast gunships armed with hexcannons, which fire curses that seize a ship's engines rather than breach its hull. A company's collections department sends them after unpaid debts, breached contracts and rivals' shipping. They stop a freighter, board it, and take what is owed from the hold.

## Bonded Armory — `HAWK-F5-002`

A bonded warehouse of salvage, surplus and quietly acquired arms, with a purchasing office at the front. It can source any weapon, warframe or crew ever built, at a premium and with a lead time of one turn. Every purchase is signed for in triplicate.

## Pact-Bound Assassins — `HAWK-F5-003`

Warlocks whose patron beyond the Material granted them the power to pass unseen. Companies retain them through intermediaries and never list them on the payroll. Their pact and their employment contract both run until the work is done.

## Battle Choir — `HAWK-F5-004`

Battle-sorcerers wired into portable Aether-Cores, since the Weave does not reach the Core. They stand behind the line and chant in unison, and every unit within hearing fights harder for it. Each choir has a conductor, a setlist and a union rep.

## Chartered Mage-Knights — `HAWK-F6-001`

An order of mage-knights chartered to a company: rune-forged plate enamelled in its livery, enchanted blades, Aether pistols, and an oath sworn to the board. Their wards turn a blow back on whoever struck it. Knighthood is granted by tourney and confirmed by performance review.

## Arcano-Linked Batteries — `HAWK-F6-002`

Aether cannon and ward-towers linked into one battery by Arcano-Tech relays, which need a live connection to the Weave. Off the Material Plane they work only on ground where a Weave Anchor holds a strand in place; there, the battery tracks every hull in the approach and fires on any without clearance. It does nothing about threats on the ground.

## Forged Siege Engines — `HAWK-F6-003`

Forged war-frames the size of houses, built for siege work and hired out on fixed-term contracts. They are people, with names, grievances and a guild, and they negotiate their own rates. They break fortified lines.

## Deep-Insertion Gate — `HAWK-F7-001`

A long-range teleportation gate that sends one unit across half the Core in a single step, on a priority clearance that overrides local customs, wards and lockdowns. The clearance is bought from whoever holds the charter for the route. Local governments are notified after arrival.

## Abjurative Firewall — `HAWK-F7-002`

A lattice of abjuration wards and scrying mirrors over a whole facility, run by the security office. Spies, illusions and infiltrators break on it, and every attempt is logged. It is no defence against rifles.

## Aether-Tech Marines — `HAWK-F7-003`

Heavy boarding troops in sealed Aether-tech armour, trained for opposed landings and ship-to-ship assaults. They go where the company sends them, whether or not the local government agrees, and the legal department follows a day behind.

## Dreadnought — `HAWK-F8-001`

A capital Aethership the size of a district, built around a spinal Aether cannon and held aloft by a floating Aether-Core in its central ring. It carries a crew of thousands, a board of directors and a full cohort of unpaid interns, who handle the Core. Its upkeep is a standing line in the annual report.
```

##### Task 2 — `campaigns/hawk/canon/glossary.md`

The Aethership entry gains its look (Decision 5).

**Find:**
```markdown
A vessel powered by Aether, able to cross the Astral Sea for exploration, trade or combat.
```
**Replace with:**
```markdown
A vessel powered by Aether, able to cross the Astral Sea for exploration, trade or combat.

An Aethership is built along a long, angular spine, often in a company's livery, around a glowing Aether-Core held in a ring or in cradling hull arms. Lit ward-domes enclose its open decks, and energy fins flare from the hull. The largest are built around a single spinal gun. Aetherships carry no sails.
```

##### Task 3 — `campaigns/hawk/canon/tags.md`

The Evangelists' entry names the patents it holds (Decision 6):

**Find:**
```markdown
The programs are free, because the locals are the product.
```
**Replace with:**
```markdown
The programs are free, because the locals are the product. The Aether-tech houses also hold the patents on Arcano-Engineered Warbeasts, and only their licensees may field them.
```

Aether Burners' entry says what running hot costs (Decision 24):

**Find:**
```markdown
The crews call it "running hot". The insurers call it something else.
```
**Replace with:**
```markdown
The crews call it "running hot". The insurers call it something else. A fleet or a garrison running hot gets anywhere fast, and its hulls and its people pay for it every turn of the trip.
```

##### Task 4 — the live turn

The overview's Live turn, steps 1, 2 and 4. The map is unchanged, so step 3 is skipped. Claude runs:

```sh
go build -o gm-toolkit ./cmd/gm-toolkit
GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit campaign add-rules --rules rulebooks/hawk --campaign hawk --replace
./cmd/scripts/reset-campaign.sh hawk
```

Then Robert runs `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit faction --campaign hawk`, plays one turn, and reports what he saw. Worth a look:

- the seeded Force units under their H.A.W.K names
- a Runebound Frigate's order crossing a lane, shown in the movement view as "In the Drift between …"
- the Radiant Throne (Aether Burners) offered "Run hot?" in its movement form
- Witherblight's Drift-Runners floor now able to build an Aethership on a sparse place

A fault that needs a code or data change is logged in the Decision Record and fixed in this commit, or raised with Robert if it is bigger than the effort.

##### Task 5 — clear the drafts

Delete `scratch/A.001.5.E4-force-drafts.md` and the measuring programs `scratch/distcheck/`, `scratch/stepcheck/`, `scratch/pacecheck/` and `scratch/regioncheck/`; the plan and the canon hold what they settled. Leave `scratch/aethership_reference/`, Robert's images. All are gitignored, so nothing here touches the commit.

##### Task 6 — this plan and the overview

- This plan's Work Breakdown Overview: `- [ ] **Phase 3 — The Force pass**` becomes `- [x] **Phase 3 — The Force pass**`.
- `A.001.5-hawk-asset-authoring-overview-plan.md`, Effort Summary:

**Find:**
```markdown
| [ ] | 4 | Force pass |
```
**Replace with:**
```markdown
| [x] | 4 | Force pass |
```

##### Commit message

```
docs(E4): write the Force canon and play the live turn

- canon/assets.md: the fiction of the 28 H.A.W.K Force assets
- canon/glossary.md: what an Aethership looks like
- canon/tags.md: the Evangelists' warbeast patents; Aether Burners'
  running hot
- force_assets.toml: Wizard Advisors and Battle Choir give their
  die once per turn per asset (Decision Record 4)
- campaigns/hawk/rulebook/: rulebooks/hawk composed in again
- the E4 live H.A.W.K turn, played by Robert
- planned-work.md: F.053, reveal stealth without a faction pick
- E4 plan: Phase 3 ticked; overview plan: Effort 4 ticked
```
