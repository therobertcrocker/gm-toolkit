package digest

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type crossKey struct {
	actorID  string
	targetID string
}

func getOrCreateBeat(beats map[string]*FactionBeat, factionID string, factionState *state.FactionState) *FactionBeat {
	if beat, ok := beats[factionID]; ok {
		return beat
	}
	beat := &FactionBeat{
		Faction: FactionRef{
			ID:   factionID,
			Name: resolveFactionName(factionID, factionState),
		},
	}
	beats[factionID] = beat
	return beat
}

func getOrCreateCross(crossMap map[crossKey]*CrossEvent, actorID, targetID string, kind CrossKind, factionState *state.FactionState) *CrossEvent {
	key := crossKey{actorID: actorID, targetID: targetID}
	if ce, ok := crossMap[key]; ok {
		return ce
	}
	ce := &CrossEvent{
		Kind:     kind,
		Attacker: FactionRef{ID: actorID, Name: resolveFactionName(actorID, factionState)},
		Defender: FactionRef{ID: targetID, Name: resolveFactionName(targetID, factionState)},
	}
	crossMap[key] = ce
	return ce
}

// findCrossTarget returns the first faction ID in a record's mutations that differs from record.FactionID.
// Only the first cross-faction target is detected; records targeting multiple distinct factions in one
// event are not supported by any current action and would be misattributed.
func findCrossTarget(record domain.EventRecord) string {
	type partial struct {
		FactionID string `json:"faction_id"`
	}
	for _, mr := range record.Mutations {
		var p partial
		if json.Unmarshal(mr.Payload, &p) == nil && p.FactionID != "" && p.FactionID != record.FactionID {
			return p.FactionID
		}
	}
	return ""
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

	beats := make(map[string]*FactionBeat)
	crossMap := make(map[crossKey]*CrossEvent)
	seenFactionIDs := make(map[string]struct{})

	for _, record := range records {
		seenFactionIDs[record.FactionID] = struct{}{}
		beat := getOrCreateBeat(beats, record.FactionID, factionState)
		targetID := findCrossTarget(record)

		// Per-record accumulators for mutation types that pair into a single composite event.
		repair := struct {
			assetID   string
			hp        int
			coin      int
			isFaction bool
		}{}
		bribe := struct {
			baseID   string
			location string
			coin     int
		}{}
		expand := struct {
			baseID   string
			location string
			newBase  bool
			hpDelta  int
			coin     int
		}{}

		// Pre-scan: find the refit-removed asset so asset_added can reference it as From.
		var refitRemovedAssetID string
		for _, mr := range record.Mutations {
			if mr.Type == "asset_removed" {
				var m domain.AssetRemoved
				if json.Unmarshal(mr.Payload, &m) == nil && m.Cause == "refit" {
					refitRemovedAssetID = m.AssetID
					break
				}
			}
		}

		for _, mr := range record.Mutations {
			switch mr.Type {

			case "coin_delta":
				var m domain.CoinDelta
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				if m.FactionID != record.FactionID && m.Cause == "ability" {
					// Defender loses coin via ability drain/steal.
					defBeat := getOrCreateBeat(beats, m.FactionID, factionState)
					seenFactionIDs[m.FactionID] = struct{}{}
					defBeat.CoinDelta += m.Delta
					if targetID != "" {
						ce := getOrCreateCross(crossMap, record.FactionID, m.FactionID, CrossAbilityStrike, factionState)
						ce.CoinDrained -= m.Delta
					}
				} else {
					beat.CoinDelta += m.Delta
					switch m.Cause {
					case "repair":
						repair.coin += m.Delta
					case "bribe":
						bribe.coin += m.Delta
					case "expand":
						expand.coin += m.Delta
					}
				}

			case "asset_added":
				var m domain.AssetAdded
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				assetName := m.Asset.ID
				if rulebook != nil {
					if def, ok := rulebook.Assets[m.Asset.DefinitionID]; ok {
						assetName = def.Name
					}
				}
				move := AssetMove{
					AssetID:      m.Asset.ID,
					AssetName:    assetName,
					DefinitionID: m.Asset.DefinitionID,
					Cause:        m.Cause,
				}
				if m.Cause == "refit" {
					move.From = refitRemovedAssetID
				}
				beat.Acquisitions = append(beat.Acquisitions, move)

			case "asset_removed":
				var m domain.AssetRemoved
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				switch m.Cause {
				case "sell":
					beat.Losses = append(beat.Losses, AssetMove{
						AssetID:   m.AssetID,
						AssetName: resolveAssetName(m.FactionID, m.AssetID, factionState, rulebook),
						Cause:     m.Cause,
					})
				case "refit":
					// already captured in refitRemovedAssetID; encoded as From on the paired AssetAdded
				case "attack":
					if m.FactionID != record.FactionID {
						getOrCreateBeat(beats, m.FactionID, factionState)
						ce := getOrCreateCross(crossMap, record.FactionID, m.FactionID, CrossAttack, factionState)
						seenFactionIDs[m.FactionID] = struct{}{}
						ce.DefenderAsset = AssetMove{
							AssetID:   m.AssetID,
							AssetName: resolveAssetName(m.FactionID, m.AssetID, factionState, rulebook),
							Cause:     m.Cause,
						}
						ce.DefenderAssetDestroyed = true
					} else {
						tgt := m.CausedByFactionID
						if tgt == "" {
							tgt = targetID
						}
						if tgt != "" {
							ce := getOrCreateCross(crossMap, record.FactionID, tgt, CrossAttack, factionState)
							ce.AttackerAsset = AssetMove{
								AssetID:   m.AssetID,
								AssetName: resolveAssetName(m.FactionID, m.AssetID, factionState, rulebook),
								Cause:     m.Cause,
							}
							ce.AttackerAssetDestroyed = true
						}
					}
				case "bookkeeping":
					// dropped — asset attrition is noise
				}

			case "asset_maintained_flag":
				// dropped

			case "faction_hp_delta":
				var m domain.FactionHPDelta
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				affectedBeat := getOrCreateBeat(beats, m.FactionID, factionState)
				seenFactionIDs[m.FactionID] = struct{}{}
				switch m.Cause {
				case "attack":
					affectedBeat.HPDelta += m.Delta
				case "repair":
					repair.isFaction = true
					repair.hp += m.Delta
				}

			case "asset_hp_delta":
				var m domain.AssetHPDelta
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				switch m.Cause {
				case "repair":
					repair.assetID = m.AssetID
					repair.hp += m.Delta
				case "attack":
					if m.FactionID == record.FactionID {
						if targetID != "" {
							ce := getOrCreateCross(crossMap, record.FactionID, targetID, CrossAttack, factionState)
							ce.DamageToAttacker -= m.Delta
						}
					} else {
						getOrCreateBeat(beats, m.FactionID, factionState)
						ce := getOrCreateCross(crossMap, record.FactionID, m.FactionID, CrossAttack, factionState)
						seenFactionIDs[m.FactionID] = struct{}{}
						ce.DamageToDefender -= m.Delta
					}
				}

			case "asset_moved":
				var m domain.AssetMoved
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				beat.Movements = append(beat.Movements, AssetMove{
					AssetID:   m.AssetID,
					AssetName: resolveAssetName(m.FactionID, m.AssetID, factionState, rulebook),
					From:      m.FromLocation,
					To:        m.ToLocation,
					Cause:     m.Cause,
				})

			case "asset_stealth_cleared":
				var m domain.AssetStealthCleared
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				affectedBeat := getOrCreateBeat(beats, m.FactionID, factionState)
				seenFactionIDs[m.FactionID] = struct{}{}
				affectedBeat.StealthOps = append(affectedBeat.StealthOps, StealthEvent{
					AssetID:   m.AssetID,
					AssetName: resolveAssetName(m.FactionID, m.AssetID, factionState, rulebook),
					Applied:   false,
				})
				if m.Cause == "ability" && m.FactionID != record.FactionID {
					ce := getOrCreateCross(crossMap, record.FactionID, m.FactionID, CrossAbilityStrike, factionState)
					ce.StealthBroken = true
				}

			case "asset_stealth_applied":
				var m domain.AssetStealthApplied
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				beat.StealthOps = append(beat.StealthOps, StealthEvent{
					AssetID:   m.AssetID,
					AssetName: resolveAssetName(m.FactionID, m.AssetID, factionState, rulebook),
					Applied:   true,
				})

			case "base_hp_delta":
				var m domain.BaseHPDelta
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				switch m.Cause {
				case "attack":
					getOrCreateBeat(beats, m.FactionID, factionState)
					ce := getOrCreateCross(crossMap, record.FactionID, m.FactionID, CrossAttack, factionState)
					seenFactionIDs[m.FactionID] = struct{}{}
					if ce.BaseHit == nil {
						ce.BaseHit = &BaseHitDetail{
							BaseID:   m.BaseID,
							Location: resolveBaseLocation(m.FactionID, m.BaseID, factionState),
						}
					}
					ce.BaseHit.Damage -= m.Delta
				case "expand":
					expand.baseID = m.BaseID
					expand.location = resolveBaseLocation(m.FactionID, m.BaseID, factionState)
					expand.hpDelta += m.Delta
				}

			case "base_destroyed":
				var m domain.BaseDestroyed
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				if m.Cause == "attack" {
					getOrCreateBeat(beats, m.FactionID, factionState)
					ce := getOrCreateCross(crossMap, record.FactionID, m.FactionID, CrossAttack, factionState)
					seenFactionIDs[m.FactionID] = struct{}{}
					if ce.BaseHit == nil {
						ce.BaseHit = &BaseHitDetail{
							BaseID:   m.BaseID,
							Location: resolveBaseLocation(m.FactionID, m.BaseID, factionState),
						}
					}
					ce.BaseHit.Destroyed = true
				}

			case "base_added":
				var m domain.BaseAdded
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				if m.Cause == "expand" {
					expand.baseID = m.Base.ID
					expand.location = m.Base.Location
					expand.newBase = true
				}

			case "base_healed":
				var m domain.BaseHealed
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				if m.Cause == "expand" {
					expand.baseID = m.BaseID
					expand.location = resolveBaseLocation(m.FactionID, m.BaseID, factionState)
					expand.hpDelta += m.Delta
				}

			case "base_expanded":
				var m domain.BaseExpanded
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				if m.Cause == "expand" {
					expand.baseID = m.BaseID
					expand.location = resolveBaseLocation(m.FactionID, m.BaseID, factionState)
					expand.hpDelta += m.Delta
				}

			case "goal_completed":
				var m domain.GoalCompleted
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				beat.GoalEvents = append(beat.GoalEvents, GoalEvent{
					Kind:      GoalCompleted,
					GoalID:    m.GoalID,
					GoalName:  resolveGoalName(m.GoalID, rulebook),
					XPAwarded: m.XPAwarded,
				})

			case "xp_awarded":
				var m domain.XPAwarded
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				beat.XPGained += m.Amount
				// Fold into matching GoalCompleted if XPAwarded wasn't set by the mutation.
				for i := range beat.GoalEvents {
					if beat.GoalEvents[i].Kind == GoalCompleted && beat.GoalEvents[i].XPAwarded == 0 {
						beat.GoalEvents[i].XPAwarded = m.Amount
						break
					}
				}

			case "homeworld_changed":
				var m domain.HomeworldChanged
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				beat.GoalEvents = append(beat.GoalEvents, GoalEvent{
					Kind:      GoalHomeworldShift,
					FromWorld: m.FromWorld,
					ToWorld:   m.ToWorld,
				})

			case "tag_added":
				var m domain.TagAdded
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				beat.GoalEvents = append(beat.GoalEvents, GoalEvent{
					Kind:    GoalTagGained,
					TagName: m.Tag.Name,
				})

			case "goal_abandoned":
				var m domain.GoalAbandoned
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				beat.GoalEvents = append(beat.GoalEvents, GoalEvent{
					Kind:     GoalAbandoned,
					GoalID:   m.GoalID,
					GoalName: resolveGoalName(m.GoalID, rulebook),
				})

			case "influence_delta":
				var m domain.InfluenceDelta
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				if m.Cause == "bribe" {
					bribe.baseID = m.BaseID
					bribe.location = resolveBaseLocation(m.FactionID, m.BaseID, factionState)
				}

			default:
				beat.Notes = append(beat.Notes, fmt.Sprintf("unknown mutation: %s", mr.Type))
			}
		}

		// Emit composite events after processing all mutations in this record.
		if repair.isFaction || repair.assetID != "" {
			re := RepairEvent{HPGained: repair.hp, Coin: repair.coin, IsFaction: repair.isFaction}
			if !repair.isFaction {
				re.Target = AssetMove{
					AssetID:   repair.assetID,
					AssetName: resolveAssetName(record.FactionID, repair.assetID, factionState, rulebook),
				}
			}
			beat.Repairs = append(beat.Repairs, re)
		}

		if bribe.baseID != "" || bribe.coin != 0 {
			beat.Bribes = append(beat.Bribes, BribeEvent{
				BaseID:   bribe.baseID,
				Location: bribe.location,
				Coin:     bribe.coin,
			})
		}

		if expand.baseID != "" || expand.hpDelta != 0 {
			beat.Expansions = append(beat.Expansions, ExpansionEvent{
				BaseID:   expand.baseID,
				Location: expand.location,
				NewBase:  expand.newBase,
				HPDelta:  expand.hpDelta,
				Coin:     expand.coin,
			})
		}
	}

	// Detect factions destroyed this cycle: seen in history but absent from live state.
	destroyedFactions := []FactionRef{}
	if factionState != nil {
		for id := range seenFactionIDs {
			if _, exists := factionState.Factions[id]; !exists {
				destroyedFactions = append(destroyedFactions, FactionRef{ID: id, Name: resolveFactionName(id, factionState)})
			}
		}
	}

	// Flatten cross-event map with stable ordering.
	crossEvents := make([]CrossEvent, 0, len(crossMap))
	for _, ce := range crossMap {
		crossEvents = append(crossEvents, *ce)
	}
	sort.Slice(crossEvents, func(i, j int) bool {
		if crossEvents[i].Attacker.Name != crossEvents[j].Attacker.Name {
			return crossEvents[i].Attacker.Name < crossEvents[j].Attacker.Name
		}
		return crossEvents[i].Defender.Name < crossEvents[j].Defender.Name
	})

	// Split beats into active and quiet.
	activeFactions := []FactionBeat{}
	quietFactions := []FactionRef{}
	for _, beat := range beats {
		if isQuiet(*beat, crossEvents) {
			quietFactions = append(quietFactions, beat.Faction)
		} else {
			activeFactions = append(activeFactions, *beat)
		}
	}
	sort.Slice(activeFactions, func(i, j int) bool {
		return activeFactions[i].Faction.Name < activeFactions[j].Faction.Name
	})
	sort.Slice(quietFactions, func(i, j int) bool {
		return quietFactions[i].Name < quietFactions[j].Name
	})

	headline := selectHeadline(activeFactions, crossEvents, destroyedFactions)

	return CycleDigest{
		Cycle:          cycleNumber,
		ActiveFactions: activeFactions,
		QuietFactions:  quietFactions,
		Cross:          crossEvents,
		Headline:       headline,
	}, nil
}
