# MyWeatherDash Deployment Guide

## Overview

MyWeatherDash is a real-time weather dashboard that displays data from a WeeWX weather station database. It supports two deployment modes:

1. **Standalone Mode** - Direct access on port 8081
2. **Hub Mode** - Integrated with MyHomeServicesHub for authentication and multi-service access

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Deployment Modes                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Standalone:                                                 │
│  Browser → nginx (443) → MyWeatherDash (8081) → MariaDB     │
│                                                              │
│  Hub Integration:                                            │
│  Browser → nginx (443) → Hub Gateway (8080)                 │
│                              ↓                               │
│                    MyWeatherDash (8081) → MariaDB           │
│                    (receives auth headers)                   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## 📋 Table of Contents
- [Prerequisites](#prerequisites)
- [Fresh Installation](#fresh-installation)
- [Deployment Modes](#deployment-modes)
  - [Standalone Mode](#standalone-mode)
  - [Hub Integration Mode](#hub-integration-mode)
- [Permission System (Hub Mode)](#permission-system-hub-mode)
- [Updating/Redeployment](#updatingredeployment)
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

## Hub Integration Mode

When integrated with MyHomeServicesHub:

### 1. Auth Header Forwarding

The hub gateway forwards these headers to MyWeatherDash:
- `X-Hub-User`: Username
- `X-Hub-Role`: User role (admin/guest)
- `X-Hub-Authenticated`: "true"

### 2. Role-Based Features

**Admin users get:**
- Full dashboard access
- NOAA Reports (monthly/yearly climate summaries)
- Easy Button (custom date range CSV export)

**Guest users get:**
- Dashboard viewing only
- No database-intensive features (prevents SQL spam)

### 3. nginx Configuration

Routes go through hub gateway (port 8080), not directly to MyWeatherDash:

```nginx
# In /etc/nginx/nginx.conf
location /weather {
    proxy_pass http://localhost:8080/weather;  # Hub gateway, not :8081
    # ... proxy headers
}

location /api/ {
    proxy_pass http://localhost:8080/api/;  # Through hub
    # ... SSE settings
}

location /static/ {
    proxy_pass http://localhost:8080/static/;  # Through hub
}
```

See [MyHomeServicesHub DEPLOYMENT.md](../MyHomeServicesHub/DEPLOYMENT.md) for complete hub setup.

## Permission System (Hub Mode Only)

### Template Conditionals

The dashboard template uses role-based rendering with Go templates:

```html
<!-- NOAA Reports button - admin only -->
<button id="noaaBtn">📊 NOAA Reports</button>

<!-- Easy Button - admin only -->
{{if .IsAdmin}}
<button id="easyBtn">
    <svg>...</svg>
    Easy Button
</button>
{{end}}

<!-- Easy Button Modal - admin only -->
{{if .IsAdmin}}
<div id="easyModal" class="noaa-modal">
    <!-- ... modal content ... -->
</div>
{{end}}

<!-- Easy Button script - admin only -->
{{if .IsAdmin}}
<script src="/static/js/easy.js?v={{ .AssetVersion }}"></script>
{{end}}
```

### Backend Permission Checks

Protected endpoints verify admin role:

```go
// handlers.go - Auth helper functions
func getUserRole(r *http.Request) string {
    role := r.Header.Get("X-Hub-Role")
    if role == "" {
        role = "admin" // Default for standalone mode
    }
    return role
}

func isAdmin(r *http.Request) bool {
    return getUserRole(r) == "admin"
}

// NOAA endpoints require admin role
func handleNOAAMonthly(w http.ResponseWriter, r *http.Request) {
    if !isAdmin(r) {
        http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
        return
    }
    // ... generate NOAA report
}
```

### Handler Registration

```go
// main.go
http.HandleFunc("/", handleWeatherDash)  // Passes .IsAdmin to template
http.HandleFunc("/api/noaa/monthly", handleNOAAMonthly)  // Protected
http.HandleFunc("/api/noaa/yearly", handleNOAAYearly)    // Protected
```

### How It Works

1. **User logs into hub** with username/password
2. **Hub creates session** with role (admin/guest)
3. **nginx routes request** to hub gateway
4. **Hub gateway checks session** and adds auth headers
5. **MyWeatherDash reads headers** and sets `IsAdmin` flag
6. **Template conditionally renders** features based on role
7. **Backend endpoints verify** admin role before processing

## Updating/Redeployment

### Quick Update Script

Use the included `deploy.sh` script:

```bash
# From your development machine
./deploy.sh 192.168.86.13
```

### Manual Update

```bash
# Build for ARM
GOOS=linux GOARCH=arm GOARM=7 go build -o weatherdash

# Deploy binary
scp weatherdash weatherdash@192.168.86.13:/home/weatherdash/weatherdash/

# Deploy template (if changed)
scp templates/index.html weatherdash@192.168.86.13:/home/weatherdash/weatherdash/templates/

# Deploy static files (if changed)
scp -r static/js weatherdash@192.168.86.13:/home/weatherdash/weatherdash/static/

# Restart service
ssh weatherdash@192.168.86.13 << 'EOF'
chmod +x /home/weatherdash/weatherdash/weatherdash
sudo systemctl restart weatherdash
sudo systemctl status weatherdash --no-pager
EOF
```

### What to Deploy After Changes

| Change Type | Files to Deploy | Restart Required |
|-------------|----------------|------------------|
| Code logic | `weatherdash` binary | Yes |
| Template UI | `templates/index.html` | Yes (templates parsed at startup) |
| Static files | `static/js/*.js` | No (but clear browser cache) |
| Configuration | `config.yaml` | Yes |
| nginx config | `/etc/nginx/nginx.conf` | nginx reload only |

## Common Deployment Issues

### 1. Static Files Return 404

**Symptom:** Browser console shows 404 for `/static/js/core.js`

**Cause:** Hub gateway doubles the `/static` path

**Fix:** In `MyHomeServicesHub/main.go`:

```go
// Correct - proxy strips path prefix
http.HandleFunc("/static/", proxyWithAuth("http://localhost:8081/"))

// Wrong - doubles /static in URL
http.HandleFunc("/static/", proxyWithAuth("http://localhost:8081/static/"))
```

### 2. Permission Denied (203/EXEC)

**Symptom:** Service fails with status 203/EXEC

**Fix:**
```bash
chmod +x /home/weatherdash/weatherdash/weatherdash
sudo systemctl restart weatherdash
```

### 3. Templates Not Updating

**Symptom:** UI changes don't appear after deployment

**Cause:** Templates are parsed at startup, not on each request

**Fix:**
```bash
# After deploying new template
sudo systemctl restart weatherdash
```

### 4. Auth Headers Not Received

**Symptom:** Guest users see Easy Button (should be hidden)

**Check hub gateway logs:**
```bash
sudo journalctl -u hub-gateway -f | grep "Proxying to"
# Should show: User: guest, Role: guest
```

**Check weatherdash logs:**
```bash
sudo journalctl -u weatherdash -f | grep "Auth headers"
# Should show: Auth headers - User: guest, Role: guest
```

**Common causes:**
- Hub gateway not restarted after code changes
- nginx routing directly to :8081 instead of through :8080
- Session cookie not being sent (check Secure flag for HTTPS)

### 5. SSE Connection Failures

**Symptom:** Data doesn't load, charts empty

**Check browser console:**
- EventSource connection errors
- 404 on /api/weather

**Check nginx SSE settings:**
```nginx
location /api/ {
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_buffering off;
    proxy_cache off;
    chunked_transfer_encoding off;
    proxy_read_timeout 3600s;
}
```

## Monitoring and Logs

### Service Status

```bash
# Check if running
sudo systemctl status weatherdash

# View recent logs
sudo journalctl -u weatherdash -n 100 --no-pager

# Follow logs in real-time
sudo journalctl -u weatherdash -f

# Filter for errors
sudo journalctl -u weatherdash | grep -i error

# Check SSE activity
sudo journalctl -u weatherdash | grep SSE | tail -20
```

### Health Check

```bash
# Service endpoint
curl http://localhost:8081/health
# Should return: OK

# Through hub
curl -H "X-Hub-Role: admin" http://localhost:8080/weather
```

### Database Connectivity

```bash
# From weatherdash logs
sudo journalctl -u weatherdash | grep -i "database\|mysql\|maria"

# Direct database test
mysql -u weewx -p weewx -e "SELECT COUNT(*) FROM archive;"
```

## Related Documentation

- [MyHomeServicesHub Deployment](../MyHomeServicesHub/DEPLOYMENT.md) - Hub gateway setup and authentication
- [WeeWX Documentation](https://weewx.com/docs.html) - Weather station software
- [nginx SSE Configuration](https://www.nginx.com/blog/nginx-nodejs-websockets-socketio/) - Server-Sent Events setup
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
