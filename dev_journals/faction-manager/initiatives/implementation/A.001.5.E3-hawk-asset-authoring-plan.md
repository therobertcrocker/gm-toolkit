# H.A.W.K Asset Authoring — Effort 3 Plan: Campaign Setup

## In Plain Words

- **What** — The H.A.W.K campaign gets a playable board. The Core Sector map is traced hex for hex from the grid image in the canon, with its places, tech levels, populations and 17 Drift lanes. The rulebook gets H.A.W.K's own settings, and a tracked seed state opens the faction game with eight factions, their tags, bases and stand-in assets.
- **Why** — The tags from Effort 2 and the asset passes to come need somewhere to run. Until the campaign has a map and factions, no H.A.W.K turn can be played, and a bad rule shows up only after every pass is done.
- **How it fits** — Effort 3 of six in `A.001.5`, the last initiative of arc `A.001`. It ends with the first live H.A.W.K turn. Efforts 4 to 6 then swap each category's stand-in assets for H.A.W.K's own and play a turn on this same board.

## Context / Goal

Turn the canon into the campaign's data: the map, the settings and the seed state, each written out in full here, then the first live turn.

- Discovery doc: [`A.001.5-hawk-asset-authoring-discovery.md`](../discovery/A.001.5-hawk-asset-authoring-discovery.md), D-1, D-2, D-10
- Overview plan: [`A.001.5-hawk-asset-authoring-overview-plan.md`](A.001.5-hawk-asset-authoring-overview-plan.md), Effort 3, Live turn, Decisions 1 and 4, Decision Record 10, 11, 14 and 15
- Effort 2 plan: [`A.001.5.E2-hawk-asset-authoring-plan.md`](A.001.5.E2-hawk-asset-authoring-plan.md), the tags and the stand-ins
- Canon: `campaigns/hawk/canon/places.md`, `factions.md`, `tags.md`, `images/core-sector-grid.webp`

## Decisions Ratified in Planning

1. **The map is a 1:1 trace of the grid image** — Robert's call, after an overlay showed the first draft (22 × 15, regions placed by label) sitting half on cloud. The image's hexes are 110 px wide in the 6,500 px original, rows 95 px apart; the sector spans about 40 hexes by 28 rows, not the canon's "about 21 by 15". Every open hex in the image is an engine hex: 277 hexes in 19 regions. Big clearings take turns to cross, so the Drift lanes are the fast road, as the canon's "fast links" says. Rejected: a half-scale trace; the schematic draft.
2. **Follow the grid: regions that share a clearing touch** — Robert's call. Helios runs into Concordia, Forgetfulness Nebula and Harmony share one teal clearing, and Oblivion, Infernus and The Maelstrom form one southern clearing. Everything else is joined only by lanes. Rejected: every region an island.
3. **A region is one connected clearing; stray patches are bridged or dropped** — The engine cannot walk between two pieces of one region, and the builder never draws a lane inside a region. Robert's calls: Lumos's teal clearing, Helios's patch on the top edge and the Cascade's glowing falls each join their region through the two thinnest cloud hexes. The five crimson hexes of the lower Maelstrom, cut off by cloud, are left off. Rejected: a new named region per patch.
4. **Region boundaries inside a shared clearing are Claude's reading of the labels** — Forgetfulness takes the west of the teal clearing, Tranquil Gate included, and Harmony the east. Oblivion ends just east of Last Light, where the hexes turn from orange to red. The crimson patch under the clouds between Infernus and Arcadia is The Maelstrom's.
5. **17 Drift lanes, drawn by line of sight** — The builder joins every pair of `*` marks in different regions that can see each other, so the layout's marks are the lane list. The marks were placed by a search that copies the builder's line-of-sight rule, then checked with the real builder: 18 warps, no warnings, every region reachable. Concordia–Prime Material has two parallel currents, which suits the Core's busiest crossroads. Concordia–Boundless Reach is left out: its every sightline crosses the thickest Drift, which the canon puts across the middle. Harmony–Enigma replaces Forgetfulness–Pax Proxima, whose marks kept seeing Nowhere. The Gulf of Null is its own clearing, so it gets the Nowhere lane.
6. **Tech levels: Low 2, Moderate 3, High 4, Highest 5** — Robert's call. Aether-Tech Evangelists (floor 4) and Weave Anchors (seat 5) matter on most places.
7. **Last Light is High, so the Thymian Armada builds its own ships** — Robert's call. Every SWN Starship is TL 4, and an exile fleet that cannot build ships at home is the wrong joke. `places.md` changes from Moderate to High.
8. **Populations: empty 0, sparse 5,000, settled 100,000, crowded 1,000,000, teeming 10,000,000; ships need 500,000** — Robert's call. Aethership and Starship share SWN's floor, so ships are built at Veil, Haven, Eventide, Last Light and Odium. Drift-Runners' 1% (5,000) opens every sparse place.
9. **Drift costs `[6, 5, 4, 3, 2]`** — Robert's call: the Drift is dangerous, and a crossing always costs more than a step. Rejected: SWN's `[5, 4, 3, 2, 1]`; a cheaper `[3, 3, 2, 2, 1]`.
10. **The four major powers are hegemons, the four smaller factions are majors** — Robert's call. Hegemon 8/7/5 with 4 + 4 starting assets and 49 HP; major 6/5/3 with 2 + 2 and 29 HP. The canon's best stat is primary, its worst tertiary, the third secondary. Rejected: major and minor.
11. **Starting Coin is one turn's income, `Wealth/2 + (Force + Cunning)/4`** — Robert's call. O.W.L. 7, the Throne and the Fey 6, the Court and Seraph 5, the Armada 4, Blackwater and Witherblight 3. Rejected: a flat amount by scale; 0.
12. **Proprietor sits on top of the one-or-two tag limit** — Robert's call; the tag canon already says it comes and goes with the places a faction governs. Every faction but Seraph holds it.
13. **Tags per faction** — Robert's picks. O.W.L.: Aether-Tech Evangelists, Monopolists. The Court: Soul-Bonded, Shell Company. The Throne: Soul-Bonded, Aether Burners. The Fey: Old Faith, Informant Network. Blackwater: Insurgents. The Armada: Godslayers. Witherblight: Drift-Runners. Seraph: War Profiteers.
14. **Bases on governed places other than the seat, case by case** — Robert's calls: O.W.L. at Arcadia 4 HP, the Throne at Lumos 49, the Court at Odium 49. IDs follow `Expand Influence`'s `<faction>-base-<world>-<n>` (`internal/faction/engine/action/actions/expand_influence.go`).
15. **Starting assets are SWN stand-ins chosen by the rules checklist** — `docs/rules/swn-faction-mechanics.md`, Faction Creation Checklist: the scale's count from the primary category and the rest from the other two. Each pick fits the faction's ratings and its seat's tech level, and sits at the seat. Claude picked for the fiction; Robert reviewed. No faction starts with a goal; each picks one at its first turn. The picks are a test board for this arc: the seed stays as written through Efforts 4 to 6, and Robert regenerates the factions for real play after the arc closes.
16. **One phase, two commits** — The map and the settings go together, because the ship floor reads the populations and the Drift costs price the lanes. The seed, the faction canon and the live turn close the effort.
17. **The data was checked before the plan was written** — In a scratch campaign outside the repo, the layout and `data.toml` below built to 19 regions, 16 worlds and 18 warps, and the seed loaded in the TUI against a copy of the rulebook with the settings below. Execution re-runs both on the real campaign.

