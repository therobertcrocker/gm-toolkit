# Campaign Manager — Implementation Plan

## Context / Goal

Introduce **campaigns** as the primary unit of data ownership in the toolkit, resolving F-007 (campaign-scoped static data) as a fall-out. Ship four CLI verbs (`create`, `register`, `set-active`, `add-rules`), a new `internal/campaigns/` package owning identity / registry / path derivation, and a wired production binary that consumes the active campaign end-to-end. After this lands, a second GM can check out the repo, scaffold their own campaign, and run the toolkit without touching source.

- Discovery doc: [`../discovery/campaign-manager-discovery.md`](../discovery/campaign-manager-discovery.md)
- Unblocks: F-005.2 (tui-manage), F-012 (spatial map CLI), and any future cross-GM sharing workflow.

## Decisions Ratified in Planning

1. **Keep the path-bundle separate from `Campaign` — `Campaign.Paths() Paths`.** The path-bundle (today `config.Config`) becomes `campaigns.Paths`. `Campaign` owns identity + root and derives `Paths` on demand. Engine and TUI continue to consume the path-bundle and gain no campaign-layer behavior — they don't reach for `Campaign`, don't branch on campaign identity, don't take any new logic. The seam is **behavioral**, not import-level: the engine *does* import `internal/campaigns` to reference the `Paths` type, because that type physically lives in the campaigns package. An import-level seam would require relocating `Paths` to a neutral upstream package (e.g. `internal/runtime/paths/`) — explicitly considered and declined: the cost of an extra one-type package outweighs the cost of an import-line change in 8 files that already need touching for the type rename. Resolves Discovery OQ1.

2. **`Campaign` composes a `Manifest` field rather than embedding it.** `Campaign{Manifest, Root}` instead of `Campaign{ID, Name, Root}`. Rationale: `Manifest` is the marshalled-to-disk shape; `Campaign` is the in-memory resolved form (Manifest plus root). Composing keeps the on-disk schema isolated in the `Manifest` type so future schema additions don't ripple through every Campaign call-site. Convenience accessors `Campaign.ID()` / `Campaign.Name()` paper over the indirection.

3. **`Paths` is derived by method, not stored as a field.** `Campaign.Paths()` recomputes each call via `filepath.Join` — six joins, no allocations worth caching. Rationale: storing the derived bundle would invite cache-coherence bugs if `Root` is ever rewritten, with zero perf payoff.

4. **Error types: sentinel `errors.Is`-checkable values, not typed error structs.** `ErrNoActiveCampaign`, `ErrCampaignNotRegistered`, `ErrCampaignRootMissing`, `ErrCampaignAlreadyExists`, `ErrRulebookEmpty`, `ErrRulebookConflict`. Callers that need to branch use `errors.Is`; callers that just need the message wrap with `%w`. Rationale: the call sites that matter (Cobra RunE returning to the user) need printable strings plus the ability to render specific hints for the no-active case — sentinels cover that without ceremony. YAGNI on typed error structs.

5. **Test harnesses build synthetic `Campaign` values via `campaigns.Scaffold` + `campaigns.CopyRulebook`.** Tests scaffold a real campaign in `t.TempDir()`, copy the SWN rulebook into it, derive `Paths` from the resulting `Campaign`. Rationale: tests exercise the same load chain production does, so production breakage doesn't pass tests. Couples tests to the campaigns API, which is acceptable because that API is small and stable. Resolves Discovery OQ2.

6. **`create --path` is required in non-interactive mode.** No default. If the positional `id` is supplied but `--path` is not, error out with a hint pointing to both the flag form and the wizard. Rationale: where the campaign lives is the most consequential choice in `create` — silent defaults invite "wait, where did it put it?" support questions. The wizard form (`gm-toolkit campaign create` with no args) prompts for path with a sensible default. Resolves Discovery OQ3.

7. **State file is `state.toml`, not `state.json`.** The discovery doc reads `state.json` (lines 56, 217, 263, 268); current code persists state via `toml.NewEncoder` with TOML-tagged structs, and existing test sites name the file `state.toml`. Plan adopts current reality — no format conversion in this initiative. Discovery doc will get a one-line post-hoc correction during the pre-merge checklist. `history.jsonl` and `narratives.jsonl` are unchanged (they really are JSON lines).

8. **`internal/faction/config/` is deleted in the same commit that introduces `campaigns.Paths`.** All eight consumers (engine/core, engine/orchestrator, engine/orchestrator_apply_test, engine/testharness/harness, engine/testharness/scenarios/spot_check_test, faction/tui/tui, faction/tui/dryrun, faction/tui/adapter/adapter) re-import from the new package atomically. Rationale: avoids a transient state where both packages exist and references diverge; per the `feedback_avoid_type_aliases` memory, no `type X = campaigns.Paths` re-export shim.

---

## Open Questions — To Ratify at Implementation Time

None. All Discovery OQs and the state-format conflict are resolved above. Any new question that surfaces during execution should pause the session per CLAUDE.md item 7 and be ratified in the decisions log.

## Shared Context

### Package layout

```
internal/campaigns/
  campaigns.go        # Manifest, Campaign, Paths; LoadManifest, SaveManifest, LoadCampaign
  registry.go         # Registry, RegistryEntry; LoadRegistry, SaveRegistry, ResolveActive
  scaffold.go         # Scaffold, CopyRulebook
  errors.go           # sentinel errors
  *_test.go           # unit tests per file
```

Lives at the same level as `internal/faction/`, `internal/spatial/`, `internal/logging/`. No dependencies on `internal/faction/` — the campaigns package is upstream of faction.

### CLI layout

```
cmd/gm-toolkit/
  main.go             # func main() only (rootCmd extracted to root.go in Commit 2)
  root.go             # rootCmd var
  faction.go          # gains --campaign override and real load chain in factionRun
  campaign.go         # thin wiring: rootCmd.AddCommand(campaign.Cmd)
  campaign/
    campaign.go       # Cmd var (parent command); subcommands self-register
    create.go         # init() self-registers; run + wizard
    register.go       # init() self-registers; run
    set_active.go     # init() self-registers; run
    add_rules.go      # init() self-registers; run
    helpers.go        # isKebabCase, createWizard
    helpers_test.go
```

