# Data-Driven Tags (A.001.7) — Effort 4 Plan: World and Setting Rules

## In Plain Words

- **What:** three tag rules start to work, and so do two setting rules they lean on.
  - **Settings.** Each setting folder gets a `settings.toml`: the drift costs (moved
    from `drift_costs.toml`) and the least world population on which each asset type
    may be bought. SWN says Starships need 500,000 people.
  - **The population rule.** Buy Asset and Pretech Logistics refuse an asset whose
    type has a floor when the world is too small. A faction whose ships come from
    another power has a waiver, set in the faction wizard.
  - **Colonists** governs its homeworld from creation, treats it as at least TL4, and
    needs only 20% of the floor. **Technical Expertise** treats every world with its
    Base as at least TL4 and needs 2%.
  - **Pirates** charges a rival 1 Coin for each asset that lands on a world with a
    Pirates Base. A rival that cannot pay waits at the edge.
- **Why:** these rules do nothing in play today, and the engine has no population
  rule at all, though the book has one.
- **How it fits:** Effort 4 of five in `A.001.7` (arc `A.001`), on branch
  `feature/data-driven-tags`. It builds on Efforts 1 and 2. Effort 5 (the A-side tag
  holder) follows and also edits `CanPurchase`.

## Context / Goal

- Overview: [`A.001.7-data-driven-tags-overview-plan.md`](./A.001.7-data-driven-tags-overview-plan.md).
  Decision 3 (one check in `CanPurchase`) and Decision 7 (values and references in
  the data) are built here. R-6 (this session) reshapes Decision 3's Starship floor
  into a population floor per asset type.
- Discovery: [`A.001.7-data-driven-tags-discovery.md`](../discovery/A.001.7-data-driven-tags-discovery.md),
  D-6, D-8, D-9, D-10.
- Effort 2 added the `homeworld` scope this effort's TL floor reuses.
- Architecture read: persistence, world-movement, hooks.

## Decisions Ratified in Planning

1. **`settings.toml` is one record of every setting rule the engine reads; a
   required key is a plain field the loader checks, an optional key is a pointer
   or a map, nil when left out.** Robert's call (option C; matches the loader's
   optional sub-tables and `filters.tech_level *int`). `drift_costs` is required:
   movement cannot price a route without it. `purchase_population` is optional: a
   type with no entry has no floor, so a setting with no such rule leaves it out.
   Each setting folder ships its own file; `_core` cannot, since two folders that
   supply one file are refused at compose. `refuseUnknownKeys` still catches a
   misspelt key. Rejected: a `required:"true"` struct tag read by reflection (no
   precedent in `rulebook/`, and it cannot tell a missing key from 0).

2. **The population floor is keyed on asset type, in the data.** Robert's call.
   `[purchase_population] Starship = 500000` in `settings.toml`;
   `CanPurchase` compares the world's population to the buyer's floor for the
   definition's type and names no type. The tag effect is `population_floor` with
   `percent` and the existing `asset_type` filter. Amends overview Decision 3
   (overview R-6). Rejected: `starship_floor` with `def.Type == TypeStarship` in Go,
   which holds a rulebook reference in engine code (Decision 3 below).

3. **Overview Decision 7 binds every rule site, not only effect entries.** Robert's
   call. Settings keys, domain checks, loaders and resolvers name no rulebook value
   (a type, tag, goal, number or ID). Engine parts are named by mechanism
   (`population_floor`), not by the book's one instance (Starship). Worked case:
   Decision 2. `R.026` widens to audit these sites too.

