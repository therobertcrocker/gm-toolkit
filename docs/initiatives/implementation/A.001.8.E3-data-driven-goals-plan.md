# Data-Driven Goals (A.001.8) — Effort 3 Plan: Turn-Start Pick

## In Plain Words

- **What:** a faction with no goal is asked to pick one at the start of its turn,
  after it spends XP on a raise and before income. The pick offers only the goals
  the faction may take now, asks for a target faction when the goal needs one, and
  may be declined. A new `GoalPicked` record sets the goal, and the story digest,
  the turn feed and the faction detail view show it. The faction wizard stops
  setting a goal.
- **Why:** today nothing gives a faction a new goal after it finishes one, Destroy
  the Foe never gets a target, and the wizard offers goals that cannot work. Effort
  1 also deleted the only record that could set a goal in play.
- **How it fits:** Effort 3, the last of `A.001.8` (arc `A.001`), on branch
  `feature/data-driven-goals`. Effort 1 made goals data and Effort 2 made them read
  every batch. This effort closes `B.039`, `B.040` and `B.041`; the initiative
  close follows.

## Context / Goal

- Overview: [`A.001.8-data-driven-goals-overview-plan.md`](./A.001.8-data-driven-goals-overview-plan.md),
  Effort 3, and its Decisions 5 (rules doc) and 7 (`needs_actions` by display name).
- Discovery: [`A.001.8-data-driven-goals-discovery.md`](../discovery/A.001.8-data-driven-goals-discovery.md),
  D-2 to D-4, D-6, D-10 to D-12 and D-25, and State the parts need.
- Effort 1: [`A.001.8.E1-data-driven-goals-plan.md`](./A.001.8.E1-data-driven-goals-plan.md)
  loads `start` (`inputs`, `needs_actions`, `gate`) and leaves it unread (its
  Decision 12). Effort 2: [`A.001.8.E2-data-driven-goals-plan.md`](./A.001.8.E2-data-driven-goals-plan.md).
- Architecture read: orchestrator, overlays and state-machine (Key Decisions);
  goals and effect-mutation through the overview.
- The discovery left no open question. Decisions 1 to 4 are Robert's calls from
  this session; the rest settle what the code showed.
- The finding that shaped the effort: Effort 1 deleted `GoalInitiated`, and the
  only other writer of `ActiveGoal` is `domain.NewFaction`, which only the wizard
  calls. The pick needs its own mutation.

## Decisions Ratified in Planning

1. **The pick is a new `GoalPicked` mutation, with a digest event and a wire line.**
   Robert's call. `GoalPicked{FactionID, GoalID, TargetFactionID, StealthedAtStart}`
   sets `ActiveGoal` when it applies; goals emit, never mutate. The digest credits
   it to the picker's beat with the target's name, and the wire renders it from one
   of two template sets, with or without a target. No headline. Rejected: a feed line
   only; the name `GoalSelected`.

2. **The overlay lists goal names, and the highlighted goal's description follows
   the cursor.** Robert's call. A second group asks for the target faction and is
   hidden unless the goal takes one. Decline is an option, and Esc is swallowed, as
   in `StatRaise`. Rejected: names only; names with each goal's computed threshold.

3. **The manage view's faction detail shows the goal's target by name.** Robert's
   call. `detail.New` takes the faction state, because the view holds none today and
   a target ID is not a name. Rejected: no target shown.

4. **Two phases: the engine pieces, then the turn step with its TUI.** Robert's call.
   Phase 1 adds the mutation, the record, the offers and the pick, with no turn
   change. Phase 2 adds `SelectGoal` to `PhaseCollector`, which forces the adapter
   ask, the overlay and the `askEntries` row into the same commit. Rejected: one
   phase (over the budget).

5. **The goal engine owns the offers and the pick.** Planning call. `Offers` checks
   the gates and lists the start-input candidates; `Pick` builds the mutation. The
   gate names and the start-record rule stay in the package that owns the goal
   vocabulary. The engine keeps the registered action names that `goal.New` already
   receives, so `needs_actions` reads them.

6. **A target candidate is any other faction above 0 HP, in faction ID order. A goal
   with an input and no candidate is not offered.** Planning call. Destroy the Foe
   leaves the pick when no rival is standing, as D-3's "living rivals" implies.

7. **`Pick` checks the choice against the offers.** Planning call. A goal not on
   offer, a missing or unknown target, or a target for a goal that takes none is an
   error. It is a collector bug, so it is not recoverable.

8. **`StealthedAtStart` is filled only for a goal whose progress uses
   `exclude_held_at_start`.** Planning call. It holds every stealthed asset the
   picker has, wherever it is, in ID order (D-10). Other saves stay clean.

9. **The completion record holds all three ratings per goal ID, written by the
   `GoalCompleted` apply arm.** Planning call; the shape for D-11. The gate is open
   when there is no record, or when the named rating is above the recorded one.

10. **`OnGoalPicked` is the one new observer method. A decline or a skip prints
    nothing.** Planning call. Stat raise prints a skip line every turn; the pick does
    not add that noise.

11. **The pick's batch goes through `applyAndRecord`.** Planning call. It reaches
    history, and every goal reads it (Effort 2 Decision 1); no goal event matches it.

12. **A pick makes the picker's digest beat non-quiet; the lede does not change.**
    Planning call. A new goal is news, but not a "major objective fulfilled".

13. **The test-harness and smoke collectors decline by default.** Planning call.
    Existing scenarios keep running with no goal and no new answers.

14. **The rules doc takes the pick, D-10 and D-11 in Commit 3, and the creation
    checklist in Commit 4.** Planning call (overview Decision 5). Each lands with the
    code that makes it true.

## Decision Record — Execution

1. **No anchor drift in Commits 1 and 2.** Every Find matched the source, and the
   test fixtures Commit 2 names (`newGoalFixture`, `shippedGoal`, `newEngine`,
   `testActionNames`) hold the factions and HP the tasks assume. Decision 12 needs
   no code: `isQuiet` already counts any goal event, so the pick's event makes the
   picker's beat non-quiet.

2. **A TOML round-trip test for the completion record (Commit 1).** Not in the plan.
   `CompletionRatings` is a saved map of maps keyed by a string type, and a field
   that fails to save fails silently. `TestFaction_CompletionRatings_TOMLRoundTrip`
   checks that it saves and loads unchanged, and that a nil record is left out.

3. **No anchor drift in Commit 3.** Every Find matched, and huh v1.0.0 has the
   `DescriptionFunc`, `OptionsFunc` and `WithHideFunc` calls the overlay uses. Two
   test details the plan left open: the overlay test names Destroy the Foe
   `CORE-G-008`, its shipped ID, and the Wealth gate scenario records all three
   ratings in `CompletionRatings`, as the `GoalCompleted` arm writes them
   (Decision 9).

4. **No anchor drift in Commit 4.** Every Find matched; `manage.go` already held
   the faction state the detail now takes, and `sort` stays in use in the wizard.

5. **Deferred: the goal pick cuts long descriptions at the panel edge.** Found in a
   live run after Commit 3. huh renders the description on one line, and no
   overlay sets a form width, so most of a goal's text is lost. Not fixed in this
   effort; it is Robert's call.

