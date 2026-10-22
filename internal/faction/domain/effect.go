package domain

// SpecialEffectType names an S-side effect. The effects engine keys its registrar
// table on it, as the ability dispatch keys on AbilityEffectType.
type SpecialEffectType string

const (
	SpecialCoinGain       SpecialEffectType = "coin_gain"
	SpecialCoinSteal      SpecialEffectType = "coin_steal"
	SpecialIgnoreCoinLoss SpecialEffectType = "ignore_coin_loss"
	SpecialSelfDamage     SpecialEffectType = "self_damage"
	SpecialApplyStealth   SpecialEffectType = "apply_stealth"
	SpecialClearStealth   SpecialEffectType = "clear_stealth"
	SpecialBonusDie       SpecialEffectType = "bonus_die"
	SpecialRemoveTagDice  SpecialEffectType = "remove_tag_dice"
	SpecialReroll         SpecialEffectType = "reroll"
	SpecialTechLevelFloor SpecialEffectType = "tech_level_floor"
	SpecialCostSurcharge  SpecialEffectType = "cost_surcharge"
	SpecialGrantMovement  SpecialEffectType = "grant_movement"
	SpecialRestrictAttack SpecialEffectType = "restrict_attack"
	SpecialCaptureTarget  SpecialEffectType = "capture_target"
	SpecialBlockAbility   SpecialEffectType = "block_ability"
	SpecialCounterstrike  SpecialEffectType = "counterstrike"
	SpecialSeizeTarget    SpecialEffectType = "seize_target"
	SpecialRestoreAsset   SpecialEffectType = "restore_asset"
	SpecialDiscountRepair SpecialEffectType = "discount_repair"
	SpecialSacrificeSelf  SpecialEffectType = "sacrifice_self"
	SpecialShelterAsset   SpecialEffectType = "shelter_asset"
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
	ScopeSelf    EffectScope = "self"
	ScopeWorld   EffectScope = "world"
	ScopeFaction EffectScope = "faction"
)

// EffectFilterName names an optional filter. owner and scope are not listed: every
// entry carries both, and a registrar accepts them by value.
type EffectFilterName string

const (
	FilterAttribute EffectFilterName = "attribute"
	FilterPhase     EffectFilterName = "phase"
	FilterAction    EffectFilterName = "action"
	FilterAssetType EffectFilterName = "asset_type"
	FilterCategory  EffectFilterName = "category"
	FilterStealthy  EffectFilterName = "stealthy"
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
// absent filter. Stealthy is a pointer because "stealthy = false" is a real
// filter, distinct from no filter.
type EffectFilters struct {
	Owner      EffectOwner
	Scope      EffectScope
	Attribute  FactionStat
	Phases     []EffectRollPhase
	Action     string
	AssetTypes []AssetType
	Categories []FactionStat
	Stealthy   *bool
}

func (filters EffectFilters) Used() []EffectFilterName {
	var present []EffectFilterName
	if filters.Attribute != "" {
		present = append(present, FilterAttribute)
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

type RerollParams struct {
	Dice     int
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

// EffectDefinition is one entry of an asset's [[effects]] list. At most one
// parameter pointer is set: the one named by Effect, when that effect takes
// parameters.
type EffectDefinition struct {
	Trigger        EffectTrigger
	Effect         SpecialEffectType
	Filters        EffectFilters
	Gates          EffectGates
	CoinGain       *CoinGainParams
	CoinSteal      *CoinStealParams
	SelfDamage     *SelfDamageParams
	BonusDie       *BonusDieParams
	Reroll         *RerollParams
	TechLevelFloor *TechLevelFloorParams
	CostSurcharge  *CostSurchargeParams
	RestrictAttack *RestrictAttackParams
	GrantMovement  *GrantMovementParams
	CaptureTarget  *CaptureTargetParams
	DiscountRepair *DiscountRepairParams
}
