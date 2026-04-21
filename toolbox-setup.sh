
#!/bin/bash
# Comprehensive Social Network Toolbox Setup (Arch Linux)
# Uses distrobox with an explicit Arch image to avoid Fedora default issues
set -e
 
CONTAINER_NAME="social-network-dev"
ARCH_IMAGE="docker.io/archlinux/archlinux:latest"
 
# ── Prerequisites check ────────────────────────────────────────────────────────
for cmd in distrobox podman; do
    if ! command -v $cmd &>/dev/null; then
        echo "❌ '$cmd' not found. Install with:"
        echo "   sudo pacman -S distrobox podman"
        exit 1
    fi
done
 
# ── Tear down any broken previous attempt ─────────────────────────────────────
if distrobox list 2>/dev/null | grep -q "$CONTAINER_NAME"; then
    echo "⚠️  Container '$CONTAINER_NAME' already exists — removing it first..."
    distrobox rm --force $CONTAINER_NAME
fi
 
echo "🐋 Creating Arch container: $CONTAINER_NAME"
distrobox create \
    --name  $CONTAINER_NAME \
    --image $ARCH_IMAGE \
    --yes
 
# ── Write the setup script to a temp file, then run it inside ─────────────────
# (heredoc piped directly into distrobox enter can stall on some setups)
SETUP_SCRIPT=$(mktemp /tmp/setup-XXXXXX.sh)
trap "rm -f $SETUP_SCRIPT" EXIT
 
cat > "$SETUP_SCRIPT" << 'INNER'
set -e
 
echo "📦 Updating system..."
sudo pacman -Syu --noconfirm
 
echo "📦 Installing system packages..."
sudo pacman -S --noconfirm \
    base-devel \
    go \
    sqlite \
    git \
    make \
    curl \
    nodejs \
    npm
 
echo "📦 Installing global Node tools..."
sudo npm install -g typescript vite create-vite
 
echo "📦 Installing Go tools..."
export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"
 
go install github.com/gorilla/websocket@latest
go install github.com/mattn/go-sqlite3@latest
go install github.com/google/uuid@latest
go install golang.org/x/crypto/bcrypt@latest
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/air-verse/air@latest
 
# Persist Go bin path
grep -qxF 'export PATH=$PATH:$HOME/go/bin' ~/.bashrc \
    || echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
 
echo ""
echo "✅ Verification:"
go version
node --version  && echo "Node: OK"
npm --version   && echo "npm:  OK"
tsc --version   && echo "tsc:  OK"
sqlite3 --version
migrate -version 2>&1 | head -n1
air -version    2>&1 | head -n1
 
echo ""
echo "✨ All dependencies installed successfully!"
INNER
 
chmod +x "$SETUP_SCRIPT"
 
echo "🔧 Running setup inside container..."
distrobox enter $CONTAINER_NAME -- bash "$SETUP_SCRIPT"
 
echo ""
echo "🎉 Container '$CONTAINER_NAME' is ready!"
echo ""
echo "Enter your dev environment any time with:"
echo "  distrobox enter $CONTAINER_NAME"
