package overlay

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type orderKind int

const (
	orderNone orderKind = iota
	orderIssue
	orderRevise
	orderCancel
)

// assetRow is one eligible asset's chosen order. Bound by pointer into the
// heap-allocated movementData.rows slice (allocated once at full length so the
// addresses stay stable for huh's bindings).
type assetRow struct {
	assetID string
	kind    orderKind
	worldID string
}

type movementData struct{ rows []assetRow }

// Movement is the sequential per-asset order form: two groups per asset (a kind
// select, then a destination select gated by group HideFunc on the row's kind).
// huh v1.0.0 supports HideFunc only at group granularity (OQ 8), so the
// destination is its own gated group, not a hidden field.
type Movement struct {
	data       *movementData
	form       *huh.Form
	spatialMap *spatial.RegionMap
}

func NewMovement(eligible []*domain.Asset, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) Movement {
	data := &movementData{rows: make([]assetRow, len(eligible))}
	worldOpts := worldOptions(spatialMap)
	names := worldNames(spatialMap)

	groups := make([]*huh.Group, 0, len(eligible)*2)
	for i, asset := range eligible {
		idx := i
		data.rows[idx].assetID = asset.ID

		loc := names[asset.Location.WorldID]
		if loc == "" {
			loc = asset.Location.WorldID
		}

		kindOpts := []huh.Option[orderKind]{
			huh.NewOption("Hold position", orderNone),
			huh.NewOption("Issue move", orderIssue),
		}
		if asset.CurrentOrder != nil { // Revise/Cancel only for an asset with a live order
			kindOpts = append(kindOpts,
				huh.NewOption("Revise move", orderRevise),
				huh.NewOption("Cancel move", orderCancel),
			)
		}

		kindGroup := huh.NewGroup(
			huh.NewSelect[orderKind]().
				Title(fmt.Sprintf("Asset %d of %d: %s (at %s)", idx+1, len(eligible), assetName(asset, rb), loc)).
				Options(kindOpts...).
				Value(&data.rows[idx].kind),
		)

		destGroup := huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("%s → destination", assetName(asset, rb))).
				Options(worldOpts...).
				Value(&data.rows[idx].worldID),
		).WithHideFunc(func() bool {
			k := data.rows[idx].kind
			return k != orderIssue && k != orderRevise
		})

		groups = append(groups, kindGroup, destGroup)
	}

	form := huh.NewForm(groups...).WithTheme(styles.FormTheme())
	return Movement{data: data, form: form, spatialMap: spatialMap}
}

func (o Movement) Init() tea.Cmd { return o.form.Init() }

func (o Movement) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	// Esc declines with the legal empty answer (no orders issued). Cancel-as-error
	// stays scoped to action overlays (Decision 2); movement has no error to raise.
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: []world.MovementDecision{}} }
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		decisions := o.decisions()
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: decisions} }
	}
	return o, cmd
}

func (o Movement) View() string { return o.form.View() }

func (o Movement) Help() help.KeyMap { return declineFormHelp{} }

// decisions assembles the engine slice. Issue/Revise set BOTH WorldID and
// RegionHex on Destination: Issue derives the hex from WorldID (movement.go:91)
// and Revise trusts RegionHex (movement.go:124), so both are populated (OQ 9).
func (o Movement) decisions() []world.MovementDecision {
	var out []world.MovementDecision
	for _, r := range o.data.rows {
		switch r.kind {
		case orderNone:
			continue
		case orderCancel:
			out = append(out, world.MovementDecision{AssetID: r.assetID, Kind: world.MovementDecisionCancel})
		case orderIssue, orderRevise:
			dest := &domain.Location{WorldID: r.worldID}
			if loc, ok := o.spatialMap.Location(r.worldID); ok {
				if rl, ok := loc.(spatial.RegionLocation); ok {
					dest.RegionHex = rl.RegionHex()
				}
			}
			kind := world.MovementDecisionIssue
			if r.kind == orderRevise {
				kind = world.MovementDecisionRevise
			}
			out = append(out, world.MovementDecision{AssetID: r.assetID, Kind: kind, Destination: dest})
		}
	}
	return out
}

// --- shared overlay helpers (also used by cargo.go) ---

func assetName(asset *domain.Asset, rb *rulebook.Rulebook) string {
	if def, ok := rb.Assets[asset.DefinitionID]; ok {
		return def.Name
	}
	return asset.DefinitionID
}

func worldOptions(spatialMap *spatial.RegionMap) []huh.Option[string] {
	worlds := spatialMap.AllWorlds()
	sort.Slice(worlds, func(i, j int) bool { return worlds[i].Name() < worlds[j].Name() })
	opts := make([]huh.Option[string], len(worlds))
	for i, w := range worlds {
		opts[i] = huh.NewOption(fmt.Sprintf("%s (TL %d)", w.Name(), w.TechLevel()), w.ID())
	}
	return opts
}

func worldNames(spatialMap *spatial.RegionMap) map[string]string {
	out := make(map[string]string)
	for _, w := range spatialMap.AllWorlds() {
		out[w.ID()] = w.Name()
	}
	return out
}

var _ Overlay = Movement{}
