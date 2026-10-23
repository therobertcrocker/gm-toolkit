package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
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

func validateWithdrawPermission(entry domain.EffectDefinition) error {
	if entry.Gates.Contest == nil || entry.Gates.Contest.Roller != domain.ContestRollerSelf {
		return fmt.Errorf("needs a contest gate with roller %q", domain.ContestRollerSelf)
	}
	return nil
}

func registerWithdrawPermission(ctx effect.RegisterContext) {
	ctx.Registry.RegisterPermissionContester(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialWithdrawPermission)+":"+ctx.Def.ID,
		&permissionContester{owner: ctx.Faction, assetID: ctx.Asset.ID, contest: *ctx.Entry.Gates.Contest, roller: ctx.Roller},
	)
}

// permissionContester holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Effort 2 Decision Record 19). It keeps the roller its
// registration carried, as a reactor does.
type permissionContester struct {
	owner   *domain.Faction
	assetID string
	contest domain.EffectContest
	roller  domain.Roller
}

func (contester *permissionContester) OfferContest(query hooks.PermissionQuery) *hooks.PermissionContest {
	self := contester.owner.Assets[contester.assetID]
	if !active(self) || query.Requester.ID == contester.owner.ID {
		return nil
	}
	if self.Location.WorldID == "" || self.Location.WorldID != query.WorldID {
		return nil
	}
	return &hooks.PermissionContest{Source: self}
}

// WinsContest rolls the bare contest, as every effect contest does (Effort 3
// Decision 12): the request path carries no rulebook for roll hooks.
func (contester *permissionContester) WinsContest(query hooks.PermissionQuery) bool {
	return contestPasses(contester.contest, contester.owner, query.Requester, contester.roller)
}
