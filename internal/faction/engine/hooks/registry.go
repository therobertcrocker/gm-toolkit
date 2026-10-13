package hooks

// Registered* types pair a hook implementation with its registration metadata.
// They live here because each carries a Scope as its primary identity field.

type RegisteredRollModifier struct {
	Scope  Scope
	Source string
	Hook   RollModifier
}

type RegisteredRollResultHook struct {
	Scope  Scope
	Source string
	Hook   RollResultHook
}

type RegisteredRollOfferFilter struct {
	Scope  Scope
	Source string
	Hook   RollOfferFilter
}

type RegisteredMutationReactor struct {
	Scope  Scope
	Source string
	Hook   MutationReactor
}

type RegisteredAssetCostModifier struct {
	Scope  Scope
	Source string
	Hook   AssetCostModifier
}

type RegisteredMaintenanceCostModifier struct {
	Scope  Scope
	Source string
	Hook   MaintenanceCostModifier
}

type RegisteredWorldTechLevelModifier struct {
	Scope  Scope
	Source string
	Hook   WorldTechLevelModifier
}

type RegisteredAttackEligibilityRule struct {
	Scope  Scope
	Source string
	Hook   AttackEligibilityRule
}

type RegisteredAssetMovementGranter struct {
	Scope  Scope
	Source string
	Hook   AssetMovementGranter
}

type RegisteredTieResolver struct {
	Scope  Scope
	Source string
	Hook   TieResolver
}

// Registry holds all registered hooks keyed by Scope, one map per category.
type Registry struct {
	rollModifiers            map[Scope][]RegisteredRollModifier
	rollResultHooks          map[Scope][]RegisteredRollResultHook
	rollOfferFilters         map[Scope][]RegisteredRollOfferFilter
	mutationReactors         map[Scope][]RegisteredMutationReactor
	assetCostModifiers       map[Scope][]RegisteredAssetCostModifier
	maintenanceCostModifiers map[Scope][]RegisteredMaintenanceCostModifier
	worldTechLevelModifiers  map[Scope][]RegisteredWorldTechLevelModifier
	attackEligibilityRules   map[Scope][]RegisteredAttackEligibilityRule
	assetMovementGranters    map[Scope][]RegisteredAssetMovementGranter
	tieResolvers             map[Scope][]RegisteredTieResolver
}

// NewRegistry returns an empty Registry ready for registration.
func NewRegistry() *Registry {
	return &Registry{
		rollModifiers:            make(map[Scope][]RegisteredRollModifier),
		rollResultHooks:          make(map[Scope][]RegisteredRollResultHook),
		rollOfferFilters:         make(map[Scope][]RegisteredRollOfferFilter),
		mutationReactors:         make(map[Scope][]RegisteredMutationReactor),
		assetCostModifiers:       make(map[Scope][]RegisteredAssetCostModifier),
		maintenanceCostModifiers: make(map[Scope][]RegisteredMaintenanceCostModifier),
		worldTechLevelModifiers:  make(map[Scope][]RegisteredWorldTechLevelModifier),
		attackEligibilityRules:   make(map[Scope][]RegisteredAttackEligibilityRule),
		assetMovementGranters:    make(map[Scope][]RegisteredAssetMovementGranter),
		tieResolvers:             make(map[Scope][]RegisteredTieResolver),
	}
}

// RegisterRollModifier registers a Cat 1 hook under the given scope.
func (registry *Registry) RegisterRollModifier(scope Scope, source string, hook RollModifier) {
	registry.rollModifiers[scope] = append(registry.rollModifiers[scope], RegisteredRollModifier{Scope: scope, Source: source, Hook: hook})
}

// RegisterRollResultHook registers a Cat 2 hook under the given scope.
func (registry *Registry) RegisterRollResultHook(scope Scope, source string, hook RollResultHook) {
	registry.rollResultHooks[scope] = append(registry.rollResultHooks[scope], RegisteredRollResultHook{Scope: scope, Source: source, Hook: hook})
}

