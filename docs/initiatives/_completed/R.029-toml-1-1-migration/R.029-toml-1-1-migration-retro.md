# Migrate Data Files to TOML 1.1 — Retro

## In Plain Words

- **Each rulebook entry is one block now.** The three asset files and `tags.toml` moved
  to TOML 1.1: one header per asset or tag, with `attack`, `ability`, `effects` and
  their filters and gates written in braces under it. A GM reads an asset top to
  bottom, like its rulebook page, and copies or deletes it as one unit.
- **Far fewer headers.** The asset files went from 253 headers to 73 and the tags from
  107 to 20. Every file loads exactly as before.
- **Saves are flatter.** `state.toml` writes each location on one line, so a saved game
  loses its three nested headers per location (115 headers to 31 on the movement
  scenario). It is still TOML, still written by the same encoder.
- **The style is written down.** `docs/contributing/engine.md` holds the house style.
  TomlSmith's Format Document at 120 columns does its layout (line breaks, open lists,
  no alignment); key order and no dotted keys stay by hand. The devcontainer now
  installs TomlSmith in place of Even Better TOML.
- **Smaller changes:** the spatial source lists its regions as one table, both
  sandboxes carry the new files, and live docs say "the `effects` list".
- **Players see no change.** Old 1.0 files still load, so a GM's own files keep working.
- **Next:** `A.001.8` writes its goal parts in this style from the start, and so do
  `A.001.5`'s H.A.W.K assets.

## 1 — Did we do what we set out to do?

Yes. `R.029` has no discovery doc by design; the yardstick is the exploration's
recommendation (H-2 with the house style, G-1, S-1, the Reach leans) and the plan's
In Plain Words. All of it landed, in the eight planned commits.

- **Landed as designed:** H-2 on all four rulebook files, each one checked by decode
  equality against the 1.0 file (`reflect.DeepEqual` on the whole `Rulebook`) and by
  a formatter fixed point. S-1 on `Location` and `RegionHex`, with a round-trip test
  that covers every place state holds a location. The style section, the devcontainer
  switch, the one-table regions, the re-compose of both sandboxes and the rename.
- **G-1 is filed, not executed here.** The goal file has no parts yet; the `A.001.8`
  row carries the amendment to D-17, as the exploration planned.
- **Changed in planning, before any code:** aligned `=` was dropped, because
  TomlSmith strips alignment everywhere (Decision 1), and the 1.1 guard test was
  dropped, because about 20 tests that load the shipped rulebook already guard it
  (Decision 3). Both amend exploration calls, on Robert's call.
- **Cut or deferred:** nothing. Scenarios, seeds, test fixtures and builder output
  stayed out of scope, as planned.
- **Added beyond the plan:** two tracker mentions of `[[effects]]` in the `A.001` arc
  status (E-6), and a key-order fix in five entries that had `ability` before
  `attack`, or `effects` before `ability` (E-4, E-5).

## 2 — Is there a better way we didn't see?

The chosen shape still holds. H-1 would have left about 150 headers that repeat their
entry's key; S-2 (a state writer of our own) would have put a few hundred lines on the
save path to fix what two small methods fix.

- **The plan's shape lists were the weak spot.** Four of the five data-file commits
  logged shapes the plan's table left out (E-2 to E-5): effect parameter tables, flat
  `gates = { once_per = … }`, `ability.relocate`. None of them caused rework, because
  the style's rules covered them and the gate checked the result. A one-line grep of
  every distinct header suffix in the Plan session would have made the lists complete,
  at almost no cost. Decision 7 (pattern and gate, not converted text) was the right
  call; the list was only a hint, and it read as a full inventory.
- **The formatter as the layout check paid off.** It turned the line rule into a
  `diff` that prints nothing. It does not check key order or dotted keys; the review
  caught the style section claiming it did, and the claim was fixed at close. The cost is a 75-line LSP client in the gitignored
  `scratch/` folder, which a later session must rebuild from the completed plan. I
  would keep it out of the repo until a second initiative needs it: `A.001.8` Effort 1
  is that second user, and can decide then.

## 3 — Does it fit the project?

Yes.

- **Persistence Key Decisions:** "TOML for static and state" holds; state did not leave
  TOML (S-3 was rejected for exactly this reason). "GMs own all data" holds: the
  sandboxes moved by `campaign add-rules --replace`, and 1.0 files still load, so no
  GM file breaks.
- **Strict loading:** `refuseUnknownKeys` reads decoder metadata, so it works the same
  for inline tables. The exploration proved it with a misspelt key inside braces.
- **Code style:** both `MarshalTOML` methods are small and named in full words; the one
  judgement call (Go `%q` quoting, not a TOML escaper) carries a fence-sign and a test
  that pins a quote and a backslash.
- **One small divergence, fixed at close:** the `MarshalTOML` doc comments ended with
  "(R.029 S-1)", a provenance tag the code rules ban. The tag is gone; the why stays.
- **Docs moved with the code:** three architecture pages and the contributing guides,
  in the same branch.
- **Process:** the branch is shared with `A.001.8`, on Robert's call, so this close
  does not merge.

## 4 — What does it change about the project?

- **New pattern to copy:** a new rulebook entry is one block. A part is never a header,
  and once a part is inline, a later header cannot add to it; the loader names the line.
- **Hand edits need care with commas and braces.** A slip gives a clear decode error
  with the line, but it is a new kind of error for a GM who knew 1.0.
- **Editor setup matters.** Format on save is off for TOML (Decision 2), so the style
  holds only when someone runs Format Document. The settings apply after a container
  rebuild.
- **Tooling limits:** Even Better TOML (1.0 only) flags the new files, and CUE's TOML
  reader rejects 1.1, so a later CUE schema check needs another reader.
- **State shape:** `Base` and `current_order` still write their own headers. Scenario
  authors can compact them by hand, as before.
- **Docs still out of date:** the completed, discovery and exploration docs keep
  "`[[effects]]`" as written, on purpose (Decision 11).

## 5 — Rating

| Metric | Score | Evidence |
|---|---|---|
| Delivered | 5 | Every part of H-2, S-1 and the Reach leans landed; every file decodes identical |
| Plan held | 4 | No ratified decision reversed in execution; 6 Decision Record rows, all drift in the plan's shape and reach lists |
| Rework | 5 | No fix commits during execution; the review found 6, all low, and the 2 picked were a comment tag and one overclaiming sentence |
| Fit | 5 | Keeps TOML for state and the strict loader; docs moved on the branch; one provenance tag in two comments, fixed at close |
| Simplicity | 5 | Two small methods and data rewrites; no new code path at load |
| **Total** | **24 / 25** | Mean 4.8: a small, well-gated refactor, held back only by the plan's incomplete shape lists |

**Verdict:** A clean data refactor: every file loads the same, reads as one block per
entry, and the formatter checks the layout for free.

**What surprised me:** the formatter decided a house-style call (alignment) that the
exploration had settled the other way, and that was the right outcome.

## Follow-ups

None.
