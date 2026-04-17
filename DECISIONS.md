# Architecture Decisions

## Language: Go
- Familiar to the primary developer
- Compiles to a single binary
- Strong standard library

## Environment: VSCode Dev Container
- Reproducible environment across rebuilds
- Go 1.26, Zsh, Oh My Zsh, Starship (pastel powerline preset)
- Air for live reload

## CLI-first
- Start with a CLI/TUI to focus on core logic
- Keeps UI and business logic cleanly separated
- Frontend (Wails, web, etc.) can be added later

## Persistence: TBD
- Not decided yet — will be driven by the needs of the faction tracker
- Candidates: SQLite, flat files (JSON/TOML), embedded key-value store

## TUI Framework: Bubble Tea (planned)
- Natural component model maps well to individual tools
- Stays within the Go ecosystem