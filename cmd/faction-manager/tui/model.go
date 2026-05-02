package tui

import (
	"fmt"
	"os"
	"sort"
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
	stateResumePrompt turnState = iota
	stateSkipPrompt
	stateGoalSelect
	stateGoalLocked
	stateBookkeeping
	stateActionSelect
	stateActionInput
	stateActionResult
	stateAttackRedirect                 // waiting for redirect confirm mid-resolution
	stateExpandInfluenceRivalConfirm    // waiting for rival free-attack y/n
	stateExpandInfluenceSelectAttackers // waiting for base attacker selection
	stateAbilityMoveDestination         // waiting for move destination mid-resolution
	stateAbilityFactionTestTarget       // waiting for faction test target mid-resolution
	stateAbilityConfirmApplied          // waiting for GM fallback confirm mid-resolution
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

// ExpandInfluenceCompletedMsg is sent by the Expand Influence resolution goroutine
// when Resolve finishes (successfully or with an error).
type ExpandInfluenceCompletedMsg struct {
	Mutations []domain.Mutation
	Err       error
}

// AbilityCompletedMsg is sent by the Use Asset Ability resolution goroutine
// when Resolve finishes (successfully or with an error).
type AbilityCompletedMsg struct {
	Mutations []domain.Mutation
	Err       error
}

type factionSnapshot struct {
	hp   int
	coin int
}

type TurnModel struct {
	engine                    *engine.Engine
	factionState              *state.FactionState
	paths                     paths.Paths
	width                     int
	height                    int
	state                     turnState
	currentFaction            *domain.Faction
	bookkeepingResult         engine.BookkeepingResult
	pendingMutations          []domain.Mutation
	actionResultText          string
	availableActions          []engine.Action
	pendingAction             engine.Action
	subModel                  tea.Model
	attackEventCh             chan tea.Msg
	attackCollector           *TUICollector
	pendingRedirect           *AttackRedirectMsg
	expandInfluenceEventCh    chan tea.Msg
	pendingRivalAttack        *ExpandInfluenceRivalMsg
	pendingBaseAttackers      *ExpandInfluenceBaseAttackersMsg
	abilityEventCh            chan tea.Msg
	pendingAbilityMove        *AbilityMoveMsg
	pendingAbilityFactionTest *AbilityFactionTestMsg
	pendingAbilityConfirm     *AbilityConfirmMsg
	goalLock                  engine.GoalLock
	pendingGoalLockMutations  []domain.Mutation
	lockedGoalDestination     string
	snapshots                 map[string]factionSnapshot // faction ID → pre-turn HP/Coin
	actionsTaken              map[string]string          // faction ID → action description
	actionResults             map[string]string          // faction ID → result summary for cycle summary
	turnLog                   []string                   // play-by-play lines for current faction's action
	err                       error
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
		if m.state == stateGoalLocked {
			return m.handleGoalLockedAck()
		}
		if m.state == stateAttackRedirect {
			return m.handleRedirectKey(msg)
		}
		if m.state == stateExpandInfluenceRivalConfirm {
			return m.handleRivalConfirmKey(msg)
		}
		if m.state == stateAbilityConfirmApplied {
			return m.handleAbilityConfirmKey(msg)
		}

	case inputs.AttackInputsSelectedMsg:
		return m.startAttackResolution(msg)

	case inputs.ExpandInfluenceOrderSelectedMsg:
		return m.startExpandInfluenceResolution(msg)

	case inputs.BaseAttackersSelectedMsg:
		return m.handleBaseAttackersSelected(msg)

	case AttackRedirectMsg:
		m.pendingRedirect = &msg
		m.state = stateAttackRedirect
		m.subModel = nil
		return m, nil

	case AttackCompletedMsg:
		return m.handleAttackCompleted(msg)

	case ExpandInfluenceRivalMsg:
		m.pendingRivalAttack = &msg
		m.state = stateExpandInfluenceRivalConfirm
		m.subModel = nil
		return m, nil

	case ExpandInfluenceBaseAttackersMsg:
		m.pendingBaseAttackers = &msg
		m.state = stateExpandInfluenceSelectAttackers
		m.subModel = m.resizeSub(inputs.NewSelectBaseAttackersModel(msg.Rival, msg.Eligible, m.engine.Rulebook))
		return m, m.subModel.Init()

	case ExpandInfluenceCompletedMsg:
		return m.handleExpandInfluenceCompleted(msg)

	case inputs.AbilityAssetsSelectedMsg:
		return m.startAbilityResolution(msg)

	case AbilityMoveMsg:
		m.pendingAbilityMove = &msg
		m.state = stateAbilityMoveDestination
		m.subModel = m.resizeSub(inputs.NewMoveDestinationModel(msg.Asset, msg.Worlds, m.engine.Rulebook))
		return m, m.subModel.Init()

	case AbilityFactionTestMsg:
		m.pendingAbilityFactionTest = &msg
		m.state = stateAbilityFactionTestTarget
		m.subModel = m.resizeSub(inputs.NewFactionTestTargetModel(msg.Asset, msg.Effect, msg.Candidates, m.engine.Rulebook))
		return m, m.subModel.Init()

	case AbilityConfirmMsg:
		m.pendingAbilityConfirm = &msg
		m.state = stateAbilityConfirmApplied
		m.subModel = nil
		return m, nil

	case inputs.AbilityMoveDestinationSelectedMsg:
		return m.handleAbilityMoveDestinationSelected(msg)

	case inputs.FactionTestTargetSelectedMsg:
		return m.handleFactionTestTargetSelected(msg)

	case AbilityCompletedMsg:
		return m.handleAbilityCompleted(msg)

	case inputs.AssetSelectedMsg:
		collector := &TUICollector{selectedAsset: msg.Asset}
		return m.runAction(actions.NewSellAsset(collector))

	case inputs.BuyOrderSelectedMsg:
		collector := &TUICollector{buyOrder: msg.Order, selectedAsset: msg.StealthTarget}
		return m.runAction(actions.NewBuyAsset(collector))

	case inputs.BribeOrderSelectedMsg:
		collector := &TUICollector{bribeBase: msg.Base, bribeAmount: msg.Amount}
		return m.runAction(actions.NewBribe(collector))

	case inputs.SeizePlanetTargetSelectedMsg:
		collector := &TUICollector{seizeWorld: msg.World}
		return m.runAction(actions.NewSeizePlanet(collector))

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

	case phases.GoalSelectedMsg:
		m.currentFaction.ActiveGoal = msg.ActiveGoal
		return m.proceedAfterGoalSelect()

	case phases.BookkeepingDoneMsg:
		available := m.engine.Action.AvailableActions(m.currentFaction, m.factionState, m.engine.Rulebook)
		if m.goalLock.Type == engine.LockRestrictActions {
			available = filterAllowedActions(available, m.goalLock.AllowedActions)
		}
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
		if m.currentFaction.ActiveGoal == nil {
			m.state = stateGoalSelect
			m.subModel = m.resizeSub(phases.NewGoalSelectModel(m.currentFaction, m.engine.Rulebook))
			return m, m.subModel.Init()
		}
		return m.proceedAfterGoalSelect()
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

func (m TurnModel) appendLog(content string) string {
	log := renderLogSection(m.turnLog)
	if log == "" {
		return content
	}
	return content + "\n\n" + log
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
		return renderSplitPanel(renderLeft(m), m.appendLog(right), m.width)

	case stateAttackRedirect:
		right := renderRedirectPrompt(m.pendingRedirect)
		return renderSplitPanel(renderLeft(m), m.appendLog(right), m.width)

	case stateExpandInfluenceRivalConfirm:
		right := renderRivalConfirmPrompt(m.pendingRivalAttack)
		return renderSplitPanel(renderLeft(m), m.appendLog(right), m.width)

	case stateExpandInfluenceSelectAttackers, stateAbilityMoveDestination, stateAbilityFactionTestTarget:
		left := renderLeft(m)
		right := ""
		if m.subModel != nil {
			right = m.subModel.View()
		}
		return renderSplitPanel(left, right, m.width)

	case stateGoalLocked:
		right := renderGoalLockedPrompt(m)
		return renderSplitPanel(renderLeft(m), m.appendLog(right), m.width)

	case stateAbilityConfirmApplied:
		right := renderAbilityConfirmPrompt(m.pendingAbilityConfirm)
		return renderSplitPanel(renderLeft(m), m.appendLog(right), m.width)

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
		} else if m.abilityEventCh != nil {
			right = style.Muted.Render("Resolving ability...")
		}
		return renderSplitPanel(left, m.appendLog(right), m.width)
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
	if f.ActiveGoal != nil {
		if goal, ok := m.engine.Rulebook.Goals[f.ActiveGoal.GoalID]; ok {
			goalName = goal.Name
		} else {
			goalName = f.ActiveGoal.GoalID
		}
	}
	fmt.Fprintf(&sb, "Goal: %s\n", style.Muted.Render(goalName))
	if f.ActiveGoal != nil && f.ActiveGoal.GoalID == "G-012" && f.ActiveGoal.TurnsRemaining > 0 {
		fmt.Fprintf(&sb, "%s\n\n", style.Muted.Render(fmt.Sprintf("→ %s (%d turns)", f.ActiveGoal.TargetWorld, f.ActiveGoal.TurnsRemaining)))
	} else {
		sb.WriteString("\n")
	}

	sb.WriteString(style.SectionTitle.Render("Stats"))
	fmt.Fprintf(&sb, "\nForce   %d\nCunning %d\nWealth  %d", f.Force, f.Cunning, f.Wealth)

	if len(f.Assets) > 0 {
		sb.WriteString("\n\n")
		sb.WriteString(style.SectionTitle.Render("Assets"))
		for _, asset := range f.Assets {
			flags := assetFlags(asset)
			name := asset.DefinitionID
			maxHP := asset.CurrentHP
			catAbbrev := "?"
			if def, ok := m.engine.Rulebook.Assets[asset.DefinitionID]; ok {
				name = def.Name
				maxHP = def.HP
				catAbbrev = assetCategoryAbbrev(def.Category)
			}
			fmt.Fprintf(&sb, "\n%-18s %s %2d/%-2d%s",
				truncate(name, 18),
				catAbbrev,
				asset.CurrentHP,
				maxHP,
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

func renderAbilityConfirmPrompt(msg *AbilityConfirmMsg) string {
	if msg == nil {
		return ""
	}
	name := msg.Asset.DefinitionID
	if msg.Def != nil {
		name = msg.Def.Name
	}
	var sb strings.Builder
	sb.WriteString(style.SectionTitle.Render("Ability Applied?"))
	sb.WriteString("\n\n")
	fmt.Fprintf(&sb, "%s\n\n", name)
	if msg.Def != nil && msg.Def.Description != "" {
		fmt.Fprintf(&sb, "%s\n\n", style.Muted.Render(msg.Def.Description))
	}
	sb.WriteString(style.Muted.Render("y — applied   n — skip"))
	return sb.String()
}

func renderRivalConfirmPrompt(msg *ExpandInfluenceRivalMsg) string {
	if msg == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(style.SectionTitle.Render("Rival Free Attack?"))
	sb.WriteString("\n\n")
	fmt.Fprintf(&sb, "%s matched or beat your contested roll.\n\n", msg.Rival.Name)
	fmt.Fprintf(&sb, "Your roll: %s   %s's roll: %s\n\n",
		style.HP.Render(fmt.Sprintf("%d", msg.FactionRoll)),
		msg.Rival.Name,
		style.LowHP.Render(fmt.Sprintf("%d", msg.RivalRoll)),
	)
	fmt.Fprintf(&sb, "Does %s make a free attack against the new Base?\n\n", msg.Rival.Name)
	sb.WriteString(style.Muted.Render("y — attack   n — pass"))
	return sb.String()
}

func assetCategoryAbbrev(category domain.FactionStat) string {
	switch category {
	case domain.StatForce:
		return "F"
	case domain.StatCunning:
		return "C"
	case domain.StatWealth:
		return "W"
	default:
		return "?"
	}
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
	case "Abandon Goal":
		return m.runAction(actions.NewAbandonGoal())
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
		m.subModel = m.resizeSub(inputs.NewBuyOrderModel(worlds, purchasable, m.currentFaction, m.engine.Rulebook))
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
	case "Expand Influence":
		m.state = stateActionInput
		m.subModel = m.resizeSub(inputs.NewExpandInfluenceModel(m.currentFaction))
		return m, m.subModel.Init()
	case "Use Asset Ability":
		candidates := tuiEligibleAbilityAssets(m.currentFaction, m.engine.Rulebook)
		m.state = stateActionInput
		m.subModel = m.resizeSub(inputs.NewAbilityAssetsModel(candidates, m.engine.Rulebook))
		return m, m.subModel.Init()
	case "Bribe":
		m.state = stateActionInput
		m.subModel = m.resizeSub(inputs.NewBribeModel(m.currentFaction))
		return m, m.subModel.Init()
	case "Seize Planet":
		worlds := tuiSeizePlanetWorlds(m.currentFaction, m.factionState)
		m.state = stateActionInput
		m.subModel = m.resizeSub(inputs.NewSeizePlanetModel(worlds))
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
	m.attackCollector = collector
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
	goalMutations := m.engine.Goal.UpdateProgress(m.currentFaction.ID, msg.Mutations, m.factionState, m.engine.Rulebook)
	allMutations := append(msg.Mutations, goalMutations...)
	if m.attackCollector != nil {
		m.turnLog = append(m.turnLog, narrateAttack(m.attackCollector, allMutations, m.factionState, m.engine.Rulebook)...)
		m.attackCollector = nil
	}
	m.turnLog = append(m.turnLog, narrateGoalEvents(allMutations, m.engine.Rulebook)...)
	m.engine.Mutation.Apply(m.factionState, allMutations)
	m.pendingMutations = append(m.pendingMutations, allMutations...)
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
	return func() tea.Msg { return <-eventCh }
}

func (m TurnModel) startExpandInfluenceResolution(msg inputs.ExpandInfluenceOrderSelectedMsg) (tea.Model, tea.Cmd) {
	eventCh := make(chan tea.Msg, 1)
	m.expandInfluenceEventCh = eventCh
	collector := &TUICollector{
		expandInfluenceOrder: msg.Order,
		eventCh:              eventCh,
	}
	action := actions.NewExpandInfluence(collector, engine.NewRandRoller())
	faction := m.currentFaction
	factionState := m.factionState
	rulebook := m.engine.Rulebook
	go func() {
		mutations, err := m.engine.Action.Run(action, faction, factionState, rulebook)
		eventCh <- ExpandInfluenceCompletedMsg{Mutations: mutations, Err: err}
	}()
	m.state = stateActionInput
	m.subModel = nil
	return m, waitForExpandInfluenceEvent(eventCh)
}

func (m TurnModel) handleExpandInfluenceCompleted(msg ExpandInfluenceCompletedMsg) (tea.Model, tea.Cmd) {
	m.expandInfluenceEventCh = nil
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	goalMutations := m.engine.Goal.UpdateProgress(m.currentFaction.ID, msg.Mutations, m.factionState, m.engine.Rulebook)
	allMutations := append(msg.Mutations, goalMutations...)
	m.turnLog = append(m.turnLog, narrateExpandInfluence(allMutations, m.currentFaction)...)
	m.turnLog = append(m.turnLog, narrateGoalEvents(allMutations, m.engine.Rulebook)...)
	m.engine.Mutation.Apply(m.factionState, allMutations)
	m.pendingMutations = append(m.pendingMutations, allMutations...)
	m.actionResultText = "Expand Influence"
	m.state = stateActionResult
	m.subModel = nil
	return m, nil
}

func (m TurnModel) handleRivalConfirmKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "y", "Y":
		m.pendingRivalAttack.ResponseCh <- true
	case "n", "N":
		m.pendingRivalAttack.ResponseCh <- false
	default:
		return m, nil
	}
	m.pendingRivalAttack = nil
	m.state = stateActionInput
	m.subModel = nil
	return m, waitForExpandInfluenceEvent(m.expandInfluenceEventCh)
}

func (m TurnModel) handleBaseAttackersSelected(msg inputs.BaseAttackersSelectedMsg) (tea.Model, tea.Cmd) {
	if m.pendingBaseAttackers != nil {
		m.pendingBaseAttackers.ResponseCh <- msg.Attackers
		m.pendingBaseAttackers = nil
	}
	m.state = stateActionInput
	m.subModel = nil
	return m, waitForExpandInfluenceEvent(m.expandInfluenceEventCh)
}

func waitForExpandInfluenceEvent(eventCh chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-eventCh }
}

func (m TurnModel) startAbilityResolution(msg inputs.AbilityAssetsSelectedMsg) (tea.Model, tea.Cmd) {
	eventCh := make(chan tea.Msg, 1)
	m.abilityEventCh = eventCh
	collector := &TUICollector{
		abilityAssets: msg.Assets,
		eventCh:       eventCh,
	}
	action := actions.NewUseAssetAbility(collector, engine.NewRandRoller(), m.engine.AbilityEngine)
	faction := m.currentFaction
	factionState := m.factionState
	rulebook := m.engine.Rulebook
	go func() {
		mutations, err := m.engine.Action.Run(action, faction, factionState, rulebook)
		eventCh <- AbilityCompletedMsg{Mutations: mutations, Err: err}
	}()
	m.state = stateActionInput
	m.subModel = nil
	return m, waitForAbilityEvent(eventCh)
}

func (m TurnModel) handleAbilityCompleted(msg AbilityCompletedMsg) (tea.Model, tea.Cmd) {
	m.abilityEventCh = nil
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	goalMutations := m.engine.Goal.UpdateProgress(m.currentFaction.ID, msg.Mutations, m.factionState, m.engine.Rulebook)
	allMutations := append(msg.Mutations, goalMutations...)
	m.turnLog = append(m.turnLog, narrateUseAssetAbility(allMutations, m.currentFaction, m.factionState, m.engine.Rulebook)...)
	m.turnLog = append(m.turnLog, narrateGoalEvents(allMutations, m.engine.Rulebook)...)
	m.engine.Mutation.Apply(m.factionState, allMutations)
	m.pendingMutations = append(m.pendingMutations, allMutations...)
	m.actionResultText = "Use Asset Ability"
	m.state = stateActionResult
	m.subModel = nil
	return m, nil
}

func (m TurnModel) handleAbilityMoveDestinationSelected(msg inputs.AbilityMoveDestinationSelectedMsg) (tea.Model, tea.Cmd) {
	if m.pendingAbilityMove != nil {
		m.pendingAbilityMove.ResponseCh <- msg.Destination
		m.pendingAbilityMove = nil
	}
	m.state = stateActionInput
	m.subModel = nil
	return m, waitForAbilityEvent(m.abilityEventCh)
}

func (m TurnModel) handleFactionTestTargetSelected(msg inputs.FactionTestTargetSelectedMsg) (tea.Model, tea.Cmd) {
	if m.pendingAbilityFactionTest != nil {
		m.pendingAbilityFactionTest.ResponseCh <- msg.Faction
		m.pendingAbilityFactionTest = nil
	}
	m.state = stateActionInput
	m.subModel = nil
	return m, waitForAbilityEvent(m.abilityEventCh)
}

func (m TurnModel) handleAbilityConfirmKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pendingAbilityConfirm == nil {
		return m, nil
	}
	switch key.String() {
	case "y", "Y":
		m.pendingAbilityConfirm.ResponseCh <- true
	case "n", "N":
		m.pendingAbilityConfirm.ResponseCh <- false
	default:
		return m, nil
	}
	m.pendingAbilityConfirm = nil
	m.state = stateActionInput
	m.subModel = nil
	return m, waitForAbilityEvent(m.abilityEventCh)
}

func waitForAbilityEvent(eventCh chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-eventCh }
}

