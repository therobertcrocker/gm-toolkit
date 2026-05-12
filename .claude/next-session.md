# Asset Movement Redesign — Plan Session

This is the Plan session for Initiative #2 (Asset Movement Redesign) in `docs/dev_journals/faction-manager/planned-work.md`. Discovery is complete; the discovery doc is at `docs/discovery/asset-movement-redesign-discovery.md` — read it first; it's the source of truth for every design decision.

**Your deliverable** is an implementation plan in `docs/implementation/asset-movement-redesign-plan.md`, broken into discrete phases (each phase = its own execution session). No code edits — Plan mode only.

## Key inputs already settled in discovery

- New Movement Phase between Phase 2B (Stat Raise) and Phase 3 (Action) in `engine/orchestrator.go`
- `Asset.Location` migrates from `string` to a `{WorldID, HexCoords}` struct
- `AssetDefinition` gains `Speed` (DriftRating already exists)
- `MovementOrder` lives on `Asset.CurrentOrder`; stores pre-computed path
- Self-movement ability steps removed; transport-pattern abilities retained (likely as a new `AbilityStepTransport` type)
- `ChangeHomeworld` stays a goal — timer becomes `spatial.Distance(homeworld, target, drift=3)`
- Cargo co-located with transport, individually targetable; attack targeting extends to hex-level lookup
- This redesign retires Spatial Phase 3 Commit 3 and Phase 4 from `docs/implementation/spatial-model-effort-2-plan.md`

## Open questions to ratify during planning

The discovery doc lists eight; all need resolution before phase breakdown is final:

1. Mid-flight revision cost
2. World-graph mutation handling
3. First-tick `WorldID` clear timing
4. `Faction.Homeworld` shape (recommendation: stay as `string`)
5. Goal-locked factions issuing orders
6. Transport step schema (recommendation: distinct `AbilityStepTransport` type)
7. Mixed-mode ability reconciliation cost (Smugglers, Blockade Runner)
8. Cross-faction same-hex resolution order

## Phasing guidance

Likely shape, for the Plan session to refine:

1. Domain types (`Location` struct, `MovementOrder`, `Speed`, `PhaseMovement`) + migrations
2. Spatial pathfinding primitive (path return, not just total distance) if not already exposed
3. Movement Phase in orchestrator + collector interface additions
4. Order lifecycle mutations + state-application
5. Transport ability schema + handler
6. Ability-step removal + asset TOML migration
7. `ChangeHomeworld` reconciliation
8. Hex-level attack targeting extension

## Session discipline

- **Model:** Opus (Plan session — always).
- Per `CLAUDE.md` and the `feedback_session_boundaries` memory: end the session at the plan file. Do not start execution work in this session, even if Robert asks — open a fresh Sonnet execution session per phase.
- Update `planned-work.md` status (Pending Discovery → Pending Plan, then → In-Progress once plan lands) at the appropriate mode boundary.
