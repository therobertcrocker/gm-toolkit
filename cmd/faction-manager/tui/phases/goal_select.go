package phases

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type GoalSelectedMsg struct{ ActiveGoal *domain.ActiveGoal }

type goalSelectStep int

const (
	goalStepGoal     goalSelectStep = iota
	goalStepWorld                   // G-012 destination or G-004 target
	goalStepDistance                // G-012 hex distance
)

type goalOption struct {
	id   string
	name string
}

type GoalSelectModel struct {
	faction      *domain.Faction
	factionState *state.FactionState
	rulebook     *loader.Rulebook

	step         goalSelectStep
	goalOptions  []goalOption
	goalCursor   int
	selectedGoal goalOption

	worldOptions []string
	worldCursor  int

	distance    int
	distanceStr string
}

func NewGoalSelectModel(faction *domain.Faction, factionState *state.FactionState, rulebook *loader.Rulebook) GoalSelectModel {
	goals := make([]goalOption, 0, len(rulebook.Goals))
	for id, g := range rulebook.Goals {
		goals = append(goals, goalOption{id: id, name: g.Name})
	}
	sort.Slice(goals, func(i, j int) bool { return goals[i].id < goals[j].id })
	return GoalSelectModel{
		faction:      faction,
		factionState: factionState,
		rulebook:     rulebook,
		goalOptions:  goals,
		distance:     1,
		distanceStr:  "1",
	}
}

func (m GoalSelectModel) Init() tea.Cmd { return nil }

func (m GoalSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch m.step {
	case goalStepGoal:
		return m.updateGoalStep(key)
	case goalStepWorld:
		return m.updateWorldStep(key)
	case goalStepDistance:
		return m.updateDistanceStep(key)
	}
	return m, nil
}

func (m GoalSelectModel) updateGoalStep(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		if m.goalCursor > 0 {
			m.goalCursor--
		}
	case "down", "j":
		if m.goalCursor < len(m.goalOptions)-1 {
			m.goalCursor++
		}
	case "enter":
		m.selectedGoal = m.goalOptions[m.goalCursor]
		switch m.selectedGoal.id {
		case "G-012":
			m.worldOptions = goalChangeHomeworldDestinations(m.faction)
			if len(m.worldOptions) == 0 {
				return m, nil
			}
			m.worldCursor = 0
			m.step = goalStepWorld
		case "G-004":
			m.worldOptions = goalSeizePlanetTargets(m.faction, m.factionState)
			if len(m.worldOptions) == 0 {
				return m, nil
			}
			m.worldCursor = 0
			m.step = goalStepWorld
		default:
			activeGoal := &domain.ActiveGoal{GoalID: m.selectedGoal.id}
			return m, func() tea.Msg { return GoalSelectedMsg{ActiveGoal: activeGoal} }
		}
	}
	return m, nil
}

func (m GoalSelectModel) updateWorldStep(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		if m.worldCursor > 0 {
			m.worldCursor--
		}
	case "down", "j":
		if m.worldCursor < len(m.worldOptions)-1 {
			m.worldCursor++
		}
	case "enter":
		selectedWorld := m.worldOptions[m.worldCursor]
		if m.selectedGoal.id == "G-012" {
			m.distance = 1
			m.distanceStr = "1"
			m.step = goalStepDistance
			// Store world selection in a temporary field via the same model.
			// We need it when emitting the final message — store it by advancing
			// the world options so worldOptions[0] is the selected world.
			m.worldOptions = []string{selectedWorld}
			m.worldCursor = 0
		} else {
			// G-004: target world selected — done
			activeGoal := &domain.ActiveGoal{
				GoalID:      m.selectedGoal.id,
				TargetWorld: selectedWorld,
			}
			return m, func() tea.Msg { return GoalSelectedMsg{ActiveGoal: activeGoal} }
		}
	}
	return m, nil
}

