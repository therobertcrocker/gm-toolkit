# H.A.W.K Asset Authoring — Effort 2 Plan: Tag Pass

## In Plain Words

- **What** — H.A.W.K's own faction tags: 16 tags that any faction in the setting could hold, each one a nature or an aptitude. The rulebook folder `rulebooks/hawk/` starts here, with SWN copies standing in for the asset files until the asset passes replace them. Six engine changes come first, because some tags do things no tag could do before: aim at a faction holding another tag, pay upkeep each turn, cross the Drift faster.
- **Why** — Tags are the first H.A.W.K rules data the engine runs. They come before the campaign setup, because factions take their tags when E3 writes the seed state.
- **How it fits** — Effort 2 of six in `A.001.5`, the last initiative of arc `A.001`. The engine work also takes SWN's asset type names out of Go, so every rulebook names its own. A Go test that loads both `rulebooks/swn` and `rulebooks/hawk` and builds the engine from each is E2's load check. The first live turn is E3's.

## Context / Goal

Design the tag pass (lore-gate, reflavor, sort; discovery D-9) and specify its build rows and data to the task.

- Discovery doc: [`A.001.5-hawk-asset-authoring-discovery.md`](../discovery/A.001.5-hawk-asset-authoring-discovery.md)
- Overview plan: [`A.001.5-hawk-asset-authoring-overview-plan.md`](A.001.5-hawk-asset-authoring-overview-plan.md), Effort 2 and Pass shape
- Canon: `campaigns/hawk/canon/`, written in Effort 1

## Decisions Ratified in Planning

1. **The tag set: 16 tags, each a nature or an aptitude that any faction could hold** — Robert's rules for the pass, from `docs/rules/swn-faction-mechanics.md:77`: a tag represents a faction's nature and special aptitudes, and a faction holds one, two at most. No tag restates one faction's identity (the gods, the fey and Blackwater are factions, not tags). Each tag joins existing effects in a way no SWN tag does, from H.A.W.K's magic, industry, travel and ideologies. Weights are mixed: light tags are an edge, heavy tags carry a cost. Final picks per faction are E3's. Rejected along the way: reflavoring SWN's 20 one for one (copies belong in `_core`), identity tags (Fey Court, Astral Divinity), tags that were only an action (Quarterly Earnings), Charter Colony (one colony in the canon), and Weave-Bound (the Weave does not reach the Core; it became Weave Anchors).
2. **Claude drafted the names and fiction from the canon; Robert edited them** — Five passes in gitignored `scratch/A.001.5.E2-tag-drafts.md` and `-pass5.md`. The fiction reaches the canon in Commit 8 and the scratch files are deleted then.
3. **Asset types are declared in `settings.toml`, `asset_types`, required in every rulebook** — Go named SWN's seven types in a closed switch (`toAssetType`) and in `domain.Type…` constants, so H.A.W.K could not name an Aethership. No Go branches on a type, so the names become data. The loader checks every name the data uses (an asset's `type`, a transport's cargo types, a relocate's types, an effects `asset_type` filter, a `purchase_population` key) against the list. Robert's call to land it in E2, since three tags filter on a type. Rejected: `rulebook.toml` (the compose never copies a manifest, `internal/campaigns/scaffold.go:89`); landing it in E4.
4. **The faction wizard names the government tag from the rulebook** — `wizard.go` titled its select "Governed world (Planetary Government)", SWN's name. Robert's call to fix it in E2.
5. **E2's load check is the engine construction test, run over both rulebooks** — `TestNewWithRulebook_ShippedRulebook` composes a rulebook, loads it and builds the engine, which runs every registrar check. It becomes a table over `swn` and `hawk`. Robert's call. Rejected: a separate H.A.W.K test; a manual compose only.
6. **`opponent_tag` filter on `bonus_die`** — Build. Matches when the faction on the other side of the roll holds the named tag: the opponent of an `own_roll`, the roller of an `enemy_roll_at_me`. Checked at load like `goal` and `asset_definition`. Godslayers and Old Faith aim at Soul-Bonded; Insurgents aim at Monopolists.
7. **`coin_gain` accepts `once_per` on `asset_destroyed` and scope `faction` on `turn_start`** — Parameterize. War Profiteers caps its income at once a turn; Aether Refiners earns at turn start. `FireTurnStart` already fires tag entries (`internal/faction/engine/effect/effect.go`), so a tag's turn-start income needs only the scope.
8. **`reroll` on `rival_roll` accepts scope `base`, read in the reroll hook** — Parameterize. Contract-Bound forces a reroll only on a world where it has a Base of Influence, true to the glossary's "contract magic works only near a Fey Royal's influence". `rollMatches` never sees a `rival_roll` owner, so the hook, which holds the faction state, reads the owner's Bases. The fence-sign in `rollMatches` is updated to say so.
9. **A drift rating bonus, through a new Cat 4 drift-rating hook** — Build. Movement prices a crossing from the asset's drift rating (`internal/faction/engine/world/movement.go`). A missing rating counts as 1, Robert's call, so Drift-Runners helps every asset. `Rulebook.DriftCost` now counts a rating above the table as the table's best cost; it counted it as the worst.
10. **Tag upkeep: a `coin_upkeep` effect on an `upkeep` hook trigger, charged in bookkeeping** — Build. Robert's calls: upkeep is paid after income and before asset maintenance, so tags are paid first; each entry names its consequence when unpaid, `on_unpaid` = `lapse`, `lose_tag` or `damage` (with a `damage` amount), one consequence per miss and no two-strike counter; damage uses the normal faction HP path, so it can destroy a faction. Only a tag holds upkeep; an asset pays maintenance. Rejected: firing upkeep at turn start, before income.
11. **A lapsed tag keeps its upkeep and loses everything else** — `Faction.LapsedTags` holds the lapsed tag IDs, set and cleared by a `TagLapsed` mutation. A lapsed tag's hooks stop firing (the holder live check), its phase and matchup entries are skipped, and its action is not offered. Its upkeep is still charged, and the payment that succeeds lifts the lapse. Losing a tag also clears its lapse.
12. **Soul-Bonded rolls one extra die a turn, on an attack or a defense, and loses 2 HP per missed upkeep** — Robert's calls. Two `bonus_die` entries on one tag share one once-per-turn budget (`budgetKey` has no trigger; the Book of Secrets fence-sign), so one entry with both phases says exactly what the engine does.
13. **Three phases, two of them build phases** — Amends the overview's Decision 6, which gave a pass one build phase. Phase 1 teaches the loader and existing effects new words; Phase 2 adds new triggers, effects and save state; Phase 3 is the data. Phase 2 runs about 1,600 lines, past the 1,500 budget: its two commits extend the same registry and loader blocks in turn, and a split would leave two one-commit phases.
14. **Stand-ins** — `cp` makes byte copies of SWN's three asset files; one Edit then removes Gengineered Slaves' `requires_tag` (discovery D-11). The stand-in `settings.toml` is SWN's with `"Aethership"` added to `asset_types`; its purchase floors stay SWN's until E3. No H.A.W.K tag needs an `asset_definition` filter, so D-11's other half does not arise in E2.
15. **Numbers a tag reuses from SWN stay SWN's** — Aether-Tech Evangelists' action (2 Coin, 1d12, a 12 raises the world to tech level 4), the 1 Coin toll and discount, the 1d6-on-5+ XP roll. Tuning waits for play.
16. **Informant Network fills the Cunning gap** — Four of the eight factions in play are best at Cunning, and no tag rewarded guile. Informant Network rolls one extra die a turn on a Cunning attack or a defense against one: two `bonus_die` entries that share one once-per-turn budget, as Soul-Bonded's do. Its 1 Coin upkeep lapses the tag when unpaid. All reuse. Robert's call. Rejected: a Cunning die added to Shell Company or Contract-Bound; leaving the gap.

## Decision Record — Execution

1. **Commit 5: a movement order still records the definition's drift rating, not the resolved one** — `BuildMovementMutations` prices the crossing from `AssetDriftRatingResolver`, but `MovementOrder.DriftRating` keeps `def.DriftRating`, as the plan wrote it. Nothing outside tests reads the field today, so the two never disagree in play. Deferred: if a later view shows an order's rating, it should store the resolved one.
2. **Commit 6: the cycle digest notes a `tag_lapsed` mutation as unknown** — `digest/build.go` has no case for it, so its `default` adds `unknown mutation: tag_lapsed` to the beat's `Notes`. Nothing outside tests reads `Notes`, so no player sees it; Out of Scope already defers a digest line for a lapse. Deferred with that line: the digest case, a drop or a real line, lands when play asks for one.

## Sort Table

Every row's surface is a tag. Primitives are the registrar names in `internal/faction/engine/effect/effects/registrars.go`; tag actions are ability effects.

| HAWK ID | Name | SWN ancestor | Call | Primitive | Notes |
|---------|------|--------------|------|-----------|-------|
| HAWK-T-001 | Proprietor | Planetary Government | reuse | `government = true` | the government tag |
| HAWK-T-002 | Weave Anchors | Deep Rooted | build | `bonus_die`, `tech_level_floor`, `coin_upkeep`, `remove_tag` | heavy; upkeep `lose_tag` |
| HAWK-T-003 | Aether Burners | Warlike, Mercenary Group | reuse | `bonus_die`, `speed_floor`, `reroll` | heavy; rerolls its own 10s |
| HAWK-T-004 | Soul-Bonded | — | build | `bonus_die`, `coin_upkeep` | heavy; upkeep `damage` 2 (Decision 12) |
| HAWK-T-005 | Contract-Bound | Psychic Academy | parameterize | `reroll` at scope `base`; action `purchase_discount` | Decision 8 |
| HAWK-T-006 | Aether Refiners | — | parameterize | `coin_gain` at turn start, `reroll`, `remove_tag` | heavy; rerolls its own 10s |
| HAWK-T-007 | War Profiteers | Scavengers | parameterize | `coin_gain` once per turn, `bonus_die`, `lose_ties` | heavy |
| HAWK-T-008 | Shell Company | Secretive, Scavengers | reuse | `apply_stealth`, `coin_gain` | |
| HAWK-T-009 | Outsourced | Exchange Consulate | build | `reroll`, `xp_roll`, `coin_upkeep` | heavy; upkeep `lapse` |
| HAWK-T-010 | Monopolists | Pirates, Preceptor Archive | reuse | `arrival_toll`, `cost_discount` | Insurgents' target |
| HAWK-T-011 | Aether-Tech Evangelists | Technical Expertise, Preceptor Archive | reuse | `tech_level_floor`; action `raise_tech_level` | |
| HAWK-T-012 | Drift-Runners | Technical Expertise | build | `drift_rating_bonus`, `population_floor` | Decision 9 |
| HAWK-T-013 | Godslayers | Fanatical | build | `reroll`, `lose_ties`, `bonus_die` + `opponent_tag` | heavy |
| HAWK-T-014 | Insurgents | Perimeter Agency | build | `bonus_die` + `opponent_tag`, `bonus_die` | |
| HAWK-T-015 | Old Faith | Theocratic | build | `bonus_die`, `bonus_die` + `opponent_tag` | counters Soul-Bonded |
| HAWK-T-016 | Informant Network | Machiavellian, Theocratic | build | `bonus_die` ×2, `coin_upkeep` | heavy; upkeep `lapse` (Decision 16) |

## Shared Context

**Gate.** Every commit lands green on `go build ./...`, `go vet ./...` and `go test ./...` from the repo root.

**Campaign commands** take the prefix `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns` and run after `go build -o gm-toolkit ./cmd/gm-toolkit` (overview Decision 5).

## Out of Scope

- **Which faction takes which tag** — E3, when it writes the seed state.
- **A bookkeeping card line for upkeep and a digest line for a lapse** — the mutations reach the history and the feed folds them; a dedicated line waits for play to ask for it.
- **A two-strike upkeep counter** — Decision 10.
- **Seat-change costs firing** — Weave Anchors and Aether Refiners lose the tag when the seat changes, but nothing changes a seat until `F.040`.
- **Renaming SWN's asset types in the H.A.W.K stand-ins and setting Aethership floors** — E3 for the floors, E4 to E6 for the assets.

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — Rulebook vocabulary**
  - Commit 1 — `feat(E2): declare asset types in settings.toml` (7 tasks)
  - Commit 2 — `feat(E2): add the opponent_tag filter` (6 tasks)
  - Commit 3 — `feat(E2): accept once_per and faction scope on coin_gain, base scope on rival_roll` (4 tasks)
  - Commit 4 — `fix(E2): name the government tag from the rulebook in the wizard` (3 tasks)
- [x] **Phase 2 — New tag mechanics**
  - Commit 5 — `feat(E2): add the drift rating bonus` (8 tasks)
  - Commit 6 — `feat(E2): add tag upkeep with lapse, loss and damage` (10 tasks)
- [x] **Phase 3 — The H.A.W.K rulebook**
  - Commit 7 — `feat(E2): start rulebooks/hawk with SWN stand-ins and H.A.W.K tags` (5 tasks)
  - Commit 8 — `docs(E2): write the tag canon and compose into campaigns/hawk` (4 tasks)

### Phase 1 — Rulebook vocabulary

**In plain words:** The loader learns to read asset type names from each rulebook's settings, and existing effects learn three new settings: aim at a faction by its tag, cap a payout at once a turn, and limit a forced reroll to your own bases. The faction wizard stops printing SWN's tag name.

#### Commit 1 — `feat(E2): declare asset types in settings.toml`

##### Task 1 — `internal/faction/domain/asset.go`

**Find:**
```go
type AssetType string

const (
	TypeMilitaryUnit      AssetType = "Military Unit"
	TypeSpecialForces     AssetType = "Special Forces"
	TypeFacility          AssetType = "Facility"
	TypeStarship          AssetType = "Starship"
	TypeTactic            AssetType = "Tactic"
	TypeLogisticsFacility AssetType = "Logistics Facility"
	TypeSpecial           AssetType = "Special"
)
```
**Replace with:**
```go
// AssetType names an asset type. Each rulebook's settings.toml declares the names
// its data uses; the engine names none (A.001.5 Effort 2 Decision 3).
type AssetType string
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

In `Load`, read the settings into one value and check every asset type the data names.

**Find:**
```go
	driftCosts, purchasePopulation, err := loadSettings(filepath.Join(dataDir, "settings.toml"))
	if err != nil {
		return nil, fmt.Errorf("loading settings: %w", err)
	}

	return &Rulebook{
		Assets:             assets,
		Tags:               tags,
		GovernmentTagID:    governmentTagID,
		Goals:              goals,
		DriftCosts:         driftCosts,
		PurchasePopulation: purchasePopulation,
	}, nil
```
**Replace with:**
```go
	loadedSettings, err := loadSettings(filepath.Join(dataDir, "settings.toml"))
	if err != nil {
		return nil, fmt.Errorf("loading settings: %w", err)
	}
	if err := checkAssetTypes(assets, tags, loadedSettings.assetTypes); err != nil {
		return nil, fmt.Errorf("checking asset types: %w", err)
	}

	return &Rulebook{
		Assets:             assets,
		Tags:               tags,
		GovernmentTagID:    governmentTagID,
		Goals:              goals,
		DriftCosts:         loadedSettings.driftCosts,
		PurchasePopulation: loadedSettings.purchasePopulation,
	}, nil
```

In `convertAsset`:

**Find:**
```go
	assetType, err := toAssetType(r.Type)
	if err != nil {
		return nil, err
	}
```
**Replace with:**
```go
	assetType := domain.AssetType(r.Type)
```

In the relocate conversion:

**Find:**
```go
	if record.AssetType != "" {
		assetType, err := toAssetType(record.AssetType)
		if err != nil {
			return nil, fmt.Errorf("relocate asset_type: %w", err)
		}
		params.AssetType = assetType
	}
```
**Replace with:**
```go
	if record.AssetType != "" {
		params.AssetType = domain.AssetType(record.AssetType)
	}
```

**Find:**
```go
	for i, value := range record.ExcludeTypes {
		assetType, err := toAssetType(value)
		if err != nil {
			return nil, fmt.Errorf("relocate exclude_types[%d]: %w", i, err)
		}
		params.ExcludeTypes = append(params.ExcludeTypes, assetType)
	}
```
**Replace with:**
```go
	for _, value := range record.ExcludeTypes {
		params.ExcludeTypes = append(params.ExcludeTypes, domain.AssetType(value))
	}
```

In `convertTransport`:

**Find:**
```go
	for i, s := range r.CargoTypes {
		t, err := toAssetType(s)
		if err != nil {
			return nil, fmt.Errorf("cargo_types[%d]: %w", i, err)
		}
		cargoTypes[i] = t
	}
