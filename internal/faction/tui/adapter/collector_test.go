package adapter_test

import (
	"errors"
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
)

func newBridge(t *testing.T) (*adapter.Adapter, *testharness.Harness) {
	t.Helper()
	harness := testharness.NewHarness(t)
	return adapter.New(harness.Engine, harness.FactionState, harness.Paths, slog.New(slog.DiscardHandler)), harness
}

// answerNext answers the next ask on the channel, as the execution view does.
func answerNext(bridge *adapter.Adapter, reply any) {
	go func() {
		ask := bridge.AskPump()().(adapter.CollectorAskMsg)
		ask.Reply <- reply
	}()
}

type collectorCase struct {
	name string
	call func() (any, error)
	good any
	want any
}

func runCollectorCases(t *testing.T, bridge *adapter.Adapter, cases []collectorCase) {
	t.Helper()
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			answerNext(bridge, testCase.good)
			got, err := testCase.call()
			if err != nil || !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("good reply: got %#v, %v; want %#v", got, err, testCase.want)
			}

			answerNext(bridge, struct{}{})
			if _, err := testCase.call(); err == nil {
				t.Error("wrong reply type: got no error")
			}

			answerNext(bridge, nil)
			if _, err := testCase.call(); err != nil {
				t.Errorf("nil reply is a legal empty answer: got %v", err)
			}
		})
	}
}

func TestActionCollector_EveryReplyType(t *testing.T) {
	bridge, harness := newBridge(t)
	alpha := harness.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := harness.AddFaction("beta", "Tartarus", 2, 2, 2)
	base := testharness.AddBase(beta, "Tartarus", 5)
	if _, err := harness.Engine.World.RebuildIndex(harness.FactionState, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("RebuildIndex: %v", err)
	}
	asset := alpha.Assets["alpha-asset-1"]
	assets := []*domain.Asset{asset}
	rules := harness.Engine.Rulebook
	definition := rules.Assets[testharness.DefSecurityPersonnel]
	factionState := harness.FactionState
	collector := bridge.Action()

	relocateOrders := []action.RelocateOrder{{Asset: asset, DestinationWorld: "Erebus"}}
	repairOrders := []action.RepairOrder{{Asset: asset, HealCount: 1}}
	buyOrder := action.BuyOrder{World: "Tartarus", Definition: definition}
	refitOrder := action.RefitOrder{OldAsset: asset, NewDefinition: definition}
	expandOrder := action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Tartarus", HPAmount: 1}
	holders := []domain.AbilityHolder{{Asset: asset}}
	bribeReply := adapter.BribeReply{Base: base, Amount: 3}

	runCollectorCases(t, bridge, []collectorCase{
		{"SelectAsset", func() (any, error) { return collector.SelectAsset(assets, rules) }, asset, asset},
		{"SelectAssetToLose", func() (any, error) { return collector.SelectAssetToLose(alpha, assets, rules) }, asset, asset},
		{"SelectRelocateOrders", func() (any, error) { return collector.SelectRelocateOrders(asset, assets, []string{"Erebus"}, 1) }, relocateOrders, relocateOrders},
		{"SelectRepairOrders", func() (any, error) { return collector.SelectRepairOrders(alpha, assets, rules) }, repairOrders, repairOrders},
		{"SelectBuyOrder", func() (any, error) {
			return collector.SelectBuyOrder(map[string][]*domain.AssetDefinition{"Tartarus": {definition}})
		}, buyOrder, buyOrder},
		{"SelectRefitOrder", func() (any, error) {
			return collector.SelectRefitOrder([]action.RefitOption{{Asset: asset, Replacements: []*domain.AssetDefinition{definition}}}, rules)
		}, refitOrder, refitOrder},
		{"SelectAttackers", func() (any, error) { return collector.SelectAttackers(assets, rules) }, assets, assets},
		{"SelectDefender", func() (any, error) { return collector.SelectDefender(asset, assets, rules) }, asset, asset},
		{"ConfirmRedirectToBase", func() (any, error) { return collector.ConfirmRedirectToBase(beta, base, 2) }, true, true},
		{"ConfirmCounterstrike", func() (any, error) { return collector.ConfirmCounterstrike(beta, asset, asset) }, true, true},
		{"ConfirmSacrifice", func() (any, error) {
			return collector.ConfirmSacrifice(hooks.SacrificeOffer{Owner: alpha, Sacrifice: asset, Saves: asset})
		}, true, true},
		{"RequestPermission", func() (any, error) {
			return collector.RequestPermission(hooks.PermissionRequest{Requester: alpha, Ruler: beta, WorldID: "Tartarus", Def: definition})
		}, true, true},
		{"ConfirmPermissionContest", func() (any, error) {
			return collector.ConfirmPermissionContest(hooks.PermissionContestOffer{Owner: beta, Source: asset, Requester: alpha, WorldID: "Tartarus", Def: definition})
		}, true, true},
		{"ConfirmPayment", func() (any, error) {
			return collector.ConfirmPayment(hooks.PaymentOffer{Payer: alpha, Asset: asset, Amount: 1})
		}, true, true},
		{"SelectExpandInfluenceOrder", func() (any, error) {
			return collector.SelectExpandInfluenceOrder(alpha, factionState, []string{"Tartarus"})
		}, expandOrder, expandOrder},
		{"ConfirmRivalFreeAttack", func() (any, error) { return collector.ConfirmRivalFreeAttack(beta, 9, 7) }, true, true},
		{"SelectBaseAttackers", func() (any, error) { return collector.SelectBaseAttackers(beta, assets, rules) }, assets, assets},
		{"SelectAbilityHolders", func() (any, error) { return collector.SelectAbilityHolders(alpha, holders, rules) }, holders, holders},
		{"ConfirmAbilityApplied", func() (any, error) { return collector.ConfirmAbilityApplied(asset, definition) }, true, true},
		{"SelectBribeTarget", func() (any, error) {
			chosen, amount, err := collector.SelectBribeTarget(alpha, factionState)
			return adapter.BribeReply{Base: chosen, Amount: amount}, err
		}, bribeReply, bribeReply},
		{"SelectSeizeTarget", func() (any, error) { return collector.SelectSeizeTarget(alpha, factionState) }, "Tartarus", "Tartarus"},
		{"SelectChangeHomeworldTarget", func() (any, error) { return collector.SelectChangeHomeworldTarget(alpha, factionState) }, "Tartarus", "Tartarus"},
		{"SelectPurchaseEnablerAsset", func() (any, error) {
			return collector.SelectPurchaseEnablerAsset("Tartarus", []action.PurchaseEnablerOption{{Def: definition, Cost: 2}})
		}, definition, definition},
	})
}

