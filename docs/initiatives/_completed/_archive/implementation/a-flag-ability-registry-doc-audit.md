# A-flag Ability Registry — Pre-Merge Architecture Doc Audit

> Pre-merge step 8.2 reconciliation for `feature/a-flag-ability-registry` (A.001.2).
> Method: every changed `.go` source file mapped back to the architecture page that
> owns it via the page's `> **Code:**` header, then each owning page compared against
> current source. Files mapping to no page are test/generated infra. Pages whose owned
> paths did not change are provably unaffected (listed at the end).

## File → owning-doc matrix

Changed source (tests/generated excluded):

| Changed source | Owning doc(s) |
|---|---|
| `ability/{dispatch,stage,reveal_stealth,coin_delta,purchase_discount,purchase_enabler,relocate}.go` | ability-catalog, actions |
| `domain/asset.go` (ability shapes, `NewPurchasedAsset`, `CanPurchase`, `Clone`) | ability-catalog (shapes), actions, effect-mutation |
| `domain/mutation.go` (+`PurchaseDiscountGranted/Consumed`) | effect-mutation |
| `domain/faction.go` (`PendingPurchaseDiscount`, `Clone`, `StatScore`) | effect-mutation, actions |
| `mutation/mutation.go` (+2 apply arms) | effect-mutation |
| `action/collector.go` (+3 methods, −1) | actions |
| `actions/{buy_asset,register,use_asset_ability,refit,repair}.go` | actions |
| `world/world.go` (`BuildRelocateMutations`, `WorldsWithin/InRegion`) | world-movement |
| `spatial/region_map.go` (`DistancesFrom`) | spatial |
| `rulebook/rulebook.go` (step parser) | persistence (owner) + ability-catalog (parser ref) |
| `tui/views/turn/overlay/{relocateorders,purchaseenabler,ability}.go` | overlays |
| `tui/adapter/{action_collector,channels,smoke}.go` | event-stream |
| `tui/views/turn/execution/{execution,format}.go` | views |
| `testharness/*`, `mocks/*` | — (test infra; no doc) |

## Findings by severity

### 🔴 HEAVY — `engine/ability-catalog.md` (status: *Design*)

The F.022 design catalog, headed *"the registry it specifies is **not yet built**."* That
framing is now false for the entire A-side. Stale claims:

- **"Two coded exemplars"** (L36-44): calls Informers a *"bespoke Go handler in a
  `map[def.ID]`"* in `ability/informers.go`. Reality: effect-keyed
  `map[AbilityEffectType]EffectHandler`; `informers.go` no longer exists (it's
  `reveal_stealth.go`).
- **"Latent vocabulary… partly dead"** table (L47-60): every row wrong — `Die`/`Outcomes`
  live (`RollTable`), `[[ability.steps]]` is the canonical shape not "silently dropped,"
  `coin_drain`/`coin_steal` **deleted** (not "dispatched nowhere").
- **A-flag effects table** (L80-89): status column pre-build. Built now: `reveal_stealth`
  (Informers + Seductress), `coin_delta`, `purchase_discount`, `relocate`,
  `purchase_enabler`. Future: Marketers/Seditionists (→A.001.3), Monopoly (S-side).
- **`purchase-enabler` row** (L88): says *"buy one **discounted** Force asset"* — wrong,
  it's a **surcharge** (1.5×). Decision Record says *"catalog row fixed at pre-merge."*
- **"Drift evidence"** (L198-209): "1 real handler + 9 stubs," "Seductress silently
  no-ops" — now historical.

**Open decision (lifecycle fork):** the page is a "Design" arc artifact. Options —
(a) fix the flagged enabler row only + add a status banner pointing to subsystem pages
for as-built truth, keep the design body frozen until the arc closes;
(b) full A-side as-built rewrite now (S-side stays design);
(c) fix the flagged enabler row only, nothing else.
**Resolved (option a):** enabler row fixed + an "As-built (A.001.2)" status banner added
to `ability-catalog.md` pointing to actions/effect-mutation/world-movement for as-built
truth; the F.022 design body stays frozen until the arc closes.

### 🟠 MODERATE — `engine/world-movement.md` (status: *Written*)

