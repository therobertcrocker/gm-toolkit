package inputs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type AttackInputsSelectedMsg struct {
	Attackers []*domain.Asset
	Defenders map[string]*domain.Asset // attacker ID → defender
}

type attackStep int

const (
	attackStepSelectAttackers attackStep = iota
	attackStepSelectDefender
)

type atkInfo struct {
	asset *domain.Asset
	name  string
	def   *domain.AssetDefinition
}

type defInfo struct {
	asset     *domain.Asset
	name      string
	ownerName string
}

type AttackInputsModel struct {
	step             attackStep
	attackers        []atkInfo
	atkSelected      []bool
	atkCursor        int
	attackerQueue    []atkInfo
	currentAtkIdx    int
	defenders        []defInfo
	defCursor        int
	selectedDefenders map[string]*domain.Asset
	factionState     *state.FactionState
	rulebook         *loader.Rulebook
}

func NewAttackInputsModel(eligible []*domain.Asset, factionState *state.FactionState, rulebook *loader.Rulebook) AttackInputsModel {
	infos := make([]atkInfo, len(eligible))
	for i, asset := range eligible {
		info := atkInfo{asset: asset, name: asset.DefinitionID}
		if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
			info.name = def.Name
			info.def = def
		}
		infos[i] = info
	}
	return AttackInputsModel{
		attackers:         infos,
		atkSelected:       make([]bool, len(infos)),
		selectedDefenders: make(map[string]*domain.Asset),
		factionState:      factionState,
		rulebook:          rulebook,
	}
}

func (m AttackInputsModel) Init() tea.Cmd { return nil }

func (m AttackInputsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch m.step {
	case attackStepSelectAttackers:
		return m.updateSelectAttackers(key)
	case attackStepSelectDefender:
		return m.updateSelectDefender(key)
	}
	return m, nil
}

func (m AttackInputsModel) updateSelectAttackers(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		if m.atkCursor > 0 {
			m.atkCursor--
		}
	case "down", "j":
		if m.atkCursor < len(m.attackers)-1 {
			m.atkCursor++
		}
	case " ":
		m.atkSelected[m.atkCursor] = !m.atkSelected[m.atkCursor]
	case "enter":
		var queue []atkInfo
		for i, info := range m.attackers {
			if m.atkSelected[i] {
				queue = append(queue, info)
			}
		}
		if len(queue) == 0 {
			return m, nil
		}
		m.attackerQueue = queue
		m.currentAtkIdx = 0
		m.step = attackStepSelectDefender
		m.defenders = m.buildDefenders(queue[0])
		m.defCursor = 0
	}
	return m, nil
}

func (m AttackInputsModel) updateSelectDefender(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		if m.defCursor > 0 {
			m.defCursor--
		}
	case "down", "j":
		if m.defCursor < len(m.defenders)-1 {
			m.defCursor++
		}
	case "enter":
		if len(m.defenders) == 0 {
			break
		}
		chosen := m.defenders[m.defCursor]
		m.selectedDefenders[m.attackerQueue[m.currentAtkIdx].asset.ID] = chosen.asset
		m.currentAtkIdx++
		if m.currentAtkIdx < len(m.attackerQueue) {
			m.defenders = m.buildDefenders(m.attackerQueue[m.currentAtkIdx])
			m.defCursor = 0
			break
		}
		attackers := make([]*domain.Asset, len(m.attackerQueue))
		for i, info := range m.attackerQueue {
			attackers[i] = info.asset
		}
		msg := AttackInputsSelectedMsg{Attackers: attackers, Defenders: m.selectedDefenders}
		return m, func() tea.Msg { return msg }
	}
	return m, nil
}

func (m AttackInputsModel) buildDefenders(attacker atkInfo) []defInfo {
	var result []defInfo
	for factionID, faction := range m.factionState.Factions {
		if factionID == attacker.asset.OwnerID {
			continue
		}
		for _, asset := range faction.Assets {
			if asset.Location != attacker.asset.Location {
				continue
			}
			if asset.Stealthy || !asset.Ready || asset.CurrentHP <= 0 || !asset.Maintained {
				continue
			}
			name := asset.DefinitionID
			if def, ok := m.rulebook.Assets[asset.DefinitionID]; ok {
				name = def.Name
			}
			result = append(result, defInfo{asset: asset, name: name, ownerName: faction.Name})
		}
	}
	return result
}

func (m AttackInputsModel) View() string {
	switch m.step {
	case attackStepSelectAttackers:
		return m.viewSelectAttackers()
	case attackStepSelectDefender:
		return m.viewSelectDefender()
	}
	return ""
}

func (m AttackInputsModel) viewSelectAttackers() string {
	var sb strings.Builder
	sb.WriteString(style.SectionTitle.Render("Attack — Select Attackers"))
	sb.WriteString("\n\n")
	for i, info := range m.attackers {
		cursor := "  "
		if i == m.atkCursor {
			cursor = "> "
		}
		check := "[ ]"
		if m.atkSelected[i] {
			check = style.HP.Render("[x]")
		}
		atkStr := ""
		if info.def != nil && info.def.Attack != nil {
			atkStr = fmt.Sprintf("  %s vs %s  %s",
				info.def.Attack.AttackerStat,
				info.def.Attack.DefenderStat,
				diceStr(info.def.Attack.Damage),
			)
		}
		fmt.Fprintf(&sb, "%s%s %-20s HP %d @ %s%s\n",
			cursor, check, info.name,
			info.asset.CurrentHP, info.asset.Location,
			style.Muted.Render(atkStr),
		)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Space select  Enter confirm"))
	return sb.String()
}

func (m AttackInputsModel) viewSelectDefender() string {
	if m.currentAtkIdx >= len(m.attackerQueue) {
		return ""
	}
	current := m.attackerQueue[m.currentAtkIdx]
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", style.SectionTitle.Render("Attack — Select Defender"))
	fmt.Fprintf(&sb, "Attacker: %s  (%d of %d)\n\n",
		current.name, m.currentAtkIdx+1, len(m.attackerQueue),
	)
	if len(m.defenders) == 0 {
		sb.WriteString(style.Muted.Render("No eligible defenders on this world."))
		return sb.String()
	}
	for i, info := range m.defenders {
		cursor := "  "
		if i == m.defCursor {
			cursor = "> "
		}
		fmt.Fprintf(&sb, "%s%-20s  HP %d  (%s)\n",
			cursor, info.name, info.asset.CurrentHP, info.ownerName,
		)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Enter select"))
	return sb.String()
}

func diceStr(d domain.DiceRoll) string {
	if d.NumDice == 0 {
		return fmt.Sprintf("%d", d.Modifier)
	}
	s := fmt.Sprintf("%dd%d", d.NumDice, d.Sides)
	if d.Modifier > 0 {
		s += fmt.Sprintf("+%d", d.Modifier)
	} else if d.Modifier < 0 {
		s += fmt.Sprintf("%d", d.Modifier)
	}
	return s
}
