# H.A.W.K Asset Authoring — Effort 6 Plan: Wealth Pass

## In Plain Words

- **What** — H.A.W.K's own Wealth assets: 27 businesses, crews, ships and schemes that replace the last SWN copy, in `rulebooks/hawk/assets/wealth_assets.toml`. About half lean into the setting's darker side (debt, labor, company towns), and half are trade, finance and long-haul shipping. A new asset type, Workforce, marks the labor. Force gains one long-range troopship. Three small engine changes come first: a public teleport network that rival factions can ride for a fare paid to its owner; income counted per worker on a place; and Coin earned each time a fleet lands a blow.
- **Why** — Wealth is the last asset category H.A.W.K borrows from SWN. When it lands, every rulebook file is H.A.W.K's own, and "H.A.W.K playable" is reached.
- **How it fits** — Effort 6 of six in `A.001.5`, the last initiative of arc `A.001`. There are two phases: the engine changes, then the Wealth data, the seed swap, the canon and a live turn on the Effort 3 board. The initiative close comes next.

## Context / Goal

Design the Wealth pass (lore-gate, reflavor, sort, review; discovery D-9) and specify the engine work and the data to the task.

- Discovery doc: [`A.001.5-hawk-asset-authoring-discovery.md`](../discovery/A.001.5-hawk-asset-authoring-discovery.md), D-7 to D-11
- Overview plan: [`A.001.5-hawk-asset-authoring-overview-plan.md`](A.001.5-hawk-asset-authoring-overview-plan.md), Effort 6, Pass shape, Live turn
- Effort 4 plan: [`A.001.5.E4-hawk-asset-authoring-plan.md`](A.001.5.E4-hawk-asset-authoring-plan.md), Decisions 2, 11, 13, 14, 17 and 20 (voice, the stand-in type gap, descriptions, the seed swap, long-range cargo, speed classes)
- Effort 5 plan: [`A.001.5.E5-hawk-asset-authoring-plan.md`](A.001.5.E5-hawk-asset-authoring-plan.md), Decisions 1, 5 and 17 (concept first, shared mechanics, extras in the seed)
- Canon: `campaigns/hawk/canon/`; the drafts in gitignored `scratch/A.001.5.E6-wealth-drafts-pass*.md` (pass 4 is the signed-off names and concepts, pass 6 the signed-off flavor and mechanics)

## Decisions Ratified in Planning

1. **Concept first; SWN gives the tiers and the numbers** — as Effort 5 Decision 1. Every heir keeps its SWN ancestor's hp, cost, maintenance, tech level, attack and counter, so seeded HP stays valid and the balance is borrowed. 27 rows: 24 SWN slots and 3 new ones (Company Coin, Graveyard Shift, Paid Agitators). Five rows are ★.
2. **Voice and the dark half** — Robert's brief: Effort 4 Decision 2's voice, and Wealth is "the natural place for the darker half of the tone, with debt, labor and company towns". Robert's call on the weight: an even split, 13 dark rows and 14 trade rows. Played straight: Debtors' Fleet and the Aether-Tech Factory stay on the right side of the canon's out-of-bounds line ("slavery played for laughs"). Names are Robert's, set over four passes.
3. **Drafts in two phases, a new file per pass** — Robert's call. Phase 1 settles names and concepts only, with an "Inspired by" column naming the SWN asset or mechanic each row builds on (passes 1 to 4). Phase 2 adds flavor and mechanics (passes 5 and 6). A pass is never rewritten.
4. **A new H.A.W.K type: Workforce** — Robert's call. Labor as a type, on Temp Workers, Prospectors and Graveyard Shift. Company Coin counts it and Mending Vats revive it. Rejected: Holding; no new type.
5. **Workforce is bought only on a settled place or bigger** — Robert's call. `purchase_population` gains `Workforce = 100000`: labor is hired on Veil, Haven, Eventide, Last Light, Odium, Lumos, Ouroboros and Soma, not on the gates, Canon IV, Somewhere, Lethe or Arcadia. It limits buying only. Rejected: no floor; 5,000.
6. **Wealth transports split as SWN's do** — Robert's call. The freight heirs (Free Traders, Merchant Convoy) refuse Force cargo; Back-Channel Freight carries Security Personnel and Operatives of any category past a refused permission. All three are long-range (19), with drift ratings. Rejected: Force cargo on every Wealth transport; none.
7. **Carrier Corps: a long-range carrier for Force** — Robert's call; amends Effort 4 Decision 17 ("Force's far movers teleport"). A new Force row, `HAWK-F6-004`, an Aethership at speed 19 and drift 3 that carries up to ten Force assets for 1 Coin. Tier 6, Robert's call: it is very powerful. Rejected: a Wealth row that carries Force only; tier 5.
8. **Speed classes as Effort 4 Decision 20** — long (19): Free Traders, Merchant Convoy, Back-Channel Freight, Debtors' Fleet, Carrier Corps. Short (10): Prospectors, Free Company (SWN's Mercenaries: speed 1).
9. **The Aether-Gate Transit Authority is open to every faction** — Robert's calls. Transit Web's free move, made public: any faction may move its non-Aethership Cunning and Wealth assets between places within 6 hexes of the Authority, with no action. The owner rides free; any other faction pays 1 Coin per asset, to the owner. The owner cannot refuse a rider; an embargo is filed as `F.054`. While the Authority is hidden, only its owner rides (Decision Record 6). Range 6, cut from 12 after the live turn (Decision Record 5). Accepted cost: the Authority sits on Veil, so the factions near it are offered a ride each turn, declined with Esc. Rejected: a flat fare; owner approval per ride; a hex-range grant for the owner only; a travelling network.
10. **A public grant is found through every faction's asset hooks, not the global scope** — `GrantedMovementAbilities` walks the rider's own granters, then every other faction's asset buckets. A granter that does not serve the rider returns nothing, so only a public grant reaches a rival. A grant registered globally would survive `ClearAsset` on a capture and be offered twice after a recapture, the double registration `registerCapturedAssets` (`internal/faction/engine/orchestrator.go`) guards against.
11. **The fare is the rider's own cost, paid to the owner** — the relocate charges the rider as today, caused by the rider, then a separate `CoinDelta` credits the owner. Liability Waivers (a rival-imposed loss) cannot cancel a fare, so no Coin is made from nothing.
12. **`B.052` folds into the public mover** — Robert's call. The relocate prompt names the asset it runs through, its owner and the fare: "via Aether-Gate Transit Authority (Fey Royals) · 1 Coin per asset". The adapter derives the line from the acting asset's definition, through `GrantMovementParams.Fare`, which the granter uses too; the collector interface and its 21 test call sites stay as they are. Rejected: a new collector argument.
13. **Company Coin pays per worker, at most 3, one per place** — Robert's calls. At turn start, 1 Coin per Workforce asset of the faction on its place, capped at 3; if a faction has several Company Coins on one place, only the lowest-ID one pays. Cost 5. Uncapped, five 2-Coin workers would out-earn the tier 7 Factory. Built as data: `turn_start`, scope `world`, an `asset_type` filter, `no_stack`, and a new `max`. Rejected: 1 Coin per two workers; no cap, priced up; flat 1 Coin a turn.
14. **Debtors' Fleet earns 1d4 Coin from no one on each hit** — Robert's calls. The crew's debt payments: each attack or counterattack hit pays the owner 1d4, taken from no one. Upkeep 1 (SWN: 2), so the top Wealth asset makes money when it fights. It carries `S`, which SWN's Scavenger Fleet did not, because the effect needs it. Rejected: `coin_steal` from the struck faction.
15. **Foreclosure Notices test Wealth vs. Wealth** — Robert's call; SWN's Marketers test Cunning vs. Wealth. It differs from the Geas-Backed NDA (Cunning vs. Cunning) in more than name.
16. **Mending Vats revive Security Personnel, Operatives and Workforce** — closes Effort 4 Decision 11's type gap, the last of it with the stand-ins gone.
17. **SWN's types leave** — `asset_types` lists H.A.W.K's eight; `Starship` leaves `purchase_population` and the `exclude_types` of Teleportation Hubs and the Deep-Insertion Gate, which the loader checks against `asset_types`.
18. **The seed swap is by suffix, plus two extras** — every seeded Wealth stand-in has its heir at the same slot (Effort 4 Decision 14), so `SWN-W` becomes `HAWK-W`. The Thymian Armada gains a Company Coin at Last Light, beside its Temp Workers and Prospectors, so it pays 2 Coin; Seraph Industries gains a Debtors' Fleet on Veil, where the Fey Royals give it a target. Both are extras beyond the scale count, as Effort 5's Glamoured Doubles, so the live turn runs the new code. From a side review.
19. **Robert plays the live turn** — overview Decision 4. Claude may test on `hawk-claude` when Robert asks.
20. **Two phases** — overview Decision 6: the builds (Decisions 9 to 14), then the data.

## Decision Record — Execution

