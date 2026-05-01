package narrative

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/narrative/digest"
)

type wireRenderer struct{}

func (renderer *wireRenderer) Render(cycleDigest digest.CycleDigest, seed int64) (string, error) {
	rng := rand.New(rand.NewSource(seed))
	var output strings.Builder

	output.WriteString(renderHeadline(cycleDigest, rng))
	output.WriteString("\n\n")
	output.WriteString(renderLede(cycleDigest, rng))
	output.WriteString("\n\n")
	for _, beat := range cycleDigest.ActiveFactions {
		output.WriteString(renderFactionSection(beat, cycleDigest.Cross, rng))
		output.WriteString("\n")
	}
	if len(cycleDigest.QuietFactions) > 0 {
		output.WriteString(renderQuietTail(cycleDigest.QuietFactions, rng))
	}
	return output.String(), nil
}

func renderHeadline(cycleDigest digest.CycleDigest, rng *rand.Rand) string {
	h := cycleDigest.Headline
	var tmpl string
	switch h.Kind {
	case digest.HeadlineMajorAttack:
		tmpl = pick(rng, headlineAttackTemplates)
	case digest.HeadlineGoalCompleted:
		tmpl = pick(rng, headlineGoalCompletedTemplates)
	case digest.HeadlineFactionDestroyed:
		tmpl = pick(rng, headlineFactionDestroyedTemplates)
	case digest.HeadlineHomeworldShift:
		tmpl = pick(rng, headlineHomeworldTemplates)
	case digest.HeadlineGoalAbandoned:
		tmpl = pick(rng, headlineGoalAbandonedTemplates)
	default:
		tmpl = pick(rng, headlineQuietTemplates)
	}
	result := strings.ReplaceAll(tmpl, "{name}", h.Subject.Name)
	result = strings.ReplaceAll(result, "{detail}", h.Detail)
	return fmt.Sprintf("# Cycle %d — %s", cycleDigest.Cycle, result)
}

func renderLede(cycleDigest digest.CycleDigest, rng *rand.Rand) string {
	if len(cycleDigest.Cross) > 0 {
		return pick(rng, ledeAttackTemplates)
	}
	for _, beat := range cycleDigest.ActiveFactions {
		for _, ge := range beat.GoalEvents {
			if ge.Kind == digest.GoalCompleted || ge.Kind == digest.GoalHomeworldShift {
				return pick(rng, ledeGoalTemplates)
			}
		}
	}
	return pick(rng, ledeQuietTemplates)
}