## Decision Record — Execution

1. **Commit 2: the live turn's three faults are filed, not fixed here** — Robert's call: none comes from the map, the settings or the seed, so all three sit outside this effort. (a) A goal's description is cut off at the panel's side: the existing `B.048`, now noted as seen again. (b) On Radiant Throne's turn the movement form's top scrolled off, so the asset each page is for could not be seen: `B.050`, likely the same missing `.Height(` as `B.035`. (c) On that form "Hold position" could not be picked for the second asset: `B.051`, needs triage. Filed in `250e9bd`.

## Shared Context

**Gate.** Each commit lands green on `go build ./...`, `go vet ./...` and `go test ./...` from the repo root. `TestNewWithRulebook_ShippedRulebook` composes `rulebooks/hawk`, so it reads the new settings.

**Campaign commands** take the prefix `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns` and run after `go build -o gm-toolkit ./cmd/gm-toolkit` (overview Decision 5).

**Image coordinates.** A layout cell at row `r`, column `c` is the image hex at row `r`, column `c + 16`, on a grid whose hex centres sit at `x = 55 + 110·col (+ 55 on odd rows)`, `y = 220 + 95·row` in the 6,500 × 3,400 original. A later edit to the map can lay the grid over the image with these numbers.

**Changing a lane.** Any new or moved `*` can change several lanes at once, because every pair of marks in sight becomes a lane. After an edit, compare `spatial build`'s warp count and `regions.toml`'s `[[warp]]` pairs with the lane list in `places.md`.

## Out of Scope

- **H.A.W.K asset types** — the stand-in assets keep SWN's types, so `asset_types` keeps them too; Efforts 4 to 6 replace both.
- **The lower Maelstrom and other image hexes left off** — Decision 3.
- **Goals in the seed** — each faction picks at its first turn.
- **A map per realm** — `F.046` and `F.045`.
- **Scenario files and a test guide** — overview Out of Scope.

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — The board**
  - Commit 1 — `feat(E3): draw the Core Sector map and set H.A.W.K's settings` (6 tasks)
  - Commit 2 — `feat(E3): seed the H.A.W.K factions` (5 tasks)

### Phase 1 — The board

**In plain words:** The H.A.W.K campaign gets its map, its own rule settings and its eight opening factions, and Robert plays the first H.A.W.K turn on it.

#### Commit 1 — `feat(E3): draw the Core Sector map and set H.A.W.K's settings`

##### Task 1 — `campaigns/hawk/spatial/source/layout.txt`

`campaigns/hawk/spatial/source/layout.txt` (new) — full contents. Odd rows start with one space; keep it.

```text
# The Core Sector of the Astral Sea, traced hex for hex from
# canon/images/core-sector-grid.webp (image column = layout column + 16).
# Regions:
#   A=Radiance's Edge, B=Helios, C=The Bastion, D=Forgetfulness Nebula
#   E=Harmony, F=Concordia, G=The Enclave, H=The Prime Material, I=Discordia
#   J=Pax Proxima, K=Enigma, L=Nowhere, M=The Gulf of Null
#   N=The Gulf of Recursion, O=The Boundless Reach, P=The Meridian Cascade
#   Q=Oblivion, R=Infernus, S=The Maelstrom
# Places: 1=Lumos, 2=Solar Gate, 3=Canon IV, 4=Axial Gate, 5=Tranquil Gate, 6=Eventide, 7=Ouroboros, 8=Veil,
#   9=Haven, a=Somewhere, b=Soma, c=Lethe, d=Last Light, e=Odium, f=Abaddon's Gate, g=Arcadia
# A * marks a Drift lane end; every pair of marks in sight of each other is a lane.
. . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . .
 . . . . . . . . . . . A A A A 2 A A A* . . B B B . . . . . . . . . . . . . . . . .
. . . . . . . . . . . A A A A . . . A . . . B B . . C* C C 4 . . . . . . . . . . .
 . . . . . . . . . . . . A A . . . . . . . B . . . . C . . C C . . . . . . . . . .
. . . . . . . . . . . . A . . . . B B B* B B . . . . . . . . 3 C . . . . . . . . .
 . . . . . . . A A A 1 A . . . . . B B B B . . . . . . . . . C C C* . . . . . . . .
. . . . . . . . A A A* A . . . . . . . . B B* . . . . . . . . . C C . . . . . . . .
 . . . . . . . . . . . . . . . . . . . F F F F . . . . . . . . . C . . . . . . . .
. . . . . . . . D D E E* E . . . . . . F* F F F 7 F . . . . . . . . . . . . . . . .
 . . . . 5 D D D D E E . . . . . G G* . . F 6 F F F . . . . . . . . N N* N . . . . .
. . . . . . . . . . . . . . . G G G G . . F F . . . . . . . . . . . N* N . . . . .
 . . . . . . . . . . . . . . . . G . . . . . . . . . . . . O O O O* . . . . . . . .
. . . . . . . . . . . . . . . . . . . . . H* H H . . . . . . O O O . . . . . . P .
 . J J . . . . . . . . K . . . . . . . . . . 8 H . . . . . . O O O . . . . . P P .
. . . J J* . . . . . . K* K K* . . H H H H H H* H H . . . . . . . . . . . b P P P P .
 . . . . . . . . . . . . . . . H H H . H . . . 9 . . . . . . . . . . . P . . . . .
. . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . P* P P . . .
 . . . . L L* L . a L L L L . . . . . . . . . . . . . . . . . . . . . . . P P . . .
. . . . . . L L L L L* L L* . M M M M . . . . I I* I . . . . . . . . . . . . . . . .
 . . . . . . . . . . . . . . M* M M M . . . I I I I I . . . S S* S g S S . . . . . .
. . . . . . . . . Q* Q . . . . M M M . . . I I* I . . . . . . S S S S S . . . . . .
 . . . . . . . Q Q Q Q Q . . . . . . . . . . . . S S S S . . . S S S . . . . . . .
. . . . . . . . Q Q Q Q c Q . . . . . . R R* R R . S S S S S S S S S . . . . . . .
 . . . . . . . . . . . . . Q d . . R R R e R R R . S S S S . . . . . . . . . . . .
. . . . . . . . . . . . . . Q R R R . . . . . R R R S S S . . . . . . . . . . . .
 . . . . . . . . . . . . . . . R R R . . . . . . . . . . . . . . . . . . . . . . .
. . . . . . . . . . . . . . R R R R R R R R R R . . . . . . . . . . . . . . . . .
 . . . . . . . . . . . . . . . . . . R R R R R R R . . . . . . . . . . . . . . . .
. . . . . . . . . . . . . . . . . . . . R f R R R . . . . . . . . . . . . . . . .
 . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . .
```

