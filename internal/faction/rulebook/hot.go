package rulebook

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// RunsHot reports whether faction holds a live tag with a hot-only effect, so its
// move orders may run hot (Aether Burners). A lapsed tag does not count.
func (rulebook *Rulebook) RunsHot(faction *domain.Faction) bool {
	if faction == nil {
		return false
	}
	for _, tagID := range faction.Tags {
		tag, ok := rulebook.Tags[tagID]
		if !ok || slices.Contains(faction.LapsedTags, tagID) {
			continue
		}
		for _, entry := range tag.Effects {
			if entry.Filters.Hot {
				return true
			}
		}
	}
	return false
}
