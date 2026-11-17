package foundry

import "github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"

const (
	ReasonAbility     = "ability not converted yet (rules in Effort 3, descriptions in Effort 4)"
	ReasonEffort3     = "needs a rule from Effort 3"
	ReasonConditional = "conditional: does not hold at baseline"
	ReasonNotCombat   = "not a combat statistic"
	ReasonState       = "sheet state, ignored (D-26)"
	ReasonInactive    = "item not worn, held or invested"
	ReasonToggle      = "toggle, imported off"
	ReasonQuantity    = "quantity 0"
	ReasonNotPlaced   = "loose item, not placed on a creature"
	ReasonEmpty       = "holds nothing the sim uses"
	ReasonUnsupported = "not supported by the importer"
)

type Skip struct {
	Item   string
	What   string
	Reason string
}

type Report struct {
	Source   string
	Creature string
	Kind     creature.Kind
	Folder   string
	Skipped  []Skip
	Unused   map[string]int
}

func (report *Report) skip(itemName, what, reason string) {
	report.Skipped = append(report.Skipped, Skip{Item: itemName, What: what, Reason: reason})
}

func (report *Report) unused(kind string) {
	if report.Unused == nil {
		report.Unused = map[string]int{}
	}
	report.Unused[kind]++
}
