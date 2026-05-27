package wizard

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/msgs"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type Model struct {
	form         *huh.Form
	rulebook     *rulebook.Rulebook
	spatialMap   *spatial.RegionMap
	factionState *state.FactionState

	name            string
	scale           domain.FactionScale
	primaryStat     domain.FactionStat
	secondaryStat   domain.FactionStat
	tertiaryStat    domain.FactionStat
	selectedTags    []*domain.Tag
	selectedGoal    *domain.Goal
	homeworldID     string
	primaryAssets   []*domain.AssetDefinition
	secondaryAssets []*domain.AssetDefinition
	coinStr         string

	confirmingDiscard bool
	discardKeys       discardKeyMap
}

type discardKeyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
}

func New(rb *rulebook.Rulebook, spatialMap *spatial.RegionMap, factionState *state.FactionState) Model {
	m := Model{
		rulebook:     rb,
		spatialMap:   spatialMap,
		factionState: factionState,
		scale:        domain.ScaleMinor,
		primaryStat:  domain.StatForce,
		secondaryStat: domain.StatCunning,
		tertiaryStat:  domain.StatWealth,
		coinStr:      "0",
		discardKeys: discardKeyMap{
			Confirm: key.NewBinding(key.WithKeys("y")),
			Cancel:  key.NewBinding(key.WithKeys("n", "esc")),
		},
	}
	m.form = buildForm(&m, rb, spatialMap)
	return m
}

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.confirmingDiscard {
		if km, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(km, m.discardKeys.Confirm):
				return m, func() tea.Msg { return msgs.CancelMsg{} }
			case key.Matches(km, m.discardKeys.Cancel):
				m.confirmingDiscard = false
				return m, nil
			}
		}
		return m, nil
	}

	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "esc" {
		m.confirmingDiscard = true
		return m, nil
	}

	formModel, cmd := m.form.Update(msg)
	m.form = formModel.(*huh.Form)
	if m.form.State == huh.StateCompleted {
		faction := m.synthesize()
		return m, func() tea.Msg { return msgs.CreatedMsg{Faction: faction} }
	}
	return m, cmd
}

func (m Model) View() string {
	if m.confirmingDiscard {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Padding(1, 2).
			Render("Discard new faction? [y/N]")
	}
	return m.form.View()
}

func (m Model) Help() help.KeyMap { return wizardHelpKeys{} }

func (m *Model) synthesize() *domain.Faction {
	coin, _ := strconv.Atoi(m.coinStr)

	var homeworld domain.Location
	homeworld.WorldID = m.homeworldID
	if loc, ok := m.spatialMap.Location(m.homeworldID); ok {
		if rl, ok := loc.(spatial.RegionLocation); ok {
			homeworld.RegionHex = rl.RegionHex()
		}
	}

	id := generateID(m.name, m.factionState.Factions)

	tempFaction := &domain.Faction{
		ID:     id,
		Assets: make(map[string]*domain.Asset),
	}
	allPicked := append(m.primaryAssets, m.secondaryAssets...)
	assets := instantiateAssets(allPicked, homeworld, tempFaction)

	return domain.NewFaction(
		id, m.name,
		m.scale,
		m.primaryStat, m.secondaryStat, m.tertiaryStat,
		m.selectedTags,
		m.selectedGoal,
		homeworld,
		assets,
		coin,
	)
}

// ---------------------------------------------------------------------------
// Form builder
// ---------------------------------------------------------------------------

