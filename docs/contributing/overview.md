# Contributing — Overview

> The front door for extending GM Toolkit. Setup, build, and the one extension
> pattern both sides share live here; the per-side guides
> ([engine](engine.md), [interface](interface.md)) carry the test tooling,
> conventions, and recipes specific to their half.

## What this is

The contributor guides answer one question: *given a subsystem, what are the
concrete steps to add one more of its kind* — one more action, one more tag
handler, one more overlay. They are organized around **extension recipes**, each
a short procedural unit keyed to the same subsystem axis as the
[architecture wiki](../architecture/architecture-overview.md).

Two side guides, because the codebase has two sides with one shared setup:

- **[engine.md](engine.md)** — the headless faction-turn engine. Read it to add
  an action, a tag/effect/goal handler, a mutation type, or a hook dispatch site.
- **[interface.md](interface.md)** — the TUI projection over the engine. Read it
  to wire a prompt, enable a mode slot, or add a manage sub-view.

A few tasks cross the boundary (adding a GM-driven action is half engine, half
interface). Those recipes carry a **See also** link to their sibling on the other
side; follow it rather than expecting one recipe to span both.

<br/>

### How to read a recipe

Every recipe has the same shape:

```markdown
### <verb> a <thing>

> Shape & why: [<subsystem>](../architecture/<side>/<page>.md)

<One sentence: what extending this gets you.>

1. **`<file path>`** — <what to add / register / wire>.
2. **`<file path>`** — <…>.

**Test:** <harness + what to assert>.

**See also:** <sibling recipe across the boundary, if the task crosses it>.
```

The cross-link under the title carries the *shape and the why* — the recipe never
repeats it. If you want to know why the subsystem is built the way it is before
you extend it, open that link first; the recipe spends its words only on *steps*.

<br/>
<br/>

# Setup

GM Toolkit ships no campaign and no rules of its own. **The GM owns all data; the
binary embeds none** — there are no compiled-in defaults and no hard-coded data
directory. (Shape & why:
[persistence & static data](../architecture/engine/persistence.md).)

- **Go** — the version in [`go.mod`](../../go.mod) (currently 1.26). `go build`
  and `go test ./...` are the only toolchain you need.
- **`$GM_TOOLKIT_HOME`** — the root the toolkit resolves campaigns under. It holds
  `campaigns-registry.toml` (the `{id, path}` list plus the active pointer). An
  unset or empty `GM_TOOLKIT_HOME` is an error, not a fallback.
- **A rulebook** — GM-authored TOML (assets, tags, goals, drift costs) the engine
  reads all turn. The repo bundles a starter set at
  [`rulebooks/swn/`](../../rulebooks/swn/); a campaign gets its own copy at create
  time (below), so editing a campaign's rules never touches the bundled source.

<br/>
<br/>

# Build & run

```sh
go build -o gm-toolkit ./cmd/gm-toolkit
```

The binary is `gm-toolkit` (a Cobra CLI; the faction TUI is one subcommand under
it). A fresh checkout has no campaign, so create one before running:

```sh
gm-toolkit campaign create my-game --path ~/games --rules ./rulebooks/swn/
```

This scaffolds the campaign directory tree, seeds it with the rulebook you point
`--rules` at, and registers it as active. From there:

```sh
gm-toolkit faction          # open the faction TUI against the active campaign
gm-toolkit campaign --help   # the full campaign command surface
```

To use a campaign someone else authored, `gm-toolkit campaign register <path>`
adds it to the registry without copying or modifying its contents.

<br/>
<br/>

# The extension pattern

Four of the engine's sub-engines — **tag**, **effect**, **goal**, and (with a
twist) **action** — share one shape, and learning it once covers most of the
[engine guide](engine.md):

1. **A registry keyed by a data ID.** Each engine holds a `map[string]Handler`
   (or, for actions, an ordered `[]ActionFactory`) populated up front and looked
   up at runtime by a rulebook ID — a tag ID, an asset-definition ID, a goal ID.
   You extend the subsystem by writing a `Handler` and registering it.
2. **Data-only entries silently skip.** A tag, asset, or goal that exists in the
   rulebook TOML with *no* registered Go handler is an intentional no-op, not an
   error. This is the seam that lets a GM add flavor data without writing Go — and
   it means **the data leads the implementation**: the TOML entry can exist before
   the handler does.

The practical consequence for you: to add behavior, the rulebook ID must already
exist in the TOML, *and* you register a handler against it. Miss the data and your
handler is dead code; miss the handler and the data is inert. Both worked
instances live in the arch wiki —
[tags & effects](../architecture/engine/effect-mutation.md) and
[goals](../architecture/engine/goals.md) — and each engine recipe says which file
holds its registration call.

<br/>
<br/>

# Index

- **[engine.md](engine.md)** — engine test tooling, conventions, and six recipes
- **[interface.md](interface.md)** — interface test tooling, conventions, and
  three recipes
- **[Architecture wiki](../architecture/architecture-overview.md)** — what each
  subsystem *is* and why it is shaped that way (every recipe links its page)
