# Migrate Data Files to TOML 1.1 — Implementation Plan

## In Plain Words

- **What:** the hand-written data files (the three asset files, `tags.toml` and the
  spatial source `data.toml`) move to TOML 1.1. Each entry becomes one block: one
  header, with every nested part written in braces under it. The saved `state.toml`
  stays TOML, but a location now writes itself on one line, so state loses its three
  nested headers per location.
- **Why:** today one asset can spread over nine headers that repeat its key. A GM
  reads and copies an entry as one unit only if it is one block.
- **How it fits:** a refactor on `feature/data-driven-goals`, beside `A.001.8`. That
  initiative writes its goal parts in the same style (its row carries the amendment
  to D-17). TomlSmith's Format Document, at 120 columns, writes the style, so the
  formatter is the style check.

## Context / Goal

- Exploration (Discovery for this refactor): [R.029 exploration](../explorations/R.029-toml-1-1-migration-exploration.md). It covers H-2 (one block per entry), the house style, G-1 (goals), S-1 (state) and the Reach table
- Related: [`A.001.8` discovery](../discovery/A.001.8-data-driven-goals-discovery.md) D-17, amended by the `A.001.8` tracker row
- Goal: every hand-written rulebook file decodes the same as before (`reflect.DeepEqual` on the loaded `Rulebook`) and is a TomlSmith fixed point at 120 columns. `state.toml` has no location headers. Live docs call the effects model "the `effects` list"

## Decisions Ratified in Planning

1. **No `=` alignment; this amends the exploration's house-style call.** TomlSmith 0.4.1's formatter strips alignment in every table, the 1.0 files too (46 edits in `tags.toml`). It was tested over its LSP in the plan session. Without alignment, the house style is exactly what the formatter writes at 120 columns: it joins a table that fits, opens one that doesn't, and keeps an array open when it has a trailing comma. Robert's call.
2. **Editor settings: TomlSmith formats TOML at 120 columns, with no format on save.** Saving with the formatter on would rewrite files R.029 leaves alone: 211 lines of the `movement.toml` scenario (it removes the indent under headers) and 51 lines of the builder's `regions.toml`. So formatting stays a deliberate Format Document. Robert's call.
3. **The 1.1 guard test is dropped; this amends the exploration's Reach lean.** Once the asset files are 1.1, about 20 tests that call `loadShippedRulebook` fail if the library drops 1.1. Robert's call.
4. **The house style lives in `docs/contributing/engine.md`**, in a "Rulebook TOML style" section that `A.001.8` Effort 1 and `A.001.5` follow. Robert's call.
5. **Three phases:** state and the style first, then the asset files, then tags, the spatial source and the rename. Robert's call.
6. **Execution changes `campaigns/sandbox` too:** its `spatial/source/data.toml` and a re-compose of its rulebook, beside `sandbox-claude`. The files are tracked, so a revert is one command. Robert's call.
7. **Data-file tasks give the pattern and the gate, not the converted text.** The refactor flow describes a pattern and a unit list. A literal copy of 2,200 converted lines would break the phase budget and repeat the work. The gate (decode equality plus the formatter fixed point) checks each file mechanically.
8. **Short `effects` entries go on one line.** The line rule says so, and the formatter joins them anyway. This differs from the trial file, which spread every effect over lines.
9. **`MarshalTOML` goes on `domain.Location` and `spatial.RegionHex`, both.** With only `RegionHex`, a `[location]` header with `world_id` stays. `Base` (`[[bases]]`) and `current_order` keep their headers: each has flat keys, and the scenario style already lets a hand author compact them. The keys keep today's spelling (`RegionID`, `Coord`, `Q`, `R`), so output and scenarios match.
10. **`MarshalTOML` quotes with Go's `%q`, not a TOML escaper.** IDs are slugs. Go and TOML escape `"`, `\` and printable Unicode the same way, and the state test pins a world ID that has a quote and a backslash. Written as a fence-sign at the spot.
11. **Rename reach: 10 live files**: 5 Go comments, 3 architecture pages, the `A.001` arc plan and the `R.019` tracker row. Completed, archived, exploration and discovery docs stay as written. Other header-style names (`[transport]`, `[ability]`) were not part of the rename call and stay.
12. **The builder's error text `[regions.B]` stays.** It is a key path and still correct. Only the doc's prose and the reference table change.

## Decision Record — Execution

