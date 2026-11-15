# H.A.W.K Asset Authoring — Effort 5 Plan: Cunning Pass

## In Plain Words

- **What** — H.A.W.K's own Cunning assets: 28 spies, cults, fronts and dirty tricks that replace the SWN copy in `rulebooks/hawk/assets/cunning_assets.toml`. They are written as corporate sci-fantasy, lean into corporate espionage and cults, and are built on SWN's balance. Two small engine changes come first. A sacrifice can be limited to some asset types, and a new "shelter" ability lets an asset hide one of its faction's own assets on purpose.
- **Why** — Cunning is the second asset category H.A.W.K owns. Robert's brief is a real take on what the Cunning factions of this setting need for their nefarious work, not SWN's assets with H.A.W.K names. SWN gives only the tiers and the numbers.
- **How it fits** — Effort 5 of six in `A.001.5`, the last initiative of arc `A.001`. There are two phases: the engine changes, then the Cunning data, the seed swap, the canon and a live turn on the Effort 3 board. Effort 6 does the same for Wealth.

## Context / Goal

Design the Cunning pass (lore-gate, reflavor, sort, review; discovery D-9) and specify the engine work and the data to the task.

- Discovery doc: [`A.001.5-hawk-asset-authoring-discovery.md`](../discovery/A.001.5-hawk-asset-authoring-discovery.md), D-7 to D-11
- Overview plan: [`A.001.5-hawk-asset-authoring-overview-plan.md`](A.001.5-hawk-asset-authoring-overview-plan.md), Effort 5, Pass shape, Live turn
- Effort 4 plan: [`A.001.5.E4-hawk-asset-authoring-plan.md`](A.001.5.E4-hawk-asset-authoring-plan.md), Decisions 2, 3, 11, 13 and 20 to 23 (voice, types, the stand-in type gap, descriptions, speed classes, teleport ranges)
- Canon: `campaigns/hawk/canon/`; the drafts in gitignored `scratch/A.001.5.E5-cunning-drafts*.md` (pass 3 is the signed-off set)

## Decisions Ratified in Planning

