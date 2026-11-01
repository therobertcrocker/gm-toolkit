package domain

// SpecialEffectType names an S-side effect. The effects engine keys its registrar
// table on it, as the ability dispatch keys on AbilityEffectType.
type SpecialEffectType string

const (
	SpecialCoinGain           SpecialEffectType = "coin_gain"
	SpecialCoinSteal          SpecialEffectType = "coin_steal"
	SpecialIgnoreCoinLoss     SpecialEffectType = "ignore_coin_loss"
	SpecialSelfDamage         SpecialEffectType = "self_damage"
	SpecialApplyStealth       SpecialEffectType = "apply_stealth"
	SpecialClearStealth       SpecialEffectType = "clear_stealth"
	SpecialBonusDie           SpecialEffectType = "bonus_die"
	SpecialRemoveTagDice      SpecialEffectType = "remove_tag_dice"
	SpecialReroll             SpecialEffectType = "reroll"
	SpecialTechLevelFloor     SpecialEffectType = "tech_level_floor"
	SpecialCostSurcharge      SpecialEffectType = "cost_surcharge"
	SpecialGrantMovement      SpecialEffectType = "grant_movement"
	SpecialRestrictAttack     SpecialEffectType = "restrict_attack"
	SpecialCaptureTarget      SpecialEffectType = "capture_target"
	SpecialBlockAbility       SpecialEffectType = "block_ability"
	SpecialCounterstrike      SpecialEffectType = "counterstrike"
	SpecialSeizeTarget        SpecialEffectType = "seize_target"
	SpecialRestoreAsset       SpecialEffectType = "restore_asset"
	SpecialDiscountRepair     SpecialEffectType = "discount_repair"
	SpecialSacrificeSelf      SpecialEffectType = "sacrifice_self"
	SpecialShelterAsset       SpecialEffectType = "shelter_asset"
	SpecialGrantStealth       SpecialEffectType = "grant_stealth"
	SpecialGrantPermission    SpecialEffectType = "grant_permission"
	SpecialWithdrawPermission SpecialEffectType = "withdraw_permission"
	SpecialLockOut            SpecialEffectType = "lock_out"
	SpecialLoseTies           SpecialEffectType = "lose_ties"
	SpecialCostDiscount       SpecialEffectType = "cost_discount"
	SpecialSpeedFloor         SpecialEffectType = "speed_floor"
	SpecialRemoveTag          SpecialEffectType = "remove_tag"
	SpecialXPRoll             SpecialEffectType = "xp_roll"
	SpecialPopulationFloor    SpecialEffectType = "population_floor"
	SpecialArrivalToll        SpecialEffectType = "arrival_toll"
)

type EffectTrigger string

const (
	TriggerTurnStart              EffectTrigger = "turn_start"
	TriggerOnPurchase             EffectTrigger = "on_purchase"
	TriggerOwnAttackHit           EffectTrigger = "own_attack_hit"
	TriggerOwnCounterHit          EffectTrigger = "own_counter_hit"
	TriggerOwnRoll                EffectTrigger = "own_roll"
	TriggerEnemyRollAtMe          EffectTrigger = "enemy_roll_at_me"
	TriggerStealthArrival         EffectTrigger = "stealth_arrival"
	TriggerCoinLoss               EffectTrigger = "coin_loss"
	TriggerTechLevelQuery         EffectTrigger = "tech_level_query"
	TriggerAssetCostQuery         EffectTrigger = "asset_cost_query"
	TriggerMovementOptionsQuery   EffectTrigger = "movement_options_query"
	TriggerAttackEligibilityQuery EffectTrigger = "attack_eligibility_query"
	TriggerOwnAttack              EffectTrigger = "own_attack"
	TriggerOwnDefenseWin          EffectTrigger = "own_defense_win"
	TriggerRepairCostQuery        EffectTrigger = "repair_cost_query"
	TriggerLethalDamage           EffectTrigger = "lethal_damage"
	TriggerPermissionQuery        EffectTrigger = "permission_query"
	TriggerPermissionGranted      EffectTrigger = "permission_granted"
	TriggerAttackTie              EffectTrigger = "attack_tie"
	TriggerAssetDestroyed         EffectTrigger = "asset_destroyed"
	TriggerSpeedQuery             EffectTrigger = "speed_query"
	TriggerRivalRoll              EffectTrigger = "rival_roll"
	TriggerHomeworldChanged       EffectTrigger = "homeworld_changed"
	TriggerGoalCompleted          EffectTrigger = "goal_completed"
	TriggerPopulationFloorQuery   EffectTrigger = "population_floor_query"
	TriggerArrivalQuery           EffectTrigger = "arrival_query"
)