func buildForm(m *Model, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) *huh.Form {
	statOptions := []huh.Option[domain.FactionStat]{
		huh.NewOption("Force", domain.StatForce),
		huh.NewOption("Cunning", domain.StatCunning),
		huh.NewOption("Wealth", domain.StatWealth),
	}

	nameGroup := huh.NewGroup(
		huh.NewInput().
			Title("Faction name").
			Value(&m.name).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("name required")
				}
				return nil
			}),
	)

	scaleGroup := huh.NewGroup(
		huh.NewSelect[domain.FactionScale]().
			Title("Scale").
			Options(
				huh.NewOption("Minor", domain.ScaleMinor),
				huh.NewOption("Major", domain.ScaleMajor),
				huh.NewOption("Hegemon", domain.ScaleHegemon),
			).
			Value(&m.scale),
	)

	statsGroup := huh.NewGroup(
		huh.NewSelect[domain.FactionStat]().
			Title("Primary attribute").
			Options(statOptions...).
			Value(&m.primaryStat),
		huh.NewSelect[domain.FactionStat]().
			Title("Secondary attribute").
			Options(statOptions...).
			Value(&m.secondaryStat),
		huh.NewSelect[domain.FactionStat]().
			Title("Tertiary attribute").
			Options(statOptions...).
			Value(&m.tertiaryStat).
			Validate(func(domain.FactionStat) error {
				if m.primaryStat == m.secondaryStat || m.primaryStat == m.tertiaryStat || m.secondaryStat == m.tertiaryStat {
					return fmt.Errorf("each attribute must be distinct")
				}
				return nil
			}),
	)

	hpGroup := huh.NewGroup(
		huh.NewNote().
			Title("Max HP").
			DescriptionFunc(func() string {
				return fmt.Sprintf("%d HP (derived from attribute ratings)", previewMaxHP(m))
			}, &m.scale),
	)

	tagsGroup := huh.NewGroup(
		huh.NewMultiSelect[*domain.Tag]().
			Title("Tags (choose up to 2)").
			Options(tagsToOptions(rb.Tags)...).
			Value(&m.selectedTags).
			Validate(func(picked []*domain.Tag) error {
				if len(picked) > 2 {
					return fmt.Errorf("at most 2 tags")
				}
				return nil
			}),
	)

	goalGroup := huh.NewGroup(
		huh.NewSelect[*domain.Goal]().
			Title("Starting goal").
			Options(goalsToOptions(rb.Goals)...).
			Value(&m.selectedGoal),
	)

	homeworldGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Homeworld").
			Options(worldsToOptions(spatialMap)...).
			Value(&m.homeworldID),
	)

	primaryAssetsGroup, secondaryAssetsGroup := buildAssetGroups(m, rb, spatialMap)

	coinGroup := huh.NewGroup(
		huh.NewInput().
			Title("Starting Coin").
			Value(&m.coinStr).
			Validate(func(s string) error {
				if _, err := strconv.Atoi(s); err != nil {
					return fmt.Errorf("must be an integer")
				}
				return nil
			}),
	)

	return huh.NewForm(
		nameGroup,
		scaleGroup,
		statsGroup,
		hpGroup,
		tagsGroup,
		goalGroup,
		homeworldGroup,
		primaryAssetsGroup,
		secondaryAssetsGroup,
		coinGroup,
	)
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func tagsToOptions(tags map[string]*domain.Tag) []huh.Option[*domain.Tag] {
	sorted := make([]*domain.Tag, 0, len(tags))
	for _, tag := range tags {
		sorted = append(sorted, tag)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	options := make([]huh.Option[*domain.Tag], len(sorted))
	for i, tag := range sorted {
		options[i] = huh.NewOption(tag.Name, tag)
	}
	return options
}

func goalsToOptions(goals map[string]*domain.Goal) []huh.Option[*domain.Goal] {
	sorted := make([]*domain.Goal, 0, len(goals))
	for _, goal := range goals {
		sorted = append(sorted, goal)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	options := make([]huh.Option[*domain.Goal], len(sorted))
	for i, goal := range sorted {
		options[i] = huh.NewOption(goal.Name, goal)
	}
	return options
}

func worldsToOptions(spatialMap *spatial.RegionMap) []huh.Option[string] {
	worlds := spatialMap.AllWorlds()
	sort.Slice(worlds, func(i, j int) bool { return worlds[i].Name() < worlds[j].Name() })

	options := make([]huh.Option[string], len(worlds))
	for i, world := range worlds {
		label := fmt.Sprintf("%s (TL %d)", world.Name(), world.TechLevel())
		options[i] = huh.NewOption(label, world.ID())
	}
	return options
}

// buildAssetGroups returns two huh.Group values: one for primary-attribute assets
// (filtered to the faction's primary stat category), one for any-attribute assets.
// Quota enforcement uses the scale-derived counts from AssetCountsFromScale.
func buildAssetGroups(m *Model, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) (*huh.Group, *huh.Group) {
	allDefs := make([]*domain.AssetDefinition, 0, len(rb.Assets))
	for _, def := range rb.Assets {
		allDefs = append(allDefs, def)
	}
	sort.Slice(allDefs, func(i, j int) bool { return allDefs[i].Name < allDefs[j].Name })

	primaryGroup := huh.NewGroup(
		huh.NewMultiSelect[*domain.AssetDefinition]().
			Title("Starting assets — primary attribute").
			OptionsFunc(func() []huh.Option[*domain.AssetDefinition] {
				return primaryAssetOptions(m, allDefs, spatialMap)
			}, &m.primaryStat).
			Value(&m.primaryAssets).
			Validate(func(picked []*domain.AssetDefinition) error {
				primaryCount, _ := domain.AssetCountsFromScale(m.scale)
				if len(picked) > primaryCount {
					return fmt.Errorf("choose at most %d primary-attribute assets for %s scale", primaryCount, m.scale)
				}
				return nil
			}),
	)

	secondaryGroup := huh.NewGroup(
		huh.NewMultiSelect[*domain.AssetDefinition]().
			Title("Starting assets — any attribute").
			OptionsFunc(func() []huh.Option[*domain.AssetDefinition] {
				return anyAssetOptions(m, allDefs, spatialMap)
			}, &m.homeworldID).
			Value(&m.secondaryAssets).
			Validate(func(picked []*domain.AssetDefinition) error {
				_, otherCount := domain.AssetCountsFromScale(m.scale)
				if len(picked) > otherCount {
					return fmt.Errorf("choose at most %d any-attribute assets for %s scale", otherCount, m.scale)
				}
				return nil
			}),
	)

	return primaryGroup, secondaryGroup
}

func primaryAssetOptions(m *Model, allDefs []*domain.AssetDefinition, spatialMap *spatial.RegionMap) []huh.Option[*domain.AssetDefinition] {
	primaryRating, _, _ := domain.RatingsFromScale(m.scale)
	techLevel := worldTechLevel(m.homeworldID, spatialMap)

	var options []huh.Option[*domain.AssetDefinition]
	for _, def := range allDefs {
		if def.Category != m.primaryStat {
			continue
		}
		if def.MinRating > primaryRating {
			continue
		}
		if def.TechLevel > techLevel {
			continue
		}
		options = append(options, huh.NewOption(
			fmt.Sprintf("%s (cost %d)", def.Name, def.Cost),
			def,
		))
	}
	return options
}

func anyAssetOptions(m *Model, allDefs []*domain.AssetDefinition, spatialMap *spatial.RegionMap) []huh.Option[*domain.AssetDefinition] {
	primaryRating, secondaryRating, tertiaryRating := domain.RatingsFromScale(m.scale)
	techLevel := worldTechLevel(m.homeworldID, spatialMap)

	ratingFor := map[domain.FactionStat]int{
		m.primaryStat:   primaryRating,
		m.secondaryStat: secondaryRating,
		m.tertiaryStat:  tertiaryRating,
	}

	var options []huh.Option[*domain.AssetDefinition]
	for _, def := range allDefs {
		rating := ratingFor[def.Category]
		if def.MinRating > rating {
			continue
		}
		if def.TechLevel > techLevel {
			continue
		}
		options = append(options, huh.NewOption(
			fmt.Sprintf("%s (%s, cost %d)", def.Name, def.Category, def.Cost),
			def,
		))
	}
	return options
}

func worldTechLevel(worldID string, spatialMap *spatial.RegionMap) int {
	if loc, ok := spatialMap.Location(worldID); ok {
		return loc.TechLevel()
	}
	return 5 // permissive default when world isn't resolved yet
}

func previewMaxHP(m *Model) int {
	primary, secondary, tertiary := domain.RatingsFromScale(m.scale)
	temp := &domain.Faction{}
	for _, stat := range []struct {
		name   domain.FactionStat
		rating int
	}{
		{m.primaryStat, primary},
		{m.secondaryStat, secondary},
		{m.tertiaryStat, tertiary},
	} {
		switch stat.name {
		case domain.StatForce:
			temp.Force = stat.rating
		case domain.StatCunning:
			temp.Cunning = stat.rating
		case domain.StatWealth:
			temp.Wealth = stat.rating
		}
	}
	return domain.CalcMaxHP(temp)
}

func instantiateAssets(defs []*domain.AssetDefinition, homeworld domain.Location, tempFaction *domain.Faction) []*domain.Asset {
	assets := make([]*domain.Asset, 0, len(defs))
	for _, def := range defs {
		id := domain.NextAssetID(tempFaction, def.ID)
		asset := &domain.Asset{
			ID:           id,
			DefinitionID: def.ID,
			OwnerID:      tempFaction.ID,
			Location:     homeworld,
			CurrentHP:    def.HP,
			Ready:        true,
			Maintained:   true,
		}
		tempFaction.Assets[id] = asset
		assets = append(assets, asset)
	}
	return assets
}

func generateID(name string, existing map[string]*domain.Faction) string {
	base := slugify(name)
	if base == "" {
		base = "faction"
	}
	id := base
	for i := 2; ; i++ {
		if _, exists := existing[id]; !exists {
			return id
		}
		id = fmt.Sprintf("%s-%d", base, i)
	}
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// ---------------------------------------------------------------------------
// Help
// ---------------------------------------------------------------------------

type wizardHelpKeys struct{}

func (wizardHelpKeys) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "advance")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "discard")),
	}
}

func (wizardHelpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{wizardHelpKeys{}.ShortHelp()}
}
