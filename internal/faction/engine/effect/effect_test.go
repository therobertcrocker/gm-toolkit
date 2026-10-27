package effect_test

import (
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var discardLog = slog.New(slog.DiscardHandler)

var (
	ownSelf    = domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf}
	rivalWorld = domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld}

	acceptsOwnSelf    = effect.Accepts{Owners: []domain.EffectOwner{domain.OwnerOwn}, Scopes: []domain.EffectScope{domain.ScopeSelf}}
	acceptsOwnWorld   = effect.Accepts{Owners: []domain.EffectOwner{domain.OwnerOwn}, Scopes: []domain.EffectScope{domain.ScopeWorld}}
	acceptsRivalWorld = effect.Accepts{Owners: []domain.EffectOwner{domain.OwnerRival}, Scopes: []domain.EffectScope{domain.ScopeWorld}}
)

// TestEffectsEngine_RegistersTransportReactor verifies that ApplyAll registers
// exactly one MutationReactor at FactionScope for a faction whose assets include
// a transport-capable asset, and zero for a faction with no transport assets.
func TestEffectsEngine_RegistersTransportReactor(t *testing.T) {
	transportDefID := "test-transport"
	nonTransportDefID := "test-soldier"

	transportDef := &domain.AssetDefinition{
		ID:    transportDefID,
		Speed: 2,
		Transport: &domain.TransportProfile{
			MaxHex:     3,
			CoinCost:   1,
			CargoTypes: []domain.AssetType{domain.TypeSpecialForces},
			MaxCargo:   1,
		},
	}
	nonTransportDef := &domain.AssetDefinition{
		ID:   nonTransportDefID,
		Type: domain.TypeSpecialForces,
	}

	rb := &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			transportDefID:    transportDef,
			nonTransportDefID: nonTransportDef,
		},
	}

	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"alpha": {
				ID: "alpha",
				Assets: map[string]*domain.Asset{
					"transport-1": {ID: "transport-1", DefinitionID: transportDefID, OwnerID: "alpha"},
				},
			},
			"beta": {
				ID: "beta",
				Assets: map[string]*domain.Asset{
					"soldier-1": {ID: "soldier-1", DefinitionID: nonTransportDefID, OwnerID: "beta"},
				},
			},
		},
	}

	e := effect.New(nil, discardLog)
	if err := e.Register(effects.NewTransportHandler(transportDef)); err != nil {
		t.Fatalf("Register: %v", err)
	}

	registry := hooks.NewRegistry()
	e.ApplyAll(factionState, rb, registry, nil, discardLog)

	wantSource := "transport:" + transportDefID

	// The transport asset carries its own reactor, at asset scope.
	transportReactors := registry.MutationReactorsForAsset("alpha", "transport-1")
	if len(transportReactors) != 1 {
		t.Fatalf("transport-1: got %d MutationReactors, want 1", len(transportReactors))
	}
	if got := transportReactors[0].Source; got != wantSource {
		t.Errorf("transport-1 reactor source = %q, want %q", got, wantSource)
	}
	if got := registry.MutationReactorsForFaction("alpha"); len(got) != 0 {
		t.Errorf("alpha: got %d faction-scoped MutationReactors, want 0", len(got))
	}

	// Faction beta has no transport asset — no reactor should be registered.
	if got := registry.MutationReactorsForAsset("beta", "soldier-1"); len(got) != 0 {
		t.Errorf("soldier-1: got %d MutationReactors, want 0", len(got))
	}
}

func TestRegister_RejectsASecondHandlerForOneDefinition(t *testing.T) {
	def := &domain.AssetDefinition{ID: "T-001", Transport: &domain.TransportProfile{MaxHex: 1}}
	e := effect.New(nil, discardLog)
	if err := e.Register(effects.NewTransportHandler(def)); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	err := e.Register(effects.NewTransportHandler(def))
	if err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("second Register: got %v, want an already-registered error", err)
	}
}

// recordingHandler records the faction/asset pairs ApplyAll hands it.
type recordingHandler struct {
	defID string
	calls *[]string
}

func (handler recordingHandler) AssetDefinitionID() string { return handler.defID }

