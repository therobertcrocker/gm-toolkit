package phases

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ResumeChoice int

const (
	ChoiceResume ResumeChoice = iota
	ChoiceStartNew
	ChoiceAbandon
)

type ResumeChoiceMsg struct{ Choice ResumeChoice }

type resumeItem struct {
	title  string
	choice ResumeChoice
}

func (i resumeItem) Title() string       { return i.title }
func (i resumeItem) Description() string { return "" }
func (i resumeItem) FilterValue() string { return i.title }

type ResumeTurnModel struct {
	list list.Model
}

func NewResumeTurnModel(inProgress bool) ResumeTurnModel {
	var items []list.Item
	if inProgress {
		items = []list.Item{
			resumeItem{"Resume Turn", ChoiceResume},
			resumeItem{"Start New Turn", ChoiceStartNew},
			resumeItem{"Abandon Turn", ChoiceAbandon},
		}
	} else {
		items = []list.Item{
			resumeItem{"Start New Turn", ChoiceStartNew},
		}
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Faction Turn"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	return ResumeTurnModel{list: l}
}

func (m ResumeTurnModel) Init() tea.Cmd { return nil }

func (m ResumeTurnModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-2)
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			if item, ok := m.list.SelectedItem().(resumeItem); ok {
				return m, func() tea.Msg { return ResumeChoiceMsg{Choice: item.choice} }
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ResumeTurnModel) View() string {
	return m.list.View()
}