### Rulebook layout (relocated)

```
rulebooks/swn/factions/
  assets/
    cunning_assets.toml
    force_assets.toml
    wealth_assets.toml
  drift_costs.toml
  goals.toml
  tags.toml
```

`rulebooks/swn/spatial/` is created as an empty placeholder for F-012 — the directory ships so users seeding from `--rules ./rulebooks/swn/` get the full skeleton.

### Naming conventions

- **`Paths`** (not `Config`) for the path-bundle. Avoids confusion with the toolkit-wide "config" mental model and signals "this is the resolved file-system layout."
- **`Campaign`** for the in-memory resolved value; **`Manifest`** for the on-disk shape. Same pattern as `rulebook.Rulebook` (loaded) vs the per-TOML structs (raw).
- Cobra commands use **kebab-case verbs** matching the discovery doc (`set-active`, `add-rules`).

### `huh` use

Wizard for bare `gm-toolkit campaign create`: id (validated kebab-case), name (default = id), path (default `~/<id>`), rules source (optional). Picker for bare `gm-toolkit campaign set-active`: `huh.Select` of registered campaigns rendered as `<id>  —  <name>`, alphabetized by name, current active highlighted. Per project convention (CLAUDE.md Code Style), interactive forms use `huh`; non-interactive output (echoes, errors) uses plain ANSI / `fmt.Println`.

### Test harness model

After Commit 3, the harness builds campaigns like this:

```go
func NewHarness(t *testing.T) *Harness {
    t.Helper()
    root := t.TempDir()
    camp, err := campaigns.Scaffold(root, "test", "Test")
    if err != nil { t.Fatalf("scaffold: %v", err) }
    if _, err := campaigns.CopyRulebook(camp, "../../../../rulebooks/swn/factions", false); err != nil {
        t.Fatalf("copy rulebook: %v", err)
    }
    paths := camp.Paths()
    rb, err := rulebook.Load(paths.FactionDataDir)
    if err != nil { t.Fatalf("rulebook.Load: %v", err) }
    // ...same engine wiring as today...
}
```

The relative path to `rulebooks/swn/factions/` is resolved from the test file's directory. Tests deeper in the tree compute their own relative path; we don't introduce a `findRepoRoot()` helper (YAGNI — six call-sites, all under the same `_test.go` file shapes).

### Pre-merge checklist additions (initiative-specific)

In addition to the standard CLAUDE.md item 7 / item 8 checklist:

- One-line correction to `docs/initiatives/discovery/campaign-manager-discovery.md` swapping `state.json` → `state.toml` in all four occurrences (lines 56, 217, 263, 268). Flagged here rather than buried in a commit.
- README updated with the new bootstrap workflow.
- Manual verification: fresh `gm-toolkit campaign create` against an empty `~/.gm-toolkit/` produces a usable campaign end-to-end (TUI smoke).

## Out of Scope

- `list`, `info`, `delete`, `rename` campaign subcommands. Filed to Backlog when concrete need surfaces. The `set-active` picker covers most of `list`.
- TUI chrome displaying the active campaign (status-bar entry, header). Deferred to tui-manage's mode-bar work.
- F-012 spatial CLI. `rulebook/spatial/` scaffolded but empty.
- Multi-user / shared-machine concerns (registry is `~`-scoped, single-user by design).
- Multi-rulebook support beyond SWN. `rulebooks/swn/` is named for plural future, but no other rulebooks ship.
- Automated migration of any existing local state. The primary developer migrates by hand.
- `faction.go` subpackage migration — moving `faction.go` into `cmd/gm-toolkit/faction/` following the `campaign/` pattern. Deferred until faction gains subcommands; a single-command file in `package main` doesn't warrant the indirection.

---

## Work Breakdown

Three commits, each = one execution session. (Phase scaffolding dropped — at three commits the per-phase grouping adds nothing.)

### Commit 1 — `feat(campaigns): introduce campaigns package; replace internal/faction/config`

The full campaigns package lands in one commit: `Manifest`, `Campaign`, `Paths`, registry, active resolution, scaffold/copy helpers, the sentinel error set, the migration of all eight `internal/faction/config` consumers to `campaigns.Paths`, and the deletion of the old package. After this commit, the engine builds against the new layer. No CLI surface and no production binary changes yet.

Re-grounding before starting: read current `internal/faction/config/config.go` and all eight consumers grep'd above. Confirm none have grown new fields since this plan was written.

##### Task 1 — `internal/campaigns/campaigns.go` (new)

```go
package campaigns

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/BurntSushi/toml"
)

// Manifest is the on-disk form of campaign.toml.
type Manifest struct {
    Campaign ManifestCampaign `toml:"campaign"`
}

type ManifestCampaign struct {
    ID   string `toml:"id"`
    Name string `toml:"name"`
}

// Campaign is the in-memory resolved campaign: its manifest plus the root path
// it was loaded from.
type Campaign struct {
    Manifest Manifest
    Root     string
}

func (c *Campaign) ID() string   { return c.Manifest.Campaign.ID }
func (c *Campaign) Name() string { return c.Manifest.Campaign.Name }

// Paths is the runtime path-bundle engine consumers receive. Derived from a
// campaign root by fixed-convention concatenation.
type Paths struct {
    FactionDataDir string
    SpatialDataDir string
    StatePath      string
    HistoryPath    string
    NarrativesPath string
    LogsDir        string
}

// Paths derives the runtime path bundle from the campaign root.
func (c *Campaign) Paths() Paths {
    return Paths{
        FactionDataDir: filepath.Join(c.Root, "rulebook", "factions"),
        SpatialDataDir: filepath.Join(c.Root, "rulebook", "spatial"),
        StatePath:      filepath.Join(c.Root, "state", "state.toml"),
        HistoryPath:    filepath.Join(c.Root, "state", "history.jsonl"),
        NarrativesPath: filepath.Join(c.Root, "state", "narratives.jsonl"),
        LogsDir:        filepath.Join(c.Root, "state", "logs"),
    }
}

// LoadManifest reads <root>/campaign.toml.
func LoadManifest(root string) (Manifest, error) {
    var m Manifest
    path := filepath.Join(root, "campaign.toml")
    if _, err := toml.DecodeFile(path, &m); err != nil {
        return Manifest{}, fmt.Errorf("load manifest %s: %w", path, err)
    }
    if m.Campaign.ID == "" {
        return Manifest{}, fmt.Errorf("load manifest %s: missing campaign.id", path)
    }
    return m, nil
}

// SaveManifest writes <root>/campaign.toml, creating the directory if needed.
func SaveManifest(root string, m Manifest) error {
    if err := os.MkdirAll(root, 0o755); err != nil {
        return fmt.Errorf("create campaign root %s: %w", root, err)
    }
    path := filepath.Join(root, "campaign.toml")
    f, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("create manifest %s: %w", path, err)
    }
    defer f.Close()
    if err := toml.NewEncoder(f).Encode(m); err != nil {
        return fmt.Errorf("encode manifest %s: %w", path, err)
    }
    return nil
}

// LoadCampaign reads <root>/campaign.toml and returns a resolved Campaign.
func LoadCampaign(root string) (*Campaign, error) {
    m, err := LoadManifest(root)
    if err != nil {
        return nil, err
    }
    return &Campaign{Manifest: m, Root: root}, nil
}
```

