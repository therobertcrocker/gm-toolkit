package manage

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/deleteconfirm"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/detail"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/list"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/msgs"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/wizard"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type view int

const (
	viewList view = iota
	viewDetail
	viewCreate
	viewDeleteConfirm
)

var backTarget = map[view]view{
	viewDetail:        viewList,
	viewCreate:        viewList,
	viewDeleteConfirm: viewDetail,
}

type Model struct {
	factionState *state.FactionState
	paths        *campaigns.Paths
	rulebook     *rulebook.Rulebook
	spatialMap   *spatial.RegionMap

	view          view
	list          list.Model
	detail        detail.Model
	create        wizard.Model
	deleteConfirm deleteconfirm.Model

	termWidth  int
	termHeight int
	saveErr    error
}

func New(
	factionState *state.FactionState,
	paths *campaigns.Paths,
	rb *rulebook.Rulebook,
	spatialMap *spatial.RegionMap,
) Model {
	return Model{
		factionState: factionState,
		paths:        paths,
		rulebook:     rb,
		spatialMap:   spatialMap,
		view:         viewList,
		list:         list.New(factionState),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.list, _ = m.list.Update(tea.WindowSizeMsg{Width: regionWidths(msg.Width)[Left], Height: msg.Height})
		if m.view == viewDetail {
			var cmd tea.Cmd
			m.detail, cmd = m.detail.Update(msg)
			return m, cmd
		}
		if m.view != viewList {
			return m.routeForward(msg)
		}
		return m, nil

	case msgs.CreatedMsg:
		if err := state.CreateFaction(m.paths.StatePath, m.factionState, msg.Faction); err != nil {
			m.saveErr = err
			return m, nil
		}
		m.saveErr = nil
		m.list = list.New(m.factionState)
		if m.termWidth > 0 {
			m.list, _ = m.list.Update(tea.WindowSizeMsg{Width: regionWidths(m.termWidth)[Left], Height: m.termHeight})
		}
		m.view = backTarget[viewCreate]
		return m, nil

	case msgs.DeletedMsg:
		if err := state.DeleteFaction(m.paths.StatePath, m.factionState, msg.ID); err != nil {
			m.saveErr = err
			return m, nil
		}
		m.saveErr = nil
		m.list = list.New(m.factionState)
		if m.termWidth > 0 {
			m.list, _ = m.list.Update(tea.WindowSizeMsg{Width: regionWidths(m.termWidth)[Left], Height: m.termHeight})
		}
		m.view = viewList
		return m, nil

	case msgs.CancelMsg:
		m.view = backTarget[m.view]
		return m, nil

	case msgs.RequestDetailMsg:
		faction, ok := m.factionState.Factions[msg.FactionID]
		if !ok {
			return m, nil
		}
		m.detail = detail.New(faction, m.rulebook, m.termWidth, m.termHeight)
		m.view = viewDetail
		return m, nil

	case msgs.RequestCreateMsg:
		m.create = wizard.New(m.rulebook, m.spatialMap, m.factionState)
		m.view = viewCreate
		return m, m.create.Init()

	case msgs.RequestDeleteMsg:
		m.deleteConfirm = deleteconfirm.New(msg.Faction)
		m.view = viewDeleteConfirm
		return m, nil
	}

	return m.routeForward(msg)
}

func (m Model) routeForward(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.view {
	case viewList:
		m.list, cmd = m.list.Update(msg)
	case viewDetail:
		m.detail, cmd = m.detail.Update(msg)
	case viewCreate:
		m.create, cmd = m.create.Update(msg)
	case viewDeleteConfirm:
		m.deleteConfirm, cmd = m.deleteConfirm.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	var content string
	switch m.view {
	case viewList:
		contentH := max(m.termHeight, 1)
		placeholder := styles.Dim.Align(lipgloss.Center).AlignVertical(lipgloss.Center)
		panels := map[Region]panel{
			Left:   {content: m.list.View(), style: lipgloss.NewStyle()},
			Center: {content: "—", style: placeholder},
			Right:  {content: "—", style: placeholder},
		}
		content = compose(panels, regionWidths(m.termWidth), contentH)
	case viewDetail:
		content = m.detail.View()
	case viewCreate:
		content = m.create.View()
	case viewDeleteConfirm:
		content = m.deleteConfirm.View()
	}
	if m.saveErr != nil {
		content += "\n\n" + styles.SaveError.Render("save error: "+m.saveErr.Error())
	}
	return content
}

// CapturesInput reports whether the active sub-view owns the full keyboard.
// The creation wizard has free-text fields (name, Coin) and its own field
// navigation, so the root must not steal q/tab/? while it is up.
func (m Model) CapturesInput() bool {
	return m.view == viewCreate
}

func (m Model) Help() help.KeyMap {
	switch m.view {
	case viewList:
		return m.list.Help()
	case viewDetail:
		return m.detail.Help()
	case viewCreate:
		return m.create.Help()
	case viewDeleteConfirm:
		return m.deleteConfirm.Help()
	}
	return nil
}
