package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/paths"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/inputs"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/phases"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/actions"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type turnState int

const (
	stateResumePrompt   turnState = iota
	stateSkipPrompt
	stateBookkeeping
	stateActionSelect
	stateActionInput
	stateActionResult
	stateAttackRedirect // waiting for redirect confirm mid-resolution
	stateCycleSummary
	stateDone
)

// AttackRedirectMsg is sent by TUICollector.ConfirmRedirectToBase to ask the
// GM whether to redirect damage to the defender's Base of Influence.
type AttackRedirectMsg struct {
	DefFaction *domain.Faction
	Base       *domain.Base
	Damage     int
	ResponseCh chan bool
}

// AttackCompletedMsg is sent by the attack resolution goroutine when Resolve
// finishes (successfully or with an error).
type AttackCompletedMsg struct {
	Mutations []domain.Mutation
	Err       error
}

type factionSnapshot struct {
	hp   int
	coin int
}

type TurnModel struct {
	engine            *engine.Engine
	factionState      *state.FactionState
	paths             paths.Paths
	width             int
	height            int
	state             turnState
	currentFaction    *domain.Faction
	bookkeepingResult engine.BookkeepingResult
	pendingMutations  []domain.Mutation
	actionResultText  string
	availableActions  []engine.Action
	pendingAction     engine.Action
	subModel          tea.Model
	attackEventCh     chan tea.Msg
	pendingRedirect   *AttackRedirectMsg
	snapshots         map[string]factionSnapshot // faction ID → pre-turn HP/Coin
	actionsTaken      map[string]string          // faction ID → action description
	err               error
}

// resizeSub forwards the current terminal dimensions to a newly created sub-model.
// Sub-models created mid-loop never receive bubbletea's initial WindowSizeMsg.
func (m TurnModel) resizeSub(sub tea.Model) tea.Model {
	if m.width == 0 {
		return sub
	}
	updated, _ := sub.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	return updated
}

func (m TurnModel) Init() tea.Cmd {
	if m.subModel != nil {
		return m.subModel.Init()
	}
	return nil
}

func (m TurnModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.subModel != nil {
			updated, cmd := m.subModel.Update(msg)
			m.subModel = updated
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		if m.state == stateActionResult {
			return m.commitAndAdvance()
		}
		if m.state == stateAttackRedirect {
			return m.handleRedirectKey(msg)
		}

	case inputs.AttackInputsSelectedMsg:
		return m.startAttackResolution(msg)

	case AttackRedirectMsg:
		m.pendingRedirect = &msg
		m.state = stateAttackRedirect
		m.subModel = nil
		return m, nil

	case AttackCompletedMsg:
		return m.handleAttackCompleted(msg)

	case inputs.AssetSelectedMsg:
		collector := &TUICollector{selectedAsset: msg.Asset}
		return m.runAction(actions.NewSellAsset(collector))

	case inputs.BuyOrderSelectedMsg:
		collector := &TUICollector{buyOrder: msg.Order}
		return m.runAction(actions.NewBuyAsset(collector))

	case inputs.RefitOrderSelectedMsg:
		collector := &TUICollector{refitOrder: msg.Order}
		return m.runAction(actions.NewRefitAsset(collector))

	case inputs.RepairOrdersSelectedMsg:
		collector := &TUICollector{repairOrders: msg.Orders}
		return m.runAction(actions.NewRepairAsset(collector))

	case phases.ResumeChoiceMsg:
		return m.handleResumeChoice(msg.Choice)

	case phases.SkipChoiceMsg:
		return m.handleSkipChoice(msg.Skip)

	case phases.BookkeepingDoneMsg:
		available := m.engine.Action.AvailableActions(m.currentFaction, m.factionState, m.engine.Rulebook)
		m.availableActions = available
		m.state = stateActionSelect
		m.subModel = m.resizeSub(phases.NewActionSelectModel(available))
		return m, m.subModel.Init()

	case phases.ActionSelectedMsg:
		return m.handleActionSelected(msg.Index)

	case phases.SummaryDoneMsg:
		m.state = stateDone
		return m, tea.Quit
	}

	if m.subModel != nil {
		updated, cmd := m.subModel.Update(msg)
		m.subModel = updated
		return m, cmd
	}

	return m, nil
}