// proceedAfterGoalSelect calls CheckLock and routes to stateGoalLocked or
// bookkeeping depending on the result.
func (m TurnModel) proceedAfterGoalSelect() (tea.Model, tea.Cmd) {
	if m.currentFaction.ActiveGoal != nil && m.currentFaction.ActiveGoal.GoalID == "G-012" {
		m.lockedGoalDestination = m.currentFaction.ActiveGoal.TargetWorld
	}
	lock, lockMutations := m.engine.Goal.CheckLock(m.currentFaction, m.factionState, m.engine.Rulebook)
	m.goalLock = lock
	m.pendingGoalLockMutations = lockMutations
	if lock.Type == engine.LockSkip {
		// Narrate homeworld completion when TurnsRemaining just hit 0.
		if len(lockMutations) > 0 {
			m.turnLog = append(m.turnLog, narrateGoalEvents(lockMutations, m.engine.Rulebook)...)
		}
		m.state = stateGoalLocked
		m.subModel = nil
		return m, nil
	}
	// LockNone / LockRestrictActions: apply any CheckLock mutations (e.g. GoalAbandoned
	// on Planetary Seizure occupation fail) before bookkeeping runs.
	if len(lockMutations) > 0 {
		m.turnLog = append(m.turnLog, narrateGoalEvents(lockMutations, m.engine.Rulebook)...)
		m.engine.Mutation.Apply(m.factionState, lockMutations)
		m.pendingMutations = append(m.pendingMutations, lockMutations...)
		m.pendingGoalLockMutations = nil
	}
	return m.startBookkeeping()
}