4. **The floor resolves per faction and definition; both purchase paths fill the
   terms.** Planning call (overview Decision 3).
   - `PurchaseTerms` gains `Population` (the purchase world's) and
     `PopulationFloor` (the buyer's, for the definition).
   - `dispatch.ResolvePopulationFloor(registry, faction, def, floors)`: 0 for a
     waived faction or a type with no floor; else the setting's floor, chained
     through the owner's `PopulationFloorModifier`s (faction bucket, then own
     assets in ID order), each answering `min(floor, base × percent / 100)`. The
     lowest floor wins: a faction with Colonists and Technical Expertise uses 2%.
   - Buy Asset reads the world's population beside its TL. Pretech Logistics reads
     the acting asset's world. No world engine, or an unknown enabler world,
     imposes no floor (`math.MaxInt`), as for the tech cap.

5. **`population_waiver` waives every floor, set by a wizard toggle.** Robert's call
   (D-9). A `Faction` field, `toml:"population_waiver,omitempty"`. The toggle names
   the book's case: ships provided or supported by another power. Rejected: a
   per-type list (a multi-select for a one-type case); `starship_waiver`.

6. **A new `base` scope: the worlds where the owner has a Base of Influence.**
   Planning call. `Faction.HasBaseOn` reads it. `tech_level_floor` accepts
   `homeworld` (Colonists) and `base` (Technical Expertise); `arrival_toll` accepts
   `base` only. `HasBaseOn` is a third base lookup beside the private
   `factionHasBaseOn` (goals) and `factionBaseOnWorld` (actions); they stay as
   they are.

7. **Colonists' government role is the tag marker `homeworld_government = true`.**
   Planning call (D-8, overview Decision 7). The loader refuses the marker when no
   tag sets `government`, and on a tag that sets `government` too. The wizard: refuses a homeworld-government tag picked
   with the government tag (Robert's call); refuses a homeworld another faction
   governs; adds the government tag's ID; and writes the homeworld as the governed
   world, so `CreateFaction` adds the record entry.

8. **Pirates is `arrival_toll`, a Cat 4 rule asked per landing asset.** Planning
   call (D-6). Owner `rival`, scope `base`, `[arrival_toll] amount`.
   - `dispatch.ArrivalTolls(registry, factions, mover, world)` walks every faction
     but the mover in ID order, its faction bucket then its own assets' in ID
     order, and returns one toll per rule that charges. Two Pirates factions on one
     world charge 1 Coin each (Robert's call).
   - `dispatch.TollCharges` pays each toll as a loss and a gain, both caused by the
     holder, as `coin_steal` does. A Bank refunds the loss and the holder keeps the
     gain (Robert's call).

9. **Who skips the toll and who pays it.** Robert's calls (rules readings).
   - A stealthed asset pays no toll. A carrier with `ignores_lockdown` (Deep Strike
     Landers) covers its whole order, cargo included, as it does for the lockdown.
     Blockade Runners pay.
   - Each asset aboard pays for itself, from its current owner.
   - `dispatch.AssetArrivalTolls` holds the stealth and owner rules once, for
     drift arrivals and relocates alike.

10. **A drift arrival checks the toll after the lockdown and before the
    government, and holds unpaid until it can pay.** Robert's call (the hold);
    planning call (the order). Lockdown, then tolls, then permission, then land. An
    order whose payers cannot all pay holds at the edge (`HoldTollUnpaid`) and tries
    again at its next tick, as an unpaid lockdown does. The toll comes before the
    government so that a faction that cannot pay is never asked. `checkArrivals`
    tracks the Coin spent per faction, since a rider's owner may differ.

11. **A relocate drops the worlds it cannot pay for before the prompt.** Robert's
    call (D-6). After the existing cost check, a world whose tolls plus one asset's
    `CoinCost` exceed the faction's Coin leaves the destinations
    (`ErrInsufficientCoin` when none is left). The total check after the picks adds
    each order's tolls; each order that moves pays its own. The pre-filter does not
    read stealth, so a short purse can lose a world a stealthed asset would reach
    free (planning call: one number per world). An asset picked for the world it
    already stands on owes no toll.

12. **The wizard's starting-asset picker stays on the raw world TL.** Robert's call.
    It has no hooks registry, so Colonists' TL4 homeworld and the population floor
    do not reach starting picks. A Backlog row records the gap at pre-merge.

13. **Three phases, one feature each, past the 800-line phase budget on purpose.**
    Robert's call. Phase 1: settings, the population rule and the waiver. Phase 2:
    the effect and scopes Colonists and Technical Expertise need, then their data.
    Phase 3: the toll rule, then the arrivals and relocates that charge it. The
    phases run about 900 to 1,020 lines. This effort is the trial of a proposed
    ceiling of about 1,500 lines and 3 to 4 commits per phase; watch for a session
    that compacts mid-phase or debugs past about 20 minutes. Rejected: five phases
    at commit seams (each under 800, but two extra sessions for no cleaner seam).

## Decision Record — Execution

- **Commit 1, Task 5 skipped.** `campaigns/sandbox-claude/` does not exist, so there
  was no rulebook to re-compose. Robert's `campaigns/sandbox/rulebook/` still holds
  `drift_costs.toml` and no `settings.toml`; it will not load until he re-composes
  it (D-4).
- **Commit 2, a test fixture the plan missed.** `TestPurchaseEnabler_PaysRivalSurcharge`
  built a `&world.WorldEngine{Index: ...}` literal with no map; Task 10's population
  read calls `Location` and panicked on the nil map. The fixture now builds the
  engine with `world.NewWithMap(&relocRouter{}, ...)` and sets the same index. Every
  production constructor sets a map, so the engine is unchanged. The enabler
  rulebook has no floor, so the test checks what it did before.
- **Commit 5, Task 7 scenario renumbered N+8.** `tags_test.go` already had a
  scenario N+7 (Exchange Consulate), so the new population-floor scenario is N+8.
  No `campaigns/sandbox-claude/` exists, so there was no rulebook to re-compose
  after the data change.
- **Commit 7, no re-compose.** Still no `campaigns/sandbox-claude/`, so the Pirates
  data reached no campaign rulebook.
- **Commit 7, a refused relocate order still counts its toll in the total check.**
  The check after the picks adds every order's tolls before the government is
  asked, so a faction can be refused with `ErrInsufficientCoin` for an order the
  government would then refuse anyway. `totalCost` already counts refused orders
  the same way, so the toll follows it. Left as is.
- **Effort close, the review fixes.** The interim review's picks changed these
  plan points:
  - **A world's population is required in the map source.** The spatial builder
    refuses a `[[worlds]]` entry with no `population`. A missing key read as 0
    and silently blocked every floored purchase on the world. TL keeps its old
    gap; out of scope.
  - **The toll rules live once, in `dispatch.AssetArrivalTolls`.** Drift
    arrivals and relocates each held a copy of "stealth pays nothing, the owner
    pays". Relocate now works out each order's tolls once, for the check and the
    charge. Decision 9 updated.
  - **An other-scope relocate to the asset's own world owes no toll.** The check
    sits in `relocate.orderTolls`, not the shared helper: only a relocate can
    pick an asset's own world, and a held drift order sits at the edge with no
    world. Decision 11 updated.
  - **The loader refuses one tag that sets both `government` and
    `homeworld_government`.** The wizard could never accept such a tag. Decision 7
    updated.
  - **The unpaid-toll wire names no tag.** "the pirates' toll" became "the toll
    to land", since any rulebook entry may charge one (Decision 3).
- **Effort close, review findings left alone.** The tracked sandbox's missing
  `settings.toml` stays Robert's (D-4). The private base lookup in `goals` stays,
  per Decision 6.
- **Initiative close: Refit checks the world's purchase rules through
  `CanPurchase`.** Robert's call, from the initiative review. Refit skipped the
  population floor, so it could put a Starship on a world too small to buy one.
  The book has a refit meet the new asset's purchase requirements, so
  `validRefitReplacements` now builds `PurchaseTerms` for the asset's world (the
  cost delta, the resolved TL, the population and the resolved floor) and asks
  `CanPurchase`, as Buy Asset does. That also closes the older gap of no TL check
  on a refit. `RefitAsset` takes the world engine. Test:
  `TestValidRefitReplacements_ChecksTheWorldsPurchaseRules`.
- **Initiative close: the wizard refuses Colonists on a governed homeworld in
  either order.** From the initiative review. The check sat on the homeworld
  select only, so a GM who went back and added Colonists after the pick learned of
  it at `CreateFaction`. The tag picker now asks the same helper
  (`homeworldGovernedElsewhere`) against the picked homeworld. Decision 7 is
  unchanged. Test: `TestHomeworldGovernedElsewhere`.

---

## Shared Context

### Branch, cadence, and checks

- Branch `feature/data-driven-tags`. Commit per session. At the effort boundary:
  `/gestalt:effort-close A.001.7.E4`.
- Every commit builds and passes: `go build ./... && go vet ./... && go test ./...`
  from the repo root. Run `gofmt -w` on each touched Go file; alignment in the
  tasks is as gofmt leaves it.
- Edit files only with Edit/Write, never `sed` or scripts. Commit 1 deletes
  `rulebooks/swn/drift_costs.toml` with `git rm`.
- No saved-state migration: `population_waiver` is `omitempty`, so old saves load.
  Robert's `campaigns/sandbox/rulebook/` needs a `settings.toml` (Commit 1) to load,
  and the new tag entries to use them (D-4). Claude does not touch it. If
  `campaigns/sandbox-claude/` exists, Claude re-composes its rulebook with
  `--replace` after Commit 1 and after each data commit.

### Pre-merge additions (beyond the overview's)

- **Graduate to Key Decisions (persistence):** `settings.toml`, one record of every
  setting rule, required plain fields and optional pointers or maps (Decision 1);
  the population floor keyed on asset type (Decision 2). The page's "drift costs
  load from `drift_costs.toml`" lines change.
- **Graduate to Key Decisions (effect-mutation):** overview Decision 7, worded for
  all engine code, not effects only (Decision 3).
- **Graduate to Key Decisions (hooks):** `ResolvePopulationFloor` and
  `ArrivalTolls`, the second a walk of every rival faction's buckets (Decisions 4,
  8).
- **Graduate to Key Decisions (world-movement):** the toll at arrival, its order
  and its bypasses; the relocate's toll filter (Decisions 9 to 11).
- **`planned-work.md`:** a Backlog row for the starting-asset picker gap (Decision
  12).

## Out of Scope

- The A-side tag holder, Teach Planetary Population, Gengineered Slaves and
  `requires_tag` (Effort 5).
- The starting-asset picker reading tag floors (Decision 12).
- Folding `factionHasBaseOn` and `factionBaseOnWorld` into `Faction.HasBaseOn`.
- The `lock_out` fee in Go (`R.026`).

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — Settings and the population rule**
  - Commit 1 — `feat(E4): settings.toml holds the drift costs and purchase populations` (5 tasks)
  - Commit 2 — `feat(E4): an asset type's population floor gates every purchase` (11 tasks)
  - Commit 3 — `feat(E4): the population waiver in the faction wizard` (1 task)
- [x] **Phase 2 — Colonists and Technical Expertise**
  - Commit 4 — `feat(E4): a population_floor effect and two TL-floor scopes` (9 tasks)
  - Commit 5 — `feat(E4): Colonists and Technical Expertise as rulebook data` (7 tasks)
- [x] **Phase 3 — Pirates**
  - Commit 6 — `feat(E4): an arrival toll rule hook` (9 tasks)
  - Commit 7 — `feat(E4): arrivals and relocates pay the toll` (8 tasks)

### Phase 1 — Settings and the population rule

**In plain words:** each setting folder states its drift costs and the population
each asset type needs, and both purchase paths refuse an asset the world is too
small for, unless the faction has a waiver, set in the faction wizard. No tag
lowers the floor yet.

#### Commit 1 — `feat(E4): settings.toml holds the drift costs and purchase populations`

##### Task 1 — `rulebooks/swn/settings.toml` (new)

Full contents:
```toml
# Setting rules for this rulebook. drift_costs is required; every other key is
# optional, and a key left out turns its rule off.

# drift_costs[i] = hex-equivalent crossing cost for drift rating (i+1)
drift_costs = [5, 4, 3, 2, 1]

# The least world population on which an asset of each type may be bought. A
# type left out has no floor. Quote a type with a space: "Military Unit".
[purchase_population]
Starship = 500000
```

##### Task 2 — `rulebooks/swn/drift_costs.toml`

Delete the file: `git rm rulebooks/swn/drift_costs.toml`. Its value moved to
`settings.toml` (D-10).

##### Task 3 — `internal/faction/rulebook/rulebook.go`

The `Rulebook` gains the population floors.

**Find:**
```go
type Rulebook struct {
	Assets          map[string]*domain.AssetDefinition
	Tags            map[string]*domain.Tag
	GovernmentTagID string // the tag with government = true, or "" when there is none
	Goals           map[string]*domain.Goal
	DriftCosts      []int
}
```
**Replace with:**
```go
type Rulebook struct {
	Assets          map[string]*domain.AssetDefinition
	Tags            map[string]*domain.Tag
	GovernmentTagID string // the tag with government = true, or "" when there is none
	Goals           map[string]*domain.Goal
	DriftCosts      []int
	// PurchasePopulation is the least world population on which an asset of each
	// type may be bought; a type with no entry has no floor (settings.toml).
	PurchasePopulation map[domain.AssetType]int
}
```

`Load` reads `settings.toml`.

**Find:**
```go
	driftCosts, err := loadDriftCosts(filepath.Join(dataDir, "drift_costs.toml"))
	if err != nil {
		return nil, fmt.Errorf("loading drift costs: %w", err)
	}

	return &Rulebook{
		Assets:          assets,
		Tags:            tags,
		GovernmentTagID: governmentTagID,
		Goals:           goals,
		DriftCosts:      driftCosts,
	}, nil
}
```
**Replace with:**
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
}
```

Replace the drift-cost loader with the settings loader.

**Find:**
```go
// ---------------------------------------------------------------------------
// Drift Costs
// ---------------------------------------------------------------------------

type driftCostFile struct {
	DriftCosts []int `toml:"drift_costs"`
}

func loadDriftCosts(path string) ([]int, error) {
	var f driftCostFile
	if _, err := toml.DecodeFile(path, &f); err != nil {
		return nil, err
	}
	return f.DriftCosts, nil
}
```
**Replace with:**
```go
// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

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

##### Task 4 — `internal/faction/rulebook/rulebook_test.go`

Append:
```go
func TestLoadSettings(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
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
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("err = %v, want it to contain %q", err, testCase.want)
			}
		})
	}
}

func TestLoadSettings_OptionalKeyLeftOut(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.toml")
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

func TestLoad_ShippedSettings(t *testing.T) {
	rb := loadShippedRulebook(t)
	if !reflect.DeepEqual(rb.DriftCosts, []int{5, 4, 3, 2, 1}) {
		t.Errorf("drift costs = %v, want [5 4 3 2 1]", rb.DriftCosts)
	}
	want := map[domain.AssetType]int{domain.TypeStarship: 500000}
	if !reflect.DeepEqual(rb.PurchasePopulation, want) {
		t.Errorf("purchase population = %v, want %v", rb.PurchasePopulation, want)
	}
}
```

##### Task 5 — sandbox-claude (not committed)

If `campaigns/sandbox-claude/` exists, re-compose its rulebook from `rulebooks/swn`
with `--replace`, then delete its `rulebook/drift_costs.toml`. Robert's
`campaigns/sandbox` is his to update (D-4).

##### Commit message

```
feat(E4): settings.toml holds the drift costs and purchase populations

- settings.toml replaces drift_costs.toml (D-10): one record of every setting
  rule; drift_costs required, purchase_population optional
- Rulebook.PurchasePopulation maps an asset type to the least world population
  it may be bought on; SWN sets Starship = 500000
- Strict keys; asset type names checked at load
```

#### Commit 2 — `feat(E4): an asset type's population floor gates every purchase`

##### Task 1 — `internal/faction/domain/asset.go`

**Find:**
```go
// PurchaseTerms are the per-path inputs to CanPurchase that differ between the Buy Asset
// action and the purchase_enabler ability: Cost is the price to test affordability against
// (base cost for a buy, surcharged for an enabler); TechCap is the highest required tech
// level a def may carry (the target world's tech level for a buy, the enabler's own cap for
// an enabler). The shared checks themselves live in CanPurchase so the two paths can't drift.
type PurchaseTerms struct {
	Cost    int
	TechCap int
}

// CanPurchase reports whether the faction may buy def under terms: it can afford Cost, meets
// the def's minimum category rating, and the def's required tech level is within TechCap.
// Government permission is asked after the choice, not here (D-35).
func CanPurchase(faction *Faction, def *AssetDefinition, terms PurchaseTerms) bool {
	switch {
	case faction.Coin < terms.Cost:
		return false
	case def.MinRating > faction.StatScore(def.Category):
		return false
	case def.TechLevel > terms.TechCap:
		return false
	}
	return true
}
```
**Replace with:**
```go
// PurchaseTerms are the per-path inputs to CanPurchase that differ between the Buy Asset
// action and the purchase_enabler ability: Cost is the price to test affordability against
// (base cost for a buy, surcharged for an enabler); TechCap is the highest required tech
// level a def may carry (the target world's tech level for a buy, the enabler's own cap for
// an enabler). Population is the purchase world's, and PopulationFloor the least the buyer
// needs for def's type (dispatch.ResolvePopulationFloor). The shared checks themselves live
// in CanPurchase so the two paths can't drift.
type PurchaseTerms struct {
	Cost            int
	TechCap         int
	Population      int
	PopulationFloor int
}

// CanPurchase reports whether the faction may buy def under terms: it can afford Cost, meets
// the def's minimum category rating, the def's required tech level is within TechCap, and
// the world's Population reaches PopulationFloor. Government permission is asked after the
// choice, not here (D-35).
func CanPurchase(faction *Faction, def *AssetDefinition, terms PurchaseTerms) bool {
	switch {
	case faction.Coin < terms.Cost:
		return false
	case def.MinRating > faction.StatScore(def.Category):
		return false
	case def.TechLevel > terms.TechCap:
		return false
	case terms.Population < terms.PopulationFloor:
		return false
	}
	return true
}
```

##### Task 2 — `internal/faction/domain/faction.go`

**Find:**
```go
	HookBudgets             map[string]int    `toml:"hook_budgets,omitempty"`
	PendingPurchaseDiscount int               `toml:"pending_purchase_discount,omitempty"`
}
```
**Replace with:**
```go
	HookBudgets             map[string]int    `toml:"hook_budgets,omitempty"`
	PendingPurchaseDiscount int               `toml:"pending_purchase_discount,omitempty"`
	// PopulationWaiver lifts every purchase-population floor: another power provides
	// or supports the faction's ships (D-9, Effort 4 Decision 5).
	PopulationWaiver bool `toml:"population_waiver,omitempty"`
}
```

##### Task 3 — `internal/faction/domain/asset_test.go`

Append:
```go
func TestCanPurchase_PopulationFloor(t *testing.T) {
	faction := &domain.Faction{ID: "f1", Coin: 10, Wealth: 3}
	def := &domain.AssetDefinition{ID: "ship", Category: domain.StatWealth, Cost: 5, TechLevel: 4, Type: domain.TypeStarship}
	cases := []struct {
		name       string
		population int
		floor      int
		want       bool
	}{
		{"no floor", 0, 0, true},
		{"the world reaches the floor", 500000, 500000, true},
		{"the world is too small", 499999, 500000, false},
	}
	for _, tc := range cases {
		terms := domain.PurchaseTerms{Cost: 5, TechCap: 5, Population: tc.population, PopulationFloor: tc.floor}
		if got := domain.CanPurchase(faction, def, terms); got != tc.want {
			t.Errorf("%s: CanPurchase = %v, want %v", tc.name, got, tc.want)
		}
	}
}
```

##### Task 4 — `internal/faction/engine/hooks/rule_modifier.go`

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
// AssetSpeedModifier adjusts the speed, in hexes a turn, of an asset faction
// owns, read by the movement phase (Cat 4).
type AssetSpeedModifier interface {
	ModifyAssetSpeed(faction *domain.Faction, def *domain.AssetDefinition, speed int) int
}

// PopulationFloorModifier lowers the world population faction needs to buy def
// (Cat 4). baseFloor is the setting's floor for def's type; floor is the chain's
// answer so far.
type PopulationFloorModifier interface {
	ModifyPopulationFloor(faction *domain.Faction, def *domain.AssetDefinition, baseFloor, floor int) int
}
```

##### Task 5 — `internal/faction/engine/hooks/registry.go`

**Find:**
```go
type RegisteredAssetSpeedModifier struct {
	Scope  Scope
	Source string
	Hook   AssetSpeedModifier
}
```
**Replace with:**
```go
type RegisteredAssetSpeedModifier struct {
	Scope  Scope
	Source string
	Hook   AssetSpeedModifier
}

type RegisteredPopulationFloorModifier struct {
	Scope  Scope
	Source string
	Hook   PopulationFloorModifier
}
```

**Find:**
```go
	assetSpeedModifiers      map[Scope][]RegisteredAssetSpeedModifier
	attackEligibilityRules   map[Scope][]RegisteredAttackEligibilityRule
```
**Replace with:**
```go
	assetSpeedModifiers      map[Scope][]RegisteredAssetSpeedModifier
	populationFloorModifiers map[Scope][]RegisteredPopulationFloorModifier
	attackEligibilityRules   map[Scope][]RegisteredAttackEligibilityRule
```

**Find:**
```go
	registry.assetSpeedModifiers = make(map[Scope][]RegisteredAssetSpeedModifier)
```
**Replace with:**
```go
	registry.assetSpeedModifiers = make(map[Scope][]RegisteredAssetSpeedModifier)
	registry.populationFloorModifiers = make(map[Scope][]RegisteredPopulationFloorModifier)
```

**Find:**
```go
	delete(registry.assetSpeedModifiers, scope)
```
**Replace with:**
```go
	delete(registry.assetSpeedModifiers, scope)
	delete(registry.populationFloorModifiers, scope)
```

**Find:**
```go
// RegisterAttackEligibilityRule registers a Cat 4 attack-eligibility hook under the given scope.
```
**Replace with:**
```go
// RegisterPopulationFloorModifier registers a Cat 4 population-floor hook under the given scope.
func (registry *Registry) RegisterPopulationFloorModifier(scope Scope, source string, hook PopulationFloorModifier) {
	registry.populationFloorModifiers[scope] = append(registry.populationFloorModifiers[scope], RegisteredPopulationFloorModifier{Scope: scope, Source: source, Hook: hook})
}

// RegisterAttackEligibilityRule registers a Cat 4 attack-eligibility hook under the given scope.
```

**Find:**
```go
// AttackEligibilityRulesFor returns all Cat 4 attack-eligibility hooks matching the given faction and asset.
```
**Replace with:**
```go
// PopulationFloorModifiersFor returns the global and faction Cat 4 population-floor
// hooks; the asset buckets come from PopulationFloorModifiersForAsset.
func (registry *Registry) PopulationFloorModifiersFor(factionID string) []RegisteredPopulationFloorModifier {
	var result []RegisteredPopulationFloorModifier
	result = append(result, registry.populationFloorModifiers[GlobalScope()]...)
	result = append(result, registry.populationFloorModifiers[FactionScope(factionID)]...)
	return result
}

// PopulationFloorModifiersForAsset returns the Cat 4 population-floor hooks
// registered against one asset instance, without the global or faction buckets.
func (registry *Registry) PopulationFloorModifiersForAsset(factionID, assetInstanceID string) []RegisteredPopulationFloorModifier {
	return registry.populationFloorModifiers[AssetScope(factionID, assetInstanceID)]
}

// AttackEligibilityRulesFor returns all Cat 4 attack-eligibility hooks matching the given faction and asset.
```

##### Task 6 — `internal/faction/engine/hooks/dispatch/rules.go`

Insert after `AssetSpeedResolver`.

**Find:**
```go
// ResolveAttackEligibility reports whether query.Attacker may deal damage to
```
**Replace with:**
```go
// ResolvePopulationFloor returns the least world population on which faction may
// buy def: the setting's floor for def's type, run through faction's
// PopulationFloorModifiers, then those of each asset it owns in asset ID order. A
// floor serves its owner only, so a rival's assets are never asked. A waived
// faction, or a type with no floor in floors, needs none (Effort 4 Decision 4).
func ResolvePopulationFloor(registry *hooks.Registry, faction *domain.Faction, def *domain.AssetDefinition, floors map[domain.AssetType]int) int {
	baseFloor := floors[def.Type]
	if faction.PopulationWaiver || baseFloor == 0 {
		return 0
	}
	if registry == nil {
		return baseFloor
	}
	floor := baseFloor
	for _, registered := range registry.PopulationFloorModifiersFor(faction.ID) {
		floor = registered.Hook.ModifyPopulationFloor(faction, def, baseFloor, floor)
	}
	for _, asset := range domain.SortedAssets(faction) {
		for _, registered := range registry.PopulationFloorModifiersForAsset(faction.ID, asset.ID) {
			floor = registered.Hook.ModifyPopulationFloor(faction, def, baseFloor, floor)
		}
	}
	return floor
}

// ResolveAttackEligibility reports whether query.Attacker may deal damage to
```

##### Task 7 — `internal/faction/engine/hooks/dispatch/rules_test.go`

Append:
```go
// stubPopulationFloor answers percent of the setting's floor, never raising it.
type stubPopulationFloor struct{ percent int }

func (stub *stubPopulationFloor) ModifyPopulationFloor(_ *domain.Faction, _ *domain.AssetDefinition, baseFloor, floor int) int {
	return min(floor, baseFloor*stub.percent/100)
}

func TestResolvePopulationFloor(t *testing.T) {
	floors := map[domain.AssetType]int{domain.TypeStarship: 500000}
	starship := &domain.AssetDefinition{ID: "ship", Type: domain.TypeStarship}
	facility := &domain.AssetDefinition{ID: "plant", Type: domain.TypeFacility}
	asset := &domain.Asset{ID: "a1", OwnerID: "f1"}

	plain := &domain.Faction{ID: "f1"}
	if got := ResolvePopulationFloor(hooks.NewRegistry(), plain, starship, floors); got != 500000 {
		t.Errorf("setting floor = %d, want 500000", got)
	}
	if got := ResolvePopulationFloor(hooks.NewRegistry(), plain, facility, floors); got != 0 {
		t.Errorf("type with no floor = %d, want 0", got)
	}
	waived := &domain.Faction{ID: "f1", PopulationWaiver: true}
	if got := ResolvePopulationFloor(hooks.NewRegistry(), waived, starship, floors); got != 0 {
		t.Errorf("waived = %d, want 0", got)
	}

	registry := hooks.NewRegistry()
	registry.RegisterPopulationFloorModifier(hooks.FactionScope("f1"), "colonists", &stubPopulationFloor{percent: 20})
	registry.RegisterPopulationFloorModifier(hooks.AssetScope("f1", "a1"), "expertise", &stubPopulationFloor{percent: 2})
	registry.RegisterPopulationFloorModifier(hooks.AssetScope("rival", "r1"), "rival", &stubPopulationFloor{percent: 1})
	owner := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
	if got := ResolvePopulationFloor(registry, owner, starship, floors); got != 10000 {
		t.Errorf("two floors = %d, want 10000 (the lowest wins; a rival's is never asked)", got)
	}
}
```

##### Task 8 — `internal/faction/engine/action/actions/buy_asset.go`

`purchasableDefinitions` reads the world's population and the buyer's floor.

**Find:**
```go
// purchasableDefinitions returns the definitions the faction may buy on fragmentID via
// domain.CanPurchase: affordable at the cost Resolve will charge, meeting the minimum
// attribute rating, and within the world's tech level. A quality also needs a target on
// the world (D-22). A P-flagged def is offered everywhere; Inputs asks its permission
// after the choice (D-35). An unknown world imposes no tech ceiling (math.MaxInt).
func purchasableDefinitions(faction *domain.Faction, rulebook *rulebook.Rulebook, fragmentID string, worldEngine *world.WorldEngine, registry *hooks.Registry, effectsEngine *effect.EffectsEngine) []*domain.AssetDefinition {
	worldTechLevel := math.MaxInt
	if worldEngine != nil {
		loc, ok := worldEngine.Location(fragmentID)
		if !ok {
			return nil
		}
		worldTechLevel = loc.TechLevel()
	}
```
**Replace with:**
```go
// purchasableDefinitions returns the definitions the faction may buy on fragmentID via
// domain.CanPurchase: affordable at the cost Resolve will charge, meeting the minimum
// attribute rating, within the world's tech level, and on a world populous enough for the
// def's type. A quality also needs a target on the world (D-22). A P-flagged def is offered
// everywhere; Inputs asks its permission after the choice (D-35). With no world engine there
// is no tech ceiling and no population floor (math.MaxInt).
func purchasableDefinitions(faction *domain.Faction, rulebook *rulebook.Rulebook, fragmentID string, worldEngine *world.WorldEngine, registry *hooks.Registry, effectsEngine *effect.EffectsEngine) []*domain.AssetDefinition {
	worldTechLevel := math.MaxInt
	worldPopulation := math.MaxInt
	if worldEngine != nil {
		loc, ok := worldEngine.Location(fragmentID)
		if !ok {
			return nil
		}
		worldTechLevel = loc.TechLevel()
		worldPopulation = loc.Population()
	}
```

**Find:**
```go
		cost := dispatch.ResolveAssetCost(registry, faction, def, fragmentID, def.Cost, onWorld).Cost
		if domain.CanPurchase(faction, def, domain.PurchaseTerms{Cost: cost, TechCap: techCap}) {
```
**Replace with:**
```go
		cost := dispatch.ResolveAssetCost(registry, faction, def, fragmentID, def.Cost, onWorld).Cost
		terms := domain.PurchaseTerms{
			Cost:            cost,
			TechCap:         techCap,
			Population:      worldPopulation,
			PopulationFloor: dispatch.ResolvePopulationFloor(registry, faction, def, rulebook.PurchasePopulation),
		}
		if domain.CanPurchase(faction, def, terms) {
```

##### Task 9 — `internal/faction/engine/action/actions/buy_asset_test.go`

Append. `testWorldLoc` reports a population of 0.
```go
func TestPurchasableDefinitions_PopulationFloorPerType(t *testing.T) {
	worldHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 0, R: 0}}
	worldEngine := world.NewWithMap(&testHexRouter{worlds: map[string]spatial.RegionHex{"a": worldHex}}, slog.New(slog.DiscardHandler))
	rb := &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"ship":  {ID: "ship", Category: domain.StatWealth, Cost: 1, TechLevel: 4, Type: domain.TypeStarship},
			"plant": {ID: "plant", Category: domain.StatWealth, Cost: 1, TechLevel: 4, Type: domain.TypeFacility},
		},
		PurchasePopulation: map[domain.AssetType]int{domain.TypeStarship: 500000},
	}
	faction := &domain.Faction{ID: "f1", Coin: 5, Homeworld: domain.Location{WorldID: "a", RegionHex: worldHex}, Assets: map[string]*domain.Asset{}}

	defs := purchasableDefinitions(faction, rb, "a", worldEngine, hooks.NewRegistry(), nil)
	if len(defs) != 1 || defs[0].ID != "plant" {
		var ids []string
		for _, def := range defs {
			ids = append(ids, def.ID)
		}
		t.Errorf("purchasable = %v, want [plant] (the empty world is below the Starship floor)", ids)
	}

	faction.PopulationWaiver = true
	if defs := purchasableDefinitions(faction, rb, "a", worldEngine, hooks.NewRegistry(), nil); len(defs) != 2 {
		t.Errorf("purchasable with the waiver = %d defs, want 2", len(defs))
	}
}
```

##### Task 10 — `internal/faction/engine/action/actions/ability/purchase_enabler.go`

Add `"math"` to the imports. `eligiblePurchaseDefs` reads the acting asset's world.

**Find:**
```go
// eligiblePurchaseDefs returns the rulebook definitions an enabler purchase may buy:
// matching category and not a quality (which adds no asset to place), required tech
// level at or below the cap, the faction meeting the
// min rating, and affordable at the surcharged price. A P-flagged def is offered; its
// permission is asked after the pick (D-35). Name-sorted for a deterministic candidate
// order.
func eligiblePurchaseDefs(ctx StepContext, params *domain.PurchaseParams) []*domain.AssetDefinition {
	var eligible []*domain.AssetDefinition
	for _, def := range ctx.Rulebook.Assets {
		if def.Category != params.Category || def.Quality {
			continue
		}
		terms := domain.PurchaseTerms{
			Cost:    enablerQuote(ctx, def, params).Cost,
			TechCap: params.MaxTechLevel,
		}
```
**Replace with:**
```go
// eligiblePurchaseDefs returns the rulebook definitions an enabler purchase may buy:
// matching category and not a quality (which adds no asset to place), required tech
// level at or below the cap, the faction meeting the min rating, affordable at the
// surcharged price, and the acting asset's world populous enough for the def's type.
// A P-flagged def is offered; its permission is asked after the pick (D-35).
// Name-sorted for a deterministic candidate order. No world engine, or an unknown
// world, imposes no population floor (math.MaxInt).
func eligiblePurchaseDefs(ctx StepContext, params *domain.PurchaseParams) []*domain.AssetDefinition {
	population := math.MaxInt
	if ctx.WorldEngine != nil {
		if loc, ok := ctx.WorldEngine.Location(ctx.Asset.Location.WorldID); ok {
			population = loc.Population()
		}
	}
	var eligible []*domain.AssetDefinition
	for _, def := range ctx.Rulebook.Assets {
		if def.Category != params.Category || def.Quality {
			continue
		}
		terms := domain.PurchaseTerms{
			Cost:            enablerQuote(ctx, def, params).Cost,
			TechCap:         params.MaxTechLevel,
			Population:      population,
			PopulationFloor: dispatch.ResolvePopulationFloor(ctx.Registry, ctx.Faction, def, ctx.Rulebook.PurchasePopulation),
		}
```

##### Task 11 — `internal/faction/engine/testharness/harness.go`

The stub worlds get a population above the SWN floor, overridable per world.

**Find:**
```go
type StubSpatialMap struct {
	PathFn func(from, to spatial.RegionHex, crossingCost int) ([]spatial.RegionHex, int, error)
	// Worlds, when set, are the worlds AllWorlds returns. Every stub location shares the
	// zero hex, so each one is at distance 0 from every other and in range of any relocate.
	Worlds []string
}

var _ world.HexRouter = (*StubSpatialMap)(nil)

func (s *StubSpatialMap) Location(id string) (spatial.Location, bool) {
	return &stubLocation{id: id}, true
}
```
**Replace with:**
```go
type StubSpatialMap struct {
	PathFn func(from, to spatial.RegionHex, crossingCost int) ([]spatial.RegionHex, int, error)
	// Worlds, when set, are the worlds AllWorlds returns. Every stub location shares the
	// zero hex, so each one is at distance 0 from every other and in range of any relocate.
	Worlds []string
	// Populations overrides a world's population; any other world has stubPopulation.
	Populations map[string]int
}

// stubPopulation is above the SWN Starship floor, so no purchase is refused for its
// world's size unless a test sets Populations.
const stubPopulation = 1_000_000

var _ world.HexRouter = (*StubSpatialMap)(nil)

func (s *StubSpatialMap) Location(id string) (spatial.Location, bool) {
	return &stubLocation{id: id, population: s.populationOf(id)}, true
}

func (s *StubSpatialMap) populationOf(id string) int {
	if population, ok := s.Populations[id]; ok {
		return population
	}
	return stubPopulation
}
```

**Find:**
```go
	for _, id := range s.Worlds {
		locations = append(locations, &stubLocation{id: id})
	}
	return locations
}

type stubLocation struct{ id string }
```
**Replace with:**
```go
	for _, id := range s.Worlds {
		locations = append(locations, &stubLocation{id: id, population: s.populationOf(id)})
	}
	return locations
}

type stubLocation struct {
	id         string
	population int
}
```

**Find:**
```go
func (l *stubLocation) Population() int              { return 0 }
```
**Replace with:**
```go
func (l *stubLocation) Population() int              { return l.population }
```

##### Commit message

```
feat(E4): an asset type's population floor gates every purchase

- PurchaseTerms gains Population and PopulationFloor; CanPurchase refuses a
  world below the floor, naming no asset type (Effort 4 Decision 2)
- Cat 4 PopulationFloorModifier and dispatch.ResolvePopulationFloor: the
  setting's floor for the type, the owner's hooks chained, the lowest wins
- Faction.PopulationWaiver lifts every floor
- Buy Asset reads the world's population; Pretech Logistics the acting
  asset's world
- The harness stub world has 1,000,000 people unless a test sets less
```

#### Commit 3 — `feat(E4): the population waiver in the faction wizard`

##### Task 1 — `internal/faction/tui/views/manage/wizard/wizard.go`

`formData` gains the waiver.

**Find:**
```go
type formData struct {
	name            string
	scale           domain.FactionScale
	primaryStat     domain.FactionStat
	secondaryStat   domain.FactionStat
	tertiaryStat    domain.FactionStat
	selectedTags    []*domain.Tag
	selectedGoal    *domain.Goal
	homeworldID     string
	governedWorldID string
	primaryAssets   []*domain.AssetDefinition
	secondaryAssets []*domain.AssetDefinition
	coinStr         string
}
```
**Replace with:**
```go
type formData struct {
	name             string
	scale            domain.FactionScale
	primaryStat      domain.FactionStat
	secondaryStat    domain.FactionStat
	tertiaryStat     domain.FactionStat
	selectedTags     []*domain.Tag
	selectedGoal     *domain.Goal
	homeworldID      string
	governedWorldID  string
	populationWaiver bool
	primaryAssets    []*domain.AssetDefinition
	secondaryAssets  []*domain.AssetDefinition
	coinStr          string
}
```

`synthesize` sets the flag on the new faction.

**Find:**
```go
	return domain.NewFaction(
		id, m.data.name,
		m.data.scale,
		m.data.primaryStat, m.data.secondaryStat, m.data.tertiaryStat,
		tagIDs(m.data.selectedTags),
		m.data.selectedGoal,
		homeworld,
		assets,
		coin,
	)
}
```
**Replace with:**
```go
	faction := domain.NewFaction(
		id, m.data.name,
		m.data.scale,
		m.data.primaryStat, m.data.secondaryStat, m.data.tertiaryStat,
		tagIDs(m.data.selectedTags),
		m.data.selectedGoal,
		homeworld,
		assets,
		coin,
	)
	faction.PopulationWaiver = m.data.populationWaiver
	return faction
}
```

The toggle goes after the governed world.

**Find:**
```go
	).WithHideFunc(func() bool { return !pickedGovernmentTag(data.selectedTags) })

	coinGroup := huh.NewGroup(
```
**Replace with:**
```go
	).WithHideFunc(func() bool { return !pickedGovernmentTag(data.selectedTags) })

	waiverGroup := huh.NewGroup(
		huh.NewConfirm().
			Title("Population waiver").
			Description("Another power provides or supports this faction's ships, so it may buy assets on a world of any size").
			Value(&data.populationWaiver),
	)

	coinGroup := huh.NewGroup(
```

**Find:**
```go
		homeworldGroup,
		governmentGroup,
		coinGroup,
	).WithTheme(styles.FormTheme())