```
**Replace with:**
```go
	for i, s := range r.CargoTypes {
		cargoTypes[i] = domain.AssetType(s)
	}
```

The settings section, from the `settingsRecord` comment to the end of `loadSettings`:

**Find:**
```go
// settingsRecord lists every setting rule the engine reads, for every rulebook. A
// required key is a plain field loadSettings checks; an optional key is a pointer
// or a map, nil when the setting leaves it out (Effort 4 Decision 1).
type settingsRecord struct {
	DriftCosts         []int          `toml:"drift_costs"`
	PurchasePopulation map[string]int `toml:"purchase_population"`
}

func loadSettings(path string) ([]int, map[domain.AssetType]int, error) {
	var record settingsRecord
	metadata, err := toml.DecodeFile(path, &record)
	if err != nil {
		return nil, nil, err
	}
	if err := refuseUnknownKeys(path, metadata); err != nil {
		return nil, nil, err
	}
	if len(record.DriftCosts) == 0 {
		return nil, nil, fmt.Errorf("%s: drift_costs is required", path)
	}
	purchasePopulation := make(map[domain.AssetType]int, len(record.PurchasePopulation))
	for name, floor := range record.PurchasePopulation {
		assetType, err := toAssetType(name)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: purchase_population: %w", path, err)
		}
		if floor <= 0 {
			return nil, nil, fmt.Errorf("%s: purchase_population %q must be > 0, got %d", path, name, floor)
		}
		purchasePopulation[assetType] = floor
	}
	return record.DriftCosts, purchasePopulation, nil
}
```
**Replace with:**
```go
// settingsRecord lists every setting rule the engine reads, for every rulebook. A
// required key is a plain field loadSettings checks; an optional key is a pointer
// or a map, nil when the setting leaves it out (Effort 4 Decision 1).
type settingsRecord struct {
	DriftCosts         []int          `toml:"drift_costs"`
	AssetTypes         []string       `toml:"asset_types"`
	PurchasePopulation map[string]int `toml:"purchase_population"`
}

// settings is settings.toml as Load uses it.
type settings struct {
	driftCosts         []int
	assetTypes         []domain.AssetType
	purchasePopulation map[domain.AssetType]int
}

func loadSettings(path string) (settings, error) {
	var record settingsRecord
	metadata, err := toml.DecodeFile(path, &record)
	if err != nil {
		return settings{}, err
	}
	if err := refuseUnknownKeys(path, metadata); err != nil {
		return settings{}, err
	}
	if len(record.DriftCosts) == 0 {
		return settings{}, fmt.Errorf("%s: drift_costs is required", path)
	}
	if len(record.AssetTypes) == 0 {
		return settings{}, fmt.Errorf("%s: asset_types is required", path)
	}
	assetTypes := make([]domain.AssetType, 0, len(record.AssetTypes))
	for _, name := range record.AssetTypes {
		assetType := domain.AssetType(name)
		if name == "" || slices.Contains(assetTypes, assetType) {
			return settings{}, fmt.Errorf("%s: asset_types: %q is empty or listed twice", path, name)
		}
		assetTypes = append(assetTypes, assetType)
	}
	purchasePopulation := make(map[domain.AssetType]int, len(record.PurchasePopulation))
	for name, floor := range record.PurchasePopulation {
		assetType := domain.AssetType(name)
		if !slices.Contains(assetTypes, assetType) {
			return settings{}, fmt.Errorf("%s: purchase_population: asset type %q is not in asset_types", path, name)
		}
		if floor <= 0 {
			return settings{}, fmt.Errorf("%s: purchase_population %q must be > 0, got %d", path, name, floor)
		}
		purchasePopulation[assetType] = floor
	}
	return settings{driftCosts: record.DriftCosts, assetTypes: assetTypes, purchasePopulation: purchasePopulation}, nil
}

// checkAssetTypes refuses an asset type settings.toml does not declare, wherever
// the data names one: an asset's type, its transport's cargo types, a relocate's
// asset type and excluded types, and an effects entry's asset_type filter, on an
// asset or a tag.
func checkAssetTypes(assets map[string]*domain.AssetDefinition, tags map[string]*domain.Tag, declared []domain.AssetType) error {
	check := func(holder string, names ...domain.AssetType) error {
		for _, name := range names {
			if !slices.Contains(declared, name) {
				return fmt.Errorf("%s: asset type %q is not in settings.toml asset_types", holder, name)
			}
		}
		return nil
	}
	checkRelocate := func(holder string, relocate *domain.RelocateParams) error {
		if relocate == nil {
			return nil
		}
		if relocate.AssetType != "" {
			if err := check(holder, relocate.AssetType); err != nil {
				return err
			}
		}
		return check(holder, relocate.ExcludeTypes...)
	}
	checkEffects := func(holder string, entries []domain.EffectDefinition) error {
		for i, entry := range entries {
			entryHolder := fmt.Sprintf("%s effects[%d]", holder, i)
			if err := check(entryHolder, entry.Filters.AssetTypes...); err != nil {
				return err
			}
			if entry.GrantMovement != nil {
				if err := checkRelocate(entryHolder, &entry.GrantMovement.Relocate); err != nil {
					return err
				}
			}
		}
		return nil
	}
	checkAbility := func(holder string, ability *domain.AbilityDefinition) error {
		if ability == nil {
			return nil
		}
		for i, step := range ability.Steps {
			if err := checkRelocate(fmt.Sprintf("%s ability step %d", holder, i), step.Relocate); err != nil {
				return err
			}
		}
		return nil
	}
	for _, id := range slices.Sorted(maps.Keys(assets)) {
		def := assets[id]
		holder := fmt.Sprintf("asset %q", id)
		if err := check(holder, def.Type); err != nil {
			return err
		}
		if def.Transport != nil {
			if err := check(holder+" transport", def.Transport.CargoTypes...); err != nil {
				return err
			}
		}
		if err := checkAbility(holder, def.Ability); err != nil {
			return err
		}
		if err := checkEffects(holder, def.Effects); err != nil {
			return err
		}
	}
	for _, id := range slices.Sorted(maps.Keys(tags)) {
		holder := fmt.Sprintf("tag %q", id)
		if err := checkAbility(holder, tags[id].Ability); err != nil {
			return err
		}
		if err := checkEffects(holder, tags[id].Effects); err != nil {
			return err
		}
	}
	return nil
}
```

Delete `toAssetType`:

**Find:**
```go
func toAssetType(s string) (domain.AssetType, error) {
	switch s {
	case "Military Unit":
		return domain.TypeMilitaryUnit, nil
	case "Special Forces":
		return domain.TypeSpecialForces, nil
	case "Facility":
		return domain.TypeFacility, nil
	case "Starship":
		return domain.TypeStarship, nil
	case "Tactic":
		return domain.TypeTactic, nil
	case "Logistics Facility":
		return domain.TypeLogisticsFacility, nil
	case "Special":
		return domain.TypeSpecial, nil
	default:
		return "", fmt.Errorf("unknown asset type: %q", s)
	}
}

```
**Replace with:** nothing.

##### Task 3 — `internal/faction/rulebook/effects.go`

In `convertEffectFilters`:

**Find:**
```go
	for i, value := range record.AssetType {
		assetType, err := toAssetType(value)
		if err != nil {
			return domain.EffectFilters{}, fmt.Errorf("filters.asset_type[%d]: %w", i, err)
		}
		filters.AssetTypes = append(filters.AssetTypes, assetType)
	}
```
**Replace with:**
```go
	for _, value := range record.AssetType {
		filters.AssetTypes = append(filters.AssetTypes, domain.AssetType(value))
	}
```

##### Task 4 — `internal/faction/rulebook/rulebook_test.go`

**Find:**
```go
		{name: "drift costs left out", body: "[purchase_population]\nStarship = 500000\n", want: "drift_costs is required"},
		{name: "unknown key", body: "drift_costs = [5]\nstarship_floor = 500000\n", want: "unknown keys in"},
		{name: "unknown asset type", body: "drift_costs = [5]\n[purchase_population]\nSpaceship = 500000\n", want: "purchase_population"},
		{name: "floor not above 0", body: "drift_costs = [5]\n[purchase_population]\nStarship = 0\n", want: "must be > 0"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "settings.toml")
			if err := os.WriteFile(path, []byte(testCase.body), 0o644); err != nil {
				t.Fatalf("writing fixture: %v", err)
			}
			_, _, err := loadSettings(path)
```
**Replace with:**
```go
		{name: "drift costs left out", body: "asset_types = [\"Starship\"]\n[purchase_population]\nStarship = 500000\n", want: "drift_costs is required"},
		{name: "asset types left out", body: "drift_costs = [5]\n", want: "asset_types is required"},
		{name: "asset type listed twice", body: "drift_costs = [5]\nasset_types = [\"Starship\", \"Starship\"]\n", want: "listed twice"},
		{name: "unknown key", body: "drift_costs = [5]\nasset_types = [\"Starship\"]\nstarship_floor = 500000\n", want: "unknown keys in"},
		{name: "undeclared asset type", body: "drift_costs = [5]\nasset_types = [\"Starship\"]\n[purchase_population]\nSpaceship = 500000\n", want: "not in asset_types"},
		{name: "floor not above 0", body: "drift_costs = [5]\nasset_types = [\"Starship\"]\n[purchase_population]\nStarship = 0\n", want: "must be > 0"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "settings.toml")
			if err := os.WriteFile(path, []byte(testCase.body), 0o644); err != nil {
				t.Fatalf("writing fixture: %v", err)
			}
			_, err := loadSettings(path)
```

**Find:**
```go
	if err := os.WriteFile(path, []byte("drift_costs = [3, 2]\n"), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	driftCosts, purchasePopulation, err := loadSettings(path)
	if err != nil {
		t.Fatalf("loadSettings: %v", err)
	}
	if !reflect.DeepEqual(driftCosts, []int{3, 2}) {
		t.Errorf("drift costs = %v, want [3 2]", driftCosts)
	}
	if len(purchasePopulation) != 0 {
		t.Errorf("purchase population = %v, want no floor for any type", purchasePopulation)
	}
}
```
**Replace with:**
```go
	if err := os.WriteFile(path, []byte("drift_costs = [3, 2]\nasset_types = [\"Starship\"]\n"), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	loaded, err := loadSettings(path)
	if err != nil {
		t.Fatalf("loadSettings: %v", err)
	}
	if !reflect.DeepEqual(loaded.driftCosts, []int{3, 2}) {
		t.Errorf("drift costs = %v, want [3 2]", loaded.driftCosts)
	}
	if !reflect.DeepEqual(loaded.assetTypes, []domain.AssetType{"Starship"}) {
		t.Errorf("asset types = %v, want [Starship]", loaded.assetTypes)
	}
	if len(loaded.purchasePopulation) != 0 {
		t.Errorf("purchase population = %v, want no floor for any type", loaded.purchasePopulation)
	}
}

func TestCheckAssetTypes(t *testing.T) {
	declared := []domain.AssetType{"Starship", "Facility"}
	starship := func(def domain.AssetDefinition) map[string]*domain.AssetDefinition {
		def.ID, def.Type = "A-1", "Starship"
		return map[string]*domain.AssetDefinition{"A-1": &def}
	}
	cases := []struct {
		name   string
		assets map[string]*domain.AssetDefinition
		tags   map[string]*domain.Tag
		want   string
	}{
		{name: "every name declared", assets: starship(domain.AssetDefinition{Transport: &domain.TransportProfile{CargoTypes: []domain.AssetType{"Facility"}}})},
		{name: "asset type", assets: map[string]*domain.AssetDefinition{"A-1": {ID: "A-1", Type: "Gunboat"}}, want: `asset "A-1": asset type "Gunboat"`},
		{name: "transport cargo", assets: starship(domain.AssetDefinition{Transport: &domain.TransportProfile{CargoTypes: []domain.AssetType{"Gunboat"}}}), want: `asset "A-1" transport`},
		{name: "ability relocate", assets: starship(domain.AssetDefinition{Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{Relocate: &domain.RelocateParams{ExcludeTypes: []domain.AssetType{"Gunboat"}}}}}}), want: `asset "A-1" ability step 0`},
		{name: "asset effects filter", assets: starship(domain.AssetDefinition{Effects: []domain.EffectDefinition{{Filters: domain.EffectFilters{AssetTypes: []domain.AssetType{"Gunboat"}}}}}), want: `asset "A-1" effects[0]`},
		{name: "tag effects filter", tags: map[string]*domain.Tag{"T-1": {ID: "T-1", Effects: []domain.EffectDefinition{{Filters: domain.EffectFilters{AssetTypes: []domain.AssetType{"Gunboat"}}}}}}, want: `tag "T-1" effects[0]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkAssetTypes(tc.assets, tc.tags, declared)
			if tc.want == "" {
				if err != nil {
					t.Errorf("err = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}
```

##### Task 5 — the asset type constants in tests

Task 1 removes the constants. In each file below, replace every use, with Edit and `replace_all`:

| Find | Replace with |
|------|--------------|
| `domain.TypeMilitaryUnit` | `domain.AssetType("Military Unit")` |
| `domain.TypeSpecialForces` | `domain.AssetType("Special Forces")` |
| `domain.TypeFacility` | `domain.AssetType("Facility")` |
| `domain.TypeStarship` | `domain.AssetType("Starship")` |
| `domain.TypeTactic` | `domain.AssetType("Tactic")` |
| `domain.TypeLogisticsFacility` | `domain.AssetType("Logistics Facility")` |
| `domain.TypeSpecial` | `domain.AssetType("Special")` |

Apply `domain.TypeSpecialForces` before `domain.TypeSpecial`, so the shorter name never matches inside the longer one.

Files: `internal/faction/domain/asset_test.go`, `internal/faction/domain/effect_test.go`, `internal/faction/rulebook/rulebook_test.go`, `internal/faction/rulebook/effects_test.go`, `internal/faction/engine/orchestrator_test.go`, `internal/faction/engine/arrival_test.go`, `internal/faction/engine/effect/effect_test.go`, `internal/faction/engine/effect/effects/shelter_asset_test.go`, `internal/faction/engine/effect/effects/grant_stealth_test.go`, `internal/faction/engine/effect/effects/restrict_attack_test.go`, `internal/faction/engine/effect/effects/population_floor_test.go`, `internal/faction/engine/effect/effects/grant_movement_test.go`, `internal/faction/engine/effect/effects/restore_asset_test.go`, `internal/faction/engine/testharness/scenarios/transport_lifecycle_test.go`, `internal/faction/engine/hooks/dispatch/rules_test.go`, `internal/faction/engine/action/actions/attack_test.go`, `internal/faction/engine/action/actions/ability/relocate_test.go`, `internal/faction/engine/relocate/relocate_test.go`, `internal/faction/engine/action/actions/refit_asset_test.go`, `internal/faction/engine/action/actions/buy_asset_test.go`.

##### Task 6 — `rulebooks/swn/settings.toml`

Full contents:

```toml
# Setting rules for this rulebook. drift_costs and asset_types are required; every
# other key is optional, and a key left out turns its rule off.

# drift_costs[i] = hex-equivalent crossing cost for drift rating (i+1)
drift_costs = [5, 4, 3, 2, 1]

# Every asset type this rulebook's data may name: an asset's type, a transport's
# cargo, a relocate's filter, an effects filter and the floors below.
asset_types = ["Military Unit", "Special Forces", "Facility", "Starship", "Tactic", "Logistics Facility", "Special"]

# The least world population on which an asset of each type may be bought. A
# type left out has no floor. Quote a type with a space: "Military Unit".
[purchase_population]
Starship = 500000
```

##### Task 7 — re-compose the sandboxes

The composed copies under `campaigns/sandbox/rulebook/` (tracked) and `campaigns/sandbox-claude/rulebook/` (gitignored) fail to load without `asset_types`. From the repo root:

```sh
go build -o gm-toolkit ./cmd/gm-toolkit
GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit campaign add-rules --rules rulebooks/swn --campaign sandbox --replace
GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit campaign add-rules --rules rulebooks/swn --campaign sandbox-claude --replace
```

`git diff campaigns/sandbox/rulebook/` shows only `settings.toml` changing.

##### Commit message

```
feat(E2): declare asset types in settings.toml

- settings.toml: asset_types is required and lists every asset type the
  rulebook's data may name; purchase_population keys must be in it
- rulebook: the loader checks asset, transport, relocate, effects-filter
  and floor names against asset_types; toAssetType is gone
- domain: the SWN asset type constants are gone; tests name types as data
- rulebooks/swn/settings.toml declares SWN's seven types; the sandbox
  rulebooks are re-composed
```

#### Commit 2 — `feat(E2): add the opponent_tag filter`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	FilterAssetDefinition   EffectFilterName = "asset_definition"
)
```
**Replace with:**
```go
	FilterAssetDefinition   EffectFilterName = "asset_definition"
	FilterOpponentTag       EffectFilterName = "opponent_tag"
)
```

**Find:**
```go
	// AssetDefinition names the rolling asset's definition ID (the Eugenics Cult
	// die on Gengineered Slaves).
	AssetDefinition string
}
```
**Replace with:**
```go
	// AssetDefinition names the rolling asset's definition ID (the Eugenics Cult
	// die on Gengineered Slaves).
	AssetDefinition string
	// OpponentTag names a tag the faction on the other side of the roll holds
	// (Godslayers' die against Soul-Bonded).
	OpponentTag string
}
```

In `Used`:

**Find:**
```go
	if filters.AssetDefinition != "" {
		present = append(present, FilterAssetDefinition)
	}
	return present
