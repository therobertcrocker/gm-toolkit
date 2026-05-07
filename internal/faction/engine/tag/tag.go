package tag

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/tag/tags"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var handlers = map[string]func(*engine.Engine, *domain.Faction){
	tags.ScavengersTagID: tags.RegisterScavengers,
	tags.WarlikeTagID:    tags.RegisterWarlike,
	tags.FanaticalTagID:  tags.RegisterFanatical,
}

// RegisterDefaultTags walks factionState and, for each faction tag with a
// known handler, registers the corresponding hook into eng.
// Tags without a handler are silently skipped (data-only).
func RegisterDefaultTags(eng *engine.Engine, factionState *state.FactionState) {
	for _, faction := range factionState.Factions {
		for _, tag := range faction.Tags {
			if register, ok := handlers[tag.ID]; ok {
				register(eng, faction)
			}
		}
	}
}
