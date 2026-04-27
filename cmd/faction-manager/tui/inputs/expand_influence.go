package inputs

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
)

type ExpandInfluenceOrderSelectedMsg struct{ Order engine.ExpandInfluenceOrder }

type expandStep int

const (
	expandStepMode     expandStep = iota
	expandStepWorld
	expandStepSubMode
	expandStepHPAmount
)

type expandModeOption struct {
	label string
	mode  engine.ExpandMode
}

type expandWorldOption struct {
	world  string
	baseID string // non-empty for reinforce mode
}

type expandSubModeOption struct {
	label string
	mode  engine.ReinforceMode
}

type ExpandInfluenceModel struct {
	faction         *domain.Faction
	step            expandStep
	modeOptions     []expandModeOption
	modeCursor      int
	selectedMode    engine.ExpandMode
	worldOptions    []expandWorldOption
	worldCursor     int
	selectedWorld   string
	selectedBaseID  string
	subModeOptions  []expandSubModeOption
	subModeCursor   int
	selectedSubMode engine.ReinforceMode
	hpAmount        int
	hpMax           int
}

func NewExpandInfluenceModel(faction *domain.Faction) ExpandInfluenceModel {
	m := ExpandInfluenceModel{faction: faction}

	canNew := faction.Coin >= 1 && len(expandWorldsForNewBase(faction)) > 0
	canReinforce := faction.Coin >= 1 && (len(expandDamagedBases(faction)) > 0 || len(expandGrowableBases(faction)) > 0)

	if canNew {
		m.modeOptions = append(m.modeOptions, expandModeOption{"New Base of Influence", engine.ExpandModeNew})
	}
	if canReinforce {
		m.modeOptions = append(m.modeOptions, expandModeOption{"Reinforce Existing Base", engine.ExpandModeReinforce})
	}

	// Skip mode step when only one mode is available.
	if len(m.modeOptions) == 1 {
		m.selectedMode = m.modeOptions[0].mode
		m.step = expandStepWorld
		m.worldOptions = m.buildWorldOptions()
	}

	return m
}

func (m ExpandInfluenceModel) Init() tea.Cmd { return nil }

func (m ExpandInfluenceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch m.step {
	case expandStepMode:
		return m.updateMode(key)
	case expandStepWorld:
		return m.updateWorld(key)
	case expandStepSubMode:
		return m.updateSubMode(key)
	case expandStepHPAmount:
		return m.updateHPAmount(key)
	}
	return m, nil
}

func (m ExpandInfluenceModel) updateMode(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		if m.modeCursor > 0 {
			m.modeCursor--
		}
	case "down", "j":
		if m.modeCursor < len(m.modeOptions)-1 {
			m.modeCursor++
		}
	case "enter":
		m.selectedMode = m.modeOptions[m.modeCursor].mode
		m.step = expandStepWorld
		m.worldOptions = m.buildWorldOptions()
		m.worldCursor = 0
	}
	return m, nil
}

func (m ExpandInfluenceModel) updateWorld(key tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if len(m.worldOptions) == 0 {
			break
		}
		opt := m.worldOptions[m.worldCursor]
		m.selectedWorld = opt.world
		m.selectedBaseID = opt.baseID
		if m.selectedMode == engine.ExpandModeReinforce {
			m.subModeOptions = m.buildSubModeOptions()
			m.subModeCursor = 0
			// Skip sub-mode step if only one option.
			if len(m.subModeOptions) == 1 {
				m.selectedSubMode = m.subModeOptions[0].mode
				m.step = expandStepHPAmount
				m.hpAmount = 1
				m.hpMax = m.computeHPMax()
			} else {
				m.step = expandStepSubMode
			}
		} else {
			m.step = expandStepHPAmount
			m.hpAmount = 1
			m.hpMax = m.computeHPMax()
		}
	}
	return m, nil
}

func (m ExpandInfluenceModel) updateSubMode(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		if m.subModeCursor > 0 {
			m.subModeCursor--
		}
	case "down", "j":
		if m.subModeCursor < len(m.subModeOptions)-1 {
			m.subModeCursor++
		}
	case "enter":
		m.selectedSubMode = m.subModeOptions[m.subModeCursor].mode
		m.step = expandStepHPAmount
		m.hpAmount = 1
		m.hpMax = m.computeHPMax()
	}
	return m, nil
}