```
**Replace with:**
```go
	if filters.AssetDefinition != "" {
		present = append(present, FilterAssetDefinition)
	}
	if filters.OpponentTag != "" {
		present = append(present, FilterOpponentTag)
	}
	return present
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	AssetDefinition   string   `toml:"asset_definition"`
}
```
**Replace with:**
```go
	AssetDefinition   string   `toml:"asset_definition"`
	OpponentTag       string   `toml:"opponent_tag"`
}
```

**Find:**
```go
		Goal: record.Goal, Detection: record.Detection, AssetDefinition: record.AssetDefinition,
	}
```
**Replace with:**
```go
		Goal: record.Goal, Detection: record.Detection, AssetDefinition: record.AssetDefinition,
		OpponentTag: record.OpponentTag,
	}
```

##### Task 3 — `internal/faction/rulebook/rulebook.go`

**Find:**
```go
// checkReferences refuses a data-to-data reference to a record the rulebook does
// not have, checked once here: an effects entry's goal filter (D-5) or
// asset_definition filter (Effort 5 Decision 9), and an asset's requires_tag
// (D-13).
```
**Replace with:**
```go
// checkReferences refuses a data-to-data reference to a record the rulebook does
// not have, checked once here: an effects entry's goal filter (D-5),
// asset_definition filter (Effort 5 Decision 9) or opponent_tag filter (A.001.5
// Effort 2 Decision 6), and an asset's requires_tag (D-13).
```

**Find:**
```go
			if definition := entry.Filters.AssetDefinition; definition != "" && assets[definition] == nil {
				return fmt.Errorf("%s effects[%d]: unknown asset %q", holder, i, definition)
			}
```
**Replace with:**
```go
			if definition := entry.Filters.AssetDefinition; definition != "" && assets[definition] == nil {
				return fmt.Errorf("%s effects[%d]: unknown asset %q", holder, i, definition)
			}
			if tag := entry.Filters.OpponentTag; tag != "" && tags[tag] == nil {
				return fmt.Errorf("%s effects[%d]: unknown tag %q", holder, i, tag)
			}
```

##### Task 4 — `internal/faction/engine/effect/effects/bonus_die.go`

**Find:**
```go
				domain.FilterAssetDefinition,
			},
```
**Replace with:**
```go
				domain.FilterAssetDefinition, domain.FilterOpponentTag,
			},
```

##### Task 5 — `internal/faction/engine/effect/effects/roll_match.go`

**Find:**
```go
// what keeps an entry to its own side. attribute reads the stat this roll uses;
// opponent_attribute the stat the other side rolls against it. tech_level reads
// the rolling asset's definition, opponent_tech_level the other side's; neither
// matches without a rulebook. detection matches only a roll made by the side
// testing to detect stealth.
func rollMatches(ctx hooks.RollContext, trigger domain.EffectTrigger, ownerFactionID string, asset *domain.Asset, filters domain.EffectFilters, rulebook *rulebook.Rulebook) bool {
	if ctx.Actor == nil {
		return false
	}
	var owner *domain.Faction
	var subject *domain.Asset
	switch trigger {
	case domain.TriggerOwnRoll:
		if ctx.Actor.ID != ownerFactionID {
			return false
		}
		owner, subject = ctx.Actor, ctx.Asset
	case domain.TriggerEnemyRollAtMe:
		if ctx.Actor.ID == ownerFactionID || ctx.Opponent == nil || ctx.Opponent.ID != ownerFactionID {
			return false
		}
		owner, subject = ctx.Opponent, ctx.OpponentAsset
```
**Replace with:**
```go
// what keeps an entry to its own side. attribute reads the stat this roll uses;
// opponent_attribute the stat the other side rolls against it. tech_level reads
// the rolling asset's definition, opponent_tech_level the other side's; neither
// matches without a rulebook. detection matches only a roll made by the side
// testing to detect stealth. opponent_tag reads the tags of the faction across the
// roll from the owner: the opponent of its own roll, the roller of an enemy roll.
func rollMatches(ctx hooks.RollContext, trigger domain.EffectTrigger, ownerFactionID string, asset *domain.Asset, filters domain.EffectFilters, rulebook *rulebook.Rulebook) bool {
	if ctx.Actor == nil {
		return false
	}
	var owner, other *domain.Faction
	var subject *domain.Asset
	switch trigger {
	case domain.TriggerOwnRoll:
		if ctx.Actor.ID != ownerFactionID {
			return false
		}
		owner, other, subject = ctx.Actor, ctx.Opponent, ctx.Asset
	case domain.TriggerEnemyRollAtMe:
		if ctx.Actor.ID == ownerFactionID || ctx.Opponent == nil || ctx.Opponent.ID != ownerFactionID {
			return false
		}
		owner, other, subject = ctx.Opponent, ctx.Actor, ctx.OpponentAsset
```

**Find:**
```go
	if filters.AssetDefinition != "" && (ctx.Asset == nil || ctx.Asset.DefinitionID != filters.AssetDefinition) {
		return false
	}
	return true
}
```
**Replace with:**
```go
	if filters.AssetDefinition != "" && (ctx.Asset == nil || ctx.Asset.DefinitionID != filters.AssetDefinition) {
		return false
	}
	if filters.OpponentTag != "" && (other == nil || !other.HasTag(filters.OpponentTag)) {
		return false
	}
	return true
}
```

##### Task 6 — tests

`internal/faction/engine/effect/effects/roll_match_test.go`:

**Find:**
```go
	gamma := &domain.Faction{ID: "gamma"}
```
**Replace with:**
```go
	gamma := &domain.Faction{ID: "gamma"}
	devout := &domain.Faction{ID: "delta", Tags: []string{"T-9"}}
```

**Find:**
```go
	slavesRoll := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, AssetDefinition: "tl0"}
```
**Replace with:**
```go
	slavesRoll := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, AssetDefinition: "tl0"}
	againstDevout := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, OpponentTag: "T-9"}
	devoutRollsAtMe := domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld, OpponentTag: "T-9"}
```

**Find:**
```go
		{"unknown trigger", domain.TriggerTurnStart, self, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
```
**Replace with:**
```go
		{"opponent tag held", domain.TriggerOwnRoll, againstDevout, hooks.RollContext{Actor: alpha, Opponent: devout}, true},
		{"opponent tag not held", domain.TriggerOwnRoll, againstDevout, hooks.RollContext{Actor: alpha, Opponent: beta}, false},
		{"opponent tag, no opponent", domain.TriggerOwnRoll, againstDevout, hooks.RollContext{Actor: alpha}, false},
		{"opponent tag, enemy roll reads the roller", domain.TriggerEnemyRollAtMe, devoutRollsAtMe, hooks.RollContext{Actor: devout, Opponent: alpha, World: "Krylos"}, true},
		{"opponent tag, enemy roll by another", domain.TriggerEnemyRollAtMe, devoutRollsAtMe, hooks.RollContext{Actor: beta, Opponent: alpha, World: "Krylos"}, false},
		{"unknown trigger", domain.TriggerTurnStart, self, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
```

`internal/faction/rulebook/rulebook_test.go`, in `TestCheckReferences`:

**Find:**
```go
	if err := checkReferences(assets, withTag(naming(domain.EffectFilters{Goal: "CORE-G-004", AssetDefinition: "A-1"})), goals); err != nil {
```
**Replace with:**
```go
	if err := checkReferences(assets, withTag(naming(domain.EffectFilters{Goal: "CORE-G-004", AssetDefinition: "A-1", OpponentTag: "T-1"})), goals); err != nil {
```

**Find:**
```go
		{"unknown required tag", map[string]*domain.AssetDefinition{"A-1": {ID: "A-1", RequiresTag: "T-404"}}, tags, `asset "A-1": requires_tag names unknown tag "T-404"`},
```
**Replace with:**
```go
		{"unknown required tag", map[string]*domain.AssetDefinition{"A-1": {ID: "A-1", RequiresTag: "T-404"}}, tags, `asset "A-1": requires_tag names unknown tag "T-404"`},
		{"unknown opponent tag", assets, withTag(naming(domain.EffectFilters{OpponentTag: "T-404"})), `tag "T-1" effects[0]: unknown tag "T-404"`},
```

`internal/faction/rulebook/rulebook_test.go`, at the end of the file:

```go
func TestConvertEffectFilters_OpponentTag(t *testing.T) {
	filters, err := convertEffectFilters(effectFiltersRecord{Owner: "own", Scope: "faction", OpponentTag: "T-1"})
	if err != nil {
		t.Fatalf("convertEffectFilters: %v", err)
	}
	if filters.OpponentTag != "T-1" || !slices.Equal(filters.Used(), []domain.EffectFilterName{domain.FilterOpponentTag}) {
		t.Errorf("filters = %+v, used = %v; want OpponentTag T-1 and only opponent_tag used", filters, filters.Used())
	}
}
```

##### Commit message

```
feat(E2): add the opponent_tag filter

- domain: EffectFilters.OpponentTag, the opponent_tag filter name
- rulebook: the filter loads, and checkReferences refuses an unknown tag
- bonus_die accepts opponent_tag; rollMatches reads the tags of the
  faction across the roll from the owner
```

#### Commit 3 — `feat(E2): accept once_per and faction scope on coin_gain, base scope on rival_roll`

##### Task 1 — `internal/faction/engine/effect/effects/coin_gain.go`

**Find:**
```go
// coinGainRegistrar pays the owner the amount. turn_start fires at the owner's
// turn start. asset_destroyed reacts to each asset or Base destroyed in a fight: owner
// reads as whose asset it was ("own" the owner's, "rival" another faction's), and
// caused_by as who destroyed it. scope = "faction" is any world.
var coinGainRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerTurnStart: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
		domain.TriggerAssetDestroyed: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn, domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterCausedBy},
		},
	},
```
**Replace with:**
```go
// coinGainRegistrar pays the owner the amount. turn_start fires at the owner's
// turn start: scope = "self" for an asset, "faction" for a tag (Aether Refiners).
// asset_destroyed reacts to each asset or Base destroyed in a fight: owner reads as
// whose asset it was ("own" the owner's, "rival" another faction's), and caused_by
// as who destroyed it. scope = "faction" is any world. once_per caps the payouts
// (War Profiteers).
var coinGainRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerTurnStart: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf, domain.ScopeFaction},
		},
		domain.TriggerAssetDestroyed: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn, domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterCausedBy},
			Gates:   []domain.EffectGateName{domain.GateOncePer},
		},
	},
```

**Find:**
```go
			owner:    ctx.Entry.Filters.Owner,
			causedBy: ctx.Entry.Filters.CausedBy,
			amount:   ctx.Entry.CoinGain.Amount,
			roller:   ctx.Roller,
		},
```
**Replace with:**
```go
			owner:    ctx.Entry.Filters.Owner,
			causedBy: ctx.Entry.Filters.CausedBy,
			oncePer:  ctx.Entry.Gates.OncePer,
			amount:   ctx.Entry.CoinGain.Amount,
			roller:   ctx.Roller,
		},
```

**Find:**
```go
type coinGainReactor struct {
	holder   holderRef
	owner    domain.EffectOwner
	causedBy domain.EffectOwner
	amount   domain.DiceRoll
	roller   dice.Roller
}

// OnMutations pays once for each combat removal in the batch that the entry's
// owner and caused_by filters match.
```
**Replace with:**
```go
type coinGainReactor struct {
	holder   holderRef
	owner    domain.EffectOwner
	causedBy domain.EffectOwner
	oncePer  domain.OncePerKey
	amount   domain.DiceRoll
	roller   dice.Roller
}

// OnMutations pays once for each combat removal in the batch that the entry's
// owner and caused_by filters match, while its once_per budget lasts.
```

**Find:**
```go
		if reactor.causedBy != "" && !onSide(reactor.causedBy, causedBy, ownerID) {
			continue
		}
```
**Replace with:**
```go
		if reactor.causedBy != "" && !onSide(reactor.causedBy, causedBy, ownerID) {
			continue
		}
		if reactor.oncePer != "" && !spendBudget(factionState.Factions[ownerID], budgetKey(reactor.oncePer, domain.SpecialCoinGain, reactor.holder.sourceID, reactor.holder.assetID, lostBy)) {
			continue
		}
```

##### Task 2 — `internal/faction/engine/effect/effects/reroll.go`

**Find:**
```go
		domain.TriggerRivalRoll: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
```
**Replace with:**
```go
		domain.TriggerRivalRoll: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeFaction, domain.ScopeBase},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
```

**Find:**
```go
	asset, live := hook.holder.live(factionState)
	if !live || !rollMatches(ctx, hook.trigger, hook.holder.factionID, asset, hook.filters, rulebook) {
		return hooks.RerollDirective{}
	}
	forced := forcedTrigger(hook.trigger)
```
**Replace with:**
```go
	asset, live := hook.holder.live(factionState)
	if !live || !rollMatches(ctx, hook.trigger, hook.holder.factionID, asset, hook.filters, rulebook) {
		return hooks.RerollDirective{}
	}
	// deliberate: scope base is read here, not in rollMatches, because a rival_roll owner takes no part in the roll and only the hook holds the faction state (A.001.5 Effort 2 Decision 8).
	if hook.filters.Scope == domain.ScopeBase && !factionState.Factions[hook.holder.factionID].HasBaseOn(ctx.World) {
		return hooks.RerollDirective{}
	}
	forced := forcedTrigger(hook.trigger)
```

##### Task 3 — `internal/faction/engine/effect/effects/roll_match.go`

**Find:**
```go
		// deliberate: return here, not read scope or filters, because the owner takes no part in the roll and the registrar accepts only scope faction and no filters (Effort 3 Decision 7).
```
**Replace with:**
```go
		// deliberate: return here, not read scope or filters, because the owner takes no part in the roll and the registrar accepts no filters; scope base is read by the reroll hook (Effort 3 Decision 7; A.001.5 Effort 2 Decision 8).
```

##### Task 4 — tests

`internal/faction/engine/effect/effects/coin_gain_test.go` imports:

**Find:**
```go
import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```
**Replace with:**
```go
import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```

Append to `coin_gain_test.go`:

```go
func TestCoinGain_OncePerPaysOneRemovalATurn(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"HAWK-T-007"}, HookBudgets: map[string]int{}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}, "gamma": {ID: "gamma"}}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialCoinGain].Register(effect.RegisterContext{
		Holder:   effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "HAWK-T-007", Name: "War Profiteers"}},
		Registry: registry,
		Roller:   dicetest.Constant{Value: 1},
		Entry: domain.EffectDefinition{
			Trigger:  domain.TriggerAssetDestroyed,
			Effect:   domain.SpecialCoinGain,
			Filters:  domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeFaction},
			Gates:    domain.EffectGates{OncePer: domain.OncePerOwnerFaction},
			CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
		},
	})
	removals := []domain.Mutation{
		domain.AssetRemoved{FactionID: "beta", Cause: "attack", CausedByFactionID: "gamma"},
		domain.AssetRemoved{FactionID: "gamma", Cause: "attack", CausedByFactionID: "beta"},
	}
	paid := func() int {
		emitted, err := dispatch.MutationReactors(registry, factionState.Factions["beta"], removals, factionState, nil)
		if err != nil {
			t.Fatalf("MutationReactors: %v", err)
		}
		total := 0
		for _, mutation := range emitted {
			if gain, ok := mutation.(domain.CoinDelta); ok && gain.FactionID == "alpha" && gain.Cause == "coin_gain" {
				total += gain.Delta
			}
		}
		return total
	}

	if got := paid(); got != 1 {
		t.Errorf("first batch: paid %d, want 1 for two removals under once_per", got)
	}
	if got := paid(); got != 0 {
		t.Errorf("second batch: paid %d, want 0 with the budget spent", got)
	}
}

