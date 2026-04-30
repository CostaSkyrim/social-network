#!/bin/bash
# Social Network Development Container Setup
# Creates and configures a container with Go 1.24+, Node.js 20+, and all dev tools

set -e

CONTAINER_NAME="social-network-dev"
BASE_IMAGE="docker.io/library/fedora:40"

# Check prerequisites
if ! command -v distrobox &>/dev/null; then
    echo "❌ distrobox not found!"
    echo ""
    echo "Install distrobox on your system:"
    echo "  Fedora/Bluefin: sudo dnf install distrobox"
    echo "  Arch Linux:     sudo pacman -S distrobox"
    echo "  Universal:      curl -s https://raw.githubusercontent.com/89luca89/distrobox/main/install | sudo sh"
    echo ""
    echo "Also ensure Podman or Docker is installed"
    exit 1
fi

if ! command -v podman &>/dev/null && ! command -v docker &>/dev/null; then
    echo "❌ Neither Podman nor Docker found!"
    exit 1
fi

# Remove existing container if present
if distrobox list 2>/dev/null | grep -q "$CONTAINER_NAME"; then
    echo "⚠️  Removing existing container: $CONTAINER_NAME"
    distrobox rm --force "$CONTAINER_NAME"
fi

echo "🐋 Creating development container: $CONTAINER_NAME"
echo "   Base image: $BASE_IMAGE"
echo ""

# Create the container
distrobox create \
    --name "$CONTAINER_NAME" \
    --image "$BASE_IMAGE" \
    --yes

# Setup script to run inside container
SETUP_SCRIPT=$(mktemp /tmp/setup-XXXXXX.sh)
trap "rm -f $SETUP_SCRIPT" EXIT

cat > "$SETUP_SCRIPT" << 'INNER_SCRIPT'
#!/bin/bash
set -e

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "   Setting up Social Network Development Environment"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Update system
echo "📦 Updating system..."
sudo dnf update -y -q

# Install base packages
echo "📦 Installing base packages..."
sudo dnf install -y \
    curl git make gcc glibc-devel \
    sqlite sqlite-devel \
    findutils procps-ng which wget tar vim \
    dnf-plugins-core > /dev/null 2>&1

# Install Go 1.24
echo "📦 Installing Go 1.24.2..."
cd /tmp
wget -q https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
rm go1.24.2.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' | sudo tee /etc/profile.d/go.sh > /dev/null
echo 'export GOPATH=$HOME/go' | sudo tee -a /etc/profile.d/go.sh > /dev/null
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export GOPATH=$HOME/go

# Setup Go environment for user
mkdir -p "$HOME/go/bin"
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc

echo "   ✓ Go $(go version | awk '{print $3}') installed"

# Install Node.js 20
echo "📦 Installing Node.js 20..."
curl -fsSL https://rpm.nodesource.com/setup_20.x | sudo bash - > /dev/null 2>&1
sudo dnf install -y nodejs > /dev/null 2>&1
echo "   ✓ Node.js $(node --version) installed"

# Install global Node tools
echo "📦 Installing global Node tools..."
sudo npm install -g typescript vite create-vite > /dev/null 2>&1
echo "   ✓ TypeScript, Vite, create-vite installed"

# Install Go development tools
echo "📦 Installing Go development tools..."
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest > /dev/null 2>&1
go install github.com/air-verse/air@latest > /dev/null 2>&1
go install github.com/go-delve/delve/cmd/dlv@latest > /dev/null 2>&1
go install golang.org/x/tools/cmd/goimports@latest > /dev/null 2>&1
go install honnef.co/go/tools/cmd/staticcheck@latest > /dev/null 2>&1
echo "   ✓ migrate, air, delve, goimports, staticcheck installed"

# Create basic air configuration
cat > ~/.air.toml << 'EOF'
# Air configuration for hot reloading
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ."
  bin = "./tmp/main"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "node_modules"]
  include_ext = ["go", "tpl", "tmpl", "html"]
  stop_on_error = true

[log]
  time = false

[color]
  build = "yellow"
  runner = "green"
  watcher = "cyan"
EOF

# Add helpful aliases
cat >> ~/.bashrc << 'EOF'

# Development aliases
alias serve="air"
alias gotest="go test -v ./..."
alias build="go build -o bin/social-network ./cmd/api"
alias migrate-up="migrate -database sqlite3://data.db -path ./migrations up"
alias migrate-down="migrate -database sqlite3://data.db -path ./migrations down"
alias lint="staticcheck ./..."

# Welcome message
echo ""
echo "🐋 Social Network Development Container"
echo "   Go: $(go version | awk '{print $3}')"
echo "   Node: $(node --version)"
echo ""
echo "   Aliases: serve, test, build, migrate-up, migrate-down, lint"
echo ""
EOF

# Final verification
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Development Environment Ready!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Go:        $(go version)"
echo "Node:      $(node --version)"
echo "npm:       v$(npm --version)"
echo "TypeScript: $(tsc --version)"
echo "SQLite:    $(sqlite3 --version 2>&1 | head -n1)"
echo "Migrate:   $(migrate -version 2>&1 | head -n1)"
echo "Air:       $(air -version 2>&1 | head -n1 | cut -d' ' -f1-3)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

INNER_SCRIPT

chmod +x "$SETUP_SCRIPT"

echo "🔧 Running setup inside container (this will take 5-8 minutes)..."
distrobox enter "$CONTAINER_NAME" -- bash "$SETUP_SCRIPT"

echo ""
echo "🎉 Setup complete!"
echo ""
echo "To enter your development environment:"
echo "  distrobox enter $CONTAINER_NAME"
echo ""
echo "Inside the container:"
echo "  - Navigate to your project (your home directory is mounted)"
echo "  - Run 'go mod download' to install dependencies"
echo "  - Run 'air' to start development server with hot reload"
echo ""
