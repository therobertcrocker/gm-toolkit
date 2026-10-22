package domain

import (
	"maps"
	"slices"
	"strings"
)

type FactionStat string

const (
	StatForce   FactionStat = "Force"
	StatCunning FactionStat = "Cunning"
	StatWealth  FactionStat = "Wealth"
)

type FactionScale string

const (
	ScaleMinor   FactionScale = "minor"
	ScaleMajor   FactionScale = "major"
	ScaleHegemon FactionScale = "hegemon"
)

type Tag struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Effect      string `toml:"effect"`
}

type Goal struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Difficulty  string `toml:"difficulty"`
}

// ActiveGoal holds the faction's current goal and live progress state.
type ActiveGoal struct {
	GoalID          string   `toml:"goal_id"`
	TargetFactionID string   `toml:"target_faction_id"`
	TargetWorld     Location `toml:"target_world"`
	Progress        int      `toml:"progress"`
	ProcessPhase    int      `toml:"process_phase"`
	TurnsRemaining  int      `toml:"turns_remaining"`
}

type Faction struct {
	ID                      string            `toml:"id"`
	Name                    string            `toml:"name"`
	Scale                   FactionScale      `toml:"scale"`
	Force                   int               `toml:"force"`
	Cunning                 int               `toml:"cunning"`
	Wealth                  int               `toml:"wealth"`
	CurrentHP               int               `toml:"current_hp"`
	MaxHP                   int               `toml:"max_hp"`
	Coin                    int               `toml:"coin"`
	XP                      int               `toml:"xp"`
	Homeworld               Location          `toml:"homeworld"`
	Tags                    []*Tag            `toml:"tags"`
	ActiveGoal              *ActiveGoal       `toml:"active_goal"`
	Assets                  map[string]*Asset `toml:"assets"`
	Bases                   []*Base           `toml:"bases"`
	HookBudgets             map[string]int    `toml:"hook_budgets,omitempty"`
	PendingPurchaseDiscount int               `toml:"pending_purchase_discount,omitempty"`
}

// Clone deep-copies a faction so a caller can apply mutations to the copy without
// touching shared state: every reference-typed field gets its own copy. Update this
// when Faction gains a reference-typed field, or the copy will alias the original.
func (faction *Faction) Clone() *Faction {
	clone := *faction
	if faction.Assets != nil {
		clone.Assets = make(map[string]*Asset, len(faction.Assets))
		for id, asset := range faction.Assets {
			clone.Assets[id] = asset.Clone()
		}
	}
	if faction.Bases != nil {
		clone.Bases = make([]*Base, len(faction.Bases))
		for i, base := range faction.Bases {
			baseCopy := *base
			clone.Bases[i] = &baseCopy
		}
	}
	if faction.Tags != nil {
		clone.Tags = make([]*Tag, len(faction.Tags))
		for i, tag := range faction.Tags {
			tagCopy := *tag
			clone.Tags[i] = &tagCopy
		}
	}
	if faction.ActiveGoal != nil {
		goalCopy := *faction.ActiveGoal
		clone.ActiveGoal = &goalCopy
	}
	if faction.HookBudgets != nil {
		clone.HookBudgets = make(map[string]int, len(faction.HookBudgets))
		maps.Copy(clone.HookBudgets, faction.HookBudgets)
	}
	return &clone
}

func RatingsFromScale(s FactionScale) (primary, secondary, tertiary int) {
	switch s {
	case ScaleMajor:
		return 6, 5, 3
	case ScaleHegemon:
		return 8, 7, 5
	default: // minor
		return 4, 3, 1
	}
}

func AssetCountsFromScale(s FactionScale) (primary, other int) {
	switch s {
	case ScaleMajor:
		return 2, 2
	case ScaleHegemon:
		return 4, 4
	default: // minor
		return 1, 1
	}
}

func CalcMaxHP(f *Faction) int {
	return 4 + HPValueForRating(f.Force) + HPValueForRating(f.Cunning) + HPValueForRating(f.Wealth)
}

// StatScore returns the faction's attribute rating for an asset category's stat
// (Force / Cunning / Wealth); an unrecognized stat scores 0.
func (faction *Faction) StatScore(stat FactionStat) int {
	switch stat {
	case StatForce:
		return faction.Force
	case StatCunning:
		return faction.Cunning
	case StatWealth:
		return faction.Wealth
	default:
		return 0
	}
}

