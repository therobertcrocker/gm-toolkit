# TUI Style System (Bubblegum) — Implementation Plan

## Context / Goal

The `internal/faction/tui` package styles itself inconsistently: `styles/styles.go`
holds app-chrome and modebar styles, but `detail.go` declares ten local style
vars, and raw `lipgloss.Color("#…")` literals are scattered through `manage.go`,
`list.go`, `layout.go`, `deleteconfirm.go`, and `wizard.go`. Colors were an
ad-hoc subset of Catppuccin Mocha plus three off-palette one-offs (`#FFFFFF`,
`#899afa` — a typo of Blue, `#fdf5fb`).

This is a **refactor** (Refactor Flow, session-modes.md section 5) bundled with a
deliberate **palette change**. It restructures styling into a three-layer
design-token system and re-themes the package to a neon cyberpunk-bubblegum
palette. It lands **before** the turn-flow initiative, which will add
substantially more styling and needs a coherent, reusable foundation to build on.

**Target shape:** a three-layer hierarchy with one hard invariant —

> **No raw hex outside the palette layer.**

- **Palette** (`palette.go`) — raw named color swatches. The *only* place hex appears.
- **Semantic tokens** (`tokens.go`) — `lipgloss.Style` values named by role
  (`Strong`, `Dim`, `Accent`, `Danger`…), each referencing the palette.
- **Component styles** (`styles.go`) — composed styles for shared widgets
  (app chrome, modebar, prompts), built from tokens. View-internal one-offs stay
  in their own file but compose from tokens — never from hex.

## Decisions Ratified in Planning

1. **Bubblegum palette (pink-forward synthwave)** — deep purple-black grounds,
   bubblegum magenta as the hero accent, candy cyan secondary. Chosen over a
   yellow-forward (true CP2077) and cyan-forward direction because the brief
   emphasized the *bubblegum* vibe. Full swatch table in Shared Context.
2. **Three-layer structure (palette → tokens → components)** — use-sites name
   *meaning*, not color. The semantic layer is what makes the system reusable for
   the turn flow and re-themeable without touching call-sites.
3. **Centralize shared, tokens for local** — the `styles` package owns the
   palette, the tokens, and genuinely-shared component styles. View-internal
   styles may stay in their file but must compose from tokens. Zero raw hex
   outside `palette.go`.
4. **The neons are accents, not the floor** — saturated colors are reserved for
   accents and state (active mode, primary stat, HP thresholds, prompts). Body
   text uses the calm bright `Text` swatch. This keeps the TUI legible rather
   than exhausting.
5. **Off-palette one-offs are dropped** — `#899afa` (typo) folds into the section
   role, `#FFFFFF`/`#fdf5fb` (titles/values) fold into the `Text` swatch via the
   `Strong` token. No bespoke whites survive.
6. **App banner becomes a neon accent** — "F A C T I O N MANAGER" renders in
   accent magenta bold (was pure-white bold). Deliberate visual change; easy to
   flip back at conversion time if it reads as too much (see Open Questions).
7. **Behavior-changing by design** — unlike a pure refactor, the visible output
   changes (new palette). Decisions log entries at execution should record any
   per-site visual decisions that deviate from the spec below.

---

## Open Questions — To Ratify at Implementation Time

1. **`HealthColor` helper vs inline** — whether the HP-bar threshold→color logic
   becomes a `styles.HealthColor(ratio) lipgloss.Color` helper or stays inline in
   `detail.go` referencing palette swatches. *Resolve: first use in C2.* Lean:
   helper, since the turn flow likely shows HP too.
2. **App-banner color** — magenta accent (specced) vs bright `Text`. *Resolve:
   when converting `view.go` chrome in C2.*
3. **`FormTheme` depth** — how many `huh.Theme` fields to map from the palette
   (minimal: prefixes + focused foreground; full: every field). *Resolve: first
   use in C2.* Lean: map the high-visibility fields (title, selected, focused
   border, prompts), leave the rest at `ThemeCharm` defaults.

## Shared Context

### The Bubblegum palette

The only place these hex values appear in the codebase is `palette.go`.

| Palette name | Hex | Role in the system |
|--------------|-----|--------------------|
| `Bg` | `#14071F` | deep purple-black ground |
| `Surface` | `#241334` | raised plum (active modebar bg) |
| `Border` | `#43275C` | rules, separators, section underlines |
| `Muted` | `#7A6B96` | dim lavender — ids, "none", bullets, hints |
| `Label` | `#B9A3D4` | field labels, descriptions, inactive modes |
| `Text` | `#FCEAFF` | bright body text, names, values |
| `Magenta` | `#FF4FD8` | accent — active mode, section headers, banner |
| `Cyan` | `#2DE2E6` | accent-alt — primary-stat highlight |
| `Mint` | `#05FFA1` | success — HP healthy |
| `Yellow` | `#FCEE0A` | warning — HP warning, placeholders |
| `Red` | `#FF3864` | danger — HP critical, destructive prompts, errors |

