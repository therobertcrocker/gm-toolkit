# Gate 1 fixture — level-14 sheet numbers derived by hand from the Foundry exports

> **Status:** confirmed. Derived by hand, then checked by Robert against the Foundry sheets on 2026-11-17 with no corrections (`F.041.E2` plan, Decision 1).

Source: `campaigns/hawk/encounters/fall-of-luna/source/actors/{kas,tavish,arabesque}.json` (pf2e system 8.6.1, Foundry 14.365).
Method: build choices and item rule elements read with python/jq, PF2e rules applied by hand. No Foundry instance was run, so every number is a derivation, not a sheet reading.
Baseline state: shields not raised, no conditions, no effect items, toggles off, gear as equipped/invested in the export.

## 1 — Summary table

"Script" is `cmd/scripts/hawk_finale_sim.py` `PARTY[14]`.

### Kas Fringebramble (legacy fighter 14, halfling)

| Number | Derived | Script | Match? |
|---|---|---|---|
| Str / Dex / Con / Int / Wis / Cha | +5 / +0 / +4 / +0 / +3 / +4 | — | — |
| Max HP | 230 | 230 | yes |
| AC | 36 | 35 | NO (+1) |
| Fortitude | +26 | 23 | NO (+3) |
| Reflex | +20 | 21 | NO (-1) |
| Will | +23 | 22 | NO (+1) |
| Perception | +23 | — | — |
| Class DC (fighter) | 33 | 32 | NO (+1) |
| Speed | 30 ft | — | — |
| Protector's Fury attack | +29 | 29 | yes |
| Protector's Fury damage | 3d12+9 bludgeoning (versatile P) | dmg_bonus 9 | yes |
| Spell attack / DC ("Arcane Charges Spells", arcane, Cha, trained, no spells) | +20 / 30 | — | — |
| Resistances | slashing 4 | — | — |

### Tavish Tyne (legacy alchemist 14, bomber, dwarf)

| Number | Derived | Script | Match? |
|---|---|---|---|
| Str / Dex / Con / Int / Wis / Cha | +0 / +4 / +3 / +4 / +2 / +3 | — | — |
| Max HP | 164 | 164 | yes |
| AC | 35 | 36 | NO (-1) |
| Fortitude | +24 | 22 | NO (+2) |
| Reflex | +23 | 21 | NO (+2) |
| Will | +21 | 20 | NO (+1) |
| Perception | +20 | — | — |
| Class DC (alchemist) | 32 | — | — |
| Speed | 30 ft | — | — |
| Bomb attack, greater bombs and Lodestone (item +2) | +24 | 24 | yes |
| Bomb attack, moderate bombs (item +1) | +23 | — | — |
| Bomb damage, example Alchemist's Fire (Greater) | 3d8+2 fire, 3 persistent fire, 7 fire splash | — | — |
| Dagger | +22, 1d4+2 piercing | — | — |
| Sling | +22, 1d6+2 bludgeoning | — | — |
| Spell attack / DC ("Arcane Innate Spells", follows alchemist class statistic, no real spells) | +22 / 32 | — | — |
| Resistances | poison 7, electricity 5 | — | — |

### Arabesque Opalheim (legacy sorcerer 14, imperial bloodline, elf)

| Number | Derived | Script | Match? |
|---|---|---|---|
| Str / Dex / Con / Int / Wis / Cha | +0 / +4 / +1 / +4 / +1 / +5 | — | — |
| Max HP | 104 | 104 | yes |
| AC | 34 | 33 | NO (+1) |
| Fortitude | +21 | 20 | NO (+1) |
| Reflex | +24 | 22 | NO (+2) |
| Will | +21 | 21 | yes |
| Perception | +19 | — | — |
| Spell attack (all entries but the locket) | +23 | — | — |
| Spell DC (all entries but the locket) | 33 | 33 | yes |
| Top spell rank | 7 | 7 | yes |
| Class DC (sorcerer) | 31 (Foundry shows trained; legacy rules give the sorcerer none) | — | — |
| Speed | 30 ft | — | — |
| Blink Blade | +24, 2d4+2 piercing (versatile S) | — | — |
| Resistances / weaknesses / immunities | none | — | — |

## 2 — Derivation trails

Proficiency bonus = level 14 + 2 / 4 / 6 / 8 (trained / expert / master / legendary).

Attributes: all three actors have `system.abilities = null` and store boosts in `system.build.attributes.boosts` (keys "1", "5", "10") plus boosts on the ancestry, background and class items. This is the remaster "modifier" form: each boost is +1, and a boost to an attribute already at +4 is a partial boost (two partials make +1).

### 2.1 — Kas

**Attributes**

| Source | Str | Dex | Con | Int | Wis | Cha |
|---|---|---|---|---|---|---|
| Ancestry Halfling: `system.alternateAncestryBoosts = ["str","cha"]` (replaces the fixed Dex/Wis/free boosts and the Str flaw) | +1 | | | | | +1 |
| Background Warrior (`boosts.0.selected=con`, `boosts.1.selected=str`) | +1 | | +1 | | | |
| Class Fighter `keyAbility.selected=str` | +1 | | | | | |
| Level 1 boosts [cha, con, str, wis] | +1 | | +1 | | +1 | +1 |
| Level 5 boosts [str, con, cha, wis] | partial | | +1 | | +1 | +1 |
| Level 10 boosts [str, con, cha, wis] | partial (completes +5) | | +1 | | +1 | +1 |
| **Total** | **+5** | **0** | **+4** | **0** | **+3** | **+4** |