1. **Commit 1: the `max` load cases sit beside the existing ones** — Claude's call; drift in Task 5. No `coin_gain = { amount = "2" }` case exists: the existing case is `amount = "1"` in `TestLoadEffects_ParsesEntries`, which still asserts `Max` 0, so it stands as the "still loads" case. The `max = 3` case joins `TestLoadEffects_ParsesEveryParameterBlock`, and the `max = -1` case joins `TestLoadEffects_LoudFailure`.
2. **Commit 1: a tag cannot count, but the loader does not say so** — a gap, not fixed here. `coin_gain` at `turn_start` accepts scope `world` for any holder, and `countedAssets` reads the holder's asset, so a tag written with scope `world` would panic at turn start. No data does it. A `Validate` that refuses scope `world` on a tag closes it; left for the effort close. Not a gap after all (Decision Record 9).
3. **Commit 2: the `relocateVia` test is internal, on a built `Adapter`** — Claude's call; Task 9 named `collector_test.go`. The harness loads a shipped rulebook with no public grant, so `TestRelocateVia` sits in `internal/faction/tui/adapter/relocate_via_test.go`, package `adapter`, on an `Adapter` with its own rulebook and factions. It covers a self-jump, a public grant for a rival and its owner, a private flat grant, an ability step, and a definition with no relocate.
4. **Commit 2: the public grant's load case joins `TestLoadEffects_ParsesEveryParameterBlock`** — Claude's call; it checks `Public` only. The existing grant case in `TestLoadEffects_ParsesEntries` still deep-equals with `Public` false.
5. **Commit 4: the Transit Authority's range drops from 12 hexes to 6** — Robert's call; reverses Decision 9's "range 12 matches Mirrorways". In the live turn, 7 of 8 factions were offered a ride. The cause: a teleport's range counts hexes with Drift lanes free (`WorldsWithin`, `internal/faction/engine/world/world.go`), so 12 hexes from Veil reach 12 of the 16 places, and Lumos, three lanes away, counts as 3 hexes. At 6 the Authority reaches Veil, Haven, Lumos, Eventide, Odium and Ouroboros, and its riders are the Fey Royals, Seraph Industries, O.W.L. and Blackwater. A stopgap: Robert's model is that a teleport stays in its region and only the top tier crosses Drift, counted in lanes, filed as `F.055`. Rejected: region only, which leaves the Authority one 2-hex hop between Veil and Haven, below WhisperGate at tier 6. The change lands in Commit 4; Commit 3's Task 1 keeps the 12 it committed.
6. **Effort close: a hidden public grant serves only its owner** — Robert's pick from the effort review; narrows Decision 9's "open to every faction". While the Authority is stealthed or sheltered, `GrantMovementAbilities` offers it to no rider, because the relocate prompt names the granting asset and its owner and would show a rival a hidden asset. The owner still rides.
7. **Effort close: other review fixes** — Robert's picks. The digest credits a CoinDelta to the faction it names, so a fare lands on the owner, not the rider. The loader refuses a public grant with relocate scope `self`. A rider who cannot pay a rival's fare is skipped with no `OnError`; the owner of a grant it cannot pay still hears it. A counting `coin_gain` rolls its amount once per counted asset, not once times the count; the shipped Company Coin pays a flat 1, so play does not change.
8. **Effort close: two review findings deferred** — Robert left them out; neither fires on shipped data. A turn_start `coin_gain` accepts `asset_type`, `no_stack` and `max` on every scope but reads them only on `world`, and `world` with no `asset_type` pays 0; a loader check for the counting params closes it. `relocateVia` in `internal/faction/tui/adapter/action_collector.go` rebuilds the fare from the definition, so a definition with two relocate sources can label one fare and charge another.
9. **Effort close: a counting tag was already refused** — reverses Decision Record 2's gap. `validateEntry` in `internal/faction/engine/effect/registrar.go` has refused scope `self` and `world` on any tag holder since `A.001.7` Effort 1, so a tag never reaches `countedAssets`. No new `Validate`; `TestNewWithRulebook_RefusesACountingTag` (`internal/faction/engine/core_test.go`) pins the `coin_gain` case.

## Sort Table

Primitives are the registrar names in `internal/faction/engine/effect/effects/registrars.go`, the ability effects in `internal/faction/domain/asset.go`, or a transport profile. Surface: `—` plain, `effects` S-side, `ability` A-side, `transport`. ★ = no SWN Wealth asset does it.

| HAWK ID | Name | SWN ancestor | Surface | Call | Primitive | Notes |
|---------|------|--------------|---------|------|-----------|-------|
| HAWK-W1-001 | Wage Garnishers | Franchise | effects | reuse | `coin_steal` | Operative |
| HAWK-W1-002 | Temp Workers | Harvesters | ability | reuse | `coin_delta` | Workforce, A |
| HAWK-W1-003 | Permit Office | Local Investments | effects | reuse | `cost_surcharge` | Facility |
| HAWK-W1-004 | Company Coin | new; Ascension Program's income | effects | build | `coin_gain`, scope world, `max` | ★ Stratagem; per Workforce, at most 3 (Decision 13); Commit 1 |
| HAWK-W2-001 | Free Traders | Freighter Contract | transport | reuse | transport profile | Aethership, long (19), drift 2, no Force cargo |
| HAWK-W2-002 | Contract Cabal | Lawyers | effects | reuse | `restrict_attack` | Operative |
| HAWK-W2-003 | Pinkertons | Union Toughs | — | reuse | — | Security Personnel |
| HAWK-W2-004 | Prospectors | Surveyors | effects | reuse | `bonus_die`, Expand Influence | Workforce, short (10) |
| HAWK-W2-005 | Graveyard Shift | new; Harvesters and Venture Capital | ability | reuse | `coin_delta` | ★ Workforce, A; a 1 destroys it |
| HAWK-W3-001 | Aether Refinery | Postech Industry | ability | reuse | `coin_delta` | Facility, A |
| HAWK-W3-002 | Artificery | Laboratory | effects | reuse | `tech_level_floor` | Facility |
| HAWK-W3-003 | Free Company | Mercenaries | — | reuse | — | Security Personnel, short (10), P, upkeep 1 |
| HAWK-W3-004 | Paid Agitators | new; Skeleton Keys (Cunning) | effects | reuse | `block_ability` | ★ Operative |
| HAWK-W4-001 | Merchant Convoy | Shipping Combine | transport | reuse | transport profile | Aethership, long (19), drift 3, ten cargo, no Force cargo |
| HAWK-W4-002 | Company Town | Monopoly | ability | reuse | `levy_coin` | Facility, A |
| HAWK-W4-003 | Mending Vats | Medical Center | effects | reuse | `restore_asset`, `discount_repair` | Security Personnel, Operative, Workforce (Decision 16) |
| HAWK-W4-004 | Liability Waivers | Bank | effects | reuse | `ignore_coin_loss` | Stratagem |
| HAWK-W5-001 | Foreclosure Notices | Marketers | ability | reuse | `disable_asset` | Stratagem, A; Wealth vs. Wealth (Decision 15) |
| HAWK-W5-002 | Archmagus Consultant | Pretech Researchers | effects | reuse | `tech_level_floor` | Operative, upkeep 1 |
| HAWK-W5-003 | Back-Channel Freight | Blockade Runners | transport | reuse | transport profile, `ignores_permission` | Aethership, long (19), drift 3 |
| HAWK-W6-001 | Adventure Capital | Venture Capital | ability | reuse | `coin_delta` | Facility, A |
| HAWK-W6-002 | Franchise Kits | R&D Department | effects | reuse | `tech_level_floor`, faction | Facility |
| HAWK-W6-003 | Aetheric Futures | Commodities Broker | ability | reuse | `purchase_discount` | Operative, A |
| HAWK-W7-001 | Aether-Tech Factory | Pretech Manufactory | ability | reuse | `coin_delta` | Facility, A |
| HAWK-W7-002 | Predatory Purchase | Hostile Takeover | effects | reuse | `seize_target` | Stratagem |
| HAWK-W7-003 | Aether-Gate Transit Authority | Transit Web | effects | build | `grant_movement` + `public` | ★ Facility; 6 hexes, 1 Coin per asset to the owner (Decisions 9 to 12); Commit 2 |
| HAWK-W8-001 | Debtors' Fleet | Scavenger Fleet | effects | build | `coin_gain` on hits | ★ Aethership, long (19), drift 2, upkeep 1 (Decision 14); Commit 1 |
| HAWK-F6-004 | Carrier Corps | none | transport | reuse | transport profile | Force Aethership, long (19), drift 3, Force cargo only (Decision 7) |

## Shared Context

**Gate.** Every commit lands green on `go build ./...`, `go vet ./...` and `go test ./...` from the repo root. `TestNewWithRulebook_ShippedRulebook` (`internal/faction/engine/core_test.go`) composes and loads `rulebooks/swn` and `rulebooks/hawk` and builds the engine from each, so it checks every new effect, filter, gate, type name and grant block. `TestHawkSeed_MatchesRulebookAndMap` (`internal/faction/engine/hawk_seed_test.go`) checks every seeded definition against the H.A.W.K rulebook, so the TOML, the settings, the Force file and the seed swap land in one commit.

**Campaign commands** take the prefix `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns` and run after `go build -o gm-toolkit ./cmd/gm-toolkit` (overview Decision 5).

## Out of Scope

- **An embargo on the public mover** — `F.054` (Decision 9).
- **Tuning numbers** — tuning waits for play (Decision 1).
- **`B.053`** — the feed's unnamed Coin lines; the fare's credit shows as one more of them.
- **Regenerating the seed for real play** — after the arc closes (Effort 3 Decision 15).
- **The tracker's "H.A.W.K playable" milestone and the arc close bookkeeping** — the initiative close.

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — Coin on hits and by count, and the public mover**
  - Commit 1 — `feat(E6): let coin_gain pay on hits and by count` (5 tasks)
  - Commit 2 — `feat(E6): open a movement grant to every faction, for a fare` (10 tasks)
- [x] **Phase 2 — The Wealth pass**
  - Commit 3 — `feat(E6): replace the Wealth stand-in with H.A.W.K's own assets` (4 tasks)
  - Commit 4 — `docs(E6): write the Wealth canon and play the live turn` (4 tasks)

### Phase 1 — Coin on hits and by count, and the public mover

**In plain words:** `coin_gain` learns to pay once per hit an asset lands, and to pay per matching asset on its place up to a cap. A movement grant can be made public, so every faction may ride it, paying its fare to the owner, and the relocate prompt names whose grant it is and what it costs.

#### Commit 1 — `feat(E6): let coin_gain pay on hits and by count`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
type CoinGainParams struct {
	Amount DiceRoll
}
```
**Replace with:**
```go
// CoinGainParams is a Coin payout. Max caps a counted payout (scope world); 0 is
// no cap.
type CoinGainParams struct {
	Amount DiceRoll
	Max    int
}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

`coin_gain` gets a record of its own, with `max`.

**Find:**
```go
	CoinGain         *amountRecord           `toml:"coin_gain"`
```
**Replace with:**
```go
	CoinGain         *coinGainRecord         `toml:"coin_gain"`
```

**Find:**
```go
type amountRecord struct {
	Amount string `toml:"amount"`
}
```
**Replace with:**
```go
type amountRecord struct {
	Amount string `toml:"amount"`
}

type coinGainRecord struct {
	Amount string `toml:"amount"`
	Max    int    `toml:"max"`
}
```

**Find:**
```go
	case domain.SpecialCoinGain:
		amount, err := convertAmount(definition.Effect, record.CoinGain)
		if err != nil {
			return err
		}
		definition.CoinGain = &domain.CoinGainParams{Amount: amount}
```
**Replace with:**
```go
	case domain.SpecialCoinGain:
		if record.CoinGain == nil {
			return missingParamBlock(definition.Effect)
		}
		amount, err := convertAmount(definition.Effect, &amountRecord{Amount: record.CoinGain.Amount})
		if err != nil {
			return err
		}
		if record.CoinGain.Max < 0 {
			return fmt.Errorf("coin_gain max must be >= 0, got %d", record.CoinGain.Max)
		}
		definition.CoinGain = &domain.CoinGainParams{Amount: amount, Max: record.CoinGain.Max}
```

##### Task 3 — `internal/faction/engine/effect/effects/coin_gain.go`

Full contents:

```go
package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// coinGainRegistrar pays the owner the amount. turn_start fires at the owner's
// turn start: scope = "self" for an asset, "faction" for a tag (Aether Refiners),
// and "world" for an asset that pays the amount once per own asset of the
// asset_type filter on its place, up to max; no_stack lets one such holder per
// place pay (Company Coin). asset_destroyed reacts to each asset or Base destroyed
// in a fight: owner reads as whose asset it was ("own" the owner's, "rival"
// another faction's), and caused_by as who destroyed it. scope = "faction" is any
// world. once_per caps the payouts (War Profiteers). own_attack_hit and
// own_counter_hit pay once per hit the declaring asset lands (Debtors' Fleet).
var coinGainRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerTurnStart: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeSelf, domain.ScopeFaction, domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
			Gates:   []domain.EffectGateName{domain.GateNoStack},
		},
		domain.TriggerAssetDestroyed: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn, domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterCausedBy},
			Gates:   []domain.EffectGateName{domain.GateOncePer},
		},
		domain.TriggerOwnAttackHit:  coinOnHitAccepts,
		domain.TriggerOwnCounterHit: coinOnHitAccepts,
	},
	Register: registerCoinGain,
	Fire:     fireCoinGain,
}

var coinOnHitAccepts = effect.Accepts{
	Owners: []domain.EffectOwner{domain.OwnerOwn},
	Scopes: []domain.EffectScope{domain.ScopeSelf},
}

func fireCoinGain(ctx effect.FireContext) []domain.Mutation {
	amount := dice.Amount(ctx.Entry.CoinGain.Amount, ctx.Roller)
	if ctx.Entry.Filters.Scope == domain.ScopeWorld {
		amount *= countedAssets(ctx)
	}
	if amount <= 0 {
		return nil
	}
	return []domain.Mutation{domain.CoinDelta{
		FactionID:         ctx.Holder.Faction.ID,
		Delta:             amount,
		Cause:             string(domain.SpecialCoinGain),
		CausedByFactionID: ctx.Holder.Faction.ID,
	}}
}

// countedAssets counts the holder's faction's active assets on the holder's place
// whose type the asset_type filter names, up to the entry's max. A holder in
// transit counts nothing. Under no_stack, a holder outranked by a lower-ID active
// asset of its own definition on the same place counts nothing.
func countedAssets(ctx effect.FireContext) int {
	holder := ctx.Holder.Asset
	worldID := holder.Location.WorldID
	if worldID == "" {
		return 0
	}
	count := 0
	for _, asset := range domain.SortedAssets(ctx.Holder.Faction) {
		if asset.Location.WorldID != worldID || !asset.Active() {
			continue
		}
		// deliberate: the same definition outranks, not any counting coin_gain, because no_stack means one Company Coin per place, and a different counting asset is a different payout.
		if ctx.Entry.Gates.NoStack && asset.DefinitionID == holder.DefinitionID && asset.ID < holder.ID {
			return 0
		}
		def := ctx.Rulebook.Assets[asset.DefinitionID]
		if def != nil && slices.Contains(ctx.Entry.Filters.AssetTypes, def.Type) {
			count++
		}
	}
	if limit := ctx.Entry.CoinGain.Max; limit > 0 {
		count = min(count, limit)
	}
	return count
}

func registerCoinGain(ctx effect.RegisterContext) {
	if ctx.Entry.Trigger == domain.TriggerOwnAttackHit || ctx.Entry.Trigger == domain.TriggerOwnCounterHit {
		ctx.Registry.RegisterMutationReactor(
			ctx.Holder.Scope(),
			ctx.Holder.HookID(domain.SpecialCoinGain),
			&coinOnHitReactor{
				holder:  refOf(ctx.Holder),
				counter: ctx.Entry.Trigger == domain.TriggerOwnCounterHit,
				amount:  ctx.Entry.CoinGain.Amount,
				roller:  ctx.Roller,
			},
		)
		return
	}
	ctx.Registry.RegisterMutationReactor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialCoinGain),
		&coinGainReactor{
			holder:   refOf(ctx.Holder),
			owner:    ctx.Entry.Filters.Owner,
			causedBy: ctx.Entry.Filters.CausedBy,
			oncePer:  ctx.Entry.Gates.OncePer,
			amount:   ctx.Entry.CoinGain.Amount,
			roller:   ctx.Roller,
		},
	)
}

// combatLossCauses are the removal causes that mean damage in a fight destroyed an
// asset or a Base: an attack or counterattack, a Cracked Comms self-attack, an
// asset's own on-hit damage, or the free attack on a new Base. A sale, a refit or
// a missed maintenance is not one.
var combatLossCauses = []string{"attack", string(domain.SpecialCounterstrike), string(domain.SpecialSelfDamage), "expand"}

// combatLoss reports whose asset or Base a mutation destroyed in a fight, and who
// destroyed it. A Base of Influence counts as an asset (Scavengers, A.001.7 close).
func combatLoss(mutation domain.Mutation) (factionID, causedByFactionID string, ok bool) {
	switch loss := mutation.(type) {
	case domain.AssetRemoved:
		return loss.FactionID, loss.CausedByFactionID, slices.Contains(combatLossCauses, loss.Cause)
	case domain.BaseDestroyed:
		return loss.FactionID, loss.CausedByFactionID, slices.Contains(combatLossCauses, loss.Cause)
	}
	return "", "", false
}

type coinGainReactor struct {
	holder   holderRef
	owner    domain.EffectOwner
	causedBy domain.EffectOwner
	oncePer  domain.OncePerKey
	amount   domain.DiceRoll
	roller   dice.Roller
}

// OnMutations pays once for each combat removal in the batch that the entry's
// owner and caused_by filters match, while its once_per budget lasts.
func (reactor *coinGainReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	ownerID := reactor.holder.factionID
	var extra []domain.Mutation
	for _, mutation := range mutations {
		lostBy, causedBy, ok := combatLoss(mutation)
		if !ok {
			continue
		}
		if !onSide(reactor.owner, lostBy, ownerID) {
			continue
		}
		if reactor.causedBy != "" && !onSide(reactor.causedBy, causedBy, ownerID) {
			continue
		}
		if reactor.oncePer != "" && !spendBudget(factionState.Factions[ownerID], budgetKey(reactor.oncePer, domain.SpecialCoinGain, reactor.holder.sourceID, reactor.holder.assetID, lostBy)) {
			continue
		}
		extra = append(extra, domain.CoinDelta{
			FactionID:         ownerID,
			Delta:             dice.Amount(reactor.amount, reactor.roller),
			Cause:             string(domain.SpecialCoinGain),
			CausedByFactionID: ownerID,
		})
	}
	return extra
}

// coinOnHitReactor pays the owner the amount for each hit its asset lands in the
// batch: attack hits, or counterattack hits when counter is set. The Coin comes
// from no one (Effort 6 Decision 14).
type coinOnHitReactor struct {
	holder  holderRef
	counter bool
	amount  domain.DiceRoll
	roller  dice.Roller
}

func (reactor *coinOnHitReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	var extra []domain.Mutation
	for range hitTargets(mutations, reactor.holder.assetID, reactor.counter) {
		extra = append(extra, domain.CoinDelta{
			FactionID:         reactor.holder.factionID,
			Delta:             dice.Amount(reactor.amount, reactor.roller),
			Cause:             string(domain.SpecialCoinGain),
			CausedByFactionID: reactor.holder.factionID,
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

##### Task 4 — `internal/faction/engine/effect/effects/coin_gain_test.go`

Add, in the file's existing style (direct `Fire` and `Register` calls, `dicetest.Constant` rolls, `dispatch.MutationReactors` for hooks):

- **`TestCoinGain_CountsWorkforceOnItsPlace`** — a faction holds a counting holder (`scope = world`, `asset_type = ["Workforce"]`, amount 1) on Veil, two active Workforce assets on Veil, one Workforce asset on Haven and one active Facility on Veil. `Fire` pays 2.
- **`TestCoinGain_CountCapsAtMax`** — five Workforce assets on the holder's place, `Max` 3: pays 3. With `Max` 0: pays 5.
- **`TestCoinGain_CountSkipsInactiveAndInTransit`** — a disabled Workforce asset is not counted; a holder with an empty `WorldID` pays nothing (no mutation).
- **`TestCoinGain_NoStackLetsTheLowestIDPay`** — two holders of the same definition on one place, `NoStack` set: firing the higher-ID holder returns no mutation, the lower-ID one pays. Without `NoStack`, both pay.
- **`TestCoinGain_PaysOnEachHit`** — a holder registered on `own_attack_hit` with amount `1d4` and a constant roll of 3; a batch with two attack hits by the holder's asset (`AssetHPDelta`, `Cause` "attack", `CausedByAssetID` the asset) and one by another asset pays two `CoinDelta`s of 3 to the owner. A counterattack hit pays nothing.
- **`TestCoinGain_PaysOnEachCounterHit`** — the same on `own_counter_hit`: counterattack hits pay, attack hits do not.
- **`TestCoinGain_HitPaysNothingWhenInactive`** — the holder's asset is disabled: no mutation.

##### Task 5 — `internal/faction/rulebook/effects_test.go`

Add, beside the existing `coin_gain` load cases:

- `coin_gain = { amount = "1", max = 3 }` loads `CoinGainParams{Amount: {Modifier: 1}, Max: 3}`.
- `coin_gain = { amount = "1", max = -1 }` refuses to load with "coin_gain max must be >= 0".
- An existing `coin_gain = { amount = "2" }` still loads with `Max` 0.

##### Commit message

```
feat(E6): let coin_gain pay on hits and by count