##### Task 2 — `internal/campaigns/errors.go` (new)

```go
package campaigns

import "errors"

var (
    ErrNoActiveCampaign      = errors.New("no active campaign")
    ErrCampaignNotRegistered = errors.New("campaign not registered")
    ErrCampaignRootMissing   = errors.New("campaign root missing on disk")
    ErrCampaignAlreadyExists = errors.New("campaign id already registered")
    ErrRulebookEmpty         = errors.New("rulebook is empty")
    ErrRulebookConflict      = errors.New("rulebook file conflict")
)
```

(`ErrRulebookConflict` is consumed by `add-rules` in Phase 2 but defined here so the sentinels live together. `ErrRulebookEmpty` is consumed by the binary wiring in Phase 3.)

##### Task 3 — `internal/campaigns/campaigns_test.go` (new)

Cover:
- `Campaign.Paths()` for a known root returns the six expected joined paths.
- `LoadManifest` round-trips with `SaveManifest` in a `t.TempDir()`.
- `LoadManifest` errors when `campaign.toml` is missing.
- `LoadManifest` errors when `campaign.id` is empty.

##### Task 4 — Migrate all eight consumers from `internal/faction/config` to `internal/campaigns`

Find with: `grep -rn "faction/config" --include="*.go"` (use Bash; this is a one-shot grep, not ongoing search).

Per file, the change is:
- Import: `"github.com/therobertcrocker/gm-toolkit/internal/faction/config"` → `"github.com/therobertcrocker/gm-toolkit/internal/campaigns"`
- Type references: `*config.Config` → `*campaigns.Paths` (or `campaigns.Paths` for value receivers — match the existing pointer-vs-value usage at each call site).
- Field references: unchanged (FactionDataDir, StatePath, etc. are identical).
- Construction: `&config.Config{...}` → `&campaigns.Paths{...}`.

Files and the construction sites to update:

| File | Site |
|---|---|
| `internal/faction/engine/core.go` | type signature only |
| `internal/faction/engine/orchestrator.go` | type signature only |
| `internal/faction/engine/orchestrator_apply_test.go:21` | literal construction |
| `internal/faction/engine/testharness/harness.go:78` | literal construction (further refactored in Commit 5) |
| `internal/faction/engine/testharness/scenarios/spot_check_test.go` | type reference |
| `internal/faction/tui/tui.go:14` | function signature: `tui.Run(eng, state, cfg *campaigns.Paths, log)` |
| `internal/faction/tui/dryrun.go:40` | literal construction |
| `internal/faction/tui/adapter/adapter.go` | type reference |

##### Task 5 — Delete `internal/faction/config/`

`rm -r internal/faction/config/`. Verify the build passes: `go build ./...`. Verify tests pass: `go test ./...`.

Test harnesses in this commit still build literal `campaigns.Paths` (the scaffold-based migration lands in Commit 5). That's intentional — keep Commit 1 focused on the type swap.

##### Task 6 — `internal/campaigns/registry.go` (new)

```go
package campaigns

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/BurntSushi/toml"
)

// Registry is the in-memory form of ~/.gm-toolkit/campaigns.toml.
type Registry struct {
    Active    string          `toml:"active,omitempty"`
    Campaigns []RegistryEntry `toml:"campaigns"`
}

type RegistryEntry struct {
    ID   string `toml:"id"`
    Path string `toml:"path"`
}

// RegistryPath returns ~/.gm-toolkit/campaigns.toml.
func RegistryPath() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("resolve home dir: %w", err)
    }
    return filepath.Join(home, ".gm-toolkit", "campaigns.toml"), nil
}

// LoadRegistry reads the registry file. Returns an empty Registry if the file is missing.
func LoadRegistry() (*Registry, error) {
    path, err := RegistryPath()
    if err != nil {
        return nil, err
    }
    var reg Registry
    if _, err := toml.DecodeFile(path, &reg); err != nil {
        if os.IsNotExist(err) {
            return &Registry{}, nil
        }
        return nil, fmt.Errorf("load registry %s: %w", path, err)
    }
    return &reg, nil
}

// SaveRegistry writes the registry, creating ~/.gm-toolkit/ if needed.
func SaveRegistry(reg *Registry) error {
    path, err := RegistryPath()
    if err != nil {
        return err
    }
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return fmt.Errorf("create registry dir: %w", err)
    }
    f, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("create registry %s: %w", path, err)
    }
    defer f.Close()
    if err := toml.NewEncoder(f).Encode(reg); err != nil {
        return fmt.Errorf("encode registry %s: %w", path, err)
    }
    return nil
}

// Register adds a campaign to the registry. Returns ErrCampaignAlreadyExists
// if the id is already present.
func (r *Registry) Register(entry RegistryEntry) error {
    for _, e := range r.Campaigns {
        if e.ID == entry.ID {
            return fmt.Errorf("%w: %s", ErrCampaignAlreadyExists, entry.ID)
        }
    }
    r.Campaigns = append(r.Campaigns, entry)
    return nil
}

// SetActive sets the active campaign id. Returns ErrCampaignNotRegistered if
// the id is not present in the registry.
func (r *Registry) SetActive(id string) error {
    if _, ok := r.Lookup(id); !ok {
        return fmt.Errorf("%w: %s", ErrCampaignNotRegistered, id)
    }
    r.Active = id
    return nil
}

// Lookup returns the entry for id, or false if not registered.
func (r *Registry) Lookup(id string) (RegistryEntry, bool) {
    for _, e := range r.Campaigns {
        if e.ID == id {
            return e, true
        }
    }
    return RegistryEntry{}, false
}

// ResolveActive returns the campaign indicated by override (if non-empty) or
// by the registry's Active pointer. Errors:
//   - both empty: ErrNoActiveCampaign
//   - id not in registry: ErrCampaignNotRegistered
//   - registered path missing on disk: ErrCampaignRootMissing
func ResolveActive(reg *Registry, override string) (*Campaign, error) {
    id := override
    if id == "" {
        id = reg.Active
    }
    if id == "" {
        return nil, ErrNoActiveCampaign
    }
    entry, ok := reg.Lookup(id)
    if !ok {
        return nil, fmt.Errorf("%w: %s", ErrCampaignNotRegistered, id)
    }
    if _, err := os.Stat(entry.Path); err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("%w: %s (%s)", ErrCampaignRootMissing, id, entry.Path)
        }
        return nil, fmt.Errorf("stat campaign root %s: %w", entry.Path, err)
    }
    return LoadCampaign(entry.Path)
}
```