// RegisterRollOfferFilter registers a Cat 1 offer filter under the given scope.
func (registry *Registry) RegisterRollOfferFilter(scope Scope, source string, hook RollOfferFilter) {
	registry.rollOfferFilters[scope] = append(registry.rollOfferFilters[scope], RegisteredRollOfferFilter{Scope: scope, Source: source, Hook: hook})
}

// RegisterMutationReactor registers a Cat 3 hook under the given scope.
func (registry *Registry) RegisterMutationReactor(scope Scope, source string, hook MutationReactor) {
	registry.mutationReactors[scope] = append(registry.mutationReactors[scope], RegisteredMutationReactor{Scope: scope, Source: source, Hook: hook})
}

// RegisterAssetCostModifier registers a Cat 4 asset-cost hook under the given scope.
func (registry *Registry) RegisterAssetCostModifier(scope Scope, source string, hook AssetCostModifier) {
	registry.assetCostModifiers[scope] = append(registry.assetCostModifiers[scope], RegisteredAssetCostModifier{Scope: scope, Source: source, Hook: hook})
}

// RegisterMaintenanceCostModifier registers a Cat 4 maintenance-cost hook under the given scope.
func (registry *Registry) RegisterMaintenanceCostModifier(scope Scope, source string, hook MaintenanceCostModifier) {
	registry.maintenanceCostModifiers[scope] = append(registry.maintenanceCostModifiers[scope], RegisteredMaintenanceCostModifier{Scope: scope, Source: source, Hook: hook})
}

// RegisterWorldTechLevelModifier registers a Cat 4 tech-level hook under the given scope.
func (registry *Registry) RegisterWorldTechLevelModifier(scope Scope, source string, hook WorldTechLevelModifier) {
	registry.worldTechLevelModifiers[scope] = append(registry.worldTechLevelModifiers[scope], RegisteredWorldTechLevelModifier{Scope: scope, Source: source, Hook: hook})
}

// RegisterAttackEligibilityRule registers a Cat 4 attack-eligibility hook under the given scope.
func (registry *Registry) RegisterAttackEligibilityRule(scope Scope, source string, hook AttackEligibilityRule) {
	registry.attackEligibilityRules[scope] = append(registry.attackEligibilityRules[scope], RegisteredAttackEligibilityRule{Scope: scope, Source: source, Hook: hook})
}

// RegisterAssetMovementGranter registers a Cat 4 movement-granter hook under the given scope.
func (registry *Registry) RegisterAssetMovementGranter(scope Scope, source string, hook AssetMovementGranter) {
	registry.assetMovementGranters[scope] = append(registry.assetMovementGranters[scope], RegisteredAssetMovementGranter{Scope: scope, Source: source, Hook: hook})
}

// RegisterTieResolver registers a Cat 5 hook under the given scope.
func (registry *Registry) RegisterTieResolver(scope Scope, source string, hook TieResolver) {
	registry.tieResolvers[scope] = append(registry.tieResolvers[scope], RegisteredTieResolver{Scope: scope, Source: source, Hook: hook})
}

// RollModifiersAt returns the Cat 1 hooks registered in one scope bucket. The roll
// dispatcher decides which buckets a roll consults.
func (registry *Registry) RollModifiersAt(scope Scope) []RegisteredRollModifier {
	return registry.rollModifiers[scope]
}

// RollResultHooksAt returns the Cat 2 hooks registered in one scope bucket.
func (registry *Registry) RollResultHooksAt(scope Scope) []RegisteredRollResultHook {
	return registry.rollResultHooks[scope]
}

// RollOfferFiltersAt returns the Cat 1 offer filters registered in one scope bucket.
func (registry *Registry) RollOfferFiltersAt(scope Scope) []RegisteredRollOfferFilter {
	return registry.rollOfferFilters[scope]
}

// MutationReactorsGlobal returns the Cat 3 hooks registered at global scope.
func (registry *Registry) MutationReactorsGlobal() []RegisteredMutationReactor {
	return registry.mutationReactors[GlobalScope()]
}

// MutationReactorsForFaction returns the Cat 3 hooks registered against one faction.
func (registry *Registry) MutationReactorsForFaction(factionID string) []RegisteredMutationReactor {
	return registry.mutationReactors[FactionScope(factionID)]
}

