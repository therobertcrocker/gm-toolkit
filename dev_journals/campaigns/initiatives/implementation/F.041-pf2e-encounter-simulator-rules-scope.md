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
| Walk the Timelines (Luna) | Two Strikes at one multiple attack penalty that count as one attack, at two targets, then Luna moves out of reach | The choice of timeline |
| Sin Vulnerability (Avatar) | Abilities carry a `loyalty` tag; using a tagged ability fires the trigger. Stepping between an ally and the avatar is a 1-action tagged move | The position test for stepping between |
| Cursed Metamorphosis (Avatar) | Failure or critical failure: "transformed", the PC can't act and saves at the end of each turn to end it. Success: sickened 1 | The animal's own stat block |
| Too Angry to Die (Kas) | When Kas stops dying and gets up, he Demoralizes as a free action | Prone and the Stand action |
| Quick Alchemy group (Tavish) | A per-fight counter of quick bombs. Debilitating Bomb and Sticky Bomb are bomb options that spend it | Crafting choices, Double Brew |
| Devouring Void (Arabesque) | Basic Fortitude damage on cast and on each Sustain to the creatures caught; it ends for a target that moves far | The check to leave the area |
| Temporal Twin (Arabesque) | The ally spends its reaction for one extra Strike with no multiple attack penalty | Repeating actions other than a Strike |
| Comforting Presence (Arabesque) | When any ally gains frightened or stupefied, its value drops by 1 and Arabesque's rises by 1 | The adjacency test |

Named approximations that are not stand-ins:

- **Wail** — the burst only; the 1-round lingering area is dropped
- **Last Stand** — the heal, the Wail recharge and the extra action; the move to The Realm Between is dropped
- **Chain Lightning** — one target; each phase has one boss, so the chain never jumps

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
| Wail | approximation | E4 | yes |
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
| Translocate | as written | E7 (position) | no |

### Kas

| Item | Model | Last rule in | Script |
|---|---|---|---|
| Protector's Fury Strike | as written | E1 | yes |
| Vicious Swing, with Furious Focus | as written | E4 | yes |
| Brutal Finish | as written | E4 | yes |
| Sudden Charge | as written | E7 (position) | yes |
| Healing Wave | as written | E4 | yes |
| Healing potions | as written | E4 | yes |
| Intimidating Strike | as written | E3 | no |
| Demoralize, with Intimidating Glare and Intimidating Prowess | as written | E3 | no |
| Battle Cry (Demoralize when rolling initiative) | as written | E4 | no |
| Propelling Winds, Umbral Wings, Winged Boots (fly) | as written | E7 (position) | no |
| Ring of the Ram | as written | E4 | no |
| Amulet of Kinship's Strength (Strength of Mind on 3 allies; +2 Will, +3 against fear) | as written | E5 | no |
| Attack of Opportunity, with its disrupt on a critical hit | as written; the move trigger comes with E7 | E4 | no |
| Tactical Reflexes | as written | E4 | no |
| Farabellus Flip | as written; the Step comes with E7 | E4 | no |
| Too Angry to Die | stand-in | E4 | no |
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
| Calculated Splash, Expanded Splash, Greater Field Discovery, Bomber | as written; splash on other creatures comes with E7 | E3 | splash only |
| Relic: Flare Bolt, Relic: Tidal Crash | as written | E4 | yes |
| Relic: Searing Wave (cone 30, recharge 1d4) | as written | E4 | no |
| Quick Alchemy group, Debilitating Bomb, Sticky Bomb | stand-in | E4 | Sticky Bomb |
| Quick Bomber | as written | E4 | no |
| Wrist Mounted Bomb Launcher (range 60) | as written | E7 (position) | no |
| Tenderizer Grenade | as written; a rules call on the amount | E3 | no |
| Inflammation Flask (weakness 3 to acid, fire and slashing) | as written | E3 | no |
| Star Grenade (splash in two lines) | as written; the line shape comes with E7 | E3 | no |
| Elixir of Life, Healing Potion | as written | E4 | yes |
| Mistform Elixir (concealed) | as written | E3 | no |
| Reaper's Spellgun | as written | E3 | no |
| Bewitching Bloom (Cherry Blossom, Bellflower) | as written | E4 | no |
| Jet Propulsion, Wyrm on the Wing (fly) | as written | E7 (position) | no |
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
| Fly | as written | E7 (position) | no |
| Fire Shield | as written | E5 | no |
| Blood Vendetta, Overselling Flourish (reaction spells) | as written | E5 | no |
| Quickened Casting | as written | E5 | no |
| Temporal Twin | stand-in | E5 | no |
| Lash Out (2d6 void and a shove); the shove comes with E7 | as written | E4 | no |
| Return Force (Bands of Force shove) | as written; the shove comes with E7 | E4 | no |
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

Basic actions (Stride, Step, Interact, Seek, Sense Motive, Balance, Take Cover, Ready) are not sheet items; movement comes with Effort 7's position seam.

## 6 — Rules calls

Each is put to Robert in the session that builds its rule.

| Call | Answer | Effort |
|---|---|---|
| Hammer critical specialization on a flier | The Fortitude save against Kas's class DC as written; on a failure the flier lands prone (off-guard, −2 to attacks, Standing costs an action). Fall damage is a named approximation: bands have no height, and Arrest a Fall (DC 15) almost always succeeds | E3 |
| Tenderizer Grenade's amount | 3 for the Greater grade, a house value: no grade states one, in print or in Foundry; 3 follows the bomb pattern of 1 to 4 by grade | E3 |
| Diehard with Inextinguishable | Kas dies at dying 6, as the export's Diehard rule element sets | E3 |
| Destiny Spores' save | Basic | E3 |
| Ancestral Suspicion | Only the controlled condition counts, as written, so it never fires in this fight | E3 |
| Wail's save | Open: whether it is basic | E4 |
| Faithful Strike | Open: the next Strike or damaging spell against the avatar only, as the text says, or every attack while it lasts, as its rule elements do | E4 |
