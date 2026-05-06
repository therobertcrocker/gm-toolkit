// Package eventhooks defines the five reactive-mechanic hook families and the
// scoped registry that binds them.
//
// # The five categories
//
//	Cat 1  RollModifier          pre-roll dice augmentation       interfaces.go
//	Cat 2  RollResultHook        post-roll inspection & rerolls   interfaces.go
//	Cat 3  MutationReactor       post-action mutation reactions   interfaces.go
//	Cat 4  RuleModifier family   query-time rule alterations      rule_modifier.go
//	Cat 5  TieResolver           tie-point rule overrides         interfaces.go
//
// Cat 4 is a typed family (AssetCostModifier, MaintenanceCostModifier,
// WorldTechLevelModifier, AssetMovementGranter) rather than a single interface,
// because each question has a different typed signature and a generic interface
// would lose compile-time safety.
//
// # Registration and dispatch
//
// Registration happens once at engine startup. The Tag Engine and Effects Engine
// call Register* for every hook their data produces; the Core Engine's dispatchers
// consult the registry at the appropriate turn pipeline point.
//
// Hooks are scoped to a faction, a specific asset instance, or globally.
// Lookup returns global hooks first, then faction-scoped, then asset-scoped —
// all in registration order within each bucket.
//
// Only Cat 3 (MutationReactor) can recurse; recursion is bounded at depth 5.
package eventhooks
