# Action Resolution — Discovery & Planning

A breakdown of the resolution logic for each of the 9 faction actions.



## Sell Asset

**Complexity:** Low — pure transaction, no dice, no targeting, no multi-turn state.

### Inputs
- Faction performing the action
- Asset selected for sale

### Resolution Steps
1. Calculate sale value: `floor(asset.cost / 2)`
2. Remove asset from faction
3. Add sale value to faction's Coin balance

### Outputs
- Asset removed from faction state
- Faction Coin balance increased by sale value

### Notes
- Bases of Influence are excluded at the validation step; by resolution time the asset is guaranteed to be sellable

<br/>


## Repair Asset

**Complexity:** Low-Medium — no dice, but escalating Coin cost across multiple assets.

### Inputs
- Faction performing the action
- List of assets to repair, each with a number of heals to apply

### Resolution Steps
1. For each asset being repaired:
   - Determine the relevant attribute score (based on asset category)
   - For each heal applied: restore HP equal to the attribute score, deduct Coin (cost = 1 for first heal, +1 for each subsequent heal on that asset)
   - Cap restored HP at asset's max HP
2. Apply all Coin deductions to faction balance

### Outputs
- Selected assets have HP restored
- Faction Coin balance reduced by total repair cost

### Notes
- Escalating cost resets per asset — a second asset starts back at 1 Coin for its first heal
- Validation ensures faction has at least 1 Coin and at least one damaged asset

<br/>

## Repair Faction

**Complexity:** Low — fixed formula, no dice, no targeting.

### Inputs
- Faction performing the action

### Resolution Steps
1. Calculate heal amount: `round((highest attribute + lowest attribute) / 2)`
2. Restore faction HP by that amount, capped at faction max HP
3. No Coin cost

### Outputs
- Faction HP restored by calculated amount

### Notes
- Cannot be accelerated by spending more Coin
- Validation ensures faction HP is below max before offering this action

<br/>


## Buy Asset

**Complexity:** Medium — no dice, but multi-step filtering before selection.

### Inputs
- Faction performing the action
- World selected for purchase
- Asset selected for purchase

### Resolution Steps
1. Build filtered asset list: assets where faction's relevant attribute ≥ `min_rating`, world tech level ≥ `tech_level`, and (if flag `P`) faction has government permission on that world
2. Present filtered list to GM
3. Deduct asset cost from faction Coin balance
4. Add asset to faction on the selected world
5. Flag asset as inactive (cannot attack, defend, or use abilities until start of next turn)

### Outputs
- Asset added to faction state on selected world
- Faction Coin balance reduced by asset cost
- Asset flagged as newly purchased (inactive this turn)

### Notes
- Only one asset may be purchased per turn
- Validation ensures at least one purchasable asset exists and faction has sufficient Coin before offering this action

<br/>

## Refit Asset

**Complexity:** Medium — no dice, cost delta calculation, same filtering logic as Buy Asset scoped to asset type.

### Inputs
- Faction performing the action
- Asset selected for refit
- Replacement asset selected

### Resolution Steps
1. Build filtered asset list: assets of the same type as the selected asset, where faction's relevant attribute ≥ `min_rating`, world tech level ≥ `tech_level`, and (if flag `P`) faction has government permission on that world
2. Present filtered list to GM
3. Calculate cost delta: `max(0, new_asset.cost - old_asset.cost)`
4. Deduct cost delta from faction Coin balance
5. Replace old asset with new asset on the same world
6. Flag asset as inactive (cannot attack or defend until start of next turn)

### Outputs
- Old asset removed from faction state
- New asset added in its place on the same world
- Faction Coin balance reduced by cost delta (if any)
- Asset flagged as newly refitted (inactive this turn)

### Notes
- If the new asset is cheaper or equal in cost, no Coin changes hands
- Validation ensures at least one valid refit target exists before offering this action

<br/>

## Expand Influence

**Complexity:** Medium-High — Coin transaction plus contested roll and optional rival interference for new bases.

### Inputs
- Faction performing the action
- World selected
- Mode: new Base of Influence or reinforce existing
- Number of HP to purchase

### Resolution Steps

