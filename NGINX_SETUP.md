# Nginx Setup Guide for MyWeatherDash Platform (Raspberry Pi / Linux)

This guide is optimized for Raspberry Pi and Linux deployments.

> **Note:** This guide currently focuses on Raspberry Pi/Linux setup. Windows and other OS deployment guides can be added in the future as needed.

## Linux Installation (Raspberry Pi)

### 1. Install nginx
```bash
# Update package list
sudo apt update

# Install nginx
sudo apt install nginx -y

# Verify installation
nginx -v
```

### 2. Copy Configuration
```bash
# Backup original config
sudo cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup

# Copy the MyWeatherDash nginx.conf
sudo cp nginx.conf /etc/nginx/nginx.conf

# Test configuration
sudo nginx -t
```

### 3. Update MyWeatherDash Port
The config.yaml file should already be set to port 8080:

```yaml
server:
  port: 8080  # HTTP server port (8080 for nginx reverse proxy, 80 for direct access)
```

If not, edit `config.yaml` and change the port to 8080.

### 4. Build MyWeatherDash for Raspberry Pi
```bash
cd ~/MyWeatherDash  # or wherever you cloned the repo

# Build the binary
go build -o weatherdash

# Test it runs
./weatherdash
# Press Ctrl+C to stop
```

### 5. Start Services

**Start nginx:**
```bash
sudo systemctl start nginx
sudo systemctl enable nginx  # Auto-start on boot
```

**Check nginx status:**
```bash
sudo systemctl status nginx
```

### 6. Test
- Main page: http://localhost/ or http://raspberrypi.local/
- Weather dashboard: http://localhost/weather
- API: http://localhost/api/weather
- From another device: http://[raspberry-pi-ip-address]/

---

## Running MyWeatherDash as a systemd Service (Auto-start)

Create a systemd service file to run MyWeatherDash automatically on boot.

### 1. Create Service File
```bash
sudo nano /etc/systemd/system/weatherdash.service
```

### 2. Add Service Configuration
```ini
[Unit]
Description=MyWeatherDash Weather Dashboard
After=network.target mariadb.service
Wants=mariadb.service

[Service]
Type=simple
User=pi
# Update paths to match your installation:
WorkingDirectory=/home/pi/MyWeatherDash
ExecStart=/home/pi/MyWeatherDash/weatherdash
Restart=always
RestartSec=10

# Environment variables (if needed)
# Environment="DB_HOST=localhost"

[Install]
WantedBy=multi-user.target
```

### 3. Enable and Start Service
```bash
# Reload systemd to recognize new service
sudo systemctl daemon-reload

# Enable auto-start on boot
sudo systemctl enable weatherdash

# Start the service now
sudo systemctl start weatherdash

# Check status
sudo systemctl status weatherdash
```

### 4. Useful Service Commands
```bash
# View logs
sudo journalctl -u weatherdash -f

# Restart service
sudo systemctl restart weatherdash

# Stop service
sudo systemctl stop weatherdash

# Disable auto-start
sudo systemctl disable weatherdash
```

---

## Raspberry Pi-Specific Notes

### Performance Optimization
```bash
# Adjust nginx worker processes for Pi (usually 1-2)
# Edit /etc/nginx/nginx.conf:
worker_processes 2;
```

### Memory Considerations
- Raspberry Pi 3/4: Can easily handle nginx + MyWeatherDash + MariaDB
- Raspberry Pi Zero: May need to reduce nginx worker_processes to 1

### Firewall (if enabled)
```bash
# Allow HTTP traffic
sudo ufw allow 80/tcp

# Check status
sudo ufw status
```

### Finding Your Pi's IP Address
```bash
# Show all network interfaces
ip addr show

# Or simpler:
hostname -I
```

---

## Alternative: Linux Installation (Other Distributions)

### CentOS/RHEL
```bash
sudo yum install nginx -y
sudo systemctl start nginx
sudo systemctl enable nginx
```

### Arch Linux
```bash
sudo pacman -S nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

Configuration file: `/etc/nginx/nginx.conf`

---

## Deployment Structure

After setup, your Raspberry Pi architecture will be:

```
Browser → http://raspberrypi (nginx :80)
           ├─ / → Go server :8080 (landing page)
           ├─ /weather → Go server :8080 (weather dashboard)
           ├─ /api/ → Go server :8080 (weather APIs)
           ├─ /static/ → Go server :8080 (assets)
           ├─ /network → Future SNMP monitor :8081
           └─ /api/network/ → Future SNMP APIs :8081
```

## Adding New Projects

To add a new project (e.g., SNMP monitor):

### 1. Create New Service
Build your new Go application listening on a different port (e.g., 8081)

### 2. Create systemd Service
```bash
sudo nano /etc/systemd/system/snmpmonitor.service
```

```ini
[Unit]
Description=SNMP Network Monitor
After=network.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi/SNMPMonitor
ExecStart=/home/pi/SNMPMonitor/snmpmonitor
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### 3. Add nginx Location Block
Edit `/etc/nginx/nginx.conf` and add:
```nginx
location /network {
    proxy_pass http://localhost:8081;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

### 4. Enable and Start Services
```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable and start new service
sudo systemctl enable snmpmonitor
sudo systemctl start snmpmonitor