1. **Concept first; SWN gives the tiers and the numbers** — Robert's brief: "This isn't 'SWN assets but HAWK flavor' this is a real take at what cunning assets would look like in a corporate sci-fantasy setting. Use the balance, tiers, etc from SWN - but that's it." Each row starts from what H.A.W.K's Cunning factions need, then takes the numbers of a same-tier SWN row. Three rows take a Wealth row's numbers: Prosperity Gospel (Local Investments), Charm Peddlers (Harvesters) and Geas-Backed NDA (Marketers). The result is 28 rows, 9 of them ★ (no SWN Cunning asset does it). Rejected: pass 1, whose names described the SWN slot they replaced.
2. **Voice: as Effort 4 Decision 2** — Corporate sci-fantasy played straight; the humour comes from the blend, never from a punchline; fit the setting, don't cite it. Name forms vary: single words, compounds and titles. No row is named after its SWN slot's job, except two kept by choice: Black-hull Smugglers and Stealth. The final names are Robert's, set over three drafts.
3. **Corporate espionage and cults, with less religion** — Robert's calls. Cults were to carry equal weight, then Robert asked for less religion in the names, so lodges, pacts, seminars and founders replace churches. The final lean is 9 cult rows and 19 espionage rows, accepted.
4. **A new H.A.W.K type: Front** — Robert's call. A business or place that covers what the faction does there; seven rows (Occult Lodge, Mirrorways, Ascension Program, Leased Demiplane, WhisperGate™ Network, Augury Desk, Oracle Engine). It replaces SWN's Logistics Facility for Cunning. `asset_types` gains Front. SWN's six types stay for the Wealth stand-in until Effort 6. Rejected: Holding; Network; reusing Facility.
5. **A shared mechanic is fine; a shared concept is not** — Robert's rule. Ritual Scapegoats were cut, because Unpaid Interns already are "a cheap person thrown in front of a killing blow", and Glamoured Doubles take the slot. 18 rows run the mechanic of their SWN number source unchanged. Robert accepts that mirroring, because each mechanic fits its new concept. Rejected: reworking mechanics only to differ.
6. **Glamoured Doubles guard Operatives only** — Robert's call. `sacrifice_self` takes an `asset_type` filter, as `shelter_asset` does; Unpaid Interns, with no filter, still guard every type. A parameterize row, Commit 1.
7. **Leased Demiplane is a hiding place, not a rescue** — Robert's calls. As an action, one of the faction's Operatives or Security Personnel on its place goes inside; it holds one asset. Inside, the asset has the existing `sheltered` condition: hidden, untouchable, unable to act or be moved. The owner may bring it out free at the start of its turn. If damage destroys the Demiplane, the asset inside is destroyed too. The Boltholes rescue is dropped. Built in Commit 2 as a `shelter` ability effect, a free release in the turn-start lift step, and an `asset_destroyed` route on `shelter_asset` that registers only its loss reactor. Rejected: keeping the rescue too; capacity 2 or unlimited; Operatives only; release by action.
8. **A door shelter is known by its source** — The release is offered when the sheltering asset's definition has a shelter step. A Demiplane that leaves play other than by combat (a sale, a refit, unpaid upkeep) lets its asset out as it goes; one destroyed in combat takes its asset with it. An asset Boltholes sheltered at 0 HP still waits for a repair, even once the Boltholes is gone. No new state. Amended at effort close: the release once also opened for any source that no longer exists, which freed a sold Boltholes' asset at 0 HP (Decision Record 4). Accepted quirk: a repair to full clears any shelter (Effort 2 Decision 12), so a damaged asset inside a Demiplane that is repaired comes out early. Reversed at the initiative close: a repair never lets an asset out of a standing door shelter (Decision Record 7).
9. **Stealth stays as is** — Robert's call. SWN's quality, pointed at Operative, at `HAWK-C3-002`; he wants a plain "add Stealth" in H.A.W.K and accepts the duplicate. This closes Effort 4 Decision 11's gap for Force's Operatives. Its description is the rules alone, and its canon entry is one plain line.
10. **Geas-Backed NDA disables a rival asset** — Robert's call. Marketers' `disable_asset` and numbers at tier 5 (`HAWK-C5-004`), with the test and the attack set to Cunning vs. Cunning. Marketers attack Wealth. Rejected: an NDA that grants Stealth (now Stealth's job); `ignore_coin_loss` (Bank's).
11. **Fey Courtiers enforce their attack limit** — Robert's call at the lore gate. `restrict_attack`, so only Operatives can attack them. SWN's Seductress had the limit in text only.
12. **Movers in Effort 4's classes** — Robert's calls. Black-hull Smugglers: speed 14 (mid), drift 2, carry one Operative for 1 Coin, `ignores_lockdown`. A tier 1 skiff lands past a tier 6 Rules of Hospitality, accepted; rejected: `ignores_permission`. Fey Courtiers jump 4 hexes and have no speed. Mirrorways move one Operative 12 hexes. WhisperGate™ Network moves any number of Operatives in its region.
13. **Four reshaped mechanics come from other rows** — Occult Lodge takes Surveyors' (Wealth) extra d10 on Expand Influence, a `faction_test` die already in the SWN data, so the pairing is not untested as pass 2 feared. Wardbreakers start Stealthed (Pact-Bound Assassins'). Membership Dues steal Coin on a hit (Hexcannon Gunships'). Prophet-Executive adds Battle Choir's once-a-turn d10 on Cunning attacks to Demagogue's attacker: a power gain at the same cost, accepted. Tuning waits for play.
14. **Charm Peddlers and Prosperity Gospel bring Wealth's tier 1 economy into Cunning** — accepted; every faction can now buy them from either category.
15. **Six Fey rows stay** — Robert's call. Old Bargains, Fey Courtiers, Mirrorways, Dreamwalkers, Rules of Hospitality and Changeling Agents draw on the Fey; the Fey Royals are a Cunning faction.
16. **WhisperGate™ Network keeps the ™** — Robert's call. Literal in the name and the canon. Renamed from WhisperWay, which clashed with Mirrorways.
17. **The seed swap is by role** — Robert's call. Each seeded stand-in becomes the row that takes its numbers, so seeded HP stays valid. Lobbyists' numbers feed two rows: O.W.L.'s becomes Occult Lodge, and Seraph's becomes Old Bargains (Seraph's seat is `veil`). Informers become Imp Familiars (`SWN-C1-002` to `HAWK-C1-001`). Every other stand-in keeps its suffix. The Radiant Throne also gains one Glamoured Doubles beside its Imp Familiars, so the build shows in play: 18 seeded Cunning units. Amended in execution: O.W.L.'s Occult Lodge sits at Eventide, not Haven (Decision Record 2).
18. **Two phases** — overview Decision 6: build rows first (Decisions 6 and 7), then the data.
19. **F.053 stays in the Backlog** — Robert's call. Imp Familiars inherit Informers' faction pick.
20. **The Wealth stand-in's type gap stays until Effort 6** — Effort 4 Decision 11. Cunning's half closes: the heirs of Covert Shipping, Boltholes and Covert Transit Net name Operatives.

## Decision Record — Execution

1. **Commit 4: Claude plays the live turn on `hawk-claude`** — Robert's call; reverses overview Decision 4 for this effort. Robert asked Claude to run the TUI itself, with debug logs. Claude ran the prep on `hawk` (Robert approved it in chat), refreshed its gitignored copy `campaigns/hawk-claude` from the new rulebook and seed, and played with `--debug`. The seeded board puts every faction alone on its seat except Veil, so most rival checks need a staged board: a gitignored scenario, `e5-cunning`, moves O.W.L.'s Occult Lodge to Lumos and the Court's Diabolic Interlopers, Membership Dues and Prophet-Executive to Solar Gate. Every check in Task 2 passed. The Leased Demiplane offered only the Shield Wardens, hid them, and at the next turn start asked "Bring Shield Wardens out of hiding?" and let them out free. The Fey Courtiers jumped 4 hexes and revealed. The Prophet-Executive's d10 came on the first Cunning attack and showed "used this turn" on the second. Glamoured Doubles took a killing blow for the Imp Familiars. The Occult Lodge's d10 came on Expand Influence. The Interlopers took 1d4 Coin and the Marines left the attacker list. Also seen: Augury Desk's reroll and forced reroll, Membership Dues stealing 4 Coin, and Old Bargains' 1 Coin upkeep.
2. **Commit 4: O.W.L.'s Occult Lodge is seeded at Eventide** — Robert's call; amends Decision 17. Expand Influence places a base only where the faction has assets and no base (`worldsForNewBase`, `internal/faction/engine/action/actions/expand_influence.go`), and the Lodge's die fires only on its own place. On Haven, O.W.L.'s seat, the seeded Lodge could never fire, and a Front cannot be moved. Eventide is 6 hexes from Haven, ungoverned and empty, so O.W.L. can expand there at once. Rejected: Arcadia (O.W.L. already has a base there); Veil and Ouroboros (a contested expand); leaving it on Haven.
3. **Commit 4: three older findings go to the tracker, none fixed here** — Robert's calls. None is an E5 fault; each is older than E5 and only shows with H.A.W.K's rows. A self-jump step (Fey Courtiers, as SWN's Seductress) offers "— do not move —", but `relocate.go` requires a self-jump to move, so the empty pick fails the action with `ErrNoSelection` and the faction loses its turn: added to `F.024`, which already holds the empty-form failure. Transit Web's granted move asks "Fey Courtiers — relocate to?" with no word of its source, so it read as the Courtiers' own jump: `B.052`. The feed's `Coin +1` names no faction (War Profiteers' gain for Seraph), and the debug log's mutation line names no asset: `B.053`.
4. **Effort close: a lost Demiplane lets its asset out at once; a lost source no longer opens a release** — Robert's call; amends Decision 8. The turn-start release opened for any source that no longer exists, and it could not tell a sold Demiplane from a sold Boltholes, so a Boltholes asset at 0 HP came out free and skipped its repair. Now the loss reactor knows a door shelter by its definition at registration: a non-combat loss clears its occupants' `sheltered` condition (`shelter_released`), and the turn-start release opens only for a source that exists and has a shelter step. Rejected: a new field on the condition to name the source's kind.
5. **Effort close: an occupied door shelter cannot be relocated** — Claude's call, from the review. Relocation skips a Demiplane that holds an asset, so the asset inside never ends up on another world. Only door shelters: a Boltholes' asset is not inside it, and blocking a Boltholes would change an SWN rule. Movement and transport need no guard today: a Front has no speed, and no transport carries one. Not true by the initiative close: Effort 6's Free Traders and Merchant Convoy carry Fronts, and Aether Burners give any asset a hot speed of 5 (Decision Record 6).
6. **Initiative close: an occupied door shelter cannot move or ride either** — from the initiative review. Effort 6 added `Front` to two Wealth transports' cargo lists, and Aether Burners let any asset move itself hot, so a Leased Demiplane with an asset inside could take a move order or board as cargo and leave its occupant on the old place. `domain.HoldsShelteredAsset` is now the one check, read by relocate, the move order list (`eligibleMovableAssets`) and the cargo list (`eligibleCargoForTransport`). Completes entry 5.
7. **Initiative close: healing never lets an asset out of a door shelter** — Robert's call; reverses Decision 8's accepted quirk. An asset inside a Leased Demiplane leaves only of its own will, by the free release at its owner's turn start. A repair to full still ends a Boltholes shelter, or one whose source is gone. `domain.InDoorShelter` is the one test, read by the release offer and by Repair Asset.
8. **Initiative close: a shelter ability needs its loss entry at load** — from the initiative review. A definition with a `shelter` ability step and no `asset_destroyed` `shelter_asset` effect would strand its occupant when the shelter is lost, since nothing clears the condition. `convertAsset` (`internal/faction/rulebook/rulebook.go`) refuses it. H.A.W.K's Leased Demiplane has both.

## Sort Table

Primitives are the registrar names in `internal/faction/engine/effect/effects/registrars.go`, the ability effects in `internal/faction/domain/asset.go`, or a transport profile. Surface: `—` plain, `effects` S-side, `ability` A-side, `transport`. ★ = no SWN Cunning asset does it.

| HAWK ID | Name | SWN ancestor | Surface | Call | Primitive | Notes |
|---------|------|--------------|---------|------|-----------|-------|
| HAWK-C1-001 | Imp Familiars | Informers | ability | reuse | `reveal_stealth` | Operative, A |
| HAWK-C1-002 | Glamoured Doubles | False Front | effects | parameterize | `sacrifice_self` + `asset_type` | Operative; guards Operatives only (Decision 6); Commit 1 |
| HAWK-C1-003 | Black-hull Smugglers | Smugglers | transport | reuse | transport profile + `ignores_lockdown` | ★ Aethership, mid (14), drift 2 (Decision 12) |
| HAWK-C1-004 | Prosperity Gospel | Local Investments (Wealth) | effects | reuse | `cost_surcharge` | ★ Stratagem |
| HAWK-C1-005 | Charm Peddlers | Harvesters (Wealth) | ability | reuse | `coin_delta` | ★ Operative, A |
| HAWK-C2-001 | Old Bargains | Lobbyists | effects | reuse | `withdraw_permission` | Stratagem, maintenance 1 |
| HAWK-C2-002 | Skeleton Keys | Saboteurs | effects | reuse | `block_ability` | Stratagem |
| HAWK-C2-003 | Corporate Auditors | Blackmail | effects | reuse | `remove_tag_dice` | Operative |
| HAWK-C2-004 | Fey Courtiers | Seductress | ability, effects | reuse | `relocate` + `reveal_stealth`; `restrict_attack` | Operative, A; 4 hexes, no speed; Operatives only may attack it (Decision 11) |
| HAWK-C2-005 | Occult Lodge | Lobbyists; Surveyors' die (Wealth) | effects | reuse | `bonus_die`, `faction_test`, Expand Influence | ★ Front (Decision 13) |
| HAWK-C3-001 | Wardbreakers | Cyberninjas | effects | reuse | `apply_stealth` | ★ Operative |
| HAWK-C3-002 | Stealth | Stealth | effects | reuse | `grant_stealth` | quality, for an Operative (Decision 9) |
| HAWK-C3-003 | Mirrorways | Covert Shipping | ability | reuse | `relocate` | Front, A; one Operative, 12 hexes |
| HAWK-C4-001 | Ascension Program | Party Machine | effects | reuse | `coin_gain` | Front |
| HAWK-C4-002 | Chosen Initiates | Vanguard Cadres | — | reuse | — | Security Personnel |
| HAWK-C4-003 | Trained Hellhounds | Tripwire Cells | effects | reuse | `clear_stealth` | Beast |
| HAWK-C4-004 | Diabolic Interlopers | Seditionists | ability | reuse | `block_attack` | Operative, A |
| HAWK-C5-001 | Membership Dues | Organization Moles | effects | reuse | `coin_steal` | ★ Stratagem (Decision 13) |
| HAWK-C5-002 | Dreamwalkers | Cracked Comms | effects | reuse | `counterstrike` | Operative |
| HAWK-C5-003 | Leased Demiplane | Boltholes | ability, effects | build | `shelter`; `shelter_asset` on `asset_destroyed` | ★ Front, A; capacity 1 (Decision 7); Commit 2 |
| HAWK-C5-004 | Geas-Backed NDA | Marketers (Wealth) | ability | reuse | `disable_asset` | ★ Stratagem, A; Cunning vs. Cunning (Decision 10) |
| HAWK-C6-001 | Rules of Hospitality | Transport Lockdown | effects | reuse | `lock_out` | Stratagem |
| HAWK-C6-002 | WhisperGate™ Network | Covert Transit Net | ability | reuse | `relocate`, region | Front, A; Operatives |
| HAWK-C6-003 | Prophet-Executive | Demagogue; Battle Choir's die | effects | reuse | `bonus_die`, Cunning attacks | ★ Operative, `once_per = "asset"` (Decision 13) |
| HAWK-C7-001 | State Faith | Popular Movement | effects | reuse | `grant_permission` | Stratagem |
| HAWK-C7-002 | Augury Desk | Book of Secrets | effects | reuse | `reroll` | Front |
| HAWK-C7-003 | Changeling Agents | Treachery | effects | reuse | `capture_target` | Stratagem |
| HAWK-C8-001 | Oracle Engine | Panopticon Matrix | effects | reuse | `bonus_die`, `clear_stealth` | Front, maintenance 2 |

## Shared Context

**Gate.** Every commit lands green on `go build ./...`, `go vet ./...` and `go test ./...` from the repo root. `TestNewWithRulebook_ShippedRulebook` (`internal/faction/engine/core_test.go`) loads `rulebooks/swn` and `rulebooks/hawk` and builds the engine from each, so it checks every new effect, filter, type name and ability block. `TestHawkSeed_MatchesRulebookAndMap` (`internal/faction/engine/hawk_seed_test.go`) checks that every seeded asset's definition is in the H.A.W.K rulebook, so the TOML and the seed swap land in one commit.

**Campaign commands** take the prefix `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns` and run after `go build -o gm-toolkit ./cmd/gm-toolkit` (overview Decision 5).

## Out of Scope

- **Tuning numbers** — Decisions 13 and 14; tuning waits for play.
- **SWN's Boltholes** — unchanged; it keeps its rescue (Decision 8).
- **The sacrifice's SWN wording** — the TUI's "false front" label and the narrative's "A false front fell in place of the asset it covered" (`internal/faction/narrative/wire_templates.go`) also show for Unpaid Interns and Glamoured Doubles, as they have since Effort 4.
- **The Wealth stand-in** — Effort 6 (Decision 20).
- **F.053** — Decision 19.
- **Regenerating the seed for real play** — after the arc closes (Effort 3 Decision 15).

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — The sacrifice filter and the shelter door**
  - Commit 1 — `feat(E5): let a sacrifice guard only listed asset types` (2 tasks)
  - Commit 2 — `feat(E5): add a shelter ability that hides an own asset` (12 tasks)
- [x] **Phase 2 — The Cunning pass**
  - Commit 3 — `feat(E5): replace the Cunning stand-in with H.A.W.K's own assets` (3 tasks)
  - Commit 4 — `docs(E5): write the Cunning canon and play the live turn` (4 tasks)

### Phase 1 — The sacrifice filter and the shelter door

**In plain words:** A sacrifice can be limited to some asset types, and an asset with the new shelter ability can hide one of its faction's own assets, let it out free at turn start, and take it down with it when destroyed.

#### Commit 1 — `feat(E5): let a sacrifice guard only listed asset types`

##### Task 1 — `internal/faction/engine/effect/effects/sacrifice_self.go`

Rewritten — full contents (Decision 6):

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// sacrificeSelfRegistrar is False Front: when damage would destroy another own
// asset on its world, the owner may sacrifice the False Front to nullify the
// killing blow, and the struck asset keeps its HP from before that hit (Effort 2
// Decision 14). owner = "own" reads as "the struck asset is the declaring asset's
// faction's", scope = "world" as "on the declaring asset's world". An asset_type
// filter guards only the listed types (Glamoured Doubles guard Operatives); with
// none it guards every type (False Front, Unpaid Interns).
var sacrificeSelfRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerLethalDamage: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
		},
	},
	Register: registerSacrificeSelf,
}

func registerSacrificeSelf(ctx effect.RegisterContext) {
	interceptor := &sacrificeSelfInterceptor{holder: refOf(ctx.Holder), assetTypes: ctx.Entry.Filters.AssetTypes}
	ctx.Registry.RegisterLethalInterceptor(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialSacrificeSelf), interceptor)
}

type sacrificeSelfInterceptor struct {
	holder     holderRef
	assetTypes []domain.AssetType
}

func (interceptor *sacrificeSelfInterceptor) OfferInterception(ctx hooks.LethalContext) *hooks.Interception {
	self, live := interceptor.holder.live(ctx.FactionState)
	if !live || !guardsTarget(self, ctx, interceptor.assetTypes) {
		return nil
	}
	cause := domain.SpecialSacrificeSelf
	// deliberate: the removal's cause is the sacrifice, not "attack", because the owner gave the False Front up and no one scores a kill (Effort 2 Decision 14).
	return &hooks.Interception{
		Source: self,
		Ask:    hooks.InterceptSacrifice,
		Mutations: append(hpTo(ctx, ctx.HPAfter+ctx.Damage, cause, self.OwnerID), domain.AssetRemoved{
			FactionID:         self.OwnerID,
			AssetID:           self.ID,
			Cause:             string(cause),
			CausedByFactionID: self.OwnerID,
		}),
	}
}
```

##### Task 2 — `internal/faction/engine/effect/effects/sacrifice_self_test.go`

A filtered sacrifice guards only its types; an unfiltered one still guards a Facility.

**Find:**
```go
			if offer := falseFrontOffer(t, factionState, hooks.LethalContext{Target: target, TargetDef: militaryUnitDef, HPAfter: -1, Damage: 4}); offer != nil {
				t.Errorf("offer = %+v, want nil", offer)
			}
		})
	}
}
```
**Replace with:**
```go
			if offer := falseFrontOffer(t, factionState, hooks.LethalContext{Target: target, TargetDef: militaryUnitDef, HPAfter: -1, Damage: 4}); offer != nil {
				t.Errorf("offer = %+v, want nil", offer)
			}
		})
	}
}

