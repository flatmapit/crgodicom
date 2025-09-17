#!/bin/bash
# Linux Post-Installation Script for CRGoDICOM
# This script runs after installation to set up the application

echo "Setting up CRGoDICOM for Linux..."

# Get the installation directory
INSTALL_DIR="$(dirname "$0")/../.."
BINARY_NAME="crgodicom"

# Determine the correct binary path
if [ -f "$INSTALL_DIR/usr/bin/$BINARY_NAME" ]; then
    BINARY_PATH="$INSTALL_DIR/usr/bin/$BINARY_NAME"
elif [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
else
    echo "Error: CRGoDICOM binary not found!"
    exit 1
fi

# Create symlink in /usr/local/bin for system-wide access
if [ ! -L "/usr/local/bin/$BINARY_NAME" ]; then
    echo "Creating system-wide command line symlink..."
    sudo ln -sf "$BINARY_PATH" "/usr/local/bin/$BINARY_NAME"
    echo "CRGoDICOM command line tool installed"
else
    echo "CRGoDICOM command line tool already installed"
fi

# Create desktop file for application menu integration
DESKTOP_FILE="/usr/share/applications/crgodicom.desktop"
if [ ! -f "$DESKTOP_FILE" ]; then
    echo "Creating desktop application entry..."
    sudo tee "$DESKTOP_FILE" > /dev/null << EOF
[Desktop Entry]
Name=CRGoDICOM
Comment=DICOM Medical Imaging Utility
Exec=$BINARY_NAME
Icon=crgodicom
Type=Application
Categories=Graphics;Medical;
Terminal=false
StartupNotify=true
MimeType=application/dicom;application/dcm;
EOF
    echo "Desktop application entry created"
else
    echo "Desktop application entry already exists"
fi

# Create MIME type association for .dcm files
MIME_PACKAGES_DIR="/usr/share/mime/packages"
MIME_FILE="$MIME_PACKAGES_DIR/crgodicom-dicom.xml"
if [ ! -f "$MIME_FILE" ]; then
    echo "Creating MIME type association..."
    sudo mkdir -p "$MIME_PACKAGES_DIR"
    sudo tee "$MIME_FILE" > /dev/null << EOF
<?xml version="1.0" encoding="UTF-8"?>
<mime-info xmlns="http://www.freedesktop.org/standards/shared-mime-info">
    <mime-type type="application/dicom">
        <comment>DICOM Medical Image</comment>
        <glob pattern="*.dcm"/>
        <glob pattern="*.dicom"/>
        <icon name="crgodicom"/>
        <sub-class-of type="application/octet-stream"/>
    </mime-type>
</mime-info>
EOF
    # Update MIME database
    sudo update-mime-database /usr/share/mime
    echo "MIME type association created"
else
    echo "MIME type association already exists"
fi

# Create application icon
ICON_DIR="/usr/share/icons/hicolor/256x256/apps"
if [ ! -f "$ICON_DIR/crgodicom.png" ]; then
    echo "Installing application icon..."
    sudo mkdir -p "$ICON_DIR"
    # Create a simple icon if none exists
    if [ -f "$INSTALL_DIR/usr/share/icons/hicolor/256x256/apps/crgodicom.png" ]; then
        sudo cp "$INSTALL_DIR/usr/share/icons/hicolor/256x256/apps/crgodicom.png" "$ICON_DIR/"
    else
        # Create a placeholder icon
        sudo convert -size 256x256 xc:blue -pointsize 24 -fill white -gravity center -annotate +0+0 "CRGo\nDICOM" "$ICON_DIR/crgodicom.png" 2>/dev/null || {
            echo "Warning: Could not create icon (ImageMagick not available)"
        }
    fi
    sudo gtk-update-icon-cache /usr/share/icons/hicolor/ 2>/dev/null || true
    echo "Application icon installed"
else
    echo "Application icon already exists"
fi

# Create user configuration directory
CONFIG_DIR="$HOME/.config/crgodicom"
if [ ! -d "$CONFIG_DIR" ]; then
    echo "Creating user configuration directory..."
    mkdir -p "$CONFIG_DIR"
    echo "Configuration directory created at $CONFIG_DIR"
fi

# Copy default configuration if it doesn't exist
if [ ! -f "$CONFIG_DIR/crgodicom.yaml" ]; then
    if [ -f "$INSTALL_DIR/usr/share/doc/crgodicom/crgodicom.yaml" ]; then
        cp "$INSTALL_DIR/usr/share/doc/crgodicom/crgodicom.yaml" "$CONFIG_DIR/"
        echo "Default configuration copied to $CONFIG_DIR"
    elif [ -f "$INSTALL_DIR/crgodicom.yaml" ]; then
        cp "$INSTALL_DIR/crgodicom.yaml" "$CONFIG_DIR/"
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

# Create desktop shortcut
DESKTOP_SHORTCUT="$HOME/Desktop/CRGoDICOM.desktop"
if [ ! -f "$DESKTOP_SHORTCUT" ]; then
    echo "Creating desktop shortcut..."
    cat > "$DESKTOP_SHORTCUT" << EOF
[Desktop Entry]
Name=CRGoDICOM
Comment=DICOM Medical Imaging Utility
Exec=$BINARY_NAME
Icon=crgodicom
Type=Application
Categories=Graphics;Medical;
Terminal=false
EOF
    chmod +x "$DESKTOP_SHORTCUT"
    echo "Desktop shortcut created"
else
    echo "Desktop shortcut already exists"
fi

# Set proper permissions
echo "Setting permissions..."
chmod +x "$BINARY_PATH"
chmod +x "/usr/local/bin/$BINARY_NAME" 2>/dev/null || true

# Update desktop database
if command -v update-desktop-database >/dev/null 2>&1; then
    sudo update-desktop-database /usr/share/applications 2>/dev/null || true
fi

echo ""
echo "CRGoDICOM installation completed successfully!"
echo ""
echo "Installation directory: $INSTALL_DIR"
echo "Binary location: $BINARY_PATH"
echo "Configuration directory: $CONFIG_DIR"
echo "Studies directory: $STUDIES_DIR"
echo "Command line tool: /usr/local/bin/$BINARY_NAME"
echo ""
echo "You can now run CRGoDICOM from:"
echo "  - Application menu (Graphics > Medical)"
echo "  - Command line: $BINARY_NAME"
echo "  - Desktop shortcut"
echo "  - File manager (double-click .dcm files)"
echo ""
echo "To uninstall, run:"
echo "  sudo rm -f /usr/local/bin/$BINARY_NAME"
echo "  sudo rm -f $DESKTOP_FILE"
echo "  sudo rm -f $MIME_FILE"
echo "  sudo rm -rf $ICON_DIR/crgodicom.png"
echo "  rm -rf $CONFIG_DIR"
echo "  rm -rf $STUDIES_DIR"
echo "  rm -f $DESKTOP_SHORTCUT"
echo ""