```
**Replace with:**
```go
		homeworldGroup,
		governmentGroup,
		waiverGroup,
		coinGroup,
	).WithTheme(styles.FormTheme())
```

##### Commit message

```
feat(E4): the population waiver in the faction wizard

- A Confirm after the governed world sets Faction.PopulationWaiver: another
  power provides or supports the faction's ships (D-9)
```

### Phase 2 — Colonists and Technical Expertise

**In plain words:** a tag can lower its holder's population floor and raise the
tech level of its homeworld or of every world with its Base, and Colonists and
Technical Expertise use both. A Colonists faction governs its homeworld from
creation.

#### Commit 4 — `feat(E4): a population_floor effect and two TL-floor scopes`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialXPRoll             SpecialEffectType = "xp_roll"
)
```
**Replace with:**
```go
	SpecialXPRoll             SpecialEffectType = "xp_roll"
	SpecialPopulationFloor    SpecialEffectType = "population_floor"
)
```

**Find:**
```go
	TriggerGoalCompleted          EffectTrigger = "goal_completed"
)
```
**Replace with:**
```go
	TriggerGoalCompleted          EffectTrigger = "goal_completed"
	TriggerPopulationFloorQuery   EffectTrigger = "population_floor_query"
)
```

**Find:**
```go
	ScopeHomeworld EffectScope = "homeworld"
	ScopeFaction   EffectScope = "faction"
)
```
**Replace with:**
```go
	ScopeHomeworld EffectScope = "homeworld"
	ScopeFaction   EffectScope = "faction"
	// ScopeBase is every world where the owner has a Base of Influence (Effort 4
	// Decision 6).
	ScopeBase EffectScope = "base"
)
```

