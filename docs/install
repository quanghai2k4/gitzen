#!/bin/bash
# GitZen Installer
# Auto-detect OS/arch, download from GitHub Releases, verify checksum, and install
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/quanghai2k4/gitzen/master/install.sh | bash
#   curl -sSL https://raw.githubusercontent.com/quanghai2k4/gitzen/master/install.sh | bash -s -- v0.1.0
#   curl -sSL https://raw.githubusercontent.com/quanghai2k4/gitzen/master/install.sh | bash -s -- --uninstall

set -e

# Configuration
REPO="quanghai2k4/gitzen"
BINARY_NAME="gitzen"
GITHUB_API="https://api.github.com/repos/${REPO}/releases"
GITHUB_DOWNLOAD="https://github.com/${REPO}/releases/download"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }

# Detect OS
detect_os() {
    local os
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        linux*) echo "linux" ;;
        darwin*) echo "darwin" ;;
        mingw*|msys*|cygwin*) error "Windows is not supported by this installer. Please download manually from GitHub Releases." ;;
        *) error "Unsupported OS: $os" ;;
    esac
}

# Detect Architecture
detect_arch() {
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) error "Unsupported architecture: $arch" ;;
    esac
}

# Get latest version from GitHub API
get_latest_version() {
    local version
    if command -v curl &> /dev/null; then
        version=$(curl -sL "${GITHUB_API}/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget &> /dev/null; then
        version=$(wget -qO- "${GITHUB_API}/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "curl or wget is required"
    fi
    
    if [ -z "$version" ]; then
        error "Failed to get latest version from GitHub"
    fi
    echo "$version"
}

# Download file
download() {
    local url="$1"
    local output="$2"
    
    if command -v curl &> /dev/null; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget &> /dev/null; then
        wget -q "$url" -O "$output"
    else
        error "curl or wget is required"
    fi
}

# Verify checksum
verify_checksum() {
    local file="$1"
    local checksums_file="$2"
    local filename
    filename=$(basename "$file")
    
    local expected_checksum
    expected_checksum=$(grep "$filename" "$checksums_file" | awk '{print $1}')
    
    if [ -z "$expected_checksum" ]; then
        warn "Checksum not found for $filename, skipping verification"
        return 0
    fi
    
    local actual_checksum
    if command -v sha256sum &> /dev/null; then
        actual_checksum=$(sha256sum "$file" | awk '{print $1}')
    elif command -v shasum &> /dev/null; then
        actual_checksum=$(shasum -a 256 "$file" | awk '{print $1}')
    else
        warn "sha256sum or shasum not found, skipping checksum verification"
        return 0
    fi
    
    if [ "$expected_checksum" != "$actual_checksum" ]; then
        error "Checksum verification failed!\nExpected: $expected_checksum\nActual: $actual_checksum"
    fi
    
    success "Checksum verified"
}

# Get install directory
get_install_dir() {
    if [ -w "/usr/local/bin" ]; then
        echo "/usr/local/bin"
    elif [ -d "$HOME/.local/bin" ]; then
        echo "$HOME/.local/bin"
    else
        mkdir -p "$HOME/.local/bin"
        echo "$HOME/.local/bin"
    fi
}

# Check if directory is in PATH
check_path() {
    local dir="$1"
    if [[ ":$PATH:" != *":$dir:"* ]]; then
        warn "$dir is not in your PATH"
        echo ""
        echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo "  export PATH=\"$dir:\$PATH\""
        echo ""
    fi
}

# Uninstall
uninstall() {
    info "Uninstalling ${BINARY_NAME}..."
    
    local locations=(
        "/usr/local/bin/${BINARY_NAME}"
        "$HOME/.local/bin/${BINARY_NAME}"
    )
    
    local found=false
    for loc in "${locations[@]}"; do
        if [ -f "$loc" ]; then
            if [ -w "$loc" ] || [ -w "$(dirname "$loc")" ]; then
                rm -f "$loc"
                success "Removed $loc"
                found=true
            else
                sudo rm -f "$loc"
                success "Removed $loc (with sudo)"
                found=true
            fi
        fi
    done
    
    if [ "$found" = false ]; then
        warn "${BINARY_NAME} is not installed"
    else
        success "${BINARY_NAME} has been uninstalled"
    fi
    exit 0
}

# Main install function
install() {
    local version="$1"
    
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║       GitZen Installer                ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════╝${NC}"
    echo ""
    
    # Detect platform
    local os arch
    os=$(detect_os)
    arch=$(detect_arch)
    info "Detected platform: ${os}/${arch}"
    
    # Get version
    if [ -z "$version" ]; then
        info "Fetching latest version..."
        version=$(get_latest_version)
    fi
    info "Version: ${version}"
    
    # Prepare download
    local version_num="${version#v}"  # Remove 'v' prefix if present
    local archive_name="${BINARY_NAME}_${version_num}_${os}_${arch}.tar.gz"
    local download_url="${GITHUB_DOWNLOAD}/${version}/${archive_name}"
    local checksums_url="${GITHUB_DOWNLOAD}/${version}/checksums.txt"
    
    # Create temp directory
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT
    
    # Download archive
    info "Downloading ${archive_name}..."
    download "$download_url" "${tmp_dir}/${archive_name}" || error "Failed to download ${archive_name}"
    success "Downloaded ${archive_name}"
    
    # Download checksums
    info "Downloading checksums..."
    download "$checksums_url" "${tmp_dir}/checksums.txt" || warn "Failed to download checksums"
    
    # Verify checksum
    if [ -f "${tmp_dir}/checksums.txt" ]; then
        verify_checksum "${tmp_dir}/${archive_name}" "${tmp_dir}/checksums.txt"
    fi
    
    # Extract archive
    info "Extracting..."
    tar -xzf "${tmp_dir}/${archive_name}" -C "${tmp_dir}"
    
    # Get install directory
    local install_dir
    install_dir=$(get_install_dir)
    
    # Install binary
    info "Installing to ${install_dir}..."
    if [ -w "$install_dir" ]; then
        mv "${tmp_dir}/${BINARY_NAME}" "${install_dir}/${BINARY_NAME}"
        chmod +x "${install_dir}/${BINARY_NAME}"
    else
        sudo mv "${tmp_dir}/${BINARY_NAME}" "${install_dir}/${BINARY_NAME}"
        sudo chmod +x "${install_dir}/${BINARY_NAME}"
    fi
    
    success "Installed ${BINARY_NAME} to ${install_dir}/${BINARY_NAME}"
    
    # Check PATH
    check_path "$install_dir"
    
    # Verify installation
    if command -v "$BINARY_NAME" &> /dev/null; then
        echo ""
        success "Installation complete!"
        echo ""
        echo "Run '${BINARY_NAME}' to start using GitZen"
        echo ""
    else
        echo ""
        success "Installation complete!"
        echo ""
        echo "Run '${install_dir}/${BINARY_NAME}' to start using GitZen"
        echo ""
    fi
}

# Parse arguments
main() {
    local version=""
    
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --uninstall|-u)
                uninstall
                ;;
            --help|-h)
                echo "GitZen Installer"
                echo ""
                echo "Usage:"
                echo "  install.sh [version]     Install specific version (e.g., v0.1.0)"
                echo "  install.sh --uninstall   Uninstall gitzen"
                echo "  install.sh --help        Show this help"
                echo ""
                echo "Examples:"
                echo "  curl -sSL https://raw.githubusercontent.com/quanghai2k4/gitzen/master/install.sh | bash"
                echo "  curl -sSL https://raw.githubusercontent.com/quanghai2k4/gitzen/master/install.sh | bash -s -- v0.1.0"
                exit 0
                ;;
            v*)
                version="$1"
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
        shift
    done
    
    install "$version"
}

main "$@"
