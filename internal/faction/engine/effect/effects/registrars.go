package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

func Registrars() map[domain.SpecialEffectType]effect.Registrar {
	return map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialCoinGain:       coinGainRegistrar,
		domain.SpecialCoinSteal:      coinStealRegistrar,
		domain.SpecialSelfDamage:     selfDamageRegistrar,
		domain.SpecialIgnoreCoinLoss: ignoreCoinLossRegistrar,
		domain.SpecialClearStealth:   clearStealthRegistrar,
		domain.SpecialBonusDie:       bonusDieRegistrar,
		domain.SpecialRemoveTagDice:  removeTagDiceRegistrar,
		domain.SpecialReroll:         rerollRegistrar,
		domain.SpecialCostSurcharge:  costSurchargeRegistrar,
		domain.SpecialTechLevelFloor: techLevelFloorRegistrar,
		domain.SpecialRestrictAttack: restrictAttackRegistrar,
		domain.SpecialGrantMovement:  grantMovementRegistrar,
		domain.SpecialApplyStealth:   applyStealthRegistrar,
		domain.SpecialCaptureTarget:  captureTargetRegistrar,
		domain.SpecialBlockAbility:   blockAbilityRegistrar,
		domain.SpecialCounterstrike:  counterstrikeRegistrar,
	}
}
