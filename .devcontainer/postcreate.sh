#!/bin/bash
set -e

# Shell setup (zsh, starship) comes from the dotfiles repo, not this script.

# Install Go tools
go install github.com/air-verse/air@latest

# Install Node.js and Claude Code
curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
sudo apt-get install -y nodejs
sudo npm install -g @anthropic-ai/claude-code