**Find:**
```go
// XPRollParams is a bonus-XP roll: roll Die, and on Target or more gain Amount XP.
```
**Replace with:**
```go
// PopulationFloorParams is a holder's floor as a percent of the setting's.
type PopulationFloorParams struct {
	Percent int
}

// XPRollParams is a bonus-XP roll: roll Die, and on Target or more gain Amount XP.
```

**Find:**
```go
	Trigger        EffectTrigger
	Effect         SpecialEffectType
	Filters        EffectFilters
	Gates          EffectGates
	CoinGain       *CoinGainParams
	CoinSteal      *CoinStealParams
	SelfDamage     *SelfDamageParams
	BonusDie       *BonusDieParams
	Reroll         *RerollParams
	TechLevelFloor *TechLevelFloorParams
	CostSurcharge  *CostSurchargeParams
	RestrictAttack *RestrictAttackParams
	GrantMovement  *GrantMovementParams
	CaptureTarget  *CaptureTargetParams
	DiscountRepair *DiscountRepairParams
	CostDiscount   *CostDiscountParams
	SpeedFloor     *SpeedFloorParams
	XPRoll         *XPRollParams
}
```
**Replace with:**
```go
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
	XPRoll          *XPRollParams
	PopulationFloor *PopulationFloorParams
}
```

##### Task 2 — `internal/faction/domain/faction.go`

**Find:**
```go
// HasTag reports whether the faction holds the tag with ID tagID.
func (faction *Faction) HasTag(tagID string) bool {
	return slices.Contains(faction.Tags, tagID)
}
```
**Replace with:**
```go
// HasTag reports whether the faction holds the tag with ID tagID.
func (faction *Faction) HasTag(tagID string) bool {
	return slices.Contains(faction.Tags, tagID)
}

// HasBaseOn reports whether the faction has a Base of Influence on worldID.
func (faction *Faction) HasBaseOn(worldID string) bool {
	return slices.ContainsFunc(faction.Bases, func(base *Base) bool { return base.Location.WorldID == worldID })
}
```

##### Task 3 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	Trigger        string                `toml:"trigger"`
	Effect         string                `toml:"effect"`
	Filters        *effectFiltersRecord  `toml:"filters"`
	Gates          *effectGatesRecord    `toml:"gates"`
	CoinGain       *amountRecord         `toml:"coin_gain"`
	CoinSteal      *amountRecord         `toml:"coin_steal"`
	SelfDamage     *amountRecord         `toml:"self_damage"`
	BonusDie       *bonusDieRecord       `toml:"bonus_die"`
	Reroll         *rerollRecord         `toml:"reroll"`
	TechLevelFloor *techLevelFloorRecord `toml:"tech_level_floor"`
	CostSurcharge  *costSurchargeRecord  `toml:"cost_surcharge"`
	RestrictAttack *restrictAttackRecord `toml:"restrict_attack"`
	GrantMovement  *grantMovementRecord  `toml:"grant_movement"`
	CaptureTarget  *amountRecord         `toml:"capture_target"`
	DiscountRepair *discountRepairRecord `toml:"discount_repair"`
	CostDiscount   *costDiscountRecord   `toml:"cost_discount"`
	SpeedFloor     *speedFloorRecord     `toml:"speed_floor"`
	XPRoll         *xpRollRecord         `toml:"xp_roll"`
}
```
**Replace with:**
```go
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
	XPRoll          *xpRollRecord          `toml:"xp_roll"`
	PopulationFloor *populationFloorRecord `toml:"population_floor"`
}
```

**Find:**
```go
type xpRollRecord struct {
```
**Replace with:**
```go
type populationFloorRecord struct {
	Percent int `toml:"percent"`
}

type xpRollRecord struct {
```

**Find:**
```go
		domain.TriggerHomeworldChanged, domain.TriggerGoalCompleted,
	}
```
**Replace with:**
```go
		domain.TriggerHomeworldChanged, domain.TriggerGoalCompleted, domain.TriggerPopulationFloorQuery,
	}
```

**Find:**
```go
		domain.SpecialRemoveTag, domain.SpecialXPRoll,
	}
	effectOwners     = []domain.EffectOwner{domain.OwnerOwn, domain.OwnerRival, domain.OwnerAny}
	effectScopes     = []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeHomeworld, domain.ScopeFaction}
```
**Replace with:**
```go
		domain.SpecialRemoveTag, domain.SpecialXPRoll, domain.SpecialPopulationFloor,
	}
	effectOwners     = []domain.EffectOwner{domain.OwnerOwn, domain.OwnerRival, domain.OwnerAny}
	effectScopes     = []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeHomeworld, domain.ScopeFaction, domain.ScopeBase}
```

**Find:**
```go
	if record.XPRoll != nil {
		present = append(present, domain.SpecialXPRoll)
	}
	return present
}
```
**Replace with:**
```go
	if record.XPRoll != nil {
		present = append(present, domain.SpecialXPRoll)
	}
	if record.PopulationFloor != nil {
		present = append(present, domain.SpecialPopulationFloor)
	}
	return present
}
```

**Find:**
```go
		definition.XPRoll = &domain.XPRollParams{Die: *die, Target: record.XPRoll.Target, Amount: record.XPRoll.Amount}
	}
	return nil
}
```
**Replace with:**
```go
		definition.XPRoll = &domain.XPRollParams{Die: *die, Target: record.XPRoll.Target, Amount: record.XPRoll.Amount}
	case domain.SpecialPopulationFloor:
		if record.PopulationFloor == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.PopulationFloor.Percent <= 0 || record.PopulationFloor.Percent > 100 {
			return fmt.Errorf("population_floor percent must be 1 to 100, got %d", record.PopulationFloor.Percent)
		}
		definition.PopulationFloor = &domain.PopulationFloorParams{Percent: record.PopulationFloor.Percent}
	}
	return nil
}
```

##### Task 4 — `internal/faction/rulebook/effects_test.go`

Append:
```go
func TestLoadTags_PopulationFloorPercent(t *testing.T) {
	entry := "[[tags.Alpha.effects]]\ntrigger = \"population_floor_query\"\neffect = \"population_floor\"\n" +
		"[tags.Alpha.effects.filters]\nowner = \"own\"\nscope = \"faction\"\n"
	cases := []struct {
		name  string
		block string
		want  string
	}{
		{name: "no block", block: "", want: "population_floor"},
		{name: "zero", block: "[tags.Alpha.effects.population_floor]\npercent = 0\n", want: "1 to 100"},
		{name: "over 100", block: "[tags.Alpha.effects.population_floor]\npercent = 101\n", want: "1 to 100"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tags.toml")
			body := "[tags.Alpha]\nid = \"T-001\"\n" + entry + testCase.block
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatalf("writing fixture: %v", err)
			}
			_, _, err := loadTags(path)
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("err = %v, want it to contain %q", err, testCase.want)
			}
		})
	}
}
```
If `effects_test.go` lacks `os`, `path/filepath` or `strings`, add them to its imports.

##### Task 5 — `internal/faction/engine/effect/effects/population_floor.go` (new)

Full contents:
```go
package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// populationFloorRegistrar reads owner = "own" as "the owner is buying", and scope
// = "faction" as "on any world" (Colonists, Technical Expertise). asset_type narrows
// which definitions it serves. The entry's percent is of the setting's floor for
// the type; the lowest floor wins (Effort 4 Decision 4).
var populationFloorRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerPopulationFloorQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
		},
	},
	Register: registerPopulationFloor,
}

func registerPopulationFloor(ctx effect.RegisterContext) {
	ctx.Registry.RegisterPopulationFloorModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialPopulationFloor),
		&populationFloorModifier{
			owner:      ctx.Holder.Faction,
			holder:     refOf(ctx.Holder),
			assetTypes: ctx.Entry.Filters.AssetTypes,
			percent:    ctx.Entry.PopulationFloor.Percent,
		},
	)
}

// populationFloorModifier holds its owner, not its holder's asset: a Cat 4 hook
// gets no FactionState (Decision Record 19).
type populationFloorModifier struct {
	owner      *domain.Faction
	holder     holderRef
	assetTypes []domain.AssetType
	percent    int
}

func (modifier *populationFloorModifier) ModifyPopulationFloor(faction *domain.Faction, def *domain.AssetDefinition, baseFloor, floor int) int {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || faction.ID != modifier.owner.ID {
		return floor
	}
	if len(modifier.assetTypes) > 0 && !slices.Contains(modifier.assetTypes, def.Type) {
		return floor
	}
	return min(floor, baseFloor*modifier.percent/100)
}
```

##### Task 6 — `internal/faction/engine/effect/effects/population_floor_test.go` (new)

Full contents:
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

func registerPopulationFloor(registry *hooks.Registry, faction *domain.Faction, tag *domain.Tag, percent int) {
	effects.Registrars()[domain.SpecialPopulationFloor].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: faction, Tag: tag},
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerPopulationFloorQuery,
			Effect:  domain.SpecialPopulationFloor,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction,
				AssetTypes: []domain.AssetType{domain.TypeStarship}},
			PopulationFloor: &domain.PopulationFloorParams{Percent: percent},
		},
		Registry: registry,
	})
}

func TestPopulationFloor_TagHolderLowersItsTypesFloor(t *testing.T) {
	colonists := &domain.Tag{ID: "SWN-T-001", Name: "Colonists"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{colonists.ID}}
	beta := &domain.Faction{ID: "beta"}
	registry := hooks.NewRegistry()
	registerPopulationFloor(registry, alpha, colonists, 20)
	floors := map[domain.AssetType]int{domain.TypeStarship: 500000, domain.TypeFacility: 1000}
	starship := &domain.AssetDefinition{ID: "ship", Type: domain.TypeStarship}
	facility := &domain.AssetDefinition{ID: "plant", Type: domain.TypeFacility}

	cases := []struct {
		name    string
		faction *domain.Faction
		def     *domain.AssetDefinition
		want    int
	}{
		{"the holder's Starship floor is 20%", alpha, starship, 100000},
		{"another type keeps its floor", alpha, facility, 1000},
		{"a rival keeps the setting's floor", beta, starship, 500000},
	}
	for _, tc := range cases {
		if got := dispatch.ResolvePopulationFloor(registry, tc.faction, tc.def, floors); got != tc.want {
			t.Errorf("%s: floor = %d, want %d", tc.name, got, tc.want)
		}
	}

	alpha.Tags = nil
	if got := dispatch.ResolvePopulationFloor(registry, alpha, starship, floors); got != 500000 {
		t.Errorf("floor after the tag is dropped = %d, want 500000", got)
	}
}

func TestPopulationFloor_TheLowestFloorWins(t *testing.T) {
	colonists := &domain.Tag{ID: "SWN-T-001", Name: "Colonists"}
	expertise := &domain.Tag{ID: "SWN-T-018", Name: "Technical Expertise"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{expertise.ID, colonists.ID}}
	registry := hooks.NewRegistry()
	registerPopulationFloor(registry, alpha, expertise, 2)
	registerPopulationFloor(registry, alpha, colonists, 20)
	floors := map[domain.AssetType]int{domain.TypeStarship: 500000}
	starship := &domain.AssetDefinition{ID: "ship", Type: domain.TypeStarship}

	if got := dispatch.ResolvePopulationFloor(registry, alpha, starship, floors); got != 10000 {
		t.Errorf("floor = %d, want 10000 (2%% beats 20%% in either order)", got)
	}
}
```

