#!/bin/bash
# Social Network Development Container Setup

set -e

CONTAINER_NAME="social-network-dev"
BASE_IMAGE="fedora:40"
PROJECT_DIR="$(pwd)"

# Check prerequisites
if ! command -v docker &>/dev/null; then
    echo "❌ Docker not found!"
    echo ""
    echo "Install Docker on Arch Linux:"
    echo "  sudo pacman -S docker"
    echo "  sudo systemctl enable --now docker"
    echo "  sudo usermod -aG docker \$USER"
    echo "  newgrp docker  (or re-login)"
    exit 1
fi

if ! docker info &>/dev/null; then
    echo "❌ Docker daemon is not running or you don't have permission!"
    echo ""
    echo "Try:"
    echo "  sudo systemctl start docker"
    echo "  sudo usermod -aG docker \$USER && newgrp docker"
    exit 1
fi

# Remove existing container if present
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "⚠️  Removing existing container: $CONTAINER_NAME"
    docker rm -f "$CONTAINER_NAME"
fi

echo "🐋 Creating development container: $CONTAINER_NAME"
echo "   Base image: $BASE_IMAGE"
echo "   Project dir: $PROJECT_DIR"
echo ""

# Setup script to run inside container
SETUP_SCRIPT=$(cat << 'INNER_SCRIPT'
#!/bin/bash
set -e

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "   Setting up Social Network Development Environment"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Update system
echo "📦 Updating system..."
dnf update -y -q

# Install base packages
echo "📦 Installing base packages..."
dnf install -y \
    curl git make gcc glibc-devel \
    sqlite sqlite-devel \
    findutils procps-ng which wget tar vim \
    dnf-plugins-core > /dev/null 2>&1

# Install Go 1.24
echo "📦 Installing Go 1.24.2..."
cd /tmp
wget -q https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
rm go1.24.2.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin:/root/go/bin' >> /etc/profile.d/go.sh
echo 'export GOPATH=/root/go' >> /etc/profile.d/go.sh
export PATH=$PATH:/usr/local/go/bin:/root/go/bin
export GOPATH=/root/go
mkdir -p /root/go/bin

echo "   ✓ Go $(go version | awk '{print $3}') installed"

# Install Node.js 20
echo "📦 Installing Node.js 20..."
curl -fsSL https://rpm.nodesource.com/setup_20.x | bash - > /dev/null 2>&1
dnf install -y nodejs > /dev/null 2>&1
echo "   ✓ Node.js $(node --version) installed"

# Install global Node tools
echo "📦 Installing global Node tools..."
npm install -g typescript vite create-vite > /dev/null 2>&1
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
cat > /root/.air.toml << 'EOF'
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
cat >> /root/.bashrc << 'EOF'
source /etc/profile.d/go.sh

# Development aliases
alias serve="air"
alias gotest="go test -v ./..."
alias build="go build -o bin/social-network ./cmd/api"
alias migrate-up="migrate -database sqlite3://data.db -path ./migrations up"
alias migrate-down="migrate -database sqlite3://data.db -path ./migrations down"
alias lint="staticcheck ./..."

echo ""
echo "🐋 Social Network Development Container"
echo "   Go: $(go version | awk '{print $3}')"
echo "   Node: $(node --version)"
echo ""
echo "   Aliases: serve, gotest, build, migrate-up, migrate-down, lint"
echo ""
EOF

# Final verification
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Development Environment Ready!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Go:         $(go version)"
echo "Node:       $(node --version)"
echo "npm:        v$(npm --version)"
echo "TypeScript: $(tsc --version)"
echo "SQLite:     $(sqlite3 --version 2>&1 | head -n1)"
echo "Migrate:    $(migrate -version 2>&1 | head -n1)"
echo "Air:        $(air -version 2>&1 | head -n1)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
INNER_SCRIPT
)

echo "🔧 Running setup inside container (this will take 5-8 minutes)..."

# Run container with project dir mounted, execute setup, keep it running
docker run -it \
    --name "$CONTAINER_NAME" \
    --mount type=bind,source="$PROJECT_DIR",target=/workspace \
    --workdir /workspace \
    "$BASE_IMAGE" \
    bash -c "$SETUP_SCRIPT"

# Commit the configured container as an image so setup isn't lost
echo ""
echo "💾 Saving container state..."
docker commit "$CONTAINER_NAME" "${CONTAINER_NAME}-image" > /dev/null
docker rm "$CONTAINER_NAME"

# Create a helper script to enter the container
cat > enter-dev.sh << EOF
#!/bin/bash
# Enter the development container
docker run -it --rm \\
    --name "${CONTAINER_NAME}" \\
    --mount type=bind,source="\$(pwd)",target=/workspace \\
    --workdir /workspace \\
    -p 8080:8080 \\
    -p 5173:5173 \\
    ${CONTAINER_NAME}-image \\
    bash --login
EOF
chmod +x enter-dev.sh

echo ""
echo "🎉 Setup complete!"
echo ""
echo "To enter your development environment:"
echo "  ./enter-dev.sh"
echo ""
echo "Inside the container:"
echo "  - Your project files are at /workspace"
echo "  - Run 'go mod download' to install dependencies"
echo "  - Run 'air' (or 'serve') to start with hot reload"
echo "  - Ports 8080 and 5173 are forwarded to your host"
echo ""
