# Sandbox — Manual Test Guide

The `sandbox` campaign is a test bed for the A.001.6 S-flag rule extensions. It has one map, one cast of four factions, and five scenario seeds. Each seed stages the assets of one effort.

<br/>
<br/>

# Setup

## Run a scenario

1. Build the binary: `go build -o gm-toolkit ./cmd/gm-toolkit`
2. Load a scenario: `./cmd/scripts/reset-campaign.sh sandbox <scenario>`
3. Start the TUI: `./gm-toolkit faction --campaign sandbox`
4. To start the same scenario again: `./cmd/scripts/reset-campaign.sh sandbox`

A reset with a scenario name copies `state/scenarios/<scenario>.toml` to `state.seed.toml` and to `state.toml`. A reset without a name restores the current seed.

| Scenario | Effort | Assets under test |
|---|---|---|
| `combat` | 1 | Saboteurs, Treachery, Cracked Comms |
| `interception` | 2 | Boltholes, False Front, Medical Center, Hostile Takeover |
| `abilities` | 3 | Marketers, Seditionists, Monopoly |
| `purchase` | 4 | Stealth, Capital Fleet, Popular Movement, P-flag assets, the government record |
| `movement` | 5 | Lobbyists, Transport Lockdown, Deep Strike Landers, Blockade Runners |

## Read this first

- **You answer every prompt.** The prompt header names the faction that the engine asks ("Asked of …"). It is often not the faction whose turn it is.
- **Dice are random.** The TUI rolls all dice. Each test uses ratings that make the wanted result likely (for example, Cunning 8 against Cunning 1 hits about 97% of the time). If a roll goes the wrong way, reset and try again.
- **Turn order is fixed for cycle 1 only.** Each seed sets the order in `[current_turn]`. From cycle 2, the engine rolls the order.
- **The board does not show conditions.** `disabled`, `no_ability`, `no_attack`, `sheltered` and `locked_out` show in the execution feed (for example, `asset X +disabled`). To see the current set, open `state/state.toml` and find `conditions` on the asset.
- **One action per turn.** Some steps below need more than one cycle.
- **Asset IDs** use the form `<faction>-<definition>-<n>`, for example `veiled-hand-SWN-C7-003-1`.

<br/>
<br/>

# The Lantern Verge

## Worlds

| World | Region | Hex | TL | Homeworld of | Governed by |
|---|---|---|---|---|---|
| Citadel | A Concord Core | (1, 2) | 5 | Iron Concord | Iron Concord |
| Ferrum | A Concord Core | (3, 2) | 4 | — | Iron Concord |
| Crossway | B Crossway Hub | (9, 2) | 4 | — | independent |
| Aurum | C Gilded Reach | (15, 5) | 5 | Gilt Exchange | independent |
| Farpost | D Far Drift | (16, 10) | 4 | — | independent |
| Gloam | E Umbral Shoals | (-1, 8) | 4 | Veiled Hand | independent |
| Mireholt | F Mire Basin | (7, 10) | 4 | Ashfall Collective | Ashfall Collective |

This gives every permission case: own-governed homeworld (Citadel, Mireholt), own-governed other world (Ferrum), rival-governed (Ferrum, Citadel, Mireholt), independent rival homeworld (Gloam, Aurum) and independent with no homeworld (Crossway, Farpost).

## Routes

Seven regions. Region G (Ironway) is a hex-only corridor from A to C with no world on it. Five warps join the regions: A–B, A–E, B–C, C–D and F–G. No SWN asset sets `drift_rating`, so each warp costs 5 for path choice. A warp is one step when an order ticks, so a warp route arrives sooner than its cost suggests.

| Length | Route | Steps | Warps |
|---|---|---|---|
| Short | Citadel – Ferrum | 2 | 0 |
| Short | Aurum – Farpost | 3 | 1 |
| Short | Crossway – Ferrum | 3 | 1 |
| Short | Citadel – Gloam | 4 | 1 |
| Medium | Crossway – Aurum, Citadel – Crossway | 5 | 1 |
| Medium | Ferrum – Gloam | 7 | 0 |
| Medium | Aurum – Ferrum, Crossway – Farpost | 8 | 2 |
| Long | Aurum – Mireholt | 11 | 1 |
| Long | Citadel – Mireholt | 14 | 1 |
| Long | Aurum – Gloam (through the Ironway) | 16 | 0 |
| Long | Citadel – Farpost | 20 | 1 |