| # | Commit | Decision | Why |
|---|---|---|---|
| E-1 | 1 | The header-count gate removes `scratch/enc/state.toml`, not `scratch/enc/*.toml` | In zsh an unmatched glob is an error, so the planned command stopped on a fresh `scratch/enc`; plan body fixed |
| E-2 | 3 | `cunning_assets.toml` holds shapes the Task 2 table leaves out: `ability.relocate` (Covert Shipping, Covert Transit Net), effect parameter tables (`coin_gain`, `lock_out`, `reroll`, `bonus_die`, `capture_target`) and a flat `gates = { once_per = … }` (Book of Secrets). Each went inline by the style's parameter-table rule | Drift in the plan's shape list only; the style already covers them, and the gate passed (IDENTICAL, fixed point) |
| E-3 | 4 | `force_assets.toml` also holds effect parameter tables beyond the planned `ability.purchase` (`self_damage`, `coin_steal`, `restrict_attack`) and a flat `gates = { once_per = … }` (Blockade Fleet). Same treatment as E-2 | Same drift as E-2; gate passed first time |
| E-4 | 5 | In `wealth_assets.toml`, `attack` now comes before `ability` in Monopoly, Marketers, Venture Capital and Commodities Broker, which had `ability` first. The file also holds `gates = { no_stack = true }` and effect parameter tables (`cost_surcharge`, `discount_repair`, `tech_level_floor`, `restrict_attack`) beyond the planned shapes | The style's key order puts `attack` before `ability`; key order does not change the decode (IDENTICAL) |

---

## Shared Context

- **Commands** run from the repo root in zsh. Quote echo banners (`echo "==== step ===="`); a bare `====` is a zsh error.
- **CLI commands** need `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns`. The registry there lists `sandbox` and `sandbox-claude`.
- **`scratch/` is gitignored.** The check tools in this plan live there and are never committed. If a session finds one missing, it creates it from the task that defines it.

## Out of Scope