##### Task 2 — `campaigns/hawk/spatial/source/data.toml`

`campaigns/hawk/spatial/source/data.toml` (new) — full contents:

```toml
# Region metadata. Every region letter that appears in the layout
# needs an entry; the CLI errors on missing ones.
[regions]
A = { name = "Radiance's Edge" }
B = { name = "Helios" }
C = { name = "The Bastion" }
D = { name = "Forgetfulness Nebula" }
E = { name = "Harmony" }
F = { name = "Concordia" }
G = { name = "The Enclave" }
H = { name = "The Prime Material" }
I = { name = "Discordia" }
J = { name = "Pax Proxima" }
K = { name = "Enigma" }
L = { name = "Nowhere" }
M = { name = "The Gulf of Null" }
N = { name = "The Gulf of Recursion" }
O = { name = "The Boundless Reach" }
P = { name = "The Meridian Cascade" }
Q = { name = "Oblivion" }
R = { name = "Infernus" }
S = { name = "The Maelstrom" }

# Worlds. `at` is the marker glyph (integer for 1-9, string for a-z).
# Tech level from the canon's magic and technology: Low 2, Moderate 3, High 4, Highest 5.
# Population from the canon's word: empty 0, sparse 5,000, settled 100,000,
# crowded 1,000,000, teeming 10,000,000.

[[worlds]]
id = "lumos"
name = "Lumos"
at = 1
region = "A"
tech_level = 4
population = 100000

[[worlds]]
id = "solar-gate"
name = "Solar Gate"
at = 2
region = "A"
tech_level = 4
population = 5000

[[worlds]]
id = "canon-iv"
name = "Canon IV"
at = 3
region = "C"
tech_level = 4
population = 5000

[[worlds]]
id = "axial-gate"
name = "Axial Gate"
at = 4
region = "C"
tech_level = 4
population = 5000

[[worlds]]
id = "tranquil-gate"
name = "Tranquil Gate"
at = 5
region = "D"
tech_level = 3
population = 0

[[worlds]]
id = "eventide"
name = "Eventide"
at = 6
region = "F"
tech_level = 4
population = 1000000

[[worlds]]
id = "ouroboros"
name = "Ouroboros"
at = 7
region = "F"
tech_level = 4
population = 100000

[[worlds]]
id = "veil"
name = "Veil"
at = 8
region = "H"
tech_level = 5
population = 10000000

[[worlds]]
id = "haven"
name = "Haven"
at = 9
region = "H"
tech_level = 5
population = 1000000

[[worlds]]
id = "somewhere"
name = "Somewhere"
at = "a"
region = "L"
tech_level = 2
population = 5000

[[worlds]]
id = "soma"
name = "Soma"
at = "b"
region = "P"
tech_level = 2
population = 100000

[[worlds]]
id = "lethe"
name = "Lethe"
at = "c"
region = "Q"
tech_level = 2
population = 5000

[[worlds]]
id = "last-light"
name = "Last Light"
at = "d"
region = "Q"
tech_level = 4
population = 1000000

[[worlds]]
id = "odium"
name = "Odium"
at = "e"
region = "R"
tech_level = 3
population = 1000000

[[worlds]]
id = "abaddons-gate"
name = "Abaddon's Gate"
at = "f"
region = "R"
tech_level = 4
population = 5000

[[worlds]]
id = "arcadia"
name = "Arcadia"
at = "g"
region = "S"
tech_level = 2
population = 5000
```

##### Task 3 — build the map

The first build, so no `--replace` (overview, Live turn step 3):

```sh
go build -o gm-toolkit ./cmd/gm-toolkit
GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit spatial build --campaign hawk
```

Expected: `19 regions (A, B, … S)`, `16 worlds`, `18 warps`, no `Warning:` lines. The build writes `campaigns/hawk/spatial/regions.toml` and `worlds.toml`; both are committed, as the sandbox's are. The 18 warps are these 17 lanes, Concordia–Prime Material twice:

```
A-B  A-E  B-C  C-N  E-K  F-G  F-H (x2)  H-I  H-K
I-R  J-K  J-L  L-M  L-Q  N-O  O-P  P-S
```

If the count or the pairs differ, a mark was copied wrong: compare Task 1 cell by cell before changing anything.

##### Task 4 — `rulebooks/hawk/settings.toml`

`rulebooks/hawk/settings.toml` (rewrite) — full contents:

```toml
# Setting rules for this rulebook. drift_costs and asset_types are required; every
# other key is optional, and a key left out turns its rule off.

# drift_costs[i] = hex-equivalent crossing cost for drift rating (i+1). The Drift
# is dangerous: a crossing always costs more than a step.
drift_costs = [6, 5, 4, 3, 2]

# Every asset type this rulebook's data may name: an asset's type, a transport's
# cargo, a relocate's filter, an effects filter and the floors below. SWN's types
# stay until Efforts 4 to 6 replace the stand-in assets.
asset_types = ["Military Unit", "Special Forces", "Facility", "Starship", "Tactic", "Logistics Facility", "Special", "Aethership"]

# The least world population on which an asset of each type may be bought. A
# type left out has no floor. Quote a type with a space: "Military Unit".
# Ships need a crowded or teeming place: 1,000,000 people or more.
[purchase_population]
Aethership = 500000
Starship = 500000
```

