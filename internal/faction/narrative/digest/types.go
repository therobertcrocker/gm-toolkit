package digest

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

type CycleDigest struct {
	Cycle          int
	ActiveFactions []FactionBeat
	QuietFactions  []FactionRef
	Cross          []CrossEvent
	Headline       Headline
}

type FactionRef struct {
	ID   string
	Name string
}

type FactionBeat struct {
	Faction      FactionRef
	CoinDelta    int
	HPDelta      int
	XPGained     int
	Acquisitions []AssetMove
	Losses       []AssetMove
	Movements    []AssetMove
	Bribes       []BribeEvent
	Repairs      []RepairEvent
	Expansions   []ExpansionEvent
	GoalEvents   []GoalEvent
	StealthOps   []StealthEvent
	Refusals     []RefusalEvent
	Notes        []string
}

// RefusalEvent is a purchase a government refused. DecidedBy is the ruler's name,
// empty when the GM decided for an independent world.
type RefusalEvent struct {
	AssetName string
	World     string
	DecidedBy string
}

type AssetMove struct {
	AssetID      string
	AssetName    string
	DefinitionID string
	From         string
	To           string
	Cause        string
}

type BribeEvent struct {
	BaseID   string
	Location string
	Coin     int
}

type RepairEvent struct {
	Target    AssetMove
	HPGained  int
	Coin      int
	IsFaction bool
}

type ExpansionEvent struct {
	BaseID   string
	Location string
	NewBase  bool
	HPDelta  int
	Coin     int
}

type GoalEventKind int

const (
	GoalCompleted GoalEventKind = iota
	GoalAbandoned
	GoalHomeworldShift
	GoalTagGained
	GoalWorldGoverned
	GoalTagLost
)

type GoalEvent struct {
	Kind      GoalEventKind
	GoalID    string
	GoalName  string
	XPAwarded int
	FromWorld string
	ToWorld   string
	TagName   string
}

type StealthEvent struct {
	AssetID   string
	AssetName string
	Applied   bool
}

type CrossKind int

const (
	CrossAttack CrossKind = iota
	CrossAbilityStrike
)

type CrossEvent struct {
	Kind                   CrossKind
	Attacker               FactionRef
	Defender               FactionRef
	AttackerAsset          AssetMove
	DefenderAsset          AssetMove
	DamageToDefender       int
	DamageToAttacker       int
	DefenderAssetDestroyed bool
	AttackerAssetDestroyed bool
	BaseHit                *BaseHitDetail
	CoinDrained            int
	StealthBroken          bool
	Conditions             []domain.ConditionKind
	Counterstruck          bool
	Interceptions          []domain.SpecialEffectType
}

type BaseHitDetail struct {
	BaseID    string
	Location  string
	Damage    int
	Destroyed bool
}

type HeadlineKind int

const (
	HeadlineMajorAttack HeadlineKind = iota
	HeadlineGoalCompleted
	HeadlineFactionDestroyed
	HeadlineHomeworldShift
	HeadlineGoalAbandoned
	HeadlineQuiet
)

type Headline struct {
	Subject FactionRef
	Kind    HeadlineKind
	Detail  string
}