func (m GoalSelectModel) updateDistanceStep(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "backspace":
		if len(m.distanceStr) > 0 {
			m.distanceStr = m.distanceStr[:len(m.distanceStr)-1]
		}
		if n, err := strconv.Atoi(m.distanceStr); err == nil && n >= 1 {
			m.distance = n
		}
	case "enter":
		if m.distance < 1 {
			return m, nil
		}
		destination := m.worldOptions[0]
		activeGoal := &domain.ActiveGoal{
			GoalID:         m.selectedGoal.id,
			TargetWorld:    destination,
			TurnsRemaining: m.distance + 1,
		}
		return m, func() tea.Msg { return GoalSelectedMsg{ActiveGoal: activeGoal} }
	default:
		if len(key.String()) == 1 && key.String() >= "0" && key.String() <= "9" {
			m.distanceStr += key.String()
			if n, err := strconv.Atoi(m.distanceStr); err == nil && n >= 1 {
				m.distance = n
			}
		}
	}
	return m, nil
}

func (m GoalSelectModel) View() string {
	switch m.step {
	case goalStepGoal:
		return m.viewGoalStep()
	case goalStepWorld:
		return m.viewWorldStep()
	case goalStepDistance:
		return m.viewDistanceStep()
	}
	return ""
}

func (m GoalSelectModel) viewGoalStep() string {
	var sb strings.Builder
	sb.WriteString(style.SectionTitle.Render("Select Goal"))
	sb.WriteString("\n\n")
	for i, opt := range m.goalOptions {
		cursor := "  "
		if i == m.goalCursor {
			cursor = "> "
		}
		fmt.Fprintf(&sb, "%s%s  %s\n", cursor, opt.id, opt.name)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Enter select"))
	return sb.String()
}

func (m GoalSelectModel) viewWorldStep() string {
	var sb strings.Builder
	var title string
	if m.selectedGoal.id == "G-012" {
		title = "Change Homeworld — Select Destination"
	} else {
		title = "Planetary Seizure — Select Target World"
	}
	sb.WriteString(style.SectionTitle.Render(title))
	sb.WriteString("\n\n")
	for i, world := range m.worldOptions {
		cursor := "  "
		if i == m.worldCursor {
			cursor = "> "
		}
		fmt.Fprintf(&sb, "%s%s\n", cursor, world)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Enter select"))
	return sb.String()
}

func (m GoalSelectModel) viewDistanceStep() string {
	var sb strings.Builder
	destination := ""
	if len(m.worldOptions) > 0 {
		destination = m.worldOptions[0]
	}
	sb.WriteString(style.SectionTitle.Render(fmt.Sprintf("Change Homeworld — %s", destination)))
	sb.WriteString("\n\n")
	sb.WriteString("Hex distance: ")
	if m.distanceStr == "" {
		sb.WriteString(style.Muted.Render("_"))
	} else {
		sb.WriteString(style.HP.Render(m.distanceStr))
	}
	fmt.Fprintf(&sb, "\n\n%s\n",
		style.Muted.Render(fmt.Sprintf("Transit time: %d turns", m.distance+1)),
	)
	sb.WriteString(style.Muted.Render("Type distance  Enter confirm"))
	return sb.String()
}

// goalChangeHomeworldDestinations returns non-homeworld base worlds for G-012.
func goalChangeHomeworldDestinations(faction *domain.Faction) []string {
	var worlds []string
	for _, base := range faction.Bases {
		if !base.IsHomeworld {
			worlds = append(worlds, base.Location)
		}
	}
	sort.Strings(worlds)
	return worlds
}

// goalSeizePlanetTargets returns worlds contested by the faction and a rival for G-004.
func goalSeizePlanetTargets(faction *domain.Faction, factionState *state.FactionState) []string {
	factionWorlds := map[string]bool{}
	for _, asset := range faction.Assets {
		if !asset.Stealthy {
			factionWorlds[asset.Location] = true
		}
	}
	contested := map[string]bool{}
	for factionID, rival := range factionState.Factions {
		if factionID == faction.ID {
			continue
		}
		for _, asset := range rival.Assets {
			if !asset.Stealthy && factionWorlds[asset.Location] {
				contested[asset.Location] = true
			}
		}
	}
	worlds := make([]string, 0, len(contested))
	for world := range contested {
		worlds = append(worlds, world)
	}
	sort.Strings(worlds)
	return worlds
}