Then compose it into the campaign, so `campaigns/hawk/rulebook/settings.toml` matches:

```sh
GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit campaign add-rules --rules rulebooks/hawk --campaign hawk --replace
```

##### Task 5 — `campaigns/hawk/canon/places.md`

Four edits: the sector's real size, the layout as traced, the lanes, and Last Light.

**Find:**
```markdown
wider east to west than north to south: about 21 hexes across and 15 rows deep.
```
**Replace with:**
```markdown
wider east to west than north to south: about 40 hexes across and 28 rows deep.
```

**Find:**
```markdown
- **North rim (dark blue)** — a band of open hexes along the top edge. West to east: Radiance's Edge (just south of Solar Gate), Helios, and The Bastion (by Axial Gate).
- **North-west (teal)** — Forgetfulness Nebula, with Harmony to its east. Tranquil Gate is on the rim at their west end.
- **Centre (purple)** — the hub. Concordia (Eventide and Ouroboros) with The Enclave to its west; south of them The Prime Material (Veil and Haven). Discordia is a separate purple clearing further south.
- **West (black)** — Pax Proxima on the western edge, Enigma between it and the centre, and Nowhere and The Gulf of Null to the south-west.
- **East (green)** — down the eastern side, north to south: The Gulf of Recursion, The Boundless Reach, and The Meridian Cascade.
- **South (red-orange)** — Oblivion in the south-west, then Infernus: the widest open stretch on the map, with Abaddon's Gate on the southern rim.
- **South-east (crimson)** — The Maelstrom.
```
**Replace with:**
```markdown
- **North rim (dark blue)** — open patches along the top edge, west to east. Radiance's Edge runs from Solar Gate down through a gap in the clouds to Lumos's teal clearing. Helios runs from the top edge down into Concordia. The Bastion curves from Axial Gate down past Canon IV.
- **North-west (teal)** — Forgetfulness Nebula and Harmony share one clearing: Forgetfulness to the west, with Tranquil Gate at its west end, and Harmony to the east.
- **Centre (purple)** — the hub. Concordia (Eventide and Ouroboros) opens straight onto Helios to its north. The Enclave lies to its west and The Prime Material (Veil and Haven) to its south, each across the Drift. Discordia is a separate purple clearing further south.
- **West (black)** — Pax Proxima on the western edge, Enigma between it and the centre, and Nowhere to the south-west. The Gulf of Null is a separate clearing of purple rock just east of Nowhere.
- **East (green)** — three separate clearings down the eastern side, north to south: The Gulf of Recursion, The Boundless Reach, and The Meridian Cascade, whose glowing falls spill east of Soma.
- **South (red-orange)** — one long clearing: Oblivion in the south-west, with Lethe and Last Light, then Infernus, the widest open stretch on the map, with Abaddon's Gate on the southern rim.
- **South-east (crimson)** — The Maelstrom, around Arcadia. Its crimson clearing runs under the clouds into Infernus.
```

**Find:**
```markdown
The grid draws no routes, and Drift surrounds every clearing. Drift currents work as fast links: one lane of current joins two regions. Which regions the lanes join is not yet fixed. The four gates are the only marked ways in or out of the sector, and each is a place on the map. The largest open areas are the north rim band and the red-orange south. The thickest Drift runs as a band across the middle, between the centre and the east, and along the western and south-western edges.
```
**Replace with:**
```markdown
The grid draws no routes, and Drift surrounds every clearing. Where two regions share a clearing, ships cross on open space. Everywhere else they ride Drift currents, the fast links: one lane of current joins two regions. The four gates are the only marked ways in or out of the sector, and each is a place on the map. The largest open areas are the red-orange south and the hub. The thickest Drift runs as a band across the middle, between the centre and the east, and no current crosses it; it also runs along the western and south-western edges.

**Drift lanes** — 17 currents:

- Radiance's Edge — Helios
- Radiance's Edge — Harmony
- Helios — The Bastion
- The Bastion — The Gulf of Recursion
- The Gulf of Recursion — The Boundless Reach
- The Boundless Reach — The Meridian Cascade
- The Meridian Cascade — The Maelstrom
- Concordia — The Enclave
- Concordia — The Prime Material, two currents side by side: the busiest crossing in the Core
- The Prime Material — Enigma
- The Prime Material — Discordia
- Discordia — Infernus
- Harmony — Enigma
- Enigma — Pax Proxima
- Pax Proxima — Nowhere
- Nowhere — Oblivion
- Nowhere — The Gulf of Null
```

**Find:**
```markdown
**Region** — `oblivion`
**Magic and technology** — Moderate
**Population** — Crowded
```
**Replace with:**
```markdown
**Region** — `oblivion`
**Magic and technology** — High
**Population** — Crowded
```

##### Task 6 — this plan

Log any execution-time change in the Decision Record before the commit.

##### Commit message

```
feat(E3): draw the Core Sector map and set H.A.W.K's settings

- campaigns/hawk/spatial/source: the Core Sector traced hex for hex from
  the canon grid image, 277 hexes in 19 regions, 16 places, 17 Drift lanes
- campaigns/hawk/spatial: regions.toml and worlds.toml built from it
- rulebooks/hawk/settings.toml: drift costs [6, 5, 4, 3, 2], ships need
  500,000 people; composed into campaigns/hawk
- canon/places.md: the sector's real size, the traced layout, the lane
  list, and Last Light raised to High
```

#### Commit 2 — `feat(E3): seed the H.A.W.K factions`

##### Task 1 — `campaigns/hawk/state/state.seed.toml`

`campaigns/hawk/state/state.seed.toml` (new) — full contents. Coordinates are the axial `hex_q` and `hex_r` that Commit 1's build writes to `worlds.toml`; if a world's coordinates there differ, the layout was copied wrong.

