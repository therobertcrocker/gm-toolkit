# PF2e Encounter Simulator (F.041) — Effort 1 Plan: Engine Core

## In Plain Words

- **What:** the first working slice of the PF2e simulator. A seeded dice roller, creature files built from layers, and an engine that can play one fight. Each creature takes three actions a turn, makes Strikes against AC, deals damage that resistance and weakness adjust, and PCs fall dying while NPCs die. The engine reports every step as an event. A `gm-toolkit pf2e sim show <creature>` command prints a merged stat block, and the wiki gets a PF2e side.
- **Why:** every later effort builds on this core. Import (Effort 2) writes these creature files, rules coverage (Effort 3) extends this engine, and the runner (Effort 6) reads these events. Effort 1 ends at gate 3: the same seed and files give the same fight, event for event.
- **How it fits:** Effort 1 of 6 in F.041, on branch `feature/pf2e-encounter-simulator`. It creates `internal/pf2e/sim/` and its first three packages. Tactics, position, import and the runner come later; until then creatures follow a fixed, simple behaviour that exists only for tests.

## Context / Goal

- Discovery: [`F.041-pf2e-encounter-simulator-discovery.md`](../discovery/F.041-pf2e-encounter-simulator-discovery.md)
- Overview plan: [`F.041-pf2e-encounter-simulator-overview-plan.md`](./F.041-pf2e-encounter-simulator-overview-plan.md). Its Decisions 1 to 7 bind this plan, and the overview's Work Breakdown, Effort 1, sets the scope.
- Ends with: a hand-written Luna fights dummy PCs, and the same seed and files give identical events (gate 3, `TestRunIsDeterministic`).

## Decisions Ratified in Planning

1. **Effort 1 builds the core rules; Effort 3 builds the rest.** Robert's call. In Effort 1: initiative, three actions and a reaction slot, the multiple attack penalty with agile, the four degrees of success with natural 20 and 1, damage types, resistance, weakness and immunity with exceptions, critical hits, and dying, wounded and recovery checks. Moved to Effort 3: bonus-type stacking, basic saves, doomed, hero points, Inextinguishable and Too Angry to Die. Nothing in Effort 1 applies them. Rejected: the whole Down row now (doomed has no source until Effort 3); stacking and basic saves now (no Effort 1 rule needs them).
2. **NPCs die at 0 HP; PCs go dying.** Robert's call; the rules default. The creature's `kind` (`pc` or `npc`) decides. Rejected: a per-creature flag (no creature needs it yet); everyone dying (bosses would need recovery checks).
3. **A creature is a folder of layers: `import.toml`, then `hand.toml`, then each variant `<variant>.toml` in the order asked.** Robert's call; settles the file names that overview Decision 6 left open. The importer (Effort 2) owns `import.toml`. At least one of `import.toml` and `hand.toml` must exist. Effort 1 writes Luna's full numbers to `campaigns/hawk/encounters/fall-of-luna/abyssal-luna/hand.toml`, so `show` works now; Effort 2 trims her to logic once the import layer carries the numbers. The dummy PCs and a copy of Luna live in `internal/pf2e/sim/engine/testdata/`, so the gate 3 test doesn't break when Effort 2 trims the campaign file. Rejected: Luna in testdata only (`show` has nothing real to show until Effort 2); layers in a `layers/` subfolder.
4. **The merge rule is one plain rule over keyed tables.** Robert's call. Strikes and abilities are keyed tables (`[strike.slash]`, `[ability.abyssal-frenzy]`). Tables merge key by key, and every other value (numbers, strings, lists) is replaced by the later layer. `ignore = true` marks an ability that the engine and the fixed behaviour skip; it stays on the creature so `show` and Effort 4's `check` can see it. Abilities live inline in the layers. Every layer declares `schema = 1` and is decoded strictly on its own, so an unknown key names its file. Rejected: `[[arrays]]` merged by `id` (a special case beside the plain merge); one file per ability.
5. **The engine reports through one `Observe(Event)` method, with one struct per event kind.** Robert's call; a deliberate break from the faction engine's method-per-event `TurnObserver`. Events are plain values that name combatants by ID, so scoring (Effort 6), replay and the gate 3 test can keep them in a list and compare them. Rejected: a method per event (a recording observer would have to turn every call back into a value).
6. **Until Effort 5, creatures use the `FirstFitting` behaviour.** Robert's call, adjusted by the merge rule. Merged keyed tables have no file order, so "first listed" becomes "first by ID". On each step the creature uses the first non-ignored ability, by ID, whose action cost fits the actions left. If no ability fits, it uses its first Strike by ID. The target is the first standing foe in setup order. A multi-Strike ability whose target falls asks the decider for a new target (`NextTarget`). Rejected: costliest ability first; the lowest-HP foe.
7. **Three packages: `internal/pf2e/sim/dice`, `.../creature` and `.../engine`; the commands go in `cmd/gm-toolkit/pf2e/`.** Robert's call; fills in overview Decision 4. `dice` holds the seeded roller and dice expressions, `creature` the formats, merge, validation and folder lookup, and `engine` the rules, the fight loop, the events and the decision seam. Rejected: the roller inside `engine`; one flat `sim` package.
8. **`campaigns.Paths` gains `PartyDir` (`pf2e/party`) and `EncountersDir` (`encounters`).** Robert's call. `Scaffold` derives folders from `Paths`, so `campaign create` now makes both folders in every new campaign. That is harmless, and git ignores empty folders. Rejected: a scaffold-skip tag; a separate PF2e path bundle.
9. **Initiative and knock-down follow the rules as written (D-18).** Initiative is d20 plus Perception. On a tie, an NPC goes before a PC; within one kind, the setup order holds. A PC reduced to 0 HP moves in the initiative order to directly before the creature that knocked it down. Recovery checks are flat checks (DC 10 plus the dying value), and natural 20 and natural 1 shift their degree like any check. The knock-down move and the natural 20 and 1 on recovery checks change when a recovery check fires; Robert confirmed both at sign-off.
10. **The wiki pages are specified by outline and required claims, not full text; the front door is given in full.** The front door reorganises prose that already exists. The three PF2e pages describe code written in this effort, so D-23 has execution write their prose after checking the code. Their Key Decisions are given here word for word. The `engine/` and `interface/` folders do not move. The front door groups them under the faction manager, and their links stay valid.

## Decision Record — Execution

1. **Commit 6, Task 1: the tree lists `engine/ability-catalog.md`.** The plan's full contents left it out, as the old front door's tree did. The page exists and the front door links it under Design principles, so the tree now names it. The Task 1 contents above are fixed to match.
2. **Commit 7, Task 3: the determinism Key Decision is reworded to match the source.** The plan's text said "nothing iterates a map without sorting first". Two loops range a map unsorted: `strongestAdjustment` (`engine/damage.go`) and `mergeInto` (`creature/layers.go`). Neither order can change a result (the first takes the largest value, the second merges key by key), so gate 3 holds, but the sentence was false. The page now says no map's order reaches a roll or an event, and names where a map is read unsorted. Sorting both loops in code would make the plan's sentence true instead; that is open for the effort close. The Task 3 text above is fixed to match.
3. **Effort close: the review-fix commit changes the plan's code in eight places.** `/code-review high` found ten issues; the picks were made by the session with Robert away and approved by him afterwards. The code blocks above are fixed to match the source.
   - **Lower-case keys** (`readLayer`). The typed decode matches field names without case, the merge does not, so `HP` beside `hp` made the load nondeterministic. Any key with a capital letter now fails the load; a fence-sign marks why. `readLayer` also reads each file once.
   - **AC and adjustment values** (`prepare`). `defense.ac` and each `resist` and `weakness` value must be at least 1. A missing AC loaded as 0.
   - **Immunity groups** (`adjustedDamage`). An `immune` entry now covers types through `adjustmentCovers`, as resistance and weakness do, so `physical` and `all` work. Validating `immune` keys is deferred: Luna's list holds conditions, which Effort 3's condition rules will need to sort out.
   - **Variant names** (`Load`). `--variant ../../x` read outside the creature folder and `--variant hand` applied `hand.toml` twice. A variant must be a plain file name and not `import` or `hand`, or the load fails with `ErrBadVariant`. `Locate` shares the name check (`isFolderName`).
   - **`NoOutcome`** (`Outcome`). `Run` returned `0`, which was `PartyWins`, beside an error. `NoOutcome` is now the zero value, with a fence-sign.
   - **Escape codes in pipes** (`show`). Bold is written only when the output is a terminal file and `NO_COLOR` is unset (`isTerminal`). `go-isatty` moved from an indirect to a direct requirement.
   - **Missing hint** (`resolveCampaign`). The `ErrCampaignRootMissing` hint the other three copies carry was added. Folding the four copies (`faction`, `spatial`, `campaign`, `pf2e`) into one shared helper is out of this effort's scope; it is a Backlog candidate for Robert.
   - **Full words**. `camp` and `dir` are renamed in production code (`pf2e.go`, `show.go`, `layers.go`); tests keep `dir`.
4. **Effort close: review findings deferred, each to the effort that needs it.**
   - **Effort 2** — the same creature ID in two fight folders fails `Locate` with `ErrCreatureAmbiguous`, so `show` cannot reach either; the importer will write such folders. `dice.Parse("3d6 + 14")` fails on the spaces with a misleading message; Foundry exports may hold spaced formulas.
   - **Effort 3** — a weakness is added to a damage part that rolled 0; fix it with the minimum-damage rule.
   - **Effort 6** — `StrikeIDs` and `AbilityIDs` sort on every call and `StandingFoes` allocates on every call; this matters once many fights are played, and `FirstFitting` is test-only until Effort 5. Two copies of one monster cannot be fielded, because a combatant's ID is its folder name.
   - **Entry 2's open point is closed without a change.** Sorting the two unsorted map loops was not picked up: neither order reaches a roll or an event.
5. **Effort close forward check: nothing changes now; each note is for the effort named.**
   - **Effort 2, import target** — `Creature` has no fields for PC build choices, class or spell DCs, attribute modifiers or skills, and layers decode strictly, so the importer cannot write a key before the struct has it. `Load` has no compute step for PC stats.
   - **Effort 3, ability schema** — `Ability.Actions` is limited to 1 to 3 and a `Step` holds only `strike`. Reactions, passives, saves, areas and spells all need new fields. A schema bump is free until Effort 4.
   - **Effort 5, `Choice`** — one `Target` per choice; areas and moves need more.
   - **Effort 6, events** — `DamageDealt` names source and target, so damage share works. No event marks a rescue; "two PCs down at once" must be derived from `DyingChanged`.

---

## Shared Context

### Gate

Every commit passes `go build ./... && go vet ./... && go test ./...` before it lands. Commit 5 also runs the manual check in its last task.

### Files this effort creates

```
internal/pf2e/sim/
├── dice/       dice.go, dice_test.go                          (Commit 1)
├── creature/   creature.go, layers.go, validate.go, tests,
│               testdata/layered/                              (Commit 2)
│               locate.go, locate_test.go                      (Commit 5)
└── engine/     combatant.go, fight.go, events.go, check.go,
                strike.go, damage.go, dying.go, tests          (Commit 3)
                decide.go, run.go, run_test.go, testdata/      (Commit 4)
cmd/gm-toolkit/pf2e/   pf2e.go, sim.go, show.go, show_test.go  (Commit 5)
campaigns/hawk/encounters/fall-of-luna/abyssal-luna/hand.toml  (Commit 5)
docs/architecture/pf2e/   overview.md, data-files.md, engine.md (Commit 7)
```

### Test conventions

- `dice` and `creature` tests are black-box (`package dice_test`, `package creature_test`), as `internal/campaigns` does.
- `engine` tests are white-box (`package engine`). The rules are unexported methods on `Fight`, and the tests drive them with a scripted roller (Commit 3, Task 8).
- No test reads campaign data. `creature` writes its error cases to `t.TempDir()`, and `engine` reads its own `testdata/`.

### Effort close review addition

`/gestalt:effort-close F.041.E1` also checks that Effort 1's shapes do not box in the later efforts. Read each shape against the overview's scope for the effort named:

- **The ability schema** (`actions` limited to 1 to 3; `Step` with only `strike`) against Effort 3's reactions, passives, saves, areas and spells. A schema bump is free until Effort 4 approves its first draft (overview Decision 1).
- **The import target** (the layer format and `creature.Load`) against Effort 2's NPC numbers, tags and PC build choices.
- **`Choice` and the `Decider`** (one target per choice) against Effort 5's areas, moves and rule-list tactics.
- **The event list** against Effort 6's scoring: rescues, two PCs down at once, damage share per source.

A finding that needs a change now goes in the review-fix commit; one that can wait goes in the Decision Record for the effort that will need it.

## Out of Scope

- **Everything Decision 1 moves to Effort 3,** and the rest of the discovery's Rules coverage table: conditions, reactions and triggers, spells, persistent damage and the other Effort 3 rows.
- **The minimum of 1 damage** when penalties would reduce a damage roll below 1. No Effort 1 creature has a penalty or a negative modifier; Effort 3 adds the rule with the first penalty. Until then a negative part is floored at 0.
- **Weapon traits other than agile,** such as deadly, fatal and versatile. They are not in the discovery's Rules coverage table; Effort 2's import report names the ones the PCs carry.
- **Position.** Every creature counts as engaged with every foe. `reach-<n>` traits are loaded and shown but not checked until Effort 5.
- **Saves, Perception beyond initiative, speed and level.** They are loaded and shown; no Effort 1 rule reads them.
- **The reaction slot's use.** `ReactionUsed` resets each turn; nothing spends it until Effort 3's reactions.
- **A command that runs a fight.** `run` is Effort 6's; gate 3 is a Go test.

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — Dice and creature files**
  - Commit 1 — `feat(E1): add seeded PF2e roller and dice expressions` (2 tasks)
  - Commit 2 — `feat(E1): add creature layers, merge and validation` (8 tasks)
