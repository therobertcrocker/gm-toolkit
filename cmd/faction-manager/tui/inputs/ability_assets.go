package inputs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

type AbilityAssetsSelectedMsg struct {
	Assets []*domain.Asset
}

type abilityAssetInfo struct {
	asset     *domain.Asset
	name      string
	catAbbrev string
}

type AbilityAssetsModel struct {
	assets         []abilityAssetInfo
	cursor         int
	selectedSet    map[int]bool
	selectionOrder []int
}

func NewAbilityAssetsModel(candidates []*domain.Asset, rulebook *loader.Rulebook) AbilityAssetsModel {
	infos := make([]abilityAssetInfo, len(candidates))
	for i, asset := range candidates {
		info := abilityAssetInfo{asset: asset, name: asset.DefinitionID, catAbbrev: "?"}
		if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
			info.name = def.Name
			info.catAbbrev = abilityCatAbbrev(def.Category)
		}
		infos[i] = info
	}
	return AbilityAssetsModel{
		assets:      infos,
		selectedSet: make(map[int]bool),
	}
}

func (m AbilityAssetsModel) Init() tea.Cmd { return nil }

func (m AbilityAssetsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.selectedSet[m.cursor] {
			m.selectedSet[m.cursor] = false
			for i, idx := range m.selectionOrder {
				if idx == m.cursor {
					m.selectionOrder = append(m.selectionOrder[:i], m.selectionOrder[i+1:]...)
					break
				}
			}
		} else {
			m.selectedSet[m.cursor] = true
			m.selectionOrder = append(m.selectionOrder, m.cursor)
		}
	case "enter":
		if len(m.selectionOrder) == 0 {
			return m, nil
		}
		assets := make([]*domain.Asset, len(m.selectionOrder))
		for i, idx := range m.selectionOrder {
			assets[i] = m.assets[idx].asset
		}
		return m, func() tea.Msg { return AbilityAssetsSelectedMsg{Assets: assets} }
	}
	return m, nil
}

func (m AbilityAssetsModel) View() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n\n", style.SectionTitle.Render("Use Asset Ability — Select Assets"))
	for i, info := range m.assets {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		check := "[ ]"
		for order, idx := range m.selectionOrder {
			if idx == i {
				check = style.HP.Render(fmt.Sprintf("[%d]", order+1))
				break
			}
		}
		fmt.Fprintf(&sb, "%s%s %-20s %s  HP %d  @ %s\n",
			cursor, check, info.name, info.catAbbrev,
			info.asset.CurrentHP, info.asset.Location,
		)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Space select (numbers show resolve order)  Enter confirm"))
	return sb.String()
}

func abilityCatAbbrev(category domain.FactionStat) string {
	switch category {
	case domain.StatForce:
		return "F"
	case domain.StatCunning:
		return "C"
	case domain.StatWealth:
		return "W"
	default:
		return "?"
	}
}
