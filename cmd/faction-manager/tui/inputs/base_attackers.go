package inputs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

type BaseAttackersSelectedMsg struct {
	Attackers []*domain.Asset
}

type baseAtkInfo struct {
	asset *domain.Asset
	name  string
}

type SelectBaseAttackersModel struct {
	rivalName string
	assets    []baseAtkInfo
	selected  []bool
	cursor    int
}

func NewSelectBaseAttackersModel(rival *domain.Faction, eligible []*domain.Asset, rulebook *loader.Rulebook) SelectBaseAttackersModel {
	infos := make([]baseAtkInfo, len(eligible))
	for i, asset := range eligible {
		info := baseAtkInfo{asset: asset, name: asset.DefinitionID}
		if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
			info.name = def.Name
		}
		infos[i] = info
	}
	return SelectBaseAttackersModel{
		rivalName: rival.Name,
		assets:    infos,
		selected:  make([]bool, len(infos)),
	}
}

func (m SelectBaseAttackersModel) Init() tea.Cmd { return nil }

func (m SelectBaseAttackersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.cursor < len(m.assets)-1 {
			m.cursor++
		}
	case " ":
		m.selected[m.cursor] = !m.selected[m.cursor]
	case "enter":
		var attackers []*domain.Asset
		for i, info := range m.assets {
			if m.selected[i] {
				attackers = append(attackers, info.asset)
			}
		}
		if len(attackers) == 0 {
			return m, nil
		}
		return m, func() tea.Msg { return BaseAttackersSelectedMsg{Attackers: attackers} }
	}
	return m, nil
}

func (m SelectBaseAttackersModel) View() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", style.SectionTitle.Render("Free Attack — Select Attackers"))
	fmt.Fprintf(&sb, "%s\n\n", style.Muted.Render(fmt.Sprintf("%s's assets attacking the new Base:", m.rivalName)))
	for i, info := range m.assets {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		check := "[ ]"
		if m.selected[i] {
			check = style.HP.Render("[x]")
		}
		fmt.Fprintf(&sb, "%s%s %-20s  HP %d  @ %s\n",
			cursor, check, info.name,
			info.asset.CurrentHP, info.asset.Location,
		)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Space select  Enter confirm"))
	return sb.String()
}