- `rulebooks/_core/goals.toml`: it has no parts yet. `A.001.8` Effort 1 writes them in the house style
- `settings.toml` and `rulebook.toml`: there are no parts to fold
- Scenarios and `state.seed.toml`: valid as they are, and already compacted by hand
- TOML strings in Go tests and the `testdata` files: they test the loader, not the style, and they keep proving that 1.0 still loads
- Builder output (`regions.toml`, `worlds.toml`), `campaigns-registry.toml` and `campaign.toml`: machine-written or flat
- A state writer of our own, or a different state format (exploration S-2, S-3); the state-format revisit row was not filed (Robert's call)
- Changing the builder's `[regions.X]` error text

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — State locations and the house style**
  - Commit 1 — `refactor: write state locations as inline tables` (3 tasks)
  - Commit 2 — `docs: add the rulebook TOML house style` (2 tasks)
- [x] **Phase 2 — Asset files**
  - Commit 3 — `refactor: move cunning assets to TOML 1.1` (3 tasks)
  - Commit 4 — `refactor: move force assets to TOML 1.1` (1 task)
  - Commit 5 — `refactor: move wealth assets to TOML 1.1` (1 task)
- [ ] **Phase 3 — Tags, spatial source and the rename**
  - Commit 6 — `refactor: move tags to TOML 1.1 and re-compose the sandboxes` (2 tasks)
  - Commit 7 — `refactor: write spatial regions as one table` (3 tasks)
  - Commit 8 — `docs: call the effects model the effects list` (10 tasks)

### Phase 1 — State locations and the house style

**In plain words:** a saved `state.toml` writes every location on one line, and the house style is written down, with the editor set to check it.

#### Commit 1 — `refactor: write state locations as inline tables`

##### Task 1 — `internal/spatial/region_map.go`

`RegionHex` writes itself as one inline table. `fmt` is already imported.

**Find:**
```go
type RegionHex struct {
	RegionID string
	Coord    HexCoord
}
```
**Replace with:**
```go
type RegionHex struct {
	RegionID string
	Coord    HexCoord
}

// MarshalTOML writes the hex as one inline table under the keys the decoder reads,
// so a saved state.toml has no region_hex or Coord headers (R.029 S-1).
func (regionHex RegionHex) MarshalTOML() ([]byte, error) {
	return fmt.Appendf(nil, "{ RegionID = %q, Coord = { Q = %d, R = %d } }", regionHex.RegionID, regionHex.Coord.Q, regionHex.Coord.R), nil
}
```

##### Task 2 — `internal/faction/domain/location.go`

`Location` writes itself as one inline table and nests the hex.

**Find:**
```go
import "github.com/therobertcrocker/gm-toolkit/internal/spatial"

type Location struct {
	WorldID   string            `toml:"world_id"`
	RegionHex spatial.RegionHex `toml:"region_hex"`
}
```
**Replace with:**
```go
import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type Location struct {
	WorldID   string            `toml:"world_id"`
	RegionHex spatial.RegionHex `toml:"region_hex"`
}

// MarshalTOML writes a location as one inline table, so a saved state.toml has no
// location headers (R.029 S-1).
func (location Location) MarshalTOML() ([]byte, error) {
	regionHex, err := location.RegionHex.MarshalTOML()
	if err != nil {
		return nil, err
	}
	// deliberate: Go %q quoting, not a TOML escaper, because IDs are slugs; TestSaveWritesLocationsInline pins quotes and backslashes
	return fmt.Appendf(nil, "{ world_id = %q, region_hex = %s }", location.WorldID, regionHex), nil
}
```

##### Task 3 — `internal/faction/state/faction_state_test.go`

A test covers every place state holds a location: the homeworld, a base, an asset, a goal target, and a movement order's origin, destination and path. One world ID has a quote and a backslash in it.

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)
```

**Find:**
```go
func TestTurnStateRoundTrip(t *testing.T) {
```
**Replace with:**
```go
func TestSaveWritesLocationsInline(t *testing.T) {
	mireholt := domain.Location{WorldID: "mireholt", RegionHex: spatial.RegionHex{RegionID: "F", Coord: spatial.HexCoord{Q: 7, R: -2}}}
	awkward := domain.Location{WorldID: `quote"back\slash`, RegionHex: spatial.RegionHex{RegionID: "A", Coord: spatial.HexCoord{Q: 1, R: 2}}}
	inFlight := domain.Location{RegionHex: spatial.RegionHex{RegionID: "B", Coord: spatial.HexCoord{Q: 3, R: 4}}}
	original := &FactionState{
		CampaignID:  "test-campaign",
		CycleNumber: 1,
		Factions: map[string]*domain.Faction{
			"ashfall": {
				ID:         "ashfall",
				Name:       "Ashfall",
				Homeworld:  mireholt,
				Tags:       []string{"SWN-T-001"},
				ActiveGoal: &domain.ActiveGoal{GoalID: "G-004", TargetWorld: awkward},
				Assets: map[string]*domain.Asset{
					"ashfall-SWN-F1-001-1": {
						ID:           "ashfall-SWN-F1-001-1",
						DefinitionID: "SWN-F1-001",
						OwnerID:      "ashfall",
						Location:     inFlight,
						CurrentHP:    3,
						CurrentOrder: &domain.MovementOrder{
							AssetID:     "ashfall-SWN-F1-001-1",
							Origin:      mireholt,
							Destination: awkward,
							Path:        []spatial.RegionHex{mireholt.RegionHex, inFlight.RegionHex, awkward.RegionHex},
							StepIdx:     1,
						},
					},
				},
				Bases: []*domain.Base{{ID: "ashfall-homeworld", OwnerID: "ashfall", Location: mireholt, CurrentHP: 20, MaxHP: 20, IsHomeworld: true}},
			},
		},
	}

	path := filepath.Join(t.TempDir(), "state.toml")
	if err := Save(path, original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading state.toml: %v", err)
	}
	for _, header := range []string{"homeworld]", "location]", "region_hex]", "Coord]", "target_world]", "origin]", "destination]", "path]]"} {
		if strings.Contains(string(written), header) {
			t.Errorf("state.toml has a %q header, want every location inline:\n%s", header, written)
		}
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !reflect.DeepEqual(loaded.Factions, original.Factions) {
		t.Errorf("Factions after round trip:\n got %+v\nwant %+v", loaded.Factions["ashfall"], original.Factions["ashfall"])
	}
}

func TestTurnStateRoundTrip(t *testing.T) {
```

##### Gate

`go build ./... && go vet ./... && go test ./...`. Then encode a real scenario through `state.Save` and count its headers. Before the change it has 115 headers. After it has 31: `[governments]`, `[current_turn]`, `[factions]`, 4 factions, 4 `assets` tables, 14 assets, 4 `[[bases]]`, 1 `current_order` and 1 `[[conditions]]`. The output dir is a scratch folder.

`scratch/encodecheck/main.go` (gitignored; create it if missing). Full contents:
```go
package main

import (
	"fmt"
	"os"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func main() {
	factionState, err := state.Load(os.Args[1])
	if err != nil {
		fmt.Println("load:", err)
		os.Exit(1)
	}
	if err := state.Save(os.Args[2], factionState); err != nil {
		fmt.Println("save:", err)
		os.Exit(1)
	}
}
```
```sh
echo "==== encode movement scenario ====" && mkdir -p scratch/enc && rm -f scratch/enc/state.toml && go run ./scratch/encodecheck campaigns/sandbox/state/scenarios/movement.toml scratch/enc/state.toml && grep -c '^\s*\[' scratch/enc/state.toml
```

##### Commit message

```
refactor: write state locations as inline tables

- Location and RegionHex implement toml.Marshaler and write one-line inline
  tables, so a saved state.toml drops its location, region_hex and Coord
  headers (R.029 exploration S-1)
- Test a save with every location kind, including a quote and a backslash
  in a world ID
```

#### Commit 2 — `docs: add the rulebook TOML house style`

##### Task 1 — `docs/contributing/engine.md`

Add the style section between "Conventions that bite" and "Test tooling", and name it in the intro.

**Find:**
```markdown
> pattern; this guide carries the engine's test tooling, the conventions you'll
> trip over, and six extension recipes.
```
**Replace with:**
```markdown
> pattern; this guide carries the engine's test tooling, the conventions you'll
> trip over, the rulebook TOML style, and six extension recipes.
```

**Find:**
````markdown
  here. ([logging & errors](../architecture/engine/logging-errors.md).)

<br/>
<br/>

# Test tooling
````
**Replace with:**
````markdown
  here. ([logging & errors](../architecture/engine/logging-errors.md).)

<br/>
<br/>

# Rulebook TOML style

The hand-written data files (`*_assets.toml`, `tags.toml`, `goals.toml`, and a
campaign's `spatial/source/data.toml`) are TOML 1.1. Each entry is one block: one
header, with every nested part written in braces under it. TomlSmith's Format
Document, at the 120-column width the devcontainer sets, writes this style, so a
formatted file is a styled file. (Why: [R.029 exploration](../initiatives/explorations/R.029-toml-1-1-migration-exploration.md), H-2.)

- **One header per entry.** `[assets.Smugglers]`, `[tags.Colonists]`,
  `[goals.MilitaryConquest]`. A part (`attack`, `transport`, `ability`, `effects`,
  `filters`, `gates`, a parameter table) never gets a header of its own.
- **The line rule.** A part that fits in 120 columns sits on one line. A part that
  does not opens over lines: one key per line, each line ending in a comma, and the
  closing brace on its own line.
- **Lists always open.** `effects`, `steps` and `outcomes` put each element on a new
  line, each ending in a comma. Each element then follows the line rule.
- **No dotted keys.** Write `gates = { contest = { … } }`, never
  `gates.contest = { … }`. Braces are the one way to nest.
- **No alignment.** One space each side of `=`. The formatter removes padding.
- **Key order.** The entry's own keys first, then its parts: `attack`, `transport`,
  `ability`, `effects`. Inside an effect: `trigger`, `effect`, `filters`, `gates`,
  then the parameter table.
- **Once a part is inline, all of it is inline.** A later `[assets.X.attack]` header
  cannot add to `attack = { … }`; the loader refuses the file and names the line.

```toml
[assets.VentureCapital]
id = "SWN-W6-001"
name = "Venture Capital"
# … the entry's own keys …
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

[assets.Lobbyists]
id = "SWN-C2-001"
# … the entry's own keys …
effects = [
  {
    trigger = "permission_granted",
    effect = "withdraw_permission",
    filters = { owner = "rival", scope = "world" },
    gates = { contest = { roller = "self", roll = "test", attacker_stat = "Cunning", defender_stat = "Cunning" } },
  },
]
```

A short effect sits on one line inside the open list:
`{ trigger = "lethal_damage", effect = "sacrifice_self", filters = { owner = "own", scope = "world" } },`.

<br/>
<br/>

# Test tooling
````

##### Task 2 — `.devcontainer/devcontainer.json`

TomlSmith becomes the TOML formatter at 120 columns, with no format on save (Decision 2). The settings apply after a container rebuild.

**Find:**
```json
        "editor.formatOnSave": true,
        "editor.defaultFormatter": "golang.go"
      }
```
**Replace with:**
```json
        "editor.formatOnSave": true,
        "editor.defaultFormatter": "golang.go",
        "[toml]": {
          "editor.defaultFormatter": "TomlSmith.tomlsmith",
          "editor.formatOnSave": false
        },
        "tomlsmith.format.lineWidth": 120,
        "tomlsmith.format.indentWidth": 2
      }
```

##### Gate

Docs and settings only. `go build ./...` stays green; check that `devcontainer.json` still parses (`python3 -c "import json,re,sys; json.loads(re.sub(r'^\s*//.*$', '', open('.devcontainer/devcontainer.json').read(), flags=re.M))"`).

##### Commit message

```
docs: add the rulebook TOML house style

- New "Rulebook TOML style" section in the engine contributing guide: one
  header per entry, the 120-column line rule, open lists, no dotted keys,
  no alignment, key order
- Devcontainer: TomlSmith formats TOML at 120 columns, no format on save
```

### Phase 2 — Asset files

**In plain words:** the three asset files are rewritten in the house style, one file per commit, and each one loads exactly as before.

#### Commit 3 — `refactor: move cunning assets to TOML 1.1`

##### Task 1 — `scratch/loadcheck/main.go` and `scratch/tomlfmt/lsp_format.py` (gitignored, not committed)

These two tools are the gate for every data-file commit (Commits 3 to 7). `loadcheck` composes two rulebook source folders, loads each, and compares the two `Rulebook` values. `lsp_format.py` runs TomlSmith's formatter on one file and prints the result.

`scratch/loadcheck/main.go` (replace the exploration's version). Full contents:
```go
package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

func load(sourceDir string) *rulebook.Rulebook {
	root, err := os.MkdirTemp("", "loadcheck-")
	if err != nil {
		fmt.Println("temp dir:", err)
		os.Exit(1)
	}
	camp := &campaigns.Campaign{Root: root}
	if _, err := campaigns.ComposeRulebook(camp, sourceDir, false); err != nil {
		fmt.Println("compose", sourceDir+":", err)
		os.Exit(1)
	}
	loaded, err := rulebook.Load(camp.Paths().FactionDataDir)
	if err != nil {
		fmt.Println("load", sourceDir+":", err)
		os.Exit(1)
	}
	return loaded
}

func main() {
	baseline := load(os.Args[1])
	candidate := load(os.Args[2])
	fmt.Printf("assets %d vs %d, tags %d vs %d, goals %d vs %d\n",
		len(baseline.Assets), len(candidate.Assets), len(baseline.Tags), len(candidate.Tags), len(baseline.Goals), len(candidate.Goals))
	if reflect.DeepEqual(baseline, candidate) {
		fmt.Println("IDENTICAL")
		return
	}
	for id, definition := range baseline.Assets {
		if !reflect.DeepEqual(definition, candidate.Assets[id]) {
			fmt.Println("asset differs:", id)
		}
	}
	for id, tag := range baseline.Tags {
		if !reflect.DeepEqual(tag, candidate.Tags[id]) {
			fmt.Println("tag differs:", id)
		}
	}
	os.Exit(1)
}
```

`scratch/tomlfmt/lsp_format.py`. Full contents:
```python
import json
import subprocess
import sys

server_path, toml_path, line_width = sys.argv[1], sys.argv[2], int(sys.argv[3])
text = open(toml_path).read()
settings = {"tomlsmith": {"toml": {"version": "1.1"}, "format": {"lineWidth": line_width, "indentWidth": 2}}}
process = subprocess.Popen([server_path, "--stdio"], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE)


def send(message):
    body = json.dumps(message).encode()
    process.stdin.write(b"Content-Length: %d\r\n\r\n" % len(body) + body)
    process.stdin.flush()


def receive():
    headers = {}
    while True:
        line = process.stdout.readline().decode().strip()
        if not line:
            break
        name, value = line.split(":", 1)
        headers[name.lower()] = value.strip()
    return json.loads(process.stdout.read(int(headers["content-length"])))


def wait_for(request_id):
    while True:
        message = receive()
        if message.get("id") == request_id and "method" not in message:
            return message
        if message.get("method") == "workspace/configuration":
            results = []
            for item in message["params"]["items"]:
                value = settings
                for part in item.get("section", "").split(".") if item.get("section") else []:
                    value = value.get(part, {}) if isinstance(value, dict) else {}
                results.append(value)
            send({"jsonrpc": "2.0", "id": message["id"], "result": results})
        elif "id" in message and "method" in message:
            send({"jsonrpc": "2.0", "id": message["id"], "result": None})


uri = "file://" + toml_path
send({"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {
    "processId": None, "rootUri": None,
    "capabilities": {"workspace": {"configuration": True}}, "initializationOptions": settings}})
wait_for(1)
send({"jsonrpc": "2.0", "method": "initialized", "params": {}})
send({"jsonrpc": "2.0", "method": "workspace/didChangeConfiguration", "params": {"settings": settings}})
send({"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": {"textDocument": {
    "uri": uri, "languageId": "toml", "version": 1, "text": text}}})
send({"jsonrpc": "2.0", "id": 2, "method": "textDocument/formatting", "params": {
    "textDocument": {"uri": uri}, "options": {"tabSize": 2, "insertSpaces": True}}})
edits = wait_for(2).get("result") or []

offsets = [0]
for line in text.split("\n"):
    offsets.append(offsets[-1] + len(line) + 1)


def offset(position):
    return offsets[position["line"]] + position["character"]


for edit in sorted(edits, key=lambda e: offset(e["range"]["start"]), reverse=True):
    start, end = offset(edit["range"]["start"]), offset(edit["range"]["end"])
    text = text[:start] + edit["newText"] + text[end:]
print(text, end="")
send({"jsonrpc": "2.0", "id": 3, "method": "shutdown"})
process.terminate()
```

**The data-file gate**, used by Commits 3 to 6. `FILE` is the converted file:
```sh
FILE=rulebooks/swn/assets/cunning_assets.toml
B=$(ls ~/.vscode-server/extensions/tomlsmith.tomlsmith-*/server/linux-x64/tomlsmith-lsp | tail -1)
echo "==== baseline from HEAD ====" && rm -rf scratch/baseline && mkdir -p scratch/baseline && git archive HEAD rulebooks | tar -x -C scratch/baseline
echo "==== decode check ====" && go run ./scratch/loadcheck scratch/baseline/rulebooks/swn rulebooks/swn
echo "==== formatter fixed point ====" && diff "$FILE" <(python3 -I scratch/tomlfmt/lsp_format.py "$B" "$PWD/$FILE" 120) && echo "FIXED POINT"
echo "==== tests ====" && go test ./...
```
It passes when the decode check prints `IDENTICAL`, `diff` prints nothing, and the tests pass. A `diff` line ending `, }` means a multi-line table that fits on one line; join it. A table the formatter opens without a final comma means the line was too long; open it by hand, with the comma.

##### Task 2 — `rulebooks/swn/assets/cunning_assets.toml`

Rewrite the file in the house style (Commit 2 Task 1): one `[assets.<Key>]` header per asset, with every part inline under it. Use the Write tool on the whole file. Keep the `# Cunning Assets` comment at the top, every entry's own keys in their current order, and one blank line between entries. The shapes this file holds, as before and after:

| 1.0 header | 1.1 part |
|---|---|
| `[assets.X.attack]` with 3 keys | `attack = { attacker_stat = …, defender_stat = …, damage = … }` |
| `[assets.X.transport]` | `transport = { max_hex = …, coin_cost = …, cargo_types = […], max_cargo = … }` |
| `[assets.X.ability]` too long for one line | `ability = {` then one key per line, then `}` |
| `[[assets.X.ability.steps]]` with `[assets.X.ability.steps.relocate]` | `ability = { steps = [ … ] }`, opened; each step is an element, its `relocate = { scope = "self", max_hex = 1 }` inline |
| `[[assets.X.effects]]` with `[assets.X.effects.filters]` | an `effects = [ … ]` element: `{ trigger = …, effect = …, filters = { … } },` |
| `[assets.X.effects.gates.contest]` | `gates = { contest = { … } }` inside the element; the element opens over lines |

The trial file `scratch/toml-format/cunning_assets.v11.toml` is a starting point only. It uses dotted keys (`gates.contest`, `ability.steps`), aligned `=` and spread short effects, all of which the house style drops.

##### Task 3 — Gate

Run the data-file gate (Task 1) with `FILE=rulebooks/swn/assets/cunning_assets.toml`.

##### Commit message

```
refactor: move cunning assets to TOML 1.1

- One [assets.X] header per asset; attack, transport, ability and effects
  written as inline tables in the house style
- Decodes identical to the 1.0 file (whole Rulebook, reflect.DeepEqual)
```