func (handler recordingHandler) Apply(faction *domain.Faction, asset *domain.Asset, _ *hooks.Registry, _ domain.Roller) {
	*handler.calls = append(*handler.calls, faction.ID+"/"+asset.ID)
}

func TestApplyAll_WalksFactionsAndAssetsInIDOrder(t *testing.T) {
	def := &domain.AssetDefinition{ID: "T-001"}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{}}
	for _, factionID := range []string{"delta", "alpha", "charlie", "bravo"} {
		faction := &domain.Faction{ID: factionID, Assets: map[string]*domain.Asset{}}
		for _, suffix := range []string{"3", "1", "2"} {
			assetID := factionID + "-" + suffix
			faction.Assets[assetID] = &domain.Asset{ID: assetID, DefinitionID: def.ID, OwnerID: factionID}
		}
		factionState.Factions[factionID] = faction
	}

	var calls []string
	e := effect.New(nil, discardLog)
	if err := e.Register(recordingHandler{defID: def.ID, calls: &calls}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	e.ApplyAll(factionState, rb, hooks.NewRegistry(), nil, discardLog)

	want := []string{
		"alpha/alpha-1", "alpha/alpha-2", "alpha/alpha-3",
		"bravo/bravo-1", "bravo/bravo-2", "bravo/bravo-3",
		"charlie/charlie-1", "charlie/charlie-2", "charlie/charlie-3",
		"delta/delta-1", "delta/delta-2", "delta/delta-3",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("ApplyAll order:\n got  %v\n want %v", calls, want)
	}
}

type constRoller struct{}

func (constRoller) Roll(int) int { return 1 }

func TestRegisterEffectsList_RegistersHookRouteEntriesOnly(t *testing.T) {
	var registered []domain.SpecialEffectType
	var roller domain.Roller = constRoller{}
	registrars := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialCoinGain: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerTurnStart: acceptsOwnSelf},
			Fire: func(effect.FireContext) []domain.Mutation {
				t.Error("ApplyAll must not fire a phase-route entry")
				return nil
			},
		},
		domain.SpecialCoinSteal: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOwnAttackHit: acceptsRivalWorld},
			Register: func(ctx effect.RegisterContext) {
				if ctx.Holder.Faction.ID != "alpha" || ctx.Holder.Asset.ID != "alpha-1" || ctx.Holder.ID() != "T-001" || ctx.Registry == nil || ctx.Roller != roller {
					t.Errorf("RegisterContext = %+v", ctx)
				}
				registered = append(registered, ctx.Entry.Effect)
			},
		},
		domain.SpecialBlockAbility: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOwnAttack: acceptsOwnSelf},
			Register: func(effect.RegisterContext) { t.Error("ApplyAll must not register a matchup-route entry") },
			Matchup:  noopMatchup,
		},
	}
	def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{
		{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain, Filters: ownSelf},
		{Trigger: domain.TriggerOwnAttackHit, Effect: domain.SpecialCoinSteal, Filters: rivalWorld},
		{Trigger: domain.TriggerOwnAttack, Effect: domain.SpecialBlockAbility, Filters: ownSelf},
	}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Assets: map[string]*domain.Asset{
			"alpha-1": {ID: "alpha-1", DefinitionID: def.ID, OwnerID: "alpha"},
		}},
	}}

	e := effect.New(registrars, discardLog)
	if err := e.RegisterEffectsList(def); err != nil {
		t.Fatalf("RegisterEffectsList: %v", err)
	}
	e.ApplyAll(factionState, rb, hooks.NewRegistry(), roller, discardLog)

	if want := []domain.SpecialEffectType{domain.SpecialCoinSteal}; !reflect.DeepEqual(registered, want) {
		t.Errorf("registered = %v, want %v", registered, want)
	}
}

