#!/bin/sh
# ForgeCrew installer for macOS / Linux
# Downloads the latest forgecrew binary from GitHub Releases.
# Falls back to local build instructions if no release is found.
set -e

REPO="1424772/ForgeCrew"
BINARY="forgecrew"
VERSION="${FORGECREW_VERSION:-latest}"

# Default install directory — user-local, no sudo required.
INSTALL_DIR="${FORGECREW_INSTALL_DIR:-$HOME/.forgecrew/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

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

# Build download URL (GitHub Releases placeholder until releases are tagged).
if [ "$VERSION" = "latest" ]; then
    RELEASE_URL="https://github.com/${REPO}/releases/latest/download/forgecrew_${OS}_${ARCH}"
else
    RELEASE_URL="https://github.com/${REPO}/releases/download/${VERSION}/forgecrew_${OS}_${ARCH}"
fi

info "Installing ForgeCrew ${VERSION} for ${OS}/${ARCH}..."
info "Install directory: ${INSTALL_DIR}"

# Create install directory.
mkdir -p "${INSTALL_DIR}"

# Try to download from GitHub Releases.
# Downloads to a temp file first, validates, then moves to final path.
TMPFILE="${INSTALL_DIR}/${BINARY}.tmp.$$"
DOWNLOAD_FAILED=0

# Clean up any leftover temp file from a previous failed run.
rm -f "${TMPFILE}"

if command -v curl > /dev/null 2>&1; then
    info "Downloading with curl..."
    # Use if/else to capture exit code without set -e triggering early exit
    # on command-substitution failure.
    if HTTP_CODE=$(curl -sL -o "${TMPFILE}" -w "%{http_code}" "${RELEASE_URL}"); then
        CURL_EXIT=0
    else
        CURL_EXIT=$?
    fi

    if [ "$CURL_EXIT" != "0" ]; then
        err "curl exited with code ${CURL_EXIT} — download failed."
        rm -f "${TMPFILE}"
        DOWNLOAD_FAILED=1
    elif [ -z "$HTTP_CODE" ]; then
        err "curl returned no HTTP status code."
        rm -f "${TMPFILE}"
        DOWNLOAD_FAILED=1
    elif ! echo "$HTTP_CODE" | grep -qE '^2[0-9][0-9]$'; then
        err "Server returned HTTP ${HTTP_CODE} (expected 2xx)."
        rm -f "${TMPFILE}"
        DOWNLOAD_FAILED=1
    elif [ ! -s "${TMPFILE}" ]; then
        err "Downloaded file is empty (HTTP ${HTTP_CODE})."
        rm -f "${TMPFILE}"
        DOWNLOAD_FAILED=1
    fi

elif command -v wget > /dev/null 2>&1; then
    info "Downloading with wget..."
    # Use if/else to capture exit code without set -e triggering early exit.
    if WGET_OUTPUT=$(wget -q -O "${TMPFILE}" "${RELEASE_URL}" 2>&1); then
        WGET_EXIT=0
    else
        WGET_EXIT=$?
    fi

    if [ "$WGET_EXIT" != "0" ]; then
        err "wget exited with code ${WGET_EXIT} — download failed."
        if [ -n "$WGET_OUTPUT" ]; then
            err "wget: ${WGET_OUTPUT}"
        fi
        rm -f "${TMPFILE}"
        DOWNLOAD_FAILED=1
    elif [ ! -s "${TMPFILE}" ]; then
        err "Downloaded file is empty."
        rm -f "${TMPFILE}"
        DOWNLOAD_FAILED=1
    fi

else
    err "Neither curl nor wget is available. Please install one of them and retry."
    DOWNLOAD_FAILED=1
fi

if [ "$DOWNLOAD_FAILED" = "1" ]; then
    warn ""
    warn "ForgeCrew binaries are not yet published to GitHub Releases."
    warn "You can build from source:"
    warn ""
    warn "  git clone https://github.com/${REPO}.git"
    warn "  cd ForgeCrew"
    warn "  go build -o ${BINARY} ./cmd/forgecrew"
    warn "  mv ${BINARY} ${INSTALL_DIR}/"
    warn ""
    if command -v go > /dev/null 2>&1; then
        info "Go detected. You can also run:"
        info "  go install github.com/${REPO}/cmd/forgecrew@latest"
    fi
    exit 1
fi

# Download validated — move temp file to final path.
mv "${TMPFILE}" "${INSTALL_DIR}/${BINARY}"

# Make executable.
chmod +x "${INSTALL_DIR}/${BINARY}"

info "ForgeCrew installed to ${INSTALL_DIR}/${BINARY}"

# Check if install dir is in PATH.
case ":$PATH:" in
    *:"${INSTALL_DIR}":*)
        # Already in PATH — nothing to do.
        ;;
    *)
        warn ""
        warn "NOTE: ${INSTALL_DIR} is not in your PATH."
        warn "Add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        warn ""
        warn "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        warn ""
        warn "Then restart your terminal or run:"
        warn "  source ~/.bashrc   # or ~/.zshrc"
        ;;
esac

info "Run 'forgecrew init' in your project to get started."