Palette identifiers name **hue/ground** (`Magenta`, `Cyan`, `Bg`, `Surface`);
token identifiers name **role** (`Accent`, `Success`, `Dim`). Same package, two
files, no overlap in identifiers.

### Re-grounding (refactor discipline)

C2 converts *existing* code. Per session-modes.md section 5, the C2 execution
session **must re-ground first**: re-read each unit's current styling before
converting, because C1 (and any intervening work) shifts the surrounding code.
The unit list below is a map of *what to convert*, not a frozen transcript of
current line numbers.

### Model selection

- **C1** — Opus suggested (design-dense: palette + token taxonomy + theme mapping).
- **C2** — Sonnet acceptable (mechanical sweep against a fixed spec), Opus fine if
  the `FormTheme` mapping proves fiddly.

Surface the recommendation at the C1→C2 transition.

### Suggested planned-work entry (for Robert to promote)

Not yet added — promotion is Robert's call.

> **Item:** TUI style system (Bubblegum) · **Type:** refactor · **Trigger:** before
> turn-flow initiative · **Detail:** three-layer palette/token/component system;
> re-theme tui to neon cyberpunk-bubblegum; unblocks turn-flow styling.

## Out of Scope

- **Turn-flow styles** — this initiative builds the foundation; the turn flow
  consumes it next and is its own initiative.
- **Non-TUI CLI output** — Cobra/`huh`+ANSI paths are unaffected; `lipgloss` is
  TUI-only (per project convention).
- **Adaptive light/dark theming** — single dark palette only; no
  `lipgloss.AdaptiveColor`.
- **Runtime theme switching / multiple palettes** — one palette, swappable by
  editing `palette.go`. No theme-selection machinery (YAGNI).

---

## Work Breakdown

### Phase 1 — Build the layer, then convert

Two commits: the new authoritative layer (nothing consumes it yet), then one
sweep converting every call-site.

#### Commit 1 — `feat(tui/styles): bubblegum palette and three-layer token system`

The new layer is *invented structure* — specced at full detail here; code conforms
to the doc.

##### Task 1 — `internal/faction/tui/styles/palette.go` (new)

Raw swatches only. The sole location of hex literals.

```go
package styles

import "github.com/charmbracelet/lipgloss"

// Bubblegum — neon cyberpunk-bubblegum palette. The ONLY place hex appears.
var (
	Bg      = lipgloss.Color("#14071F")
	Surface = lipgloss.Color("#241334")
	Border  = lipgloss.Color("#43275C")
	Muted   = lipgloss.Color("#7A6B96")
	Label   = lipgloss.Color("#B9A3D4")
	Text    = lipgloss.Color("#FCEAFF")

	Magenta = lipgloss.Color("#FF4FD8")
	Cyan    = lipgloss.Color("#2DE2E6")
	Mint    = lipgloss.Color("#05FFA1")
	Yellow  = lipgloss.Color("#FCEE0A")
	Red     = lipgloss.Color("#FF3864")
)
```

##### Task 2 — `internal/faction/tui/styles/tokens.go` (new)

Semantic `lipgloss.Style` tokens. Each references the palette; none reference hex.

```go
package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Text weight / brightness
	Strong = lipgloss.NewStyle().Bold(true).Foreground(Text) // names, values, items
	Body   = lipgloss.NewStyle().Foreground(Text)            // default body
	Subtle = lipgloss.NewStyle().Foreground(Label)           // descriptions, scale, inactive
	Dim    = lipgloss.NewStyle().Foreground(Muted)           // ids, "none", bullets, hints

	// Accents
	Accent    = lipgloss.NewStyle().Foreground(Magenta)            // section headers, active mode
	AccentAlt = lipgloss.NewStyle().Foreground(Cyan)               // primary-stat highlight

	// State
	Success = lipgloss.NewStyle().Foreground(Mint)
	Warning = lipgloss.NewStyle().Foreground(Yellow)
	Danger  = lipgloss.NewStyle().Foreground(Red)

	// Structure
	Rule = lipgloss.NewStyle().Foreground(Border) // dividers, separators, section underlines
)
```

##### Task 3 — `internal/faction/tui/styles/styles.go` (rewrite)

Component styles built from tokens. Replaces the current hex-laden file.

