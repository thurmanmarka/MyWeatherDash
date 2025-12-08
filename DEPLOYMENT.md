# MyWeatherDash Deployment Guide

This guide covers deploying MyWeatherDash as a standalone service on a Raspberry Pi or Linux system.

## 📋 Table of Contents
- [Prerequisites](#prerequisites)
- [Standalone Deployment](#standalone-deployment)
- [Manual Installation](#manual-installation)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements
- Raspberry Pi (or Linux server)
- Go 1.19+ (for building from source)
- WeeWX weather station with MariaDB database
- Network access to WeeWX database

### User Setup
Create a dedicated user for the service:
```bash
sudo useradd -m -s /bin/bash weatherdash
sudo usermod -aG sudo weatherdash
sudo su - weatherdash
```

## Standalone Deployment

### Quick Install (Automated)

The automated installation script handles everything:

```bash
# Clone the repository
git clone https://github.com/thurmanmarka/MyWeatherDash.git
cd MyWeatherDash

# Run the installation script
./install.sh

# Set up SSL certificates
./setup-ssl.sh

# Configure your database connection
nano /home/weatherdash/weatherdash/config.yaml

# Enable and start the service
sudo systemctl enable weatherdash
sudo systemctl start weatherdash
```

The dashboard will be available at:
- **HTTPS:** `https://your-pi-ip/`
- **HTTP:** `http://your-pi-ip/` (redirects to HTTPS)

### What Gets Installed

The installation script:

1. **Builds the binary** for ARM architecture (Raspberry Pi)
2. **Creates installation directory** at `/home/weatherdash/weatherdash/`
3. **Copies files:**
   - Binary (`weatherdash`)
   - Static files (`static/`)
   - Templates (`templates/`)
   - Configuration (`config.yaml`)
4. **Installs nginx** as reverse proxy
5. **Configures nginx** with SSL support
6. **Creates systemd service** for auto-start

### SSL Certificate Setup

The `setup-ssl.sh` script creates a self-signed SSL certificate:

```bash
./setup-ssl.sh
```

This generates:
- Certificate: `/etc/nginx/ssl/weatherdash.crt`
- Private key: `/etc/nginx/ssl/weatherdash.key`

**Note:** Browsers will show a security warning for self-signed certificates. Click "Advanced" → "Proceed" to continue.

For production deployments, consider using [Let's Encrypt](https://letsencrypt.org/) for trusted certificates.

## Manual Installation

If you prefer manual setup or need to customize the installation:

### 1. Build the Binary

```bash
# For Raspberry Pi (ARM)
GOOS=linux GOARCH=arm GOARM=7 go build -o weatherdash

# For standard Linux (x86_64)
GOOS=linux GOARCH=amd64 go build -o weatherdash
```

### 2. Create Directory Structure

```bash
sudo mkdir -p /home/weatherdash/weatherdash/{static,templates}
sudo chown -R weatherdash:weatherdash /home/weatherdash/weatherdash
```

### 3. Copy Files

```bash
cp weatherdash /home/weatherdash/weatherdash/
cp -r static/* /home/weatherdash/weatherdash/static/
cp -r templates/* /home/weatherdash/weatherdash/templates/
cp config.yaml /home/weatherdash/weatherdash/
chmod +x /home/weatherdash/weatherdash/weatherdash
```

### 4. Install and Configure nginx

```bash
# Install nginx
sudo apt update
sudo apt install nginx -y

# Backup existing config
sudo cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup

# Copy MyWeatherDash config
sudo cp nginx.conf /etc/nginx/nginx.conf

# Test configuration
sudo nginx -t

# Enable and start nginx
sudo systemctl enable nginx
sudo systemctl restart nginx
```

### 5. Create Systemd Service

Create `/etc/systemd/system/weatherdash.service`:

```ini
[Unit]
Description=MyWeatherDash Weather Dashboard
After=network.target mariadb.service
Wants=mariadb.service

[Service]
Type=simple
User=weatherdash
WorkingDirectory=/home/weatherdash/weatherdash
ExecStart=/home/weatherdash/weatherdash/weatherdash
Restart=always
RestartSec=10

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/home/weatherdash/weatherdash

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable weatherdash
sudo systemctl start weatherdash
```

## Configuration

### Database Connection

Edit `/home/weatherdash/weatherdash/config.yaml`:

```yaml
db:
  user: weewx
  password: "your_password"
  host: "192.168.1.100"
  port: 3306
  name: "weewx"
  params: "parseTime=false"

server:
  port: 8080  # Must be 8080 for nginx reverse proxy
  sse_poll_seconds: 60
  client_poll_seconds: 0

location:
  name: "My Weather Station"
  latitude: 33.4484
  longitude: -112.0740
  altitude: 1086
```

### Port Configuration

**Important:** When using nginx as a reverse proxy:
- MyWeatherDash listens on **port 8080** (localhost only)
- nginx listens on **ports 80/443** (public)
- nginx forwards requests to port 8080

### Nginx Configuration

The included `nginx.conf` provides:
- **HTTP to HTTPS redirect** (all HTTP traffic → HTTPS)
- **SSL/TLS termination** (ports 443)
- **Reverse proxy** to MyWeatherDash on port 8080
- **WebSocket/SSE support** for live updates
- **Static file caching**
- **Gzip compression**

## Troubleshooting

### Check Service Status

```bash
# Check weatherdash status
sudo systemctl status weatherdash

# View weatherdash logs
sudo journalctl -u weatherdash -f

# Check nginx status
sudo systemctl status nginx

# View nginx logs
sudo tail -f /var/log/nginx/error.log
sudo tail -f /var/log/nginx/access.log
```

### Common Issues

#### Service Won't Start
```bash
# Check for port conflicts
sudo netstat -tlnp | grep 8080

# Check configuration
cd /home/weatherdash/weatherdash
./weatherdash  # Run manually to see errors
```

#### Database Connection Errors
```bash
# Test database connection
mysql -h 192.168.1.100 -u weewx -p weewx

# Check config.yaml has correct credentials
cat /home/weatherdash/weatherdash/config.yaml
```

#### Nginx Configuration Errors
```bash
# Test nginx config
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx

# Check SSL certificate paths
ls -l /etc/nginx/ssl/
```

#### SSL Certificate Issues
```bash
# Regenerate certificate
./setup-ssl.sh

# Verify certificate
openssl x509 -in /etc/nginx/ssl/weatherdash.crt -text -noout
```

### Restart Services

```bash
# Restart weatherdash
sudo systemctl restart weatherdash

# Restart nginx
sudo systemctl restart nginx

# Restart both
sudo systemctl restart weatherdash nginx
```

### Update Application

```bash
cd ~/MyWeatherDash
git pull
go build -o weatherdash
sudo cp weatherdash /home/weatherdash/weatherdash/
sudo cp -r static/* /home/weatherdash/weatherdash/static/
sudo cp -r templates/* /home/weatherdash/weatherdash/templates/
sudo systemctl restart weatherdash
```

## Hub Integration

To deploy MyWeatherDash as part of MyHomeServicesHub instead:

1. Configure MyWeatherDash to run on port **8081** (not 8080)
2. Don't install the standalone nginx configuration
3. Let MyHomeServicesHub handle the reverse proxy routing
4. See [MyHomeServicesHub](../MyHomeServicesHub/README.md) for details

## Security Considerations

### Firewall Configuration

```bash
# Allow HTTP and HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Block direct access to backend port
sudo ufw deny 8080/tcp
```

### SSL Best Practices

For production deployments:

1. **Use Let's Encrypt** for trusted certificates
2. **Configure proper DNS** for your domain
3. **Update nginx config** with your domain name
4. **Enable HSTS** headers in nginx
5. **Set up certificate auto-renewal**

### Service Hardening

The systemd service includes security hardening:
- `NoNewPrivileges=true` - Prevents privilege escalation
- `PrivateTmp=true` - Isolates /tmp
- `ProtectSystem=strict` - Read-only system directories
- `ProtectHome=true` - Protects other user home directories

## Monitoring

### Log Locations

- **Application logs:** `sudo journalctl -u weatherdash`
- **Nginx access:** `/var/log/nginx/access.log`
- **Nginx errors:** `/var/log/nginx/error.log`

### Health Checks

```bash
# Check if service is responding
curl http://localhost:8080/api/ping

# Check through nginx
curl -k https://localhost/api/ping

# Check from another machine
curl -k https://192.168.1.100/api/ping
```

---

**Need help?** Open an issue on [GitHub](https://github.com/thurmanmarka/MyWeatherDash/issues)
