package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// withdrawPermissionRegistrar is Lobbyists: when a rival is granted a
// government's permission on the declaring asset's world, its owner may test the
// rival, and a win withdraws the permission. owner = "rival" reads as "the
// requester is a rival", scope = "world" as "the request is for the declaring
// asset's world". The contest gate carries the test (Effort 5 Decision 10).
var withdrawPermissionRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerPermissionGranted: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeWorld},
			Gates:  []domain.EffectGateName{domain.GateContest},
		},
	},
	Register: registerWithdrawPermission,
	Validate: validateWithdrawPermission,
}

func validateWithdrawPermission(entry domain.EffectDefinition, _ effect.HolderKind) error {
	if entry.Gates.Contest == nil || entry.Gates.Contest.Roller != domain.ContestRollerSelf {
		return fmt.Errorf("needs a contest gate with roller %q", domain.ContestRollerSelf)
	}
	return nil
}

func registerWithdrawPermission(ctx effect.RegisterContext) {
	ctx.Registry.RegisterPermissionContester(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialWithdrawPermission),
		&permissionContester{owner: ctx.Holder.Faction, holder: refOf(ctx.Holder), contest: *ctx.Entry.Gates.Contest, roller: ctx.Roller, registry: ctx.Registry},
	)
}

// permissionContester holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Effort 2 Decision Record 19). It keeps the roller and registry
// its registration carried, as a reactor does.
type permissionContester struct {
	owner    *domain.Faction
	holder   holderRef
	contest  domain.EffectContest
	roller   dice.Roller
	registry *hooks.Registry
}

func (contester *permissionContester) OfferContest(query hooks.PermissionQuery) *hooks.PermissionContest {
	self, live := contester.holder.liveIn(contester.owner)
	if !live || query.Requester.ID == contester.owner.ID {
		return nil
	}
	if self.Location.WorldID == "" || self.Location.WorldID != query.WorldID {
		return nil
	}
	return &hooks.PermissionContest{Source: self}
}

// WinsContest rolls the contest through the roll hooks (Effort 3 Decision 3),
// with the state and rulebook the permission dispatcher holds. The requester has
// no asset on the roll: it asks for one that is not there yet.
func (contester *permissionContester) WinsContest(query hooks.PermissionQuery, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool {
	self, _ := contester.holder.liveIn(contester.owner)
	rolls := dice.Engine{
		// deliberate: a nil collector, not a prompt, because the permission path holds only a PermissionCollector; the roll-card follow-on row gives it one (overview Decision 2).
		Collector:    nil,
		Registry:     contester.registry,
		Roller:       contester.roller,
		FactionState: factionState,
		Rulebook:     rulebook,
	}
	return rolls.Gate(
		contester.contest.Contest, contester.contest.Roller,
		dice.Side{Faction: contester.owner, Asset: self},
		dice.Side{Faction: query.Requester},
		query.WorldID, contestAction,
	)
}
