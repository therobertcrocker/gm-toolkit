#!/bin/bash
set -e

# Install zsh
sudo apt-get update && sudo apt-get install -y zsh

# Install Starship
curl -sS https://starship.rs/install.sh | sudo sh -s -- --yes

# Configure Starship
mkdir -p ~/.config

# Apply pastel powerline preset
starship preset pastel-powerline -o ~/.config/starship.toml

# Add container module to starship config
sed -i 's/\$username\\/\$username\\\n\$container\\/' ~/.config/starship.toml

cat >> ~/.config/starship.toml << 'EOF'

[container]
disabled = false
style = "bg:#9A348E"
format = '[$symbol \[$name\] ]($style)'
EOF

# Set Starship as shell init
echo 'eval "$(starship init zsh)"' >> ~/.zshrc

# Set zsh as default shell
sudo chsh -s $(which zsh) vscode

# Install Go tools
go install github.com/air-verse/air@latest

# Install Node.js and Claude Code
curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
sudo apt-get install -y nodejs
sudo npm install -g @anthropic-ai/claude-code