```go
package styles

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	// App chrome
	AppTitle    = Accent.Bold(true)         // neon banner (was pure-white bold)
	AppCampaign = Dim
	AppDivider  = Rule

	// Mode bar
	ModeBarActive    = Accent.Bold(true).Background(Surface).Padding(0, 2)
	ModeBarInactive  = Subtle.Padding(0, 2)
	ModeBarDisabled  = Dim.Padding(0, 2)
	ModeBarSeparator = Rule.SetString(" │ ")

	// Section heading (the underline rule is composed at the call-site from Rule)
	SectionHeading = Accent.Bold(true)

	// Prompts
	DangerPrompt = Danger.Bold(true).Padding(1, 2) // quit, delete-confirm
	WarnPrompt   = Warning.Bold(true).Padding(1, 2) // discard-new-faction

	// Misc
	Placeholder = Warning.Padding(2, 4).Italic(true)
	HelpStub    = Dim.Padding(1, 2)
	SaveError   = Danger.Bold(true)
)

// HealthColor maps an HP ratio (0..1) to a state color. (Optional — see Open Q1.)
func HealthColor(ratio float64) lipgloss.Color {
	switch {
	case ratio > 0.5:
		return Mint
	case ratio > 0.25:
		return Yellow
	default:
		return Red
	}
}

// FormTheme adapts huh's Charm theme to the Bubblegum palette and keeps the
// checkbox-prefix convention. Replaces wizard.wizardTheme. (Depth: see Open Q3.)
func FormTheme() *huh.Theme {
	theme := huh.ThemeCharm()
	// map high-visibility fields to palette: focused title/selected -> Magenta,
	// borders -> Border, prompts -> Text; keep [✓]/[ ] prefixes.
	// (exact field assignments determined at first use)
	theme.Focused.SelectedPrefix = theme.Focused.SelectedPrefix.SetString("[✓] ")
	theme.Focused.UnselectedPrefix = theme.Focused.UnselectedPrefix.SetString("[ ] ")
	theme.Blurred.SelectedPrefix = theme.Blurred.SelectedPrefix.SetString("[✓] ")
	theme.Blurred.UnselectedPrefix = theme.Blurred.UnselectedPrefix.SetString("[ ] ")
	return theme
}
```

Notes for execution:
- `ConfirmExit` and `SaveError` are renamed/consolidated: the old `ConfirmExit`
  becomes `DangerPrompt` (same styling, shared with delete-confirm). C2 updates
  the two `view.go` call-sites.
- `lipgloss.Style` methods return copies, so deriving `AppTitle = Accent.Bold(true)`
  does not mutate the `Accent` token. Safe to layer.

**End state of C1:** package compiles; nothing outside `styles` consumes the new
identifiers yet. No behavior change visible.

#### Commit 2 — `refactor(tui): convert all views to the token system`

Each unit below stops declaring/embedding raw color and consumes tokens,
components, or (for bare colors) palette swatches. **Re-ground each unit before
converting.** The invariant to verify on completion: `grep` for
`lipgloss.Color("#"` returns hits *only* in `palette.go`.

Units to convert:

- **`views/manage/detail/detail.go`** — delete all ten local style vars and the
  three hex literals in `hpBar`. Map:
  - `titleStyle`/`valueStyle`/`itemStyle` → `styles.Strong`
  - `idStyle`/`dimStyle` → `styles.Dim`
  - `scaleStyle` → `styles.Subtle`
  - `labelStyle` → `styles.Label`-backed: use `styles.Subtle` for labels, or add a
    dedicated `Label` token if needed at re-grounding (current `labelStyle` is
    `#9399B2`, between Dim and Subtle — map to `styles.Subtle`)
  - `sectionStyle` → `styles.SectionHeading`; `ruleStyle` → `styles.Rule`
  - `primaryStyle` (highest stat) → `styles.AccentAlt`; non-primary → `styles.Strong`
  - `hpBar` fill → `styles.HealthColor(ratio)` (or palette swatches); empty track → `styles.Dim`
  - `wrap(text, indent, color)` call-sites pass `styles.Label` / `styles.Muted`
    (palette swatches) instead of hex.
- **`views/manage/manage.go`** — the centered panel placeholder (`#6C7086`) →
  `styles.Dim` plus its alignment; `SaveError` already a component (rename if C1
  changed it — it did not).
- **`views/manage/list/list.go`** — empty-state hint (`#6C7086`) → `styles.Dim`.
- **`views/manage/layout.go`** — vertical separator (`#6C7086`) → `styles.Rule`.
- **`views/manage/deleteconfirm/deleteconfirm.go`** — inline red prompt →
  `styles.DangerPrompt`.
- **`views/manage/wizard/wizard.go`** — discard prompt (`#F9E2AF` bold) →
  `styles.WarnPrompt`; `wizardTheme()` deleted in favor of `styles.FormTheme()`
  at both `WithTheme(...)` call-sites.
- **`view.go`** — `ConfirmExit` references → `styles.DangerPrompt`; confirm
  `AppTitle`/`AppCampaign`/`AppDivider`/`Placeholder` still resolve (component
  names stable). Ratify the banner color here (Open Q2).
- **`views/modebar/modebar.go`** — already consumes `styles.ModeBar*`; verify no
  change needed beyond the component names staying stable (they do).
- **`views/turn/turn.go`** — consumes `styles.Placeholder`; no change.

**Completion check:**
```
grep -rn 'lipgloss.Color("#' internal/faction/tui --include="*.go"
# expect: only internal/faction/tui/styles/palette.go
```

**End state of C2:** the package renders in the Bubblegum palette; all color flows
from `palette.go` through tokens; the turn flow can build on `styles.*` directly.
