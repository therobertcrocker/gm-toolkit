# Campaign Manager — Discovery

Initiative type: feature. Introduces "campaign" as a first-class concept in the toolkit and resolves F-007 (Campaign-Scoped Static Data) by folding it into this work — there is no existing production consumer of `internal/faction/config/`, so F-007 is not a refactor; it falls out of building the campaign-management layer.

## Problem

The toolkit today has no concept of a campaign. The production binary (`cmd/gm-toolkit/faction.go:35`) calls `tui.Run(nil, nil, nil, log)` — there is no path-construction site in production code at all. Path values exist only in inline literals across three test/dev sites (`internal/faction/tui/dryrun.go:40`, `internal/faction/engine/testharness/harness.go:78`, `internal/faction/engine/orchestrator_apply_test.go:21`), each hardcoding its own paths against `t.TempDir()` or `os.MkdirTemp()`.

Static data lives at `internal/faction/data/` inside the source tree — six TOML files (assets ×3, drift costs, goals, tags) treated as authoritative by the binary's compiled-in assumption. This violates the standing principle "GMs own all rulebook TOML; never embed defaults or hard-code the data path in the binary": the binary effectively does both today, because the data is in the source tree and tests reference it by path.

Three downstream consequences:

1. **The toolkit cannot be used by anyone other than the primary developer.** A second GM checking out the repo gets the same SWN rulebook, hardcoded; there is no mechanism for them to substitute their own. No CLI surface exists for creating, naming, or pointing at a campaign.
2. **A single user cannot run multiple campaigns side-by-side.** State is wherever it happens to be written; rulebook is whatever the binary was built against; there is no switch.
3. **The next TUI initiative (F-005.2, tui-manage) is blocked.** Manage mode adds faction CRUD against persistent state, which requires *somewhere* to persist to and *some* rulebook to validate against. Both have to exist as first-class concepts before tui-manage can be designed against them.

This initiative resolves all three by introducing campaigns as the primary unit of data ownership in the toolkit, building the four-command CLI surface to manage them, and wiring the production binary to resolve and consume the active campaign end-to-end.

## Design Summary

A **campaign** is a user-owned directory containing two top-level subdirectories — `rulebook/` for user-authored static content, `state/` for tool-generated dynamic data — plus a small `campaign.toml` manifest identifying it. Campaigns live wherever the user chose at creation time; the toolkit makes no assumptions about location. A small **registry** at `~/.gm-toolkit/campaigns.toml` maps campaign ids to their paths and tracks which one is "active."

Four CLI commands form the campaign-management surface:

- `create <id>` scaffolds a new campaign directory at the user-supplied `--path`, writes the manifest, registers it, and optionally seeds the rulebook from `--rules <path>`.
- `register <path>` adds an existing campaign directory (received from another GM, restored from backup, etc.) to the registry without modifying its contents.
- `set-active` switches which registered campaign is active; interactive `huh` picker by default, `--campaign <id>` for scripting.
- `add-rules --rules <path>` copies TOML files into a campaign's `rulebook/`; the active campaign by default, or `--campaign <id>` override. Errors on filename conflicts unless `--replace` is passed.

All non-campaign-manager commands (today: `gm-toolkit faction`) resolve the active campaign at invocation, build the engine's `Config` from its root by **fixed layout convention** (no env vars, no per-path configuration), load rulebook and state, and pass real arguments to the TUI. If no campaign is active and no `--campaign` override is supplied, the command fails loud with an actionable hint.

The current `internal/faction/data/*.toml` files relocate to `rulebooks/swn/` at the repo root, becoming a separately-distributable starter pack that the binary does not embed and does not know about by path. The convention is documented; new users seed their campaign with `gm-toolkit campaign create ... --rules ./rulebooks/swn/`. The binary stays content-free.

The `internal/faction/config/` package — currently a six-field path struct with no logic — is deleted. A new `internal/campaigns/` package at the same level as `faction`, `spatial`, and `logging` owns campaign identity, registry I/O, active resolution, and Config construction.