func TestCoinGain_TagPaysAtTurnStartUnlessLapsed(t *testing.T) {
	refiners := &domain.Tag{ID: "HAWK-T-006", Name: "Aether Refiners", Effects: []domain.EffectDefinition{{
		Trigger:  domain.TriggerTurnStart,
		Effect:   domain.SpecialCoinGain,
		Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
	}}}
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))
	if err := engine.CheckTagEffects(refiners); err != nil {
		t.Fatalf("CheckTagEffects: %v", err)
	}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{refiners.ID}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha}}
	book := &rulebook.Rulebook{Tags: map[string]*domain.Tag{refiners.ID: refiners}}

	got := engine.FireTurnStart(alpha, factionState, book, hooks.NewRegistry(), dicetest.Constant{Value: 1})
	want := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 1, Cause: "coin_gain", CausedByFactionID: "alpha"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FireTurnStart = %+v, want %+v", got, want)
	}
}
```

Commit 6 adds the lapsed half of the second test.

`internal/faction/engine/effect/effects/reroll_test.go`, after `TestReroll_RivalRollIsForTagsOnly`:

```go
func TestReroll_RivalRollAtBaseScopeNeedsTheOwnersBase(t *testing.T) {
	alpha := &domain.Faction{
		ID: "alpha", Tags: []string{"HAWK-T-005"}, HookBudgets: map[string]int{},
		Bases: []*domain.Base{{ID: "b-1", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}}},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}, "gamma": {ID: "gamma"}}}
	entry := psychicAcademyEntry()
	entry.Filters.Scope = domain.ScopeBase
	tag := &domain.Tag{ID: "HAWK-T-005", Name: "Contract-Bound", Effects: []domain.EffectDefinition{entry}}
	if err := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler)).CheckTagEffects(tag); err != nil {
		t.Fatalf("CheckTagEffects: %v", err)
	}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialReroll].Register(effect.RegisterContext{Holder: effect.Holder{Faction: alpha, Tag: tag}, Entry: entry, Registry: registry})
	collector := &rollCollector{accept: true}

	away := rollThroughDispatcher(factionState, registry, collector, rollOn(factionState, "beta", "gamma", "Tartarus"), 8)
	if away.Sum != 8 || collector.rerolls != 0 {
		t.Errorf("roll off alpha's bases: Sum = %d, rerolls asked = %d; want 8 and 0", away.Sum, collector.rerolls)
	}
	atBase := rollThroughDispatcher(factionState, registry, collector, rollOn(factionState, "beta", "gamma", "Krylos"), 8, 2)
	if atBase.Sum != 2 || collector.rerolls != 1 {
		t.Errorf("roll on alpha's base: Sum = %d, rerolls asked = %d; want 2 and 1", atBase.Sum, collector.rerolls)
	}
}
```

##### Commit message

```
feat(E2): accept once_per and faction scope on coin_gain, base scope on rival_roll

- coin_gain: asset_destroyed takes a once_per gate (War Profiteers);
  turn_start takes scope faction, so a tag earns at turn start (Aether
  Refiners)
- reroll: rival_roll takes scope base, read by the reroll hook against
  the owner's live Bases (Contract-Bound); the rollMatches fence-sign
  says so
```

#### Commit 4 — `fix(E2): name the government tag from the rulebook in the wizard`

##### Task 1 — `internal/faction/tui/views/manage/wizard/wizard.go`

In `buildForm`:

**Find:**
```go
			Title("Governed world (Planetary Government)").
```
**Replace with:**
```go
			Title(governmentTitle(rb)).
```

After `pickedGovernmentTag`:

**Find:**
```go
func pickedGovernmentTag(tags []*domain.Tag) bool {
	return slices.ContainsFunc(tags, func(tag *domain.Tag) bool { return tag.Government })
}
```
**Replace with:**
```go
func pickedGovernmentTag(tags []*domain.Tag) bool {
	return slices.ContainsFunc(tags, func(tag *domain.Tag) bool { return tag.Government })
}

// governmentTitle names the governed-world select after the rulebook's government
// tag, or plainly when the rulebook has none.
func governmentTitle(rb *rulebook.Rulebook) string {
	if tag := rb.Tags[rb.GovernmentTagID]; tag != nil {
		return fmt.Sprintf("Governed world (%s)", tag.Name)
	}
	return "Governed world"
}
```

##### Task 2 — `internal/faction/tui/views/manage/wizard/wizard_test.go`

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```

Append:

```go
func TestGovernmentTitle(t *testing.T) {
	proprietor := &rulebook.Rulebook{GovernmentTagID: "HAWK-T-001", Tags: map[string]*domain.Tag{"HAWK-T-001": {ID: "HAWK-T-001", Name: "Proprietor"}}}
	if got := governmentTitle(proprietor); got != "Governed world (Proprietor)" {
		t.Errorf("title = %q, want the rulebook's government tag named", got)
	}
	if got := governmentTitle(&rulebook.Rulebook{}); got != "Governed world" {
		t.Errorf("title with no government tag = %q, want %q", got, "Governed world")
	}
}
```

##### Task 3 — this plan

In the Work Breakdown Overview, `- [ ] **Phase 1 — Rulebook vocabulary**` becomes `- [x] **Phase 1 — Rulebook vocabulary**`.

##### Commit message

```
fix(E2): name the government tag from the rulebook in the wizard

- the governed-world select took SWN's tag name, "Planetary Government";
  it now reads the rulebook's government tag
- E2 plan: Phase 1 ticked
```

### Phase 2 — New tag mechanics

**In plain words:** Two new kinds of effect a tag can carry: a bonus to an asset's drift rating, so crossing the Drift costs less, and an upkeep each faction pays in bookkeeping, with a lapse, a lost tag or damage when it can't. A lapsed tag keeps paying, and the payment that succeeds turns it back on.

#### Commit 5 — `feat(E2): add the drift rating bonus`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialArrivalToll        SpecialEffectType = "arrival_toll"
)
```
**Replace with:**
```go
	SpecialArrivalToll        SpecialEffectType = "arrival_toll"
	SpecialDriftRatingBonus   SpecialEffectType = "drift_rating_bonus"
)
```

**Find:**
```go
	TriggerArrivalQuery           EffectTrigger = "arrival_query"
)
```
**Replace with:**
```go
	TriggerArrivalQuery           EffectTrigger = "arrival_query"
	TriggerDriftRatingQuery       EffectTrigger = "drift_rating_query"
)
```

**Find:**
```go
// PopulationFloorParams is a holder's floor as a percent of the setting's.
type PopulationFloorParams struct {
	Percent int
}
```
**Replace with:**
```go
// PopulationFloorParams is a holder's floor as a percent of the setting's.
type PopulationFloorParams struct {
	Percent int
}

// DriftRatingBonusParams raises the drift rating the holder's assets cross with.
type DriftRatingBonusParams struct {
	Amount int
}
```

The new field is the longest, so every field realigns.

**Find:**
```go
type EffectDefinition struct {
	Trigger         EffectTrigger
	Effect          SpecialEffectType
	Filters         EffectFilters
	Gates           EffectGates
	CoinGain        *CoinGainParams
	CoinSteal       *CoinStealParams
	SelfDamage      *SelfDamageParams
	BonusDie        *BonusDieParams
	Reroll          *RerollParams
	TechLevelFloor  *TechLevelFloorParams
	CostSurcharge   *CostSurchargeParams
	RestrictAttack  *RestrictAttackParams
	GrantMovement   *GrantMovementParams
	CaptureTarget   *CaptureTargetParams
	DiscountRepair  *DiscountRepairParams
	CostDiscount    *CostDiscountParams
	SpeedFloor      *SpeedFloorParams
	XPRoll          *TargetRoll
	PopulationFloor *PopulationFloorParams
	ArrivalToll     *ArrivalTollParams
	LockOut         *LockOutParams
}
```
**Replace with:**
```go
type EffectDefinition struct {
	Trigger          EffectTrigger
	Effect           SpecialEffectType
	Filters          EffectFilters
	Gates            EffectGates
	CoinGain         *CoinGainParams
	CoinSteal        *CoinStealParams
	SelfDamage       *SelfDamageParams
	BonusDie         *BonusDieParams
	Reroll           *RerollParams
	TechLevelFloor   *TechLevelFloorParams
	CostSurcharge    *CostSurchargeParams
	RestrictAttack   *RestrictAttackParams
	GrantMovement    *GrantMovementParams
	CaptureTarget    *CaptureTargetParams
	DiscountRepair   *DiscountRepairParams
	CostDiscount     *CostDiscountParams
	SpeedFloor       *SpeedFloorParams
	XPRoll           *TargetRoll
	PopulationFloor  *PopulationFloorParams
	ArrivalToll      *ArrivalTollParams
	LockOut          *LockOutParams
	DriftRatingBonus *DriftRatingBonusParams
}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

The new field is the longest, so every field realigns.

**Find:**
```go
type effectRecord struct {
	Trigger         string                 `toml:"trigger"`
	Effect          string                 `toml:"effect"`
	Filters         *effectFiltersRecord   `toml:"filters"`
	Gates           *effectGatesRecord     `toml:"gates"`
	CoinGain        *amountRecord          `toml:"coin_gain"`
	CoinSteal       *amountRecord          `toml:"coin_steal"`
	SelfDamage      *amountRecord          `toml:"self_damage"`
	BonusDie        *bonusDieRecord        `toml:"bonus_die"`
	Reroll          *rerollRecord          `toml:"reroll"`
	TechLevelFloor  *techLevelFloorRecord  `toml:"tech_level_floor"`
	CostSurcharge   *costSurchargeRecord   `toml:"cost_surcharge"`
	RestrictAttack  *restrictAttackRecord  `toml:"restrict_attack"`
	GrantMovement   *grantMovementRecord   `toml:"grant_movement"`
	CaptureTarget   *amountRecord          `toml:"capture_target"`
	DiscountRepair  *discountRepairRecord  `toml:"discount_repair"`
	CostDiscount    *costDiscountRecord    `toml:"cost_discount"`
	SpeedFloor      *speedFloorRecord      `toml:"speed_floor"`
	XPRoll          *targetRollRecord      `toml:"xp_roll"`
	PopulationFloor *populationFloorRecord `toml:"population_floor"`
	ArrivalToll     *arrivalTollRecord     `toml:"arrival_toll"`
	LockOut         *amountRecord          `toml:"lock_out"`
}
```
**Replace with:**
```go
type effectRecord struct {
	Trigger          string                  `toml:"trigger"`
	Effect           string                  `toml:"effect"`
	Filters          *effectFiltersRecord    `toml:"filters"`
	Gates            *effectGatesRecord      `toml:"gates"`
	CoinGain         *amountRecord           `toml:"coin_gain"`
	CoinSteal        *amountRecord           `toml:"coin_steal"`
	SelfDamage       *amountRecord           `toml:"self_damage"`
	BonusDie         *bonusDieRecord         `toml:"bonus_die"`
	Reroll           *rerollRecord           `toml:"reroll"`
	TechLevelFloor   *techLevelFloorRecord   `toml:"tech_level_floor"`
	CostSurcharge    *costSurchargeRecord    `toml:"cost_surcharge"`
	RestrictAttack   *restrictAttackRecord   `toml:"restrict_attack"`
	GrantMovement    *grantMovementRecord    `toml:"grant_movement"`
	CaptureTarget    *amountRecord           `toml:"capture_target"`
	DiscountRepair   *discountRepairRecord   `toml:"discount_repair"`
	CostDiscount     *costDiscountRecord     `toml:"cost_discount"`
	SpeedFloor       *speedFloorRecord       `toml:"speed_floor"`
	XPRoll           *targetRollRecord       `toml:"xp_roll"`
	PopulationFloor  *populationFloorRecord  `toml:"population_floor"`
	ArrivalToll      *arrivalTollRecord      `toml:"arrival_toll"`
	LockOut          *amountRecord           `toml:"lock_out"`
	DriftRatingBonus *driftRatingBonusRecord `toml:"drift_rating_bonus"`
}
```

**Find:**
```go
type populationFloorRecord struct {
	Percent int `toml:"percent"`
}
```
**Replace with:**
```go
type populationFloorRecord struct {
	Percent int `toml:"percent"`
}

type driftRatingBonusRecord struct {
	Amount int `toml:"amount"`
}
```

**Find:**
```go
		domain.TriggerArrivalQuery,
	}
```
**Replace with:**
```go
		domain.TriggerArrivalQuery, domain.TriggerDriftRatingQuery,
	}
```

**Find:**
```go
		domain.SpecialRemoveTag, domain.SpecialXPRoll, domain.SpecialPopulationFloor, domain.SpecialArrivalToll,
	}
```
**Replace with:**
```go
		domain.SpecialRemoveTag, domain.SpecialXPRoll, domain.SpecialPopulationFloor, domain.SpecialArrivalToll,
		domain.SpecialDriftRatingBonus,
	}
```

In `presentParamBlocks`:

**Find:**
```go
	if record.LockOut != nil {
		present = append(present, domain.SpecialLockOut)
	}
	return present
```
**Replace with:**
```go
	if record.LockOut != nil {
		present = append(present, domain.SpecialLockOut)
	}
	if record.DriftRatingBonus != nil {
		present = append(present, domain.SpecialDriftRatingBonus)
	}
	return present
```

In `convertEffectParams`:

**Find:**
```go
		definition.ArrivalToll = &domain.ArrivalTollParams{Amount: record.ArrivalToll.Amount}
	}
	return nil
}
```
**Replace with:**
```go
		definition.ArrivalToll = &domain.ArrivalTollParams{Amount: record.ArrivalToll.Amount}
	case domain.SpecialDriftRatingBonus:
		if record.DriftRatingBonus == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.DriftRatingBonus.Amount <= 0 {
			return fmt.Errorf("drift_rating_bonus amount must be > 0, got %d", record.DriftRatingBonus.Amount)
		}
		definition.DriftRatingBonus = &domain.DriftRatingBonusParams{Amount: record.DriftRatingBonus.Amount}
	}
	return nil
}
```

##### Task 3 — `internal/faction/rulebook/rulebook.go`

**Find:**
```go
// DriftCost returns the hex-equivalent crossing cost for the given drift rating.
// Out-of-range ratings fall back to DriftCosts[0] (worst cost).
func (rulebook *Rulebook) DriftCost(driftRating int) int {
	if driftRating < 1 || driftRating > len(rulebook.DriftCosts) {
		return rulebook.DriftCosts[0]
	}
	return rulebook.DriftCosts[driftRating-1]
}
```
**Replace with:**
```go
// DriftCost returns the hex-equivalent crossing cost for the given drift rating. A
// rating below 1 counts as 1, the worst cost; one above the table counts as its
// last, the best cost, so a drift rating bonus never makes a crossing dearer.
func (rulebook *Rulebook) DriftCost(driftRating int) int {
	return rulebook.DriftCosts[min(max(driftRating, 1), len(rulebook.DriftCosts))-1]
}
```

##### Task 4 — hooks

`internal/faction/engine/hooks/rule_modifier.go`:

**Find:**
```go
type AssetSpeedModifier interface {
	ModifyAssetSpeed(faction *domain.Faction, def *domain.AssetDefinition, speed int) int
}
```
**Replace with:**
```go
type AssetSpeedModifier interface {
	ModifyAssetSpeed(faction *domain.Faction, def *domain.AssetDefinition, speed int) int
}

// AssetDriftRatingModifier adjusts the drift rating an asset faction owns crosses
// the Drift with, read when a movement order is priced (Cat 4).
type AssetDriftRatingModifier interface {
	ModifyAssetDriftRating(faction *domain.Faction, def *domain.AssetDefinition, rating int) int
}
```

`internal/faction/engine/hooks/registry.go`:

**Find:**
```go
type RegisteredArrivalTollRule struct {
```
**Replace with:**
```go
type RegisteredAssetDriftRatingModifier struct {
	Scope  Scope
	Source string
	Hook   AssetDriftRatingModifier
}

type RegisteredArrivalTollRule struct {
```

**Find:**
```go
	permissionContesters     map[Scope][]RegisteredPermissionContester
}
```
**Replace with:**
```go
	permissionContesters     map[Scope][]RegisteredPermissionContester
	driftRatingModifiers     map[Scope][]RegisteredAssetDriftRatingModifier
}
```

In `Reset`:

