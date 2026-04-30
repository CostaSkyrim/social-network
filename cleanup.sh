#!/bin/bash
# Social Network Development Container Cleanup
# Stops and removes the development container

set -e

CONTAINER_NAME="social-network-dev"

echo "🐋 Cleaning up development container: $CONTAINER_NAME"

# Check if distrobox is available
if ! command -v distrobox &>/dev/null; then
    echo "❌ distrobox not found"
    exit 1
fi

# Check if container exists
if distrobox list 2>/dev/null | grep -q "$CONTAINER_NAME"; then
    echo "⚠️  Found container: $CONTAINER_NAME"
    
    # Stop if running
    echo "   Stopping container..."
    distrobox stop "$CONTAINER_NAME" 2>/dev/null || true
    
    # Remove container
    echo "   Removing container..."
    distrobox rm --force "$CONTAINER_NAME"
    
    echo "✅ Container removed successfully!"
else
    echo "ℹ️  Container '$CONTAINER_NAME' not found"
fi

echo ""
echo "🧹 Cleanup complete!"

# Optional: Show remaining containers
REMAINING=$(distrobox list 2>/dev/null | tail -n +2 | head -n -1 || true)
if [ -n "$REMAINING" ]; then
    echo ""
    echo "Remaining containers:"
    distrobox list
else
    echo "No remaining containers."
fi