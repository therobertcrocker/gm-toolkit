# Rules scope — what the rules efforts model from the Fall of Luna sheets

> **Status:** settled with Robert on 2026-11-17, in the `F.041.E3` Plan session (overview plan, Decisions 8 to 12).

Source: the five actor exports and two loose items in `campaigns/hawk/encounters/fall-of-luna/source/` (pf2e system 8.6.1), read item by item, and the import report of `gm-toolkit pf2e sim import fall-of-luna`.
Method: three filters, each Robert's call. Step 1 kept what is useful in this fight. Step 2 took each item that is hard to build on its own and picked as written, stand-in or drop. Step 3 took each survivor that needs code of its own and is not in `cmd/scripts/hawk_finale_sim.py`, and picked keep or drop.

This file binds Efforts 3 to 5 (what each rule must cover) and Effort 6 (which abilities need mechanics; every other ability is marked `ignore` and listed in the report, D-8). A later fight's needs are found by its own import report and `check`, not added here.

## 1 — How to read it

- **Model** — `as written` follows the rules text (D-18). `stand-in` is the simpler effect in section 2 (overview Decision 9). `approximation` is as written except for the named part.
- **Last rule in** — the effort whose building block completes the item. The item's own mechanics are written by hand in Effort 6 (overview Decision 13; conversion moved to `A.004.4`), or earlier where section 4 says so. `E7 (position)` now means `A.004.2`.
- **Script** — `yes` when the script models it today, in any form.

## 2 — Stand-ins

Each is built only from shared building blocks, lives in the creature's hand layer with an `approximation` note, and is printed in the run report (overview Decision 9).

| Item | Stand-in | What is lost |
|---|---|---|
| Walk the Timelines (Luna) | Two Strikes at one multiple attack penalty that count as one attack; the second goes at another standing foe, or the same one when it stands alone. No move | The choice of timeline; the move out of reach (`A.004.2`) |
| Sin Vulnerability (Avatar) | An act of loyalty fires it: using an ability tagged `loyalty`, reviving an ally, or passing a save against a boss effect tagged `tests-loyalty`; once per creature per round. The acting PC gets Faithful Strike. No step-between ability | Stepping between an ally and the avatar (`A.004.2`) |
| Cursed Metamorphosis (Avatar) | Failure or critical failure: "transformed", the PC can't act and saves at the end of each turn to end it. Success: sickened 1 | The animal's own stat block |
| Quick Alchemy group (Tavish) | A pool of 6 quick bombs, a dial for `A.004.2`. Each additive (Debilitating Bomb, Sticky Bomb) is its own 2-action ability that spends it | Crafting choices, Double Brew |
| Devouring Void (Arabesque) | Basic Fortitude damage on cast and on each Sustain to the creatures caught; it ends for a target that moves far | The check to leave the area |
| Temporal Twin (Arabesque) | The ally spends its reaction for one extra Strike with no multiple attack penalty | Repeating actions other than a Strike |
| Comforting Presence (Arabesque) | A reaction: when an ally gains frightened or stupefied, that ally's value drops by 1 and Arabesque's rises by 1. One ability per condition | The adjacency test |

Named approximations that are not stand-ins:

- **Wail** — the burst only; the 3-action version and the 1-round lingering area are dropped
- **Last Stand** — the heal, the Wail recharge and the extra action; the move to The Realm Between is dropped
- **Chain Lightning** — one target; each phase has one boss, so the chain never jumps
- **Attack of Opportunity** — every foe counts as in reach; a manipulate action, a ranged Strike or a Stand provokes it, and Stand is the only move action until `A.004.2`
- **Searing Wave** — the cone is a basic Reflex save by the struck foe on a hit or a miss, so it catches foes only; each phase has one boss
- **Healing Wave** — the cone catches Kas and every ally, dying ones included
- **Ring of the Ram** — the damage; the push is dropped
- **Farabellus Flip** — the +2 AC; the Step waits for `A.004.2`
- **Lash Out** — the 2d6 void; the shove waits for `A.004.2`

## 3 — Kept

### Abyssal Luna