6. **Lists put their "do nothing" option first (fix after Commit 4).** Not in the
   plan; Robert's call. The same live run showed the goal pick opening with only
   "Decline goal" in view. huh puts the cursor on the option that matches the
   bound value and sets the list's scroll offset to that row without clamping it.
   A nil starting value matched the nil "do nothing" option, which was last, so
   the list opened scrolled to its bottom row. Select action, stat raise and the
   goal pick had this shape; relocate and repair orders already list theirs first.
   The three now list "do nothing" first, so Enter on an unread prompt still does
   nothing. Rejected: start the cursor on the first real option (Enter would then
   act); patch huh.

---

## Shared Context

### Branch, cadence and checks

- Branch `feature/data-driven-goals`. One commit per plan commit, scope `E3`. At the
  effort boundary: `/gestalt:effort-close A.001.8.E3`, then `/gestalt:initiative-close A.001.8`.
- Every commit builds and passes `go build ./... && go vet ./... && go test ./...`
  from the repo root. Run `gofmt -w` on each touched Go file.
- Edit files only with Edit and Write.
- Test tasks are written in plain words. A dice sequence or a scenario detail that
  runs short or long against the harness is fixed there and logged.
- A task that adds a cross-package name ends with "add the import the compiler
  asks for".
- Paths in task titles are relative to `internal/faction/` unless they start with
  `docs/`.

### Pre-merge additions (beyond the overview's)

- **Graduate to Key Decisions (goals):** the goal engine owns the offers, the gates
  and the start record; the choice is checked against the offers (Decisions 5 to 8).
- **Graduate to Key Decisions (effect-mutation):** `GoalPicked` is the one writer of
  `ActiveGoal`, and `GoalCompleted` writes the completion record (Decisions 1 and 9).
- **Overlays page body:** add `GoalPick` to the bespoke overlays; it follows the
  in-modal decline pattern (Decision 2).
- **Orchestrator page body:** the goal-pick phase between stat raise and
  bookkeeping, and `OnGoalPicked` (Decision 10).

## Out of Scope

- A dedicated goal-pick card in the execution view's right panel; the base detail
  shows during the pick.
- Goal names in the turn feed. `formatMutation` has no rulebook, so the feed shows
  goal IDs, as it shows asset definition IDs.
- The rules' "select a new goal" after a faction merge (rules doc, Merging); the
  engine has no merge.
- The architecture pages: the initiative close rewrites them (overview Decision 6).

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — The pick in the engine**
  - Commit 1 — `feat(E3): GoalPicked and the completion record` (10 tasks)
  - Commit 2 — `feat(E3): goal offers, gates and the pick` (4 tasks)
- [x] **Phase 2 — The pick in the turn**
  - Commit 3 — `feat(E3): turn-start goal pick` (16 tasks)
  - Commit 4 — `feat(E3): goals are picked only at turn start` (7 tasks)

### Phase 1 — The pick in the engine

**In plain words:** the engine gains a record that sets a faction's goal, a record
of its ratings at each goal's completion, and a goal-engine step that lists the
goals a faction may pick and builds the pick. Nothing in a turn calls them yet.

#### Commit 1 — `feat(E3): GoalPicked and the completion record`

##### Task 1 — `domain/mutation.go`

Add the mutation before `GoalAbandoned`.

**Find:**
```go
// GoalAbandoned records that a faction abandoned their active goal. Income for
```
**Replace with:**
```go
// GoalPicked records the goal a faction picked at the start of its turn, with the
// target faction the goal asked for and the assets stealthed when it started
// (D-10). Its apply sets ActiveGoal.
type GoalPicked struct {
	FactionID        string   `json:"faction_id"`
	GoalID           string   `json:"goal_id"`
	TargetFactionID  string   `json:"target_faction_id,omitempty"`
	StealthedAtStart []string `json:"stealthed_at_start,omitempty"`
}

func (m GoalPicked) Type() string { return "goal_picked" }

// GoalAbandoned records that a faction abandoned their active goal. Income for
```

##### Task 2 — `domain/faction.go`

The completion record, a ratings helper, and the record's deep copy.

**Find:**
```go
	// WorldTechLevels holds the worlds that count as a higher tech level for this
	// faction's purchases, by world ID (raise_tech_level, D-11).
	WorldTechLevels map[string]int `toml:"world_tech_levels,omitempty"`
}
```
**Replace with:**
```go
	// WorldTechLevels holds the worlds that count as a higher tech level for this
	// faction's purchases, by world ID (raise_tech_level, D-11).
	WorldTechLevels map[string]int `toml:"world_tech_levels,omitempty"`
	// CompletionRatings holds the faction's ratings when it last completed each
	// goal, by goal ID; a rating_above_last_completion gate reads it (D-11).
	CompletionRatings map[string]map[FactionStat]int `toml:"completion_ratings,omitempty"`
}
```

**Find:**
```go
	if faction.WorldTechLevels != nil {
		clone.WorldTechLevels = maps.Clone(faction.WorldTechLevels)
	}
	return &clone
}
```
**Replace with:**
```go
	if faction.WorldTechLevels != nil {
		clone.WorldTechLevels = maps.Clone(faction.WorldTechLevels)
	}
	if faction.CompletionRatings != nil {
		clone.CompletionRatings = make(map[string]map[FactionStat]int, len(faction.CompletionRatings))
		for goalID, ratings := range faction.CompletionRatings {
			clone.CompletionRatings[goalID] = maps.Clone(ratings)
		}
	}
	return &clone
}
```

**Find:**
```go
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
```
**Replace with:**
```go
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

// Ratings is the faction's three attribute ratings by stat.
func (faction *Faction) Ratings() map[FactionStat]int {
	return map[FactionStat]int{StatForce: faction.Force, StatCunning: faction.Cunning, StatWealth: faction.Wealth}
}
```

##### Task 3 — `domain/faction_test.go`

Add `TestFactionClone_CopiesTheCompletionRecord`: a faction with
`CompletionRatings{"CORE-G-011": {Wealth: 2}}` is cloned; setting the clone's
`["CORE-G-011"][Wealth]` to 5 and adding a `"CORE-G-001"` entry to the clone leaves
the original with one entry and Wealth 2. A faction with a nil record clones to a
nil record.

##### Task 4 — `engine/mutation/mutation.go`

`GoalPicked` sets the goal; `GoalCompleted` also writes the completion record.

**Find:**
```go
		case domain.GoalCompleted:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.ActiveGoal = nil
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
```
**Replace with:**
```go
		case domain.GoalPicked:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.ActiveGoal = &domain.ActiveGoal{
					GoalID:           v.GoalID,
					TargetFactionID:  v.TargetFactionID,
					StealthedAtStart: slices.Clone(v.StealthedAtStart),
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.GoalCompleted:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.ActiveGoal = nil
				if faction.CompletionRatings == nil {
					faction.CompletionRatings = map[string]map[domain.FactionStat]int{}
				}
				faction.CompletionRatings[v.GoalID] = faction.Ratings()
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
```

##### Task 5 — `engine/mutation/mutation_test.go`

Add two tests beside `TestApply_GoalFailedClearsTheGoal`, in its style:

- `TestApply_GoalPickedSetsTheGoal`:
  - "the picker holds the goal": `f1` holds no goal; applying
    `GoalPicked{FactionID: "f1", GoalID: "CORE-G-009", StealthedAtStart: []string{"a2"}}`
    leaves `ActiveGoal` equal to `{GoalID: "CORE-G-009", Progress: 0,
    StealthedAtStart: ["a2"]}`. Changing the mutation's slice afterwards does not
    change the faction's.
  - "the target is kept": `GoalPicked{FactionID: "f1", GoalID: "CORE-G-008",
    TargetFactionID: "f2"}` leaves `ActiveGoal.TargetFactionID == "f2"`.
  - "an unknown faction is a miss": one miss, as in the `GoalFailed` test.
- `TestApply_GoalCompletedRecordsTheRatings`: `f1` at Force 3, Cunning 2, Wealth 4
  holds `CORE-G-011`. Applying `GoalCompleted{FactionID: "f1", GoalID: "CORE-G-011"}`
  clears `ActiveGoal` and sets `CompletionRatings["CORE-G-011"]` to
  `{Force: 3, Cunning: 2, Wealth: 4}`. Then `f1.Wealth = 5`, it picks and completes
  `CORE-G-001`: a second entry appears and the `CORE-G-011` entry still reads
  Wealth 4.

##### Task 6 — `narrative/digest/types.go` and `narrative/digest/build.go`

In `types.go`, add the event kind and the target's name.

**Find:**
```go
	GoalTechLevelRaised
	GoalFailed
)
```
**Replace with:**
```go
	GoalTechLevelRaised
	GoalFailed
	GoalPicked
)
```

**Find:**
```go
	TechLevel int
	Cause     string
}
```
**Replace with:**
```go
	TechLevel  int
	Cause      string
	TargetName string
}
```

In `build.go`, credit the pick to the picker's beat, after the `goal_failed` case.

**Find:**
```go
					Cause:    failed.Cause,
				})

			case "influence_delta":
```
**Replace with:**
```go
					Cause:    failed.Cause,
				})

			case "goal_picked":
				var picked domain.GoalPicked
				if err := json.Unmarshal(mr.Payload, &picked); err != nil {
					return CycleDigest{}, err
				}
				picker := getOrCreateBeat(beats, picked.FactionID, factionState)
				seenFactionIDs[picked.FactionID] = struct{}{}
				event := GoalEvent{Kind: GoalPicked, GoalID: picked.GoalID, GoalName: resolveGoalName(picked.GoalID, rulebook)}
				if picked.TargetFactionID != "" {
					event.TargetName = resolveFactionName(picked.TargetFactionID, factionState)
				}
				picker.GoalEvents = append(picker.GoalEvents, event)

			case "influence_delta":
```

##### Task 7 — `narrative/digest/digest_test.go`

Add `TestGoalPicked` beside `TestGoalFailed`, with the same helpers (`eventRec`,
`mutRec`, `buildOne`, `minState`, `minRulebook`, `findBeat`):

- A `goal_picked` record by `faction-a` for `goal-1` ("Destroy the Foe") with
  target `faction-b` gives `faction-a` one `GoalPicked` event with that goal name
  and `TargetName == "Beta"`. `faction-b` gets no goal event, and the headline is
  `HeadlineQuiet`.
- The same record with no target gives an event with an empty `TargetName`.

##### Task 8 — `narrative/wire_templates.go` and `narrative/wire_renderer.go`

In `wire_templates.go`, add the two template sets after `goalFailedTemplates`.

**Find:**
```go
var goalHomeworldTemplates = []string{
```
**Replace with:**
```go
var goalPickedTemplates = []string{
	"{faction} took up a new objective: {goal}.",
	"New orders for {faction}: {goal}.",
}

// goalPickedTargetTemplates name the target faction a picked goal is aimed at.
var goalPickedTargetTemplates = []string{
	"{faction} took up {goal}, with {target} in their sights.",
	"{faction} marked {target} for {goal}.",
}

var goalHomeworldTemplates = []string{
```

In `wire_renderer.go`, render the event.

**Find:**
```go
	case digest.GoalHomeworldShift:
		tmpl := pick(rng, goalHomeworldTemplates)
```
**Replace with:**
```go
	case digest.GoalPicked:
		templates := goalPickedTemplates
		if ge.TargetName != "" {
			templates = goalPickedTargetTemplates
		}
		result := strings.ReplaceAll(pick(rng, templates), "{faction}", factionName)
		result = strings.ReplaceAll(result, "{goal}", ge.GoalName)
		return strings.ReplaceAll(result, "{target}", ge.TargetName)
	case digest.GoalHomeworldShift:
		tmpl := pick(rng, goalHomeworldTemplates)
```

##### Task 9 — `narrative/wire_renderer_test.go`

Add `TestWireRendererGoalPicked`, built like `goalFailedDigest`: a beat for "Iron
Hegemony" with one `GoalPicked` event for "Destroy the Foe". For seeds 1 to 10:
with `TargetName: "Free Traders"` the output holds the faction, the goal and the
target; with no target it holds the faction and the goal and no `{target}`.

##### Task 10 — `tui/views/turn/execution/format.go` and `format_test.go`

In `format.go`, a feed line for the pick, before the `GoalCompleted` case.

**Find:**
```go
	case domain.GoalCompleted:
		return fmt.Sprintf("goal completed: %s (+%d XP)", mut.FactionID, mut.XPAwarded), true
```
**Replace with:**
```go
	case domain.GoalPicked:
		if mut.TargetFactionID != "" {
			return fmt.Sprintf("goal picked: %s → %s", mut.GoalID, mut.TargetFactionID), true
		}
		return fmt.Sprintf("goal picked: %s", mut.GoalID), true
	case domain.GoalCompleted:
		return fmt.Sprintf("goal completed: %s (+%d XP)", mut.FactionID, mut.XPAwarded), true
```

In `format_test.go`, add two rows to the surfaced cases.

**Find:**
```go
		{"GoalCompleted", domain.GoalCompleted{FactionID: "beta", XPAwarded: 5}, "goal completed: beta (+5 XP)"},
```
**Replace with:**
```go
		{"GoalPicked", domain.GoalPicked{GoalID: "CORE-G-001"}, "goal picked: CORE-G-001"},
		{"GoalPicked target", domain.GoalPicked{GoalID: "CORE-G-008", TargetFactionID: "beta"}, "goal picked: CORE-G-008 → beta"},
		{"GoalCompleted", domain.GoalCompleted{FactionID: "beta", XPAwarded: 5}, "goal completed: beta (+5 XP)"},
```

##### Commit message

```
feat(E3): GoalPicked and the completion record

- GoalPicked{FactionID, GoalID, TargetFactionID, StealthedAtStart} sets a
  faction's goal when it applies; it is now the one writer of ActiveGoal
- GoalCompleted records the holder's ratings for that goal in
  Faction.CompletionRatings, which the Wealth of Worlds gate reads (D-11)
- the digest credits a pick to the picker with the target's name, the wire
  renders it with or without a target, and the feed shows a goal-picked line
```

#### Commit 2 — `feat(E3): goal offers, gates and the pick`

##### Task 1 — `domain/goal.go`

The pick's offer and answer, after `GoalGate`.

