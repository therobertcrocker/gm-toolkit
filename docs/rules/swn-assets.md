# SWN Faction Assets — Reference Transcription

Source: `reference/Faction_Rules_SWN.pdf`, PDF pages 7–12 (book pages 218–223). This file is the project's source of truth for SWN asset data.

The transcription is verbatim. It keeps the book's own typos and inconsistencies (for example "Saboteurs: Minions rained…", the Capital Fleet type "Spaceship", and the Monopoly Note "S" while its text says "As an action"). Table cells drop trailing periods, per the house style. Italics in the descriptions follow the book.

> **Project note:** The book says "FacCred(s)". This project calls FacCreds **Coin**. The transcribed text keeps "FacCred".

<br/>
<br/>

# Flag Legend

A "P" special code indicates a need for planetary governmental permission to raise or transport in the asset. "A" means that the asset can perform a special action, and "S" indicates that the asset has a special feature or cost.

<br/>
<br/>

# Cunning Assets

## Cunning Asset Table

| Asset | ID | Rating | HP | Cost | TL | Type | Attack | Counter | Note |
|-------|----|--------|----|------|----|------|--------|---------|------|
| Smugglers | SWN-C1-001 | 1 | 4 | 2 | 4 | Starship | Cunning vs. Wealth, 1d4 damage | None | A |
| Informers | SWN-C1-002 | 1 | 3 | 2 | 0 | Special Forces | Cunning vs. Cunning, special | None | A, S |
| False Front | SWN-C1-003 | 1 | 2 | 1 | 0 | Logistics Facility | None | None | S |
| Base of Influence | — | 1 | * | * | 0 | Special | None | None | S |
| Lobbyists | SWN-C2-001 | 2 | 4 | 4 | 0 | Special Forces | Cunning vs. Cunning, special | None | S |
| Saboteurs | SWN-C2-002 | 2 | 6 | 5 | 0 | Special Forces | Cunning vs. Cunning, 2d4 damage | None | S |
| Blackmail | SWN-C2-003 | 2 | 4 | 4 | 0 | Tactic | Cunning vs. Cunning, 1d4+1 damage | None | S |
| Seductress | SWN-C2-004 | 2 | 4 | 4 | 0 | Special Forces | Cunning vs. Cunning, special | None | A, S |
| Cyberninjas | SWN-C3-001 | 3 | 4 | 6 | 4 | Special Forces | Cunning vs. Cunning, 2d6 damage | None | - |
| Stealth | SWN-C3-002 | 3 | - | 2 | 0 | Tactic | None | None | S |
| Covert Shipping | SWN-C3-003 | 3 | 4 | 8 | 4 | Logistics Facility | None | None | A, S |
| Party Machine | SWN-C4-001 | 4 | 10 | 8 | 0 | Logistics Facility | Cunning vs. Cunning, 2d6 damage | 1d6 damage | S |
| Vanguard Cadres | SWN-C4-002 | 4 | 12 | 8 | 3 | Military Unit | Cunning vs. Cunning, 1d6 damage | 1d6 damage | - |
| Tripwire Cells | SWN-C4-003 | 4 | 8 | 12 | 4 | Special Forces | None | 1d4 damage | A, S |
| Seditionists | SWN-C4-004 | 4 | 8 | 12 | 0 | Special Forces | None | None | A |
| Organization Moles | SWN-C5-001 | 5 | 8 | 10 | 0 | Tactic | Cunning vs. Cunning, 2d6 damage | None | |
| Cracked Comms | SWN-C5-002 | 5 | 6 | 14 | 0 | Tactic | None | Special | S |
| Boltholes | SWN-C5-003 | 5 | 6 | 12 | 4 | Logistics Facility | None | 2d6 damage | S |
| Transport Lockdown | SWN-C6-001 | 6 | 10 | 20 | 4 | Tactic | Cunning vs. Cunning, special | None | S |
| Covert Transit Net | SWN-C6-002 | 6 | 15 | 18 | 4 | Logistics Facility | None | None | A |
| Demagogue | SWN-C6-003 | 6 | 10 | 20 | 0 | Special Forces | Cunning vs. Cunning, 2d8 damage | 1d8 damage | - |
| Popular Movement | SWN-C7-001 | 7 | 16 | 25 | 4 | Tactic | Cunning vs. Cunning, 2d6 damage | 1d6 damage | S |
| Book of Secrets | SWN-C7-002 | 7 | 10 | 20 | 4 | Tactic | None | 2d8 damage | S |
| Treachery | SWN-C7-003 | 7 | 5 | 10 | 0 | Tactic | Cunning vs. Cunning, Special | None | S |
| Panopticon Matrix | SWN-C8-001 | 8 | 20 | 30 | 5 | Logistics Facility | None | 1d6 damage | S |