func TestSacrificeSelf_AssetTypeFilter(t *testing.T) {
	t.Run("an unfiltered sacrifice guards a Facility", func(t *testing.T) {
		factionState := falseFrontState()
		x := factionState.Factions["alpha"].Assets["x"]
		if offer := falseFrontOffer(t, factionState, hooks.LethalContext{Target: x, TargetDef: facilityDef, HPAfter: -1, Damage: 4}); offer == nil {
			t.Error("offer = nil, want one")
		}
	})

	cases := []struct {
		name      string
		targetDef *domain.AssetDefinition
		wantOffer bool
	}{
		{"a listed type is guarded", militaryUnitDef, true},
		{"an unlisted type is not", facilityDef, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := falseFrontState()
			alpha := factionState.Factions["alpha"]
			registry := hooks.NewRegistry()
			effects.Registrars()[domain.SpecialSacrificeSelf].Register(effect.RegisterContext{
				Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["ff"], Def: &domain.AssetDefinition{ID: "HAWK-C1-002"}}, Registry: registry,
				Entry: domain.EffectDefinition{
					Trigger: domain.TriggerLethalDamage,
					Effect:  domain.SpecialSacrificeSelf,
					Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld, AssetTypes: []domain.AssetType{"Military Unit"}},
				},
			})
			interceptors := registry.LethalInterceptorsForAsset("alpha", "ff")
			if len(interceptors) != 1 {
				t.Fatalf("interceptors = %d, want 1", len(interceptors))
			}
			offer := interceptors[0].Hook.OfferInterception(hooks.LethalContext{Target: alpha.Assets["x"], TargetDef: tc.targetDef, HPAfter: -1, Damage: 4, FactionState: factionState})
			if (offer != nil) != tc.wantOffer {
				t.Errorf("offer = %+v, want an offer: %v", offer, tc.wantOffer)
			}
		})
	}
}
```

##### Commit message

```
feat(E5): let a sacrifice guard only listed asset types

- sacrifice_self accepts an asset_type filter and guards only those
  types, as shelter_asset does; with none it guards every type
- tests: a filtered sacrifice skips an unlisted type; an unfiltered
  one still guards a Facility
```

#### Commit 2 — `feat(E5): add a shelter ability that hides an own asset`

##### Task 1 — `internal/faction/domain/asset.go`

The effect name:

**Find:**
```go
	EffectRaiseTechLevel   AbilityEffectType = "raise_tech_level"
)
```
**Replace with:**
```go
	EffectRaiseTechLevel   AbilityEffectType = "raise_tech_level"
	EffectShelter          AbilityEffectType = "shelter"
)
```

The step's parameters:

**Find:**
```go
type RaiseTechLevelParams struct {
	CoinCost  int
	TechLevel int
}
```
**Replace with:**
```go
type RaiseTechLevelParams struct {
	CoinCost  int
	TechLevel int
}

// ShelterParams configures a shelter step (Leased Demiplane): AssetTypes are the
// own asset types it may take in, and Capacity is how many it holds at once.
type ShelterParams struct {
	AssetTypes []AssetType
	Capacity   int
}
```

**Find:**
```go
	RaiseTechLevel *RaiseTechLevelParams
}
```
**Replace with:**
```go
	RaiseTechLevel *RaiseTechLevelParams
	Shelter        *ShelterParams
}
```

The definition reports its shelter step, for the turn-start release (Task 8):

**Find:**
```go
func (def *AssetDefinition) HalfPrice() int {
	return max(1, def.Cost/2)
}
```
**Replace with:**
```go
func (def *AssetDefinition) HalfPrice() int {
	return max(1, def.Cost/2)
}

// Shelter returns the parameters of the definition's shelter step (Leased
// Demiplane), or nil when it has none.
func (def *AssetDefinition) Shelter() *ShelterParams {
	if def == nil || def.Ability == nil {
		return nil
	}
	for _, step := range def.Ability.Steps {
		if step.Effect == EffectShelter {
			return step.Shelter
		}
	}
	return nil
}
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

The record:

**Find:**
```go
	RaiseTechLevel *raiseTechLevelRecord `toml:"raise_tech_level"`
}

type raiseTechLevelRecord struct {
```
**Replace with:**
```go
	RaiseTechLevel *raiseTechLevelRecord `toml:"raise_tech_level"`
	Shelter        *shelterRecord        `toml:"shelter"`
}

type shelterRecord struct {
	AssetTypes []string `toml:"asset_types"`
	Capacity   int      `toml:"capacity"`
}

type raiseTechLevelRecord struct {
```

In `convertAbilityStep`, the block converts:

**Find:**
```go
		step.RaiseTechLevel = &domain.RaiseTechLevelParams{CoinCost: raise.CoinCost, TechLevel: raise.TechLevel}
	}
```
**Replace with:**
```go
		step.RaiseTechLevel = &domain.RaiseTechLevelParams{CoinCost: raise.CoinCost, TechLevel: raise.TechLevel}
	}
	if shelter := record.Shelter; shelter != nil {
		if len(shelter.AssetTypes) == 0 || shelter.Capacity <= 0 {
			return domain.AbilityStep{}, fmt.Errorf("shelter needs asset_types and capacity > 0, got %d types and capacity %d", len(shelter.AssetTypes), shelter.Capacity)
		}
		params := &domain.ShelterParams{Capacity: shelter.Capacity}
		for _, name := range shelter.AssetTypes {
			params.AssetTypes = append(params.AssetTypes, domain.AssetType(name))
		}
		step.Shelter = params
	}
```

and the step's shape is checked:

**Find:**
```go
		return domain.AbilityStep{}, fmt.Errorf("levy_coin requires target = %q and no contest or die", "faction")
	}
```
**Replace with:**
```go
		return domain.AbilityStep{}, fmt.Errorf("levy_coin requires target = %q and no contest or die", "faction")
	}
	if step.Effect == domain.EffectShelter && (step.Target != domain.TargetAsset || step.Shelter == nil || step.Contest != nil || step.Amount != nil || step.Table != nil || step.TargetRoll != nil) {
		return domain.AbilityStep{}, fmt.Errorf("shelter requires target = %q, a [shelter] block and no contest or die", "asset")
	}
	if step.Shelter != nil && step.Effect != domain.EffectShelter {
		return domain.AbilityStep{}, fmt.Errorf("a [shelter] block needs effect = %q, got %q", domain.EffectShelter, step.Effect)
	}
```

In `toAbilityEffect`:

**Find:**
```go
	case "raise_tech_level":
		return domain.EffectRaiseTechLevel, nil
```
**Replace with:**
```go
	case "raise_tech_level":
		return domain.EffectRaiseTechLevel, nil
	case "shelter":
		return domain.EffectShelter, nil
```

In `checkAssetTypes`, a shelter's types must be declared:

**Find:**
```go
// asset type and excluded types, and an effects entry's asset_type filter, on an
// asset or a tag.
```
**Replace with:**
```go
// asset type and excluded types, a shelter's asset types, and an effects entry's
// asset_type filter, on an asset or a tag.
```

**Find:**
```go
		for i, step := range ability.Steps {
			if err := checkRelocate(fmt.Sprintf("%s ability step %d", holder, i), step.Relocate); err != nil {
				return err
			}
		}
```
**Replace with:**
```go
		for i, step := range ability.Steps {
			stepHolder := fmt.Sprintf("%s ability step %d", holder, i)
			if err := checkRelocate(stepHolder, step.Relocate); err != nil {
				return err
			}
			if step.Shelter != nil {
				if err := check(stepHolder, step.Shelter.AssetTypes...); err != nil {
					return err
				}
			}
		}
```

##### Task 3 — `internal/faction/engine/effect/effects/shelter_asset.go`

An `asset_destroyed` route registers only the loss reactor, so a Leased Demiplane takes its occupant down with it and rescues no one (Decision 7).

**Find:**
```go
// declaring asset's faction's", scope = "world" as "on the declaring asset's world".
var shelterAssetRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerLethalDamage: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
		},
	},
	Register: registerShelterAsset,
}

func registerShelterAsset(ctx effect.RegisterContext) {
	scope := ctx.Holder.Scope()
	source := ctx.Holder.HookID(domain.SpecialShelterAsset)
	holder := refOf(ctx.Holder)
	ctx.Registry.RegisterLethalInterceptor(scope, source, &shelterAssetInterceptor{holder: holder, assetTypes: ctx.Entry.Filters.AssetTypes})
	ctx.Registry.RegisterMutationReactor(scope, source, &shelterLossReactor{holder: holder})
}
```
**Replace with:**
```go
// declaring asset's faction's", scope = "world" as "on the declaring asset's world".
// The asset_destroyed route (owner = "own", scope = "self") keeps only the second
// rule: an asset that shelters by its ability (Leased Demiplane) takes what it
// shelters with it when damage destroys it, and rescues no one.
var shelterAssetRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerLethalDamage: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
		},
		domain.TriggerAssetDestroyed: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Register: registerShelterAsset,
}

func registerShelterAsset(ctx effect.RegisterContext) {
	scope := ctx.Holder.Scope()
	source := ctx.Holder.HookID(domain.SpecialShelterAsset)
	holder := refOf(ctx.Holder)
	if ctx.Entry.Trigger == domain.TriggerLethalDamage {
		ctx.Registry.RegisterLethalInterceptor(scope, source, &shelterAssetInterceptor{holder: holder, assetTypes: ctx.Entry.Filters.AssetTypes})
	}
	ctx.Registry.RegisterMutationReactor(scope, source, &shelterLossReactor{holder: holder})
}
```

##### Task 4 — `internal/faction/engine/action/actions/ability/shelter.go` and `dispatch.go`

`shelter.go` (new) — full contents:

```go
package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// shelter is Leased Demiplane: the target, one of the faction's own assets on the
// holder's world, goes inside. A sheltered asset is hidden, untouchable and cannot
// act (domain.Asset.Targetable). Its owner may bring it out free at its turn start
// (turn.OfferShelterReleases), and it is destroyed if damage destroys the holder
// (shelter_asset's asset_destroyed route).
func shelter(ctx StepContext) ([]domain.Mutation, error) {
	if ctx.Asset == nil {
		return nil, fmt.Errorf("shelter: needs an asset holder")
	}
	return []domain.Mutation{domain.AssetConditionApplied{
		FactionID:         ctx.Faction.ID,
		AssetID:           ctx.Target.Asset.ID,
		Condition:         domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: ctx.Asset.ID},
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}}, nil
}
```

`dispatch.go`:

**Find:**
```go
	domain.EffectRaiseTechLevel:   raiseTechLevel,
}
```
**Replace with:**
```go
	domain.EffectRaiseTechLevel:   raiseTechLevel,
	domain.EffectShelter:          shelter,
}
```

##### Task 5 — `internal/faction/engine/action/actions/ability/stage.go`

The target stage reads the rulebook, because a shelter's candidates are filtered by their definitions' types.

**Find:**
```go
	worldEngine *world.WorldEngine,
	registry *hooks.Registry,
) (domain.AbilityTarget, error) {
	candidates, picks := stepCandidates(step, faction, holder.Asset, factionState, worldEngine, registry)
```
**Replace with:**
```go
	worldEngine *world.WorldEngine,
	registry *hooks.Registry,
	rulebook *rulebook.Rulebook,
) (domain.AbilityTarget, error) {
	candidates, picks := stepCandidates(step, faction, holder.Asset, factionState, worldEngine, registry, rulebook)
```

