#!/bin/sh
# ============================================================================
# Shifter Install Script
# Version: 1.0.0
#
# One-liner installation:
#   curl -fsSL https://raw.githubusercontent.com/Moximxxx/shifter/release/install.sh | sh
#
# Or with a specific version:
#   curl -fsSL https://raw.githubusercontent.com/Moximxxx/shifter/release/install.sh | VERSION=1.0.0 sh
# ============================================================================

set -e

# --- Configuration ---
REPO="Moximxxx/shifter"
DEFAULT_VERSION="1.0.2"
VERSION="${VERSION:-$DEFAULT_VERSION}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY="shifter"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# --- Helpers ---
# info: white text, no color
info()    { printf "→ %s\n" "$1"; }
# success: green checkmark line
success() { printf "${GREEN}✓ %s${NC}\n" "$1"; }
# warn: yellow warning line
warn()    { printf "${YELLOW}⚠ %s${NC}\n" "$1"; }
# error: red error line
error()   { printf "${RED}✗ %s${NC}\n" "$1"; exit 1; }

# --- Platform Detection ---
detect_platform() {
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
    printf "${MAGENTA}"
    cat << 'SHIFTER_LOGO'
███████╗██╗  ██╗██╗███████╗████████╗███████╗██████╗
██╔════╝██║  ██║██║██╔════╝╚══██╔══╝██╔════╝██╔══██╗
███████╗███████║██║█████╗     ██║   █████╗  ██████╔╝
╚════██║██╔══██║██║██╔══╝     ██║   ██╔══╝  ██╔══██╗
███████║██║  ██║██║██║        ██║   ███████╗██║  ██║
╚══════╝╚═╝  ╚═╝╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═╝
SHIFTER_LOGO
    printf "${NC}\n"

    # Ensure install dir exists and is in PATH
    mkdir -p "$INSTALL_DIR"

    info "Installing Shifter v${VERSION}..."
    # Clean up old installations (ignore permission errors)
    for old in "$HOME/.local/bin/$BINARY" "$HOME/go/bin/$BINARY"; do
        rm -f "$old" 2>/dev/null || true
    done
    # /usr/local/bin needs sudo — skip if not writable
    if [ -f "/usr/local/bin/$BINARY" ] && [ -w "/usr/local/bin/$BINARY" ]; then
        rm -f "/usr/local/bin/$BINARY"
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

    # Extract (tar contains shifter-<platform> binary)
    info "Extracting..."
    tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

    # Find the extracted binary (named shifter-<platform>)
    EXTRACTED_BIN=$(find "$TMP_DIR" -name "shifter-*" -type f | head -1)
    if [ -z "$EXTRACTED_BIN" ]; then
        error "Binary not found in archive"
    fi

    # Install
    if [ -w "$INSTALL_DIR" ]; then
        cp "$EXTRACTED_BIN" "$INSTALL_DIR/$BINARY"
        chmod +x "$INSTALL_DIR/$BINARY"
    else
        info "Need sudo to install to $INSTALL_DIR"
        sudo cp "$EXTRACTED_BIN" "$INSTALL_DIR/$BINARY"
        sudo chmod +x "$INSTALL_DIR/$BINARY"
    fi

    success "Installed to $INSTALL_DIR/$BINARY"

    # Verify
    if command -v "$BINARY" >/dev/null 2>&1; then
        printf "${GREEN}✓${NC} Installed: "
        "$BINARY" --version 2>/dev/null || true
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
    # Clear shell command cache so 'shifter' resolves to new binary
    hash -r 2>/dev/null || true

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
