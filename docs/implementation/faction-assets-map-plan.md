# Implementation Plan — `Faction.Assets` Slice → Map

Converts `Faction.Assets` from `[]*Asset` to `map[string]*Asset` keyed by instance ID,
replacing O(n) linear scans in the mutation engine, goal engine, and action layer with O(1) lookups.

Source write-up: `docs/dev_journal/planned-work.md` → **`Faction.Assets` Slice → Map**

<br/>

## Phase 1 — Core type change

**Commit:** `refactor: convert Faction.Assets to map[string]*Asset`

All compile-breaking sites and the initialization paths. After this commit the build passes and the full test suite is green.

<br/>

### 1a. `internal/faction/domain/faction.go`

Change the field:

```go
// before
Assets []*Asset `toml:"assets"`

// after
Assets map[string]*Asset `toml:"assets"`
```

Add a `SortedAssets` helper (used in Phase 1c bookkeeping and Phase 1d sell_asset). Placed in `domain` so both callers can import it without a dependency cycle:

```go
func SortedAssets(faction *Faction) []*Asset {
    sorted := make([]*Asset, 0, len(faction.Assets))
    for _, asset := range faction.Assets {
        sorted = append(sorted, asset)
    }
    slices.SortFunc(sorted, func(a, b *Asset) int {
        return strings.Compare(a.ID, b.ID)
    })
    return sorted
}
```

Imports needed: `slices`, `strings`.

<br/>

### 1b. `internal/faction/engine/mutation/mutation.go`

Seven asset mutations change. Five scan loops become direct lookups; two structural cases change shape.

**`AssetRemoved`** — filter-loop becomes a single delete:

```go
case domain.AssetRemoved:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        delete(faction.Assets, v.AssetID)
    }
```

**`AssetAdded`** — append becomes a map assign:

```go
case domain.AssetAdded:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        asset := v.Asset
        faction.Assets[asset.ID] = &asset
    }
```

**`AssetMaintainedFlag`, `AssetHPDelta`, `AssetStealthCleared`, `AssetMoved`, `AssetStealthApplied`** — each replaces its `for _, asset := range ... break` scan with a direct lookup:

```go
// pattern for all five
case domain.AssetMaintainedFlag:
    if faction, ok := factionState.Factions[v.FactionID]; ok {
        if asset := faction.Assets[v.AssetID]; asset != nil {
            asset.Maintained = v.Maintained
        }
    }
```

Apply the same pattern for the other four: `AssetHPDelta` (`.CurrentHP += v.Delta`), `AssetStealthCleared` (`.Stealthy = false`), `AssetMoved` (`.Location = v.ToLocation`), `AssetStealthApplied` (`.Stealthy = true`).

<br/>

### 1c. `internal/faction/state/faction_state.go`

After decode, initialize nil `Assets` maps on each faction. Mirrors the existing `Factions` nil check:

```go
func Load(path string) (*FactionState, error) {
    // ... existing decode ...
    if s.Factions == nil {
        s.Factions = make(map[string]*domain.Faction)
    }
    for _, faction := range s.Factions {
        if faction.Assets == nil {
            faction.Assets = make(map[string]*domain.Asset)
        }
    }
    return &s, nil
}
```

<br/>

### 1d. `internal/faction/engine/turn/bookkeeping.go`

The `applyMaintenance` function iterates `faction.Assets` twice. The first pass (count by category) has no ordering requirement. The second pass (charge each asset) has a determinism requirement: when coin runs out mid-loop, which assets are dropped must be reproducible across runs.

Replace both loops with sorted iteration using `domain.SortedAssets`:

```go
// First pass — count by category
for _, asset := range domain.SortedAssets(faction) {
    ...
}

// Second pass — charge maintenance
for _, asset := range domain.SortedAssets(faction) {
    ...
}
```

The sort order is by asset ID (ascending). This matches the invariant that asset IDs encode ownership and purchase order (`alpha-F1-001-1`), so the sort is semantically stable across runs.

<br/>

### 1e. `internal/faction/engine/action/actions/sell_asset.go`

`SellAsset.Inputs` passes `faction.Assets` directly to `SelectAsset`. The `action.Collector` interface keeps its `[]*domain.Asset` signature — the conversion is internal:

```go
func (sa *SellAsset) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
    selected, err := sa.collector.SelectAsset(domain.SortedAssets(faction), rulebook)
    ...
}
```

<br/>

### 1f. `internal/faction/engine/testharness/harness.go`

Two sites:

**`AddFaction`** — slice literal becomes a map literal with the asset ID as key:

```go
faction.Assets = map[string]*domain.Asset{
    id + "-asset-1": {
        ID:           id + "-asset-1",
        DefinitionID: DefSecurityPersonnel,
        OwnerID:      id,
        Location:     homeworld,
        CurrentHP:    3,
        Ready:        true,
        Maintained:   true,
    },
}
```

**`AddAssetOnWorld`** — append becomes map assign:

```go
faction.Assets[asset.ID] = asset
```

<br/>

### 1g. Test files — compile-breaking sites

Three test files use positional access or slice literals that no longer compile.

**`internal/faction/engine/mutation/mutation_test.go`**

`newMutationTestState` — slice literal → map:

```go
Assets: map[string]*domain.Asset{
    "a1": {ID: "a1", DefinitionID: "infantry", Location: "Anchorage", Maintained: true},
    "a2": {ID: "a2", DefinitionID: "spy_net",  Location: "Anchorage", Maintained: false},
},
```

Positional assertions — replace each `Assets[0]` / `Assets[1]` with the known asset ID:

- `assets[0].ID != "a2"` → `assets["a2"] == nil`
- `s.Factions["f1"].Assets[0].Maintained` → `s.Factions["f1"].Assets["a1"].Maintained`
- `s.Factions["f1"].Assets[1].Maintained` → `s.Factions["f1"].Assets["a2"].Maintained`
- `s.Factions["f1"].Assets[0].Stealthy = true` → `s.Factions["f1"].Assets["a1"].Stealthy = true`
- `s.Factions["f1"].Assets[0].Stealthy` → `s.Factions["f1"].Assets["a1"].Stealthy`

**`internal/faction/state/faction_state_test.go`**

Slice literal in the `original` fixture → map:

```go
Assets: map[string]*domain.Asset{
    "iron-collective-asset-0": {
        ID:           "iron-collective-asset-0",
        DefinitionID: "F1-001",
        OwnerID:      "iron-collective",
        Location:     "Tartarus",
        CurrentHP:    3,
        Stealthy:     false,
        Ready:        true,
        Maintained:   true,
    },
},
```

Positional access in assertions:

```go
// before
a  := got.Assets[0]
wa := want.Assets[0]

// after
a  := got.Assets["iron-collective-asset-0"]
wa := want.Assets["iron-collective-asset-0"]
```

**`internal/faction/narrative/digest/digest_test.go`** (line 162)

```go
// before
fs.Factions["faction-a"].Assets = []*domain.Asset{{ID: "asset-1", DefinitionID: "def-1"}}

// after
fs.Factions["faction-a"].Assets = map[string]*domain.Asset{
    "asset-1": {ID: "asset-1", DefinitionID: "def-1"},
}
```

<br/>

### 1h. `cmd/faction-manager/commands/faction/create.go`

The faction struct literal doesn't set `Assets`. A newly created faction has no assets, so initializing to an empty map is correct. Add explicitly to avoid a nil map if the faction is used before being saved and reloaded:

```go
faction := &domain.Faction{
    ...
    Assets: make(map[string]*domain.Asset),
}
```

<br/>
<br/>

## Phase 2 — Lookup simplifications

**Commit:** `refactor: replace Assets linear scans with direct map lookups`

Call sites that were doing O(n) ID scans can now be simplified. These are correctness-neutral — the scan and the lookup produce the same result — but remove noise from the code.

<br/>

### 2a. `internal/faction/engine/goal/progress.go`

`findAsset` is a helper that linearly scans a faction's assets by ID. Replace its body with a direct map lookup:

```go
func findAsset(faction *domain.Faction, assetID string) *domain.Asset {
    return faction.Assets[assetID]
}
```

All callers (`progressInsideEnemyTerritory`, `progressInvincibleValor`, `countAssetKillsByCategory`) are unchanged — they already check for nil return.

<br/>

### 2b. `internal/faction/narrative/digest/resolve.go`

`resolveAssetName` scans the faction's assets for the matching ID. Replace the loop with a direct lookup:

```go
func resolveAssetName(factionID, assetID string, factionState *state.FactionState, rulebook *rulebook.Rulebook) string {
    if factionState == nil {
        return assetID
    }
    f, ok := factionState.Factions[factionID]
    if !ok {
        return assetID
    }
    asset, ok := f.Assets[assetID]
    if !ok {
        return assetID
    }
    if rulebook != nil {
        if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
            return def.Name
        }
    }
    return assetID
}
```

<br/>

### 2c. `internal/faction/engine/action/actions/attack.go`

`ownerFaction` scans all factions to find the one that owns a given asset. Since `Asset.OwnerID` exists, the nested loop is unnecessary:

```go
func ownerFaction(factionState *state.FactionState, asset *domain.Asset) *domain.Faction {
    return factionState.Factions[asset.OwnerID]
}
```

<br/>
<br/>

## Phase 3 — Decisions log

**Commit:** `docs: log decisions for Faction.Assets map conversion`

Add a new branch section to `docs/dev_journal/decisions-log.md`:

### Branch: `refactor/faction-assets-map`

| # | Decision | Rationale |
|---|----------|-----------|
| 191 | `Faction.Assets` converted from `[]*Asset` to `map[string]*Asset` | Matches the precedent of `FactionState.Factions` (Decision 64); pre-1.0 breaking change is acceptable; O(1) lookups in mutation engine remove 5+ linear scans per mutation batch |
| 192 | Canonical sort order in `applyMaintenance` is ascending asset ID | Map iteration is unordered; maintenance charging is coin-budget-sensitive (first asset wins when coin runs out); sorting by ID makes the result deterministic and reproducible across runs |
| 193 | `action.Collector.SelectAsset` keeps `[]*domain.Asset` parameter; map→slice conversion is internal to `SellAsset.Inputs` | The collector interface should not know about the internal storage shape; `domain.SortedAssets` produces a consistent ordered slice without exposing the map |
| 194 | `ownerFaction` in `attack.go` simplified to `factionState.Factions[asset.OwnerID]` | `Asset.OwnerID` already encodes the owner; the previous O(n × factions) scan was a holdover from when the field may not have been trusted; direct lookup is correct and faster |

<br/>
<br/>

## TOML schema note

The TOML encoding of `Faction.Assets` changes shape:

```toml
# before (array of tables)
[[factions.iron-collective.assets]]
id = "iron-collective-asset-0"
...

# after (inline map)
[factions.iron-collective.assets."iron-collective-asset-0"]
id = "iron-collective-asset-0"
...
```

No live campaign files exist as of this writing. If any do at execution time, they need a one-time migration: re-save them after loading with the new binary (Load decodes the old format; Save writes the new format). Document this in the commit message if migration is performed.
