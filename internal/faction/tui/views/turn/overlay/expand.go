package overlay

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// expandData is heap-allocated so huh's pointer bindings survive the overlay's
// value-copy. The new-base fields (world, newHP) and reinforce fields (base,
// subMode, reinfHP) are gated by mode via group HideFuncs.
type expandData struct {
	mode    action.ExpandMode
	world   string
	newHP   string
	base    *domain.Base
	subMode action.ReinforceMode
	reinfHP string
}

// Expand is the Expand Influence wizard: a mode Select, then one of two gated
// branches. New: world Select + HP Input (HP == Coin cost). Reinforce: base
// Select → submode Select (options derive from the base's caps via OptionsFunc)
// → HP Input (cap derives from base+submode via the Validate closure). Group
// HideFunc is the only branching primitive huh exposes (movement.go). Answer:
// action.ExpandInfluenceOrder. Esc cancels.
type Expand struct {
	data *expandData
	form *huh.Form
}

func NewExpand(newBaseWorlds []string, worldNames map[string]string, reinforce []adapter.ReinforceTarget, coin int) Expand {
	data := &expandData{}

	capsByBase := make(map[string]adapter.ReinforceTarget, len(reinforce))
	for _, target := range reinforce {
		capsByBase[target.Base.ID] = target
	}

	modeOpts := make([]huh.Option[action.ExpandMode], 0, 2)
	if len(newBaseWorlds) > 0 {
		modeOpts = append(modeOpts, huh.NewOption("Place a new base", action.ExpandModeNew))
	}
	if len(reinforce) > 0 {
		modeOpts = append(modeOpts, huh.NewOption("Reinforce a base", action.ExpandModeReinforce))
	}
	modeGroup := huh.NewGroup(
		huh.NewSelect[action.ExpandMode]().
			Title("Expand influence — how?").
			Options(modeOpts...).
			Value(&data.mode),
	)

	// --- New-base branch (gated mode == New) ---
	worldOpts := make([]huh.Option[string], len(newBaseWorlds))
	for i, worldID := range newBaseWorlds {
		label := worldNames[worldID]
		if label == "" {
			label = worldID
		}
		worldOpts[i] = huh.NewOption(label, worldID)
	}
	newWorldGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Place new base on which world?").
			Options(worldOpts...).
			Value(&data.world),
	).WithHideFunc(func() bool { return data.mode != action.ExpandModeNew })

	newHPGroup := huh.NewGroup(
		huh.NewInput().
			Title(fmt.Sprintf("New base HP (= Coin cost, 1–%d)", coin)).
			Value(&data.newHP).
			Validate(amountValidator(1, coin)),
	).WithHideFunc(func() bool { return data.mode != action.ExpandModeNew })

	// --- Reinforce branch (gated mode == Reinforce) ---
	baseOpts := make([]huh.Option[*domain.Base], len(reinforce))
	for i, target := range reinforce {
		baseOpts[i] = huh.NewOption(reinforceBaseLabel(target), target.Base)
	}
	reinfBaseGroup := huh.NewGroup(
		huh.NewSelect[*domain.Base]().
			Title("Reinforce which base?").
			Options(baseOpts...).
			Value(&data.base),
	).WithHideFunc(func() bool { return data.mode != action.ExpandModeReinforce })

	reinfSubGroup := huh.NewGroup(
		huh.NewSelect[action.ReinforceMode]().
			Title("Heal damage or raise max HP?").
			Value(&data.subMode).
			OptionsFunc(func() []huh.Option[action.ReinforceMode] {
				var opts []huh.Option[action.ReinforceMode]
				if data.base == nil {
					return opts
				}
				target := capsByBase[data.base.ID]
				if target.CanHeal {
					opts = append(opts, huh.NewOption(fmt.Sprintf("Heal damage (up to %d HP)", target.HealCap), action.ReinforceHeal))
				}
				if target.CanMax {
					opts = append(opts, huh.NewOption(fmt.Sprintf("Raise max HP (up to %d)", target.MaxCap), action.ReinforceMax))
				}
				return opts
			}, &data.base),
	).WithHideFunc(func() bool { return data.mode != action.ExpandModeReinforce })

	reinfHPGroup := huh.NewGroup(
		huh.NewInput().
			Title("HP amount (Coin cost)").
			Value(&data.reinfHP).
			Validate(func(s string) error {
				return amountValidator(1, reinforceLimit(capsByBase, data.base, data.subMode, coin))(s)
			}),
	).WithHideFunc(func() bool { return data.mode != action.ExpandModeReinforce })

	form := huh.NewForm(modeGroup, newWorldGroup, newHPGroup, reinfBaseGroup, reinfSubGroup, reinfHPGroup).WithTheme(styles.FormTheme())
	return Expand{data: data, form: form}
}