<br/>

## Cunning Asset Descriptions

**Base of Influence:** This asset is special, and is required for purchasing or upgrading units on a particular world. Any damage done to a Base of Influence is also done to a faction's hit points. The cost of a Base of Influence equals its maximum hit points, which can be any number up to the total maximum hit points of its owning faction. A faction's bases of influence don't count against their maximum assets. A Base of Influence can only be purchased with the *Expand Influence* action.

**Blackmail:** Selectively degrade the effectiveness of an asset. Any attempt to attack or defend against Blackmail loses any bonus dice earned by tags.

**Boltholes:** Equipped with a number of postech innovations to make cleaning them out a costly and dangerous pursuit. If a faction Special Forces or Military Unit asset on the same planet as the Boltholes suffers damage sufficient to destroy it, it is instead set at 0 HP and rendered untouchable and unusable until it is repaired to full strength. If the Boltholes are destroyed before this happens, the asset is destroyed with them.

**Book of Secrets:** Exhaustively cataloged psychometric records on important and influential local figures, allowing uncanny accuracy in predicting their actions. Once per turn, a Book of Secrets allows the faction to reroll one die for an action taken on that world or force an enemy faction to reroll one die. This reroll can only be forced once per turn, no matter how many Books of Secrets are owned.

**Covert Shipping:** Quiet interstellar asset transport. Any one Special Forces unit can be moved between any worlds within three hexes of the Covert Shipping at the cost of one FacCred.

**Covert Transit Net:** Facilities web an area of space with a network of smugglers and gray-market freighter captains. As an action, any Special Forces assets can be moved between any worlds within three hexes of the Covert Transit Net.

**Cracked Comms:** A cryptographic asset for the interception and deciphering of enemy communications. Friendly fire can be induced with the right interference. If the Cracked Comms succeeds in defending against an attack, it can immediately cause the attacking asset to make an attack against itself for normal damage or counterattack results.

**Cyberninjas:** Outfitted with the latest in personal stealth cyberware, cyberninjas are designed to avoid all but the most careful scans.

**Demagogues:** Popular leaders of a particular faith or ideology that can be relied upon to point their followers in the direction of maximum utility.

**False Front:** This asset allows a faction to preserve more valuable resources. If another asset on the planet suffers enough damage to destroy it, the faction can sacrifice the false front instead to nullify the killing blow.

**Informers:** Minions that lace a planet's underworld, watchful for intruders. They can choose to Attack a faction without specifying a target asset. On a successful Cunning vs. Cunning attack, all Stealthed assets on the planet belonging to that faction are revealed. Informers can target a faction even if none of their assets are visible on a world; at worst, they simply learn that there are no stealthed assets.

**Lobbyists:** Usable to block the governmental permission that is sometimes required to buy an asset or transport it into a system. When a rival faction gains permission to do so, the Lobbyists can make an immediate Cunning vs. Cunning test against the faction; if successful, the permission is withdrawn and cannot be re-attempted until next turn.

**Organization Moles :** These can subvert and confuse enemy assets, striking to damage their cohesion.

**Panopticon Matrix:** These facilities weave braked-AI intelligence analysts into a web of observation capable of detecting the slightest evidence of intruders on a world. Every rival Stealthed asset on the planet must succeed in a Cunning vs. Cunning test at the beginning of every turn or lose their Stealth. The owner also gains an additional die on all Cunning attacks and defenses on that planet.

**Party Machines:** Political blocks control particular cities or regions… blocks that are firmly in control of the faction. Each turn, a Party Machine provides 1 FacCred to its owning faction.

