#!/bin/sh
# ForgeCrew installer for macOS / Linux
# Downloads the latest forgecrew binary from GitHub Releases.
# Falls back to go build if no release is found.
set -e

REPO="1424772/ForgeCrew"
BINARY="forgecrew"
VERSION="${FORGECREW_VERSION:-latest}"
INSTALL_DIR="${FORGECREW_INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info()  { printf "${GREEN}[forgecrew]${NC} %s\n" "$*"; }
warn()  { printf "${YELLOW}[forgecrew]${NC} %s\n" "$*"; }
err()   { printf "${RED}[forgecrew]${NC} %s\n" "$*"; }

# Detect OS and architecture.
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) err "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux|darwin) ;;
    *) err "Unsupported OS: $OS (this script supports Linux and macOS)"; exit 1 ;;
esac

# Build download URL.
if [ "$VERSION" = "latest" ]; then
    RELEASE_URL="https://github.com/${REPO}/releases/latest/download/${BINARY}-${OS}-${ARCH}"
else
    RELEASE_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}-${OS}-${ARCH}"
fi

info "Installing ForgeCrew ${VERSION} for ${OS}/${ARCH}..."

# Try to download from GitHub Releases.
if command -v curl > /dev/null 2>&1; then
    HTTP_CODE=$(curl -sL -o /tmp/${BINARY} -w "%{http_code}" "${RELEASE_URL}")
    DOWNLOAD_OK=$?
elif command -v wget > /dev/null 2>&1; then
    wget -q -O /tmp/${BINARY} "${RELEASE_URL}" 2>&1
    DOWNLOAD_OK=$?
else
    warn "Neither curl nor wget found. Falling back to go build."
    DOWNLOAD_OK=1
fi

if [ "$DOWNLOAD_OK" != "0" ] || [ "$HTTP_CODE" = "404" ]; then
    warn "Release not found (${RELEASE_URL})."
    warn ""

    if command -v go > /dev/null 2>&1; then
        info "Building from source with 'go install'..."
        go install "github.com/${REPO}/cmd/forgecrew@latest"
        info "Installed via go install."
        exit 0
    fi

    warn "Option 1: Install Go (https://go.dev/dl/) and run:"
    warn "  go install github.com/${REPO}/cmd/forgecrew@latest"
    warn ""
    warn "Option 2: Clone and build manually:"
    warn "  git clone https://github.com/${REPO}.git"
    warn "  cd ForgeCrew"
    warn "  go build -o ${BINARY} ./cmd/forgecrew"
    warn "  mv ${BINARY} ${INSTALL_DIR}/"
    exit 1
fi

# Install the binary.
chmod +x /tmp/${BINARY}
if [ -w "${INSTALL_DIR}" ]; then
    mv /tmp/${BINARY} "${INSTALL_DIR}/${BINARY}"
else
    warn "Need sudo to install to ${INSTALL_DIR}."
    sudo mv /tmp/${BINARY} "${INSTALL_DIR}/${BINARY}"
fi

info "ForgeCrew installed successfully!"
info "Run '${BINARY} init' in your project to get started."