func (m ExpandInfluenceModel) updateHPAmount(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "+", "=", "up", "k":
		if m.hpAmount < m.hpMax {
			m.hpAmount++
		}
	case "-", "down", "j":
		if m.hpAmount > 1 {
			m.hpAmount--
		}
	case "enter":
		order := engine.ExpandInfluenceOrder{
			Mode:     m.selectedMode,
			World:    m.selectedWorld,
			BaseID:   m.selectedBaseID,
			SubMode:  m.selectedSubMode,
			HPAmount: m.hpAmount,
		}
		return m, func() tea.Msg { return ExpandInfluenceOrderSelectedMsg{Order: order} }
	}
	return m, nil
}

func (m ExpandInfluenceModel) View() string {
	switch m.step {
	case expandStepMode:
		return m.viewMode()
	case expandStepWorld:
		return m.viewWorld()
	case expandStepSubMode:
		return m.viewSubMode()
	case expandStepHPAmount:
		return m.viewHPAmount()
	}
	return ""
}

func (m ExpandInfluenceModel) viewMode() string {
	var sb strings.Builder
	sb.WriteString(style.SectionTitle.Render("Expand Influence"))
	sb.WriteString("\n\n")
	for i, opt := range m.modeOptions {
		cursor := "  "
		if i == m.modeCursor {
			cursor = "> "
		}
		fmt.Fprintf(&sb, "%s%s\n", cursor, opt.label)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Enter select"))
	return sb.String()
}

func (m ExpandInfluenceModel) viewWorld() string {
	var sb strings.Builder
	title := "Expand Influence — Select World"
	if m.selectedMode == engine.ExpandModeReinforce {
		title = "Reinforce Base — Select World"
	}
	sb.WriteString(style.SectionTitle.Render(title))
	sb.WriteString("\n\n")
	for i, opt := range m.worldOptions {
		cursor := "  "
		if i == m.worldCursor {
			cursor = "> "
		}
		label := opt.world
		if opt.baseID != "" {
			if base := expandFindBase(m.faction, opt.baseID); base != nil {
				label = fmt.Sprintf("%-20s  HP %d/%d  Max %d",
					opt.world, base.CurrentHP, base.EffectiveMaxHP(m.faction), base.MaxHP)
			}
		}
		fmt.Fprintf(&sb, "%s%s\n", cursor, label)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Enter select"))
	return sb.String()
}

func (m ExpandInfluenceModel) viewSubMode() string {
	var sb strings.Builder
	base := expandFindBase(m.faction, m.selectedBaseID)
	sb.WriteString(style.SectionTitle.Render("Reinforce: " + m.selectedWorld))
	sb.WriteString("\n")
	if base != nil {
		fmt.Fprintf(&sb, "%s\n\n", style.Muted.Render(
			fmt.Sprintf("HP %d/%d  Max %d  (Faction max: %d)",
				base.CurrentHP, base.EffectiveMaxHP(m.faction), base.MaxHP, m.faction.MaxHP),
		))
	}
	for i, opt := range m.subModeOptions {
		cursor := "  "
		if i == m.subModeCursor {
			cursor = "> "
		}
		fmt.Fprintf(&sb, "%s%s\n", cursor, opt.label)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Enter select"))
	return sb.String()
}

func (m ExpandInfluenceModel) viewHPAmount() string {
	var sb strings.Builder
	title := m.hpAmountTitle()
	sb.WriteString(style.SectionTitle.Render(title))
	sb.WriteString("\n\n")
	fmt.Fprintf(&sb, "HP:   %s  (max %d)\n", style.HP.Render(fmt.Sprintf("%d", m.hpAmount)), m.hpMax)
	fmt.Fprintf(&sb, "Cost: %s\n\n", style.Coin.Render(fmt.Sprintf("%d Coin", m.hpAmount)))
	fmt.Fprintf(&sb, "%s\n\n", style.Muted.Render(fmt.Sprintf("Coin available: %d", m.faction.Coin)))
	sb.WriteString(style.Muted.Render("+/- adjust  Enter confirm"))
	return sb.String()
}

func (m ExpandInfluenceModel) hpAmountTitle() string {
	switch m.selectedMode {
	case engine.ExpandModeNew:
		return fmt.Sprintf("New Base on %s", m.selectedWorld)
	case engine.ExpandModeReinforce:
		switch m.selectedSubMode {
		case engine.ReinforceHeal:
			return fmt.Sprintf("Heal Base on %s", m.selectedWorld)
		case engine.ReinforceMax:
			return fmt.Sprintf("Increase Max HP — %s", m.selectedWorld)
		}
	}
	return "Expand Influence"
}

func (m ExpandInfluenceModel) buildWorldOptions() []expandWorldOption {
	var result []expandWorldOption
	if m.selectedMode == engine.ExpandModeNew {
		for _, world := range expandWorldsForNewBase(m.faction) {
			result = append(result, expandWorldOption{world: world})
		}
	} else {
		seen := map[string]bool{}
		for _, base := range m.faction.Bases {
			if base.IsHomeworld || seen[base.Location] {
				continue
			}
			canHeal := base.CurrentHP < base.EffectiveMaxHP(m.faction)
			canGrow := base.MaxHP < m.faction.MaxHP
			if canHeal || canGrow {
				result = append(result, expandWorldOption{world: base.Location, baseID: base.ID})
				seen[base.Location] = true
			}
		}
	}
	return result
}

func (m ExpandInfluenceModel) buildSubModeOptions() []expandSubModeOption {
	var result []expandSubModeOption
	base := expandFindBase(m.faction, m.selectedBaseID)
	if base == nil {
		return result
	}
	if base.CurrentHP < base.EffectiveMaxHP(m.faction) {
		result = append(result, expandSubModeOption{"Heal HP", engine.ReinforceHeal})
	}
	if base.MaxHP < m.faction.MaxHP {
		result = append(result, expandSubModeOption{"Increase Max HP", engine.ReinforceMax})
	}
	return result
}

func (m ExpandInfluenceModel) computeHPMax() int {
	switch m.selectedMode {
	case engine.ExpandModeNew:
		return min(m.faction.MaxHP, m.faction.Coin)
	case engine.ExpandModeReinforce:
		base := expandFindBase(m.faction, m.selectedBaseID)
		if base == nil {
			return 0
		}
		switch m.selectedSubMode {
		case engine.ReinforceHeal:
			return min(base.EffectiveMaxHP(m.faction)-base.CurrentHP, m.faction.Coin)
		case engine.ReinforceMax:
			return min(m.faction.MaxHP-base.MaxHP, m.faction.Coin)
		}
	}
	return 0
}

// Local helpers — mirror engine/actions helpers; kept here to avoid cross-package coupling.

func expandFindBase(faction *domain.Faction, baseID string) *domain.Base {
	for _, base := range faction.Bases {
		if base.ID == baseID {
			return base
		}
	}
	return nil
}

func expandWorldsForNewBase(faction *domain.Faction) []string {
	withAssets := map[string]struct{}{}
	for _, asset := range faction.Assets {
		withAssets[asset.Location] = struct{}{}
	}
	withBase := map[string]struct{}{}
	for _, base := range faction.Bases {
		withBase[base.Location] = struct{}{}
	}
	var result []string
	for world := range withAssets {
		if _, exists := withBase[world]; !exists {
			result = append(result, world)
		}
	}
	sort.Strings(result)
	return result
}

func expandDamagedBases(faction *domain.Faction) []*domain.Base {
	var result []*domain.Base
	for _, base := range faction.Bases {
		if !base.IsHomeworld && base.CurrentHP < base.EffectiveMaxHP(faction) {
			result = append(result, base)
		}
	}
	return result
}

func expandGrowableBases(faction *domain.Faction) []*domain.Base {
	var result []*domain.Base
	for _, base := range faction.Bases {
		if !base.IsHomeworld && base.MaxHP < faction.MaxHP {
			result = append(result, base)
		}
	}
	return result
}