- [x] **Phase 2 — Engine core**
  - Commit 3 — `feat(E1): add checks, Strikes, damage and dying rules` (9 tasks)
  - Commit 4 — `feat(E1): add the fight loop, events and fixed behaviour` (9 tasks)
- [x] **Phase 3 — show and the wiki**
  - Commit 5 — `feat(E1): add pf2e sim show and Luna's hand layer` (11 tasks)
  - Commit 6 — `docs(E1): restructure the architecture front door by product` (1 task)
  - Commit 7 — `docs(E1): add the pf2e overview, data-files and engine pages` (5 tasks)

### Phase 1 — Dice and creature files

**In plain words:** adds a dice roller that gives the same rolls for the same seed, and creature files that load from layers merged in order. When it lands, a creature folder loads into one checked stat block, but nothing fights yet.

#### Commit 1 — `feat(E1): add seeded PF2e roller and dice expressions`

##### Task 1 — `internal/pf2e/sim/dice/dice.go`

`internal/pf2e/sim/dice/dice.go` (new) — full contents:

```go
package dice

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

var ErrBadExpression = errors.New("bad dice expression")

type Roller interface {
	Die(sides int) int
}

type Seeded struct {
	source *rand.Rand
}

func NewSeeded(seed uint64) *Seeded {
	return &Seeded{source: rand.New(rand.NewPCG(seed, seed))}
}

func (roller *Seeded) Die(sides int) int {
	return roller.source.IntN(sides) + 1
}

type Expression struct {
	Count    int
	Sides    int
	Modifier int
}

func Parse(text string) (Expression, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return Expression{}, fmt.Errorf("%w: empty", ErrBadExpression)
	}
	countText, rest, hasDice := strings.Cut(trimmed, "d")
	if !hasDice {
		modifier, err := strconv.Atoi(trimmed)
		if err != nil {
			return Expression{}, fmt.Errorf("%w: %q", ErrBadExpression, text)
		}
		return Expression{Modifier: modifier}, nil
	}
	sidesText, modifierText := rest, ""
	if index := strings.IndexAny(rest, "+-"); index >= 0 {
		sidesText, modifierText = rest[:index], rest[index:]
	}
	count, err := strconv.Atoi(countText)
	if err != nil || count < 1 {
		return Expression{}, fmt.Errorf("%w: %q: dice count must be a whole number of at least 1", ErrBadExpression, text)
	}
	sides, err := strconv.Atoi(sidesText)
	if err != nil || sides < 2 {
		return Expression{}, fmt.Errorf("%w: %q: die size must be a whole number of at least 2", ErrBadExpression, text)
	}
	modifier := 0
	if modifierText != "" {
		modifier, err = strconv.Atoi(modifierText)
		if err != nil {
			return Expression{}, fmt.Errorf("%w: %q: bad modifier", ErrBadExpression, text)
		}
	}
	return Expression{Count: count, Sides: sides, Modifier: modifier}, nil
}

func (expression Expression) Roll(roller Roller) int {
	total := expression.Modifier
	for range expression.Count {
		total += roller.Die(expression.Sides)
	}
	return total
}

func (expression Expression) String() string {
	if expression.Count == 0 {
		return strconv.Itoa(expression.Modifier)
	}
	text := fmt.Sprintf("%dd%d", expression.Count, expression.Sides)
	switch {
	case expression.Modifier > 0:
		text += fmt.Sprintf("+%d", expression.Modifier)
	case expression.Modifier < 0:
		text += strconv.Itoa(expression.Modifier)
	}
	return text
}
```

##### Task 2 — `internal/pf2e/sim/dice/dice_test.go`

`internal/pf2e/sim/dice/dice_test.go` (new) — full contents:

```go
package dice_test

import (
	"errors"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

type fixedRoller int

func (value fixedRoller) Die(int) int { return int(value) }

func TestParse(t *testing.T) {
	valid := []struct {
		text string
		want dice.Expression
	}{
		{"3d6+14", dice.Expression{Count: 3, Sides: 6, Modifier: 14}},
		{"1d8", dice.Expression{Count: 1, Sides: 8}},
		{"2d4-1", dice.Expression{Count: 2, Sides: 4, Modifier: -1}},
		{"5", dice.Expression{Modifier: 5}},
		{" 1d20 ", dice.Expression{Count: 1, Sides: 20}},
	}
	for _, testCase := range valid {
		got, err := dice.Parse(testCase.text)
		if err != nil {
			t.Errorf("Parse(%q): %v", testCase.text, err)
			continue
		}
		if got != testCase.want {
			t.Errorf("Parse(%q) = %+v, want %+v", testCase.text, got, testCase.want)
		}
	}
	for _, text := range []string{"", "d6", "0d6", "1d1", "1d", "1d6+", "two", "1d6+x"} {
		if _, err := dice.Parse(text); !errors.Is(err, dice.ErrBadExpression) {
			t.Errorf("Parse(%q) err = %v, want ErrBadExpression", text, err)
		}
	}
}

func TestExpressionStringRoundTrips(t *testing.T) {
	for _, text := range []string{"3d6+14", "1d8", "2d4-1", "5"} {
		expression, err := dice.Parse(text)
		if err != nil {
			t.Fatalf("Parse(%q): %v", text, err)
		}
		if got := expression.String(); got != text {
			t.Errorf("String() = %q, want %q", got, text)
		}
	}
}

func TestRollAddsDiceAndModifier(t *testing.T) {
	expression := dice.Expression{Count: 3, Sides: 6, Modifier: 2}
	if got := expression.Roll(fixedRoller(4)); got != 14 {
		t.Errorf("Roll = %d, want 14", got)
	}
}

func TestSeededRepeatsForTheSameSeed(t *testing.T) {
	first, second, other := dice.NewSeeded(42), dice.NewSeeded(42), dice.NewSeeded(43)
	differs := false
	for range 200 {
		value := first.Die(20)
		if value < 1 || value > 20 {
			t.Fatalf("Die(20) = %d, out of range", value)
		}
		if again := second.Die(20); again != value {
			t.Fatalf("same seed gave %d then %d", value, again)
		}
		if other.Die(20) != value {
			differs = true
		}
	}
	if !differs {
		t.Error("seeds 42 and 43 gave the same 200 rolls")
	}
}
```

##### Commit message

```
feat(E1): add seeded PF2e roller and dice expressions

- dice.Roller interface and dice.Seeded, a PCG roller that repeats for a seed
- dice.Expression with Parse, Roll and String for NdS+M and flat values
```

#### Commit 2 — `feat(E1): add creature layers, merge and validation`

##### Task 1 — `internal/pf2e/sim/creature/creature.go`

`internal/pf2e/sim/creature/creature.go` (new) — full contents:

```go
package creature

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const SchemaVersion = 1

type Kind string

const (
	KindPC  Kind = "pc"
	KindNPC Kind = "npc"
)

type Creature struct {
	ID         string             `toml:"-"`
	Layers     []string           `toml:"-"`
	Schema     int                `toml:"schema"`
	Name       string             `toml:"name"`
	Kind       Kind               `toml:"kind"`
	Level      int                `toml:"level"`
	Traits     []string           `toml:"traits"`
	Perception int                `toml:"perception"`
	Speed      Speed              `toml:"speed"`
	Defense    Defense            `toml:"defense"`
	Strikes    map[string]Strike  `toml:"strike"`
	Abilities  map[string]Ability `toml:"ability"`
}

type Speed struct {
	Land int `toml:"land"`
	Fly  int `toml:"fly"`
}

type Defense struct {
	HP        int                   `toml:"hp"`
	AC        int                   `toml:"ac"`
	Fortitude int                   `toml:"fortitude"`
	Reflex    int                   `toml:"reflex"`
	Will      int                   `toml:"will"`
	Immune    []string              `toml:"immune"`
	Resist    map[string]Adjustment `toml:"resist"`
	Weakness  map[string]Adjustment `toml:"weakness"`
}

type Adjustment struct {
	Value  int      `toml:"value"`
	Except []string `toml:"except"`
}

type Strike struct {
	ID     string       `toml:"-"`
	Name   string       `toml:"name"`
	Bonus  int          `toml:"bonus"`
	Traits []string     `toml:"traits"`
	Damage []DamagePart `toml:"damage"`
}

type DamagePart struct {
	Dice string          `toml:"dice"`
	Type string          `toml:"type"`
	Roll dice.Expression `toml:"-"`
}

type Ability struct {
	ID      string `toml:"-"`
	Name    string `toml:"name"`
	Actions int    `toml:"actions"`
	Ignore  bool   `toml:"ignore"`
	Steps   []Step `toml:"steps"`
}

type Step struct {
	Strike string `toml:"strike"`
}

func (strike Strike) Agile() bool {
	return slices.Contains(strike.Traits, "agile")
}

func (creature *Creature) StrikeIDs() []string {
	return slices.Sorted(maps.Keys(creature.Strikes))
}

func (creature *Creature) AbilityIDs() []string {
	return slices.Sorted(maps.Keys(creature.Abilities))
}

var (
	damageTypes   = []string{"acid", "bleed", "bludgeoning", "cold", "electricity", "fire", "force", "mental", "piercing", "poison", "slashing", "sonic", "spirit", "vitality", "void"}
	physicalTypes = []string{"bludgeoning", "piercing", "slashing"}
)

func IsDamageType(name string) bool {
	return slices.Contains(damageTypes, name)
}

func IsPhysical(damageType string) bool {
	return slices.Contains(physicalTypes, damageType)
}
```

##### Task 2 — `internal/pf2e/sim/creature/layers.go`

`internal/pf2e/sim/creature/layers.go` (new) — full contents:

```go
package creature

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	importLayer = "import.toml"
	handLayer   = "hand.toml"
)

var (
	ErrNoLayers   = errors.New("creature has neither an import nor a hand layer")
	ErrInvalid    = errors.New("invalid creature")
	ErrBadVariant = errors.New("bad variant name")
)

func Load(directory string, variants ...string) (*Creature, error) {
	var names []string
	for _, name := range []string{importLayer, handLayer} {
		_, err := os.Stat(filepath.Join(directory, name))
		if err == nil {
			names = append(names, name)
			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("load creature %s: %w", directory, err)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("load creature %s: %w", directory, ErrNoLayers)
	}
	for _, variant := range variants {
		name := variant + ".toml"
		if !isFolderName(variant) || name == importLayer || name == handLayer {
			return nil, fmt.Errorf("load creature %s: %w: %q", directory, ErrBadVariant, variant)
		}
		names = append(names, name)
	}

	merged := map[string]any{}
	for _, name := range names {
		layer, err := readLayer(filepath.Join(directory, name))
		if err != nil {
			return nil, fmt.Errorf("load creature %s: %w", directory, err)
		}
		mergeInto(merged, layer)
	}

	var buffer bytes.Buffer
	if err := toml.NewEncoder(&buffer).Encode(merged); err != nil {
		return nil, fmt.Errorf("load creature %s: encode merged layers: %w", directory, err)
	}
	var creature Creature
	if _, err := toml.Decode(buffer.String(), &creature); err != nil {
		return nil, fmt.Errorf("load creature %s: decode merged layers: %w", directory, err)
	}
	creature.ID = filepath.Base(directory)
	creature.Layers = names
	if err := creature.prepare(); err != nil {
		return nil, fmt.Errorf("load creature %s: %w", directory, err)
	}
	return &creature, nil
}

func isFolderName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, `/\`)
}

func readLayer(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("layer %s: %w", path, err)
	}
	var typed Creature
	metadata, err := toml.Decode(string(content), &typed)
	if err != nil {
		return nil, fmt.Errorf("layer %s: %w", path, err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		return nil, fmt.Errorf("layer %s: unknown keys: %s", path, joinKeys(undecoded))
	}
	// deliberate: a key with a capital letter is an error, not accepted, because the decoder matches fields without case but the merge does not, so `HP` and `hp` would both survive the merge and either could win.
	var capitalised []toml.Key
	for _, key := range metadata.Keys() {
		if text := key.String(); text != strings.ToLower(text) {
			capitalised = append(capitalised, key)
		}
	}
	if len(capitalised) > 0 {
		return nil, fmt.Errorf("layer %s: keys must be lower case: %s", path, joinKeys(capitalised))
	}
	if typed.Schema != SchemaVersion {
		return nil, fmt.Errorf("layer %s: schema %d, want %d", path, typed.Schema, SchemaVersion)
	}
	var layer map[string]any
	if _, err := toml.Decode(string(content), &layer); err != nil {
		return nil, fmt.Errorf("layer %s: %w", path, err)
	}
	return layer, nil
}

func joinKeys(keys []toml.Key) string {
	texts := make([]string, len(keys))
	for index, key := range keys {
		texts[index] = key.String()
	}
	return strings.Join(texts, ", ")
}

func mergeInto(base, overlay map[string]any) {
	for key, value := range overlay {
		overlayTable, overlayIsTable := value.(map[string]any)
		baseTable, baseIsTable := base[key].(map[string]any)
		if overlayIsTable && baseIsTable {
			mergeInto(baseTable, overlayTable)
			continue
		}
		base[key] = value
	}
}
```

##### Task 3 — `internal/pf2e/sim/creature/validate.go`

`prepare` fills the derived fields (IDs, parsed dice) and collects every problem, so one load reports them all.

`internal/pf2e/sim/creature/validate.go` (new) — full contents:

```go
package creature

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

func (creature *Creature) prepare() error {
	var problems []string
	report := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	if creature.Name == "" {
		report("name is empty")
	}
	if creature.Kind != KindPC && creature.Kind != KindNPC {
		report("kind %q: want pc or npc", creature.Kind)
	}
	if creature.Defense.HP < 1 {
		report("defense.hp must be at least 1")
	}
	if creature.Defense.AC < 1 {
		report("defense.ac must be at least 1")
	}
	for _, table := range []struct {
		name        string
		adjustments map[string]Adjustment
	}{
		{"resist", creature.Defense.Resist},
		{"weakness", creature.Defense.Weakness},
	} {
		for _, key := range slices.Sorted(maps.Keys(table.adjustments)) {
			if key != "all" && key != "physical" && !IsDamageType(key) {
				report("defense.%s.%s: not a damage type, physical or all", table.name, key)
			}
			if table.adjustments[key].Value < 1 {
				report("defense.%s.%s: value must be at least 1", table.name, key)
			}
		}
	}

	for _, id := range creature.StrikeIDs() {
		strike := creature.Strikes[id]
		strike.ID = id
		if strike.Name == "" {
			report("strike.%s: name is empty", id)
		}
		if len(strike.Damage) == 0 {
			report("strike.%s: no damage", id)
		}
		for index := range strike.Damage {
			part := &strike.Damage[index]
			expression, err := dice.Parse(part.Dice)
			if err != nil {
				report("strike.%s.damage[%d]: %v", id, index, err)
			}
			part.Roll = expression
			if !IsDamageType(part.Type) {
				report("strike.%s.damage[%d]: unknown damage type %q", id, index, part.Type)
			}
		}
		creature.Strikes[id] = strike
	}

	for _, id := range creature.AbilityIDs() {
		ability := creature.Abilities[id]
		ability.ID = id
		creature.Abilities[id] = ability
		if ability.Ignore {
			continue
		}
		if ability.Name == "" {
			report("ability.%s: name is empty", id)
		}
		if ability.Actions < 1 || ability.Actions > 3 {
			report("ability.%s: actions %d: want 1 to 3", id, ability.Actions)
		}
		if len(ability.Steps) == 0 {
			report("ability.%s: no steps", id)
		}
		for index, step := range ability.Steps {
			if _, found := creature.Strikes[step.Strike]; !found {
				report("ability.%s.steps[%d]: unknown strike %q", id, index, step.Strike)
			}
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalid, strings.Join(problems, "; "))
	}
	return nil
}
```

##### Task 4 — `internal/pf2e/sim/creature/testdata/layered/import.toml`

`internal/pf2e/sim/creature/testdata/layered/import.toml` (new) — full contents:

```toml
schema = 1

name       = "Layered Beast"
kind       = "npc"
level      = 5
traits     = ["beast"]
perception = 12

[speed]
land = 30

[defense]
hp        = 60
ac        = 22
fortitude = 14
reflex    = 12
will      = 10

[strike.claw]
name   = "Claw"
bonus  = 12
traits = ["agile"]
damage = [{ dice = "2d6+4", type = "slashing" }]

[strike.bite]
name   = "Bite"
bonus  = 12
damage = [{ dice = "2d8+4", type = "piercing" }]

[ability.howl]
name    = "Howl"
actions = 1
steps   = [{ strike = "bite" }]
```

##### Task 5 — `internal/pf2e/sim/creature/testdata/layered/hand.toml`

`internal/pf2e/sim/creature/testdata/layered/hand.toml` (new) — full contents:

```toml
schema = 1

[defense.resist.fire]
value = 5

[strike.claw]
bonus = 14

[ability.rend]
name    = "Rend"
actions = 2
steps   = [{ strike = "claw" }, { strike = "claw" }]

[ability.howl]
ignore = true
```

##### Task 6 — `internal/pf2e/sim/creature/testdata/layered/elite.toml`

`internal/pf2e/sim/creature/testdata/layered/elite.toml` (new) — full contents:

```toml
schema = 1
traits = ["elite"]

[defense]
hp = 80
```

##### Task 7 — `internal/pf2e/sim/creature/layers_test.go`

`internal/pf2e/sim/creature/layers_test.go` (new) — full contents:

```go
package creature_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const minimalHand = `schema = 1
name = "Minimal"
kind = "pc"
level = 1
perception = 3

[defense]
hp = 10
ac = 15

[strike.fist]
name = "Fist"
bonus = 5
damage = [{ dice = "1d4", type = "bludgeoning" }]
`

func writeLayers(t *testing.T, layers map[string]string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "subject")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	for name, content := range layers {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	return dir
}

func TestLoadMergesImportAndHand(t *testing.T) {
	loaded, err := creature.Load("testdata/layered")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ID != "layered" {
		t.Errorf("ID = %q, want layered", loaded.ID)
	}
	if want := []string{"import.toml", "hand.toml"}; !slices.Equal(loaded.Layers, want) {
		t.Errorf("Layers = %v, want %v", loaded.Layers, want)
	}
	claw := loaded.Strikes["claw"]
	if claw.Bonus != 14 {
		t.Errorf("claw bonus = %d, want 14 from the hand layer", claw.Bonus)
	}
	if claw.Name != "Claw" || claw.ID != "claw" {
		t.Errorf("claw name, id = %q, %q, want Claw, claw kept from the import layer", claw.Name, claw.ID)
	}
	if want := (dice.Expression{Count: 2, Sides: 6, Modifier: 4}); claw.Damage[0].Roll != want {
		t.Errorf("claw damage = %+v, want %+v", claw.Damage[0].Roll, want)
	}
	if !claw.Agile() {
		t.Error("claw is not agile")
	}
	if _, found := loaded.Strikes["bite"]; !found {
		t.Error("bite from the import layer is missing")
	}
	if loaded.Defense.HP != 60 {
		t.Errorf("HP = %d, want 60", loaded.Defense.HP)
	}
	if got := loaded.Defense.Resist["fire"].Value; got != 5 {
		t.Errorf("fire resistance = %d, want 5", got)
	}
	if rend := loaded.Abilities["rend"]; rend.ID != "rend" || rend.Actions != 2 || len(rend.Steps) != 2 {
		t.Errorf("rend = %+v, want a 2-action ability with 2 steps", rend)
	}
	if howl := loaded.Abilities["howl"]; !howl.Ignore || howl.Name != "Howl" {
		t.Errorf("howl = %+v, want the import layer's Howl marked ignore", howl)
	}
	if want := []string{"howl", "rend"}; !slices.Equal(loaded.AbilityIDs(), want) {
		t.Errorf("AbilityIDs = %v, want %v", loaded.AbilityIDs(), want)
	}
}

func TestLoadAppliesVariantsLast(t *testing.T) {
	loaded, err := creature.Load("testdata/layered", "elite")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Defense.HP != 80 || loaded.Defense.AC != 22 {
		t.Errorf("HP, AC = %d, %d, want 80, 22", loaded.Defense.HP, loaded.Defense.AC)
	}
	if want := []string{"elite"}; !slices.Equal(loaded.Traits, want) {
		t.Errorf("Traits = %v, want %v (a list is replaced, not merged)", loaded.Traits, want)
	}
	if want := []string{"import.toml", "hand.toml", "elite.toml"}; !slices.Equal(loaded.Layers, want) {
		t.Errorf("Layers = %v, want %v", loaded.Layers, want)
	}
}

func TestLoadHandOnly(t *testing.T) {
	dir := writeLayers(t, map[string]string{"hand.toml": minimalHand})
	loaded, err := creature.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ID != "subject" || !slices.Equal(loaded.Layers, []string{"hand.toml"}) {
		t.Errorf("ID, Layers = %q, %v, want subject, [hand.toml]", loaded.ID, loaded.Layers)
	}
}

func TestLoadErrors(t *testing.T) {
	const flurry = "\n[ability.flurry]\nname = \"Flurry\"\nactions = 2\nsteps = [{ strike = \"kick\" }]\n"
	const slowFlurry = "\n[ability.flurry]\nname = \"Flurry\"\nactions = 4\nsteps = [{ strike = \"fist\" }]\n"
	cases := []struct {
		name     string
		layers   map[string]string
		variants []string
		wantErr  error
		wantText string
	}{
		{name: "no layers", layers: map[string]string{}, wantErr: creature.ErrNoLayers},
		{name: "unknown key", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, "level = 1", "level = 1\nlevle = 2", 1)}, wantText: "unknown keys: levle"},
		{name: "wrong schema", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, "schema = 1", "schema = 2", 1)}, wantText: "schema 2, want 1"},
		{name: "missing variant", layers: map[string]string{"hand.toml": minimalHand}, variants: []string{"elite"}, wantText: "elite.toml"},
		{name: "bad kind", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, `kind = "pc"`, `kind = "boss"`, 1)}, wantErr: creature.ErrInvalid, wantText: `kind "boss"`},
		{name: "bad dice", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, `"1d4"`, `"1d"`, 1)}, wantErr: creature.ErrInvalid, wantText: "strike.fist.damage[0]"},
		{name: "bad damage type", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, "bludgeoning", "blunt", 1)}, wantErr: creature.ErrInvalid, wantText: `unknown damage type "blunt"`},
		{name: "unknown step strike", layers: map[string]string{"hand.toml": minimalHand + flurry}, wantErr: creature.ErrInvalid, wantText: `unknown strike "kick"`},
		{name: "too many actions", layers: map[string]string{"hand.toml": minimalHand + slowFlurry}, wantErr: creature.ErrInvalid, wantText: "actions 4"},
		{name: "bad resist key", layers: map[string]string{"hand.toml": minimalHand + "\n[defense.resist.blunt]\nvalue = 5\n"}, wantErr: creature.ErrInvalid, wantText: "defense.resist.blunt"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := writeLayers(t, testCase.layers)
			_, err := creature.Load(dir, testCase.variants...)
			if err == nil {
				t.Fatal("Load succeeded, want an error")
			}
			if testCase.wantErr != nil && !errors.Is(err, testCase.wantErr) {
				t.Errorf("err = %v, want %v", err, testCase.wantErr)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("err = %v, want it to mention %q", err, testCase.wantText)
			}
		})
	}
}
```

##### Task 8 — `dev_journals/campaigns/initiatives/implementation/F.041.E1-pf2e-encounter-simulator-plan.md`

Tick Phase 1 in this plan's Overview.

**Find:**
```markdown
- [ ] **Phase 1 — Dice and creature files**
```
**Replace with:**
```markdown
- [x] **Phase 1 — Dice and creature files**
```

##### Commit message

```
feat(E1): add creature layers, merge and validation

- creature.Creature: the stat block, Strikes and abilities as keyed tables
- creature.Load merges import.toml, hand.toml and variant layers in order
- tables merge key by key; other values are replaced by the later layer
- each layer must declare schema = 1 and is decoded strictly on its own
- validation parses damage dice and reports every problem at once
```

### Phase 2 — Engine core

**In plain words:** adds the rules and the fight loop. When it lands, Luna and three dummy PCs play a full fight from files, and the same seed gives the same events every time (gate 3).

#### Commit 3 — `feat(E1): add checks, Strikes, damage and dying rules`

##### Task 1 — `internal/pf2e/sim/engine/combatant.go`

`internal/pf2e/sim/engine/combatant.go` (new) — full contents:

```go
package engine

import "github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"

type Side int

const (
	SideParty Side = iota
	SideEnemies
)

func (side Side) String() string {
	if side == SideParty {
		return "party"
	}
	return "enemies"
}

type Combatant struct {
	ID              string
	Side            Side
	Creature        *creature.Creature
	HP              int
	Dying           int
	Wounded         int
	Dead            bool
	Initiative      int
	ActionsLeft     int
	AttacksThisTurn int
	ReactionUsed    bool
}

func NewCombatant(side Side, source *creature.Creature) *Combatant {
	return &Combatant{ID: source.ID, Side: side, Creature: source, HP: source.Defense.HP}
}

func (combatant *Combatant) Standing() bool {
	return !combatant.Dead && combatant.HP > 0
}
```

##### Task 2 — `internal/pf2e/sim/engine/fight.go`

`internal/pf2e/sim/engine/fight.go` (new) — full contents. Commit 4 adds the decider and round limit to the struct.

```go
package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

type Fight struct {
	roller     dice.Roller
	observer   Observer
	combatants []*Combatant
	order      []*Combatant
	round      int
}

func (fight *Fight) moveBefore(moved, anchor *Combatant) {
	if moved == anchor {
		return
	}
	from := slices.Index(fight.order, moved)
	fight.order = slices.Delete(fight.order, from, from+1)
	to := slices.Index(fight.order, anchor)
	fight.order = slices.Insert(fight.order, to, moved)
	fight.observer.Observe(InitiativeMoved{Combatant: moved.ID, Before: anchor.ID})
}
```

##### Task 3 — `internal/pf2e/sim/engine/events.go`

All event kinds, Commit 4's included, so the event list has one home.

`internal/pf2e/sim/engine/events.go` (new) — full contents:

```go
package engine

type Observer interface {
	Observe(event Event)
}

type Event interface {
	isEvent()
}

type Outcome int

// deliberate: the zero value is NoOutcome, not PartyWins, because Run returns it beside an error and a tally must not count a failed fight as a win.
const (
	NoOutcome Outcome = iota
	PartyWins
	EnemiesWin
	Stalemate
)

func (outcome Outcome) String() string {
	switch outcome {
	case PartyWins:
		return "party wins"
	case EnemiesWin:
		return "enemies win"
	case Stalemate:
		return "stalemate"
	default:
		return "no outcome"
	}
}

type InitiativeRolled struct {
	Combatant string
	Die       int
	Total     int
}

type RoundStarted struct {
	Round int
}

type TurnStarted struct {
	Round     int
	Combatant string
}

type ActionUsed struct {
	Combatant string
	Action    string
	Actions   int
	Target    string
}

type StrikeRolled struct {
	Attacker string
	Target   string
	Strike   string
	Die      int
	Bonus    int
	Penalty  int
	Total    int
	AC       int
	Degree   Degree
}

type DamageAmount struct {
	Type   string
	Rolled int
	Taken  int
}

type DamageDealt struct {
	Source   string
	Target   string
	Strike   string
	Critical bool
	Parts    []DamageAmount
	Total    int
	HPAfter  int
}

type InitiativeMoved struct {
	Combatant string
	Before    string
}

type DyingChanged struct {
	Combatant string
	From      int
	To        int
}

type RecoveryChecked struct {
	Combatant string
	Die       int
	DC        int
	Degree    Degree
}

type Stabilized struct {
	Combatant string
	Wounded   int
}

type Died struct {
	Combatant string
}

type FightEnded struct {
	Rounds  int
	Outcome Outcome
}

func (InitiativeRolled) isEvent() {}
func (RoundStarted) isEvent() {}
func (TurnStarted) isEvent() {}
func (ActionUsed) isEvent() {}
func (StrikeRolled) isEvent() {}
func (DamageDealt) isEvent() {}
func (InitiativeMoved) isEvent() {}
func (DyingChanged) isEvent() {}
func (RecoveryChecked) isEvent() {}
func (Stabilized) isEvent() {}
func (Died) isEvent() {}
func (FightEnded) isEvent() {}
```

