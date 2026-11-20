package creature

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

type Limit struct {
	PerFight     int             `toml:"per_fight,omitzero"`
	PerRound     int             `toml:"per_round,omitzero"`
	Recharge     string          `toml:"recharge,omitempty"`
	Pool         string          `toml:"pool,omitempty"`
	RechargeRoll dice.Expression `toml:"-"`
}

func (ability Ability) CostRange() (int, int) {
	if ability.Free || ability.Trigger != nil {
		return 0, 0
	}
	return ability.Actions, max(ability.Actions, ability.MaxActions)
}

func (ability Ability) HasTrait(trait string) bool {
	return slices.Contains(ability.Traits, trait)
}

func (creature *Creature) checkPools(report func(string, ...any)) {
	for _, name := range slices.Sorted(maps.Keys(creature.Pools)) {
		if creature.Pools[name] < 0 {
			report("pool.%s: must not be negative", name)
		}
	}
}

func checkCost(path string, ability Ability, report func(string, ...any)) {
	switch {
	case ability.Free && (ability.Actions != 0 || ability.MaxActions != 0):
		report("%s: a free ability takes no actions", path)
	case ability.Free && ability.Limit == nil:
		report("%s: a free ability needs a limit", path)
	case !ability.Free && (ability.Actions < 1 || ability.Actions > 3):
		report("%s: actions %d: want 1 to 3", path, ability.Actions)
	case ability.MaxActions != 0 && (ability.MaxActions <= ability.Actions || ability.MaxActions > 3):
		report("%s: max_actions %d: want more than actions and at most 3", path, ability.MaxActions)
	}
}

func (creature *Creature) checkLimit(path string, limit *Limit, report func(string, ...any)) {
	if limit == nil {
		return
	}
	if limit.PerFight < 0 || limit.PerRound < 0 {
		report("%s.limit: per_fight and per_round must not be negative", path)
	}
	if limit.PerFight == 0 && limit.PerRound == 0 && limit.Recharge == "" && limit.Pool == "" {
		report("%s.limit: sets nothing", path)
	}
	if limit.Recharge != "" {
		limit.RechargeRoll = parseAmount(path+".limit.recharge", limit.Recharge, report)
	}
	if _, found := creature.Pools[limit.Pool]; limit.Pool != "" && !found {
		report("%s.limit: unknown pool %q", path, limit.Pool)
	}
}
