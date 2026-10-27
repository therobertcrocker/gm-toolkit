# Data-Driven Tags (A.001.7) — Effort 1 Plan: Holder Foundation

## In Plain Words

- **What:** the base that lets a tag carry rules as data. Five parts:
  - **Tag IDs.** Factions store tag IDs (`SWN-T-020`), not copies of whole tags,
    and the government role is a `government = true` flag in the tag data.
  - **Effects that do not care who holds them.** One `Holder` value says "this
    asset" or "this faction, through a tag". All 25 effects register and check
    "am I still in play?" through it. The entry's scope decides who may hold it.
  - **Tags run effects.** A tag's `[[effects]]` list is loaded, checked when the
    engine starts, and registered and fired each cycle, as an asset's list is.
  - **The rule queries reach everyone.** Ties, a buyer's own purchase cost, and
    rolls against a faction now ask every relevant faction and asset.
  - **The four coded tags become data.** Warlike, Fanatical, Preceptor Archive and
    Scavengers move to `tags.toml`; Scavengers stops paying for fights it was not
    in. The old tag engine is deleted.
- **Why:** today a tag works only when Go code knows its ID, and every effect is
  written for an asset. After this effort, a new tag needs no code when existing
  effects can state its rule.
- **How it fits:** Effort 1 of four in `A.001.7` (arc `A.001`), on branch
  `feature/data-driven-tags`. Efforts 2 to 4 add the other 16 tags on top of the
  holder built here. `A.001.8` (goals) reuses it.

## Context / Goal

- Overview: [`A.001.7-data-driven-tags-overview-plan.md`](./A.001.7-data-driven-tags-overview-plan.md).
  Its Decisions 1, 2, 4, 6 and 7 stand. Decision 5 is replaced (Decision 1 below).
  Under Decision 7, every new effect here reads its values from its entry:
  `cost_discount` (amount, `min_tech_level`), `reroll` (`result`), `bonus_die`
  (`once_per`), `coin_gain` on `asset_destroyed` (amount, `caused_by`).
- Discovery: [`A.001.7-data-driven-tags-discovery.md`](../discovery/A.001.7-data-driven-tags-discovery.md),
  Design Summary, Effort 1; D-2, D-4, D-14, D-20.
- Architecture read: effect-mutation, hooks, persistence, narrative-digest.

## Decisions Ratified in Planning