**Find:**
```go
// HasTargets reports whether the first step of a holder's ability can pick a
// target now. A step that picks no target always can.
func HasTargets(faction *domain.Faction, holder domain.AbilityHolder, abilityDefinition *domain.AbilityDefinition, factionState *state.FactionState, worldEngine *world.WorldEngine, registry *hooks.Registry) bool {
	if abilityDefinition == nil || len(abilityDefinition.Steps) == 0 {
		return true
	}
	// deliberate: only the first step, not every step, because a later step's candidates depend on what earlier steps change (Seductress moves before it reveals).
	candidates, picks := stepCandidates(abilityDefinition.Steps[0], faction, holder.Asset, factionState, worldEngine, registry)
```
**Replace with:**
```go
// shelterCandidates returns the faction's own assets on the acting asset's world
// that a shelter step may take in, in asset-ID order: of a listed type, not the
// holder, not sheltered already, and with no open move order. None while the
// holder already shelters Capacity assets, or is in transit.
func shelterCandidates(step domain.AbilityStep, faction *domain.Faction, acting *domain.Asset, rulebook *rulebook.Rulebook) []*domain.Asset {
	params := step.Shelter
	if params == nil || rulebook == nil || acting == nil || acting.Location.WorldID == "" {
		return nil
	}
	inside := domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: acting.ID}
	held := 0
	for _, asset := range domain.SortedAssets(faction) {
		if slices.ContainsFunc(asset.Conditions, inside.Is) {
			held++
		}
	}
	if held >= params.Capacity {
		return nil
	}
	var candidates []*domain.Asset
	for _, candidate := range domain.SortedAssets(faction) {
		if candidate.ID == acting.ID || candidate.Location.WorldID != acting.Location.WorldID ||
			candidate.CurrentOrder != nil || candidate.HasCondition(domain.ConditionSheltered) {
			continue
		}
		if def := rulebook.Assets[candidate.DefinitionID]; def != nil && slices.Contains(params.AssetTypes, def.Type) {
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

// HasTargets reports whether the first step of a holder's ability can pick a
// target now. A step that picks no target always can.
func HasTargets(faction *domain.Faction, holder domain.AbilityHolder, abilityDefinition *domain.AbilityDefinition, factionState *state.FactionState, worldEngine *world.WorldEngine, registry *hooks.Registry, rulebook *rulebook.Rulebook) bool {
	if abilityDefinition == nil || len(abilityDefinition.Steps) == 0 {
		return true
	}
	// deliberate: only the first step, not every step, because a later step's candidates depend on what earlier steps change (Seductress moves before it reveals).
	candidates, picks := stepCandidates(abilityDefinition.Steps[0], faction, holder.Asset, factionState, worldEngine, registry, rulebook)
```

**Find:**
```go
func stepCandidates(step domain.AbilityStep, faction *domain.Faction, acting *domain.Asset, factionState *state.FactionState, worldEngine *world.WorldEngine, registry *hooks.Registry) (candidates domain.AbilityTargetCandidates, picks bool) {
	switch step.Target {
	case domain.TargetFaction:
		return domain.AbilityTargetCandidates{Factions: factionCandidates(step.Effect, faction.ID, acting, factionState)}, true
	case domain.TargetAsset:
		return domain.AbilityTargetCandidates{Assets: assetCandidates(step.Effect, faction.ID, acting, factionState)}, true
```
**Replace with:**
```go
func stepCandidates(step domain.AbilityStep, faction *domain.Faction, acting *domain.Asset, factionState *state.FactionState, worldEngine *world.WorldEngine, registry *hooks.Registry, rulebook *rulebook.Rulebook) (candidates domain.AbilityTargetCandidates, picks bool) {
	switch step.Target {
	case domain.TargetFaction:
		return domain.AbilityTargetCandidates{Factions: factionCandidates(step.Effect, faction.ID, acting, factionState)}, true
	case domain.TargetAsset:
		// deliberate: own assets, not rivals', because a shelter hides the faction's own asset; every other asset step acts on a rival.
		if step.Effect == domain.EffectShelter {
			return domain.AbilityTargetCandidates{Assets: shelterCandidates(step, faction, acting, rulebook)}, true
		}
		return domain.AbilityTargetCandidates{Assets: assetCandidates(step.Effect, faction.ID, acting, factionState)}, true
```

##### Task 6 — the target stage's callers

`internal/faction/engine/action/actions/ability/dispatch.go`:

**Find:**
```go
		target, err := resolveTarget(step, workingFaction, domain.AbilityHolder{Asset: workingAsset, Tag: holder.Tag}, view, collector, worldEngine, registry)
```
**Replace with:**
```go
		target, err := resolveTarget(step, workingFaction, domain.AbilityHolder{Asset: workingAsset, Tag: holder.Tag}, view, collector, worldEngine, registry, rulebook)
```

`internal/faction/engine/action/actions/use_asset_ability.go`, in `abilityCandidates`:

**Find:**
```go
ability.HasTargets(faction, holder, def.Ability, factionState, worldEngine, registry) {
```
**Replace with:**
```go
ability.HasTargets(faction, holder, def.Ability, factionState, worldEngine, registry, rulebook) {
```

**Find:**
```go
		if ability.HasTargets(faction, holder, tag.Ability, factionState, worldEngine, registry) {
```
**Replace with:**
```go
		if ability.HasTargets(faction, holder, tag.Ability, factionState, worldEngine, registry, rulebook) {
```

The tests pass `nil`. In `internal/faction/engine/action/actions/ability/stage_test.go`, with the Edit tool and `replace_all`:

| Find | Replace with | Lines |
|------|--------------|-------|
| `factionState, collector, nil, nil)` | `factionState, collector, nil, nil, nil)` | 3 |
| `factionState, collector, engine, nil)` | `factionState, collector, engine, nil, nil)` | 1 |
| `factionState, nil, nil) {` | `factionState, nil, nil, nil) {` | 4 |

In `internal/faction/engine/action/actions/ability/raise_tech_level_test.go`:

**Find:**
```go
tag.Ability, factionState, engine, nil) {
```
**Replace with:**
```go
tag.Ability, factionState, engine, nil, nil) {
```

##### Task 7 — `internal/faction/engine/hooks/collector.go`

**Find:**
```go
	PaymentRestore PaymentReason = "restore"
)
```
**Replace with:**
```go
	PaymentRestore PaymentReason = "restore"
	// PaymentReleaseShelter: bring an asset out of hiding at the owner's turn start,
	// at no cost (Leased Demiplane).
	PaymentReleaseShelter PaymentReason = "release_shelter"
)
```

##### Task 8 — `internal/faction/engine/turn/conditions.go`

The free release (Decisions 7 and 8):

**Find:**
```go
				domain.AssetConditionCleared{FactionID: faction.ID, AssetID: asset.ID, Condition: condition, Cause: "condition_lifted", CausedByFactionID: faction.ID},
			)
		}
	}
	return mutations, nil
}
```
**Replace with:**
```go
				domain.AssetConditionCleared{FactionID: faction.ID, AssetID: asset.ID, Condition: condition, Cause: "condition_lifted", CausedByFactionID: faction.ID},
			)
		}
	}
	return mutations, nil
}

// OfferShelterReleases asks faction, for each of its sheltered assets in ID order
// that may leave, whether to bring it out, at no cost (Effort 5 Decision 7).
func OfferShelterReleases(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, collector hooks.PaymentCollector) ([]domain.Mutation, error) {
	var mutations []domain.Mutation
	for _, asset := range domain.SortedAssets(faction) {
		for _, condition := range asset.Conditions {
			if condition.Kind != domain.ConditionSheltered || !releasable(condition, factionState, rulebook) {
				continue
			}
			release, err := collector.ConfirmPayment(hooks.PaymentOffer{Payer: faction, Asset: asset, Amount: 0, Reason: hooks.PaymentReleaseShelter})
			if err != nil {
				return nil, fmt.Errorf("offer shelter releases: %w", err)
			}
			if !release {
				continue
			}
			mutations = append(mutations, domain.AssetConditionCleared{FactionID: faction.ID, AssetID: asset.ID, Condition: condition, Cause: "shelter_released", CausedByFactionID: faction.ID})
		}
	}
	return mutations, nil
}

// releasable reports whether a sheltered condition opens at its owner's turn
// start: its source has a shelter ability (Leased Demiplane), or its source no
// longer exists. An asset Boltholes sheltered waits for a repair to full
// (Effort 5 Decision 8).
func releasable(condition domain.Condition, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool {
	for _, holder := range factionState.Factions {
		if source, ok := holder.Assets[condition.SourceAssetID]; ok {
			return rulebook.Assets[source.DefinitionID].Shelter() != nil
		}
	}
	return true
}
```

##### Task 9 — `internal/faction/engine/orchestrator.go`

The releases are offered with the lifts, before the turn-start effects, so a released asset's own `turn_start` entries fire this turn.

**Find:**
```go
// runConditionLifts offers the faction's disabled assets their price and applies
// the lifts it pays for. A recoverable failure is reported and lifts nothing.
func (e *Engine) runConditionLifts(faction *domain.Faction, factionState *state.FactionState, paths *campaigns.Paths, collector PhaseCollector, observer TurnObserver, phaseLog *slog.Logger) ([]domain.Mutation, error) {
	mutations, err := turn.OfferConditionLifts(faction, e.Rulebook, collector)
	if err == nil && len(mutations) > 0 {
```
**Replace with:**
```go
// runConditionLifts offers the faction's disabled assets their price and its
// sheltered assets a free release, and applies what it accepts. A recoverable
// failure is reported and lifts nothing.
func (e *Engine) runConditionLifts(faction *domain.Faction, factionState *state.FactionState, paths *campaigns.Paths, collector PhaseCollector, observer TurnObserver, phaseLog *slog.Logger) ([]domain.Mutation, error) {
	mutations, err := turn.OfferConditionLifts(faction, e.Rulebook, collector)
	if err == nil {
		var releases []domain.Mutation
		releases, err = turn.OfferShelterReleases(faction, factionState, e.Rulebook, collector)
		mutations = append(mutations, releases...)
	}
	if err == nil && len(mutations) > 0 {
```

##### Task 10 — `internal/faction/tui/views/turn/overlay/ability.go`

**Find:**
```go
	case hooks.PaymentRestore:
		return NoCancelConfirm{NewConfirm(fmt.Sprintf("Pay %d Coin to restore %s at 1 HP?", offer.Amount, name), header, "Pay", "Let it be destroyed")}
	}
```
**Replace with:**
```go
	case hooks.PaymentRestore:
		return NoCancelConfirm{NewConfirm(fmt.Sprintf("Pay %d Coin to restore %s at 1 HP?", offer.Amount, name), header, "Pay", "Let it be destroyed")}
	case hooks.PaymentReleaseShelter:
		return NoCancelConfirm{NewConfirm(fmt.Sprintf("Bring %s out of hiding?", name), header, "Bring it out", "Stay hidden")}
	}
```

##### Task 11 — tests

`internal/faction/rulebook/rulebook_test.go` — the shelter step loads, and its shape is checked:

**Find:**
```go
func TestConvertAbilityStep_RaiseTechLevelBlockOnAnotherEffect(t *testing.T) {
```
**Replace with:**
```go
func TestConvertAbilityStep_Shelter(t *testing.T) {
	const header = "\n[assets.Machine.ability]\neffect = \"shelter\"\n"
	const target = "target = \"asset\"\n"
	const block = "[assets.Machine.ability.shelter]\nasset_types = [\"Operative\", \"Security Personnel\"]\ncapacity = 1\n"

	t.Run("loads", func(t *testing.T) {
		assets, err := loadAssetsFromTOML(t, effectTestAsset+header+target+block)
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		want := &domain.ShelterParams{AssetTypes: []domain.AssetType{"Operative", "Security Personnel"}, Capacity: 1}
		if got := assets["T-001"].Ability.Steps[0].Shelter; !reflect.DeepEqual(got, want) {
			t.Errorf("shelter = %+v, want %+v", got, want)
		}
	})

	cases := []struct {
		name string
		body string
		want string
	}{
		{"no asset target", header + block, "shelter requires"},
		{"no block", header + target, "shelter requires"},
		{"a contest", header + target + "roll = \"test\"\nattacker_stat = \"Cunning\"\ndefender_stat = \"Cunning\"\n" + block, "shelter requires"},
		{"no capacity", header + target + "[assets.Machine.ability.shelter]\nasset_types = [\"Operative\"]\n", "capacity > 0"},
		{"no types", header + target + "[assets.Machine.ability.shelter]\ncapacity = 1\n", "asset_types"},
		{"a block on another effect", "\n[assets.Machine.ability]\neffect = \"levy_coin\"\ntarget = \"faction\"\n" + block, "needs effect = \"shelter\""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := loadAssetsFromTOML(t, effectTestAsset+tc.body)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestConvertAbilityStep_RaiseTechLevelBlockOnAnotherEffect(t *testing.T) {
```

In `TestCheckAssetTypes`:

**Find:**
```go
		{name: "ability relocate", assets: starship(domain.AssetDefinition{Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{Relocate: &domain.RelocateParams{ExcludeTypes: []domain.AssetType{"Gunboat"}}}}}}), want: `asset "A-1" ability step 0`},
```
**Replace with:**
```go
		{name: "ability relocate", assets: starship(domain.AssetDefinition{Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{Relocate: &domain.RelocateParams{ExcludeTypes: []domain.AssetType{"Gunboat"}}}}}}), want: `asset "A-1" ability step 0`},
		{name: "ability shelter", assets: starship(domain.AssetDefinition{Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{Shelter: &domain.ShelterParams{AssetTypes: []domain.AssetType{"Gunboat"}, Capacity: 1}}}}}), want: `asset "A-1" ability step 0`},
```

`internal/faction/engine/effect/effects/shelter_asset_test.go` — the `asset_destroyed` route registers no rescue, and still takes the sheltered asset down:

**Find:**
```go
			got := runReactors(reactors, tc.mutations, factionState)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}
```
**Replace with:**
```go
			got := runReactors(reactors, tc.mutations, factionState)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestShelterAsset_AssetDestroyedRoute(t *testing.T) {
	factionState := boltholesState()
	alpha := factionState.Factions["alpha"]
	alpha.Assets["mu"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "bolt"}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialShelterAsset].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Asset: alpha.Assets["bolt"], Def: boltholesDef}, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerAssetDestroyed,
			Effect:  domain.SpecialShelterAsset,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
		},
	})

	if interceptors := registry.LethalInterceptorsForAsset("alpha", "bolt"); len(interceptors) != 0 {
		t.Errorf("interceptors = %d, want 0: the asset_destroyed route rescues no one", len(interceptors))
	}
	var reactors []hooks.MutationReactor
	for _, registered := range registry.MutationReactorsForAsset("alpha", "bolt") {
		reactors = append(reactors, registered.Hook)
	}
	killed := domain.AssetRemoved{FactionID: "alpha", AssetID: "bolt", Cause: "attack", CausedByFactionID: "beta"}
	got := runReactors(reactors, []domain.Mutation{killed}, factionState)
	want := []domain.Mutation{domain.AssetRemoved{FactionID: "alpha", AssetID: "mu", Cause: "attack", CausedByFactionID: "beta"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mutations = %+v, want %+v", got, want)
	}
}
```

`internal/faction/engine/action/actions/ability/shelter_test.go` (new) — full contents:

```go
package ability

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
)

var shelterStep = domain.AbilityStep{
	Effect:  domain.EffectShelter,
	Target:  domain.TargetAsset,
	Shelter: &domain.ShelterParams{AssetTypes: []domain.AssetType{"Operative"}, Capacity: 1},
}

// makeDemiplaneFixture builds f1's Leased Demiplane d1 on Anchorage, with f1's
// Operatives op1 and op3 and Facility fac1 there, Operative op2 on Elsewhere, and
// f2's Operative b2 on Anchorage.
func makeDemiplaneFixture() (*domain.Faction, *domain.Asset, *state.FactionState, *rulebook.Rulebook) {
	demiplane := &domain.AssetDefinition{
		ID:      "demiplane",
		Name:    "Leased Demiplane",
		Flags:   []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{shelterStep}},
	}
	rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"demiplane": demiplane,
		"operative": {ID: "operative", Type: "Operative"},
		"facility":  {ID: "facility", Type: "Facility"},
	}}
	at := func(id, definitionID, ownerID, worldID string) *domain.Asset {
		return &domain.Asset{ID: id, DefinitionID: definitionID, OwnerID: ownerID, Location: domain.Location{WorldID: worldID}}
	}
	acting := at("d1", "demiplane", "f1", "Anchorage")
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{
		"d1":   acting,
		"op1":  at("op1", "operative", "f1", "Anchorage"),
		"op3":  at("op3", "operative", "f1", "Anchorage"),
		"fac1": at("fac1", "facility", "f1", "Anchorage"),
		"op2":  at("op2", "operative", "f1", "Elsewhere"),
	}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"f1": faction,
		"f2": {ID: "f2", Assets: map[string]*domain.Asset{"b2": at("b2", "operative", "f2", "Anchorage")}},
	}}
	return faction, acting, factionState, rulebook
}

func TestShelterCandidates(t *testing.T) {
	cases := []struct {
		name  string
		setup func(faction *domain.Faction)
		want  []string
	}{
		{name: "own assets of a listed type on its world", want: []string{"op1", "op3"}},
		{name: "an asset sheltered elsewhere is skipped", setup: func(faction *domain.Faction) {
			faction.Assets["op1"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "other"}}
		}, want: []string{"op3"}},
		{name: "an asset with a move order is skipped", setup: func(faction *domain.Faction) {
			faction.Assets["op1"].CurrentOrder = &domain.MovementOrder{}
		}, want: []string{"op3"}},
		{name: "a full Demiplane takes no one", setup: func(faction *domain.Faction) {
			faction.Assets["op2"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "d1"}}
		}, want: []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faction, acting, factionState, rulebook := makeDemiplaneFixture()
			if tc.setup != nil {
				tc.setup(faction)
			}
			candidates, picks := stepCandidates(shelterStep, faction, acting, factionState, nil, nil, rulebook)
			if !picks {
				t.Fatal("picks = false, want a pick")
			}
			if got := assetIDs(candidates.Assets); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("candidates = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDispatch_Shelter(t *testing.T) {
	faction, acting, factionState, rulebook := makeDemiplaneFixture()
	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ domain.AbilityHolder, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
			return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: candidates.Assets[0]}, nil
		})

	view := NewView(factionState)
	mutations, err := Dispatch(faction, domain.AbilityHolder{Asset: acting}, collector, &dicetest.Cycle{Values: []int{1}}, nil, factionState, view, rulebook, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	want := []domain.Mutation{domain.AssetConditionApplied{
		FactionID:         "f1",
		AssetID:           "op1",
		Condition:         domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: "d1"},
		Cause:             "ability",
		CausedByFactionID: "f1",
	}}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want %+v", mutations, want)
	}
}

func TestHasTargets_FullShelter(t *testing.T) {
	faction, acting, factionState, rulebook := makeDemiplaneFixture()
	faction.Assets["op1"].Conditions = []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "d1"}}
	if HasTargets(faction, domain.AbilityHolder{Asset: acting}, &domain.AbilityDefinition{Steps: []domain.AbilityStep{shelterStep}}, factionState, nil, nil, rulebook) {
		t.Error("HasTargets = true, want false: the Demiplane is full")
	}
}
```

`internal/faction/engine/turn/conditions_test.go` — append:

**Find:**
```go
	t.Run("a collector error is returned", func(t *testing.T) {
		promptErr := errors.New("prompt failed")
		_, err := OfferConditionLifts(makeFaction(), rulebook, &paymentFake{err: promptErr})
		if !errors.Is(err, promptErr) {
			t.Errorf("err = %v, want it to wrap %v", err, promptErr)
		}
	})
}
```
**Replace with:**
```go
	t.Run("a collector error is returned", func(t *testing.T) {
		promptErr := errors.New("prompt failed")
		_, err := OfferConditionLifts(makeFaction(), rulebook, &paymentFake{err: promptErr})
		if !errors.Is(err, promptErr) {
			t.Errorf("err = %v, want it to wrap %v", err, promptErr)
		}
	})
}

func TestOfferShelterReleases(t *testing.T) {
	byDemiplane := domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: "d1"}
	byBoltholes := domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: "b1"}
	bySoldSource := domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: "sold"}
	rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"demiplane": {ID: "demiplane", Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
			Effect:  domain.EffectShelter,
			Target:  domain.TargetAsset,
			Shelter: &domain.ShelterParams{AssetTypes: []domain.AssetType{"Operative"}, Capacity: 1},
		}}}},
		"boltholes": {ID: "boltholes"},
		"operative": {ID: "operative"},
	}}
	makeState := func() (*domain.Faction, *state.FactionState) {
		faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{
			"d1": {ID: "d1", DefinitionID: "demiplane"},
			"b1": {ID: "b1", DefinitionID: "boltholes"},
			"x1": {ID: "x1", DefinitionID: "operative", Conditions: []domain.Condition{byDemiplane}},
			"x2": {ID: "x2", DefinitionID: "operative", Conditions: []domain.Condition{byBoltholes}},
			"x3": {ID: "x3", DefinitionID: "operative", Conditions: []domain.Condition{bySoldSource}},
		}}
		return faction, &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
	}

	t.Run("offers a Demiplane's asset and an orphan, free; a Boltholes asset waits", func(t *testing.T) {
		collector := &paymentFake{answer: true}
		faction, factionState := makeState()
		mutations, err := OfferShelterReleases(faction, factionState, rulebook, collector)
		if err != nil {
			t.Fatalf("OfferShelterReleases: %v", err)
		}
		want := []domain.Mutation{
			domain.AssetConditionCleared{FactionID: "f1", AssetID: "x1", Condition: byDemiplane, Cause: "shelter_released", CausedByFactionID: "f1"},
			domain.AssetConditionCleared{FactionID: "f1", AssetID: "x3", Condition: bySoldSource, Cause: "shelter_released", CausedByFactionID: "f1"},
		}
		if !reflect.DeepEqual(mutations, want) {
			t.Errorf("mutations = %+v, want %+v", mutations, want)
		}
		for _, offer := range collector.offers {
			if offer.Amount != 0 || offer.Reason != hooks.PaymentReleaseShelter {
				t.Errorf("offer = %+v, want a free release_shelter offer", offer)
			}
		}
	})

	t.Run("declined offers emit nothing", func(t *testing.T) {
		faction, factionState := makeState()
		mutations, err := OfferShelterReleases(faction, factionState, rulebook, &paymentFake{answer: false})
		if err != nil || mutations != nil {
			t.Errorf("mutations, err = %+v, %v; want nil, nil", mutations, err)
		}
	})

	t.Run("a collector error is returned", func(t *testing.T) {
		promptErr := errors.New("prompt failed")
		faction, factionState := makeState()
		_, err := OfferShelterReleases(faction, factionState, rulebook, &paymentFake{err: promptErr})
		if !errors.Is(err, promptErr) {
			t.Errorf("err = %v, want it to wrap %v", err, promptErr)
		}
	})
}
```