**Find:**
```go
	registry.permissionContesters = make(map[Scope][]RegisteredPermissionContester)
}
```
**Replace with:**
```go
	registry.permissionContesters = make(map[Scope][]RegisteredPermissionContester)
	registry.driftRatingModifiers = make(map[Scope][]RegisteredAssetDriftRatingModifier)
}
```

In `ClearAsset`:

**Find:**
```go
	delete(registry.permissionContesters, scope)
}
```
**Replace with:**
```go
	delete(registry.permissionContesters, scope)
	delete(registry.driftRatingModifiers, scope)
}
```

**Find:**
```go
// RegisterArrivalTollRule registers a Cat 4 arrival-toll hook under the given scope.
```
**Replace with:**
```go
// RegisterAssetDriftRatingModifier registers a Cat 4 drift-rating hook under the given scope.
func (registry *Registry) RegisterAssetDriftRatingModifier(scope Scope, source string, hook AssetDriftRatingModifier) {
	registry.driftRatingModifiers[scope] = append(registry.driftRatingModifiers[scope], RegisteredAssetDriftRatingModifier{Scope: scope, Source: source, Hook: hook})
}

// RegisterArrivalTollRule registers a Cat 4 arrival-toll hook under the given scope.
```

**Find:**
```go
// ArrivalTollRulesAt returns the Cat 4 arrival-toll hooks of one bucket. There is
```
**Replace with:**
```go
// AssetDriftRatingModifiersFor returns the global and faction Cat 4 drift-rating
// hooks; the asset buckets come from AssetDriftRatingModifiersForAsset.
func (registry *Registry) AssetDriftRatingModifiersFor(factionID string) []RegisteredAssetDriftRatingModifier {
	var result []RegisteredAssetDriftRatingModifier
	result = append(result, registry.driftRatingModifiers[GlobalScope()]...)
	result = append(result, registry.driftRatingModifiers[FactionScope(factionID)]...)
	return result
}

// AssetDriftRatingModifiersForAsset returns the Cat 4 drift-rating hooks registered
// against one asset instance, without the global or faction buckets.
func (registry *Registry) AssetDriftRatingModifiersForAsset(factionID, assetInstanceID string) []RegisteredAssetDriftRatingModifier {
	return registry.driftRatingModifiers[AssetScope(factionID, assetInstanceID)]
}

// ArrivalTollRulesAt returns the Cat 4 arrival-toll hooks of one bucket. There is
```

##### Task 5 — `internal/faction/engine/hooks/dispatch/rules.go`

**Find:**
```go
// ArrivalTolls returns the tolls one of mover's assets owes on landing on world:
```
**Replace with:**
```go
// AssetDriftRatingResolver returns the drift rating an asset faction owns crosses
// the Drift with, by definition: the definition's own, a missing rating counted as
// 1 (A.001.5 Effort 2 Decision 9), run through faction's drift-rating hooks, then
// those registered against each asset faction owns, in asset ID order. With a nil
// registry it returns the definition's rating, counted the same way.
func AssetDriftRatingResolver(registry *hooks.Registry, faction *domain.Faction) func(*domain.AssetDefinition) int {
	base := func(definition *domain.AssetDefinition) int { return max(definition.DriftRating, 1) }
	if registry == nil {
		return base
	}
	modifiers := registry.AssetDriftRatingModifiersFor(faction.ID)
	for _, asset := range domain.SortedAssets(faction) {
		modifiers = append(modifiers, registry.AssetDriftRatingModifiersForAsset(faction.ID, asset.ID)...)
	}
	return func(definition *domain.AssetDefinition) int {
		rating := base(definition)
		for _, registered := range modifiers {
			rating = registered.Hook.ModifyAssetDriftRating(faction, definition, rating)
		}
		return rating
	}
}

// ArrivalTolls returns the tolls one of mover's assets owes on landing on world:
```

##### Task 6 — the registrar

`internal/faction/engine/effect/effects/drift_rating_bonus.go` (new) — full contents:

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// driftRatingBonusRegistrar reads owner = "own" as "the moving asset is the
// owner's", and scope = "faction" as "every asset the owner holds" (Drift-Runners).
// The entry carries the bonus; a movement order prices its crossing from the raised
// rating.
var driftRatingBonusRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerDriftRatingQuery: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerDriftRatingBonus,
}

func registerDriftRatingBonus(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetDriftRatingModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialDriftRatingBonus),
		&driftRatingBonusModifier{
			owner:  ctx.Holder.Faction,
			holder: refOf(ctx.Holder),
			amount: ctx.Entry.DriftRatingBonus.Amount,
		},
	)
}

// driftRatingBonusModifier holds its owner, not its holder's asset: a Cat 4 hook
// gets no FactionState (Decision Record 19).
type driftRatingBonusModifier struct {
	owner  *domain.Faction
	holder holderRef
	amount int
}

func (modifier *driftRatingBonusModifier) ModifyAssetDriftRating(faction *domain.Faction, _ *domain.AssetDefinition, rating int) int {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || faction.ID != modifier.owner.ID {
		return rating
	}
	return rating + modifier.amount
}
```

`internal/faction/engine/effect/effects/registrars.go`:

**Find:**
```go
		domain.SpecialArrivalToll:        arrivalTollRegistrar,
	}
```
**Replace with:**
```go
		domain.SpecialArrivalToll:        arrivalTollRegistrar,
		domain.SpecialDriftRatingBonus:   driftRatingBonusRegistrar,
	}
```

##### Task 7 — movement prices crossings from the resolved rating

`internal/faction/engine/world/movement.go`:

**Find:**
```go
func (engine *WorldEngine) BuildMovementMutations(decisions []MovementDecision, faction *domain.Faction, rulebook *rulebook.Rulebook, log *slog.Logger) ([]domain.Mutation, error) {
	var mutations []domain.Mutation

	for _, decision := range decisions {
```
**Replace with:**
```go
// BuildMovementMutations issues, revises and cancels movement orders. A crossing
// is priced from the drift rating registry's drift-rating hooks resolve
// (Drift-Runners).
func (engine *WorldEngine) BuildMovementMutations(decisions []MovementDecision, faction *domain.Faction, rulebook *rulebook.Rulebook, registry *hooks.Registry, log *slog.Logger) ([]domain.Mutation, error) {
	var mutations []domain.Mutation

	driftRatingOf := dispatch.AssetDriftRatingResolver(registry, faction)
	for _, decision := range decisions {
```

Replace both occurrences, with Edit and `replace_all`:

**Find:**
```go
			crossingCost := rulebook.DriftCost(def.DriftRating)
```
**Replace with:**
```go
			crossingCost := rulebook.DriftCost(driftRatingOf(def))
```

`internal/faction/engine/orchestrator.go`, in `prepareMovementDecisions`:

**Find:**
```go
	return worldEngine.BuildMovementMutations(decisions, faction, rulebook, worldLog)
```
**Replace with:**
```go
	return worldEngine.BuildMovementMutations(decisions, faction, rulebook, registry, worldLog)
```

`internal/faction/engine/world/world_test.go`:

**Find:**
```go
	mutations, err := we.BuildMovementMutations(decisions, faction, rb, slog.New(slog.DiscardHandler))
```
**Replace with:**
```go
	mutations, err := we.BuildMovementMutations(decisions, faction, rb, nil, slog.New(slog.DiscardHandler))
```

`internal/faction/engine/world/movement_test.go`:

**Find:**
```go
	mutations, err := we.BuildMovementMutations(decisions, faction, reviseRulebook(), slog.New(slog.DiscardHandler))
```
**Replace with:**
```go
	mutations, err := we.BuildMovementMutations(decisions, faction, reviseRulebook(), nil, slog.New(slog.DiscardHandler))
```

##### Task 8 — tests

`internal/faction/engine/effect/effects/drift_rating_bonus_test.go` (new) — full contents:

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

func TestDriftRatingBonus_RaisesTheOwnersAssets(t *testing.T) {
	runners := &domain.Tag{ID: "HAWK-T-012", Name: "Drift-Runners"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{runners.ID}}
	beta := &domain.Faction{ID: "beta"}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialDriftRatingBonus].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: runners},
		Entry: domain.EffectDefinition{
			Trigger:          domain.TriggerDriftRatingQuery,
			Effect:           domain.SpecialDriftRatingBonus,
			Filters:          domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
			DriftRatingBonus: &domain.DriftRatingBonusParams{Amount: 1},
		},
		Registry: registry,
	})
	unrated := &domain.AssetDefinition{ID: "unrated"}
	rated := &domain.AssetDefinition{ID: "rated", DriftRating: 3}

	cases := []struct {
		name       string
		faction    *domain.Faction
		definition *domain.AssetDefinition
		want       int
	}{
		{"a missing rating counts as 1, then gains the bonus", alpha, unrated, 2},
		{"a rated asset gains the bonus", alpha, rated, 4},
		{"a rival's asset keeps its rating", beta, rated, 3},
		{"a rival's unrated asset counts as 1", beta, unrated, 1},
	}
	for _, tc := range cases {
		if got := dispatch.AssetDriftRatingResolver(registry, tc.faction)(tc.definition); got != tc.want {
			t.Errorf("%s: drift rating = %d, want %d", tc.name, got, tc.want)
		}
	}

	alpha.Tags = nil
	if got := dispatch.AssetDriftRatingResolver(registry, alpha)(rated); got != 3 {
		t.Errorf("drift rating after the tag is dropped = %d, want 3", got)
	}
	if got := dispatch.AssetDriftRatingResolver(nil, alpha)(unrated); got != 1 {
		t.Errorf("nil registry, unrated asset: drift rating = %d, want 1", got)
	}
}
```

`internal/faction/rulebook/rulebook_test.go`, at the end of the file:

```go
func TestDriftCost_ClampsToTheTable(t *testing.T) {
	rb := &Rulebook{DriftCosts: []int{5, 4, 3, 2, 1}}
	cases := []struct {
		rating int
		want   int
	}{{0, 5}, {1, 5}, {3, 3}, {5, 1}, {6, 1}}
	for _, tc := range cases {
		if got := rb.DriftCost(tc.rating); got != tc.want {
			t.Errorf("DriftCost(%d) = %d, want %d", tc.rating, got, tc.want)
		}
	}
}
```

##### Commit message

```
feat(E2): add the drift rating bonus

- drift_rating_bonus on a drift_rating_query trigger raises the drift
  rating the owner's assets cross with (Drift-Runners)
- hooks: a Cat 4 drift-rating family and AssetDriftRatingResolver, which
  counts a missing rating as 1
- movement prices a crossing from the resolved rating; DriftCost counts
  a rating above the table as its best cost
```

#### Commit 6 — `feat(E2): add tag upkeep with lapse, loss and damage`

##### Task 1 — domain

`internal/faction/domain/effect.go`:

**Find:**
```go
	SpecialDriftRatingBonus   SpecialEffectType = "drift_rating_bonus"
)
```
**Replace with:**
```go
	SpecialDriftRatingBonus   SpecialEffectType = "drift_rating_bonus"
	SpecialCoinUpkeep         SpecialEffectType = "coin_upkeep"
)
```

**Find:**
```go
	TriggerDriftRatingQuery       EffectTrigger = "drift_rating_query"
)
```
**Replace with:**
```go
	TriggerDriftRatingQuery       EffectTrigger = "drift_rating_query"
	TriggerUpkeep                 EffectTrigger = "upkeep"
)
```

**Find:**
```go
// DriftRatingBonusParams raises the drift rating the holder's assets cross with.
type DriftRatingBonusParams struct {
	Amount int
}
```
**Replace with:**
```go
// DriftRatingBonusParams raises the drift rating the holder's assets cross with.
type DriftRatingBonusParams struct {
	Amount int
}

// UpkeepUnpaid names what an unpaid upkeep costs the tag's holder.
type UpkeepUnpaid string

const (
	UnpaidLapse   UpkeepUnpaid = "lapse"
	UnpaidLoseTag UpkeepUnpaid = "lose_tag"
	UnpaidDamage  UpkeepUnpaid = "damage"
)

// CoinUpkeepParams is a tag's turn upkeep: Amount Coin, charged in bookkeeping
// after income and before asset maintenance. OnUnpaid is what a miss costs; Damage
// is the faction HP an UnpaidDamage miss takes, and 0 otherwise.
type CoinUpkeepParams struct {
	Amount   int
	OnUnpaid UpkeepUnpaid
	Damage   int
}
```

**Find:**
```go
	DriftRatingBonus *DriftRatingBonusParams
}
```
**Replace with:**
```go
	DriftRatingBonus *DriftRatingBonusParams
	CoinUpkeep       *CoinUpkeepParams
}
```

`internal/faction/domain/faction.go`:

**Find:**
```go
	CompletionRatings map[string]map[FactionStat]int `toml:"completion_ratings,omitempty"`
}
```
**Replace with:**
```go
	CompletionRatings map[string]map[FactionStat]int `toml:"completion_ratings,omitempty"`
	// LapsedTags holds the held tags whose effects are off until their upkeep is
	// paid (A.001.5 Effort 2 Decision 11).
	LapsedTags []string `toml:"lapsed_tags,omitempty"`
}
```

**Find:**
```go
// HasBaseOn reports whether the faction has a Base of Influence on worldID.
```
**Replace with:**
```go
// HasLapsedTag reports whether the faction's tag tagID is lapsed: held, with its
// effects off until its upkeep is paid.
func (faction *Faction) HasLapsedTag(tagID string) bool {
	return slices.Contains(faction.LapsedTags, tagID)
}

// HasBaseOn reports whether the faction has a Base of Influence on worldID.
```

In `Clone`:

**Find:**
```go
	clone.Tags = slices.Clone(faction.Tags)
```
**Replace with:**
```go
	clone.Tags = slices.Clone(faction.Tags)
	clone.LapsedTags = slices.Clone(faction.LapsedTags)
```

`internal/faction/domain/mutation.go`:

**Find:**
```go
func (m TagRemoved) Type() string { return "tag_removed" }
```
**Replace with:**
```go
func (m TagRemoved) Type() string { return "tag_removed" }

// TagLapsed turns a held tag's effects off (Lapsed) or back on: an upkeep the
// faction could not pay, then one it paid.
type TagLapsed struct {
	FactionID         string `json:"faction_id"`
	TagID             string `json:"tag_id"`
	Lapsed            bool   `json:"lapsed"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m TagLapsed) Type() string { return "tag_lapsed" }
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	DriftRatingBonus *driftRatingBonusRecord `toml:"drift_rating_bonus"`
}
```
**Replace with:**
```go
	DriftRatingBonus *driftRatingBonusRecord `toml:"drift_rating_bonus"`
	CoinUpkeep       *coinUpkeepRecord       `toml:"coin_upkeep"`
}
```

**Find:**
```go
type driftRatingBonusRecord struct {
	Amount int `toml:"amount"`
}
```
**Replace with:**
```go
type driftRatingBonusRecord struct {
	Amount int `toml:"amount"`
}

type coinUpkeepRecord struct {
	Amount   int    `toml:"amount"`
	OnUnpaid string `toml:"on_unpaid"`
	Damage   int    `toml:"damage"`
}
```

**Find:**
```go
		domain.TriggerArrivalQuery, domain.TriggerDriftRatingQuery,
	}
```
**Replace with:**
```go
		domain.TriggerArrivalQuery, domain.TriggerDriftRatingQuery, domain.TriggerUpkeep,
	}
```

**Find:**
```go
		domain.SpecialDriftRatingBonus,
	}
```
**Replace with:**
```go
		domain.SpecialDriftRatingBonus, domain.SpecialCoinUpkeep,
	}
```

**Find:**
```go
	restrictAllows = []domain.RestrictAllow{domain.AllowOnlyListed, domain.AllowExceptListed}
)
```
**Replace with:**
```go
	restrictAllows = []domain.RestrictAllow{domain.AllowOnlyListed, domain.AllowExceptListed}
	upkeepUnpaids  = []domain.UpkeepUnpaid{domain.UnpaidLapse, domain.UnpaidLoseTag, domain.UnpaidDamage}
)
```

In `presentParamBlocks`:

**Find:**
```go
	if record.DriftRatingBonus != nil {
		present = append(present, domain.SpecialDriftRatingBonus)
	}
	return present
```
**Replace with:**
```go
	if record.DriftRatingBonus != nil {
		present = append(present, domain.SpecialDriftRatingBonus)
	}
	if record.CoinUpkeep != nil {
		present = append(present, domain.SpecialCoinUpkeep)
	}
	return present
