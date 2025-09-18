#!/bin/bash
# Create macOS DMG installer for CRGoDICOM

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DIST_DIR="$PROJECT_ROOT/dist"
MACOS_DIR="$DIST_DIR/macos"

# Configuration
APP_NAME="CRGoDICOM"
DMG_NAME="crgodicom-macos-$(date +%Y%m%d-%H%M%S)"
DMG_PATH="$DIST_DIR/$DMG_NAME.dmg"

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

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    log_error "This script can only be run on macOS"
    exit 1
fi

# Check if hdiutil is available
if ! command -v hdiutil >/dev/null 2>&1; then
    log_error "hdiutil not found. This script requires macOS."
    exit 1
fi

log_info "Creating macOS DMG installer for $APP_NAME"

# Create temporary directory for DMG contents
TEMP_DIR=$(mktemp -d)
log_info "Using temporary directory: $TEMP_DIR"

# Copy files to temporary directory
log_info "Copying application files..."

# Create Applications symlink
ln -s /Applications "$TEMP_DIR/Applications"

# Copy the appropriate binary based on architecture
ARCH=$(uname -m)
if [[ "$ARCH" == "arm64" ]]; then
    BINARY_NAME="crgodicom-darwin-arm64"
    log_info "Detected Apple Silicon (arm64), using $BINARY_NAME"
else
    BINARY_NAME="crgodicom-darwin-amd64"
    log_info "Detected Intel (amd64), using $BINARY_NAME"
fi

# Copy binary and rename it
cp "$MACOS_DIR/$BINARY_NAME" "$TEMP_DIR/$APP_NAME"
chmod +x "$TEMP_DIR/$APP_NAME"

# Copy configuration and documentation
cp "$MACOS_DIR/crgodicom.yaml" "$TEMP_DIR/"
cp "$MACOS_DIR/README.md" "$TEMP_DIR/"
cp "$MACOS_DIR/LICENSE" "$TEMP_DIR/"
cp "$MACOS_DIR/CHANGELOG.md" "$TEMP_DIR/"

# Copy examples directory
cp -r "$MACOS_DIR/examples" "$TEMP_DIR/"

# Create installer information
cat > "$TEMP_DIR/INSTALL.txt" << EOF
CRGoDICOM Installation Instructions
==================================

1. Drag CRGoDICOM to the Applications folder
2. Open Terminal and run: crgodicom --help
3. For DCMTK setup, run: crgodicom check-dcmtk --install-help

Features:
- Create synthetic DICOM studies
- Export to PDF/PNG with metadata overlay
- Send to PACS systems via DCMTK
- Template-based study generation
- Cross-platform compatibility

Documentation:
- README.md: Complete usage guide
- CHANGELOG.md: Version history
- examples/: Template examples

Support:
- GitHub: https://github.com/flatmapit/crgodicom
- Website: https://flatmapit.com

© 2025 FlatMapIt.com - Licensed under MIT License
EOF

# Set up DMG layout
log_info "Setting up DMG layout..."

# Create a simple background (optional)
# You could add a background image here if desired

# Create the DMG
log_info "Creating DMG: $DMG_PATH"

# Remove existing DMG if it exists
if [ -f "$DMG_PATH" ]; then
    rm "$DMG_PATH"
fi

# Create the DMG
hdiutil create -srcfolder "$TEMP_DIR" -volname "$APP_NAME" -fs HFS+ -fsargs "-c c=64,a=16,e=16" -format UDRW -size 100m "$DIST_DIR/temp-$DMG_NAME.dmg"

# Mount the DMG
log_info "Mounting DMG for customization..."
MOUNT_POINT=$(hdiutil attach "$DIST_DIR/temp-$DMG_NAME.dmg" -readwrite -noverify -noautoopen | grep -E '^/dev/' | sed 1q | awk '{print $3}')

# Wait for mount
sleep 2

# Set DMG properties
log_info "Setting DMG properties..."

# Simple DMG setup without complex Finder customization
log_info "DMG layout set up successfully"

# Wait for operations to complete
sleep 3

# Unmount the DMG
log_info "Unmounting DMG..."
hdiutil detach "$MOUNT_POINT"

# Convert to final compressed DMG
log_info "Converting to compressed DMG..."
hdiutil convert "$DIST_DIR/temp-$DMG_NAME.dmg" -format UDZO -imagekey zlib-level=9 -o "$DMG_PATH"

# Clean up temporary files
rm "$DIST_DIR/temp-$DMG_NAME.dmg"
rm -rf "$TEMP_DIR"

# Verify the DMG
if [ -f "$DMG_PATH" ]; then
    log_success "DMG created successfully: $DMG_PATH"
    log_info "DMG size: $(du -h "$DMG_PATH" | cut -f1)"
    log_info "DMG info:"
    hdiutil imageinfo "$DMG_PATH" | grep -E "(format|size|compressed)"
else
    log_error "Failed to create DMG"
    exit 1
fi

log_success "macOS DMG installer created successfully!"
log_info "You can now distribute: $DMG_PATH"
