# Roll Card — Retro

## In Plain Words

- **What works now:** on each stat roll with someone to ask, the GM sees a card before the
  dice, with who rolls, for what, against whom, and each extra die as a toggle. An offer that
  cannot be taken shows too, with the reason. After the dice, a second card shows them, the
  total, the total to beat and each reroll on offer with its owner. With no reroll on offer,
  the dice go to the feed with no stop. Each roll leaves one line in the feed.
- **Psychic Academy is live:** the last SWN tag with no data entry can now force a reroll on
  a rival's roll, in the rulebook and the sandbox copy.
- **Fixed along the way:** a canceled action no longer costs a once-per-turn use, and Book of
  Secrets has one shared use per turn. A wrong reply type from a prompt is now an error in
  every collector method, not a silent zero value. The feed wraps long lines.
- **Removed:** the two old thin prompts and their overlay file.
- **Found, not fixed:** the first test through the real channels proves the `B.038` data
  race, so `go test -race` on the execution view stays red until that row is fixed.

## 1 — Did we do what we set out to do?

Yes. Each part of the discovery's In Plain Words and Design Summary shipped.

**Landed:** the before card on each hooked roll with a collector, the after card that waits
only on a choice, the forced-then-elective reroll order, struck offers with their reasons,
the wording in the adapter and the drawing in the overlay (D-12), and Psychic Academy's
entry. All of D-11's scope landed too: `F.032`, `B.016`, `B.024`, `R.015`, the card's part
of `B.027`, `B.029` and `B.043`, and the planned part of `F.027`.

**Narrowed:** a nil registry asks no one (Decision 14). Production always has a registry, so
no real roll loses its cards. Struck offers show only on the before card, and a spent reroll
stays silent, so the after card never waits with nothing to choose (Decision 5).

**Cut or deferred:** nothing from scope. Two gaps went to the tracker as `B.046` (the footer
under overlays) and `B.047` (a full card can overrun the band).

**Added, not planned:**

- Cards sit flush left, draw their own keys, and the feed wraps for every event; the manual
  check found all three (Decision Record 5 to 7)
- Expand Influence's rival cards show the expander's total (Decision 6)
- `tmux` in `.devcontainer/postcreate.sh` (Decision Record 19)
- Two stale architecture claims corrected, past the plan's `D.010` line (Decision Record 19)
- One test fix commit, `d515d00`, for a flaky Psychic Academy scenario

**One thing not shown:** Psychic Academy's row has four harness scenarios, but the Decision
Record logs no TUI run with that row on a card. No sandbox seed holds `SWN-T-014`, and
`TESTING.md` has no step for it, though Decision 9 put the entry in the sandbox copy so it
could be tested there.

## 2 — Is there a better way we didn't see?

The chosen way still holds. `R.028` put every hooked roll through `rollWithHooks`, so the
engine change stayed in one function, and the hand-drawn cards removed `B.029` and the
no-offers special case with no form to fight.

Two smaller spots:

- **Fixing `B.038` on this branch.** It would give a clean `-race` run on the execution view,
  so a new race there could not hide behind the known one. D-11 and the plan kept the fix out
  on purpose, and the cost is not sized. Leaving it is a fair call, but the red run now has a
  price that it did not have before Commit 8.
- **`RerollDirective.Slot`.** A field the dispatcher needs for its own bookkeeping sits on a
  type that every collector reads. Keeping it apart would take a wrapper type at each card
  and test, which costs more than the wart does. It stays.

## 3 — Does it fit the project?

| Key Decision or pattern | Fit |
|---|---|
| Elective rerolls keep per-roll timing (dice) | Matches; the fence-sign stays at the roll |
| Rerolls reach stat rolls only (dice) | Matches; amount, table and target rolls show no card |
| A roll with no collector asks no one (dice) | Matches; the three contests wait for `F.037` |
| A new roll extends a route (dice) | Matches; the asks changed only in `rollWithHooks` |
| A budget lives on its owner (hooks) | Matches for the owner; the trigger left the key, on purpose (`B.024`), and the page says so |
| A prompt an action cannot unwind ignores Esc (overlays) | Matches; both cards ignore it (D-7) |
| `askEntries` is the one table of asks (overlays) | Matches; two renamed rows, and the test walks every kind |
| The adapter copies what views render (event stream) | Matches more strictly than most: card payloads hold only strings, and a test walks them |
| The message enums are append-only (event stream) | Matches; asks renamed in place, `EvtRollSettled` appended |
| Three archetypes, the rest `huh` forms (overlays) | Diverges: the cards use no form (Decision 7); the page records it |
| Collectors return errors | Diverges on the two card asks: a bad reply declines and reports, because `hooks.Collector` has no error (Decision 1) |

`A.001.7` D-7 expected a die picker on the card. D-1 narrowed that: one kept die, so nothing
to pick. That is the one earlier design note this work overrode, and it was Robert's call.

## 4 — What does it change about the project?

- **More stops.** Each hooked roll with a collector now takes at least one key press, so a
  contest takes two. A holder who keeps declining Psychic Academy is asked on every rival roll
  until the use is spent (discovery finding 2). The wider policy stays with `F.018`.
- **New seam.** `F.037` gets the card by passing a collector, `F.038` can reuse the wording
  in `roll_card.go`, and a planner under `A.003` can read the same card structs.
- **New rule for rulebook authors.** Two entries of one effect on one definition now share a
  use. `A.001.5` cannot give them separate budgets without a change to `budgetKey`.
- **New contract for hooks and tests.** An offer filter must report what it strikes. A mock
  test that rolls with a real registry uses `mocks.NewRollingCollector`.
- **New red run.** `go test -race` on the execution package fails until `B.038` is fixed. The
  plain gate does not run it.
- **Docs now out of date.** `overlays.md` still says "confirm a reroll" in its Purpose, and the
  `hooks.md` diagram labels the collector "elective select / confirm". Neither is wrong enough
  to block; `D.010` covers `hooks.md`.

## 5 — Rating

| Metric | Score | Evidence |
|---|---|---|
| Delivered | 5 | All of the discovery's In Plain Words and D-11's scope shipped; nothing cut |
| Plan held | 4 | One doc-text reversal (Record 15) and one ruling reversed (Record 17, by 19); no design reversed, but the manual check added three display fixes |
| Rework | | |
| Fit | 4 | Every binding Key Decision matches; two recorded divergences, both with a reason |
| Simplicity | 4 | One roll function and two small overlays; `Slot` is the one wart |
| **Total** | | |

**Verdict:** the card does what the discovery asked, in the shape it asked for, and the unseen part is Psychic Academy's row in the real TUI.

**What surprised me:** nothing draws an overlay's `Help()`, so the first hand-drawn card had no keys on screen, and every test passed anyway.

## Follow-ups

| ID | Item | Trigger | Detail |
|---|---|---|---|
| `D.011` | Sandbox step for Psychic Academy | -- | No sandbox seed holds `SWN-T-014` and `TESTING.md` has no step for it, so its row has not been seen in the TUI; add a seed and a step: a rival roll shows the row with its owner, a take spends the use, a decline keeps it. Found in the `F.039` close |
| `B.038` | Move the row out of triage | -- | Existing row; `F.039` Commit 8 confirmed the race, so its `needs triage` trigger is stale, and the fix also turns the execution package's `-race` run green. Found in the `F.039` close |
