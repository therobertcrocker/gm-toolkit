package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/relocate"
)

// relocateAssets runs a relocate step through the shared relocate code, with the
// acting asset as the range origin. The caller spends the action.
func relocateAssets(ctx StepContext) ([]domain.Mutation, error) {
	if ctx.Step.Relocate == nil {
		return nil, fmt.Errorf("relocate: step has no relocate params")
	}
	return relocate.Run(relocate.Request{
		Faction:   ctx.Faction,
		Acting:    ctx.Asset,
		Params:    ctx.Step.Relocate,
		World:     ctx.WorldEngine,
		Rulebook:  ctx.Rulebook,
		Collector: ctx.Collector,
	})
}