1. **Effects are blind to their holder: a full port.** Robert's call. `effect.Holder`
   (`Faction`, `Asset`, `Def`, `Tag`) replaces `Faction`/`Asset`/`Def` in
   `RegisterContext`, and is added to `FireContext` and `MatchupContext`. Every
   registrar signs up through it (`Holder.Scope()`, `Holder.HookID()`,
   `Holder.ID()`, `Holder.Name()`). Every hook keeps a `holderRef` (IDs only) and
   checks at fire time through `holderRef.live` / `liveIn`: the asset is active, or
   the faction still holds the tag. Asset-bound hooks are ported too, so a later
   effort that opens an effect to tags adds only the scope's meaning.
   - **Reverses overview Decision 5, constraint 1** ("a registrar that a tag does
     not use stays unchanged"), and the discovery's "`Accepts` gains holders".
     Recorded in the overview's Decision Record.
   - Rejected: a split where only tag-capable hooks get the holder check (cheaper;
     every later opening also rewrites the check); a `Holders` list on `Accepts`
     (Decision 2 makes it redundant).

2. **Scope decides the holder; there is no holders list.** Robert's call. `self`
   and `world` both measure from the declaring asset, so a tag cannot use them.
   `validateEntry` refuses a tag entry with scope `self` or `world`, and a tag entry
   with `once_per = "asset"`. `Registrar.Validate` gains the holder kind. Two
   registrars accept `faction` scope but read the declaring asset, so they refuse a
   tag holder for now: `grant_movement` (range origin; Effort 2 builds Mercenary
   Group's self-move mode) and `ignore_coin_loss` (lead copy).

3. **`faction` scope is added to a registrar when a tag needs it.** Robert's call.
   Effort 1 adds it to `bonus_die` and `reroll` (`own_roll`). The new effects
   (`lose_ties`, `cost_discount`, `coin_gain` on `asset_destroyed`) accept only
   `faction`. After Effort 4, a sweep audit checks which registrars still lack it:
   a follow-on row added at pre-merge.

4. **All four dispatcher gaps close in Effort 1.** Robert's call.
   - `ResolveTie` asks each side's global, faction and asset buckets (the owner's
     assets in ID order), attacker side first. **The first answer that is not
     `TieStandard` wins** (planning call): with asset buckets in the walk, a
     resolver that does not match must not end the search. It also keeps a tag lost
     mid-cycle from hiding the defender's resolver.
   - `ResolveAssetCost` asks the buyer's own asset buckets (ID order) after its
     global and faction buckets.
   - The roll dispatcher reads the opponent's faction bucket. **Moved from Effort
     2.** It opens in Phase 5, after the Warlike and Fanatical Go hooks are gone:
     neither checks whose roll it is.
   - `FireTurnStart`, `FirePurchase` and `FireMatchup` walk the faction's tags
     after the asset entries. No registrar accepts `faction` on those triggers yet,
     so only tests reach the tag walk until an effort opens one.

5. **Scavengers is two entries with a new `caused_by` filter.** Robert's call.
   `owner` says whose asset was destroyed; `caused_by` says who destroyed it. Own
   losses: `owner = "own"`. Own kills: `owner = "rival", caused_by = "own"`. This
   closes `B.042` (D-2). The payout's cause becomes `coin_gain`.

6. **Fanatical's reroll is `result = 1`.** Robert's call. The `reroll` block takes
   exactly one of `dice` (the N lowest, or a forced rival's N highest) and `result`
   (every die showing that number).

7. **Fanatical's tie loss is a new effect, `lose_ties`, on a new trigger,
   `attack_tie`.** Planning call. The resolver answers `TieDefenderWins` when the
   owner attacks and `TieAttackerWins` when it defends, as the Go handler does.

8. **Preceptor Archive's discount is a new effect, `cost_discount`, with a new
   filter, `min_tech_level`.** Planning call. The cost never goes below 0.

9. **The government role is `government = true` on a tag record.** Planning call
   (discovery group E). The field `Rulebook.GovernmentTagID`, set by the loader,
   names it (Decision Record 4). The loader refuses two. A rulebook with none is valid: Planetary Seizure then changes the record
   only, and the wizard hides the governed-world pick.

10. **A held tag that the rulebook does not know registers nothing.** Planning
    call. As for an asset with no definition. The digest and the Manage detail view
    show the bare ID. A `state.toml` of the old shape fails to load (D-4).

11. **Six phases.** Robert's call. Phase 2 (sign-up) and Phase 3 (hook bodies) split
    the port at a compile seam. Each tag's move to data adds its entry and deletes
    its Go handler in one commit, so nothing fires twice.

12. **Robert's `campaigns/sandbox` is not edited.** Overview Shared Context and D-4.
    Claude re-seeds `campaigns/sandbox-claude` at the end of Phase 6.

13. **Phases 1 and 3 to 6 run 2 to 13% past the 800-line phase budget.** The
    excess is literal test code; a further split adds a session with no real seam.

## Decision Record — Execution

1. **Phase 1, Commit 1: `scratch/checkseeds/main.go` reads tag IDs.** Not in the
   plan. The directory is gitignored, but `go build ./...` compiles it, so the
   check gate fails until its tag loop ranges over IDs. A three-line change; not
   committed. Commit 2 does the same for its government check, which now reads
   `book.GovernmentTagID()`.
2. **Phase 5, Commit 9: Blackmail's Warlike test registers the data entry.** Not in
   the plan. `remove_tag_dice_test.go` (`TestRemoveTagDice_WarlikeAttackOnBlackmailRollsOneDie`)
   called `tags.WarlikeHandler{}.Apply`, so deleting the handler broke `go vet`. The
   test now gives beta the Warlike tag ID and registers Warlike's `bonus_die` entry
   (scope `faction`, `once_per = "owner_faction"`) on a tag holder. It checks the
   new budget key, `effect:bonus_die:own_roll:SWN-T-020`. The test still proves the
   same rule: Blackmail strips a tag's die from a roll against it.
3. **Phase 6, Commit 13: `engine/tag/tag.go` drops its `tags` import.** Not in the
   plan. With both handlers gone, `New` uses nothing from `tag/tags`, so the
   unused import fails `go build`. Commit 14 still deletes the whole package.
4. **Interim review: the government tag ID is a `Rulebook` field, not a method.**
   Robert's call. `GovernmentTagID()` sorted every tag key on each call, but
   `loadTags` already finds the one government tag. `loadTags` now returns the ID,
   and `Load` stores it in the exported field `GovernmentTagID`. An exported field,
   not an unexported one behind the method: Go does not allow a field and a method
   with one name, and test rulebooks built as literals (`goals/lock_test.go`) set
   the field directly. Commit 2's task code shows the method as first built.
   `scratch/checkseeds` reads the field (gitignored; not committed).
5. **Interim review: the Expand Influence free attack asks `ResolveTie`.** Robert's
   rules call. Fanatical's "always loses ties during attacks" covers the free
   Attack on a new Base. Before, `baseAttack.resolve` landed every tie with a
   hard-coded `>=`, so a Fanatical rival's tied free attack still hit. This was
   older than Effort 1, but `lose_ties` (Decision 7) made it reachable from data.
   The tie now asks both sides' resolvers as `attack.go` does: a Fanatical rival
   misses a tie; a Fanatical expanding faction still takes the hit.
6. **Interim review: `caused_by = "any"` stays valid.** Robert's call. It loads
   through `effectOwners` and reads as "no filter" in `onSide`, as `owner = "any"`
   does. Neither the loader nor `Accepts` refuses it.

---

## Shared Context

### Branch, cadence, and checks

- Branch `feature/data-driven-tags`. Commit per session. At the effort boundary:
  interim `/code-review` of the effort's diff, `fix:` commits, then a squash into
  one `effort 1` commit (overview Shared Context).
- Every commit builds and passes: `go build ./... && go vet ./... && go test ./...`
  from the repo root.
- Edit files only with Edit/Write, never `sed` or scripts.

### Holder vocabulary

Used in every phase from Phase 2 on. The type is defined in Phase 2, Commit 3,
Task 1; the hook-side `holderRef` in Phase 3, Commit 4, Task 1.

- **Holder** — who declared an entry: an asset through its definition, or a
  faction through a tag. A quality's purchase has a holder with a `Def` and no
  `Asset`.
- **Subject asset** — `FireContext.Asset` and `MatchupContext.Asset`: the asset
  the event is about (the bought asset, the quality's target, the fighting asset).
  For an asset holder at turn start it is the holder's own asset.

### Pre-merge additions (beyond the overview's)

- `planned-work.md`: add the follow-on row "registrar `faction`-scope sweep audit"
  (Decision 3).
- Graduate to Key Decisions: Decisions 1, 2 and 4 (hooks, effect-mutation), and
  Decision 9 (persistence, which replaces the `T-011` wording).

## Out of Scope

- The other 16 tags, and `faction` scope on any registrar not named in Decision 3.
- Mercenary Group's self-move mode (Effort 2) and lifting the two refusals in
  Decision 2.
- Robert's `campaigns/sandbox` files (`rulebook/tags.toml`, `state.seed.toml`,
  `scenarios/*.toml`, `TESTING.md`).

---

## Work Breakdown

**Overview**

- **Phase 1 — Tag IDs**
  - Commit 1 — `refactor: hold tag IDs on factions` (10 tasks)
  - Commit 2 — `feat: mark the government tag in rulebook data` (6 tasks)
- **Phase 2 — Holder contract**
  - Commit 3 — `refactor: name the effect holder in the effect contexts` (5 tasks)
- **Phase 3 — Holder-neutral hooks**
  - Commit 4 — `refactor: check roll and reactor hooks through their holder` (3 tasks)
  - Commit 5 — `refactor: check rule and lethal hooks through their holder` (3 tasks)
- **Phase 4 — Tag holder**
  - Commit 6 — `feat: load and check tag effects` (5 tasks)
  - Commit 7 — `feat: register and fire tag effects` (4 tasks)
  - Commit 8 — `feat: ask asset hooks at ties and on the buyer's cost` (3 tasks)
- **Phase 5 — Roll tags to data**
  - Commit 9 — `feat: Warlike as rulebook data` (5 tasks)
  - Commit 10 — `feat: Fanatical as rulebook data` (6 tasks)
  - Commit 11 — `feat: let the opponent's faction hooks reach a roll` (2 tasks)
- **Phase 6 — Economy tags to data**
  - Commit 12 — `feat: Preceptor Archive as rulebook data` (5 tasks)
  - Commit 13 — `fix: Scavengers pays only for its own kills and losses` (5 tasks)
  - Commit 14 — `refactor: delete the tag engine` (3 tasks)

### Phase 1 — Tag IDs

**In plain words:** a faction stores the IDs of its tags (`SWN-T-020`), not full
copies, so a rulebook change reaches every faction. The government role is a
`government = true` flag on a tag record, not a Go constant.

#### Commit 1 — `refactor: hold tag IDs on factions`

##### Task 1 — `rulebooks/swn/tags.toml`

Prefix every tag ID. Each of the 20 lines `id = "T-0NN"` (001 to 020) becomes
`id = "SWN-T-0NN"`. Example, the first and the last:

**Find:**
```toml
[tags.Colonists]
id = "T-001"
```
**Replace with:**
```toml
[tags.Colonists]
id = "SWN-T-001"
```

**Find:**
```toml
[tags.Warlike]
id = "T-020"
```
**Replace with:**
```toml
[tags.Warlike]
id = "SWN-T-020"
```

##### Task 2 — `internal/faction/domain/faction.go`

`Tag` is rulebook data only now, so it drops its struct tags. Factions hold IDs.

**Find:**
```go
type Tag struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Effect      string `toml:"effect"`
}

// TagPlanetaryGovernment marks a faction that governs at least one world; the
// government record says which (A.001.6 overview Decision 10).
const TagPlanetaryGovernment = "T-011"
```
**Replace with:**
```go
type Tag struct {
	ID          string
	Name        string
	Description string
	Effect      string
}

// TagPlanetaryGovernment marks a faction that governs at least one world; the
// government record says which (A.001.6 overview Decision 10).
const TagPlanetaryGovernment = "SWN-T-011"
```

**Find:**
```go
	Tags                    []*Tag            `toml:"tags"`
```
**Replace with:**
```go
	Tags                    []string          `toml:"tags"`
```

**Find:**
```go
// HasTag reports whether the faction holds the tag with ID tagID.
func (faction *Faction) HasTag(tagID string) bool {
	return slices.ContainsFunc(faction.Tags, func(tag *Tag) bool { return tag.ID == tagID })
}
```
**Replace with:**
```go
// HasTag reports whether the faction holds the tag with ID tagID.
func (faction *Faction) HasTag(tagID string) bool {
	return slices.Contains(faction.Tags, tagID)
}
```

**Find:**
```go
	if faction.Tags != nil {
		clone.Tags = make([]*Tag, len(faction.Tags))
		for i, tag := range faction.Tags {
			tagCopy := *tag
			clone.Tags[i] = &tagCopy
		}
	}
```
**Replace with:**
```go
	clone.Tags = slices.Clone(faction.Tags)
```

**Find:**
```go
	tags []*Tag,
	goal *Goal,
```
**Replace with:**
```go
	tagIDs []string,
	goal *Goal,
```

**Find:**
```go
		Tags:      tags,
		Coin:      coin,
```
**Replace with:**
```go
		Tags:      tagIDs,
		Coin:      coin,
```

##### Task 3 — `internal/faction/domain/mutation.go`

**Find:**
```go
// TagAdded grants a tag to a faction. The full Tag is embedded so Apply does
// not need a Rulebook lookup.
type TagAdded struct {
	FactionID         string `json:"faction_id"`
	Tag               Tag    `json:"tag"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m TagAdded) Type() string { return "tag_added" }

// TagRemoved takes one tag from a faction. The full Tag rides along, as on
// TagAdded, so the digest can name it after the faction no longer holds it.
type TagRemoved struct {
	FactionID         string `json:"faction_id"`
	Tag               Tag    `json:"tag"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}
```
**Replace with:**
```go
// TagAdded grants a tag to a faction. A reader that needs the name reads the
// rulebook (A.001.7 overview Decision 4).
type TagAdded struct {
	FactionID         string `json:"faction_id"`
	TagID             string `json:"tag_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m TagAdded) Type() string { return "tag_added" }

// TagRemoved takes one tag from a faction.
type TagRemoved struct {
	FactionID         string `json:"faction_id"`
	TagID             string `json:"tag_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}
```

##### Task 4 — `internal/faction/engine/mutation/mutation.go`

**Find:**
```go
		case domain.TagAdded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				tag := v.Tag
				faction.Tags = append(faction.Tags, &tag)
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.TagRemoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				index := slices.IndexFunc(faction.Tags, func(tag *domain.Tag) bool { return tag.ID == v.Tag.ID })
				if index >= 0 {
					faction.Tags = slices.Delete(faction.Tags, index, index+1)
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.Tag.ID})
				}
```
**Replace with:**
```go
		case domain.TagAdded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.Tags = append(faction.Tags, v.TagID)
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.TagRemoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				index := slices.Index(faction.Tags, v.TagID)
				if index >= 0 {
					faction.Tags = slices.Delete(faction.Tags, index, index+1)
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.TagID})
				}
```

##### Task 5 — `internal/faction/engine/goal/goals/planetary_seizure.go`

In `takeGovernment`, the mutations carry the ID. Commit 2 replaces the constant.

**Find:**
```go
	if !seizer.HasTag(domain.TagPlanetaryGovernment) {
		var tag domain.Tag
		if found, ok := rulebook.Tags[domain.TagPlanetaryGovernment]; ok {
			tag = *found
		}
		mutations = append(mutations, domain.TagAdded{FactionID: seizer.ID, Tag: tag, Cause: "goal_completed", CausedByFactionID: seizer.ID})
	}
	previous := factionState.Factions[previousRuler]
	if previous == nil || !slices.Equal(factionState.WorldsRuledBy(previousRuler), []string{worldID}) {
		return mutations
	}
	for _, held := range previous.Tags {
		if held.ID == domain.TagPlanetaryGovernment {
			return append(mutations, domain.TagRemoved{FactionID: previousRuler, Tag: *held, Cause: "goal_completed", CausedByFactionID: seizer.ID})
		}
	}
	return mutations
}
```
**Replace with:**
```go
	if !seizer.HasTag(domain.TagPlanetaryGovernment) {
		mutations = append(mutations, domain.TagAdded{FactionID: seizer.ID, TagID: domain.TagPlanetaryGovernment, Cause: "goal_completed", CausedByFactionID: seizer.ID})
	}
	previous := factionState.Factions[previousRuler]
	if previous == nil || !previous.HasTag(domain.TagPlanetaryGovernment) || !slices.Equal(factionState.WorldsRuledBy(previousRuler), []string{worldID}) {
		return mutations
	}
	return append(mutations, domain.TagRemoved{FactionID: previousRuler, TagID: domain.TagPlanetaryGovernment, Cause: "goal_completed", CausedByFactionID: seizer.ID})
}
```

##### Task 6 — the four Go handlers (`internal/faction/engine/tag/tags/`)

They keep working until Phases 5 and 6 delete them.

| File | Find | Replace with |
|---|---|---|
| `fanatical.go` | `const FanaticalTagID = "T-005"` | `const FanaticalTagID = "SWN-T-005"` |
| `preceptor_archive.go` | `const PreceptorArchiveTagID = "T-013"` | `const PreceptorArchiveTagID = "SWN-T-013"` |
| `scavengers.go` | `const ScavengersTagID = "T-016"` | `const ScavengersTagID = "SWN-T-016"` |
| `warlike.go` | `const WarlikeTagID = "T-020"` | `const WarlikeTagID = "SWN-T-020"` |

In `internal/faction/engine/tag/tag.go`, `ApplyAll` ranges over IDs:

**Find:**
```go
		for _, tag := range faction.Tags {
			handler, ok := e.handlers[tag.ID]
```
**Replace with:**
```go
		for _, tagID := range faction.Tags {
			handler, ok := e.handlers[tagID]
```

##### Task 7 — `internal/faction/narrative/digest/`

`resolve.go` — after `resolveGoalName`, add:

```go
func resolveTagName(tagID string, rulebook *rulebook.Rulebook) string {
	if rulebook == nil {
		return tagID
	}
	if tag, ok := rulebook.Tags[tagID]; ok {
		return tag.Name
	}
	return tagID
}
```

`build.go`:

**Find:**
```go
				beat.GoalEvents = append(beat.GoalEvents, GoalEvent{
					Kind:    GoalTagGained,
					TagName: m.Tag.Name,
				})
```
**Replace with:**
```go
				beat.GoalEvents = append(beat.GoalEvents, GoalEvent{
					Kind:    GoalTagGained,
					TagName: resolveTagName(m.TagID, rulebook),
				})
```

**Find:**
```go
				loser.GoalEvents = append(loser.GoalEvents, GoalEvent{Kind: GoalTagLost, TagName: m.Tag.Name})
```
**Replace with:**
```go
				loser.GoalEvents = append(loser.GoalEvents, GoalEvent{Kind: GoalTagLost, TagName: resolveTagName(m.TagID, rulebook)})
```

##### Task 8 — `internal/faction/tui/views/manage/detail/detail.go`

**Find:**
```go
	for _, tag := range f.Tags {
		b.WriteString("  " + styles.Strong.Render(tag.Name) + "\n")
```
**Replace with:**
```go
	for _, tagID := range f.Tags {
		tag, known := m.rulebook.Tags[tagID]
		if !known {
			b.WriteString("  " + styles.Strong.Render(tagID) + "\n")
			continue
		}
		b.WriteString("  " + styles.Strong.Render(tag.Name) + "\n")
```

##### Task 9 — `internal/faction/tui/views/manage/wizard/wizard.go`

The picker keeps rulebook pointers; the faction gets their IDs.

**Find:**
```go
		m.data.selectedTags,
		m.data.selectedGoal,
```
**Replace with:**
```go
		tagIDs(m.data.selectedTags),
		m.data.selectedGoal,
```

After `tagsToOptions`, add:

```go
func tagIDs(tags []*domain.Tag) []string {
	ids := make([]string, len(tags))
	for i, tag := range tags {
		ids[i] = tag.ID
	}
	return ids
}
```

##### Task 10 — tests

Tests name the government tag by its literal ID, so Commit 2 can delete the
constant without touching them again.

`internal/faction/domain/faction_test.go`:

**Find:**
```go
	goal := &domain.Goal{ID: "goal-expand", Name: "Expand Influence"}
	tag := &domain.Tag{ID: "tag-warlike", Name: "Warlike"}

	faction := domain.NewFaction(
		"test-faction", "Test Faction",
		domain.ScaleMinor,
		domain.StatForce, domain.StatCunning, domain.StatWealth,
		[]*domain.Tag{tag},
```
**Replace with:**
```go
	goal := &domain.Goal{ID: "goal-expand", Name: "Expand Influence"}

	faction := domain.NewFaction(
		"test-faction", "Test Faction",
		domain.ScaleMinor,
		domain.StatForce, domain.StatCunning, domain.StatWealth,
		[]string{"SWN-T-020"},
```

**Find:**
```go
	primaryRating, secondaryRating, tertiaryRating := domain.RatingsFromScale(domain.ScaleMinor)
```
**Replace with:**
```go
	if !faction.HasTag("SWN-T-020") || len(faction.Tags) != 1 {
		t.Errorf("Tags = %v, want [SWN-T-020]", faction.Tags)
	}

	primaryRating, secondaryRating, tertiaryRating := domain.RatingsFromScale(domain.ScaleMinor)
```

`internal/faction/engine/mutation/mutation_test.go`:

**Find:**
```go
	s.Factions["f1"].Tags = []*domain.Tag{{ID: "T-001"}, {ID: domain.TagPlanetaryGovernment}}
	me := New(slog.New(slog.DiscardHandler))
	removal := domain.TagRemoved{FactionID: "f1", Tag: domain.Tag{ID: domain.TagPlanetaryGovernment}}

	if err := me.Apply(s, []domain.Mutation{removal}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if tags := s.Factions["f1"].Tags; len(tags) != 1 || tags[0].ID != "T-001" {
		t.Errorf("Tags = %v, want [T-001]", tags)
	}
	err := me.Apply(s, []domain.Mutation{removal}, slog.New(slog.DiscardHandler))
	if err == nil || len(err.Misses) != 1 || err.Misses[0].EntityID != domain.TagPlanetaryGovernment {
		t.Errorf("removing a tag not held = %v, want one miss naming %s", err, domain.TagPlanetaryGovernment)
	}
}
```
**Replace with:**
```go
	s.Factions["f1"].Tags = []string{"SWN-T-001", "SWN-T-011"}
	me := New(slog.New(slog.DiscardHandler))
	removal := domain.TagRemoved{FactionID: "f1", TagID: "SWN-T-011"}

	if err := me.Apply(s, []domain.Mutation{removal}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if tags := s.Factions["f1"].Tags; len(tags) != 1 || tags[0] != "SWN-T-001" {
		t.Errorf("Tags = %v, want [SWN-T-001]", tags)
	}
	err := me.Apply(s, []domain.Mutation{removal}, slog.New(slog.DiscardHandler))
	if err == nil || len(err.Misses) != 1 || err.Misses[0].EntityID != "SWN-T-011" {
		t.Errorf("removing a tag not held = %v, want one miss naming SWN-T-011", err)
	}
}

func TestMutationEngine_TagAdded(t *testing.T) {
	s := newMutationTestState()
	me := New(slog.New(slog.DiscardHandler))

	if err := me.Apply(s, []domain.Mutation{domain.TagAdded{FactionID: "f1", TagID: "SWN-T-020"}}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !s.Factions["f1"].HasTag("SWN-T-020") {
		t.Errorf("Tags = %v, want SWN-T-020 held", s.Factions["f1"].Tags)
	}
}
```

`internal/faction/engine/goal/goals/lock_test.go`:

**Find:**
```go
			"T-011": {ID: "T-011", Name: "Planetary Government"},
```
**Replace with:**
```go
			"SWN-T-011": {ID: "SWN-T-011", Name: "Planetary Government"},
```

**Find:**
```go
	rival := &domain.Faction{ID: "f2", Force: 4, Cunning: 4, Wealth: 4,
		Tags: []*domain.Tag{{ID: domain.TagPlanetaryGovernment, Name: "Planetary Government"}}}
```
**Replace with:**
```go
	rival := &domain.Faction{ID: "f2", Force: 4, Cunning: 4, Wealth: 4, Tags: []string{"SWN-T-011"}}
```

**Find:**
```go
	tag := domain.Tag{ID: domain.TagPlanetaryGovernment, Name: "Planetary Government"}
	changed := func(from string) domain.WorldGovernmentChanged {
		return domain.WorldGovernmentChanged{WorldID: "Krylos", FromFactionID: from, ToFactionID: "f1", Cause: "goal_completed", CausedByFactionID: "f1"}
	}
	gained := domain.TagAdded{FactionID: "f1", Tag: tag, Cause: "goal_completed", CausedByFactionID: "f1"}
```
**Replace with:**
```go
	const tagID = "SWN-T-011"
	changed := func(from string) domain.WorldGovernmentChanged {
		return domain.WorldGovernmentChanged{WorldID: "Krylos", FromFactionID: from, ToFactionID: "f1", Cause: "goal_completed", CausedByFactionID: "f1"}
	}
	gained := domain.TagAdded{FactionID: "f1", TagID: tagID, Cause: "goal_completed", CausedByFactionID: "f1"}
```

**Find:**
```go
				domain.TagRemoved{FactionID: "f2", Tag: tag, Cause: "goal_completed", CausedByFactionID: "f1"}}, completed(2)...),
```
**Replace with:**
```go
				domain.TagRemoved{FactionID: "f2", TagID: tagID, Cause: "goal_completed", CausedByFactionID: "f1"}}, completed(2)...),
```

**Find:**
```go
				faction.Tags = []*domain.Tag{&tag}
```
**Replace with:**
```go
				faction.Tags = []string{tagID}
```

`internal/faction/state/faction_state_test.go`:

**Find:**
```go
				Tags: []*domain.Tag{
					{ID: "T-001", Name: "Colonists", Description: "Settler faction", Effect: "+1 Wealth"},
				},
```
**Replace with:**
```go
				Tags: []string{"SWN-T-001"},
```

**Find:**
```go
	if len(got.Tags) != 1 {
		t.Fatalf("Faction.Tags: got %d, want 1", len(got.Tags))
	}
	if got.Tags[0].ID != want.Tags[0].ID || got.Tags[0].Name != want.Tags[0].Name {
		t.Errorf("Faction.Tags[0]: got %+v, want %+v", got.Tags[0], want.Tags[0])
	}
```
**Replace with:**
```go
	if !reflect.DeepEqual(got.Tags, want.Tags) {
		t.Errorf("Faction.Tags: got %v, want %v", got.Tags, want.Tags)
	}
```

After `TestLoadMissingFile`, add (D-4: the old shape fails loud):

```go
func TestLoad_RefusesTagCopies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.toml")
	body := "[factions.alpha]\nid = \"alpha\"\ntags = [{ id = \"T-011\", name = \"Planetary Government\" }]\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "tags") {
		t.Fatalf("Load = %v, want an error naming the tags key", err)
	}
}
```

`internal/faction/narrative/digest/digest_test.go`:

**Find:**
```go
			mutRec(t, "tag_added", domain.TagAdded{
				FactionID: "faction-a",
				Tag:       domain.Tag{ID: "tag-1", Name: "Guerrilla Tactics"},
				Cause:     "goal_completed",
			}),
		),
	}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), nil)
```
**Replace with:**
```go
			mutRec(t, "tag_added", domain.TagAdded{
				FactionID: "faction-a",
				TagID:     "tag-1",
				Cause:     "goal_completed",
			}),
		),
	}
	rb := &rulebook.Rulebook{Tags: map[string]*domain.Tag{"tag-1": {ID: "tag-1", Name: "Guerrilla Tactics"}}}
	d := buildOne(t, records, minState(map[string]string{"faction-a": "Alpha"}), rb)
```

**Find:**
```go
			mutRec(t, "tag_removed", domain.TagRemoved{
				FactionID: "f2", Tag: domain.Tag{ID: domain.TagPlanetaryGovernment, Name: "Planetary Government"},
				Cause: "goal_completed", CausedByFactionID: "f1",
			}),
		),
	}
	cycleDigest := buildOne(t, records, factionState, nil)
```
**Replace with:**
```go
			mutRec(t, "tag_removed", domain.TagRemoved{
				FactionID: "f2", TagID: "SWN-T-011",
				Cause: "goal_completed", CausedByFactionID: "f1",
			}),
		),
	}
	// deliberate: no rulebook, so the line falls back to the bare ID (Effort 1 Decision 10).
	cycleDigest := buildOne(t, records, factionState, nil)
```

**Find:**
```go
		return event.Kind == GoalTagLost && event.TagName == "Planetary Government"
```
**Replace with:**
```go
		return event.Kind == GoalTagLost && event.TagName == "SWN-T-011"
```

`internal/faction/rulebook/rulebook_test.go`:

**Find:**
```go
		tag, ok := rb.Tags["T-001"]
		if !ok {
			t.Fatal("expected tag T-001 (Colonists) to be present")
		}
```
**Replace with:**
```go
		tag, ok := rb.Tags["SWN-T-001"]
		if !ok {
			t.Fatal("expected tag SWN-T-001 (Colonists) to be present")
		}
```

`internal/faction/engine/testharness/scenarios/tags_test.go` — five fixtures:

| Find | Replace with |
|---|---|
| `alpha.Tags = []*domain.Tag{{ID: "T-016"}}` | `alpha.Tags = []string{"SWN-T-016"}` |
| `alpha.Tags = []*domain.Tag{{ID: "T-020"}}` | `alpha.Tags = []string{"SWN-T-020"}` |
| `alpha.Tags = []*domain.Tag{{ID: "T-005"}}` (two tests; Edit each) | `alpha.Tags = []string{"SWN-T-005"}` |
| `alpha.Tags = []*domain.Tag{{ID: "T-013"}}` | `alpha.Tags = []string{"SWN-T-013"}` |

##### Commit message

```
refactor: hold tag IDs on factions

- Faction.Tags is a list of tag IDs; TagAdded and TagRemoved carry the ID
- SWN tags take the SWN- prefix (SWN-T-001 to SWN-T-020)
- the digest and the Manage detail view read tag names from the rulebook
- a state.toml that holds tag copies fails to load (D-4)
```

#### Commit 2 — `feat: mark the government tag in rulebook data`

##### Task 1 — `internal/faction/domain/faction.go`

**Find:**
```go
type Tag struct {
	ID          string
	Name        string
	Description string
	Effect      string
}

// TagPlanetaryGovernment marks a faction that governs at least one world; the
// government record says which (A.001.6 overview Decision 10).
const TagPlanetaryGovernment = "SWN-T-011"
```
**Replace with:**
```go
// Tag is one rulebook tag. Government marks the tag a faction holds while it
// governs at least one world; the government record says which (A.001.6
// overview Decision 10).
type Tag struct {
	ID          string
	Name        string
	Description string
	Effect      string
	Government  bool
}
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

Add `"maps"` to the imports.

**Find:**
```go
// Load reads all static data files from dataDir and returns a populated Rulebook.
```
**Replace with:**
```go
// GovernmentTagID names the tag with government = true, or "" when the rulebook
// has none. The loader refuses two.
func (rulebook *Rulebook) GovernmentTagID() string {
	for _, tagID := range slices.Sorted(maps.Keys(rulebook.Tags)) {
		if rulebook.Tags[tagID].Government {
			return tagID
		}
	}
	return ""
}

// Load reads all static data files from dataDir and returns a populated Rulebook.
```

**Find:**
```go
type tagRecord struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Effect      string `toml:"effect"`
}
```
**Replace with:**
```go
type tagRecord struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Effect      string `toml:"effect"`
	Government  bool   `toml:"government"`
}
```

**Find:**
```go
	result := make(map[string]*domain.Tag, len(f.Tags))
	for _, r := range f.Tags {
		if _, exists := result[r.ID]; exists {
			return nil, fmt.Errorf("duplicate tag ID %q in %s", r.ID, path)
		}
		result[r.ID] = &domain.Tag{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Effect:      r.Effect,
		}
	}
	return result, nil
```
**Replace with:**
```go
	result := make(map[string]*domain.Tag, len(f.Tags))
	government := ""
	for _, r := range f.Tags {
		if _, exists := result[r.ID]; exists {
			return nil, fmt.Errorf("duplicate tag ID %q in %s", r.ID, path)
		}
		if r.Government {
			if government != "" {
				return nil, fmt.Errorf("tags %q and %q both set government in %s", government, r.ID, path)
			}
			government = r.ID
		}
		result[r.ID] = &domain.Tag{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Effect:      r.Effect,
			Government:  r.Government,
		}
	}
	return result, nil
```

##### Task 3 — `rulebooks/swn/tags.toml`

**Find:**
```toml
[tags.PlanetaryGovernment]
id = "SWN-T-011"
name = "Planetary Government"
```
**Replace with:**
```toml
[tags.PlanetaryGovernment]
id = "SWN-T-011"
name = "Planetary Government"
government = true
```

##### Task 4 — `internal/faction/engine/goal/goals/planetary_seizure.go`

**Find:**
```go
// takeGovernment makes seizer the ruler of worldID (overview Decision 10): the
// record changes hands, the seizer gains T-011 unless it holds it, and a previous
// ruler that governs no other world loses the tag it holds. A seizer that already
// rules the world changes nothing.
```
**Replace with:**
```go
// takeGovernment makes seizer the ruler of worldID (overview Decision 10): the
// record changes hands, the seizer gains the government tag unless it holds it,
// and a previous ruler that governs no other world loses it. A seizer that already
// rules the world changes nothing. A rulebook with no government tag changes the
// record only.
```

**Find:**
```go
	if !seizer.HasTag(domain.TagPlanetaryGovernment) {
		mutations = append(mutations, domain.TagAdded{FactionID: seizer.ID, TagID: domain.TagPlanetaryGovernment, Cause: "goal_completed", CausedByFactionID: seizer.ID})
	}
	previous := factionState.Factions[previousRuler]
	if previous == nil || !previous.HasTag(domain.TagPlanetaryGovernment) || !slices.Equal(factionState.WorldsRuledBy(previousRuler), []string{worldID}) {
		return mutations
	}
	return append(mutations, domain.TagRemoved{FactionID: previousRuler, TagID: domain.TagPlanetaryGovernment, Cause: "goal_completed", CausedByFactionID: seizer.ID})
}
```
**Replace with:**
```go
	governmentTagID := rulebook.GovernmentTagID()
	if governmentTagID == "" {
		return mutations
	}
	if !seizer.HasTag(governmentTagID) {
		mutations = append(mutations, domain.TagAdded{FactionID: seizer.ID, TagID: governmentTagID, Cause: "goal_completed", CausedByFactionID: seizer.ID})
	}
	previous := factionState.Factions[previousRuler]
	if previous == nil || !previous.HasTag(governmentTagID) || !slices.Equal(factionState.WorldsRuledBy(previousRuler), []string{worldID}) {
		return mutations
	}
	return append(mutations, domain.TagRemoved{FactionID: previousRuler, TagID: governmentTagID, Cause: "goal_completed", CausedByFactionID: seizer.ID})
}
```

##### Task 5 — `internal/faction/tui/views/manage/wizard/wizard.go`

**Find:**
```go
func pickedGovernmentTag(tags []*domain.Tag) bool {
	return slices.ContainsFunc(tags, func(tag *domain.Tag) bool { return tag.ID == domain.TagPlanetaryGovernment })
}
```
**Replace with:**
```go
func pickedGovernmentTag(tags []*domain.Tag) bool {
	return slices.ContainsFunc(tags, func(tag *domain.Tag) bool { return tag.Government })
}
```

##### Task 6 — tests

`internal/faction/engine/goal/goals/lock_test.go`:

**Find:**
```go
			"SWN-T-011": {ID: "SWN-T-011", Name: "Planetary Government"},
```
**Replace with:**
```go
			"SWN-T-011": {ID: "SWN-T-011", Name: "Planetary Government", Government: true},
```

After `TestPlanetarySeizure_CompletionTakesTheGovernment`, add:

```go
func TestPlanetarySeizure_NoGovernmentTagChangesTheRecordOnly(t *testing.T) {
	factionState, faction := makeOccupationState()
	factionState.Governments = map[string]string{"Krylos": "f2"}

	_, mutations := PlanetarySeizure{}.CheckLock(faction, factionState, &rulebook.Rulebook{})

	for _, mutation := range mutations {
		switch mutation.(type) {
		case domain.TagAdded, domain.TagRemoved:
			t.Errorf("mutation %+v, want no tag change without a government tag", mutation)
		}
	}
}
```

`internal/faction/rulebook/rulebook_test.go` — in `TestLoadTagsAndGoals_Strict`, add a case after "tag duplicate ID":

```go
		{
			name: "two government tags",
			body: "[tags.Alpha]\nid = \"T-001\"\ngovernment = true\n\n[tags.Beta]\nid = \"T-002\"\ngovernment = true\n",
			load: loadTagFile,
			want: "both set government",
		},
```

In the shipped-rulebook test, after the "tags populated" subtest, add:

```go
	t.Run("government tag", func(t *testing.T) {
		if got := rb.GovernmentTagID(); got != "SWN-T-011" {
			t.Errorf("GovernmentTagID = %q, want SWN-T-011", got)
		}
	})
```

##### Commit message

```
feat: mark the government tag in rulebook data

- a tag record may set government = true; the loader refuses two
- Rulebook.GovernmentTagID replaces domain.TagPlanetaryGovernment
- Planetary Seizure and the wizard read the flag; a rulebook with none
  changes the government record only
```

### Phase 2 — Holder contract

**In plain words:** the three effect contexts name the entry's holder in one
`Holder` value, in place of loose faction, asset and definition fields. Nothing
changes in play. Phase 3 then moves each hook onto the holder.

#### Commit 3 — `refactor: name the effect holder in the effect contexts`

##### Task 1 — `internal/faction/engine/effect/registrar.go`

**Find:**
```go
// RegisterContext carries Roller because a Cat 3 reactor's OnMutations takes none;
// a reactor that rolls dice keeps this one.
type RegisterContext struct {
	Faction  *domain.Faction
	Asset    *domain.Asset
	Def      *domain.AssetDefinition
	Entry    domain.EffectDefinition
	Registry *hooks.Registry
	Roller   domain.Roller
}

