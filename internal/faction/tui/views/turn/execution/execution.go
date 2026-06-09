package execution

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/chrome"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/layout"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/msgs"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/overlay"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type keyMap struct {
	Scroll key.Binding
	Follow key.Binding
}

// movableLine is a resolved, race-safe summary of one eligible asset for the
// movement detail pane. Built from the SelectMovementDecisions ask payload while
// the engine is parked on the reply; never read from the live faction map.
type movableLine struct{ name, location, order string }

// factionSnapshot is a copy of the acting faction's display fields, refreshed
// from each *domain.Faction event. The view never holds the engine pointer.
type factionSnapshot struct {
	name, scale            string
	force, cunning, wealth int
	curHP, maxHP, coin     int
}

// wizardBandHeight reserves the top of the center column for the active overlay.
// huh caps a Select at 10 visible options, so the tallest form (title + options +
// footer) is ~13 rows; 14 leaves one breathing row. The event stream takes the
// remaining center height below the band.
const wizardBandHeight = 14

// bannerHeight is the phase banner at the very top of the center column: one row
// of text plus a blank pad row that separates it from the wizard band below.
const bannerHeight = 2

// streamHeight is the event stream's row budget: the center content height minus
// the banner, the fixed wizard band, and the one-row divider between band and
// stream. Clamped to 1.
func streamHeight(totalHeight int) int {
	return max(totalHeight-bannerHeight-wizardBandHeight-1, 1)
}

// Model renders a running cycle as a left rail of the turn order, a center column
// split into a fixed wizard band (the active overlay) over a persistent event
// stream, and a right detail card for the acting faction.
type Model struct {
	order     []adapter.RailEntry
	currentID string

	stream viewport.Model
	lines  []string
	follow bool

	detail      factionSnapshot
	cycleNumber int
	movables    []movableLine

	overlay      overlay.Overlay
	activeAsk    adapter.AskKind // which modal is mounted; read only while overlay != nil
	pendingReply chan<- any

	latestErr error
	fatal     bool

	rulebook   *rulebook.Rulebook
	spatialMap *spatial.RegionMap

	keys          keyMap
	width, height int
}

func New(width, height int, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) Model {
	centerW := layout.RegionWidths(width)[layout.Center]
	return Model{
		rulebook:   rb,
		spatialMap: spatialMap,
		stream:     viewport.New(centerW, streamHeight(height)),
		follow:     true,
		keys: keyMap{
			Scroll: key.NewBinding(key.WithKeys("up", "down", "j", "k"), key.WithHelp("↑/↓", "scroll")),
			Follow: key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G", "follow")),
		},
		width:  width,
		height: height,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.stream.Width = layout.RegionWidths(msg.Width)[layout.Center]
		m.stream.Height = streamHeight(msg.Height)
		m.stream.SetContent(strings.Join(m.lines, "\n"))
		if m.follow {
			m.stream.GotoBottom()
		}
		return m, nil

	case adapter.ObserverEventMsg:
		m.applyEvent(msg)
		m.lines = append(m.lines, renderEvent(msg))
		m.stream.SetContent(strings.Join(m.lines, "\n"))
		if m.follow {
			m.stream.GotoBottom()
		}
		return m, nil

	case adapter.EngineDoneMsg:
		if msg.Err != nil {
			m.latestErr = msg.Err
			m.fatal = true
		}
		return m, nil

	case adapter.CollectorAskMsg:
		if msg.Kind == adapter.AskSelectMovementDecisions {
			m.movables = resolveMovables(msg.Payload.(adapter.SelectMovementDecisionsPayload).Eligible, m.rulebook, m.spatialMap)
		}
		m.pendingReply = msg.Reply
		m.activeAsk = msg.Kind
		m.overlay = m.newOverlay(msg)
		return m, m.overlay.Init()

	case overlay.OverlayDoneMsg:
		if m.pendingReply != nil {
			m.pendingReply <- msg.Answer // cap-1 buffered (phase_collector.go:18), non-blocking
			m.pendingReply = nil
		}
		m.overlay = nil
		return m, nil

	case tea.KeyMsg:
		if m.fatal {
			return m, func() tea.Msg { return msgs.ReturnToSetupMsg{} }
		}
		if m.overlay != nil {
			var cmd tea.Cmd
			m.overlay, cmd = m.overlay.Update(msg)
			return m, cmd
		}
		if key.Matches(msg, m.keys.Follow) {
			m.stream.GotoBottom()
			m.follow = true
			return m, nil
		}
		var cmd tea.Cmd
		m.stream, cmd = m.stream.Update(msg)
		m.follow = m.stream.AtBottom()
		return m, cmd
	}
	// Forward unmatched messages to the overlay so that huh's internal messages
	// (focus, nextField, etc.) reach the form after Init dispatches them.
	if m.overlay != nil {
		var cmd tea.Cmd
		m.overlay, cmd = m.overlay.Update(msg)
		return m, cmd
	}
	return m, nil
}

