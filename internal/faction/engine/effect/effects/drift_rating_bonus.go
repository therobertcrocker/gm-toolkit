package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// driftRatingBonusRegistrar reads owner = "own" as "the moving asset is the
// owner's", and scope = "faction" as "every asset the owner holds" (Drift-Runners).
// The entry carries the bonus; a movement order prices its crossing from the raised
// rating.
var driftRatingBonusRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerDriftRatingQuery: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerDriftRatingBonus,
}

func registerDriftRatingBonus(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetDriftRatingModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialDriftRatingBonus),
		&driftRatingBonusModifier{
			owner:  ctx.Holder.Faction,
			holder: refOf(ctx.Holder),
			amount: ctx.Entry.DriftRatingBonus.Amount,
		},
	)
}

// driftRatingBonusModifier holds its owner, not its holder's asset: a Cat 4 hook
// gets no FactionState (Decision Record 19).
type driftRatingBonusModifier struct {
	owner  *domain.Faction
	holder holderRef
	amount int
}

func (modifier *driftRatingBonusModifier) ModifyAssetDriftRating(faction *domain.Faction, _ *domain.AssetDefinition, rating int) int {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || faction.ID != modifier.owner.ID {
		return rating
	}
	return rating + modifier.amount
}
