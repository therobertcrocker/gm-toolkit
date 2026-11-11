# Phase 2 Commit 1 — Session State

**Goal:** `refactor: migrate Asset.Location to Location struct`

**Plan:** `asset-movement-redesign-effort-1-plan.md` → Phase 2 Commit 1

---

## Done

- `internal/faction/domain/asset.go` — `Location string` → `Location Location` ✓
- `internal/faction/engine/mutation/mutation.go` — `case domain.AssetMoved` → `asset.Location = domain.Location{WorldID: v.ToLocation}` ✓

---

## Remaining — Source Files

### `internal/faction/engine/mutation/mutation.go`

`case domain.MovementOrderIssued` (line ~192):
```
asset.Location = ""
```
→
```
asset.Location = domain.Location{}
```

`case domain.MovementOrderCompleted` (line ~218):
```
asset.Location = v.FinalLocation.WorldID
```
→
```
asset.Location = v.FinalLocation
```

---

### `internal/faction/engine/turn/bookkeeping.go`

`AssetRef` struct field (line ~22):
```
Location     string
```
→
```
Location     domain.Location
```

Line ~92 (`ref := AssetRef{...}`) — no change needed; `asset.Location` is already `domain.Location`.

---

### `internal/faction/engine/testharness/harness.go`

`AddFaction` (line ~110):
```
Location:     homeworld,
```
→
```
Location:     domain.Location{WorldID: homeworld},
```

`AddAssetOnWorld` (line ~228):
```
Location:     world,
```
→
```
Location:     domain.Location{WorldID: world},
```

---

### `internal/faction/engine/ability/steps/movement.go`

Line ~31 (`AssetMoved` struct literal):
```
FromLocation:      asset.Location,
```
→
```
FromLocation:      asset.Location.WorldID,
```

Lines ~43–44 (`worldsFromState`):
```
if asset.Location != "" {
    seen[asset.Location] = true
```
→
```
if asset.Location.WorldID != "" {
    seen[asset.Location.WorldID] = true
```

---

### `internal/faction/engine/ability/steps/faction_check.go`

Line ~21:
```
candidates := factionTestCandidates(factionState, faction.ID, asset.Location, step.Effect)
```
→
```
candidates := factionTestCandidates(factionState, faction.ID, asset.Location.WorldID, step.Effect)
```

Line ~51:
```
if asset.Location == world {
```
→
```
if asset.Location.WorldID == world {
```

Line ~74:
```
if targetAsset.Location == asset.Location && targetAsset.Stealthy {
```
→
```
if targetAsset.Location.WorldID == asset.Location.WorldID && targetAsset.Stealthy {
```

---

### `internal/faction/engine/action/actions/buy_asset.go`

`Resolve` (line ~90):
```
Location:     ba.buyOrder.World,
```
→
```
Location:     domain.Location{WorldID: ba.buyOrder.World},
```

`eligibleStealthTargets` (line ~123):
```
if asset.Location != world || asset.Stealthy {
```
→
```
if asset.Location.WorldID != world || asset.Stealthy {
```

`availableWorlds` (line ~139):
```
seen[asset.Location] = struct{}{}
```
→
```
seen[asset.Location.WorldID] = struct{}{}
```

---

### `internal/faction/engine/action/actions/attack.go`

Line ~79:
```
defenders := liveDefenders(faction.ID, attacker.Location, assetHPTracker, attack.index)
```
→
```
defenders := liveDefenders(faction.ID, attacker.Location.WorldID, assetHPTracker, attack.index)
```

Lines ~126, ~140, ~152 (`World:` field in `RollContext` / `tieCtx`):
```
World:     attacker.Location,
```
→
```
World:     attacker.Location.WorldID,
```

Line ~159:
```
base := factionBaseOnWorld(defenderFaction, attacker.Location)
```
→
```
base := factionBaseOnWorld(defenderFaction, attacker.Location.WorldID)
```

Line ~234 (`attackerHasTarget`):
```
return len(eligibleDefenders(faction.ID, attacker.Location, index)) > 0
```
→
```
return len(eligibleDefenders(faction.ID, attacker.Location.WorldID, index)) > 0
```

---

### `internal/faction/engine/action/actions/expand_influence.go`

`eligibleNewBaseWorlds` (line ~219):
```
if asset.Location != worldID {
```
→
```
if asset.Location.WorldID != worldID {
```

`worldsForNewBase` (line ~239):
```
worldsWithAssets[asset.Location] = struct{}{}
```
→
```
worldsWithAssets[asset.Location.WorldID] = struct{}{}
```

`rivalAssetsOnWorld` (line ~297):
```
if asset.Location == world && !asset.Stealthy && asset.Ready && asset.CurrentHP > 0 && asset.Maintained {
```
→
```
if asset.Location.WorldID == world && !asset.Stealthy && asset.Ready && asset.CurrentHP > 0 && asset.Maintained {
```

