package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// counterstrikeRegistrar is Cracked Comms: after a successful defense, its owner
// may turn the attacking asset on itself. The Attack action asks and rolls.
var counterstrikeRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnDefenseWin: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Matchup: func(effect.MatchupContext) effect.MatchupOutcome {
		return effect.MatchupOutcome{Counterstrike: true}
	},
}