**Popular Movements:** A planet-wide surge of enthusiasm for a cause controlled by the faction. This support pervades all levels of government, and the government always grants any asset purchase or movement requests made by the faction.

**Saboteurs:** Minions rained in launching strikes against enemy operations. An asset attacked by saboteurs cannot apply any *Use Asset Ability* action until the start of the attacking faction's next turn. This applies whether or not the attack was successful.

**Seditionists:** These asset sap a target's loyalty and will to obey. For a cost of 1d4 FacCreds, the Seditionists can attach themselves to an enemy asset. Until they attach to a different asset or no longer share the same planet, the affected asset cannot attack. If the asset is destroyed, the Seditionists survive.

**Seductresses:** They and their male equivalents subvert the leadership of enemy assets. As an action, a Seductress can travel to any world within one hex. As an attack, a Seductress does no damage, but an asset that has been successfully attacked immediately reveals any other Stealthed assets of that faction on the planet. Only Special Forces units can attack a Seductress.

**Smugglers:** Men and women skilled in extracting personnel. For one FacCred, the smugglers asset can transport itself and/or any one Special Forces unit to a planet up to two hexes away.

**Stealth:** Not an asset, per se, but a special quality that can be purchased for another Special Forces asset on the planet. An asset that has been Stealthed cannot be detected or attacked by other factions. If the unit normally requires the permission of a planetary government to be moved onto a planet, that permission may be foregone. An asset loses its Stealth if it is used to attack or defend.

**Transport Lockdown:** These techniques involve selective pressure on local routing and shipping companies. On a successful Cunning vs. Cunning attack against a rival faction, the rival faction cannot transport assets onto that planet without spending 1d4 FacCreds and waiting one turn.

**Treachery:** Traitors can attack an enemy asset. On a successful attack, the Treachery asset is lost, 5 FacCreds are gained, and the targeted asset switches sides to join the traitor's faction, even if the faction does not otherwise have the attributes necessary.

**Tripwire Cells:** These observers are alert to the arrival of stealthed units. Whenever a stealthed asset lands or is purchased on a planet with tripwire cells, the Cells make an immediate Cunning vs. Cunning attack against the owning faction. If successful, the asset loses its stealth.

**Vanguard Cadres:** Followers of the movement are inspired sufficiently to take up arms and fight on behalf of their leadership.

<br/>
<br/>

# Force Assets

## Force Asset Table

