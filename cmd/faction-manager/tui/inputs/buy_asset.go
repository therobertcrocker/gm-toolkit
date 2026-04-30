package inputs

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

type BuyOrderSelectedMsg struct {
	Order         engine.BuyOrder
	StealthTarget *domain.Asset // non-nil only when buying C3-002 with an eligible target
}

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

type stealthTargetItem struct {
	asset *domain.Asset
	def   *domain.AssetDefinition
}

func (i stealthTargetItem) Title() string       { return i.def.Name }
func (i stealthTargetItem) Description() string { return fmt.Sprintf("ID: %s  HP: %d", i.asset.ID, i.asset.CurrentHP) }
func (i stealthTargetItem) FilterValue() string { return i.def.Name }

type BuyOrderModel struct {
	step          int
	worlds        []string
	purchasable   []*domain.AssetDefinition
	faction       *domain.Faction
	rulebook      *loader.Rulebook
	worldList     list.Model
	defList       list.Model
	stealthList   list.Model
	selectedWorld string
	selectedDef   *domain.AssetDefinition
	width         int
	height        int
}

func NewBuyOrderModel(worlds []string, purchasable []*domain.AssetDefinition, faction *domain.Faction, rulebook *loader.Rulebook) BuyOrderModel {
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

	sList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	sList.SetShowHelp(false)
	sList.SetFilteringEnabled(false)
	sList.DisableQuitKeybindings()

	return BuyOrderModel{
		worlds:      worlds,
		purchasable: sorted,
		faction:     faction,
		rulebook:    rulebook,
		worldList:   wList,
		defList:     dList,
		stealthList: sList,
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
		m.stealthList.SetSize(msg.Width, msg.Height-2)
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			switch m.step {
			case 0:
				if item, ok := m.worldList.SelectedItem().(worldItem); ok {
					m.selectedWorld = item.world
					m.step = 1
					return m, nil
				}
			case 1:
				if item, ok := m.defList.SelectedItem().(defItem); ok {
					m.selectedDef = item.def
					order := engine.BuyOrder{World: m.selectedWorld, Definition: item.def}

					if item.def.ID != "C3-002" {
						return m, func() tea.Msg { return BuyOrderSelectedMsg{Order: order} }
					}

					targets := eligibleStealthTargets(m.faction, m.selectedWorld, m.rulebook)
					switch len(targets) {
					case 0:
						return m, func() tea.Msg { return BuyOrderSelectedMsg{Order: order} }
					case 1:
						t := targets[0]
						return m, func() tea.Msg { return BuyOrderSelectedMsg{Order: order, StealthTarget: t} }
					default:
						m.stealthList = buildStealthList(targets, m.rulebook)
						m.stealthList.SetSize(m.width, m.height-2)
						m.step = 2
						return m, nil
					}
				}
			case 2:
				if item, ok := m.stealthList.SelectedItem().(stealthTargetItem); ok {
					order := engine.BuyOrder{World: m.selectedWorld, Definition: m.selectedDef}
					t := item.asset
					return m, func() tea.Msg { return BuyOrderSelectedMsg{Order: order, StealthTarget: t} }
				}
			}
		}
	}
	var cmd tea.Cmd
	switch m.step {
	case 0:
		m.worldList, cmd = m.worldList.Update(msg)
	case 1:
		m.defList, cmd = m.defList.Update(msg)
	case 2:
		m.stealthList, cmd = m.stealthList.Update(msg)
	}
	return m, cmd
}

func (m BuyOrderModel) View() string {
	switch m.step {
	case 0:
		return m.worldList.View()
	case 1:
		return m.defList.View()
	default:
		return m.stealthList.View()
	}
}

func buildStealthList(targets []*domain.Asset, rulebook *loader.Rulebook) list.Model {
	items := make([]list.Item, 0, len(targets))
	for _, asset := range targets {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			continue
		}
		items = append(items, stealthTargetItem{asset: asset, def: def})
	}
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Apply Stealth to which asset?"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	return l
}

func eligibleStealthTargets(faction *domain.Faction, world string, rulebook *loader.Rulebook) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range faction.Assets {
		if asset.Location != world || asset.Stealthy {
			continue
		}
		def, ok := rulebook.Assets[asset.DefinitionID]
		if ok && def.Type == domain.TypeSpecialForces {
			result = append(result, asset)
		}
	}
	return result
}
