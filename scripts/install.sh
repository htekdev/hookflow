#!/bin/sh
# hookflow install script
# Usage: curl -sSL https://raw.githubusercontent.com/htekdev/hookflow/main/scripts/install.sh | sh

set -e

REPO="htekdev/hookflow"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    printf "${GREEN}[INFO]${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}[WARN]${NC} %s\n" "$1"
}

error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1"
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*)  echo "darwin" ;;
        Linux*)   echo "linux" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *)        error "Unsupported OS: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)     echo "amd64" ;;
        arm64|aarch64)    echo "arm64" ;;
        *)                error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest release version from GitHub
get_latest_version() {
    curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | \
        grep '"tag_name":' | \
        sed -E 's/.*"([^"]+)".*/\1/' || echo ""
}

# Main installation
main() {
    info "Installing hookflow CLI..."

    OS=$(detect_os)
    ARCH=$(detect_arch)
    
    info "Detected: ${OS}-${ARCH}"

    # Get version
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(get_latest_version)
        if [ -z "$VERSION" ]; then
            warn "Could not fetch latest version, trying direct download..."
            VERSION="latest"
        fi
    fi

    # Build binary name
    EXT=""
    if [ "$OS" = "windows" ]; then
        EXT=".exe"
    fi
    BINARY_NAME="hookflow-${OS}-${ARCH}${EXT}"

    # Build download URL
    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}"
    else
        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"
    fi

    info "Downloading from: ${DOWNLOAD_URL}"

    # Create temp file
    TMP_FILE=$(mktemp)
    trap "rm -f ${TMP_FILE}" EXIT

    # Download
    if command -v curl >/dev/null 2>&1; then
        curl -sL -o "${TMP_FILE}" "${DOWNLOAD_URL}" || error "Download failed"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "${TMP_FILE}" "${DOWNLOAD_URL}" || error "Download failed"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi

    # Check if download succeeded
    if [ ! -s "${TMP_FILE}" ]; then
        error "Downloaded file is empty. The release may not exist yet."
    fi

    # Determine install location
    TARGET="${INSTALL_DIR}/hookflow"
    
    # Check if we can write to install directory
    if [ ! -w "${INSTALL_DIR}" ]; then
        # Try to create with sudo
        if command -v sudo >/dev/null 2>&1; then
            info "Installing to ${TARGET} (requires sudo)..."
            sudo mkdir -p "${INSTALL_DIR}"
            sudo mv "${TMP_FILE}" "${TARGET}"
            sudo chmod +x "${TARGET}"
        else
            # Fall back to user directory
            INSTALL_DIR="${HOME}/.local/bin"
            TARGET="${INSTALL_DIR}/hookflow"
            mkdir -p "${INSTALL_DIR}"
            mv "${TMP_FILE}" "${TARGET}"
            chmod +x "${TARGET}"
            warn "Installed to ${TARGET}"
            warn "Make sure ${INSTALL_DIR} is in your PATH"
        fi
    else
        mv "${TMP_FILE}" "${TARGET}"
        chmod +x "${TARGET}"
    fi

    info "✓ Installed hookflow to ${TARGET}"

    # Verify installation
    if command -v hookflow >/dev/null 2>&1; then
        info "✓ $(hookflow version)"
    else
        warn "hookflow installed but not in PATH"
        warn "Add ${INSTALL_DIR} to your PATH or run: ${TARGET} version"
    fi

    info ""
    info "Get started:"
    info "  cd your-project"
    info "  hookflow init"
    info "  hookflow create \"block edits to .env files\""
}

main "$@"