| Asset | ID | Rating | HP | Cost | TL | Type | Attack | Counter | Note |
|-------|----|--------|----|------|----|------|--------|---------|------|
| Security Personnel | SWN-F1-001 | 1 | 3 | 2 | 0 | Military Unit | Force vs. Force, 1d3+1 damage | 1d4 damage | - |
| Hitmen | SWN-F1-002 | 1 | 1 | 2 | 0 | Special Forces | Force vs. Cunning, 1d6 damage | None | - |
| Militia Unit | SWN-F1-003 | 1 | 4 | 4 | 3 | Military Unit | Force vs. Force, 1d6 damage | 1d4+1 damage | P |
| Base of Influence | — | 1 | * | * | 0 | Special | None | None | S |
| Heavy Drop Assets | SWN-F2-001 | 2 | 6 | 4 | 4 | Facility | None | None | A |
| Elite Skirmishers | SWN-F2-002 | 2 | 5 | 5 | 4 | Military Unit | Force vs. Force, 2d4 damage | 1d4+1 damage | P |
| Hardened Personnel | SWN-F2-003 | 2 | 4 | 4 | 3 | Special Forces | None | 1d4+1 damage | - |
| Guerrilla Populace | SWN-F2-004 | 2 | 6 | 4 | 0 | Military Unit | Force vs. Cunning, 1d4+1 damage | None | - |
| Zealots | SWN-F3-001 | 3 | 4 | 6 | 0 | Special Forces | Force vs Force, 2d6 damage | 2d6 damage | S |
| Cunning Trap | SWN-F3-002 | 3 | 2 | 5 | 0 | Tactic | None | 1d6+3 damage | - |
| Counterintel Unit | SWN-F3-003 | 3 | 4 | 6 | 4 | Special Forces | Cunning vs Cunning, 1d4+1 damage | 1d6 damage | - |
| Beachhead Landers | SWN-F4-001 | 4 | 10 | 10 | 4 | Facility | None | None | A |
| Extended Theater | SWN-F4-002 | 4 | 10 | 10 | 4 | Facility | None | None | A |
| Strike Fleet | SWN-F4-003 | 4 | 8 | 12 | 4 | Starship | Force vs. Force, 2d6 damage | 1d8 damage | A |
| Postech Infantry | SWN-F4-004 | 4 | 12 | 8 | 4 | Military Unit | Force vs. Force, 1d8 damage | 1d8 damage | P |
| Blockade Fleet | SWN-F5-001 | 5 | 8 | 10 | 4 | Starship | Force vs. Wealth, 1d6 damage * | None | S |
| Pretech Logistics | SWN-F5-002 | 5 | 6 | 14 | 0 | Facility | None | None | A |
| Psychic Assassins | SWN-F5-003 | 5 | 4 | 12 | 4 | Special Forces | Cunning vs. Cunning, 2d6+2 damage | None | S |
| Pretech Infantry | SWN-F6-001 | 6 | 16 | 20 | 5 | Military Unit | Force vs. Force, 2d8 damage | 2d8+2 damage | P |
| Planetary Defenses | SWN-F6-002 | 6 | 20 | 18 | 4 | Facility | None | 2d6+6 damage * | S |
| Gravtank Formation | SWN-F6-003 | 6 | 14 | 25 | 4 | Military Unit | Force vs. Force, 2d10+4 damage | 1d10 damage | P |
| Deep Strike Landers | SWN-F7-001 | 7 | 10 | 25 | 4 | Facility | None | None | A |
| Integral Protocols | SWN-F7-002 | 7 | 10 | 20 | 5 | Facility | None | 2d8+2 damage * | S |
| Space Marines | SWN-F7-003 | 7 | 16 | 30 | 4 | Military Unit | Force vs. Force 2d8+2 damage | 2d8 damage | A |
| Capital Fleet | SWN-F8-001 | 8 | 30 | 40 | 4 | Spaceship | Force vs. Force, 3d10+4 damage | 3d8 damage | A, S |

<br/>

## Force Asset Descriptions

**Base of Influence:** This asset is special, and is required for purchasing or upgrading units on a particular world. Any damage done to a Base of Influence is also done to a faction's hit points. The cost of a Base of Influence equals its maximum hit points, which can be any number up to the total maximum hit points of its owning faction. A faction's bases of influence don't count against their maximum assets. A Base of Influence can only be purchased with the *Expand Influence* action.

**Beachhead Landers:** A collection of short-range, high-capacity spike drive ships capable of moving large numbers of troops. As an action, the asset may move any number of assets on the planet, including itself, to any world within one hex at a cost of one FacCred per asset moved.

**Blockade Fleets:** A ragtag lot of corsairs, pirates, privateers, and other deniable assets. When they successfully *Attack* an enemy faction asset, they steal 1d4 FacCreds from the target faction as well. This theft can occur to a faction only once per turn, no matter how many blockade fleets attack. As an action, this asset may also move itself to a world within one hex.

**Capital Fleets:** The pride of an empire, this is a collection of massive capital warships without peer in most sectors. Capital fleets are expensive to keep flying, and cost an additional 2 FacCreds of maintenance each turn. As an action, they may move to any world within three hexes of their current location. Planetary government permission is required to raise a capital fleet, but not to move one into a system.

**Counterintel Units:** These security units specialize in code breaking, internal security, and monitoring duties. They can crack open enemy plots long before they have time to come to fruition.

**Cunning Traps:** This asset covers all the myriad stratagems of war, from induced landslides to spreading local diseases.

**Deep Strike Landers:** These advanced transport ships are capable of moving an asset long distances. As an action, any one non-Starship asset, including itself, can be moved between any two worlds within three hexes of the deep strike landers, at a cost of 2 FacCreds. This movement can be done even if the local planetary government objects, albeit doing so is usually an act of open war.

**Elite Skirmishers:** These lightly-equipped troops are trained for guerrilla warfare and quick raids.

**Extended Theater:** These facilities allow for transporting assets long distances. As an action, any one non-Starship asset, including itself, can be moved between any two worlds within two hexes of the extended theater, at a cost of 1 FacCred.

