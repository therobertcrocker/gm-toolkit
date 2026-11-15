package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// coinUpkeepRegistrar is a tag's turn upkeep: bookkeeping charges it after income
// and before asset maintenance, so tags are paid first (A.001.5 Effort 2 Decision
// 10). owner = "own" reads as "the holder pays", scope = "faction" as "once for the
// faction".
var coinUpkeepRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerUpkeep: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerCoinUpkeep,
	Validate: validateCoinUpkeep,
}

func validateCoinUpkeep(_ domain.EffectDefinition, holder effect.HolderKind) error {
	if holder != effect.HolderTag {
		return fmt.Errorf("only a tag pays upkeep; an asset pays maintenance")
	}
	return nil
}

func registerCoinUpkeep(ctx effect.RegisterContext) {
	ctx.Registry.RegisterUpkeepCharger(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialCoinUpkeep),
		&coinUpkeepCharger{tagID: ctx.Holder.Tag.ID, params: *ctx.Entry.CoinUpkeep},
	)
}

type coinUpkeepCharger struct {
	tagID  string
	params domain.CoinUpkeepParams
}

// deliberate: held, not live, because a lapsed tag still owes its upkeep and paying it is how the lapse lifts (A.001.5 Effort 2 Decision 11).
func (charger *coinUpkeepCharger) UpkeepDue(faction *domain.Faction) (hooks.TagUpkeep, bool) {
	if !faction.HasTag(charger.tagID) {
		return hooks.TagUpkeep{}, false
	}
	return hooks.TagUpkeep{TagID: charger.tagID, Params: charger.params}, true
}
