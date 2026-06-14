#!/bin/sh
# ============================================================================
# Shifter Install Script
# Version: 0.1.0
#
# One-liner installation:
#   curl -fsSL https://raw.githubusercontent.com/Moximxxx/shifter/main/install.sh | sh
#
# Or with a specific version:
#   curl -fsSL https://raw.githubusercontent.com/Moximxxx/shifter/main/install.sh | VERSION=0.1.0 sh
# ============================================================================

set -e

# --- Configuration ---
REPO="Moximxxx/shifter"
DEFAULT_VERSION="0.1.0"
VERSION="${VERSION:-$DEFAULT_VERSION}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY="shifter"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# --- Helpers ---
info()    { printf "${CYAN}→ %s${NC}\n" "$1"; }
success() { printf "${GREEN}✓ %s${NC}\n" "$1"; }
warn()    { printf "${YELLOW}⚠ %s${NC}\n" "$1"; }
error()   { printf "${RED}✗ %s${NC}\n" "$1"; exit 1; }

# --- Platform Detection ---
detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux)  os="linux" ;;
        Darwin) os="darwin" ;;
        *)      error "Unsupported OS: $(uname -s)" ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac

    echo "${os}-${arch}"
}

# --- Main ---
main() {
    echo ""
    printf "${CYAN}╔══════════════════════════════════════╗${NC}\n"
    printf "${CYAN}║       Shifter Installer v${VERSION}       ║${NC}\n"
    printf "${CYAN}║  Cross-agent config migration tool  ║${NC}\n"
    printf "${CYAN}╚══════════════════════════════════════╝${NC}\n"
    echo ""

    # Check for existing installation
    if command -v "$BINARY" >/dev/null 2>&1; then
        existing_version=$("$BINARY" --version 2>/dev/null || echo "unknown")
        info "Found existing installation: $existing_version"
        if [ "$existing_version" = "shifter version $VERSION" ]; then
            success "Already up to date (v$VERSION)"
            exit 0
        fi
        info "Upgrading to v$VERSION..."
    fi

    PLATFORM=$(detect_platform)
    ARCHIVE="shifter_${VERSION}_${PLATFORM}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ARCHIVE}"

    info "Platform: $PLATFORM"
    info "Downloading: $DOWNLOAD_URL"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    # Download
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE" || {
            warn "Download failed — falling back to go install"
            fallback_go_install
            return
        }
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_DIR/$ARCHIVE" || {
            warn "Download failed — falling back to go install"
            fallback_go_install
            return
        }
    else
        warn "Neither curl nor wget found — trying go install"
        fallback_go_install
        return
    fi

    # Extract
    info "Extracting..."
    tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

    # Install
    if [ -w "$INSTALL_DIR" ]; then
        cp "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
    else
        info "Need sudo to install to $INSTALL_DIR"
        sudo cp "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
    fi

    chmod +x "$INSTALL_DIR/$BINARY"
    success "Installed to $INSTALL_DIR/$BINARY"

    # Verify
    if command -v "$BINARY" >/dev/null 2>&1; then
        installed_version=$("$BINARY" --version 2>/dev/null || echo "ok")
        success "Installation verified: $installed_version"
    else
        warn "Binary installed but not in PATH"
        info "Add to PATH: export PATH=\"$INSTALL_DIR:\$PATH\""
    fi

    # Post-install
    echo ""
    info "Quick start:"
    echo "  shifter detect                     # Scan for configured agents"
    echo "  shifter port claude-code --to codex  # Port your first config"
    echo "  shifter ui                         # Interactive wizard"
    echo ""
    info "Setup environment defaults (optional):"
    echo "  shifter env init >> ~/.bashrc"

    echo ""
    success "Shifter v$VERSION installed successfully!"
}

# --- Go install fallback ---
fallback_go_install() {
    if command -v go >/dev/null 2>&1; then
        info "Installing via go install..."
        go install "github.com/${REPO}@v${VERSION}"
        success "Installed via go install"
    else
        error "Go is not installed. Please install Go first: https://go.dev/dl/"
    fi
}

main