#### Commit 4 — `refactor: move force assets to TOML 1.1`

##### Task 1 — `rulebooks/swn/assets/force_assets.toml`

Same conversion as Commit 3 Task 2, on this file. It adds `[assets.PretechLogistics.ability.purchase]`, which becomes `purchase = { … }` inside `ability`. Gate: the data-file gate (Commit 3 Task 1) with `FILE=rulebooks/swn/assets/force_assets.toml`.

##### Commit message

```
refactor: move force assets to TOML 1.1

- One [assets.X] header per asset; parts written as inline tables in the
  house style
- Decodes identical to the 1.0 file
```

#### Commit 5 — `refactor: move wealth assets to TOML 1.1`

##### Task 1 — `rulebooks/swn/assets/wealth_assets.toml`

Same conversion as Commit 3 Task 2, on this file. It adds `ability.outcomes` lists (Venture Capital is the example in the style section), `ability.target_roll` (Harvesters) and `effects.grant_movement.relocate` (Transit Web), which becomes `grant_movement = { relocate = { … } }` inside its element. Gate: the data-file gate (Commit 3 Task 1) with `FILE=rulebooks/swn/assets/wealth_assets.toml`.

##### Commit message

```
refactor: move wealth assets to TOML 1.1

- One [assets.X] header per asset; parts written as inline tables in the
  house style
- Decodes identical to the 1.0 file
```