// EffectOwner names whose side of the trigger's event an entry matches. The
// reading differs per trigger; each registrar documents its own.
type EffectOwner string

const (
	OwnerOwn   EffectOwner = "own"
	OwnerRival EffectOwner = "rival"
	OwnerAny   EffectOwner = "any"
)

type EffectScope string

const (
	ScopeSelf      EffectScope = "self"
	ScopeWorld     EffectScope = "world"
	ScopeHomeworld EffectScope = "homeworld"
	ScopeFaction   EffectScope = "faction"
	// ScopeBase is every world where the owner has a Base of Influence (Effort 4
	// Decision 6).
	ScopeBase EffectScope = "base"
)

// EffectFilterName names an optional filter. owner and scope are not listed: every
// entry carries both, and a registrar accepts them by value.
type EffectFilterName string

const (
	FilterAttribute         EffectFilterName = "attribute"
	FilterOpponentAttribute EffectFilterName = "opponent_attribute"
	FilterPhase             EffectFilterName = "phase"
	FilterAction            EffectFilterName = "action"
	FilterAssetType         EffectFilterName = "asset_type"
	FilterCategory          EffectFilterName = "category"
	FilterStealthy          EffectFilterName = "stealthy"
	FilterMinTechLevel      EffectFilterName = "min_tech_level"
	FilterCausedBy          EffectFilterName = "caused_by"
	FilterTechLevel         EffectFilterName = "tech_level"
	FilterOpponentTechLevel EffectFilterName = "opponent_tech_level"
	FilterGoal              EffectFilterName = "goal"
	FilterDetection         EffectFilterName = "detection"
	FilterAssetDefinition   EffectFilterName = "asset_definition"
)

type EffectGateName string

const (
	GateOncePer EffectGateName = "once_per"
	GateNoStack EffectGateName = "no_stack"
	GateContest EffectGateName = "contest"
)

// EffectRollPhase names a hooks.RollPhase; domain cannot import hooks.
type EffectRollPhase string

const (
	RollPhaseAttack      EffectRollPhase = "attack"
	RollPhaseDefense     EffectRollPhase = "defense"
	RollPhaseFactionTest EffectRollPhase = "faction_test"
	RollPhaseContested   EffectRollPhase = "contested"
)

type OncePerKey string

const (
	OncePerOwnerFaction  OncePerKey = "owner_faction"
	OncePerTargetFaction OncePerKey = "target_faction"
	OncePerAsset         OncePerKey = "asset"
)

// ContestRoller picks which side rolls as the contest's attacker: the declaring
// asset (Tripwire Cells), or the affected asset's faction testing to keep what it
// has (Panopticon Matrix).
type ContestRoller string

const (
	ContestRollerSelf   ContestRoller = "self"
	ContestRollerTarget ContestRoller = "target"
)

// ContestRoll names the roll phases a contest rolls with, as the book words it: an
// attack and its defense (Tripwire Cells', Informers' and Seductress' "Cunning vs.
// Cunning attack"), or a faction test and its contested roll (Panopticon Matrix's,
// Lobbyists' and Marketers' "test"). An attack contest's tie asks the tie
// resolvers; a test's never does.
type ContestRoll string

const (
	ContestRollAttack ContestRoll = "attack"
	ContestRollTest   ContestRoll = "test"
)

type RestrictSide string

const (
	RestrictDefense          RestrictSide = "defense"
	RestrictAttackAndCounter RestrictSide = "attack_and_counter"
)

type RestrictAllow string

const (
	AllowOnlyListed   RestrictAllow = "only_listed"
	AllowExceptListed RestrictAllow = "except_listed"
)

// EffectFilters narrows which events fire an entry. A zero optional field is an
// absent filter. Stealthy, TechLevel and OpponentTechLevel are pointers because
// "stealthy = false" and "tech_level = 0" are real filters, distinct from no
// filter. Detection is a plain bool: only "detection = true" filters.
type EffectFilters struct {
	Owner             EffectOwner
	Scope             EffectScope
	Attribute         FactionStat
	OpponentAttribute FactionStat
	Phases            []EffectRollPhase
	Action            string
	AssetTypes        []AssetType
	Categories        []FactionStat
	Stealthy          *bool
	MinTechLevel      int
	TechLevel         *int
	OpponentTechLevel *int
	CausedBy          EffectOwner
	Goal              string
	Detection         bool
	// AssetDefinition names the rolling asset's definition ID (the Eugenics Cult
	// die on Gengineered Slaves).
	AssetDefinition string
}

