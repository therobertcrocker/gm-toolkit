# Regions & Chrome

> **Code:** `internal/faction/tui/layout/`, `internal/faction/tui/chrome/`, `internal/faction/tui/styles/`; root layout budget in `internal/faction/tui/{model,view,update}.go`

## Purpose

This page covers how the TUI carves the terminal into space: the three-column
working area every mode renders into, the fixed chrome (title, mode bar, status/
help footer) the root draws around it, and the single styling vocabulary all of
it speaks. The load-bearing idea is that the **root owns the layout budget** — it
measures the chrome, hands each sub-model only the height that's left, and the
subs never see the frame around them.

## Shape

### The three-column working area

`layout` defines the working area as three columns identified by a `Region`:

```go
type Region int

const (
	Left Region = iota
	Center
	Right
)
```

A `Panel` pairs the content string for a region with the base `lipgloss.Style`
that governs its alignment and color. `RegionWidths(termWidth)` splits the area
into a wide center flanked by equal quarter-width `Left` and `Right` columns,
reserving two single columns for the vertical separators between them and giving
the remainder to the center. `Compose(panels, widths, height)` renders each
panel into its width at the shared height and joins them left-to-right with `│`
rules between.

The column a component lives in is just which `Region` key its content is filed
under — *"to move a component to a different column, assign its View to a
different `Region` key; widths and separators follow"* (`layout.go`). The layout
package knows nothing about factions, turns, or overlays; it is a pure
content-into-columns renderer, and each mode's [view](views.md) decides what goes
in which column.

### The root chrome and the layout budget

The root `Model` (`view.go`) draws a fixed frame around the working area:

```
─────────────────────────────────────  ← rule
  F A C T I O N   MANAGER      campaign · Cycle n→n+1 · k Factions
─────────────────────────────────────  ← rule
            Manage │ Turn │ Spatial     ← mode bar, centered
                                        ← blank
  <active sub-model's View()>           ← the content area
─────────────────────────────────────  ← rule
  <status line, or help bar>            ← footer
─────────────────────────────────────  ← rule
```

The header is a fixed five rows (`headerHeight = 5` in `update.go`: rule, title,
rule, mode bar, blank). The footer is *not* a constant — its height changes when
the help bar toggles between compact and full — so `contentHeight()` subtracts
`headerHeight + lipgloss.Height(m.footerView())`, measuring the footer from the
rendered bar rather than assuming a number. `footerView` is called by both
`View` and `contentHeight`, so the height reserved always matches the height
drawn.

Whatever height is left is the subs' entire world. `resizeSubs` forwards it as a
synthetic `tea.WindowSizeMsg{Width, Height: contentHeight()}` to every
sub-model, and fires on a real terminal resize *and* whenever help is toggled
(which changes the budget). A sub-model receiving a `WindowSizeMsg` is being told
the size of the content area, never the size of the terminal — it cannot see, and
must not account for, the chrome.

### The status-line / help footer

The footer shows one of two things. If the active sub-model implements
`chrome.StatusLiner` and returns non-empty text, that status line is rendered;
otherwise the footer falls through to the help bar (global key bindings plus the
sub's own `Help()` key map). `StatusLiner` carries a severity the root maps to a
style:

```go
type Severity int // Info, Recoverable, Fatal

type StatusLiner interface {
	StatusLine() (text string, severity Severity)
}
```

`Info → Body`, `Recoverable → Warning`, `Fatal → Danger`. This is how a turn's
recoverable error surfaces in the bottom bar without the sub-model reaching for a
style itself — it reports *what severity*, the root owns *what color*. The
`chrome` package is deliberately tiny (just `Severity` and `StatusLiner`): it is
the contract between subs and the root frame, nothing more.

### The styling vocabulary

Everything visual goes through `styles`, layered in three files:

- **`palette.go`** — the Catppuccin Frappé colors, and *"the ONLY place hex
  appears."* Named roles (`Bg`, `Surface`, `Border`, `Text`, `Mauve`, `Sky`,
  `Teal`, `Yellow`, `Red`).
- **`tokens.go`** — palette colors bound to text-weight and accent roles:
  `Strong`/`Body`/`Subtle`/`Dim` for brightness, `Accent`/`AccentAlt` for
  highlights, `Success`/`Warning`/`Danger` for state, `Rule` for dividers.
- **`styles.go`** — semantic chrome styles composed from the tokens
  (`AppTitle`, the `ModeBar*` set, `DangerPrompt`), plus two helpers: `HealthColor`
  (HP ratio → palette color) and `FormTheme` (huh's Charm theme adapted to the
  palette so overlay forms match the rest of the UI).

The cascade is one-directional: hex lives only in the palette, tokens name the
palette, and the rest of the TUI names tokens. Adding a color means touching the
palette; re-theming a role means touching tokens; no view file holds a raw color.

## Key Decisions

- **The root owns the layout budget; subs never see chrome.** The root measures
  its own header and footer and forwards only the remaining content height to
  sub-models as a `WindowSizeMsg`. A sub-model lays itself out against the size it
  is handed and is oblivious to the title, mode bar, and footer around it. This is
  what lets the chrome change (help toggling height, a future status line) without
  any sub-model knowing. (Frozen log 274–284.)
- **Footer height is measured, not assumed.** Because the help bar's height
  varies (compact vs. full), `contentHeight` measures the rendered footer rather
  than subtracting a constant, and the single `footerView` is shared by `View` and
  `contentHeight` so the reserved and drawn heights can never disagree.
- **A component's column is its `Region` key.** Placement is data, not control
  flow: re-filing a view's content under a different `Region` moves it, and
  `RegionWidths`/`Compose` handle widths and separators. The layout package is
  game-agnostic.
- **Severity in, style out.** Sub-models report a `chrome.Severity` with their
  status text; the root maps severity to a palette style. Subs never choose
  footer colors, which keeps error styling consistent across modes.
- **One styling vocabulary, hex in one place.** All lipgloss flows through the
  `styles` sub-package, layered palette → tokens → semantic styles, with raw hex
  confined to `palette.go`. This is the concrete form of the *lipgloss-only-inside-
  `tui/`* convention. (Frozen log 255–261.)

## Dependencies

**Depends on** `lipgloss` and `huh` (the only place the TUI's styling library is
wrapped) and the `bubbles/help` renderer for the footer's help bar. Nothing in
the engine.

**Depended on by** every interface subsystem that draws. The
[state machine](state-machine.md) page covers the root `Update`/`View` that
assemble the chrome and run the budget; the [views](views.md) compose their
content into `layout.Region`s and surface status through `chrome.StatusLiner`;
the [overlays](overlays.md) render with the same `styles` vocabulary. The
`layout`, `chrome`, and `styles` packages are leaves — they import no other
interface subsystem, which is what keeps them reusable across modes.
