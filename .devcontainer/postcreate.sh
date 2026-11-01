#!/bin/bash
set -e

# Shell setup (zsh, starship) comes from the dotfiles repo, not this script.

# Install Go tools
go install github.com/air-verse/air@latest

# Install Claude Code (native installer: it owns ~/.local, so it can update itself)
curl -fsSL https://claude.ai/install.sh | bash
