package detail

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/msgs"
)

const hpBarWidth = 16

type keyMap struct {
	Scroll key.Binding
	Delete key.Binding
	Back   key.Binding
}

type Model struct {
	faction  *domain.Faction
	rulebook *rulebook.Rulebook
	viewport viewport.Model
	width    int
	height   int
	keys     keyMap
}

func New(faction *domain.Faction, rb *rulebook.Rulebook, width, height int) Model {
	m := Model{
		faction:  faction,
		rulebook: rb,
		width:    width,
		height:   height,
		keys: keyMap{
			Scroll: key.NewBinding(key.WithKeys("up", "down", "j", "k"), key.WithHelp("↑/↓", "scroll")),
			Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
			Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		},
	}
	m.viewport = viewport.New(width, height)
	m.viewport.SetContent(m.renderBody())
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		m.viewport.SetContent(m.renderBody())
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Delete):
			return m, func() tea.Msg { return msgs.RequestDeleteMsg{Faction: m.faction} }
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return msgs.CancelMsg{} }
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.faction == nil {
		return "(no faction)"
	}
	return m.viewport.View()
}

func (m Model) renderBody() string {
	if m.faction == nil {
		return ""
	}
	f := m.faction
	var b strings.Builder

	b.WriteString(styles.Strong.Render(f.Name) + styles.Dim.Render("  "+f.ID) + "\n")
	b.WriteString(styles.Subtle.Render(titleCase(string(f.Scale))) + "\n\n")
	b.WriteString(styles.Subtle.Render("Homeworld  ") + styles.Strong.Render(f.Homeworld.WorldID) + "\n\n")

	b.WriteString(m.statRow(f) + "\n\n")
	b.WriteString(m.metaRow(f) + "\n")

	b.WriteString("\n" + m.section("TAGS") + "\n")
	if len(f.Tags) == 0 {
		b.WriteString("  " + styles.Dim.Render("none") + "\n")
	}
	for _, tag := range f.Tags {
		b.WriteString("  " + styles.Strong.Render(tag.Name) + "\n")
		b.WriteString(m.wrap(tag.Description, 4, styles.Label) + "\n")
		if tag.Effect != "" {
			b.WriteString(m.wrap("Effect: "+tag.Effect, 4, styles.Muted) + "\n")
		}
	}

	b.WriteString("\n" + m.section("GOAL") + "\n")
	if f.ActiveGoal != nil {
		goalName := f.ActiveGoal.GoalID
		var goalDesc string
		if goal, ok := m.rulebook.Goals[f.ActiveGoal.GoalID]; ok {
			goalName = goal.Name
			goalDesc = goal.Description
		}
		b.WriteString("  " + styles.Strong.Render(goalName) +
			styles.Dim.Render(fmt.Sprintf("   ·   progress %d", f.ActiveGoal.Progress)) + "\n")
		if goalDesc != "" {
			b.WriteString(m.wrap(goalDesc, 4, styles.Label) + "\n")
		}
	} else {
		b.WriteString("  " + styles.Dim.Render("none") + "\n")
	}

	b.WriteString("\n" + m.section("ASSETS") + "\n")
	if len(f.Assets) == 0 {
		b.WriteString("  " + styles.Dim.Render("none") + "\n")
	}
	for _, asset := range domain.SortedAssets(f) {
		name := asset.DefinitionID
		if def, ok := m.rulebook.Assets[asset.DefinitionID]; ok {
			name = def.Name
		}
		b.WriteString("  " + styles.Dim.Render("• ") + styles.Strong.Render(name) + "\n")
	}

	b.WriteString("\n" + m.section("BASES") + "\n")
	if len(f.Bases) == 0 {
		b.WriteString("  " + styles.Dim.Render("none") + "\n")
	}
	for _, base := range f.Bases {
		b.WriteString("  " + styles.Dim.Render("• ") + styles.Strong.Render(base.Location.WorldID) + "\n")
	}

	return b.String()
}

func (m Model) statRow(f *domain.Faction) string {
	stats := []struct {
		name string
		val  int
	}{
		{"Force", f.Force},
		{"Cunning", f.Cunning},
		{"Wealth", f.Wealth},
	}
	primary := f.Force
	for _, s := range stats {
		if s.val > primary {
			primary = s.val
		}
	}
	parts := make([]string, 0, len(stats))
	for _, s := range stats {
		statStyle := styles.Strong
		if s.val == primary {
			statStyle = styles.AccentAlt
		}
		parts = append(parts, styles.Subtle.Render(s.name+" ")+statStyle.Render(strconv.Itoa(s.val)))
	}
	return strings.Join(parts, "    ")
}

func (m Model) metaRow(f *domain.Faction) string {
	hp := styles.Subtle.Render("HP ") + hpBar(f.CurrentHP, f.MaxHP) +
		"  " + styles.Strong.Render(fmt.Sprintf("%d/%d", f.CurrentHP, f.MaxHP))
	coin := styles.Subtle.Render("Coin ") + styles.Strong.Render(strconv.Itoa(f.Coin))
	xp := styles.Subtle.Render("XP ") + styles.Strong.Render(strconv.Itoa(f.XP))
	return hp + "\n\n" + coin + "\n" + xp
}

func hpBar(current, maximum int) string {
	if maximum < 1 {
		maximum = 1
	}
	ratio := float64(current) / float64(maximum)
	ratio = math.Max(0, math.Min(1, ratio))
	filled := int(math.Round(ratio * hpBarWidth))

	return lipgloss.NewStyle().Foreground(styles.HealthColor(ratio)).Render(strings.Repeat("█", filled)) +
		styles.Dim.Render(strings.Repeat("░", hpBarWidth-filled))
}

func (m Model) section(title string) string {
	head := styles.SectionHeading.Render(title)
	ruleLen := max(m.contentWidth()-lipgloss.Width(head)-1, 0)
	return head + " " + styles.Rule.Render(strings.Repeat("─", ruleLen))
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (m Model) wrap(text string, indent int, color lipgloss.Color) string {
	return lipgloss.NewStyle().
		Width(m.contentWidth()).
		PaddingLeft(indent).
		Foreground(color).
		Render(text)
}

func (m Model) contentWidth() int {
	if m.width > 0 {
		return m.width
	}
	return 80
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }

func (h helpKeys) ShortHelp() []key.Binding {
	return []key.Binding{h.keys.Scroll, h.keys.Delete, h.keys.Back}
}
func (h helpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{h.keys.Scroll, h.keys.Delete, h.keys.Back}}
}