func (o Expand) Init() tea.Cmd { return o.form.Init() }

func (o Expand) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		order := o.order()
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: order} }
	}
	return o, cmd
}

func (o Expand) View() string      { return o.form.View() }
func (o Expand) Help() help.KeyMap { return cancelFormHelp{} }

func (o Expand) order() action.ExpandInfluenceOrder {
	switch o.data.mode {
	case action.ExpandModeNew:
		hp, _ := strconv.Atoi(o.data.newHP)
		return action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: o.data.world, HPAmount: hp}
	case action.ExpandModeReinforce:
		hp, _ := strconv.Atoi(o.data.reinfHP)
		baseID := ""
		if o.data.base != nil {
			baseID = o.data.base.ID
		}
		return action.ExpandInfluenceOrder{Mode: action.ExpandModeReinforce, BaseID: baseID, SubMode: o.data.subMode, HPAmount: hp}
	}
	return action.ExpandInfluenceOrder{}
}

func reinforceBaseLabel(target adapter.ReinforceTarget) string {
	return fmt.Sprintf("Base @ %s · HP %d/%d", target.Base.Location.WorldID, target.Base.CurrentHP, target.Base.MaxHP)
}

// reinforceLimit is the live Coin/HP cap for the chosen base+submode, floored by
// Coin. Defensive coin fallback when nothing is selected yet (Validate fires
// before the binding resolves on the first paint).
func reinforceLimit(capsByBase map[string]adapter.ReinforceTarget, base *domain.Base, sub action.ReinforceMode, coin int) int {
	if base == nil {
		return coin
	}
	target := capsByBase[base.ID]
	limit := coin
	switch sub {
	case action.ReinforceHeal:
		if target.HealCap < limit {
			limit = target.HealCap
		}
	case action.ReinforceMax:
		if target.MaxCap < limit {
			limit = target.MaxCap
		}
	}
	return limit
}

// NewConfirmRivalFreeAttack asks (on the rival's behalf — single-GM tool) whether
// a tying/beating rival makes its free attack on the new base. Answer: bool.
func NewConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) ConfirmOverlay {
	title := fmt.Sprintf("Let %s make a free attack on the new base?", rival.Name)
	header := fmt.Sprintf("Contested roll — %s rolled %d vs your %d", rival.Name, rivalRoll, factionRoll)
	return NewConfirm(title, header, "Allow attack", "Decline")
}

// NewSelectBaseAttackers picks which of the rival's assets attack the new base.
// All eligible belong to the one rival, so labels need no owner. Answer:
// []*domain.Asset (uncapped).
func NewSelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rb *rulebook.Rulebook) MultiSelectOverlay[*domain.Asset] {
	opts := make([]huh.Option[*domain.Asset], 0, len(eligible))
	for _, asset := range eligible {
		opts = append(opts, huh.NewOption(attackerLabel(asset, rb), asset))
	}
	return NewMultiSelect[*domain.Asset]("Commit which attackers?", fmt.Sprintf("%s's assets attacking the new base", rival.Name), opts, 0)
}

var _ Overlay = Expand{}
