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

// formData holds all values that huh binds to via pointer. It must be
// heap-allocated and shared between the Model and the form so that copying
// the Model (value semantics) does not orphan the bindings.
type formData struct {
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
}

type Model struct {
	form         *huh.Form // phase 1: name through coin (no assets)
	assetsForm   *huh.Form // phase 2: asset selection, built after phase 1 completes
	phase        int       // 0 = main form, 1 = asset form
	rulebook     *rulebook.Rulebook
	spatialMap   *spatial.RegionMap
	factionState *state.FactionState
	data         *formData

	confirmingDiscard bool
	discardKeys       discardKeyMap
}

type discardKeyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
}

func New(rb *rulebook.Rulebook, spatialMap *spatial.RegionMap, factionState *state.FactionState) Model {
	data := &formData{
		scale:         domain.ScaleMinor,
		primaryStat:   domain.StatForce,
		secondaryStat: domain.StatCunning,
		tertiaryStat:  domain.StatWealth,
		coinStr:       "0",
	}
	m := Model{
		rulebook:     rb,
		spatialMap:   spatialMap,
		factionState: factionState,
		data:         data,
		discardKeys: discardKeyMap{
			Confirm: key.NewBinding(key.WithKeys("y")),
			Cancel:  key.NewBinding(key.WithKeys("n", "esc")),
		},
	}
	m.form = buildForm(data, rb, spatialMap)
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

	if m.phase == 0 {
		formModel, cmd := m.form.Update(msg)
		m.form = formModel.(*huh.Form)
		if m.form.State == huh.StateCompleted {
			m.phase = 1
			m.assetsForm = buildAssetsForm(m.data, m.rulebook, m.spatialMap)
			return m, m.assetsForm.Init()
		}
		return m, cmd
	}

	formModel, cmd := m.assetsForm.Update(msg)
	m.assetsForm = formModel.(*huh.Form)
	if m.assetsForm.State == huh.StateCompleted {
		faction := m.synthesize()
		return m, func() tea.Msg { return msgs.CreatedMsg{Faction: faction} }
	}
	return m, cmd
}

func (m Model) View() string {
	if m.confirmingDiscard {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F9E2AF")).Padding(1, 2).
			Render("Discard new faction? [y/N]")
	}
	if m.phase == 1 {
		return m.assetsForm.View()
	}
	return m.form.View()
}

func (m Model) Help() help.KeyMap { return wizardHelpKeys{} }

func (m *Model) synthesize() *domain.Faction {
	coin, _ := strconv.Atoi(m.data.coinStr)

	var homeworld domain.Location
	homeworld.WorldID = m.data.homeworldID
	if loc, ok := m.spatialMap.Location(m.data.homeworldID); ok {
		if rl, ok := loc.(spatial.RegionLocation); ok {
			homeworld.RegionHex = rl.RegionHex()
		}
	}

	id := generateID(m.data.name, m.factionState.Factions)

	tempFaction := &domain.Faction{
		ID:     id,
		Assets: make(map[string]*domain.Asset),
	}
	allPicked := append(m.data.primaryAssets, m.data.secondaryAssets...)
	assets := instantiateAssets(allPicked, homeworld, tempFaction)

	return domain.NewFaction(
		id, m.data.name,
		m.data.scale,
		m.data.primaryStat, m.data.secondaryStat, m.data.tertiaryStat,
		m.data.selectedTags,
		m.data.selectedGoal,
		homeworld,
		assets,
		coin,
	)
}

// ---------------------------------------------------------------------------
// Form builder
// ---------------------------------------------------------------------------

func buildForm(data *formData, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) *huh.Form {
	statOptions := []huh.Option[domain.FactionStat]{
		huh.NewOption("Force", domain.StatForce),
		huh.NewOption("Cunning", domain.StatCunning),
		huh.NewOption("Wealth", domain.StatWealth),
	}

	nameGroup := huh.NewGroup(
		huh.NewInput().
			Title("Faction name").
			Value(&data.name).
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
			Value(&data.scale),
	)

	statsGroup := huh.NewGroup(
		huh.NewSelect[domain.FactionStat]().
			Title("Primary attribute").
			Options(statOptions...).
			Value(&data.primaryStat),
		huh.NewSelect[domain.FactionStat]().
			Title("Secondary attribute").
			Options(statOptions...).
			Value(&data.secondaryStat),
		huh.NewSelect[domain.FactionStat]().
			Title("Tertiary attribute").
			Options(statOptions...).
			Value(&data.tertiaryStat).
			Validate(func(domain.FactionStat) error {
				if data.primaryStat == data.secondaryStat || data.primaryStat == data.tertiaryStat || data.secondaryStat == data.tertiaryStat {
					return fmt.Errorf("each attribute must be distinct")
				}
				return nil
			}),
	)

	hpGroup := huh.NewGroup(
		huh.NewNote().
			Title("Max HP").
			DescriptionFunc(func() string {
				return fmt.Sprintf("%d HP (derived from attribute ratings)", previewMaxHP(data))
			}, &data.scale),
	)

	tagsGroup := huh.NewGroup(
		huh.NewMultiSelect[*domain.Tag]().
			Title("Tags (choose up to 2)").
			Options(tagsToOptions(rb.Tags)...).
			Value(&data.selectedTags).
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
			Value(&data.selectedGoal),
	)

	homeworldGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Homeworld").
			Options(worldsToOptions(spatialMap)...).
			Value(&data.homeworldID),
	)

	coinGroup := huh.NewGroup(
		huh.NewInput().
			Title("Starting Coin").
			Value(&data.coinStr).
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
		coinGroup,
	).WithTheme(wizardTheme())
}