##### Task 12 — this plan

This plan's Work Breakdown Overview: `- [ ] **Phase 1 — The sacrifice filter and the shelter door**` becomes `- [x] **Phase 1 — The sacrifice filter and the shelter door**`.

##### Commit message

```
feat(E5): add a shelter ability that hides an own asset

- domain: the shelter ability effect, ShelterParams and
  AssetDefinition.Shelter
- rulebook: a [shelter] block (asset_types, capacity) on a shelter
  step; its shape and its types are checked
- ability: shelter takes in one of the faction's own assets on its
  place; the target stage reads the rulebook for the types
- shelter_asset: an asset_destroyed route takes what an asset
  shelters with it, and rescues no one
- turn start: a sheltered asset whose source has a shelter ability,
  or no longer exists, is offered a free release
- TUI: "Bring X out of hiding?"
- tests: loading, candidates, dispatch, the destroyed route and the
  release
- E5 plan: Phase 1 ticked
```

### Phase 2 — The Cunning pass

**In plain words:** H.A.W.K's 28 Cunning assets replace the SWN copy, the new Front type is declared, the seed's Cunning units become H.A.W.K's (plus one Glamoured Doubles), the canon gets their fiction, and Robert plays a turn with them.

#### Commit 3 — `feat(E5): replace the Cunning stand-in with H.A.W.K's own assets`

##### Task 1 — `rulebooks/hawk/assets/cunning_assets.toml`

Rewritten — full contents:

```toml
# Cunning Assets — H.A.W.K (A.001.5 Effort 5)

[assets.ImpFamiliars]
id = "HAWK-C1-001"
name = "Imp Familiars"
category = "Cunning"
min_rating = 1
hp = 3
cost = 2
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["A", "S"]
counter = ""
description = "Contract imps issued as assistants. As an action, make a Cunning vs. Cunning attack on a faction; on a success, all its Stealthed assets on this place are revealed."
ability = {
  roll = "attack",
  attacker_stat = "Cunning",
  defender_stat = "Cunning",
  effect = "reveal_stealth",
  target = "faction",
}

[assets.GlamouredDoubles]
id = "HAWK-C1-002"
name = "Glamoured Doubles"
category = "Cunning"
min_rating = 1
hp = 2
cost = 1
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["S"]
counter = ""
description = "Stand-ins glamoured to wear an agent's face. If another of the faction's Operatives on this place takes enough damage to destroy it, the faction may sacrifice the Doubles instead to nullify the killing blow."
effects = [
  {
    trigger = "lethal_damage",
    effect = "sacrifice_self",
    filters = { owner = "own", scope = "world", asset_type = ["Operative"] },
  },
]

[assets.BlackHullSmugglers]
id = "HAWK-C1-003"
name = "Black-hull Smugglers"
category = "Cunning"
min_rating = 1
hp = 4
cost = 2
maintenance = 0
tech_level = 4
speed = 14
drift_rating = 2
type = "Aethership"
flags = ["S"]
counter = ""
description = "Unlit smuggling skiffs with no registry. Moves at speed 14 (mid-range). For 1 Coin, carries itself and one Operative, landing even past a lockdown."
attack = { attacker_stat = "Cunning", defender_stat = "Wealth", damage = "1d4" }
transport = { coin_cost = 1, cargo_types = ["Operative"], max_cargo = 1, ignores_lockdown = true }

[assets.ProsperityGospel]
id = "HAWK-C1-004"
name = "Prosperity Gospel"
category = "Cunning"
min_rating = 1
hp = 2
cost = 1
maintenance = 0
tech_level = 2
type = "Stratagem"
flags = ["S"]
counter = ""
description = "A doctrine that wealth is grace. Any other faction buying an asset on this place pays 1 extra Coin, which is lost. This applies once, however many are present."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "1d4-1" }
effects = [
  {
    trigger = "asset_cost_query",
    effect = "cost_surcharge",
    filters = { owner = "rival", scope = "world" },
    gates = { no_stack = true },
    cost_surcharge = { amount = 1 },
  },
]

[assets.CharmPeddlers]
id = "HAWK-C1-005"
name = "Charm Peddlers"
category = "Cunning"
min_rating = 1
hp = 4
cost = 2
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["A"]
counter = "1d4"
description = "Luck sold door to door. As an action, roll 1d6; on a 3 or higher, gain 1 Coin."
ability = { effect = "coin_delta", target_roll = { die = "1d6", target = 3, amount = 1 } }

[assets.OldBargains]
id = "HAWK-C2-001"
name = "Old Bargains"
category = "Cunning"
min_rating = 2
hp = 4
cost = 4
maintenance = 1
tech_level = 0
type = "Stratagem"
flags = ["S"]
counter = ""
description = "Favours the Fey are owed. When a rival gains permission to buy or bring in an asset on this place, make an immediate Cunning vs. Cunning test; on a success, the permission is withdrawn and cannot be asked again until next turn."
effects = [
  {
    trigger = "permission_granted",
    effect = "withdraw_permission",
    filters = { owner = "rival", scope = "world" },
    gates = { contest = { roller = "self", roll = "test", attacker_stat = "Cunning", defender_stat = "Cunning" } },
  },
]

[assets.SkeletonKeys]
id = "HAWK-C2-002"
name = "Skeleton Keys"
category = "Cunning"
min_rating = 2
hp = 6
cost = 5
maintenance = 0
tech_level = 0
type = "Stratagem"
flags = ["S"]
counter = ""
description = "Master keys to an old security system. An asset attacked with the Keys cannot use its ability until the start of the attacking faction's next turn, whether or not the attack succeeds."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "2d4" }
effects = [
  { trigger = "own_attack", effect = "block_ability", filters = { owner = "own", scope = "self" } },
]

[assets.CorporateAuditors]
id = "HAWK-C2-003"
name = "Corporate Auditors"
category = "Cunning"
min_rating = 2
hp = 4
cost = 4
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["S"]
counter = ""
description = "Auditors sworn to disallow. Any attack on or defence against the Auditors loses the bonus dice the rival's tags would give."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "1d4+1" }
effects = [
  { trigger = "enemy_roll_at_me", effect = "remove_tag_dice", filters = { owner = "rival", scope = "self" } },
]

[assets.FeyCourtiers]
id = "HAWK-C2-004"
name = "Fey Courtiers"
category = "Cunning"
min_rating = 2
hp = 4
cost = 4
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["A", "S"]
counter = ""
description = "Courtiers who charm a rival's staff. As an action, move to a place within 4 hexes, then make a Cunning vs. Cunning attack on a faction there; on a success, it takes no damage but its Stealthed assets on that place are revealed. Only Operatives can attack Fey Courtiers."
ability = {
  steps = [
    { effect = "relocate", relocate = { scope = "self", max_hex = 4 } },
    {
      effect = "reveal_stealth",
      target = "faction",
      roll = "attack",
      attacker_stat = "Cunning",
      defender_stat = "Cunning",
    },
  ],
}
effects = [
  {
    trigger = "attack_eligibility_query",
    effect = "restrict_attack",
    filters = { owner = "rival", scope = "self", asset_type = ["Operative"] },
    restrict_attack = { side = "defense", allow = "only_listed" },
  },
]

[assets.OccultLodge]
id = "HAWK-C2-005"
name = "Occult Lodge"
category = "Cunning"
min_rating = 2
hp = 4
cost = 4
maintenance = 0
tech_level = 0
type = "Front"
flags = ["S"]
counter = ""
description = "A secret society, run as a franchise. Roll an extra d10 on the test to Expand Influence on this place."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "world", phase = ["faction_test"], action = "Expand Influence" },
    bonus_die = { sides = 10, count = 1 },
  },
]

[assets.Wardbreakers]
id = "HAWK-C3-001"
name = "Wardbreakers"
category = "Cunning"
min_rating = 3
hp = 4
cost = 6
maintenance = 0
tech_level = 4
type = "Operative"
flags = ["S"]
counter = ""
description = "Engineers who walk through wards. Wardbreakers start Stealthed when bought."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "2d6" }
effects = [
  { trigger = "on_purchase", effect = "apply_stealth", filters = { owner = "own", scope = "self" } },
]

[assets.Stealth]
id = "HAWK-C3-002"
name = "Stealth"
category = "Cunning"
min_rating = 3
hp = 0
cost = 2
maintenance = 0
tech_level = 0
type = "Stratagem"
flags = ["S"]
quality = true
counter = ""
description = "Not a standard asset but a special quality, bought for one of the faction's Operatives on this place. A Stealthed asset cannot be detected or attacked. If it would need government permission to land, that permission may be foregone. Stealth is lost if the asset attacks or defends."
effects = [
  {
    trigger = "on_purchase",
    effect = "grant_stealth",
    filters = { owner = "own", scope = "world", asset_type = ["Operative"], stealthy = false },
  },
]

[assets.Mirrorways]
id = "HAWK-C3-003"
name = "Mirrorways"
category = "Cunning"
min_rating = 3
hp = 4
cost = 8
maintenance = 0
tech_level = 4
type = "Front"
flags = ["A", "S"]
counter = ""
description = "Paired mirrors to step through. As an action, for 1 Coin, move one Operative between any places within 12 hexes of the Mirrorways."
ability = {
  effect = "relocate",
  relocate = { scope = "other", max_hex = 12, coin_cost = 1, asset_type = "Operative", max_assets = 1 },
}

[assets.AscensionProgram]
id = "HAWK-C4-001"
name = "Ascension Program"
category = "Cunning"
min_rating = 4
hp = 10
cost = 8
maintenance = 0
tech_level = 0
type = "Front"
flags = ["S"]
counter = "1d6"
description = "A path to godhood in twelve paid levels. Each turn, gain 1 Coin."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "2d6" }
effects = [
  {
    trigger = "turn_start",
    effect = "coin_gain",
    filters = { owner = "own", scope = "self" },
    coin_gain = { amount = "1" },
  },
]

[assets.ChosenInitiates]
id = "HAWK-C4-002"
name = "Chosen Initiates"
category = "Cunning"
min_rating = 4
hp = 12
cost = 8
maintenance = 0
tech_level = 3
type = "Security Personnel"
flags = []
counter = "1d6"
description = "The inner circle's sworn guard, who keep the doors, collect the dues and remove disruptive guests."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "1d6" }

[assets.TrainedHellhounds]
id = "HAWK-C4-003"
name = "Trained Hellhounds"
category = "Cunning"
min_rating = 4
hp = 8
cost = 12
maintenance = 0
tech_level = 4
type = "Beast"
flags = ["S"]
counter = "1d4"
description = "Hellhounds trained to smell glamour. Whenever a rival's Stealthed asset arrives or is bought on this place, make an immediate Cunning vs. Cunning attack on its faction; on a success, the asset loses its Stealth."
effects = [
  {
    trigger = "stealth_arrival",
    effect = "clear_stealth",
    filters = { owner = "rival", scope = "world", stealthy = true },
    gates = { contest = { roller = "self", roll = "attack", attacker_stat = "Cunning", defender_stat = "Cunning" } },
  },
]

[assets.DiabolicInterlopers]
id = "HAWK-C4-004"
name = "Diabolic Interlopers"
category = "Cunning"
min_rating = 4
hp = 8
cost = 12
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["A"]
counter = ""
description = "Devils that slip into a host. As an action, for 1d4 Coin, attach to a rival asset on this place. Until they attach to another asset or no longer share its place, that asset cannot attack. If it is destroyed, the Interlopers survive."
ability = { effect = "block_attack", target = "asset", die = "1d4" }

[assets.MembershipDues]
id = "HAWK-C5-001"
name = "Membership Dues"
category = "Cunning"
min_rating = 5
hp = 8
cost = 10
maintenance = 0
tech_level = 0
type = "Stratagem"
flags = ["S"]
counter = ""
description = "A rival's staff, secretly members. When the Dues successfully attack an enemy asset, they steal 1d4 Coin from its faction, once per turn per faction, however many attack."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "2d6" }
effects = [
  {
    trigger = "own_attack_hit",
    effect = "coin_steal",
    filters = { owner = "rival", scope = "world" },
    gates = { once_per = "target_faction" },
    coin_steal = { amount = "1d4" },
  },
]

[assets.Dreamwalkers]
id = "HAWK-C5-002"
name = "Dreamwalkers"
category = "Cunning"
min_rating = 5
hp = 6
cost = 14
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["S"]
counter = ""
description = "They rewrite a commander's plan in their sleep. If the Dreamwalkers defend successfully against an attack, the attacking asset immediately attacks itself, for normal damage or counterattack results."
effects = [
  { trigger = "own_defense_win", effect = "counterstrike", filters = { owner = "own", scope = "self" } },
]

[assets.LeasedDemiplane]
id = "HAWK-C5-003"
name = "Leased Demiplane"
category = "Cunning"
min_rating = 5
hp = 6
cost = 12
maintenance = 0
tech_level = 4
type = "Front"
flags = ["A", "S"]
counter = "2d6"
description = "A room outside the world, rented monthly. As an action, move one of the faction's Operatives or Security Personnel on this place inside; the Demiplane holds one asset at a time. While inside, the asset cannot be detected, attacked or moved, and cannot act. At the start of the faction's turn, it may bring the asset out at no cost. If the Demiplane is destroyed, the asset inside is destroyed with it."
ability = {
  effect = "shelter",
  target = "asset",
  shelter = { asset_types = ["Operative", "Security Personnel"], capacity = 1 },
}
effects = [
  { trigger = "asset_destroyed", effect = "shelter_asset", filters = { owner = "own", scope = "self" } },
]

[assets.GeasBackedNDA]
id = "HAWK-C5-004"
name = "Geas-Backed NDA"
category = "Cunning"
min_rating = 5
hp = 8
cost = 10
maintenance = 0
tech_level = 0
type = "Stratagem"
flags = ["A"]
counter = ""
description = "A geas that gags a rival asset's staff. As an action, test Cunning vs. Cunning against a rival faction's asset; on a success, its faction must at once pay half that asset's purchase cost (rounded down), or the asset is disabled until the price is paid."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "1d6" }
ability = {
  roll = "test",
  attacker_stat = "Cunning",
  defender_stat = "Cunning",
  effect = "disable_asset",
  target = "asset",
}

[assets.RulesOfHospitality]
id = "HAWK-C6-001"
name = "Rules of Hospitality"
category = "Cunning"
min_rating = 6
hp = 10
cost = 20
maintenance = 0
tech_level = 4
type = "Stratagem"
flags = ["S"]
counter = ""
description = "No one enters uninvited. On a successful Cunning vs. Cunning attack on a rival faction, it cannot bring assets onto this place without paying 1d4 Coin and waiting one turn."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning" }
effects = [
  {
    trigger = "own_attack_hit",
    effect = "lock_out",
    filters = { owner = "own", scope = "self" },
    lock_out = { amount = "1d4" },
  },
]

[assets.WhisperGateNetwork]
id = "HAWK-C6-002"
name = "WhisperGate™ Network"
category = "Cunning"
min_rating = 6
hp = 15
cost = 18
maintenance = 0
tech_level = 4
type = "Front"
flags = ["A"]
counter = ""
description = "Gates across a region that open to a password. As an action, move any number of the faction's Operatives between places in this region."
ability = {
  effect = "relocate",
  relocate = { scope = "other", range = "region", asset_type = "Operative", max_assets = 0 },
}

[assets.ProphetExecutive]
id = "HAWK-C6-003"
name = "Prophet-Executive"
category = "Cunning"
min_rating = 6
hp = 10
cost = 20
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["S"]
counter = "1d8"
description = "Prophet and chief executive in one. Once per turn, roll an extra d10 on one of the faction's Cunning attacks on this place."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "2d8" }
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "world", phase = ["attack"], attribute = "Cunning" },
    gates = { once_per = "asset" },
    bonus_die = { sides = 10, count = 1 },
  },
]

[assets.StateFaith]
id = "HAWK-C7-001"
name = "State Faith"
category = "Cunning"
min_rating = 7
hp = 16
cost = 25
maintenance = 0
tech_level = 4
type = "Stratagem"
flags = ["S"]
counter = "1d6"
description = "The faith is the place's religion. The government always grants the faction's requests to buy or bring in assets here."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "2d6" }
effects = [
  { trigger = "permission_query", effect = "grant_permission", filters = { owner = "own", scope = "world" } },
]

[assets.AuguryDesk]
id = "HAWK-C7-002"
name = "Augury Desk"
category = "Cunning"
min_rating = 7
hp = 10
cost = 20
maintenance = 0
tech_level = 4
type = "Front"
flags = ["S"]
counter = "2d8"
description = "A trading desk of augurs. Once per turn, the faction may reroll one die for an action on this place, or force a rival to reroll one die."
effects = [
  {
    trigger = "own_roll",
    effect = "reroll",
    filters = { owner = "own", scope = "world" },
    gates = { once_per = "owner_faction" },
    reroll = { dice = 1, elective = true },
  },
  {
    trigger = "enemy_roll_at_me",
    effect = "reroll",
    filters = { owner = "rival", scope = "world" },
    gates = { once_per = "owner_faction" },
    reroll = { dice = 1, elective = true },
  },
]

[assets.ChangelingAgents]
id = "HAWK-C7-003"
name = "Changeling Agents"
category = "Cunning"
min_rating = 7
hp = 5
cost = 10
maintenance = 0
tech_level = 0
type = "Stratagem"
flags = ["S"]
counter = ""
description = "A rival officer swapped for a changeling. On a successful Cunning vs. Cunning attack, the Agents are lost, the faction gains 5 Coin, and the target asset joins the faction, even if the faction could not otherwise field it."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning" }
effects = [
  {
    trigger = "own_attack_hit",
    effect = "capture_target",
    filters = { owner = "own", scope = "self" },
    capture_target = { amount = "5" },
  },
]

[assets.OracleEngine]
id = "HAWK-C8-001"
name = "Oracle Engine"
category = "Cunning"
min_rating = 8
hp = 20
cost = 30
maintenance = 2
tech_level = 5
type = "Front"
flags = ["S"]
counter = "1d6"
description = "An Arcane Intelligence that sees one place whole. Every rival Stealthed asset on this place must pass a Cunning vs. Cunning test at the start of the faction's turn or lose its Stealth. The faction rolls an extra die on all Cunning attacks and defences on this place. Costs 2 Coin of maintenance each turn."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "world", phase = ["attack", "defense"], attribute = "Cunning" },
    bonus_die = { sides = 10, count = 1 },
  },
  {
    trigger = "turn_start",
    effect = "clear_stealth",
    filters = { owner = "rival", scope = "world", stealthy = true },
    gates = { contest = { roller = "target", roll = "test", attacker_stat = "Cunning", defender_stat = "Cunning" } },
  },
]
```