### Phase 3 — Tags, spatial source and the rename

**In plain words:** tags and the spatial source move to the house style, both sandboxes get the new files, and live docs call the effects model "the `effects` list".

#### Commit 6 — `refactor: move tags to TOML 1.1 and re-compose the sandboxes`

##### Task 1 — `rulebooks/swn/tags.toml`

Same conversion as Commit 3 Task 2, with `[tags.<Key>]` headers. Tags carry `effects` with parameter tables (`tech_level_floor`, `population_floor`, `bonus_die`, `gates`) and two `ability` parts (`[tags.PreceptorArchive.ability.target_roll]` and `.raise_tech_level`). The file has no comments. Gate: the data-file gate (Commit 3 Task 1) with `FILE=rulebooks/swn/tags.toml`.

##### Task 2 — Re-compose both sandboxes

Copy the converted rulebook into both campaigns. `sandbox` is tracked, so its new copies land in this commit. `sandbox-claude` is gitignored; this also fixes its stale `tags.toml`.

```sh
echo "==== compose sandbox ====" && GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns go run ./cmd/gm-toolkit campaign add-rules --rules rulebooks/swn --campaign sandbox --replace
echo "==== compose sandbox-claude ====" && GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns go run ./cmd/gm-toolkit campaign add-rules --rules rulebooks/swn --campaign sandbox-claude --replace
echo "==== copies match ====" && for f in assets/cunning_assets.toml assets/force_assets.toml assets/wealth_assets.toml tags.toml; do cmp rulebooks/swn/$f campaigns/sandbox/rulebook/$f && cmp rulebooks/swn/$f campaigns/sandbox-claude/rulebook/$f; done && echo "ALL MATCH"
```