func TestRegisterEffectsList_QualityNeedsTargets(t *testing.T) {
	entry := domain.EffectDefinition{Trigger: domain.TriggerOnPurchase, Effect: domain.SpecialApplyStealth, Filters: ownSelf}
	def := &domain.AssetDefinition{ID: "Q-001", Quality: true, Effects: []domain.EffectDefinition{entry}}

	withoutTargets := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialApplyStealth: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOnPurchase: acceptsOwnSelf},
			Fire:     noopFire,
		},
	}
	if err := effect.New(withoutTargets, discardLog).RegisterEffectsList(def); err == nil {
		t.Error("a quality whose effect has no Targets registered; want an error")
	}

	withTargets := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialApplyStealth: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOnPurchase: acceptsOwnSelf},
			Fire:     noopFire,
			Targets:  func(effect.TargetContext) []*domain.Asset { return nil },
		},
	}
	if err := effect.New(withTargets, discardLog).RegisterEffectsList(def); err != nil {
		t.Errorf("RegisterEffectsList: %v", err)
	}
}

func noopFire(effect.FireContext) []domain.Mutation { return nil }

func noopRegister(effect.RegisterContext) {}

func noopMatchup(effect.MatchupContext) effect.MatchupOutcome { return effect.MatchupOutcome{} }

func TestHolder_NamesItsSource(t *testing.T) {
	faction := &domain.Faction{ID: "alpha"}
	asset := effect.Holder{Faction: faction, Asset: &domain.Asset{ID: "a-1"}, Def: &domain.AssetDefinition{ID: "SWN-C7-002", Name: "Book of Secrets"}}
	tag := effect.Holder{Faction: faction, Tag: &domain.Tag{ID: "SWN-T-020", Name: "Warlike"}}

	if asset.Scope() != hooks.AssetScope("alpha", "a-1") || asset.HookID(domain.SpecialReroll) != "reroll:SWN-C7-002" || asset.Name() != "Book of Secrets" {
		t.Errorf("asset holder: scope %+v, hook %q, name %q", asset.Scope(), asset.HookID(domain.SpecialReroll), asset.Name())
	}
	if tag.Scope() != hooks.FactionScope("alpha") || tag.HookID(domain.SpecialBonusDie) != "bonus_die:SWN-T-020" || tag.Name() != "Warlike" {
		t.Errorf("tag holder: scope %+v, hook %q, name %q", tag.Scope(), tag.HookID(domain.SpecialBonusDie), tag.Name())
	}
}

func validationRegistrars() map[domain.SpecialEffectType]effect.Registrar {
	return map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialCoinGain: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerTurnStart: acceptsOwnSelf},
			Fire:     noopFire,
		},
		domain.SpecialCoinSteal: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{
				domain.TriggerOwnAttackHit: {
					Owners:  []domain.EffectOwner{domain.OwnerRival},
					Scopes:  []domain.EffectScope{domain.ScopeWorld},
					Filters: []domain.EffectFilterName{domain.FilterStealthy},
					Gates:   []domain.EffectGateName{domain.GateOncePer},
				},
			},
			Register: noopRegister,
		},
		// Book of Secrets' shape: each trigger accepts a different owner
		domain.SpecialReroll: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{
				domain.TriggerOwnRoll:       acceptsOwnWorld,
				domain.TriggerEnemyRollAtMe: acceptsRivalWorld,
			},
			Register: noopRegister,
		},
		// declares a hook trigger but has only a phase route
		domain.SpecialClearStealth: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerStealthArrival: acceptsRivalWorld},
			Fire:     noopFire,
		},
		// declares a phase trigger but has only a hook route
		domain.SpecialApplyStealth: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOnPurchase: acceptsOwnSelf},
			Register: noopRegister,
		},
		domain.SpecialBlockAbility: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOwnAttack: acceptsOwnSelf},
			Matchup:  noopMatchup,
		},
		// declares a matchup trigger but has only a hook route
		domain.SpecialCaptureTarget: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOwnDefenseWin: acceptsOwnSelf},
			Register: noopRegister,
		},
	}
}