```toml
# H.A.W.K's opening position at the start of season 3 (A.001.5 Effort 3).
# Load: ./cmd/scripts/reset-campaign.sh hawk
# Starting assets are SWN stand-ins until Efforts 4 to 6 swap in HAWK- assets.
# Each faction holds its scale's count: hegemon 4 primary + 4 other, major 2 + 2.

campaign_id = "hawk"
cycle_number = 1

[governments]
  haven = "occult-workers-league"
  arcadia = "occult-workers-league"
  abaddons-gate = "court-of-lies"
  odium = "court-of-lies"
  solar-gate = "radiant-throne"
  lumos = "radiant-throne"
  veil = "fey-royals"
  ouroboros = "blackwater-industries"
  last-light = "thymian-armada"
  canon-iv = "witherblight-collective"

[factions.occult-workers-league]
  id = "occult-workers-league"
  name = "Occult Worker's League"
  scale = "hegemon"
  force = 5
  cunning = 7
  wealth = 8
  current_hp = 49
  max_hp = 49
  coin = 7
  xp = 0
  homeworld = { world_id = "haven", region_hex = { RegionID = "H", Coord = { Q = 16, R = 15 } } }
  tags = ["HAWK-T-001", "HAWK-T-011", "HAWK-T-010"]
  bases = [
    { id = "occult-workers-league-homeworld", owner_id = "occult-workers-league", current_hp = 49, max_hp = 49, influence = 0, ready = false, is_homeworld = true, location = { world_id = "haven", region_hex = { RegionID = "H", Coord = { Q = 16, R = 15 } } } },
    { id = "occult-workers-league-base-arcadia-1", owner_id = "occult-workers-league", current_hp = 4, max_hp = 4, influence = 0, ready = false, is_homeworld = false, location = { world_id = "arcadia", region_hex = { RegionID = "S", Coord = { Q = 23, R = 19 } } } },
  ]
  [factions.occult-workers-league.assets.occult-workers-league-SWN-W3-001-1]
    id = "occult-workers-league-SWN-W3-001-1"
    definition_id = "SWN-W3-001"
    owner_id = "occult-workers-league"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "haven", region_hex = { RegionID = "H", Coord = { Q = 16, R = 15 } } }
  [factions.occult-workers-league.assets.occult-workers-league-SWN-W4-001-1]
    id = "occult-workers-league-SWN-W4-001-1"
    definition_id = "SWN-W4-001"
    owner_id = "occult-workers-league"
    current_hp = 10
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "haven", region_hex = { RegionID = "H", Coord = { Q = 16, R = 15 } } }
  [factions.occult-workers-league.assets.occult-workers-league-SWN-W6-002-1]
    id = "occult-workers-league-SWN-W6-002-1"
    definition_id = "SWN-W6-002"
    owner_id = "occult-workers-league"
    current_hp = 15
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "haven", region_hex = { RegionID = "H", Coord = { Q = 16, R = 15 } } }
  [factions.occult-workers-league.assets.occult-workers-league-SWN-W7-001-1]
    id = "occult-workers-league-SWN-W7-001-1"
    definition_id = "SWN-W7-001"
    owner_id = "occult-workers-league"
    current_hp = 16
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "haven", region_hex = { RegionID = "H", Coord = { Q = 16, R = 15 } } }
  [factions.occult-workers-league.assets.occult-workers-league-SWN-C2-001-1]
    id = "occult-workers-league-SWN-C2-001-1"
    definition_id = "SWN-C2-001"
    owner_id = "occult-workers-league"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "haven", region_hex = { RegionID = "H", Coord = { Q = 16, R = 15 } } }
  [factions.occult-workers-league.assets.occult-workers-league-SWN-C3-003-1]
    id = "occult-workers-league-SWN-C3-003-1"
    definition_id = "SWN-C3-003"
    owner_id = "occult-workers-league"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "haven", region_hex = { RegionID = "H", Coord = { Q = 16, R = 15 } } }
  [factions.occult-workers-league.assets.occult-workers-league-SWN-F1-001-1]
    id = "occult-workers-league-SWN-F1-001-1"
    definition_id = "SWN-F1-001"
    owner_id = "occult-workers-league"
    current_hp = 3
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "haven", region_hex = { RegionID = "H", Coord = { Q = 16, R = 15 } } }
  [factions.occult-workers-league.assets.occult-workers-league-SWN-F3-003-1]
    id = "occult-workers-league-SWN-F3-003-1"
    definition_id = "SWN-F3-003"
    owner_id = "occult-workers-league"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "haven", region_hex = { RegionID = "H", Coord = { Q = 16, R = 15 } } }

[factions.court-of-lies]
  id = "court-of-lies"
  name = "The Court of Lies"
  scale = "hegemon"
  force = 7
  cunning = 8
  wealth = 5
  current_hp = 49
  max_hp = 49
  coin = 5
  xp = 0
  homeworld = { world_id = "abaddons-gate", region_hex = { RegionID = "R", Coord = { Q = 7, R = 28 } } }
  tags = ["HAWK-T-001", "HAWK-T-004", "HAWK-T-008"]
  bases = [
    { id = "court-of-lies-homeworld", owner_id = "court-of-lies", current_hp = 49, max_hp = 49, influence = 0, ready = false, is_homeworld = true, location = { world_id = "abaddons-gate", region_hex = { RegionID = "R", Coord = { Q = 7, R = 28 } } } },
    { id = "court-of-lies-base-odium-1", owner_id = "court-of-lies", current_hp = 49, max_hp = 49, influence = 0, ready = false, is_homeworld = false, location = { world_id = "odium", region_hex = { RegionID = "R", Coord = { Q = 9, R = 23 } } } },
  ]
  [factions.court-of-lies.assets.court-of-lies-SWN-C4-004-1]
    id = "court-of-lies-SWN-C4-004-1"
    definition_id = "SWN-C4-004"
    owner_id = "court-of-lies"
    current_hp = 8
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "abaddons-gate", region_hex = { RegionID = "R", Coord = { Q = 7, R = 28 } } }
  [factions.court-of-lies.assets.court-of-lies-SWN-C5-001-1]
    id = "court-of-lies-SWN-C5-001-1"
    definition_id = "SWN-C5-001"
    owner_id = "court-of-lies"
    current_hp = 8
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "abaddons-gate", region_hex = { RegionID = "R", Coord = { Q = 7, R = 28 } } }
  [factions.court-of-lies.assets.court-of-lies-SWN-C6-003-1]
    id = "court-of-lies-SWN-C6-003-1"
    definition_id = "SWN-C6-003"
    owner_id = "court-of-lies"
    current_hp = 10
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "abaddons-gate", region_hex = { RegionID = "R", Coord = { Q = 7, R = 28 } } }
  [factions.court-of-lies.assets.court-of-lies-SWN-C7-003-1]
    id = "court-of-lies-SWN-C7-003-1"
    definition_id = "SWN-C7-003"
    owner_id = "court-of-lies"
    current_hp = 5
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "abaddons-gate", region_hex = { RegionID = "R", Coord = { Q = 7, R = 28 } } }
  [factions.court-of-lies.assets.court-of-lies-SWN-F3-001-1]
    id = "court-of-lies-SWN-F3-001-1"
    definition_id = "SWN-F3-001"
    owner_id = "court-of-lies"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "abaddons-gate", region_hex = { RegionID = "R", Coord = { Q = 7, R = 28 } } }
  [factions.court-of-lies.assets.court-of-lies-SWN-F5-003-1]
    id = "court-of-lies-SWN-F5-003-1"
    definition_id = "SWN-F5-003"
    owner_id = "court-of-lies"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "abaddons-gate", region_hex = { RegionID = "R", Coord = { Q = 7, R = 28 } } }
  [factions.court-of-lies.assets.court-of-lies-SWN-F7-003-1]
    id = "court-of-lies-SWN-F7-003-1"
    definition_id = "SWN-F7-003"
    owner_id = "court-of-lies"
    current_hp = 16
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "abaddons-gate", region_hex = { RegionID = "R", Coord = { Q = 7, R = 28 } } }
  [factions.court-of-lies.assets.court-of-lies-SWN-W2-002-1]
    id = "court-of-lies-SWN-W2-002-1"
    definition_id = "SWN-W2-002"
    owner_id = "court-of-lies"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "abaddons-gate", region_hex = { RegionID = "R", Coord = { Q = 7, R = 28 } } }

[factions.radiant-throne]
  id = "radiant-throne"
  name = "The Radiant Throne"
  scale = "hegemon"
  force = 8
  cunning = 5
  wealth = 7
  current_hp = 49
  max_hp = 49
  coin = 6
  xp = 0
  homeworld = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }
  tags = ["HAWK-T-001", "HAWK-T-004", "HAWK-T-003"]
  bases = [
    { id = "radiant-throne-homeworld", owner_id = "radiant-throne", current_hp = 49, max_hp = 49, influence = 0, ready = false, is_homeworld = true, location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } } },
    { id = "radiant-throne-base-lumos-1", owner_id = "radiant-throne", current_hp = 49, max_hp = 49, influence = 0, ready = false, is_homeworld = false, location = { world_id = "lumos", region_hex = { RegionID = "A", Coord = { Q = 8, R = 5 } } } },
  ]
  [factions.radiant-throne.assets.radiant-throne-SWN-F3-001-1]
    id = "radiant-throne-SWN-F3-001-1"
    definition_id = "SWN-F3-001"
    owner_id = "radiant-throne"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }
  [factions.radiant-throne.assets.radiant-throne-SWN-F4-003-1]
    id = "radiant-throne-SWN-F4-003-1"
    definition_id = "SWN-F4-003"
    owner_id = "radiant-throne"
    current_hp = 8
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }
  [factions.radiant-throne.assets.radiant-throne-SWN-F6-002-1]
    id = "radiant-throne-SWN-F6-002-1"
    definition_id = "SWN-F6-002"
    owner_id = "radiant-throne"
    current_hp = 20
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }
  [factions.radiant-throne.assets.radiant-throne-SWN-F7-003-1]
    id = "radiant-throne-SWN-F7-003-1"
    definition_id = "SWN-F7-003"
    owner_id = "radiant-throne"
    current_hp = 16
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }
  [factions.radiant-throne.assets.radiant-throne-SWN-C1-002-1]
    id = "radiant-throne-SWN-C1-002-1"
    definition_id = "SWN-C1-002"
    owner_id = "radiant-throne"
    current_hp = 3
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }
  [factions.radiant-throne.assets.radiant-throne-SWN-C4-001-1]
    id = "radiant-throne-SWN-C4-001-1"
    definition_id = "SWN-C4-001"
    owner_id = "radiant-throne"
    current_hp = 10
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }
  [factions.radiant-throne.assets.radiant-throne-SWN-W4-002-1]
    id = "radiant-throne-SWN-W4-002-1"
    definition_id = "SWN-W4-002"
    owner_id = "radiant-throne"
    current_hp = 12
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }
  [factions.radiant-throne.assets.radiant-throne-SWN-W4-003-1]
    id = "radiant-throne-SWN-W4-003-1"
    definition_id = "SWN-W4-003"
    owner_id = "radiant-throne"
    current_hp = 8
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }

[factions.fey-royals]
  id = "fey-royals"
  name = "The Fey Royals"
  scale = "hegemon"
  force = 5
  cunning = 8
  wealth = 7
  current_hp = 49
  max_hp = 49
  coin = 6
  xp = 0
  homeworld = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  tags = ["HAWK-T-001", "HAWK-T-015", "HAWK-T-016"]
  bases = [
    { id = "fey-royals-homeworld", owner_id = "fey-royals", current_hp = 49, max_hp = 49, influence = 0, ready = false, is_homeworld = true, location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } } },
  ]
  [factions.fey-royals.assets.fey-royals-SWN-C2-004-1]
    id = "fey-royals-SWN-C2-004-1"
    definition_id = "SWN-C2-004"
    owner_id = "fey-royals"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  [factions.fey-royals.assets.fey-royals-SWN-C5-001-1]
    id = "fey-royals-SWN-C5-001-1"
    definition_id = "SWN-C5-001"
    owner_id = "fey-royals"
    current_hp = 8
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  [factions.fey-royals.assets.fey-royals-SWN-C7-001-1]
    id = "fey-royals-SWN-C7-001-1"
    definition_id = "SWN-C7-001"
    owner_id = "fey-royals"
    current_hp = 16
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  [factions.fey-royals.assets.fey-royals-SWN-C7-002-1]
    id = "fey-royals-SWN-C7-002-1"
    definition_id = "SWN-C7-002"
    owner_id = "fey-royals"
    current_hp = 10
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  [factions.fey-royals.assets.fey-royals-SWN-F2-003-1]
    id = "fey-royals-SWN-F2-003-1"
    definition_id = "SWN-F2-003"
    owner_id = "fey-royals"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  [factions.fey-royals.assets.fey-royals-SWN-F3-003-1]
    id = "fey-royals-SWN-F3-003-1"
    definition_id = "SWN-F3-003"
    owner_id = "fey-royals"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  [factions.fey-royals.assets.fey-royals-SWN-W2-002-1]
    id = "fey-royals-SWN-W2-002-1"
    definition_id = "SWN-W2-002"
    owner_id = "fey-royals"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  [factions.fey-royals.assets.fey-royals-SWN-W7-003-1]
    id = "fey-royals-SWN-W7-003-1"
    definition_id = "SWN-W7-003"
    owner_id = "fey-royals"
    current_hp = 5
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }

[factions.blackwater-industries]
  id = "blackwater-industries"
  name = "Blackwater Industries"
  scale = "major"
  force = 5
  cunning = 6
  wealth = 3
  current_hp = 29
  max_hp = 29
  coin = 3
  xp = 0
  homeworld = { world_id = "ouroboros", region_hex = { RegionID = "F", Coord = { Q = 19, R = 8 } } }
  tags = ["HAWK-T-001", "HAWK-T-014"]
  bases = [
    { id = "blackwater-industries-homeworld", owner_id = "blackwater-industries", current_hp = 29, max_hp = 29, influence = 0, ready = false, is_homeworld = true, location = { world_id = "ouroboros", region_hex = { RegionID = "F", Coord = { Q = 19, R = 8 } } } },
  ]
  [factions.blackwater-industries.assets.blackwater-industries-SWN-C5-003-1]
    id = "blackwater-industries-SWN-C5-003-1"
    definition_id = "SWN-C5-003"
    owner_id = "blackwater-industries"
    current_hp = 6
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "ouroboros", region_hex = { RegionID = "F", Coord = { Q = 19, R = 8 } } }
  [factions.blackwater-industries.assets.blackwater-industries-SWN-C6-002-1]
    id = "blackwater-industries-SWN-C6-002-1"
    definition_id = "SWN-C6-002"
    owner_id = "blackwater-industries"
    current_hp = 15
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "ouroboros", region_hex = { RegionID = "F", Coord = { Q = 19, R = 8 } } }
  [factions.blackwater-industries.assets.blackwater-industries-SWN-F2-003-1]
    id = "blackwater-industries-SWN-F2-003-1"
    definition_id = "SWN-F2-003"
    owner_id = "blackwater-industries"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "ouroboros", region_hex = { RegionID = "F", Coord = { Q = 19, R = 8 } } }
  [factions.blackwater-industries.assets.blackwater-industries-SWN-W2-001-1]
    id = "blackwater-industries-SWN-W2-001-1"
    definition_id = "SWN-W2-001"
    owner_id = "blackwater-industries"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "ouroboros", region_hex = { RegionID = "F", Coord = { Q = 19, R = 8 } } }

[factions.thymian-armada]
  id = "thymian-armada"
  name = "The Thymian Armada"
  scale = "major"
  force = 6
  cunning = 3
  wealth = 5
  current_hp = 29
  max_hp = 29
  coin = 4
  xp = 0
  homeworld = { world_id = "last-light", region_hex = { RegionID = "Q", Coord = { Q = 3, R = 23 } } }
  tags = ["HAWK-T-001", "HAWK-T-013"]
  bases = [
    { id = "thymian-armada-homeworld", owner_id = "thymian-armada", current_hp = 29, max_hp = 29, influence = 0, ready = false, is_homeworld = true, location = { world_id = "last-light", region_hex = { RegionID = "Q", Coord = { Q = 3, R = 23 } } } },
  ]
  [factions.thymian-armada.assets.thymian-armada-SWN-F4-003-1]
    id = "thymian-armada-SWN-F4-003-1"
    definition_id = "SWN-F4-003"
    owner_id = "thymian-armada"
    current_hp = 8
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "last-light", region_hex = { RegionID = "Q", Coord = { Q = 3, R = 23 } } }
  [factions.thymian-armada.assets.thymian-armada-SWN-F3-001-1]
    id = "thymian-armada-SWN-F3-001-1"
    definition_id = "SWN-F3-001"
    owner_id = "thymian-armada"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "last-light", region_hex = { RegionID = "Q", Coord = { Q = 3, R = 23 } } }
  [factions.thymian-armada.assets.thymian-armada-SWN-W1-002-1]
    id = "thymian-armada-SWN-W1-002-1"
    definition_id = "SWN-W1-002"
    owner_id = "thymian-armada"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "last-light", region_hex = { RegionID = "Q", Coord = { Q = 3, R = 23 } } }
  [factions.thymian-armada.assets.thymian-armada-SWN-W2-004-1]
    id = "thymian-armada-SWN-W2-004-1"
    definition_id = "SWN-W2-004"
    owner_id = "thymian-armada"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "last-light", region_hex = { RegionID = "Q", Coord = { Q = 3, R = 23 } } }

[factions.witherblight-collective]
  id = "witherblight-collective"
  name = "Witherblight Collective"
  scale = "major"
  force = 5
  cunning = 6
  wealth = 3
  current_hp = 29
  max_hp = 29
  coin = 3
  xp = 0
  homeworld = { world_id = "canon-iv", region_hex = { RegionID = "C", Coord = { Q = 28, R = 4 } } }
  tags = ["HAWK-T-001", "HAWK-T-012"]
  bases = [
    { id = "witherblight-collective-homeworld", owner_id = "witherblight-collective", current_hp = 29, max_hp = 29, influence = 0, ready = false, is_homeworld = true, location = { world_id = "canon-iv", region_hex = { RegionID = "C", Coord = { Q = 28, R = 4 } } } },
  ]
  [factions.witherblight-collective.assets.witherblight-collective-SWN-C2-002-1]
    id = "witherblight-collective-SWN-C2-002-1"
    definition_id = "SWN-C2-002"
    owner_id = "witherblight-collective"
    current_hp = 6
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "canon-iv", region_hex = { RegionID = "C", Coord = { Q = 28, R = 4 } } }
  [factions.witherblight-collective.assets.witherblight-collective-SWN-C3-001-1]
    id = "witherblight-collective-SWN-C3-001-1"
    definition_id = "SWN-C3-001"
    owner_id = "witherblight-collective"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "canon-iv", region_hex = { RegionID = "C", Coord = { Q = 28, R = 4 } } }
  [factions.witherblight-collective.assets.witherblight-collective-SWN-F2-004-1]
    id = "witherblight-collective-SWN-F2-004-1"
    definition_id = "SWN-F2-004"
    owner_id = "witherblight-collective"
    current_hp = 6
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "canon-iv", region_hex = { RegionID = "C", Coord = { Q = 28, R = 4 } } }
  [factions.witherblight-collective.assets.witherblight-collective-SWN-F3-003-1]
    id = "witherblight-collective-SWN-F3-003-1"
    definition_id = "SWN-F3-003"
    owner_id = "witherblight-collective"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "canon-iv", region_hex = { RegionID = "C", Coord = { Q = 28, R = 4 } } }

[factions.seraph-industries]
  id = "seraph-industries"
  name = "Seraph Industries"
  scale = "major"
  force = 5
  cunning = 3
  wealth = 6
  current_hp = 29
  max_hp = 29
  coin = 5
  xp = 0
  homeworld = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  tags = ["HAWK-T-007"]
  bases = [
    { id = "seraph-industries-homeworld", owner_id = "seraph-industries", current_hp = 29, max_hp = 29, influence = 0, ready = false, is_homeworld = true, location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } } },
  ]
  [factions.seraph-industries.assets.seraph-industries-SWN-W3-001-1]
    id = "seraph-industries-SWN-W3-001-1"
    definition_id = "SWN-W3-001"
    owner_id = "seraph-industries"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  [factions.seraph-industries.assets.seraph-industries-SWN-W6-001-1]
    id = "seraph-industries-SWN-W6-001-1"
    definition_id = "SWN-W6-001"
    owner_id = "seraph-industries"
    current_hp = 10
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  [factions.seraph-industries.assets.seraph-industries-SWN-F2-002-1]
    id = "seraph-industries-SWN-F2-002-1"
    definition_id = "SWN-F2-002"
    owner_id = "seraph-industries"
    current_hp = 5
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  [factions.seraph-industries.assets.seraph-industries-SWN-C2-001-1]
    id = "seraph-industries-SWN-C2-001-1"
    definition_id = "SWN-C2-001"
    owner_id = "seraph-industries"
    current_hp = 4
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
```

