package turn

import (
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// buildFactionOrder rolls a starting index via roller and returns faction IDs as
// a rotation. Keys are sorted before randomizing so the rotation is deterministic
// given the same set and seed.
func buildFactionOrder(factions map[string]*domain.Faction, roller domain.Roller) []string {
	ids := make([]string, 0, len(factions))
	for id := range factions {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	n := len(ids)
	start := roller.Roll(n) - 1 // Roll(n) → [1,n]; subtract 1 for [0,n-1]
	order := make([]string, n)
	for i := range n {
		order[i] = ids[(start+i)%n]
	}
	return order
}