```

In `convertEffectParams`:

**Find:**
```go
		definition.DriftRatingBonus = &domain.DriftRatingBonusParams{Amount: record.DriftRatingBonus.Amount}
	}
	return nil
}
```
**Replace with:**
```go
		definition.DriftRatingBonus = &domain.DriftRatingBonusParams{Amount: record.DriftRatingBonus.Amount}
	case domain.SpecialCoinUpkeep:
		if record.CoinUpkeep == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.CoinUpkeep.Amount <= 0 {
			return fmt.Errorf("coin_upkeep amount must be > 0, got %d", record.CoinUpkeep.Amount)
		}
		onUnpaid, err := toEnum("coin_upkeep on_unpaid", record.CoinUpkeep.OnUnpaid, upkeepUnpaids)
		if err != nil {
			return err
		}
		if onUnpaid == domain.UnpaidDamage && record.CoinUpkeep.Damage <= 0 {
			return fmt.Errorf("coin_upkeep damage must be > 0 with on_unpaid %q, got %d", onUnpaid, record.CoinUpkeep.Damage)
		}
		if onUnpaid != domain.UnpaidDamage && record.CoinUpkeep.Damage != 0 {
			return fmt.Errorf("coin_upkeep damage is only for on_unpaid %q, got on_unpaid %q", domain.UnpaidDamage, onUnpaid)
		}
		definition.CoinUpkeep = &domain.CoinUpkeepParams{Amount: record.CoinUpkeep.Amount, OnUnpaid: onUnpaid, Damage: record.CoinUpkeep.Damage}
	}
	return nil
}
```

##### Task 3 — hooks

`internal/faction/engine/hooks/rule_modifier.go`:

**Find:**
```go
// AssetDriftRatingModifier adjusts the drift rating an asset faction owns crosses
```
**Replace with:**
```go
// UpkeepCharger is a tag's turn upkeep, read by bookkeeping after income and
// before asset maintenance (Cat 4). UpkeepDue reports the upkeep while faction
// still holds the tag, lapsed or not: a lapsed tag keeps paying, so it can lift
// its own lapse.
type UpkeepCharger interface {
	UpkeepDue(faction *domain.Faction) (TagUpkeep, bool)
}

// TagUpkeep is one tag's upkeep for one turn.
type TagUpkeep struct {
	TagID  string
	Params domain.CoinUpkeepParams
}

// AssetDriftRatingModifier adjusts the drift rating an asset faction owns crosses
```

`internal/faction/engine/hooks/registry.go`:

**Find:**
```go
type RegisteredAssetDriftRatingModifier struct {
```
**Replace with:**
```go
type RegisteredUpkeepCharger struct {
	Scope  Scope
	Source string
	Hook   UpkeepCharger
}

type RegisteredAssetDriftRatingModifier struct {
```

**Find:**
```go
	driftRatingModifiers     map[Scope][]RegisteredAssetDriftRatingModifier
}
```
**Replace with:**
```go
	driftRatingModifiers     map[Scope][]RegisteredAssetDriftRatingModifier
	upkeepChargers           map[Scope][]RegisteredUpkeepCharger
}
```

**Find:**
```go
	registry.driftRatingModifiers = make(map[Scope][]RegisteredAssetDriftRatingModifier)
}
```
**Replace with:**
```go
	registry.driftRatingModifiers = make(map[Scope][]RegisteredAssetDriftRatingModifier)
	registry.upkeepChargers = make(map[Scope][]RegisteredUpkeepCharger)
}
```

**Find:**
```go
	delete(registry.driftRatingModifiers, scope)
}
```
**Replace with:**
```go
	delete(registry.driftRatingModifiers, scope)
	delete(registry.upkeepChargers, scope)
}
```

**Find:**
```go
// RegisterAssetDriftRatingModifier registers a Cat 4 drift-rating hook under the given scope.
```
**Replace with:**
```go
// RegisterUpkeepCharger registers a Cat 4 upkeep hook under the given scope.
func (registry *Registry) RegisterUpkeepCharger(scope Scope, source string, hook UpkeepCharger) {
	registry.upkeepChargers[scope] = append(registry.upkeepChargers[scope], RegisteredUpkeepCharger{Scope: scope, Source: source, Hook: hook})
}

// RegisterAssetDriftRatingModifier registers a Cat 4 drift-rating hook under the given scope.
```

**Find:**
```go
// AssetDriftRatingModifiersFor returns the global and faction Cat 4 drift-rating
```
**Replace with:**
```go
// UpkeepChargersFor returns the faction's Cat 4 upkeep hooks, in registration
// order. Only a tag holds upkeep, so there is no global or asset bucket.
func (registry *Registry) UpkeepChargersFor(factionID string) []RegisteredUpkeepCharger {
	return registry.upkeepChargers[FactionScope(factionID)]
}

// AssetDriftRatingModifiersFor returns the global and faction Cat 4 drift-rating
```

##### Task 4 — the registrar

`internal/faction/engine/effect/effects/coin_upkeep.go` (new) — full contents:

```go
package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// coinUpkeepRegistrar is a tag's turn upkeep: bookkeeping charges it after income
// and before asset maintenance, so tags are paid first (A.001.5 Effort 2 Decision
// 10). owner = "own" reads as "the holder pays", scope = "faction" as "once for the
// faction".
var coinUpkeepRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerUpkeep: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerCoinUpkeep,
	Validate: validateCoinUpkeep,
}

func validateCoinUpkeep(_ domain.EffectDefinition, holder effect.HolderKind) error {
	if holder != effect.HolderTag {
		return fmt.Errorf("only a tag pays upkeep; an asset pays maintenance")
	}
	return nil
}

func registerCoinUpkeep(ctx effect.RegisterContext) {
	ctx.Registry.RegisterUpkeepCharger(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialCoinUpkeep),
		&coinUpkeepCharger{tagID: ctx.Holder.Tag.ID, params: *ctx.Entry.CoinUpkeep},
	)
}

type coinUpkeepCharger struct {
	tagID  string
	params domain.CoinUpkeepParams
}

// deliberate: held, not live, because a lapsed tag still owes its upkeep and paying it is how the lapse lifts (A.001.5 Effort 2 Decision 11).
func (charger *coinUpkeepCharger) UpkeepDue(faction *domain.Faction) (hooks.TagUpkeep, bool) {
	if !faction.HasTag(charger.tagID) {
		return hooks.TagUpkeep{}, false
	}
	return hooks.TagUpkeep{TagID: charger.tagID, Params: charger.params}, true
}
```

`internal/faction/engine/effect/effects/registrars.go`:

**Find:**
```go
		domain.SpecialDriftRatingBonus:   driftRatingBonusRegistrar,
	}
```
**Replace with:**
```go
		domain.SpecialDriftRatingBonus:   driftRatingBonusRegistrar,
		domain.SpecialCoinUpkeep:         coinUpkeepRegistrar,
	}
```

##### Task 5 — `internal/faction/engine/turn/bookkeeping.go`

**Find:**
```go
	result := applyMaintenance(registry, faction, faction.Coin+total, &mutations, t.rulebook)
```
**Replace with:**
```go
	runningCoin := faction.Coin + total
	applyUpkeep(registry, faction, &runningCoin, &mutations)
	result := applyMaintenance(registry, faction, runningCoin, &mutations, t.rulebook)
```

**Find:**
```go
// applyMaintenance evaluates per-asset maintenance costs against startCoin,
```
**Replace with:**
```go
// applyUpkeep charges each upkeep a held tag declares against runningCoin, after
// income and before asset maintenance, so tags are paid first (A.001.5 Effort 2
// Decision 10). A paid upkeep lifts the tag's lapse; an unpaid one runs the entry's
// consequence: the tag lapses, the tag is lost, or the faction takes damage.
func applyUpkeep(registry *hooks.Registry, faction *domain.Faction, runningCoin *int, mutations *[]domain.Mutation) {
	if registry == nil {
		return
	}
	for _, registered := range registry.UpkeepChargersFor(faction.ID) {
		upkeep, due := registered.Hook.UpkeepDue(faction)
		if !due {
			continue
		}
		params := upkeep.Params
		lapsed := faction.HasLapsedTag(upkeep.TagID)
		if *runningCoin >= params.Amount {
			*runningCoin -= params.Amount
			*mutations = append(*mutations, domain.CoinDelta{FactionID: faction.ID, Delta: -params.Amount, Cause: "upkeep"})
			if lapsed {
				*mutations = append(*mutations, domain.TagLapsed{FactionID: faction.ID, TagID: upkeep.TagID, Lapsed: false, Cause: "upkeep", CausedByFactionID: faction.ID})
			}
			continue
		}
		switch params.OnUnpaid {
		case domain.UnpaidLapse:
			if !lapsed {
				*mutations = append(*mutations, domain.TagLapsed{FactionID: faction.ID, TagID: upkeep.TagID, Lapsed: true, Cause: "upkeep", CausedByFactionID: faction.ID})
			}
		case domain.UnpaidLoseTag:
			*mutations = append(*mutations, domain.TagRemoved{FactionID: faction.ID, TagID: upkeep.TagID, Cause: "upkeep", CausedByFactionID: faction.ID})
		case domain.UnpaidDamage:
			*mutations = append(*mutations, domain.FactionHPDelta{FactionID: faction.ID, Delta: -params.Damage, Cause: "upkeep", CausedByFactionID: faction.ID})
		}
	}
}

// applyMaintenance evaluates per-asset maintenance costs against startCoin,
```

##### Task 6 — `internal/faction/engine/mutation/mutation.go`

**Find:**
```go
		case domain.TagRemoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				index := slices.Index(faction.Tags, v.TagID)
				if index >= 0 {
					faction.Tags = slices.Delete(faction.Tags, index, index+1)
				} else {
```
**Replace with:**
```go
		case domain.TagRemoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				index := slices.Index(faction.Tags, v.TagID)
				if index >= 0 {
					faction.Tags = slices.Delete(faction.Tags, index, index+1)
					faction.LapsedTags = slices.DeleteFunc(faction.LapsedTags, func(tagID string) bool { return tagID == v.TagID })
				} else {
```

**Find:**
```go
		case domain.WorldGovernmentChanged:
```
**Replace with:**
```go
		case domain.TagLapsed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.LapsedTags = slices.DeleteFunc(faction.LapsedTags, func(tagID string) bool { return tagID == v.TagID })
				if v.Lapsed {
					faction.LapsedTags = append(faction.LapsedTags, v.TagID)
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.WorldGovernmentChanged:
```

##### Task 7 — a lapsed tag's effects are off

`internal/faction/engine/effect/effects/holder.go`:

**Find:**
```go
// live returns the declaring asset from current state, and whether the holder
// still applies: its asset is active, or its faction still holds the tag. A tag
// has no asset. A nil state (a dispatcher test) is no holder.
```
**Replace with:**
```go
// live returns the declaring asset from current state, and whether the holder
// still applies: its asset is active, or its faction still holds the tag and the
// tag is not lapsed. A tag has no asset. A nil state (a dispatcher test) is no
// holder.
```

**Find:**
```go
	if ref.tagID != "" {
		return nil, owner.HasTag(ref.tagID)
	}
```
**Replace with:**
```go
	if ref.tagID != "" {
		return nil, owner.HasTag(ref.tagID) && !owner.HasLapsedTag(ref.tagID)
	}
```

`internal/faction/engine/effect/effect.go`:

**Find:**
```go
func heldTags(faction *domain.Faction, rulebook *rulebook.Rulebook) []*domain.Tag {
	var tags []*domain.Tag
	for _, tagID := range faction.Tags {
		if tag := rulebook.Tags[tagID]; tag != nil {
			tags = append(tags, tag)
		}
	}
	return tags
}
```
**Replace with:**
```go
func heldTags(faction *domain.Faction, rulebook *rulebook.Rulebook) []*domain.Tag {
	var tags []*domain.Tag
	for _, tagID := range faction.Tags {
		if tag := rulebook.Tags[tagID]; tag != nil {
			tags = append(tags, tag)
		}
	}
	return tags
}

// activeTags returns the held tags whose effects are on: every held tag but a
// lapsed one. The phase and matchup routes read it. Hook registration reads
// heldTags, so a lapsed tag keeps its upkeep hook, and its other hooks wait on
// the holder's live check.
func activeTags(faction *domain.Faction, rulebook *rulebook.Rulebook) []*domain.Tag {
	return slices.DeleteFunc(heldTags(faction, rulebook), func(tag *domain.Tag) bool { return faction.HasLapsedTag(tag.ID) })
}
```

In `FireTurnStart`:

**Find:**
```go
	for _, tag := range heldTags(faction, rulebook) {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Tag: tag}, tag.Effects, domain.TriggerTurnStart, nil, factionState, rulebook, hookRegistry, roller)...)
	}
```
**Replace with:**
```go
	for _, tag := range activeTags(faction, rulebook) {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Tag: tag}, tag.Effects, domain.TriggerTurnStart, nil, factionState, rulebook, hookRegistry, roller)...)
	}
```

In `FirePurchase`:

**Find:**
```go
	for _, tag := range heldTags(faction, rulebook) {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Tag: tag}, tag.Effects, domain.TriggerOnPurchase, asset, factionState, rulebook, hookRegistry, roller)...)
	}
```
**Replace with:**
```go
	for _, tag := range activeTags(faction, rulebook) {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Tag: tag}, tag.Effects, domain.TriggerOnPurchase, asset, factionState, rulebook, hookRegistry, roller)...)
	}
```

In `FireMatchup`:

**Find:**
```go
	for _, tag := range heldTags(faction, rulebook) {
		fire(Holder{Faction: faction, Tag: tag}, tag.Effects)
	}
```
**Replace with:**
```go
	for _, tag := range activeTags(faction, rulebook) {
		fire(Holder{Faction: faction, Tag: tag}, tag.Effects)
	}
```

`internal/faction/engine/action/actions/use_asset_ability.go`, in `abilityCandidates`:

**Find:**
```go
		tag, ok := rulebook.Tags[tagID]
		if !ok || tag.Ability == nil {
			continue
		}
```
**Replace with:**
```go
		tag, ok := rulebook.Tags[tagID]
		if !ok || tag.Ability == nil || faction.HasLapsedTag(tagID) {
			continue
		}
```

##### Task 8 — `internal/faction/tui/views/turn/execution/format.go`

**Find:**
```go
		// XPSpent, MovementOrderProgressed, InfluenceDelta, TagAdded,
		// TagRemoved.
```
**Replace with:**
```go
		// XPSpent, MovementOrderProgressed, InfluenceDelta, TagAdded,
		// TagRemoved, TagLapsed.