func (m TurnModel) handleResumeChoice(choice phases.ResumeChoice) (tea.Model, tea.Cmd) {
	switch choice {
	case phases.ChoiceAbandon:
		m.engine.Turn.Abandon(m.factionState)
		if err := state.Save(m.paths.State, m.factionState); err != nil {
			m.err = err
			return m, nil
		}
		m.state = stateDone
		return m, tea.Quit

	case phases.ChoiceResume:
		faction, err := m.engine.Turn.CurrentFaction(m.factionState)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.currentFaction = faction

	case phases.ChoiceStartNew:
		if err := m.engine.Turn.Start(m.factionState); err != nil {
			m.err = err
			return m, nil
		}
		if err := state.Save(m.paths.State, m.factionState); err != nil {
			m.err = err
			return m, nil
		}
		faction, err := m.engine.Turn.CurrentFaction(m.factionState)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.currentFaction = faction
	}

	m.snapshots[m.currentFaction.ID] = factionSnapshot{hp: m.currentFaction.CurrentHP, coin: m.currentFaction.Coin}
	m.state = stateSkipPrompt
	m.subModel = m.resizeSub(phases.NewSkipTurnModel(m.currentFaction.Name))
	return m, m.subModel.Init()
}

func (m TurnModel) handleSkipChoice(skip bool) (tea.Model, tea.Cmd) {
	if !skip {
		result, err := m.engine.Turn.ApplyBookkeeping(m.factionState)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.bookkeepingResult = result
		m.pendingMutations = result.RecordedMutations
		m.state = stateBookkeeping
		m.subModel = m.resizeSub(phases.NewBookkeepingModel(result, m.engine.Rulebook))
		return m, m.subModel.Init()
	}

	m.actionsTaken[m.currentFaction.ID] = "Skipped"
	event, err := buildEventRecord(m.factionState, m.currentFaction, nil)
	if err != nil {
		m.err = err
		return m, nil
	}
	done, err := m.engine.Turn.Advance(m.factionState)
	if err != nil {
		m.err = err
		return m, nil
	}
	if err := state.Save(m.paths.State, m.factionState); err != nil {
		m.err = err
		return m, nil
	}
	if err := m.engine.History.Record(m.paths.History, event); err != nil {
		m.err = err
		return m, nil
	}
	if done {
		m.state = stateCycleSummary
		m.subModel = m.resizeSub(phases.NewCycleSummaryModel(m.factionState.CycleNumber, m.buildSummaryRows()))
		return m, m.subModel.Init()
	}

	faction, err := m.engine.Turn.CurrentFaction(m.factionState)
	if err != nil {
		m.err = err
		return m, nil
	}
	m.currentFaction = faction
	m.snapshots[faction.ID] = factionSnapshot{hp: faction.CurrentHP, coin: faction.Coin}
	m.subModel = phases.NewSkipTurnModel(m.currentFaction.Name)
	return m, m.subModel.Init()
}

func (m TurnModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	switch m.state {
	case stateResumePrompt:
		if m.subModel != nil {
			return m.subModel.View()
		}
		return "Loading..."

	case stateActionResult:
		right := fmt.Sprintf("%s\n\n%s",
			style.SectionTitle.Render(m.actionResultText),
			style.Muted.Render("Press any key to continue"),
		)
		return renderSplitPanel(renderLeft(m), right, m.width)

	case stateAttackRedirect:
		right := renderRedirectPrompt(m.pendingRedirect)
		return renderSplitPanel(renderLeft(m), right, m.width)

	case stateCycleSummary:
		if m.subModel != nil {
			return m.subModel.View()
		}
		return ""

	case stateDone:
		return ""

	default:
		left := renderLeft(m)
		right := ""
		if m.subModel != nil {
			right = m.subModel.View()
		} else if m.attackEventCh != nil {
			right = style.Muted.Render("Resolving attack...")
		}
		return renderSplitPanel(left, right, m.width)
	}
}

