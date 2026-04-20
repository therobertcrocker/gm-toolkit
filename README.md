# GM Toolkit

A modular collection of tools for tabletop RPG game masters, built in Go.

## Philosophy
- CLI-first: core logic is decoupled from any UI layer
- Each tool is self-contained in its own package
- Built to grow: a frontend (TUI, web, or desktop) can be added later without touching core logic

## Tools
- [ ] Faction Tracker — track factions, their relationships, and agendas

## Project Structure
```
gm-toolkit/
├── cmd/              # CLI entry points
│   └── faction/      # Faction tracker CLI
├── internal/         # Core logic (not importable externally)
│   └── faction/      # Faction tracker models and logic
└── main.go
```

## Development
This project uses a VSCode Dev Container. To get started:
1. Install Docker Desktop and the VSCode Dev Containers extension
2. Open the project in VSCode
3. When prompted, reopen in container