**Find:**
```go
type GoalGate struct {
	RatingAboveLastCompletion FactionStat
}
```
**Replace with:**
```go
type GoalGate struct {
	RatingAboveLastCompletion FactionStat
}

// GoalOffer is one goal the turn-start pick may offer, with the candidates for
// the start inputs it declares. Targets is set only for a target_faction input.
type GoalOffer struct {
	Goal    *Goal
	Targets []*Faction
}

// GoalChoice is a pick's answer: the goal, and the target faction when the goal
// takes one.
type GoalChoice struct {
	GoalID          string
	TargetFactionID string
}
```

##### Task 2 — `engine/goal/goal.go`

The engine keeps the registered action names for the `needs_actions` gate.

**Find:**
```go
type GoalEngine struct {
	goals    map[string]*compiledGoal
	rulebook *rulebook.Rulebook
	log      *slog.Logger
}

// New compiles every goal in the rulebook. A goal whose parts name something the
// engine cannot run is an error, not a goal that does nothing. actionNames are the
// registered actions a goal's action filters may name.
func New(rulebook *rulebook.Rulebook, actionNames []string, log *slog.Logger) (*GoalEngine, error) {
	goalEngine := &GoalEngine{goals: make(map[string]*compiledGoal, len(rulebook.Goals)), rulebook: rulebook, log: log}
```
**Replace with:**
```go
type GoalEngine struct {
	goals       map[string]*compiledGoal
	actionNames []string
	rulebook    *rulebook.Rulebook
	log         *slog.Logger
}

// New compiles every goal in the rulebook. A goal whose parts name something the
// engine cannot run is an error, not a goal that does nothing. actionNames are the
// registered actions: a goal's action filters may name only these, and its
// needs_actions gate is open only when it names these.
func New(rulebook *rulebook.Rulebook, actionNames []string, log *slog.Logger) (*GoalEngine, error) {
	goalEngine := &GoalEngine{
		goals:       make(map[string]*compiledGoal, len(rulebook.Goals)),
		actionNames: actionNames,
		rulebook:    rulebook,
		log:         log,
	}
```

##### Task 3 — `engine/goal/offers.go` (new)

Full contents:
```go
package goal

import (
	"fmt"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// Offers lists the goals faction may pick, in goal ID order: each goal whose
// gates are open (D-11, D-12), with the candidates for the start inputs it
// declares (D-3). A goal with an input that has no candidate is left out.
func (goalEngine *GoalEngine) Offers(faction *domain.Faction, factionState *state.FactionState) []domain.GoalOffer {
	var offers []domain.GoalOffer
	for _, goalID := range slices.Sorted(maps.Keys(goalEngine.goals)) {
		goal := goalEngine.goals[goalID].goal
		if !goalEngine.gatesOpen(goal, faction) {
			continue
		}
		offer := domain.GoalOffer{Goal: goal}
		if hasInput(goal, inputTargetFaction) {
			offer.Targets = livingRivals(faction.ID, factionState)
			if len(offer.Targets) == 0 {
				continue
			}
		}
		offers = append(offers, offer)
	}
	return offers
}

// gatesOpen reports whether goal's start gates let faction pick it: every action
// it needs is registered, and its rating, when it has a rating gate, is above the
// faction's record. A goal the faction never completed has no record, so its
// rating gate is open.
func (goalEngine *GoalEngine) gatesOpen(goal *domain.Goal, faction *domain.Faction) bool {
	if goal.Start == nil {
		return true
	}
	for _, action := range goal.Start.NeedsActions {
		if !slices.Contains(goalEngine.actionNames, action) {
			return false
		}
	}
	if gate := goal.Start.Gate; gate != nil {
		stat := gate.RatingAboveLastCompletion
		recorded, completed := faction.CompletionRatings[goal.ID]
		if completed && faction.StatScore(stat) <= recorded[stat] {
			return false
		}
	}
	return true
}

// livingRivals is every other faction above 0 HP, in faction ID order.
func livingRivals(factionID string, factionState *state.FactionState) []*domain.Faction {
	var rivals []*domain.Faction
	for _, rivalID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		rival := factionState.Factions[rivalID]
		if rivalID != factionID && rival.CurrentHP > 0 {
			rivals = append(rivals, rival)
		}
	}
	return rivals
}

// Pick builds the mutation that gives faction the goal it chose. The choice must
// name a goal in offers, with a target from that offer when the goal takes one and
// none when it does not; anything else is a collector bug. A goal whose progress
// leaves out the assets held stealthed at the start records them here (D-10).
func (goalEngine *GoalEngine) Pick(faction *domain.Faction, offers []domain.GoalOffer, choice domain.GoalChoice) (domain.GoalPicked, error) {
	index := slices.IndexFunc(offers, func(offer domain.GoalOffer) bool { return offer.Goal.ID == choice.GoalID })
	if index < 0 {
		return domain.GoalPicked{}, fmt.Errorf("goal %q is not on offer to %s", choice.GoalID, faction.ID)
	}
	offer := offers[index]
	targetOffered := slices.ContainsFunc(offer.Targets, func(target *domain.Faction) bool { return target.ID == choice.TargetFactionID })
	switch {
	case len(offer.Targets) == 0 && choice.TargetFactionID != "":
		return domain.GoalPicked{}, fmt.Errorf("goal %q takes no target faction, got %q", choice.GoalID, choice.TargetFactionID)
	case len(offer.Targets) > 0 && !targetOffered:
		return domain.GoalPicked{}, fmt.Errorf("goal %q needs a target faction on offer, got %q", choice.GoalID, choice.TargetFactionID)
	}
	picked := domain.GoalPicked{FactionID: faction.ID, GoalID: offer.Goal.ID, TargetFactionID: choice.TargetFactionID}
	if progress := offer.Goal.Progress; progress != nil && progress.Filters.ExcludeHeldAtStart {
		picked.StealthedAtStart = stealthedAssetIDs(faction)
	}
	return picked, nil
}

// stealthedAssetIDs is the IDs of faction's stealthed assets, wherever they are,
// in ID order.
func stealthedAssetIDs(faction *domain.Faction) []string {
	var assetIDs []string
	for _, assetID := range slices.Sorted(maps.Keys(faction.Assets)) {
		if faction.Assets[assetID].Stealthy {
			assetIDs = append(assetIDs, assetID)
		}
	}
	return assetIDs
}
```

##### Task 4 — `engine/goal/offers_test.go` (new)

Package `goal`. Each test builds its factions from `newGoalFixture(t, "Military
Conquest", 0)` (holder `f1` at Force 3, Cunning 3, Wealth 2; rival `f2` at HP 5;
third `f3` at HP 2) and reads goal IDs by name with `shippedGoal(...).goal.ID`.
The engine comes from `newEngine` (`testActionNames`) unless a test says
otherwise. Tests:

- `TestOffers_ShippedGoals`: 10 offers in goal ID order, and World Seizure is not
  among them. Destroy the Foe's `Targets` are `[f2, f3]`; every other offer has
  none.
- `TestOffers_NeedsActions`: an engine built with `testActionNames` plus
  `"Seize Planet"` offers World Seizure.
- `TestOffers_TargetsAreLivingRivals`: with `f3.CurrentHP = 0`, Destroy the Foe's
  targets are `[f2]`; with `f2` and `f3` both at 0, Destroy the Foe is not offered
  and the other nine are.