No apex item. The ancestry item still shows `boosts.2.selected = "cha"` and a Str flaw; both are overridden by `alternateAncestryBoosts`. Cross-check: only the alternate reading gives Str +5, which the script's atk 29 and dmg_bonus 9 need.

**Proficiency ranks used**

| Stat | Rank | Source |
|---|---|---|
| Perception | master | class `perception: 2`; Battlefield Surveyor (L7) ActiveEffectLike upgrade `system.perception.rank` 3 |
| Fortitude | master | class `savingThrows.fortitude: 2`; Juggernaut (L9) `subfeatures.proficiencies.fortitude.rank 3` |
| Reflex | expert | class `savingThrows.reflex: 2` (Evasion is L15, not on the sheet) |
| Will | expert | class `will: 1`; Bravery (L3) ActiveEffectLike upgrade `system.saves.will.rank` 2 |
| Heavy armor | expert | class `defenses.heavy: 1`; Armor Expertise (L11) `subfeatures.proficiencies.heavy.rank 2` |
| Hammer group, simple/martial | legendary | Weapon Legend (Hammer) (L13) MartialProficiency value 4, `selection: hammer` |
| Other simple/martial/unarmed | master | Weapon Legend `subfeatures.proficiencies` rank 3 |
| Class DC | expert | Fighter Expertise (L11) `subfeatures.proficiencies.fighter.rank 2` |

**Numbers**

