package inputs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

type RepairOrdersSelectedMsg struct{ Orders []engine.RepairOrder }

type repairAssetInfo struct {
	asset     *domain.Asset
	name      string
	maxHP     int
	statScore int
}

type RepairOrdersModel struct {
	step       int
	assets     []repairAssetInfo
	selected   []bool
	cursor     int
	faction    *domain.Faction
	rulebook   *loader.Rulebook
	// count step
	countOrder []repairAssetInfo
	countIndex int
	healCounts []int
}

func NewRepairOrdersModel(faction *domain.Faction, damagedAssets []*domain.Asset, rulebook *loader.Rulebook) RepairOrdersModel {
	infos := make([]repairAssetInfo, len(damagedAssets))
	for i, asset := range damagedAssets {
		info := repairAssetInfo{asset: asset, name: asset.DefinitionID}
		if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
			info.name = def.Name
			info.maxHP = def.HP
			info.statScore = statScoreFor(faction, def.Category)
		}
		infos[i] = info
	}
	return RepairOrdersModel{
		assets:   infos,
		selected: make([]bool, len(infos)),
		faction:  faction,
		rulebook: rulebook,
	}
}

func statScoreFor(faction *domain.Faction, category domain.FactionStat) int {
	switch category {
	case domain.StatForce:
		return faction.Force
	case domain.StatCunning:
		return faction.Cunning
	case domain.StatWealth:
		return faction.Wealth
	default:
		return 0
	}
}

func (m RepairOrdersModel) Init() tea.Cmd { return nil }

func (m RepairOrdersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.step == 0 {
		return m.updateSelect(key)
	}
	return m.updateCount(key)
}

func (m RepairOrdersModel) updateSelect(key tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		// collect selected assets in order
		var order []repairAssetInfo
		for i, info := range m.assets {
			if m.selected[i] {
				order = append(order, info)
			}
		}
		if len(order) == 0 {
			return m, nil
		}
		m.step = 1
		m.countOrder = order
		m.countIndex = 0
		m.healCounts = make([]int, len(order))
		for i := range m.healCounts {
			m.healCounts[i] = 1
		}
	}
	return m, nil
}

func (m RepairOrdersModel) updateCount(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	info := m.countOrder[m.countIndex]
	missing := info.maxHP - info.asset.CurrentHP

	switch key.String() {
	case "+", "=", "up", "k":
		maxCount := maxHealCount(missing, info.statScore)
		coinRemaining := m.faction.Coin - m.confirmedCoinCost()
		if m.healCounts[m.countIndex] < maxCount && healCost(m.healCounts[m.countIndex]+1) <= coinRemaining {
			m.healCounts[m.countIndex]++
		}
	case "-", "down", "j":
		if m.healCounts[m.countIndex] > 1 {
			m.healCounts[m.countIndex]--
		}
	case "enter":
		if m.countIndex < len(m.countOrder)-1 {
			m.countIndex++
		} else {
			orders := make([]engine.RepairOrder, len(m.countOrder))
			for i, info := range m.countOrder {
				orders[i] = engine.RepairOrder{Asset: info.asset, HealCount: m.healCounts[i]}
			}
			return m, func() tea.Msg { return RepairOrdersSelectedMsg{Orders: orders} }
		}
	}
	return m, nil
}

func (m RepairOrdersModel) View() string {
	if m.step == 0 {
		return m.viewSelect()
	}
	return m.viewCount()
}

func (m RepairOrdersModel) viewSelect() string {
	var sb strings.Builder
	sb.WriteString(style.SectionTitle.Render("Repair Assets"))
	fmt.Fprintf(&sb, "\n%s\n\n", style.Muted.Render(fmt.Sprintf("Coin: %d", m.faction.Coin)))

	for i, info := range m.assets {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		check := "[ ]"
		if m.selected[i] {
			check = style.HP.Render("[x]")
		}
		fmt.Fprintf(&sb, "%s%s %s  HP %d/%d  @ %s\n",
			cursor, check, info.name,
			info.asset.CurrentHP, info.maxHP,
			info.asset.Location,
		)
	}

	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Space select  Enter confirm"))
	return sb.String()
}

func (m RepairOrdersModel) viewCount() string {
	info := m.countOrder[m.countIndex]
	missing := info.maxHP - info.asset.CurrentHP
	count := m.healCounts[m.countIndex]

	restored := simulateHeals(count, missing, info.statScore)
	thisCost := healCost(count)
	runningCost := m.confirmedCoinCost() + thisCost

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s  %s\n\n",
		style.SectionTitle.Render("Repair: "+info.name),
		style.Muted.Render(fmt.Sprintf("(%d of %d)", m.countIndex+1, len(m.countOrder))),
	)
	fmt.Fprintf(&sb, "HP: %s/%d  Missing: %d  Stat: %d\n\n",
		style.LowHP.Render(fmt.Sprintf("%d", info.asset.CurrentHP)),
		info.maxHP, missing, info.statScore,
	)

	restoredStr := fmt.Sprintf("%d", restored)
	if restored >= missing {
		restoredStr = style.HP.Render(fmt.Sprintf("%d (full)", restored))
	}
	fmt.Fprintf(&sb, "Heals:    %d\n", count)
	fmt.Fprintf(&sb, "Restored: %s\n", restoredStr)
	fmt.Fprintf(&sb, "Cost:     %s\n\n",
		style.Coin.Render(fmt.Sprintf("%d Coin", thisCost)),
	)

	coinStr := fmt.Sprintf("Running total: %d / %d Coin", runningCost, m.faction.Coin)
	if runningCost > m.faction.Coin {
		coinStr = style.LowHP.Render(coinStr)
	} else {
		coinStr = style.Muted.Render(coinStr)
	}
	sb.WriteString(coinStr)
	sb.WriteString("\n\n")
	sb.WriteString(style.Muted.Render("+/- adjust heals  Enter confirm"))
	return sb.String()
}

// confirmedCoinCost returns the total Coin cost for all assets already confirmed
// in the count step (i.e., those before countIndex).
func (m RepairOrdersModel) confirmedCoinCost() int {
	total := 0
	for i := 0; i < m.countIndex; i++ {
		total += healCost(m.healCounts[i])
	}
	return total
}

// healCost returns the total Coin cost for n heals on a single asset.
// Cost escalates: 1st heal = 1, 2nd = 2, ..., n-th = n. Total = n*(n+1)/2.
func healCost(count int) int {
	return count * (count + 1) / 2
}

// maxHealCount returns how many heals are needed to fully repair an asset
// given its missing HP and the faction stat score per heal.
func maxHealCount(missing, statScore int) int {
	if statScore <= 0 {
		return 0
	}
	count := 0
	totalHeal := 0
	for totalHeal < missing {
		healAmt := min(statScore, missing-totalHeal)
		if healAmt <= 0 {
			break
		}
		totalHeal += healAmt
		count++
	}
	return count
}

// simulateHeals returns the total HP restored by count heals on an asset with
// the given missing HP and stat score, mirroring the engine's Resolve logic.
func simulateHeals(count, missing, statScore int) int {
	totalHeal := 0
	for range count {
		healAmt := min(statScore, missing-totalHeal)
		if healAmt <= 0 {
			break
		}
		totalHeal += healAmt
	}
	return totalHeal
}