## Per-Area Design Details

### Campaign Identity & On-Disk Layout

Each campaign is a single directory the user owns. Structure:

```
<campaign-root>/
  campaign.toml              # manifest (id, name)
  rulebook/                  # STATIC — user-authored content
    factions/
      assets/
        cunning_assets.toml
        force_assets.toml
        wealth_assets.toml
      drift_costs.toml
      goals.toml
      tags.toml
    spatial/                 # populated when F-012 lands; empty otherwise
  state/                     # DYNAMIC — tool-generated
    state.json
    history.jsonl
    narratives.jsonl
    logs/
```

The `rulebook/` vs `state/` split is structural and load-bearing. The two halves have fundamentally different lifecycles:

- **`rulebook/`** is content the GM owns and may version-control, share between campaigns, or hand to other GMs. Edits are deliberate; lost content is recoverable from a starter pack or another GM's copy.
- **`state/`** is tool-generated, changes every turn, and is irreplaceable if lost. It wants different backup cadence, is typically gitignored, and never travels in shareable rulebook packs.

Encoding the distinction in the directory layout means external tools (backup, git, share-by-tar) can target either half cleanly without per-file exclude rules.

`campaign.toml` schema:

```toml
[campaign]
id = "crimson-lance"           # kebab-case, stable, matches registry key
name = "The Crimson Lance"     # display name shown in set-active picker
```

Two fields. Everything else (description, created date, ruleset version, etc.) is YAGNI — adding fields later costs nothing, and shipping with extra fields invites their misuse.

**The manifest is authoritative.** The registry derives the id↔path mapping from the manifest on `create` and `register`; if the registry is ever corrupted or deleted, it can be rebuilt by re-scanning known paths and reading each `campaign.toml`. The inverse (registry-authoritative manifest-derivative) would mean a registry corruption silently divorces directories from their identities with no recovery path.

### Registry

The registry lives at `~/.gm-toolkit/campaigns.toml`. The path is resolved from `os.UserHomeDir()` joined with `.gm-toolkit/` — no XDG, just the home-dotfolder convention. The directory is created lazily on the first command that needs to write to it.

Schema:

```toml
active = "crimson-lance"

[[campaigns]]
id = "crimson-lance"
path = "/home/robert/SWN/lance"

[[campaigns]]
id = "shadow-protocol"
path = "/home/robert/SWN/shadow"
```

Two responsibilities: track which campaigns are registered (id↔path map) and which one is active (single pointer). Nothing else. Campaign metadata stays in each campaign's manifest.

### Commands

#### `create`

```
gm-toolkit campaign create <id> [--name <name>] [--path <dir>] [--rules <path>] [--activate]
gm-toolkit campaign create                                                                    # interactive wizard
```

Behavior:
1. Resolve `<id>` from positional arg, or open `huh` wizard if no args. The wizard collects id (validated kebab-case) and name (any string).
2. `--name` defaults to `id` when omitted.
3. `--path` default — see Open Question 3 below.
4. Create the campaign directory tree (`rulebook/{factions,spatial}/`, `state/logs/`).
5. Write `campaign.toml` to the campaign root.
6. Register in `~/.gm-toolkit/campaigns.toml`.
7. If `--rules <path>` was provided, copy contents into `rulebook/`. (Rulebook is fresh, so no conflicts are expected; behavior on conflict matches `add-rules`.)
8. If no campaign is currently active, set this one active. Otherwise leave the active pointer alone unless `--activate` was passed.
9. If `--rules` was not provided, print: `rulebook/ is empty — seed it with 'gm-toolkit campaign add-rules --rules <path>' or copy TOML manually before running the toolkit`.

#### `register`

```
gm-toolkit campaign register <path> [--activate]
```

Behavior:
1. Read `<path>/campaign.toml` — error if not present or malformed.
2. Add `{id, path}` entry to registry. If `id` collides with an already-registered campaign, error out and suggest the user remove the existing entry first. (No `--force`; that's `delete`'s job once it exists.)
3. Auto-activate per the same rule as `create`.