An order arrives after `ceil(steps / speed)` of the owner's turns. Speed-1 carriers (Heavy Drop Assets, Mercenaries) take one turn per step. Speed-3 carriers (Deep Strike Landers, Blockade Runners) take three steps per turn.

To change the map, edit `spatial/source/layout.txt` and `data.toml`, then run `./gm-toolkit spatial build --campaign sandbox --replace`. A warp mark (`*`) links to every marked hex of another region that it can see across empty cells, so put region hexes behind a mark to aim it. If you move a world, update its hex in all five seeds.

<br/>
<br/>

# The Cast

| Faction | ID | Scale | F / C / W | Role |
|---|---|---|---|---|
| Iron Concord | `iron-concord` | Hegemon | 8 / 5 / 7 | Planetary Government of Citadel and Ferrum; the Force side |
| Veiled Hand | `veiled-hand` | Hegemon | 5 / 8 / 7 | The Cunning side |
| Gilt Exchange | `gilt-exchange` | Hegemon | 5 / 7 / 8 | The Wealth side |
| Ashfall Collective | `ashfall` | Minor | 4 / 1 / 3 | Planetary Government of Mireholt; the target (Cunning 1) |

Every seed uses this cast, these homeworlds and these governments. Only the assets, Coin and goals change.

<br/>
<br/>

# Scenario: combat

Order: Veiled Hand, Ashfall, Gilt Exchange, Iron Concord. Veiled Hand has Planetary Seizure (G-004) at phase 1 on Farpost, so its only action is Attack.

## Board

- **Mireholt** — Veiled Hand: Saboteurs, Treachery, Cracked Comms. Ashfall: Harvesters (its only A-flag asset), Militia Unit, Book of Secrets, Party Machine. Ashfall's homeworld Base is here.
- **Farpost** — Veiled Hand: Treachery, Informers. Ashfall: Security Personnel (its only asset there).

## Steps

1. **Cycle 1, Veiled Hand: Attack.** Commit Saboteurs, both Treacheries.
   - Saboteurs → Harvesters. On a hit, a prompt asks whether to redirect the damage to Ashfall's Base on Mireholt. Pick **Redirect to base** so Harvesters lives.
   - Treachery (Mireholt) → Book of Secrets.
   - Treachery (Farpost) → Security Personnel.
2. **Cycle 1, Ashfall.** Look at the action list, then Attack: Party Machine → Cracked Comms.
3. **Cycle 2, at Veiled Hand's turn start.** Read the execution feed.

## Expected

- **Saboteurs** — `asset ashfall-SWN-W1-002-1 +no_ability` on a hit **or** a miss. On Ashfall's turn, **Use Asset Ability is not in the action list**. At Veiled Hand's next turn start: `-no_ability` (`condition_expired`). If the rolled order puts Ashfall before Veiled Hand in cycle 2, the block holds through that Ashfall turn too.
- **Treachery** — the label reads `Treachery · Cunning · special`, with no redirect prompt. On a hit: `hit (no damage)`, then the Treachery is removed, Veiled Hand gains 5 Coin and the target is `captured by veiled-hand`. The captured asset keeps its ID (`ashfall-SWN-C7-002-1`) in Veiled Hand's list. On a miss, the counter hits the Treachery.
- **Hook re-registration** — after the capture, the Book of Secrets reroll on Mireholt is offered to Veiled Hand, not to Ashfall. Ashfall's Party Machine attack in step 2 is a roll on that world.
- **Planetary Seizure** — the Farpost capture takes Ashfall's last asset there, so the goal moves to phase 2 with 3 turns in the same turn (`goal_phase_advanced`).
- **Cracked Comms** — if Veiled Hand defends outright, a prompt shows, asked of Veiled Hand: "Turn Party Machine's attack back on itself?". **Turn it**: Party Machine rolls Cunning against its own Cunning. A self-hit deals its 2d6 and a self-miss deals its 1d6 counter, both as `HP -N (counterstrike)`. **Let it pass**: nothing happens. On a tie there is no prompt.

## Edge cases

- Attack with Treachery against Cracked Comms: no prompt (Treachery has no damage and no counter).
- Treachery against the Militia Unit and a tie: the capture lands even if the counter kills the Treachery.

<br/>
<br/>

# Scenario: interception

Order: Iron Concord, Gilt Exchange, Veiled Hand, Ashfall. Iron Concord (Force 8) holds one Militia Unit on each Veiled Hand world. Veiled Hand defends with Force 5.

## Board