func (filters EffectFilters) Used() []EffectFilterName {
	var present []EffectFilterName
	if filters.Attribute != "" {
		present = append(present, FilterAttribute)
	}
	if filters.OpponentAttribute != "" {
		present = append(present, FilterOpponentAttribute)
	}
	if len(filters.Phases) > 0 {
		present = append(present, FilterPhase)
	}
	if filters.Action != "" {
		present = append(present, FilterAction)
	}
	if len(filters.AssetTypes) > 0 {
		present = append(present, FilterAssetType)
	}
	if len(filters.Categories) > 0 {
		present = append(present, FilterCategory)
	}
	if filters.Stealthy != nil {
		present = append(present, FilterStealthy)
	}
	if filters.MinTechLevel > 0 {
		present = append(present, FilterMinTechLevel)
	}
	if filters.TechLevel != nil {
		present = append(present, FilterTechLevel)
	}
	if filters.OpponentTechLevel != nil {
		present = append(present, FilterOpponentTechLevel)
	}
	if filters.CausedBy != "" {
		present = append(present, FilterCausedBy)
	}
	if filters.Goal != "" {
		present = append(present, FilterGoal)
	}
	if filters.Detection {
		present = append(present, FilterDetection)
	}
	if filters.AssetDefinition != "" {
		present = append(present, FilterAssetDefinition)
	}
	return present
}

type EffectGates struct {
	OncePer OncePerKey
	NoStack bool
	Contest *EffectContest
}

func (gates EffectGates) Used() []EffectGateName {
	var present []EffectGateName
	if gates.OncePer != "" {
		present = append(present, GateOncePer)
	}
	if gates.NoStack {
		present = append(present, GateNoStack)
	}
	if gates.Contest != nil {
		present = append(present, GateContest)
	}
	return present
}

type EffectContest struct {
	Contest
	Roller ContestRoller
}

type CoinGainParams struct {
	Amount DiceRoll
}

type CoinStealParams struct {
	Amount DiceRoll
}

type SelfDamageParams struct {
	Amount DiceRoll
}

type BonusDieParams struct {
	Sides int
	Count int
}

// RerollParams picks the dice a reroll targets: the Dice lowest (a forced
// reroll's highest), or every die showing Result. Exactly one is set.
type RerollParams struct {
	Dice     int
	Result   int
	Elective bool
}

type TechLevelFloorParams struct {
	Level int
}

type CostSurchargeParams struct {
	Amount int
}

type RestrictAttackParams struct {
	Side  RestrictSide
	Allow RestrictAllow
}

type GrantMovementParams struct {
	Relocate RelocateParams
}

type CaptureTargetParams struct {
	Coin DiceRoll
}

type DiscountRepairParams struct {
	Amount int
}

type CostDiscountParams struct {
	Amount int
}

type SpeedFloorParams struct {
	Speed int
}

// ArrivalTollParams is the Coin a landing rival asset pays the holder.
type ArrivalTollParams struct {
	Amount int
}

// PopulationFloorParams is a holder's floor as a percent of the setting's.
type PopulationFloorParams struct {
	Percent int
}

// XPRollParams is a bonus-XP roll: roll Die, and on Target or more gain Amount XP.
type XPRollParams struct {
	Die    DiceRoll
	Target int
	Amount int
}

// EffectDefinition is one entry of an asset's [[effects]] list. At most one
// parameter pointer is set: the one named by Effect, when that effect takes
// parameters.
type EffectDefinition struct {
	Trigger         EffectTrigger
	Effect          SpecialEffectType
	Filters         EffectFilters
	Gates           EffectGates
	CoinGain        *CoinGainParams
	CoinSteal       *CoinStealParams
	SelfDamage      *SelfDamageParams
	BonusDie        *BonusDieParams
	Reroll          *RerollParams
	TechLevelFloor  *TechLevelFloorParams
	CostSurcharge   *CostSurchargeParams
	RestrictAttack  *RestrictAttackParams
	GrantMovement   *GrantMovementParams
	CaptureTarget   *CaptureTargetParams
	DiscountRepair  *DiscountRepairParams
	CostDiscount    *CostDiscountParams
	SpeedFloor      *SpeedFloorParams
	XPRoll          *XPRollParams
	PopulationFloor *PopulationFloorParams
	ArrivalToll     *ArrivalTollParams
}