// MutationReactorsForAsset returns the Cat 3 hooks registered against one asset instance.
func (registry *Registry) MutationReactorsForAsset(factionID, assetInstanceID string) []RegisteredMutationReactor {
	return registry.mutationReactors[AssetScope(factionID, assetInstanceID)]
}

// AssetCostModifiersFor returns all Cat 4 asset-cost hooks matching the given faction and asset.
func (registry *Registry) AssetCostModifiersFor(factionID, assetInstanceID string) []RegisteredAssetCostModifier {
	var result []RegisteredAssetCostModifier
	result = append(result, registry.assetCostModifiers[GlobalScope()]...)
	result = append(result, registry.assetCostModifiers[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.assetCostModifiers[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}

// AssetCostModifiersForAsset returns the Cat 4 asset-cost hooks registered
// against one asset instance, without the global or faction buckets.
func (registry *Registry) AssetCostModifiersForAsset(factionID, assetInstanceID string) []RegisteredAssetCostModifier {
	return registry.assetCostModifiers[AssetScope(factionID, assetInstanceID)]
}

// MaintenanceCostModifiersFor returns all Cat 4 maintenance-cost hooks matching the given faction and asset.
func (registry *Registry) MaintenanceCostModifiersFor(factionID, assetInstanceID string) []RegisteredMaintenanceCostModifier {
	var result []RegisteredMaintenanceCostModifier
	result = append(result, registry.maintenanceCostModifiers[GlobalScope()]...)
	result = append(result, registry.maintenanceCostModifiers[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.maintenanceCostModifiers[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}

// WorldTechLevelModifiersFor returns all Cat 4 tech-level hooks matching the given faction and asset.
func (registry *Registry) WorldTechLevelModifiersFor(factionID, assetInstanceID string) []RegisteredWorldTechLevelModifier {
	var result []RegisteredWorldTechLevelModifier
	result = append(result, registry.worldTechLevelModifiers[GlobalScope()]...)
	result = append(result, registry.worldTechLevelModifiers[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.worldTechLevelModifiers[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}

// WorldTechLevelModifiersForAsset returns the Cat 4 tech-level hooks registered
// against one asset instance, without the global or faction buckets.
func (registry *Registry) WorldTechLevelModifiersForAsset(factionID, assetInstanceID string) []RegisteredWorldTechLevelModifier {
	return registry.worldTechLevelModifiers[AssetScope(factionID, assetInstanceID)]
}

// AttackEligibilityRulesFor returns all Cat 4 attack-eligibility hooks matching the given faction and asset.
func (registry *Registry) AttackEligibilityRulesFor(factionID, assetInstanceID string) []RegisteredAttackEligibilityRule {
	var result []RegisteredAttackEligibilityRule
	result = append(result, registry.attackEligibilityRules[GlobalScope()]...)
	result = append(result, registry.attackEligibilityRules[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.attackEligibilityRules[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}

// AttackEligibilityRulesForAsset returns the Cat 4 attack-eligibility hooks
// registered against one asset instance, without the global or faction buckets.
func (registry *Registry) AttackEligibilityRulesForAsset(factionID, assetInstanceID string) []RegisteredAttackEligibilityRule {
	return registry.attackEligibilityRules[AssetScope(factionID, assetInstanceID)]
}

// AssetMovementGrantersFor returns all Cat 4 movement-granter hooks matching the given faction and asset.
func (registry *Registry) AssetMovementGrantersFor(factionID, assetInstanceID string) []RegisteredAssetMovementGranter {
	var result []RegisteredAssetMovementGranter
	result = append(result, registry.assetMovementGranters[GlobalScope()]...)
	result = append(result, registry.assetMovementGranters[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.assetMovementGranters[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}

// TieResolversFor returns all Cat 5 hooks matching the given faction and asset.
func (registry *Registry) TieResolversFor(factionID, assetInstanceID string) []RegisteredTieResolver {
	var result []RegisteredTieResolver
	result = append(result, registry.tieResolvers[GlobalScope()]...)
	result = append(result, registry.tieResolvers[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.tieResolvers[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}