func TestRegisterEffectsList_Validates(t *testing.T) {
	stealthy := true
	cases := []struct {
		name    string
		entry   domain.EffectDefinition
		wantErr string
	}{
		{
			name:  "accepted phase-route entry",
			entry: domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain, Filters: ownSelf},
		},
		{
			name: "accepted hook-route entry with a filter and a gate",
			entry: domain.EffectDefinition{
				Trigger: domain.TriggerOwnAttackHit,
				Effect:  domain.SpecialCoinSteal,
				Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld, Stealthy: &stealthy},
				Gates:   domain.EffectGates{OncePer: domain.OncePerTargetFaction},
			},
		},
		{
			name:  "accepted entry on a second trigger with its own owner",
			entry: domain.EffectDefinition{Trigger: domain.TriggerEnemyRollAtMe, Effect: domain.SpecialReroll, Filters: rivalWorld},
		},
		{
			name:    "no registrar",
			entry:   domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialBonusDie, Filters: ownSelf},
			wantErr: `effect "bonus_die" has no registrar`,
		},
		{
			name:    "trigger not accepted",
			entry:   domain.EffectDefinition{Trigger: domain.TriggerOwnAttackHit, Effect: domain.SpecialCoinGain, Filters: ownSelf},
			wantErr: `does not accept trigger "own_attack_hit"`,
		},
		{
			name: "owner not accepted",
			entry: domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain,
				Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeSelf}},
			wantErr: `does not accept owner "rival"`,
		},
		{
			name:    "owner accepted only under another trigger",
			entry:   domain.EffectDefinition{Trigger: domain.TriggerOwnRoll, Effect: domain.SpecialReroll, Filters: rivalWorld},
			wantErr: `does not accept owner "rival" for trigger "own_roll"`,
		},
		{
			name: "scope not accepted",
			entry: domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain,
				Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld}},
			wantErr: `does not accept scope "world"`,
		},
		{
			name: "filter not accepted",
			entry: domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain,
				Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf, Attribute: domain.StatCunning}},
			wantErr: `does not accept filter "attribute"`,
		},
		{
			name: "gate not accepted",
			entry: domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain,
				Filters: ownSelf, Gates: domain.EffectGates{NoStack: true}},
			wantErr: `does not accept gate "no_stack"`,
		},
		{
			name:    "hook trigger without a hook route",
			entry:   domain.EffectDefinition{Trigger: domain.TriggerStealthArrival, Effect: domain.SpecialClearStealth, Filters: rivalWorld},
			wantErr: "has no hook route",
		},
		{
			name:    "phase trigger without a phase route",
			entry:   domain.EffectDefinition{Trigger: domain.TriggerOnPurchase, Effect: domain.SpecialApplyStealth, Filters: ownSelf},
			wantErr: "has no phase route",
		},
		{
			name:  "accepted matchup-route entry",
			entry: domain.EffectDefinition{Trigger: domain.TriggerOwnAttack, Effect: domain.SpecialBlockAbility, Filters: ownSelf},
		},
		{
			name:    "matchup trigger without a matchup route",
			entry:   domain.EffectDefinition{Trigger: domain.TriggerOwnDefenseWin, Effect: domain.SpecialCaptureTarget, Filters: ownSelf},
			wantErr: "has no matchup route",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{tc.entry}}
			err := effect.New(validationRegistrars(), discardLog).RegisterEffectsList(def)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("RegisterEffectsList: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("RegisterEffectsList: got %v, want an error containing %q", err, tc.wantErr)
			}
			if !strings.Contains(err.Error(), `asset "T-001" effects[0]`) {
				t.Errorf("error %q does not name the entry", err.Error())
			}
		})
	}
}

func TestCheckTagEffects_ScopeDecidesTheHolder(t *testing.T) {
	acceptsOwnAny := effect.Accepts{
		Owners: []domain.EffectOwner{domain.OwnerOwn},
		Scopes: []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeFaction},
		Gates:  []domain.EffectGateName{domain.GateOncePer},
	}
	registrars := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialBonusDie: {Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOwnRoll: acceptsOwnAny}, Register: noopRegister},
	}
	entry := func(scope domain.EffectScope, oncePer domain.OncePerKey) domain.EffectDefinition {
		return domain.EffectDefinition{Trigger: domain.TriggerOwnRoll, Effect: domain.SpecialBonusDie,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: scope}, Gates: domain.EffectGates{OncePer: oncePer}}
	}
	cases := []struct {
		name    string
		entry   domain.EffectDefinition
		wantErr string
	}{
		{"faction scope", entry(domain.ScopeFaction, domain.OncePerOwnerFaction), ""},
		{"self scope", entry(domain.ScopeSelf, ""), `scope "self" cannot apply`},
		{"world scope", entry(domain.ScopeWorld, ""), `scope "world" cannot apply`},
		{"once per asset", entry(domain.ScopeFaction, domain.OncePerAsset), `once_per "asset" cannot apply`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tag := &domain.Tag{ID: "SWN-T-020", Effects: []domain.EffectDefinition{tc.entry}}
			err := effect.New(registrars, discardLog).CheckTagEffects(tag)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("CheckTagEffects: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) || !strings.Contains(err.Error(), `tag "SWN-T-020" effects[0]`) {
				t.Fatalf("CheckTagEffects: got %v, want an error containing %q", err, tc.wantErr)
			}
		})
	}

	t.Run("an asset may use self scope", func(t *testing.T) {
		def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{entry(domain.ScopeSelf, domain.OncePerAsset)}}
		if err := effect.New(registrars, discardLog).RegisterEffectsList(def); err != nil {
			t.Errorf("RegisterEffectsList: %v", err)
		}
	})
}

