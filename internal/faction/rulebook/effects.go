package rulebook

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

type effectRecord struct {
	Trigger        string                `toml:"trigger"`
	Effect         string                `toml:"effect"`
	Filters        *effectFiltersRecord  `toml:"filters"`
	Gates          *effectGatesRecord    `toml:"gates"`
	CoinGain       *amountRecord         `toml:"coin_gain"`
	CoinSteal      *amountRecord         `toml:"coin_steal"`
	SelfDamage     *amountRecord         `toml:"self_damage"`
	BonusDie       *bonusDieRecord       `toml:"bonus_die"`
	Reroll         *rerollRecord         `toml:"reroll"`
	TechLevelFloor *techLevelFloorRecord `toml:"tech_level_floor"`
	CostSurcharge  *costSurchargeRecord  `toml:"cost_surcharge"`
	RestrictAttack *restrictAttackRecord `toml:"restrict_attack"`
	GrantMovement  *grantMovementRecord  `toml:"grant_movement"`
	CaptureTarget  *amountRecord         `toml:"capture_target"`
	DiscountRepair *discountRepairRecord `toml:"discount_repair"`
}

type effectFiltersRecord struct {
	Owner     string   `toml:"owner"`
	Scope     string   `toml:"scope"`
	Attribute string   `toml:"attribute"`
	Phase     []string `toml:"phase"`
	Action    string   `toml:"action"`
	AssetType []string `toml:"asset_type"`
	Category  []string `toml:"category"`
	Stealthy  *bool    `toml:"stealthy"`
}

type effectGatesRecord struct {
	OncePer string               `toml:"once_per"`
	NoStack bool                 `toml:"no_stack"`
	Contest *effectContestRecord `toml:"contest"`
}

type effectContestRecord struct {
	Roller       string `toml:"roller"`
	AttackerStat string `toml:"attacker_stat"`
	DefenderStat string `toml:"defender_stat"`
}

type amountRecord struct {
	Amount string `toml:"amount"`
}

type bonusDieRecord struct {
	Sides int `toml:"sides"`
	Count int `toml:"count"`
}

type rerollRecord struct {
	Dice     int  `toml:"dice"`
	Elective bool `toml:"elective"`
}

type techLevelFloorRecord struct {
	Level int `toml:"level"`
}

type costSurchargeRecord struct {
	Amount int `toml:"amount"`
}

type restrictAttackRecord struct {
	Side  string `toml:"side"`
	Allow string `toml:"allow"`
}

type grantMovementRecord struct {
	Relocate *relocateRecord `toml:"relocate"`
}

type discountRepairRecord struct {
	Amount int `toml:"amount"`
}

var (
	effectTriggers = []domain.EffectTrigger{
		domain.TriggerTurnStart, domain.TriggerOnPurchase, domain.TriggerOwnAttackHit,
		domain.TriggerOwnCounterHit, domain.TriggerOwnRoll, domain.TriggerEnemyRollAtMe,
		domain.TriggerStealthArrival, domain.TriggerCoinLoss, domain.TriggerTechLevelQuery,
		domain.TriggerAssetCostQuery, domain.TriggerMovementOptionsQuery,
		domain.TriggerAttackEligibilityQuery, domain.TriggerOwnAttack, domain.TriggerOwnDefenseWin, domain.TriggerLethalDamage, domain.TriggerRepairCostQuery,
		domain.TriggerPermissionQuery,
	}
	specialEffects = []domain.SpecialEffectType{
		domain.SpecialCoinGain, domain.SpecialCoinSteal, domain.SpecialIgnoreCoinLoss,
		domain.SpecialSelfDamage, domain.SpecialApplyStealth, domain.SpecialClearStealth,
		domain.SpecialBonusDie, domain.SpecialRemoveTagDice, domain.SpecialReroll,
		domain.SpecialTechLevelFloor, domain.SpecialCostSurcharge, domain.SpecialGrantMovement,
		domain.SpecialRestrictAttack, domain.SpecialCaptureTarget, domain.SpecialBlockAbility,
		domain.SpecialCounterstrike, domain.SpecialShelterAsset, domain.SpecialSacrificeSelf,
		domain.SpecialRestoreAsset, domain.SpecialDiscountRepair, domain.SpecialSeizeTarget,
		domain.SpecialGrantStealth, domain.SpecialGrantPermission,
	}
	effectOwners     = []domain.EffectOwner{domain.OwnerOwn, domain.OwnerRival, domain.OwnerAny}
	effectScopes     = []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeFaction}
	effectRollPhases = []domain.EffectRollPhase{
		domain.RollPhaseAttack, domain.RollPhaseDefense, domain.RollPhaseFactionTest, domain.RollPhaseContested,
	}
	oncePerKeys    = []domain.OncePerKey{domain.OncePerOwnerFaction, domain.OncePerTargetFaction, domain.OncePerAsset}
	contestRollers = []domain.ContestRoller{domain.ContestRollerSelf, domain.ContestRollerTarget}
	restrictSides  = []domain.RestrictSide{domain.RestrictDefense, domain.RestrictAttackAndCounter}
	restrictAllows = []domain.RestrictAllow{domain.AllowOnlyListed, domain.AllowExceptListed}
)

