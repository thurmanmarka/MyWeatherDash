#!/bin/bash
# Deploy MyWeatherDash to Raspberry Pi
# Usage: ./deploy.sh [host]

set -e

HOST="${1:-192.168.86.13}"
USER="weatherdash"
REMOTE_DIR="/home/weatherdash/weatherdash"
SERVICE_NAME="weatherdash"

echo "🏗️  Building myweatherdash for ARM..."
GOOS=linux GOARCH=arm GOARM=7 go build -o weatherdash

echo "📦 Copying files to $HOST..."
scp weatherdash config.yaml ${USER}@${HOST}:${REMOTE_DIR}/
scp templates/index.html ${USER}@${HOST}:${REMOTE_DIR}/templates/

echo "🔧 Setting permissions and restarting service..."
ssh ${USER}@${HOST} << 'EOF'
chmod +x /home/weatherdash/weatherdash/weatherdash
sudo systemctl restart weatherdash
sudo systemctl status weatherdash --no-pager -l
EOF

echo "✅ Deployment complete!"
echo "🌐 Hub access: https://${HOST}"
echo "🌤️  Direct access: https://${HOST}:8081 (standalone mode)"
