package narrative

import "math/rand"

func pick(rng *rand.Rand, pool []string) string {
	return pool[rng.Intn(len(pool))]
}

var headlineAttackTemplates = []string{
	"{name} draws blood — open warfare across the sector.",
	"The sector trembles as {name} strikes.",
}

var headlineGoalCompletedTemplates = []string{
	"{name} achieves a long-sought objective.",
	"Victory: {name} completes their goal.",
}

var headlineFactionDestroyedTemplates = []string{
	"{name} is no more — a faction erased from the board.",
	"The fall of {name} reshapes the sector.",
}

var headlineHomeworldTemplates = []string{
	"{name} stakes a new claim — homeworld designation shifts.",
	"A new homeworld rises under {name}'s banner.",
}

var headlineGoalAbandonedTemplates = []string{
	"{name} retreats from a costly objective.",
	"Ambitions cut short: {name} abandons their goal.",
}

var headlineQuietTemplates = []string{
	"A cycle of consolidation — no major moves recorded.",
	"Quiet maneuvering this cycle. No bold strikes.",
}

var ledeAttackTemplates = []string{
	"This cycle saw open conflict between rival factions, with assets destroyed and territory contested.",
	"Weapons hot: the cycle closed with multiple engagements logged across the sector.",
}

var ledeGoalTemplates = []string{
	"This cycle saw major objectives fulfilled, reshaping factional power across the sector.",
	"Progress and peril: factions closed in on long-pursued goals this cycle.",
}

var ledeQuietTemplates = []string{
	"A quiet cycle. Factions consolidated gains and prepared for moves yet to come.",
	"Little open action this cycle — factions moved in shadow, consolidating resources.",
}

var goalCompletedTemplates = []string{
	"{faction} completed their objective: {goal}.",
	"Mission accomplished — {faction} closed out {goal}.",
}

var goalAbandonedTemplates = []string{
	"{faction} abandoned the goal {goal}, cutting their losses.",
	"{faction} walked away from {goal} this cycle.",
}

var goalHomeworldTemplates = []string{
	"{faction} shifted homeworld designation from {from} to {to}.",
	"A new base of power: {faction} moved home from {from} to {to}.",
}

var goalTagTemplates = []string{
	"{faction} gained the tag: {tag}.",
	"New descriptor logged for {faction}: {tag}.",
}

var acquisitionTemplates = []string{
	"{faction} added {asset} to their order of battle.",
	"New acquisition: {faction} brought {asset} online.",
}

var lossTemplates = []string{
	"{faction} divested {asset}.",
	"{asset} struck from {faction}'s roster.",
}

var movementTemplates = []string{
	"{faction} redeployed {asset} from {from} to {to}.",
	"{asset} moved from {from} to {to} under {faction}'s orders.",
}

var bribeTemplates = []string{
	"{faction} spread coin at {location} to buy loyalty.",
	"Influence purchased: {faction} dropped {coin} coin at {location}.",
}

var repairAssetTemplates = []string{
	"{faction} spent {cost} Coin repairing {asset}, restoring {hp} HP.",
	"Maintenance logged: {faction} brought {asset} back to readiness (+{hp} HP) at a cost of {cost} Coin.",
}

var repairFactionTemplates = []string{
	"{faction} spent {cost} Coin on recovery operations, restoring {hp} faction HP.",
	"Recovery underway: {faction} invested {cost} Coin and restored {hp} HP.",
}

var expansionNewBaseTemplates = []string{
	"{faction} established a new base at {location}.",
	"New foothold: {faction} planted roots at {location}.",
}

var expansionTemplates = []string{
	"{faction} expanded holdings at {location}.",
	"{faction} reinforced their position at {location}.",
}

var stealthAppliedTemplates = []string{
	"{faction} cloaked {asset}, moving it off the grid.",
	"{asset} went dark under {faction}'s orders.",
}

var stealthClearedTemplates = []string{
	"{faction}'s asset {asset} had its cover blown.",
	"Exposure: {asset} stripped of stealth ({faction}).",
}

var crossAttackTemplates = []string{
	"{attacker} sent {attacker_asset} against {defender}'s {defender_asset}, dealing {damage} damage.",
	"Conflict: {attacker} struck at {defender} — {attacker_asset} vs {defender_asset}, {damage} damage recorded.",
}

var crossAbilityTemplates = []string{
	"{attacker} leveraged a special ability against {defender}.",
	"Tactical action: {attacker} deployed an ability against {defender}.",
}

var quietTailTemplates = []string{
	"No major action reported from: {factions}.",
	"The following factions remained quiet this cycle: {factions}.",
}