// wizardTheme renders multiselect rows as checkboxes ([✓] / [ ]) so selection
// state is unambiguous.
func wizardTheme() *huh.Theme {
	theme := huh.ThemeCharm()
	theme.Focused.SelectedPrefix = theme.Focused.SelectedPrefix.SetString("[✓] ")
	theme.Focused.UnselectedPrefix = theme.Focused.UnselectedPrefix.SetString("[ ] ")
	theme.Blurred.SelectedPrefix = theme.Blurred.SelectedPrefix.SetString("[✓] ")
	theme.Blurred.UnselectedPrefix = theme.Blurred.UnselectedPrefix.SetString("[ ] ")
	return theme
}

// buildAssetsForm constructs the asset-selection form after the main form
// completes, so options are computed from the user's final scale/stats/homeworld.
func buildAssetsForm(data *formData, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) *huh.Form {
	allDefs := make([]*domain.AssetDefinition, 0, len(rb.Assets))
	for _, def := range rb.Assets {
		allDefs = append(allDefs, def)
	}
	sort.Slice(allDefs, func(i, j int) bool { return allDefs[i].Name < allDefs[j].Name })

	primaryOpts := primaryAssetOptions(data, allDefs, spatialMap)
	secondaryOpts := anyAssetOptions(data, allDefs, spatialMap)
	primaryCount, otherCount := domain.AssetCountsFromScale(data.scale)

	primaryGroup := huh.NewGroup(
		huh.NewMultiSelect[*domain.AssetDefinition]().
			Title(fmt.Sprintf("Starting assets — primary attribute (choose up to %d)", primaryCount)).
			Options(primaryOpts...).
			Value(&data.primaryAssets).
			Validate(func(picked []*domain.AssetDefinition) error {
				if len(picked) > primaryCount {
					return fmt.Errorf("choose at most %d primary-attribute assets for %s scale", primaryCount, data.scale)
				}
				return nil
			}),
	)

	secondaryGroup := huh.NewGroup(
		huh.NewMultiSelect[*domain.AssetDefinition]().
			Title(fmt.Sprintf("Starting assets — any attribute (choose up to %d)", otherCount)).
			Options(secondaryOpts...).
			Value(&data.secondaryAssets).
			Validate(func(picked []*domain.AssetDefinition) error {
				if len(picked) > otherCount {
					return fmt.Errorf("choose at most %d any-attribute assets for %s scale", otherCount, data.scale)
				}
				return nil
			}),
	)

	return huh.NewForm(primaryGroup, secondaryGroup).WithTheme(wizardTheme())
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
func primaryAssetOptions(data *formData, allDefs []*domain.AssetDefinition, spatialMap *spatial.RegionMap) []huh.Option[*domain.AssetDefinition] {
	primaryRating, _, _ := domain.RatingsFromScale(data.scale)
	techLevel := worldTechLevel(data.homeworldID, spatialMap)

	var options []huh.Option[*domain.AssetDefinition]
	for _, def := range allDefs {
		if def.Category != data.primaryStat {
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

func anyAssetOptions(data *formData, allDefs []*domain.AssetDefinition, spatialMap *spatial.RegionMap) []huh.Option[*domain.AssetDefinition] {
	primaryRating, secondaryRating, tertiaryRating := domain.RatingsFromScale(data.scale)
	techLevel := worldTechLevel(data.homeworldID, spatialMap)

	ratingFor := map[domain.FactionStat]int{
		data.primaryStat:   primaryRating,
		data.secondaryStat: secondaryRating,
		data.tertiaryStat:  tertiaryRating,
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

func previewMaxHP(data *formData) int {
	primary, secondary, tertiary := domain.RatingsFromScale(data.scale)
	temp := &domain.Faction{}
	for _, stat := range []struct {
		name   domain.FactionStat
		rating int
	}{
		{data.primaryStat, primary},
		{data.secondaryStat, secondary},
		{data.tertiaryStat, tertiary},
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
		key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "advance")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "discard")),
	}
}

func (wizardHelpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{wizardHelpKeys{}.ShortHelp()}
}