##### Task 2 — `rulebooks/hawk/settings.toml`

Front joins H.A.W.K's types (Decision 4); SWN's stay for the Wealth stand-in.

**Find:**
```toml
# cargo, a relocate's filter, an effects filter and the floors below. H.A.W.K's
# types come first; SWN's stay for the Cunning and Wealth stand-ins until
# Efforts 5 and 6 replace them.
asset_types = [
  "Security Personnel", "Operative", "Facility", "Aethership", "Stratagem", "Beast",
```
**Replace with:**
```toml
# cargo, a relocate's filter, a shelter, an effects filter and the floors below.
# H.A.W.K's types come first; SWN's stay for the Wealth stand-in until Effort 6
# replaces it.
asset_types = [
  "Security Personnel", "Operative", "Facility", "Aethership", "Stratagem", "Beast", "Front",
```

##### Task 3 — `campaigns/hawk/state/state.seed.toml`

The seed's 17 Cunning units become H.A.W.K's, and the Radiant Throne gains a Glamoured Doubles (Decision 17).

| Faction | Stand-in | Heir |
|---------|----------|------|
| O.W.L. | `SWN-C2-001` Lobbyists | `HAWK-C2-005` Occult Lodge |
| O.W.L. | `SWN-C3-003` Covert Shipping | `HAWK-C3-003` Mirrorways |
| Court of Lies | `SWN-C4-004` Seditionists | `HAWK-C4-004` Diabolic Interlopers |
| Court of Lies | `SWN-C5-001` Organization Moles | `HAWK-C5-001` Membership Dues |
| Court of Lies | `SWN-C6-003` Demagogue | `HAWK-C6-003` Prophet-Executive |
| Court of Lies | `SWN-C7-003` Treachery | `HAWK-C7-003` Changeling Agents |
| Radiant Throne | `SWN-C1-002` Informers | `HAWK-C1-001` Imp Familiars |
| Radiant Throne | `SWN-C4-001` Party Machine | `HAWK-C4-001` Ascension Program |
| Radiant Throne | — | `HAWK-C1-002` Glamoured Doubles (new) |
| Fey Royals | `SWN-C2-004` Seductress | `HAWK-C2-004` Fey Courtiers |
| Fey Royals | `SWN-C5-001` Organization Moles | `HAWK-C5-001` Membership Dues |
| Fey Royals | `SWN-C7-001` Popular Movement | `HAWK-C7-001` State Faith |
| Fey Royals | `SWN-C7-002` Book of Secrets | `HAWK-C7-002` Augury Desk |
| Blackwater | `SWN-C5-003` Boltholes | `HAWK-C5-003` Leased Demiplane |
| Blackwater | `SWN-C6-002` Covert Transit Net | `HAWK-C6-002` WhisperGate™ Network |
| Witherblight | `SWN-C2-002` Saboteurs | `HAWK-C2-002` Skeleton Keys |
| Witherblight | `SWN-C3-001` Cyberninjas | `HAWK-C3-001` Wardbreakers |
| Seraph | `SWN-C2-001` Lobbyists | `HAWK-C2-001` Old Bargains |

The two units whose heir changes number go first. O.W.L.'s Lobbyists:

**Find:**
```toml
  [factions.occult-workers-league.assets.occult-workers-league-SWN-C2-001-1]
    id = "occult-workers-league-SWN-C2-001-1"
    definition_id = "SWN-C2-001"
```
**Replace with:**
```toml
  [factions.occult-workers-league.assets.occult-workers-league-HAWK-C2-005-1]
    id = "occult-workers-league-HAWK-C2-005-1"
    definition_id = "HAWK-C2-005"
```

The Radiant Throne's Informers, with the new Glamoured Doubles after them:

**Find:**
```toml
  [factions.radiant-throne.assets.radiant-throne-SWN-C1-002-1]
    id = "radiant-throne-SWN-C1-002-1"
    definition_id = "SWN-C1-002"
    owner_id = "radiant-throne"
    current_hp = 3
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }
```
**Replace with:**
```toml
  [factions.radiant-throne.assets.radiant-throne-HAWK-C1-001-1]
    id = "radiant-throne-HAWK-C1-001-1"
    definition_id = "HAWK-C1-001"
    owner_id = "radiant-throne"
    current_hp = 3
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }
  [factions.radiant-throne.assets.radiant-throne-HAWK-C1-002-1]
    id = "radiant-throne-HAWK-C1-002-1"
    definition_id = "HAWK-C1-002"
    owner_id = "radiant-throne"
    current_hp = 2
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "solar-gate", region_hex = { RegionID = "A", Coord = { Q = 15, R = 1 } } }
```

