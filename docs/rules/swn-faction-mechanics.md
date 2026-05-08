# SWN Faction Rules — Mechanics Reference

Sourced from *Stars Without Number Revised Edition*, pp. 212–229.

> **Project note:** The rulebook calls the currency "Coins." This project renames it to **Coin** throughout.

---

## Faction Attributes

Every faction has three attributes rated 1–8:

| Attribute | Governs |
|-----------|---------|
| **Force** | Military assets, direct combat, planetary assault |
| **Cunning** | Espionage, stealth, infiltration, information warfare |
| **Wealth** | Commerce, logistics, manufacturing, financial attacks |

Ratings determine which assets the faction can own (`min_rating` ≤ relevant attribute) and are added to dice rolls for attacks and defenses.

---

## Faction HP

Max HP = **4 + HP value of Force + HP value of Cunning + HP value of Wealth**

HP values per rating:

| Rating | XP Cost to Reach | HP Value |
|--------|-----------------|----------|
| 1 | — | 1 |
| 2 | 2 | 2 |
| 3 | 4 | 4 |
| 4 | 6 | 6 |
| 5 | 9 | 9 |
| 6 | 12 | 12 |
| 7 | 16 | 16 |
| 8 | 20 | 20 |

Typical HP totals: Minor = 15, Major = 29, Hegemon = 49.

When a faction's HP reaches 0, it is destroyed. Damage to a Base of Influence is also dealt directly to faction HP.

---

## Faction Scale (Starting Ratings)

| Scale | Primary | Secondary | Tertiary |
|-------|---------|-----------|----------|
| Minor | 4 | 3 | 1 |
| Major | 6 | 5 | 3 |
| Hegemon | 8 | 7 | 5 |

Secondary is one less than primary. Tertiary is three less than primary. The GM assigns which attribute receives each tier.

---

## Coin (Coins)

Coin represents logistics capability, available resources, and managerial focus — not literal credits.

**Income per turn:**
`floor(Wealth / 2) + floor((Force + Cunning) / 4)`

- Half Wealth rounded down, plus one-quarter of total Force and Cunning rounded down.

> **Source note:** The PDF rule text (p. 213) reads "half their Wealth rating rounded **up**," but the play example (p. 228) and Vothite income (W3 → 2 Coin/turn) both use rounded down. This project follows the example.

**Maintenance:** Paid at the start of each turn. An asset that cannot be maintained is unusable. If unpaid for two consecutive turns, the asset is lost. Factions cannot voluntarily skip maintenance.

**Asset limit:** A faction may own no more assets of a given attribute type than their rating in that attribute (e.g. Force 3 = max 3 Force assets). Each asset over the limit costs 1 extra Coin per turn in maintenance. Bases of Influence do not count toward this limit.

---

## Faction Tags

Tags represent a faction's nature and special aptitudes. A faction generally receives **one tag**, or two at most for particularly versatile organizations. The Planetary Government tag is an exception — it can be acquired and lost multiple times (once per world controlled).

Tags grant bonus dice: when a tag's domain is relevant to an attack or defense, the faction rolls an **additional d10 and keeps the highest result**.

Tags do not normally change without a drastic, organization-shaping event.

---

## Faction Goals

A faction may pursue **one goal at a time**. On completion, the faction earns XP equal to the goal's difficulty and may select a new goal at the start of their next turn.

Abandoning a goal costs the faction that turn's Coin income and they may take no other action that turn.

### Goal List