// applyEvent updates the rail/detail state captured from each event. Only the
// four *domain.Faction events carry a faction; mutation events touch neither
// the rail nor the card.
func (m *Model) applyEvent(msg adapter.ObserverEventMsg) {
	switch msg.Kind {
	case adapter.EvtCycleStarted:
		payload := msg.Payload.(adapter.CycleStartedPayload)
		m.order = payload.Order
		m.cycleNumber = payload.CycleNumber
	case adapter.EvtFactionTurnStarted:
		faction := msg.Payload.(*domain.Faction)
		m.currentID = faction.ID
		m.detail = snapshotOf(faction)
		m.movables = nil
		if !m.fatal {
			m.latestErr = nil // a recoverable error scopes to the faction it occurred under
		}
	case adapter.EvtStatRaiseSkipped, adapter.EvtFactionSkipped, adapter.EvtFactionTurnCompleted:
		m.detail = snapshotOf(msg.Payload.(*domain.Faction))
	case adapter.EvtError:
		m.latestErr = msg.Payload.(adapter.ErrorPayload).Err
	}
}

func snapshotOf(faction *domain.Faction) factionSnapshot {
	return factionSnapshot{
		name:    faction.Name,
		scale:   string(faction.Scale),
		force:   faction.Force,
		cunning: faction.Cunning,
		wealth:  faction.Wealth,
		curHP:   faction.CurrentHP,
		maxHP:   faction.MaxHP,
		coin:    faction.Coin,
	}
}

func (m Model) View() string {
	contentH := max(m.height, 1)
	widths := layout.RegionWidths(m.width)
	centerW := widths[layout.Center]

	bandH := min(wizardBandHeight, contentH)
	divider := styles.Rule.Render(strings.Repeat("─", centerW))
	centerContent := lipgloss.JoinVertical(lipgloss.Left,
		m.bannerView(centerW),
		m.wizardBand(centerW, bandH),
		divider,
		m.stream.View(),
	)

	panels := map[layout.Region]layout.Panel{
		layout.Left:   {Content: m.railView(), Style: lipgloss.NewStyle()},
		layout.Center: {Content: centerContent, Style: lipgloss.NewStyle()},
		layout.Right:  {Content: m.detailView(), Style: lipgloss.NewStyle()},
	}
	return layout.Compose(panels, widths, contentH)
}

// bannerView renders the one-row phase header. The phase is anchored to the
// active modal (a genuine pause point); between modals it reads "resolving…" so
// the burst of non-interactive events never flickers the banner text.
func (m Model) bannerView(width int) string {
	style := lipgloss.NewStyle().Width(width).Height(bannerHeight)
	if m.detail.name == "" {
		return style.Render("")
	}
	phase := "resolving…"
	if m.overlay != nil {
		phase = askPhaseLabel(m.activeAsk)
	}
	return style.Render(styles.AccentAlt.Render("▶ "+m.detail.name) + styles.Subtle.Render(" · "+phase))
}