##### Task 4 — `internal/pf2e/sim/engine/check.go`

`internal/pf2e/sim/engine/check.go` (new) — full contents:

```go
package engine

type Degree int

const (
	CriticalFailure Degree = iota
	Failure
	Success
	CriticalSuccess
)

func (degree Degree) String() string {
	switch degree {
	case CriticalFailure:
		return "critical failure"
	case Failure:
		return "failure"
	case Success:
		return "success"
	default:
		return "critical success"
	}
}

func DegreeOf(die, total, dc int) Degree {
	var degree Degree
	switch {
	case total >= dc+10:
		degree = CriticalSuccess
	case total >= dc:
		degree = Success
	case total <= dc-10:
		degree = CriticalFailure
	default:
		degree = Failure
	}
	switch die {
	case 20:
		degree = min(degree+1, CriticalSuccess)
	case 1:
		degree = max(degree-1, CriticalFailure)
	}
	return degree
}
```

##### Task 5 — `internal/pf2e/sim/engine/strike.go`

`internal/pf2e/sim/engine/strike.go` (new) — full contents:

```go
package engine

import "github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"

func attackPenalty(attacksMade int, agile bool) int {
	switch {
	case attacksMade == 0:
		return 0
	case attacksMade == 1 && agile:
		return -4
	case attacksMade == 1:
		return -5
	case agile:
		return -8
	default:
		return -10
	}
}

func (fight *Fight) strike(attacker, target *Combatant, strike creature.Strike) {
	penalty := attackPenalty(attacker.AttacksThisTurn, strike.Agile())
	attacker.AttacksThisTurn++
	die := fight.roller.Die(20)
	total := die + strike.Bonus + penalty
	armorClass := target.Creature.Defense.AC
	degree := DegreeOf(die, total, armorClass)
	fight.observer.Observe(StrikeRolled{
		Attacker: attacker.ID,
		Target:   target.ID,
		Strike:   strike.ID,
		Die:      die,
		Bonus:    strike.Bonus,
		Penalty:  penalty,
		Total:    total,
		AC:       armorClass,
		Degree:   degree,
	})
	if degree < Success {
		return
	}
	fight.dealDamage(attacker, target, strike, degree == CriticalSuccess)
}
```

##### Task 6 — `internal/pf2e/sim/engine/damage.go`

Damage is totalled per type first, then immunity, the highest weakness and the highest resistance apply to each type, in that order.

`internal/pf2e/sim/engine/damage.go` (new) — full contents:

```go
package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func (fight *Fight) dealDamage(attacker, target *Combatant, strike creature.Strike, critical bool) {
	rolledByType := map[string]int{}
	var types []string
	for _, part := range strike.Damage {
		amount := max(part.Roll.Roll(fight.roller), 0)
		if critical {
			amount *= 2
		}
		if _, seen := rolledByType[part.Type]; !seen {
			types = append(types, part.Type)
		}
		rolledByType[part.Type] += amount
	}

	parts := make([]DamageAmount, 0, len(types))
	total := 0
	for _, damageType := range types {
		rolled := rolledByType[damageType]
		taken := adjustedDamage(target.Creature.Defense, damageType, strike.Traits, rolled)
		parts = append(parts, DamageAmount{Type: damageType, Rolled: rolled, Taken: taken})
		total += taken
	}

	wasStanding := target.Standing()
	target.HP = max(target.HP-total, 0)
	fight.observer.Observe(DamageDealt{
		Source:   attacker.ID,
		Target:   target.ID,
		Strike:   strike.ID,
		Critical: critical,
		Parts:    parts,
		Total:    total,
		HPAfter:  target.HP,
	})
	if total == 0 {
		return
	}
	switch {
	case wasStanding && target.HP == 0:
		fight.knockDown(target, attacker, critical)
	case !wasStanding:
		fight.hitWhileDown(target, critical)
	}
}

func adjustedDamage(defense creature.Defense, damageType string, traits []string, amount int) int {
	immune := slices.ContainsFunc(defense.Immune, func(key string) bool { return adjustmentCovers(key, damageType) })
	if immune {
		return 0
	}
	if weakness, found := strongestAdjustment(defense.Weakness, damageType, traits); found {
		amount += weakness
	}
	if resistance, found := strongestAdjustment(defense.Resist, damageType, traits); found {
		amount -= resistance
	}
	return max(amount, 0)
}

func strongestAdjustment(adjustments map[string]creature.Adjustment, damageType string, traits []string) (int, bool) {
	strongest, found := 0, false
	for key, adjustment := range adjustments {
		if !adjustmentCovers(key, damageType) {
			continue
		}
		excepted := slices.Contains(adjustment.Except, damageType) ||
			slices.ContainsFunc(traits, func(trait string) bool { return slices.Contains(adjustment.Except, trait) })
		if excepted {
			continue
		}
		if !found || adjustment.Value > strongest {
			strongest, found = adjustment.Value, true
		}
	}
	return strongest, found
}

func adjustmentCovers(key, damageType string) bool {
	switch key {
	case "all":
		return true
	case "physical":
		return creature.IsPhysical(damageType)
	default:
		return key == damageType
	}
}
```

##### Task 7 — `internal/pf2e/sim/engine/dying.go`

`internal/pf2e/sim/engine/dying.go` (new) — full contents:

```go
package engine

import "github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"

const deathThreshold = 4

func (fight *Fight) knockDown(target, attacker *Combatant, critical bool) {
	if target.Creature.Kind == creature.KindNPC {
		fight.kill(target)
		return
	}
	fight.moveBefore(target, attacker)
	fight.setDying(target, dyingOnKnockDown(target, critical))
}

func (fight *Fight) hitWhileDown(target *Combatant, critical bool) {
	if target.Dead {
		return
	}
	if target.Dying == 0 {
		fight.setDying(target, dyingOnKnockDown(target, critical))
		return
	}
	increase := 1
	if critical {
		increase = 2
	}
	fight.setDying(target, target.Dying+increase)
}

func dyingOnKnockDown(target *Combatant, critical bool) int {
	value := 1 + target.Wounded
	if critical {
		value++
	}
	return value
}

func (fight *Fight) setDying(target *Combatant, value int) {
	from := target.Dying
	target.Dying = min(value, deathThreshold)
	fight.observer.Observe(DyingChanged{Combatant: target.ID, From: from, To: target.Dying})
	if target.Dying >= deathThreshold {
		fight.kill(target)
	}
}

func (fight *Fight) kill(target *Combatant) {
	target.Dead = true
	fight.observer.Observe(Died{Combatant: target.ID})
}

func (fight *Fight) recoveryCheck(target *Combatant) {
	dc := 10 + target.Dying
	die := fight.roller.Die(20)
	degree := DegreeOf(die, die, dc)
	fight.observer.Observe(RecoveryChecked{Combatant: target.ID, Die: die, DC: dc, Degree: degree})

	change := 0
	switch degree {
	case CriticalSuccess:
		change = -2
	case Success:
		change = -1
	case Failure:
		change = 1
	case CriticalFailure:
		change = 2
	}
	next := max(target.Dying+change, 0)
	if next > 0 {
		fight.setDying(target, next)
		return
	}
	fight.setDying(target, 0)
	target.Wounded++
	fight.observer.Observe(Stabilized{Combatant: target.ID, Wounded: target.Wounded})
}
```

##### Task 8 — `internal/pf2e/sim/engine/helpers_test.go`

`internal/pf2e/sim/engine/helpers_test.go` (new) — full contents:

```go
package engine

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

type scriptedRoller struct {
	t      *testing.T
	values []int
}

func script(t *testing.T, values ...int) *scriptedRoller {
	return &scriptedRoller{t: t, values: values}
}

func (roller *scriptedRoller) Die(sides int) int {
	roller.t.Helper()
	if len(roller.values) == 0 {
		roller.t.Fatalf("scripted roller ran out on a d%d", sides)
	}
	value := roller.values[0]
	roller.values = roller.values[1:]
	if value < 1 || value > sides {
		roller.t.Fatalf("scripted value %d is not a d%d result", value, sides)
	}
	return value
}

type recorder struct {
	events []Event
}

func (recorder *recorder) Observe(event Event) {
	recorder.events = append(recorder.events, event)
}

func fist() creature.Strike {
	return creature.Strike{
		ID:    "fist",
		Name:  "Fist",
		Bonus: 10,
		Damage: []creature.DamagePart{
			{Dice: "1d6+2", Type: "bludgeoning", Roll: dice.Expression{Count: 1, Sides: 6, Modifier: 2}},
		},
	}
}

func testCombatant(id string, side Side, kind creature.Kind, hp, armorClass int) *Combatant {
	source := &creature.Creature{
		ID:         id,
		Name:       id,
		Kind:       kind,
		Level:      1,
		Perception: 5,
		Defense:    creature.Defense{HP: hp, AC: armorClass},
		Strikes:    map[string]creature.Strike{"fist": fist()},
	}
	return NewCombatant(side, source)
}

func testFight(roller dice.Roller, combatants ...*Combatant) (*Fight, *recorder) {
	events := &recorder{}
	fight := &Fight{
		roller:     roller,
		observer:   events,
		combatants: combatants,
		order:      slices.Clone(combatants),
	}
	return fight, events
}

func orderIDs(fight *Fight) []string {
	ids := make([]string, len(fight.order))
	for index, combatant := range fight.order {
		ids[index] = combatant.ID
	}
	return ids
}
```

##### Task 9 — `internal/pf2e/sim/engine/rules_test.go`

`internal/pf2e/sim/engine/rules_test.go` (new) — full contents:

```go
package engine

import (
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func TestDegreeOf(t *testing.T) {
	cases := []struct {
		name             string
		die, total, dc   int
		want             Degree
	}{
		{"beat by 10", 12, 30, 20, CriticalSuccess},
		{"meet the DC", 10, 20, 20, Success},
		{"miss by 1", 10, 19, 20, Failure},
		{"miss by 10", 5, 10, 20, CriticalFailure},
		{"natural 20 lifts a failure", 20, 25, 30, Success},
		{"natural 20 stops at critical success", 20, 40, 20, CriticalSuccess},
		{"natural 1 drops a success", 1, 21, 20, Failure},
		{"natural 1 stops at critical failure", 1, 5, 20, CriticalFailure},
	}
	for _, testCase := range cases {
		if got := DegreeOf(testCase.die, testCase.total, testCase.dc); got != testCase.want {
			t.Errorf("%s: DegreeOf(%d, %d, %d) = %v, want %v", testCase.name, testCase.die, testCase.total, testCase.dc, got, testCase.want)
		}
	}
}

func TestAttackPenalty(t *testing.T) {
	cases := []struct {
		attacksMade int
		agile       bool
		want        int
	}{
		{0, false, 0}, {1, false, -5}, {2, false, -10}, {3, false, -10},
		{0, true, 0}, {1, true, -4}, {2, true, -8},
	}
	for _, testCase := range cases {
		if got := attackPenalty(testCase.attacksMade, testCase.agile); got != testCase.want {
			t.Errorf("attackPenalty(%d, %v) = %d, want %d", testCase.attacksMade, testCase.agile, got, testCase.want)
		}
	}
}

func TestStrikeMissDealsNoDamage(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 30, 22)
	fight, events := testFight(script(t, 5), attacker, target)

	fight.strike(attacker, target, fist())

	want := []Event{StrikeRolled{Attacker: "attacker", Target: "target", Strike: "fist", Die: 5, Bonus: 10, Total: 15, AC: 22, Degree: Failure}}
	if !reflect.DeepEqual(events.events, want) {
		t.Errorf("events = %+v, want %+v", events.events, want)
	}
	if target.HP != 30 {
		t.Errorf("HP = %d, want 30", target.HP)
	}
}

func TestStrikeCriticalDoublesDamage(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 30, 15)
	fight, events := testFight(script(t, 15, 4), attacker, target)

	fight.strike(attacker, target, fist())

	want := []Event{
		StrikeRolled{Attacker: "attacker", Target: "target", Strike: "fist", Die: 15, Bonus: 10, Total: 25, AC: 15, Degree: CriticalSuccess},
		DamageDealt{Source: "attacker", Target: "target", Strike: "fist", Critical: true, Parts: []DamageAmount{{Type: "bludgeoning", Rolled: 12, Taken: 12}}, Total: 12, HPAfter: 18},
	}
	if !reflect.DeepEqual(events.events, want) {
		t.Errorf("events = %+v, want %+v", events.events, want)
	}
}

func TestSecondAttackTakesPenalty(t *testing.T) {
	attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
	target := testCombatant("target", SideParty, creature.KindPC, 30, 40)
	fight, events := testFight(script(t, 10, 10), attacker, target)

	fight.strike(attacker, target, fist())
	fight.strike(attacker, target, fist())

	second := events.events[1].(StrikeRolled)
	if second.Penalty != -5 || second.Total != 15 {
		t.Errorf("second strike penalty, total = %d, %d, want -5, 15", second.Penalty, second.Total)
	}
	if attacker.AttacksThisTurn != 2 {
		t.Errorf("AttacksThisTurn = %d, want 2", attacker.AttacksThisTurn)
	}
}

func TestAdjustedDamage(t *testing.T) {
	defense := creature.Defense{
		Immune: []string{"poison"},
		Resist: map[string]creature.Adjustment{
			"physical": {Value: 5},
			"slashing": {Value: 10},
			"all":      {Value: 2, Except: []string{"holy", "fire"}},
		},
		Weakness: map[string]creature.Adjustment{"cold": {Value: 7}},
	}
	cases := []struct {
		name       string
		damageType string
		traits     []string
		amount     int
		want       int
	}{
		{"immunity", "poison", nil, 20, 0},
		{"only the highest resistance", "slashing", nil, 20, 10},
		{"physical covers piercing", "piercing", nil, 20, 15},
		{"weakness then resistance", "cold", nil, 20, 25},
		{"exception by damage type", "fire", nil, 20, 20},
		{"exception by strike trait", "mental", []string{"holy"}, 20, 20},
		{"resistance stops at zero", "mental", nil, 1, 0},
	}
	for _, testCase := range cases {
		if got := adjustedDamage(defense, testCase.damageType, testCase.traits, testCase.amount); got != testCase.want {
			t.Errorf("%s: adjustedDamage = %d, want %d", testCase.name, got, testCase.want)
		}
	}
}

func TestNPCDiesAtZero(t *testing.T) {
	attacker := testCombatant("attacker", SideParty, creature.KindPC, 50, 20)
	target := testCombatant("target", SideEnemies, creature.KindNPC, 5, 20)
	fight, events := testFight(script(t, 15, 6), attacker, target)

	fight.strike(attacker, target, fist())

	if !target.Dead || target.Dying != 0 {
		t.Errorf("Dead, Dying = %v, %d, want true, 0", target.Dead, target.Dying)
	}
	if last := events.events[len(events.events)-1]; last != (Died{Combatant: "target"}) {
		t.Errorf("last event = %+v, want Died", last)
	}
}

func TestPCKnockedDownGoesDying(t *testing.T) {
	cases := []struct {
		name      string
		wounded   int
		die       int
		wantDying int
	}{
		{"hit", 0, 12, 1},
		{"critical hit", 0, 20, 2},
		{"hit while wounded 2", 2, 12, 3},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			target := testCombatant("target", SideParty, creature.KindPC, 5, 20)
			target.Wounded = testCase.wounded
			other := testCombatant("other", SideParty, creature.KindPC, 30, 20)
			attacker := testCombatant("attacker", SideEnemies, creature.KindNPC, 50, 20)
			fight, events := testFight(script(t, testCase.die, 6), target, other, attacker)

			fight.strike(attacker, target, fist())

			if target.Dying != testCase.wantDying || target.Dead {
				t.Errorf("Dying, Dead = %d, %v, want %d, false", target.Dying, target.Dead, testCase.wantDying)
			}
			if want := []string{"other", "target", "attacker"}; !slices.Equal(orderIDs(fight), want) {
				t.Errorf("order = %v, want %v", orderIDs(fight), want)
			}
			if !slices.Contains(events.events, Event(InitiativeMoved{Combatant: "target", Before: "attacker"})) {
				t.Errorf("events = %+v, want an InitiativeMoved", events.events)
			}
		})
	}
}

func TestDamageWhileDown(t *testing.T) {
	cases := []struct {
		name      string
		dying     int
		wounded   int
		critical  bool
		wantDying int
		wantDead  bool
	}{
		{"hit while dying 1", 1, 0, false, 2, false},
		{"critical while dying 1", 1, 0, true, 3, false},
		{"hit while dying 3", 3, 0, false, 4, true},
		{"hit while stable and wounded 1", 0, 1, false, 2, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			target := testCombatant("target", SideParty, creature.KindPC, 30, 20)
			target.HP, target.Dying, target.Wounded = 0, testCase.dying, testCase.wounded
			fight, _ := testFight(script(t), target)

			fight.hitWhileDown(target, testCase.critical)

			if target.Dying != testCase.wantDying || target.Dead != testCase.wantDead {
				t.Errorf("Dying, Dead = %d, %v, want %d, %v", target.Dying, target.Dead, testCase.wantDying, testCase.wantDead)
			}
		})
	}
}

func TestRecoveryCheck(t *testing.T) {
	cases := []struct {
		name        string
		dying       int
		die         int
		wantDying   int
		wantWounded int
		wantDead    bool
	}{
		{"natural 20 from dying 1", 1, 20, 0, 1, false},
		{"success from dying 1", 1, 12, 0, 1, false},
		{"success from dying 2", 2, 12, 1, 0, false},
		{"failure from dying 1", 1, 5, 2, 0, false},
		{"natural 1 from dying 1", 1, 1, 3, 0, false},
		{"natural 1 from dying 3", 3, 1, 4, 0, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			target := testCombatant("target", SideParty, creature.KindPC, 30, 20)
			target.HP, target.Dying = 0, testCase.dying
			fight, _ := testFight(script(t, testCase.die), target)

			fight.recoveryCheck(target)

			if target.Dying != testCase.wantDying || target.Wounded != testCase.wantWounded || target.Dead != testCase.wantDead {
				t.Errorf("Dying, Wounded, Dead = %d, %d, %v, want %d, %d, %v",
					target.Dying, target.Wounded, target.Dead, testCase.wantDying, testCase.wantWounded, testCase.wantDead)
			}
		})
	}
}
```

##### Commit message

```
feat(E1): add checks, Strikes, damage and dying rules

- engine.Fight with the roller, the observer and the initiative order
- events.go: Observe(Event) and one struct per event kind
- DegreeOf: four degrees, natural 20 and 1
- Strikes with the multiple attack penalty (agile -4/-8)
- damage per type with immunity, then the highest weakness and resistance,
  exceptions by damage type or Strike trait; critical hits double
- NPCs die at 0 HP; PCs go dying (1 + wounded, +1 on a critical) and move
  to just before the attacker in initiative
- damage while down raises dying; dying 4 kills; recovery checks stabilize
  and add wounded
```

#### Commit 4 — `feat(E1): add the fight loop, events and fixed behaviour`

##### Task 1 — `internal/pf2e/sim/engine/fight.go`

Add the decider and the round limit to the struct (the struct from Commit 3, Task 2).

**Find:**
```go
type Fight struct {
	roller     dice.Roller
	observer   Observer
	combatants []*Combatant
	order      []*Combatant
	round      int
}
```
**Replace with:**
```go
type Fight struct {
	roller     dice.Roller
	observer   Observer
	decider    Decider
	maxRounds  int
	combatants []*Combatant
	order      []*Combatant
	round      int
}

func (fight *Fight) StandingFoes(actor *Combatant) []*Combatant {
	var foes []*Combatant
	for _, combatant := range fight.combatants {
		if combatant.Side != actor.Side && combatant.Standing() {
			foes = append(foes, combatant)
		}
	}
	return foes
}
```

##### Task 2 — `internal/pf2e/sim/engine/decide.go`

The decision seam. Effort 5's rule-list tactics become a second `Decider`.

`internal/pf2e/sim/engine/decide.go` (new) — full contents:

```go
package engine

type ChoiceKind int

const (
	ChooseStrike ChoiceKind = iota
	ChooseAbility
)

type Choice struct {
	Kind   ChoiceKind
	ID     string
	Target *Combatant
}

type Decider interface {
	Decide(fight *Fight, actor *Combatant) (Choice, bool)
	NextTarget(fight *Fight, actor *Combatant) (*Combatant, bool)
}

// deliberate: abilities and Strikes are tried in ID order, not file order, because merged keyed tables keep no order (E1 Decision 6).
type FirstFitting struct{}

func (behaviour FirstFitting) Decide(fight *Fight, actor *Combatant) (Choice, bool) {
	target, found := behaviour.NextTarget(fight, actor)
	if !found {
		return Choice{}, false
	}
	for _, id := range actor.Creature.AbilityIDs() {
		ability := actor.Creature.Abilities[id]
		if !ability.Ignore && ability.Actions <= actor.ActionsLeft {
			return Choice{Kind: ChooseAbility, ID: id, Target: target}, true
		}
	}
	strikeIDs := actor.Creature.StrikeIDs()
	if len(strikeIDs) == 0 {
		return Choice{}, false
	}
	return Choice{Kind: ChooseStrike, ID: strikeIDs[0], Target: target}, true
}

func (FirstFitting) NextTarget(fight *Fight, actor *Combatant) (*Combatant, bool) {
	foes := fight.StandingFoes(actor)
	if len(foes) == 0 {
		return nil, false
	}
	return foes[0], true
}
```

##### Task 3 — `internal/pf2e/sim/engine/run.go`

`internal/pf2e/sim/engine/run.go` (new) — full contents:

```go
package engine

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const actionsPerTurn = 3

var (
	ErrDuplicateCombatant = errors.New("duplicate combatant id")
	ErrBadSetup           = errors.New("bad fight setup")
	ErrBadChoice          = errors.New("bad choice")
)

type Setup struct {
	Party     []*creature.Creature
	Enemies   []*creature.Creature
	Roller    dice.Roller
	Observer  Observer
	Decider   Decider
	MaxRounds int
}

func Run(setup Setup) (Outcome, error) {
	fight, err := newFight(setup)
	if err != nil {
		return NoOutcome, err
	}
	return fight.run()
}

func newFight(setup Setup) (*Fight, error) {
	switch {
	case len(setup.Party) == 0 || len(setup.Enemies) == 0:
		return nil, fmt.Errorf("%w: a fight needs at least one party member and one enemy", ErrBadSetup)
	case setup.MaxRounds < 1:
		return nil, fmt.Errorf("%w: max rounds %d: want at least 1", ErrBadSetup, setup.MaxRounds)
	case setup.Roller == nil || setup.Observer == nil || setup.Decider == nil:
		return nil, fmt.Errorf("%w: roller, observer and decider are all required", ErrBadSetup)
	}
	fight := &Fight{
		roller:    setup.Roller,
		observer:  setup.Observer,
		decider:   setup.Decider,
		maxRounds: setup.MaxRounds,
	}
	seen := map[string]bool{}
	for _, group := range []struct {
		side      Side
		creatures []*creature.Creature
	}{
		{SideParty, setup.Party},
		{SideEnemies, setup.Enemies},
	} {
		for _, source := range group.creatures {
			if seen[source.ID] {
				return nil, fmt.Errorf("%w: %s", ErrDuplicateCombatant, source.ID)
			}
			seen[source.ID] = true
			fight.combatants = append(fight.combatants, NewCombatant(group.side, source))
		}
	}
	fight.rollInitiative()
	return fight, nil
}

func (fight *Fight) rollInitiative() {
	for _, combatant := range fight.combatants {
		die := fight.roller.Die(20)
		combatant.Initiative = die + combatant.Creature.Perception
		fight.observer.Observe(InitiativeRolled{Combatant: combatant.ID, Die: die, Total: combatant.Initiative})
	}
	fight.order = slices.Clone(fight.combatants)
	slices.SortStableFunc(fight.order, func(first, second *Combatant) int {
		return cmp.Or(
			cmp.Compare(second.Initiative, first.Initiative),
			cmp.Compare(tieRank(first), tieRank(second)),
		)
	})
}

func tieRank(combatant *Combatant) int {
	if combatant.Creature.Kind == creature.KindNPC {
		return 0
	}
	return 1
}

func (fight *Fight) run() (Outcome, error) {
	for fight.round = 1; fight.round <= fight.maxRounds; fight.round++ {
		fight.observer.Observe(RoundStarted{Round: fight.round})
		for index := 0; index < len(fight.order); {
			actor := fight.order[index]
			if err := fight.takeTurn(actor); err != nil {
				return NoOutcome, err
			}
			if outcome, over := fight.outcome(); over {
				fight.observer.Observe(FightEnded{Rounds: fight.round, Outcome: outcome})
				return outcome, nil
			}
			// deliberate: the next actor is found from the actor's current place, not the old index, because a knock-down moves the target before the attacker mid-round.
			index = slices.Index(fight.order, actor) + 1
		}
	}
	fight.observer.Observe(FightEnded{Rounds: fight.maxRounds, Outcome: Stalemate})
	return Stalemate, nil
}

func (fight *Fight) takeTurn(actor *Combatant) error {
	if actor.Dead {
		return nil
	}
	fight.observer.Observe(TurnStarted{Round: fight.round, Combatant: actor.ID})
	if actor.Dying > 0 {
		fight.recoveryCheck(actor)
		return nil
	}
	if !actor.Standing() {
		return nil
	}
	actor.ActionsLeft = actionsPerTurn
	actor.AttacksThisTurn = 0
	actor.ReactionUsed = false
	for actor.ActionsLeft > 0 {
		choice, acting := fight.decider.Decide(fight, actor)
		if !acting {
			return nil
		}
		if err := fight.perform(actor, choice); err != nil {
			return err
		}
		if _, over := fight.outcome(); over {
			return nil
		}
	}
	return nil
}

func (fight *Fight) perform(actor *Combatant, choice Choice) error {
	if choice.Target == nil || !choice.Target.Standing() {
		return fmt.Errorf("%w: %s chose %q with no standing target", ErrBadChoice, actor.ID, choice.ID)
	}
	switch choice.Kind {
	case ChooseStrike:
		strike, found := actor.Creature.Strikes[choice.ID]
		if !found {
			return fmt.Errorf("%w: %s has no strike %q", ErrBadChoice, actor.ID, choice.ID)
		}
		actor.ActionsLeft--
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: 1, Target: choice.Target.ID})
		fight.strike(actor, choice.Target, strike)
		return nil
	case ChooseAbility:
		ability, found := actor.Creature.Abilities[choice.ID]
		if !found || ability.Ignore {
			return fmt.Errorf("%w: %s has no usable ability %q", ErrBadChoice, actor.ID, choice.ID)
		}
		if ability.Actions > actor.ActionsLeft {
			return fmt.Errorf("%w: %s chose %s (%d actions) with %d left", ErrBadChoice, actor.ID, choice.ID, ability.Actions, actor.ActionsLeft)
		}
		actor.ActionsLeft -= ability.Actions
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: ability.Actions, Target: choice.Target.ID})
		target := choice.Target
		for _, step := range ability.Steps {
			if !target.Standing() {
				next, found := fight.decider.NextTarget(fight, actor)
				if !found {
					return nil
				}
				target = next
			}
			fight.strike(actor, target, actor.Creature.Strikes[step.Strike])
		}
		return nil
	default:
		return fmt.Errorf("%w: %s made a choice of unknown kind %d", ErrBadChoice, actor.ID, choice.Kind)
	}
}

func (fight *Fight) outcome() (Outcome, bool) {
	partyStanding, enemiesStanding := false, false
	for _, combatant := range fight.combatants {
		if !combatant.Standing() {
			continue
		}
		if combatant.Side == SideParty {
			partyStanding = true
		} else {
			enemiesStanding = true
		}
	}
	switch {
	case !enemiesStanding:
		return PartyWins, true
	case !partyStanding:
		return EnemiesWin, true
	default:
		return NoOutcome, false
	}
}
```

