package effects_test

import (
	"log/slog"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

var (
	planetaryDefensesDef = &domain.AssetDefinition{ID: "SWN-F6-002", Category: domain.StatForce, Type: domain.TypeFacility}
	lawyersDef           = &domain.AssetDefinition{ID: "SWN-W2-002", Category: domain.StatWealth, Type: domain.TypeSpecialForces,
		Attack: &domain.AttackProfile{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth}}

	starshipDef = &domain.AssetDefinition{ID: "starship", Category: domain.StatForce, Type: domain.TypeStarship,
		Attack: &domain.AttackProfile{AttackerStat: domain.StatForce, DefenderStat: domain.StatForce}}
	infantryDef = &domain.AssetDefinition{ID: "infantry", Category: domain.StatForce, Type: domain.TypeMilitaryUnit,
		Attack: &domain.AttackProfile{AttackerStat: domain.StatForce, DefenderStat: domain.StatForce}}
	infiltratorDef = &domain.AssetDefinition{ID: "infiltrator", Category: domain.StatCunning, Type: domain.TypeSpecialForces,
		Attack: &domain.AttackProfile{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning}}
	bankDef = &domain.AssetDefinition{ID: "bank", Category: domain.StatWealth, Type: domain.TypeFacility}
)

func planetaryDefensesEntry() domain.EffectDefinition {
	return restrictEntry(domain.RestrictDefense, domain.AllowOnlyListed,
		domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeSelf, AssetTypes: []domain.AssetType{domain.TypeStarship}})
}

func lawyersEntry() domain.EffectDefinition {
	return restrictEntry(domain.RestrictAttackAndCounter, domain.AllowExceptListed,
		domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeSelf, Categories: []domain.FactionStat{domain.StatForce}})
}

func integralProtocolsEntry() domain.EffectDefinition {
	return restrictEntry(domain.RestrictDefense, domain.AllowOnlyListed,
		domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeSelf, Attribute: domain.StatCunning})
}

func restrictEntry(side domain.RestrictSide, allow domain.RestrictAllow, filters domain.EffectFilters) domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger:        domain.TriggerAttackEligibilityQuery,
		Effect:         domain.SpecialRestrictAttack,
		Filters:        filters,
		RestrictAttack: &domain.RestrictAttackParams{Side: side, Allow: allow},
	}
}

// restriction registers entry on alpha's asset "r1" and returns the registry and
// that asset. Opponents are built with rival or ownAsset.
func restriction(def *domain.AssetDefinition, entry domain.EffectDefinition) (*hooks.Registry, *domain.Asset) {
	asset := activeAsset("r1", def.ID)
	alpha := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{asset.ID: asset}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialRestrictAttack].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Asset: asset, Def: def}, Entry: entry, Registry: registry,
	})
	return registry, asset
}

func rival(def *domain.AssetDefinition) *domain.Asset {
	return &domain.Asset{ID: "b1", DefinitionID: def.ID, OwnerID: "beta", CurrentHP: 4, Ready: true, Maintained: true}
}

func fight(registry *hooks.Registry, attacker *domain.Asset, attackerDef *domain.AssetDefinition, defender *domain.Asset, defenderDef *domain.AssetDefinition, side hooks.AttackSide) bool {
	return dispatch.ResolveAttackEligibility(registry, hooks.AttackEligibilityQuery{
		Attacker: attacker, AttackerDef: attackerDef, Defender: defender, DefenderDef: defenderDef, Side: side,
	})
}