func TestPhaseCollector_EveryReplyType(t *testing.T) {
	bridge, harness := newBridge(t)
	alpha := harness.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := harness.AddFaction("beta", "Tartarus", 2, 2, 2)
	if _, err := harness.Engine.World.RebuildIndex(harness.FactionState, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("RebuildIndex: %v", err)
	}
	asset := alpha.Assets["alpha-asset-1"]
	assets := []*domain.Asset{asset}
	rules := harness.Engine.Rulebook
	definition := rules.Assets[testharness.DefSecurityPersonnel]
	collector := bridge.Phase()

	available := harness.Engine.Action.AvailableActions(alpha, harness.FactionState, rules, bridge.Action(), slog.New(slog.DiscardHandler))
	if len(available) == 0 {
		t.Fatal("no action available for alpha")
	}
	chosenAction := available[0]
	force := domain.StatForce
	decisions := []world.MovementDecision{{AssetID: asset.ID}}
	relocateOrders := []action.RelocateOrder{{Asset: asset, DestinationWorld: "Erebus"}}

	runCollectorCases(t, bridge, []collectorCase{
		{"SelectAction", func() (any, error) { return collector.SelectAction(alpha, available) }, chosenAction, chosenAction},
		{"SelectStatRaise", func() (any, error) { return collector.SelectStatRaise(alpha, []domain.FactionStat{force}) }, &force, &force},
		{"SelectMovementDecisions", func() (any, error) { return collector.SelectMovementDecisions(alpha, assets) }, decisions, decisions},
		{"SelectTransportCargo", func() (any, error) { return collector.SelectTransportCargo(asset, assets, nil) }, assets, assets},
		{"SelectRelocateOrders", func() (any, error) { return collector.SelectRelocateOrders(asset, assets, []string{"Erebus"}, 1) }, relocateOrders, relocateOrders},
		{"ConfirmPayment", func() (any, error) {
			return collector.ConfirmPayment(hooks.PaymentOffer{Payer: alpha, Asset: asset, Amount: 1})
		}, true, true},
		{"RequestPermission", func() (any, error) {
			return collector.RequestPermission(hooks.PermissionRequest{Requester: alpha, Ruler: beta, WorldID: "Tartarus", Def: definition})
		}, true, true},
		{"ConfirmPermissionContest", func() (any, error) {
			return collector.ConfirmPermissionContest(hooks.PermissionContestOffer{Owner: beta, Source: asset, Requester: alpha, WorldID: "Tartarus", Def: definition})
		}, true, true},
	})
}