func renderLeft(m TurnModel) string {
	if m.currentFaction == nil {
		return ""
	}
	f := m.currentFaction

	var sb strings.Builder

	sb.WriteString(style.Header.Render(f.Name))
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "Scale: %s\n", f.Scale)

	hpStr := fmt.Sprintf("%d/%d", f.CurrentHP, f.MaxHP)
	if f.CurrentHP <= f.MaxHP/4 {
		hpStr = style.LowHP.Render(hpStr)
	} else {
		hpStr = style.HP.Render(hpStr)
	}
	fmt.Fprintf(&sb, "HP: %s\n", hpStr)
	fmt.Fprintf(&sb, "Coin: %s\n", style.Coin.Render(fmt.Sprintf("%d", f.Coin)))

	goalName := "(none)"
	if f.Goal != nil {
		goalName = f.Goal.Name
	}
	fmt.Fprintf(&sb, "Goal: %s\n\n", style.Muted.Render(goalName))

	sb.WriteString(style.SectionTitle.Render("Stats"))
	fmt.Fprintf(&sb, "\nForce   %d\nCunning %d\nWealth  %d", f.Force, f.Cunning, f.Wealth)

	if len(f.Assets) > 0 {
		sb.WriteString("\n\n")
		sb.WriteString(style.SectionTitle.Render("Assets"))
		for _, asset := range f.Assets {
			flags := assetFlags(asset)
			fmt.Fprintf(&sb, "\n%-18s %s %2d/%-2d%s",
				truncate(asset.DefinitionID, 18),
				assetCategoryAbbrev(asset),
				asset.CurrentHP,
				asset.CurrentHP,
				flags,
			)
		}
	}

	if len(f.Bases) > 0 {
		sb.WriteString("\n\n")
		sb.WriteString(style.SectionTitle.Render("Bases"))
		for _, base := range f.Bases {
			tag := ""
			if base.IsHomeworld {
				tag = style.Muted.Render(" (home)")
			}
			fmt.Fprintf(&sb, "\n%-16s HP %d/%d%s",
				truncate(base.Location, 16),
				base.CurrentHP,
				base.EffectiveMaxHP(f),
				tag,
			)
		}
	}

	return sb.String()
}

func renderRedirectPrompt(msg *AttackRedirectMsg) string {
	if msg == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(style.SectionTitle.Render("Redirect Damage?"))
	sb.WriteString("\n\n")
	fmt.Fprintf(&sb, "%s has a Base of Influence here.\n\n", msg.DefFaction.Name)
	fmt.Fprintf(&sb, "Redirect %s to the Base instead of the asset?\n\n",
		style.Coin.Render(fmt.Sprintf("%d damage", msg.Damage)),
	)
	sb.WriteString(style.Muted.Render("y — redirect to Base   n — hit the asset"))
	return sb.String()
}

func assetCategoryAbbrev(_ *domain.Asset) string {
	return "?"
}

func assetFlags(asset *domain.Asset) string {
	var flags []string
	if !asset.Ready {
		flags = append(flags, "NR")
	}
	if !asset.Maintained {
		flags = append(flags, "UM")
	}
	if asset.Stealthy {
		flags = append(flags, "ST")
	}
	if len(flags) == 0 {
		return ""
	}
	return " " + style.Muted.Render("["+strings.Join(flags, ",")+"]")
}