- `TestOffers_WealthGate`, a table on whether Wealth of Worlds is offered to `f1`
  at Wealth 2:
  - no record: offered
  - `CompletionRatings[wealthOfWorlds] = {Wealth: 2}`: not offered
  - `CompletionRatings[wealthOfWorlds] = {Wealth: 1}`: offered
  - a record for Military Conquest only, at Wealth 2: offered
- `TestPick`, a table over `Pick(f1, offers, choice)` with the offers from
  `Offers`:
  - Destroy the Foe with target `f2`: `GoalPicked{FactionID: "f1", GoalID:
    destroyTheFoe, TargetFactionID: "f2"}`, no error
  - Destroy the Foe with no target: an error that names "needs a target faction"
  - Destroy the Foe with target `f9`: the same error
  - Military Conquest with target `f2`: an error that names "takes no target"
  - World Seizure (not on offer): an error that names "not on offer"
- `TestPick_RecordsStealthedAtStart`: `f1` holds stealthed `a2` and `a1` (on any
  worlds) and an unstealthed `a3`. Picking Inside Enemy Territory gives
  `StealthedAtStart == ["a1", "a2"]`; picking Military Conquest gives nil.

##### Commit message

```
feat(E3): goal offers, gates and the pick

- domain.GoalOffer and GoalChoice carry the pick's offer and answer
- GoalEngine.Offers lists the goals a faction may pick: the needs_actions
  gate reads the registered actions (D-12), the Wealth of Worlds gate reads the
  completion record (D-11), and a target input lists the living rivals (D-3)
- GoalEngine.Pick checks the choice against the offers and builds GoalPicked,
  recording the assets stealthed at the start where the goal needs them (D-10)
```

### Phase 2 — The pick in the turn

**In plain words:** each faction's turn gains a goal-pick step after stat raise,
which asks a faction with no goal to pick one through a new prompt in the TUI. The
wizard stops setting goals, and the faction detail shows a goal's target.

#### Commit 3 — `feat(E3): turn-start goal pick`

##### Task 1 — `engine/collector.go`

**Find:**
```go
// PhaseCollector abstracts the orchestrator-owned input prompts: the phase
// checkpoints, action choice, stat raise, movement decisions, transport cargo and
```
**Replace with:**
```go
// PhaseCollector abstracts the orchestrator-owned input prompts: the phase
// checkpoints, action choice, stat raise, goal pick, movement decisions, transport cargo and
```

**Find:**
```go
	SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error)
```
**Replace with:**
```go
	SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error)
	SelectGoal(faction *domain.Faction, offers []domain.GoalOffer) (*domain.GoalChoice, error)
```

##### Task 2 — `engine/observer.go`

**Find:**
```go
	OnStatRaiseSkipped(faction *domain.Faction)
```
**Replace with:**
```go
	OnStatRaiseSkipped(faction *domain.Faction)
	OnGoalPicked(faction *domain.Faction, mutations []domain.Mutation)
```

##### Task 3 — `engine/orchestrator.go`

Run the pick after stat raise.

**Find:**
```go
// RunFactionTurn drives one faction's turn from setup through state save:
// stat-raise, bookkeeping, movement, action.
```
**Replace with:**
```go
// RunFactionTurn drives one faction's turn from setup through state save:
// stat-raise, goal pick, bookkeeping, movement, action.
```

**Find:**
```go
	if err := e.runStatRaisePhase(faction, factionState, paths, collectors, observer, turnLog); err != nil {
		return false, err
	}
```
**Replace with:**
```go
	if err := e.runStatRaisePhase(faction, factionState, paths, collectors, observer, turnLog); err != nil {
		return false, err
	}

	// deliberate: the pick follows the raise, not precedes it, because a Wealth raise this turn can open Wealth of Worlds for the same turn's pick (D-2).
	if err := e.runGoalPickPhase(faction, factionState, paths, collectors, observer, turnLog); err != nil {
		return false, err
	}
```

Add the phase and its helper after `runStatRaisePhase`.

**Find:**
```go
	phaseLog.Info("phase end", "outcome", "raised", "stat", *statToRaise)
	return nil
}
```
**Replace with:**
```go
	phaseLog.Info("phase end", "outcome", "raised", "stat", *statToRaise)
	return nil
}

func (e *Engine) runGoalPickPhase(
	faction *domain.Faction,
	factionState *state.FactionState,
	paths *campaigns.Paths,
	collectors Collectors,
	observer TurnObserver,
	turnLog *slog.Logger,
) error {
	phaseLog := turnLog.With("phase", "goal_pick")
	logging.PhaseStart(phaseLog, "goal_pick")
	phaseLog.Info("phase begin")

	pickMutations, err := e.prepareGoalPick(faction, factionState, collectors.Phase)
	if err == nil && len(pickMutations) > 0 {
		pickMutations, err = e.applyAndRecord(factionState, faction, pickMutations, paths, phaseLog)
	}
	if err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}
	if len(pickMutations) == 0 {
		phaseLog.Info("phase end", "outcome", "skipped")
		return nil
	}
	observer.OnGoalPicked(faction, pickMutations)
	phaseLog.Info("phase end", "outcome", "picked")
	return nil
}

// prepareGoalPick asks a faction with no goal to pick one of the goals on offer.
// It asks nothing when the faction holds a goal or nothing is on offer. A nil
// answer declines, and the ask comes back next turn (D-4).
func (e *Engine) prepareGoalPick(faction *domain.Faction, factionState *state.FactionState, collector PhaseCollector) ([]domain.Mutation, error) {
	if faction.ActiveGoal != nil {
		return nil, nil
	}
	offers := e.Goal.Offers(faction, factionState)
	if len(offers) == 0 {
		return nil, nil
	}
	choice, err := collector.SelectGoal(faction, offers)
	if err != nil {
		return nil, fmt.Errorf("selecting goal: %w", err)
	}
	if choice == nil {
		return nil, nil
	}
	picked, err := e.Goal.Pick(faction, offers, *choice)
	if err != nil {
		return nil, fmt.Errorf("picking goal: %w", err)
	}
	return []domain.Mutation{picked}, nil
}
```

##### Task 4 — `engine/orchestrator_test.go`

The test doubles gain the new methods.

**Find:**
```go
func (c *minimalPhaseCollector) SelectStatRaise(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
	return nil, nil
}
```
**Replace with:**
```go
func (c *minimalPhaseCollector) SelectStatRaise(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
	return nil, nil
}
func (c *minimalPhaseCollector) SelectGoal(_ *domain.Faction, _ []domain.GoalOffer) (*domain.GoalChoice, error) {
	return nil, nil
}
```

**Find:**
```go
func (o *errorObserver) OnStatRaiseSkipped(_ *domain.Faction) {}
```
**Replace with:**
```go
func (o *errorObserver) OnStatRaiseSkipped(_ *domain.Faction) {}

func (o *errorObserver) OnGoalPicked(_ *domain.Faction, _ []domain.Mutation) {}
```

##### Task 5 — `engine/testharness/collector.go`

**Find:**
```go
	SelectStatRaiseFn            func(*domain.Faction, []domain.FactionStat) (*domain.FactionStat, error)
```
**Replace with:**
```go
	SelectStatRaiseFn            func(*domain.Faction, []domain.FactionStat) (*domain.FactionStat, error)
	SelectGoalFn                 func(*domain.Faction, []domain.GoalOffer) (*domain.GoalChoice, error)
```