##### Task 7 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialXPRoll:             xpRollRegistrar,
	}
}
```
**Replace with:**
```go
		domain.SpecialXPRoll:             xpRollRegistrar,
		domain.SpecialPopulationFloor:    populationFloorRegistrar,
	}
}
```

##### Task 8 — `internal/faction/engine/effect/effects/tech_level_floor.go`

The registrar accepts the two new scopes.

**Find:**
```go
// techLevelFloorRegistrar reads owner = "own" as "the owner is buying", scope =
// "world" as "the declaring asset stands on the purchase world", and scope =
// "faction" as "any world". category narrows which definitions the floor serves.
var techLevelFloorRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerTechLevelQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld, domain.ScopeFaction},
```
**Replace with:**
```go
// techLevelFloorRegistrar reads owner = "own" as "the owner is buying", scope =
// "world" as "the declaring asset stands on the purchase world", scope =
// "homeworld" as "the owner's homeworld" (Colonists), scope = "base" as "a world
// with the owner's Base" (Technical Expertise), and scope = "faction" as "any
// world". category narrows which definitions the floor serves.
var techLevelFloorRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerTechLevelQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld, domain.ScopeHomeworld, domain.ScopeBase, domain.ScopeFaction},
```

**Find:**
```go
	if modifier.scope == domain.ScopeWorld && asset.Location.WorldID != world {
		return baseTL
	}
```
**Replace with:**
```go
	switch modifier.scope {
	case domain.ScopeWorld:
		if asset.Location.WorldID != world {
			return baseTL
		}
	case domain.ScopeHomeworld:
		if modifier.owner.Homeworld.WorldID != world {
			return baseTL
		}
	case domain.ScopeBase:
		if !modifier.owner.HasBaseOn(world) {
			return baseTL
		}
	}
```

##### Task 9 — `internal/faction/engine/effect/effects/tech_level_floor_test.go`

Append:
```go
func registerTagFloor(faction *domain.Faction, tag *domain.Tag, scope domain.EffectScope) *hooks.Registry {
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialTechLevelFloor].Register(effect.RegisterContext{
		Holder:   effect.Holder{Faction: faction, Tag: tag},
		Entry:    techLevelFloorEntry(scope, 4),
		Registry: registry,
	})
	return registry
}

func TestTechLevelFloor_HomeworldScopeRaisesTheHomeworldOnly(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-001"}, Homeworld: tartarus, Assets: map[string]*domain.Asset{}}
	registry := registerTagFloor(alpha, &domain.Tag{ID: "SWN-T-001", Name: "Colonists"}, domain.ScopeHomeworld)

	if got := dispatch.ResolveWorldTechLevel(registry, alpha, forceAsset, "Tartarus", 2); got != 4 {
		t.Errorf("homeworld tech level = %d, want 4", got)
	}
	if got := dispatch.ResolveWorldTechLevel(registry, alpha, forceAsset, "Krylos", 2); got != 2 {
		t.Errorf("other world tech level = %d, want 2", got)
	}
}

func TestTechLevelFloor_BaseScopeRaisesWorldsWithABase(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-018"}, Homeworld: tartarus, Assets: map[string]*domain.Asset{},
		Bases: []*domain.Base{{ID: "alpha-krylos", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}}}}
	registry := registerTagFloor(alpha, &domain.Tag{ID: "SWN-T-018", Name: "Technical Expertise"}, domain.ScopeBase)

	if got := dispatch.ResolveWorldTechLevel(registry, alpha, forceAsset, "Krylos", 2); got != 4 {
		t.Errorf("tech level on a Base world = %d, want 4", got)
	}
	if got := dispatch.ResolveWorldTechLevel(registry, alpha, forceAsset, "Hadrian", 2); got != 2 {
		t.Errorf("tech level with no Base = %d, want 2", got)
	}
}
```

##### Commit message

```
feat(E4): a population_floor effect and two TL-floor scopes

- population_floor on population_floor_query: a percent of the setting's
  floor, narrowed by asset_type; owner own, scope faction
- New scope base: the worlds where the owner has a Base (Faction.HasBaseOn)
- tech_level_floor accepts homeworld and base
- No rulebook entry yet: Colonists and Technical Expertise land next
```

#### Commit 5 — `feat(E4): Colonists and Technical Expertise as rulebook data`

##### Task 1 — `internal/faction/domain/faction.go`

**Find:**
```go
// Tag is one rulebook tag. Government marks the tag a faction holds while it
// governs at least one world; the government record says which (A.001.6
// overview Decision 10).
type Tag struct {
	ID          string
	Name        string
	Description string
	Effect      string
	Government  bool
	Effects     []EffectDefinition
}
```
**Replace with:**
```go
// Tag is one rulebook tag. Government marks the tag a faction holds while it
// governs at least one world; the government record says which (A.001.6
// overview Decision 10). HomeworldGovernment marks a tag whose holder governs
// its homeworld from creation and holds the Government tag with it (Colonists,
// Effort 4 Decision 7).
type Tag struct {
	ID                  string
	Name                string
	Description         string
	Effect              string
	Government          bool
	HomeworldGovernment bool
	Effects             []EffectDefinition
}
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

**Find:**
```go
type tagRecord struct {
	ID          string         `toml:"id"`
	Name        string         `toml:"name"`
	Description string         `toml:"description"`
	Effect      string         `toml:"effect"`
	Government  bool           `toml:"government"`
	Effects     []effectRecord `toml:"effects"`
}
```
**Replace with:**
```go
type tagRecord struct {
	ID                  string         `toml:"id"`
	Name                string         `toml:"name"`
	Description         string         `toml:"description"`
	Effect              string         `toml:"effect"`
	Government          bool           `toml:"government"`
	HomeworldGovernment bool           `toml:"homeworld_government"`
	Effects             []effectRecord `toml:"effects"`
}
```

**Find:**
```go
		result[r.ID] = &domain.Tag{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Effect:      r.Effect,
			Government:  r.Government,
			Effects:     effects,
		}
	}
	return result, government, nil
}
```
**Replace with:**
```go
		result[r.ID] = &domain.Tag{
			ID:                  r.ID,
			Name:                r.Name,
			Description:         r.Description,
			Effect:              r.Effect,
			Government:          r.Government,
			HomeworldGovernment: r.HomeworldGovernment,
			Effects:             effects,
		}
	}
	for _, tag := range result {
		if tag.HomeworldGovernment && government == "" {
			return nil, "", fmt.Errorf("tag %q sets homeworld_government, but no tag in %s sets government", tag.ID, path)
		}
	}
	return result, government, nil
}
```

##### Task 3 — `rulebooks/swn/tags.toml`

**Find:**
```toml
[tags.Colonists]
id = "SWN-T-001"
name = "Colonists"
description = "This faction is a fresh colony on an otherwise largely untouched planet."
effect = "Has all benefits of Planetary Government for its homeworld. Homeworld treated as at least tech level 4. Colonies under 100,000 citizens cannot build Spaceship-type assets."
```
**Replace with:**
```toml
[tags.Colonists]
id = "SWN-T-001"
name = "Colonists"
homeworld_government = true
description = "This faction is a fresh colony on an otherwise largely untouched planet."
effect = "Has all benefits of Planetary Government for its homeworld. Homeworld treated as at least tech level 4. Colonies under 100,000 citizens cannot build Spaceship-type assets."

[[tags.Colonists.effects]]
trigger = "tech_level_query"
effect  = "tech_level_floor"
[tags.Colonists.effects.filters]
owner = "own"
scope = "homeworld"
[tags.Colonists.effects.tech_level_floor]
level = 4

[[tags.Colonists.effects]]
trigger = "population_floor_query"
effect  = "population_floor"
[tags.Colonists.effects.filters]
owner      = "own"
scope      = "faction"
asset_type = ["Starship"]
[tags.Colonists.effects.population_floor]
percent = 20
```

**Find:**
```toml
effect = "Treats all planets with a Base of Influence as at least tech level 4. Can build Starship-type assets on any world with at least ten thousand occupants."
```
**Replace with:**
```toml
effect = "Treats all planets with a Base of Influence as at least tech level 4. Can build Starship-type assets on any world with at least ten thousand occupants."

[[tags.TechnicalExpertise.effects]]
trigger = "tech_level_query"
effect  = "tech_level_floor"
[tags.TechnicalExpertise.effects.filters]
owner = "own"
scope = "base"
[tags.TechnicalExpertise.effects.tech_level_floor]
level = 4

[[tags.TechnicalExpertise.effects]]
trigger = "population_floor_query"
effect  = "population_floor"
[tags.TechnicalExpertise.effects.filters]
owner      = "own"
scope      = "faction"
asset_type = ["Starship"]
[tags.TechnicalExpertise.effects.population_floor]
percent = 2
```

##### Task 4 — `internal/faction/rulebook/rulebook_test.go`

In `TestLoadTagsAndGoals_Strict`, add a case after "two government tags".

**Find:**
```go
			want: "both set government",
		},
```
**Replace with:**
```go
			want: "both set government",
		},
		{
			name: "homeworld government with no government tag",
			body: "[tags.Alpha]\nid = \"T-001\"\nhomeworld_government = true\n",
			load: loadTagFile,
			want: "sets homeworld_government",
		},
```

Append:
```go
func TestLoad_ColonistsAndTechnicalExpertise(t *testing.T) {
	rb := loadShippedRulebook(t)
	floor := func(scope domain.EffectScope) domain.EffectDefinition {
		return domain.EffectDefinition{
			Trigger:        domain.TriggerTechLevelQuery,
			Effect:         domain.SpecialTechLevelFloor,
			Filters:        domain.EffectFilters{Owner: domain.OwnerOwn, Scope: scope},
			TechLevelFloor: &domain.TechLevelFloorParams{Level: 4},
		}
	}
	population := func(percent int) domain.EffectDefinition {
		return domain.EffectDefinition{
			Trigger: domain.TriggerPopulationFloorQuery,
			Effect:  domain.SpecialPopulationFloor,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction,
				AssetTypes: []domain.AssetType{domain.TypeStarship}},
			PopulationFloor: &domain.PopulationFloorParams{Percent: percent},
		}
	}
	colonists := rb.Tags["SWN-T-001"]
	if !colonists.HomeworldGovernment {
		t.Error("Colonists does not set homeworld_government")
	}
	if want := []domain.EffectDefinition{floor(domain.ScopeHomeworld), population(20)}; !reflect.DeepEqual(colonists.Effects, want) {
		t.Errorf("Colonists effects = %+v, want %+v", colonists.Effects, want)
	}
	if want := []domain.EffectDefinition{floor(domain.ScopeBase), population(2)}; !reflect.DeepEqual(rb.Tags["SWN-T-018"].Effects, want) {
		t.Errorf("Technical Expertise effects = %+v, want %+v", rb.Tags["SWN-T-018"].Effects, want)
	}
}
```

##### Task 5 — `internal/faction/tui/views/manage/wizard/wizard.go`

`Update` sends the governed world through one helper.

**Find:**
```go
		faction := m.synthesize()
		governedWorldID := ""
		if pickedGovernmentTag(m.data.selectedTags) {
			governedWorldID = m.data.governedWorldID
		}
		return m, func() tea.Msg { return msgs.CreatedMsg{Faction: faction, GovernedWorldID: governedWorldID} }
```
**Replace with:**
```go
		faction := m.synthesize()
		governedWorldID := governedWorld(m.data)
		return m, func() tea.Msg { return msgs.CreatedMsg{Faction: faction, GovernedWorldID: governedWorldID} }
```

`synthesize` adds the government tag for a homeworld-government pick.

**Find:**
```go
	allPicked := append(m.data.primaryAssets, m.data.secondaryAssets...)
	assets := instantiateAssets(allPicked, homeworld, tempFaction)

	faction := domain.NewFaction(
		id, m.data.name,
		m.data.scale,
		m.data.primaryStat, m.data.secondaryStat, m.data.tertiaryStat,
		tagIDs(m.data.selectedTags),
```
**Replace with:**
```go
	allPicked := append(m.data.primaryAssets, m.data.secondaryAssets...)
	assets := instantiateAssets(allPicked, homeworld, tempFaction)

	tags := tagIDs(m.data.selectedTags)
	if homeworldGovernmentTag(m.data.selectedTags) != nil && m.rulebook.GovernmentTagID != "" {
		tags = append(tags, m.rulebook.GovernmentTagID)
	}

	faction := domain.NewFaction(
		id, m.data.name,
		m.data.scale,
		m.data.primaryStat, m.data.secondaryStat, m.data.tertiaryStat,
		tags,
```

The tag picker refuses the pair.

