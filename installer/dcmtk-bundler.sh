#!/bin/bash
# DCMTK Bundler Script for CRGoDICOM Installers
# This script handles DCMTK bundling for different platforms

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/dist"

# DCMTK versions to bundle
DCMTK_VERSION="3.6.7"
DCMTK_PREBUILT_VERSION="3.6.7"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to detect platform
detect_platform() {
    case "$(uname -s)" in
        Linux*)     PLATFORM="linux";;
        Darwin*)    PLATFORM="macos";;
        CYGWIN*|MINGW32*|MSYS*|MINGW*) PLATFORM="windows";;
        *)          PLATFORM="unknown";;
    esac
    
    case "$(uname -m)" in
        x86_64)     ARCH="amd64";;
        arm64|aarch64) ARCH="arm64";;
        *)          ARCH="unknown";;
    esac
    
    echo "${PLATFORM}-${ARCH}"
}

# Function to download DCMTK prebuilt binaries
download_dcmtk_prebuilt() {
    local platform="$1"
    local target_dir="$2"
    
    log_info "Downloading DCMTK prebuilt binaries for $platform..."
    
    case "$platform" in
        "linux-amd64")
            # Try to download from GitHub releases or official DCMTK site
            # For now, we'll provide instructions for manual download
            log_warning "Prebuilt DCMTK binaries not available for automatic download"
            log_info "Please download DCMTK manually from: https://dicom.offis.de/download/dcmtk/"
            return 1
            ;;
        "macos-amd64"|"macos-arm64")
            log_info "Checking Homebrew for DCMTK..."
            if command -v brew >/dev/null 2>&1; then
                log_info "Installing DCMTK via Homebrew..."
                brew install dcmtk
                return 0
            else
                log_warning "Homebrew not available. Please install DCMTK manually."
                return 1
            fi
            ;;
        "windows-amd64")
            log_info "Checking Chocolatey for DCMTK..."
            if command -v choco >/dev/null 2>&1; then
                log_info "Installing DCMTK via Chocolatey..."
                choco install dcmtk -y
                return 0
            else
                log_warning "Chocolatey not available. Please install DCMTK manually."
                return 1
            fi
            ;;
        *)
            log_error "Unsupported platform: $platform"
            return 1
            ;;
    esac
}

# Function to create DCMTK bundle structure
create_dcmtk_bundle() {
    local platform="$1"
    local bundle_dir="$2"
    
    log_info "Creating DCMTK bundle structure for $platform..."
    
    mkdir -p "$bundle_dir/dcmtk/bin"
    mkdir -p "$bundle_dir/dcmtk/lib"
    mkdir -p "$bundle_dir/dcmtk/share"
    
    # Create version file
    echo "$DCMTK_VERSION" > "$bundle_dir/dcmtk/VERSION"
    
    # Create README
    cat > "$bundle_dir/dcmtk/README.md" << EOF
# DCMTK Bundle for CRGoDICOM

This directory contains DCMTK (DICOM Toolkit) binaries bundled with CRGoDICOM.

## Version
DCMTK Version: $DCMTK_VERSION

## Included Tools
- storescu: Send DICOM files to PACS
- echoscu: Test DICOM connectivity
- dcmdump: Inspect DICOM files
- dcmodify: Modify DICOM files

## Platform
Platform: $platform

## Usage
CRGoDICOM will automatically detect and use these bundled DCMTK tools
when system DCMTK is not available.

## License
DCMTK is licensed under the BSD 3-Clause License.
See: https://dicom.offis.de/dcmtk.php.en

## Source
Official DCMTK website: https://dicom.offis.de/download/dcmtk/
EOF

    log_success "DCMTK bundle structure created"
}