# Reload nginx configuration
sudo nginx -t
sudo systemctl reload nginx
```

---

## Troubleshooting

### Port 80 already in use
```bash
# Find process using port 80
sudo lsof -i :80
# or
sudo netstat -tulpn | grep :80

# Stop the conflicting service (e.g., Apache)
sudo systemctl stop apache2
sudo systemctl disable apache2
```

### nginx won't start
```bash
# Test configuration syntax
sudo nginx -t

# Check error log
sudo tail -f /var/log/nginx/error.log

# Check if nginx process is running
ps aux | grep nginx
```

### Can't access from other devices on network
```bash
# Check firewall status
sudo ufw status

# Allow HTTP traffic
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Check nginx is listening on all interfaces (0.0.0.0)
sudo netstat -tulpn | grep nginx
```

### MyWeatherDash service won't start
```bash
# Check service logs
sudo journalctl -u weatherdash -n 50

# Check if binary is executable
chmod +x /home/pi/MyWeatherDash/weatherdash

# Verify config.yaml exists and is readable
ls -la /home/pi/MyWeatherDash/config.yaml
```

### MariaDB connection issues
```bash
# Check MariaDB is running
sudo systemctl status mariadb

# Test database connection
mysql -u weewx -p -h localhost weewx

# Check config.yaml has correct database credentials
```

---

## SSL/HTTPS Setup with Let's Encrypt (Production)

### 1. Install Certbot
```bash
sudo apt install certbot python3-certbot-nginx -y
```

### 2. Get SSL Certificate
```bash
# Replace with your domain name
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

### 3. Auto-renewal
```bash
# Certbot auto-renewal is configured automatically
# Test renewal process:
sudo certbot renew --dry-run
```

### Manual SSL Setup (Alternative)
If using your own certificates:

1. Uncomment HTTPS section in `/etc/nginx/nginx.conf`
2. Update certificate paths:
```nginx
ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
```
3. Add HTTP→HTTPS redirect:
```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}
```
4. Test and reload:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

---

## Performance Tuning for Raspberry Pi

### Basic Optimization
Edit `/etc/nginx/nginx.conf`:

```nginx
# Raspberry Pi 3/4: Use 2-4 workers
# Raspberry Pi Zero: Use 1 worker
worker_processes 2;

# Adjust based on RAM
worker_connections 512;

# Enable for HTTP/2 (if using HTTPS)
# listen 443 ssl http2;
```

### Proxy Buffering
Already configured in the provided `nginx.conf`:
```nginx
proxy_buffering on;
proxy_buffer_size 4k;
proxy_buffers 8 4k;
```

### Static Content Caching
The `/static/` location already has caching enabled:
```nginx
location /static/ {
    proxy_pass http://localhost:8080;
    proxy_cache_valid 200 1h;
    expires 1h;
    add_header Cache-Control "public, immutable";
}
```

### Monitor Performance
```bash
# Watch nginx processes
htop

# Check memory usage
free -h

# Monitor logs for errors
sudo tail -f /var/log/nginx/error.log
sudo journalctl -u weatherdash -f
```

---

## Complete Setup Checklist

- [ ] Install nginx: `sudo apt install nginx`
- [ ] Copy nginx.conf to `/etc/nginx/nginx.conf`
- [ ] Test nginx config: `sudo nginx -t`
- [ ] Update config.yaml port to 8080
- [ ] Build MyWeatherDash: `go build -o weatherdash`
- [ ] Create weatherdash.service in `/etc/systemd/system/`
- [ ] Enable services: `sudo systemctl enable nginx weatherdash`
- [ ] Start services: `sudo systemctl start nginx weatherdash`
- [ ] Test locally: `curl http://localhost/weather`
- [ ] Find Pi IP: `hostname -I`
- [ ] Test from network: `http://[pi-ip-address]/weather`
- [ ] Configure firewall: `sudo ufw allow 80/tcp`
- [ ] (Optional) Set up SSL with certbot
- [ ] (Optional) Configure static IP for Pi
- [ ] Monitor logs: `sudo journalctl -u weatherdash -f`

---

## Next Steps

1. **Create a landing page** at `/` to navigate between modules (Weather, Network Monitor, etc.)
2. **Build SNMP monitor** service on port 8081
3. **Set up monitoring alerts** (email, webhook, etc.)
4. **Configure automatic database backups**
5. **Add authentication** layer for sensitive data
6. **Set up reverse DNS** or dynamic DNS for external access

Your Raspberry Pi is now a powerful multi-service monitoring platform! 🎉