**Find:**
```go
	return nil, nil // default: skip raise
}
```
**Replace with:**
```go
	return nil, nil // default: skip raise
}

func (c *ScriptedCollector) SelectGoal(faction *domain.Faction, offers []domain.GoalOffer) (*domain.GoalChoice, error) {
	if c.SelectGoalFn != nil {
		return c.SelectGoalFn(faction, offers)
	}
	return nil, nil // default: decline, so a scenario runs with no goal unless it picks one
}
```

##### Task 6 — `engine/testharness/observer.go`

**Find:**
```go
func (r *RecordingObserver) OnStatRaiseSkipped(faction *domain.Faction) {
	r.Events = append(r.Events, ObservedEvent{Kind: "StatRaiseSkipped", Faction: faction})
}
```
**Replace with:**
```go
func (r *RecordingObserver) OnStatRaiseSkipped(faction *domain.Faction) {
	r.Events = append(r.Events, ObservedEvent{Kind: "StatRaiseSkipped", Faction: faction})
}

func (r *RecordingObserver) OnGoalPicked(faction *domain.Faction, mutations []domain.Mutation) {
	r.Events = append(r.Events, ObservedEvent{Kind: "GoalPicked", Faction: faction, Payload: mutations})
}
```

##### Task 7 — `engine/testharness/scenarios/goal_pick_test.go` (new)

Package `scenarios`, in the style of `stat_raise_test.go`: `Turn.Start`, then
`RunCycle`, with every faction passing its action. Goal IDs come from `goalID` in
`goals_test.go`. Tests:

- `TestRunCycle_GoalPick_PicksAGoalWithATarget`: `alpha` and `beta`, neither holds
  a goal. `SelectGoalFn` records the offers made to `alpha` and picks Destroy the
  Foe with target `beta`; it declines for `beta`. Checks:
  - the offers to `alpha` leave out World Seizure, and Destroy the Foe's targets
    are exactly `[beta]`
  - `alpha.ActiveGoal` is Destroy the Foe with `TargetFactionID == "beta"`, and
    `beta.ActiveGoal` is nil
  - history holds a `goal_picked` mutation for `alpha`
  - the observer's `GoalPicked` event for `alpha` comes after its
    `StatRaiseSkipped` and before its `BookkeepingApplied`
- `TestRunCycle_GoalPick_HeldGoalIsNotAsked`: `alpha` alone, holding Military
  Conquest. `SelectGoalFn` fails the test if called. `alpha` still holds Military
  Conquest.
- `TestRunCycle_GoalPick_DeclineAsksAgainNextTurn`: `alpha` alone. `SelectGoalFn`
  counts its calls and declines. After two cycles (`Turn.Start` and `RunCycle`
  twice) it was called twice, `alpha.ActiveGoal` is nil, and history holds no
  `goal_picked`.
- `TestRunCycle_GoalPick_RaiseOpensTheWealthGate`: `alpha` alone at Force 3,
  Cunning 2, Wealth 1, with XP 2 and `CompletionRatings` for Wealth of Worlds at
  Wealth 1. `SelectGoalFn` records the offered goal IDs and declines. Two subtests:
  - the stat raise raises Wealth: Wealth of Worlds is offered
  - the stat raise declines: Wealth of Worlds is not offered

##### Task 8 — `tui/adapter/channels.go`

**Find:**
```go
	AskSelectAbilityWorldTarget
	// AskKindCount counts the kinds above; keep it last.
```
**Replace with:**
```go
	AskSelectAbilityWorldTarget
	AskSelectGoal
	// AskKindCount counts the kinds above; keep it last.
```

**Find:**
```go
	EvtRollSettled
)
```
**Replace with:**
```go
	EvtRollSettled
	EvtGoalPicked
)
```

**Find:**
```go
type SelectStatRaisePayload struct{ Eligible []domain.FactionStat }
```
**Replace with:**
```go
type SelectStatRaisePayload struct{ Eligible []domain.FactionStat }
type SelectGoalPayload struct{ Offers []domain.GoalOffer }
```

**Find:**
```go
type MovementTickedPayload struct{ Mutations []domain.Mutation }
```
**Replace with:**
```go
type GoalPickedPayload struct{ Mutations []domain.Mutation }
type MovementTickedPayload struct{ Mutations []domain.Mutation }
```

##### Task 9 — `tui/adapter/phase_collector.go` and `tui/adapter/observer.go`

In `phase_collector.go`:

**Find:**
```go
	return replyAs[*domain.FactionStat]("SelectStatRaise", raw)
}
```
**Replace with:**
```go
	return replyAs[*domain.FactionStat]("SelectStatRaise", raw)
}

func (p *phaseCollector) SelectGoal(faction *domain.Faction, offers []domain.GoalOffer) (*domain.GoalChoice, error) {
	raw, err := p.adapter.ask(AskSelectGoal, faction, SelectGoalPayload{Offers: offers})
	if err != nil {
		return nil, err
	}
	return replyAs[*domain.GoalChoice]("SelectGoal", raw)
}
```

In `observer.go`:

**Find:**
```go
func (o *observer) OnMovementTicked(faction *domain.Faction, muts []domain.Mutation) {
```
**Replace with:**
```go
func (o *observer) OnGoalPicked(faction *domain.Faction, muts []domain.Mutation) {
	o.adapter.emit(EvtGoalPicked, GoalPickedPayload{Mutations: muts})
}
func (o *observer) OnMovementTicked(faction *domain.Faction, muts []domain.Mutation) {
```

##### Task 10 — `tui/adapter/smoke.go`

**Find:**
```go
func (s *SmokePhaseCollector) SelectStatRaise(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
	return nil, nil
}
```
**Replace with:**
```go
func (s *SmokePhaseCollector) SelectStatRaise(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
	return nil, nil
}

func (s *SmokePhaseCollector) SelectGoal(_ *domain.Faction, _ []domain.GoalOffer) (*domain.GoalChoice, error) {
	return nil, nil
}
```

**Find:**
```go
func (o *SmokeObserver) OnStatRaiseSkipped(faction *domain.Faction) {
	o.log.Info("EVT StatRaiseSkipped", "faction", faction.ID)
}
```
**Replace with:**
```go
func (o *SmokeObserver) OnStatRaiseSkipped(faction *domain.Faction) {
	o.log.Info("EVT StatRaiseSkipped", "faction", faction.ID)
}
func (o *SmokeObserver) OnGoalPicked(faction *domain.Faction, mutations []domain.Mutation) {
	o.log.Info("EVT GoalPicked", "faction", faction.ID, "mutations", len(mutations))
}
```

##### Task 11 — `tui/adapter/collector_test.go`

Add a case for the new ask.

**Find:**
```go
	force := domain.StatForce
	decisions := []world.MovementDecision{{AssetID: asset.ID}}
```
**Replace with:**
```go
	force := domain.StatForce
	offers := []domain.GoalOffer{{Goal: &domain.Goal{ID: "CORE-G-001"}}}
	choice := &domain.GoalChoice{GoalID: "CORE-G-001"}
	decisions := []world.MovementDecision{{AssetID: asset.ID}}
```