| World | Veiled Hand | Notes |
|---|---|---|
| Crossway | Boltholes, Security Personnel (1 HP), Zealots (1 HP) | |
| Ferrum | False Front, Hitmen (1 HP) | |
| Farpost | Medical Center, Militia Unit (1 HP) | 10 Coin covers the restore (half price 2) |
| Gloam | False Front, Medical Center, Militia Unit (1 HP) | Veiled Hand's homeworld Base: the redirect prompt shows first |
| Aurum | — | Gilt Exchange Hostile Takeover; Ashfall Franchise (3 HP) |

## Steps

1. **Cycle 1, Iron Concord: Attack.** Commit all four Militia Units. Defenders: Security Personnel (Crossway), Hitmen (Ferrum), Militia Unit (Farpost), Militia Unit (Gloam).
   - Gloam: at the redirect prompt, pick **Hit the asset**. At the False Front prompt, pick **Let it fall**. The Medical Center prompt then shows: pick **Pay**.
2. **Cycle 1, Gilt Exchange: Attack.** Hostile Takeover → Franchise.
3. **Cycle 1, Veiled Hand: Repair Asset.** Repair the sheltered Security Personnel.

## Expected

- **Boltholes (Crossway)** — no prompt. `HP +N (sheltered)` and `+sheltered`. The asset stays at 0 HP and drops out of the defender list. After the repair to full HP: `-sheltered` (`repair`).
- **False Front (Ferrum)** — "Sacrifice False Front to save Hitmen?", asked of Veiled Hand. **Sacrifice**: Hitmen back to 1 HP, `-asset … (false front)`.
- **Medical Center (Farpost)** — "Pay 2 Coin to restore Militia Unit at 1 HP?". **Pay**: `HP +N (restored)` and `Coin -2`.
- **Ask order (Gloam)** — the False Front is asked first, because `…-SWN-C1-003-2` sorts before `…-SWN-W4-003-2`. A decline moves on to the Medical Center.
- **Hostile Takeover (Aurum)** — no prompt. `HP +N (seized)` and `captured by gilt-exchange`. The Franchise is now Gilt Exchange's, at 1 HP.
- **Misses** — the defender still counters. An intercepted defender still counters on a tie.

## Edge cases

- Iron Concord attacks the same Farpost Militia again before Veiled Hand's next turn: the Medical Center's budget is spent, so the asset is destroyed with no prompt. The budget clears at Veiled Hand's bookkeeping.
- Veiled Hand at less than 2 Coin: the restore prompt does not show.
- Zealots (Crossway) take damage from their own hit or counter-hit (1d4). At 1 HP that is lethal, and only Boltholes can save them: False Front and Medical Center are not asked. This is chance-based, because Zealots (Force 5) rarely beat Iron Concord (Force 8).
- Destroy the Boltholes with an attack: the sheltered asset is removed with it.
- Medical Center repair discount: a Special Forces or Military Unit repair on its world costs 1 less (floor 0). Known gap `B.017`: the overlay shows the full price, and Repair Asset still needs 1 Coin.

<br/>
<br/>

# Scenario: abilities

Order: Gilt Exchange, Veiled Hand, Ashfall, Iron Concord. Everything is on Crossway (independent, no Bases, so no redirect prompts).

## Board

- **Gilt Exchange** — Marketers, Monopoly; 10 Coin.
- **Veiled Hand** — Seditionists; 10 Coin.
- **Ashfall** — Union Toughs (half price 2), Strike Fleet (half price 6); 4 Coin.
- **Iron Concord** — Security Personnel; 0 Coin.

## Steps

1. **Cycle 1, Gilt Exchange: Use Asset Ability.** Select Marketers and Monopoly.
   - Marketers → Strike Fleet (Cunning 7 against Wealth 3).
   - Monopoly → Iron Concord.
2. **Cycle 1, Veiled Hand: Use Asset Ability.** Seditionists → Union Toughs.
3. **Cycle 1, Ashfall.** Read the bookkeeping feed, then look at the action list.
4. **Cycle 2, Gilt Exchange: Use Asset Ability.** Marketers → Union Toughs; Monopoly → Ashfall.
5. **Cycle 2, Veiled Hand: Use Asset Ability.** Seditionists → the next target.

## Expected