##### Commit message

```
refactor: move tags to TOML 1.1 and re-compose the sandboxes

- One [tags.X] header per tag; effects and ability written as inline
  tables in the house style; decodes identical to the 1.0 file
- Re-compose sandbox (tracked) and sandbox-claude from rulebooks/swn, so
  both carry the 1.1 asset and tag files
```

#### Commit 7 — `refactor: write spatial regions as one table`

##### Task 1 — `campaigns/sandbox/spatial/source/data.toml`

The regions become one `[regions]` table of one-line entries. `[[worlds]]` stays as it is.

**Find:**
```toml
[regions.A]
name = "Concord Core"

[regions.B]
name = "Crossway Hub"

[regions.C]
name = "Gilded Reach"

[regions.D]
name = "Far Drift"

[regions.E]
name = "Umbral Shoals"

[regions.F]
name = "Mire Basin"

[regions.G]
name = "Ironway"
```
**Replace with:**
```toml
[regions]
A = { name = "Concord Core" }
B = { name = "Crossway Hub" }
C = { name = "Gilded Reach" }
D = { name = "Far Drift" }
E = { name = "Umbral Shoals" }
F = { name = "Mire Basin" }
G = { name = "Ironway" }
```

##### Task 2 — `campaigns/sandbox-claude/spatial/source/data.toml`

The two files are byte-identical today. Make the same edit as Task 1, then check with `cmp campaigns/sandbox/spatial/source/data.toml campaigns/sandbox-claude/spatial/source/data.toml`.

##### Task 3 — `docs/contributing/spatial-maps.md`

Step 3's block, its error row and its reference row follow the new shape. The builder output blocks stay as they are.

**Find:**
````markdown
Every region letter in the layout needs an entry in `data.toml`:

```toml
[regions.A]
name = "Ember Reach"

[regions.B]
name = "Cold Harbor"
```
````
**Replace with:**
````markdown
Every region letter in the layout needs an entry under `[regions]` in `data.toml`:

```toml
[regions]
A = { name = "Ember Reach" }
B = { name = "Cold Harbor" }
```
````

**Find:**
```markdown
| A region letter with no `[regions.X]` block | Add the block, or remove the letter from the layout |
```
**Replace with:**
```markdown
| A region letter with no entry under `[regions]` | Add the entry, or remove the letter from the layout |
```

**Find:**
```markdown
| `name` | `[regions.X]` | yes | Display name; `X` is a single uppercase letter |
```
**Replace with:**
```markdown
| `name` | `[regions]`, key `X` | yes | Display name; `X` is a single uppercase letter |
```

##### Gate

Before Task 1, run the build once at `HEAD` to confirm it reproduces the tracked output (`git diff` empty). After the edits:
```sh
B=$(ls ~/.vscode-server/extensions/tomlsmith.tomlsmith-*/server/linux-x64/tomlsmith-lsp | tail -1)
FILE=campaigns/sandbox/spatial/source/data.toml
echo "==== formatter fixed point ====" && diff "$FILE" <(python3 -I scratch/tomlfmt/lsp_format.py "$B" "$PWD/$FILE" 120) && echo "FIXED POINT"
echo "==== rebuild sandbox map ====" && GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns go run ./cmd/gm-toolkit spatial build --campaign sandbox --replace
echo "==== output unchanged ====" && git diff --exit-code campaigns/sandbox/spatial/regions.toml campaigns/sandbox/spatial/worlds.toml && echo "UNCHANGED"
echo "==== tests ====" && go test ./...
```

