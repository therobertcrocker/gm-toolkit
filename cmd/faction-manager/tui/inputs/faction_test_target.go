package inputs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

type FactionTestTargetSelectedMsg struct {
	Faction *domain.Faction
}

type FactionTestTargetModel struct {
	assetName string
	effect    domain.AbilityEffectType
	factions  []*domain.Faction
	cursor    int
}

func NewFactionTestTargetModel(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction, rulebook *loader.Rulebook) FactionTestTargetModel {
	name := asset.DefinitionID
	if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
		name = def.Name
	}
	return FactionTestTargetModel{
		assetName: name,
		effect:    effect,
		factions:  candidates,
	}
}

func (m FactionTestTargetModel) Init() tea.Cmd { return nil }

func (m FactionTestTargetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.factions)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.factions) == 0 {
			return m, nil
		}
		faction := m.factions[m.cursor]
		return m, func() tea.Msg { return FactionTestTargetSelectedMsg{Faction: faction} }
	}
	return m, nil
}

func (m FactionTestTargetModel) View() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", style.SectionTitle.Render("Faction Test — Select Target"))
	fmt.Fprintf(&sb, "%s\n\n", style.Muted.Render(
		fmt.Sprintf("%s — %s", m.assetName, abilityEffectLabel(m.effect)),
	))
	if len(m.factions) == 0 {
		sb.WriteString(style.Muted.Render("No eligible targets."))
		return sb.String()
	}
	for i, faction := range m.factions {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		fmt.Fprintf(&sb, "%s%s\n", cursor, faction.Name)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Enter confirm"))
	return sb.String()
}

func abilityEffectLabel(effect domain.AbilityEffectType) string {
	switch effect {
	case domain.EffectRevealStealth:
		return "Reveal Stealth"
	case domain.EffectCoinDrain:
		return "Coin Drain"
	case domain.EffectCoinSteal:
		return "Coin Steal"
	default:
		return string(effect)
	}
}