func (m TurnModel) handleActionSelected(index int) (tea.Model, tea.Cmd) {
	if index == -1 {
		m.actionResultText = "No Action"
		return m.commitAndAdvance()
	}
	m.pendingAction = m.availableActions[index]
	switch m.pendingAction.Name() {
	case "Repair Faction":
		return m.runAction(actions.NewRepairFaction())
	case "Repair Asset":
		var damagedAssets []*domain.Asset
		for _, asset := range m.currentFaction.Assets {
			if def, ok := m.engine.Rulebook.Assets[asset.DefinitionID]; ok && asset.CurrentHP < def.HP {
				damagedAssets = append(damagedAssets, asset)
			}
		}
		m.state = stateActionInput
		m.subModel = m.resizeSub(inputs.NewRepairOrdersModel(m.currentFaction, damagedAssets, m.engine.Rulebook))
		return m, m.subModel.Init()
	case "Sell Asset":
		m.state = stateActionInput
		m.subModel = m.resizeSub(inputs.NewSelectAssetModel(m.currentFaction.Assets, m.engine.Rulebook))
		return m, m.subModel.Init()
	case "Buy Asset":
		worlds := tuiAvailableWorlds(m.currentFaction)
		purchasable := tuiPurchasableDefinitions(m.currentFaction, m.engine.Rulebook)
		m.state = stateActionInput
		m.subModel = m.resizeSub(inputs.NewBuyOrderModel(worlds, purchasable))
		return m, m.subModel.Init()
	case "Refit Asset":
		options := tuiRefitOptions(m.currentFaction, m.engine.Rulebook)
		m.state = stateActionInput
		m.subModel = m.resizeSub(inputs.NewRefitOrderModel(options, m.engine.Rulebook))
		return m, m.subModel.Init()
	case "Attack":
		eligible := tuiEligibleAttackers(m.currentFaction, m.factionState, m.engine.Rulebook)
		m.state = stateActionInput
		m.subModel = m.resizeSub(inputs.NewAttackInputsModel(eligible, m.factionState, m.engine.Rulebook))
		return m, m.subModel.Init()
	default:
		m.actionResultText = m.pendingAction.Name() + " — not yet implemented in TUI"
		m.state = stateActionResult
		m.subModel = nil
		return m, nil
	}
}

func (m TurnModel) startAttackResolution(msg inputs.AttackInputsSelectedMsg) (tea.Model, tea.Cmd) {
	eventCh := make(chan tea.Msg, 1)
	m.attackEventCh = eventCh
	collector := &TUICollector{
		attackers: msg.Attackers,
		defenders: msg.Defenders,
		eventCh:   eventCh,
	}
	action := actions.NewAttack(collector, engine.NewRandRoller())
	faction := m.currentFaction
	factionState := m.factionState
	rulebook := m.engine.Rulebook
	go func() {
		mutations, err := m.engine.Action.Run(action, faction, factionState, rulebook)
		eventCh <- AttackCompletedMsg{Mutations: mutations, Err: err}
	}()
	m.state = stateActionInput
	m.subModel = nil
	return m, waitForAttackEvent(eventCh)
}

func (m TurnModel) handleAttackCompleted(msg AttackCompletedMsg) (tea.Model, tea.Cmd) {
	m.attackEventCh = nil
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.engine.Mutation.Apply(m.factionState, msg.Mutations)
	m.pendingMutations = append(m.pendingMutations, msg.Mutations...)
	m.actionResultText = "Attack"
	m.state = stateActionResult
	m.subModel = nil
	return m, nil
}

func (m TurnModel) handleRedirectKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "y", "Y":
		m.pendingRedirect.ResponseCh <- true
	case "n", "N":
		m.pendingRedirect.ResponseCh <- false
	default:
		return m, nil
	}
	m.pendingRedirect = nil
	m.state = stateActionInput
	m.subModel = nil
	return m, waitForAttackEvent(m.attackEventCh)
}

