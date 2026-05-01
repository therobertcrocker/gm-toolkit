package digest

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

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
	Goal         *GoalRef
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
	Notes        []string
}

type GoalRef struct {
	ID   string
	Name string
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

func Build(
	records []domain.EventRecord,
	cycleNumber int,
	factionState *state.FactionState,
	rulebook *loader.Rulebook,
) (CycleDigest, error) {
	if len(records) == 0 {
		return CycleDigest{}, fmt.Errorf("no history records for cycle %d", cycleNumber)
	}
	return CycleDigest{Cycle: cycleNumber}, nil
}