The stand-ins by faction, for review at a glance:

| Faction | Primary | Other |
|---|---|---|
| O.W.L. | Postech Industry, Shipping Combine, R&D Department, Pretech Manufactory | Lobbyists, Covert Shipping, Security Personnel, Counterintel Unit |
| The Court | Seditionists, Organization Moles, Demagogue, Treachery | Zealots, Psychic Assassins, Space Marines, Lawyers |
| The Throne | Zealots, Strike Fleet, Planetary Defenses, Space Marines | Informers, Party Machine, Monopoly, Medical Center |
| The Fey | Seductress, Organization Moles, Popular Movement, Book of Secrets | Hardened Personnel, Counterintel Unit, Lawyers, Transit Web |
| Blackwater | Boltholes, Covert Transit Net | Hardened Personnel, Freighter Contract |
| The Armada | Strike Fleet, Zealots | Harvesters, Surveyors |
| Witherblight | Saboteurs, Cyberninjas | Guerrilla Populace, Counterintel Unit |
| Seraph | Postech Industry, Venture Capital | Elite Skirmishers, Lobbyists |

##### Task 2 — `campaigns/hawk/canon/factions.md`

Each entry gains a **Tags** line after **Controls**, so the canon names each faction's tags; the tag fiction stays in `tags.md`. Eight edits.

