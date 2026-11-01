package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// arrivalTollRegistrar reads owner = "rival" as "the landing asset is a rival's",
// and scope = "base" as "on a world where the owner has a Base of Influence"
// (Pirates). The entry carries the toll (Effort 4 Decision 8).
var arrivalTollRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerArrivalQuery: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeBase},
		},
	},
	Register: registerArrivalToll,
}

func registerArrivalToll(ctx effect.RegisterContext) {
	ctx.Registry.RegisterArrivalTollRule(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialArrivalToll),
		&arrivalTollRule{
			owner:  ctx.Holder.Faction,
			holder: refOf(ctx.Holder),
			amount: ctx.Entry.ArrivalToll.Amount,
		},
	)
}

// arrivalTollRule holds its owner, not its holder's asset: a Cat 4 hook gets no
// FactionState (Decision Record 19). The owner's Bases are read live at query time.
type arrivalTollRule struct {
	owner  *domain.Faction
	holder holderRef
	amount int
}

func (rule *arrivalTollRule) ArrivalToll(mover *domain.Faction, world string) int {
	if _, live := rule.holder.liveIn(rule.owner); !live || mover.ID == rule.owner.ID || !rule.owner.HasBaseOn(world) {
		return 0
	}
	return rule.amount
}