```

##### Task 9 — tests

`internal/faction/engine/turn/bookkeeping_test.go`, imports:

**Find:**
```go
import (
	"log/slog"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```
**Replace with:**
```go
import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```

Append to `bookkeeping_test.go`:

```go
type stubUpkeep struct{ upkeep hooks.TagUpkeep }

func (stub stubUpkeep) UpkeepDue(*domain.Faction) (hooks.TagUpkeep, bool) { return stub.upkeep, true }

// upkeepBookkeeping runs one bookkeeping for faction f1, which holds tag T-1 with
// upkeep params, and returns the mutations after the income line.
func upkeepBookkeeping(t *testing.T, engine *TurnEngine, faction *domain.Faction, params domain.CoinUpkeepParams) []domain.Mutation {
	t.Helper()
	s := &state.FactionState{CampaignID: "test", Factions: map[string]*domain.Faction{"f1": faction}}
	if err := engine.Start(s); err != nil {
		t.Fatalf("Start: %v", err)
	}
	registry := hooks.NewRegistry()
	registry.RegisterUpkeepCharger(hooks.FactionScope("f1"), "coin_upkeep:T-1", stubUpkeep{hooks.TagUpkeep{TagID: "T-1", Params: params}})
	_, mutations, err := engine.ApplyBookkeeping(s, registry, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("ApplyBookkeeping: %v", err)
	}
	return mutations[1:]
}

func TestApplyBookkeeping_Upkeep(t *testing.T) {
	lapse := domain.CoinUpkeepParams{Amount: 1, OnUnpaid: domain.UnpaidLapse}
	paid := domain.CoinDelta{FactionID: "f1", Delta: -1, Cause: "upkeep"}
	cases := []struct {
		name    string
		faction *domain.Faction
		params  domain.CoinUpkeepParams
		want    []domain.Mutation
	}{
		{"paid", &domain.Faction{ID: "f1", Coin: 2, Tags: []string{"T-1"}}, lapse, []domain.Mutation{paid}},
		{"paid from this turn's income", &domain.Faction{ID: "f1", Wealth: 2, Tags: []string{"T-1"}}, lapse, []domain.Mutation{paid}},
		{"paid, lifting a lapse", &domain.Faction{ID: "f1", Coin: 1, Tags: []string{"T-1"}, LapsedTags: []string{"T-1"}}, lapse, []domain.Mutation{
			paid, domain.TagLapsed{FactionID: "f1", TagID: "T-1", Lapsed: false, Cause: "upkeep", CausedByFactionID: "f1"},
		}},
		{"unpaid, lapses", &domain.Faction{ID: "f1", Tags: []string{"T-1"}}, lapse, []domain.Mutation{
			domain.TagLapsed{FactionID: "f1", TagID: "T-1", Lapsed: true, Cause: "upkeep", CausedByFactionID: "f1"},
		}},
		{"unpaid, already lapsed", &domain.Faction{ID: "f1", Tags: []string{"T-1"}, LapsedTags: []string{"T-1"}}, lapse, nil},
		{"unpaid, loses the tag", &domain.Faction{ID: "f1", Tags: []string{"T-1"}}, domain.CoinUpkeepParams{Amount: 1, OnUnpaid: domain.UnpaidLoseTag}, []domain.Mutation{
			domain.TagRemoved{FactionID: "f1", TagID: "T-1", Cause: "upkeep", CausedByFactionID: "f1"},
		}},
		{"unpaid, takes damage", &domain.Faction{ID: "f1", Tags: []string{"T-1"}}, domain.CoinUpkeepParams{Amount: 1, OnUnpaid: domain.UnpaidDamage, Damage: 2}, []domain.Mutation{
			domain.FactionHPDelta{FactionID: "f1", Delta: -2, Cause: "upkeep", CausedByFactionID: "f1"},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := upkeepBookkeeping(t, newTurn(), tc.faction, tc.params)
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestApplyBookkeeping_UpkeepIsPaidBeforeMaintenance(t *testing.T) {
	book := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"gear": {ID: "gear", Category: domain.StatForce, Maintenance: 1}}}
	faction := &domain.Faction{
		ID: "f1", Force: 1, Coin: 1, Tags: []string{"T-1"},
		Assets: map[string]*domain.Asset{"g-1": {ID: "g-1", DefinitionID: "gear", Maintained: true}},
	}
	got := upkeepBookkeeping(t, New(&randRoller{}, book, slog.New(slog.DiscardHandler)), faction, domain.CoinUpkeepParams{Amount: 1, OnUnpaid: domain.UnpaidLapse})
	want := []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: -1, Cause: "upkeep"},
		domain.AssetMaintainedFlag{FactionID: "f1", AssetID: "g-1", Maintained: false, Cause: "bookkeeping"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mutations = %+v, want the upkeep paid and the asset unmaintained", got)
	}
}
```

`internal/faction/engine/effect/effects/coin_upkeep_test.go` (new) — full contents:

```go
package effects_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

func upkeepEntry() domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger:    domain.TriggerUpkeep,
		Effect:     domain.SpecialCoinUpkeep,
		Filters:    domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		CoinUpkeep: &domain.CoinUpkeepParams{Amount: 1, OnUnpaid: domain.UnpaidLapse},
	}
}

func TestCoinUpkeep_DueWhileHeldLapsedOrNot(t *testing.T) {
	tag := &domain.Tag{ID: "HAWK-T-009", Name: "Outsourced"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{tag.ID}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialCoinUpkeep].Register(effect.RegisterContext{Holder: effect.Holder{Faction: alpha, Tag: tag}, Entry: upkeepEntry(), Registry: registry})
	chargers := registry.UpkeepChargersFor("alpha")
	if len(chargers) != 1 {
		t.Fatalf("chargers = %d, want 1", len(chargers))
	}

	upkeep, due := chargers[0].Hook.UpkeepDue(alpha)
	if !due || upkeep.TagID != tag.ID || upkeep.Params.Amount != 1 {
		t.Errorf("held: upkeep = %+v, due = %v; want the tag's upkeep due", upkeep, due)
	}
	alpha.LapsedTags = []string{tag.ID}
	if _, due := chargers[0].Hook.UpkeepDue(alpha); !due {
		t.Error("lapsed: due = false, want the upkeep still due")
	}
	alpha.Tags = nil
	if _, due := chargers[0].Hook.UpkeepDue(alpha); due {
		t.Error("dropped: due = true, want no upkeep once the tag is gone")
	}
}

func TestCoinUpkeep_IsForTagsOnly(t *testing.T) {
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))
	asset := &domain.AssetDefinition{ID: "X-1", Effects: []domain.EffectDefinition{upkeepEntry()}}
	if err := engine.RegisterEffectsList(asset); err == nil || !strings.Contains(err.Error(), "only a tag pays upkeep") {
		t.Errorf("asset holder: err = %v, want an only-a-tag refusal", err)
	}
	if err := engine.CheckTagEffects(&domain.Tag{ID: "T-1", Effects: []domain.EffectDefinition{upkeepEntry()}}); err != nil {
		t.Errorf("tag holder: err = %v, want none", err)
	}
}
```

`internal/faction/engine/effect/effects/coin_gain_test.go`, at the end of `TestCoinGain_TagPaysAtTurnStartUnlessLapsed` (Commit 3, Task 4):

**Find:**
```go
	got := engine.FireTurnStart(alpha, factionState, book, hooks.NewRegistry(), dicetest.Constant{Value: 1})
	want := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 1, Cause: "coin_gain", CausedByFactionID: "alpha"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FireTurnStart = %+v, want %+v", got, want)
	}
}
```
**Replace with:**
```go
	got := engine.FireTurnStart(alpha, factionState, book, hooks.NewRegistry(), dicetest.Constant{Value: 1})
	want := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 1, Cause: "coin_gain", CausedByFactionID: "alpha"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FireTurnStart = %+v, want %+v", got, want)
	}

	alpha.LapsedTags = []string{refiners.ID}
	if got := engine.FireTurnStart(alpha, factionState, book, hooks.NewRegistry(), dicetest.Constant{Value: 1}); len(got) != 0 {
		t.Errorf("lapsed: FireTurnStart = %+v, want nothing", got)
	}
}
```

`internal/faction/engine/effect/effects/speed_floor_test.go`, at the end of `TestSpeedFloor_RaisesTheOwnersSlowAssets`:

**Find:**
```go
	alpha.Tags = nil
	if got := dispatch.AssetSpeedResolver(registry, alpha)(facility); got != 0 {
		t.Errorf("speed after the tag is dropped = %d, want 0", got)
	}
}
```
**Replace with:**
```go
	alpha.LapsedTags = []string{mercenaries.ID}
	if got := dispatch.AssetSpeedResolver(registry, alpha)(facility); got != 0 {
		t.Errorf("speed while the tag is lapsed = %d, want 0", got)
	}

	alpha.Tags = nil
	if got := dispatch.AssetSpeedResolver(registry, alpha)(facility); got != 0 {
		t.Errorf("speed after the tag is dropped = %d, want 0", got)
	}
}
```

`internal/faction/engine/mutation/mutation_test.go`, after `TestMutationEngine_TagRemoved`:

```go
func TestMutationEngine_TagLapsed(t *testing.T) {
	s := newMutationTestState()
	faction := s.Factions["f1"]
	faction.Tags = []string{"T-1", "T-2"}
	me := New(slog.New(slog.DiscardHandler))
	apply := func(mutation domain.Mutation) {
		t.Helper()
		if err := me.Apply(s, []domain.Mutation{mutation}, slog.New(slog.DiscardHandler)); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	}

	apply(domain.TagLapsed{FactionID: "f1", TagID: "T-1", Lapsed: true})
	apply(domain.TagLapsed{FactionID: "f1", TagID: "T-1", Lapsed: true})
	if !reflect.DeepEqual(faction.LapsedTags, []string{"T-1"}) {
		t.Errorf("after two lapses: LapsedTags = %v, want [T-1] once", faction.LapsedTags)
	}
	apply(domain.TagLapsed{FactionID: "f1", TagID: "T-1", Lapsed: false})
	if len(faction.LapsedTags) != 0 {
		t.Errorf("after the lift: LapsedTags = %v, want none", faction.LapsedTags)
	}
	apply(domain.TagLapsed{FactionID: "f1", TagID: "T-2", Lapsed: true})
	apply(domain.TagRemoved{FactionID: "f1", TagID: "T-2"})
	if len(faction.LapsedTags) != 0 {
		t.Errorf("after removing a lapsed tag: LapsedTags = %v, want none", faction.LapsedTags)
	}
}
```

`internal/faction/engine/action/actions/use_asset_ability_test.go`, after `TestUseAssetAbility_TagHolder`:

```go
func TestUseAssetAbility_LapsedTagIsNotOffered(t *testing.T) {
	tag := &domain.Tag{ID: "T-1", Name: "Brokers", Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
		Effect: domain.EffectPurchaseDiscount,
		Target: domain.TargetNone,
		Amount: &domain.DiceRoll{NumDice: 1, Sides: 4},
	}}}}
	book := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{}, Tags: map[string]*domain.Tag{"T-1": tag}}
	faction := &domain.Faction{ID: "f1", Tags: []string{"T-1"}, LapsedTags: []string{"T-1"}, Assets: map[string]*domain.Asset{}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	collector := mocks.NewMockCollector(gomock.NewController(t))
	if NewUseAssetAbility(collector, nil, nil, nil).Validate(faction, factionState, book) {
		t.Error("Validate = true, want false: the only ability belongs to a lapsed tag")
	}
}
```

##### Task 10 — this plan

In the Work Breakdown Overview, `- [ ] **Phase 2 — New tag mechanics**` becomes `- [x] **Phase 2 — New tag mechanics**`.

##### Commit message

```
feat(E2): add tag upkeep with lapse, loss and damage

- coin_upkeep on an upkeep trigger: a tag's turn upkeep, charged in
  bookkeeping after income and before asset maintenance; on_unpaid
  lapses the tag, removes it, or damages the faction
- domain: Faction.LapsedTags and the TagLapsed mutation; removing a tag
  clears its lapse
- a lapsed tag's hooks, phase and matchup entries and action are off;
  its upkeep is still charged, and paying it lifts the lapse
- E2 plan: Phase 2 ticked
```

### Phase 3 — The H.A.W.K rulebook

**In plain words:** `rulebooks/hawk/` comes into being with H.A.W.K's 16 tags, SWN's asset files and settings standing in for the rest, and a test that loads it and builds the engine from it. The tag fiction joins the canon, and the rulebook is composed into the H.A.W.K campaign.

#### Commit 7 — `feat(E2): start rulebooks/hawk with SWN stand-ins and H.A.W.K tags`

##### Task 1 — `rulebooks/hawk/rulebook.toml`

New — full contents:

```toml
# Folders composed into a campaign with this rulebook. Each path resolves from this folder.
includes = ["../_core"]
```

##### Task 2 — the asset stand-ins

From the repo root (byte copies; Decision 14):

```sh
mkdir -p rulebooks/hawk/assets
cp rulebooks/swn/assets/force_assets.toml rulebooks/hawk/assets/force_assets.toml
cp rulebooks/swn/assets/cunning_assets.toml rulebooks/hawk/assets/cunning_assets.toml
cp rulebooks/swn/assets/wealth_assets.toml rulebooks/hawk/assets/wealth_assets.toml
```

Then in `rulebooks/hawk/assets/force_assets.toml`, Gengineered Slaves drops its tag requirement (discovery D-11):

**Find:**
```toml
description = "Genetically engineered servitors bred by a eugenics cult for labor and war. Only a faction with the Eugenics Cult tag can buy them."
requires_tag = "SWN-T-003"
```
**Replace with:**
```toml
description = "Genetically engineered servitors bred by a eugenics cult for labor and war. Only a faction with the Eugenics Cult tag can buy them."
```

##### Task 3 — `rulebooks/hawk/settings.toml`

New — full contents:

```toml
# Setting rules for this rulebook. drift_costs and asset_types are required; every
# other key is optional, and a key left out turns its rule off.
#
# A stand-in: SWN's values, plus Aethership for the H.A.W.K tags that name it.
# A.001.5 Effort 3 writes H.A.W.K's own values.

# drift_costs[i] = hex-equivalent crossing cost for drift rating (i+1)
drift_costs = [5, 4, 3, 2, 1]

# Every asset type this rulebook's data may name: an asset's type, a transport's
# cargo, a relocate's filter, an effects filter and the floors below.
asset_types = ["Military Unit", "Special Forces", "Facility", "Starship", "Tactic", "Logistics Facility", "Special", "Aethership"]

# The least world population on which an asset of each type may be bought. A
# type left out has no floor. Quote a type with a space: "Military Unit".
[purchase_population]
Starship = 500000
```

##### Task 4 — `rulebooks/hawk/tags.toml`

New — full contents:

```toml
[tags.Proprietor]
id = "HAWK-T-001"
name = "Proprietor"
government = true
description = "This faction holds the deed to a place in the Core, a legal fiction nobody believes and everybody enforces."
effect = "Held while the faction governs at least one place; the campaign's government record lists each place it governs. On those places, the faction grants or refuses permission to buy or bring in assets marked as needing government permission."

[tags.WeaveAnchors]
id = "HAWK-T-002"
name = "Weave Anchors"
description = "This faction has pinned a strand of the Weave to its seat, the one place in the Astral Sea where its casters are at full strength."
effect = "Roll an extra d10 on every roll made on its seat. Its seat counts as tech level 5. Upkeep: 1 Coin each turn; unpaid, the tag is lost. Also lost if the faction's seat changes."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "homeworld" },
    bonus_die = { sides = 10, count = 1 },
  },
  {
    trigger = "tech_level_query",
    effect = "tech_level_floor",
    filters = { owner = "own", scope = "homeworld" },
    tech_level_floor = { level = 5 },
  },
  {
    trigger = "upkeep",
    effect = "coin_upkeep",
    filters = { owner = "own", scope = "faction" },
    coin_upkeep = { amount = 1, on_unpaid = "lose_tag" },
  },
  { trigger = "homeworld_changed", effect = "remove_tag", filters = { owner = "own", scope = "faction" } },
]

[tags.AetherBurners]
id = "HAWK-T-003"
name = "Aether Burners"
description = "This faction runs everything on raw, barely refined Aether: fast, loud, and liable to burn out at the worst moment."
effect = "Once per turn, roll an extra d10 on a Force attack. Every asset can move itself one hex as an action. Every die that comes up 10 must be rerolled."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "faction", phase = ["attack"], attribute = "Force" },
    gates = { once_per = "owner_faction" },
    bonus_die = { sides = 10, count = 1 },
  },
  {
    trigger = "speed_query",
    effect = "speed_floor",
    filters = { owner = "own", scope = "faction" },
    speed_floor = { speed = 1 },
  },
  {
    trigger = "own_roll",
    effect = "reroll",
    filters = { owner = "own", scope = "faction" },
    reroll = { result = 10, elective = false },
  },
]

[tags.SoulBonded]
id = "HAWK-T-004"
name = "Soul-Bonded"
description = "This faction's people are bound to an Astral God by Soul Bonds: the god's power anywhere, and the god slowly eats them for it."
effect = "Once per turn, roll an extra d10 when attacking or defending. Upkeep: 1 Coin each turn; unpaid, the faction takes 2 damage."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "faction", phase = ["attack", "defense"] },
    gates = { once_per = "owner_faction" },
    bonus_die = { sides = 10, count = 1 },
  },
  {
    trigger = "upkeep",
    effect = "coin_upkeep",
    filters = { owner = "own", scope = "faction" },
    coin_upkeep = { amount = 1, on_unpaid = "damage", damage = 2 },
  },
]

[tags.ContractBound]
id = "HAWK-T-005"
name = "Contract-Bound"
description = "This faction lives by bargains sealed with a patron, and always knows which clause matters."
effect = "Once per turn, force a rival rolling on a place where this faction has a Base of Influence to reroll one die. Action: roll 1d6; that many Coin come off its next asset purchase, down to half price."
ability = { effect = "purchase_discount", die = "1d6" }
effects = [
  {
    trigger = "rival_roll",
    effect = "reroll",
    filters = { owner = "rival", scope = "base" },
    gates = { once_per = "owner_faction" },
    reroll = { dice = 1, elective = true },
  },
]