**Gravtank Formations:** These tank formations are composed of advanced postech gravtank units that are capable of covering almost any terrain and cracking even the toughest defensive positions.

**Guerrilla Populace:** Such assets reflect popular support among the locals and a cadre of men and women willing to fight as partisans.

**Hardened Personnel:** These are the employees and support staff of the faction that have been trained in defensive fighting and equipped with supply caches and hardened fallback positions.

**Heavy Drop Assets:** These craft allow for the transport of resources from one world to another. As an action, any one non-Starship asset, including this one, may be moved to any world within one hex for one FacCred.

**Hitmen:** Crudely-equipped thugs and assassins with minimal training that have been aimed at a rival faction's leadership.

**Integral Protocols:** A complex web of braked-AI supported sensors and redundant security checks used to defeat attempts to infiltrate an area. They can defend only against attacks versus Cunning, but they add an additional die to the defender's roll.

**Militia Units:** Groups of lightly-equipped irregular troops with rudimentary military training but no heavy support.

**Planetary Defenses:** Massive mag cannons and gravitic braker gun arrays designed to defend against starship bombardments and repel unauthorized landing. Planetary defenses can only defend against attacks by Starship-type assets.

**Postech Infantry:** The backbone of most planetary armies, these well-trained soldiers are usually equipped with mag weaponry and combat field uniforms, and have heavy support units attached.

**Pretech Infantry:** The cream of the stellar ground forces, elite troops kitted out in the best pretech weaponry and armor available, with sophisticated heavy support weaponry integral to the unit.

**Pretech Logistics:** These assets represent caches, smugglers, or internal research and salvage programs. As an action, a pretech logistics asset allows the owner to buy one Force asset on that world that requires up to tech level 5 to purchase. This asset costs half again as many FacCreds as usual, rounded up. Only one asset can be purchased per turn.

**Psychic Assassins:** Combat-trained psychics equipped with advanced pretech stealth gear and psitech weaponry. Psychic assassins automatically start Stealthed when purchased.

**Security Personnel:** Standard civilian guards or policemen, usually equipped with nonlethal weaponry or personal sidearms.

**Space Marines:** Heavily-armored specialist troops trained for ship boarding actions and opposed landings. As an action, they can move to any world within one hex of their current location, whether or not the planetary government permits it.

**Strike Fleets:** Forces composed of frigate or cruiser-class vessels equipped with space-to-ground weaponry and sophisticated defenses against light planetary weaponry. As an action, they can move to any world within one hex of their current location.

**Zealots:** Members of the faction so utterly dedicated that they are willing to launch suicide attacks or hold positions to the death. Zealots take 1d4 damage every time they launch a successful attack or perform a counterattack.

<br/>
<br/>

# Wealth Assets

## Wealth Asset Table