#### `set-active`

```
gm-toolkit campaign set-active                    # interactive picker
gm-toolkit campaign set-active --campaign <id>    # non-interactive
```

Behavior:
- Bare invocation: `huh.Select` listing all registered campaigns as `<id>  —  <name>`, alphabetized by name. The currently-active campaign is highlighted.
- `--campaign <id>`: validates that the id is registered, updates the `active` pointer, no prompt.
- Either form echoes: `active campaign: <id>`.

#### `add-rules`

```
gm-toolkit campaign add-rules --rules <path> [--campaign <id>] [--replace]
```

Behavior:
1. Resolve target campaign: `--campaign` if supplied, else active. Error if neither is set.
2. Walk `<path>` recursively, identify TOML files and their relative positions.
3. For each file, compute the destination as `<target-campaign-root>/rulebook/<relative-path>`.
4. If any destination file already exists, abort with the full list of conflicts; require `--replace` to overwrite. With `--replace`, overwrite silently.
5. Echo a summary: `copied <N> files into <campaign>/rulebook/`.

### Bootstrap & Default Rulebook

The current `internal/faction/data/*.toml` content moves to `rulebooks/swn/` at the repo root:

```
<repo-root>/
  rulebooks/
    swn/
      factions/
        assets/
          cunning_assets.toml
          force_assets.toml
          wealth_assets.toml
        drift_costs.toml
        goals.toml
        tags.toml
      spatial/                  # placeholder; populated when F-012 lands
```

This directory ships alongside the binary in the same repo but is **not embedded** in it. Tests that reference `internal/faction/data/` switch to referencing `rulebooks/swn/factions/` directly. No compile-time path is baked into the binary.

A new user's bootstrap workflow:

```
$ gm-toolkit campaign create my-game --path ~/games/my-game --rules ./rulebooks/swn/factions/
```

This creates the campaign, scaffolds the directories, copies the SWN starter pack into `rulebook/`, and auto-activates it (no prior active).

A user with their own custom rulebook simply points `--rules` at their directory instead of `./rulebooks/swn/factions/`. The binary never assumes any source.

### Runtime Resolution

The binary's `factionRun` handler (currently one line: `return tui.Run(nil, nil, nil, log)`) becomes:

1. Read `~/.gm-toolkit/campaigns.toml` via `campaigns.LoadRegistry()`.
2. Resolve active: `campaigns.ResolveActive(registry, cliCampaignOverride)`. Returns `(*campaigns.Campaign, error)`. Errors if no active and no override.
3. Build `Config` from the campaign root via `campaigns.BuildConfig(campaign)` (or the equivalent — see Open Question 1).
4. Load rulebook from `cfg.FactionDataDir` via existing `rulebook.Load(...)`. Errors if the directory is empty or missing required files.
5. Load `FactionState` from `cfg.StatePath` if the file exists; otherwise initialize an empty `FactionState{}` in memory (new campaigns start empty; saved on first state change).
6. Construct engine via existing `engine.NewWithRulebook(rb, log)`.
7. Call `tui.Run(eng, factionState, cfg, log)` — real args, no nils.

Error states:
- **No registry file:** treat as empty registry. Engine commands fail with the no-active-campaign hint.
- **Registry references a missing directory:** error with the path. Suggest `register` again or fixing the path.
- **Active campaign has empty `rulebook/`:** error with hint to run `add-rules`.
- **Active campaign's `state/state.json` is malformed:** error with the parse failure. Do not silently overwrite.

A `--campaign <id>` flag at the engine-command level (`gm-toolkit faction --campaign <id>`) overrides the active pointer for that invocation only. Useful for scripting against a non-active campaign without disturbing the active pointer.

### Path Derivation by Convention

Given a campaign root, the six `Config` paths are computed by fixed concatenation:

| `Config` field    | Derived as                            |
|-------------------|---------------------------------------|
| `FactionDataDir`  | `<root>/rulebook/factions/`           |
| `SpatialDataDir`  | `<root>/rulebook/spatial/`            |
| `StatePath`       | `<root>/state/state.json`             |
| `HistoryPath`     | `<root>/state/history.jsonl`          |
| `NarrativesPath`  | `<root>/state/narratives.jsonl`       |
| `LogsDir`         | `<root>/state/logs/`                  |

No env vars. No config file with overrides. No per-path knobs. The layout *is* the contract, and a campaign directory is self-contained because of it: hand it to another GM and their toolkit knows exactly where to look without any user-side wiring.

### `internal/campaigns/` Package

Replaces `internal/faction/config/`. Lives at the same level as `internal/faction/`, `internal/spatial/`, `internal/logging/`.

Public surface (sketch — exact signatures are a Plan-session decision):

- **`Campaign`** — value type carrying campaign identity and root. At minimum `{ID, Name, Root string}`. Whether the six derived paths live as fields, methods, or are produced by a helper function is open (see Open Question 1).
- **`Manifest`** — TOML-bound type matching `campaign.toml`. Distinct from `Campaign`: a Manifest is the on-disk form; a Campaign is the in-memory resolved form (Manifest + path).
- **`Registry`** — in-memory form of `~/.gm-toolkit/campaigns.toml`. Carries id↔path map and active id.
- **`LoadRegistry() (*Registry, error)`** — reads the registry file. Returns an empty registry if the file is missing.
- **`SaveRegistry(*Registry) error`** — writes the registry file. Creates `~/.gm-toolkit/` if missing.
- **`ResolveActive(reg *Registry, override string) (*Campaign, error)`** — resolves the active campaign or returns the no-active error.
- **`LoadManifest(path string) (Manifest, error)`** — reads `<path>/campaign.toml`.

### CLI Wiring

A new Cobra subcommand `campaign` registers under the root command. Its children: `create`, `register`, `set-active`, `add-rules`. Lives at `cmd/gm-toolkit/campaign.go` alongside `faction.go`.

The `--campaign <id>` flag on `gm-toolkit faction` is added in `cmd/gm-toolkit/faction.go` alongside `--dryrun`.

### TUI Integration

Within this initiative, the only TUI-side change is that `tui.Run` in `cmd/gm-toolkit/faction.go` is called with real `(engine, state, cfg, log)` arguments instead of `nil, nil, nil, log`. Foundation TUI does not currently reach into these values, so the change is invisible at the UI layer; tui-manage inherits a wired-up binary and can focus purely on UI consumption.

Display of the active campaign in the TUI chrome (e.g. a status-bar entry, header banner) is deferred to tui-manage's scope — it's a UX concern that fits naturally with that initiative's mode-bar work.

## Domain Model Changes

New types in `internal/campaigns/`:

- **`Manifest{ID, Name string}`** — on-disk form of `campaign.toml`. TOML-tagged for marshal/unmarshal.
- **`Campaign{ID, Name, Root string}`** — in-memory resolved form. Produced by combining a Manifest with the path it was loaded from.
- **`Registry{Active string, Campaigns []RegistryEntry}`** — in-memory form of `~/.gm-toolkit/campaigns.toml`.
- **`RegistryEntry{ID, Path string}`** — single row in the registry.

Engine signatures stay unchanged; the engine continues to accept a path-bundle struct (renamed from `config.Config` to whichever name lives in `internal/campaigns/`).

## State Storage

New files on disk:

- **`<campaign-root>/campaign.toml`** — per-campaign manifest. Created by `create`; read by `register` and on every active resolution.
- **`~/.gm-toolkit/campaigns.toml`** — global registry. Created lazily by the first command that needs to write it.
- **`<campaign-root>/state/state.json`** — per-campaign state. Existed conceptually before; now lives under a campaign root by convention.
- **`<campaign-root>/state/history.jsonl`** — per-campaign event log. Same as above.
- **`<campaign-root>/state/narratives.jsonl`** and **`<campaign-root>/state/logs/`** — per-campaign tool output.

Relocations:

- `internal/faction/data/*.toml` → `rulebooks/swn/factions/*.toml` (with `assets/` subdirectory preserved).

No mutation surface changes. No state schema changes.

## User-Facing Impact

**New CLI surface:** `gm-toolkit campaign {create, register, set-active, add-rules}`. Four verbs under a new subcommand group.

**Behavior change on `gm-toolkit faction`:** the command now requires an active campaign. Without one, it errors out with an actionable hint. With one, it loads rulebook and state from the active campaign's root.

**New bootstrap workflow:** a fresh checkout no longer "just works." Users (including the primary developer) must `create` a campaign and either seed it with `--rules ./rulebooks/swn/factions/` or run `add-rules` before they can use the toolkit. The README documents this.

**Sharing workflow:** a GM who receives another GM's campaign directory runs `gm-toolkit campaign register <path>` and (if appropriate) `set-active` to point the toolkit at it.

## Replaces / Retires

- `internal/faction/config/` — package deleted entirely. The six-field path struct moves to `internal/campaigns/` under whichever name Plan settles on. References across the codebase are updated.
- `internal/faction/data/` — directory relocated to `rulebooks/swn/factions/` at the repo root. All test sites that referenced the old location are updated. The binary loses its compiled-in assumption that this content exists at any particular path.
- `tui.Run(nil, nil, nil, log)` in `cmd/gm-toolkit/faction.go` — replaced with a real load chain.

## Out of Scope

- **`list`, `info`, `delete`, `rename` commands.** Filed for the Backlog if real need surfaces; barebones V1 ships without them. The `set-active` picker's interactive listing covers most of `list`'s purpose; the others have no concrete user need today.
- **TUI chrome displaying the active campaign.** Deferred to tui-manage scope — more naturally a UX concern alongside the mode-bar work.
- **F-012 (Spatial Map CLI).** The `rulebook/spatial/` directory is scaffolded but stays empty; F-012 will populate it when it lands.
- **Multi-user / shared-machine concerns.** Registry lives in `~/.gm-toolkit/`, scoped per-user. Multi-user on a shared filesystem is not addressed.
- **Multi-rulebook (non-SWN) support.** Only SWN exists today. `rulebooks/swn/` is named with future plurals in mind but no other rulebooks ship in this initiative.
- **Migration tooling.** No automated migration of existing local state. The primary developer's current data (if any) is moved by hand.

## Open Questions

1. **Exact API surface of `internal/campaigns/`.** Signature shape — `Campaign` embedding `Manifest` or composed via field; paths as struct fields vs. methods (`campaign.FactionDataDir` vs `campaign.FactionDataDir()`); error types for the load chain (typed errors? wrapped sentinel errors?); whether the path-bundle struct (`Config` today) survives as a separate type or collapses into `Campaign`. Resolution: Plan session.
2. **Test-harness migration scope.** `internal/faction/engine/testharness/harness.go` and the inline test sites build `Config` literals with hardcoded paths against `t.TempDir()`. Two paths forward: (a) keep building literal Configs with an explicit `dataDir` parameter — minimal change, tests stay decoupled from the campaigns package; (b) switch to building synthetic `Campaign` values in a temp dir — more uniform with production code, but couples tests to the campaigns package's API. Resolution: Plan session.
3. **`--path` default on `create`.** Should it be CWD-relative (`./<id>/`), home-relative (`~/<id>/`), or required (error if omitted in non-interactive mode)? The wizard always prompts with a sensible default; the question is what the non-interactive form does. Resolution: Plan session.

## Reference Exemplars

- **`git`** — the canonical content/state split, per-repo directory, plus global config pattern. Per-repo working tree maps to `rulebook/`; `~/.gitconfig` maps to `~/.gm-toolkit/campaigns.toml`.
- **Obsidian vaults** — per-vault directory the user picks, plus a small registry the application maintains tracking known vaults. Exact match for Pattern B.
- **Cargo workspaces / npm projects** — per-project directory + global state in `$CARGO_HOME` / `$NPM_PREFIX`. Same structural shape, different domain.