**Find:**
```go
			Validate(func(picked []*domain.Tag) error {
				if len(picked) > 2 {
					return fmt.Errorf("at most 2 tags")
				}
				return nil
			}),
	)
```
**Replace with:**
```go
			Validate(func(picked []*domain.Tag) error {
				if len(picked) > 2 {
					return fmt.Errorf("at most 2 tags")
				}
				if tag := homeworldGovernmentTag(picked); tag != nil && pickedGovernmentTag(picked) {
					return fmt.Errorf("%s already governs its homeworld; drop the government tag", tag.Name)
				}
				return nil
			}),
	)
```

The homeworld picker refuses a world another faction governs.

**Find:**
```go
	homeworldGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Homeworld").
			Options(worldsToOptions(spatialMap)...).
			Value(&data.homeworldID),
	)
```
**Replace with:**
```go
	homeworldGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Homeworld").
			Options(worldsToOptions(spatialMap)...).
			Value(&data.homeworldID).
			Validate(func(worldID string) error {
				if tag := homeworldGovernmentTag(data.selectedTags); tag != nil && factionState.Ruler(worldID) != "" {
					return fmt.Errorf("another faction governs this world; %s needs its homeworld free", tag.Name)
				}
				return nil
			}),
	)
```

Add the helpers after `pickedGovernmentTag`.

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

// homeworldGovernmentTag returns the picked tag that governs its homeworld from
// creation (Colonists), or nil.
func homeworldGovernmentTag(tags []*domain.Tag) *domain.Tag {
	for _, tag := range tags {
		if tag.HomeworldGovernment {
			return tag
		}
	}
	return nil
}

// governedWorld is the world the new faction rules: the picked one for the
// government tag, the homeworld for a homeworld-government tag, else none. The
// tag picker refuses the two together (Effort 4 Decision 7).
func governedWorld(data *formData) string {
	switch {
	case pickedGovernmentTag(data.selectedTags):
		return data.governedWorldID
	case homeworldGovernmentTag(data.selectedTags) != nil:
		return data.homeworldID
	}
	return ""
}
```

##### Task 6 — `internal/faction/engine/testharness/harness.go`

**Find:**
```go
	DefSecurityPersonnel = "SWN-F1-001"
```
**Replace with:**
```go
	DefSecurityPersonnel = "SWN-F1-001"
	DefSmugglers         = "SWN-C1-001"
```

##### Task 7 — `internal/faction/engine/testharness/scenarios/tags_test.go`

Append:
```go
// --- scenario N+8: Colonists and Technical Expertise — Starships on a small world ---

func TestPopulationFloor_TagsOpenASmallWorldToStarships(t *testing.T) {
	cases := []struct {
		name  string
		tags  []string
		offer bool
	}{
		{"no tag: the setting's 500,000 floor", nil, false},
		{"Colonists: 20% of the floor", []string{"SWN-T-001"}, true},
		{"Technical Expertise: 2% of the floor", []string{"SWN-T-018"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := testharness.NewHarness(t)
			h.SpatialMap.Populations = map[string]int{"Tartarus": 150_000}
			alpha := h.AddFaction("alpha", "Tartarus", 3, 3, 3)
			alpha.Tags = tc.tags
			alpha.Coin = 20
			offered := false
			h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
				if faction.ID != "alpha" {
					return nil, nil
				}
				for _, candidate := range available {
					if candidate.Name() == "Buy Asset" {
						return candidate, nil
					}
				}
				t.Fatal("Buy Asset not available")
				return nil, nil
			}
			h.Collector.SelectBuyOrderFn = func(purchasable map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
				var fallback *domain.AssetDefinition
				for _, def := range purchasable["Tartarus"] {
					if def.ID == testharness.DefSmugglers {
						offered = true
					}
					if def.ID == testharness.DefSecurityPersonnel {
						fallback = def
					}
				}
				return action.BuyOrder{World: "Tartarus", Definition: fallback}, nil
			}

			if err := h.Engine.Turn.Start(h.FactionState); err != nil {
				t.Fatalf("Turn.Start: %v", err)
			}
			if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
				t.Fatalf("RunCycle: %v", err)
			}

			testharness.CheckStep(t, "Smugglers offered on a world of 150,000", offered == tc.offer,
				fmt.Sprintf("Smugglers offered = %v, want %v", offered, tc.offer))
		})
	}
}
```

##### Commit message

```
feat(E4): Colonists and Technical Expertise as rulebook data

- Colonists: TL4 homeworld (tech_level_floor, scope homeworld) and a 20%
  Starship floor; homeworld_government = true
- Technical Expertise: TL4 on every Base world (scope base) and a 2% floor
- The tag loader refuses homeworld_government with no government tag
- Wizard: a homeworld-government pick governs the homeworld and holds the
  government tag; refused with the government tag, or on a governed world
```

### Phase 3 — Pirates

**In plain words:** a rival asset that lands on a world with a Pirates Base pays the
Pirates 1 Coin, by drift order or by relocate. A drift order that cannot pay waits
at the edge; a relocate does not offer a world it cannot pay for.

#### Commit 6 — `feat(E4): an arrival toll rule hook`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialPopulationFloor    SpecialEffectType = "population_floor"
)
```
**Replace with:**
```go
	SpecialPopulationFloor    SpecialEffectType = "population_floor"
	SpecialArrivalToll        SpecialEffectType = "arrival_toll"
)
```

**Find:**
```go
	TriggerPopulationFloorQuery   EffectTrigger = "population_floor_query"
)
```
**Replace with:**
```go
	TriggerPopulationFloorQuery   EffectTrigger = "population_floor_query"
	TriggerArrivalQuery           EffectTrigger = "arrival_query"
)
```

**Find:**
```go
// PopulationFloorParams is a holder's floor as a percent of the setting's.
```
**Replace with:**
```go
// ArrivalTollParams is the Coin a landing rival asset pays the holder.
type ArrivalTollParams struct {
	Amount int
}

// PopulationFloorParams is a holder's floor as a percent of the setting's.
```

**Find:**
```go
	PopulationFloor *PopulationFloorParams
}
```
**Replace with:**
```go
	PopulationFloor *PopulationFloorParams
	ArrivalToll     *ArrivalTollParams
}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	PopulationFloor *populationFloorRecord `toml:"population_floor"`
}
```
**Replace with:**
```go
	PopulationFloor *populationFloorRecord `toml:"population_floor"`
	ArrivalToll     *arrivalTollRecord     `toml:"arrival_toll"`
}
```

**Find:**
```go
type populationFloorRecord struct {
```
**Replace with:**
```go
type arrivalTollRecord struct {
	Amount int `toml:"amount"`
}

type populationFloorRecord struct {
```

**Find:**
```go
		domain.TriggerHomeworldChanged, domain.TriggerGoalCompleted, domain.TriggerPopulationFloorQuery,
	}
```
**Replace with:**
```go
		domain.TriggerHomeworldChanged, domain.TriggerGoalCompleted, domain.TriggerPopulationFloorQuery,
		domain.TriggerArrivalQuery,
	}
```

**Find:**
```go
		domain.SpecialRemoveTag, domain.SpecialXPRoll, domain.SpecialPopulationFloor,
	}
```
**Replace with:**
```go
		domain.SpecialRemoveTag, domain.SpecialXPRoll, domain.SpecialPopulationFloor, domain.SpecialArrivalToll,
	}
```

**Find:**
```go
	if record.PopulationFloor != nil {
		present = append(present, domain.SpecialPopulationFloor)
	}
	return present
}
```
**Replace with:**
```go
	if record.PopulationFloor != nil {
		present = append(present, domain.SpecialPopulationFloor)
	}
	if record.ArrivalToll != nil {
		present = append(present, domain.SpecialArrivalToll)
	}
	return present
}
```

**Find:**
```go
		definition.PopulationFloor = &domain.PopulationFloorParams{Percent: record.PopulationFloor.Percent}
	}
	return nil
}
```
**Replace with:**
```go
		definition.PopulationFloor = &domain.PopulationFloorParams{Percent: record.PopulationFloor.Percent}
	case domain.SpecialArrivalToll:
		if record.ArrivalToll == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.ArrivalToll.Amount <= 0 {
			return fmt.Errorf("arrival_toll amount must be > 0, got %d", record.ArrivalToll.Amount)
		}
		definition.ArrivalToll = &domain.ArrivalTollParams{Amount: record.ArrivalToll.Amount}
	}
	return nil
}
```

##### Task 3 — `internal/faction/engine/hooks/rule_modifier.go`

**Find:**
```go
// PopulationFloorModifier lowers the world population faction needs to buy def
```
**Replace with:**
```go
// ArrivalToll is the Coin a landing asset owes FactionID, the faction that
// charges it.
type ArrivalToll struct {
	FactionID string
	Amount    int
}

// ArrivalTollRule charges mover for one of its assets landing on world (Cat 4).
// It returns 0 when it charges nothing.
type ArrivalTollRule interface {
	ArrivalToll(mover *domain.Faction, world string) int
}

// PopulationFloorModifier lowers the world population faction needs to buy def
```

##### Task 4 — `internal/faction/engine/hooks/registry.go`

**Find:**
```go
type RegisteredPopulationFloorModifier struct {
```
**Replace with:**
```go
type RegisteredArrivalTollRule struct {
	Scope  Scope
	Source string
	Hook   ArrivalTollRule
}

type RegisteredPopulationFloorModifier struct {
```

**Find:**
```go
	populationFloorModifiers map[Scope][]RegisteredPopulationFloorModifier
```
**Replace with:**
```go
	populationFloorModifiers map[Scope][]RegisteredPopulationFloorModifier
	arrivalTollRules         map[Scope][]RegisteredArrivalTollRule
```

**Find:**
```go
	registry.populationFloorModifiers = make(map[Scope][]RegisteredPopulationFloorModifier)
```
**Replace with:**
```go
	registry.populationFloorModifiers = make(map[Scope][]RegisteredPopulationFloorModifier)
	registry.arrivalTollRules = make(map[Scope][]RegisteredArrivalTollRule)
```

**Find:**
```go
	delete(registry.populationFloorModifiers, scope)
```
**Replace with:**
```go
	delete(registry.populationFloorModifiers, scope)
	delete(registry.arrivalTollRules, scope)
```

**Find:**
```go
// RegisterPopulationFloorModifier registers a Cat 4 population-floor hook under the given scope.
```
**Replace with:**
```go
// RegisterArrivalTollRule registers a Cat 4 arrival-toll hook under the given scope.
func (registry *Registry) RegisterArrivalTollRule(scope Scope, source string, hook ArrivalTollRule) {
	registry.arrivalTollRules[scope] = append(registry.arrivalTollRules[scope], RegisteredArrivalTollRule{Scope: scope, Source: source, Hook: hook})
}

// RegisterPopulationFloorModifier registers a Cat 4 population-floor hook under the given scope.
```

**Find:**
```go
// PopulationFloorModifiersFor returns the global and faction Cat 4 population-floor
```
**Replace with:**
```go
// ArrivalTollRulesAt returns the Cat 4 arrival-toll hooks of one bucket. There is
// no ...For lookup: the toll dispatcher walks every rival faction's buckets, so a
// global bucket would charge once per rival.
func (registry *Registry) ArrivalTollRulesAt(scope Scope) []RegisteredArrivalTollRule {
	return registry.arrivalTollRules[scope]
}

// PopulationFloorModifiersFor returns the global and faction Cat 4 population-floor
```

##### Task 5 — `internal/faction/engine/hooks/dispatch/rules.go`

Add `"maps"` and `"slices"` to the imports.

**Find:**
```go
// ResolvePopulationFloor returns the least world population on which faction may
```
**Replace with:**
```go
// ArrivalTolls returns the tolls one of mover's assets owes on landing on world:
// one per rule that charges, each paid to the faction that registered it. It asks
// every faction but mover, in ID order, its faction bucket and then its own
// assets' buckets in ID order, because a toll is a rival's effect (Effort 4
// Decision 8). Who skips a toll is the caller's to decide.
func ArrivalTolls(registry *hooks.Registry, factions map[string]*domain.Faction, mover *domain.Faction, world string) []hooks.ArrivalToll {
	if registry == nil || mover == nil {
		return nil
	}
	var tolls []hooks.ArrivalToll
	for _, factionID := range slices.Sorted(maps.Keys(factions)) {
		if factionID == mover.ID {
			continue
		}
		rules := slices.Clone(registry.ArrivalTollRulesAt(hooks.FactionScope(factionID)))
		for _, asset := range domain.SortedAssets(factions[factionID]) {
			rules = append(rules, registry.ArrivalTollRulesAt(hooks.AssetScope(factionID, asset.ID))...)
		}
		for _, registered := range rules {
			if amount := registered.Hook.ArrivalToll(mover, world); amount > 0 {
				tolls = append(tolls, hooks.ArrivalToll{FactionID: factionID, Amount: amount})
			}
		}
	}
	return tolls
}

// TollTotal is the Coin tolls charge in all.
func TollTotal(tolls []hooks.ArrivalToll) int {
	total := 0
	for _, toll := range tolls {
		total += toll.Amount
	}
	return total
}

// TollCharges pays each toll from payerID to the faction that charges it: a loss
// and a gain, both caused by that faction, as a steal is, so a Bank absorbs the
// loss and the holder keeps the gain (Effort 4 Decision 8).
func TollCharges(payerID string, tolls []hooks.ArrivalToll) []domain.Mutation {
	var charges []domain.Mutation
	for _, toll := range tolls {
		charges = append(charges,
			domain.CoinDelta{FactionID: payerID, Delta: -toll.Amount, Cause: string(domain.SpecialArrivalToll), CausedByFactionID: toll.FactionID},
			domain.CoinDelta{FactionID: toll.FactionID, Delta: toll.Amount, Cause: string(domain.SpecialArrivalToll), CausedByFactionID: toll.FactionID},
		)
	}
	return charges
}

// ResolvePopulationFloor returns the least world population on which faction may
```