##### Task 7 — `internal/campaigns/scaffold.go` (new)

```go
package campaigns

import (
    "fmt"
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
)

// Scaffold creates a fresh campaign directory tree at root and writes a
// campaign.toml manifest. The tree:
//   <root>/rulebook/factions/
//   <root>/rulebook/spatial/
//   <root>/state/logs/
// Errors if root exists and is non-empty.
func Scaffold(root, id, name string) (*Campaign, error) {
    if entries, err := os.ReadDir(root); err == nil && len(entries) > 0 {
        return nil, fmt.Errorf("scaffold: %s is not empty", root)
    } else if err != nil && !os.IsNotExist(err) {
        return nil, fmt.Errorf("scaffold: stat %s: %w", root, err)
    }
    for _, sub := range []string{
        filepath.Join("rulebook", "factions"),
        filepath.Join("rulebook", "spatial"),
        filepath.Join("state", "logs"),
    } {
        if err := os.MkdirAll(filepath.Join(root, sub), 0o755); err != nil {
            return nil, fmt.Errorf("scaffold: create %s: %w", sub, err)
        }
    }
    m := Manifest{Campaign: ManifestCampaign{ID: id, Name: name}}
    if err := SaveManifest(root, m); err != nil {
        return nil, err
    }
    return &Campaign{Manifest: m, Root: root}, nil
}

// CopyRulebook copies every .toml file under srcDir into <campaign-root>/rulebook/
// at the matching relative path. With replace=false, errors with ErrRulebookConflict
// (wrapped, message listing all conflicts) if any destination already exists.
// Returns the count of files copied.
func CopyRulebook(c *Campaign, srcDir string, replace bool) (int, error) {
    destBase := filepath.Join(c.Root, "rulebook")

    var conflicts []string
    var toCopy []struct{ src, dest string }

    err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        if !strings.HasSuffix(strings.ToLower(d.Name()), ".toml") {
            return nil
        }
        rel, relErr := filepath.Rel(srcDir, path)
        if relErr != nil {
            return relErr
        }
        dest := filepath.Join(destBase, rel)
        if _, statErr := os.Stat(dest); statErr == nil {
            conflicts = append(conflicts, rel)
        }
        toCopy = append(toCopy, struct{ src, dest string }{path, dest})
        return nil
    })
    if err != nil {
        return 0, fmt.Errorf("walk %s: %w", srcDir, err)
    }

    if len(conflicts) > 0 && !replace {
        return 0, fmt.Errorf("%w in %s: %s", ErrRulebookConflict, destBase, strings.Join(conflicts, ", "))
    }

    copied := 0
    for _, p := range toCopy {
        if err := copyFile(p.src, p.dest); err != nil {
            return copied, fmt.Errorf("copy %s -> %s: %w", p.src, p.dest, err)
        }
        copied++
    }
    return copied, nil
}

func copyFile(src, dest string) error {
    if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
        return err
    }
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    out, err := os.Create(dest)
    if err != nil {
        return err
    }
    defer out.Close()
    if _, err := io.Copy(out, in); err != nil {
        return err
    }
    return nil
}
```

##### Task 8 — `internal/campaigns/registry_test.go` and `internal/campaigns/scaffold_test.go` (new)

Registry tests cover:
- `LoadRegistry` returns empty Registry when file is missing.
- `SaveRegistry` then `LoadRegistry` round-trips active + entries.
- `Register` rejects duplicate id with `ErrCampaignAlreadyExists`.
- `SetActive` rejects unregistered id with `ErrCampaignNotRegistered`.
- `ResolveActive` cases: empty registry → `ErrNoActiveCampaign`; override → that id; override not registered → `ErrCampaignNotRegistered`; registered but path gone → `ErrCampaignRootMissing`; happy path → returns loaded Campaign.

Scaffold tests cover:
- `Scaffold` creates the full tree and writable manifest in `t.TempDir()`.
- `Scaffold` rejects non-empty root.
- `CopyRulebook` copies a fixture rulebook into a scaffolded campaign and returns the file count.
- `CopyRulebook` with conflicts and `replace=false` returns `ErrRulebookConflict` listing the conflicts.
- `CopyRulebook` with `replace=true` overwrites silently.

Use `rulebooks/swn/factions/` as the fixture source after Commit 5 lands; for this commit, build a small fixture in `t.TempDir()` to avoid coupling to the (not-yet-relocated) data directory.