func askPhaseLabel(kind adapter.AskKind) string {
	switch kind {
	case adapter.AskSelectStatRaise:
		return "Stat Raise"
	case adapter.AskSelectAction:
		return "Action"
	case adapter.AskSelectMovementDecisions:
		return "Movement"
	case adapter.AskSelectTransportCargo:
		return "Load Cargo"
	case adapter.AskAwaitCheckpoint:
		return "Cycle Checkpoint"
	case adapter.AskSelectAsset:
		return "Select Asset"
	case adapter.AskSelectModifiers:
		return "Pre-roll Modifiers"
	case adapter.AskConfirmReroll:
		return "Reroll?"
	case adapter.AskSelectBuyOrder:
		return "Buy Asset"
	case adapter.AskSelectRefitOrder:
		return "Refit Asset"
	case adapter.AskSelectRepairOrders:
		return "Repair Asset"
	case adapter.AskSelectBribeTarget:
		return "Bribe"
	case adapter.AskSelectSeizeTarget:
		return "Seize Planet"
	case adapter.AskSelectChangeHomeworldTarget:
		return "Change Homeworld"
	case adapter.AskSelectAttackers:
		return "Select Attackers"
	case adapter.AskSelectDefender:
		return "Select Defender"
	case adapter.AskConfirmRedirectToBase:
		return "Redirect to Base?"
	case adapter.AskSelectExpandInfluenceOrder:
		return "Expand Influence"
	case adapter.AskConfirmRivalFreeAttack:
		return "Rival Free Attack?"
	case adapter.AskSelectBaseAttackers:
		return "Base Attackers"
	case adapter.AskSelectAbilityAssets:
		return "Ability Assets"
	case adapter.AskConfirmAbilityApplied:
		return "Apply Ability?"
	case adapter.AskSelectFactionTestTarget:
		return "Ability Target"
	default:
		return "Action"
	}
}

// wizardBand renders the reserved top of the center column: the active overlay,
// or a dim placeholder during non-interactive phases.
func (m Model) wizardBand(width, height int) string {
	if m.overlay == nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, styles.Dim.Render("—"))
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Top, m.overlay.View())
}

