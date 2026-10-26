# Goals `_core` Home + Strict Tag/Goal Loader — Implementation Plan

## In Plain Words

- **What:** the goals file moves out of the SWN rulebook folder into a shared folder,
  `rulebooks/_core/`. A small manifest in the SWN folder (`rulebook.toml`) names that
  shared folder. When a campaign is created or re-seeded, one compose step copies
  both folders into the campaign, and it checks everything before it writes
  anything. The goal text loses its SWN words ("planet", "planetary government",
  "hex distance"). The tag and goal loader also becomes strict: it refuses unknown
  keys and duplicate IDs.
- **Why:** goals are the same rules in every game this tool runs, so a second
  rulebook (H.A.W.K) must not need its own copy. Shared text must read correctly in
  both settings. The strict loader catches typos and silent overwrites at startup.
- **How it fits:** initiative `A.001.4` in the data-driven-rulebook arc. It has no
  dependency on the registries. A.001.7 and A.001.8 (tags and goals as data) build
  on its strict loader. A.001.5 creates `rulebooks/hawk/`, which includes `_core`.

## Context / Goal

- Discovery doc: [A.001.4-goals-tags-core-discovery.md](../discovery/A.001.4-goals-tags-core-discovery.md) (D-1 to D-13, binding)
- Arc plan: [data-driven-rulebook-arc-plan.md](../arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md), the A.001.4 boundary and AP-8
- Arch pages: [persistence](../../architecture/engine/persistence.md), [goals](../../architecture/engine/goals.md)

One phase, three commits: the strict loader, then the compose step with its test
sites, then the data move, the goal text, the docs and the sandbox re-seed.

## Decisions Ratified in Planning