**Find:**
```go
		{"SelectStatRaise", func() (any, error) { return collector.SelectStatRaise(alpha, []domain.FactionStat{force}) }, &force, &force},
```
**Replace with:**
```go
		{"SelectStatRaise", func() (any, error) { return collector.SelectStatRaise(alpha, []domain.FactionStat{force}) }, &force, &force},
		{"SelectGoal", func() (any, error) { return collector.SelectGoal(alpha, offers) }, choice, choice},
```

##### Task 12 — `tui/views/turn/overlay/goalpick.go` (new)

Full contents:
```go
package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// goalPickData is heap-allocated and pointer-bound to the huh fields so the
// value-copy of the overlay (bubbletea semantics) never orphans the bindings.
type goalPickData struct {
	offer    *domain.GoalOffer
	targetID string
}

// GoalPick is the turn-start goal pick: a select over the goals on offer, whose
// description follows the cursor, then a target select shown only for a goal
// that takes a target faction. Decline is an option (a nil offer), never Esc (D-4).
type GoalPick struct {
	data *goalPickData
	form *huh.Form
}

func NewGoalPick(offers []domain.GoalOffer) GoalPick {
	data := &goalPickData{}
	goalOptions := make([]huh.Option[*domain.GoalOffer], 0, len(offers)+1)
	for i := range offers {
		goalOptions = append(goalOptions, huh.NewOption(offers[i].Goal.Name, &offers[i]))
	}
	goalOptions = append(goalOptions, huh.NewOption("Decline goal", (*domain.GoalOffer)(nil)))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*domain.GoalOffer]().
				Title("Pick a goal").
				DescriptionFunc(func() string { return goalPickDescription(data.offer) }, &data.offer).
				Options(goalOptions...).
				Value(&data.offer),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Target faction").
				OptionsFunc(func() []huh.Option[string] { return targetOptions(data.offer) }, &data.offer).
				Value(&data.targetID).
				Validate(func(targetID string) error {
					if targetID == "" {
						return fmt.Errorf("pick a target faction")
					}
					return nil
				}),
		).WithHideFunc(func() bool { return data.offer == nil || len(data.offer.Targets) == 0 }),
	).WithTheme(styles.FormTheme())
	return GoalPick{data: data, form: form}
}

func goalPickDescription(offer *domain.GoalOffer) string {
	if offer == nil {
		return "No goal this turn; the pick comes back next turn."
	}
	return offer.Goal.Description
}

func targetOptions(offer *domain.GoalOffer) []huh.Option[string] {
	if offer == nil {
		return nil
	}
	options := make([]huh.Option[string], len(offer.Targets))
	for i, target := range offer.Targets {
		options[i] = huh.NewOption(target.Name, target.ID)
	}
	return options
}

func (o GoalPick) Init() tea.Cmd { return o.form.Init() }

func (o GoalPick) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return o, nil // decline is an option, so Esc must not abort the form
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.answer()
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

// answer is nil for a decline. A target is sent only for a goal that takes one,
// so a target chosen for an earlier highlighted goal is never carried over.
func (o GoalPick) answer() *domain.GoalChoice {
	offer := o.data.offer
	if offer == nil {
		return nil
	}
	choice := &domain.GoalChoice{GoalID: offer.Goal.ID}
	if len(offer.Targets) > 0 {
		choice.TargetFactionID = o.data.targetID
	}
	return choice
}

func (o GoalPick) View() string { return o.form.View() }

func (o GoalPick) Help() help.KeyMap { return formHelp{} }

var _ Overlay = GoalPick{}
```

##### Task 13 — `tui/views/turn/overlay/goalpick_test.go` (new)

Package `overlay`, with two offers: Military Conquest (no targets) and Destroy the
Foe (targets `f2` "Beta" and `f3` "Gamma").

- `TestGoalPickIgnoresEsc`: as `TestSelectAssetToLoseIgnoresEsc`; Esc gives no
  command and the overlay stays a `GoalPick`.
- `TestGoalPickAnswer`, a table that sets `data.offer` and `data.targetID` and
  reads `answer()`:
  - no offer (decline): nil
  - Military Conquest: `{GoalID: militaryConquest}`, no target
  - Destroy the Foe with `targetID "f2"`: `{GoalID: destroyTheFoe, TargetFactionID: "f2"}`
  - Military Conquest with a leftover `targetID "f2"`: no target
- `TestGoalPickTargetOptions`: `targetOptions` gives "Beta" and "Gamma" for
  Destroy the Foe, and nothing for Military Conquest or a nil offer.

##### Task 14 — `tui/views/turn/execution/execution.go`

The ask's row in `askEntries`.

**Find:**
```go
	adapter.AskSelectStatRaise: {"Stat Raise", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewStatRaise(msg.Faction, msg.Payload.(adapter.SelectStatRaisePayload).Eligible)
	}},
```
**Replace with:**
```go
	adapter.AskSelectStatRaise: {"Stat Raise", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewStatRaise(msg.Faction, msg.Payload.(adapter.SelectStatRaisePayload).Eligible)
	}},
	adapter.AskSelectGoal: {"Goal", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewGoalPick(msg.Payload.(adapter.SelectGoalPayload).Offers)
	}},
```

##### Task 15 — `tui/views/turn/execution/format.go`

The feed header for the pick.

**Find:**
```go
	case adapter.EvtStatRaiseApplied:
		return phaseHeader("stat raise applied"), msg.Payload.(adapter.StatRaiseAppliedPayload).Mutations
```
**Replace with:**
```go
	case adapter.EvtStatRaiseApplied:
		return phaseHeader("stat raise applied"), msg.Payload.(adapter.StatRaiseAppliedPayload).Mutations
	case adapter.EvtGoalPicked:
		return phaseHeader("goal pick"), msg.Payload.(adapter.GoalPickedPayload).Mutations
```

##### Task 16 — `tui/views/turn/execution/cycle_test.go` and `docs/rules/swn-faction-mechanics.md`

In `cycle_test.go`, the GM declines the pick, so the cycle runs as before.

**Find:**
```go
	case adapter.AskSelectStatRaise:
		return nil
```
**Replace with:**
```go
	case adapter.AskSelectStatRaise, adapter.AskSelectGoal:
		return nil
```

In the rules doc, the pick, D-10 and D-11.

**Find:**
```markdown
- **Inside Enemy Territory** — the goal counts the stealthed assets the faction
  holds at one time on worlds a rival governs, or on a rival's homeworld that has no
  faction government. An asset that loses stealth or leaves stops counting.
```
**Replace with:**
```markdown
- **Inside Enemy Territory** — the goal counts the stealthed assets the faction
  holds at one time on worlds a rival governs, or on a rival's homeworld that has no
  faction government. An asset that loses stealth or leaves stops counting. Assets
  already stealthed when the goal is picked never count, wherever they go.
- **Wealth of Worlds** — the goal is offered again only when the faction's Wealth
  is above its Wealth when it last completed the goal. Abandoning the goal sets no
  such limit.
- **Picking a goal** — a faction with no goal is asked at the start of its turn,
  after any XP spend and before income. It may decline; the question comes back
  each turn until it picks. A goal that names a target faction asks for one among
  the factions still standing, and is not offered when there is none. Planetary
  Seizure is not offered while the Seize Planet action is off.
```

##### Commit message