[tags.AetherRefiners]
id = "HAWK-T-006"
name = "Aether Refiners"
description = "This faction turns raw Aether into Aetherstones and Aether-Cores, the fuel of everything in the Astral Sea."
effect = "At the start of each turn, gain 1 Coin. Every die that comes up 10 must be rerolled. Lost if the faction's seat changes."
effects = [
  {
    trigger = "turn_start",
    effect = "coin_gain",
    filters = { owner = "own", scope = "faction" },
    coin_gain = { amount = "1" },
  },
  {
    trigger = "own_roll",
    effect = "reroll",
    filters = { owner = "own", scope = "faction" },
    reroll = { result = 10, elective = false },
  },
  { trigger = "homeworld_changed", effect = "remove_tag", filters = { owner = "own", scope = "faction" } },
]

[tags.WarProfiteers]
id = "HAWK-T-007"
name = "War Profiteers"
description = "This faction profits from everyone's war, and fights its own as little as it can."
effect = "Once per turn, gain 1 Coin when any rival's asset is destroyed, by anyone. Once per turn, roll an extra d10 when defending against a Wealth attack. Always loses ties in an attack."
effects = [
  {
    trigger = "asset_destroyed",
    effect = "coin_gain",
    filters = { owner = "rival", scope = "faction" },
    gates = { once_per = "owner_faction" },
    coin_gain = { amount = "1" },
  },
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "faction", phase = ["defense"], opponent_attribute = "Wealth" },
    gates = { once_per = "owner_faction" },
    bonus_die = { sides = 10, count = 1 },
  },
  { trigger = "attack_tie", effect = "lose_ties", filters = { owner = "own", scope = "faction" } },
]

[tags.ShellCompany]
id = "HAWK-T-008"
name = "Shell Company"
description = "Nothing this faction owns is on paper, and everything it loses is insured."
effect = "Every asset it buys starts Stealthed. Gain 1 Coin whenever one of its own assets is destroyed."
effects = [
  { trigger = "on_purchase", effect = "apply_stealth", filters = { owner = "own", scope = "faction" } },
  {
    trigger = "asset_destroyed",
    effect = "coin_gain",
    filters = { owner = "own", scope = "faction" },
    coin_gain = { amount = "1" },
  },
]

[tags.Outsourced]
id = "HAWK-T-009"
name = "Outsourced"
description = "This faction hires its problems out to H.A.W.K through Questing as a Service, and pays a subscription for the privilege."
effect = "Once per turn, reroll one of its own lowest dice. When it completes any goal, roll 1d6; on 5+, gain 1 bonus XP. Upkeep: 1 Coin each turn; unpaid, the tag lapses until paid."
effects = [
  {
    trigger = "own_roll",
    effect = "reroll",
    filters = { owner = "own", scope = "faction" },
    gates = { once_per = "owner_faction" },
    reroll = { dice = 1, elective = true },
  },
  {
    trigger = "goal_completed",
    effect = "xp_roll",
    filters = { owner = "own", scope = "faction" },
    xp_roll = { die = "1d6", target = 5, amount = 1 },
  },
  {
    trigger = "upkeep",
    effect = "coin_upkeep",
    filters = { owner = "own", scope = "faction" },
    coin_upkeep = { amount = 1, on_unpaid = "lapse" },
  },
]

[tags.Monopolists]
id = "HAWK-T-010"
name = "Monopolists"
description = "This faction owns the docks, the fuel and the roads, and everyone else pays to use them."
effect = "Any rival asset that moves onto a place where this faction has a Base of Influence pays it 1 Coin. Assets that need tech level 4 or more cost it 1 less."
effects = [
  {
    trigger = "arrival_query",
    effect = "arrival_toll",
    filters = { owner = "rival", scope = "base" },
    arrival_toll = { amount = 1 },
  },
  {
    trigger = "asset_cost_query",
    effect = "cost_discount",
    filters = { owner = "own", scope = "faction", min_tech_level = 4 },
    cost_discount = { amount = 1 },
  },
]

[tags.AetherTechEvangelists]
id = "HAWK-T-011"
name = "Aether-Tech Evangelists"
description = "This faction believes Aether-tech lifts every Fragment it touches, and will not stop explaining this."
effect = "Every place with this faction's Base of Influence counts as at least tech level 4. Action: pay 2 Coin and roll 1d12; on a 12, the place's tech level becomes 4 for this faction."
ability = {
  effect = "raise_tech_level",
  target = "world",
  target_roll = { die = "1d12", target = 12 },
  raise_tech_level = { coin_cost = 2, tech_level = 4 },
}
effects = [
  {
    trigger = "tech_level_query",
    effect = "tech_level_floor",
    filters = { owner = "own", scope = "base" },
    tech_level_floor = { level = 4 },
  },
]

[tags.DriftRunners]
id = "HAWK-T-012"
name = "Drift-Runners"
description = "This faction lives in the Drift and reads its currents: smugglers, pilots, nomads."
effect = "Its assets cross the Drift as if their drift rating were 1 higher. It can build Aetherships on any place with at least 1% of the usual population floor."
effects = [
  {
    trigger = "drift_rating_query",
    effect = "drift_rating_bonus",
    filters = { owner = "own", scope = "faction" },
    drift_rating_bonus = { amount = 1 },
  },
  {
    trigger = "population_floor_query",
    effect = "population_floor",
    filters = { owner = "own", scope = "faction", asset_type = ["Aethership"] },
    population_floor = { percent = 1 },
  },
]

[tags.Godslayers]
id = "HAWK-T-013"
name = "Godslayers"
description = "This faction is sworn to destroy the Astral Gods and everyone who feeds them."
effect = "Always rerolls dice that come up 1. Once per turn, roll an extra d10 when attacking a faction that holds Soul-Bonded. Always loses ties in an attack."
effects = [
  {
    trigger = "own_roll",
    effect = "reroll",
    filters = { owner = "own", scope = "faction" },
    reroll = { result = 1, elective = false },
  },
  { trigger = "attack_tie", effect = "lose_ties", filters = { owner = "own", scope = "faction" } },
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "faction", phase = ["attack"], opponent_tag = "HAWK-T-004" },
    gates = { once_per = "owner_faction" },
    bonus_die = { sides = 10, count = 1 },
  },
]

[tags.Insurgents]
id = "HAWK-T-014"
name = "Insurgents"
description = "This faction is rebel cells that fight whoever owns the docks, the fuel and the roads."
effect = "Once per turn, roll an extra d10 when attacking a faction that holds Monopolists. Roll an extra d10 when testing to find Stealthed assets."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "faction", phase = ["attack"], opponent_tag = "HAWK-T-010" },
    gates = { once_per = "owner_faction" },
    bonus_die = { sides = 10, count = 1 },
  },
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "faction", detection = true },
    bonus_die = { sides = 10, count = 1 },
  },
]

[tags.OldFaith]
id = "HAWK-T-015"
name = "Old Faith"
description = "This faction holds to the old gods of mortalkind, the Fey Royals, against the Astral Gods who replaced them."
effect = "Roll an extra d10 on the test to Expand Influence. Once per turn, roll an extra d10 when defending against a faction that holds Soul-Bonded."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "faction", phase = ["faction_test"], action = "Expand Influence" },
    bonus_die = { sides = 10, count = 1 },
  },
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "faction", phase = ["defense"], opponent_tag = "HAWK-T-004" },
    gates = { once_per = "owner_faction" },
    bonus_die = { sides = 10, count = 1 },
  },
]

[tags.InformantNetwork]
id = "HAWK-T-016"
name = "Informant Network"
description = "This faction runs on paid informants: it hears everything, as long as it keeps paying."
effect = "Once per turn, roll an extra d10 on a Cunning attack or when defending against one. Upkeep: 1 Coin each turn; unpaid, the tag lapses until paid."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "faction", phase = ["attack"], attribute = "Cunning" },
    gates = { once_per = "owner_faction" },
    bonus_die = { sides = 10, count = 1 },
  },
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "faction", phase = ["defense"], opponent_attribute = "Cunning" },
    gates = { once_per = "owner_faction" },
    bonus_die = { sides = 10, count = 1 },
  },
  {
    trigger = "upkeep",
    effect = "coin_upkeep",
    filters = { owner = "own", scope = "faction" },
    coin_upkeep = { amount = 1, on_unpaid = "lapse" },
  },
]
```

##### Task 5 — `internal/faction/engine/core_test.go`

**Find:**
```go
// TestNewWithRulebook_ShippedRulebook catches an authored effects entry that
// loads but fails registrar validation, which otherwise surfaces only at runtime.
func TestNewWithRulebook_ShippedRulebook(t *testing.T) {
	camp := &campaigns.Campaign{Root: t.TempDir()}
	if _, err := campaigns.ComposeRulebook(camp, "../../../rulebooks/swn", false); err != nil {
		t.Fatalf("compose shipped rulebook: %v", err)
	}
	rb, err := rulebook.Load(camp.Paths().FactionDataDir)
	if err != nil {
		t.Fatalf("load shipped rulebook: %v", err)
	}
	if _, err := NewWithRulebook(rb, nil, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("construct engine from shipped rulebook: %v", err)
	}
}
```
**Replace with:**
```go
// TestNewWithRulebook_ShippedRulebook catches an authored effects entry that
// loads but fails registrar validation, which otherwise surfaces only at runtime.
// It is H.A.W.K's load check too (A.001.5 Effort 2 Decision 5).
func TestNewWithRulebook_ShippedRulebook(t *testing.T) {
	for _, name := range []string{"swn", "hawk"} {
		t.Run(name, func(t *testing.T) {
			camp := &campaigns.Campaign{Root: t.TempDir()}
			if _, err := campaigns.ComposeRulebook(camp, "../../../rulebooks/"+name, false); err != nil {
				t.Fatalf("compose shipped rulebook: %v", err)
			}
			rb, err := rulebook.Load(camp.Paths().FactionDataDir)
			if err != nil {
				t.Fatalf("load shipped rulebook: %v", err)
			}
			if _, err := NewWithRulebook(rb, nil, slog.New(slog.DiscardHandler)); err != nil {
				t.Fatalf("construct engine from shipped rulebook: %v", err)
			}
		})
	}
}
```

##### Commit message

```
feat(E2): start rulebooks/hawk with SWN stand-ins and H.A.W.K tags

- rulebooks/hawk/: rulebook.toml includes ../_core; SWN copies of the
  three asset files (Gengineered Slaves without requires_tag) and
  settings.toml (plus Aethership) stand in until Efforts 3 to 6
- tags.toml: H.A.W.K's 16 tags, HAWK-T-001 to HAWK-T-016
- the shipped-rulebook engine test runs over swn and hawk
```

#### Commit 8 — `docs(E2): write the tag canon and compose into campaigns/hawk`

##### Task 1 — `campaigns/hawk/canon/tags.md`

Full contents:

```markdown
# H.A.W.K Tags

The fiction behind each faction tag, one entry per `HAWK-T-` ID.

A tag is a faction's nature or a special aptitude, and any faction in the setting could hold one. A faction holds one tag, or two if it is unusually versatile; Proprietor comes and goes with the places a faction governs. The rules for each tag are in `rulebooks/hawk/tags.toml`.

## Proprietor — `HAWK-T-001`

The Core Sector is neutral ground, which in practice means everything in it is for sale. Since Blackwater fell, its places belong to whoever holds the deed, and a Proprietor holds one. The title is a legal fiction that nobody in the Core believes and everybody enforces.

## Weave Anchors — `HAWK-T-002`

The Weave exists only on the Material Plane, and only mortals control it. A Weave Anchor has done the impossible on one Fragment: a spire, a circle, a city wired back to Myr, with a strand of the Weave pinned to it. On that ground its casters are terrifying. Off it, they are tourists.

The strand always wants to snap back to Myr. Holding it costs every day, in Aetherstones, in casters and in the occasional casters' funeral. If the anchor fails, or the faction moves its seat, the strand snaps home.

## Aether Burners — `HAWK-T-003`

Refined Aether is for people with patience. Aether Burners pour it in raw, so their ships are fast, their guns are loud and their best moments go up in smoke. The crews call it "running hot". The insurers call it something else.

## Soul-Bonded — `HAWK-T-004`

A Soul Bond gives raw access to a god's realm wherever the believer stands, and the faithful fight with that weight behind them. The god collects every turn, in faith, in Coin and in years of life. The faithful call it worship. Godslayers call it a feeding schedule.

## Contract-Bound — `HAWK-T-005`

Every contract has a clause, and a contract-bound faction always knows which one matters. Near its influence, rivals find their plans quietly voided on a technicality. Elsewhere, it calls in favours: somebody, somewhere, signed something.

## Aether Refiners — `HAWK-T-006`

Everything in the Astral Sea runs on refined Aether, and a refinery sells it by the stone. The work is hot, dangerous and only occasionally explosive. Raw Aether is volatile, and so is everything near it: refinery towns run hot, the work never stops, and neither do the accidents. A refinery can't be moved, so the faction lives where its stills are.

## War Profiteers — `HAWK-T-007`

Somebody sells the guns, the armour, the replacement hulls and the memorial plaques, and every side of the War in Heaven is a customer. War profiteers love a long war and hate a close fight. Nobody's books look better than theirs, and nobody wants to be seen with them.

## Shell Company — `HAWK-T-008`

A shell company's warehouse belongs to a holding company that belongs to a trust that belongs to nobody. When the warehouse burns down, the insurance pays out in full, and the investigators can't find who to blame.

## Outsourced — `HAWK-T-009`

Why keep a security team when a sanctioned party is one request away? An outsourced faction keeps H.A.W.K on retainer through Questing as a Service, and gets professionals at the worst moment for its rivals. The subscription renews automatically. Collateral damage is billed separately. Miss a payment and the service is suspended until the account is settled.

## Monopolists — `HAWK-T-010`

Every dock fee, fuel tax and landing permit in a monopolist's territory goes to the same account. It doesn't need to fight anyone; it only needs them to keep travelling.

## Aether-Tech Evangelists — `HAWK-T-011`

The evangelists believe Aether-tech lifts every Fragment it touches, and they will not stop explaining this. Wherever they set up, the stones are cut, the cores are wired in and the lights come on. The programs are free, because the locals are the product.

## Drift-Runners — `HAWK-T-012`

Drift-runners read the clouds the way other people read a street sign. Their pilots ride currents other crews won't touch, so they cut through murk that turns everyone else back, and they build their next ship from whatever the last one left behind.

## Godslayers — `HAWK-T-013`

Godslayers have sworn that no Astral God survives them, and the faithful are the gods' supply line. They don't retreat, they don't negotiate, and they don't know when to stop, even when they should.

## Insurgents — `HAWK-T-014`

Insurgents live in the gaps between the people in charge: maintenance crews, dock hands, the engineer who wrote the security and kept a back door. Their enemy is whoever owns the docks, the fuel and the roads, and makes everyone else pay to use them.

## Old Faith — `HAWK-T-015`

Old prayers and old bargains still run deep in mortal hearts, and since Diem Veritas they run deeper. Everyone now knows what the Astral Gods do to their faithful. An Old Faith faction stands in their way, offering mortals older gods who at least ask first, and it is very hard to talk out of its convictions.

## Informant Network — `HAWK-T-016`

H.A.W.K stands for Hunts, Acquisitions, Whispers and Knowledge, and the whispers are for sale. Every dock hand, bartender and junior clerk in the Core sells something, and a faction with an informant network buys it all. Its rivals' plans arrive before its rivals do. The network costs money every week, and the week it doesn't pay, the whispers dry up and everyone goes quiet at once.
```

##### Task 2 — compose into `campaigns/hawk`

From the repo root:

```sh
go build -o gm-toolkit ./cmd/gm-toolkit
GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit campaign add-rules --rules rulebooks/hawk --campaign hawk
```

The first compose leaves out `--replace`. `campaigns/hawk/rulebook/` gains `rulebook/assets/` (three files), `goals.toml`, `settings.toml` and `tags.toml`.

##### Task 3 — delete the drafts

Delete `scratch/A.001.5.E2-tag-drafts.md` and `scratch/A.001.5.E2-tag-drafts-pass5.md` once Robert has read the canon from Task 1. Both are gitignored, so the deletion is not part of the commit.

##### Task 4 — the plans

- This plan's Work Breakdown Overview: `- [ ] **Phase 3 — The H.A.W.K rulebook**` becomes `- [x] **Phase 3 — The H.A.W.K rulebook**`.
- The overview plan's Effort Summary: the row `| [ ] | 2 | Tag pass |` becomes `| [x] | 2 | Tag pass |`.

##### Commit message

```
docs(E2): write the tag canon and compose into campaigns/hawk

- canon/tags.md: the fiction of the 16 H.A.W.K tags
- campaigns/hawk/rulebook/: rulebooks/hawk composed in with add-rules
- E2 plan: Phase 3 ticked; overview plan: Effort 2 ticked
```