// FireContext carries FactionState because a phase effect may reach other
// factions' assets (Panopticon Matrix tests every rival on its world).
type FireContext struct {
	Faction      *domain.Faction
	Asset        *domain.Asset
	Def          *domain.AssetDefinition
	Entry        domain.EffectDefinition
	Roller       domain.Roller
	FactionState *state.FactionState
}
```
**Replace with:**
```go
// Holder is who declared an [[effects]] entry: an asset through its definition,
// or a faction through a tag (Tag set, Asset and Def nil). A registrar reads its
// holder only through this value, so one registrar serves both. A quality's
// purchase has a Def and no Asset.
type Holder struct {
	Faction *domain.Faction
	Asset   *domain.Asset
	Def     *domain.AssetDefinition
	Tag     *domain.Tag
}

// ID names the source in a hook ID and a budget key: the definition ID, or the
// tag ID.
func (holder Holder) ID() string {
	if holder.Tag != nil {
		return holder.Tag.ID
	}
	return holder.Def.ID
}

// Name is the source's display name, for an offer's Source.
func (holder Holder) Name() string {
	if holder.Tag != nil {
		return holder.Tag.Name
	}
	return holder.Def.Name
}

// Scope is the registry bucket the holder's hooks go in: its asset's, or its
// faction's for a tag.
func (holder Holder) Scope() hooks.Scope {
	if holder.Tag != nil {
		return hooks.FactionScope(holder.Faction.ID)
	}
	return hooks.AssetScope(holder.Faction.ID, holder.Asset.ID)
}

// HookID names a registered hook for diagnostics.
func (holder Holder) HookID(effectName domain.SpecialEffectType) string {
	return string(effectName) + ":" + holder.ID()
}

// RegisterContext carries Roller because a Cat 3 reactor's OnMutations takes none;
// a reactor that rolls dice keeps this one.
type RegisterContext struct {
	Holder   Holder
	Entry    domain.EffectDefinition
	Registry *hooks.Registry
	Roller   domain.Roller
}

