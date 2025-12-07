#!/bin/bash
# Self-signed SSL certificate setup for MyWeatherDash
# Run this on the Raspberry Pi after installation

set -e

echo "🔒 Setting up self-signed SSL certificate for MyWeatherDash..."

# Create directory for SSL certificates
SSL_DIR="/etc/nginx/ssl"
sudo mkdir -p $SSL_DIR

# Generate self-signed certificate (valid for 365 days)
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout $SSL_DIR/weatherdash.key \
    -out $SSL_DIR/weatherdash.crt \
    -subj "/C=US/ST=State/L=City/O=MyWeatherDash/CN=192.168.86.13"

# Set proper permissions
sudo chmod 600 $SSL_DIR/weatherdash.key
sudo chmod 644 $SSL_DIR/weatherdash.crt

echo "✅ SSL certificate created at $SSL_DIR/weatherdash.crt"
echo "✅ SSL key created at $SSL_DIR/weatherdash.key"
echo ""
echo "Next steps:"
echo "1. Copy the updated nginx.conf to /etc/nginx/nginx.conf"
echo "2. Test nginx config: sudo nginx -t"
echo "3. Reload nginx: sudo systemctl reload nginx"
echo "4. Access your dashboard at: https://192.168.86.13"
echo ""
echo "⚠️  Your browser will show a security warning because this is self-signed."
echo "    Click 'Advanced' → 'Proceed to 192.168.86.13' to continue."