| Item | Model | Last rule in | Script |
|---|---|---|---|
| Slash, Stab | as written | E1 | yes |
| Impending Fate (doomed on a critical hit or a critically failed Destiny Spores save) | as written | E4 | yes |
| Abyssal Frenzy | as written | E1 | yes |
| Destiny Spores (cone 40, Will DC 39, slowed, recharge 1d4) | as written | E4 | yes |
| Fey Momentum (a critical Strike recharges Destiny Spores) | as written | E4 | yes |
| Walk the Timelines | stand-in | E4 | yes |
| Challenge Fate (misfortune reaction) | as written | E4 | yes |
| +2 status to all saves against occult effects | as written | E3 | no |
| Ill Omen | as written | E4 | yes |
| True Target (2 per day) | as written | E5 | no |

### Avatar of Lilith

| Item | Model | Last rule in | Script |
|---|---|---|---|
| Hand of Chaos, Slash, Stinger | as written | E1 | yes |
| Terrifying Touch (Will DC 38, or 43 after a critical hit) | as written | E5 | yes |
| Warpwave Strike, with the Warpwave table (section 4) | as written | E5 | no |
| Demonic Frenzy | as written | E1 | yes |
| Vengeful Spite | as written | E4 | yes |
| Wail (Wails of the Damned's degrees, not basic: nothing, full, full and drained 1d4, double and drained 4) | approximation | E4 | yes |
| Fast Healing 10 | as written | E4 | no |
| Protean Anatomy 25 | as written | E3 | flat resist 12 |
| Sin Vulnerability | stand-in | E4 | yes |
| Last Stand | approximation | E4 | yes |
| Resistance 12 to all damage except holy and cold iron; weakness holy 15; immunities | as written | E3 | yes |
| Warp Mind (3 per day) | as written | E5 | partly |
| Disintegrate | as written | E5 | no |
| Dispel Magic | as written | E5 | yes |
| Cursed Metamorphosis | stand-in | E5 | no |
| Confusion | as written | E5 | partly |
| Unfathomable Song | as written | E5 | yes |
| Divine Wrath | as written | E5 | yes |
| Translocate | as written | `A.004.2` | no |

### Kas

| Item | Model | Last rule in | Script |
|---|---|---|---|
| Protector's Fury Strike | as written | E1 | yes |
| Vicious Swing, with Furious Focus | as written | E4 | yes |
| Brutal Finish | as written | E4 | yes |
| Sudden Charge | as written | `A.004.2` | yes |
| Healing Wave | approximation | E4 | yes |
| Healing potions | as written | E4 | yes |
| Intimidating Strike | as written | E3 | no |
| Demoralize, with Intimidating Glare and Intimidating Prowess | as written | E4 | no |
| Battle Cry (Demoralize when rolling initiative) | as written | E4 | no |
| Propelling Winds, Umbral Wings, Winged Boots (fly) | as written | `A.004.2` | no |
| Ring of the Ram | approximation | E4 | no |
| Amulet of Kinship's Strength (Strength of Mind on 3 allies; +2 Will, +3 against fear) | as written | E5 | no |
| Attack of Opportunity, with its disrupt on a critical hit | approximation; the move trigger comes with `A.004.2` | E4 | no |
| Tactical Reflexes | as written | E4 | no |
| Farabellus Flip | approximation; the Step comes with `A.004.2` | E4 | no |
| Too Angry to Die | as written | E4 | no |
| Inextinguishable, Diehard (dying max) | as written | E3 | yes |
| Toughness, Mountain's Stoutness (recovery DC) | as written | E3 | no |
| Bravery, Juggernaut, Gutsy Halfling, Irrepressible | as written | E3 | Bravery |
| Fearsome Brute | as written | E3 | no |
| Battlefield Surveyor (+2 circumstance to initiative) | as written | E3 | no |
| Greater Crushing and Greater Fearsome runes (section 4) | as written | E3 | no |
| Hammer critical specialization | as written; a rules call on prone fliers | E3 | no |
| Hero points (Heroic Recovery) | as written | E3 | yes |

### Tavish

| Item | Model | Last rule in | Script |
|---|---|---|---|
| Bomb Strikes, with splash and persistent damage | as written | E3 | yes |
| Calculated Splash, Expanded Splash, Greater Field Discovery, Bomber | as written; splash on other creatures comes with `A.004.2` | E3 | splash only |
| Relic: Flare Bolt, Relic: Tidal Crash | as written | E4 | yes |
| Relic: Searing Wave (cone 30, recharge 1d4) | approximation | E4 | no |
| Quick Alchemy group, Debilitating Bomb, Sticky Bomb | stand-in | E4 | Sticky Bomb |
| Quick Bomber | as written | E4 | no |
| Wrist Mounted Bomb Launcher (range 60) | as written | `A.004.2` | no |
| Tenderizer Grenade | as written; a rules call on the amount | E3 | no |
| Inflammation Flask (weakness 3 to acid, fire and slashing) | as written | E3 | no |
| Star Grenade (splash in two lines) | as written; the line shape comes with `A.004.2` | E3 | no |
| Elixir of Life, Healing Potion | as written | E4 | yes |
| Mistform Elixir (concealed) | as written | E3 | no |
| Reaper's Spellgun | as written | E3 | no |
| Bewitching Bloom (Cherry Blossom) | as written | E4 | no |
| Bewitching Bloom (Bellflower) | as written | E5 | no |
| Jet Propulsion, Wyrm on the Wing (fly) | as written | `A.004.2` | no |
| Juggernaut | as written | E3 | no |
| Hero points | as written | E3 | yes |

### Arabesque

| Item | Model | Last rule in | Script |
|---|---|---|---|
| Haste | as written | E4 | yes |
| Vampiric Exsanguination, with Unsullied Blood | as written | E5 | yes |
| Fireball, Disintegrate, Missed Cue | as written | E5 | no |
| Chain Lightning | approximation | E5 | no |
| Devouring Void | stand-in | E5 | no |
| Force Barrage, Sudden Bolt, Bursting Bloom | as written | E5 | no |
| Cantrips: Electric Arc, Needle Darts, Telekinetic Projectile, Ignition | as written | E5 | no |
| True Target | as written | E5 | no |
| Dispel Magic | as written | E5 | no |
| Fly | as written | `A.004.2` | no |
| Fire Shield | as written | E5 | no |
| Blood Vendetta, Overselling Flourish (reaction spells) | as written | E5 | no |
| Quickened Casting | as written | E5 | no |
| Temporal Twin | stand-in | E5 | no |
| Lash Out (2d6 void and a shove); the shove comes with `A.004.2` | approximation | E4 | no |
| Return Force (Bands of Force shove) | as written | `A.004.2` | no |
| Comforting Presence | stand-in | E4 | no |
| Calm and Centered, Ancestral Suspicion, Desolation Locket armor (+3 against emotion) | as written | E3 | no |
| Therapeutic Snap Peas | as written | E4 | no |
| Blink Blade Strike | as written | E1 | no |
| Hero points | as written | E3 | yes |

### Everyone

| Item | Model | Last rule in | Script |
|---|---|---|---|
| Swear Fidelity (the loose item, given to every PC) | as written | E4 | yes |

## 4 — Data the exports do not hold

Typed into the creature's hand layer by the effort that builds its rule (overview Decision 11).

| Data | Where | Source |
|---|---|---|
| The Warpwave table: on a failed save roll 1d8; the effect lasts 1d4 rounds unless stated; a new warpwave ends the previous one. 1 clumsy 2 (3 on a critical failure); 2 confused and 4d6 temporary HP; 3 dazzled (permanent on a critical failure); 4 enfeebled 2 (3); 5 immobilized; 6 quickened (Stride, Strike or Step only); 7 slowed 1; 8 stupefied 2 (3) | `avatar-of-lilith/hand.toml` | Monster Core p. 270, Protean family (AoN MonsterFamilies ID 457) |
| Wails of the Damned: area, degrees of success | `avatar-of-lilith/hand.toml` | Compendium spell, read when Effort 4 builds Wail |
| Greater Crushing and Greater Fearsome rune effects | `kas/hand.toml` | Compendium runes, read when Effort 3 builds critical riders |
| Strength of Mind | `kas/hand.toml` | Compendium spell, read when Effort 5 builds it |

## 5 — Dropped

Every dropped ability is marked `ignore` in a hand layer by Effort 6, so D-8 holds and the report lists it.

| Reason | Items |
|---|---|
| Not combat | Kas: Earn Income, Scout, Search, Gain Contact, Practical Research, Scout Location, Learn Name, Gossip, Feather Steps, Grab an Edge. Tavish: Investigate, Treat Wounds, Subsist, Earn Income, Identify Alchemy, Make an Impression, Coerce, Rally, Research, Decipher Writing, Long-Term Rest, Sense Direction, Craft, Identify Magic, Avoid Notice, Search, Scout. Arabesque: Kreighton's Cognitive Crossover, Linguistics Nexus, Identify Magic, Perception Filter, Hypercognition, Message, Detect Magic, Figment, Light, Alarm, Empathic Link, Scouting Eye, Umbral Journey, Retrocognition, Ancestral Memories, Blood Component Substitution. Luna: Prophetic Wings, Retrocognition, Mindlink. Avatar: Hears Heartbeats |
| No target in this fight | Spellsap Grenade, Umbral Extraction (the bosses cast innate spells); Vitality Lash (undead only); Ring of Counterspells (stores Vision of Death, which no boss casts); Diverting Vortex (no boss makes a ranged attack); Lodestone Bomb's metal rider (the bomb stays); Unfettered Halfling, Escape (no boss grabs); Terrified Retreat (the bosses are higher level) |
| The bosses fly | Trip, Shockwave, Scatter Scree, Tangle Vine |
| Too weak at this level | Demon Mask (DC 29); Desolation Locket spells (DC 30: Haunting Hymn, Agonizing Despair, Wave of Despair); Explosive Missive (DC 18); Explosive Ammunition (DC 25); Worm Vial; Elysian Whimsy, Glimmer of Charm (low-rank incapacitation); Clear Mind, Heal (rank 1 and 2 innate); Bloodspray Curse; Fleet Step; Eagle Eye Elixir; Alchemist Goggles (+1 item, the bombs carry +2); Illusory Creature; Backfire Mantle |
| Can't be used | Shield Block (the shield is destroyed); unprepared Grimoire spells (Phantasmal Killer, Vision of Death, Hideous Laughter, Agonizing Despair); ammunition with no launcher (Magnetic Shot, Disintegration Bolt, Life Shot, Ghost Ammunition, Starshot Arrow, Tidal Crash Round, Transposition Ammunition); Fire in the Hole (its feat is not on the sheet); loose Shock and Impactful runes |
| Needs a map (D-11) | Wall of Stone; Arabesque's Translocate and the Blink Blade teleports |
| Hard on its own (step 2) | Grapple (the PCs don't grapple enemies) |
| Own code, not in the script (step 3) | Untethered to Fate, Guidance, Shatter Defenses, Spellwatch rune, Dread rune, Infectious Enthusiasm, Aid, Wand of Contagious Frailty, Reach Spell, Widen Spell, Safeguarded Spell |

Basic actions (Stride, Step, Interact, Seek, Sense Motive, Balance, Take Cover, Ready) are not sheet items; movement comes with `A.004.2`'s position seam.

## 6 — Rules calls

Each is put to Robert in the session that builds its rule.

| Call | Answer | Effort |
|---|---|---|
| Hammer critical specialization on a flier | The Fortitude save against Kas's class DC as written; on a failure the flier lands prone (off-guard, −2 to attacks, Standing costs an action). Fall damage is a named approximation: bands have no height, and Arrest a Fall (DC 15) almost always succeeds | E3 |
| Tenderizer Grenade's amount | 3 for the Greater grade, a house value: no grade states one, in print or in Foundry; 3 follows the bomb pattern of 1 to 4 by grade | E3 |
| Diehard with Inextinguishable | Kas dies at dying 6, as the export's Diehard rule element sets | E3 |
| Destiny Spores' save | Basic | E3 |
| Ancestral Suspicion | Only the controlled condition counts, as written, so it never fires in this fight | E3 |
| Wail's save | Not basic: Wails of the Damned as written (Player Core p. 367) with the Avatar's own numbers, Fortitude DC 38, 8d10 spirit. Nothing on a critical success, full damage on a success, full damage and drained 1d4 on a failure, double damage and drained 4 on a critical failure. The 3-action version and the lingering round stay dropped | E4 |
| Faithful Strike | The text, not the rule elements: spent by the next Strike or damaging spell against the avatar, hit or miss; it ends unused at the end of the holder's next turn. Until Effort 5 defines spells, Strikes only | E4 |
| Recharge | Counts down at the start of the user's turn; a roll of 1 means ready next turn. Destiny Spores, Wail and Searing Wave share it | E4 |
| The Avatar's turn start | Last Stand is checked first, then Fast Healing | E4 |
| Impending Fate and Fey Momentum | Fire on critical Strikes only, not on other attack rolls | E4 |
| Vengeful Spite's conditions | Frightened and stunned follow the save result alone; Terrifying Touch's "not already frightened" limit does not carry over | E4 |
| Sin Vulnerability's acts | Once per creature per initiative round; a `loyalty` ability, a revive of an ally, and a passed save against a boss effect marked for it. Faithful Strike goes to the acting PC. No step-between ability until `A.004.2` | E4 |
| Hands | Not run, a table rule that beats rules as written: drinking or feeding a potion or elixir is 1 action; no draw, re-grip or dropped weapon | E4 |
| A knocked-out PC | Is prone, and Stand costs 1 action. Too Angry to Die is a reaction on that Stand, as written | E4 |
| Starting counts | From the sheets: Kas 3 Moderate and 1 Greater healing potion; Tavish 1 Elixir of Life and 1 potion; Arabesque 5 Therapeutic Snap Peas; Tavish 6 quick bombs. Each a dial for `A.004.2`. Sheet stock does not set a pool: Tavish's later sheet shows 3 infused reagents, mid-adventure state, and the pool stays 6 | E4 |
| Relic gifts and Searing Wave | One gift per bomb. Searing Wave's cone fires on a hit or a miss, not a critical failure, and catches foes only | E4 |
| Attack of Opportunity | Fires on every manipulate action, ranged Strike and Stand by a foe; every creature counts as engaged with every foe. A critical hit disrupts a manipulate action: the actions and the use are both lost | E4 |
| Position stand-ins | Walk the Timelines is two Strikes and no move; Farabellus Flip is the +2 AC without the Step; Lash Out is the damage without the shove; Return Force waits for `A.004.2`; Healing Wave catches Kas and every ally, dying ones included; Ring of the Ram's push is dropped | E4 |
| Challenge Fate | Answers Strikes and spell attacks | E4 |
| Comforting Presence | A reaction that helps the triggering ally only | E4 |
| Skill modifiers | Computed from the build: the importer reads skill ranks, the load computes each modifier | E4 |
| Bewitching Bloom (Cherry Blossom) | A PC rolls recovery against 10 or its own DC, whichever is lower | E4 |
| "Once per day, hour or minute" | Once per fight. Slowed "for 1 round" is a 1-round duration | E4 |
| Quickened with lost actions | A quickened creature that loses actions loses the quickened one first | E4 |
| Actions while down | A downed PC still regains actions (Player Core pg. 415, 435, 411), so stunned and slowed count down while it is down | E4 |
| Stunned on the bearer's own turn | Takes its remaining actions at once, as written (Player Core pg. 442); the ability in progress completes | E4 |
| Stunned for a duration | Loses every action for the duration and overrides slowed, as written (Player Core pg. 446). Inference, not rules text: a valueless stunned beside a valued one loses every action, and the valued one counts down by what it would have taken | E4 |
| Flourish and once per round | One clock: back at the start of the user's next turn, as a 1-round duration is. The strict top-of-initiative round is not built | E4 |
| Reactions before the first turn | Allowed: a table rule, as Player Core pg. 436 leaves it to the GM. Tactical Reflexes' extra reaction arrives at the first turn, as its text says | E4 |
| Fast healing on a downed creature | Works, as written (Player Core pg. 410: no exception for 0 HP or dying). Two inferences, not rules text: the heal comes before the recovery check, since the start-of-turn steps run in any order and healing first is never worse; a creature healed above 0 acts that turn, still prone | E4 |
| Sustain | An effect lasts to the end of the user's next turn even if the user drops, as written. Two readings wait for Effort 5: a second Sustain in a turn, and a disrupted Sustain ending the effect | E4 |
| Haste | Reaches the caster: "Target 1 creature", heightened to 6, nothing keeps the caster out | E4 |
| Demoralize | Visual, through Intimidating Glare (Player Core pg. 257) | E4 |
| Untrained Improvisation | The importer reads the sheet's formula, the legacy text; at level 14 the legacy and Remaster numbers agree (+14) | E4 |
| A bomb with no stock | Imported when it is a bomb, the PC has Quick Alchemy and its formula is in the formula book, as the legacy Quick Alchemy requires (Core Rulebook pg. 72). Any other quantity-0 weapon is skipped | E4 |
| A bare bomb Strike | Spends nothing today; a design question for Effort 6 | E6 |