func TestCheckTagEffects_AssetBoundRegistrarsRefuseATag(t *testing.T) {
	engine := effect.New(effects.Registrars(), discardLog)
	for name, entry := range map[string]domain.EffectDefinition{
		"grant_movement": {Trigger: domain.TriggerMovementOptionsQuery, Effect: domain.SpecialGrantMovement,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction}, GrantMovement: &domain.GrantMovementParams{}},
		"ignore_coin_loss": {Trigger: domain.TriggerCoinLoss, Effect: domain.SpecialIgnoreCoinLoss,
			Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeFaction}},
	} {
		t.Run(name, func(t *testing.T) {
			err := engine.CheckTagEffects(&domain.Tag{ID: "SWN-T-008", Effects: []domain.EffectDefinition{entry}})
			if err == nil || !strings.Contains(err.Error(), "a tag cannot hold this effect yet") {
				t.Errorf("CheckTagEffects: got %v, want the tag refusal", err)
			}
		})
	}
}

func TestFireTurnStart_FiresActiveTurnStartEntriesInAssetOrder(t *testing.T) {
	registrars := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialCoinGain: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerTurnStart: acceptsOwnSelf},
			Fire: func(ctx effect.FireContext) []domain.Mutation {
				return []domain.Mutation{domain.CoinDelta{FactionID: ctx.Holder.Faction.ID, Delta: 1, Cause: ctx.Asset.ID}}
			},
		},
	}
	def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{
		{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain, Filters: ownSelf},
		{Trigger: domain.TriggerOwnAttackHit, Effect: domain.SpecialCoinSteal, Filters: rivalWorld},
	}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	faction := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{
		"a-2":            {ID: "a-2", DefinitionID: def.ID, Ready: true, Maintained: true, CurrentHP: 1},
		"a-1":            {ID: "a-1", DefinitionID: def.ID, Ready: true, Maintained: true, CurrentHP: 1},
		"a-not-ready":    {ID: "a-not-ready", DefinitionID: def.ID, Ready: false, Maintained: true, CurrentHP: 1},
		"a-unmaintained": {ID: "a-unmaintained", DefinitionID: def.ID, Ready: true, Maintained: false, CurrentHP: 1},
		"a-destroyed":    {ID: "a-destroyed", DefinitionID: def.ID, Ready: true, Maintained: true, CurrentHP: 0},
	}}

	mutations := effect.New(registrars, discardLog).FireTurnStart(faction, nil, rb, nil)

	var causes []string
	for _, mutation := range mutations {
		causes = append(causes, mutation.(domain.CoinDelta).Cause)
	}
	if want := []string{"a-1", "a-2"}; !reflect.DeepEqual(causes, want) {
		t.Errorf("fired for %v, want %v", causes, want)
	}
}

