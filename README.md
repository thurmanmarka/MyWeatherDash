# 🌩️ MyWeatherDash

A modern, feature-rich weather dashboard powered by **Go**, **WeeWX**, **MariaDB**, and **Chart.js**

MyWeatherDash is a self-hosted personal weather dashboard that transforms WeeWX weather station data into beautiful, interactive visualizations with real-time updates and comprehensive statistics.

## ✨ Features

### 📊 Interactive Charts (Chart.js)
- **12 synchronized charts** with shared time axis and day/night shading
- Temperature, dewpoint, feels-like (heat index/wind chill)
- Barometric pressure with trend indicators
- Humidity (inside & outside)
- Wind speed, gusts, direction, and **vector charts** (WeeWX-style)
- Rain amount and rate
- Lightning strikes
- Inside temperature & humidity
- **Range selection**: Day, Week, or Month views

### 🌡️ Live Current Conditions
- Real-time temperature, dewpoint, humidity
- Active feels-like calculation (heat index or wind chill)
- Barometric pressure with trend forecast
- Wind speed, gust, and direction
- Rain totals and rate
- Lightning strike detection
- Inside conditions monitoring
- Alert indicators for severe conditions

### 📈 Comprehensive Statistics Panel
- **High/Low tracking** for all measurements
- Daily and range-based statistics
- Wind metrics: max, average, RMS, vector average, and direction
- Rain accumulation and rate tracking
- Lightning strike counts and distance
- Organized by category (temperature, precipitation, wind, indoor)

### 📄 NOAA Climatological Reports
- **Monthly and yearly summaries** in standard NOAA format
- Automatic report generation from historical data
- Download as text files
- Force recompile option for updated data
- Configurable location metadata (name, coordinates, elevation)

### ⚡ Real-Time Updates
- **Server-Sent Events (SSE)** for live data streaming
- Configurable polling intervals
- Automatic chart refresh
- No page reload required

### 🎨 Modern UI/UX
- Clean, responsive design
- Color-coded data visualization
- Intuitive iconography
- Mobile-friendly layout
- Professional styling with smooth animations

## 🛠️ Project Structure

```
MyWeatherDash/
├── main.go              # Server entry point
├── handlers.go          # API endpoint handlers
├── types.go             # Data structures
├── config.go            # Configuration loader
├── noaa.go              # NOAA report generator
├── sse.go               # Server-Sent Events broker
├── config.yaml          # Your configuration (gitignored)
├── config.example.yaml  # Configuration template
├── templates/
│   └── index.html       # Dashboard UI
├── static/
│   ├── js/
│   │   ├── core.js      # Chart.js plugins & utilities
│   │   ├── charts.js    # Chart rendering logic
│   │   └── noaa.js      # NOAA reports interface
│   └── noaa/            # Generated NOAA reports
└── README.md
```

## 🔧 Setup

### 1. Prerequisites
- Go 1.19+ installed
- WeeWX weather station with MariaDB/MySQL archive
- Network access to WeeWX database

### 2. Clone the Repository
```bash
git clone https://github.com/thurmanmarka/MyWeatherDash.git
cd MyWeatherDash
```

### 3. Configure Database Connection
Copy the example configuration:
```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` with your settings:
```yaml
# Database configuration
db:
  user: your_db_user
  password: "your_password"
  host: "192.168.1.100"
  port: 3306
  name: "weewx"
  params: "parseTime=false"

# Server configuration
server:
  port: 8080                    # HTTP server port
  sse_poll_seconds: 60          # Server polling interval
  client_poll_seconds: 0        # Client polling (0 = disabled)

# Location configuration
location:
  name: "My Weather Station"
  latitude: 32.0
  longitude: -110.0
  altitude: 2800                # feet above sea level
```

⚠️ **Note**: `config.yaml` is gitignored to protect your credentials.

### 4. Install Dependencies
```bash
go mod download
```

### 5. Run the Server
Development mode:
```bash
go run .
```

Or build and run:
```bash
go build -o weatherdash
./weatherdash
```

The dashboard will be available at `http://localhost:8080` (or your configured port).

## 🧪 API Endpoints

All data endpoints support time range queries:
- `?range=day` - Last 24 hours (default)
- `?range=week` - Last 7 days
- `?range=month` - Last 30 days

### Core Endpoints
- `GET /` - Dashboard UI
- `GET /api/ping` - Health check
- `GET /api/weather` - Temperature & dewpoint
- `GET /api/barometer` - Barometric pressure
- `GET /api/feelslike` - Heat index & wind chill
- `GET /api/humidity` - Outside humidity
- `GET /api/wind` - Wind speed, gust, direction
- `GET /api/rain` - Rain rate & amount
- `GET /api/lightning` - Lightning strikes
- `GET /api/insideTemp` - Inside temperature
- `GET /api/insideHumidity` - Inside humidity
- `GET /api/statistics` - Comprehensive statistics
- `GET /api/stream` - SSE live updates

### NOAA Reports
- `GET /api/noaa/monthly?year=2025&month=11` - Monthly summary
- `GET /api/noaa/yearly?year=2025` - Yearly summary

## ⚙️ Configuration Options

### Database (`db`)
- `user`, `password`, `host`, `port`, `name` - Database connection settings
- `params` - Additional MySQL parameters

### Server (`server`)
- `port` - HTTP server port (default: 8080)
- `sse_poll_seconds` - How often server checks for new data (default: 60)
- `client_poll_seconds` - Client-side polling interval, 0 to disable (default: 0)

### Location (`location`)
- `name` - Station name (shown in page header and NOAA reports)
- `latitude` - Decimal degrees
- `longitude` - Decimal degrees
- `altitude` - Elevation in feet above sea level

## 🎯 Key Features Explained

### Wind Vector Chart
Plots wind vectors using WeeWX methodology:
- Converts wind direction to vector components
- Applies -90° rotation for proper orientation
- Auto-scaling Y-axis based on wind speed
- Individual vector lines for each data point

### Day/Night Shading
- Automatically calculates sunrise/sunset times
- Shades night periods across all charts
- Synchronized across all visualizations

### Statistics Panel
- Real-time calculation of highs, lows, and averages
- Separate tracking for "Today" vs selected range
- Split display for temperature ranges (high/low on separate rows)
- Wind direction for max gust events

### NOAA Reports
- Standard NOAA climatological format
- Daily summaries for monthly reports
- Monthly summaries for yearly reports
- Automatic file generation and caching
- Force recompile option for data updates

## 🏗️ Future Enhancements
- [ ] Dark mode toggle
- [ ] Systemd service configuration
- [ ] Docker containerization
- [ ] Weather alerts and notifications
- [ ] Export data to CSV
- [ ] Historical data comparisons
- [ ] Customizable chart colors
- [ ] Multi-language support

## 📜 License
This project is for personal use. Feel free to fork and modify for your own weather station.

## 🙌 Contributing
Contributions, issues, and feature requests are welcome! Feel free to open an issue or submit a pull request.

## 🔗 Related Projects
- [WeeWX](https://weewx.com/) - Weather station software
- [Chart.js](https://www.chartjs.org/) - JavaScript charting library

---

**Built with ❤️ for weather enthusiasts**
