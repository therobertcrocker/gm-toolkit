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
	}
}