**New Base of Influence:**
1. GM selects world and HP amount (1 Coin per HP, up to faction max HP)
2. Deduct Coin from faction balance
3. Roll contested check: faction rolls `1d10 + Cunning`; each rival faction on that world rolls the same
4. For each rival that ties or beats the faction's roll: present the GM the option for that rival to make a free Attack against the new Base (rival may decline)
5. Resolve any free Attacks (see Attack resolution)
6. Place new Base of Influence on the world, flagged as inactive until start of next turn

**Reinforce Existing Base:**
1. GM selects world and HP amount to add (1 Coin per HP, up to faction max HP)
2. Deduct Coin from faction balance
3. Increase Base HP by selected amount
4. No contested roll or rival interference

### Outputs
- New Base placed on world (or existing Base HP increased)
- Faction Coin balance reduced by purchase cost
- New Base flagged as inactive this turn (new base only)

### Notes
- **Depends on Attack resolution** — free Attack triggered by rivals uses Attack action logic; design Attack before finalizing this flow
- Rival free Attack is optional — the GM decides whether each eligible rival chooses to attack or pass
- The homeworld Base cannot be shrunk; HP can otherwise be reduced without refund
- Validation ensures faction has at least one asset on the target world and sufficient Coin

<br/>

## Use Asset Ability

**Complexity:** Medium — delegates to a bespoke Ability Engine per asset; ordering constraint adds sequencing logic.

### Inputs
- Faction performing the action
- Ordered list of assets whose abilities to trigger

### Resolution Steps
1. Build list of usable `A`-flagged assets (not unmaintained, not inactive)
2. Present list to GM for selection and ordering
3. For each selected asset in order: fully resolve its ability via the Ability Engine before proceeding to the next
4. If an ability requires a faction test, call the shared faction test utility

### Outputs
- Ability effects applied per asset (varies by ability)
- Any state changes recorded to history

### Notes
- **Depends on Ability Engine (planned)** — each asset ability has bespoke resolution logic; the Ability Engine is a sub-engine responsible for resolving named abilities
- **Faction test** (`1d10 + acting attribute` vs. `1d10 + target attribute`, ties to defender, tags add +1d10) is a shared engine utility — not owned by the Ability Engine, as it may surface in other actions
- Tests are not attacks — no damage or counterattack triggers
- All targets must be in the same stellar system unless the ability specifies otherwise

<br/>

## Attack

**Complexity:** High — dice rolls, tag modifiers, defender choice, damage redirection, multi-asset sequencing.

### Inputs
- Faction performing the action
- One or more attacking assets
- Target world
- Per matchup: defending asset (chosen by GM for the rival faction)

### Resolution Steps
1. GM selects one or more attacking assets on the target world
2. For each attacking asset in sequence:
   a. GM selects the defending asset (chosen by the defending faction's GM)
   b. Check tag relevance for attacker and defender via Tag Engine; apply +1d10 keep highest if relevant
   c. Roll attack: `1d10 + attacker's relevant attribute` (plus tag die if applicable)
   d. Roll defense: `1d10 + defender's relevant attribute` (plus tag die if applicable)
   e. Resolve outcome:
      - **Attacker strictly greater:** defending asset takes listed attack damage; GM may redirect damage to Base of Influence on that world instead; if asset HP = 0 it is destroyed
      - **Defender wins or ties:** attacking asset takes counterattack damage (if any listed on defending asset)
      - **Tie:** both attack damage and counterattack damage apply simultaneously
3. Update HP for all affected assets; remove destroyed assets from state

### Outputs
- Asset HP updated or assets destroyed
- Base of Influence HP updated if damage redirected
- All attack outcomes recorded to history

### Notes
- **Depends on Tag Engine (planned)** — tag relevance per roll is evaluated by the Tag Engine; tags surface across multiple actions not just Attack
- Defender asset selection is manual (GM decides for rival faction); AI agent defender logic is a future concern
- A defending asset can defend multiple times in one turn; an attacking asset can only attack once
- Only known (non-stealthed) assets can be targeted; stealth is lost if an asset attacks or defends
- Stealth loss on defend should be applied before damage resolution so state is consistent