- **Marketers, cannot pay** — Ashfall has 4 Coin, less than 6, so there is no prompt: `asset ashfall-SWN-F4-003-1 +disabled`.
- **Monopoly, 0 Coin** — "Lose an Asset: Iron Concord cannot pay: lose which asset?". The pick removes Security Personnel (`levy_coin`). Esc does nothing here.
- **Seditionists** — 1d4 Coin from Veiled Hand, then `+no_attack` on the Union Toughs. No contest.
- **Ashfall turn start** — no lift offer. The offer runs first in bookkeeping, before income, and Ashfall has 4 Coin, less than the 6 it needs. To see it, set Ashfall's `coin` to 6 or more in `state.toml` before its turn: "Pay 6 Coin to bring Strike Fleet back into use?". **Leave it disabled** means the offer comes back each turn.
- **Ashfall's action list** — **Attack is not offered.** The Toughs are blocked and the Strike Fleet is disabled.
- **Marketers, can pay** — "Pay 2 Coin, or Union Toughs is disabled?", asked of Ashfall. **Pay**: `Coin -2`, no condition.
- **Monopoly, Coin ≥ 1** — Ashfall −1, Gilt Exchange +1, with no prompt.
- **Seditionists moves** — the Union Toughs are not offered again. Picking the Strike Fleet clears the Toughs' `no_attack`, and their attack comes back.

## Edge cases

- A disabled asset is not a Marketers candidate. When it is the only rival asset on the world, Marketers is not offered at all.
- A disabled asset can still be picked as a defender. It rolls defense but deals no counter.
- A blocked asset still counters when attacked.
- Known gap `R.024`: destroy the Seditionists and the `no_attack` record stays in `state.toml`, but the target can attack.
- Known gap `F.029`: a paid lift reads only as Coin spent; the digest has no line for it.

<br/>
<br/>

# Scenario: purchase

Order: Ashfall, Iron Concord, Veiled Hand, Gilt Exchange. Veiled Hand has Inside Enemy Territory (G-009). Gilt Exchange has Planetary Seizure (G-004) at phase 2 with 1 turn left on Mireholt.

## Board

- **Ferrum** (Iron Concord) — Ashfall: 2x Pretech Logistics. Veiled Hand: Lobbyists, Informers, Hardened Personnel.
- **Citadel** (Iron Concord) — Veiled Hand: Popular Movement. Iron Concord: Lobbyists.
- **Crossway** (independent) — Iron Concord: Security Personnel. Veiled Hand: Informers.
- **Aurum** (independent, Gilt Exchange homeworld) — Iron Concord: Security Personnel. Veiled Hand: Informers.
- **Mireholt** (Ashfall) — Gilt Exchange: Franchise.

Buy Asset lists only the homeworld and the worlds where the faction has an asset. A Base does not count.

## Steps and expected

1. **Cycle 1, Ashfall: Use Asset Ability.** Select both Pretech Logistics. Each buys a Militia Unit on Ferrum (150% cost: 6 Coin).
   - First buy: "Permit Ashfall Collective to raise Militia Unit on Ferrum?", asked of Iron Concord, the government of Ferrum. **Grant**.
   - Then "Contest Ashfall Collective's permission…?", asked of Veiled Hand, for its Lobbyists. **Contest** (Cunning 8 against 1). A win: `permission_refused` (`withdrawn`).
   - Second buy: refused with no prompt, as `barred`.
2. **Cycle 1, Iron Concord: Buy Asset.** Capital Fleet on Citadel: no prompt (own homeworld, and Iron Concord governs it).
3. **Cycle 1, Veiled Hand: Buy Asset.** Militia Unit on Citadel: no prompt, because the Popular Movement grants it. Then the contest offer shows, asked of Iron Concord for its Lobbyists. Cunning 5 against 8, so the grant usually stands.
4. **Cycle 1, Gilt Exchange turn start.** The seizure completes with no input: `world_government_changed` (Mireholt to Gilt Exchange), `tag_added` T-011 for Gilt Exchange, `tag_removed` T-011 for Ashfall (Mireholt was its only world), `goal_completed`, `xp_awarded`.
5. **Later cycles** (one action per turn):
   - Iron Concord: Capital Fleet on Crossway, then on Aurum. Each asks the GM: "Asked of the GM (no faction governs …)".
   - Veiled Hand: Stealth on Ferrum. Three own Special Forces there, so a select prompt shows. Inside Enemy Territory +1, because a rival governs Ferrum.
   - Veiled Hand: Stealth on Aurum. One target, so no prompt. +1, because Aurum is a rival's homeworld.
   - Veiled Hand: Stealth on Crossway. One target, no prompt, and no progress (independent, no homeworld).
   - Veiled Hand: Refit Hardened Personnel to Postech Infantry on Ferrum. This asks Iron Concord (cause `refit`).
   - Veiled Hand: Militia Unit on Gloam. No prompt (own homeworld).
   - Gilt Exchange: Mercenaries on Mireholt. No prompt, because Gilt Exchange now governs it.