| Asset | ID | Rating | HP | Cost | TL | Type | Attack | Counter | Note |
|-------|----|--------|----|------|----|------|--------|---------|------|
| Franchise | SWN-W1-001 | 1 | 3 | 2 | 2 | Facility | Wealth vs. Wealth, 1d4 damage | 1d4-1 damage | S |
| Harvesters | SWN-W1-002 | 1 | 4 | 2 | 0 | Facility | None | 1d4 | A |
| Local Investments | SWN-W1-003 | 1 | 2 | 1 | 2 | Facility | Wealth vs. Wealth, 1d4-1 damage | None | S |
| Base of Influence | — | 1 | * | * | 0 | Special | None | None | S |
| Freighter Contract | SWN-W2-001 | 2 | 4 | 5 | 4 | Starship | Wealth vs. Wealth, 1d4 damage | None | A |
| Lawyers | SWN-W2-002 | 2 | 4 | 6 | 0 | Special Forces | Cunning vs. Wealth, 2d4 damage | 1d6 damage | S |
| Union Toughs | SWN-W2-003 | 2 | 6 | 4 | 0 | Military Unit | Wealth vs. Force, 1d4+1 damage | 1d4 damage | |
| Surveyors | SWN-W2-004 | 2 | 4 | 4 | 4 | Special Forces | None | 1d4 damage | A, S |
| Postech Industry | SWN-W3-001 | 3 | 4 | 8 | 4 | Facility | None | 1d4 damage | A |
| Laboratory | SWN-W3-002 | 3 | 4 | 6 | 0 | Facility | None | None | S |
| Mercenaries | SWN-W3-003 | 3 | 6 | 8 | 4 | Military Unit | Wealth vs. Force, 2d4+2 damage | 1d6 damage | A, S, P |
| Shipping Combine | SWN-W4-001 | 4 | 10 | 10 | 4 | Facility | None | 1d6 damage | A |
| Monopoly | SWN-W4-002 | 4 | 12 | 8 | 3 | Facility | Wealth vs. Wealth, 1d6 damage | 1d6 damage | S |
| Medical Center | SWN-W4-003 | 4 | 8 | 12 | 4 | Facility | None | None | S |
| Bank | SWN-W4-004 | 4 | 8 | 12 | 3 | Facility | None | None | S |
| Marketers | SWN-W5-001 | 5 | 8 | 10 | 0 | Tactic | Cunning vs. Wealth, 1d6 damage | None | A |
| Pretech Researchers | SWN-W5-002 | 5 | 6 | 14 | 4 | Special Forces | None | None | S |
| Blockade Runners | SWN-W5-003 | 5 | 6 | 12 | 4 | Starship | None | 2d4 damage | A |
| Venture Capital | SWN-W6-001 | 6 | 10 | 15 | 4 | Facility | Wealth vs. Wealth, 2d6 | 1d6 damage | A |
| R&D Department | SWN-W6-002 | 6 | 15 | 18 | 4 | Facility | None | None | S |
| Commodities Broker | SWN-W6-003 | 6 | 10 | 20 | 0 | Special Forces | Wealth vs. Wealth, 2d8 damage | 1d8 damage | A |
| Pretech Manufactory | SWN-W7-001 | 7 | 16 | 25 | 5 | Facility | None | None | S |
| Hostile Takeover | SWN-W7-002 | 7 | 10 | 20 | 4 | Tactic | Wealth vs. Wealth, 2d10 damage | 2d8 damage | S |
| Transit Web | SWN-W7-003 | 7 | 5 | 15 | 5 | Facility | Cunning vs. Cunning, 1d12 damage | None | S |
| Scavenger Fleet | SWN-W8-001 | 8 | 20 | 30 | 5 | Starship | Wealth vs. Wealth, 2d10+4 damage | 2d10 damage | |

<br/>

## Wealth Asset Descriptions

**Bank Assets:** Once per turn, the faction can ignore one cost or FacCred loss imposed by another faction. This does not require an action. Multiple bank assets allow multiple losses to be ignored.

**Base of Influence:** These follow the standard rules for Bases of Influence.

**Blockade Runners:** These starship captains excel at transporting goods through unfriendly lines. As an action, a blockade runner can transfer itself or any one Military Unit or Special Forces to a world within three hexes for a cost of two FacCreds. They can even move units that would otherwise require planetary government permission to enter.

**Commodities Brokers:** They substantially lessen the cost of large-scale investments by timing materials purchases properly. As an action, the owner of a commodities broker can roll 1d8; that many FacCreds are subtracted from the cost of their next asset purchase, down to a minimum of half normal price, rounded down.

**Franchise:** This asset reflects a deniable connection with a local licensee for the faction's goods and services. When a Franchise successfully attacks a enemy asset, the enemy faction loses one FacCred (if available), which is gained by the Franchise's owner. This loss can happen only once a turn, no matter how many Franchises attack.

**Freighter Contract:** A special link with heavy shipping spacers. As an action, the faction may move any one non-Force asset, including this one, to any world within two hexes at a cost of one FacCred.

**Harvesters:** These gather the natural resources of a world, whether ore, biologicals, or other unprocessed goods. As an action, the Harvesters' owning faction may roll 1d6. On 3+, gain one FacCred.

**Hostile Takeover:** This asset can seize control of damaged and poorly-controlled assets. If a Hostile Takeover does enough damage to destroy an asset, the target is instead reduced to 1 hit point and acquired by the Hostile Takeover's owning faction.

