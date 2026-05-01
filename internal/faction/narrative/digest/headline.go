package digest

import "sort"

func isQuiet(beat FactionBeat, crossEvents []CrossEvent) bool {
	if len(beat.Acquisitions) > 0 || len(beat.Losses) > 0 || len(beat.Movements) > 0 ||
		len(beat.Bribes) > 0 || len(beat.Repairs) > 0 || len(beat.Expansions) > 0 ||
		len(beat.GoalEvents) > 0 || len(beat.StealthOps) > 0 {
		return false
	}
	for _, ce := range crossEvents {
		if ce.Attacker.ID == beat.Faction.ID || ce.Defender.ID == beat.Faction.ID {
			return false
		}
	}
	return true
}

func selectHeadline(activeFactions []FactionBeat, crossEvents []CrossEvent, destroyedFactions []FactionRef) Headline {
	// Priority 1: Major Attack
	if len(crossEvents) > 0 {
		best := crossEvents[0]
		for _, ce := range crossEvents[1:] {
			bestDmg := best.DamageToDefender + best.DamageToAttacker
			ceDmg := ce.DamageToDefender + ce.DamageToAttacker
			if ceDmg > bestDmg || (ceDmg == bestDmg && ce.Attacker.Name < best.Attacker.Name) {
				best = ce
			}
		}
		return Headline{Subject: best.Attacker, Kind: HeadlineMajorAttack, Detail: best.Defender.Name}
	}

	// Priority 2: Goal Completed
	var bestBeat *FactionBeat
	var bestGoal *GoalEvent
	for i := range activeFactions {
		for j := range activeFactions[i].GoalEvents {
			ge := &activeFactions[i].GoalEvents[j]
			if ge.Kind != GoalCompleted {
				continue
			}
			if bestGoal == nil ||
				ge.XPAwarded > bestGoal.XPAwarded ||
				(ge.XPAwarded == bestGoal.XPAwarded && activeFactions[i].Faction.Name < bestBeat.Faction.Name) {
				bestBeat = &activeFactions[i]
				bestGoal = ge
			}
		}
	}
	if bestGoal != nil {
		return Headline{Subject: bestBeat.Faction, Kind: HeadlineGoalCompleted, Detail: bestGoal.GoalName}
	}

	// Priority 3: Faction Destroyed
	if len(destroyedFactions) > 0 {
		sort.Slice(destroyedFactions, func(i, j int) bool {
			return destroyedFactions[i].Name < destroyedFactions[j].Name
		})
		return Headline{Subject: destroyedFactions[0], Kind: HeadlineFactionDestroyed}
	}

	// Priority 4: Homeworld Shift
	for _, beat := range activeFactions {
		for _, ge := range beat.GoalEvents {
			if ge.Kind == GoalHomeworldShift {
				return Headline{Subject: beat.Faction, Kind: HeadlineHomeworldShift, Detail: ge.ToWorld}
			}
		}
	}

	// Priority 5: Goal Abandoned
	for _, beat := range activeFactions {
		for _, ge := range beat.GoalEvents {
			if ge.Kind == GoalAbandoned {
				return Headline{Subject: beat.Faction, Kind: HeadlineGoalAbandoned, Detail: ge.GoalName}
			}
		}
	}

	return Headline{Kind: HeadlineQuiet}
}
