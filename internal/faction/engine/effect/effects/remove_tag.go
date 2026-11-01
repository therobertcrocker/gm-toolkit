package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// removeTagRegistrar is Deep Rooted's loss: when the owner's homeworld changes,
// the tag that declared the entry leaves the faction (D-19). owner = "own" reads
// as "the owner's homeworld changed". Only a tag holds it: an asset has no tag of
// its own to remove.
var removeTagRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerHomeworldChanged: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerRemoveTag,
	Validate: validateRemoveTag,
}

func validateRemoveTag(_ domain.EffectDefinition, holder effect.HolderKind) error {
	if holder != effect.HolderTag {
		return fmt.Errorf("only a tag can remove itself")
	}
	return nil
}

func registerRemoveTag(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialRemoveTag), &removeTagReactor{holder: refOf(ctx.Holder)})
}

type removeTagReactor struct {
	holder holderRef
}

// OnMutations removes the tag once, on the first HomeworldChanged of its owner in
// the batch, while the owner still holds it.
func (reactor *removeTagReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	for _, mutation := range mutations {
		changed, ok := mutation.(domain.HomeworldChanged)
		if !ok || changed.FactionID != reactor.holder.factionID {
			continue
		}
		return []domain.Mutation{domain.TagRemoved{
			FactionID:         reactor.holder.factionID,
			TagID:             reactor.holder.tagID,
			Cause:             string(domain.SpecialRemoveTag),
			CausedByFactionID: reactor.holder.factionID,
		}}
	}
	return nil
}