func TestFirePurchase_FiresOnPurchaseEntriesOnly(t *testing.T) {
	coinPerTrigger := func(ctx effect.FireContext) []domain.Mutation {
		return []domain.Mutation{domain.CoinDelta{FactionID: ctx.Holder.Faction.ID, Delta: 1, Cause: string(ctx.Entry.Trigger)}}
	}
	registrars := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialApplyStealth: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOnPurchase: acceptsOwnSelf},
			Fire:     coinPerTrigger,
		},
		domain.SpecialCoinGain: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerTurnStart: acceptsOwnSelf},
			Fire:     coinPerTrigger,
		},
	}
	def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{
		{Trigger: domain.TriggerOnPurchase, Effect: domain.SpecialApplyStealth, Filters: ownSelf},
		{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain, Filters: ownSelf},
	}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	faction := &domain.Faction{ID: "alpha"}
	engine := effect.New(registrars, discardLog)

	t.Run("only the purchase entry fires", func(t *testing.T) {
		bought := &domain.Asset{ID: "a-1", DefinitionID: def.ID, Ready: false, CurrentHP: 1}
		got := engine.FirePurchase(faction, bought, nil, rb, nil)
		want := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 1, Cause: "on_purchase"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("mutations = %+v, want %+v", got, want)
		}
	})

	t.Run("an unknown definition fires nothing", func(t *testing.T) {
		unknown := &domain.Asset{ID: "a-2", DefinitionID: "T-404"}
		if got := engine.FirePurchase(faction, unknown, nil, rb, nil); got != nil {
			t.Errorf("mutations = %+v, want nil", got)
		}
	})
}

func TestFireTurnStart_PassesFactionStateToThePhaseRoute(t *testing.T) {
	var seen *state.FactionState
	registrars := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialCoinGain: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerTurnStart: acceptsOwnSelf},
			Fire: func(ctx effect.FireContext) []domain.Mutation {
				seen = ctx.FactionState
				return nil
			},
		},
	}
	def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{
		{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain, Filters: ownSelf},
	}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	faction := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{
		"a-1": {ID: "a-1", DefinitionID: def.ID, Ready: true, Maintained: true, CurrentHP: 1},
	}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": faction}}

	effect.New(registrars, discardLog).FireTurnStart(faction, factionState, rb, nil)

	if seen != factionState {
		t.Errorf("FireContext.FactionState = %p, want the state passed in (%p)", seen, factionState)
	}
}

func TestFireMatchup_FiresEntriesOfTheTriggerAndOrsCounterstrike(t *testing.T) {
	coinPerEffect := func(counterstrike bool) func(effect.MatchupContext) effect.MatchupOutcome {
		return func(ctx effect.MatchupContext) effect.MatchupOutcome {
			return effect.MatchupOutcome{
				Mutations:     []domain.Mutation{domain.CoinDelta{FactionID: ctx.Holder.Faction.ID, Delta: 1, Cause: string(ctx.Entry.Effect)}},
				Counterstrike: counterstrike,
			}
		}
	}
	registrars := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialBlockAbility: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOwnAttack: acceptsOwnSelf},
			Matchup:  coinPerEffect(false),
		},
		domain.SpecialCaptureTarget: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOwnDefenseWin: acceptsOwnSelf},
			Matchup:  coinPerEffect(true),
		},
		domain.SpecialReroll: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOwnDefenseWin: acceptsOwnSelf},
			Matchup:  coinPerEffect(false),
		},
	}
	def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{
		{Trigger: domain.TriggerOwnAttack, Effect: domain.SpecialBlockAbility, Filters: ownSelf},
		{Trigger: domain.TriggerOwnDefenseWin, Effect: domain.SpecialCaptureTarget, Filters: ownSelf},
		{Trigger: domain.TriggerOwnDefenseWin, Effect: domain.SpecialReroll, Filters: ownSelf},
	}}
	engine := effect.New(registrars, discardLog)
	ctx := effect.MatchupContext{Holder: effect.Holder{Faction: &domain.Faction{ID: "alpha"}, Def: def}}

	t.Run("own_attack fires its one entry", func(t *testing.T) {
		got := engine.FireMatchup(domain.TriggerOwnAttack, ctx, &rulebook.Rulebook{})
		want := effect.MatchupOutcome{Mutations: []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 1, Cause: "block_ability"}}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("outcome = %+v, want %+v", got, want)
		}
	})

	t.Run("own_defense_win fires both entries and ors Counterstrike", func(t *testing.T) {
		got := engine.FireMatchup(domain.TriggerOwnDefenseWin, ctx, &rulebook.Rulebook{})
		want := effect.MatchupOutcome{
			Mutations: []domain.Mutation{
				domain.CoinDelta{FactionID: "alpha", Delta: 1, Cause: "capture_target"},
				domain.CoinDelta{FactionID: "alpha", Delta: 1, Cause: "reroll"},
			},
			Counterstrike: true,
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("outcome = %+v, want %+v", got, want)
		}
	})
}