- **Max HP 230** = ancestry 6 (`Halfling.system.hp`) + 14 x (class 10 + Con 4) = 196 + Toughness FlatModifier hp `@actor.level` 14 + Mountain's Stoutness FlatModifier hp `@actor.level` 14.
- **AC 36** = 10 + Dex 0 (Full Plate `dexCap: 0`, Dex is 0 anyway) + heavy expert 18 + Full Plate `acBonus` 6 + armor potency rune 2 (`runes.potency: 2`, armor worn and invested). Tower shield "Bramblecrest (Destroyed)" is stowed in the backpack: no shield. Farabellus Flip +2 circumstance is a toggle, off.
- **Fortitude +26** = 14 + 6 (master) + Con 4 + item 2 (greater resilient, Full Plate `runes.resilient: 2`).
- **Reflex +20** = 14 + 4 (expert) + Dex 0 + item 2 (resilient). Bulwark trait (+3 instead of Dex vs damaging Reflex effects) is conditional and not counted; with it the number is +23.
- **Will +23** = 14 + 4 (expert) + Wis 3 + item 2. Two item bonuses are present (resilient 2; Amulet of Kinship's Strength FlatModifier will item 2, invested); same type, so they do not stack. The amulet's +1 vs fear is conditional.
- **Perception +23** = 14 + 6 (master) + Wis 3. Conditional, skipped: Keen Eyes +2 circumstance (Seek, hidden/undetected within 30 ft), Umbral Wings +1 item (Seek). Initiative with Perception gets +2 circumstance (Battlefield Surveyor) = +25.
- **Class DC 33** = 10 + 14 + 4 (expert) + Str 5.
- **Speed 30 ft** = Halfling 25 + Fleet 5 (FlatModifier land-speed). Full Plate `speedPenalty: -10`, reduced to -5 because Str +5 meets `strength: 4`, then suppressed by Unburdened Iron (AdjustModifier suppress `armor-speed-penalty`). Winged Boots and Umbral Wings give a fly Speed only when activated.
- **Protector's Fury attack +29** = 14 + 8 (legendary, hammer group) + Str 5 + weapon potency 2 (`runes.potency: 2`). Held in 2 hands.
- **Protector's Fury damage 3d12+9 bludgeoning** = dice 3 (base 1 + greater striking, `runes.striking: 2`) x d12 + Str 5 + Weapon Specialization 4 (FlatModifier 2, AdjustModifier upgrade to 4 at `item:proficiency:rank:4`). Property runes `greaterCrushing`, `greaterFearsome` add no dice; they act on a critical hit only. Critical specialization (hammer: knock prone) is on (Fighter Weapon Mastery CriticalSpecialization, rank >= 3). Versatile P lets the damage be piercing.
  - Toggles, off in the export: Vicious Swing +2 dice at L10+ (5d12+9), Brutal Finish +1 die. Conditional: Fearsome Brute circumstance damage vs frightened targets (2 x frightened value, x1.5 because Intimidation is master: +3 at frightened 1, +6 at frightened 2).
- **Unarmed fist** (not on the sheet, `showBasicUnarmed: false`): +25 (14 + 6 master + Str 5), 1d4+8 bludgeoning (Str 5 + specialization 3).
- **Spellcasting entry "Arcane Charges Spells"**: tradition arcane, `ability: cha`, `proficiency.value: 1`, `prepared: charges`, zero spells. If used: attack +20 = 14 + 2 + Cha 4, DC 30. Not a real caster.
- **Resistances**: slashing 4 = Armor Expertise Resistance rule, predicate heavy + plate, value `2 + @armor.system.runes.potency` (2 + 2). The armor material is dragonhide (`material.type`), which Foundry does not automate. No weaknesses or immunities.

**Skipped (conditional, toggled off or not a sheet number)**: Keen Eyes; Gutsy Halfling, Bravery, Juggernaut, Irrepressible, Inextinguishable AdjustDegreeOfSuccess; Battlefield Surveyor initiative bonus; Farabellus Flip (toggle off); Vicious Swing (toggle off); Brutal Finish (toggle off); Fearsome Brute; Intimidating Prowess; Demon Mask +2 item Intimidation (skill only); Feather Steps; Aethership Navigation; amulet fear rider; Umbral Wings Seek bonus; Aetherlight Lantern.

**Ignored state on the actor**: effect "Spell Effect: Haste", condition "Slowed 1", current HP 230, hero points.

### 2.2 — Tavish

**Attributes**

| Source | Str | Dex | Con | Int | Wis | Cha |
|---|---|---|---|---|---|---|
| Ancestry Dwarf (legacy): Con, Wis, free = cha; flaw Cha (`alternateAncestryBoosts: null`) | | | +1 | | +1 | +1 - 1 |
| Background Saboteur (dex, int) | | +1 | | +1 | | |
| Class Alchemist key ability: `keyAbility.value = ["int"]`, `selected: null` (single option, so Int) | | | | +1 | | |
| Level 1 boosts [int, cha, con, dex] | | +1 | +1 | +1 | | +1 |
| Level 5 boosts [dex, int, cha, wis] | | +1 | | +1 | +1 | +1 |
| Level 10 boosts [dex, cha, int, con] | | +1 | +1 | partial (stays +4) | | +1 |
| **Total** | **0** | **+4** | **+3** | **+4** | **+2** | **+3** |

**Proficiency ranks used**

| Stat | Rank | Source |
|---|---|---|
| Perception | expert | class `perception: 1`; Alertness (L9) `subfeatures.proficiencies.perception.rank 2` |
| Fortitude | master | class 2; Juggernaut (L11) `subfeatures.proficiencies.fortitude.rank 3` |
| Reflex | expert | class 2 (Evasion is L15) |
| Will | expert | class 1; Iron Will (L7) ActiveEffectLike upgrade `system.saves.will.rank` 2 |
| Light armor | expert | class 1; Alchemist Armor Expertise (L13) `subfeatures.proficiencies.light.rank 2` |
| Alchemical bombs | expert | class item rule ActiveEffectLike `weapon-base-alchemical-bomb.rank` 1; Alchemical Weapon Expertise (L7) upgrade to 2 |
| Simple weapons, unarmed | expert | class 1; Alchemical Weapon Expertise upgrade to 2 |
| Class DC | expert | Alchemical Expertise (L9) `subfeatures.proficiencies.alchemist.rank 2` |

**Numbers**

- **Max HP 164** = ancestry 10 + 14 x (class 8 + Con 3).
- **AC 35** = 10 + Dex 3 (Dex +4 capped by Studded Leather `dexCap: 3`) + light expert 18 + `acBonus` 2 + potency 2 (`runes.potency: 2`, worn and invested).
- **Fortitude +24** = 14 + 6 + Con 3 + item 1 (resilient, `runes.resilient: 1`).
- **Reflex +23** = 14 + 4 + Dex 4 + item 1. Backfire Mantle +1 circumstance applies only vs own/ally effects; skipped.
- **Will +21** = 14 + 4 + Wis 2 + item 1.
- **Perception +20** = 14 + 4 + Wis 2. Conditional, skipped: Eye Slash +1 item (visual), Earplugs -2 (auditory).
- **Class DC 32** = 10 + 14 + 4 + Int 4.
- **Speed 30 ft** = Dwarf 20 + Boots of Bounding (Greater) FlatModifier land-speed item 10 (worn, invested). Armor `speedPenalty: 0`.
- **Bomb attack** = 14 + 4 (expert) + Dex 4 + item bonus. The item bonus is the larger of the bomb's own `system.bonus.value` and Alchemist Goggles' +1 (FlatModifier attack item 1, predicate alchemical + bomb, invested); they do not stack. Greater bombs and Lodestone (bonus 2): **+24**. Moderate bombs (bonus 1): **+23**.
- **Bomb damage** = bomb dice + Weapon Specialization 2 (expert) on the direct hit; splash = bomb `splashDamage.value` + Int 4 (Expanded Splash FlatModifier, splash category). Calculated Splash is disabled by its own predicate `not feat:expanded-splash`. No Strength to damage (Str is 0 in any case).

| Bomb | Qty / carry | Attack | Direct hit | Splash |
|---|---|---|---|---|
| Alchemist's Fire (Greater) | 6, worn | +24 | 3d8+2 fire, 3 persistent fire | 7 fire |
| Bottled Lightning (Greater) | 3, worn | +24 | 3d6+2 electricity | 7 |
| Tenderizer Grenade (Greater) | 3, held | +24 | 3d6+2 acid | 7 |
| Nail Bomb (Greater) | 2, held | +24 | 6d4+2 piercing (cold iron via AdjustStrike) | 7 |
| Star Grenade (Greater) | 5, stowed in haversack | +24 | 3d8+2 fire | 7 |
| Sticky Algae Bomb (Greater) | 2, worn | +24 | 3d8+2 poison | 7 |
| Inflammation Flask (Greater) | 2, worn | +24 | 3d6+2 acid | 7 |
| Lodestone Bomb | 4, worn | +24 | 3d4+2 force | 6 |
| Crystal Shards (Moderate) | 2, worn | +23 | 2d4+2 piercing, 1 persistent bleed | 8 |
| Frost Vial (Moderate) | 0, held | +23 | 2d6+2 cold | 6 |
| Bottled Lightning (Moderate) | 0, held | +23 | 2d6+2 electricity | 6 |

- **Sticky Bomb is toggled ON in the export** (`RollOption sticky-bomb value: true`). With it on, Foundry adds persistent damage of the bomb's type to every splash bomb: max(bomb splash, Int 4) + Int 4 = 8 by the rule elements as written (rules-as-written it equals the splash, 7 for greater bombs). It is a once-per-round additive, so the fixture above leaves it out; the importer must decide how to treat a stored `value: true` toggle. Debilitating Bomb is also toggled on but is only a Note.
- **Dagger** (worn, not held; agile, finesse, thrown-10, versatile S): +22 = 14 + 4 + Dex 4; 1d4+2 piercing (Str 0 + specialization 2).
- **Sling** (inside the haversack; propulsive, reload 1): +22; 1d6+2 bludgeoning.
- **Spellsap Grenade** (4, worn): homebrew placeholder. `category: simple`, `group: null`, `baseItem: null`, `range: null`, 1d6 slashing, no splash, traits alchemical/bomb/consumable/infused. Read literally it is a melee simple weapon: +19 (14 + 4 + Str 0 + goggles 1), 1d6+2 slashing. That is clearly not what the item is meant to do. Low confidence; exclude from the gate.
- **Clan Pistol**: the Clan Pistol feat's GrantItem points at item id `932yv7n6s4w7rk7w`, which is not on the actor. There is ammunition for it (and a "Tidal Crash Round" with a `clan-pistol-damage` selector) but no pistol weapon. Proficiency is trained (`system.proficiencies.attacks.weapon-base-clan-pistol.rank 1`). No strike derived.
- **Torch** Strike rule: stowed, improvised; ignored.
- **Unarmed fist**: +22, 1d4+2 bludgeoning.
- **Spellcasting entry "Arcane Innate Spells"**: `ability: ""`, `proficiency.slug: "alchemist"`, so it rides the alchemist class statistic: DC 32, attack +22. The two spells on the actor are blank placeholders ("New 1st Level Spell", "New 2nd Level Spell") with `location: null`. Not a real caster.
- **Resistances**: poison 7 (Strong-Blooded Dwarf Resistance `max(1, floor(level/2))`), electricity 5 (Wyrm on the Wing, ChoiceSet `selection: electricity`, invested). `system.attributes.resistances` on the actor is an empty list.

**Skipped**: Clan Pistol diplomacy bonus; Earplugs; Backfire Mantle; Eye Slash; Untrained Improvisation, Artisan's Mark, Alchemist Goggles crafting, Haversack crafting, Disguise Kit, Boots of Bounding jump bonus, Stone Soul Siblings, Aethership Weapons Systems (skills only); Relic: Tidal Crash, Flare Bolt, Searing Wave (toggles, off); Sticky Bomb and Debilitating Bomb (toggles, ON, see above); Tidal Crash Round, Magnetic Shot, Ghost Ammunition, Explosive Ammunition (ammo for a weapon that is not there); actor-level `customModifiers.stealth` "Shadow Property" +1 item (skill only).

**Ignored state on the actor**: current HP 90 of 164; infused reagents 6; no effects or conditions. Unused: Brooch of Shielding and Bewitching Bloom (Cherry Blossom) are `invested: false`. Ten items are invested, which is the limit.

### 2.3 — Arabesque

**Attributes**

| Source | Str | Dex | Con | Int | Wis | Cha |
|---|---|---|---|---|---|---|
| Ancestry Elf (remaster): Dex, Int, free = cha; flaw Con | | +1 | -1 | +1 | | +1 |
| Background Noble (int, cha) | | | | +1 | | +1 |
| Class Sorcerer `keyAbility.selected = cha` | | | | | | +1 |
| Level 1 boosts [int, cha, dex, con] | | +1 | +1 | +1 | | +1 |
| Level 5 boosts [cha, int, dex, wis] | | +1 | | +1 | +1 | partial |
| Level 10 boosts [con, dex, int, cha] | | +1 | +1 | partial (stays +4) | | partial (completes +5) |
| **Total** | **0** | **+4** | **+1** | **+4** | **+1** | **+5** |

Cross-check: the Haste effect on Kas, cast by Arabesque, records `spellcasting.attribute = {type: cha, mod: 5}`.

**Proficiency ranks used**

| Stat | Rank | Source |
|---|---|---|
| Perception | expert | class 1; Alertness (L11) subfeature rank 2 |
| Fortitude | expert | class 1; Magical Fortitude (L5) subfeature rank 2 |
| Reflex | expert | class 1; Reflex Expertise (L9) subfeature rank 2 |
| Will | expert | class `savingThrows.will: 2` (Resolve is L17) |
| Unarmored | expert | class 1; Defensive Robes (L13) subfeature rank 2 |
| Simple weapons, unarmed | expert | class 1; Weapon Expertise (L11) subfeature rank 2 |
| Spellcasting | expert | class `spellcasting: 1`; Expert Spellcaster (L7) `subfeatures.proficiencies.spellcasting.rank 2` (Master is L15) |

**Numbers**

- **Max HP 104** = ancestry 6 + 14 x (class 6 + Con 1).
- **AC 34** = 10 + Dex 4 (Bands of Force DexterityModifierCap 5, not binding) + unarmored expert 18 + item 2 (Bands of Force (Greater) FlatModifier ac item 2, worn and invested). The older Bands of Force (+1) is on the sheet with `invested: false` and is not counted. Diverting Vortex +1 status vs ranged mundane attacks is conditional.
- **Fortitude +21** = 14 + 4 + Con 1 + item 2 (greater bands, selector `saving-throw`).
- **Reflex +24** = 14 + 4 + Dex 4 + item 2.
- **Will +21** = 14 + 4 + Wis 1 + item 2.
- **Perception +19** = 14 + 4 + Wis 1.
- **Spell attack +23** = 14 + 4 (expert) + Cha 5. **Spell DC 33** = 10 + 23. Applies to every entry with `ability: cha`:

| Entry | Tradition | Kind |
|---|---|---|
| Sorcerer | arcane | spontaneous; slots max by rank 1-7: 4 / 4 / 5 / 6 / 4 / 4 / 4 (ranks 3 and 4 include Ring of Wizardry IV) |
| Arcane Focus Tradition | arcane | focus (Ancestral Memories) |
| Grimoire: The Endless | arcane | prepared |
| Seer Elf | arcane | stored as "prepared" with one cantrip slot |
| Axial Awakening | arcane | innate |
| Wildborn Adept | primal | stored as "prepared" with cantrip slots |
| Otherworldy Acumen | primal | innate |

  The entries store `proficiency.value: 1`; the expert rank comes from the actor-level spellcasting proficiency, which innate spells also use.
- **Desolation Locket entry**: `ability: ""`, `proficiency.slug: "sorcerer"`. The item text says its spells use a fixed DC 30. Read as a Foundry statistic it would ride the sorcerer class DC (31, attack +21). Low confidence; treat the DC as 30 per the item text and exclude it from the gate.
- **Class DC 31** = 10 + 14 + 2 (trained) + Cha 5, on the assumption that system 8.6.1 gives every class a trained class DC (the class items carry no class DC field). The legacy sorcerer has no class DC in the rules. Low value for the simulator.
- **Top spell rank 7**: Sorcerer entry `slots.slot7.max = 4`.
- **Speed 30 ft** = Elf 30. No modifiers.
- **Blink Blade** (+2 striking dagger, held in 1 hand; agile, finesse, magical, teleportation, thrown-10, versatile S): attack +24 = 14 + 4 (simple expert) + Dex 4 (finesse) + potency 2. Damage 2d4+2 piercing = 2 dice (striking 1) + Str 0 + Weapon Specialization 2.
- **Unarmed fist**: +22, 1d4+2 bludgeoning.
- **Resistances / weaknesses / immunities**: none in any rule element.

**Skipped**: Seer Elf and Spectacles circumstance/item skill bonuses; Ancestral Suspicion (+2 circumstance vs controlled, Sense Motive); Diverting Vortex (toggle stored ON, but the modifier's predicate is ranged + mundane, so it never shows on the baseline AC); Hireling Manager; Ring of Wizardry +2 item Arcana; Grimoire Learn a Spell; Bands of Force (Greater) Athletics 21 for its reaction; Quickened Casting and Widen Spell toggles; Weapon Expertise class-gated rules; Desolation Locket +3 item vs emotion (text only, no rule element).

**Ignored state on the actor**: effect "Spell Effect: Haste"; condition "Frightened 1" (-1 status to everything; with it the open sheet would read AC 33, saves 20 / 23 / 20, DC 32); condition "Persistent Damage (4d8+12 piercing)"; current HP 104; focus 1/1; hero points 0. Stale top-level fields `system.savingThrows` and `system.martial` are leftovers from an old schema and disagree with the class item; ignore them.

### 2.4 — The two extra items in `source/items/`

- `effect-faithful-strike.json` — a temporary effect (1 round). Two ItemAlteration rules add the `holy` trait to weapons and spells. No number changes. Not baseline.
- `swear-fidelity.json` — a one-action ability with no rule elements. No number changes.

## 3 — Mismatches against PARTY[14], and confidence

### 3.1 — Mismatches

| PC | Number | Derived | Script | Best explanation | More likely right |
|---|---|---|---|---|---|
| Arabesque | AC | 34 | 33 | The script predates Bands of Force (Greater). Its own DEFAULTS say `ara_ac` "+N (greater bands of force = 1)", so the 33 was read with the +1 bands, which are still on the sheet, uninvested. Frightened 1 on the current actor would also show 33. | Derived |
| Arabesque | Fortitude | +21 | 20 | Same: +1 bands then, +2 now. | Derived |
| Arabesque | Reflex | +24 | 22 | +1 of the gap is the bands. The other +1 is not explained. With the old bands the export gives 23, not 22. | Derived (medium) |
| Arabesque | Will | +21 | 21 | Matches now, but by the bands logic the script should have read 20. So the script's Ref/Will pair looks like Dex +3 / Wis +2 rather than the export's Dex +4 / Wis +1; yet that would make the script's AC 32, not 33. The script values are not self-consistent with any single state I can build. | Derived (medium) |
| Tavish | AC | 35 | 36 | 36 needs the Dex cap ignored (Dex +4 instead of the Studded Leather cap of 3) or a +3 potency rune. The export clearly has `dexCap: 3`, `potency: 2`. | Derived |
| Tavish | Fortitude | +24 | 22 | Not explained by one change. Without the resilient rune the export gives 23. The script's DEFAULTS mention "greater resilient rune = 2" as a what-if, which suggests the script numbers have no resilient bonus at all. | Derived |
| Tavish | Reflex | +23 | 21 | Without resilient: 22. Still 1 off. | Derived |
| Tavish | Will | +21 | 20 | Without resilient: 20. Consistent with "script has no resilient rune". | Derived |
| Kas | AC | 36 | 35 | Probably an older armor potency rune (+1) or a 1-point penalty when read. | Derived |
| Kas | Fortitude | +26 | 23 | Not explainable at level 14: master Fortitude with Con +4 is +24 before any item bonus, and HP 230 proves Con +4. 23 is the level-13 number without a resilient rune. | Derived |
| Kas | Reflex | +20 | 21 | Script is higher, which no older-gear story explains. Expert + Dex 0 is +18, +2 resilient. Bulwark would give 23 vs damage. | Derived (medium) |
| Kas | Will | +23 | 22 | 22 is the level-13 number with the amulet (13 + 4 + 3 + 2). | Derived |
| Kas | Class DC | 33 | 32 | 32 is the level-13 number, or level 14 with Str +4. The attack (+29) and damage (+9) need Str +5 at level 14, so 33 follows. | Derived |

Overall reading: HP, Kas's attack and damage, Tavish's bomb attack and Arabesque's DC and top rank all match, so the build reading (boosts, levels, weapon ranks) is sound. The defensive numbers in the script look like a mix of an earlier sheet state (before the greater bands, resilient runes, possibly one level back for Kas's saves) and hand estimates; the level-15 row adds a flat +1 to each and ignores Evasion, which shows the same looseness. I could not make every script value come out of one consistent earlier state, so I say plainly: the Kas Reflex 21, Tavish AC 36 and Arabesque Reflex 22 / Will 21 pair remain unexplained.

### 3.2 — Confidence per derived number

| PC | High | Medium | Low |
|---|---|---|---|
| Kas | attributes, HP, attack, damage, Fortitude, Will, Perception, speed, slashing resistance | AC (36 rests on potency 2 counting, which it should), Reflex (plain +20; the bulwark +23 variant is situational), Class DC (33; script says 32) | spell attack/DC of the empty "charges" entry |
| Tavish | attributes, HP, bomb attack, Perception, Class DC, speed, resistances, dagger, sling | AC (Dex cap is the deciding term), all three saves (correct for the export; the script disagrees by 1-2), bomb damage (Weapon Specialization +2 on bombs and splash = base + Int are how the rule elements read; the Sticky Bomb toggle is a judgement call) | Spellsap Grenade, innate entry DC, Clan Pistol (absent) |
| Arabesque | attributes, HP, spell attack, spell DC, top rank, Perception, speed, Blink Blade | AC and saves (the +2 from the greater bands is certain in the export; the script disagrees) | Class DC 31, Desolation Locket DC |

One global caveat: I did not run Foundry. Where the remaster "trained class DC for all" and "statistic slug on a spellcasting entry" behaviours matter, I am working from memory of the system, not from this repo.

## 4 — Rule element kinds and weapon traits

### 4.1 — Rule element `key` counts (items on the three actors)

| Key | Kas | Tavish | Arabesque | Total | Touches the gate numbers? |
|---|---|---|---|---|---|
| FlatModifier | 14 | 23 | 14 | 51 | Yes: hp (Toughness, Mountain's Stoutness), land-speed (Fleet, Boots of Bounding), will (amulet), ac + saving-throw (Bands of Force), weapon-damage (Weapon Specialization), attack (Alchemist Goggles), alchemical-bomb-damage (Expanded Splash). Most others are predicated skill or situational bonuses. |
| ActiveEffectLike | 7 | 11 | 4 | 22 | Yes: proficiency rank upgrades (`system.saves.will.rank`, `system.perception.rank`, `system.proficiencies.attacks.*.rank`). Others set dying values, crafting data, languages, flags. |
| Note | 6 | 13 | 3 | 22 | No (text only). |
| GrantItem | 4 | 14 | 1 | 19 | Indirectly: granted feats are already on the actor as items. Two use dynamic uuids (`{item|flags...}`, `{actor|flags...}`). |
| AdjustModifier | 7 | 5 | 3 | 15 | Yes: Weapon Specialization upgrade to 3/4, Unburdened Iron suppresses the armor speed penalty, Calculated Splash / Sticky Bomb, amulet fear rider, Fearsome Brute multiplier. |
| RollOption | 4 | 6 | 3 | 13 | Indirectly: toggles that gate other rules. Three are stored ON (Tavish Sticky Bomb, Tavish Debilitating Bomb, Arabesque Diverting Vortex). |
| ItemAlteration | 0 | 3 | 8 | 11 | No (traits, descriptions, tags, area size). |
| Resistance | 8 | 2 | 0 | 10 | Yes: Kas slashing 4 (one of eight predicated rules fires), Tavish poison 7 and electricity 5. Values are formulas. |
| ChoiceSet | 2 | 3 | 3 | 8 | Indirectly: `selection` feeds other rules (weapon group hammer, electricity, clan-pistol, research field). |
| AdjustDegreeOfSuccess | 6 | 1 | 0 | 7 | No (degree of success on saves; matters to the sim's save logic, not the numbers). |
| DamageDice | 2 | 4 | 0 | 6 | Only behind toggles (Vicious Swing, Brutal Finish, relics, ammo). Baseline: none. |
| MartialProficiency | 4 | 0 | 0 | 4 | Yes: Kas hammer group master then legendary. |
| AdjustStrike | 0 | 3 | 0 | 3 | No for the numbers (range increment 60, cold iron material, ghost touch). |
| CriticalSpecialization | 1 | 0 | 1 | 2 | No for the numbers (Kas's fires; Arabesque's is class-gated off). |
| DexterityModifierCap | 0 | 0 | 2 | 2 | Yes in principle (cap 5; Dex is 4, so not binding). |
| Sense | 1 | 1 | 0 | 2 | No. |
| TokenLight | 1 | 1 | 0 | 2 | No. |
| TokenEffectIcon | 1 | 1 | 0 | 2 | No. |
| CraftingAbility | 0 | 1 | 0 | 1 | No. |
| Strike | 0 | 1 | 0 | 1 | No (stowed torch). |

Not rule elements but equally load-bearing: `system.subfeatures.proficiencies` on class-feature items carries most rank increases (Juggernaut, Armor Expertise, Fighter Expertise, Weapon Legend base ranks, Alertness, Alchemical Expertise, Alchemist Armor Expertise, Magical Fortitude, Reflex Expertise, Expert Spellcaster, Weapon Expertise, Defensive Robes). An importer that reads only `rules` misses them.

The two condition items (Frightened: FlatModifier selector `all`) and the Haste effects (GrantItem) are counted above under Arabesque and Kas; they are state, not build.

### 4.2 — Weapon traits

- Kas, Protector's Fury: gnome, trip, versatile-p.
- Arabesque, Blink Blade: agile, finesse, magical, teleportation, thrown-10, versatile-s.
- Tavish, dagger: agile, finesse, thrown-10, versatile-s. Sling: propulsive.
- Tavish, bombs (union): alchemical, bomb, consumable, splash, infused, plus acid, cold, electricity, fire, force, earth, plant, water, disease.

## 5 — Surprises and importer must-handles

1. **Proficiency increases live in two places.** Some are ActiveEffectLike rules (`system.saves.will.rank`, `system.perception.rank`, `system.proficiencies.attacks.<x>.rank`); most are in `system.subfeatures.proficiencies` on the feat item, keyed by save name, armor category, weapon category, `spellcasting`, or the class slug for class DC. Base ranks come from the class item (`perception`, `savingThrows`, `attacks`, `defenses`, `spellcasting`). The class items have no class DC field.
2. **Class-feature list on the class item includes future levels** (Evasion 15, Greater Weapon Specialization 15, Master Spellcaster 15...). Only the feat items actually on the actor count.
3. **`system.abilities` is null; attributes are boosts only.** Ancestry boosts need three cases: fixed (`boosts.N.value` has one entry, `selected` may be null), free (`selected` set), and `alternateAncestryBoosts` (Kas), which replaces all ancestry boosts and the flaw even though the old `selected` values are still stored. Flaws apply even with `selected: null` (Tavish Cha, Arabesque Con). Partial boosts above +4 must be tracked (Kas Str, Arabesque Cha complete to +5; Tavish Int and Arabesque Int stay at +4 with a partial).
4. **`system.details.keyability.value` is unreliable.** Tavish's says "str" while the class is Int-only with `keyAbility.selected: null`. Use the class item: `selected`, or the only entry in `value`.
5. **Mixed editions inside one PC.** Kas: legacy Fighter (Attack of Opportunity, Juggernaut, Bravery) with the remaster-named Vicious Swing. Arabesque: legacy Sorcerer class item, remaster Elf, remaster-named Reflex Expertise (the class item lists the legacy name Lightning Reflexes), remaster spell names. Tavish: fully legacy alchemist (infused reagents, perpetual infusions) on remaster-style boosts. The class items say `publication.remaster: false`.
6. **Stored toggles.** RollOption rules keep their state in `value`. Three are ON in the exports (Sticky Bomb, Debilitating Bomb, Diverting Vortex). A "sheet parity" target would include Sticky Bomb's persistent damage on Tavish's bombs; a "baseline" target would not. The fixture above is baseline.
7. **State on the actors that must be dropped**: Haste effect on Kas and Arabesque, Slowed 1 on Kas, Frightened 1 and Persistent Damage on Arabesque, Tavish at 90 of 164 HP. Frightened carries a FlatModifier with selector `all`, so a naive rule-element pass would lower every Arabesque number by 1.
8. **Max HP is not stored**; only `attributes.hp.value` (current). It happens to equal max for Kas and Arabesque, not for Tavish.
9. **Item bonuses of the same type do not stack**: Kas has resilient +2 and the amulet's +2 item to Will; Tavish's bombs carry their own item bonus beside the goggles' +1.
10. **Rule values are formulas, not numbers**: `@actor.level`, `max(1,floor(@actor.level/2))`, `2 + @armor.system.runes.potency`, `@actor.abilities.int.mod`, `@weapon.system.splashDamage.value`, `ternary(...)`, `match(when(...))`, and brace templates such as `{item|flags.pf2e.rulesSelections.wotw}` for a resistance type. Predicates use nested `or` / `not` / `gte` / `lte` objects.
11. **Runes are small integers and string lists on the item**: weapon `runes.potency`, `runes.striking` (1 striking, 2 greater), `runes.property`; armor `runes.potency`, `runes.resilient`, `runes.property`. Separate "equipment" items named Dread (Moderate), Spellwatch, Snagging, Shock, Impactful are loose rune stones or records, not active bonuses.
12. **Gear state**: `equipped.carryType` (held / worn / stowed), `handsHeld`, `invested` (true / false / null for not investable), `containerId`. An item can be `carryType: worn` and still sit in a container (Tavish's sling, healer's tools). Kas's shield is destroyed and stowed. Arabesque carries two Bands of Force; only the invested one counts. Quantities are used as counters (Brooch of Shielding qty 20 and 30) and can be 0 (two of Tavish's bombs).
13. **Odd weapon data**: Protector's Fury has `slug: gnome-hooked-hammer` but `baseItem: maul`, d12, group hammer, and neither the two-hand nor the shove trait: a homebrew blend. Blink Blade keeps a `specific` block that repeats its runes. Tavish's Spellsap Grenade is a shell (simple, no group, no range, 1d6 slashing). Infused bombs carry `temporary: true`. Bombs have `category: martial`; Tavish's proficiency comes only from the `weapon-base-alchemical-bomb` key.
14. **Dangling references**: Tavish's Clan Pistol feat grants an item id that is not on the actor; pistol ammunition and a custom "Tidal Crash Round" (selectors `clan-pistol-damage`, `tavish-selector-damage`) remain.
15. **Spellcasting entries are many and messy.** Arabesque has eight (two innate elf entries are stored as "prepared"); Kas and Tavish each have a non-caster entry ("charges", empty innate). Two entries use `ability: ""` with `proficiency.slug` naming a class. Entry `proficiency.value` stays at 1; the real rank comes from the actor's spellcasting proficiency. Tavish has two blank placeholder spells with `location: null`. Sorcerer slot `max` values already include the Ring of Wizardry.
16. **Legacy leftovers on the actor**: Arabesque's `system.savingThrows` and `system.martial`; Tavish's `system.customModifiers.stealth` (a hand-entered +1 item bonus) and `system.proficiencies.attacks`; `system.saves: {}`.
17. **Homebrew campaign content with rule elements**: "Gift"/"Relic" bonus feats in `location: campaign`, actions carrying rules (Feather Steps, Inextinguishable, Linguistics Nexus), a consumable carrying damage rules, and a module trait `xdy-pf2e-housepatched` on Kas's armor.

## 6 — Skills

Read by Robert from the Foundry sheets at level 14, for `F.041.E4` Commit 10 (`TestGateSkills`). Each cell is the total modifier the sheet shows. Sheet state as in the header: no conditions, no effect items, toggles off, gear as equipped.

| Skill | Kas | Tavish | Arabesque |
|---|---|---|---|
| Acrobatics |20|19|22|
| Arcana |0|20|22|
| Athletics |25|13|0|
| Crafting |0|26|4|
| Deception |20|23|25|
| Diplomacy |20|17|5|
| Intimidation |26|17|21|
| Medicine |3|18|1|
| Nature |3|18|1|
| Occultism |0|18|24|
| Performance |4|17|23|
| Religion |3|16|1|
| Society |0|20|20|
| Stealth |0|24|20|
| Survival |19|16|1|
| Thievery |0|19|20|