func renderFactionSection(beat digest.FactionBeat, crossEvents []digest.CrossEvent, rng *rand.Rand) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "### %s\n\n", beat.Faction.Name)

	for _, ge := range beat.GoalEvents {
		sb.WriteString(renderGoalEvent(ge, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
	for _, acq := range beat.Acquisitions {
		sb.WriteString(renderAcquisition(acq, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
	for _, loss := range beat.Losses {
		sb.WriteString(renderLoss(loss, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
	for _, mov := range beat.Movements {
		sb.WriteString(renderMovement(mov, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
	for _, bribe := range beat.Bribes {
		sb.WriteString(renderBribe(bribe, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
	for _, rep := range beat.Repairs {
		sb.WriteString(renderRepair(rep, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
	for _, exp := range beat.Expansions {
		sb.WriteString(renderExpansion(exp, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
	for _, stealth := range beat.StealthOps {
		sb.WriteString(renderStealthOp(stealth, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
	for _, ce := range crossEvents {
		if ce.Attacker.ID == beat.Faction.ID {
			sb.WriteString(renderCrossEvent(ce, rng))
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

func renderGoalEvent(ge digest.GoalEvent, factionName string, rng *rand.Rand) string {
	switch ge.Kind {
	case digest.GoalCompleted:
		tmpl := pick(rng, goalCompletedTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		result = strings.ReplaceAll(result, "{goal}", ge.GoalName)
		if ge.XPAwarded > 0 {
			result += fmt.Sprintf(" (+%d XP)", ge.XPAwarded)
		}
		return result
	case digest.GoalAbandoned:
		tmpl := pick(rng, goalAbandonedTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{goal}", ge.GoalName)
	case digest.GoalHomeworldShift:
		tmpl := pick(rng, goalHomeworldTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		result = strings.ReplaceAll(result, "{from}", ge.FromWorld)
		return strings.ReplaceAll(result, "{to}", ge.ToWorld)
	case digest.GoalTagGained:
		tmpl := pick(rng, goalTagTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{tag}", ge.TagName)
	}
	return ""
}

func renderAcquisition(move digest.AssetMove, factionName string, rng *rand.Rand) string {
	tmpl := pick(rng, acquisitionTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	result = strings.ReplaceAll(result, "{asset}", move.AssetName)
	if move.From != "" {
		result += fmt.Sprintf(" (upgraded from %s)", move.From)
	}
	return result
}

func renderLoss(move digest.AssetMove, factionName string, rng *rand.Rand) string {
	tmpl := pick(rng, lossTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	return strings.ReplaceAll(result, "{asset}", move.AssetName)
}

func renderMovement(move digest.AssetMove, factionName string, rng *rand.Rand) string {
	tmpl := pick(rng, movementTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	result = strings.ReplaceAll(result, "{asset}", move.AssetName)
	result = strings.ReplaceAll(result, "{from}", move.From)
	return strings.ReplaceAll(result, "{to}", move.To)
}

func renderBribe(bribe digest.BribeEvent, factionName string, rng *rand.Rand) string {
	tmpl := pick(rng, bribeTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	result = strings.ReplaceAll(result, "{location}", bribe.Location)
	return strings.ReplaceAll(result, "{coin}", fmt.Sprintf("%d", -bribe.Coin))
}

func renderRepair(rep digest.RepairEvent, factionName string, rng *rand.Rand) string {
	if rep.IsFaction {
		tmpl := pick(rng, repairFactionTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{hp}", fmt.Sprintf("%d", rep.HPGained))
	}
	tmpl := pick(rng, repairAssetTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	result = strings.ReplaceAll(result, "{asset}", rep.Target.AssetName)
	return strings.ReplaceAll(result, "{hp}", fmt.Sprintf("%d", rep.HPGained))
}

func renderExpansion(exp digest.ExpansionEvent, factionName string, rng *rand.Rand) string {
	if exp.NewBase {
		tmpl := pick(rng, expansionNewBaseTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{location}", exp.Location)
	}
	tmpl := pick(rng, expansionTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	return strings.ReplaceAll(result, "{location}", exp.Location)
}

func renderStealthOp(stealth digest.StealthEvent, factionName string, rng *rand.Rand) string {
	if stealth.Applied {
		tmpl := pick(rng, stealthAppliedTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{asset}", stealth.AssetName)
	}
	tmpl := pick(rng, stealthClearedTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	return strings.ReplaceAll(result, "{asset}", stealth.AssetName)
}

func renderCrossEvent(ce digest.CrossEvent, rng *rand.Rand) string {
	switch ce.Kind {
	case digest.CrossAttack:
		tmpl := pick(rng, crossAttackTemplates)
		result := strings.ReplaceAll(tmpl, "{attacker}", ce.Attacker.Name)
		result = strings.ReplaceAll(result, "{defender}", ce.Defender.Name)
		result = strings.ReplaceAll(result, "{attacker_asset}", ce.AttackerAsset.AssetName)
		result = strings.ReplaceAll(result, "{defender_asset}", ce.DefenderAsset.AssetName)
		return strings.ReplaceAll(result, "{damage}", fmt.Sprintf("%d", ce.DamageToDefender))
	case digest.CrossAbilityStrike:
		tmpl := pick(rng, crossAbilityTemplates)
		result := strings.ReplaceAll(tmpl, "{attacker}", ce.Attacker.Name)
		return strings.ReplaceAll(result, "{defender}", ce.Defender.Name)
	}
	return ""
}

func renderQuietTail(quietFactions []digest.FactionRef, rng *rand.Rand) string {
	names := make([]string, len(quietFactions))
	for i, f := range quietFactions {
		names[i] = f.Name
	}
	tmpl := pick(rng, quietTailTemplates)
	return strings.ReplaceAll(tmpl, "{factions}", strings.Join(names, ", "))
}