```
feat(E3): turn-start goal pick

- a goal-pick phase after stat raise asks a faction with no goal to pick one
  of the goals on offer; a nil answer declines and the ask returns next turn
  (D-2, D-4)
- PhaseCollector.SelectGoal and TurnObserver.OnGoalPicked, across the adapter,
  smoke, test-harness and orchestrator-test implementations
- the GoalPick overlay: goal names with the highlighted goal's description,
  then a target faction when the goal takes one; decline in-modal, Esc ignored
- docs: the rules doc covers the pick, the Inside Enemy Territory start record
  and the Wealth of Worlds gate
```

#### Commit 4 — `feat(E3): goals are picked only at turn start`

##### Task 1 — `tui/views/manage/wizard/wizard.go`

The wizard sets no goal (D-25). Four edits.

**Find:**
```go
	selectedTags     []*domain.Tag
	selectedGoal     *domain.Goal
	homeworldID      string
```
**Replace with:**
```go
	selectedTags     []*domain.Tag
	homeworldID      string
```

**Find:**
```go
		tags,
		m.data.selectedGoal,
		homeworld,
```
**Replace with:**
```go
		tags,
		homeworld,
```

**Find:**
```go
	goalGroup := huh.NewGroup(
		huh.NewSelect[*domain.Goal]().
			Title("Starting goal").
			Options(goalsToOptions(rb.Goals)...).
			Value(&data.selectedGoal),
	)

	homeworldGroup := huh.NewGroup(
```
**Replace with:**
```go
	homeworldGroup := huh.NewGroup(
```

**Find:**
```go
		tagsGroup,
		goalGroup,
		homeworldGroup,
```
**Replace with:**
```go
		tagsGroup,
		homeworldGroup,
```

**Find:**
```go
func goalsToOptions(goals map[string]*domain.Goal) []huh.Option[*domain.Goal] {
	sorted := make([]*domain.Goal, 0, len(goals))
	for _, goal := range goals {
		sorted = append(sorted, goal)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	options := make([]huh.Option[*domain.Goal], len(sorted))
	for i, goal := range sorted {
		options[i] = huh.NewOption(goal.Name, goal)
	}
	return options
}

```
**Replace with:** nothing (delete the function and its trailing blank line). If
`rb` is now unused in `buildForm`, the compiler says so; it is still read by the
tags group, so expect no change there.

##### Task 2 — `domain/faction.go`

`NewFaction` loses its goal parameter.

**Find:**
```go
// NewFaction synthesizes a Faction from creation-time inputs. Applies SWN
// derivations: attribute ratings from scale, MaxHP from attributes,
// CurrentHP = MaxHP, homeworld Base of Influence at MaxHP.
func NewFaction(
	id, name string,
	scale FactionScale,
	primaryStat, secondaryStat, tertiaryStat FactionStat,
	tagIDs []string,
	goal *Goal,
	homeworld Location,
```
**Replace with:**
```go
// NewFaction synthesizes a Faction from creation-time inputs. Applies SWN
// derivations: attribute ratings from scale, MaxHP from attributes,
// CurrentHP = MaxHP, homeworld Base of Influence at MaxHP. A new faction holds
// no goal; it picks one at its first turn (D-25).
func NewFaction(
	id, name string,
	scale FactionScale,
	primaryStat, secondaryStat, tertiaryStat FactionStat,
	tagIDs []string,
	homeworld Location,
```

**Find:**
```go
	if goal != nil {
		faction.ActiveGoal = &ActiveGoal{GoalID: goal.ID}
	}

```
**Replace with:** nothing (delete the block and its trailing blank line).

##### Task 3 — `domain/faction_test.go`

**Find:**
```go
	goal := &domain.Goal{ID: "goal-expand", Name: "Expand Influence"}

```
**Replace with:** nothing.

**Find:**
```go
		[]string{"SWN-T-020"},
		goal,
		homeworld,
```
**Replace with:**
```go
		[]string{"SWN-T-020"},
		homeworld,
```

**Find:**
```go
	if faction.ActiveGoal == nil || faction.ActiveGoal.GoalID != goal.ID {
		t.Errorf("ActiveGoal = %v, want GoalID=%q", faction.ActiveGoal, goal.ID)
	}
```
**Replace with:**
```go
	if faction.ActiveGoal != nil {
		t.Errorf("ActiveGoal = %+v, want nil", faction.ActiveGoal)
	}
```

##### Task 4 — `tui/views/manage/detail/detail.go`

The detail takes the faction state and names the goal's target.

**Find:**
```go
type Model struct {
	faction  *domain.Faction
	rulebook *rulebook.Rulebook
	viewport viewport.Model
	width    int
	height   int
	keys     keyMap
}

func New(faction *domain.Faction, rb *rulebook.Rulebook, width, height int) Model {
	m := Model{
		faction:  faction,
		rulebook: rb,
		width:    width,
		height:   height,
```
**Replace with:**
```go
type Model struct {
	faction      *domain.Faction
	factionState *state.FactionState
	rulebook     *rulebook.Rulebook
	viewport     viewport.Model
	width        int
	height       int
	keys         keyMap
}

func New(faction *domain.Faction, factionState *state.FactionState, rb *rulebook.Rulebook, width, height int) Model {
	m := Model{
		faction:      faction,
		factionState: factionState,
		rulebook:     rb,
		width:        width,
		height:       height,
```

**Find:**
```go
		b.WriteString("  " + styles.Strong.Render(goalName) +
			styles.Dim.Render(fmt.Sprintf("   ·   progress %d", f.ActiveGoal.Progress)) + "\n")
```
**Replace with:**
```go
		target := ""
		if targetID := f.ActiveGoal.TargetFactionID; targetID != "" {
			targetName := targetID
			if targetFaction, ok := m.factionState.Factions[targetID]; ok {
				targetName = targetFaction.Name
			}
			target = "   →   " + targetName
		}
		b.WriteString("  " + styles.Strong.Render(goalName) +
			styles.Dim.Render(fmt.Sprintf("%s   ·   progress %d", target, f.ActiveGoal.Progress)) + "\n")
```

Add the import the compiler asks for (`internal/faction/state`).

##### Task 5 — `tui/views/manage/manage.go`

**Find:**
```go
		m.detail = detail.New(faction, m.rulebook, m.termWidth, m.termHeight)
```
**Replace with:**
```go
		m.detail = detail.New(faction, m.factionState, m.rulebook, m.termWidth, m.termHeight)
```

##### Task 6 — `docs/rules/swn-faction-mechanics.md`

**Find:**
```markdown
5. **Select starting goal**
```
**Replace with:**
```markdown
5. **Select starting goal** — at the faction's first turn, through the goal pick;
   the faction wizard sets none
```

##### Task 7 — the plan docs

In `docs/initiatives/implementation/A.001.8-data-driven-goals-overview-plan.md`,
tick Effort 3.

**Find:**
```markdown
| [ ] | 3 | Turn-start pick |
```
**Replace with:**
```markdown
| [x] | 3 | Turn-start pick |
```

In this plan's Work Breakdown Overview, tick Phase 2.

##### Commit message

```
feat(E3): goals are picked only at turn start

- the faction wizard's goal step and goalsToOptions are removed, and
  NewFaction takes no goal: a new faction picks at its first turn (D-25)
- the manage view's faction detail names a goal's target faction
- docs: the creation checklist's starting goal moves to the first turn; the
  overview ticks Effort 3
```