##### Commit message

```
refactor: write spatial regions as one table

- sandbox and sandbox-claude data.toml: [regions] holds one-line entries
  in place of a [regions.X] header per region; the built map is unchanged
- spatial-maps.md: Step 3 example, error row and field table follow
```

#### Commit 8 — `docs: call the effects model the effects list`

In 1.1 files the data writes `effects = [ … ]`, so `[[effects]]` no longer matches what a reader finds in a data file.

##### Task 1 — `internal/faction/engine/core_test.go`

**Find:** `// TestNewWithRulebook_ShippedRulebook catches an authored [[effects]] entry that`
**Replace with:** `// TestNewWithRulebook_ShippedRulebook catches an authored effects entry that`

##### Task 2 — `internal/faction/engine/effect/effect.go`

**Find:** `// CheckTagEffects checks a tag's [[effects]] entries for a faction holder. A tag`
**Replace with:** `// CheckTagEffects checks a tag's effects entries for a faction holder. A tag`

##### Task 3 — `internal/faction/engine/effect/registrar.go`

**Find:** `// Holder is who declared an [[effects]] entry: an asset through its definition,`
**Replace with:** `// Holder is who declared an effects entry: an asset through its definition,`

**Find:** `// effectsListHandler holds a definition's whole [[effects]] list, because the`
**Replace with:** `// effectsListHandler holds a definition's whole effects list, because the`

##### Task 4 — `internal/faction/rulebook/rulebook.go`

**Find:** `// not have, checked once here: an [[effects]] entry's goal filter (D-5) or`
**Replace with:** `// not have, checked once here: an effects entry's goal filter (D-5) or`

##### Task 5 — `internal/faction/domain/effect.go`

**Find:** `// EffectDefinition is one entry of an asset's [[effects]] list. At most one`
**Replace with:** `// EffectDefinition is one entry of an asset's effects list. At most one`

##### Task 6 — `docs/architecture/engine/ability-catalog.md`

**Find:** ``> `[[effects]]` list; the effects engine keys a registrar table on the effect name``
**Replace with:** ``> `effects` list; the effects engine keys a registrar table on the effect name``

**Find:** ``> S-flag table marks an authored `[[effects]]` entry.``
**Replace with:** ``> S-flag table marks an authored `effects` entry.``

##### Task 7 — `docs/architecture/engine/orchestrator.md`

**Find:** ``  rulebook's `[[effects]]` data and the effect registrars, so it checks every``
**Replace with:** ``  rulebook's `effects` data and the effect registrars, so it checks every``

##### Task 8 — `docs/architecture/engine/effect-mutation.md`

**Find:** ``Every other `S`-flag feature is an `[[effects]]` list in asset data. The effects``
**Replace with:** ``Every other `S`-flag feature is an `effects` list in asset data. The effects``

**Find:** ``  hook category.** The `[[effects]]` data never names a hook category. Each``
**Replace with:** ``  hook category.** The `effects` data never names a hook category. Each``

**Find:** ``  it.** An asset definition and a tag record carry the same `[[effects]]` list,``
**Replace with:** ``  it.** An asset definition and a tag record carry the same `effects` list,``

**Find:** ``  type) is read from the `[[effects]]` entry or from `settings.toml`. The test: a``
**Replace with:** ``  type) is read from the `effects` entry or from `settings.toml`. The test: a``

##### Task 9 — `docs/initiatives/arcs/A.001-data-driven-rulebook/A.001-data-driven-rulebook-arc-plan.md`

**Find:** ``*resolved in its Discovery:* an `[[effects]]` list of trigger``
**Replace with:** ``*resolved in its Discovery:* an `effects` list of trigger``

##### Task 10 — `docs/planned-work.md`

In the `R.019` row.

**Find:** ``beside the `[[effects]]` model; fold it in``
**Replace with:** ``beside the `effects` model; fold it in``

##### Gate

`go build ./... && go test ./...`. Then confirm no live file still says `[[effects]]`:
```sh
grep -rln '\[\[effects\]\]' . --exclude-dir=.git | grep -v '_completed\|explorations/\|discovery/\|scratch/\|R.029-toml-1-1-migration-plan.md' ; echo "exit $? (1 = none left)"
```

##### Commit message

```
docs: call the effects model the effects list

- TOML 1.1 data writes effects = [ ... ], so "[[effects]]" no longer
  matches the files; rename it in 5 Go comments, 3 architecture pages,
  the A.001 arc plan and the R.019 tracker row
- Completed, exploration and discovery docs stay as written
```