// BudgetSpent reports whether the hook budget key is spent this turn. An empty
// key is unlimited and is never spent.
func (faction *Faction) BudgetSpent(key string) bool {
	return key != "" && faction.HookBudgets[key] >= 1
}

// SpendBudget marks the hook budget key as spent this turn. An empty key is a
// no-op.
func (faction *Faction) SpendBudget(key string) {
	if key == "" {
		return
	}
	if faction.HookBudgets == nil {
		faction.HookBudgets = make(map[string]int)
	}
	faction.HookBudgets[key]++
}

func HPValueForRating(rating int) int {
	switch rating {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 4
	case 4:
		return 6
	case 5:
		return 9
	case 6:
		return 12
	case 7:
		return 16
	case 8:
		return 20
	default:
		return 0
	}
}

// NewFaction synthesizes a Faction from creation-time inputs. Applies SWN
// derivations: attribute ratings from scale, MaxHP from attributes,
// CurrentHP = MaxHP, homeworld Base of Influence at MaxHP.
func NewFaction(
	id, name string,
	scale FactionScale,
	primaryStat, secondaryStat, tertiaryStat FactionStat,
	tags []*Tag,
	goal *Goal,
	homeworld Location,
	assets []*Asset,
	coin int,
) *Faction {
	if primaryStat == secondaryStat || primaryStat == tertiaryStat || secondaryStat == tertiaryStat {
		panic("domain.NewFaction: primary, secondary, and tertiary stats must be distinct")
	}

	primary, secondary, tertiary := RatingsFromScale(scale)

	faction := &Faction{
		ID:        id,
		Name:      name,
		Scale:     scale,
		Tags:      tags,
		Coin:      coin,
		Homeworld: homeworld,
		Assets:    make(map[string]*Asset, len(assets)),
	}

	for _, stat := range []struct {
		name   FactionStat
		rating int
	}{
		{primaryStat, primary},
		{secondaryStat, secondary},
		{tertiaryStat, tertiary},
	} {
		switch stat.name {
		case StatForce:
			faction.Force = stat.rating
		case StatCunning:
			faction.Cunning = stat.rating
		case StatWealth:
			faction.Wealth = stat.rating
		}
	}

	faction.MaxHP = CalcMaxHP(faction)
	faction.CurrentHP = faction.MaxHP

	if goal != nil {
		faction.ActiveGoal = &ActiveGoal{GoalID: goal.ID}
	}

	for _, asset := range assets {
		faction.Assets[asset.ID] = asset
	}

	faction.Bases = []*Base{
		{
			ID:          id + "-homeworld",
			OwnerID:     id,
			Location:    homeworld,
			CurrentHP:   faction.MaxHP,
			MaxHP:       faction.MaxHP,
			IsHomeworld: true,
		},
	}

	return faction
}

func SortedAssets(faction *Faction) []*Asset {
	sorted := make([]*Asset, 0, len(faction.Assets))
	for _, asset := range faction.Assets {
		sorted = append(sorted, asset)
	}
	slices.SortFunc(sorted, func(a, b *Asset) int {
		return strings.Compare(a.ID, b.ID)
	})
	return sorted
}

// AssetOwner returns the faction that holds assetID, or nil. Asset IDs are unique
// across factions (NextAssetID), so at most one matches.
func AssetOwner(factions map[string]*Faction, assetID string) *Faction {
	for _, faction := range factions {
		if _, ok := faction.Assets[assetID]; ok {
			return faction
		}
	}
	return nil
}

// AttackBlocked reports whether a no_attack condition on asset holds now. It is
// checked when read: the source must still exist, share the asset's world, and
// belong to another faction. The source need not be active.
func AttackBlocked(factions map[string]*Faction, asset *Asset) bool {
	for _, condition := range asset.Conditions {
		if condition.Kind != ConditionNoAttack {
			continue
		}
		owner := AssetOwner(factions, condition.SourceAssetID)
		if owner == nil || owner.ID == asset.OwnerID {
			continue
		}
		source := owner.Assets[condition.SourceAssetID]
		if source.Location.WorldID != "" && source.Location.WorldID == asset.Location.WorldID {
			return true
		}
	}
	return false
}