- coin_gain on own_attack_hit and own_counter_hit pays once per hit
  its asset lands, from no one (Debtors' Fleet)
- coin_gain at turn_start with scope world pays per own asset of the
  asset_type filter on its place, capped by a new max; no_stack lets
  the lowest-ID holder of a definition on a place pay (Company Coin)
- rulebook: coin_gain gets its own record with max
- tests for both, and for the max field's load
```

#### Commit 2 — `feat(E6): open a movement grant to every faction, for a fare`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
type GrantMovementParams struct {
	Relocate RelocateParams
}
```
**Replace with:**
```go
// GrantMovementParams is a free move an asset offers. A public grant is offered to
// every faction: its owner rides free, and any other faction pays the relocate's
// coin cost to the owner (Aether-Gate Transit Authority).
type GrantMovementParams struct {
	Relocate RelocateParams
	Public   bool
}

// Fare is what riderID pays per use of a grant an asset of ownerID offers: nothing
// for the owner of a public grant, the relocate's coin cost otherwise.
func (params GrantMovementParams) Fare(ownerID, riderID string) int {
	if params.Public && riderID == ownerID {
		return 0
	}
	return params.Relocate.CoinCost
}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
type grantMovementRecord struct {
	Relocate *relocateRecord `toml:"relocate"`
}
```
**Replace with:**
```go
type grantMovementRecord struct {
	Relocate *relocateRecord `toml:"relocate"`
	Public   bool            `toml:"public"`
}
```

**Find:**
```go
		definition.GrantMovement = &domain.GrantMovementParams{Relocate: *relocate}
```
**Replace with:**
```go
		definition.GrantMovement = &domain.GrantMovementParams{Relocate: *relocate, Public: record.GrantMovement.Public}
```

##### Task 3 — `internal/faction/engine/hooks/rule_modifier.go`

**Find:**
```go
// MovementAbility is a free relocation a granter offers its faction this turn.
// Source is the granting asset; range is measured from it.
type MovementAbility struct {
	Source   *domain.Asset
	Relocate domain.RelocateParams
}
```
**Replace with:**
```go
// MovementAbility is a free relocation a granter offers a faction this turn.
// Source is the granting asset; range is measured from it. Payee is the faction
// the relocate's coin cost is paid to, or empty when the Coin is spent.
type MovementAbility struct {
	Source   *domain.Asset
	Relocate domain.RelocateParams
	Payee    string
}
```

##### Task 4 — `internal/faction/engine/effect/effects/grant_movement.go`

Full contents:

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// grantMovementRegistrar reads owner = "own" as "the grant is the declaring
// asset's faction's", and scope = "faction" as "the riding faction's own assets
// move". The range and the asset filters come from the relocate block, measured
// from the declaring asset. A public grant serves every faction (Effort 6
// Decision 9).
var grantMovementRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerMovementOptionsQuery: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerGrantMovement,
	// deliberate: refuse a tag, not accept it, because the range is measured from the declaring asset and a tag has none; Mercenary Group is a speed floor, not a grant (Effort 2 Decision 7).
	Validate: refuseTagHolder,
}

func registerGrantMovement(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetMovementGranter(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialGrantMovement),
		&movementGranter{
			owner:  ctx.Holder.Faction,
			holder: refOf(ctx.Holder),
			grant:  *ctx.Entry.GrantMovement,
		},
	)
}

// movementGranter reads its asset through its owner at query time (Decision
// Record 19), so a grant from an asset removed or made inactive mid-cycle lapses.
type movementGranter struct {
	owner  *domain.Faction
	holder holderRef
	grant  domain.GrantMovementParams
}

// GrantMovementAbilities offers the grant to its owner, and to any other faction
// when it is public. A rider other than the owner pays the fare to the owner.
func (granter *movementGranter) GrantMovementAbilities(faction *domain.Faction) []hooks.MovementAbility {
	rider := faction.ID != granter.owner.ID
	if rider && !granter.grant.Public {
		return nil
	}
	asset, live := granter.holder.liveIn(granter.owner)
	if !live {
		return nil
	}
	relocate := granter.grant.Relocate
	relocate.CoinCost = granter.grant.Fare(granter.owner.ID, faction.ID)
	ability := hooks.MovementAbility{Source: asset, Relocate: relocate}
	if rider {
		ability.Payee = granter.owner.ID
	}
	return []hooks.MovementAbility{ability}
}
```

##### Task 5 — `internal/faction/engine/hooks/dispatch/rules.go`

**Find:**
```go
// GrantedMovementAbilities collects the free movement offered to faction: its
// global and faction granters, then those registered against each asset it owns,
// in asset ID order. A grant serves its owner only, so a rival's assets are never
// asked. Returns nil when registry is nil or no granter offers anything.
func GrantedMovementAbilities(registry *hooks.Registry, faction *domain.Faction) []hooks.MovementAbility {
	if registry == nil {
		return nil
	}
	var abilities []hooks.MovementAbility
	for _, registered := range registry.AssetMovementGrantersFor(faction.ID, "") {
		abilities = append(abilities, registered.Hook.GrantMovementAbilities(faction)...)
	}
	for _, asset := range domain.SortedAssets(faction) {
		for _, registered := range registry.AssetMovementGrantersForAsset(faction.ID, asset.ID) {
			abilities = append(abilities, registered.Hook.GrantMovementAbilities(faction)...)
		}
	}
	return abilities
}
```
**Replace with:**
```go
// GrantedMovementAbilities collects the free movement offered to faction: its
// global and faction granters, then those registered against each asset it owns,
// in asset ID order, then those registered against every other faction's assets,
// by faction and asset ID. A granter that does not serve faction offers nothing,
// so a rival's asset offers only a public grant (Effort 6 Decision 10). Returns
// nil when registry is nil or no granter offers anything.
func GrantedMovementAbilities(registry *hooks.Registry, factions map[string]*domain.Faction, faction *domain.Faction) []hooks.MovementAbility {
	if registry == nil {
		return nil
	}
	var abilities []hooks.MovementAbility
	for _, registered := range registry.AssetMovementGrantersFor(faction.ID, "") {
		abilities = append(abilities, registered.Hook.GrantMovementAbilities(faction)...)
	}
	for _, asset := range domain.SortedAssets(faction) {
		for _, registered := range registry.AssetMovementGrantersForAsset(faction.ID, asset.ID) {
			abilities = append(abilities, registered.Hook.GrantMovementAbilities(faction)...)
		}
	}
	for _, ownerID := range slices.Sorted(maps.Keys(factions)) {
		if ownerID == faction.ID {
			continue
		}
		for _, asset := range domain.SortedAssets(factions[ownerID]) {
			for _, registered := range registry.AssetMovementGrantersForAsset(ownerID, asset.ID) {
				abilities = append(abilities, registered.Hook.GrantMovementAbilities(faction)...)
			}
		}
	}
	return abilities
}
```

`maps` and `slices` are already imported.

##### Task 6 — `internal/faction/engine/relocate/relocate.go`

`Request` gains the payee.

**Find:**
```go
	FactionState *state.FactionState
	Collector    Collector
}
```
**Replace with:**
```go
	FactionState *state.FactionState
	Collector    Collector
	// Payee is the faction each moved order's coin cost is paid to, or empty when
	// the Coin is spent (a public grant's fare, Effort 6 Decision 11).
	Payee string
}
```

**Find:**
```go
		mutations = append(mutations, muts...)
		mutations = append(mutations, dispatch.TollCharges(order.Asset.OwnerID, tolls[i])...)
		moved++
```
**Replace with:**
```go
		mutations = append(mutations, muts...)
		mutations = append(mutations, dispatch.TollCharges(order.Asset.OwnerID, tolls[i])...)
		if request.Payee != "" && orderCost > 0 {
			// deliberate: the rider's charge stays caused by the rider, not the payee, because the rider chose the fare; Liability Waivers would otherwise cancel it while the payee is still paid.
			mutations = append(mutations, domain.CoinDelta{FactionID: request.Payee, Delta: orderCost, Cause: "fare", CausedByFactionID: request.Faction.ID})
		}
		moved++
```

##### Task 7 — `internal/faction/engine/orchestrator.go`

In `runGrantedMovement`, the dispatch reads every faction, and the relocate carries the payee.

**Find:**
```go
	for _, grant := range dispatch.GrantedMovementAbilities(e.Hooks, faction) {
		params := grant.Relocate
		mutations, err := relocate.Run(relocate.Request{
			Faction:      faction,
			Acting:       grant.Source,
			Params:       &params,
			World:        e.World,
			Rulebook:     e.Rulebook,
			Registry:     e.Hooks,
			FactionState: factionState,
			Collector:    collector,
		})
```
**Replace with:**
```go
	for _, grant := range dispatch.GrantedMovementAbilities(e.Hooks, factionState.Factions, faction) {
		params := grant.Relocate
		mutations, err := relocate.Run(relocate.Request{
			Faction:      faction,
			Acting:       grant.Source,
			Params:       &params,
			World:        e.World,
			Rulebook:     e.Rulebook,
			Registry:     e.Hooks,
			FactionState: factionState,
			Collector:    collector,
			Payee:        grant.Payee,
		})
```

##### Task 8 — the prompt names the source and the fare (`B.052`)

`internal/faction/tui/adapter/channels.go`:

**Find:**
```go
type SelectRelocateOrdersPayload struct {
	Acting          *domain.Asset
	Candidates      []*domain.Asset
	CandidateWorlds []string
	WorldNames      map[string]string // worldID -> display name (adapter-built)
	MaxAssets       int               // 0 = unbounded
}
```
**Replace with:**
```go
type SelectRelocateOrdersPayload struct {
	Acting          *domain.Asset
	Candidates      []*domain.Asset
	CandidateWorlds []string
	WorldNames      map[string]string // worldID -> display name (adapter-built)
	MaxAssets       int               // 0 = unbounded
	Via             string            // the asset the move runs through, its owner and the fare; empty for a self-jump
}
```

`internal/faction/tui/adapter/action_collector.go`:

**Find:**
```go
		WorldNames:      a.worldNames(candidateWorlds),
		MaxAssets:       maxAssets,
	})
	if err != nil {
		return nil, err
	}
	return replyAs[[]action.RelocateOrder]("SelectRelocateOrders", raw)
}
```
**Replace with:**
```go
		WorldNames:      a.worldNames(candidateWorlds),
		MaxAssets:       maxAssets,
		Via:             a.relocateVia(acting, candidates),
	})
	if err != nil {
		return nil, err
	}
	return replyAs[[]action.RelocateOrder]("SelectRelocateOrders", raw)
}

// relocateVia names the asset a relocation runs through, its owner and what the
// move costs the rider: "via Transit Web (Fey Royals) · 1 Coin". A self-jump runs
// through nothing and names nothing (B.052).
func (a *Adapter) relocateVia(acting *domain.Asset, candidates []*domain.Asset) string {
	if len(candidates) == 0 || slices.Contains(candidates, acting) || a.engine.Rulebook == nil {
		return ""
	}
	def := a.engine.Rulebook.Assets[acting.DefinitionID]
	if def == nil {
		return ""
	}
	params, fare, ok := relocateCost(def, acting.OwnerID, candidates[0].OwnerID)
	if !ok {
		return ""
	}
	owner := acting.OwnerID
	if faction := a.factionState.Factions[acting.OwnerID]; faction != nil {
		owner = faction.Name
	}
	return fmt.Sprintf("via %s (%s) · %s", def.Name, owner, fareLabel(fare, params.CoinCostMode))
}

// relocateCost finds the relocate block def offers, from a movement grant or an
// ability step, and the fare riderID pays for it.
func relocateCost(def *domain.AssetDefinition, ownerID, riderID string) (domain.RelocateParams, int, bool) {
	for _, entry := range def.Effects {
		if entry.GrantMovement != nil {
			return entry.GrantMovement.Relocate, entry.GrantMovement.Fare(ownerID, riderID), true
		}
	}
	if def.Ability != nil {
		for _, step := range def.Ability.Steps {
			if step.Relocate != nil {
				return *step.Relocate, step.Relocate.CoinCost, true
			}
		}
	}
	return domain.RelocateParams{}, 0, false
}