##### Task 4 — `internal/pf2e/sim/engine/testdata/abyssal-luna/hand.toml`

A copy of the campaign's Luna (Commit 5, Task 9), kept apart so Effort 2 can trim the campaign file (Decision 3). Numbers from `source/actors/abyssal-luna.json`.

`internal/pf2e/sim/engine/testdata/abyssal-luna/hand.toml` (new) — full contents:

```toml
schema = 1

name       = "Abyssal Luna"
kind       = "npc"
level      = 16
traits     = ["unique", "fey", "occult"]
perception = 29

[speed]
land = 60
fly  = 180

[defense]
hp        = 300
ac        = 36
fortitude = 27
reflex    = 28
will      = 29
immune    = ["confused", "doomed", "paralyzed", "sleep"]

[strike.slash]
name   = "Slash"
bonus  = 30
traits = ["agile", "magical", "reach-10", "unarmed"]
damage = [
  { dice = "3d6+14", type = "slashing" },
  { dice = "1d8", type = "mental" },
]

[strike.stab]
name   = "Stab"
bonus  = 28
traits = ["magical", "reach-20"]
damage = [
  { dice = "3d8+14", type = "bludgeoning" },
  { dice = "1d8", type = "mental" },
]

[ability.abyssal-frenzy]
name    = "Abyssal Frenzy"
actions = 2
steps   = [{ strike = "slash" }, { strike = "slash" }, { strike = "stab" }]
```

##### Task 5 — `internal/pf2e/sim/engine/testdata/dummy-fighter/hand.toml`

Level-14 numbers from the script's `PARTY[14]` row for Kas; the Strike is a stand-in.

`internal/pf2e/sim/engine/testdata/dummy-fighter/hand.toml` (new) — full contents:

```toml
schema = 1

name       = "Dummy Fighter"
kind       = "pc"
level      = 14
perception = 24

[speed]
land = 25

[defense]
hp        = 230
ac        = 35
fortitude = 23
reflex    = 21
will      = 22

[strike.greatsword]
name   = "Greatsword"
bonus  = 29
damage = [{ dice = "3d12+9", type = "slashing" }]
```

##### Task 6 — `internal/pf2e/sim/engine/testdata/dummy-alchemist/hand.toml`

`internal/pf2e/sim/engine/testdata/dummy-alchemist/hand.toml` (new) — full contents:

```toml
schema = 1

name       = "Dummy Alchemist"
kind       = "pc"
level      = 14
perception = 22

[speed]
land = 25

[defense]
hp        = 164
ac        = 36
fortitude = 22
reflex    = 21
will      = 20

[strike.crossbow]
name   = "Crossbow"
bonus  = 24
damage = [{ dice = "3d8+2", type = "piercing" }]
```

##### Task 7 — `internal/pf2e/sim/engine/testdata/dummy-sorcerer/hand.toml`

`internal/pf2e/sim/engine/testdata/dummy-sorcerer/hand.toml` (new) — full contents:

```toml
schema = 1

name       = "Dummy Sorcerer"
kind       = "pc"
level      = 14
perception = 21

[speed]
land = 25

[defense]
hp        = 104
ac        = 33
fortitude = 20
reflex    = 22
will      = 21

[strike.staff]
name   = "Staff"
bonus  = 20
damage = [{ dice = "2d4+1", type = "bludgeoning" }]
```

##### Task 8 — `internal/pf2e/sim/engine/run_test.go`

`TestRunIsDeterministic` is gate 3.

`internal/pf2e/sim/engine/run_test.go` (new) — full contents:

```go
package engine

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

func loadTestCreatures(t *testing.T, ids ...string) []*creature.Creature {
	t.Helper()
	creatures := make([]*creature.Creature, len(ids))
	for index, id := range ids {
		loaded, err := creature.Load("testdata/" + id)
		if err != nil {
			t.Fatalf("load %s: %v", id, err)
		}
		creatures[index] = loaded
	}
	return creatures
}

func runLunaFight(t *testing.T, seed uint64) (Outcome, []Event) {
	t.Helper()
	events := &recorder{}
	outcome, err := Run(Setup{
		Party:     loadTestCreatures(t, "dummy-fighter", "dummy-alchemist", "dummy-sorcerer"),
		Enemies:   loadTestCreatures(t, "abyssal-luna"),
		Roller:    dice.NewSeeded(seed),
		Observer:  events,
		Decider:   FirstFitting{},
		MaxRounds: 20,
	})
	if err != nil {
		t.Fatalf("Run(seed %d): %v", seed, err)
	}
	return outcome, events.events
}

func TestRunIsDeterministic(t *testing.T) {
	firstSeedEvents := []Event(nil)
	seedsDiffer := false
	for seed := uint64(1); seed <= 10; seed++ {
		outcome, events := runLunaFight(t, seed)
		againOutcome, againEvents := runLunaFight(t, seed)
		if outcome != againOutcome || !reflect.DeepEqual(events, againEvents) {
			t.Fatalf("seed %d gave two different fights", seed)
		}
		ended, isEnd := events[len(events)-1].(FightEnded)
		if !isEnd || ended.Outcome != outcome {
			t.Errorf("seed %d: last event = %+v, want FightEnded with %v", seed, events[len(events)-1], outcome)
		}
		if firstSeedEvents == nil {
			firstSeedEvents = events
		} else if !reflect.DeepEqual(events, firstSeedEvents) {
			seedsDiffer = true
		}
	}
	if !seedsDiffer {
		t.Error("seeds 1 to 10 all gave the same fight")
	}
}

func TestInitiativeTieGoesToTheNPC(t *testing.T) {
	pc := testCombatant("pc", SideParty, creature.KindPC, 30, 20).Creature
	npc := testCombatant("npc", SideEnemies, creature.KindNPC, 30, 20).Creature
	fight, err := newFight(Setup{
		Party:     []*creature.Creature{pc},
		Enemies:   []*creature.Creature{npc},
		Roller:    script(t, 10, 10),
		Observer:  &recorder{},
		Decider:   FirstFitting{},
		MaxRounds: 1,
	})
	if err != nil {
		t.Fatalf("newFight: %v", err)
	}
	if want := []string{"npc", "pc"}; !slices.Equal(orderIDs(fight), want) {
		t.Errorf("order = %v, want %v", orderIDs(fight), want)
	}
}

func TestFirstFittingTakesAbilityThenStrike(t *testing.T) {
	luna := loadTestCreatures(t, "abyssal-luna")[0]
	fighter := loadTestCreatures(t, "dummy-fighter")[0]
	actor := NewCombatant(SideEnemies, luna)
	foe := NewCombatant(SideParty, fighter)
	fight := &Fight{combatants: []*Combatant{foe, actor}}

	actor.ActionsLeft = 3
	choice, acting := FirstFitting{}.Decide(fight, actor)
	if !acting || choice.Kind != ChooseAbility || choice.ID != "abyssal-frenzy" || choice.Target != foe {
		t.Errorf("with 3 actions: %+v, %v, want abyssal-frenzy at the fighter", choice, acting)
	}
	actor.ActionsLeft = 1
	choice, acting = FirstFitting{}.Decide(fight, actor)
	if !acting || choice.Kind != ChooseStrike || choice.ID != "slash" {
		t.Errorf("with 1 action: %+v, %v, want the slash Strike", choice, acting)
	}
}

func TestRoundLimitIsAStalemate(t *testing.T) {
	wall := func(id string, kind creature.Kind) *creature.Creature {
		return testCombatant(id, SideParty, kind, 30, 100).Creature
	}
	events := &recorder{}
	outcome, err := Run(Setup{
		Party:     []*creature.Creature{wall("left", creature.KindPC)},
		Enemies:   []*creature.Creature{wall("right", creature.KindNPC)},
		Roller:    dice.NewSeeded(1),
		Observer:  events,
		Decider:   FirstFitting{},
		MaxRounds: 2,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if outcome != Stalemate {
		t.Errorf("outcome = %v, want stalemate", outcome)
	}
	if last := events.events[len(events.events)-1]; last != (FightEnded{Rounds: 2, Outcome: Stalemate}) {
		t.Errorf("last event = %+v, want FightEnded after 2 rounds", last)
	}
}

func TestRunRejectsBadSetups(t *testing.T) {
	fighter := loadTestCreatures(t, "dummy-fighter")[0]
	luna := loadTestCreatures(t, "abyssal-luna")[0]
	valid := Setup{
		Party:     []*creature.Creature{fighter},
		Enemies:   []*creature.Creature{luna},
		Roller:    dice.NewSeeded(1),
		Observer:  &recorder{},
		Decider:   FirstFitting{},
		MaxRounds: 5,
	}
	duplicate := valid
	duplicate.Enemies = []*creature.Creature{luna, luna}
	noRounds := valid
	noRounds.MaxRounds = 0
	noEnemies := valid
	noEnemies.Enemies = nil

	if _, err := Run(duplicate); !errors.Is(err, ErrDuplicateCombatant) {
		t.Errorf("duplicate ids: err = %v, want ErrDuplicateCombatant", err)
	}
	if _, err := Run(noRounds); !errors.Is(err, ErrBadSetup) {
		t.Errorf("no rounds: err = %v, want ErrBadSetup", err)
	}
	if _, err := Run(noEnemies); !errors.Is(err, ErrBadSetup) {
		t.Errorf("no enemies: err = %v, want ErrBadSetup", err)
	}
}
```

##### Task 9 — `dev_journals/campaigns/initiatives/implementation/F.041.E1-pf2e-encounter-simulator-plan.md`

Tick Phase 2 in this plan's Overview.

**Find:**
```markdown
- [ ] **Phase 2 — Engine core**
```
**Replace with:**
```markdown
- [x] **Phase 2 — Engine core**
```

##### Commit message

```
feat(E1): add the fight loop, events and fixed behaviour

- engine.Run plays a fight from a Setup: party, enemies, roller, observer,
  decider and a round limit
- initiative is d20 + Perception; on a tie the NPC goes first
- three actions a turn; dying creatures make a recovery check instead
- the Decider seam, and FirstFitting: first fitting ability by ID, else the
  first Strike, at the first standing foe
- the fight ends when one side has no standing creature, or as a
  stalemate at the round limit
- gate 3: TestRunIsDeterministic plays Luna against three dummy PCs
```

### Phase 3 — show and the wiki

**In plain words:** adds `gm-toolkit pf2e sim show <creature>`, puts Luna's hand layer in the H.A.W.K campaign, and writes the PF2e side of the architecture wiki. When it lands, Effort 1 is done.

#### Commit 5 — `feat(E1): add pf2e sim show and Luna's hand layer`

##### Task 1 — `internal/campaigns/campaigns.go`

Add the PF2e folders (Decision 8).

**Find:**
```go
	NarrativesDir  string `path:"narratives"           kind:"dir"`
	LogsDir        string `path:"state/logs"           kind:"dir"`
}
```
**Replace with:**
```go
	NarrativesDir  string `path:"narratives"           kind:"dir"`
	LogsDir        string `path:"state/logs"           kind:"dir"`
	PartyDir       string `path:"pf2e/party"           kind:"dir"`
	EncountersDir  string `path:"encounters"           kind:"dir"`
}
```

##### Task 2 — `internal/campaigns/campaigns_test.go`

In `TestCampaignPaths`, cover the new fields.

**Find:**
```go
		{"LogsDir", paths.LogsDir, filepath.Join(root, "state", "logs")},
	}
```
**Replace with:**
```go
		{"LogsDir", paths.LogsDir, filepath.Join(root, "state", "logs")},
		{"PartyDir", paths.PartyDir, filepath.Join(root, "pf2e", "party")},
		{"EncountersDir", paths.EncountersDir, filepath.Join(root, "encounters")},
	}
```

##### Task 3 — `internal/campaigns/scaffold_test.go`

In `TestScaffoldCreatesTree`, check that the new folders are made.

**Find:**
```go
		filepath.Join(root, "state", "logs"),
	} {
```
**Replace with:**
```go
		filepath.Join(root, "state", "logs"),
		filepath.Join(root, "pf2e", "party"),
		filepath.Join(root, "encounters"),
	} {
```

##### Task 4 — `internal/pf2e/sim/creature/locate.go`

`internal/pf2e/sim/creature/locate.go` (new) — full contents:

```go
package creature

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrCreatureNotFound  = errors.New("creature not found")
	ErrCreatureAmbiguous = errors.New("creature found in more than one folder")
	ErrBadCreatureID     = errors.New("bad creature id")
)

func Locate(partyDir, encountersDir, id string) (string, error) {
	if !isFolderName(id) {
		return "", fmt.Errorf("%w: %q", ErrBadCreatureID, id)
	}
	candidates := []string{filepath.Join(partyDir, id)}
	fights, err := os.ReadDir(encountersDir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("locate %s: %w", id, err)
	}
	for _, fight := range fights {
		if fight.IsDir() {
			candidates = append(candidates, filepath.Join(encountersDir, fight.Name(), id))
		}
	}
	var found []string
	for _, candidate := range candidates {
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			found = append(found, candidate)
		}
	}
	switch len(found) {
	case 0:
		return "", fmt.Errorf("%w: %s (looked in %s and %s/*)", ErrCreatureNotFound, id, partyDir, encountersDir)
	case 1:
		return found[0], nil
	default:
		return "", fmt.Errorf("%w: %s: %s", ErrCreatureAmbiguous, id, strings.Join(found, ", "))
	}
}
```

##### Task 5 — `internal/pf2e/sim/creature/locate_test.go`

`internal/pf2e/sim/creature/locate_test.go` (new) — full contents:

```go
package creature_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func TestLocate(t *testing.T) {
	root := t.TempDir()
	partyDir := filepath.Join(root, "pf2e", "party")
	encountersDir := filepath.Join(root, "encounters")
	for _, dir := range []string{
		filepath.Join(partyDir, "kas"),
		filepath.Join(encountersDir, "fight-a", "luna"),
		filepath.Join(encountersDir, "fight-a", "lilith"),
		filepath.Join(encountersDir, "fight-b", "lilith"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	cases := []struct {
		id      string
		want    string
		wantErr error
	}{
		{id: "kas", want: filepath.Join(partyDir, "kas")},
		{id: "luna", want: filepath.Join(encountersDir, "fight-a", "luna")},
		{id: "lilith", wantErr: creature.ErrCreatureAmbiguous},
		{id: "nobody", wantErr: creature.ErrCreatureNotFound},
		{id: "../kas", wantErr: creature.ErrBadCreatureID},
	}
	for _, testCase := range cases {
		got, err := creature.Locate(partyDir, encountersDir, testCase.id)
		if testCase.wantErr != nil {
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("Locate(%q) err = %v, want %v", testCase.id, err, testCase.wantErr)
			}
			continue
		}
		if err != nil || got != testCase.want {
			t.Errorf("Locate(%q) = %q, %v, want %q", testCase.id, got, err, testCase.want)
		}
	}
}

func TestLocateWithoutEncountersFolder(t *testing.T) {
	root := t.TempDir()
	partyDir := filepath.Join(root, "party")
	if err := os.MkdirAll(filepath.Join(partyDir, "kas"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := creature.Locate(partyDir, filepath.Join(root, "missing"), "kas"); err != nil {
		t.Errorf("Locate: %v", err)
	}
}
```

##### Task 6 — `cmd/gm-toolkit/pf2e/pf2e.go`

`cmd/gm-toolkit/pf2e/pf2e.go` (new) — full contents:

```go
package pf2e

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

func NewCmd() *cobra.Command {
	var campaignOverride string
	cmd := &cobra.Command{
		Use:   "pf2e",
		Short: "Pathfinder 2e tools",
		Long:  "Pathfinder 2e tools for the active campaign. Use --campaign <id> to override the active pointer for this invocation.",
	}
	cmd.PersistentFlags().StringVar(&campaignOverride, "campaign", "", "campaign id to use instead of the active one")
	cmd.AddCommand(newSimCmd(&campaignOverride))
	return cmd
}

func resolveCampaign(override string) (*campaigns.Campaign, error) {
	registry, err := campaigns.LoadRegistry()
	if err != nil {
		return nil, err
	}
	campaign, err := campaigns.ResolveActive(registry, override)
	switch {
	case err == nil:
		return campaign, nil
	case errors.Is(err, campaigns.ErrNoActiveCampaign):
		return nil, fmt.Errorf("no active campaign and no --campaign given\n  hint: gm-toolkit campaign set-active --campaign <id>")
	case errors.Is(err, campaigns.ErrCampaignNotRegistered):
		return nil, fmt.Errorf("%w\n  hint: gm-toolkit campaign register <path>", err)
	case errors.Is(err, campaigns.ErrCampaignRootMissing):
		return nil, fmt.Errorf("%w\n  hint: re-register at its new location or restore the directory", err)
	default:
		return nil, err
	}
}
```

##### Task 7 — `cmd/gm-toolkit/pf2e/sim.go`

`cmd/gm-toolkit/pf2e/sim.go` (new) — full contents:

```go
package pf2e

import "github.com/spf13/cobra"

func newSimCmd(campaignOverride *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sim",
		Short: "Simulate PF2e fights",
		Long:  "Simulate Pathfinder 2e fights from the campaign's creature files.",
	}
	cmd.AddCommand(newShowCmd(campaignOverride))
	return cmd
}
```

##### Task 8 — `cmd/gm-toolkit/pf2e/show.go`

`cmd/gm-toolkit/pf2e/show.go` (new) — full contents:

```go
package pf2e

import (
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const (
	ansiBold  = "\033[1m"
	ansiReset = "\033[0m"
)

func newShowCmd(campaignOverride *string) *cobra.Command {
	var variants []string
	cmd := &cobra.Command{
		Use:   "show <creature>",
		Short: "Print a creature's merged stat block",
		Long:  "Merges the creature's layers (import.toml, hand.toml, then each --variant in order) and prints the result. Looks in pf2e/party/ and then in each encounters/<fight>/ folder.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			campaign, err := resolveCampaign(*campaignOverride)
			if err != nil {
				return fmt.Errorf("pf2e sim show: %w", err)
			}
			paths := campaign.Paths()
			directory, err := creature.Locate(paths.PartyDir, paths.EncountersDir, args[0])
			if err != nil {
				return fmt.Errorf("pf2e sim show: %w", err)
			}
			loaded, err := creature.Load(directory, variants...)
			if err != nil {
				return fmt.Errorf("pf2e sim show: %w", err)
			}
			out := cmd.OutOrStdout()
			printCreature(out, loaded, isTerminal(out))
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&variants, "variant", nil, "variant layer to apply on top, by name; repeat to stack, applied in order")
	return cmd
}

func isTerminal(out io.Writer) bool {
	file, isFile := out.(*os.File)
	return isFile && os.Getenv("NO_COLOR") == "" && isatty.IsTerminal(file.Fd())
}

func printCreature(out io.Writer, shown *creature.Creature, styled bool) {
	bold, reset := "", ""
	if styled {
		bold, reset = ansiBold, ansiReset
	}
	fmt.Fprintf(out, "%s%s%s  %s %d", bold, shown.Name, reset, shown.Kind, shown.Level)
	if len(shown.Traits) > 0 {
		fmt.Fprintf(out, "  [%s]", strings.Join(shown.Traits, ", "))
	}
	fmt.Fprintln(out)

	defense := shown.Defense
	fmt.Fprintf(out, "Perception %+d\n", shown.Perception)
	fmt.Fprintf(out, "HP %d  AC %d  Fort %+d  Ref %+d  Will %+d\n", defense.HP, defense.AC, defense.Fortitude, defense.Reflex, defense.Will)
	if len(defense.Immune) > 0 {
		fmt.Fprintf(out, "Immune %s\n", strings.Join(defense.Immune, ", "))
	}
	printAdjustments(out, "Resist", defense.Resist)
	printAdjustments(out, "Weakness", defense.Weakness)
	fmt.Fprintf(out, "Speed %s\n", speedText(shown.Speed))

	if len(shown.Strikes) > 0 {
		fmt.Fprintf(out, "%sStrikes%s\n", bold, reset)
		for _, id := range shown.StrikeIDs() {
			strike := shown.Strikes[id]
			fmt.Fprintf(out, "  %s %+d", strike.Name, strike.Bonus)
			if len(strike.Traits) > 0 {
				fmt.Fprintf(out, " (%s)", strings.Join(strike.Traits, ", "))
			}
			fmt.Fprintf(out, "  %s\n", damageText(strike.Damage))
		}
	}

	if len(shown.Abilities) > 0 {
		fmt.Fprintf(out, "%sAbilities%s\n", bold, reset)
		for _, id := range shown.AbilityIDs() {
			ability := shown.Abilities[id]
			if ability.Ignore {
				fmt.Fprintf(out, "  %s (ignored)\n", id)
				continue
			}
			steps := make([]string, len(ability.Steps))
			for index, step := range ability.Steps {
				steps[index] = "strike " + step.Strike
			}
			fmt.Fprintf(out, "  %s [%d]  %s\n", ability.Name, ability.Actions, strings.Join(steps, ", "))
		}
	}

	fmt.Fprintf(out, "Layers %s\n", strings.Join(shown.Layers, " + "))
}

func printAdjustments(out io.Writer, label string, adjustments map[string]creature.Adjustment) {
	if len(adjustments) == 0 {
		return
	}
	entries := make([]string, 0, len(adjustments))
	for _, key := range slices.Sorted(maps.Keys(adjustments)) {
		adjustment := adjustments[key]
		entry := fmt.Sprintf("%s %d", key, adjustment.Value)
		if len(adjustment.Except) > 0 {
			entry += fmt.Sprintf(" (except %s)", strings.Join(adjustment.Except, ", "))
		}
		entries = append(entries, entry)
	}
	fmt.Fprintf(out, "%s %s\n", label, strings.Join(entries, ", "))
}

func speedText(speed creature.Speed) string {
	text := fmt.Sprintf("%d ft", speed.Land)
	if speed.Fly > 0 {
		text += fmt.Sprintf(", fly %d ft", speed.Fly)
	}
	return text
}

func damageText(parts []creature.DamagePart) string {
	texts := make([]string, len(parts))
	for index, part := range parts {
		texts[index] = part.Roll.String() + " " + part.Type
	}
	return strings.Join(texts, " plus ")
}
```

##### Task 9 — `campaigns/hawk/encounters/fall-of-luna/abyssal-luna/hand.toml`

Luna as the sim sees her in Effort 1: the numbers from `source/actors/abyssal-luna.json`, plus Abyssal Frenzy. Same content as the engine's testdata copy (Commit 4, Task 4). Effort 2 trims this file to logic once `import.toml` carries the numbers (Decision 3).

`campaigns/hawk/encounters/fall-of-luna/abyssal-luna/hand.toml` (new) — full contents:

```toml
schema = 1

name       = "Abyssal Luna"
kind       = "npc"
level      = 16
traits     = ["unique", "fey", "occult"]
perception = 29

[speed]
land = 60
fly  = 180

[defense]
hp        = 300
ac        = 36
fortitude = 27
reflex    = 28
will      = 29
immune    = ["confused", "doomed", "paralyzed", "sleep"]

[strike.slash]
name   = "Slash"
bonus  = 30
traits = ["agile", "magical", "reach-10", "unarmed"]
damage = [
  { dice = "3d6+14", type = "slashing" },
  { dice = "1d8", type = "mental" },
]

[strike.stab]
name   = "Stab"
bonus  = 28
traits = ["magical", "reach-20"]
damage = [
  { dice = "3d8+14", type = "bludgeoning" },
  { dice = "1d8", type = "mental" },
]

[ability.abyssal-frenzy]
name    = "Abyssal Frenzy"
actions = 2
steps   = [{ strike = "slash" }, { strike = "slash" }, { strike = "stab" }]
```

##### Task 10 — `cmd/gm-toolkit/pf2e/show_test.go`

`cmd/gm-toolkit/pf2e/show_test.go` (new) — full contents:

```go
package pf2e

import (
	"bytes"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

func TestPrintCreature(t *testing.T) {
	shown := &creature.Creature{
		Name:       "Abyssal Luna",
		Kind:       creature.KindNPC,
		Level:      16,
		Perception: 29,
		Speed:      creature.Speed{Land: 60, Fly: 180},
		Defense: creature.Defense{
			HP: 300, AC: 36, Fortitude: 27, Reflex: 28, Will: 29,
			Resist: map[string]creature.Adjustment{"all": {Value: 10, Except: []string{"holy"}}},
		},
		Strikes: map[string]creature.Strike{
			"slash": {Name: "Slash", Bonus: 30, Traits: []string{"agile"}, Damage: []creature.DamagePart{
				{Type: "slashing", Roll: dice.Expression{Count: 3, Sides: 6, Modifier: 14}},
				{Type: "mental", Roll: dice.Expression{Count: 1, Sides: 8}},
			}},
		},
		Abilities: map[string]creature.Ability{
			"abyssal-frenzy":   {Name: "Abyssal Frenzy", Actions: 2, Steps: []creature.Step{{Strike: "slash"}, {Strike: "slash"}}},
			"prophetic-wings":  {Ignore: true},
		},
		Layers: []string{"import.toml", "hand.toml"},
	}
	var out bytes.Buffer
	printCreature(&out, shown, false)
	if strings.Contains(out.String(), "\033") {
		t.Errorf("unstyled output holds an escape code:\n%q", out.String())
	}
	for _, want := range []string{
		"Abyssal Luna  npc 16",
		"HP 300  AC 36  Fort +27  Ref +28  Will +29",
		"Resist all 10 (except holy)",
		"Speed 60 ft, fly 180 ft",
		"Slash +30 (agile)  3d6+14 slashing plus 1d8 mental",
		"Abyssal Frenzy [2]  strike slash, strike slash",
		"prophetic-wings (ignored)",
		"Layers import.toml + hand.toml",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
}
```

##### Task 11 — `cmd/gm-toolkit/root.go`

