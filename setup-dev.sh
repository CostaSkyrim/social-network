#!/bin/bash
# Social Network Development Container Setup
# Creates and configures a container with Go 1.24+, Node.js 20+, and all dev tools

set -e

# Refuse to run as root — distrobox handles privilege escalation on its own
if [ "$EUID" -eq 0 ]; then
    echo "❌ Do not run this script with sudo!"
    echo "   Run it as your normal user: ./setup-dev.sh"
    echo "   If rootful mode is needed, the script will prompt for elevation."
    exit 1
fi

CONTAINER_NAME="social-network-dev"
BASE_IMAGE="${1:-docker.io/library/fedora:40}"

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

# On Arch, rootless podman may need cgroup v2 delegation
if command -v pacman &>/dev/null; then
    echo "🔧 Ensuring rootless podman is set up on Arch..."
    loginctl enable-linger "$(whoami)" 2>/dev/null || true
    podman system migrate 2>/dev/null || true
fi

# Test if rootless actually works (create + start a container)
ROOT_FLAGS=""
TEST_NAME="distrobox-test-$$"
echo "🧪 Testing rootless container support..."
if distrobox create --name "$TEST_NAME" --image "$BASE_IMAGE" --yes 2>/dev/null && \
   timeout 10 distrobox enter "$TEST_NAME" -- echo "ok" 2>/dev/null | grep -q "ok"; then
    echo "   ✓ Rootless available"
    distrobox rm --force "$TEST_NAME" 2>/dev/null || true
else
    echo "   ⚠️  Rootless not available — will use --root mode"
    ROOT_FLAGS="--root"
    distrobox rm --force "$TEST_NAME" 2>/dev/null || true
fi

# Remove existing container if present
if [ -z "$ROOT_FLAGS" ]; then
    distrobox list 2>/dev/null | grep -q "$CONTAINER_NAME" && distrobox rm --force "$CONTAINER_NAME"
else
    distrobox list --root 2>/dev/null | grep -q "$CONTAINER_NAME" && distrobox rm --root --force "$CONTAINER_NAME"
fi

echo "🐋 Creating development container: $CONTAINER_NAME"
echo "   Base image: $BASE_IMAGE"
echo ""

# Create the container
distrobox create $ROOT_FLAGS \
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

SUDO=""
if [ "$(id -u)" -ne 0 ]; then
    SUDO="sudo"
fi

# Detect package manager
if command -v pacman &>/dev/null; then
    PM="pacman"
    PM_INSTALL="$SUDO pacman -S --noconfirm"
    PM_UPDATE="$SUDO pacman -Syu --noconfirm"
elif command -v dnf &>/dev/null; then
    PM="dnf"
    PM_INSTALL="$SUDO dnf install -y"
    PM_UPDATE="$SUDO dnf update -y -q"
else
    echo "❌ No supported package manager found (pacman or dnf)"
    exit 1
fi
echo "   Detected package manager: $PM"

# Update system
echo "📦 Updating system..."
$PM_UPDATE

# Install base packages
echo "📦 Installing base packages..."
if [ "$PM" = "pacman" ]; then
    $PM_INSTALL \
        curl git make gcc \
        sqlite sqlite3 \
        findutils procps-ng which wget tar vim \
        redis
elif [ "$PM" = "dnf" ]; then
    $PM_INSTALL \
        curl git make gcc glibc-devel \
        sqlite sqlite-devel \
        findutils procps-ng which wget tar vim \
        dnf-plugins-core \
        redis
fi

# Install Go 1.24
echo "📦 Installing Go 1.24.2..."
cd /tmp
wget -q https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
$SUDO rm -rf /usr/local/go
$SUDO tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
rm go1.24.2.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' | $SUDO tee /etc/profile.d/go.sh > /dev/null
echo 'export GOPATH=$HOME/go' | $SUDO tee -a /etc/profile.d/go.sh > /dev/null
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export GOPATH=$HOME/go

# Setup Go environment for user
mkdir -p "$HOME/go/bin"
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc

echo "   ✓ Go $(go version | awk '{print $3}') installed"

# Install Node.js 22 (LTS)
echo "📦 Installing Node.js 22..."
if [ "$PM" = "pacman" ]; then
    $PM_INSTALL nodejs npm
elif [ "$PM" = "dnf" ]; then
    curl -fsSL https://rpm.nodesource.com/setup_22.x | $SUDO bash -
    $SUDO dnf install -y nodejs
fi
echo "   ✓ Node.js $(node --version) installed"

# Install global Node tools
echo "📦 Installing global Node tools..."
$SUDO npm install -g typescript vite create-vite
echo "   ✓ TypeScript, Vite, create-vite installed"

# Install Go development tools
echo "📦 Installing Go development tools..."
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/air-verse/air@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
echo "   ✓ migrate, air, delve, goimports, staticcheck installed"

# Create basic air configuration
cat > ~/.air.toml << 'EOF'
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

# Install frontend dependencies
echo "📦 Installing frontend npm dependencies..."
if [ -d "$(find / -maxdepth 4 -name 'frontend' -type d 2>/dev/null | head -1)" ]; then
    cd "$(find / -maxdepth 4 -name 'frontend' -type d 2>/dev/null | head -1)"
    npm install
    echo "   ✓ Frontend dependencies installed"
fi

# Add helpful aliases
cat >> ~/.bashrc << 'EOF'

# Development aliases
alias serve="air"
alias gotest="go test -v ./..."
alias build="go build -o bin/social-network ./cmd/api"
alias lint="staticcheck ./..."

# Welcome message
echo ""
echo "🐋 Social Network Development Container"
echo "   Go: $(go version | awk '{print $3}')"
echo "   Node: $(node --version)"
echo "   Make: $(make --version 2>&1 | head -1)"
echo ""
echo "   Commands: make dev, make check, make frontend-dev, make backend-run"
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
echo "Make:      $(make --version 2>&1 | head -1)"
echo "TypeScript: $(tsc --version)"
echo "SQLite:    $(sqlite3 --version 2>&1 | head -n1)"
echo "Redis:     $(redis-cli --version 2>&1 | head -n1 || echo 'not found')"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

INNER_SCRIPT

chmod +x "$SETUP_SCRIPT"

echo "🔧 Running setup inside container (this will take 5-8 minutes)..."
distrobox enter $ROOT_FLAGS "$CONTAINER_NAME" -- bash "$SETUP_SCRIPT"

echo ""
echo "🎉 Setup complete!"
echo ""
echo "To enter your development environment:"
echo "  distrobox enter $ROOT_FLAGS $CONTAINER_NAME"
echo ""
echo "Inside the container, from the project directory:"
echo "  make dev        # Start backend + frontend"
echo "  make check      # Run all checks"
echo "  make frontend-dev  # Start frontend only"
echo "  make backend-run   # Start backend only"
echo ""