var acceptsOwnFaction = effect.Accepts{Owners: []domain.EffectOwner{domain.OwnerOwn}, Scopes: []domain.EffectScope{domain.ScopeFaction}}

// tagRoutes records the holder of every entry each route runs.
func tagRoutes(seen *[]string) map[domain.SpecialEffectType]effect.Registrar {
	name := func(holder effect.Holder) string {
		if holder.Tag != nil {
			return "tag:" + holder.Tag.ID
		}
		return "asset:" + holder.Asset.ID
	}
	return map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialBonusDie: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOwnRoll: acceptsOwnFaction},
			Register: func(ctx effect.RegisterContext) { *seen = append(*seen, "register "+name(ctx.Holder)) },
		},
		domain.SpecialCoinGain: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerTurnStart: acceptsOwnFaction},
			Fire: func(ctx effect.FireContext) []domain.Mutation {
				*seen = append(*seen, "turn_start "+name(ctx.Holder))
				return nil
			},
		},
		domain.SpecialApplyStealth: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOnPurchase: acceptsOwnFaction},
			Fire: func(ctx effect.FireContext) []domain.Mutation {
				*seen = append(*seen, "on_purchase "+name(ctx.Holder)+" bought "+ctx.Asset.ID)
				return nil
			},
		},
		domain.SpecialBlockAbility: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOwnAttack: acceptsOwnFaction},
			Matchup: func(ctx effect.MatchupContext) effect.MatchupOutcome {
				*seen = append(*seen, "own_attack "+name(ctx.Holder)+" fighting "+ctx.Asset.ID)
				return effect.MatchupOutcome{}
			},
		},
	}
}

func TestTagEntries_RunOnEveryRouteAfterTheAssets(t *testing.T) {
	entries := []domain.EffectDefinition{
		{Trigger: domain.TriggerOwnRoll, Effect: domain.SpecialBonusDie, Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction}},
		{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain, Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction}},
		{Trigger: domain.TriggerOnPurchase, Effect: domain.SpecialApplyStealth, Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction}},
		{Trigger: domain.TriggerOwnAttack, Effect: domain.SpecialBlockAbility, Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction}},
	}
	def := &domain.AssetDefinition{ID: "T-001", Effects: entries}
	tag := &domain.Tag{ID: "SWN-T-099", Effects: entries}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}, Tags: map[string]*domain.Tag{tag.ID: tag}}
	asset := &domain.Asset{ID: "a-1", DefinitionID: def.ID, Ready: true, Maintained: true, CurrentHP: 1}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-404", tag.ID}, Assets: map[string]*domain.Asset{"a-1": asset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha}}

	var seen []string
	engine := effect.New(tagRoutes(&seen), discardLog)
	if err := engine.RegisterEffectsList(def); err != nil {
		t.Fatalf("RegisterEffectsList: %v", err)
	}
	if err := engine.CheckTagEffects(tag); err != nil {
		t.Fatalf("CheckTagEffects: %v", err)
	}

	engine.ApplyAll(factionState, rb, hooks.NewRegistry(), nil, discardLog)
	engine.FireTurnStart(alpha, factionState, rb, nil)
	engine.FirePurchase(alpha, &domain.Asset{ID: "a-2", DefinitionID: "unknown"}, factionState, rb, nil)
	engine.FireMatchup(domain.TriggerOwnAttack, effect.MatchupContext{Holder: effect.Holder{Faction: alpha, Asset: asset, Def: def}, Asset: asset}, rb)

	want := []string{
		"register asset:a-1", "register tag:SWN-T-099",
		"turn_start asset:a-1", "turn_start tag:SWN-T-099",
		"on_purchase tag:SWN-T-099 bought a-2",
		"own_attack asset:a-1 fighting a-1", "own_attack tag:SWN-T-099 fighting a-1",
	}
	if !reflect.DeepEqual(seen, want) {
		t.Errorf("routes:\n got  %v\n want %v", seen, want)
	}
}