---

### Commit 2 — `feat(cli): campaign subcommands (create / register / set-active / add-rules)`

Commands live in `cmd/gm-toolkit/campaign/` (package `campaign`); each subcommand file self-registers onto `campaign.Cmd` via its own `init()`. A thin `cmd/gm-toolkit/campaign.go` in `package main` wires `campaign.Cmd` to `rootCmd`. This commit also extracts `rootCmd` from `main.go` to a new `root.go`. After this commit, the user can create, register, switch, and seed campaigns; `gm-toolkit faction` still calls `tui.Run(nil, nil, nil, log)` — production wiring is Commit 3.

Re-grounding: confirm `cmd/gm-toolkit/main.go` still has `rootCmd` inline (no `root.go` exists yet). Re-read the final shapes of `internal/campaigns/scaffold.go` and `internal/campaigns/registry.go` to confirm `Scaffold`, `CopyRulebook`, `LoadRegistry`, `SaveRegistry`, `ResolveActive`, and the sentinel errors haven't drifted since Commit 1.

##### Task 1 — Extract `rootCmd` to `cmd/gm-toolkit/root.go` (new)

Move `var rootCmd = &cobra.Command{...}` verbatim from `main.go` to a new `root.go` in the same package. `main.go` retains only `func main()` and its imports. No logic change.

##### Task 2 — `cmd/gm-toolkit/campaign/campaign.go` (new)

```go
package campaign

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
    Use:   "campaign",
    Short: "Manage campaigns",
    Long:  "Create, register, switch between, and seed gm-toolkit campaigns.",
}
```

No `AddCommand` calls here — each subcommand file adds itself to `Cmd` from its own `init()`.

##### Task 3 — `cmd/gm-toolkit/campaign/helpers.go` (new)

```go
package campaign

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/charmbracelet/huh"
)

func isKebabCase(s string) bool {
    if s == "" {
        return false
    }
    for i, r := range s {
        switch {
        case r >= 'a' && r <= 'z':
        case r >= '0' && r <= '9':
        case r == '-' && i != 0 && i != len(s)-1:
        default:
            return false
        }
    }
    return true
}

func createWizard() (id, name, path, rules string, err error) {
    home, _ := os.UserHomeDir()
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().Title("Campaign id (kebab-case)").Value(&id).Validate(func(s string) error {
                if !isKebabCase(s) {
                    return fmt.Errorf("must be lowercase letters, digits, hyphens")
                }
                return nil
            }),
            huh.NewInput().Title("Display name (default = id)").Value(&name),
            huh.NewInput().Title("Campaign path").Value(&path).Suggestions([]string{filepath.Join(home, id)}),
            huh.NewInput().Title("Rulebook source dir (optional)").Value(&rules),
        ),
    )
    if err = form.Run(); err != nil {
        return
    }
    if name == "" {
        name = id
    }
    if path == "" {
        path = filepath.Join(home, id)
    }
    return
}
```

##### Task 4 — `cmd/gm-toolkit/campaign/create.go` (new)

```go
package campaign

import (
    "fmt"
    "path/filepath"

    "github.com/spf13/cobra"

    "github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

var (
    createName     string
    createPath     string
    createRules    string
    createActivate bool
)

var createCmd = &cobra.Command{
    Use:   "create [id]",
    Short: "Create a new campaign",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runCreate,
}

func init() {
    Cmd.AddCommand(createCmd)
    createCmd.Flags().StringVar(&createName, "name", "", "display name (defaults to id)")
    createCmd.Flags().StringVar(&createPath, "path", "", "campaign directory (required in non-interactive mode)")
    createCmd.Flags().StringVar(&createRules, "rules", "", "directory of rulebook TOML to seed into rulebook/")
    createCmd.Flags().BoolVar(&createActivate, "activate", false, "set as active even if another campaign is already active")
}

func runCreate(cmd *cobra.Command, args []string) error {
    var id, name, path, rules string
    if len(args) == 0 && createPath == "" && createName == "" && createRules == "" {
        var err error
        id, name, path, rules, err = createWizard()
        if err != nil {
            return err
        }
    } else {
        if len(args) == 0 {
            return fmt.Errorf("create: id is required in non-interactive mode\n  hint: gm-toolkit campaign create <id> --path <dir>\n  or:   gm-toolkit campaign create   # interactive wizard")
        }
        id = args[0]
        if !isKebabCase(id) {
            return fmt.Errorf("create: id %q must be kebab-case (lowercase letters, digits, hyphens)", id)
        }
        if createPath == "" {
            return fmt.Errorf("create: --path is required in non-interactive mode\n  hint: gm-toolkit campaign create %s --path ~/games/%s\n  or:   gm-toolkit campaign create   # interactive wizard", id, id)
        }
        path = createPath
        name = createName
        if name == "" {
            name = id
        }
        rules = createRules
    }

    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("create: resolve path: %w", err)
    }

    camp, err := campaigns.Scaffold(absPath, id, name)
    if err != nil {
        return fmt.Errorf("create: %w", err)
    }

    reg, err := campaigns.LoadRegistry()
    if err != nil {
        return fmt.Errorf("create: %w", err)
    }
    if err := reg.Register(campaigns.RegistryEntry{ID: id, Path: absPath}); err != nil {
        return fmt.Errorf("create: %w", err)
    }
    if reg.Active == "" || createActivate {
        if err := reg.SetActive(id); err != nil {
            return fmt.Errorf("create: %w", err)
        }
    }
    if err := campaigns.SaveRegistry(reg); err != nil {
        return fmt.Errorf("create: %w", err)
    }

    fmt.Printf("Created campaign %q at %s\n", id, absPath)
    if reg.Active == id {
        fmt.Printf("active campaign: %s\n", id)
    }

    if rules != "" {
        copied, err := campaigns.CopyRulebook(camp, rules, false)
        if err != nil {
            return fmt.Errorf("create: seed rulebook: %w", err)
        }
        fmt.Printf("Copied %d rulebook file(s) into %s/rulebook/\n", copied, absPath)
    } else {
        fmt.Printf("rulebook/ is empty — seed it with 'gm-toolkit campaign add-rules --rules <path>' or copy TOML manually before running the toolkit\n")
    }
    return nil
}
```

