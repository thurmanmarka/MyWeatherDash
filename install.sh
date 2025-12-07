#!/bin/bash
# MyWeatherDash Installation Script for Raspberry Pi
# This script installs nginx, weatherdash binary, and systemd services

set -e

INSTALL_DIR="/home/pi/weatherdash"
SERVICE_FILE="weatherdash.service"
BINARY_NAME="weatherdash"
CONFIG_FILE="config.yaml"
NGINX_CONF="nginx.conf"

echo "=== MyWeatherDash Platform Installation ==="
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

# Install nginx
echo ""
echo "Installing nginx..."
if command -v nginx &> /dev/null; then
    echo "nginx is already installed."
else
    sudo apt update
    sudo apt install nginx -y
    echo "nginx installed successfully."
fi

# Backup existing nginx config
if [ -f /etc/nginx/nginx.conf ]; then
    echo "Backing up existing nginx config..."
    sudo cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup.$(date +%Y%m%d_%H%M%S)
fi

# Install nginx configuration
echo "Installing nginx configuration..."
sudo cp $NGINX_CONF /etc/nginx/nginx.conf

# Test nginx configuration
echo "Testing nginx configuration..."
sudo nginx -t

# Install systemd service
echo "Installing weatherdash systemd service..."
sudo cp $SERVICE_FILE /etc/systemd/system/
sudo systemctl daemon-reload

# Enable and start nginx
echo "Enabling nginx service..."
sudo systemctl enable nginx
sudo systemctl restart nginx

echo ""
echo "=== Installation complete! ==="
echo ""
echo "Services installed:"
echo "  - nginx (reverse proxy on port 80)"
echo "  - weatherdash (backend on port 8080)"
echo ""
echo "Next steps:"
echo "1. Edit the config file: nano $INSTALL_DIR/$CONFIG_FILE"
echo "   - Ensure port is set to 8080"
echo "   - Configure database credentials"
echo "2. Enable weatherdash: sudo systemctl enable weatherdash"
echo "3. Start weatherdash: sudo systemctl start weatherdash"
echo "4. Check status: sudo systemctl status weatherdash"
echo "5. View logs: sudo journalctl -u weatherdash -f"
echo ""
echo "Access your dashboard:"
echo "  - Landing page: http://$(hostname -I | awk '{print $1}')/"
echo "  - Weather Dashboard: http://$(hostname -I | awk '{print $1}')/weather"
echo "  - From this Pi: http://localhost/"
echo ""
echo "Troubleshooting:"
echo "  - nginx logs: sudo tail -f /var/log/nginx/error.log"
echo "  - Test nginx config: sudo nginx -t"
echo "  - Reload nginx: sudo systemctl reload nginx"
echo ""