6. **Refuse any ask.** The action is used up, no Coin is spent and no asset is added. The feed shows `permission refused: … on …`.

## Faction wizard (T-011)

In Manage mode, create a faction and pick the Planetary Government tag. A "Governed world" select shows. It lists the worlds that no faction governs, with the homeworld first. Delete the faction afterwards. Known gap `B.020`: when every world is governed, the select has no options.

<br/>
<br/>

# Scenario: movement

Order: Ashfall, Iron Concord, Veiled Hand, Gilt Exchange. A movement phase runs at the start of every faction turn, before the action: tick, arrival check, new orders, then granted moves (Transit Web).

## Board

- **Crossway** — Ashfall: Heavy Drop Assets, Blockade Runners, Deep Strike Landers, three Militia Units (cargo), Mercenaries, Transit Web; 15 Coin.
- **Ferrum** (Iron Concord) — Veiled Hand: Lobbyists.
- **Citadel** (Iron Concord) — Veiled Hand: Transport Lockdown, seeded with `locked_out` against Ashfall on Citadel. Ashfall: Informers. Iron Concord: Heavy Drop Assets, Militia Unit.
- **In flight** — Ashfall: Mercenaries, one step short of Aurum.

## Steps and expected

1. **Cycle 1, Ashfall movement — tick.** The in-flight Mercenaries arrive at Aurum (independent): "Permit Ashfall Collective to land Mercenaries on Aurum?", asked of the GM.
   - **Refuse**: `move held short of aurum (permission_refused)` and `permission refused`. The next tick asks again.
   - **Grant**: it lands.
2. **Cycle 1, Ashfall movement — new orders.** Each carrier loads one Militia Unit and costs 2 Coin. Send them all to Citadel (5 steps):
   - Deep Strike Landers (speed 3) arrive on Ashfall's third turn. They skip the Lockdown and permission: no fee, no prompt.
   - Blockade Runners (speed 3) reach Citadel on the third turn. The Lockdown fee (1d4) is charged with no prompt, and the order holds with `lockdown_paid`. The next tick lands with no permission ask.
   - Heavy Drop Assets (speed 1) take 5 turns. The first arrival pays the fee and holds. The next tick asks Iron Concord for permission. A refusal holds with `LockdownPaid` still set, so the tick after asks for permission only.
3. **Cycle 1, Ashfall movement — granted move.** Transit Web relocates the other Mercenaries to Ferrum (in range: warps count 0). This asks Iron Concord. **Grant**, then the Lobbyists contest shows, asked of Veiled Hand. A win: `permission_refused` (`withdrawn`), the Mercenaries stay, and Transit Web charges nothing.
4. **Cycle 1, Iron Concord movement.** Heavy Drop Assets with Militia to Ferrum (2 steps). Iron Concord governs Ferrum, so there is no ask, but Veiled Hand's Lobbyists can still contest (Cunning 8 against 5).

## Edge cases

- Ashfall at 0–1 Coin when a carrier reaches Citadel: the order holds with `lockdown_unpaid` and nothing is spent. A new 1d4 is rolled at the next tick. Edit `coin` in `state.toml` to set this up.
- Two carriers locked out in one tick: their fees count against one running Coin total.
- A live Lockdown strike: delete the seeded `conditions` line, reset, then on Veiled Hand's turn attack the Ashfall Informers on Citadel with the Transport Lockdown (special attack, no damage).
- Known gap `B.023`: transports do not enforce their `max_hex` range. An order to any world is accepted, so route length changes only the travel time.
- Known gap `B.019`: the digest reads most `movement_order_*` mutations as "unknown mutation".
- Known gap `F.030`: no view shows which factions a Lockdown holds out.

<br/>
<br/>

# Maintaining the Seeds

- Seeds use inline tables to stay short. The engine writes `state.toml` in the long form, and both load the same way.
- Each asset table key must equal its `id`, and `owner_id` must equal the faction key. Each `location` must match the world's region and hex in `spatial/worlds.toml`.
- Every asset needs `ready = true`. A resumed turn does not ready assets.
- A faction that governs a world in `[governments]` must hold tag T-011, and only those factions hold it.
- `min_rating` applies to purchase and refit only, so a seed can give a faction any asset.