func fareLabel(fare int, mode domain.RelocateCostMode) string {
	switch {
	case fare == 0:
		return "free"
	case mode == domain.RelocateCostFlat:
		return fmt.Sprintf("%d Coin", fare)
	default:
		return fmt.Sprintf("%d Coin per asset", fare)
	}
}
```

`fmt` and `slices` are already imported.

`internal/faction/tui/views/turn/overlay/relocateorders.go`:

**Find:**
```go
		sel := huh.NewSelect[string]().
			Title(fmt.Sprintf("%s — relocate to?", name)).
			Options(worldOpts...).
			Value(&data.dests[idx])
```
**Replace with:**
```go
		sel := huh.NewSelect[string]().
			Title(fmt.Sprintf("%s — relocate to?", name)).
			Description(payload.Via).
			Options(worldOpts...).
			Value(&data.dests[idx])
```

##### Task 9 — tests

`internal/faction/engine/hooks/dispatch/rules_test.go`, in `TestGrantedMovementAbilities_WalksFactionThenAssets`:

**Find:**
```go
	abilities := GrantedMovementAbilities(registry, faction)
```
**Replace with:**
```go
	abilities := GrantedMovementAbilities(registry, map[string]*domain.Faction{"f1": faction}, faction)
```

In `TestGrantedMovementAbilities_NilRegistry`:

**Find:**
```go
	if got := GrantedMovementAbilities(nil, &domain.Faction{ID: "f1"}); got != nil {
```
**Replace with:**
```go
	if got := GrantedMovementAbilities(nil, nil, &domain.Faction{ID: "f1"}); got != nil {
```

Add:

- **`TestGrantedMovementAbilities_AsksOtherFactionsAssetsLast`** — `f1` with asset `a`, rivals `r2` (asset `x`) and `r1` (asset `y`), each with a stub granter registered at its asset scope; the factions map holds all three. The sources come back as `a`, `y`, `x`: own assets first, then rivals by faction ID.

`internal/faction/engine/effect/effects/grant_movement_test.go`, with a public variant of `registerWeb` (`GrantMovementParams{Relocate: transitWebRelocate(), Public: true}`):

- **`TestGrantMovement_PublicGrantServesARivalForAFare`** — beta is offered one ability from alpha's web, with `Relocate.CoinCost` 1 and `Payee` "alpha".
- **`TestGrantMovement_PublicGrantIsFreeForItsOwner`** — alpha is offered the grant with `CoinCost` 0 and no `Payee`.
- **`TestGrantMovement_PrivateGrantKeepsItsCost`** — the existing private web still offers alpha `CoinCost` 1 and no `Payee`, and beta nothing (`TestGrantMovement_RivalFactionGetsNothing` already covers beta).

`internal/faction/domain/effect_test.go`:

- **`TestGrantMovementParams_Fare`** — a public grant: the owner pays 0, a rider pays `CoinCost`. A private grant: both pay `CoinCost`.

`internal/faction/engine/relocate/relocate_test.go`:

- **`TestRun_PayeeIsPaidEachMovedOrdersCost`** — a per-asset relocate with `CoinCost` 1 and `Payee` "owner": two orders move, the mover is charged 2 as today, and two `CoinDelta`s of 1 to "owner" with `Cause` "fare" and `CausedByFactionID` the mover. A refused order (permission) pays the payee nothing.
- **`TestRun_NoPayeeCreditsNoOne`** — without `Payee`, no `fare` mutation.

`internal/faction/rulebook/effects_test.go`:

- `grant_movement = { public = true, relocate = { … } }` loads with `Public` true; the existing grant case loads with `Public` false.

`internal/faction/tui/adapter/collector_test.go`:

- **`TestRelocateVia`** — `relocateCost` and `fareLabel` through `relocateVia` where a harness definition allows: a candidate list that holds the acting asset gives ""; a grant source gives "via <name> (<owner name>) · 1 Coin per asset"; the owner of a public grant reads "free". If the harness cannot build a public grant, test `relocateCost` and `fareLabel` directly in an internal `_test.go` in package `adapter`.

##### Task 10 — the tracker and this plan

This plan's Work Breakdown Overview: `- [ ] **Phase 1 — Coin on hits and by count, and the public mover**` becomes `- [x] **Phase 1 — Coin on hits and by count, and the public mover**`.

`dev_journals/faction-manager/planned-work.md`: `B.052` is fixed, so its row leaves Bugfixes.

**Find:**
```markdown
| `B.052` | Granted move names no source | -- | A movement grant (Transit Web) asks "Fey Courtiers — relocate to?" once per eligible asset (`overlay/relocateorders.go:52`) and never says which asset grants the move or what it costs, so it reads as the asset's own jump; name the granting asset and the Coin. Found in the `A.001.5` Effort 5 live turn. |
```
**Replace with:** nothing (the line is removed).

##### Commit message

```
feat(E6): open a movement grant to every faction, for a fare

- grant_movement takes public: a public grant is offered to every
  faction; its owner rides free, and any other faction pays the
  relocate's coin cost to the owner (GrantMovementParams.Fare)
- GrantedMovementAbilities also asks every other faction's asset
  granters, so a rival's public grant reaches the rider
- relocate.Request takes a Payee: each moved order's cost is
  credited to it, while the rider's charge stays the rider's own
- the relocate prompt names the asset it runs through, its owner and
  the fare (B.052)
- tests; B.052 leaves the tracker; Phase 1 ticked
```

### Phase 2 — The Wealth pass

**In plain words:** H.A.W.K's 27 Wealth assets replace the SWN copy, Force gains Carrier Corps, SWN's asset types leave the settings, and the seed swaps its Wealth units to their heirs and gains a Company Coin and a Debtors' Fleet. Then the canon is written and Robert plays a live turn: every rulebook file is H.A.W.K's own.

#### Commit 3 — `feat(E6): replace the Wealth stand-in with H.A.W.K's own assets`

##### Task 1 — `rulebooks/hawk/assets/wealth_assets.toml`

Full contents:

```toml
# Wealth Assets — H.A.W.K (A.001.5 Effort 6)

[assets.WageGarnishers]
id = "HAWK-W1-001"
name = "Wage Garnishers"
category = "Wealth"
min_rating = 1
hp = 3
cost = 2
maintenance = 0
tech_level = 2
type = "Operative"
flags = ["S"]
counter = "1d4-1"
description = "Lenders to a rival's staff, repaid from the rival's payroll. When the Garnishers hit a rival asset, that faction loses 1 Coin to this one. This happens once per turn, however many Garnishers hit."
attack = { attacker_stat = "Wealth", defender_stat = "Wealth", damage = "1d4" }
effects = [
  {
    trigger = "own_attack_hit",
    effect = "coin_steal",
    filters = { owner = "rival", scope = "world" },
    gates = { once_per = "target_faction" },
    coin_steal = { amount = "1" },
  },
]

[assets.TempWorkers]
id = "HAWK-W1-002"
name = "Temp Workers"
category = "Wealth"
min_rating = 1
hp = 4
cost = 2
maintenance = 0
tech_level = 0
type = "Workforce"
flags = ["A"]
counter = "1d4"
description = "Workers on temporary contracts, called in when there is work. As an action, roll 1d6; on a 3 or higher, gain 1 Coin."
ability = { effect = "coin_delta", target_roll = { die = "1d6", target = 3, amount = 1 } }

[assets.PermitOffice]
id = "HAWK-W1-003"
name = "Permit Office"
category = "Wealth"
min_rating = 1
hp = 2
cost = 1
maintenance = 0
tech_level = 2
type = "Facility"
flags = ["S"]
counter = ""
description = "The office that approves building on this place. Any other faction that buys an asset here pays 1 extra Coin, which is lost. Only one Permit Office on a place counts."
attack = { attacker_stat = "Wealth", defender_stat = "Wealth", damage = "1d4-1" }
effects = [
  {
    trigger = "asset_cost_query",
    effect = "cost_surcharge",
    filters = { owner = "rival", scope = "world" },
    gates = { no_stack = true },
    cost_surcharge = { amount = 1 },
  },
]

[assets.CompanyCoin]
id = "HAWK-W1-004"
name = "Company Coin"
category = "Wealth"
min_rating = 1
hp = 2
cost = 5
maintenance = 0
tech_level = 2
type = "Stratagem"
flags = ["S"]
counter = ""
description = "Wages paid in coin only the company's stores accept. At the start of each turn, gain 1 Coin for each of the faction's Workforce assets on this place, up to 3. Only one Company Coin on a place pays."
effects = [
  {
    trigger = "turn_start",
    effect = "coin_gain",
    filters = { owner = "own", scope = "world", asset_type = ["Workforce"] },
    gates = { no_stack = true },
    coin_gain = { amount = "1", max = 3 },
  },
]

[assets.FreeTraders]
id = "HAWK-W2-001"
name = "Free Traders"
category = "Wealth"
min_rating = 2
hp = 4
cost = 5
maintenance = 0
tech_level = 4
speed = 19
drift_rating = 2
type = "Aethership"
flags = ["S"]
counter = ""
description = "Independent long-haul freighters. Moves at speed 19 (long-range). When it moves, it can carry one non-Force asset from its place with it, for 1 Coin."
attack = { attacker_stat = "Wealth", defender_stat = "Wealth", damage = "1d4" }
transport = {
  coin_cost = 1,
  cargo_types = ["Facility", "Aethership", "Security Personnel", "Operative", "Stratagem", "Front", "Workforce", "Beast"],
  max_cargo = 1,
  exclude_categories = ["Force"],
}

[assets.ContractCabal]
id = "HAWK-W2-002"
name = "Contract Cabal"
category = "Wealth"
min_rating = 2
hp = 4
cost = 6
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["S"]
counter = "1d6"
description = "Lawyers fluent in contract magic. They cannot attack or counterattack Force assets, and Force assets cannot attack them."
attack = { attacker_stat = "Cunning", defender_stat = "Wealth", damage = "2d4" }
effects = [
  {
    trigger = "attack_eligibility_query",
    effect = "restrict_attack",
    filters = { owner = "rival", scope = "self", category = ["Force"] },
    restrict_attack = { side = "attack_and_counter", allow = "except_listed" },
  },
]

[assets.Pinkertons]
id = "HAWK-W2-003"
name = "Pinkertons"
category = "Wealth"
min_rating = 2
hp = 6
cost = 4
maintenance = 0
tech_level = 0
type = "Security Personnel"
flags = []
counter = "1d4"
description = "Hired muscle sent to end a strike before the next shift."
attack = { attacker_stat = "Wealth", defender_stat = "Force", damage = "1d4+1" }

[assets.Prospectors]
id = "HAWK-W2-004"
name = "Prospectors"
category = "Wealth"
min_rating = 2
hp = 4
cost = 4
maintenance = 0
tech_level = 0
speed = 10
type = "Workforce"
flags = ["S"]
counter = "1d4"
description = "Prospectors who test a seam with the cheapest crew first. Moves at speed 10 (short-range). Expand Influence on this place rolls one extra die."
effects = [
  {
    trigger = "own_roll",
    effect = "bonus_die",
    filters = { owner = "own", scope = "world", phase = ["faction_test"], action = "Expand Influence" },
    bonus_die = { sides = 10, count = 1 },
  },
]

[assets.GraveyardShift]
id = "HAWK-W2-005"
name = "Graveyard Shift"
category = "Wealth"
min_rating = 2
hp = 4
cost = 4
maintenance = 0
tech_level = 0
type = "Workforce"
flags = ["A"]
counter = "1d4"
description = "The crew that works through the night. As an action, roll 1d6: on a 1, the asset is destroyed; on a 2–4, gain 1 Coin; on a 5–6, gain 2 Coin."
ability = {
  effect = "coin_delta",
  die = "1d6",
  outcomes = [
    { from = 1, to = 1, destroy_self = true },
    { from = 2, to = 4, coin = 1 },
    { from = 5, to = 6, coin = 2 },
  ],
}

[assets.AetherRefinery]
id = "HAWK-W3-001"
name = "Aether Refinery"
category = "Wealth"
min_rating = 3
hp = 4
cost = 8
maintenance = 0
tech_level = 4
type = "Facility"
flags = ["A"]
counter = "1d4"
description = "A refinery that boils raw Aether down into stones. As an action, roll 1d6: on a 1, lose 1 Coin, and if it can't be paid, the asset is destroyed; on a 2–4, gain 1 Coin; on a 5–6, gain 2 Coin."
ability = {
  effect = "coin_delta",
  die = "1d6",
  outcomes = [
    { from = 1, to = 1, coin = -1 },
    { from = 2, to = 4, coin = 1 },
    { from = 5, to = 6, coin = 2 },
  ],
}

[assets.Artificery]
id = "HAWK-W3-002"
name = "Artificery"
category = "Wealth"
min_rating = 3
hp = 4
cost = 6
maintenance = 0
tech_level = 0
type = "Facility"
flags = ["S"]
counter = ""
description = "A workshop licensed to build Arcano-Tech. Assets may be bought on this place as if it were tech level 4."
effects = [
  {
    trigger = "tech_level_query",
    effect = "tech_level_floor",
    filters = { owner = "own", scope = "world" },
    tech_level_floor = { level = 4 },
  },
]

[assets.FreeCompany]
id = "HAWK-W3-003"
name = "Free Company"
category = "Wealth"
min_rating = 3
hp = 6
cost = 8
maintenance = 1
tech_level = 4
speed = 10
type = "Security Personnel"
flags = ["S", "P"]
counter = "1d6"
description = "A mercenary company with a rate card. Costs 1 Coin per turn in upkeep. Moves at speed 10 (short-range). Buying it on a place, or moving it there, needs the government's permission."
attack = { attacker_stat = "Wealth", defender_stat = "Force", damage = "2d4+2" }

[assets.PaidAgitators]
id = "HAWK-W3-004"
name = "Paid Agitators"
category = "Wealth"
min_rating = 3
hp = 6
cost = 7
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["S"]
counter = ""
description = "Organizers on the faction's payroll. An asset they attack cannot use its ability until the start of this faction's next turn, whether or not the attack hits."
attack = { attacker_stat = "Wealth", defender_stat = "Wealth", damage = "2d4" }
effects = [
  { trigger = "own_attack", effect = "block_ability", filters = { owner = "own", scope = "self" } },
]

[assets.MerchantConvoy]
id = "HAWK-W4-001"
name = "Merchant Convoy"
category = "Wealth"
min_rating = 4
hp = 10
cost = 10
maintenance = 0
tech_level = 4
speed = 19
drift_rating = 3
type = "Aethership"
flags = ["S"]
counter = "1d6"
description = "A scheduled convoy of freighters. Moves at speed 19 (long-range). When it moves, it can carry up to ten non-Force assets from its place with it, for 1 Coin."
transport = {
  coin_cost = 1,
  cargo_types = ["Facility", "Aethership", "Security Personnel", "Operative", "Stratagem", "Front", "Workforce", "Beast"],
  max_cargo = 10,
  exclude_categories = ["Force"],
}

[assets.CompanyTown]
id = "HAWK-W4-002"
name = "Company Town"
category = "Wealth"
min_rating = 4
hp = 12
cost = 8
maintenance = 0
tech_level = 3
type = "Facility"
flags = ["A"]
counter = "1d6"
description = "A town owned by one company. As an action, force one other faction with unstealthed assets on this place to pay this faction 1 Coin. If it can't pay, it loses one of its assets here, of its choice."
attack = { attacker_stat = "Wealth", defender_stat = "Wealth", damage = "1d6" }
ability = { effect = "levy_coin", target = "faction" }

[assets.MendingVats]
id = "HAWK-W4-003"
name = "Mending Vats"
category = "Wealth"
min_rating = 4
hp = 8
cost = 12
maintenance = 0
tech_level = 4
type = "Facility"
flags = ["S"]
counter = ""
description = "Alchemical vats that knit the injured back together. Once per asset, when a Security Personnel, Operative or Workforce asset on this place is destroyed, the faction may pay half its cost to restore it at 1 HP. Repairing those assets here costs 1 less Coin."
effects = [
  {
    trigger = "lethal_damage",
    effect = "restore_asset",
    filters = { owner = "own", scope = "world", asset_type = ["Security Personnel", "Operative", "Workforce"] },
    gates = { once_per = "asset" },
  },
  {
    trigger = "repair_cost_query",
    effect = "discount_repair",
    filters = { owner = "own", scope = "world", asset_type = ["Security Personnel", "Operative", "Workforce"] },
    gates = { no_stack = true },
    discount_repair = { amount = 1 },
  },
]

[assets.LiabilityWaivers]
id = "HAWK-W4-004"
name = "Liability Waivers"
category = "Wealth"
min_rating = 4
hp = 8
cost = 12
maintenance = 0
tech_level = 3
type = "Stratagem"
flags = ["S"]
counter = ""
description = "Waivers signed in every deal. Once per turn, the faction can ignore one Coin cost or loss another faction imposes. Several Waivers ignore several losses."
effects = [
  {
    trigger = "coin_loss",
    effect = "ignore_coin_loss",
    filters = { owner = "rival", scope = "faction" },
    gates = { once_per = "asset" },
  },
]

[assets.ForeclosureNotices]
id = "HAWK-W5-001"
name = "Foreclosure Notices"
category = "Wealth"
min_rating = 5
hp = 8
cost = 10
maintenance = 0
tech_level = 0
type = "Stratagem"
flags = ["A"]
counter = ""
description = "A claim served for a debt. As an action, test Wealth vs. Wealth against a rival faction's asset; on a success, that faction must at once pay half the asset's purchase cost (rounded down), or the asset is disabled until the price is paid."
attack = { attacker_stat = "Wealth", defender_stat = "Wealth", damage = "1d6" }
ability = {
  roll = "test",
  attacker_stat = "Wealth",
  defender_stat = "Wealth",
  effect = "disable_asset",
  target = "asset",
}

[assets.ArchmagusConsultant]
id = "HAWK-W5-002"
name = "Archmagus Consultant"
category = "Wealth"
min_rating = 5
hp = 6
cost = 12
maintenance = 1
tech_level = 5
type = "Operative"
flags = ["S"]
counter = ""
description = "A famous archmage on retainer. This place counts as tech level 5 for buying Cunning and Wealth assets. Costs 1 Coin per turn in upkeep."
effects = [
  {
    trigger = "tech_level_query",
    effect = "tech_level_floor",
    filters = { owner = "own", scope = "world", category = ["Cunning", "Wealth"] },
    tech_level_floor = { level = 5 },
  },
]

[assets.BackChannelFreight]
id = "HAWK-W5-003"
name = "Back-Channel Freight"
category = "Wealth"
min_rating = 5
hp = 6
cost = 12
maintenance = 0
tech_level = 4
speed = 19
drift_rating = 3
type = "Aethership"
flags = ["S"]
counter = "2d4"
description = "Freight that moves on arrangements made off the record. Moves at speed 19 (long-range). When it moves, it can carry one Security Personnel or Operative from its place with it, for 2 Coin, even where the government refuses permission."
transport = {
  coin_cost = 2,
  cargo_types = ["Security Personnel", "Operative"],
  max_cargo = 1,
  ignores_permission = true,
}

[assets.AdventureCapital]
id = "HAWK-W6-001"
name = "Adventure Capital"
category = "Wealth"
min_rating = 6
hp = 10
cost = 15
maintenance = 0
tech_level = 4
type = "Facility"
flags = ["A"]
counter = "1d6"
description = "A fund that backs adventuring expeditions. As an action, roll 1d8: on a 1, the asset is destroyed; on a 2–3, gain 1 Coin; on a 4–7, gain 2 Coin; on an 8, gain 3 Coin."
attack = { attacker_stat = "Wealth", defender_stat = "Wealth", damage = "2d6" }
ability = {
  effect = "coin_delta",
  die = "1d8",
  outcomes = [
    { from = 1, to = 1, destroy_self = true },
    { from = 2, to = 3, coin = 1 },
    { from = 4, to = 7, coin = 2 },
    { from = 8, to = 8, coin = 3 },
  ],
}

[assets.FranchiseKits]
id = "HAWK-W6-002"
name = "Franchise Kits"
category = "Wealth"
min_rating = 6
hp = 15
cost = 18
maintenance = 0
tech_level = 4
type = "Facility"
flags = ["S"]
counter = ""
description = "Workshops shipped as kits. The faction may treat every place as tech level 4 for buying Wealth assets."
effects = [
  {
    trigger = "tech_level_query",
    effect = "tech_level_floor",
    filters = { owner = "own", scope = "faction", category = ["Wealth"] },
    tech_level_floor = { level = 4 },
  },
]

[assets.AethericFutures]
id = "HAWK-W6-003"
name = "Aetheric Futures"
category = "Wealth"
min_rating = 6
hp = 10
cost = 20
maintenance = 0
tech_level = 0
type = "Operative"
flags = ["A"]
counter = "1d8"
description = "Traders who buy Aetherstones ahead at a low price. As an action, roll 1d8; that many Coin come off the cost of the faction's next asset, down to half its normal price, rounded down."
attack = { attacker_stat = "Wealth", defender_stat = "Wealth", damage = "2d8" }
ability = { effect = "purchase_discount", die = "1d8" }

[assets.AetherTechFactory]
id = "HAWK-W7-001"
name = "Aether-Tech Factory"
category = "Wealth"
min_rating = 7
hp = 16
cost = 25
maintenance = 0
tech_level = 5
type = "Facility"
flags = ["A"]
counter = ""
description = "A plant the size of a district. As an action, roll 1d8 and gain half that many Coin, rounded up."
ability = {
  effect = "coin_delta",
  die = "1d8",
  outcomes = [
    { from = 1, to = 2, coin = 1 },
    { from = 3, to = 4, coin = 2 },
    { from = 5, to = 6, coin = 3 },
    { from = 7, to = 8, coin = 4 },
  ],
}

[assets.PredatoryPurchase]
id = "HAWK-W7-002"
name = "Predatory Purchase"
category = "Wealth"
min_rating = 7
hp = 10
cost = 20
maintenance = 0
tech_level = 4
type = "Stratagem"
flags = ["S"]
counter = "2d8"
description = "A raid on a rival business, paid for with the rival's own debt. If its attack would destroy a rival asset, the asset is instead reduced to 1 HP and becomes this faction's."
attack = { attacker_stat = "Wealth", defender_stat = "Wealth", damage = "2d10" }
effects = [
  { trigger = "lethal_damage", effect = "seize_target", filters = { owner = "rival", scope = "self" } },
]

[assets.AetherGateTransitAuthority]
id = "HAWK-W7-003"
name = "Aether-Gate Transit Authority"
category = "Wealth"
min_rating = 7
hp = 5
cost = 15
maintenance = 0
tech_level = 5
type = "Facility"
flags = ["S"]
counter = ""
description = "A public network of Aether-gates. Every faction may move any number of its non-Aethership Cunning and Wealth assets between places within 12 hexes of the Authority, with no action. The owner rides free; any other faction pays the owner 1 Coin per asset moved."
attack = { attacker_stat = "Cunning", defender_stat = "Cunning", damage = "1d12" }
effects = [
  {
    trigger = "movement_options_query",
    effect = "grant_movement",
    filters = { owner = "own", scope = "faction" },
    grant_movement = {
      public = true,
      relocate = {
        scope = "other",
        range = "hexes",
        max_hex = 12,
        coin_cost = 1,
        coin_cost_mode = "per_asset",
        categories = ["Cunning", "Wealth"],
        exclude_types = ["Aethership"],
      },
    },
  },
]

[assets.DebtorsFleet]
id = "HAWK-W8-001"
name = "Debtors' Fleet"
category = "Wealth"
min_rating = 8
hp = 20
cost = 30
maintenance = 1
tech_level = 5
speed = 19
drift_rating = 2
type = "Aethership"
flags = ["S"]
counter = "2d10"
description = "An armada crewed by people working off their debts. Moves at speed 19 (long-range). Costs 1 Coin per turn in upkeep. Each time it deals damage, attacking or counterattacking, the faction gains 1d4 Coin."
attack = { attacker_stat = "Wealth", defender_stat = "Wealth", damage = "2d10+4" }
effects = [
  { trigger = "own_attack_hit", effect = "coin_gain", filters = { owner = "own", scope = "self" }, coin_gain = { amount = "1d4" } },
  { trigger = "own_counter_hit", effect = "coin_gain", filters = { owner = "own", scope = "self" }, coin_gain = { amount = "1d4" } },
]
```

##### Task 2 — `rulebooks/hawk/settings.toml`

Full contents:

```toml
# Setting rules for this rulebook. drift_costs and asset_types are required; every
# other key is optional, and a key left out turns its rule off.

# drift_costs[i] = movement points a Drift jump costs at drift rating (i+1). The
# Drift is dangerous: a jump always costs more than a step. Front-loaded, so an
# unrated ship struggles and the first rating upgrades matter most.
drift_costs = [10, 7, 5, 3, 2]

# Every asset type this rulebook's data may name: an asset's type, a transport's
# cargo, a relocate's filter, a shelter, an effects filter and the floors below.
asset_types = [
  "Security Personnel", "Operative", "Facility", "Aethership", "Stratagem", "Beast", "Front", "Workforce",
]

# The least world population on which an asset of each type may be bought. A
# type left out has no floor. Quote a type with a space: "Security Personnel".
# Ships need 500,000 people or more, SWN's floor: on this map, a crowded or
# teeming place. Labor needs a settled place or bigger.
[purchase_population]
Aethership = 500000
Workforce = 100000
```

##### Task 3 — `rulebooks/hawk/assets/force_assets.toml`

Teleportation Hubs drops `Starship`:

**Find:**
```toml
  relocate = { scope = "other", max_hex = 8, coin_cost = 1, exclude_types = ["Aethership", "Starship"], max_assets = 1 },
```
**Replace with:**
```toml
  relocate = { scope = "other", max_hex = 8, coin_cost = 1, exclude_types = ["Aethership"], max_assets = 1 },
```

The Deep-Insertion Gate drops `Starship`:

**Find:**
```toml
    exclude_types = ["Aethership", "Starship"],
```
**Replace with:**
```toml
    exclude_types = ["Aethership"],
```

Carrier Corps joins after Forged Siege Engines, before the tier 7 rows:

**Find:**
```toml
[assets.DeepInsertionGate]
```
**Replace with:**
```toml
[assets.CarrierCorps]
id = "HAWK-F6-004"
name = "Carrier Corps"
category = "Force"
min_rating = 6
hp = 15
cost = 20
maintenance = 0
tech_level = 4
speed = 19
drift_rating = 3
type = "Aethership"
flags = ["S"]
counter = "1d8"
description = "Long-range troopships for a whole army. Moves at speed 19 (long-range). When it moves, it can carry up to ten Force assets from its place with it, for 1 Coin."
transport = {
  coin_cost = 1,
  cargo_types = ["Facility", "Security Personnel", "Operative", "Stratagem", "Beast"],
  max_cargo = 10,
  exclude_categories = ["Cunning", "Wealth"],
}

[assets.DeepInsertionGate]
```

##### Task 4 — `campaigns/hawk/state/state.seed.toml`

The seed's 14 Wealth units become H.A.W.K's by suffix, and two extras join (Decision 18).

| Faction | Stand-in | Heir |
|---------|----------|------|
| O.W.L. | `SWN-W3-001` Postech Industry | `HAWK-W3-001` Aether Refinery |
| O.W.L. | `SWN-W4-001` Shipping Combine | `HAWK-W4-001` Merchant Convoy |
| O.W.L. | `SWN-W6-002` R&D Department | `HAWK-W6-002` Franchise Kits |
| O.W.L. | `SWN-W7-001` Pretech Manufactory | `HAWK-W7-001` Aether-Tech Factory |
| Court of Lies | `SWN-W2-002` Lawyers | `HAWK-W2-002` Contract Cabal |
| Radiant Throne | `SWN-W4-002` Monopoly | `HAWK-W4-002` Company Town |
| Radiant Throne | `SWN-W4-003` Medical Center | `HAWK-W4-003` Mending Vats |
| Fey Royals | `SWN-W2-002` Lawyers | `HAWK-W2-002` Contract Cabal |
| Fey Royals | `SWN-W7-003` Transit Web | `HAWK-W7-003` Aether-Gate Transit Authority |
| Blackwater | `SWN-W2-001` Freighter Contract | `HAWK-W2-001` Free Traders |
| Thymian Armada | `SWN-W1-002` Harvesters | `HAWK-W1-002` Temp Workers |
| Thymian Armada | `SWN-W2-004` Surveyors | `HAWK-W2-004` Prospectors |
| Thymian Armada | — | `HAWK-W1-004` Company Coin (new) |
| Seraph Industries | `SWN-W3-001` Postech Industry | `HAWK-W3-001` Aether Refinery |
| Seraph Industries | `SWN-W6-001` Venture Capital | `HAWK-W6-001` Adventure Capital |
| Seraph Industries | — | `HAWK-W8-001` Debtors' Fleet (new) |

With the Edit tool and `replace_all`, replace every `SWN-W` with `HAWK-W`: the 14 units' table header, `id` and `definition_id` lines (42 lines).

Then the Thymian Armada's Company Coin, after its Prospectors:

**Find:**
```toml
    location = { world_id = "last-light", region_hex = { RegionID = "Q", Coord = { Q = 3, R = 23 } } }

[factions.witherblight-collective]
```
**Replace with:**
```toml
    location = { world_id = "last-light", region_hex = { RegionID = "Q", Coord = { Q = 3, R = 23 } } }
  [factions.thymian-armada.assets.thymian-armada-HAWK-W1-004-1]
    id = "thymian-armada-HAWK-W1-004-1"
    definition_id = "HAWK-W1-004"
    owner_id = "thymian-armada"
    current_hp = 2
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "last-light", region_hex = { RegionID = "Q", Coord = { Q = 3, R = 23 } } }

[factions.witherblight-collective]
```

Then Seraph Industries' Debtors' Fleet, before its Old Bargains:

**Find:**
```toml
  [factions.seraph-industries.assets.seraph-industries-HAWK-C2-001-1]
```
**Replace with:**
```toml
  [factions.seraph-industries.assets.seraph-industries-HAWK-W8-001-1]
    id = "seraph-industries-HAWK-W8-001-1"
    definition_id = "HAWK-W8-001"
    owner_id = "seraph-industries"
    current_hp = 20
    stealthy = false
    ready = true
    maintained = true
    location = { world_id = "veil", region_hex = { RegionID = "H", Coord = { Q = 16, R = 13 } } }
  [factions.seraph-industries.assets.seraph-industries-HAWK-C2-001-1]
```

Then the header:

**Find:**
```toml
# Force and Cunning starting assets are H.A.W.K's own (Efforts 4 and 5); Wealth
# assets are SWN stand-ins until Effort 6 swaps in HAWK- assets.
# Each faction holds its scale's count: hegemon 4 primary + 4 other, major 2 + 2.
# The Radiant Throne also holds one Glamoured Doubles, to show it in play (Effort 5).
```
**Replace with:**
```toml
# Every starting asset is H.A.W.K's own (Efforts 4 to 6).
# Each faction holds its scale's count: hegemon 4 primary + 4 other, major 2 + 2.
# The Radiant Throne also holds one Glamoured Doubles, to show it in play (Effort 5).
# The Thymian Armada also holds a Company Coin beside its two Workforce assets, and
# Seraph Industries a Debtors' Fleet on Veil, to show them in play (Effort 6).
```

Check: `grep -c "SWN-" campaigns/hawk/state/state.seed.toml` prints `0`, and `grep -c "definition_id = \"HAWK-W" campaigns/hawk/state/state.seed.toml` prints `16`.

##### Commit message

```
feat(E6): replace the Wealth stand-in with H.A.W.K's own assets

- wealth_assets.toml: 27 H.A.W.K Wealth assets with HAWK-W IDs, the
  new Workforce type, long-range Wealth transports, Company Coin's
  count, the public Aether-Gate Transit Authority and the Debtors'
  Fleet's Coin on hits
- force_assets.toml: Carrier Corps, a long-range Force carrier; the
  teleports drop Starship from exclude_types
- settings.toml: H.A.W.K's eight types only; Workforce floor 100,000;
  Starship's floor leaves
- state.seed.toml: the 14 seeded Wealth units swap to their heirs by
  suffix; the Thymian Armada gains a Company Coin, Seraph Industries
  a Debtors' Fleet
```

#### Commit 4 — `docs(E6): write the Wealth canon and play the live turn`

##### Task 1 — `campaigns/hawk/canon/assets.md`

The fiction of the 27 Wealth assets and Carrier Corps, from the signed-off pass 6 drafts.

Carrier Corps joins the Force entries:

**Find:**
```markdown
## Deep-Insertion Gate — `HAWK-F7-001`
```
**Replace with:**
```markdown
## Carrier Corps — `HAWK-F6-004`

Troopships built along a long spine of berths, mess decks and drop bays, with an Aether-Core aft and a quartermaster's office forward. Each carries a whole army across the Core, with its staff, its records and its beasts. The army pays for its own berths, by department.

## Deep-Insertion Gate — `HAWK-F7-001`
```

Append at the end of the file, after Oracle Engine's entry:

```markdown

## Wage Garnishers — `HAWK-W1-001`

Clerks with a ledger and a friendly manner, who set up a desk outside a rival's gates on payday-eve. Any worker short of money may take an advance against their wages, at a modest fee. The repayment is drawn from the rival's payroll office directly, by a standing order the workers signed without reading.

## Temp Workers — `HAWK-W1-002`

A roster of workers on contracts that promise no hours, kept on call by a dispatch office. Each morning the office sends word to the ones it needs. The rest wait by the board, in case there is work tomorrow.

## Permit Office — `HAWK-W1-003`

The office that approves every new building on a place, staffed by clerks who have never approved one quickly. An outside builder's application is lost, delayed, or sent back for one more form. The fee for each new form is modest, and it is not refunded.

## Company Coin — `HAWK-W1-004`

The company mints its own coin, stamped with its own seal, and pays every wage in it. The company's stores accept it at face value. No one else accepts it at all, so every wage comes back.

## Free Traders — `HAWK-W2-001`

Independent freighters, each owned by its captain and mortgaged to someone else, that go wherever a contract pays. They post their routes on the dispatch boards at every dock in the Core and leave when the hold is full. They carry anything with a manifest, and most things without one.

## Contract Cabal — `HAWK-W2-002`

Lawyers trained in the clauses of contract magic, where a term once agreed binds the body as well as the purse. They read a rival's agreements more closely than the rival ever has, and find the clauses that bind its own people. Soldiers are outside their jurisdiction, and they prefer to keep it that way.

## Pinkertons — `HAWK-W2-003`

Private agents in good coats, hired by the day to break a strike, guard a payroll or escort a foreman home. They carry clubs, ward-batons and a letter of authority from the company. The letter is not a legal document, and no one has ever checked.

## Prospectors — `HAWK-W2-004`

Prospectors who survey new Fragments for Aether seams. They test each seam by sending in the cheapest crew first, with a lamp and a pick. If the crew comes back, the claim is filed.

## Graveyard Shift — `HAWK-W2-005`

The crew that keeps the works running from dusk to dawn, at a premium rate and with half the lamps. They are short on rest and long on accidents. The company replaces the shift on the morning after an accident, and the work never stops.

## Aether Refinery — `HAWK-W3-001`

Stills, condensers and cutting floors that boil raw Aether down into Aetherstones, run by crews who work in heavy aprons and light shoes. The stones sell well, and the refinery pays well. Now and then it explodes, and the next crew is hired the same day.

## Artificery — `HAWK-W3-002`

A workshop licensed to build Arcano-Tech, with benches, a forge, a Weave-link and an artificer who holds the licence. It sets up wherever the faction sends it. Within a season, the place is making things it never could before, under licence and for a fee.

## Free Company — `HAWK-W3-003`

A mercenary company with a charter, a banner, a rate card and a loyalty clause. The clause runs exactly as long as the payments do. They are good at their work, and they will say so in writing.

## Paid Agitators — `HAWK-W3-004`

Organizers on the faction's payroll, who take jobs on a rival's floor and start talking about pay. Within a week the floor is out on strike, and the rival's machines stand still. The organizers are paid a bonus for every day the strike holds.

## Merchant Convoy — `HAWK-W4-001`

A line of freighters that sails on a published schedule, escorted, insured and always on time. It moves whole operations across the Core: warehouses, staff, their files and their furniture. Departure is final, and late cargo waits for the next convoy.

## Company Town — `HAWK-W4-002`

A town owned and run by a single company: the houses, the shops, the school and the clinic. Every part of life there is tied to working for it. Outsiders are welcome to do business, and they pay the same rents and fees as everyone else.

## Mending Vats — `HAWK-W4-003`

Alchemical vats in a long, warm room, where injured workers and soldiers are knitted back together overnight. They come out ready for their next shift, mostly the same as before. The treatment is taken out of their wages, in instalments.

## Liability Waivers — `HAWK-W4-004`

Waivers signed in every deal the faction makes, by everyone it deals with, in a hand nobody reads. When a rival costs the faction money, the fine print says the loss doesn't count. The rival's lawyers have read the fine print, and agree.

## Foreclosure Notices — `HAWK-W5-001`

A claim served on a rival's asset for a debt, with the amount, the interest and the penalty for late payment. Until the debt is paid, the asset is sealed and stands idle. The debt is usually real, and the amount is usually correct.

## Archmagus Consultant — `HAWK-W5-002`

A famous archmage on a retainer, with a tower on loan, a staff of apprentices and a standing invoice. Their presence lets a place make things it has no business making. The retainer is never cheap, and it is never late.

## Back-Channel Freight — `HAWK-W5-003`

Freight that moves on arrangements made off the record. Every customs officer, harbour master and patrol captain on the route has been paid through a back channel, in advance and in full. A hold full of soldiers lands without a question, and the paperwork says machine parts.

## Adventure Capital — `HAWK-W6-001`

A fund that backs adventuring expeditions for a share of the find. Its money comes from angels, devils and companies alike, and no one asks which is which. Some expeditions come back, and the fund's prospectus lists their names.

## Franchise Kits — `HAWK-W6-002`

Workshops shipped in crates, with every part numbered and a book of instructions in six languages. The faction can put up a branch on any place, whatever the local craft. The instructions are clear, and the spare screws are always missing.

## Aetheric Futures — `HAWK-W6-003`

Traders who buy next season's Aetherstones at this season's price, and sell them to their own faction at a quiet discount. Their desks face a wall of price boards, and none of them looks away from it. When they are wrong, it is someone else's loss.

## Aether-Tech Factory — `HAWK-W7-001`

A plant the size of a district, turning out Aether-tech goods around the clock. Its workers are born in the company hospital, schooled on the factory floor and buried in the company plot. The plant has never closed, and no one remembers who opened it.

## Predatory Purchase — `HAWK-W7-002`

A raid on a rival business, financed with the rival's own debt and timed for its worst quarter. When a blow would ruin the business, the faction buys it at the bottom instead. The staff keep their jobs, and their new contracts start in the morning.

## Aether-Gate Transit Authority — `HAWK-W7-003`

A public network of Aether-gates in stations across the Core, each with a ticket hall, a timetable and a fare board. Anyone may ride, and every fare goes to the faction that runs it. Delays are announced, and apologised for, and not refunded.

## Debtors' Fleet — `HAWK-W8-001`

An armada of salvaged warships and armed freighters, crewed by people working off their debts, one voyage at a time. Each crew member's balance is posted on the mess deck, and every blow the fleet lands pays some of it down. No one has yet reached zero, but the board shows progress.
```

##### Task 2 — the live turn

The overview's Live turn, steps 1, 2 and 4. The map is unchanged, so step 3 is skipped. Claude runs:

```sh
go build -o gm-toolkit ./cmd/gm-toolkit
GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit campaign add-rules --rules rulebooks/hawk --campaign hawk --replace
./cmd/scripts/reset-campaign.sh hawk
```

Then Robert runs `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit faction --campaign hawk`, plays a turn or two, and reports what he saw (Decision 19). Claude tests on `hawk-claude` if Robert asks. Worth a look:

- the 16 seeded Wealth units under their H.A.W.K names
- the Thymian Armada's Company Coin paying 2 Coin at its turn start, for its Temp Workers and Prospectors at Last Light
- a faction with Cunning or Wealth assets within 12 hexes of Veil offered a move "via Aether-Gate Transit Authority (Fey Royals) · 1 Coin per asset"; a ride paying the Fey Royals; the Fey Royals' own offer reading "free"
- Seraph Industries' Debtors' Fleet attacking a Fey Royals asset on Veil and gaining 1d4 Coin per hit, or on a counterattack hit
- buying Temp Workers refused on Solar Gate (population) and offered on Lumos
- Carrier Corps in the Force purchase list for a faction rated Force 6 or higher

A fault that needs a code or data change is logged in the Decision Record and fixed in this commit, or raised with Robert if it is bigger than the effort.

##### Task 3 — clear the drafts

Delete `scratch/A.001.5.E6-wealth-drafts-pass1.md` to `scratch/A.001.5.E6-wealth-drafts-pass6.md`; the plan and the canon hold what they settled. They are gitignored, so nothing here touches the commit.

##### Task 4 — this plan and the overview

- This plan's Work Breakdown Overview: `- [ ] **Phase 2 — The Wealth pass**` becomes `- [x] **Phase 2 — The Wealth pass**`.
- `A.001.5-hawk-asset-authoring-overview-plan.md`, Effort Summary:

**Find:**
```markdown
| [ ] | 6 | Wealth pass |
```
**Replace with:**
```markdown
| [x] | 6 | Wealth pass |
```

##### Commit message

```
docs(E6): write the Wealth canon and play the live turn

- canon/assets.md: the fiction of the 27 H.A.W.K Wealth assets and
  Carrier Corps
- campaigns/hawk/rulebook/: rulebooks/hawk composed in again
- the E6 live H.A.W.K turn, played by Robert
- wealth_assets.toml: the Transit Authority's range drops from 12
  hexes to 6; reach counted in Drift lanes is filed as F.055
- E6 plan: Phase 2 ticked; overview plan: Effort 6 ticked
```