func convertEffects(records []effectRecord) ([]domain.EffectDefinition, error) {
	if len(records) == 0 {
		return nil, nil
	}
	definitions := make([]domain.EffectDefinition, 0, len(records))
	for i, record := range records {
		definition, err := convertEffect(record)
		if err != nil {
			return nil, fmt.Errorf("effects[%d]: %w", i, err)
		}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

func convertEffect(record effectRecord) (domain.EffectDefinition, error) {
	trigger, err := toEnum("effect trigger", record.Trigger, effectTriggers)
	if err != nil {
		return domain.EffectDefinition{}, err
	}
	effectType, err := toEnum("special effect", record.Effect, specialEffects)
	if err != nil {
		return domain.EffectDefinition{}, err
	}
	if record.Filters == nil {
		return domain.EffectDefinition{}, fmt.Errorf("missing [filters] block: owner and scope are required")
	}
	filters, err := convertEffectFilters(*record.Filters)
	if err != nil {
		return domain.EffectDefinition{}, err
	}
	gates, err := convertEffectGates(record.Gates)
	if err != nil {
		return domain.EffectDefinition{}, err
	}
	for _, block := range presentParamBlocks(record) {
		if block != effectType {
			return domain.EffectDefinition{}, fmt.Errorf("parameter block [%s] does not match effect %q", block, effectType)
		}
	}
	definition := domain.EffectDefinition{Trigger: trigger, Effect: effectType, Filters: filters, Gates: gates}
	if err := convertEffectParams(record, &definition); err != nil {
		return domain.EffectDefinition{}, err
	}
	return definition, nil
}

func convertEffectFilters(record effectFiltersRecord) (domain.EffectFilters, error) {
	if record.Owner == "" {
		return domain.EffectFilters{}, fmt.Errorf("filters.owner is required")
	}
	if record.Scope == "" {
		return domain.EffectFilters{}, fmt.Errorf("filters.scope is required")
	}
	owner, err := toEnum("effect owner", record.Owner, effectOwners)
	if err != nil {
		return domain.EffectFilters{}, err
	}
	scope, err := toEnum("effect scope", record.Scope, effectScopes)
	if err != nil {
		return domain.EffectFilters{}, err
	}
	filters := domain.EffectFilters{Owner: owner, Scope: scope, Action: record.Action, Stealthy: record.Stealthy}
	if record.Attribute != "" {
		attribute, err := toFactionStat(record.Attribute)
		if err != nil {
			return domain.EffectFilters{}, fmt.Errorf("filters.attribute: %w", err)
		}
		filters.Attribute = attribute
	}
	for i, value := range record.Phase {
		phase, err := toEnum("roll phase", value, effectRollPhases)
		if err != nil {
			return domain.EffectFilters{}, fmt.Errorf("filters.phase[%d]: %w", i, err)
		}
		filters.Phases = append(filters.Phases, phase)
	}
	for i, value := range record.AssetType {
		assetType, err := toAssetType(value)
		if err != nil {
			return domain.EffectFilters{}, fmt.Errorf("filters.asset_type[%d]: %w", i, err)
		}
		filters.AssetTypes = append(filters.AssetTypes, assetType)
	}
	for i, value := range record.Category {
		category, err := toFactionStat(value)
		if err != nil {
			return domain.EffectFilters{}, fmt.Errorf("filters.category[%d]: %w", i, err)
		}
		filters.Categories = append(filters.Categories, category)
	}
	return filters, nil
}

func convertEffectGates(record *effectGatesRecord) (domain.EffectGates, error) {
	if record == nil {
		return domain.EffectGates{}, nil
	}
	gates := domain.EffectGates{NoStack: record.NoStack}
	if record.OncePer != "" {
		key, err := toEnum("once_per key", record.OncePer, oncePerKeys)
		if err != nil {
			return domain.EffectGates{}, err
		}
		gates.OncePer = key
	}
	if contest := record.Contest; contest != nil {
		roller, err := toEnum("contest roller", contest.Roller, contestRollers)
		if err != nil {
			return domain.EffectGates{}, err
		}
		attacker, err := toFactionStat(contest.AttackerStat)
		if err != nil {
			return domain.EffectGates{}, fmt.Errorf("contest attacker_stat: %w", err)
		}
		defender, err := toFactionStat(contest.DefenderStat)
		if err != nil {
			return domain.EffectGates{}, fmt.Errorf("contest defender_stat: %w", err)
		}
		gates.Contest = &domain.EffectContest{
			Contest: domain.Contest{AttackerStat: attacker, DefenderStat: defender},
			Roller:  roller,
		}
	}
	return gates, nil
}

// presentParamBlocks names the parameter blocks set on record. Each block is named
// after the effect it configures, so a block for any other effect is a mismatch.
func presentParamBlocks(record effectRecord) []domain.SpecialEffectType {
	var present []domain.SpecialEffectType
	if record.CoinGain != nil {
		present = append(present, domain.SpecialCoinGain)
	}
	if record.CoinSteal != nil {
		present = append(present, domain.SpecialCoinSteal)
	}
	if record.SelfDamage != nil {
		present = append(present, domain.SpecialSelfDamage)
	}
	if record.BonusDie != nil {
		present = append(present, domain.SpecialBonusDie)
	}
	if record.Reroll != nil {
		present = append(present, domain.SpecialReroll)
	}
	if record.TechLevelFloor != nil {
		present = append(present, domain.SpecialTechLevelFloor)
	}
	if record.CostSurcharge != nil {
		present = append(present, domain.SpecialCostSurcharge)
	}
	if record.RestrictAttack != nil {
		present = append(present, domain.SpecialRestrictAttack)
	}
	if record.GrantMovement != nil {
		present = append(present, domain.SpecialGrantMovement)
	}
	if record.CaptureTarget != nil {
		present = append(present, domain.SpecialCaptureTarget)
	}
	if record.DiscountRepair != nil {
		present = append(present, domain.SpecialDiscountRepair)
	}
	return present
}

// convertEffectParams fills the parameter pointer for definition.Effect. An effect
// that takes parameters must carry its block; an effect that takes none falls
// through.
func convertEffectParams(record effectRecord, definition *domain.EffectDefinition) error {
	switch definition.Effect {
	case domain.SpecialCoinGain:
		amount, err := convertAmount(definition.Effect, record.CoinGain)
		if err != nil {
			return err
		}
		definition.CoinGain = &domain.CoinGainParams{Amount: amount}
	case domain.SpecialCoinSteal:
		amount, err := convertAmount(definition.Effect, record.CoinSteal)
		if err != nil {
			return err
		}
		definition.CoinSteal = &domain.CoinStealParams{Amount: amount}
	case domain.SpecialSelfDamage:
		amount, err := convertAmount(definition.Effect, record.SelfDamage)
		if err != nil {
			return err
		}
		definition.SelfDamage = &domain.SelfDamageParams{Amount: amount}
	case domain.SpecialBonusDie:
		if record.BonusDie == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.BonusDie.Sides <= 0 || record.BonusDie.Count <= 0 {
			return fmt.Errorf("bonus_die sides and count must be > 0, got %d and %d", record.BonusDie.Sides, record.BonusDie.Count)
		}
		definition.BonusDie = &domain.BonusDieParams{Sides: record.BonusDie.Sides, Count: record.BonusDie.Count}
	case domain.SpecialReroll:
		if record.Reroll == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.Reroll.Dice <= 0 {
			return fmt.Errorf("reroll dice must be > 0, got %d", record.Reroll.Dice)
		}
		definition.Reroll = &domain.RerollParams{Dice: record.Reroll.Dice, Elective: record.Reroll.Elective}
	case domain.SpecialTechLevelFloor:
		if record.TechLevelFloor == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.TechLevelFloor.Level <= 0 {
			return fmt.Errorf("tech_level_floor level must be > 0, got %d", record.TechLevelFloor.Level)
		}
		definition.TechLevelFloor = &domain.TechLevelFloorParams{Level: record.TechLevelFloor.Level}
	case domain.SpecialCostSurcharge:
		if record.CostSurcharge == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.CostSurcharge.Amount <= 0 {
			return fmt.Errorf("cost_surcharge amount must be > 0, got %d", record.CostSurcharge.Amount)
		}
		definition.CostSurcharge = &domain.CostSurchargeParams{Amount: record.CostSurcharge.Amount}
	case domain.SpecialRestrictAttack:
		if record.RestrictAttack == nil {
			return missingParamBlock(definition.Effect)
		}
		side, err := toEnum("restrict_attack side", record.RestrictAttack.Side, restrictSides)
		if err != nil {
			return err
		}
		allow, err := toEnum("restrict_attack allow", record.RestrictAttack.Allow, restrictAllows)
		if err != nil {
			return err
		}
		definition.RestrictAttack = &domain.RestrictAttackParams{Side: side, Allow: allow}
	case domain.SpecialGrantMovement:
		if record.GrantMovement == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.GrantMovement.Relocate == nil {
			return fmt.Errorf("grant_movement requires a [grant_movement.relocate] block")
		}
		relocate, err := convertRelocate(record.GrantMovement.Relocate)
		if err != nil {
			return fmt.Errorf("grant_movement: %w", err)
		}
		definition.GrantMovement = &domain.GrantMovementParams{Relocate: *relocate}
	case domain.SpecialCaptureTarget:
		coin, err := convertAmount(definition.Effect, record.CaptureTarget)
		if err != nil {
			return err
		}
		definition.CaptureTarget = &domain.CaptureTargetParams{Coin: coin}
	case domain.SpecialDiscountRepair:
		if record.DiscountRepair == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.DiscountRepair.Amount <= 0 {
			return fmt.Errorf("discount_repair amount must be > 0, got %d", record.DiscountRepair.Amount)
		}
		definition.DiscountRepair = &domain.DiscountRepairParams{Amount: record.DiscountRepair.Amount}
	}
	return nil
}

func convertAmount(effectType domain.SpecialEffectType, record *amountRecord) (domain.DiceRoll, error) {
	if record == nil {
		return domain.DiceRoll{}, missingParamBlock(effectType)
	}
	amount, err := parseAmount(record.Amount)
	if err != nil {
		return domain.DiceRoll{}, fmt.Errorf("%s amount: %w", effectType, err)
	}
	if amount == nil {
		return domain.DiceRoll{}, fmt.Errorf("%s amount is required", effectType)
	}
	if amount.NumDice == 0 && amount.Modifier <= 0 {
		return domain.DiceRoll{}, fmt.Errorf("%s amount must be > 0, got %q", effectType, record.Amount)
	}
	return *amount, nil
}

func missingParamBlock(effectType domain.SpecialEffectType) error {
	return fmt.Errorf("effect %q requires a [%s] block", effectType, effectType)
}

func toEnum[T ~string](kind, value string, allowed []T) (T, error) {
	for _, candidate := range allowed {
		if value == string(candidate) {
			return candidate, nil
		}
	}
	var zero T
	return zero, fmt.Errorf("unknown %s: %q (want one of %v)", kind, value, allowed)
}
