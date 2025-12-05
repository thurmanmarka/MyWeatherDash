#!/bin/bash
# MyWeatherDash Installation Script for Raspberry Pi
# This script installs the weatherdash binary and systemd service

set -e

INSTALL_DIR="/home/pi/weatherdash"
SERVICE_FILE="weatherdash.service"
BINARY_NAME="weatherdash"
CONFIG_FILE="config.yaml"

echo "=== MyWeatherDash Installation ==="
echo ""

# Check if running as pi user
if [ "$USER" != "pi" ]; then
    echo "Warning: This script is designed to run as the 'pi' user."
    echo "Current user: $USER"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Build the binary for ARM (Raspberry Pi)
echo "Building binary for ARM..."
GOOS=linux GOARCH=arm GOARM=7 go build -o $BINARY_NAME .

# Create installation directory
echo "Creating installation directory: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR/static"
mkdir -p "$INSTALL_DIR/templates"

# Copy files
echo "Copying files..."
cp $BINARY_NAME "$INSTALL_DIR/"
cp -r static/* "$INSTALL_DIR/static/"
cp -r templates/* "$INSTALL_DIR/templates/"

# Copy config if it doesn't exist
if [ ! -f "$INSTALL_DIR/$CONFIG_FILE" ]; then
    if [ -f "$CONFIG_FILE" ]; then
        cp $CONFIG_FILE "$INSTALL_DIR/"
        echo "Copied $CONFIG_FILE to $INSTALL_DIR"
    else
        echo "Warning: $CONFIG_FILE not found. You'll need to create it manually."
    fi
else
    echo "$CONFIG_FILE already exists in $INSTALL_DIR, skipping."
fi

# Set permissions
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Install systemd service
echo "Installing systemd service..."
sudo cp $SERVICE_FILE /etc/systemd/system/
sudo systemctl daemon-reload

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "1. Edit the config file: $INSTALL_DIR/$CONFIG_FILE"
echo "2. Enable the service: sudo systemctl enable weatherdash"
echo "3. Start the service: sudo systemctl start weatherdash"
echo "4. Check status: sudo systemctl status weatherdash"
echo "5. View logs: sudo journalctl -u weatherdash -f"
echo ""
