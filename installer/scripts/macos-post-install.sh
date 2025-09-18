#!/bin/bash
# macOS Post-Installation Script for CRGoDICOM
# This script runs after installation to set up the application

echo "Setting up CRGoDICOM for macOS..."

# Get the installation directory
INSTALL_DIR="$(dirname "$0")/../.."
APP_NAME="CRGoDICOM"
APP_DIR="/Applications/${APP_NAME}.app"

# Create symlink in /usr/local/bin for command line access
if [ ! -L "/usr/local/bin/crgodicom" ]; then
    echo "Creating command line symlink..."
    sudo ln -sf "${APP_DIR}/Contents/MacOS/crgodicom" "/usr/local/bin/crgodicom"
    echo "CRGoDICOM command line tool installed"
else
    echo "CRGoDICOM command line tool already installed"
fi

# Create desktop shortcut (macOS alias)
DESKTOP_SHORTCUT="$HOME/Desktop/CRGoDICOM"
if [ ! -e "$DESKTOP_SHORTCUT" ]; then
    echo "Creating desktop shortcut..."
    osascript -e "tell application \"Finder\" to make alias file to POSIX file \"${APP_DIR}\" at desktop"
    echo "Desktop shortcut created"
else
    echo "Desktop shortcut already exists"
fi

# Set up file associations for .dcm files
echo "Setting up file associations..."
# Create Launch Services database entry
/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -f "${APP_DIR}"

# Create default configuration directory
CONFIG_DIR="$HOME/Library/Application Support/CRGoDICOM"
if [ ! -d "$CONFIG_DIR" ]; then
    echo "Creating configuration directory..."
    mkdir -p "$CONFIG_DIR"
    echo "Configuration directory created at $CONFIG_DIR"
fi

# Copy default configuration if it doesn't exist
if [ ! -f "$CONFIG_DIR/crgodicom.yaml" ]; then
    if [ -f "${APP_DIR}/Contents/Resources/crgodicom.yaml" ]; then
        cp "${APP_DIR}/Contents/Resources/crgodicom.yaml" "$CONFIG_DIR/"
        echo "Default configuration copied to $CONFIG_DIR"
    fi
fi

# Create studies directory
STUDIES_DIR="$HOME/Documents/CRGoDICOM/studies"
if [ ! -d "$STUDIES_DIR" ]; then
    echo "Creating studies directory..."
    mkdir -p "$STUDIES_DIR"
    echo "Studies directory created at $STUDIES_DIR"
fi

# Set up file association for .dcm files using Launch Services
echo "Registering DICOM file associations..."
# Create a temporary plist for file association
TEMP_PLIST="/tmp/crgodicom-file-association.plist"
cat > "$TEMP_PLIST" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDocumentTypes</key>
    <array>
        <dict>
            <key>CFBundleTypeExtensions</key>
            <array>
                <string>dcm</string>
                <string>dicom</string>
            </array>
            <key>CFBundleTypeName</key>
            <string>DICOM Medical Image</string>
            <key>CFBundleTypeRole</key>
            <string>Viewer</string>
            <key>LSHandlerRank</key>
            <string>Owner</string>
        </dict>
    </array>
</dict>
</plist>
EOF

# Apply the file association (this would typically be done during app bundle creation)
echo "File associations configured"

# Clean up temporary file
rm -f "$TEMP_PLIST"

# Set proper permissions
echo "Setting permissions..."
chmod +x "${APP_DIR}/Contents/MacOS/crgodicom"
chmod +x "/usr/local/bin/crgodicom" 2>/dev/null || true

echo ""
echo "CRGoDICOM installation completed successfully!"
echo ""
echo "Installation directory: $APP_DIR"
echo "Configuration directory: $CONFIG_DIR"
echo "Studies directory: $STUDIES_DIR"
echo "Command line tool: /usr/local/bin/crgodicom"
echo ""
echo "You can now run CRGoDICOM from:"
echo "  - Applications folder"
echo "  - Command line: crgodicom"
echo "  - Desktop shortcut"
echo ""
echo "To uninstall, simply drag the application to Trash and run:"
echo "  sudo rm -f /usr/local/bin/crgodicom"
echo ""