##### Task 6 — `internal/faction/engine/hooks/dispatch/rules_test.go`

Append. Add `"reflect"` to the imports if it is missing.
```go
// stubToll charges a fixed amount on one world.
type stubToll struct {
	world  string
	amount int
}

func (stub *stubToll) ArrivalToll(_ *domain.Faction, world string) int {
	if world != stub.world {
		return 0
	}
	return stub.amount
}

func TestArrivalTolls_EveryRivalChargesInIDOrder(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterArrivalTollRule(hooks.FactionScope("zulu"), "pirates", &stubToll{world: "Krylos", amount: 1})
	registry.RegisterArrivalTollRule(hooks.AssetScope("bravo", "b1"), "toll-asset", &stubToll{world: "Krylos", amount: 2})
	registry.RegisterArrivalTollRule(hooks.FactionScope("alpha"), "own", &stubToll{world: "Krylos", amount: 5})
	registry.RegisterArrivalTollRule(hooks.FactionScope("bravo"), "elsewhere", &stubToll{world: "Hadrian", amount: 3})
	alpha := &domain.Faction{ID: "alpha"}
	factions := map[string]*domain.Faction{
		"alpha": alpha,
		"bravo": {ID: "bravo", Assets: map[string]*domain.Asset{"b1": {ID: "b1", OwnerID: "bravo"}}},
		"zulu":  {ID: "zulu"},
	}

	got := ArrivalTolls(registry, factions, alpha, "Krylos")

	want := []hooks.ArrivalToll{{FactionID: "bravo", Amount: 2}, {FactionID: "zulu", Amount: 1}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tolls = %+v, want %+v (the mover's own rule never charges it)", got, want)
	}
}

func TestTollCharges_PairsALossAndAGainCausedByTheHolder(t *testing.T) {
	got := TollCharges("alpha", []hooks.ArrivalToll{{FactionID: "bravo", Amount: 1}})
	want := []domain.Mutation{
		domain.CoinDelta{FactionID: "alpha", Delta: -1, Cause: "arrival_toll", CausedByFactionID: "bravo"},
		domain.CoinDelta{FactionID: "bravo", Delta: 1, Cause: "arrival_toll", CausedByFactionID: "bravo"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("charges = %+v, want %+v", got, want)
	}
}
```

##### Task 7 — `internal/faction/engine/effect/effects/arrival_toll.go` (new)

Full contents:
```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// arrivalTollRegistrar reads owner = "rival" as "the landing asset is a rival's",
// and scope = "base" as "on a world where the owner has a Base of Influence"
// (Pirates). The entry carries the toll (Effort 4 Decision 8).
var arrivalTollRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerArrivalQuery: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeBase},
		},
	},
	Register: registerArrivalToll,
}

func registerArrivalToll(ctx effect.RegisterContext) {
	ctx.Registry.RegisterArrivalTollRule(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialArrivalToll),
		&arrivalTollRule{
			owner:  ctx.Holder.Faction,
			holder: refOf(ctx.Holder),
			amount: ctx.Entry.ArrivalToll.Amount,
		},
	)
}

// arrivalTollRule holds its owner, not its holder's asset: a Cat 4 hook gets no
// FactionState (Decision Record 19). The owner's Bases are read live at query time.
type arrivalTollRule struct {
	owner  *domain.Faction
	holder holderRef
	amount int
}

func (rule *arrivalTollRule) ArrivalToll(mover *domain.Faction, world string) int {
	if _, live := rule.holder.liveIn(rule.owner); !live || mover.ID == rule.owner.ID || !rule.owner.HasBaseOn(world) {
		return 0
	}
	return rule.amount
}
```

##### Task 8 — `internal/faction/engine/effect/effects/arrival_toll_test.go` (new)

Full contents:
```go
package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

func TestArrivalToll_TagHolderChargesRivalsOnItsBaseWorlds(t *testing.T) {
	pirates := &domain.Tag{ID: "SWN-T-010", Name: "Pirates"}
	bravo := &domain.Faction{ID: "bravo", Tags: []string{pirates.ID},
		Bases: []*domain.Base{{ID: "bravo-krylos", OwnerID: "bravo", Location: domain.Location{WorldID: "Krylos"}}}}
	alpha := &domain.Faction{ID: "alpha"}
	factions := map[string]*domain.Faction{"alpha": alpha, "bravo": bravo}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialArrivalToll].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: bravo, Tag: pirates},
		Entry: domain.EffectDefinition{
			Trigger:     domain.TriggerArrivalQuery,
			Effect:      domain.SpecialArrivalToll,
			Filters:     domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeBase},
			ArrivalToll: &domain.ArrivalTollParams{Amount: 1},
		},
		Registry: registry,
	})

	if got, want := dispatch.ArrivalTolls(registry, factions, alpha, "Krylos"), []hooks.ArrivalToll{{FactionID: "bravo", Amount: 1}}; !reflect.DeepEqual(got, want) {
		t.Errorf("tolls on the Base world = %+v, want %+v", got, want)
	}
	if got := dispatch.ArrivalTolls(registry, factions, alpha, "Hadrian"); len(got) != 0 {
		t.Errorf("tolls with no Base = %+v, want none", got)
	}
	if got := dispatch.ArrivalTolls(registry, factions, bravo, "Krylos"); len(got) != 0 {
		t.Errorf("tolls on the holder's own asset = %+v, want none", got)
	}
	bravo.Tags = nil
	if got := dispatch.ArrivalTolls(registry, factions, alpha, "Krylos"); len(got) != 0 {
		t.Errorf("tolls after the tag is dropped = %+v, want none", got)
	}
}
```

##### Task 9 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialPopulationFloor:    populationFloorRegistrar,
	}
}
```
**Replace with:**
```go
		domain.SpecialPopulationFloor:    populationFloorRegistrar,
		domain.SpecialArrivalToll:        arrivalTollRegistrar,
	}
}
```

##### Commit message

```
feat(E4): an arrival toll rule hook

- arrival_toll on arrival_query: owner rival, scope base, amount from the
  entry
- Cat 4 ArrivalTollRule and dispatch.ArrivalTolls, a walk of every rival
  faction in ID order, one toll per rule that charges
- dispatch.TollCharges pays each toll as a loss and a gain caused by the
  holder, so a Bank absorbs the loss and the holder keeps the gain
- No caller yet: arrivals and relocates charge it next
```

#### Commit 7 — `feat(E4): arrivals and relocates pay the toll`

##### Task 1 — `internal/faction/domain/mutation.go`

**Find:**
```go
	// HoldLockdownUnpaid: the faction cannot pay the 1d4 yet (overview Decision 15).
	HoldLockdownUnpaid HoldReason = "lockdown_unpaid"
)
```
**Replace with:**
```go
	// HoldLockdownUnpaid: the faction cannot pay the 1d4 yet (overview Decision 15).
	HoldLockdownUnpaid HoldReason = "lockdown_unpaid"
	// HoldTollUnpaid: an owner aboard cannot pay its arrival tolls yet (Effort 4 Decision 10).
	HoldTollUnpaid HoldReason = "toll_unpaid"
)
```

##### Task 2 — `internal/faction/engine/arrival.go`

Rewritten: the tolls join the arrival, and the Coin spent is tracked per faction.
Full contents:
```go
package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// checkArrivals decides, for each order the tick completes, whether it may land,
// and returns the tick's mutations with each completion replaced by what the
// arrival does (D-24, D-27). The government asked is the one ruling at arrival. It
// keeps a running total of the Coin each faction spends this tick on lockdowns and
// tolls, so two arrivals cannot spend the same Coin. It runs between the tick and
// its reactors, so the transport reactor moves the cargo to the edge with its
// carrier (D-28).
func (e *Engine) checkArrivals(faction *domain.Faction, factionState *state.FactionState, tickMutations []domain.Mutation, collector hooks.PermissionCollector) ([]domain.Mutation, error) {
	checked := make([]domain.Mutation, 0, len(tickMutations))
	spent := map[string]int{}
	for _, mutation := range tickMutations {
		completed, ok := mutation.(domain.MovementOrderCompleted)
		if !ok {
			checked = append(checked, mutation)
			continue
		}
		arrival, costs, err := e.resolveArrival(faction, factionState, completed, collector, spent)
		if err != nil {
			return nil, err
		}
		for payerID, cost := range costs {
			spent[payerID] += cost
		}
		checked = append(checked, arrival...)
	}
	return checked, nil
}

// resolveArrival decides one arriving order and returns the mutations that take
// the completion's place, with the Coin each faction spends on them. A Transport
// Lockdown on the destination comes first (Effort 5 Decision 14), then the tolls,
// then the government, asked for each asset aboard, the carrier first, then its
// cargo in manifest order, each for its own owner; the first refusal holds the
// whole order (D-28). The tolls come before the government so a faction that
// cannot pay is never asked (Effort 4 Decision 10). spent is the Coin each faction
// has spent this tick. An order that lands keeps its completion and pays its
// tolls after it.
func (e *Engine) resolveArrival(faction *domain.Faction, factionState *state.FactionState, completed domain.MovementOrderCompleted, collector hooks.PermissionCollector, spent map[string]int) ([]domain.Mutation, map[string]int, error) {
	carrier := faction.Assets[completed.AssetID]
	if carrier == nil || carrier.CurrentOrder == nil {
		return []domain.Mutation{completed}, nil, nil
	}
	order := carrier.CurrentOrder
	var profile *domain.TransportProfile
	if carrierDef := e.Rulebook.Assets[carrier.DefinitionID]; carrierDef != nil {
		profile = carrierDef.Transport
	}
	ignoresLockdown := profile != nil && profile.IgnoresLockdown
	if !order.LockdownPaid && !ignoresLockdown {
		if lockerID := domain.LockedOutBy(factionState.Factions, order.Destination.WorldID, faction.ID); lockerID != "" {
			hold, cost := lockdownHold(completed, order, lockerID, e.Rand.Roll(4), faction.Coin-spent[faction.ID])
			return hold, map[string]int{faction.ID: cost}, nil
		}
	}
	assets := aboard(factionState, carrier)
	var tolls []domain.Mutation
	owed := map[string]int{}
	// deliberate: the carrier's lockdown bypass covers every toll aboard, cargo included, because Robert reads the toll as shipping pressure, as a lockdown is (Effort 4 Decision 9).
	if !ignoresLockdown {
		tolls, owed = e.tollCharges(factionState, assets, order.Destination.WorldID)
	}
	for payerID, amount := range owed {
		if factionState.Factions[payerID].Coin-spent[payerID] < amount {
			return []domain.Mutation{holdAtEdge(completed, order, domain.HoldTollUnpaid)}, nil, nil
		}
	}
	// deliberate: the carrier's bypass covers its cargo, because the book lets these transports land "any one … asset" past the government's objection (D-25).
	if profile == nil || !profile.IgnoresPermission {
		for _, asset := range assets {
			def := e.Rulebook.Assets[asset.DefinitionID]
			requester := factionState.Factions[asset.OwnerID]
			if def == nil || requester == nil {
				continue
			}
			refusal, err := dispatch.RequestArrivalPermission(e.Hooks, collector, factionState, e.Rulebook, requester, asset, def, order.Destination.WorldID, domain.PermissionCauseArrival)
			if err != nil {
				return nil, nil, err
			}
			if refusal != nil {
				return []domain.Mutation{*refusal, holdAtEdge(completed, order, domain.HoldRefused)}, nil, nil
			}
		}
	}
	return append([]domain.Mutation{completed}, tolls...), owed, nil
}

// tollCharges returns the tolls the assets aboard owe on landing at worldID, each
// paid by the asset's current owner, and the Coin each owner owes in all.
func (e *Engine) tollCharges(factionState *state.FactionState, assets []*domain.Asset, worldID string) ([]domain.Mutation, map[string]int) {
	var charges []domain.Mutation
	owed := map[string]int{}
	for _, asset := range assets {
		// deliberate: a stealthed asset pays no toll, though it still pays a lockdown, because Robert reads the toll as a levy on what the holder can see (Effort 4 Decision 9).
		if asset.Stealthy {
			continue
		}
		tolls := dispatch.ArrivalTolls(e.Hooks, factionState.Factions, factionState.Factions[asset.OwnerID], worldID)
		if len(tolls) == 0 {
			continue
		}
		charges = append(charges, dispatch.TollCharges(asset.OwnerID, tolls)...)
		owed[asset.OwnerID] += dispatch.TollTotal(tolls)
	}
	return charges, owed
}

// lockdownHold charges the order's 1d4 fee and holds it a turn at the edge; a
// faction that cannot pay holds unpaid and tries again at its next tick (overview
// Decision 15). The fee names lockerID, the Lockdown's owner, as its cause. It
// returns the mutations and the Coin they spend.
func lockdownHold(completed domain.MovementOrderCompleted, order *domain.MovementOrder, lockerID string, fee, coin int) ([]domain.Mutation, int) {
	if coin < fee {
		return []domain.Mutation{holdAtEdge(completed, order, domain.HoldLockdownUnpaid)}, 0
	}
	return []domain.Mutation{
		domain.CoinDelta{FactionID: completed.FactionID, Delta: -fee, Cause: "transport_lockdown", CausedByFactionID: lockerID},
		holdAtEdge(completed, order, domain.HoldLockdownPaid),
	}, fee
}

// aboard returns the carrier and the cargo its order carries, each found under its
// current owner: a capture can make a rider's owner differ from the carrier's.
func aboard(factionState *state.FactionState, carrier *domain.Asset) []*domain.Asset {
	assets := []*domain.Asset{carrier}
	for _, cargoID := range carrier.CurrentOrder.CargoAssetIDs {
		if owner := domain.AssetOwner(factionState.Factions, cargoID); owner != nil {
			assets = append(assets, owner.Assets[cargoID])
		}
	}
	return assets
}