func TestSelectAbilityTarget_EachTargetKind(t *testing.T) {
	bridge, harness := newBridge(t)
	alpha := harness.AddFaction("alpha", "Tartarus", 4, 3, 2)
	asset := alpha.Assets["alpha-asset-1"]
	holder := domain.AbilityHolder{Asset: asset}
	collector := bridge.Action()
	target := func(kind domain.TargetKind) func() (any, error) {
		return func() (any, error) {
			return collector.SelectAbilityTarget(holder, domain.AbilityStep{Target: kind}, domain.AbilityTargetCandidates{})
		}
	}

	runCollectorCases(t, bridge, []collectorCase{
		{"faction", target(domain.TargetFaction), alpha, domain.AbilityTarget{Kind: domain.TargetFaction, Faction: alpha}},
		{"asset", target(domain.TargetAsset), asset, domain.AbilityTarget{Kind: domain.TargetAsset, Asset: asset}},
		{"world", target(domain.TargetWorld), "Tartarus", domain.AbilityTarget{Kind: domain.TargetWorld, World: "Tartarus"}},
	})
}

// checkpointAsks reports whether AwaitCheckpoint asked for phase. With no ask, its
// answering goroutine stays on the channel, so a test checks a silent phase last.
func checkpointAsks(t *testing.T, bridge *adapter.Adapter, phase string) bool {
	t.Helper()
	answered := make(chan struct{}, 1)
	go func() {
		ask, ok := bridge.AskPump()().(adapter.CollectorAskMsg)
		if !ok {
			return
		}
		answered <- struct{}{}
		ask.Reply <- true
	}()
	if err := bridge.Phase().AwaitCheckpoint(phase); err != nil {
		t.Errorf("AwaitCheckpoint(%q): %v", phase, err)
	}
	select {
	case <-answered:
		return true
	default:
		return false
	}
}

func TestAwaitCheckpoint_PerFactionCadenceAsksAtEveryPhase(t *testing.T) {
	bridge, _ := newBridge(t)
	bridge.SetPerFactionCadence(true)

	for _, phase := range []string{engine.CheckpointMovement, engine.CheckpointCycleSummary} {
		if !checkpointAsks(t, bridge, phase) {
			t.Errorf("%s: got no ask", phase)
		}
	}
}

func TestAwaitCheckpoint_OnlyTheCycleSummaryAsksOtherwise(t *testing.T) {
	bridge, _ := newBridge(t)

	if !checkpointAsks(t, bridge, engine.CheckpointCycleSummary) {
		t.Error("cycle summary: got no ask")
	}
	if checkpointAsks(t, bridge, engine.CheckpointMovement) {
		t.Error("movement: asked outside per-faction cadence")
	}
}

func TestCanceledReplyComesBackAsTheError(t *testing.T) {
	bridge, harness := newBridge(t)
	alpha := harness.AddFaction("alpha", "Tartarus", 4, 3, 2)

	answerNext(bridge, action.ErrTurnCanceled)
	if _, err := bridge.Action().SelectAsset(nil, harness.Engine.Rulebook); !errors.Is(err, action.ErrTurnCanceled) {
		t.Errorf("Action(): got %v, want ErrTurnCanceled", err)
	}

	answerNext(bridge, action.ErrTurnCanceled)
	if _, err := bridge.Phase().SelectAction(alpha, nil); !errors.Is(err, action.ErrTurnCanceled) {
		t.Errorf("Phase(): got %v, want ErrTurnCanceled", err)
	}
}