Every other heir keeps its stand-in's suffix. With the Edit tool and `replace_all`, replace every `SWN-C` with `HAWK-C`: the remaining 15 units' table header, `id` and `definition_id` lines (45 lines).

Then the header:

**Find:**
```toml
# Force starting assets are H.A.W.K's own (Effort 4); Cunning and Wealth assets
# are SWN stand-ins until Efforts 5 and 6 swap in HAWK- assets.
# Each faction holds its scale's count: hegemon 4 primary + 4 other, major 2 + 2.
```
**Replace with:**
```toml
# Force and Cunning starting assets are H.A.W.K's own (Efforts 4 and 5); Wealth
# assets are SWN stand-ins until Effort 6 swaps in HAWK- assets.
# Each faction holds its scale's count: hegemon 4 primary + 4 other, major 2 + 2.
# The Radiant Throne also holds one Glamoured Doubles, to show it in play (Effort 5).
```

Check: `grep -c "SWN-C" campaigns/hawk/state/state.seed.toml` prints `0`, and `grep -c "definition_id = \"HAWK-C" campaigns/hawk/state/state.seed.toml` prints `18`.

##### Commit message

```
feat(E5): replace the Cunning stand-in with H.A.W.K's own assets

- cunning_assets.toml: 28 H.A.W.K Cunning assets with HAWK-C IDs,
  H.A.W.K types, the new Front type, Effort 4's speed classes and
  teleport ranges, Stealth for Operatives and the Leased Demiplane's
  shelter door
- settings.toml: Front joins asset_types
- state.seed.toml: the 17 seeded Cunning units swap to their heirs
  by role; the Radiant Throne gains a Glamoured Doubles
```

#### Commit 4 — `docs(E5): write the Cunning canon and play the live turn`

##### Task 1 — `campaigns/hawk/canon/assets.md`

The fiction of the 28 assets, from the signed-off pass 3 drafts. Stealth's entry is one plain line (Decision 9).

**Find:**
```markdown
It carries a crew of thousands, a board of directors and a full cohort of unpaid interns, who handle the Core. Its upkeep is a standing line in the annual report.
```
**Replace with:**
```markdown
It carries a crew of thousands, a board of directors and a full cohort of unpaid interns, who handle the Core. Its upkeep is a standing line in the annual report.

## Imp Familiars — `HAWK-C1-001`

Imps bound by contract and issued to middle managers as personal assistants, small enough to sit on a shoulder and take the minutes. They attend every meeting and file a report on each one. The contract says who reads the reports, and it is never the manager.

## Glamoured Doubles — `HAWK-C1-002`

Temporary staff glamoured to look, sound and walk like one of the faction's agents, and sent out on the agent's usual routes. When a rival finally strikes, the blow lands on the double. The glamour fades from the body within the hour, and the agent goes back to work.

## Black-hull Smugglers — `HAWK-C1-003`

Small Aetherships with black hulls, no lights and no registry runes, crewed by smugglers who fly the Drift with one passenger at a time. They carry no manifest and take payment in Aetherstones. Every company denies owning them.

## Prosperity Gospel — `HAWK-C1-004`

A doctrine preached on a place: wealth is a sign of grace, and grace is earned by working for the faithful. Workers who have heard it ask an outsider for a blessing bonus before they sign. Most outsiders pay it.

## Charm Peddlers — `HAWK-C1-005`

Licensed peddlers of luck charms, curse wards and lucky Aetherstones, who work the docks and the staff canteens with a case of samples. Each charm is certified for one year. The curses are oddly common wherever the peddlers work.

## Old Bargains — `HAWK-C2-001`

Favours owed to the Fey by the officials of a place, or by their grandparents. When a permit is signed that the Fey dislike, the official remembers a promise, and the signature fades from the page. No one recalls making the bargain, and everyone honours it.

## Skeleton Keys — `HAWK-C2-002`

Master keys to wards, runes and Aether-Core housings, cut for a security system that half the Core still runs on. One key opens a machine room; one turn of it stops the machine. The locks are changed after every visit, and the keys still work.

## Corporate Auditors — `HAWK-C2-003`

Chartered auditors bound by an oath of office that is also a spell, sent to examine a rival's books under a writ that no one remembers signing. In their presence, a faction's usual advantages are disallowed as irregular. Every finding is minuted, and every penalty is invoiced.

## Fey Courtiers — `HAWK-C2-004`

Courtiers from the Feywild, sent to a rival's offices with gifts, compliments and an invitation to a party. Everyone they speak to tells them something true. Soldiers find they cannot raise a hand against them; it takes someone trained to resist a glamour.

## Occult Lodge — `HAWK-C2-005`

A chapter of a secret society, opened under franchise with a standard lodge hall, a standard handshake and a fee to the grand lodge. Its members are the place's clerks, magistrates and harbour masters. When the faction's new offices open, the paperwork is already done.

## Wardbreakers — `HAWK-C3-001`

Engineers who learned wards from the inside and now pass through them for hire. They cross a rival's abjurations the way a clerk crosses an unlocked office, and leave nothing in the log. A company hires them once to break in, and again to explain how.

## Stealth — `HAWK-C3-002`

The ordinary means of keeping one agent out of sight: a cover name, a quiet route and records that go missing. It lasts until the agent acts.

## Mirrorways — `HAWK-C3-003`

Mirrors in the washrooms and lobbies of friendly businesses, paired so that a person can step into one and out of the other. Visitors are asked to knock first. The cleaning staff have strict instructions about the glass.

## Ascension Program — `HAWK-C4-001`

A path to godhood in twelve levels, each with its own course, robe and fee. Members recruit their own students, and part of every fee flows up the levels to the top. No member has reached level twelve yet, but the founder says several are close.

## Chosen Initiates — `HAWK-C4-002`

Members picked from the lower circles to guard the inner one: they keep the doors, collect the dues and remove disruptive guests. Each carries a wand, a sealed oath and a short club, and uses all three. An initiate who leaves the circle is not seen again.

## Trained Hellhounds — `HAWK-C4-003`

Hellhounds leased from an infernal kennel, trained to smell glamour, illusion and forged papers, and posted at the docks with a handler at the other end of the leash. Every arrival is sniffed. The handler stamps whatever the hound allows.

## Diabolic Interlopers — `HAWK-C4-004`

Small devils that slip inside a host and argue with it. A soldier with a devil inside spends the battle in a debate, and loses. When the host dies, the devil steps out and looks for the next one.

## Membership Dues — `HAWK-C5-001`

A rival's workforce, enrolled one by one in the faction's society and still on the rival's payroll. Each payday, the membership dues go out by standing order before the wages arrive. The rival's books record the loss as staff welfare.

## Dreamwalkers — `HAWK-C5-002`

Fey-trained sleepers who walk into a rival commander's dreams the night before an attack and walk out with the plan. By morning the plan has changed in small ways the commander does not notice. The attack lands on its own lines.

## Leased Demiplane — `HAWK-C5-003`

A pocket demiplane rented by the month, reached through a cupboard door, with a bed, a lamp and no windows. An agent in trouble steps inside and closes the door. If the lease lapses, the room closes with the agent in it.

## Geas-Backed NDA — `HAWK-C5-004`

A non-disclosure agreement backed by a geas, served on everyone who works a rival's asset at once. Nobody bound by it can speak about their own work, write it down or discuss it with each other, so the work stops. The faction's lawyers will lift the geas for half of what the asset cost.

## Rules of Hospitality — `HAWK-C6-001`

An old Fey law: no one enters a household uninvited, and a whole place can be a household if the right words are said over it. Once they are said, a rival's ships and staff wait at the threshold until they are asked in. Invitations are available, for a fee.

## WhisperGate™ Network — `HAWK-C6-002`

Service doors in warehouses, offices and back alleys across a region, each fitted with a gate-rune that opens onto any other when the right word is whispered to it. The network is sold as a premium courier service. The passwords change daily and are never written down.

## Prophet-Executive — `HAWK-C6-003`

A prophet who is also a chief executive, or the other way round; the board has never settled which. They speak once a year in a plain robe, and the quarterly results improve. Every follower is a stakeholder, and every stakeholder is family.

## State Faith — `HAWK-C7-001`

The faith has become the place's official religion: the governor is a member, the council opens with a prayer, and the tax office collects the tithe. Requests from the faithful are approved as a matter of course. Requests from anyone else are prayed over.

## Augury Desk — `HAWK-C7-002`

A trading desk of augurs who read omens in Aether-flow charts and the entrails of birds, and forecast a place's events a day ahead. When something goes wrong, they have already sold it. Rivals who act on a forecast find that it was the desk's.

## Changeling Agents — `HAWK-C7-003`

Changelings trained to replace a rival officer overnight, with the same face, the same habits and different loyalties. The unit under them notices nothing until it reports for duty somewhere new. The real officer is returned some years later, well rested.

## Oracle Engine — `HAWK-C8-001`

An Arcane Intelligence built around a ring of scrying-glass and fed everything that happens on one place: every room, every ledger, every conversation. It answers any question about the place within the hour. It runs hot, and its upkeep is paid in Coin and in Aetherstones.
```

##### Task 2 — the live turn

The overview's Live turn, steps 1, 2 and 4. The map is unchanged, so step 3 is skipped. Claude runs:

```sh
go build -o gm-toolkit ./cmd/gm-toolkit
GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit campaign add-rules --rules rulebooks/hawk --campaign hawk --replace
./cmd/scripts/reset-campaign.sh hawk
```

Then Robert runs `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit faction --campaign hawk`, plays a turn or two, and reports what he saw. Amended in execution: Claude played it on `hawk-claude`, with a staged scenario (Decision Record 1). Worth a look:

- the 18 seeded Cunning units under their H.A.W.K names
- Blackwater's Leased Demiplane at Ouroboros offered as an ability, hiding its Shield Wardens; at Blackwater's next turn start, "Bring Shield Wardens out of hiding?"
- O.W.L.'s Expand Influence near its Occult Lodge at Eventide rolling the extra d10 (Decision Record 2)
- the Fey Royals' Fey Courtiers jumping up to 4 hexes before they reveal
- the Court of Lies' Prophet-Executive adding a d10 to one Cunning attack, and its Diabolic Interlopers attaching to a rival asset
- the Radiant Throne's Glamoured Doubles offered as a sacrifice if its Imp Familiars take a killing blow

A fault that needs a code or data change is logged in the Decision Record and fixed in this commit, or raised with Robert if it is bigger than the effort.

##### Task 3 — clear the drafts

Delete `scratch/A.001.5.E5-cunning-drafts.md`, `scratch/A.001.5.E5-cunning-drafts-pass2.md` and `scratch/A.001.5.E5-cunning-drafts-pass3.md`; the plan and the canon hold what they settled. They are gitignored, so nothing here touches the commit.

##### Task 4 — this plan and the overview

- This plan's Work Breakdown Overview: `- [ ] **Phase 2 — The Cunning pass**` becomes `- [x] **Phase 2 — The Cunning pass**`.
- `A.001.5-hawk-asset-authoring-overview-plan.md`, Effort Summary:

**Find:**
```markdown
| [ ] | 5 | Cunning pass |
```
**Replace with:**
```markdown
| [x] | 5 | Cunning pass |
```

##### Commit message

```
docs(E5): write the Cunning canon and play the live turn

- canon/assets.md: the fiction of the 28 H.A.W.K Cunning assets
- campaigns/hawk/rulebook/: rulebooks/hawk composed in again
- the E5 live H.A.W.K turn, played by Robert
- E5 plan: Phase 2 ticked; overview plan: Effort 5 ticked
```