Register `pf2e` and reword the root text, which still says the tool is only for Stars Without Number.

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/cmd/gm-toolkit/campaign"
	"github.com/therobertcrocker/gm-toolkit/cmd/gm-toolkit/spatial"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "gm-toolkit",
		Short: "GM Toolkit — tools for running Stars Without Number factions and worlds",
		Long:  "GM Toolkit bundles utilities for the GM running Stars Without Number campaigns. Run `gm-toolkit <subcommand> --help` for per-subcommand details.",
	}
	root.AddCommand(campaign.NewCmd())
	root.AddCommand(spatial.NewCmd())
	root.AddCommand(newFactionCmd())
	return root
}
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/cmd/gm-toolkit/campaign"
	"github.com/therobertcrocker/gm-toolkit/cmd/gm-toolkit/pf2e"
	"github.com/therobertcrocker/gm-toolkit/cmd/gm-toolkit/spatial"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "gm-toolkit",
		Short: "GM Toolkit — tools for running tabletop campaigns",
		Long:  "GM Toolkit bundles utilities for the GM: campaign management, Stars Without Number factions and worlds, and a Pathfinder 2e encounter simulator. Run `gm-toolkit <subcommand> --help` for per-subcommand details.",
	}
	root.AddCommand(campaign.NewCmd())
	root.AddCommand(spatial.NewCmd())
	root.AddCommand(newFactionCmd())
	root.AddCommand(pf2e.NewCmd())
	return root
}
```

After the gate passes, run the manual check and read the output against Luna's export:

```
go run ./cmd/gm-toolkit pf2e sim show abyssal-luna --campaign hawk
```

##### Commit message

```
feat(E1): add pf2e sim show and Luna's hand layer

- campaigns.Paths gains PartyDir (pf2e/party) and EncountersDir (encounters)
- creature.Locate finds a creature in the party folder or any fight folder
- gm-toolkit pf2e sim show <creature> [--variant ...] prints the merged
  stat block; --campaign is persistent on pf2e
- the root command text names all three tool sets
- Abyssal Luna's hand layer in campaigns/hawk/encounters/fall-of-luna/
```

#### Commit 6 — `docs(E1): restructure the architecture front door by product`

##### Task 1 — `docs/architecture/architecture-overview.md`

Load `/gestalt:style-guide` and `/gestalt:diagram-kit` first. The diagram moves unchanged, but it sits inside an edited page. The front door is reorganised around the three products (D-24, Decision 10). The faction text moves under the faction manager unchanged, and the `engine/` and `interface/` folders stay where they are.

`docs/architecture/architecture-overview.md` (rewritten) — full contents. Copy the mermaid block from the current file verbatim where marked.

````markdown
# Architecture

> The front door to `docs/architecture/`. This page orients and routes — it
> does not re-narrate the subsystems. For the cross-subsystem story of a side,
> follow its spine; for a single subsystem, go straight to its page.

## The products

gm-toolkit holds three products. Each owns its code and, where its shape is
worth recording, its own side of this wiki.

| Product | Code | Wiki |
|---------|------|------|
| `campaigns` — the campaign as a whole, and campaign-level tools | `internal/campaigns/`, `internal/pf2e/` | [PF2e simulator](pf2e/overview.md) |
| `faction-manager` — the SWN faction-turn system and its TUI | `internal/faction/`, `internal/spatial/` | [Engine](engine/overview.md), [Interface](interface/overview.md) |
| `codex` — a campaign knowledge base (planned) | — | — |

<br/>

## Campaigns

- **[PF2e simulator](pf2e/overview.md)** — plays a Pathfinder 2e fight many
  times and reports how it tends to go. A headless engine reads creature files
  the GM owns, built from layers; the `gm-toolkit pf2e sim` commands drive it.
  It is being built in `F.041`.

<br/>

## Faction manager

The faction manager splits along one seam: a headless engine, and whatever
drives it.

- **[Engine](engine/overview.md)** — the SWN faction-turn system. A composition
  root (`engine.Engine`) owning a rulebook, a roller, a hook registry, and seven
  sub-engines, sequenced by a phase-oriented orchestrator. It is **headless by
  design**: it talks to its driver through exactly two interfaces — a collector
  supplying decisions in, a fire-and-forget observer reporting events out — and
  cannot tell who is on the other side.

- **[Interface](interface/overview.md)** — whatever drives the engine through
  that seam. Today that is a single Bubble Tea TUI (`internal/faction/tui/`); the
  side is named "interface" rather than "tui" because the same seam is meant to
  back a rebuilt CLI and any future frontend, none of which the engine can tell
  apart.

The two spines never duplicate each other: the engine pages describe what
happens when a turn resolves, the interface pages describe how a human (or, one
day, an AI planner) drives that resolution.

<!-- the existing ```mermaid block, verbatim -->

<br/>

## Design principles

The standing commitments behind the products. Each subsystem page's Key
Decisions apply them; these are the reasons a change that breaks one needs a
conversation first.

- **Headless engines** — an engine takes decisions in and reports events out
  through interfaces, and cannot tell who is on the other side. The faction
  engine uses a collector and an observer; the PF2e engine a decider and an
  observer. So a human, a script or an AI planner drives either the same way.
- **Extensible static data** — rules content lives in data the GM owns, and
  changing it needs no Go. Faction rules live in the campaign's rulebook TOML
  (see [ability catalog](engine/ability-catalog.md) and [goals](engine/goals.md));
  PF2e creatures live in layered creature files (see
  [data files](pf2e/data-files.md)).

The faction manager adds two of its own:

- **Robust history** — every state change is a mutation applied at one choke
  point and recorded in the append-only `history.jsonl`; nothing is lost. See
  [effect-mutation](engine/effect-mutation.md) and
  [persistence](engine/persistence.md).
- **Review apart from play** — browsing faction state (Manage mode) is a
  separate flow from running a turn (Turn mode), the only mode that drives the
  engine. See [interface overview](interface/overview.md).

<br/>

## How these pages work

Each side has the same shape:

- **A spine (`overview.md`)** — the cross-subsystem narrative plus a page index.
  Read it first to get the whole side in one pass.
- **One page per subsystem** — each follows a fixed template: **Purpose** (is
  this the page you need?), **Shape** (the structural story — types, data flow,
  boundaries), **Key Decisions** (the durable "why it's shaped this way"), and
  **Dependencies** (what it leans on / what leans on it). Pages backed by
  rulebook data may also carry a **catalogue** (e.g. the Actions, Goals, and
  Mutation catalogues on the engine side).

**Key Decisions sections are the primary read surface.** This is a push-not-pull
system: rather than hunting a decisions log, you read the decisions where you'll
trip over them — on the page for the subsystem you're changing. Discovery and
Plan sessions read the relevant page's Key Decisions as binding context before
designing, and the pre-merge checklist *appends* newly-durable decisions there.

**Routing rule.** Anything about a single subsystem lives on that subsystem's
page; only narrative that genuinely spans subsystems lives on a spine. If you're
unsure where something belongs, it belongs on a page.

<br/>

## Where to start

| If you want… | Read |
|--------------|------|
| The PF2e simulator in one pass | [pf2e/overview.md](pf2e/overview.md) |
| How a creature file becomes a stat block | [pf2e/data-files.md](pf2e/data-files.md) |
| The whole faction engine in one pass | [engine/overview.md](engine/overview.md) |
| The whole faction interface in one pass | [interface/overview.md](interface/overview.md) |
| How a faction turn resolves end to end | [engine/overview.md](engine/overview.md) → [orchestrator](engine/orchestrator.md) |
| How the TUI drives the faction engine | [interface/overview.md](interface/overview.md) → [event stream](interface/event-stream.md) |
| The "why" behind a subsystem before changing it | that subsystem's page, **Key Decisions** |

<br/>

## The tree

```
docs/architecture/
  architecture-overview.md   ← you are here (front door)
  pf2e/                      campaigns: the PF2e simulator
    overview.md          data-files.md        engine.md
  engine/                    faction manager: the engine side
    overview.md          orchestrator.md      turn-pipeline.md
    actions.md           ability-catalog.md   goals.md
    hooks.md             world-movement.md    spatial.md
    effect-mutation.md   persistence.md       narrative-digest.md
    logging-errors.md    dice.md
  interface/                 faction manager: the interface side
    overview.md          regions.md           state-machine.md
    event-stream.md      overlays.md          views.md
```
````

##### Commit message

```
docs(E1): restructure the architecture front door by product

- the front door opens with the three products and routes to each side
- the faction engine and interface sit under the faction manager, unchanged
- design principles split into shared ones and faction-only ones
- the tree and the start table add the pf2e/ side
```

#### Commit 7 — `docs(E1): add the pf2e overview, data-files and engine pages`

Load `/gestalt:style-guide` and read the arch-page template (`templates/arch-page.md` in the session-modes skill) first. Each page follows its sections. Before writing a claim in Shape, check it against the code written in Commits 1 to 5. Key Decisions are given word for word; add the source tag shown.

##### Task 1 — `docs/architecture/pf2e/overview.md`

The spine (D-22). Required content:

- **Purpose** — the PF2e side: a simulator that plays one fight many times. Version 1 is `F.041`; the ruleset as data and an encounter designer are later work.
- **Shape** — the pipeline `import → files → load → engine → runner`. Mark import and runner "(planned)" and name the effort that brings each. Built today: the files (`creature.Load`), the engine (`engine.Run`) and `show`. Give the package map (`internal/pf2e/sim/dice`, `creature`, `engine`; `cmd/gm-toolkit/pf2e`) and the two seams: the `Decider` (decisions in) and the `Observer` (events out). Position is "(planned)", Effort 5.
- **Page index** — `data-files.md`, `engine.md`; `sim.md` "(planned)", Effort 5.
- **Key Decisions:**
  - **The simulator has its own data model; Foundry is an import source.** The engine reads only its own files, so imported, hand-written and designer-made creatures are the same kind of file. (`F.041` D-1, D-6.)
  - **All v1 code lives under `internal/pf2e/sim/`.** `internal/pf2e/` is the PF2e root; `internal/pf2e/rules/` is left for the ruleset as data. (`F.041` overview Decision 4.)
  - **The commands are `gm-toolkit pf2e sim <command>`, with `--campaign` persistent on `pf2e`.** (`F.041` overview Decision 5.)

##### Task 2 — `docs/architecture/pf2e/data-files.md`

Required content:

- **Purpose** — how a creature's files become one stat block.
- **Shape** — the campaign folders (`pf2e/party/<pc>/`, `encounters/<fight>/<creature>/`, from `campaigns.Paths`); the layer order (`import.toml`, `hand.toml`, variants as `--variant` names); the merge rule; the file format with a short real example (Luna's Slash and Abyssal Frenzy); `schema = 1` in every layer and the strict per-layer decode; what `prepare` checks; `ignore = true`; `creature.Locate` and its error for a name found twice.
- **Key Decisions:**
  - **A creature is an import layer the importer owns plus hand layers, merged on load.** Re-import stays one safe step; fixes, homebrew and level-up variants sit in hand layers. (`F.041` D-7.)
  - **One merge rule over keyed tables.** Strikes and abilities are keyed tables, so tables merge key by key and every other value is replaced by the later layer. No list is merged by ID, and merged tables keep no file order. (`F.041.E1` Decision 4.)
  - **Every layer declares its schema version and is decoded strictly on its own.** An unknown key names its file, and Effort 4's approval cache keys on the schema version. (`F.041.E1` Decision 4.)
  - **An ignored ability stays on the creature.** The engine skips it, but `show` and `check` can still see that it was ignored on purpose. (`F.041.E1` Decision 4; D-8.)

##### Task 3 — `docs/architecture/pf2e/engine.md`

Required content:

- **Purpose** — the rules and the fight loop.
- **Shape** — the seeded roller (`dice.Seeded`, one per fight); `engine.Setup` and `engine.Run`; initiative and the tie rule; the turn (three actions, recovery check when dying); `DegreeOf`; Strikes and the multiple attack penalty; damage per type with immunity, weakness and resistance; dying, wounded and the knock-down move in initiative; the end of a fight and the stalemate; the `Decider` seam with `FirstFitting`; the event list (a table of the event kinds and when each fires). Name the rules Effort 3 adds as "(planned)".
- **Key Decisions:**
  - **The engine is headless: decisions come in through a `Decider`, events go out through one `Observe(Event)`.** Events are plain values that name combatants by ID, so scoring, replay and the determinism test can keep and compare them. Unlike the faction engine's method-per-event observer. (`F.041.E1` Decision 5.)
  - **One seeded roller per fight; the same seed and files give the same events.** Every roll goes through it, and no map's order reaches a roll or an event: Strikes and abilities are walked in sorted ID order, and a map is read unsorted only where order cannot change the result (picking the strongest resistance or weakness). (`F.041` gate 3; reworded in execution, Decision Record 2.)
  - **Rules are modelled as written, and nothing is faked.** NPCs die at 0 HP and PCs go dying; an initiative tie goes to the NPC; a PC knocked out moves to just before the creature that did it. (`F.041` D-18; `F.041.E1` Decisions 2 and 9.)
  - **`FirstFitting` is a test behaviour, not tactics.** It takes the first fitting ability by ID, else the first Strike, at the first standing foe. Effort 5's rule lists replace it behind the same seam. (`F.041.E1` Decision 6.)

##### Task 4 — `dev_journals/campaigns/initiatives/implementation/F.041-pf2e-encounter-simulator-overview-plan.md`

Tick Effort 1 in the Effort Summary.

**Find:**
```markdown
| [ ] | 1 | Engine core |
```
**Replace with:**
```markdown
| [x] | 1 | Engine core |
```

##### Task 5 — `dev_journals/campaigns/initiatives/implementation/F.041.E1-pf2e-encounter-simulator-plan.md`

Tick Phase 3 in this plan's Overview.

**Find:**
```markdown
- [ ] **Phase 3 — show and the wiki**
```
**Replace with:**
```markdown
- [x] **Phase 3 — show and the wiki**
```

##### Commit message

```
docs(E1): add the pf2e overview, data-files and engine pages

- pf2e/overview.md: the spine, packages, seams and page index
- pf2e/data-files.md: layers, the merge rule, the file format, lookup
- pf2e/engine.md: roller, rules, fight loop, decider and events
- tick Effort 1 in the overview plan's Effort Summary
```