**Find:** ``**Controls** — `haven`, `arcadia` ``
**Replace with:**
```markdown
**Controls** — `haven`, `arcadia`
**Tags** — Proprietor (`HAWK-T-001`), Aether-Tech Evangelists (`HAWK-T-011`), Monopolists (`HAWK-T-010`)
```

**Find:** ``**Controls** — `abaddons-gate`, `odium` ``
**Replace with:**
```markdown
**Controls** — `abaddons-gate`, `odium`
**Tags** — Proprietor (`HAWK-T-001`), Soul-Bonded (`HAWK-T-004`), Shell Company (`HAWK-T-008`)
```

**Find:** ``**Controls** — `solar-gate`, `lumos` ``
**Replace with:**
```markdown
**Controls** — `solar-gate`, `lumos`
**Tags** — Proprietor (`HAWK-T-001`), Soul-Bonded (`HAWK-T-004`), Aether Burners (`HAWK-T-003`)
```

**Find:** ``**Controls** — `veil` ``
**Replace with:**
```markdown
**Controls** — `veil`
**Tags** — Proprietor (`HAWK-T-001`), Old Faith (`HAWK-T-015`), Informant Network (`HAWK-T-016`)
```

**Find:** ``**Controls** — `ouroboros` ``
**Replace with:**
```markdown
**Controls** — `ouroboros`
**Tags** — Proprietor (`HAWK-T-001`), Insurgents (`HAWK-T-014`)
```

