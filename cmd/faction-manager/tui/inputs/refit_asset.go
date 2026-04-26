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

type RefitOrderSelectedMsg struct{ Order engine.RefitOrder }

type refitAssetItem struct {
	option engine.RefitOption
	name   string
}

func (i refitAssetItem) Title() string { return i.name }
func (i refitAssetItem) Description() string {
	return fmt.Sprintf("@ %s  HP %d", i.option.Asset.Location, i.option.Asset.CurrentHP)
}
func (i refitAssetItem) FilterValue() string { return i.name }

type replacementItem struct {
	def     *domain.AssetDefinition
	oldCost int
}

func (i replacementItem) Title() string { return i.def.Name }
func (i replacementItem) Description() string {
	delta := i.def.Cost - i.oldCost
	if delta > 0 {
		return fmt.Sprintf("Cost: %d  (+%d)", i.def.Cost, delta)
	}
	return fmt.Sprintf("Cost: %d", i.def.Cost)
}
func (i replacementItem) FilterValue() string { return i.def.Name }

type RefitOrderModel struct {
	step           int
	options        []engine.RefitOption
	rulebook       *loader.Rulebook
	assetList      list.Model
	replList       list.Model
	selectedOption engine.RefitOption
	width          int
	height         int
}

func NewRefitOrderModel(options []engine.RefitOption, rulebook *loader.Rulebook) RefitOrderModel {
	aItems := make([]list.Item, len(options))
	for i, opt := range options {
		name := opt.Asset.DefinitionID
		if def, ok := rulebook.Assets[opt.Asset.DefinitionID]; ok {
			name = def.Name
		}
		aItems[i] = refitAssetItem{option: opt, name: name}
	}
	aDelegate := list.NewDefaultDelegate()
	aList := list.New(aItems, aDelegate, 0, 0)
	aList.Title = "Select Asset to Refit"
	aList.SetShowHelp(false)
	aList.SetFilteringEnabled(false)
	aList.DisableQuitKeybindings()

	rDelegate := list.NewDefaultDelegate()
	rList := list.New(nil, rDelegate, 0, 0)
	rList.Title = "Select Replacement"
	rList.SetShowHelp(false)
	rList.SetFilteringEnabled(false)
	rList.DisableQuitKeybindings()

	return RefitOrderModel{
		options:   options,
		rulebook:  rulebook,
		assetList: aList,
		replList:  rList,
	}
}

func (m RefitOrderModel) Init() tea.Cmd { return nil }

func (m RefitOrderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.assetList.SetSize(msg.Width, msg.Height-2)
		m.replList.SetSize(msg.Width, msg.Height-2)
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			if m.step == 0 {
				if item, ok := m.assetList.SelectedItem().(refitAssetItem); ok {
					m.selectedOption = item.option
					m.step = 1
					oldDef := m.rulebook.Assets[item.option.Asset.DefinitionID]
					replacements := make([]*domain.AssetDefinition, len(item.option.Replacements))
					copy(replacements, item.option.Replacements)
					sort.Slice(replacements, func(i, j int) bool { return replacements[i].Name < replacements[j].Name })
					rItems := make([]list.Item, len(replacements))
					for i, def := range replacements {
						oldCost := 0
						if oldDef != nil {
							oldCost = oldDef.Cost
						}
						rItems[i] = replacementItem{def: def, oldCost: oldCost}
					}
					m.replList.SetItems(rItems)
					if m.width > 0 {
						m.replList.SetSize(m.width, m.height-2)
					}
					return m, nil
				}
			} else {
				if item, ok := m.replList.SelectedItem().(replacementItem); ok {
					order := engine.RefitOrder{
						OldAsset:      m.selectedOption.Asset,
						NewDefinition: item.def,
					}
					return m, func() tea.Msg { return RefitOrderSelectedMsg{Order: order} }
				}
			}
		}
	}
	var cmd tea.Cmd
	if m.step == 0 {
		m.assetList, cmd = m.assetList.Update(msg)
	} else {
		m.replList, cmd = m.replList.Update(msg)
	}
	return m, cmd
}

func (m RefitOrderModel) View() string {
	if m.step == 0 {
		return m.assetList.View()
	}
	return m.replList.View()
}