1. **One phase, three commits** (resolves the discovery's slicing question). C1
   (strict loader) and C2 (compose) each build green on their own: the SWN folder
   has no manifest until C3, so C2 composes it as a plain folder. C3 holds every
   change that needs `_core` to exist. The work is small, so it does not need two
   sessions. The phase is about 1000 lines, past the 800 budget; it is not split,
   because the extra lines are literal file bodies (`scaffold.go`, `goals.toml`).
2. **`CopyRulebook` is renamed to `ComposeRulebook`, same signature.** It is one
   entry point; the name says that it follows includes. A wrapper is rejected: a
   per-folder pass cannot check before it writes (D-1). Keeping the old name is
   rejected because the name would hide the include behavior.
3. **`ErrRulebookDuplicate` wraps `ErrRulebookConflict`.** D-1 holds as written:
   `errors.Is(err, ErrRulebookConflict)` is true for both refusals. `add-rules`
   checks `ErrRulebookDuplicate` first, so its hint does not tell the GM to pass
   `--replace`, which cannot fix a duplicate. Rejected: a separate sentinel (that
   reverses D-1); one sentinel with the old hint (the hint is wrong).
4. **The manifest is strict and small.** File name `rulebook.toml`, one key
   (`includes`). Unknown keys fail, as in `loadAssets`, so `include = [...]` does
   not compose silently without `_core`. A missing or non-directory include is a
   plain wrapped error that names the include and the manifest. It gets no
   sentinel, because no caller branches on it (YAGNI).
5. **Every composed folder's `rulebook.toml` is skipped**, not only the top
   folder's. Includes are one level deep (D-1), so an included folder's manifest is
   neither read nor copied. It is never campaign data.
6. **Check order:** manifest and include errors, then the walk, then duplicates
   (always refused), then campaign clashes (refused unless `replace`), then the
   copy. Each refusal lists every offending file, as the old conflict error did.
7. **The strict loader shares one helper with `loadAssets`.**
   `refuseUnknownKeys(path, metadata)` replaces the inline block in `loadAssets`;
   the existing fence-sign moves with it, word for word. The duplicate errors read
   `duplicate tag ID %q in %s` and `duplicate goal ID %q in %s`, matching the asset
   loader.
8. **Test sites compose with `&campaigns.Campaign{Root: t.TempDir()}`, without
   `Scaffold`.** `copyFile` makes every parent folder. The harness keeps its
   `Scaffold` call. `rulebook_test.go` and `core_test.go` each compose inline,
   because a shared `testharness` helper would make an import cycle (`testharness`
   imports `engine`, which imports `rulebook`). `scenarios/helpers.go` gets its own
   `composedRulebookDir(t)` in place of `testDataDir`.
9. **Goal text (D-13, reviewed by Robert).**
   - G-004 is renamed "World Seizure".
   - "planet" becomes "world" (G-004, G-005), and "planetary government" becomes
     "government" (G-004, G-009).
   - G-012 "hex distance" becomes "the distance between the two worlds". This
     matches the code better than the old text: `ChangeHomeworld.Resolve` sets
     `TurnsRemaining = 1 + world.Distance(...)`, and that distance includes the
     drift crossing cost.
   - **"Base of Influence" stays (G-005, G-012).** This is Robert's call and an
     exception to D-5: the exact UI label stays in the shared text.
10. **TOML table keys follow the renamed goal names** (`[goals.WorldSeizure]`). The
    key is data, not Go, so D-10 does not cover it, and the loader ignores it.
11. **The docs show the data name where they list goals.** In the `goals.md`
    catalogue, G-004 becomes "World Seizure". Prose that names the Go handler
    (`PlanetarySeizure`, "Planetary Seizure's combat phase") stays until A.001.8
    deletes the handler (D-10). `docs/rules/swn-faction-mechanics.md` quotes the SWN
    book and does not change.
12. **Docs that describe the bundled rulebook mention `_core`** (README,
    `contributing/overview.md`, `contributing/engine.md`). The `ability-catalog.md`
    "Open reconciliations" section is deleted: this initiative resolves both of its
    entries.

## Decision Record — Execution

- **C3 / Task 9 — the `sandbox` re-seed is pending Robert.** The `sandbox-claude`
  re-seed ran as planned (`Copied 6 file(s)`, `goals.toml` matches `_core`, no
  manifest copied). The `sandbox` re-seed was blocked: the auto-mode permission
  check refused the overwrite of Robert's campaign. A check before the run showed
  that every file in both sandbox rulebooks was byte-for-byte the same as the
  pre-move `rulebooks/swn/`, so the re-seed only changes `goals.toml`. C3 therefore
  does not hold the `campaigns/sandbox/rulebook/goals.toml` change. Robert runs the
  `sandbox` command from Task 9, and the change lands as its own
  `chore: re-seed sandbox from the composed rulebook` commit on this branch before
  the merge.
- **C3 / Task 10 — smoke check through the create wizard.** No faction in
  `sandbox-claude` has an active goal, so the check used the create wizard's goal
  list (one of the two routes the task names). The TUI started, and the list showed
  "World Seizure". The wizard was discarded; no state changed.

---

## Shared Context

- **Branch:** `feature/goals-tags-core`, cut from `main` at the start of the session.
- **Green check per commit:** `go build ./... && go test ./...`. The baseline at
  plan time is green.
- **Running the app:** only `--campaign sandbox-claude`, through tmux. The one
  exception is the D-12 re-seed of `sandbox` (C3, Task 9).
- **`campaigns/sandbox/` is tracked in git; `campaigns/sandbox-claude/` is
  ignored.** The re-seed changes `campaigns/sandbox/rulebook/goals.toml`, and C3
  commits that change.
- **Pre-merge additions** (CLAUDE.md, Collaboration 9):
  - Graduate to the `persistence.md` Key Decisions: the compose manifest (D-1 and
    Decisions 3 to 6) and the strict tag/goal loader (D-8).
  - The `goals.md` "data-only goals" Key Decision is rewritten in C3; check it
    against the source at pre-merge.
  - `planned-work.md`: remove the `A.001.4` row; change "after `A.001.4`" in the
    A.001.7 and A.001.8 rows.

## Out of Scope

- Tag text and tag behavior (D-7); `tags.toml` does not change.
- Goal behavior, thresholds, XP, locks, and the `E.002` goal rules calls (A.001.8).
- Rulebook ID prefixes (D-9) and Go renames, including `PlanetarySeizure` and
  `planetary_seizure.go` (D-10).
- `rulebooks/hawk/` and its manifest (A.001.5).
- Nested includes, a rulebook file that overrides an included file, and a
  merge-by-record loader (D-1, D-7).
- The unused `ErrRulebookEmpty` sentinel; this plan does not touch it.

---

## Work Breakdown

**Overview**

- **Phase 1 — Shared goals and strict loader**
  - Commit 1 — `feat: refuse unknown keys and duplicate IDs in tags and goals` (3 tasks)
  - Commit 2 — `feat: compose rulebook includes when seeding a campaign` (9 tasks)
  - Commit 3 — `feat: move goals to a shared _core rulebook with setting-free text` (10 tasks)

### Phase 1 — Shared goals and strict loader

**In plain words:** the tag and goal loader refuses typos and duplicate IDs; campaign
seeding follows a rulebook's `includes`; the goals live in `rulebooks/_core/` with
setting-free text; both sandboxes are re-seeded from the new layout.

#### Commit 1 — `feat: refuse unknown keys and duplicate IDs in tags and goals`

##### Task 1 — `internal/faction/rulebook/rulebook.go`

In `loadAssets`, move the inline unknown-key block into a shared helper.

**Find:**
```go
		// deliberate: unknown keys fail the load, not skip, because the decoder drops them silently and a misspelt filter or gate would do nothing.
		if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
			keys := make([]string, len(undecoded))
			for i, key := range undecoded {
				keys[i] = key.String()
			}
			return nil, fmt.Errorf("unknown keys in %s: %s", path, strings.Join(keys, ", "))
		}
		for _, record := range f.Assets {
```
**Replace with:**
```go
		if err := refuseUnknownKeys(path, metadata); err != nil {
			return nil, err
		}
		for _, record := range f.Assets {
```

Add the helper directly above `convertAsset`.

**Find:**
```go
func convertAsset(r assetRecord) (*domain.AssetDefinition, error) {
```
**Replace with:**
```go
// deliberate: unknown keys fail the load, not skip, because the decoder drops them silently and a misspelt filter or gate would do nothing.
func refuseUnknownKeys(path string, metadata toml.MetaData) error {
	undecoded := metadata.Undecoded()
	if len(undecoded) == 0 {
		return nil
	}
	keys := make([]string, len(undecoded))
	for i, key := range undecoded {
		keys[i] = key.String()
	}
	return fmt.Errorf("unknown keys in %s: %s", path, strings.Join(keys, ", "))
}

func convertAsset(r assetRecord) (*domain.AssetDefinition, error) {
```

Make `loadTags` and `loadGoals` strict.

**Find:**
```go
func loadTags(path string) (map[string]*domain.Tag, error) {
	var f tagFile
	if _, err := toml.DecodeFile(path, &f); err != nil {
		return nil, err
	}
	result := make(map[string]*domain.Tag, len(f.Tags))
	for _, r := range f.Tags {
		result[r.ID] = &domain.Tag{
```
**Replace with:**
```go
func loadTags(path string) (map[string]*domain.Tag, error) {
	var f tagFile
	metadata, err := toml.DecodeFile(path, &f)
	if err != nil {
		return nil, err
	}
	if err := refuseUnknownKeys(path, metadata); err != nil {
		return nil, err
	}
	result := make(map[string]*domain.Tag, len(f.Tags))
	for _, r := range f.Tags {
		if _, exists := result[r.ID]; exists {
			return nil, fmt.Errorf("duplicate tag ID %q in %s", r.ID, path)
		}
		result[r.ID] = &domain.Tag{
```

**Find:**
```go
func loadGoals(path string) (map[string]*domain.Goal, error) {
	var f goalFile
	if _, err := toml.DecodeFile(path, &f); err != nil {
		return nil, err
	}
	result := make(map[string]*domain.Goal, len(f.Goals))
	for _, r := range f.Goals {
		result[r.ID] = &domain.Goal{
```
**Replace with:**
```go
func loadGoals(path string) (map[string]*domain.Goal, error) {
	var f goalFile
	metadata, err := toml.DecodeFile(path, &f)
	if err != nil {
		return nil, err
	}
	if err := refuseUnknownKeys(path, metadata); err != nil {
		return nil, err
	}
	result := make(map[string]*domain.Goal, len(f.Goals))
	for _, r := range f.Goals {
		if _, exists := result[r.ID]; exists {
			return nil, fmt.Errorf("duplicate goal ID %q in %s", r.ID, path)
		}
		result[r.ID] = &domain.Goal{
```

##### Task 2 — `internal/faction/rulebook/rulebook_test.go`

Add `"os"` and `"path/filepath"` to the imports.

**Find:**
```go
import (
	"reflect"
	"slices"
	"strings"
	"testing"
```
**Replace with:**
```go
import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
```

Add the tests after `TestLoadDuplicateAssetID`.

**Find:**
```go
	if !strings.Contains(err.Error(), "duplicate asset ID") {
		t.Errorf("error message %q does not mention duplicate asset ID", err.Error())
	}
}
```
**Replace with:**
```go
	if !strings.Contains(err.Error(), "duplicate asset ID") {
		t.Errorf("error message %q does not mention duplicate asset ID", err.Error())
	}
}

func TestLoadTagsAndGoals_Strict(t *testing.T) {
	loadTagFile := func(path string) error { _, err := loadTags(path); return err }
	loadGoalFile := func(path string) error { _, err := loadGoals(path); return err }
	cases := []struct {
		name string
		body string
		load func(path string) error
		want string
	}{
		{
			name: "tag unknown key",
			body: "[tags.Alpha]\nid = \"T-001\"\nname = \"Alpha\"\neffekt = \"none\"\n",
			load: loadTagFile,
			want: "unknown keys in",
		},
		{
			name: "tag duplicate ID",
			body: "[tags.Alpha]\nid = \"T-001\"\n\n[tags.Beta]\nid = \"T-001\"\n",
			load: loadTagFile,
			want: `duplicate tag ID "T-001"`,
		},
		{
			name: "goal unknown key",
			body: "[goals.Alpha]\nid = \"G-001\"\nname = \"Alpha\"\ndificulty = \"low\"\n",
			load: loadGoalFile,
			want: "unknown keys in",
		},
		{
			name: "goal duplicate ID",
			body: "[goals.Alpha]\nid = \"G-001\"\n\n[goals.Beta]\nid = \"G-001\"\n",
			load: loadGoalFile,
			want: `duplicate goal ID "G-001"`,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "data.toml")
			if err := os.WriteFile(path, []byte(testCase.body), 0o644); err != nil {
				t.Fatalf("writing fixture: %v", err)
			}
			err := testCase.load(path)
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("err = %v, want it to contain %q", err, testCase.want)
			}
		})
	}
}
```

##### Task 3 — `docs/architecture/engine/persistence.md`

**Find:**
```markdown
- **Tags, goals, drift costs** load from single files (`tags.toml`, `goals.toml`,
  `drift_costs.toml`).
```
**Replace with:**
```markdown
- **Tags, goals, drift costs** load from single files (`tags.toml`, `goals.toml`,
  `drift_costs.toml`). The tag and goal loaders refuse unknown keys and a
  duplicate ID, as the asset loader does.
```

##### Commit message

```
feat: refuse unknown keys and duplicate IDs in tags and goals

- loadTags and loadGoals fail on unknown keys and on a repeated id
- shared refuseUnknownKeys helper, also used by loadAssets
- persistence.md notes the strict tag and goal loaders
```

#### Commit 2 — `feat: compose rulebook includes when seeding a campaign`

##### Task 1 — `internal/campaigns/campaigns.go`

**Find:**
```go
	ErrRulebookConflict      = errors.New("rulebook file conflict")
)
```
**Replace with:**
```go
	ErrRulebookConflict      = errors.New("rulebook file conflict")
	ErrRulebookDuplicate     = fmt.Errorf("%w: file supplied by two rulebook folders", ErrRulebookConflict)
)
```

##### Task 2 — `internal/campaigns/scaffold.go`

`internal/campaigns/scaffold.go` (rewritten) — full contents:
```go
package campaigns

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

const rulebookManifestFile = "rulebook.toml"

// rulebookManifest is the on-disk form of a rulebook folder's rulebook.toml.
type rulebookManifest struct {
	Includes []string `toml:"includes"`
}

// Scaffold creates a fresh campaign directory tree at root and writes a
// campaign.toml manifest. Directory layout is derived from Paths() so the two
// stay in sync automatically. Errors if root exists and is non-empty.
func Scaffold(root, id, name string) (*Campaign, error) {
	if entries, err := os.ReadDir(root); err == nil && len(entries) > 0 {
		return nil, fmt.Errorf("scaffold: %s is not empty", root)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("scaffold: stat %s: %w", root, err)
	}
	camp := &Campaign{Manifest: Manifest{Campaign: ManifestCampaign{ID: id, Name: name}}, Root: root}
	paths := camp.Paths()
	pv := reflect.ValueOf(paths)
	pt := pv.Type()
	for i := range pt.NumField() {
		field := pt.Field(i)
		full := pv.Field(i).String()
		switch field.Tag.Get("kind") {
		case "dir":
			if err := os.MkdirAll(full, 0o755); err != nil {
				return nil, fmt.Errorf("scaffold: create %s: %w", full, err)
			}
		case "file":
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				return nil, fmt.Errorf("scaffold: create parent for %s: %w", full, err)
			}
		}
	}
	if err := SaveManifest(root, camp.Manifest); err != nil {
		return nil, err
	}
	return camp, nil
}

// ComposeRulebook copies every .toml file under srcDir, and under each folder
// that srcDir's rulebook.toml names in includes, into <campaign-root>/rulebook/
// at the matching relative path. The manifest itself is not copied. Every check
// runs before any write, so a refused compose copies nothing: a file that two
// folders both supply errors with ErrRulebookDuplicate whatever replace says,
// and with replace=false an existing destination errors with
// ErrRulebookConflict. Returns the count of files copied.
func ComposeRulebook(c *Campaign, srcDir string, replace bool) (int, error) {
	folders, err := composeFolders(srcDir)
	if err != nil {
		return 0, err
	}
	destBase := filepath.Join(c.Root, "rulebook")

	suppliedBy := make(map[string]string)
	var duplicates, conflicts []string
	var toCopy []struct{ src, dest string }

	for _, folder := range folders {
		err := filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(strings.ToLower(d.Name()), ".toml") {
				return nil
			}
			rel, relErr := filepath.Rel(folder, path)
			if relErr != nil {
				return relErr
			}
			// deliberate: every folder's manifest is skipped, not only srcDir's, because includes are one level deep (A.001.4 D-1) and a manifest is never campaign data.
			if rel == rulebookManifestFile {
				return nil
			}
			if first, supplied := suppliedBy[rel]; supplied {
				duplicates = append(duplicates, fmt.Sprintf("%s (%s and %s)", rel, first, folder))
				return nil
			}
			suppliedBy[rel] = folder
			dest := filepath.Join(destBase, rel)
			if _, statErr := os.Stat(dest); statErr == nil {
				conflicts = append(conflicts, rel)
			}
			toCopy = append(toCopy, struct{ src, dest string }{path, dest})
			return nil
		})
		if err != nil {
			return 0, fmt.Errorf("walk %s: %w", folder, err)
		}
	}

	if len(duplicates) > 0 {
		return 0, fmt.Errorf("%w: %s", ErrRulebookDuplicate, strings.Join(duplicates, ", "))
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

// composeFolders returns srcDir followed by each folder its rulebook.toml
// includes, resolved from srcDir. With no manifest, srcDir is a plain folder.
func composeFolders(srcDir string) ([]string, error) {
	manifestPath := filepath.Join(srcDir, rulebookManifestFile)
	var manifest rulebookManifest
	metadata, err := toml.DecodeFile(manifestPath, &manifest)
	if errors.Is(err, fs.ErrNotExist) {
		return []string{srcDir}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", manifestPath, err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, key := range undecoded {
			keys[i] = key.String()
		}
		return nil, fmt.Errorf("unknown keys in %s: %s", manifestPath, strings.Join(keys, ", "))
	}
	folders := []string{srcDir}
	for _, include := range manifest.Includes {
		folder := filepath.Join(srcDir, include)
		info, statErr := os.Stat(folder)
		if statErr != nil {
			return nil, fmt.Errorf("include %q in %s: %w", include, manifestPath, statErr)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("include %q in %s: %s is not a directory", include, manifestPath, folder)
		}
		folders = append(folders, folder)
	}
	return folders, nil
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

##### Task 3 — `internal/campaigns/scaffold_test.go`

Rename the existing calls and tests (Edit with `replace_all`):

- `campaigns.CopyRulebook(` → `campaigns.ComposeRulebook(` (5 calls)
- `TestCopyRulebook` → `TestComposeRulebook` (3 test names)
- `t.Fatalf("CopyRulebook: %v", err)` → `t.Fatalf("ComposeRulebook: %v", err)`; `"first CopyRulebook: %v"` → `"first ComposeRulebook: %v"`; `"CopyRulebook with replace=true: %v"` → `"ComposeRulebook with replace=true: %v"`

Add `"strings"` to the imports.

**Find:**
```go
	"path/filepath"
	"testing"
```
**Replace with:**
```go
	"path/filepath"
	"strings"
	"testing"
```

Append the include tests at the end of the file, after `TestComposeRulebookReplaceOverwrites`:
```go

func TestComposeRulebookCopiesIncludes(t *testing.T) {
	fixture := makeTOMLFixture(t, map[string]string{
		"swn/rulebook.toml": "includes = [\"../_core\"]\n",
		"swn/tags.toml":     "[tags]\n",
		"_core/goals.toml":  "[goals]\n",
	})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "include-test", "Include Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	n, err := campaigns.ComposeRulebook(camp, filepath.Join(fixture, "swn"), false)
	if err != nil {
		t.Fatalf("ComposeRulebook: %v", err)
	}
	if n != 2 {
		t.Errorf("copied %d files, want 2", n)
	}
	for _, name := range []string{"tags.toml", "goals.toml"} {
		if _, err := os.Stat(filepath.Join(root, "rulebook", name)); err != nil {
			t.Errorf("%s not copied: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "rulebook", "rulebook.toml")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("rulebook.toml copied into the campaign, want it left out (stat err = %v)", err)
	}
}

func TestComposeRulebookRefusesDuplicateEvenWithReplace(t *testing.T) {
	fixture := makeTOMLFixture(t, map[string]string{
		"swn/rulebook.toml": "includes = [\"../_core\"]\n",
		"swn/tags.toml":     "[tags]\n",
		"swn/goals.toml":    "[goals]\n",
		"_core/goals.toml":  "[goals]\n",
	})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "duplicate-test", "Duplicate Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	_, err = campaigns.ComposeRulebook(camp, filepath.Join(fixture, "swn"), true)
	if !errors.Is(err, campaigns.ErrRulebookDuplicate) {
		t.Fatalf("expected ErrRulebookDuplicate, got %v", err)
	}
	if !errors.Is(err, campaigns.ErrRulebookConflict) {
		t.Errorf("ErrRulebookDuplicate does not wrap ErrRulebookConflict: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "rulebook", "tags.toml")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("tags.toml written by a refused compose (stat err = %v)", err)
	}
}

func TestComposeRulebookRefusesMissingInclude(t *testing.T) {
	fixture := makeTOMLFixture(t, map[string]string{
		"swn/rulebook.toml": "includes = [\"../missing\"]\n",
		"swn/tags.toml":     "[tags]\n",
	})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "missing-test", "Missing Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	_, err = campaigns.ComposeRulebook(camp, filepath.Join(fixture, "swn"), false)
	if err == nil || !strings.Contains(err.Error(), `include "../missing"`) {
		t.Fatalf("err = %v, want a missing-include error", err)
	}
	if _, err := os.Stat(filepath.Join(root, "rulebook", "tags.toml")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("tags.toml written by a refused compose (stat err = %v)", err)
	}
}

func TestComposeRulebookRefusesUnknownManifestKey(t *testing.T) {
	fixture := makeTOMLFixture(t, map[string]string{
		"swn/rulebook.toml": "include = [\"../_core\"]\n",
		"swn/tags.toml":     "[tags]\n",
		"_core/goals.toml":  "[goals]\n",
	})
	root := t.TempDir()
	camp, err := campaigns.Scaffold(root, "manifest-test", "Manifest Test")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	_, err = campaigns.ComposeRulebook(camp, filepath.Join(fixture, "swn"), false)
	if err == nil || !strings.Contains(err.Error(), "unknown keys in") {
		t.Fatalf("err = %v, want an unknown-key error", err)
	}
}
```

##### Task 4 — `cmd/gm-toolkit/campaign/create.go`

**Find:**
```go
				copied, err := campaigns.CopyRulebook(camp, rules, false)
```
**Replace with:**
```go
				copied, err := campaigns.ComposeRulebook(camp, rules, false)
```

##### Task 5 — `cmd/gm-toolkit/campaign/add_rules.go`

**Find:**
```go
			copied, err := campaigns.CopyRulebook(camp, source, replace)
			if err != nil {
				if errors.Is(err, campaigns.ErrRulebookConflict) {
```
**Replace with:**
```go
			copied, err := campaigns.ComposeRulebook(camp, source, replace)
			if err != nil {
				if errors.Is(err, campaigns.ErrRulebookDuplicate) {
					return fmt.Errorf("add-rules: %w\n  hint: remove the file from one of the rulebook folders", err)
				}
				if errors.Is(err, campaigns.ErrRulebookConflict) {
```

**Find:**
```go
	cmd.Flags().StringVar(&source, "rules", "", "source directory of TOML files (required)")
```
**Replace with:**
```go
	cmd.Flags().StringVar(&source, "rules", "", "rulebook directory to copy, with the folders its rulebook.toml includes (required)")
```

##### Task 6 — `internal/faction/engine/testharness/harness.go`

**Find:**
```go
	if _, err := campaigns.CopyRulebook(camp, testRulebookDir(t), false); err != nil {
		t.Fatalf("copy rulebook: %v", err)
	}
```
**Replace with:**
```go
	if _, err := campaigns.ComposeRulebook(camp, testRulebookDir(t), false); err != nil {
		t.Fatalf("compose rulebook: %v", err)
	}
```

##### Task 7 — `internal/faction/engine/testharness/scenarios/helpers.go` and `spot_check_test.go`

`internal/faction/engine/testharness/scenarios/helpers.go` (rewritten) — full contents:
```go
package scenarios

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

// composedRulebookDir composes rulebooks/swn/ and the folders it includes into a
// temp campaign and returns that campaign's rulebook dir. The source is resolved
// from this file so it works regardless of the working directory when tests run.
func composedRulebookDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "..")
	camp := &campaigns.Campaign{Root: t.TempDir()}
	if _, err := campaigns.ComposeRulebook(camp, filepath.Join(repoRoot, "rulebooks", "swn"), false); err != nil {
		t.Fatalf("compose rulebook: %v", err)
	}
	return camp.Paths().FactionDataDir
}
```

In `spot_check_test.go`, `TestSpotCheck_RenderedOutput`:

**Find:**
```go
			FactionDataDir: testDataDir,
```
**Replace with:**
```go
			FactionDataDir: composedRulebookDir(t),
```

##### Task 8 — `internal/faction/rulebook/rulebook_test.go` and `internal/faction/engine/core_test.go`

In `rulebook_test.go`, add the campaigns import.

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

const testDataDir = "../../../rulebooks/swn"
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// loadShippedRulebook composes rulebooks/swn/ and the folders it includes into a
// temp campaign, as campaign create does, and loads it.
func loadShippedRulebook(t *testing.T) *Rulebook {
	t.Helper()
	camp := &campaigns.Campaign{Root: t.TempDir()}
	if _, err := campaigns.ComposeRulebook(camp, "../../../rulebooks/swn", false); err != nil {
		t.Fatalf("compose shipped rulebook: %v", err)
	}
	rulebook, err := Load(camp.Paths().FactionDataDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	return rulebook
}
```

Replace every occurrence (4: `TestLoad`, `TestLoad_TransportBypasses`,
`TestLoad_LobbyistsWithdrawPermission`, `TestLoad_TransportLockdown`; Edit with
`replace_all`):

**Find:**
```go
	rb, err := Load(testDataDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
```
**Replace with:**
```go
	rb := loadShippedRulebook(t)
```

In `core_test.go`, add `"github.com/therobertcrocker/gm-toolkit/internal/campaigns"`
to the imports (above the `domain` import), then in
`TestNewWithRulebook_ShippedRulebook`:

**Find:**
```go
	rb, err := rulebook.Load("../../../rulebooks/swn")
	if err != nil {
		t.Fatalf("load shipped rulebook: %v", err)
	}
```
**Replace with:**
```go
	camp := &campaigns.Campaign{Root: t.TempDir()}
	if _, err := campaigns.ComposeRulebook(camp, "../../../rulebooks/swn", false); err != nil {
		t.Fatalf("compose shipped rulebook: %v", err)
	}
	rb, err := rulebook.Load(camp.Paths().FactionDataDir)
	if err != nil {
		t.Fatalf("load shipped rulebook: %v", err)
	}
```

##### Task 9 — `docs/architecture/engine/persistence.md`

**Find:**
```markdown
(registered but the directory is gone). Creating a campaign is `Scaffold` (the
tree + a `campaign.toml` manifest) followed by `CopyRulebook`, which walks a
source rulebook directory and copies every `.toml` into the campaign's
`rulebook/`, refusing to clobber existing files unless `replace` is set.
```
**Replace with:**
```markdown
(registered but the directory is gone). Creating a campaign is `Scaffold` (the
tree + a `campaign.toml` manifest) followed by `ComposeRulebook`. It walks a
source rulebook directory, plus each folder that the directory's `rulebook.toml`
manifest names in `includes` (one level deep), and copies every `.toml` except
the manifest into the campaign's `rulebook/`. Every check runs before any write:
a file that two folders both supply is always refused (`ErrRulebookDuplicate`),
and an existing campaign file is refused unless `replace` is set
(`ErrRulebookConflict`). A folder with no manifest copies as a plain folder.
```

##### Commit message

```
feat: compose rulebook includes when seeding a campaign

- CopyRulebook becomes ComposeRulebook: follows rulebook.toml includes,
  checks every folder before any write, never copies the manifest
- ErrRulebookDuplicate (wraps ErrRulebookConflict) refuses a file two
  folders supply, even with --replace; add-rules gives it its own hint
- shipped-rulebook tests and the harness load through the real compose step
- persistence.md describes the compose step
```

#### Commit 3 — `feat: move goals to a shared _core rulebook with setting-free text`

##### Task 1 — move `goals.toml`

```sh
mkdir -p rulebooks/_core
git mv rulebooks/swn/goals.toml rulebooks/_core/goals.toml
```

##### Task 2 — `rulebooks/_core/goals.toml`

`rulebooks/_core/goals.toml` (rewritten) — full contents. G-004, G-005, G-009 and
G-012 change; the other goals are byte-for-byte the same.
```toml
[goals.MilitaryConquest]
id = "G-001"
name = "Military Conquest"
description = "Destroy a number of Force assets of rival factions equal to your faction's Force rating."
difficulty = "half_assets_destroyed"

[goals.CommercialExpansion]
id = "G-002"
name = "Commercial Expansion"
description = "Destroy a number of Wealth assets of rival factions equal to your faction's Wealth rating."
difficulty = "half_assets_destroyed"

[goals.IntelligenceCoup]
id = "G-003"
name = "Intelligence Coup"
description = "Destroy a number of Cunning assets of rival factions equal to your faction's Cunning rating."
difficulty = "half_assets_destroyed"

[goals.WorldSeizure]
id = "G-004"
name = "World Seizure"
description = "Take control of a world, becoming the world's legitimate government. If no opposing faction resists, counts as difficulty low."
difficulty = "half_avg_ruling_faction"

[goals.ExpandInfluence]
id = "G-005"
name = "Expand Influence"
description = "Plant a Base of Influence on a new world. Difficulty increases by one if the attempt is contested by a rival faction."
difficulty = "low_plus_contested"

[goals.BloodTheEnemy]
id = "G-006"
name = "Blood the Enemy"
description = "Inflict a number of hit points of damage on enemy faction assets or bases equal to your faction's total Force, Cunning, and Wealth ratings."
difficulty = "moderate"

[goals.PeaceableKingdom]
id = "G-007"
name = "Peaceable Kingdom"
description = "Do not take an Attack action for four consecutive turns."
difficulty = "low"

[goals.DestroyTheFoe]
id = "G-008"
name = "Destroy the Foe"
description = "Destroy a rival faction entirely."
difficulty = "one_plus_avg_target"

[goals.InsideEnemyTerritory]
id = "G-009"
name = "Inside Enemy Territory"
description = "Have a number of stealthed assets on worlds with other governments equal to your Cunning score. Units already stealthed when this goal is adopted do not count."
difficulty = "moderate"

[goals.InvincibleValor]
id = "G-010"
name = "Invincible Valor"
description = "Destroy a Force asset with a minimum purchase rating higher than your faction's Force rating."
difficulty = "moderate"

[goals.WealthOfWorlds]
id = "G-011"
name = "Wealth of Worlds"
description = "Spend Coins equal to four times your faction's Wealth rating on bribes and influence. The faction's Wealth rating must increase before this goal can be selected again."
difficulty = "moderate"

[goals.ChangeHomeworld]
id = "G-012"
name = "Change Homeworld"
description = "Relocate the faction's homeworld to another world where it has a Base of Influence. The move takes a number of turns equal to the distance between the two worlds plus one."
difficulty = "none"
```

##### Task 3 — `rulebooks/swn/rulebook.toml`

`rulebooks/swn/rulebook.toml` (new) — full contents:
```toml
# Folders composed into a campaign with this rulebook. Each path resolves from this folder.
includes = ["../_core"]
```

##### Task 4 — `docs/architecture/engine/goals.md`

**Find:**
```markdown
the tag and effect engines use. A faction's `ActiveGoal.GoalID` with no
registered handler is a **data-only goal** — present in `goals.toml` for flavor,
silently skipped, not an error (the fence-signed twin of the tag/effect skip).
```
**Replace with:**
```markdown
the tag and effect engines use. A faction's `ActiveGoal.GoalID` with no
registered handler is a **data-only goal**: it is silently skipped, not an error
(the fence-signed twin of the tag/effect skip). The goal definitions live in the
shared `rulebooks/_core/goals.toml`, which the SWN rulebook composes into its
campaigns.
```

**Find:**
```markdown
`PlanetarySeizure` (G-004) exercises every mechanism across a multi-turn
```
**Replace with:**
```markdown
`PlanetarySeizure` (G-004, named World Seizure in `goals.toml`) exercises every mechanism across a multi-turn
```

**Find:**
```markdown
| Planetary Seizure (G-004) | Multi-phase: clear, then occupy the target world | Occupation countdown elapses (grants the T-011 tag) |
```
**Replace with:**
```markdown
| World Seizure (G-004) | Multi-phase: clear, then occupy the target world | Occupation countdown elapses (grants the T-011 tag) |
```

**Find:**
```markdown
- **Data-only goals are intentionally no-ops.** A goal ID in `goals.toml` with no
  registered handler returns `LockNone` and no progress — flavor data a GM can add
  without writing Go, the same silent-skip contract as tags and effects.
```
**Replace with:**
```markdown
- **Data-only goals are a silent skip, as an incremental seam.** A goal ID in
  `goals.toml` with no registered handler returns `LockNone` and no progress, the
  same silent-skip contract as tags and effects. This is not an end state: a goal
  a GM adds without Go does nothing until goal behavior comes from data (A.001.8,
  planned).
```

##### Task 5 — `docs/architecture/engine/ability-catalog.md`

Delete the section; this initiative resolves both entries.

**Find:**
```markdown
## Open reconciliations (carried into the arc)

- **`goals.md` "intentional flavor" wording** — the data-only skip is an
  incremental seam, not a deliberate end state; reconcile the wording at
  pre-merge (already flagged in the discovery).
- **Shared goals/tags source home** — every ruleset composes the *same* goals and
  tags into its campaign dir. A `campaigns`/scaffold change (how `Scaffold` /
  `CopyRulebook` compose), not a loader change. Separate arc effort.

## Dependencies
```
**Replace with:**
```markdown
## Dependencies
```

##### Task 6 — `README.md`

**Find:**
```markdown
This scaffolds the campaign directory, seeds it with the bundled SWN starter rulebook, and registers it as your active campaign. After this, `gm-toolkit faction` opens the TUI against your campaign.
```
**Replace with:**
```markdown
This scaffolds the campaign directory, seeds it with the bundled SWN starter rulebook, and registers it as your active campaign. After this, `gm-toolkit faction` opens the TUI against your campaign.

The SWN rulebook's `rulebook.toml` pulls in the shared goals from `rulebooks/_core/`, so the campaign gets both.
```

##### Task 7 — `docs/contributing/overview.md`

**Find:**
```markdown
  reads all turn. The repo bundles a starter set at
  [`rulebooks/swn/`](../../rulebooks/swn/); a campaign gets its own copy at create
  time (below), so editing a campaign's rules never touches the bundled source.
```
**Replace with:**
```markdown
  reads all turn. The repo bundles a starter set at
  [`rulebooks/swn/`](../../rulebooks/swn/), which pulls in the shared goals from
  [`rulebooks/_core/`](../../rulebooks/_core/); a campaign gets its own copy at
  create time (below), so editing a campaign's rules never touches the bundled
  source.
```

##### Task 8 — `docs/contributing/engine.md`

**Find:**
```markdown
campaign, copies the bundled `rulebooks/swn/` rulebook, and builds a *live* engine
```
**Replace with:**
```markdown
campaign, composes the bundled `rulebooks/swn/` rulebook (with the shared
`rulebooks/_core/` goals), and builds a *live* engine
```

**Find:**
```markdown
real `rulebooks/swn/` at the repo root, so your data ID must exist there (or in a
test rulebook you point the harness at) for the handler to fire.
```
**Replace with:**
```markdown
real `rulebooks/swn/` at the repo root, composed with `rulebooks/_core/`, so your
data ID must exist in one of them (or in a test rulebook you point the harness
at) for the handler to fire.
```

##### Task 9 — re-seed both sandboxes (D-12)

Run after `go test ./...` is green. Each command must print
`Copied 6 file(s)` (SWN: three asset files, `drift_costs.toml`, `tags.toml`;
`_core`: `goals.toml`).

```sh
go run ./cmd/gm-toolkit campaign add-rules --rules ./rulebooks/swn --replace --campaign sandbox-claude
go run ./cmd/gm-toolkit campaign add-rules --rules ./rulebooks/swn --replace --campaign sandbox
echo "==== goals match _core ===="
diff rulebooks/_core/goals.toml campaigns/sandbox-claude/rulebook/goals.toml
diff rulebooks/_core/goals.toml campaigns/sandbox/rulebook/goals.toml
echo "==== no manifest copied ===="
ls campaigns/sandbox-claude/rulebook/rulebook.toml campaigns/sandbox/rulebook/rulebook.toml 2>&1
```

Both `diff` commands print nothing; `ls` reports both manifests missing.

##### Task 10 — smoke-check `sandbox-claude`

Open the TUI through tmux with `--campaign sandbox-claude`, confirm it starts (the
strict loader accepts the composed rulebook), and open Manage on a faction with
an active goal or the create wizard's goal list to see the new goal text. Close
the session. `sandbox` is not run.

##### Commit message

```
feat: move goals to a shared _core rulebook with setting-free text

- rulebooks/_core/goals.toml holds the goals; rulebooks/swn/rulebook.toml
  includes it
- goal text drops SWN words: World Seizure (G-004), "world" and "government"
  for planet terms, G-012 distance between the two worlds
- goals.md: data-only goals are an incremental seam; catalogue shows the
  data name; ability-catalog open reconciliations resolved
- README and contributing docs name the shared _core folder
- re-seed sandbox and sandbox-claude from the composed rulebook
```