# Function to copy DCMTK binaries
copy_dcmtk_binaries() {
    local platform="$1"
    local bundle_dir="$2"
    
    log_info "Copying DCMTK binaries for $platform..."
    
    # List of required DCMTK tools
    local tools=("storescu" "echoscu" "dcmdump" "dcmodify")
    
    for tool in "${tools[@]}"; do
        local tool_path=""
        
        case "$platform" in
            "linux-amd64"|"macos-amd64"|"macos-arm64")
                tool_path=$(which "$tool" 2>/dev/null || echo "")
                ;;
            "windows-amd64")
                tool_path=$(which "${tool}.exe" 2>/dev/null || echo "")
                ;;
        esac
        
        if [ -n "$tool_path" ] && [ -f "$tool_path" ]; then
            log_info "Copying $tool from $tool_path"
            cp "$tool_path" "$bundle_dir/dcmtk/bin/"
            
            # Copy dependencies if needed
            case "$platform" in
                "linux-amd64")
                    # Copy shared libraries
                    ldd "$tool_path" 2>/dev/null | grep "=>" | awk '{print $3}' | while read lib; do
                        if [ -f "$lib" ]; then
                            cp "$lib" "$bundle_dir/dcmtk/lib/" 2>/dev/null || true
                        fi
                    done
                    ;;
            esac
        else
            log_warning "Tool $tool not found in PATH"
        fi
    done
    
    # Copy libraries and data files if they exist
    case "$platform" in
        "linux-amd64")
            # Copy DCMTK data files
            if [ -d "/usr/share/dcmtk" ]; then
                cp -r /usr/share/dcmtk/* "$bundle_dir/dcmtk/share/" 2>/dev/null || true
            fi
            ;;
        "macos-amd64"|"macos-arm64")
            # Copy DCMTK data files from Homebrew
            if [ -d "/usr/local/share/dcmtk" ]; then
                cp -r /usr/local/share/dcmtk/* "$bundle_dir/dcmtk/share/" 2>/dev/null || true
            elif [ -d "/opt/homebrew/share/dcmtk" ]; then
                cp -r /opt/homebrew/share/dcmtk/* "$bundle_dir/dcmtk/share/" 2>/dev/null || true
            fi
            ;;
    esac
    
    log_success "DCMTK binaries copied"
}

# Function to create installer integration
create_installer_integration() {
    local platform="$1"
    local bundle_dir="$2"
    
    log_info "Creating installer integration for $platform..."
    
    # Create post-install script that checks for DCMTK
    case "$platform" in
        "windows-amd64")
            cat > "$bundle_dir/check-dcmtk.bat" << 'EOF'
@echo off
REM Check DCMTK availability for CRGoDICOM

echo Checking DCMTK installation...

REM Check if DCMTK is in PATH
where storescu >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ DCMTK found in system PATH
    echo Running CRGoDICOM DCMTK check...
    crgodicom check-dcmtk
    goto :end
)

REM Check bundled DCMTK
if exist "%~dp0dcmtk\bin\storescu.exe" (
    echo ✅ Bundled DCMTK found
    echo Setting up bundled DCMTK...
    set PATH=%~dp0dcmtk\bin;%PATH%
    crgodicom check-dcmtk
    goto :end
)

echo ❌ DCMTK not found
echo.
echo Please install DCMTK or use the bundled version.
echo Run 'crgodicom check-dcmtk --install-help' for instructions.

:end
pause
EOF
            ;;
        "macos-amd64"|"macos-arm64")
            cat > "$bundle_dir/check-dcmtk.sh" << 'EOF'
#!/bin/bash
# Check DCMTK availability for CRGoDICOM

echo "Checking DCMTK installation..."

# Check if DCMTK is in PATH
if command -v storescu >/dev/null 2>&1; then
    echo "✅ DCMTK found in system PATH"
    echo "Running CRGoDICOM DCMTK check..."
    crgodicom check-dcmtk
    exit 0
fi

# Check bundled DCMTK
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/dcmtk/bin/storescu" ]; then
    echo "✅ Bundled DCMTK found"
    echo "Setting up bundled DCMTK..."
    export PATH="$SCRIPT_DIR/dcmtk/bin:$PATH"
    crgodicom check-dcmtk
    exit 0
fi

echo "❌ DCMTK not found"
echo ""
echo "Please install DCMTK or use the bundled version."
echo "Run 'crgodicom check-dcmtk --install-help' for instructions."
EOF
            chmod +x "$bundle_dir/check-dcmtk.sh"
            ;;
        "linux-amd64")
            cat > "$bundle_dir/check-dcmtk.sh" << 'EOF'
#!/bin/bash
# Check DCMTK availability for CRGoDICOM

echo "Checking DCMTK installation..."

# Check if DCMTK is in PATH
if command -v storescu >/dev/null 2>&1; then
    echo "✅ DCMTK found in system PATH"
    echo "Running CRGoDICOM DCMTK check..."
    crgodicom check-dcmtk
    exit 0
fi

# Check bundled DCMTK
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/dcmtk/bin/storescu" ]; then
    echo "✅ Bundled DCMTK found"
    echo "Setting up bundled DCMTK..."
    export PATH="$SCRIPT_DIR/dcmtk/bin:$PATH"
    crgodicom check-dcmtk
    exit 0
fi

echo "❌ DCMTK not found"
echo ""
echo "Please install DCMTK or use the bundled version."
echo "Run 'crgodicom check-dcmtk --install-help' for instructions."
EOF
            chmod +x "$bundle_dir/check-dcmtk.sh"
            ;;
    esac
    
    log_success "Installer integration created"
}

# Main function
main() {
    local platform="$1"
    local bundle_dir="$2"
    
    if [ -z "$platform" ]; then
        platform=$(detect_platform)
    fi
    
    if [ -z "$bundle_dir" ]; then
        bundle_dir="$BUILD_DIR/dcmtk-bundle-$platform"
    fi
    
    log_info "Creating DCMTK bundle for platform: $platform"
    log_info "Bundle directory: $bundle_dir"
    
    # Create bundle directory
    mkdir -p "$bundle_dir"
    
    # Create DCMTK bundle structure
    create_dcmtk_bundle "$platform" "$bundle_dir"
    
    # Try to download/copy DCMTK binaries
    if ! download_dcmtk_prebuilt "$platform" "$bundle_dir"; then
        log_info "Attempting to copy system DCMTK binaries..."
        copy_dcmtk_binaries "$platform" "$bundle_dir"
    fi
    
    # Create installer integration
    create_installer_integration "$platform" "$bundle_dir"
    
    log_success "DCMTK bundle created successfully!"
    log_info "Bundle location: $bundle_dir"
    
    # Show bundle contents
    log_info "Bundle contents:"
    find "$bundle_dir" -type f | sort
}

# Script usage
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Usage: $0 [platform] [bundle_dir]"
    echo ""
    echo "Creates a DCMTK bundle for CRGoDICOM installers."
    echo ""
    echo "Arguments:"
    echo "  platform    Target platform (auto-detected if not specified)"
    echo "  bundle_dir  Output directory (default: dist/dcmtk-bundle-<platform>)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Auto-detect platform"
    echo "  $0 linux-amd64                       # Specify platform"
    echo "  $0 windows-amd64 /tmp/dcmtk-bundle   # Specify platform and directory"
    echo ""
    echo "Supported platforms:"
    echo "  linux-amd64, macos-amd64, macos-arm64, windows-amd64"
    exit 0
fi

# Run main function
main "$1" "$2"
