# GM Toolkit

A modular CLI toolkit for tabletop RPG game masters, built in Go. Currently focused on a **Faction Tracker** — a tool for managing faction turns, assets, and campaign history inspired by the Stars Without Number faction rules.

## Philosophy
- CLI-first: core logic is decoupled from any UI layer
- Each tool is self-contained in its own package
- Built to grow: a TUI or web frontend can be added later without touching core logic

## Faction Tracker

Track factions across a campaign — stats, assets, Coin, goals, and turn history. Mechanics are inspired by SWN faction rules, adapted into the GM's own system.

### Current Features
- `faction create` — interactive wizard to create a new faction (name, homeworld, scale, stats, starting assets, tags, goal)
- `faction list` — summary view of all factions in a campaign (name, scale, HP, current goal)
- `faction delete` — remove a faction with confirmation

### In Progress
- Turn engine (income, maintenance, action resolution)
- History/event log (append-only JSONL)
- Narrative summary renderer
- Review Mode and Edit Mode

## Project Structure
```
gm-toolkit/
├── cmd/
│   └── faction-manager/          # Faction tracker CLI entrypoint
│       ├── main.go
│       ├── bin/                  # Compiled binary
│       ├── campaigns/            # Campaign state files (gitignored)
│       └── commands/             # Cobra command tree
│           ├── app.go            # App struct, engine init, command wiring
│           ├── review.go         # Review Mode (stub)
│           ├── turn.go           # Turn Mode (stub)
│           └── faction/          # Faction subcommands (create, list, delete)
│               └── wizard/       # Interactive wizard steps
├── internal/
│   └── faction/
│       ├── domain/               # Pure data types
│       ├── loader/               # Rulebook loader (static TOML data)
│       ├── state/                # Campaign state load/save
│       ├── engine/               # Game logic (growing)
│       └── data/                 # Static asset, tag, and goal TOML files
└── docs/
    ├── swn-faction-mechanics.md  # Rules reference
    ├── dev-journal-factions.md   # Living design doc and decisions log
    ├── discovery/                # Feature discovery and planning docs
    └── tracking/                 # Development tracking journals
```

## Development

This project uses a VSCode Dev Container. To get started:
1. Install Docker Desktop and the VSCode Dev Containers extension
2. Open the project in VSCode
3. When prompted, reopen in container

See `.claude/DEV_SETUP.md` for build and run instructions.