func (m TurnModel) startBookkeeping() (tea.Model, tea.Cmd) {
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

func (m TurnModel) handleGoalLockedAck() (tea.Model, tea.Cmd) {
	if len(m.pendingGoalLockMutations) > 0 {
		m.engine.Mutation.Apply(m.factionState, m.pendingGoalLockMutations)
	}
	m.actionsTaken[m.currentFaction.ID] = "Change Homeworld (transit)"
	event, err := buildEventRecord(m.factionState, m.currentFaction, m.pendingGoalLockMutations)
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
	m.pendingGoalLockMutations = nil
	m.lockedGoalDestination = ""
	m.goalLock = engine.GoalLock{}
	m.turnLog = nil
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
	m.subModel = m.resizeSub(phases.NewSkipTurnModel(faction.Name))
	m.state = stateSkipPrompt
	return m, m.subModel.Init()
}

func renderGoalLockedPrompt(m TurnModel) string {
	var sb strings.Builder
	sb.WriteString(style.SectionTitle.Render("Change Homeworld — In Transit"))
	sb.WriteString("\n\n")
	if m.currentFaction.ActiveGoal != nil {
		fmt.Fprintf(&sb, "Moving to %s\n", m.lockedGoalDestination)
		fmt.Fprintf(&sb, "%s turns remaining\n\n",
			style.HP.Render(fmt.Sprintf("%d", m.currentFaction.ActiveGoal.TurnsRemaining)),
		)
	} else {
		fmt.Fprintf(&sb, "Homeworld relocated to %s\n\n", m.lockedGoalDestination)
	}
	sb.WriteString(style.Muted.Render("Press any key to continue"))
	return sb.String()
}

func filterAllowedActions(available []engine.Action, allowed []string) []engine.Action {
	set := make(map[string]bool, len(allowed))
	for _, name := range allowed {
		set[name] = true
	}
	filtered := make([]engine.Action, 0, len(allowed))
	for _, action := range available {
		if set[action.Name()] {
			filtered = append(filtered, action)
		}
	}
	return filtered
}

func (m TurnModel) runAction(action engine.Action) (tea.Model, tea.Cmd) {
	mutations, err := m.engine.Action.Run(action, m.currentFaction, m.factionState, m.engine.Rulebook)
	if err != nil {
		m.err = err
		return m, nil
	}
	goalMutations := m.engine.Goal.UpdateProgress(m.currentFaction.ID, mutations, m.factionState, m.engine.Rulebook)
	mutations = append(mutations, goalMutations...)
	m.turnLog = append(m.turnLog, narrateAction(action, mutations, m.currentFaction, m.engine.Rulebook)...)
	m.turnLog = append(m.turnLog, narrateGoalEvents(mutations, m.engine.Rulebook)...)
	m.engine.Mutation.Apply(m.factionState, mutations)
	m.pendingMutations = append(m.pendingMutations, mutations...)
	m.actionResultText = action.Name()
	m.state = stateActionResult
	m.subModel = nil
	return m, nil
}

func (m TurnModel) commitAndAdvance() (tea.Model, tea.Cmd) {
	m.actionsTaken[m.currentFaction.ID] = m.actionResultText
	m.actionResults[m.currentFaction.ID] = logSummary(m.turnLog)
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
	m.turnLog = nil

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
			Name:          faction.Name,
			EndHP:         faction.CurrentHP,
			EndCoin:       faction.Coin,
			Action:        m.actionsTaken[faction.ID],
			ResultSummary: m.actionResults[faction.ID],
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
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
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
		engine:        e,
		factionState:  factionState,
		paths:         p,
		state:         stateResumePrompt,
		subModel:      phases.NewResumeTurnModel(e.Turn.InProgress(factionState)),
		snapshots:     make(map[string]factionSnapshot),
		actionsTaken:  make(map[string]string),
		actionResults: make(map[string]string),
	}

	prog := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stderr),
	)
	_, err := prog.Run()
	return err
}