| Goal | Completion Condition | Difficulty |
|------|---------------------|------------|
| Military Conquest | Destroy Force assets of rivals equal to your Force rating | half assets destroyed |
| Commercial Expansion | Destroy Wealth assets of rivals equal to your Wealth rating | half assets destroyed |
| Intelligence Coup | Destroy Cunning assets of rivals equal to your Cunning rating | half assets destroyed |
| Planetary Seizure | Take control of a planet, becoming its government | half avg of ruling faction's F/C/W (min 1) |
| Expand Influence | Plant a Base of Influence on a new planet | 1, +1 if contested |
| Blood the Enemy | Inflict HP damage on enemy assets/bases equal to total F+C+W | 2 |
| Peaceable Kingdom | Don't take an Attack action for four turns | 1 |
| Destroy the Foe | Destroy a rival faction | 1 + avg of target faction's F/C/W |
| Inside Enemy Territory | Have Stealthed assets on worlds with other governments equal to Cunning score (assets already stealthed when this goal is adopted don't count) | 2 |
| Invincible Valor | Destroy a Force asset with min_rating higher than your Force rating | 2 |
| Wealth of Worlds | Spend Coin equal to 4× Wealth rating on bribes (lost); Wealth must increase before selecting again | 2 |

---

## Faction Advancement (XP)

XP is earned by completing goals. It can be saved or spent at the start of any turn to raise an attribute rating. XP is consumed on spend.

Cost to raise a rating = the **HP value of the new rating** (e.g. Force 2→3 costs 4 XP, Force 3→4 costs 6 XP).

Raising an attribute increases Max HP and unlocks higher-tier assets.

Optionally, if a faction's deeds justify it, the GM may allow spending XP to acquire a new tag.

---

## The Faction Turn

### Turn Order

At the start of each turn, roll a die no smaller than the number of factions. The result determines which faction acts first; then proceed in order down the list, wrapping around.

### Each Faction's Turn (in sequence)

1. **Collect income** per the Coin formula above
2. **Pay maintenance** for any assets that require it
3. **Take one action** (see below)
4. At turn end, GM narrates a brief news item summarizing visible events

A faction with no current goal may pick one. A faction **can take one action type per turn**, but may apply that action to as many assets/worlds as it wishes (e.g. Attack on multiple worlds in one turn).

---

## Faction Actions

### Attack
- Select one or more of your assets; target a rival faction's assets on the same world
- Each attacking asset is matched one-at-a-time against a defending asset chosen by the **defender**
- Each attacking asset can attack only once per turn; a defending asset can defend multiple times
- **Attack roll:** `1d10 + attacker's relevant attribute` (per the asset's Attack line)
- **Defense roll:** `1d10 + defender's relevant attribute` (per the asset's Attack line)
- Tags add +1d10 (keep highest) when relevant to the roll
- **Attacker wins (strictly greater):** defending asset takes listed damage; if HP = 0, it's destroyed
  - Defender may redirect damage to their Base of Influence on that world instead
- **Defender wins (strictly greater):** attacker's asset takes the defender's Counterattack damage (if any)
- **Tie:** both Attack and Counterattack succeed — both sides take damage
- Only **known** assets can be attacked; Stealthed assets cannot be targeted until revealed

### Buy Asset
- Purchase one asset on the homeworld or any planet with a Base of Influence
- Requirements: faction's relevant attribute ≥ `min_rating`; planet tech level ≥ `tech_level`
- Assets with flag `P` require government permission to purchase or import
- Only **one asset per turn** may be purchased
- Newly bought asset cannot attack, defend, or use special abilities until the **start of next turn**

### Change Homeworld
- Relocate homeworld to a planet where the faction already has a Base of Influence
- Takes 1 turn + 1 more per hex of distance between old and new homeworld
- Faction may take **no actions** during the move
- If the destination already has a Base of Influence, the new homeworld's base is set to maximum HP and the **old homeworld's base inherits the destination base's previous HP value** (the bases swap their HP roles)

### Expand Influence
- Purchase a Base of Influence on a planet where the faction has at least one other asset
- Cost: **1 Coin per HP** of the base, up to the faction's maximum HP
- Faction rolls `1d10 + Cunning`; every rival faction on the planet makes the same roll — any rival that ties or beats the roll may immediately make a free Attack against the new Base of Influence
- The new Base of Influence cannot be used until the start of the faction's next turn
- This action can also be used to **add HP** to an existing Base of Influence (pay 1 Coin per HP added, up to faction's max HP). HP can be reduced without refund. The homeworld's base cannot be shrunk.

### Refit Asset
- Convert one asset to any other asset of the **same type**
- Pay the difference in cost if the new asset is more expensive (minimum 0)
- Planet must support the purchase requirements (tech level, government permission) of the new asset
- Refitted asset cannot attack or defend until the start of the **next turn**

### Repair Asset / Faction
- **Asset repair:** 1 Coin heals HP equal to the faction's score in the asset's ruling attribute. Additional amounts can be healed in the same action, but each further amount costs +1 Coin (1 Coin for first, 2 Coin for second, 3 Coin for third, etc.). Any number of assets may be repaired in a single action.
- **Faction repair:** Regain HP equal to the rounded average of the faction's highest and lowest attribute ratings. This cannot be accelerated by spending more Coin.

### Sell Asset
- Remove an asset and gain **half its purchase cost in Coin**, rounded down

### Seize Planet
- Attempt to become the ruling government of a world
- Must destroy all **unstealthed** opposing assets on the planet; if not accomplished in one turn, must continue next turn (no other actions allowed in the meantime). The attempt ends when **either** all resistance is destroyed **or** all of the seizing faction's own assets on that planet are destroyed or have left
- If all resistance is eliminated, the faction must maintain at least one unstealthed asset on the world for **three turns**
- On success: gain the **Planetary Government** tag for that world

### Use Asset Ability
- Trigger the special abilities of any assets with the `A` (Action) flag
- Each form of asset must be fully used before the next type is triggered (order matters for chaining)
- Unless specified otherwise, all asset targets must be in the same stellar system as the acting assets
- Some abilities call for a faction test: roll `1d10 + acting attribute` vs. `1d10 + target attribute`; ties go to the defender. Tags may add +1d10. Tests are **not attacks** — no damage or counterattack triggers.

---

## Assets

### Asset Fields

| Field | Meaning |
|-------|---------|
| `id` | Unique ID, format `{Category}{MinRating}-{Seq}` (e.g. F3-002, C5-001, W4-003) |
| `category` | Force / Cunning / Wealth |
| `min_rating` | Minimum attribute rating to own |
| `hp` | Hit points; destroyed at 0 |
| `cost` | Coin to purchase |
| `tech_level` | Minimum planet tech level to purchase (assets can be transported to lower TL worlds) |
| `type` | Asset type (see below) |
| `flags` | P = government permission required; A = has Action ability; S = has Special feature |
| `counter` | Damage dealt to a failed attacker |
| `attack` | Attack profile: attacker stat, defender stat, damage dice |

### Asset Types

- **Military Unit** — conventional armed forces
- **Special Forces** — elite, covert, or unconventional units
- **Facility** — fixed installations providing support functions
- **Logistics Facility** — supply/transport infrastructure
- **Starship** — space vessels; can move between worlds
- **Tactic** — situational or one-off effects

### Maintenance Costs

Some individual assets have recurring maintenance costs noted in their descriptions (e.g. Mercenaries: 1 Coin/turn, Scavenger Fleet: 2 Coin/turn). First missed payment: asset is unusable. Second consecutive missed payment: asset is lost. Factions cannot voluntarily skip maintenance.

### Stealth

- A Stealthed asset cannot be detected or attacked by rival factions
- If the asset normally requires government permission to land, that permission may be foregone while Stealthed
- Stealth is **lost** if the asset attacks or defends
- The Stealth Cunning asset (purchased separately) applies Stealth to a Special Forces unit on the same planet

### Government Permission (Flag P)

Assets with flag `P` require the permission of the local planetary government to purchase or import. The standing government can physically prevent recruitment or shoot down assets in entry phase. Bribes of 1d4 Coin can occasionally succeed. No government will willingly permit assets powerful enough to overthrow them.

The **Planetary Government** tag grants or denies this permission for that world.

### Bases of Influence

- Required to purchase assets on a planet (exception: a faction can always buy on its homeworld even if its Base there is destroyed, provided it still has a Base elsewhere)
- Only one Base of Influence per world at a time
- Cannot be moved once placed; cannot be purchased with Buy Asset — only with Expand Influence
- Sale value: 0 Coin
- Damage to a Base is also dealt to the **faction's HP** directly; if a Base is brought below 0 HP, the **overflow damage is not** counted against the faction's HP
- Bases do not count toward the faction's maximum asset totals
- A faction's homeworld always has a Base of Influence at **maximum HP**
- Starship-type assets cannot be purchased on worlds with fewer than several hundred thousand inhabitants

---

## Faction Creation Checklist

1. **Choose scale** — Minor, Major, or Hegemon
2. **Assign attributes** — primary, secondary, tertiary per scale
3. **Calculate Max HP** — 4 + HP value of each attribute
4. **Select tags** — generally one tag; two at most
5. **Select starting goal**
6. **Place homeworld Base of Influence** — automatically at max HP
7. **Select starting assets:**
   - Minor: 1 asset in primary attribute + 1 asset in any other attribute
   - Major: 2 assets in primary attribute + 2 assets in other attributes
   - Hegemon: 4 assets in primary attribute + 4 assets in other attributes
   - Assets must meet attribute rating and tech level requirements
   - If the faction controls one or more worlds, add **Planetary Government** tags
8. **Set starting Coin** — no explicit starting amount is specified in the rules; GM discretion (income begins on Turn 1)

---

## Newly-Founded PC Factions

PC factions (typically formed around character level 9) start as:

- Primary attribute 2, others 1
- 8 HP
- One asset in the primary attribute
- One tag appropriate to the faction's nature

---

## Faction Mergers

When two factions merge, assign a compatibility rating 1–9 (6 = average similarity). For each of Force, Cunning, Wealth, and each asset, roll 1d10 vs. compatibility:

- **Roll ≤ compatibility:** use the higher attribute / asset transfers to new faction
- **Roll > compatibility:** use the lower attribute / asset is forcibly sold

Coin totals are combined. All Bases of Influence are inherited. Select a new goal. Assets whose ratings the new faction can't support are kept but cannot be Repaired.

---

## PC Adventures and Factions

PC adventures occur outside the faction turn economy. Consequences (destroyed assets, gained Coin) are applied directly without rolls. The GM should translate turn events into brief news items for the PCs at the end of each turn.

One Coin ≈ 100,000 credits. PCs can donate up to 1d4 × 100,000 credits per turn (each 100,000 = 1 Coin). Excess in one turn is wasted; larger donations require a full turn to process.
