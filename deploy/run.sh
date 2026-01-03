#!/usr/bin/env bash
set -euo pipefail

# This script builds the Go binary, then builds and launches the Docker container
# Usage: ./run.sh

# Determine script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR}/.."

# Build the Go binary into the deploy folder
echo "🛠  Building Go binary..."
cd "${PROJECT_ROOT}"
go build -o deploy/mcp-server main.go

# Build the Docker image using the deploy/docker-compose.yml
echo "🐳  Building Docker image..."
cd "${SCRIPT_DIR}"
docker-compose -f docker-compose.yml build

# Start up the container in detached mode
echo "🚀  Starting container..."
docker-compose -f docker-compose.yml up -d

echo "✅  Deployment complete. Container 'mcp-server' is running and listening on port 3333."
