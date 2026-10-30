# GM Toolkit — Claude Instructions

@~/.claude/plugins/marketplaces/gestalt/plugins/gestalt/rules/working-together.md
@~/.claude/plugins/marketplaces/gestalt/plugins/gestalt/rules/git-and-delivery.md
@~/.claude/plugins/marketplaces/gestalt/plugins/gestalt/rules/code-and-docs.md

## This project

- **Override (working-together, Files and tools):** search the codebase with the LSP tools, not grep. Grep stays fine for docs.

## Extra docs

Before any feature discussion or implementation, read the doc(s) relevant to the work at hand:

| If you're working on…                | Read first                                  |
|--------------------------------------|---------------------------------------------|
| Overall design or open questions     | `docs/dev_journals/faction-manager/dev-journal-factions.md`          |
| Feature Request or Deferred Decision | `docs/dev_journals/faction-manager/planned-work.md`                      |
| Setup, workflow, or dev conventions  | `docs/contributing/overview.md` (front door; routes to engine/interface guides) |
| Historical decision archaeology      | `docs/dev_journals/faction-manager/archive/decisions-log.md` — frozen archive, never appended |
| Rules behavior                       | `docs/rules/swn-faction-mechanics.md`       |

**Searching the decisions log (historical only — log is frozen):** Do not read the full file. Use one of:
- **Index first:** Read the index at the top (~25 lines), find the relevant section by topic, then read only that section using `offset`/`limit`.
- **Keyword grep:** `grep -n "keyword" docs/dev_journals/faction-manager/archive/decisions-log.md` to find the line, then read the surrounding section.