##### Task 5 — `cmd/gm-toolkit/campaign/register.go` (new)

```go
package campaign

import (
    "fmt"
    "path/filepath"

    "github.com/spf13/cobra"

    "github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

var registerActivate bool

var registerCmd = &cobra.Command{
    Use:   "register <path>",
    Short: "Register an existing campaign directory",
    Args:  cobra.ExactArgs(1),
    RunE:  runRegister,
}

func init() {
    Cmd.AddCommand(registerCmd)
    registerCmd.Flags().BoolVar(&registerActivate, "activate", false, "set as active even if another campaign is already active")
}

func runRegister(cmd *cobra.Command, args []string) error {
    absPath, err := filepath.Abs(args[0])
    if err != nil {
        return fmt.Errorf("register: resolve path: %w", err)
    }
    m, err := campaigns.LoadManifest(absPath)
    if err != nil {
        return fmt.Errorf("register: %w", err)
    }
    reg, err := campaigns.LoadRegistry()
    if err != nil {
        return fmt.Errorf("register: %w", err)
    }
    if err := reg.Register(campaigns.RegistryEntry{ID: m.Campaign.ID, Path: absPath}); err != nil {
        return fmt.Errorf("register: %w", err)
    }
    if reg.Active == "" || registerActivate {
        if err := reg.SetActive(m.Campaign.ID); err != nil {
            return fmt.Errorf("register: %w", err)
        }
    }
    if err := campaigns.SaveRegistry(reg); err != nil {
        return fmt.Errorf("register: %w", err)
    }
    fmt.Printf("Registered campaign %q (%s) at %s\n", m.Campaign.ID, m.Campaign.Name, absPath)
    if reg.Active == m.Campaign.ID {
        fmt.Printf("active campaign: %s\n", m.Campaign.ID)
    }
    return nil
}
```

##### Task 6 — `cmd/gm-toolkit/campaign/set_active.go` (new)

```go
package campaign

import (
    "fmt"
    "sort"

    "github.com/charmbracelet/huh"
    "github.com/spf13/cobra"

    "github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

var setActiveCampaign string

var setActiveCmd = &cobra.Command{
    Use:   "set-active",
    Short: "Set the active campaign",
    RunE:  runSetActive,
}

func init() {
    Cmd.AddCommand(setActiveCmd)
    setActiveCmd.Flags().StringVar(&setActiveCampaign, "campaign", "", "campaign id (non-interactive)")
}

func runSetActive(cmd *cobra.Command, args []string) error {
    reg, err := campaigns.LoadRegistry()
    if err != nil {
        return fmt.Errorf("set-active: %w", err)
    }
    if len(reg.Campaigns) == 0 {
        return fmt.Errorf("set-active: no campaigns registered\n  hint: gm-toolkit campaign create <id> --path <dir>")
    }

    id := setActiveCampaign
    if id == "" {
        type displayEntry struct{ id, label string }
        displays := make([]displayEntry, 0, len(reg.Campaigns))
        for _, entry := range reg.Campaigns {
            m, mErr := campaigns.LoadManifest(entry.Path)
            name := entry.ID
            if mErr == nil {
                name = m.Campaign.Name
            }
            displays = append(displays, displayEntry{entry.ID, fmt.Sprintf("%s  —  %s", entry.ID, name)})
        }
        sort.Slice(displays, func(i, j int) bool { return displays[i].label < displays[j].label })
        opts := make([]huh.Option[string], 0, len(displays))
        for _, d := range displays {
            opts = append(opts, huh.NewOption(d.label, d.id))
        }
        if reg.Active != "" {
            id = reg.Active
        }
        sel := huh.NewSelect[string]().Title("Select active campaign").Options(opts...).Value(&id)
        if err := huh.NewForm(huh.NewGroup(sel)).Run(); err != nil {
            return fmt.Errorf("set-active: %w", err)
        }
    }

    if err := reg.SetActive(id); err != nil {
        return fmt.Errorf("set-active: %w", err)
    }
    if err := campaigns.SaveRegistry(reg); err != nil {
        return fmt.Errorf("set-active: %w", err)
    }
    fmt.Printf("active campaign: %s\n", id)
    return nil
}
```

##### Task 7 — `cmd/gm-toolkit/campaign/add_rules.go` (new)

```go
package campaign

import (
    "errors"
    "fmt"

    "github.com/spf13/cobra"

    "github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

var (
    addRulesSource   string
    addRulesCampaign string
    addRulesReplace  bool
)

var addRulesCmd = &cobra.Command{
    Use:   "add-rules",
    Short: "Copy rulebook TOML into a campaign's rulebook/",
    RunE:  runAddRules,
}

func init() {
    Cmd.AddCommand(addRulesCmd)
    addRulesCmd.Flags().StringVar(&addRulesSource, "rules", "", "source directory of TOML files (required)")
    addRulesCmd.Flags().StringVar(&addRulesCampaign, "campaign", "", "target campaign id (defaults to active)")
    addRulesCmd.Flags().BoolVar(&addRulesReplace, "replace", false, "overwrite conflicting files in destination")
    _ = addRulesCmd.MarkFlagRequired("rules")
}

func runAddRules(cmd *cobra.Command, args []string) error {
    reg, err := campaigns.LoadRegistry()
    if err != nil {
        return fmt.Errorf("add-rules: %w", err)
    }
    camp, err := campaigns.ResolveActive(reg, addRulesCampaign)
    if err != nil {
        if errors.Is(err, campaigns.ErrNoActiveCampaign) {
            return fmt.Errorf("add-rules: no active campaign and no --campaign specified\n  hint: gm-toolkit campaign set-active --campaign <id>")
        }
        return fmt.Errorf("add-rules: %w", err)
    }
    copied, err := campaigns.CopyRulebook(camp, addRulesSource, addRulesReplace)
    if err != nil {
        if errors.Is(err, campaigns.ErrRulebookConflict) {
            return fmt.Errorf("add-rules: %w\n  hint: pass --replace to overwrite", err)
        }
        return fmt.Errorf("add-rules: %w", err)
    }
    fmt.Printf("Copied %d file(s) into %s/rulebook/\n", copied, camp.Root)
    return nil
}
```

