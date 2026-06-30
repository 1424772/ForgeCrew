#!/bin/sh
# ForgeCrew installer for macOS / Linux
# Downloads the latest forgecrew binary from GitHub Releases,
# verifies SHA256 checksum, and provides PATH setup instructions.
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

# ── Detect OS and architecture ──
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

# ── Build download URLs ──
if [ "$VERSION" = "latest" ]; then
    RELEASE_BASE="https://github.com/${REPO}/releases/latest/download"
else
    RELEASE_BASE="https://github.com/${REPO}/releases/download/${VERSION}"
fi

BINARY_NAME="forgecrew_${OS}_${ARCH}"
BINARY_URL="${RELEASE_BASE}/${BINARY_NAME}"
CHECKSUM_URL="${RELEASE_BASE}/checksums.txt"

info "Installing ForgeCrew ${VERSION} for ${OS}/${ARCH}..."
info "Install directory: ${INSTALL_DIR}"

# ── Create install directory ──
mkdir -p "${INSTALL_DIR}"

# ── Download binary ──
TMPFILE="${INSTALL_DIR}/${BINARY}.tmp.$$"
TMP_CHECKSUM="${INSTALL_DIR}/checksums.tmp.$$"
DOWNLOAD_FAILED=0

# Clean up any leftover temp files from a previous failed run.
rm -f "${TMPFILE}" "${TMP_CHECKSUM}"

if command -v curl > /dev/null 2>&1; then
    info "Downloading binary with curl..."
    # Use if/else to capture exit code without set -e triggering early exit
    # on command-substitution failure.
    if HTTP_CODE=$(curl -sL -o "${TMPFILE}" -w "%{http_code}" "${BINARY_URL}"); then
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
    info "Downloading binary with wget..."
    # Use if/else to capture exit code without set -e triggering early exit.
    if WGET_OUTPUT=$(wget -q -O "${TMPFILE}" "${BINARY_URL}" 2>&1); then
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

# ── SHA256 checksum verification ──
CHECKSUM_OK=0

# Download checksums.txt
if command -v curl > /dev/null 2>&1; then
    if HTTP_CODE=$(curl -sL -o "${TMP_CHECKSUM}" -w "%{http_code}" "${CHECKSUM_URL}"); then
        if echo "$HTTP_CODE" | grep -qE '^2[0-9][0-9]$' && [ -s "${TMP_CHECKSUM}" ]; then
            CHECKSUM_OK=1
        fi
    fi
elif command -v wget > /dev/null 2>&1; then
    if wget -q -O "${TMP_CHECKSUM}" "${CHECKSUM_URL}" 2>/dev/null; then
        if [ -s "${TMP_CHECKSUM}" ]; then
            CHECKSUM_OK=1
        fi
    fi
fi

if [ "$CHECKSUM_OK" = "1" ]; then
    info "Verifying SHA256 checksum..."

    # Detect checksum tool (sha256sum on Linux, shasum -a 256 on macOS).
    if command -v sha256sum > /dev/null 2>&1; then
        EXPECTED=$(grep "${BINARY_NAME}" "${TMP_CHECKSUM}" | awk '{print $1}')
        ACTUAL=$(sha256sum "${TMPFILE}" | awk '{print $1}')
    elif command -v shasum > /dev/null 2>&1; then
        EXPECTED=$(grep "${BINARY_NAME}" "${TMP_CHECKSUM}" | awk '{print $1}')
        ACTUAL=$(shasum -a 256 "${TMPFILE}" | awk '{print $1}')
    else
        warn "No sha256sum or shasum found — skipping checksum verification."
        EXPECTED=""
        ACTUAL=""
    fi

    if [ -n "$EXPECTED" ] && [ "$EXPECTED" != "$ACTUAL" ]; then
        err "SHA256 checksum mismatch!"
        err "Expected: $EXPECTED"
        err "Actual:   $ACTUAL"
        rm -f "${TMPFILE}" "${TMP_CHECKSUM}"
        exit 1
    fi

    if [ -n "$EXPECTED" ]; then
        info "Checksum verified."
    fi
    rm -f "${TMP_CHECKSUM}"
else
    warn "Skipping checksum verification (checksums.txt not available)."
    rm -f "${TMP_CHECKSUM}"
fi

# ── Install binary ──
mv "${TMPFILE}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"

info "ForgeCrew installed to ${INSTALL_DIR}/${BINARY}"

# ── Self-test ──
info "Running self-test: ${INSTALL_DIR}/${BINARY} version"
if VERSION_OUTPUT=$("${INSTALL_DIR}/${BINARY}" version 2>&1); then
    info "Self-test passed: ${VERSION_OUTPUT}"
else
    err "Self-test failed: ${VERSION_OUTPUT}"
    err "The binary may be corrupted or incompatible with this system."
    rm -f "${INSTALL_DIR}/${BINARY}"
    exit 1
fi

# ── PATH check ──
case ":$PATH:" in
    *:"${INSTALL_DIR}":*)
        # Already in PATH — nothing to do.
        ;;
    *)
        warn ""
        warn "NOTE: ${INSTALL_DIR} is not in your PATH."
        warn "Add the following line to your shell profile (~/.bashrc, ~/.zshrc, ~/.profile, etc.):"
        warn ""
        warn "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        warn ""
        warn "Then restart your terminal or run:"
        warn "  source ~/.bashrc   # or ~/.zshrc / ~/.profile"
        warn ""
        warn "To use forgecrew in this terminal session without restarting, run:"
        warn "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        ;;
esac

info "Run 'forgecrew init' in your project to get started."