// FireContext carries FactionState because a phase effect may reach other
// factions' assets (Panopticon Matrix tests every rival on its world). Asset is
// the subject: the bought asset, a quality's target, or at turn start the
// holder's own asset.
type FireContext struct {
	Holder       Holder
	Asset        *domain.Asset
	Entry        domain.EffectDefinition
	Roller       domain.Roller
	FactionState *state.FactionState
}
```

**Find:**
```go
// MatchupContext is one matchup of the Attack action, seen from the declaring
// asset: its attacker for own_attack, its defender for own_defense_win.
type MatchupContext struct {
	Faction       *domain.Faction
	Asset         *domain.Asset
	Def           *domain.AssetDefinition
	Entry         domain.EffectDefinition
```
**Replace with:**
```go
// MatchupContext is one matchup of the Attack action, seen from one side. Asset is
// that side's fighting asset: its attacker for own_attack, its defender for
// own_defense_win. Holder declared the entry.
type MatchupContext struct {
	Holder        Holder
	Asset         *domain.Asset
	Entry         domain.EffectDefinition
```

In `effectsListHandler.Apply`:

**Find:**
```go
		handler.registrars[entry.Effect].Register(RegisterContext{
			Faction:  faction,
			Asset:    asset,
			Def:      handler.def,
			Entry:    entry,
```
**Replace with:**
```go
		handler.registrars[entry.Effect].Register(RegisterContext{
			Holder:   Holder{Faction: faction, Asset: asset, Def: handler.def},
			Entry:    entry,
```

##### Task 2 — `internal/faction/engine/effect/effect.go`

In `FireTurnStart`:

**Find:**
```go
			mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
				Faction:      faction,
				Asset:        asset,
				Def:          def,
				Entry:        entry,
```
**Replace with:**
```go
			mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
				Holder:       Holder{Faction: faction, Asset: asset, Def: def},
				Asset:        asset,
				Entry:        entry,
```

In `FirePurchase`:

**Find:**
```go
		mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
			Faction:      faction,
			Asset:        asset,
			Def:          def,
			Entry:        entry,
```
**Replace with:**
```go
		mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
			Holder:       Holder{Faction: faction, Asset: asset, Def: def},
			Asset:        asset,
			Entry:        entry,
```

In `FireQuality`:

**Find:**
```go
// direct definition read raised by Buy Asset: a quality adds no asset, so
// FirePurchase never sees it. FireContext.Asset is the target, not a declaring asset.
func (e *EffectsEngine) FireQuality(faction *domain.Faction, def *domain.AssetDefinition, target *domain.Asset, factionState *state.FactionState) []domain.Mutation {
	entry := def.Effects[0]
	// deliberate: no roller, because no quality effect rolls; give FireQuality one when one does.
	return e.registrars[entry.Effect].Fire(FireContext{
		Faction:      faction,
		Asset:        target,
		Def:          def,
		Entry:        entry,
```
**Replace with:**
```go
// direct definition read raised by Buy Asset: a quality adds no asset, so
// FirePurchase never sees it. The holder has no asset; FireContext.Asset is the target.
func (e *EffectsEngine) FireQuality(faction *domain.Faction, def *domain.AssetDefinition, target *domain.Asset, factionState *state.FactionState) []domain.Mutation {
	entry := def.Effects[0]
	// deliberate: no roller, because no quality effect rolls; give FireQuality one when one does.
	return e.registrars[entry.Effect].Fire(FireContext{
		Holder:       Holder{Faction: faction, Def: def},
		Asset:        target,
		Entry:        entry,
```

In `FireMatchup`:

**Find:**
```go
// FireMatchup runs the declaring asset's entries for trigger through each
// registrar's matchup route. A direct definition read, raised by the Attack
// action, which owns the matchup.
func (e *EffectsEngine) FireMatchup(trigger domain.EffectTrigger, ctx MatchupContext) MatchupOutcome {
	var outcome MatchupOutcome
	for _, entry := range ctx.Def.Effects {
```
**Replace with:**
```go
// FireMatchup runs the holder's entries for trigger through each registrar's
// matchup route. A direct definition read, raised by the Attack action, which
// owns the matchup.
func (e *EffectsEngine) FireMatchup(trigger domain.EffectTrigger, ctx MatchupContext) MatchupOutcome {
	var outcome MatchupOutcome
	for _, entry := range ctx.Holder.Def.Effects {
```

##### Task 3 — `internal/faction/engine/effect/effects/*.go` (non-test)

A rename the compiler checks. Apply it to every error `go build ./...` reports in
this package, and nowhere else:

| In | Old | New |
|---|---|---|
| any `RegisterContext`, `FireContext` or `MatchupContext` value | `ctx.Faction` | `ctx.Holder.Faction` |
| a `RegisterContext` value (the `register…` functions) | `ctx.Asset` | `ctx.Holder.Asset` |
| a `RegisterContext` value | `ctx.Def` | `ctx.Holder.Def` |

`ctx.Asset` in a `Fire` or `Matchup` function stays: it is the subject asset.
`hooks.RollContext` values (`rollMatches`) are not touched. The files that change:
`apply_stealth.go`, `block_ability.go`, `bonus_die.go`, `capture_target.go`,
`clear_stealth.go`, `coin_gain.go`, `coin_steal.go`, `cost_surcharge.go`,
`discount_repair.go`, `grant_movement.go`, `grant_permission.go`, `grant_stealth.go`,
`ignore_coin_loss.go`, `lock_out.go`, `remove_tag_dice.go`, `reroll.go`,
`restore_asset.go`, `restrict_attack.go`, `sacrifice_self.go`, `seize_target.go`,
`self_damage.go`, `shelter_asset.go`, `tech_level_floor.go`,
`withdraw_permission.go`. Phase 3 rewrites each `register…` function in full.

##### Task 4 — `internal/faction/engine/action/actions/attack.go`

**Find:**
```go
	return attack.effects.FireMatchup(trigger, effect.MatchupContext{
		Faction:       faction,
		Asset:         asset,
		Def:           def,
		Opponent:      opponent,
```
**Replace with:**
```go
	return attack.effects.FireMatchup(trigger, effect.MatchupContext{
		Holder:        effect.Holder{Faction: faction, Asset: asset, Def: def},
		Asset:         asset,
		Opponent:      opponent,
```

##### Task 5 — tests

Every `RegisterContext` literal wraps its three holder fields. The rows are exact
one-line replacements:

| File (line today) | Find | Replace with |
|---|---|---|
| `effects/bonus_die_test.go` (65) | `Faction: alpha, Asset: alpha.Assets[assetID], Def: def, Entry: entry, Registry: registry,` | `Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets[assetID], Def: def}, Entry: entry, Registry: registry,` |
| `effects/capture_target_test.go` (23) | `Faction: alpha, Asset: asset, Def: &domain.AssetDefinition{ID: "SWN-C7-003"}, Registry: registry, Roller: fixedRoller{value: 1},` | `Holder: effect.Holder{Faction: alpha, Asset: asset, Def: &domain.AssetDefinition{ID: "SWN-C7-003"}}, Registry: registry, Roller: fixedRoller{value: 1},` |
| `effects/clear_stealth_test.go` (58, 240: Edit each) | `Faction: alpha, Asset: alpha.Assets["cells"], Def: &domain.AssetDefinition{ID: "SWN-C4-003"}, Registry: registry, Roller: roller,` | `Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["cells"], Def: &domain.AssetDefinition{ID: "SWN-C4-003"}}, Registry: registry, Roller: roller,` |
| `effects/coin_steal_test.go` (56) | `Faction: alpha, Asset: asset, Def: def, Entry: def.Effects[0], Registry: registry, Roller: roller,` | `Holder: effect.Holder{Faction: alpha, Asset: asset, Def: def}, Entry: def.Effects[0], Registry: registry, Roller: roller,` |
| `effects/cost_surcharge_test.go` (52) | `Faction: alpha, Asset: asset, Def: localInvestmentsDef, Entry: entry, Registry: registry,` | `Holder: effect.Holder{Faction: alpha, Asset: asset, Def: localInvestmentsDef}, Entry: entry, Registry: registry,` |
| `effects/discount_repair_test.go` (18) | `Faction: owner, Asset: medicalCenter, Def: &domain.AssetDefinition{ID: "SWN-W4-003"}, Registry: registry,` | `Holder: effect.Holder{Faction: owner, Asset: medicalCenter, Def: &domain.AssetDefinition{ID: "SWN-W4-003"}}, Registry: registry,` |
| `effects/grant_permission_test.go` (16) | `Faction: owner, Asset: movement, Def: &domain.AssetDefinition{ID: "SWN-C7-001"}, Registry: registry,` | `Holder: effect.Holder{Faction: owner, Asset: movement, Def: &domain.AssetDefinition{ID: "SWN-C7-001"}}, Registry: registry,` |
| `effects/ignore_coin_loss_test.go` (48) | `Faction: alpha, Asset: asset, Def: def, Entry: entry, Registry: registry,` | `Holder: effect.Holder{Faction: alpha, Asset: asset, Def: def}, Entry: entry, Registry: registry,` |
| `effects/lock_out_test.go` (28) | `Faction: alpha, Asset: alpha.Assets["lock"], Def: &domain.AssetDefinition{ID: "SWN-C6-001"}, Registry: registry,` | `Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["lock"], Def: &domain.AssetDefinition{ID: "SWN-C6-001"}}, Registry: registry,` |
| `effects/remove_tag_dice_test.go` (37) | `Faction: alpha, Asset: alpha.Assets["bm"], Def: &domain.AssetDefinition{ID: "SWN-C2-003", Name: "Blackmail"},` | `Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["bm"], Def: &domain.AssetDefinition{ID: "SWN-C2-003", Name: "Blackmail"}},` |
| `effects/reroll_test.go` (57) | `Faction: alpha, Asset: alpha.Assets["bos"], Def: bookOfSecretsDef, Entry: entry, Registry: registry,` | `Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["bos"], Def: bookOfSecretsDef}, Entry: entry, Registry: registry,` |
| `effects/restore_asset_test.go` (37) | `Faction: alpha, Asset: alpha.Assets["mc"], Def: &domain.AssetDefinition{ID: "SWN-W4-003"}, Registry: registry,` | `Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["mc"], Def: &domain.AssetDefinition{ID: "SWN-W4-003"}}, Registry: registry,` |
| `effects/restrict_attack_test.go` (59) | `Faction: alpha, Asset: asset, Def: def, Entry: entry, Registry: registry,` | `Holder: effect.Holder{Faction: alpha, Asset: asset, Def: def}, Entry: entry, Registry: registry,` |
| `effects/sacrifice_self_test.go` (35) | `Faction: alpha, Asset: alpha.Assets["ff"], Def: &domain.AssetDefinition{ID: "SWN-C1-003"}, Registry: registry,` | `Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["ff"], Def: &domain.AssetDefinition{ID: "SWN-C1-003"}}, Registry: registry,` |
| `effects/seize_target_test.go` (35) | `Faction: alpha, Asset: alpha.Assets["ht"], Def: &domain.AssetDefinition{ID: "SWN-W7-002"}, Registry: registry,` | `Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["ht"], Def: &domain.AssetDefinition{ID: "SWN-W7-002"}}, Registry: registry,` |
| `effects/self_damage_test.go` (28) | `Faction: alpha, Asset: asset, Def: def, Registry: registry, Roller: fixedRoller{value: 3},` | `Holder: effect.Holder{Faction: alpha, Asset: asset, Def: def}, Registry: registry, Roller: fixedRoller{value: 3},` |
| `effects/shelter_asset_test.go` (46) | `Faction: alpha, Asset: alpha.Assets["bolt"], Def: boltholesDef, Registry: registry,` | `Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["bolt"], Def: boltholesDef}, Registry: registry,` |
| `effects/tech_level_floor_test.go` (43) | `Faction: alpha, Asset: alpha.Assets["lab"], Def: laboratoryDef, Entry: entry, Registry: registry,` | `Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["lab"], Def: laboratoryDef}, Entry: entry, Registry: registry,` |
| `effects/withdraw_permission_test.go` (29) | `Faction: owner, Asset: lobbyists, Def: &domain.AssetDefinition{ID: "SWN-C2-001"}, Registry: registry, Roller: roller,` | `Holder: effect.Holder{Faction: owner, Asset: lobbyists, Def: &domain.AssetDefinition{ID: "SWN-C2-001"}}, Registry: registry, Roller: roller,` |
| `actions/attack_test.go` (1097) | `Faction: f2, Asset: falseFront, Def: &domain.AssetDefinition{ID: "SWN-C1-003"}, Registry: registry,` | `Holder: effect.Holder{Faction: f2, Asset: falseFront, Def: &domain.AssetDefinition{ID: "SWN-C1-003"}}, Registry: registry,` |
| `actions/attack_test.go` (1175) | `Faction: faction, Asset: asset, Def: &domain.AssetDefinition{ID: "SWN-W7-002"}, Registry: registry,` | `Holder: effect.Holder{Faction: faction, Asset: asset, Def: &domain.AssetDefinition{ID: "SWN-W7-002"}}, Registry: registry,` |
| `actions/attack_test.go` (1196) | `Faction: f2, Asset: boltholes, Def: &domain.AssetDefinition{ID: "SWN-C5-003"}, Registry: registry,` | `Holder: effect.Holder{Faction: f2, Asset: boltholes, Def: &domain.AssetDefinition{ID: "SWN-C5-003"}}, Registry: registry,` |

`effects/grant_movement_test.go`:

**Find:**
```go
		Faction: alpha,
		Asset:   web,
		Def:     transitWebDef,
```
**Replace with:**
```go
		Holder: effect.Holder{Faction: alpha, Asset: web, Def: transitWebDef},
```

`FireContext` and `MatchupContext` literals keep `Asset` as the subject:

| File (line today) | Find | Replace with |
|---|---|---|
| `effects/apply_stealth_test.go` (15) | `Faction: alpha, Asset: asset, Def: &domain.AssetDefinition{ID: "SWN-F5-003"},` | `Holder: effect.Holder{Faction: alpha, Asset: asset, Def: &domain.AssetDefinition{ID: "SWN-F5-003"}}, Asset: asset,` |
| `effects/clear_stealth_test.go` (256, 279: Edit each) | `Faction: alpha, Asset: alpha.Assets["cells"], Def: &domain.AssetDefinition{ID: "SWN-C8-001"}, Roller: roller, FactionState: factionState,` | `Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["cells"], Def: &domain.AssetDefinition{ID: "SWN-C8-001"}}, Asset: alpha.Assets["cells"], Roller: roller, FactionState: factionState,` |
| `effects/coin_gain_test.go` (35) | `Faction: faction,` | `Holder: effect.Holder{Faction: faction},` |

`effects/block_ability_test.go`:

**Find:**
```go
		Faction:       alpha,
		Asset:         activeAsset("saboteurs", "SWN-C2-002"),
```
**Replace with:**
```go
		Holder:        effect.Holder{Faction: alpha, Asset: saboteurs},
		Asset:         saboteurs,
```
and before the `got :=` line, add `saboteurs := activeAsset("saboteurs", "SWN-C2-002")`.

`internal/faction/engine/effect/effect_test.go`:

**Find:**
```go
				if ctx.Faction.ID != "alpha" || ctx.Asset.ID != "alpha-1" || ctx.Def.ID != "T-001" || ctx.Registry == nil || ctx.Roller != roller {
```
**Replace with:**
```go
				if ctx.Holder.Faction.ID != "alpha" || ctx.Holder.Asset.ID != "alpha-1" || ctx.Holder.ID() != "T-001" || ctx.Registry == nil || ctx.Roller != roller {
```

**Find:**
```go
				return []domain.Mutation{domain.CoinDelta{FactionID: ctx.Faction.ID, Delta: 1, Cause: ctx.Asset.ID}}
```
**Replace with:**
```go
				return []domain.Mutation{domain.CoinDelta{FactionID: ctx.Holder.Faction.ID, Delta: 1, Cause: ctx.Asset.ID}}
```

**Find:**
```go
		return []domain.Mutation{domain.CoinDelta{FactionID: ctx.Faction.ID, Delta: 1, Cause: string(ctx.Entry.Trigger)}}
```
**Replace with:**
```go
		return []domain.Mutation{domain.CoinDelta{FactionID: ctx.Holder.Faction.ID, Delta: 1, Cause: string(ctx.Entry.Trigger)}}
```

**Find:**
```go
				Mutations:     []domain.Mutation{domain.CoinDelta{FactionID: ctx.Faction.ID, Delta: 1, Cause: string(ctx.Entry.Effect)}},
```
**Replace with:**
```go
				Mutations:     []domain.Mutation{domain.CoinDelta{FactionID: ctx.Holder.Faction.ID, Delta: 1, Cause: string(ctx.Entry.Effect)}},
```

**Find:**
```go
	ctx := effect.MatchupContext{Faction: &domain.Faction{ID: "alpha"}, Def: def}
```
**Replace with:**
```go
	ctx := effect.MatchupContext{Holder: effect.Holder{Faction: &domain.Faction{ID: "alpha"}, Def: def}}
```

Add a test for the holder's naming, after `noopMatchup`:

```go
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
```

##### Commit message

```
refactor: name the effect holder in the effect contexts

- effect.Holder (faction, asset, definition or tag) replaces the loose
  Faction/Asset/Def fields in RegisterContext, FireContext and MatchupContext
- FireContext.Asset and MatchupContext.Asset name the subject asset
- no change in play
```

### Phase 3 — Holder-neutral hooks

**In plain words:** every hook keeps its holder as IDs and asks, when it fires,
"is my holder still in play?" through one check that works for an asset and for a
tag. Nothing changes in play; a tag can now hold any effect whose scope allows it.

**The pattern for every file in this phase.** Each task gives (a) the new
`register…` function and hook type in full, which replace the old ones, and (b) the
fire-time check as a Find/Replace. Then (c) a rename the compiler checks: in the
file's hook methods, `<receiver>.ownerFactionID` and `<receiver>.factionID` become
`<receiver>.holder.factionID`, `<receiver>.assetID` becomes
`<receiver>.holder.assetID`, and `<receiver>.defID` becomes
`<receiver>.holder.sourceID`. A file whose task says "drop `hooks`" loses that
import, because `ctx.Holder.Scope()` replaces its only use.

#### Commit 4 — `refactor: check roll and reactor hooks through their holder`

##### Task 1 — `internal/faction/engine/effect/effects/holder.go` (new)

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// holderRef is what a hook keeps of its holder: IDs, read again at fire time,
// because a held pointer keeps a removed asset in effect (Decision Record 19).
// assetID is empty for a tag; tagID is empty for an asset. sourceID is the
// definition or tag ID that budget keys name.
type holderRef struct {
	factionID string
	assetID   string
	tagID     string
	sourceID  string
}

func refOf(holder effect.Holder) holderRef {
	ref := holderRef{factionID: holder.Faction.ID, sourceID: holder.ID()}
	if holder.Asset != nil {
		ref.assetID = holder.Asset.ID
	}
	if holder.Tag != nil {
		ref.tagID = holder.Tag.ID
	}
	return ref
}

// live returns the declaring asset from current state, and whether the holder
// still applies: its asset is active, or its faction still holds the tag. A tag
// has no asset. A nil state (a dispatcher test) is no holder.
func (ref holderRef) live(factionState *state.FactionState) (*domain.Asset, bool) {
	if factionState == nil {
		return nil, false
	}
	return ref.liveIn(factionState.Factions[ref.factionID])
}

// liveIn is live for a Cat 4 hook, which gets no FactionState: it reads the owner
// it kept from registration, whose maps stay live all cycle.
func (ref holderRef) liveIn(owner *domain.Faction) (*domain.Asset, bool) {
	if owner == nil {
		return nil, false
	}
	if ref.tagID != "" {
		return nil, owner.HasTag(ref.tagID)
	}
	asset := owner.Assets[ref.assetID]
	return asset, active(asset)
}
```

`internal/faction/engine/effect/effects/holder_test.go` (new):

```go
package effects

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestHolderRef_Live(t *testing.T) {
	asset := &domain.Asset{ID: "a-1", Ready: true, Maintained: true, CurrentHP: 1}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-020"}, Assets: map[string]*domain.Asset{"a-1": asset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha}}
	assetRef := refOf(effect.Holder{Faction: alpha, Asset: asset, Def: &domain.AssetDefinition{ID: "SWN-C7-002"}})
	tagRef := refOf(effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-020"}})

	if got, live := assetRef.live(factionState); !live || got != asset {
		t.Errorf("active asset: live = %v, asset = %v", live, got)
	}
	if got, live := tagRef.live(factionState); !live || got != nil {
		t.Errorf("held tag: live = %v, asset = %v; want live and no asset", live, got)
	}
	if _, live := tagRef.live(nil); live {
		t.Error("nil state: live, want not")
	}

	asset.Ready = false
	if _, live := assetRef.live(factionState); live {
		t.Error("unready asset: live, want not")
	}
	alpha.Tags = nil
	if _, live := tagRef.liveIn(alpha); live {
		t.Error("dropped tag: live, want not")
	}
	delete(factionState.Factions, "alpha")
	if _, live := assetRef.live(factionState); live {
		t.Error("removed faction: live, want not")
	}
}
```

##### Task 2 — roll hooks

`bonus_die.go` — (a):

```go
func registerBonusDie(ctx effect.RegisterContext) {
	ctx.Registry.RegisterRollModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialBonusDie),
		&bonusDieModifier{
			holder:  refOf(ctx.Holder),
			source:  ctx.Holder.Name(),
			trigger: ctx.Entry.Trigger,
			filters: ctx.Entry.Filters,
			params:  *ctx.Entry.BonusDie,
		},
	)
}

type bonusDieModifier struct {
	holder  holderRef
	source  string
	trigger domain.EffectTrigger
	filters domain.EffectFilters
	params  domain.BonusDieParams
}
```
(b):

**Find:**
```go
	asset := liveAsset(factionState, modifier.ownerFactionID, modifier.assetID)
	if !active(asset) || !rollMatches(ctx, modifier.trigger, modifier.ownerFactionID, asset, modifier.filters) {
```
**Replace with:**
```go
	asset, live := modifier.holder.live(factionState)
	if !live || !rollMatches(ctx, modifier.trigger, modifier.holder.factionID, asset, modifier.filters) {
```

`reroll.go` — (a):

```go
func registerReroll(ctx effect.RegisterContext) {
	source := ctx.Holder.Name()
	if ctx.Entry.Trigger == domain.TriggerEnemyRollAtMe {
		source += " (forced)"
	}
	ctx.Registry.RegisterRollResultHook(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialReroll),
		&rerollHook{
			holder:  refOf(ctx.Holder),
			source:  source,
			trigger: ctx.Entry.Trigger,
			filters: ctx.Entry.Filters,
			oncePer: ctx.Entry.Gates.OncePer,
			params:  *ctx.Entry.Reroll,
		},
	)
}

type rerollHook struct {
	holder  holderRef
	source  string
	trigger domain.EffectTrigger
	filters domain.EffectFilters
	oncePer domain.OncePerKey
	params  domain.RerollParams
}
```
(b):

**Find:**
```go
	asset := liveAsset(factionState, hook.ownerFactionID, hook.assetID)
	if !active(asset) || !rollMatches(ctx, hook.trigger, hook.ownerFactionID, asset, hook.filters) {
```
**Replace with:**
```go
	asset, live := hook.holder.live(factionState)
	if !live || !rollMatches(ctx, hook.trigger, hook.holder.factionID, asset, hook.filters) {
```

`remove_tag_dice.go` — (a):

```go
func registerRemoveTagDice(ctx effect.RegisterContext) {
	ctx.Registry.RegisterRollOfferFilter(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialRemoveTagDice),
		&tagDiceFilter{
			holder:  refOf(ctx.Holder),
			trigger: ctx.Entry.Trigger,
			filters: ctx.Entry.Filters,
		},
	)
}

type tagDiceFilter struct {
	holder  holderRef
	trigger domain.EffectTrigger
	filters domain.EffectFilters
}
```
(b):

**Find:**
```go
	asset := liveAsset(factionState, filter.ownerFactionID, filter.assetID)
	if !active(asset) || !rollMatches(ctx, filter.trigger, filter.ownerFactionID, asset, filter.filters) {
```
**Replace with:**
```go
	asset, live := filter.holder.live(factionState)
	if !live || !rollMatches(ctx, filter.trigger, filter.holder.factionID, asset, filter.filters) {
```

##### Task 3 — reactors

`capture_target.go` — drop `hooks`. (a):

```go
func registerCaptureTarget(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialCaptureTarget),
		&captureTargetReactor{
			holder: refOf(ctx.Holder),
			coin:   ctx.Entry.CaptureTarget.Coin,
			roller: ctx.Roller,
		},
	)
}

type captureTargetReactor struct {
	holder holderRef
	coin   domain.DiceRoll
	roller domain.Roller
}
```
(b):

**Find:**
```go
	self := liveAsset(factionState, reactor.factionID, reactor.assetID)
	if !active(self) {
		return nil
	}
	allHits := hits(mutations, false)
```
**Replace with:**
```go
	self, live := reactor.holder.live(factionState)
	if !live {
		return nil
	}
	allHits := hits(mutations, false)
```

`clear_stealth.go` — drop `hooks`. (a):

```go
func registerClearStealth(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialClearStealth),
		&clearStealthReactor{
			holder:   refOf(ctx.Holder),
			stealthy: ctx.Entry.Filters.Stealthy,
			contest:  ctx.Entry.Gates.Contest,
			roller:   ctx.Roller,
		},
	)
}

type clearStealthReactor struct {
	holder   holderRef
	stealthy *bool
	contest  *domain.EffectContest
	roller   domain.Roller
}
```
(b):

**Find:**
```go
	owner := factionState.Factions[reactor.ownerFactionID]
	if owner == nil || !active(owner.Assets[reactor.assetID]) {
		return nil
	}
	worldID := owner.Assets[reactor.assetID].Location.WorldID
```
**Replace with:**
```go
	self, live := reactor.holder.live(factionState)
	if !live {
		return nil
	}
	owner := factionState.Factions[reactor.holder.factionID]
	worldID := self.Location.WorldID
```

`coin_steal.go` — drop `hooks`. The lead rule is its fire-time check: the hook
reacts only when its asset is the owner's lowest-ID active thief. (a):

```go
func registerCoinSteal(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialCoinSteal),
		&coinStealReactor{holder: refOf(ctx.Holder), roller: ctx.Roller},
	)
}

// coinStealReactor reads its entries from the rulebook, not from registration,
// because the lead thief runs the entries of every thief its faction owns.
type coinStealReactor struct {
	holder holderRef
	roller domain.Roller
}
```
(b): none; (c) covers both fields.

`ignore_coin_loss.go` — drop `hooks`. The lead copy is its fire-time check, as for
`coin_steal`. (a):

```go
func registerIgnoreCoinLoss(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialIgnoreCoinLoss),
		&ignoreCoinLossReactor{
			holder:  refOf(ctx.Holder),
			trigger: ctx.Entry.Trigger,
			oncePer: ctx.Entry.Gates.OncePer,
		},
	)
}

type ignoreCoinLossReactor struct {
	holder  holderRef
	trigger domain.EffectTrigger
	oncePer domain.OncePerKey
}
```
(b): none.

`lock_out.go` — drop `hooks`. (a):

```go
func registerLockOut(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialLockOut), &lockOutReactor{holder: refOf(ctx.Holder)})
}

type lockOutReactor struct {
	holder holderRef
}
```
(b):

**Find:**
```go
	self := liveAsset(factionState, reactor.factionID, reactor.assetID)
	if !active(self) {
		return nil
	}
	// deliberate: no lock from a Lockdown the batch removes or captures, because its condition would apply after its removal and miss.
```
**Replace with:**
```go
	self, live := reactor.holder.live(factionState)
	if !live {
		return nil
	}
	// deliberate: no lock from a Lockdown the batch removes or captures, because its condition would apply after its removal and miss.
```

`self_damage.go` — (a):

```go
func registerSelfDamage(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialSelfDamage),
		&selfDamageReactor{
			holder:   refOf(ctx.Holder),
			def:      ctx.Holder.Def,
			counter:  ctx.Entry.Trigger == domain.TriggerOwnCounterHit,
			amount:   ctx.Entry.SelfDamage.Amount,
			roller:   ctx.Roller,
			registry: ctx.Registry,
		},
	)
}

// selfDamageReactor keeps its definition and the registry, because a lethal
// self-hit asks the Cat 6 interceptors (D-10).
type selfDamageReactor struct {
	holder   holderRef
	def      *domain.AssetDefinition
	counter  bool
	amount   domain.DiceRoll
	roller   domain.Roller
	registry *hooks.Registry
}
```
(b):

**Find:**
```go
	faction := factionState.Factions[reactor.factionID]
	if faction == nil {
		return nil
	}
	asset := faction.Assets[reactor.assetID]
	if !active(asset) {
		return nil
	}
```
**Replace with:**
```go
	asset, live := reactor.holder.live(factionState)
	if !live {
		return nil
	}
```

`shelter_asset.go` — (a) replaces `registerShelterAsset`, the
`shelterAssetInterceptor` type and the `shelterLossReactor` type (keep the loss
reactor's comment):

```go
func registerShelterAsset(ctx effect.RegisterContext) {
	scope := ctx.Holder.Scope()
	source := ctx.Holder.HookID(domain.SpecialShelterAsset)
	holder := refOf(ctx.Holder)
	ctx.Registry.RegisterLethalInterceptor(scope, source, &shelterAssetInterceptor{holder: holder, assetTypes: ctx.Entry.Filters.AssetTypes})
	ctx.Registry.RegisterMutationReactor(scope, source, &shelterLossReactor{holder: holder})
}

type shelterAssetInterceptor struct {
	holder     holderRef
	assetTypes []domain.AssetType
}

// shelterLossReactor destroys the assets its Boltholes shelters when damage
// destroys the Boltholes. A sale or a refit does not (Effort 2 Decision 13).
type shelterLossReactor struct {
	holder holderRef
}
```
(b) — the interceptor. The loss reactor keeps no check (its fence-sign stands):

**Find:**
```go
	self := liveAsset(ctx.FactionState, interceptor.factionID, interceptor.assetID)
	if !guardsTarget(self, ctx, interceptor.assetTypes) {
		return nil
	}
	cause := domain.SpecialShelterAsset
```
**Replace with:**
```go
	self, live := interceptor.holder.live(ctx.FactionState)
	if !live || !guardsTarget(self, ctx, interceptor.assetTypes) {
		return nil
	}
	cause := domain.SpecialShelterAsset
```

##### Commit message

```
refactor: check roll and reactor hooks through their holder

- holderRef keeps a hook's holder as IDs; live and liveIn answer "still in
  play?" for an asset (active) and a tag (still held)
- roll hooks, reactors and Boltholes sign up through Holder.Scope and HookID
- no change in play
```

#### Commit 5 — `refactor: check rule and lethal hooks through their holder`

##### Task 1 — Cat 4 hooks

A Cat 4 hook gets no `FactionState`, so it keeps its owner and calls `liveIn`.

`cost_surcharge.go` — (a):

```go
func registerCostSurcharge(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetCostModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialCostSurcharge),
		&costSurchargeModifier{
			owner:   ctx.Holder.Faction,
			holder:  refOf(ctx.Holder),
			noStack: ctx.Entry.Gates.NoStack,
			amount:  ctx.Entry.CostSurcharge.Amount,
		},
	)
}

// costSurchargeModifier holds its owner, not its asset, because a Cat 4 hook gets
// no FactionState: the owner pointer is live all cycle, and an asset removed
// mid-cycle reads as nil from owner.Assets where a held asset pointer would not.
type costSurchargeModifier struct {
	owner   *domain.Faction
	holder  holderRef
	noStack bool
	amount  int
}
```
(b):

**Find:**
```go
	asset := modifier.owner.Assets[modifier.assetID]
	if !active(asset) || buyer.ID == modifier.owner.ID || asset.Location.WorldID != world {
```
**Replace with:**
```go
	asset, live := modifier.holder.liveIn(modifier.owner)
	if !live || buyer.ID == modifier.owner.ID || asset.Location.WorldID != world {
```

`discount_repair.go` — (a):

```go
func registerDiscountRepair(ctx effect.RegisterContext) {
	ctx.Registry.RegisterRepairCostModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialDiscountRepair),
		&discountRepairModifier{
			owner:      ctx.Holder.Faction,
			holder:     refOf(ctx.Holder),
			noStack:    ctx.Entry.Gates.NoStack,
			amount:     ctx.Entry.DiscountRepair.Amount,
			assetTypes: ctx.Entry.Filters.AssetTypes,
		},
	)
}

// discountRepairModifier holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Decision Record 19).
type discountRepairModifier struct {
	owner      *domain.Faction
	holder     holderRef
	noStack    bool
	amount     int
	assetTypes []domain.AssetType
}
```
(b):

**Find:**
```go
	self := modifier.owner.Assets[modifier.assetID]
	if !active(self) || repairer.ID != modifier.owner.ID {
```
**Replace with:**
```go
	self, live := modifier.holder.liveIn(modifier.owner)
	if !live || repairer.ID != modifier.owner.ID {
```

`grant_movement.go` — (a):

```go
func registerGrantMovement(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetMovementGranter(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialGrantMovement),
		&movementGranter{
			owner:    ctx.Holder.Faction,
			holder:   refOf(ctx.Holder),
			relocate: ctx.Entry.GrantMovement.Relocate,
		},
	)
}

// movementGranter reads its asset through its owner at query time (Decision
// Record 19), so a grant from an asset removed or made inactive mid-cycle lapses.
type movementGranter struct {
	owner    *domain.Faction
	holder   holderRef
	relocate domain.RelocateParams
}
```
(b):

**Find:**
```go
	asset := granter.owner.Assets[granter.assetID]
	if !active(asset) {
		return nil
	}
```
**Replace with:**
```go
	asset, live := granter.holder.liveIn(granter.owner)
	if !live {
		return nil
	}
```

`grant_permission.go` — (a):

```go
func registerGrantPermission(ctx effect.RegisterContext) {
	ctx.Registry.RegisterPermissionGranter(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialGrantPermission),
		&permissionGranter{owner: ctx.Holder.Faction, holder: refOf(ctx.Holder)},
	)
}

// permissionGranter holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Effort 2 Decision Record 19).
type permissionGranter struct {
	owner  *domain.Faction
	holder holderRef
}
```
(b):

**Find:**
```go
	self := granter.owner.Assets[granter.assetID]
	if !active(self) || query.Requester.ID != granter.owner.ID {
```
**Replace with:**
```go
	self, live := granter.holder.liveIn(granter.owner)
	if !live || query.Requester.ID != granter.owner.ID {
```

`restrict_attack.go` — (a) replaces `registerRestrictAttack` and the
`restrictAttackRule` type:

```go
func registerRestrictAttack(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAttackEligibilityRule(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialRestrictAttack),
		&restrictAttackRule{
			owner:      ctx.Holder.Faction,
			holder:     refOf(ctx.Holder),
			side:       ctx.Entry.RestrictAttack.Side,
			allow:      ctx.Entry.RestrictAttack.Allow,
			assetTypes: ctx.Entry.Filters.AssetTypes,
			categories: ctx.Entry.Filters.Categories,
			attribute:  ctx.Entry.Filters.Attribute,
		},
	)
}

// restrictAttackRule holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Decision Record 19).
type restrictAttackRule struct {
	owner      *domain.Faction
	holder     holderRef
	side       domain.RestrictSide
	allow      domain.RestrictAllow
	assetTypes []domain.AssetType
	categories []domain.FactionStat
	attribute  domain.FactionStat
}
```
(b):

**Find:**
```go
	if !active(rule.owner.Assets[rule.assetID]) {
		return true
	}
```
**Replace with:**
```go
	if _, live := rule.holder.liveIn(rule.owner); !live {
		return true
	}
```

`tech_level_floor.go` — drop `hooks`. (a):

```go
func registerTechLevelFloor(ctx effect.RegisterContext) {
	ctx.Registry.RegisterWorldTechLevelModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialTechLevelFloor),
		&techLevelFloorModifier{
			owner:      ctx.Holder.Faction,
			holder:     refOf(ctx.Holder),
			scope:      ctx.Entry.Filters.Scope,
			categories: ctx.Entry.Filters.Categories,
			level:      ctx.Entry.TechLevelFloor.Level,
		},
	)
}

// techLevelFloorModifier holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Decision Record 19).
type techLevelFloorModifier struct {
	owner      *domain.Faction
	holder     holderRef
	scope      domain.EffectScope
	categories []domain.FactionStat
	level      int
}
```
(b):

**Find:**
```go
	asset := modifier.owner.Assets[modifier.assetID]
	if !active(asset) || faction.ID != modifier.owner.ID {
		return baseTL
	}
```
**Replace with:**
```go
	asset, live := modifier.holder.liveIn(modifier.owner)
	if !live || faction.ID != modifier.owner.ID {
		return baseTL
	}
```

`withdraw_permission.go` — (a) replaces `registerWithdrawPermission` and the
`permissionContester` type:

```go
func registerWithdrawPermission(ctx effect.RegisterContext) {
	ctx.Registry.RegisterPermissionContester(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialWithdrawPermission),
		&permissionContester{owner: ctx.Holder.Faction, holder: refOf(ctx.Holder), contest: *ctx.Entry.Gates.Contest, roller: ctx.Roller},
	)
}

// permissionContester holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Effort 2 Decision Record 19). It keeps the roller its
// registration carried, as a reactor does.
type permissionContester struct {
	owner   *domain.Faction
	holder  holderRef
	contest domain.EffectContest
	roller  domain.Roller
}
```
(b):

**Find:**
```go
	self := contester.owner.Assets[contester.assetID]
	if !active(self) || query.Requester.ID == contester.owner.ID {
```
**Replace with:**
```go
	self, live := contester.holder.liveIn(contester.owner)
	if !live || query.Requester.ID == contester.owner.ID {
```

##### Task 2 — Cat 6 interceptors

`restore_asset.go` — (a):

```go
func registerRestoreAsset(ctx effect.RegisterContext) {
	ctx.Registry.RegisterLethalInterceptor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialRestoreAsset),
		&restoreAssetInterceptor{
			holder:     refOf(ctx.Holder),
			assetTypes: ctx.Entry.Filters.AssetTypes,
			oncePer:    ctx.Entry.Gates.OncePer,
		},
	)
}

type restoreAssetInterceptor struct {
	holder     holderRef
	assetTypes []domain.AssetType
	oncePer    domain.OncePerKey
}
```
(b):

**Find:**
```go
	self := liveAsset(ctx.FactionState, interceptor.factionID, interceptor.assetID)
	if !guardsTarget(self, ctx, interceptor.assetTypes) || ctx.TargetDef == nil {
```
**Replace with:**
```go
	self, live := interceptor.holder.live(ctx.FactionState)
	if !live || !guardsTarget(self, ctx, interceptor.assetTypes) || ctx.TargetDef == nil {
```

`sacrifice_self.go` — (a):

```go
func registerSacrificeSelf(ctx effect.RegisterContext) {
	ctx.Registry.RegisterLethalInterceptor(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialSacrificeSelf), &sacrificeSelfInterceptor{holder: refOf(ctx.Holder)})
}

type sacrificeSelfInterceptor struct {
	holder holderRef
}
```
(b):

**Find:**
```go
	self := liveAsset(ctx.FactionState, interceptor.factionID, interceptor.assetID)
	if !guardsTarget(self, ctx, nil) {
```
**Replace with:**
```go
	self, live := interceptor.holder.live(ctx.FactionState)
	if !live || !guardsTarget(self, ctx, nil) {
```

`seize_target.go` — (a):

```go
func registerSeizeTarget(ctx effect.RegisterContext) {
	ctx.Registry.RegisterLethalInterceptor(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialSeizeTarget), &seizeTargetInterceptor{holder: refOf(ctx.Holder)})
}

type seizeTargetInterceptor struct {
	holder holderRef
}
```
(b):

**Find:**
```go
	self := liveAsset(ctx.FactionState, interceptor.factionID, interceptor.assetID)
	if !active(self) || ctx.Striker == nil || ctx.Striker.ID != self.ID || ctx.Target.OwnerID == self.OwnerID {
```
**Replace with:**
```go
	self, live := interceptor.holder.live(ctx.FactionState)
	if !live || ctx.Striker == nil || ctx.Striker.ID != self.ID || ctx.Target.OwnerID == self.OwnerID {
```

##### Task 3 — `internal/faction/engine/effect/effects/tech_level_floor_test.go`

The one registrar that already accepts a holder-free scope now works for a tag
with no other change. Add:

```go
func TestTechLevelFloor_TagHolderRaisesEveryWorldWhileHeld(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-018"}, Assets: map[string]*domain.Asset{}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialTechLevelFloor].Register(effect.RegisterContext{
		Holder:   effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-018", Name: "Technical Expertise"}},
		Entry:    techLevelFloorEntry(domain.ScopeFaction, 4),
		Registry: registry,
	})

	if got := dispatch.ResolveWorldTechLevel(registry, alpha, forceAsset, "Krylos", 2); got != 4 {
		t.Errorf("tech level = %d, want 4 while the tag is held", got)
	}
	alpha.Tags = nil
	if got := dispatch.ResolveWorldTechLevel(registry, alpha, forceAsset, "Krylos", 2); got != 2 {
		t.Errorf("tech level = %d, want 2 after the tag is dropped", got)
	}
}
```

##### Commit message

```
refactor: check rule and lethal hooks through their holder

- Cat 4 hooks call holderRef.liveIn on the owner they keep; Cat 6
  interceptors call holderRef.live
- every registrar now signs up and checks through its holder, so a tag can
  hold any effect whose scope allows it (tech_level_floor at scope faction)
- no change in play
```

### Phase 4 — Tag holder

**In plain words:** a tag record carries `[[effects]]`, engine construction checks
each entry for a faction holder, and the cycle registers and fires a held tag's
entries as it does an asset's. Ties and the buyer's own asset cost also ask asset
hooks now. No shipped tag has entries yet.

#### Commit 6 — `feat: load and check tag effects`

##### Task 1 — `internal/faction/domain/faction.go`

**Find:**
```go
	Effect      string
	Government  bool
}
```
**Replace with:**
```go
	Effect      string
	Government  bool
	Effects     []EffectDefinition
}
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

**Find:**
```go
	Government  bool   `toml:"government"`
}
```
**Replace with:**
```go
	Government  bool           `toml:"government"`
	Effects     []effectRecord `toml:"effects"`
}
```

**Find:**
```go
		if r.Government {
			if government != "" {
				return nil, fmt.Errorf("tags %q and %q both set government in %s", government, r.ID, path)
			}
			government = r.ID
		}
		result[r.ID] = &domain.Tag{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Effect:      r.Effect,
			Government:  r.Government,
		}
```
**Replace with:**
```go
		if r.Government {
			if government != "" {
				return nil, fmt.Errorf("tags %q and %q both set government in %s", government, r.ID, path)
			}
			government = r.ID
		}
		effects, err := convertEffects(r.Effects)
		if err != nil {
			return nil, fmt.Errorf("tag %q in %s: %w", r.ID, path, err)
		}
		result[r.ID] = &domain.Tag{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Effect:      r.Effect,
			Government:  r.Government,
			Effects:     effects,
		}
```

##### Task 3 — `internal/faction/engine/effect/registrar.go`

**Find:**
```go
type Registrar struct {
	Triggers map[domain.EffectTrigger]Accepts
	Register func(RegisterContext)
	Fire     func(FireContext) []domain.Mutation
	Matchup  func(MatchupContext) MatchupOutcome
	Validate func(domain.EffectDefinition) error
	Targets  func(TargetContext) []*domain.Asset
}
```
**Replace with:**
```go
type Registrar struct {
	Triggers map[domain.EffectTrigger]Accepts
	Register func(RegisterContext)
	Fire     func(FireContext) []domain.Mutation
	Matchup  func(MatchupContext) MatchupOutcome
	Validate func(domain.EffectDefinition, HolderKind) error
	Targets  func(TargetContext) []*domain.Asset
}

// HolderKind is what declared an entry, as engine construction checks it.
type HolderKind int

const (
	HolderAsset HolderKind = iota
	HolderTag
)
```

In `validateEntry`, take the holder and refuse what a tag cannot use:

**Find:**
```go
func (e *EffectsEngine) validateEntry(entry domain.EffectDefinition) error {
```
**Replace with:**
```go
func (e *EffectsEngine) validateEntry(entry domain.EffectDefinition, holder HolderKind) error {
```

**Find:**
```go
	if !slices.Contains(accepts.Scopes, entry.Filters.Scope) {
		return fmt.Errorf("effect %q does not accept scope %q for trigger %q", entry.Effect, entry.Filters.Scope, entry.Trigger)
	}
```
**Replace with:**
```go
	if !slices.Contains(accepts.Scopes, entry.Filters.Scope) {
		return fmt.Errorf("effect %q does not accept scope %q for trigger %q", entry.Effect, entry.Filters.Scope, entry.Trigger)
	}
	// deliberate: scope decides the holder, not a holders list, because self and world measure from the declaring asset and a tag has none (Effort 1 Decision 2).
	if holder == HolderTag && (entry.Filters.Scope == domain.ScopeSelf || entry.Filters.Scope == domain.ScopeWorld) {
		return fmt.Errorf("effect %q: a tag has no asset, so scope %q cannot apply", entry.Effect, entry.Filters.Scope)
	}
	if holder == HolderTag && entry.Gates.OncePer == domain.OncePerAsset {
		return fmt.Errorf("effect %q: a tag has no asset, so once_per %q cannot apply", entry.Effect, domain.OncePerAsset)
	}
```

**Find:**
```go
		if err := registrar.Validate(entry); err != nil {
```
**Replace with:**
```go
		if err := registrar.Validate(entry, holder); err != nil {
```

##### Task 4 — `internal/faction/engine/effect/effect.go` and the registrars

`effect.go`, in `RegisterEffectsList`:

**Find:**
```go
		if err := e.validateEntry(entry); err != nil {
			return fmt.Errorf("asset %q effects[%d]: %w", def.ID, i, err)
		}
```
**Replace with:**
```go
		if err := e.validateEntry(entry, HolderAsset); err != nil {
			return fmt.Errorf("asset %q effects[%d]: %w", def.ID, i, err)
		}
```

After `RegisterEffectsList`, add:

```go
// CheckTagEffects checks a tag's [[effects]] entries for a faction holder. A tag
// registers no handler: ApplyAll and the phase routes read its entries from the
// rulebook, whose every tag NewWithRulebook has checked.
func (e *EffectsEngine) CheckTagEffects(tag *domain.Tag) error {
	for i, entry := range tag.Effects {
		if err := e.validateEntry(entry, HolderTag); err != nil {
			return fmt.Errorf("tag %q effects[%d]: %w", tag.ID, i, err)
		}
	}
	return nil
}
```

`effects/holder.go` — add `"fmt"` to the imports, and at the end:

```go
// refuseTagHolder is the Validate of a registrar whose faction scope still reads
// the declaring asset, so a tag cannot hold it yet (Effort 1 Decision 2).
func refuseTagHolder(_ domain.EffectDefinition, holder effect.HolderKind) error {
	if holder == effect.HolderTag {
		return fmt.Errorf("a tag cannot hold this effect yet: it reads the declaring asset")
	}
	return nil
}
```

`effects/grant_movement.go`:

**Find:**
```go
	Register: registerGrantMovement,
}
```
**Replace with:**
```go
	Register: registerGrantMovement,
	// deliberate: refuse a tag, not accept it, because the range is measured from the declaring asset; Effort 2's self-move mode lifts this (Effort 1 Decision 2).
	Validate: refuseTagHolder,
}
```

`effects/ignore_coin_loss.go`:

**Find:**
```go
	Register: registerIgnoreCoinLoss,
}
```
**Replace with:**
```go
	Register: registerIgnoreCoinLoss,
	// deliberate: refuse a tag, not accept it, because only the lowest-ID active copy reacts and a tag has no copies (Effort 1 Decision 2).
	Validate: refuseTagHolder,
}
```

`effects/restrict_attack.go`:

**Find:**
```go
func validateRestrictAttack(entry domain.EffectDefinition) error {
```
**Replace with:**
```go
func validateRestrictAttack(entry domain.EffectDefinition, _ effect.HolderKind) error {
```

`effects/withdraw_permission.go`:

**Find:**
```go
func validateWithdrawPermission(entry domain.EffectDefinition) error {
```
**Replace with:**
```go
func validateWithdrawPermission(entry domain.EffectDefinition, _ effect.HolderKind) error {
```

##### Task 5 — `internal/faction/engine/core.go` and tests

`core.go`, in `NewWithRulebook`:

**Find:**
```go
	actions.RegisterDefaultActions(e.Action, func() domain.Roller { return e.Rand }, e.Hooks, e.World, e.Rulebook, e.Effect)
```
**Replace with:**
```go
	for _, tagID := range slices.Sorted(maps.Keys(rulebook.Tags)) {
		if err := e.Effect.CheckTagEffects(rulebook.Tags[tagID]); err != nil {
			return nil, fmt.Errorf("engine: %w", err)
		}
	}
	actions.RegisterDefaultActions(e.Action, func() domain.Roller { return e.Rand }, e.Hooks, e.World, e.Rulebook, e.Effect)
```

`core_test.go` — after `TestNewWithRulebook_ValidatesEffects`, add:

```go
func TestNewWithRulebook_ValidatesTagEffects(t *testing.T) {
	tag := &domain.Tag{ID: "SWN-T-020", Effects: []domain.EffectDefinition{{
		Trigger:  domain.TriggerOwnRoll,
		Effect:   domain.SpecialBonusDie,
		Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
		BonusDie: &domain.BonusDieParams{Sides: 10, Count: 1},
	}}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{}, Tags: map[string]*domain.Tag{tag.ID: tag}}
	_, err := NewWithRulebook(rb, nil, slog.New(slog.DiscardHandler))
	if err == nil || !strings.Contains(err.Error(), `tag "SWN-T-020" effects[0]`) || !strings.Contains(err.Error(), "a tag has no asset") {
		t.Fatalf("got %v, want a tag validation error naming the entry", err)
	}
}
```

`internal/faction/engine/effect/effect_test.go` — after `TestRegisterEffectsList_Validates`, add:

```go
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
```

`internal/faction/rulebook/rulebook_test.go` — after `TestLoadTagsAndGoals_Strict`, add:

```go
func TestLoadTags_Effects(t *testing.T) {
	write := func(body string) string {
		path := filepath.Join(t.TempDir(), "tags.toml")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("writing fixture: %v", err)
		}
		return path
	}
	const header = "[tags.Warlike]\nid = \"SWN-T-020\"\nname = \"Warlike\"\n\n[[tags.Warlike.effects]]\ntrigger = \"own_roll\"\n"

	tags, err := loadTags(write(header + "effect  = \"bonus_die\"\n[tags.Warlike.effects.filters]\nowner = \"own\"\nscope = \"faction\"\n[tags.Warlike.effects.bonus_die]\nsides = 10\ncount = 1\n"))
	if err != nil {
		t.Fatalf("loadTags: %v", err)
	}
	want := []domain.EffectDefinition{{
		Trigger:  domain.TriggerOwnRoll,
		Effect:   domain.SpecialBonusDie,
		Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		BonusDie: &domain.BonusDieParams{Sides: 10, Count: 1},
	}}
	if got := tags["SWN-T-020"].Effects; !reflect.DeepEqual(got, want) {
		t.Errorf("Effects = %+v, want %+v", got, want)
	}

	_, err = loadTags(write(header + "effect  = \"no_such_effect\"\n[tags.Warlike.effects.filters]\nowner = \"own\"\nscope = \"faction\"\n"))
	if err == nil || !strings.Contains(err.Error(), `tag "SWN-T-020"`) {
		t.Errorf("loadTags: got %v, want an error naming the tag", err)
	}
}
```

##### Commit message

```
feat: load and check tag effects

- a tag record carries [[effects]], converted as an asset's are
- engine construction checks every tag entry for a faction holder: scope
  self or world and once_per asset are refused, since a tag has no asset
- grant_movement and ignore_coin_loss refuse a tag for now
```

#### Commit 7 — `feat: register and fire tag effects`

##### Task 1 — `internal/faction/engine/effect/effect.go`: registration

**Find:**
```go
// ApplyAll walks every asset across all factions — factions by ID, then assets by
// ID, so hook registration order is the same every run — looks up the handler for
// that asset's definition ID, and invokes it. Assets whose definitions are missing
// or whose def has no registered handler are silently skipped.
func (e *EffectsEngine) ApplyAll(factionState *state.FactionState, rulebook *rulebook.Rulebook, hookRegistry *hooks.Registry, roller domain.Roller, log *slog.Logger) {
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		faction := factionState.Factions[factionID]
		for _, asset := range domain.SortedAssets(faction) {
			e.ApplyAsset(faction, asset, rulebook, hookRegistry, roller)
		}
	}
}
```
**Replace with:**
```go
// ApplyAll walks every faction by ID, then its assets by ID, then the tags it
// holds, so hook registration order is the same every run. An asset invokes the
// handler for its definition; a tag registers its hook-route entries at its
// faction's scope. A missing definition, handler or tag registers nothing.
func (e *EffectsEngine) ApplyAll(factionState *state.FactionState, rulebook *rulebook.Rulebook, hookRegistry *hooks.Registry, roller domain.Roller, log *slog.Logger) {
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		faction := factionState.Factions[factionID]
		for _, asset := range domain.SortedAssets(faction) {
			e.ApplyAsset(faction, asset, rulebook, hookRegistry, roller)
		}
		for _, tag := range heldTags(faction, rulebook) {
			for _, entry := range tag.Effects {
				if IsPhaseTrigger(entry.Trigger) || IsMatchupTrigger(entry.Trigger) {
					continue
				}
				// deliberate: no nil check on Register, because NewWithRulebook ran CheckTagEffects on every tag, which rejects a hook-route entry without one.
				e.registrars[entry.Effect].Register(RegisterContext{
					Holder:   Holder{Faction: faction, Tag: tag},
					Entry:    entry,
					Registry: hookRegistry,
					Roller:   roller,
				})
			}
		}
	}
}

// heldTags returns the rulebook tags faction holds, in the order it holds them. A
// held tag the rulebook does not know is skipped, as an asset with no definition
// is (Effort 1 Decision 10).
func heldTags(faction *domain.Faction, rulebook *rulebook.Rulebook) []*domain.Tag {
	var tags []*domain.Tag
	for _, tagID := range faction.Tags {
		if tag := rulebook.Tags[tagID]; tag != nil {
			tags = append(tags, tag)
		}
	}
	return tags
}
```

##### Task 2 — `internal/faction/engine/effect/effect.go`: phase and matchup routes

In `FireTurnStart`, fire the held tags after the assets:

**Find:**
```go
// FireTurnStart runs every turn_start entry on faction's active assets, in asset ID
// order, through its registrar's phase route. A direct definition read, not a hook:
// turn start is a phase event the orchestrator raises once per turn.
```
**Replace with:**
```go
// FireTurnStart runs every turn_start entry on faction's active assets, in asset ID
// order, then on the tags it holds, through each registrar's phase route. A direct
// definition read, not a hook: turn start is a phase event the orchestrator raises
// once per turn.
```

**Find:**
```go
				FactionState: factionState,
			})...)
		}
	}
	return mutations
}
```
**Replace with:**
```go
				FactionState: factionState,
			})...)
		}
	}
	for _, tag := range heldTags(faction, rulebook) {
		mutations = append(mutations, e.fireTag(faction, tag, domain.TriggerTurnStart, nil, factionState, roller)...)
	}
	return mutations
}

// fireTag runs tag's entries for trigger through each registrar's phase route.
// asset is the subject: the bought asset at purchase, nil at turn start.
func (e *EffectsEngine) fireTag(faction *domain.Faction, tag *domain.Tag, trigger domain.EffectTrigger, asset *domain.Asset, factionState *state.FactionState, roller domain.Roller) []domain.Mutation {
	var mutations []domain.Mutation
	for _, entry := range tag.Effects {
		if entry.Trigger != trigger {
			continue
		}
		// deliberate: no nil check on Fire, because CheckTagEffects rejects a phase-route entry without one.
		mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
			Holder:       Holder{Faction: faction, Tag: tag},
			Asset:        asset,
			Entry:        entry,
			Roller:       roller,
			FactionState: factionState,
		})...)
	}
	return mutations
}
```

Replace `FirePurchase` in full:

```go
// FirePurchase runs every on_purchase entry of a just-bought asset, then of each tag
// the buyer holds, through its registrar's phase route. A direct definition read,
// not a hook: the asset was bought mid-turn, so it has no hooks until the next
// cycle. No active check: a bought asset is never ready, and its purchase effect is
// part of what was bought.
func (e *EffectsEngine) FirePurchase(faction *domain.Faction, asset *domain.Asset, factionState *state.FactionState, rulebook *rulebook.Rulebook, roller domain.Roller) []domain.Mutation {
	var mutations []domain.Mutation
	if def := rulebook.Assets[asset.DefinitionID]; def != nil {
		for _, entry := range def.Effects {
			if entry.Trigger != domain.TriggerOnPurchase {
				continue
			}
			// deliberate: no nil check on Fire, because RegisterEffectsList rejects a phase-route entry without one.
			mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
				Holder:       Holder{Faction: faction, Asset: asset, Def: def},
				Asset:        asset,
				Entry:        entry,
				Roller:       roller,
				FactionState: factionState,
			})...)
		}
	}
	for _, tag := range heldTags(faction, rulebook) {
		mutations = append(mutations, e.fireTag(faction, tag, domain.TriggerOnPurchase, asset, factionState, roller)...)
	}
	return mutations
}
```

Replace `FireMatchup` in full:

```go
// FireMatchup runs the entries for trigger of the fighting asset's definition
// (ctx.Holder on the way in), then of each tag its faction holds, through each
// registrar's matchup route. A direct definition read, raised by the Attack
// action, which owns the matchup.
func (e *EffectsEngine) FireMatchup(trigger domain.EffectTrigger, ctx MatchupContext, rulebook *rulebook.Rulebook) MatchupOutcome {
	var outcome MatchupOutcome
	fire := func(holder Holder, entries []domain.EffectDefinition) {
		for _, entry := range entries {
			if entry.Trigger != trigger {
				continue
			}
			ctx.Holder, ctx.Entry = holder, entry
			// deliberate: no nil check on Matchup, because RegisterEffectsList and CheckTagEffects reject a matchup-route entry without one.
			result := e.registrars[entry.Effect].Matchup(ctx)
			outcome.Mutations = append(outcome.Mutations, result.Mutations...)
			outcome.Counterstrike = outcome.Counterstrike || result.Counterstrike
		}
	}
	fire(ctx.Holder, ctx.Holder.Def.Effects)
	faction := ctx.Holder.Faction
	for _, tag := range heldTags(faction, rulebook) {
		fire(Holder{Faction: faction, Tag: tag}, tag.Effects)
	}
	return outcome
}
```

##### Task 3 — `internal/faction/engine/action/actions/attack.go`

**Find:**
```go
		attack.mutations = append(attack.mutations, attack.fireMatchup(domain.TriggerOwnAttack, faction, attacker, attackerDef, defenderFaction, defender, factionState).Mutations...)
```
**Replace with:**
```go
		attack.mutations = append(attack.mutations, attack.fireMatchup(domain.TriggerOwnAttack, faction, attacker, attackerDef, defenderFaction, defender, factionState, rulebook).Mutations...)
```

**Find:**
```go
			outcome := attack.fireMatchup(domain.TriggerOwnDefenseWin, defenderFaction, defender, defenderDef, faction, attacker, factionState)
```
**Replace with:**
```go
			outcome := attack.fireMatchup(domain.TriggerOwnDefenseWin, defenderFaction, defender, defenderDef, faction, attacker, factionState, rulebook)
```

**Find:**
```go
func (attack *AttackAction) fireMatchup(trigger domain.EffectTrigger, faction *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, opponent *domain.Faction, opponentAsset *domain.Asset, factionState *state.FactionState) effect.MatchupOutcome {
```
**Replace with:**
```go
func (attack *AttackAction) fireMatchup(trigger domain.EffectTrigger, faction *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, opponent *domain.Faction, opponentAsset *domain.Asset, factionState *state.FactionState, rulebook *rulebook.Rulebook) effect.MatchupOutcome {
```

**Find:**
```go
		OpponentAsset: opponentAsset,
		FactionState:  factionState,
	})
}
```
**Replace with:**
```go
		OpponentAsset: opponentAsset,
		FactionState:  factionState,
	}, rulebook)
}
```

##### Task 4 — `internal/faction/engine/effect/effect_test.go`

In `TestFireMatchup_FiresEntriesOfTheTriggerAndOrsCounterstrike`, the two calls
take a rulebook:

| Find | Replace with |
|---|---|
| `got := engine.FireMatchup(domain.TriggerOwnAttack, ctx)` | `got := engine.FireMatchup(domain.TriggerOwnAttack, ctx, &rulebook.Rulebook{})` |
| `got := engine.FireMatchup(domain.TriggerOwnDefenseWin, ctx)` | `got := engine.FireMatchup(domain.TriggerOwnDefenseWin, ctx, &rulebook.Rulebook{})` |

At the end of the file, add:

```go
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
```

##### Commit message

```
feat: register and fire tag effects

- ApplyAll registers each held tag's hook-route entries at its faction scope,
  after the faction's assets
- turn start, purchase and the Attack matchup fire held tags' entries after
  the asset's
- a held tag the rulebook does not know registers nothing
```

#### Commit 8 — `feat: ask asset hooks at ties and on the buyer's cost`

##### Task 1 — `internal/faction/engine/hooks/registry.go`

After `RollResultHooksAt`, add:

```go
// TieResolversAt returns the Cat 5 hooks registered in one scope bucket.
// ResolveTie decides which buckets a tie consults.
func (registry *Registry) TieResolversAt(scope Scope) []RegisteredTieResolver {
	return registry.tieResolvers[scope]
}
```

##### Task 2 — `internal/faction/engine/hooks/dispatch/rules.go`

Replace `ResolveAssetCost`'s doc comment and its first loop:

**Find:**
```go
// ResolveAssetCost runs the buyer's own AssetCostModifiers, then those registered
// against every asset standing on world — including a rival's, because a cost
// hook may belong to a faction other than the buyer (Local Investments). Hooks
// run in registration order, chaining outputs so each sees the previous result.
// onWorld is the assets on world in ID order (world.AssetsOnWorld), so the chain
// is the same every run; a nil slice adds nothing.
func ResolveAssetCost(registry *hooks.Registry, buyer *domain.Faction, def *domain.AssetDefinition, world string, baseCost int, onWorld []*domain.Asset) hooks.AssetCostQuery {
	query := &hooks.AssetCostQuery{Cost: baseCost, Applied: map[string]bool{}}
	if registry == nil {
		return *query
	}
	for _, registered := range registry.AssetCostModifiersFor(buyer.ID, "") {
		registered.Hook.ModifyAssetCost(buyer, def, world, query)
	}
```
**Replace with:**
```go
// ResolveAssetCost runs the buyer's global and faction AssetCostModifiers, then
// those of each of the buyer's own assets in ID order, then those registered
// against every rival asset standing on world, because a cost hook may belong to
// another faction (Local Investments). Hooks chain, each seeing the previous
// result. onWorld is the assets on world in ID order (world.AssetsOnWorld), so the
// chain is the same every run; a nil slice adds nothing.
func ResolveAssetCost(registry *hooks.Registry, buyer *domain.Faction, def *domain.AssetDefinition, world string, baseCost int, onWorld []*domain.Asset) hooks.AssetCostQuery {
	query := &hooks.AssetCostQuery{Cost: baseCost, Applied: map[string]bool{}}
	if registry == nil {
		return *query
	}
	for _, registered := range registry.AssetCostModifiersFor(buyer.ID, "") {
		registered.Hook.ModifyAssetCost(buyer, def, world, query)
	}
	for _, asset := range domain.SortedAssets(buyer) {
		for _, registered := range registry.AssetCostModifiersForAsset(buyer.ID, asset.ID) {
			registered.Hook.ModifyAssetCost(buyer, def, world, query)
		}
	}
```

Replace `ResolveTie` in full:

```go
// ResolveTie asks the global hooks, then the attacker's side, then the defender's:
// on each side its faction's hooks, then each of its assets' in ID order. The first
// answer that is not TieStandard wins, because with asset buckets in the walk a
// resolver that does not match must not end the search (Effort 1 Decision 4).
// TieStandard when none answers.
func ResolveTie(registry *hooks.Registry, ctx hooks.RollContext, factionState *state.FactionState) hooks.TieOutcome {
	if registry == nil {
		return hooks.TieStandard
	}
	scopes := []hooks.Scope{hooks.GlobalScope()}
	for _, side := range []*domain.Faction{ctx.Actor, ctx.Opponent} {
		if side == nil {
			continue
		}
		scopes = append(scopes, hooks.FactionScope(side.ID))
		for _, asset := range domain.SortedAssets(side) {
			scopes = append(scopes, hooks.AssetScope(side.ID, asset.ID))
		}
	}
	for _, scope := range scopes {
		for _, registered := range registry.TieResolversAt(scope) {
			if outcome := registered.Hook.ResolveTie(ctx, factionState); outcome != hooks.TieStandard {
				return outcome
			}
		}
	}
	return hooks.TieStandard
}
```

##### Task 3 — `internal/faction/engine/hooks/dispatch/rules_test.go`

**Find:**
```go
func TestResolveAssetCost_OwnAssetOnWorldSkipped(t *testing.T) {
	registry := hooks.NewRegistry()
	buyer := &domain.Faction{ID: "buyer"}
	def := &domain.AssetDefinition{ID: "cheap", Cost: 5}
	ownAsset := &domain.Asset{ID: "own-a1", OwnerID: "buyer"}
	registry.RegisterAssetCostModifier(hooks.AssetScope("buyer", "own-a1"), "surcharge", &stubAssetCostModifier{delta: -1})
	got := ResolveAssetCost(registry, buyer, def, "Tartarus", 5, []*domain.Asset{ownAsset}).Cost
	if got != 5 {
		t.Errorf("ResolveAssetCost = %d, want 5 (buyer's own asset is not asked)", got)
	}
}
```
**Replace with:**
```go
func TestResolveAssetCost_BuyersOwnAssetAskedOnce(t *testing.T) {
	registry := hooks.NewRegistry()
	ownAsset := &domain.Asset{ID: "own-a1", OwnerID: "buyer"}
	buyer := &domain.Faction{ID: "buyer", Assets: map[string]*domain.Asset{ownAsset.ID: ownAsset}}
	def := &domain.AssetDefinition{ID: "cheap", Cost: 5}
	registry.RegisterAssetCostModifier(hooks.AssetScope("buyer", "own-a1"), "discount", &stubAssetCostModifier{delta: 1})
	got := ResolveAssetCost(registry, buyer, def, "Tartarus", 5, []*domain.Asset{ownAsset}).Cost
	if got != 4 {
		t.Errorf("ResolveAssetCost = %d, want 4 (the buyer's own asset asked once, not again as a world asset)", got)
	}
}
```

After `TestResolveTie_DefenderWins`, add:

```go
func TestResolveTie_AssetResolverAnswers(t *testing.T) {
	registry := hooks.NewRegistry()
	attacker := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": {ID: "a1"}}}
	registry.RegisterTieResolver(hooks.AssetScope("f1", "a1"), "asset", &stubTieResolver{outcome: hooks.TieAttackerWins})
	if got := ResolveTie(registry, hooks.RollContext{Actor: attacker}, nil); got != hooks.TieAttackerWins {
		t.Errorf("ResolveTie = %v, want TieAttackerWins from the attacker's asset", got)
	}
}

func TestResolveTie_StandardDoesNotEndTheSearch(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterTieResolver(hooks.FactionScope("f1"), "silent", &stubTieResolver{outcome: hooks.TieStandard})
	registry.RegisterTieResolver(hooks.FactionScope("f2"), "loses", &stubTieResolver{outcome: hooks.TieAttackerWins})
	ctx := hooks.RollContext{Actor: &domain.Faction{ID: "f1"}, Opponent: &domain.Faction{ID: "f2"}}
	if got := ResolveTie(registry, ctx, nil); got != hooks.TieAttackerWins {
		t.Errorf("ResolveTie = %v, want the defender's TieAttackerWins past the attacker's TieStandard", got)
	}
}
```

##### Commit message

```
feat: ask asset hooks at ties and on the buyer's cost

- ResolveTie walks global, then each side's faction and asset buckets; the
  first answer that is not TieStandard wins
- ResolveAssetCost asks the buyer's own asset buckets after its faction's
```

### Phase 5 — Roll tags to data

**In plain words:** Warlike and Fanatical run from `tags.toml` entries, and their
Go handlers are gone. With no Go hook left that assumes "this roll is mine", the
roll dispatcher reads the opponent's faction bucket too.

#### Commit 9 — `feat: Warlike as rulebook data`

##### Task 1 — `internal/faction/engine/effect/effects/bonus_die.go`

`bonus_die` takes `scope = "faction"` and a `once_per` gate, and marks a tag's die.

**Find:**
```go
// bonusDieRegistrar reads owner = "own" as "the owner is rolling". scope =
// "self" is a roll by the declaring asset; "world" any owner roll on its world.
var bonusDieRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnRoll: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAttribute, domain.FilterPhase, domain.FilterAction},
		},
	},
```
**Replace with:**
```go
// bonusDieRegistrar reads owner = "own" as "the owner is rolling". scope =
// "self" is a roll by the declaring asset; "world" any owner roll on its world;
// "faction" any owner roll anywhere. A tag's die is FromTag, so Blackmail strips it.
var bonusDieRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnRoll: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterAttribute, domain.FilterPhase, domain.FilterAction},
			Gates:   []domain.EffectGateName{domain.GateOncePer},
		},
	},
```

Replace `registerBonusDie` and `bonusDieModifier` (from Phase 3) with:

```go
func registerBonusDie(ctx effect.RegisterContext) {
	ctx.Registry.RegisterRollModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialBonusDie),
		&bonusDieModifier{
			holder:  refOf(ctx.Holder),
			source:  ctx.Holder.Name(),
			trigger: ctx.Entry.Trigger,
			filters: ctx.Entry.Filters,
			oncePer: ctx.Entry.Gates.OncePer,
			params:  *ctx.Entry.BonusDie,
		},
	)
}

type bonusDieModifier struct {
	holder  holderRef
	source  string
	trigger domain.EffectTrigger
	filters domain.EffectFilters
	oncePer domain.OncePerKey
	params  domain.BonusDieParams
}
```

**Find:**
```go
	return []hooks.ModifierOffer{{
		Source:      modifier.source,
		Description: fmt.Sprintf("roll %d extra d%d, keep highest", params.Count, params.Sides),
		Apply: func(rollState *hooks.RollState) {
			for range params.Count {
				rollState.AddDie(params.Sides)
			}
			// deliberate: keep 1, not 1 + count, because every faction roll is one d10 and an extra die is pick-the-best ("an additional d10, keep the highest").
			rollState.SetKeepHighest(1)
		},
	}}
}
```
**Replace with:**
```go
	offer := hooks.ModifierOffer{
		Source:      modifier.source,
		Description: fmt.Sprintf("roll %d extra d%d, keep highest", params.Count, params.Sides),
		FromTag:     modifier.holder.tagID != "",
		Apply: func(rollState *hooks.RollState) {
			for range params.Count {
				rollState.AddDie(params.Sides)
			}
			// deliberate: keep 1, not 1 + count, because every faction roll is one d10 and an extra die is pick-the-best ("an additional d10, keep the highest").
			rollState.SetKeepHighest(1)
		},
	}
	if modifier.oncePer != "" {
		offer.BudgetKey = budgetKey(modifier.oncePer, domain.SpecialBonusDie, modifier.trigger, modifier.holder.sourceID, modifier.holder.assetID, otherSideID(ctx, false))
	}
	return []hooks.ModifierOffer{offer}
}
```

##### Task 2 — `internal/faction/engine/effect/effects/roll_match.go`

**Find:**
```go
// fires on the owner's own roll; enemy_roll_at_me on a rival's roll made against
// the owner. scope = "self" ties the roll to the declaring asset (rolling it for
// own_roll, rolled against for enemy_roll_at_me); scope = "world" to any such roll
// on the asset's world. The dispatcher reaches both sides' assets, so this check,
// not the registration scope, is what keeps an entry to its own side.
```
**Replace with:**
```go
// fires on the owner's own roll; enemy_roll_at_me on a rival's roll made against
// the owner. scope = "self" ties the roll to the declaring asset (rolling it for
// own_roll, rolled against for enemy_roll_at_me); scope = "world" to any such roll
// on the asset's world; scope = "faction" to any such roll, so asset is nil for a
// tag. The dispatcher reaches both sides, so this check, not the registration
// scope, is what keeps an entry to its own side.
```

**Find:**
```go
	case domain.ScopeWorld:
		if ctx.World == "" || ctx.World != asset.Location.WorldID {
			return false
		}
	default:
		return false
	}
```
**Replace with:**
```go
	case domain.ScopeWorld:
		if ctx.World == "" || ctx.World != asset.Location.WorldID {
			return false
		}
	case domain.ScopeFaction:
	default:
		return false
	}
```

##### Task 3 — `rulebooks/swn/tags.toml`

**Find:**
```toml
effect = "Once per turn, roll an additional d10 when making a Force attack."
```
**Replace with:**
```toml
effect = "Once per turn, roll an additional d10 when making a Force attack."

[[tags.Warlike.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[tags.Warlike.effects.filters]
owner     = "own"
scope     = "faction"
phase     = ["attack"]
attribute = "Force"
[tags.Warlike.effects.gates]
once_per = "owner_faction"
[tags.Warlike.effects.bonus_die]
sides = 10
count = 1
```

##### Task 4 — delete the Go handler

- Delete `internal/faction/engine/tag/tags/warlike.go` and `warlike_test.go`
  (`git rm`).
- `internal/faction/engine/tag/tag.go`, in `New`, delete the line
  `e.Register(tags.WarlikeHandler{})`.

##### Task 5 — tests

`internal/faction/engine/effect/effects/bonus_die_test.go` — at the end, add:

```go
func TestBonusDie_TagHolderOffersAFromTagDieOncePerTurn(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-020"}}
	beta := &domain.Faction{ID: "beta"}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": beta}}
	entry := bonusDieEntry(domain.EffectFilters{
		Owner: domain.OwnerOwn, Scope: domain.ScopeFaction,
		Attribute: domain.StatForce, Phases: []domain.EffectRollPhase{domain.RollPhaseAttack},
	}, 1)
	entry.Gates.OncePer = domain.OncePerOwnerFaction
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialBonusDie].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-020", Name: "Warlike"}}, Entry: entry, Registry: registry,
	})
	hook := registry.RollModifiersAt(hooks.FactionScope("alpha"))[0].Hook
	attack := hooks.RollContext{Phase: hooks.PhaseAttack, Actor: alpha, Opponent: beta, Attribute: "Force", World: "Krylos"}

	offers := hook.OfferModifiers(attack, factionState, nil)
	if len(offers) != 1 || !offers[0].FromTag || offers[0].Source != "Warlike" || offers[0].BudgetKey != "effect:bonus_die:own_roll:SWN-T-020" {
		t.Fatalf("offers = %+v, want one FromTag Warlike offer with the tag's budget key", offers)
	}

	collector := &rollCollector{accept: true}
	first := dispatch.RollWithHooks(attack, domain.DiceRoll{NumDice: 1, Sides: 10}, registry, collector, &seqRoller{values: []int{2, 9}}, alpha, factionState, nil)
	second := dispatch.RollWithHooks(attack, domain.DiceRoll{NumDice: 1, Sides: 10}, registry, collector, &seqRoller{values: []int{2}}, alpha, factionState, nil)
	if first.Sum != 9 || second.Sum != 2 {
		t.Errorf("sums = %d, %d; want 9 (die taken) then 2 (spent for the turn)", first.Sum, second.Sum)
	}

	alpha.Tags = nil
	if offers := hook.OfferModifiers(attack, factionState, nil); len(offers) != 0 {
		t.Errorf("offers after the tag is dropped = %+v, want none", offers)
	}
}
```

`internal/faction/engine/effect/effects/roll_match_test.go` — after the table loop,
add:

```go
	faction := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction}
	if !rollMatches(hooks.RollContext{Actor: alpha, World: "Tartarus"}, domain.TriggerOwnRoll, "alpha", nil, faction) {
		t.Error("faction scope: the owner's roll on any world, with no asset, does not match")
	}
	if rollMatches(hooks.RollContext{Actor: beta, World: "Tartarus"}, domain.TriggerOwnRoll, "alpha", nil, faction) {
		t.Error("faction scope: a rival's roll matches")
	}
```

`internal/faction/engine/testharness/scenarios/tags_test.go`:

**Find:**
```go
	budgets := h.FactionState.Factions["alpha"].HookBudgets
	testharness.CheckStep(t, "Warlike budget consumed once", budgets["tag:Warlike"] == 1,
		fmt.Sprintf("HookBudgets[tag:Warlike] = %d, want 1", budgets["tag:Warlike"]))
```
**Replace with:**
```go
	const warlikeBudget = "effect:bonus_die:own_roll:SWN-T-020"
	budgets := h.FactionState.Factions["alpha"].HookBudgets
	testharness.CheckStep(t, "Warlike budget consumed once", budgets[warlikeBudget] == 1,
		fmt.Sprintf("HookBudgets[%s] = %d, want 1", warlikeBudget, budgets[warlikeBudget]))
```

##### Commit message

```
feat: Warlike as rulebook data

- bonus_die accepts scope faction and a once_per gate; a tag's die is
  FromTag, so Blackmail still strips it
- rollMatches reads scope faction as any roll of the owner's
- Warlike's tags.toml entry replaces its Go handler
```

#### Commit 10 — `feat: Fanatical as rulebook data`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialLockOut            SpecialEffectType = "lock_out"
)
```
**Replace with:**
```go
	SpecialLockOut            SpecialEffectType = "lock_out"
	SpecialLoseTies           SpecialEffectType = "lose_ties"
)
```

**Find:**
```go
	TriggerPermissionGranted      EffectTrigger = "permission_granted"
)
```
**Replace with:**
```go
	TriggerPermissionGranted      EffectTrigger = "permission_granted"
	TriggerAttackTie              EffectTrigger = "attack_tie"
)
```

**Find:**
```go
type RerollParams struct {
	Dice     int
	Elective bool
}
```
**Replace with:**
```go
// RerollParams picks the dice a reroll targets: the Dice lowest (a forced
// reroll's highest), or every die showing Result. Exactly one is set.
type RerollParams struct {
	Dice     int
	Result   int
	Elective bool
}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
type rerollRecord struct {
	Dice     int  `toml:"dice"`
	Elective bool `toml:"elective"`
}
```
**Replace with:**
```go
type rerollRecord struct {
	Dice     int  `toml:"dice"`
	Result   int  `toml:"result"`
	Elective bool `toml:"elective"`
}
```

**Find:**
```go
		domain.TriggerPermissionQuery, domain.TriggerPermissionGranted,
	}
```
**Replace with:**
```go
		domain.TriggerPermissionQuery, domain.TriggerPermissionGranted, domain.TriggerAttackTie,
	}
```

**Find:**
```go
		domain.SpecialLockOut,
	}
```
**Replace with:**
```go
		domain.SpecialLockOut, domain.SpecialLoseTies,
	}
```

**Find:**
```go
		if record.Reroll.Dice <= 0 {
			return fmt.Errorf("reroll dice must be > 0, got %d", record.Reroll.Dice)
		}
		definition.Reroll = &domain.RerollParams{Dice: record.Reroll.Dice, Elective: record.Reroll.Elective}
```
**Replace with:**
```go
		dice, result := record.Reroll.Dice, record.Reroll.Result
		if dice < 0 || result < 0 || (dice > 0) == (result > 0) {
			return fmt.Errorf("reroll needs exactly one of dice or result > 0, got dice %d and result %d", dice, result)
		}
		definition.Reroll = &domain.RerollParams{Dice: dice, Result: result, Elective: record.Reroll.Elective}
```

##### Task 3 — effects

`internal/faction/engine/effect/effects/reroll.go`:

**Find:**
```go
// rerollRegistrar reads own_roll as the owner rerolling its own dice, and
// enemy_roll_at_me as the owner forcing a rival rolling against it to reroll.
// owner and scope read as in rollMatches.
var rerollRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnRoll: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
```
**Replace with:**
```go
// rerollRegistrar reads own_roll as the owner rerolling its own dice, and
// enemy_roll_at_me as the owner forcing a rival rolling against it to reroll.
// owner and scope read as in rollMatches.
var rerollRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnRoll: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeFaction},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
```

**Find:**
```go
		Indices:  pickDice(result.Dice, hook.params.Dice, forced),
```
**Replace with:**
```go
		Indices:  rerollIndices(result.Dice, hook.params, forced),
```

**Find:**
```go
// pickDice returns the indices of the n dice a reroll targets: the owner's own
// lowest dice, or a forced rival's highest.
```
**Replace with:**
```go
// rerollIndices returns the dice a reroll targets: every die showing
// params.Result (Fanatical's 1s), or else the params.Dice that pickDice picks.
func rerollIndices(dice []int, params domain.RerollParams, forced bool) []int {
	if params.Result == 0 {
		return pickDice(dice, params.Dice, forced)
	}
	var indices []int
	for i, die := range dice {
		if die == params.Result {
			indices = append(indices, i)
		}
	}
	return indices
}

// pickDice returns the indices of the n dice a reroll targets: the owner's own
// lowest dice, or a forced rival's highest.
```

`internal/faction/engine/effect/effects/lose_ties.go` (new):

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// loseTiesRegistrar is Fanatical's tie rule: the owner loses every tie in an
// attack, as attacker or defender. owner = "own" reads as "the owner is on either
// side of the tie", scope = "faction" as "anywhere".
var loseTiesRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerAttackTie: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerLoseTies,
}

func registerLoseTies(ctx effect.RegisterContext) {
	ctx.Registry.RegisterTieResolver(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialLoseTies), &loseTiesResolver{holder: refOf(ctx.Holder)})
}

type loseTiesResolver struct {
	holder holderRef
}

func (resolver *loseTiesResolver) ResolveTie(ctx hooks.RollContext, factionState *state.FactionState) hooks.TieOutcome {
	if _, live := resolver.holder.live(factionState); !live {
		return hooks.TieStandard
	}
	switch {
	case ctx.Actor != nil && ctx.Actor.ID == resolver.holder.factionID:
		return hooks.TieDefenderWins
	case ctx.Opponent != nil && ctx.Opponent.ID == resolver.holder.factionID:
		return hooks.TieAttackerWins
	}
	return hooks.TieStandard
}
```

`internal/faction/engine/effect/effects/registrars.go`:

**Find:**
```go
		domain.SpecialLockOut:            lockOutRegistrar,
```
**Replace with:**
```go
		domain.SpecialLockOut:            lockOutRegistrar,
		domain.SpecialLoseTies:           loseTiesRegistrar,
```

##### Task 4 — `rulebooks/swn/tags.toml`

**Find:**
```toml
effect = "Always rerolls any dice that come up as 1. Always loses ties during attacks."
```
**Replace with:**
```toml
effect = "Always rerolls any dice that come up as 1. Always loses ties during attacks."

[[tags.Fanatical.effects]]
trigger = "own_roll"
effect  = "reroll"
[tags.Fanatical.effects.filters]
owner = "own"
scope = "faction"
[tags.Fanatical.effects.reroll]
result   = 1
elective = false

[[tags.Fanatical.effects]]
trigger = "attack_tie"
effect  = "lose_ties"
[tags.Fanatical.effects.filters]
owner = "own"
scope = "faction"
```

##### Task 5 — delete the Go handler

- Delete `internal/faction/engine/tag/tags/fanatical.go` and `fanatical_test.go`
  (`git rm`).
- `internal/faction/engine/tag/tag.go`, in `New`, delete the line
  `e.Register(tags.FanaticalHandler{})`.

##### Task 6 — tests

`internal/faction/engine/effect/effects/reroll_test.go` — at the end, add:

```go
func TestReroll_TagHolderRerollsEveryDieShowingTheResult(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-005"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialReroll].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-005", Name: "Fanatical"}},
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerOwnRoll, Effect: domain.SpecialReroll,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
			Reroll:  &domain.RerollParams{Result: 1},
		},
		Registry: registry,
	})
	hook := registry.RollResultHooksAt(hooks.FactionScope("alpha"))[0].Hook
	roll := rollOn(factionState, "alpha", "beta", "Tartarus")

	directive := hook.OnRollResult(roll, hooks.RollResult{Dice: []int{1, 5, 1}}, factionState, nil)
	if !reflect.DeepEqual(directive.Indices, []int{0, 2}) || directive.Elective || directive.BudgetKey != "" || directive.Source != "Fanatical" {
		t.Errorf("directive = %+v, want indices [0 2], forced, unlimited, from Fanatical", directive)
	}
	if directive := hook.OnRollResult(rollOn(factionState, "beta", "alpha", "Tartarus"), hooks.RollResult{Dice: []int{1}}, factionState, nil); len(directive.Indices) != 0 {
		t.Errorf("rival's roll: directive = %+v, want none", directive)
	}
	alpha.Tags = nil
	if directive := hook.OnRollResult(roll, hooks.RollResult{Dice: []int{1}}, factionState, nil); len(directive.Indices) != 0 {
		t.Errorf("dropped tag: directive = %+v, want none", directive)
	}
}
```

`internal/faction/engine/effect/effects/lose_ties_test.go` (new):

```go
package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestLoseTies_OwnerLosesOnEitherSide(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-005"}}
	beta := &domain.Faction{ID: "beta"}
	gamma := &domain.Faction{ID: "gamma"}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": beta, "gamma": gamma}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialLoseTies].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-005", Name: "Fanatical"}},
		Entry: domain.EffectDefinition{Trigger: domain.TriggerAttackTie, Effect: domain.SpecialLoseTies,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction}},
		Registry: registry,
	})

	cases := []struct {
		name string
		ctx  hooks.RollContext
		want hooks.TieOutcome
	}{
		{"owner attacks", hooks.RollContext{Actor: alpha, Opponent: beta}, hooks.TieDefenderWins},
		{"owner defends", hooks.RollContext{Actor: beta, Opponent: alpha}, hooks.TieAttackerWins},
		{"owner turns an asset on itself", hooks.RollContext{Actor: alpha}, hooks.TieDefenderWins},
		{"others fight", hooks.RollContext{Actor: beta, Opponent: gamma}, hooks.TieStandard},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := dispatch.ResolveTie(registry, tc.ctx, factionState); got != tc.want {
				t.Errorf("ResolveTie = %v, want %v", got, tc.want)
			}
		})
	}

	alpha.Tags = nil
	if got := dispatch.ResolveTie(registry, hooks.RollContext{Actor: alpha, Opponent: beta}, factionState); got != hooks.TieStandard {
		t.Errorf("dropped tag: ResolveTie = %v, want TieStandard", got)
	}
}
```

`internal/faction/rulebook/effects_test.go`:

**Find:**
```go
		entry("own_roll", "reroll", "dice = 1\nelective = true")+
```
**Replace with:**
```go
		entry("own_roll", "reroll", "dice = 1\nelective = true")+
		entry("own_roll", "reroll", "result = 1")+
```

**Find:**
```go
	if len(effects) != 6 {
		t.Fatalf("effects: got %d, want 6", len(effects))
	}
```
**Replace with:**
```go
	if len(effects) != 7 {
		t.Fatalf("effects: got %d, want 7", len(effects))
	}
```

**Find:**
```go
		{"reroll", effects[2].Reroll, &domain.RerollParams{Dice: 1, Elective: true}},
		{"tech_level_floor", effects[3].TechLevelFloor, &domain.TechLevelFloorParams{Level: 4}},
		{"cost_surcharge", effects[4].CostSurcharge, &domain.CostSurchargeParams{Amount: 1}},
		{"restrict_attack", effects[5].RestrictAttack, &domain.RestrictAttackParams{Side: domain.RestrictDefense, Allow: domain.AllowOnlyListed}},
```
**Replace with:**
```go
		{"reroll", effects[2].Reroll, &domain.RerollParams{Dice: 1, Elective: true}},
		{"reroll result", effects[3].Reroll, &domain.RerollParams{Result: 1}},
		{"tech_level_floor", effects[4].TechLevelFloor, &domain.TechLevelFloorParams{Level: 4}},
		{"cost_surcharge", effects[5].CostSurcharge, &domain.CostSurchargeParams{Amount: 1}},
		{"restrict_attack", effects[6].RestrictAttack, &domain.RestrictAttackParams{Side: domain.RestrictDefense, Allow: domain.AllowOnlyListed}},
```

**Find:**
```go
[assets.Machine.effects.reroll]
dice = 0
`,
			wantErr: "reroll dice must be > 0, got 0",
		},
```
**Replace with:**
```go
[assets.Machine.effects.reroll]
dice = 0
`,
			wantErr: "reroll needs exactly one of dice or result > 0, got dice 0 and result 0",
		},
		{
			name: "reroll with dice and result",
			body: `
[[assets.Machine.effects]]
trigger = "own_roll"
effect  = "reroll"` + filters + `
[assets.Machine.effects.reroll]
dice   = 1
result = 1
`,
			wantErr: "reroll needs exactly one of dice or result > 0, got dice 1 and result 1",
		},
```

##### Commit message

```
feat: Fanatical as rulebook data

- reroll takes result = N: every die showing N (exactly one of dice, result)
- new lose_ties effect on a new attack_tie trigger: the owner loses ties
  as attacker and defender
- Fanatical's tags.toml entries replace its Go handler
```

#### Commit 11 — `feat: let the opponent's faction hooks reach a roll`

##### Task 1 — `internal/faction/engine/hooks/dispatch/roll.go`

**Find:**
```go
// RollWithHooks executes a dice roll with Cat 1 (RollModifier) hooks applied
// before rolling and Cat 2 (RollResultHook) hooks applied after. It consults the
// buckets rollScopes lists, so an opponent's asset hooks and hooks on the roll's
// world reach the roll. Each offer and directive is stamped with the faction that
// registered it, and its budget is read and spent on that faction.
```
**Replace with:**
```go
// RollWithHooks executes a dice roll with Cat 1 (RollModifier) hooks applied
// before rolling and Cat 2 (RollResultHook) hooks applied after. It consults the
// buckets rollScopes lists, so the opponent's faction and asset hooks and hooks on
// the roll's world reach the roll. Each offer and directive is stamped with the
// faction that registered it, and its budget is read and spent on that faction.
```

**Find:**
```go
// rollScopes lists the registry buckets a roll consults, in order: global, the
// roller's faction, the roller's assets, then the opponent's assets.
// deliberate: the opponent's faction bucket is left out, not added, because tag hooks register there and assume the roll is their own faction's (Fanatical would reroll the opponent's 1s).
func rollScopes(ctx hooks.RollContext, roller *domain.Faction) []hooks.Scope {
	scopes := []hooks.Scope{hooks.GlobalScope(), hooks.FactionScope(roller.ID)}
	scopes = append(scopes, assetScopes(roller, ctx.World, ctx.Asset)...)
	if ctx.Opponent != nil {
		scopes = append(scopes, assetScopes(ctx.Opponent, ctx.World, ctx.OpponentAsset)...)
	}
	return scopes
}
```
**Replace with:**
```go
// rollScopes lists the registry buckets a roll consults, in order: global, the
// roller's faction, the roller's assets, the opponent's faction, then the
// opponent's assets. Every S hook checks its owner relation itself (rollMatches),
// so a hook in the opponent's bucket acts only on rolls made against its owner.
func rollScopes(ctx hooks.RollContext, roller *domain.Faction) []hooks.Scope {
	scopes := []hooks.Scope{hooks.GlobalScope(), hooks.FactionScope(roller.ID)}
	scopes = append(scopes, assetScopes(roller, ctx.World, ctx.Asset)...)
	if ctx.Opponent != nil {
		scopes = append(scopes, hooks.FactionScope(ctx.Opponent.ID))
		scopes = append(scopes, assetScopes(ctx.Opponent, ctx.World, ctx.OpponentAsset)...)
	}
	return scopes
}
```

##### Task 2 — tests

`internal/faction/engine/hooks/dispatch/roll_test.go`:

**Find:**
```go
func TestRollWithHooks_OpponentFactionHookDoesNotFire(t *testing.T) {
	alpha := factionWithAssets("alpha", nil)
	beta := factionWithAssets("beta", map[string]string{"b-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.FactionScope("beta"), "tag:Warlike", addDieModifier("tag:Warlike", ""))

	result := RollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, failOnSelectCollector{t}, &fixedRoller{values: []int{5}}, alpha, rollState(alpha, beta), nil)

	if len(result.Dice) != 2 {
		t.Fatalf("expected 2 dice (opponent faction hook left out), got %d", len(result.Dice))
	}
}
```
**Replace with:**
```go
func TestRollWithHooks_OpponentFactionHookFires(t *testing.T) {
	alpha := factionWithAssets("alpha", nil)
	beta := factionWithAssets("beta", nil)
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.FactionScope("beta"), "beta-tag", addDieModifier("beta-tag", ""))
	collector := &recordingCollector{}

	result := RollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, &fixedRoller{values: []int{5}}, alpha, rollState(alpha, beta), nil)

	if len(result.Dice) != 3 {
		t.Fatalf("expected 3 dice (2 base + 1 from beta's faction hook), got %d", len(result.Dice))
	}
	if len(collector.offers) != 1 || collector.offers[0].FactionID != "beta" {
		t.Errorf("expected one offer stamped FactionID beta, got %+v", collector.offers)
	}
}
```

`internal/faction/engine/effect/effects/bonus_die_test.go` — at the end, add:

```go
func TestBonusDie_TagReachesNoRivalRollAgainstItsOwner(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-020"}}
	beta := &domain.Faction{ID: "beta"}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": beta}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialBonusDie].Register(effect.RegisterContext{
		Holder:   effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-020", Name: "Warlike"}},
		Entry:    bonusDieEntry(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction}, 1),
		Registry: registry,
	})
	collector := &rollCollector{accept: true}
	betaAttacks := hooks.RollContext{Phase: hooks.PhaseAttack, Actor: beta, Opponent: alpha, Attribute: "Force", World: "Krylos"}

	result := dispatch.RollWithHooks(betaAttacks, domain.DiceRoll{NumDice: 1, Sides: 10}, registry, collector, &seqRoller{values: []int{3}}, beta, factionState, nil)

	if result.Sum != 3 || collector.selections != 0 {
		t.Errorf("Sum = %d, selections = %d; want 3 and 0: alpha's tag is read but does not serve beta", result.Sum, collector.selections)
	}
}
```

##### Commit message

```
feat: let the opponent's faction hooks reach a roll

- rollScopes adds the opponent's faction bucket; the fence-sign goes, since
  no Go tag hook remains and every S hook checks its owner itself
- moved here from Effort 2 (Effort 1 Decision 4)
```

### Phase 6 — Economy tags to data

**In plain words:** Preceptor Archive and Scavengers run from `tags.toml`.
Scavengers now pays only for the holder's own kills and losses (`B.042`). The tag
engine is deleted, and Claude's sandbox campaign is re-seeded to the new shape.

#### Commit 12 — `feat: Preceptor Archive as rulebook data`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialLoseTies           SpecialEffectType = "lose_ties"
)
```
**Replace with:**
```go
	SpecialLoseTies           SpecialEffectType = "lose_ties"
	SpecialCostDiscount       SpecialEffectType = "cost_discount"
)
```

**Find:**
```go
	FilterStealthy  EffectFilterName = "stealthy"
)
```
**Replace with:**
```go
	FilterStealthy     EffectFilterName = "stealthy"
	FilterMinTechLevel EffectFilterName = "min_tech_level"
)
```

**Find:**
```go
	Categories []FactionStat
	Stealthy   *bool
}
```
**Replace with:**
```go
	Categories   []FactionStat
	Stealthy     *bool
	MinTechLevel int
}
```

**Find:**
```go
	if filters.Stealthy != nil {
		present = append(present, FilterStealthy)
	}
	return present
}
```
**Replace with:**
```go
	if filters.Stealthy != nil {
		present = append(present, FilterStealthy)
	}
	if filters.MinTechLevel > 0 {
		present = append(present, FilterMinTechLevel)
	}
	return present
}
```

**Find:**
```go
type DiscountRepairParams struct {
	Amount int
}
```
**Replace with:**
```go
type DiscountRepairParams struct {
	Amount int
}

type CostDiscountParams struct {
	Amount int
}
```

**Find:**
```go
	DiscountRepair *DiscountRepairParams
}
```
**Replace with:**
```go
	DiscountRepair *DiscountRepairParams
	CostDiscount   *CostDiscountParams
}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	DiscountRepair *discountRepairRecord `toml:"discount_repair"`
}
```
**Replace with:**
```go
	DiscountRepair *discountRepairRecord `toml:"discount_repair"`
	CostDiscount   *costDiscountRecord   `toml:"cost_discount"`
}
```

**Find:**
```go
	Category  []string `toml:"category"`
	Stealthy  *bool    `toml:"stealthy"`
}
```
**Replace with:**
```go
	Category     []string `toml:"category"`
	Stealthy     *bool    `toml:"stealthy"`
	MinTechLevel int      `toml:"min_tech_level"`
}
```

**Find:**
```go
type discountRepairRecord struct {
	Amount int `toml:"amount"`
}
```
**Replace with:**
```go
type discountRepairRecord struct {
	Amount int `toml:"amount"`
}

type costDiscountRecord struct {
	Amount int `toml:"amount"`
}
```

**Find:**
```go
		domain.SpecialLockOut, domain.SpecialLoseTies,
	}
```
**Replace with:**
```go
		domain.SpecialLockOut, domain.SpecialLoseTies, domain.SpecialCostDiscount,
	}
```

In `convertEffectFilters`:

**Find:**
```go
	filters := domain.EffectFilters{Owner: owner, Scope: scope, Action: record.Action, Stealthy: record.Stealthy}
```
**Replace with:**
```go
	if record.MinTechLevel < 0 {
		return domain.EffectFilters{}, fmt.Errorf("filters.min_tech_level must be >= 0, got %d", record.MinTechLevel)
	}
	filters := domain.EffectFilters{Owner: owner, Scope: scope, Action: record.Action, Stealthy: record.Stealthy, MinTechLevel: record.MinTechLevel}
```

In `presentParamBlocks`:

**Find:**
```go
	if record.DiscountRepair != nil {
		present = append(present, domain.SpecialDiscountRepair)
	}
	return present
```
**Replace with:**
```go
	if record.DiscountRepair != nil {
		present = append(present, domain.SpecialDiscountRepair)
	}
	if record.CostDiscount != nil {
		present = append(present, domain.SpecialCostDiscount)
	}
	return present
```

In `convertEffectParams`:

**Find:**
```go
		definition.DiscountRepair = &domain.DiscountRepairParams{Amount: record.DiscountRepair.Amount}
	}
	return nil
}
```
**Replace with:**
```go
		definition.DiscountRepair = &domain.DiscountRepairParams{Amount: record.DiscountRepair.Amount}
	case domain.SpecialCostDiscount:
		if record.CostDiscount == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.CostDiscount.Amount <= 0 {
			return fmt.Errorf("cost_discount amount must be > 0, got %d", record.CostDiscount.Amount)
		}
		definition.CostDiscount = &domain.CostDiscountParams{Amount: record.CostDiscount.Amount}
	}
	return nil
}
```

##### Task 3 — effects

`internal/faction/engine/effect/effects/cost_discount.go` (new):

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// costDiscountRegistrar lowers the owner's purchase cost by amount, never below 0.
// owner = "own" reads as "the owner is buying", scope = "faction" as "on any
// world". min_tech_level narrows it to definitions that need at least that tech
// level (Preceptor Archive: TL4+).
var costDiscountRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerAssetCostQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterMinTechLevel},
		},
	},
	Register: registerCostDiscount,
}

func registerCostDiscount(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetCostModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialCostDiscount),
		&costDiscountModifier{
			owner:        ctx.Holder.Faction,
			holder:       refOf(ctx.Holder),
			minTechLevel: ctx.Entry.Filters.MinTechLevel,
			amount:       ctx.Entry.CostDiscount.Amount,
		},
	)
}

// costDiscountModifier holds its owner: a Cat 4 hook gets no FactionState
// (Decision Record 19).
type costDiscountModifier struct {
	owner        *domain.Faction
	holder       holderRef
	minTechLevel int
	amount       int
}

func (modifier *costDiscountModifier) ModifyAssetCost(buyer *domain.Faction, def *domain.AssetDefinition, _ string, query *hooks.AssetCostQuery) {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || buyer.ID != modifier.owner.ID || def.TechLevel < modifier.minTechLevel {
		return
	}
	query.Cost = max(query.Cost-modifier.amount, 0)
}
```

`registrars.go`:

**Find:**
```go
		domain.SpecialLoseTies:           loseTiesRegistrar,
```
**Replace with:**
```go
		domain.SpecialLoseTies:           loseTiesRegistrar,
		domain.SpecialCostDiscount:       costDiscountRegistrar,
```

##### Task 4 — data, and delete the Go handler

`rulebooks/swn/tags.toml`:

**Find:**
```toml
effect = "Purchasing an asset requiring tech level 4+ costs one fewer Coin. May take the special action Teach Planetary Population (2 Coins, roll 1d12; on 12 the world's tech level permanently becomes 4 for this faction)."
```
**Replace with:**
```toml
effect = "Purchasing an asset requiring tech level 4+ costs one fewer Coin. May take the special action Teach Planetary Population (2 Coins, roll 1d12; on 12 the world's tech level permanently becomes 4 for this faction)."

[[tags.PreceptorArchive.effects]]
trigger = "asset_cost_query"
effect  = "cost_discount"
[tags.PreceptorArchive.effects.filters]
owner          = "own"
scope          = "faction"
min_tech_level = 4
[tags.PreceptorArchive.effects.cost_discount]
amount = 1
```

- Delete `internal/faction/engine/tag/tags/preceptor_archive.go` and
  `preceptor_archive_test.go` (`git rm`).
- `internal/faction/engine/tag/tag.go`, in `New`, delete the line
  `e.Register(tags.PreceptorArchiveHandler{})`.

##### Task 5 — tests

`internal/faction/engine/effect/effects/cost_discount_test.go` (new):

```go
package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

func TestCostDiscount_TagLowersTheOwnersTechLevelPurchases(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-013"}}
	beta := &domain.Faction{ID: "beta"}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialCostDiscount].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-013", Name: "Preceptor Archive"}},
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerAssetCostQuery, Effect: domain.SpecialCostDiscount,
			Filters:      domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, MinTechLevel: 4},
			CostDiscount: &domain.CostDiscountParams{Amount: 1},
		},
		Registry: registry,
	})
	tl4 := &domain.AssetDefinition{ID: "tl4", Cost: 4, TechLevel: 4}
	tl3 := &domain.AssetDefinition{ID: "tl3", Cost: 4, TechLevel: 3}
	free := &domain.AssetDefinition{ID: "free", Cost: 0, TechLevel: 5}

	cases := []struct {
		name  string
		buyer *domain.Faction
		def   *domain.AssetDefinition
		want  int
	}{
		{"TL4 purchase", alpha, tl4, 3},
		{"TL3 purchase", alpha, tl3, 4},
		{"never below 0", alpha, free, 0},
		{"a rival's purchase", beta, tl4, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := dispatch.ResolveAssetCost(registry, tc.buyer, tc.def, "Tartarus", tc.def.Cost, nil).Cost; got != tc.want {
				t.Errorf("cost = %d, want %d", got, tc.want)
			}
		})
	}

	alpha.Tags = nil
	if got := dispatch.ResolveAssetCost(registry, alpha, tl4, "Tartarus", 4, nil).Cost; got != 4 {
		t.Errorf("dropped tag: cost = %d, want 4", got)
	}
}
```

`internal/faction/domain/effect_test.go`:

**Find:**
```go
				Stealthy:   &stealthy,
			},
			want: []domain.EffectFilterName{
				domain.FilterAttribute, domain.FilterPhase, domain.FilterAction,
				domain.FilterAssetType, domain.FilterCategory, domain.FilterStealthy,
			},
```
**Replace with:**
```go
				Stealthy:     &stealthy,
				MinTechLevel: 4,
			},
			want: []domain.EffectFilterName{
				domain.FilterAttribute, domain.FilterPhase, domain.FilterAction,
				domain.FilterAssetType, domain.FilterCategory, domain.FilterStealthy,
				domain.FilterMinTechLevel,
			},
```
(Run `gofmt -l` on the file; re-align the struct literal's keys if it lists it.)

`internal/faction/rulebook/rulebook_test.go` — at the end of the shipped-rulebook
`TestLoad`, after the "government tag" subtest, add:

```go
	t.Run("preceptor archive discount", func(t *testing.T) {
		want := []domain.EffectDefinition{{
			Trigger:      domain.TriggerAssetCostQuery,
			Effect:       domain.SpecialCostDiscount,
			Filters:      domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, MinTechLevel: 4},
			CostDiscount: &domain.CostDiscountParams{Amount: 1},
		}}
		if got := rb.Tags["SWN-T-013"].Effects; !reflect.DeepEqual(got, want) {
			t.Errorf("Preceptor Archive effects = %+v, want %+v", got, want)
		}
	})
```

The scenario tests `TestPreceptorArchive_*` stay as they are: they now run the
data entry.

##### Commit message

```
feat: Preceptor Archive as rulebook data

- new cost_discount effect on asset_cost_query, with a min_tech_level filter;
  the cost never goes below 0
- Preceptor Archive's tags.toml entry replaces its Go handler
```

#### Commit 13 — `fix: Scavengers pays only for its own kills and losses`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	TriggerAttackTie              EffectTrigger = "attack_tie"
)
```
**Replace with:**
```go
	TriggerAttackTie              EffectTrigger = "attack_tie"
	TriggerAssetDestroyed         EffectTrigger = "asset_destroyed"
)
```

**Find:**
```go
	FilterMinTechLevel EffectFilterName = "min_tech_level"
)
```
**Replace with:**
```go
	FilterMinTechLevel EffectFilterName = "min_tech_level"
	FilterCausedBy     EffectFilterName = "caused_by"
)
```

**Find:**
```go
	Stealthy     *bool
	MinTechLevel int
}
```
**Replace with:**
```go
	Stealthy     *bool
	MinTechLevel int
	CausedBy     EffectOwner
}
```

**Find:**
```go
	if filters.MinTechLevel > 0 {
		present = append(present, FilterMinTechLevel)
	}
	return present
```
**Replace with:**
```go
	if filters.MinTechLevel > 0 {
		present = append(present, FilterMinTechLevel)
	}
	if filters.CausedBy != "" {
		present = append(present, FilterCausedBy)
	}
	return present
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
		domain.TriggerPermissionQuery, domain.TriggerPermissionGranted, domain.TriggerAttackTie,
	}
```
**Replace with:**
```go
		domain.TriggerPermissionQuery, domain.TriggerPermissionGranted, domain.TriggerAttackTie,
		domain.TriggerAssetDestroyed,
	}
```

**Find:**
```go
	MinTechLevel int      `toml:"min_tech_level"`
}
```
**Replace with:**
```go
	MinTechLevel int      `toml:"min_tech_level"`
	CausedBy     string   `toml:"caused_by"`
}
```

In `convertEffectFilters`:

**Find:**
```go
	for i, value := range record.Phase {
```
**Replace with:**
```go
	if record.CausedBy != "" {
		causedBy, err := toEnum("filters.caused_by", record.CausedBy, effectOwners)
		if err != nil {
			return domain.EffectFilters{}, err
		}
		filters.CausedBy = causedBy
	}
	for i, value := range record.Phase {
```

##### Task 3 — `internal/faction/engine/effect/effects/coin_gain.go`

Replace the file in full:

```go
package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// coinGainRegistrar pays the owner the amount. turn_start fires at the owner's
// turn start. asset_destroyed reacts to each asset destroyed in a fight: owner
// reads as whose asset it was ("own" the owner's, "rival" another faction's), and
// caused_by as who destroyed it. scope = "faction" is any world.
var coinGainRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerTurnStart: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
		domain.TriggerAssetDestroyed: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn, domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterCausedBy},
		},
	},
	Register: registerCoinGain,
	Fire:     fireCoinGain,
}

func fireCoinGain(ctx effect.FireContext) []domain.Mutation {
	return []domain.Mutation{domain.CoinDelta{
		FactionID:         ctx.Holder.Faction.ID,
		Delta:             ctx.Entry.CoinGain.Amount.Roll(ctx.Roller),
		Cause:             string(domain.SpecialCoinGain),
		CausedByFactionID: ctx.Holder.Faction.ID,
	}}
}

func registerCoinGain(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialCoinGain),
		&coinGainReactor{
			holder:   refOf(ctx.Holder),
			owner:    ctx.Entry.Filters.Owner,
			causedBy: ctx.Entry.Filters.CausedBy,
			amount:   ctx.Entry.CoinGain.Amount,
			roller:   ctx.Roller,
		},
	)
}

// combatLossCauses are the removal causes that mean an asset was destroyed in a
// fight: an attack or counterattack, or a Cracked Comms self-attack.
var combatLossCauses = []string{"attack", string(domain.SpecialCounterstrike)}

type coinGainReactor struct {
	holder   holderRef
	owner    domain.EffectOwner
	causedBy domain.EffectOwner
	amount   domain.DiceRoll
	roller   domain.Roller
}

// OnMutations pays once for each combat removal in the batch that the entry's
// owner and caused_by filters match.
func (reactor *coinGainReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	ownerID := reactor.holder.factionID
	var extra []domain.Mutation
	for _, mutation := range mutations {
		removed, ok := mutation.(domain.AssetRemoved)
		if !ok || !slices.Contains(combatLossCauses, removed.Cause) {
			continue
		}
		if !onSide(reactor.owner, removed.FactionID, ownerID) {
			continue
		}
		if reactor.causedBy != "" && !onSide(reactor.causedBy, removed.CausedByFactionID, ownerID) {
			continue
		}
		extra = append(extra, domain.CoinDelta{
			FactionID:         ownerID,
			Delta:             reactor.amount.Roll(reactor.roller),
			Cause:             string(domain.SpecialCoinGain),
			CausedByFactionID: ownerID,
		})
	}
	return extra
}

// onSide reports whether factionID stands on side relative to ownerID: "own" is
// the owner, "rival" another faction, "any" either.
func onSide(side domain.EffectOwner, factionID, ownerID string) bool {
	switch side {
	case domain.OwnerOwn:
		return factionID == ownerID
	case domain.OwnerRival:
		return factionID != "" && factionID != ownerID
	}
	return true
}
```

##### Task 4 — data, and delete the Go handler

`rulebooks/swn/tags.toml`:

**Find:**
```toml
effect = "Whenever the faction destroys an asset or has one of their assets destroyed, gain 1 Coin."
```
**Replace with:**
```toml
effect = "Whenever the faction destroys an asset or has one of their assets destroyed, gain 1 Coin."

[[tags.Scavengers.effects]]
trigger = "asset_destroyed"
effect  = "coin_gain"
[tags.Scavengers.effects.filters]
owner = "own"
scope = "faction"
[tags.Scavengers.effects.coin_gain]
amount = "1"

[[tags.Scavengers.effects]]
trigger = "asset_destroyed"
effect  = "coin_gain"
[tags.Scavengers.effects.filters]
owner     = "rival"
scope     = "faction"
caused_by = "own"
[tags.Scavengers.effects.coin_gain]
amount = "1"
```

- Delete `internal/faction/engine/tag/tags/scavengers.go` and `scavengers_test.go`
  (`git rm`).
- `internal/faction/engine/tag/tag.go`, in `New`, delete the line
  `e.Register(tags.ScavengersHandler{})`. `New` now registers nothing; Commit 14
  deletes the package.

##### Task 5 — tests

`internal/faction/engine/effect/effects/coin_gain_test.go` — at the end, add:

```go
func TestCoinGain_ScavengersPaysOwnKillsAndLossesOnly(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-016"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}, "gamma": {ID: "gamma"}}}
	registry := hooks.NewRegistry()
	holder := effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-016", Name: "Scavengers"}}
	for _, filters := range []domain.EffectFilters{
		{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		{Owner: domain.OwnerRival, Scope: domain.ScopeFaction, CausedBy: domain.OwnerOwn},
	} {
		effects.Registrars()[domain.SpecialCoinGain].Register(effect.RegisterContext{
			Holder: holder, Registry: registry, Roller: fixedRoller{value: 1},
			Entry: domain.EffectDefinition{Trigger: domain.TriggerAssetDestroyed, Effect: domain.SpecialCoinGain,
				Filters: filters, CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}}},
		})
	}
	paid := func(removed domain.AssetRemoved) int {
		emitted, err := dispatch.MutationReactors(registry, factionState.Factions["beta"], []domain.Mutation{removed}, factionState, nil)
		if err != nil {
			t.Fatalf("MutationReactors: %v", err)
		}
		total := 0
		for _, mutation := range emitted {
			if gain, ok := mutation.(domain.CoinDelta); ok && gain.FactionID == "alpha" && gain.Cause == "coin_gain" {
				total += gain.Delta
			}
		}
		return total
	}

	cases := []struct {
		name    string
		removed domain.AssetRemoved
		want    int
	}{
		{"own loss in an attack", domain.AssetRemoved{FactionID: "alpha", Cause: "attack", CausedByFactionID: "beta"}, 1},
		{"own kill in an attack", domain.AssetRemoved{FactionID: "beta", Cause: "attack", CausedByFactionID: "alpha"}, 1},
		{"own kill by Cracked Comms", domain.AssetRemoved{FactionID: "beta", Cause: "counterstrike", CausedByFactionID: "alpha"}, 1},
		{"a fight between two others (B.042)", domain.AssetRemoved{FactionID: "beta", Cause: "attack", CausedByFactionID: "gamma"}, 0},
		{"a sale", domain.AssetRemoved{FactionID: "alpha", Cause: "sell", CausedByFactionID: "alpha"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := paid(tc.removed); got != tc.want {
				t.Errorf("paid %d, want %d", got, tc.want)
			}
		})
	}

	alpha.Tags = nil
	if got := paid(domain.AssetRemoved{FactionID: "alpha", Cause: "attack", CausedByFactionID: "beta"}); got != 0 {
		t.Errorf("dropped tag: paid %d, want 0", got)
	}
}
```

Add these to that file's imports:
`"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"`,
`"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"`
and `"github.com/therobertcrocker/gm-toolkit/internal/faction/state"`.

`internal/faction/domain/effect_test.go`:

**Find:**
```go
				MinTechLevel: 4,
			},
			want: []domain.EffectFilterName{
				domain.FilterAttribute, domain.FilterPhase, domain.FilterAction,
				domain.FilterAssetType, domain.FilterCategory, domain.FilterStealthy,
				domain.FilterMinTechLevel,
			},
```
**Replace with:**
```go
				MinTechLevel: 4,
				CausedBy:     domain.OwnerOwn,
			},
			want: []domain.EffectFilterName{
				domain.FilterAttribute, domain.FilterPhase, domain.FilterAction,
				domain.FilterAssetType, domain.FilterCategory, domain.FilterStealthy,
				domain.FilterMinTechLevel, domain.FilterCausedBy,
			},
```

`internal/faction/engine/testharness/scenarios/tags_test.go` — the payout's cause
is `coin_gain` now (two places):

| Find | Replace with |
|---|---|
| `if _, ok := testharness.FindMutationByTypeAndCause(records, "coin_delta", "scavengers"); !ok {` | `if _, ok := testharness.FindMutationByTypeAndCause(records, "coin_delta", "coin_gain"); !ok {` |
| `t.Error("expected coin_delta with cause=scavengers in history")` | `t.Error("expected coin_delta with cause=coin_gain in history")` |
| `if _, ok := testharness.FindMutationByTypeAndCause(records, "coin_delta", "scavengers"); ok {` | `if _, ok := testharness.FindMutationByTypeAndCause(records, "coin_delta", "coin_gain"); ok {` |
| `t.Error("expected no coin_delta with cause=scavengers when Scavengers tag is absent")` | `t.Error("expected no coin_delta with cause=coin_gain when Scavengers tag is absent")` |

##### Commit message

```
fix: Scavengers pays only for its own kills and losses

- coin_gain gains an asset_destroyed reactor route; owner says whose asset
  died, the new caused_by filter says who destroyed it
- Scavengers' two tags.toml entries replace its Go handler, so a fight
  between two other factions no longer pays it (closes B.042)
```

#### Commit 14 — `refactor: delete the tag engine`

##### Task 1 — `internal/faction/engine/core.go`

Delete the `tag` import line
`"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/tag"`, the
`Tag      *tag.TagEngine` field of `Engine`, and the line `e.Tag = tag.New(log)`
in `NewWithRulebook`. Run `gofmt -w` on the file to re-align the struct.

##### Task 2 — `internal/faction/engine/orchestrator.go`

**Find:**
```go
	e.Hooks.Reset()
	e.Tag.ApplyAll(factionState, e.Hooks, e.log.With("engine", "tag"))
	e.Effect.ApplyAll(factionState, e.Rulebook, e.Hooks, e.Rand, e.log.With("engine", "effect"))
```
**Replace with:**
```go
	e.Hooks.Reset()
	e.Effect.ApplyAll(factionState, e.Rulebook, e.Hooks, e.Rand, e.log.With("engine", "effect"))
```

##### Task 3 — delete the package, then re-seed sandbox-claude

- `git rm -r internal/faction/engine/tag`. The directory now holds only `tag.go`.
- Re-seed Claude's campaign (gitignored; nothing to commit):
  1. `go run ./cmd/gm-toolkit campaign add-rules --campaign sandbox-claude --rules rulebooks/swn --replace`
  2. In `campaigns/sandbox-claude/state/state.seed.toml` and each
     `campaigns/sandbox-claude/state/scenarios/*.toml`, use Edit to turn every tag
     list of inline tables into a list of prefixed IDs. Example:
     `tags = [ { id = "T-011", name = "Planetary Government", … }, ]` becomes
     `tags = ["SWN-T-011"]`.
  3. `./cmd/scripts/reset-campaign.sh sandbox-claude`
  4. Run one cycle against `--campaign sandbox-claude` (tmux, per the run skill),
     and confirm the campaign loads and the cycle completes.

##### Commit message

```
refactor: delete the tag engine

- engine/tag and Engine.Tag are gone; the effects engine registers every
  [[effects]] entry, asset or tag (D-14)
```
