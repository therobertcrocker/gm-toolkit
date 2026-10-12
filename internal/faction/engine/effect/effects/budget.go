package effects

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// spendBudget reports whether key is unspent on owner's HookBudgets this turn,
// and spends it when it is. The budget lives on the owner's map, whatever the
// target, so it resets in the owner's own bookkeeping.
func spendBudget(owner *domain.Faction, key string) bool {
	if owner.BudgetSpent(key) {
		return false
	}
	owner.SpendBudget(key)
	return true
}

// budgetKey names the once_per budget an entry spends. The effect name, trigger
// and definition ID keep two effects, two entries of one definition, or two asset
// kinds from sharing a budget.
func budgetKey(oncePer domain.OncePerKey, effectName domain.SpecialEffectType, trigger domain.EffectTrigger, defID, assetID, targetFactionID string) string {
	base := "effect:" + string(effectName) + ":" + string(trigger) + ":" + defID
	switch oncePer {
	case domain.OncePerTargetFaction:
		return base + ":" + targetFactionID
	case domain.OncePerAsset:
		return base + ":" + assetID
	default:
		return base
	}
}