Instant-relocate primitive entirely absent:
- No mention of `BuildRelocateMutations`, `WorldsWithin`, `WorldsInRegion`,
  `worldsMatching` — a second world-engine entry point, distinct from the drift-order
  `BuildMovementMutations`/`TickMovementOrders` it documents.
- **HexRouter seam** (L52-58) says *"just `Distance` and `Path`"* — now also
  `DistancesFrom` and `AllWorlds`.
- **Edit:** new "Instant relocation" subsection + 2 Key Decisions graduating plan
  Decision 3 (*instant, distinct from drift path*) and Decision 4 (*reuses `AssetMoved`,
  no new mutation*).

### 🟠 MODERATE — `engine/effect-mutation.md` (status: *Written*)

- Mutation catalogue (L136-165) **missing** `purchase_discount_granted` and
  `purchase_discount_consumed`. "Exhaustive by panic" claim still holds; the list doesn't.
- `asset_moved` (L150) now also emitted by ability relocate (`Cause: "ability"`) — note.
- L70 "P is a buy-time government-permission check" now backed by graduated domain
  (`HasPlanetaryGovernment`/`CanPurchase`) — accurate, no change.

### 🟡 MINOR — `engine/actions.md` (status: *Written*)

- "**sixteen** action-specific methods" (L71): collector dropped `SelectFactionTestTarget`,
  added `SelectAbilityTarget` + `SelectRelocateOrders` + `SelectPurchaseEnablerAsset` →
  now **eighteen**.
- "**The ability fold**" (L85-89) understates Dispatch — now a **step pipeline** (target →
  gate → effect, halt-on-failure, cross-step working clone). Existing "A-dispatch keys on
  the effect" Key Decision (L144) accurate but predates the step model; augment.
- Buy Asset row (L104): now conditionally emits `PurchaseDiscountConsumed`; Buy/enabler
  share `domain.NewPurchasedAsset`/`CanPurchase` — durable "two purchase paths can't
  drift" decision worth a line.

### 🟡 MINOR — `engine/spatial.md` (status: *Written*)

- "the only **three** questions" (L6-8) + Pathing (L60-71): `DistancesFrom` is a new
  sibling to `Distance`/`Path` (one Dijkstra pass, all reachable hexes). One sentence.
  (`AllWorlds` pre-existed; only `DistancesFrom` is new here.)

### 🟢 LOW — `interface/overlays.md`

- Two new bespoke overlays (`RelocateOrders`, `PurchaseEnabler`) follow the documented
  complete-and-emit pattern; "~18 overlays" (L59) has a tilde. Optional: add to bespoke
  list. No structural/Key-Decision drift.

### ✅ NO CHANGE — verified, not assumed

- **`persistence.md`** — owns `rulebook/` (+203). Its only ability reference (L91,
  "unknown ability effect is a load-time error") **remains true** (`toAbilityEffect` still
  errors). Parser internals not enumerated.
- **`interface/event-stream.md`** — owns the adapter (changed). Describes the ask
  mechanism narratively; does **not** name collector methods or `AskKind`s, so the
  rename/additions drift no text.
- **`interface/views.md`** — owns `execution/` (changed). Describes three-band view +
  snapshot rendering; does **not** enumerate `formatMutation` arms, so new `AssetMoved`
  formatting is invisible to it.
- **`orchestrator.md`** — `orchestrator.go` unchanged; world-engine threading is in
  `register.go` (actions); `Engine.World` already existed.
- **`engine/overview.md`** — cycle/conventions unchanged; only the page-index "Design"
  status for the catalog is in play (tied to the fork above).
- **`goals.md`, `hooks.md`, `logging-errors.md`, `narrative-digest.md`,
  `turn-pipeline.md`, `interface/{regions,state-machine,overview}.md`,
  `architecture-overview.md`** — no owned paths in the changed-file set.

## Pre-merge step 8 — beyond reconciliation

- **8.2 graduation candidates** (plan Decision Record → arch Key Decisions):
  relocate-is-instant + reuses-`AssetMoved` (→ world-movement); per-step pipeline +
  halt-on-failure + world-engine threading (→ actions); shared purchase path
  (→ actions/effect-mutation); enumerated-payouts-no-formula (→ catalog *if* updated).
- **8.3** planned-work.md: log `F.024` (relocate back-nav, review C11); confirm A.001.2
  items struck.
- **8.4** move discovery + 5 plan docs to `completed/`.