func TestRestrictAttack_ValidateRefusesAttributeWhenStriking(t *testing.T) {
	cases := map[string]struct {
		side    domain.RestrictSide
		wantErr bool
	}{
		"attack_and_counter refuses attribute": {side: domain.RestrictAttackAndCounter, wantErr: true},
		"defense accepts attribute":            {side: domain.RestrictDefense, wantErr: false},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			entry := restrictEntry(testCase.side, domain.AllowOnlyListed,
				domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeSelf, Attribute: domain.StatCunning})
			def := &domain.AssetDefinition{ID: "restricted", Effects: []domain.EffectDefinition{entry}}
			err := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler)).RegisterEffectsList(def)
			if (err != nil) != testCase.wantErr {
				t.Errorf("RegisterEffectsList error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

func TestRestrictAttack_PlanetaryDefensesDefendsAgainstStarship(t *testing.T) {
	registry, defenses := restriction(planetaryDefensesDef, planetaryDefensesEntry())
	if !fight(registry, rival(starshipDef), starshipDef, defenses, planetaryDefensesDef, hooks.SideAttack) {
		t.Error("a Starship may attack Planetary Defenses")
	}
}

func TestRestrictAttack_PlanetaryDefensesRefusesNonStarship(t *testing.T) {
	registry, defenses := restriction(planetaryDefensesDef, planetaryDefensesEntry())
	if fight(registry, rival(infantryDef), infantryDef, defenses, planetaryDefensesDef, hooks.SideAttack) {
		t.Error("a Military Unit may not attack Planetary Defenses (asset_type, only_listed)")
	}
}

func TestRestrictAttack_LawyersRefusesForceTarget(t *testing.T) {
	registry, lawyers := restriction(lawyersDef, lawyersEntry())
	if fight(registry, lawyers, lawyersDef, rival(infantryDef), infantryDef, hooks.SideAttack) {
		t.Error("Lawyers may not attack a Force asset (category, except_listed)")
	}
}

func TestRestrictAttack_LawyersAttacksOtherCategory(t *testing.T) {
	registry, lawyers := restriction(lawyersDef, lawyersEntry())
	if !fight(registry, lawyers, lawyersDef, rival(bankDef), bankDef, hooks.SideAttack) {
		t.Error("Lawyers may attack a Wealth asset")
	}
}

func TestRestrictAttack_LawyersRefusesForceCounter(t *testing.T) {
	registry, lawyers := restriction(lawyersDef, lawyersEntry())
	if fight(registry, lawyers, lawyersDef, rival(infantryDef), infantryDef, hooks.SideCounter) {
		t.Error("Lawyers may not counterattack a Force asset")
	}
}

func TestRestrictAttack_IntegralProtocolsDefendsAgainstCunning(t *testing.T) {
	registry, protocols := restriction(integralProtocolsDef, integralProtocolsEntry())
	if !fight(registry, rival(infiltratorDef), infiltratorDef, protocols, integralProtocolsDef, hooks.SideAttack) {
		t.Error("an attack versus Cunning may target Integral Protocols")
	}
}

func TestRestrictAttack_IntegralProtocolsRefusesOtherAttribute(t *testing.T) {
	registry, protocols := restriction(integralProtocolsDef, integralProtocolsEntry())
	if fight(registry, rival(infantryDef), infantryDef, protocols, integralProtocolsDef, hooks.SideAttack) {
		t.Error("an attack versus Force may not target Integral Protocols (attribute, only_listed)")
	}
}

func TestRestrictAttack_DefenseSideIgnoredWhenStriking(t *testing.T) {
	registry, defenses := restriction(planetaryDefensesDef, planetaryDefensesEntry())
	if !fight(registry, defenses, planetaryDefensesDef, rival(infantryDef), infantryDef, hooks.SideCounter) {
		t.Error("a defense-side restriction must not limit its asset's own counterattack")
	}
}

func TestRestrictAttack_AttackSideIgnoredWhenDefending(t *testing.T) {
	registry, lawyers := restriction(lawyersDef, lawyersEntry())
	if !fight(registry, rival(infantryDef), infantryDef, lawyers, lawyersDef, hooks.SideAttack) {
		t.Error("an attack_and_counter restriction must not stop a Force asset attacking Lawyers")
	}
}

func TestRestrictAttack_IgnoresFightItIsNotIn(t *testing.T) {
	registry, _ := restriction(lawyersDef, lawyersEntry())
	sibling := activeAsset("r2", lawyersDef.ID)
	rule := registry.AttackEligibilityRulesForAsset("alpha", "r1")[0].Hook
	if !rule.MayFight(hooks.AttackEligibilityQuery{
		Attacker: sibling, AttackerDef: lawyersDef, Defender: rival(infantryDef), DefenderDef: infantryDef, Side: hooks.SideAttack,
	}) {
		t.Error("scope = self: a restriction must not answer for another asset's fight")
	}
}

func TestRestrictAttack_IgnoresOwnFactionOpponent(t *testing.T) {
	registry, lawyers := restriction(lawyersDef, lawyersEntry())
	ownForce := activeAsset("r3", infantryDef.ID)
	if !fight(registry, lawyers, lawyersDef, ownForce, infantryDef, hooks.SideAttack) {
		t.Error("owner = rival: a restriction must not answer against its own faction's asset")
	}
}

func TestRestrictAttack_InactiveAssetRestrictsNothing(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			registry, lawyers := restriction(lawyersDef, lawyersEntry())
			deactivate(lawyers)
			if !fight(registry, lawyers, lawyersDef, rival(infantryDef), infantryDef, hooks.SideAttack) {
				t.Error("an inactive Lawyers restricts nothing")
			}
		})
	}
}