// holdAtEdge replaces completed with a hold at the last hex before the destination
// (overview Decision 16). Path[0] is the asset's own hex, so on a two-hex path the
// hold is the in-orbit state right after issue.
func holdAtEdge(completed domain.MovementOrderCompleted, order *domain.MovementOrder, reason domain.HoldReason) domain.MovementOrderHeld {
	step := max(len(order.Path)-2, 0)
	return domain.MovementOrderHeld{
		FactionID:         completed.FactionID,
		AssetID:           completed.AssetID,
		WorldID:           order.Destination.WorldID,
		NewStepIdx:        step,
		RegionHex:         order.Path[step],
		Reason:            reason,
		Cause:             domain.CauseMovementTick,
		CausedByFactionID: completed.CausedByFactionID,
	}
}
```

##### Task 3 — `internal/faction/engine/arrival_test.go`

A carrier that ignores lockdowns joins the test rulebook.

**Find:**
```go
		"mercs":   {ID: "mercs", Name: "Mercs", Speed: 2, Type: domain.TypeMilitaryUnit, Flags: []domain.AssetFlag{domain.FlagPermission}},
	}}
}
```
**Replace with:**
```go
		"mercs":   {ID: "mercs", Name: "Mercs", Speed: 2, Type: domain.TypeMilitaryUnit, Flags: []domain.AssetFlag{domain.FlagPermission}},
		"landers": {ID: "landers", Name: "Landers", Speed: 3, Transport: &domain.TransportProfile{
			MaxCargo: 1, CargoTypes: []domain.AssetType{domain.TypeMilitaryUnit}, IgnoresLockdown: true,
		}},
	}}
}
```

Append:
```go
// stubPirates charges 1 Coin on every world but its owner's own assets.
type stubPirates struct{ ownerID string }

func (stub *stubPirates) ArrivalToll(mover *domain.Faction, _ string) int {
	if mover.ID == stub.ownerID {
		return 0
	}
	return 1
}

func TestCheckArrivals_Tolls(t *testing.T) {
	toll := func(payerID string) []domain.Mutation {
		return []domain.Mutation{
			domain.CoinDelta{FactionID: payerID, Delta: -1, Cause: "arrival_toll", CausedByFactionID: "bravo"},
			domain.CoinDelta{FactionID: "bravo", Delta: 1, Cause: "arrival_toll", CausedByFactionID: "bravo"},
		}
	}
	cases := []struct {
		name          string
		carrierDef    string
		coin          int
		stealthyCargo bool
		capturedCargo bool
		wantTolls     []string
		wantHeld      bool
	}{
		{name: "carrier and cargo each pay", carrierDef: "ship", coin: 5, wantTolls: []string{"alpha", "alpha"}},
		{name: "a stealthed rider pays nothing", carrierDef: "ship", coin: 5, stealthyCargo: true, wantTolls: []string{"alpha"}},
		{name: "a captured rider's owner pays for it", carrierDef: "ship", coin: 5, capturedCargo: true, wantTolls: []string{"alpha", "charlie"}},
		{name: "a lockdown bypass covers the whole order", carrierDef: "landers", coin: 5},
		{name: "an owner who cannot pay holds the order", carrierDef: "ship", coin: 1, wantHeld: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eng, factionState, alpha, completed := arrivalFixture(tc.carrierDef, "militia", 3)
			delete(factionState.Governments, "Krylos")
			alpha.Coin = tc.coin
			alpha.Assets["c1"].Stealthy = tc.stealthyCargo
			if tc.capturedCargo {
				rider := alpha.Assets["c1"]
				delete(alpha.Assets, "c1")
				rider.OwnerID = "charlie"
				factionState.Factions["charlie"] = &domain.Faction{ID: "charlie", Coin: 5, Assets: map[string]*domain.Asset{"c1": rider}}
			}
			eng.Hooks.RegisterArrivalTollRule(hooks.FactionScope("bravo"), "pirates", &stubPirates{ownerID: "bravo"})

			got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, &minimalPhaseCollector{})

			if err != nil {
				t.Fatalf("checkArrivals: %v", err)
			}
			if tc.wantHeld {
				want := []domain.Mutation{domain.MovementOrderHeld{FactionID: "alpha", AssetID: "t1", WorldID: "Krylos", NewStepIdx: 1, RegionHex: arrivalHex(1), Reason: domain.HoldTollUnpaid, Cause: domain.CauseMovementTick, CausedByFactionID: "alpha"}}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("mutations = %+v, want %+v", got, want)
				}
				return
			}
			want := []domain.Mutation{completed}
			for _, payerID := range tc.wantTolls {
				want = append(want, toll(payerID)...)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("mutations = %+v, want %+v", got, want)
			}
		})
	}
}

func TestCheckArrivals_TwoTollsCannotSpendTheSameCoin(t *testing.T) {
	eng, factionState, alpha, first := arrivalFixture("ship", "", 3)
	delete(factionState.Governments, "Krylos")
	alpha.Coin = 1
	second := alpha.Assets["t1"].Clone()
	second.ID = "t2"
	second.CurrentOrder.AssetID = "t2"
	second.DefinitionID = "militia"
	alpha.Assets["t2"] = second
	secondCompleted := first
	secondCompleted.AssetID = "t2"
	eng.Hooks.RegisterArrivalTollRule(hooks.FactionScope("bravo"), "pirates", &stubPirates{ownerID: "bravo"})

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{first, secondCompleted}, &minimalPhaseCollector{})

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	held, ok := got[len(got)-1].(domain.MovementOrderHeld)
	if !ok || held.AssetID != "t2" || held.Reason != domain.HoldTollUnpaid {
		t.Errorf("last mutation = %+v, want t2 held unpaid: the first toll spent the one Coin", got[len(got)-1])
	}
}
```

##### Task 4 — `internal/faction/engine/relocate/relocate.go`

The relocate drops the worlds it cannot pay for, adds the tolls to its total, and
pays each order's tolls.

**Find:**
```go
	if request.Faction.Coin < params.CoinCost {
		return nil, fmt.Errorf("relocate: %w: need %d, have %d", ErrInsufficientCoin, params.CoinCost, request.Faction.Coin)
	}
```
**Replace with:**
```go
	if request.Faction.Coin < params.CoinCost {
		return nil, fmt.Errorf("relocate: %w: need %d, have %d", ErrInsufficientCoin, params.CoinCost, request.Faction.Coin)
	}
	destinations = affordableDestinations(request, destinations)
	if len(destinations) == 0 {
		return nil, fmt.Errorf("relocate: every destination's toll is more than the faction holds: %w", ErrInsufficientCoin)
	}
```

**Find:**
```go
	if cost := totalCost(params, len(orders)); request.Faction.Coin < cost {
		return nil, fmt.Errorf("relocate: %w: need %d, have %d", ErrInsufficientCoin, cost, request.Faction.Coin)
	}
```
**Replace with:**
```go
	cost := totalCost(params, len(orders))
	for _, order := range orders {
		cost += dispatch.TollTotal(orderTolls(request, order))
	}
	if request.Faction.Coin < cost {
		return nil, fmt.Errorf("relocate: %w: need %d, have %d", ErrInsufficientCoin, cost, request.Faction.Coin)
	}
```

**Find:**
```go
		mutations = append(mutations, muts...)
		moved++
```
**Replace with:**
```go
		mutations = append(mutations, muts...)
		mutations = append(mutations, dispatch.TollCharges(request.Faction.ID, orderTolls(request, order))...)
		moved++
```

**Find:**
```go
// eligibleAssets returns the faction's assets that an other-scope relocation may
```
**Replace with:**
```go
// affordableDestinations drops each world whose tolls, beside the relocate's own
// cost for one asset, are more than the faction holds: a mover with no Coin cannot
// order the move (D-6).
func affordableDestinations(request Request, destinations []string) []string {
	if request.FactionState == nil {
		return destinations
	}
	return slices.DeleteFunc(slices.Clone(destinations), func(worldID string) bool {
		// deliberate: stealth is not read here, so a short purse can lose a world a stealthed asset would reach free, because the filter keeps one number per world (Effort 4 Decision 11).
		tolls := dispatch.ArrivalTolls(request.Registry, request.FactionState.Factions, request.Faction, worldID)
		return request.Faction.Coin < request.Params.CoinCost+dispatch.TollTotal(tolls)
	})
}

// orderTolls returns the tolls order's asset owes on landing. A stealthed asset
// owes none (Effort 4 Decision 9).
func orderTolls(request Request, order action.RelocateOrder) []hooks.ArrivalToll {
	if request.FactionState == nil || order.Asset.Stealthy {
		return nil
	}
	return dispatch.ArrivalTolls(request.Registry, request.FactionState.Factions, request.Faction, order.DestinationWorld)
}

// eligibleAssets returns the faction's assets that an other-scope relocation may
```

##### Task 5 — `internal/faction/engine/relocate/relocate_test.go`

Append. Add `"reflect"` to the imports.
```go
// stubPirates charges 1 Coin to every faction but its owner, on "mid" only.
type stubPirates struct{}

func (stubPirates) ArrivalToll(mover *domain.Faction, world string) int {
	if mover.ID == "pirate" || world != "mid" {
		return 0
	}
	return 1
}

func tollRequest(t *testing.T, coin int, collector *mocks.MockCollector) (Request, *domain.Asset) {
	t.Helper()
	acting := &domain.Asset{ID: "sf", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	faction := &domain.Faction{ID: "f1", Coin: coin, Assets: map[string]*domain.Asset{"sf": acting}}
	registry := hooks.NewRegistry()
	registry.RegisterArrivalTollRule(hooks.FactionScope("pirate"), "pirates", stubPirates{})
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "pirate": {ID: "pirate"}}}
	params := &domain.RelocateParams{Scope: domain.RelocateSelf, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1}
	return Request{
		Faction: faction, Acting: acting, Params: params, World: relocEngine(), Rulebook: relocRulebook(),
		Registry: registry, FactionState: factionState, Collector: collector,
	}, acting
}

func TestRun_UnaffordableTollWorldIsNotADestination(t *testing.T) {
	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	request, acting := tollRequest(t, 1, collector)
	var offered []string
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ *domain.Asset, _ []*domain.Asset, worlds []string, _ int) ([]action.RelocateOrder, error) {
			offered = worlds
			return []action.RelocateOrder{{Asset: acting, DestinationWorld: "near"}}, nil
		})

	if _, err := Run(request); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if slices.Contains(offered, "mid") || !slices.Contains(offered, "near") {
		t.Errorf("destinations = %v, want near and not mid (1 Coin pays the move, not the toll)", offered)
	}
}

func TestRun_TollWorldPaysTheHolder(t *testing.T) {
	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	request, acting := tollRequest(t, 5, collector)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: acting, DestinationWorld: "mid"}}, nil)

	mutations, err := Run(request)

	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: -1, Cause: "arrival_toll", CausedByFactionID: "pirate"},
		domain.CoinDelta{FactionID: "pirate", Delta: 1, Cause: "arrival_toll", CausedByFactionID: "pirate"},
	}
	if got := mutations[len(mutations)-2:]; !reflect.DeepEqual(got, want) {
		t.Errorf("last mutations = %+v, want the toll pair %+v", got, want)
	}
}
```

##### Task 6 — `internal/faction/narrative/wire_templates.go`

**Find:**
```go
	domain.HoldLockdownUnpaid: {
		"{faction}'s {asset} cannot pay its way through the shipping lockdown at {world} and waits in orbit.",
	},
}
```
**Replace with:**
```go
	domain.HoldLockdownUnpaid: {
		"{faction}'s {asset} cannot pay its way through the shipping lockdown at {world} and waits in orbit.",
	},
	domain.HoldTollUnpaid: {
		"{faction}'s {asset} cannot pay the pirates' toll at {world} and waits in orbit.",
	},
}
```

##### Task 7 — `rulebooks/swn/tags.toml`

**Find:**
```toml
effect = "Any movement of an asset onto a world that has a Base of Influence for this faction costs one extra Coin, paid to this faction."
```
**Replace with:**
```toml
effect = "Any movement of an asset onto a world that has a Base of Influence for this faction costs one extra Coin, paid to this faction."

[[tags.Pirates.effects]]
trigger = "arrival_query"
effect  = "arrival_toll"
[tags.Pirates.effects.filters]
owner = "rival"
scope = "base"
[tags.Pirates.effects.arrival_toll]
amount = 1
```

##### Task 8 — `internal/faction/rulebook/rulebook_test.go`

Append:
```go
func TestLoad_PiratesToll(t *testing.T) {
	rb := loadShippedRulebook(t)
	want := []domain.EffectDefinition{{
		Trigger:     domain.TriggerArrivalQuery,
		Effect:      domain.SpecialArrivalToll,
		Filters:     domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeBase},
		ArrivalToll: &domain.ArrivalTollParams{Amount: 1},
	}}
	if got := rb.Tags["SWN-T-010"].Effects; !reflect.DeepEqual(got, want) {
		t.Errorf("Pirates effects = %+v, want %+v", got, want)
	}
}
```

##### Commit message

```
feat(E4): arrivals and relocates pay the toll

- A drift arrival checks the tolls after the lockdown and before the
  government; an owner aboard who cannot pay holds the order at the edge
  (HoldTollUnpaid) until its next tick
- Each asset aboard pays from its current owner; a stealthed asset pays
  nothing; a carrier that ignores lockdowns covers its whole order
- checkArrivals tracks the Coin spent per faction
- A relocate drops the worlds it cannot pay for before the prompt, adds the
  tolls to its total, and pays each order's tolls
- Pirates as rulebook data: arrival_toll, amount 1
```