**Laboratory:** The lab allows a world to make hesitant progress in tech. The presence of a Laboratory allows assets to be purchased on that world as if it had Tech Level 4.

**Lawyers:** Sophists in immaculate suits or charismatic tribal skalds, lawyers have the ability to tie an enemy up in the coils of their own internal rules, damaging assets with confusion and red tape. Lawyers cannot attack or counterattack Force assets.

**Local Investments:** These give the faction substantial influence over the commerce on a world. Any other faction that tries to buy an asset on that planet must pay one extra FacCred. This money is not given to the investments' owner, but is lost. This penalty is only applied once.

**Marketers:** Deployed to confuse enemy factions into untimely investments. As an action, the marketers may test Cunning vs. Wealth against a rival faction's asset. If successful, the target faction must immediately pay half the asset's purchase cost, rounded down, or have it become disabled and useless until this price is paid.

**Medical Center:** Salvage and repair damaged assets. Once between turns, if a Special Forces or Military Unit asset on the world is destroyed, the faction may immediately pay half its purchase cost to restore it with one hit point. Any *Repair Asset* action taken on that world costs one less FacCred for Special Forces and Military Units.

**Mercenaries:** Groups of well-equipped, highly-trained soldiers willing to serve the highest bidder. Mercenaries have a maintenance cost of one FacCred per turn. As an action, Mercenaries can move to any world within one hex of their current location. To purchase or move a Mercenary asset to a planet requires government permission.

**Monopoly:** An open or tacit stranglehold on certain vital businesses or resources on a world. As an action, owners of a monopoly may force one other faction with unstealthed assets on that world to pay them one FacCred. If the target faction can't pay, they lose one asset of their choice on the world.

**Postech Industry:** As an action, the owning faction can roll 1d6 for a Postech Industry asset. On a 1, one FacCred is lost, on a 2-4 one FacCred is earned, and a 5-6 returns two FacCreds. If money is lost and no resources are available to pay it, the Postech Industry is destroyed.

**Pretech Manufactories:** Rare, precious examples of functioning pretech industrial facilities, retrofitted to work without the benefit of specialized psychic disciplines. As an action, the owning faction can roll 1d8 for a Pretech Manufactory, and gain half that many FacCreds, rounded up.

**Pretech Researchers:** A highly versatile team of research and design specialists capable of supporting limited pretech… as long as they're adequately funded. Any world with Pretech Researchers on it is treated as tech level 5 for the purpose of buying Cunning and Wealth assets. Pretech researchers have a maintenance cost of 1 FacCred per turn.

**R&D Departments:** These allow the smooth extension of wealth-creation and industrial principles to the farthest reaches of the faction's operations. A faction with an R&D department may treat all planets as having tech level 4 for purposes of buying Wealth assets.

**Scavenger Fleets:** These rag-tag armadas bring enormous technical and mercantile resources to their patrons, along with a facility with heavy guns. As an action, a Scavenger Fleet can be moved to any world within three hexes. Scavenger Fleets cost 2 FacCreds a turn in maintenance.

**Shipping Combine:** Transport large amounts of equipment and personnel between worlds. As an action, the combine may move any number of non-Force assets, including itself, to any world within two hexes at a cost of one FacCred per asset.

**Surveyors:** Explore potential resource and investment options on worlds. The presence of a Surveyor crew allows one additional die to be rolled on *Expand Influence* actions. As an action, a surveyor crew can be moved to any world within two hexes.

**Transit Web:** These facilities allow almost effortless relocation of all assets. For one FacCred, any number of non-starship Cunning or Wealth assets may be moved between any two worlds within three hexes of the Transit Web. This may be done freely on the owner's turn so long as the fee can be paid, and using the ability doesn't require an action.

**Union Toughs:** These bruisers don't much like scabs and management, and they're willing to take the faction's word on which people are which. They're lightly armed and poorly trained, but they can infiltrate to perform sabotage.

**Venture Capital:** This asset grows resources out of seemingly nowhere, harvesting the best of entrepreneurship for the faction's benefit. As an action, venture capital can be tapped. 1d8 is rolled; on a 1, the asset is destroyed, while on a 2-3 one FacCred is gained, 4-7 yields two FacCreds and 8 grants three FacCreds.