---

### `internal/faction/engine/action/actions/seize_planet.go`

`seizePlanetTargetWorlds` (line ~71):
```
factionFragments[asset.Location] = struct{}{}
```
→
```
factionFragments[asset.Location.WorldID] = struct{}{}
```

---

### `internal/faction/engine/goal/goals/helpers.go`

`factionHasUnstealthedAssetOn` (line ~107):
```
if asset.Location == world && !asset.Stealthy {
```
→
```
if asset.Location.WorldID == world && !asset.Stealthy {
```

---

### `internal/faction/engine/goal/goals/inside_enemy_territory.go`

Line ~30:
```
if !rivalHasPlanetaryGovernmentOnWorld(actingFaction.ID, asset.Location, factionState, index) {
```
→
```
if !rivalHasPlanetaryGovernmentOnWorld(actingFaction.ID, asset.Location.WorldID, factionState, index) {
```

---

### `internal/faction/engine/goal/goals/planetary_seizure.go`

Line ~71:
```
if asset.Location == goal.TargetWorld && !asset.Stealthy && !removedIDs[asset.ID] {
```
→
```
if asset.Location.WorldID == goal.TargetWorld && !asset.Stealthy && !removedIDs[asset.ID] {
```

---

### `internal/faction/engine/world/world.go`

Lines ~38, ~42:
```
if _, ok := engine.spatialMap.Location(asset.Location); !ok {
    continue
}
index.AssetsByLocation[asset.Location] = append(index.AssetsByLocation[asset.Location], asset)
```
→
```
if _, ok := engine.spatialMap.Location(asset.Location.WorldID); !ok {
    continue
}
index.AssetsByLocation[asset.Location.WorldID] = append(index.AssetsByLocation[asset.Location.WorldID], asset)
```

---

## Remaining — Test Files

**Pattern:** `Location: "world-id"` → `Location: domain.Location{WorldID: "world-id"}`

Apply this pattern to every `domain.Asset` struct literal in:

| File | Notes |
|---|---|
| `internal/faction/engine/mutation/mutation_test.go` | Lines ~18–19, `"Anchorage"` |
| `internal/faction/state/faction_state_test.go` | Line ~38, `"Tartarus"` |
| `internal/faction/engine/ability/ability_test.go` | Lines ~41, ~51, field is `assetLocation` (string var) |
| `internal/faction/engine/action/actions/attack_test.go` | Lines ~68, ~77 struct literals; line ~138 assignment `defenderAsset.Location = "Tartarus"` → `= domain.Location{WorldID: "Tartarus"}` |
| `internal/faction/engine/action/actions/buy_asset_test.go` | Line ~84 comparison: `!= "Tartarus"` → `!= (domain.Location{WorldID: "Tartarus"})`; line ~85 format `%q` → `%v` |
| `internal/faction/engine/action/actions/expand_influence_test.go` | Lines ~35, ~45, ~57, ~70, ~193, ~196 |
| `internal/faction/engine/action/actions/refit_asset_test.go` | Lines ~36, ~45, ~59; lines ~92–93 comparison + format same as buy_asset_test |
| `internal/faction/engine/action/actions/seize_planet_test.go` | Lines ~24, ~38, ~52, ~53, ~70, ~71 |
| `internal/faction/engine/action/actions/use_asset_ability_test.go` | Lines ~80, ~93, `"Anchorage"` |

**Index/helper files:**

`internal/faction/engine/action/actions/spatial_index_test.go`:
- `liveAsset` helper: `Location: location` → `Location: domain.Location{WorldID: location}`
- `indexWithAssets`: `idx.AssetsByLocation[asset.Location]` → `[asset.Location.WorldID]` (2 occurrences)

`internal/faction/engine/action/actions/test_helpers_test.go`:
- `indexFromState`: `idx.AssetsByLocation[asset.Location]` → `[asset.Location.WorldID]` (2 occurrences)

`internal/faction/engine/goal/goals/world_index_test.go`:
- Line ~14: `{ID: "a1", OwnerID: ownerID, Location: locationID}` → `Location: domain.Location{WorldID: locationID}`

---

## Verify

```
go build ./... && go test ./...
```

All tests must pass. No behavior changes — this is a pure type migration.

---

## Not Touched This Commit

- `refit_asset.go` line ~75 `Location: ra.refitOrder.OldAsset.Location` — already a struct copy, no change needed
- `base.Location` sites — Commit 2
- `faction.Homeworld` sites — Commit 3
- `AssetMoved.FromLocation/ToLocation` fields — still `string`; Commit 4 migrates them
- Narrative / digest files — Commit 4