func (m Model) railView() string {
	var b strings.Builder
	for _, entry := range m.order {
		if entry.ID == m.currentID {
			b.WriteString(styles.AccentAlt.Render("▶ " + entry.Name))
		} else {
			b.WriteString(styles.Subtle.Render("  " + entry.Name))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) detailView() string {
	if m.detail.name == "" {
		return styles.Dim.Render("—")
	}
	// Specialized cards support an active decision, so they key off the mounted
	// modal — the only ground-truth phase signal. (Deriving them from turn events
	// instead is unreliable: the engine ticks in-flight orders and fires
	// MovementTicked *before* the movement ask, so an event-derived phase would
	// already have advanced past movement by the time its modal opens.)
	if m.overlay != nil {
		switch m.activeAsk {
		case adapter.AskSelectStatRaise:
			return m.statRaiseDetail()
		case adapter.AskSelectMovementDecisions, adapter.AskSelectTransportCargo:
			return m.movementDetail()
		case adapter.AskSelectAction:
			return m.actionDetail()
		}
	}
	return m.baseDetail()
}

func (m Model) baseDetail() string {
	d := m.detail
	var b strings.Builder
	b.WriteString(styles.Strong.Render(d.name) + "\n")
	b.WriteString(styles.Subtle.Render(d.scale) + "\n\n")
	b.WriteString(styles.Subtle.Render("Force ") + styles.Strong.Render(strconv.Itoa(d.force)) + "\n")
	b.WriteString(styles.Subtle.Render("Cunning ") + styles.Strong.Render(strconv.Itoa(d.cunning)) + "\n")
	b.WriteString(styles.Subtle.Render("Wealth ") + styles.Strong.Render(strconv.Itoa(d.wealth)) + "\n\n")
	b.WriteString(styles.Subtle.Render("HP ") + styles.Strong.Render(fmt.Sprintf("%d/%d", d.curHP, d.maxHP)) + "\n")
	b.WriteString(styles.Subtle.Render("Coin ") + styles.Strong.Render(strconv.Itoa(d.coin)))
	return b.String()
}

// statRaiseDetail emphasizes the three ratings (the raise targets) over the base
// card. Renders from the race-safe snapshot scalars.
func (m Model) statRaiseDetail() string {
	d := m.detail
	var b strings.Builder
	b.WriteString(styles.Strong.Render(d.name) + "\n")
	b.WriteString(styles.Subtle.Render(d.scale) + "\n\n")
	b.WriteString(styles.Subtle.Render("Raise a stat") + "\n")
	b.WriteString(styles.Subtle.Render("Force ") + styles.AccentAlt.Render(strconv.Itoa(d.force)) + "\n")
	b.WriteString(styles.Subtle.Render("Cunning ") + styles.AccentAlt.Render(strconv.Itoa(d.cunning)) + "\n")
	b.WriteString(styles.Subtle.Render("Wealth ") + styles.AccentAlt.Render(strconv.Itoa(d.wealth)))
	return b.String()
}

// movementDetail lists the faction's movable assets with location and current
// order, from the stashed ask summary. Falls back to the base card if empty.
func (m Model) movementDetail() string {
	if len(m.movables) == 0 {
		return m.baseDetail()
	}
	var b strings.Builder
	b.WriteString(styles.Strong.Render(m.detail.name) + "\n")
	b.WriteString(styles.Subtle.Render("Movable assets") + "\n\n")
	for _, line := range m.movables {
		b.WriteString(styles.Strong.Render(line.name) + "\n")
		b.WriteString(styles.Subtle.Render("  at "+line.location) + "\n")
		b.WriteString(styles.Subtle.Render("  "+line.order) + "\n")
	}
	return b.String()
}

// actionDetail augments the base card with asset/base counts during the action
// pick — a light "what can this faction bring to bear" cue. The per-prompt
// overlays carry the fine-grained decision context (attacker stats, costs).
func (m Model) actionDetail() string {
	return m.baseDetail() // Effort 3: counts deferred — base card is sufficient for the pick
}

// resolveMovables builds the race-safe movement summary from the ask payload.
// Safe to read the live asset pointers here: the engine is parked on the ask's
// reply when this runs (phase_collector.go ask round-trip).
func resolveMovables(eligible []*domain.Asset, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) []movableLine {
	names := make(map[string]string)
	for _, w := range spatialMap.AllWorlds() {
		names[w.ID()] = w.Name()
	}
	lines := make([]movableLine, 0, len(eligible))
	for _, asset := range eligible {
		name := asset.DefinitionID
		if def, ok := rb.Assets[asset.DefinitionID]; ok {
			name = def.Name
		}
		location := names[asset.Location.WorldID]
		if location == "" {
			location = asset.Location.WorldID
		}
		order := "no order"
		if asset.CurrentOrder != nil {
			dest := asset.CurrentOrder.Destination.WorldID
			if n, ok := names[dest]; ok {
				dest = n
			}
			order = "→ " + dest
		}
		lines = append(lines, movableLine{name: name, location: location, order: order})
	}
	return lines
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }

func (h helpKeys) ShortHelp() []key.Binding  { return []key.Binding{h.keys.Scroll, h.keys.Follow} }
func (h helpKeys) FullHelp() [][]key.Binding { return [][]key.Binding{{h.keys.Scroll, h.keys.Follow}} }

func (m Model) CapturesInput() bool { return m.overlay != nil }

func (m Model) StatusLine() (string, chrome.Severity) {
	if m.latestErr == nil {
		return "", chrome.Info
	}
	if m.fatal {
		return m.latestErr.Error(), chrome.Fatal
	}
	return m.latestErr.Error(), chrome.Recoverable
}

// newOverlay builds the overlay for an ask. One arm per AskKind so Efforts 2-3
// swap real overlays in kind-by-kind; the AskSelectAction arm stays on the
// placeholder until Effort 3. A method so it can supply the cycle number,
// rulebook, and spatial map the real overlays need but the ask does not carry.
func (m Model) newOverlay(msg adapter.CollectorAskMsg) overlay.Overlay {
	name := ""
	if msg.Faction != nil {
		name = msg.Faction.Name
	}
	switch msg.Kind {
	case adapter.AskAwaitCheckpoint:
		return overlay.NewCheckpoint(m.cycleNumber)
	case adapter.AskSelectStatRaise:
		return overlay.NewStatRaise(msg.Faction, msg.Payload.(adapter.SelectStatRaisePayload).Eligible)
	case adapter.AskSelectAction:
		return overlay.NewSelectAction(msg.Payload.(adapter.SelectActionPayload).Available)
	case adapter.AskSelectAsset:
		return overlay.NewSelectAsset("Select asset", msg.Payload.(adapter.SelectAssetPayload).Assets, m.rulebook)
	case adapter.AskSelectModifiers:
		return overlay.NewSelectModifiers(msg.Payload.(adapter.SelectModifiersPayload).Offers)
	case adapter.AskConfirmReroll:
		return overlay.NewConfirmReroll(msg.Payload.(adapter.ConfirmRerollPayload).Directive)
	case adapter.AskSelectBuyOrder:
		p := msg.Payload.(adapter.SelectBuyOrderPayload)
		return overlay.NewBuyOrder(p.PurchasableByWorld, p.WorldNames, p.Coin)
	case adapter.AskSelectRefitOrder:
		return overlay.NewRefitOrder(msg.Payload.(adapter.SelectRefitOrderPayload).Options, m.rulebook)
	case adapter.AskSelectRepairOrders:
		return overlay.NewRepairOrders(msg.Payload.(adapter.SelectRepairOrdersPayload).Targets, m.rulebook)
	case adapter.AskSelectBribeTarget:
		p := msg.Payload.(adapter.SelectBribeTargetPayload)
		return overlay.NewBribe(p.Bases, p.OwnerNames, p.Coin)
	case adapter.AskSelectSeizeTarget:
		p := msg.Payload.(adapter.SelectSeizeTargetPayload)
		return overlay.NewWorldSelect("Seize which world?", "", p.Worlds, p.WorldNames)
	case adapter.AskSelectChangeHomeworldTarget:
		p := msg.Payload.(adapter.SelectChangeHomeworldTargetPayload)
		return overlay.NewWorldSelect("New homeworld?", "", p.Worlds, p.WorldNames)
	case adapter.AskSelectAttackers:
		return overlay.NewSelectAttackers(msg.Payload.(adapter.SelectAttackersPayload).Eligible, m.rulebook)
	case adapter.AskSelectDefender:
		p := msg.Payload.(adapter.SelectDefenderPayload)
		return overlay.NewSelectDefender(p.Attacker, p.Eligible, p.OwnerNames, m.rulebook)
	case adapter.AskConfirmRedirectToBase:
		p := msg.Payload.(adapter.ConfirmRedirectToBasePayload)
		return overlay.NewConfirmRedirectToBase(p.DefenderFaction, p.Base, p.Damage)
	case adapter.AskSelectExpandInfluenceOrder:
		p := msg.Payload.(adapter.SelectExpandInfluenceOrderPayload)
		return overlay.NewExpand(p.NewBaseWorlds, p.WorldNames, p.ReinforceBases, p.Coin)
	case adapter.AskConfirmRivalFreeAttack:
		p := msg.Payload.(adapter.ConfirmRivalFreeAttackPayload)
		return overlay.NewConfirmRivalFreeAttack(p.Rival, p.RivalRoll, p.FactionRoll)
	case adapter.AskSelectBaseAttackers:
		p := msg.Payload.(adapter.SelectBaseAttackersPayload)
		return overlay.NewSelectBaseAttackers(p.Rival, p.Eligible, m.rulebook)
	case adapter.AskSelectAbilityAssets:
		return overlay.NewSelectAbilityAssets(msg.Payload.(adapter.SelectAbilityAssetsPayload).Candidates, m.rulebook)
	case adapter.AskConfirmAbilityApplied:
		p := msg.Payload.(adapter.ConfirmAbilityAppliedPayload)
		return overlay.NewConfirmAbilityApplied(p.Asset, p.Def)
	case adapter.AskSelectFactionTestTarget:
		p := msg.Payload.(adapter.SelectFactionTestTargetPayload)
		return overlay.NewSelectFactionTestTarget(p.Asset, p.Effect, p.Candidates)
	case adapter.AskSelectMovementDecisions:
		return overlay.NewMovement(msg.Payload.(adapter.SelectMovementDecisionsPayload).Eligible, m.rulebook, m.spatialMap)
	case adapter.AskSelectTransportCargo:
		p := msg.Payload.(adapter.SelectTransportCargoPayload)
		return overlay.NewTransportCargo(p.Transport, p.EligibleCargo, p.Profile, m.rulebook)
	default:
		return overlay.NewPlaceholder(msg.Kind, name)
	}
}
