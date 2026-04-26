package inputs

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

type AssetSelectedMsg struct{ Asset *domain.Asset }

type assetItem struct {
	asset *domain.Asset
	name  string
}

func (i assetItem) Title() string { return i.name }
func (i assetItem) Description() string {
	return fmt.Sprintf("@ %s  HP %d", i.asset.Location, i.asset.CurrentHP)
}
func (i assetItem) FilterValue() string { return i.name }

type SelectAssetModel struct {
	list list.Model
}

func NewSelectAssetModel(assets []*domain.Asset, rulebook *loader.Rulebook) SelectAssetModel {
	items := make([]list.Item, len(assets))
	for i, asset := range assets {
		name := asset.DefinitionID
		if def, ok := rulebook.Assets[asset.DefinitionID]; ok {
			name = def.Name
		}
		items[i] = assetItem{asset: asset, name: name}
	}
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Select Asset to Sell"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	return SelectAssetModel{list: l}
}

func (m SelectAssetModel) Init() tea.Cmd { return nil }

func (m SelectAssetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-2)
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			if item, ok := m.list.SelectedItem().(assetItem); ok {
				return m, func() tea.Msg { return AssetSelectedMsg{Asset: item.asset} }
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SelectAssetModel) View() string {
	return m.list.View()
}
