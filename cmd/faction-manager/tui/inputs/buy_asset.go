package inputs

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
)

type BuyOrderSelectedMsg struct{ Order engine.BuyOrder }

type worldItem struct {
	world string
}

func (i worldItem) Title() string       { return i.world }
func (i worldItem) Description() string { return "" }
func (i worldItem) FilterValue() string { return i.world }

type defItem struct {
	def *domain.AssetDefinition
}

func (i defItem) Title() string { return i.def.Name }
func (i defItem) Description() string {
	return fmt.Sprintf("%s  Cost: %d  Min Rating: %d", i.def.Category, i.def.Cost, i.def.MinRating)
}
func (i defItem) FilterValue() string { return i.def.Name }

type BuyOrderModel struct {
	step          int
	worlds        []string
	purchasable   []*domain.AssetDefinition
	worldList     list.Model
	defList       list.Model
	selectedWorld string
	width         int
	height        int
}

func NewBuyOrderModel(worlds []string, purchasable []*domain.AssetDefinition) BuyOrderModel {
	sorted := make([]*domain.AssetDefinition, len(purchasable))
	copy(sorted, purchasable)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	wItems := make([]list.Item, len(worlds))
	for i, world := range worlds {
		wItems[i] = worldItem{world: world}
	}
	wDelegate := list.NewDefaultDelegate()
	wList := list.New(wItems, wDelegate, 0, 0)
	wList.Title = "Select World"
	wList.SetShowHelp(false)
	wList.SetFilteringEnabled(false)
	wList.DisableQuitKeybindings()

	dItems := make([]list.Item, len(sorted))
	for i, def := range sorted {
		dItems[i] = defItem{def: def}
	}
	dDelegate := list.NewDefaultDelegate()
	dList := list.New(dItems, dDelegate, 0, 0)
	dList.Title = "Select Asset to Buy"
	dList.SetShowHelp(false)
	dList.SetFilteringEnabled(false)
	dList.DisableQuitKeybindings()

	return BuyOrderModel{
		worlds:      worlds,
		purchasable: sorted,
		worldList:   wList,
		defList:     dList,
	}
}

func (m BuyOrderModel) Init() tea.Cmd { return nil }

func (m BuyOrderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.worldList.SetSize(msg.Width, msg.Height-2)
		m.defList.SetSize(msg.Width, msg.Height-2)
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			if m.step == 0 {
				if item, ok := m.worldList.SelectedItem().(worldItem); ok {
					m.selectedWorld = item.world
					m.step = 1
					return m, nil
				}
			} else {
				if item, ok := m.defList.SelectedItem().(defItem); ok {
					order := engine.BuyOrder{World: m.selectedWorld, Definition: item.def}
					return m, func() tea.Msg { return BuyOrderSelectedMsg{Order: order} }
				}
			}
		}
	}
	var cmd tea.Cmd
	if m.step == 0 {
		m.worldList, cmd = m.worldList.Update(msg)
	} else {
		m.defList, cmd = m.defList.Update(msg)
	}
	return m, cmd
}

func (m BuyOrderModel) View() string {
	if m.step == 0 {
		return m.worldList.View()
	}
	return m.defList.View()
}