func waitForAttackEvent(eventCh chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-eventCh
	}
}

func (m TurnModel) runAction(action engine.Action) (tea.Model, tea.Cmd) {
	mutations, err := m.engine.Action.Run(action, m.currentFaction, m.factionState, m.engine.Rulebook)
	if err != nil {
		m.err = err
		return m, nil
	}
	m.engine.Mutation.Apply(m.factionState, mutations)
	m.pendingMutations = append(m.pendingMutations, mutations...)
	m.actionResultText = action.Name()
	m.state = stateActionResult
	m.subModel = nil
	return m, nil
}

func (m TurnModel) commitAndAdvance() (tea.Model, tea.Cmd) {
	m.actionsTaken[m.currentFaction.ID] = m.actionResultText
	event, err := buildEventRecord(m.factionState, m.currentFaction, m.pendingMutations)
	if err != nil {
		m.err = err
		return m, nil
	}
	done, err := m.engine.Turn.Advance(m.factionState)
	if err != nil {
		m.err = err
		return m, nil
	}
	if err := state.Save(m.paths.State, m.factionState); err != nil {
		m.err = err
		return m, nil
	}
	if err := m.engine.History.Record(m.paths.History, event); err != nil {
		m.err = err
		return m, nil
	}
	m.pendingMutations = nil
	m.actionResultText = ""

	if done {
		m.state = stateCycleSummary
		m.subModel = m.resizeSub(phases.NewCycleSummaryModel(m.factionState.CycleNumber, m.buildSummaryRows()))
		return m, m.subModel.Init()
	}
	faction, err := m.engine.Turn.CurrentFaction(m.factionState)
	if err != nil {
		m.err = err
		return m, nil
	}
	m.currentFaction = faction
	m.snapshots[faction.ID] = factionSnapshot{hp: faction.CurrentHP, coin: faction.Coin}
	m.state = stateSkipPrompt
	m.subModel = m.resizeSub(phases.NewSkipTurnModel(faction.Name))
	return m, m.subModel.Init()
}

func (m TurnModel) buildSummaryRows() []phases.FactionSummaryRow {
	rows := make([]phases.FactionSummaryRow, 0, len(m.factionState.Factions))
	for _, faction := range m.factionState.Factions {
		snap, hasSnap := m.snapshots[faction.ID]
		row := phases.FactionSummaryRow{
			Name:   faction.Name,
			EndHP:  faction.CurrentHP,
			EndCoin: faction.Coin,
			Action: m.actionsTaken[faction.ID],
		}
		if hasSnap {
			row.StartHP = snap.hp
			row.StartCoin = snap.coin
		} else {
			row.StartHP = faction.CurrentHP
			row.StartCoin = faction.Coin
		}
		rows = append(rows, row)
	}
	return rows
}

func buildEventRecord(factionState *state.FactionState, faction *domain.Faction, mutations []domain.Mutation) (domain.EventRecord, error) {
	records := make([]domain.MutationRecord, 0, len(mutations))
	for _, mutation := range mutations {
		record, err := domain.NewMutationRecord(mutation)
		if err != nil {
			return domain.EventRecord{}, fmt.Errorf("building mutation record: %w", err)
		}
		records = append(records, record)
	}
	return domain.EventRecord{
		Cycle:     factionState.CycleNumber,
		FactionID: faction.ID,
		Timestamp: time.Now().UTC(),
		Mutations: records,
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func RunTurnTUI(e *engine.Engine, factionState *state.FactionState, p paths.Paths) error {
	m := TurnModel{
		engine:       e,
		factionState: factionState,
		paths:        p,
		state:        stateResumePrompt,
		subModel:     phases.NewResumeTurnModel(e.Turn.InProgress(factionState)),
		snapshots:    make(map[string]factionSnapshot),
		actionsTaken: make(map[string]string),
	}

	prog := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stderr),
	)
	_, err := prog.Run()
	return err
}