**Find:** ``**Controls** — `last-light` ``
**Replace with:**
```markdown
**Controls** — `last-light`
**Tags** — Proprietor (`HAWK-T-001`), Godslayers (`HAWK-T-013`)
```

**Find:** ``**Controls** — `canon-iv` ``
**Replace with:**
```markdown
**Controls** — `canon-iv`
**Tags** — Proprietor (`HAWK-T-001`), Drift-Runners (`HAWK-T-012`)
```

**Find:** `**Controls** — none`
**Replace with:**
```markdown
**Controls** — none
**Tags** — War Profiteers (`HAWK-T-007`)
```

##### Task 3 — clear the plan session's scratch images

The plan session left three overlay images in the gitignored `scratch/`: `A.001.5.E3-draft-on-image.png`, `A.001.5.E3-trace-on-image.png` and `A.001.5.E3-map-on-image.png`. Delete them; the plan holds what they showed.

##### Task 4 — the live turn

The overview's Live turn, steps 1, 2 and 4; the map has not changed since Commit 1, so step 3 is skipped. Claude runs:

```sh
go build -o gm-toolkit ./cmd/gm-toolkit
GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit campaign add-rules --rules rulebooks/hawk --campaign hawk --replace
./cmd/scripts/reset-campaign.sh hawk
```

Then Robert runs `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit faction --campaign hawk`, plays one turn and reports what he saw. Each faction picks its goal at its first turn. A fault that needs a code or data change is logged in the Decision Record and fixed in this commit, or raised with Robert if it is bigger than the effort.

##### Task 5 — this plan and the overview

- This plan: tick Phase 1 in the Overview.
- `A.001.5-hawk-asset-authoring-overview-plan.md`, Effort Summary: tick Effort 3's Done box.

**Find:**
```markdown
| [ ] | 3 | Campaign setup |
```
**Replace with:**
```markdown
| [x] | 3 | Campaign setup |
```

##### Commit message

```
feat(E3): seed the H.A.W.K factions

- campaigns/hawk/state/state.seed.toml: eight factions at the start of
  season 3, with scale, stats, Coin, tags, governments, bases and SWN
  stand-in assets
- canon/factions.md: each faction names its tags
- the first live H.A.W.K turn, played by Robert
```