##### Task 8 — `cmd/gm-toolkit/campaign.go` (new — thin wiring, package main)

```go
package main

import "github.com/therobertcrocker/gm-toolkit/cmd/gm-toolkit/campaign"

func init() {
    rootCmd.AddCommand(campaign.Cmd)
}
```

##### Task 9 — `cmd/gm-toolkit/campaign/helpers_test.go` (new)

Test `isKebabCase` directly. Don't drive Cobra commands from tests — the campaigns package tests cover underlying behavior.

Valid inputs: `"my-camp"`, `"camp1"`, `"a"`, `"foo-bar-baz"`. Invalid: `""`, `"-start"`, `"end-"`, `"has space"`, `"CamelCase"`, `"under_score"`.

##### Task 10 — Manual smoke: create and register

Document in commit body: in a clean `$HOME`, run `gm-toolkit campaign create test-camp --path /tmp/test-camp`. Verify the directory tree, the manifest, the registry, and the "active campaign" echo. (`--rules` requires Commit 3 data — skip or substitute any TOML-bearing dir.)

##### Task 11 — Manual smoke: set-active and add-rules

Document in commit body: register a second campaign; verify `set-active` picker shows both alphabetized with current highlighted; verify `--campaign` switches without prompt; verify `add-rules` conflict error then succeed with `--replace`.

---

### Commit 3 — `feat: wire faction binary to active campaign; relocate rulebook to rulebooks/swn`

Move the rulebook out of the source tree, migrate test harnesses to scaffold-based campaigns, wire `factionRun` to load real arguments from the active campaign, add the `--campaign` override, and document the bootstrap workflow in the README. The "production cutover" commit.

Re-grounding: confirm `internal/faction/data/` still contains exactly the six files listed in the discovery doc. Confirm `internal/faction/engine/testharness/harness.go:75`, the three inline test sites, and the current shape of `cmd/gm-toolkit/faction.go` haven't grown additional path literals or flags since this plan was written.

##### Task 1 — Move files

```
internal/faction/data/cunning_assets.toml → rulebooks/swn/assets/cunning_assets.toml
internal/faction/data/force_assets.toml   → rulebooks/swn/assets/force_assets.toml
internal/faction/data/wealth_assets.toml  → rulebooks/swn/assets/wealth_assets.toml
internal/faction/data/drift_costs.toml    → rulebooks/swn/drift_costs.toml
internal/faction/data/goals.toml          → rulebooks/swn/goals.toml
internal/faction/data/tags.toml           → rulebooks/swn/tags.toml
```

`rm -r internal/faction/data/` after the moves.

##### Task 2 — Migrate `internal/faction/engine/testharness/harness.go`

Replace the `NewHarness(t *testing.T, dataDir string) *Harness` signature with `NewHarness(t *testing.T) *Harness` — the `dataDir` parameter goes away because the harness now scaffolds its own campaign and copies the standard SWN rulebook in.

```go
func NewHarness(t *testing.T) *Harness {
    t.Helper()
    root := t.TempDir()
    camp, err := campaigns.Scaffold(root, "test", "Test")
    if err != nil { t.Fatalf("scaffold: %v", err) }
    _, err = campaigns.CopyRulebook(camp, testRulebookDir(t), false)
    if err != nil { t.Fatalf("copy rulebook: %v", err) }
    paths := camp.Paths()
    rb, err := rulebook.Load(paths.FactionDataDir)
    if err != nil { t.Fatalf("rulebook.Load: %v", err) }
    log := slog.New(slog.DiscardHandler)
    eng := engine.NewWithRulebook(rb, log)
    spatialMap := &StubSpatialMap{}
    eng.World = world.NewWithMap(spatialMap, log)
    actions.RegisterDefaultActions(eng)

    scriptedCollector := &ScriptedCollector{}
    return &Harness{
        Engine:       eng,
        FactionState: &state.FactionState{CampaignID: "test", Factions: make(map[string]*domain.Faction)},
        Paths:        &paths,
        Collector:    scriptedCollector,
        Collectors:   engine.Collectors{Phase: scriptedCollector, Action: scriptedCollector},
        // ... unchanged tail
    }
}

// testRulebookDir locates rulebooks/swn/factions/ relative to this test file.
func testRulebookDir(t *testing.T) string {
    t.Helper()
    _, file, _, _ := runtime.Caller(0)
    repoRoot := filepath.Join(filepath.Dir(file), "..", "..", "..", "..")
    return filepath.Join(repoRoot, "rulebooks", "swn")
}
```

Update the `Harness` struct: `Cfg *config.Config` → `Paths *campaigns.Paths` (the type swap landed in Commit 1; rename the field for clarity).

##### Task 3 — Update call-sites of `NewHarness`

Find with: `grep -rn "testharness.NewHarness\|harness.NewHarness" --include="*.go"`. Each call drops the `dataDir` argument. Update `internal/faction/engine/testharness/scenarios/spot_check_test.go` and any other consumer surfaced by grep.

##### Task 4 — Update `internal/faction/engine/orchestrator_apply_test.go`

The inline `cfg := &campaigns.Paths{...}` at line ~21 currently points at `internal/faction/data/`. Switch to the same scaffold pattern as the harness, or — if the test only needs the data dir — compute the relative path inline. Pick the smaller change; if the test only uses `FactionDataDir`, just update that string.

##### Task 5 — Update `internal/faction/tui/dryrun.go`

Same fix as the harness: scaffold + copy rulebook into `t.TempDir()`-equivalent (or `os.MkdirTemp` — match what dryrun already does). Adjust the construction at line ~40.

##### Task 6 — Update `state.FactionState.CampaignID`

Audit whether the existing `CampaignID` field on `FactionState` (which today carries the literal string `"test"` in tests) should now be populated from the active `Campaign.ID()` at load time. If it's used anywhere consequential, populate it; if not, leave the audit note in the commit body. (This is the kind of small-but-real question that should surface during execution re-grounding — flagged here so the execution session looks for it.)

