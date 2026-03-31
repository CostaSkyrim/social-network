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

# Update system
sudo dnf update -y

# Install Go and backend dependencies
sudo dnf install -y \
    golang \
    sqlite \
    sqlite-devel \
    git \
    make \
    curl \
    gcc \
    glibc-devel

# Install Node.js 20 LTS (includes npm)
curl -fsSL https://rpm.nodesource.com/setup_20.x | sudo bash -
sudo dnf install -y nodejs

# Install TypeScript and React dependencies globally
sudo npm install -g typescript
sudo npm install -g react
sudo npm install -g react-dom
sudo npm install -g vite
sudo npm install -g create-vite

# Install Go packages for backend
go install github.com/gorilla/websocket@latest
go install github.com/mattn/go-sqlite3@latest
go install github.com/google/uuid@latest
go install golang.org/x/crypto/bcrypt@latest
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/air-verse/air@latest

# Add Go bin to PATH for this session
export PATH=$PATH:~/go/bin

# Verify installations
echo ""
echo "✅ Verification:"
echo "Go version: $(go version)"
echo "Node version: $(node --version)"
echo "npm version: $(npm --version)"
echo "TypeScript version: $(tsc --version)"
echo "SQLite version: $(sqlite3 --version)"
echo "migrate version: $(migrate -version 2>&1 | head -n1)"
echo "air version: $(air -version 2>&1 | head -n1)"

# Add Go bin to PATH permanently
echo 'export PATH=$PATH:~/go/bin' >> ~/.bashrc

echo ""
echo "✨ All dependencies installed successfully!"
echo "   Your toolbox is ready."

EOF

echo ""
echo "🎉 Toolbox container '$CONTAINER_NAME' created!"
echo ""
echo "To enter your development environment:"
echo "  toolbox enter $CONTAINER_NAME"
echo ""
echo "All dependencies are installed and ready to use."