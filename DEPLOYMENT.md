# MyWeatherDash Deployment Guide

## Quick Install on Raspberry Pi

### Prerequisites
- Raspberry Pi with Raspberry Pi OS
- Go 1.21+ installed
- MariaDB/MySQL with WeeWX database
- Git

### Installation Steps

1. **Clone the repository:**
   ```bash
   git clone https://github.com/thurmanmarka/MyWeatherDash.git
   cd MyWeatherDash
   ```

2. **Configure the application:**
   ```bash
   cp config.example.yaml config.yaml
   nano config.yaml
   ```
   Edit database credentials and other settings.

3. **Run the install script:**
   ```bash
   chmod +x install.sh
   ./install.sh
   ```

4. **Enable and start the service:**
   ```bash
   sudo systemctl enable weatherdash
   sudo systemctl start weatherdash
   ```

5. **Check status:**
   ```bash
   sudo systemctl status weatherdash
   sudo journalctl -u weatherdash -f
   ```

6. **Access the dashboard:**
   - Open `http://<raspberry-pi-ip>:80` in your browser
   - Default port is 80 (configurable in `config.yaml`)

## Manual Installation

If you prefer manual installation:

1. Build for ARM:
   ```bash
   GOOS=linux GOARCH=arm GOARM=7 go build -o weatherdash .
   ```

2. Create directories:
   ```bash
   sudo mkdir -p /home/pi/weatherdash
   sudo cp weatherdash /home/pi/weatherdash/
   sudo cp -r static templates config.yaml /home/pi/weatherdash/
   ```

3. Install systemd service:
   ```bash
   sudo cp weatherdash.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable weatherdash
   sudo systemctl start weatherdash
   ```

## Health Check

The application exposes a `/health` endpoint for monitoring:

```bash
curl http://localhost/health
```

Response when healthy:
```json
{"status":"ok"}
```

## Service Management

- **Start:** `sudo systemctl start weatherdash`
- **Stop:** `sudo systemctl stop weatherdash`
- **Restart:** `sudo systemctl restart weatherdash`
- **Status:** `sudo systemctl status weatherdash`
- **Logs:** `sudo journalctl -u weatherdash -f`
- **Disable autostart:** `sudo systemctl disable weatherdash`

## Configuration

Edit `/home/pi/weatherdash/config.yaml`:
- Database credentials
- Server port (default: 80)
- Location coordinates
- Polling intervals

Restart the service after configuration changes:
```bash
sudo systemctl restart weatherdash
```

## Troubleshooting

### Service won't start
```bash
# Check logs
sudo journalctl -u weatherdash -n 50

# Verify database connection
mysql -h <host> -u <user> -p<password> <database> -e "SELECT COUNT(*) FROM archive;"

# Check permissions
ls -la /home/pi/weatherdash/
```

### Port 80 already in use
Edit `config.yaml` and change the port, or stop the conflicting service.

### Database connection errors
- Verify MariaDB is running: `sudo systemctl status mariadb`
- Check credentials in `config.yaml`
- Ensure WeeWX database exists and has data

## Uninstall

```bash
sudo systemctl stop weatherdash
sudo systemctl disable weatherdash
sudo rm /etc/systemd/system/weatherdash.service
sudo systemctl daemon-reload
sudo rm -rf /home/pi/weatherdash
```