##### Task 7 — Rewrite `cmd/gm-toolkit/faction.go`

```go
package main

import (
    "errors"
    "fmt"
    "log/slog"
    "os"

    "github.com/spf13/cobra"

    "github.com/therobertcrocker/gm-toolkit/internal/campaigns"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui"
)

var (
    factionDryRun         bool
    factionCampaignOverride string
)

var factionCmd = &cobra.Command{
    Use:   "faction",
    Short: "Open the faction TUI",
    Long:  "Launches the faction-manager TUI against the active campaign. Use --campaign <id> to override the active pointer for this invocation. Use --dryrun for a one-faction smoke cycle.",
    RunE:  factionRun,
}

func init() {
    factionCmd.Flags().BoolVar(&factionDryRun, "dryrun", false, "run a one-faction smoke cycle and exit")
    factionCmd.Flags().StringVar(&factionCampaignOverride, "campaign", "", "campaign id to use instead of the active one")
    rootCmd.AddCommand(factionCmd)
}

func factionRun(cmd *cobra.Command, args []string) error {
    log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

    if factionDryRun {
        return tui.RunDryRun(log)
    }

    reg, err := campaigns.LoadRegistry()
    if err != nil {
        return fmt.Errorf("faction: %w", err)
    }
    camp, err := campaigns.ResolveActive(reg, factionCampaignOverride)
    if err != nil {
        return formatActiveResolveError(err)
    }
    paths := camp.Paths()

    rb, err := rulebook.Load(paths.FactionDataDir)
    if err != nil {
        return fmt.Errorf("faction: load rulebook from %s: %w\n  hint: seed it with 'gm-toolkit campaign add-rules --rules <path>'", paths.FactionDataDir, err)
    }

    factionState, err := state.Load(paths.StatePath)
    if err != nil {
        return fmt.Errorf("faction: load state from %s: %w", paths.StatePath, err)
    }
    if factionState.CampaignID == "" {
        factionState.CampaignID = camp.ID()
    }

    eng := engine.NewWithRulebook(rb, log)

    return tui.Run(eng, factionState, &paths, log)
}

func formatActiveResolveError(err error) error {
    switch {
    case errors.Is(err, campaigns.ErrNoActiveCampaign):
        return fmt.Errorf("faction: no active campaign\n  hint: gm-toolkit campaign create <id> --path <dir>\n  or:   gm-toolkit campaign set-active --campaign <id>")
    case errors.Is(err, campaigns.ErrCampaignNotRegistered):
        return fmt.Errorf("faction: %w\n  hint: gm-toolkit campaign register <path>", err)
    case errors.Is(err, campaigns.ErrCampaignRootMissing):
        return fmt.Errorf("faction: %w\n  hint: re-register at its new location or restore the directory", err)
    default:
        return fmt.Errorf("faction: %w", err)
    }
}
```

##### Task 8 — Update README.md

Add a "Bootstrap" section near the top:

```markdown
## Bootstrap

A fresh checkout doesn't ship a campaign. Create one before running the toolkit:

    gm-toolkit campaign create my-game --path ~/games/my-game --rules ./rulebooks/swn/factions/

This scaffolds the campaign directory, seeds it with the bundled SWN starter rulebook, and registers it as your active campaign. After this, `gm-toolkit faction` opens the TUI against your campaign.

If you receive another GM's campaign directory, register it without modifying its contents:

    gm-toolkit campaign register ~/incoming/their-campaign

See `gm-toolkit campaign --help` for the full command surface.
```

If the repo doesn't have a `README.md` yet (verify with `ls`), create one with that section and a one-line description of the toolkit. Otherwise insert into the existing README at the most relevant point.

##### Task 9 — Manual end-to-end smoke

Document in the commit body:
1. Clean `~/.gm-toolkit/` (`mv ~/.gm-toolkit ~/.gm-toolkit.bak`).
2. `gm-toolkit faction` → errors with "no active campaign" + hint.
3. `gm-toolkit campaign create smoke --path /tmp/smoke --rules ./rulebooks/swn/factions/` → success, sets active.
4. `gm-toolkit faction` → TUI opens with real engine + state + paths.
5. Restore `~/.gm-toolkit/` (`mv ~/.gm-toolkit.bak ~/.gm-toolkit`).

##### Task 10 — Build & test

Full `go build ./...` + `go test ./...` clean. Tests run against `rulebooks/swn/factions/` via the new scaffold helpers.

---

## Pre-Merge Checklist (initiative-specific)

Beyond CLAUDE.md item 7 and item 8 (decisions log update, code review, planned-work update):

- [ ] One-line correction to `docs/initiatives/discovery/campaign-manager-discovery.md`: `state.json` → `state.toml` in all four occurrences (lines 56, 217, 263, 268). Per `feedback_discovery_docs` this is a factual correction, not a rewrite — confirmed scope is one find/replace.
- [ ] README bootstrap section verified by following it on a clean checkout.
- [ ] Manual smoke per Commit 6 Task 3 captured in the merge commit's body.
- [ ] Version bump assessment (CLAUDE.md item 9): this is a user-facing breaking change to `gm-toolkit faction` — minor bump at minimum, possibly major depending on the project's current 0.x vs 1.x stance.

## Model Selection for Execution

Per CLAUDE.md and `feedback_model_selection`:

- **Commit 1** (campaigns package + 8-consumer migration + delete old config): suggest **Opus** at session open. Interlocking signatures across the package plus a wide blast radius — a missed type-swap consumer cascades through the build.
- **Commit 2** (four CLI subcommands): **Sonnet**. Bounded to one file, against a stable campaigns API; mostly Cobra ceremony plus `huh` wiring.
- **Commit 3** (relocate + harness migration + binary wiring + README): suggest **Opus**. Production cutover touching multiple test files, runtime path resolution, and the binary entry. Cost of a missed migration site is the build going red across the suite.
- **Pre-merge checklist**: **Sonnet** (always, per the table in session-modes section 9).

Each execution session should open by re-grounding against the actual code (per session-modes section 9) and confirm the model before starting.
