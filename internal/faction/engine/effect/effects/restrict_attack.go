package effects

import (
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// restrictAttackRegistrar reads owner = "rival" as "the opponent belongs to
// another faction", and scope = "self" as "the declaring asset is in the fight".
// side picks the role the restriction answers for: defense when the asset is
// struck, attack_and_counter when it strikes. The filters name the opponent.
var restrictAttackRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerAttackEligibilityQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeSelf},
			Filters: []domain.EffectFilterName{domain.FilterAssetType, domain.FilterCategory, domain.FilterAttribute},
		},
	},
	Register: registerRestrictAttack,
	Validate: validateRestrictAttack,
}

// validateRestrictAttack refuses attribute with attack_and_counter: the attribute is
// read from the striker's attack profile, and on that side the striker is the
// declaring asset, so the filter would test the asset's own fixed profile.
func validateRestrictAttack(entry domain.EffectDefinition, _ effect.HolderKind) error {
	if entry.RestrictAttack != nil && entry.RestrictAttack.Side == domain.RestrictAttackAndCounter && entry.Filters.Attribute != "" {
		return fmt.Errorf("filter %q is not allowed with side %q", domain.FilterAttribute, domain.RestrictAttackAndCounter)
	}
	return nil
}

func registerRestrictAttack(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAttackEligibilityRule(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialRestrictAttack),
		&restrictAttackRule{
			owner:      ctx.Holder.Faction,
			holder:     refOf(ctx.Holder),
			side:       ctx.Entry.RestrictAttack.Side,
			allow:      ctx.Entry.RestrictAttack.Allow,
			assetTypes: ctx.Entry.Filters.AssetTypes,
			categories: ctx.Entry.Filters.Categories,
			attribute:  ctx.Entry.Filters.Attribute,
		},
	)
}

// restrictAttackRule holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Decision Record 19).
type restrictAttackRule struct {
	owner      *domain.Faction
	holder     holderRef
	side       domain.RestrictSide
	allow      domain.RestrictAllow
	assetTypes []domain.AssetType
	categories []domain.FactionStat
	attribute  domain.FactionStat
}

func (rule *restrictAttackRule) MayFight(query hooks.AttackEligibilityQuery) bool {
	if _, live := rule.holder.liveIn(rule.owner); !live {
		return true
	}
	opponentDef, answers := rule.opponentDef(query)
	if !answers {
		return true
	}
	return rule.permits(query, opponentDef)
}

// opponentDef returns the other asset's definition when this rule answers the
// query: its asset holds the role its side names, and the other asset is a rival's.
func (rule *restrictAttackRule) opponentDef(query hooks.AttackEligibilityQuery) (*domain.AssetDefinition, bool) {
	var self, opponent *domain.Asset
	var opponentDef *domain.AssetDefinition
	switch {
	case rule.side == domain.RestrictDefense && query.Side == hooks.SideDefense:
		self, opponent, opponentDef = query.Defender, query.Attacker, query.AttackerDef
	case rule.side == domain.RestrictAttackAndCounter && query.Side != hooks.SideDefense:
		self, opponent, opponentDef = query.Attacker, query.Defender, query.DefenderDef
	default:
		return nil, false
	}
	if self.OwnerID != rule.owner.ID || self.ID != rule.holder.assetID || opponent.OwnerID == rule.owner.ID {
		return nil, false
	}
	return opponentDef, true
}

// permits applies the filters to the opponent and the allow direction to the result.
func (rule *restrictAttackRule) permits(query hooks.AttackEligibilityQuery, opponentDef *domain.AssetDefinition) bool {
	matched := rule.matches(query, opponentDef)
	if rule.allow == domain.AllowOnlyListed {
		return matched
	}
	return !matched
}

// matches reports whether every filter the rule sets holds for this fight. The
// attribute belongs to the attack, so it is read from the striker's profile; a
// striker with no attack profile has no attribute and does not match.
func (rule *restrictAttackRule) matches(query hooks.AttackEligibilityQuery, opponentDef *domain.AssetDefinition) bool {
	if len(rule.assetTypes) > 0 && !slices.Contains(rule.assetTypes, opponentDef.Type) {
		return false
	}
	if len(rule.categories) > 0 && !slices.Contains(rule.categories, opponentDef.Category) {
		return false
	}
	if rule.attribute != "" && (query.AttackerDef.Attack == nil || query.AttackerDef.Attack.DefenderStat != rule.attribute) {
		return false
	}
	return true
}
