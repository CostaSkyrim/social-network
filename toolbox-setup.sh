#!/bin/bash

# Comprehensive Social Network Toolbox Setup
# Installs all dependencies

set -e

CONTAINER_NAME="social-network-dev"

echo "🐋 Creating comprehensive toolbox container: $CONTAINER_NAME"

# Create the container
toolbox create $CONTAINER_NAME

# Install all dependencies inside the container
toolbox enter $CONTAINER_NAME << 'EOF'

echo "📦 Installing system packages..."

# Update system and install base tools (Go, SQLite, Node, etc.)
sudo dnf update -y
sudo dnf install -y golang sqlite sqlite-devel git make curl gcc glibc-devel

# Install Node
curl -fsSL https://rpm.nodesource.com/setup_20.x | sudo bash -
sudo dnf install -y nodejs
sudo npm install -g typescript react react-dom vite create-vite

echo "🔧 Setting up Go environment..."

# 1. Add Go bin directory to PATH permanently
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
export PATH=$PATH:$HOME/go/bin

# 2. Install the 'migrate' CLI tool
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 3. Install Air for live reloading
go install github.com/air-verse/air@latest

echo ""
echo "✅ Go toolchain is ready."
echo "   When you are inside the container and in your project directory,"
echo "   run 'go mod download' to install all your project dependencies."
echo ""
echo "✨ Setup complete."

EOF

echo ""
echo "🎉 Toolbox container '$CONTAINER_NAME' created!"
echo ""
echo "To enter your development environment:"
echo "  toolbox enter $CONTAINER_NAME"
echo ""
echo "All dependencies are installed and ready to use."