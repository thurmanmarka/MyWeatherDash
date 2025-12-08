package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sync"

	"github.com/thurmanmarka/astroglide"
	"golang.org/x/sync/singleflight"
)

// -------------------- auth helpers --------------------

// getUserRole extracts the user role from hub auth headers
func getUserRole(r *http.Request) string {
	role := r.Header.Get("X-Hub-Role")
	user := r.Header.Get("X-Hub-User")
	auth := r.Header.Get("X-Hub-Authenticated")
	log.Printf("Auth headers - User: %s, Role: %s, Authenticated: %s", user, role, auth)
	if role == "" {
		role = "admin" // Default to admin if no header (standalone mode)
	}
	return role
}

// isAdmin checks if the user has admin role
func isAdmin(r *http.Request) bool {
	return getUserRole(r) == "admin"
}

// requireAdmin is middleware that blocks non-admin users
func requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAdmin(r) {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// -------------------- cached celestial --------------------
type cachedCelestial struct {
	data   CelestialData
	expiry time.Time
}

// celestialCache stores cached celestial results keyed by date+timezone
var celestialCache = struct {
	sync.RWMutex
	m map[string]*cachedCelestial
}{m: make(map[string]*cachedCelestial)}

// celestialGroup deduplicates concurrent compute requests for the same date
var celestialGroup singleflight.Group

// -------------------- helpers --------------------

func handleLanding(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>MyWeatherDash Platform</title>
    <style>
        :root {
            --bg: #e7edf4;
            --page-bg: #f5f7fb;
            --panel-bg: #ffffff;
            --panel-border: #d2d7e0;
            --accent: #2563eb;
            --accent-hover: #1d4ed8;
            --text-main: #1f2933;
            --text-muted: #6b7280;
        }

        [data-theme="dark"] {
            --bg: #0f172a;
            --page-bg: #1e293b;
            --panel-bg: #334155;
            --panel-border: #475569;
            --accent: #3b82f6;
            --accent-hover: #2563eb;
            --text-main: #f1f5f9;
            --text-muted: #94a3b8;
        }

        * {
            box-sizing: border-box;
        }

        body {
            font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
            margin: 0;
            padding: 20px;
            background: radial-gradient(circle at top left, var(--bg), var(--page-bg) 70%);
            color: var(--text-main);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .container {
            max-width: 800px;
            width: 100%;
            background: var(--panel-bg);
            border-radius: 16px;
            padding: 48px;
            box-shadow: 0 8px 24px rgba(15, 23, 42, 0.1);
            border: 1px solid var(--panel-border);
        }

        h1 {
            margin: 0 0 12px 0;
            font-size: 2.5rem;
            font-weight: 700;
            color: var(--text-main);
            text-align: center;
        }

        .subtitle {
            text-align: center;
            color: var(--text-muted);
            font-size: 1.1rem;
            margin-bottom: 48px;
        }

        .modules {
            display: grid;
            gap: 20px;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
        }

        .module-card {
            background: var(--page-bg);
            border: 2px solid var(--panel-border);
            border-radius: 12px;
            padding: 32px;
            text-decoration: none;
            color: var(--text-main);
            transition: all 0.2s ease;
            display: flex;
            flex-direction: column;
            align-items: center;
            text-align: center;
        }

        .module-card:hover {
            border-color: var(--accent);
            transform: translateY(-4px);
            box-shadow: 0 12px 24px rgba(37, 99, 235, 0.15);
        }

        .module-icon {
            font-size: 3rem;
            margin-bottom: 16px;
        }

        .module-title {
            font-size: 1.5rem;
            font-weight: 600;
            margin-bottom: 8px;
        }

        .module-description {
            color: var(--text-muted);
            font-size: 0.95rem;
        }

        .module-card.disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }

        .module-card.disabled:hover {
            transform: none;
            border-color: var(--panel-border);
            box-shadow: none;
        }

        .coming-soon {
            display: inline-block;
            background: var(--accent);
            color: white;
            font-size: 0.7rem;
            padding: 2px 8px;
            border-radius: 999px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-top: 8px;
        }

        .footer {
            margin-top: 48px;
            text-align: center;
            color: var(--text-muted);
            font-size: 0.9rem;
        }

        @media (max-width: 600px) {
            .container {
                padding: 32px 24px;
            }
            h1 {
                font-size: 2rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🏠 Home Services Hub</h1>
        <p class="subtitle">Your centralized monitoring and management platform</p>
        
        <div class="modules">
            <a href="/weather" class="module-card">
                <div class="module-icon">☀️</div>
                <div class="module-title">Weather Dashboard</div>
                <div class="module-description">Real-time weather data, charts, and statistics</div>
            </a>
            
            <div class="module-card disabled">
                <div class="module-icon">🌐</div>
                <div class="module-title">Network Monitor</div>
                <div class="module-description">SNMP monitoring for network devices</div>
                <span class="coming-soon">Coming Soon</span>
            </div>
        </div>

        <div class="footer">
            Home Services Hub v2.0.0 | Powered by Go & nginx
        </div>
    </div>
</body>
</html>`
	w.Write([]byte(tmpl))
}

func handleWeatherDash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	// Provide client-side polling interval (ms) to the template
	data := struct {
		ClientPollMs int
		AssetVersion string
		LocationName string
		ExtremeHeat  float64
		ExtremeCold  float64
		IsAdmin      bool
		UserRole     string
	}{
		ClientPollMs: appConfig.Server.ClientPollSeconds * 1000,
		AssetVersion: time.Now().Format("20060102T150405"),
		LocationName: appConfig.Location.Name,
		ExtremeHeat:  appConfig.Alerts.ExtremeHeat,
		ExtremeCold:  appConfig.Alerts.ExtremeCold,
		IsAdmin:      isAdmin(r),
		UserRole:     getUserRole(r),
	}

	if err := tmplIndex.Execute(w, data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(PingResponse{Message: "pong"})
}

// handleHealth returns 200 OK for health checks (systemd, monitoring, load balancers)
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Optional: check DB connection, add more sophisticated health checks
	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": "database unreachable"})
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Range helper: day / week / month
func getRangeDuration(r *http.Request) time.Duration {
	q := strings.ToLower(r.URL.Query().Get("range"))
	switch q {
	case "week":
		return 7 * 24 * time.Hour
	case "month":
		// simple 30-day month
		return 30 * 24 * time.Hour
	default:
		// "day" or anything else
		return 24 * time.Hour
	}
}

// -------------------- /api/barometer --------------------

func handleBarometer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dur := getRangeDuration(r)
	since := time.Now().Add(-dur).Unix()

	rows, err := db.Query(`
		SELECT dateTime, barometer
		FROM archive
		WHERE dateTime >= ?
		ORDER BY dateTime ASC
	`, since)
	if err != nil {
		log.Println("DB query error (barometer):", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var readings []BarometerReading
	for rows.Next() {
		var epochSec int64
		var pressure float64

		if err := rows.Scan(&epochSec, &pressure); err != nil {
			log.Println("DB scan error (barometer):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		ts := time.Unix(epochSec, 0)

		readings = append(readings, BarometerReading{
			Timestamp: ts,
			Pressure:  pressure,
		})
	}
	if err := rows.Err(); err != nil {
		log.Println("DB rows error (barometer):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
	}

	// Calculate trend, level, and forecast for the latest reading
	if len(readings) > 4 {
		latest := &readings[len(readings)-1]
		current := latest.Pressure
		prior := readings[len(readings)-5].Pressure
		change := current - prior

		// Determine pressure level
		if current > 30.20 {
			latest.Level = "high"
		} else if current < 29.80 {
			latest.Level = "low"
		} else {
			latest.Level = "normal"
		}

		// Determine rate of change (per hour)
		// Assuming 5 intervals in data; typical WeeWX is 5-min intervals = 12 per hour
		// So 5 intervals = ~25 minutes. Change per hour ≈ change * (60/25) = change * 2.4
		changePerHour := math.Abs(change) * 2.4

		// Categorize trend based on rate
		if change > 0 {
			if changePerHour >= 0.06 {
				latest.Trend = "rapid-rise"
			} else if changePerHour >= 0.02 {
				latest.Trend = "slow-rise"
			} else {
				latest.Trend = "steady"
			}
		} else if change < 0 {
			if changePerHour >= 0.06 {
				latest.Trend = "rapid-fall"
			} else if changePerHour >= 0.02 {
				latest.Trend = "slow-fall"
			} else {
				latest.Trend = "steady"
			}
		} else {
			latest.Trend = "steady"
		}

		// Generate forecast based on level + trend
		switch latest.Level {
		case "high":
			switch latest.Trend {
			case "steady", "slow-rise":
				latest.Forecast = "Fair weather"
			case "rapid-rise":
				latest.Forecast = "Fair, improving"
			case "slow-fall":
				latest.Forecast = "Cloudy later"
			case "rapid-fall":
				latest.Forecast = "Warmer, cloudier"
			}
		case "normal":
			switch latest.Trend {
			case "steady", "slow-rise":
				latest.Forecast = "Conditions continue"
			case "rapid-rise":
				latest.Forecast = "Improving"
			case "slow-fall":
				latest.Forecast = "Minor changes"
			case "rapid-fall":
				latest.Forecast = "Rain/snow likely"
			}
		case "low":
			switch latest.Trend {
			case "steady", "slow-rise":
				latest.Forecast = "Cooler, clearing"
			case "rapid-rise":
				latest.Forecast = "Improving quickly"
			case "slow-fall":
				latest.Forecast = "Rain coming"
			case "rapid-fall":
				latest.Forecast = "Stormy weather"
			}
		}
	}

	if err := json.NewEncoder(w).Encode(readings); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
}

// -------------------- /api/weather --------------------

func handleWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dur := getRangeDuration(r)
	since := time.Now().Add(-dur).Unix()

	rows, err := db.Query(`
		SELECT dateTime, outTemp, dewpoint
		FROM archive
		WHERE dateTime >= ?
		ORDER BY dateTime ASC
	`, since)
	if err != nil {
		log.Println("DB query error (weather):", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var readings []WeatherReading
	for rows.Next() {
		var epochSec int64
		var tempF, dewF float64

		if err := rows.Scan(&epochSec, &tempF, &dewF); err != nil {
			log.Println("DB scan error (weather):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		ts := time.Unix(epochSec, 0)

		readings = append(readings, WeatherReading{
			Timestamp:   ts,
			Temperature: tempF,
			Dewpoint:    dewF,
		})
	}
	if err := rows.Err(); err != nil {
		log.Println("DB rows error (weather):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(readings); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
}

// -------------------- /api/feelslike --------------------

func handleFeelsLike(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dur := getRangeDuration(r)
	since := time.Now().Add(-dur).Unix()

	rows, err := db.Query(`
		SELECT dateTime, heatindex, windchill, outTemp
		FROM archive
		WHERE dateTime >= ?
		ORDER BY dateTime ASC
	`, since)
	if err != nil {
		log.Println("DB query error (feelslike):", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var readings []FeelsLikeReading
	for rows.Next() {
		var epochSec int64
		var heatF, chillF, tempF float64

		if err := rows.Scan(&epochSec, &heatF, &chillF, &tempF); err != nil {
			log.Println("DB scan error (feelslike):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		ts := time.Unix(epochSec, 0)

		reading := FeelsLikeReading{
			Timestamp: ts,
			HeatIndex: heatF,
			WindChill: chillF,
		}

		// For the latest reading, compute active feels-like value
		readings = append(readings, reading)
	}
	if err := rows.Err(); err != nil {
		log.Println("DB rows error (feelslike):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
	}

	// Compute active feels-like for the latest reading
	if len(readings) > 0 {
		latest := &readings[len(readings)-1]

		// Need to get the temperature for the latest record
		var tempF float64
		err := db.QueryRow(`
			SELECT outTemp FROM archive 
			WHERE dateTime = ?
		`, latest.Timestamp.Unix()).Scan(&tempF)

		if err == nil {
			active := pickFeelsLikeSource(tempF, latest.HeatIndex, latest.WindChill)
			latest.ActiveValue = active.value
			latest.ActiveSource = active.source
			latest.ActiveLabel = active.label
		}
	}

	if err := json.NewEncoder(w).Encode(readings); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
}

// Helper to pick the active feels-like source
type feelsLikeSource struct {
	value  float64
	source string
	label  string
}

func pickFeelsLikeSource(tempF, heatIndexF, windChillF float64) feelsLikeSource {
	const HEAT_THRESHOLD_F = 80.0
	const CHILL_THRESHOLD_F = 50.0

	activeValue := tempF
	sourceLabel := "Air Temp"
	sourceKey := "air"

	if !math.IsNaN(heatIndexF) && heatIndexF != 0 && tempF >= HEAT_THRESHOLD_F {
		activeValue = heatIndexF
		sourceLabel = "Heat Index"
		sourceKey = "heat"
	} else if !math.IsNaN(windChillF) && windChillF != 0 && tempF <= CHILL_THRESHOLD_F {
		activeValue = windChillF
		sourceLabel = "Wind Chill"
		sourceKey = "chill"
	}

	return feelsLikeSource{
		value:  activeValue,
		source: sourceKey,
		label:  sourceLabel,
	}
}

// -------------------- /api/humidity --------------------

func handleHumidity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dur := getRangeDuration(r)
	since := time.Now().Add(-dur).Unix()

	rows, err := db.Query(`
   	    SELECT dateTime, outHumidity
   	    FROM archive
   	    WHERE dateTime >= ?
   	    ORDER BY dateTime ASC
   	`, since)
	if err != nil {
		log.Println("DB query error (humidity):", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var readings []HumidityReading
	for rows.Next() {
		var epochSec int64
		var hum float64

		if err := rows.Scan(&epochSec, &hum); err != nil {
			log.Println("DB scan error (humidity):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		ts := time.Unix(epochSec, 0)

		readings = append(readings, HumidityReading{
			Timestamp: ts,
			Humidity:  hum,
		})
	}

	if err := rows.Err(); err != nil {
		log.Println("DB rows error (humidity):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(readings)
}

// -------------------- /api/wind --------------------

func handleWind(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dur := getRangeDuration(r)
	since := time.Now().Add(-dur).Unix()

	rows, err := db.Query(`
		SELECT dateTime, windSpeed, windGust, windDir
		FROM archive
		WHERE dateTime >= ?
		ORDER BY dateTime ASC
	`, since)
	if err != nil {
		log.Println("DB query error (wind):", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var readings []WindReading
	for rows.Next() {
		var epochSec int64
		var speed, gust float64
		var dir sql.NullFloat64

		if err := rows.Scan(&epochSec, &speed, &gust, &dir); err != nil {
			log.Println("DB scan error (wind):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		ts := time.Unix(epochSec, 0)

		var dirPtr *float64
		if dir.Valid {
			dirCopy := dir.Float64
			dirPtr = &dirCopy
		}

		readings = append(readings, WindReading{
			Timestamp: ts,
			Speed:     speed,
			Gust:      gust,
			Direction: dirPtr,
		})
	}
	if err := rows.Err(); err != nil {
		log.Println("DB rows error (wind):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
	}

	// Compute compass direction and strong flag for latest reading
	if len(readings) > 0 {
		latest := &readings[len(readings)-1]

		// Compass direction
		if latest.Direction != nil {
			latest.Compass = degreesToCompass(*latest.Direction)
		} else {
			latest.Compass = "--"
		}

		// Strong wind detection using config thresholds
		latest.Strong = (latest.Speed >= appConfig.Alerts.WindSpeed || latest.Gust >= appConfig.Alerts.WindGust)
	}

	if err := json.NewEncoder(w).Encode(readings); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
}

// Helper to convert wind direction degrees to 16-point compass
func degreesToCompass(deg float64) string {
	if math.IsNaN(deg) {
		return "--"
	}

	directions := []string{
		"N", "NNE", "NE", "ENE",
		"E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW",
		"W", "WNW", "NW", "NNW",
	}

	index := int(math.Round(math.Mod(deg, 360.0)/22.5)) % 16
	return directions[index]
}

// -------------------- /api/rain --------------------

func handleRain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dur := getRangeDuration(r)
	since := time.Now().Add(-dur).Unix()

	rows, err := db.Query(`
		SELECT dateTime, rainRate, rain
		FROM archive
		WHERE dateTime >= ?
		ORDER BY dateTime ASC
	`, since)
	if err != nil {
		log.Println("DB query error (rain):", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var readings []RainReading
	for rows.Next() {
		var epochSec int64
		var rate, amount float64

		if err := rows.Scan(&epochSec, &rate, &amount); err != nil {
			log.Println("DB scan error (rain):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		ts := time.Unix(epochSec, 0)

		readings = append(readings, RainReading{
			Timestamp: ts,
			Rate:      rate,
			Amount:    amount,
		})
	}
	if err := rows.Err(); err != nil {
		log.Println("DB rows error (rain):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
	}

	// Compute recently-active flag for latest reading
	if len(readings) > 0 {
		now := time.Now()
		tenMinutesAgo := now.Add(-10 * time.Minute)
		recentlyActive := false

		// Scan backwards through readings looking for rain in last 10 minutes
		for i := len(readings) - 1; i >= 0; i-- {
			if readings[i].Timestamp.Before(tenMinutesAgo) {
				break
			}
			if readings[i].Rate > 0 || readings[i].Amount > 0 {
				recentlyActive = true
				break
			}
		}

		readings[len(readings)-1].RecentlyActive = recentlyActive
	}

	if err := json.NewEncoder(w).Encode(readings); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
}

// -------------------- /api/lightning --------------------

func handleLightning(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dur := getRangeDuration(r)
	since := time.Now().Add(-dur).Unix()

	rows, err := db.Query(`
        SELECT dateTime, lightning_strike_count, lightning_distance
        FROM archive
        WHERE dateTime >= ? AND lightning_strike_count IS NOT NULL
        ORDER BY dateTime ASC
    `, since)
	if err != nil {
		log.Println("DB query error (lightning):", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var readings []LightningReading
	for rows.Next() {
		var epochSec int64
		var strikes float64
		var distance sql.NullFloat64

		if err := rows.Scan(&epochSec, &strikes, &distance); err != nil {
			log.Println("DB scan error (lightning):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		ts := time.Unix(epochSec, 0)

		var distPtr *float64
		if distance.Valid {
			distPtr = &distance.Float64
		}

		readings = append(readings, LightningReading{
			Timestamp: ts,
			Strikes:   strikes,
			Distance:  distPtr,
		})
	}

	if err := rows.Err(); err != nil {
		log.Println("DB rows error (lightning):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
	}

	// Compute recently-active flag for latest reading
	if len(readings) > 0 {
		now := time.Now()
		tenMinutesAgo := now.Add(-10 * time.Minute)
		recentlyActive := false

		// Scan backwards through readings looking for strikes in last 10 minutes
		for i := len(readings) - 1; i >= 0; i-- {
			if readings[i].Timestamp.Before(tenMinutesAgo) {
				break
			}
			if readings[i].Strikes > 0 {
				recentlyActive = true
				break
			}
		}

		readings[len(readings)-1].RecentlyActive = recentlyActive
	}

	if err := json.NewEncoder(w).Encode(readings); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
}

// -------------------- /api/insideTemp --------------------

func handleInsideTemp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dur := getRangeDuration(r)
	since := time.Now().Add(-dur).Unix()

	rows, err := db.Query(`
        SELECT dateTime, inTemp
        FROM archive
        WHERE dateTime >= ? AND inTemp IS NOT NULL
        ORDER BY dateTime ASC
    `, since)
	if err != nil {
		log.Println("DB query error (insideTemp):", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	readings := make([]InsideTemperature, 0, 1024)

	for rows.Next() {
		var epochSec int64
		var tempF float64

		if err := rows.Scan(&epochSec, &tempF); err != nil {
			log.Println("DB scan error (insideTemp):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		ts := time.Unix(epochSec, 0)

		readings = append(readings, InsideTemperature{
			Timestamp:   ts,
			InsideTempF: tempF,
		})
	}
	if err := rows.Err(); err != nil {
		log.Println("DB rows error (insideTemp):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(readings); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
		return
	}
}

// -------------------- /api/insideHumidity --------------------

func handleInsideHumidity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dur := getRangeDuration(r)
	since := time.Now().Add(-dur).Unix()

	rows, err := db.Query(`
        SELECT dateTime, inHumidity
        FROM archive
        WHERE dateTime >= ? AND inHumidity IS NOT NULL
        ORDER BY dateTime ASC
   	`, since)
	if err != nil {
		log.Println("DB query error (insideHumidity):", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var readings []InsideHumidityReading
	for rows.Next() {
		var epochSec int64
		var hum float64

		if err := rows.Scan(&epochSec, &hum); err != nil {
			log.Println("DB scan error (insideHumidity):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		ts := time.Unix(epochSec, 0)

		readings = append(readings, InsideHumidityReading{
			Timestamp:      ts,
			InsideHumidity: hum,
		})
	}

	if err := rows.Err(); err != nil {
		log.Println("DB rows error (insideHumidity):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(readings)
}

// -------------------- /api/noaa/monthly --------------------

func handleNOAAMonthly(w http.ResponseWriter, r *http.Request) {
	// Require admin role for NOAA reports
	if !isAdmin(r) {
		http.Error(w, "Forbidden: NOAA reports require admin access", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")
	forceStr := r.URL.Query().Get("force")
	force := forceStr == "1" || forceStr == "true"
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	if yearStr != "" {
		if v, err := strconv.Atoi(yearStr); err == nil {
			year = v
		}
	}
	if monthStr != "" {
		if v, err := strconv.Atoi(monthStr); err == nil {
			month = v
		}
	}
	content, err := GetOrGenerateMonthly(db, NOAAMonthlyParams{Year: year, Month: month}, force)
	if err != nil {
		log.Println("NOAA monthly error:", err)
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(content))
}

// -------------------- /api/noaa/yearly --------------------

func handleNOAAYearly(w http.ResponseWriter, r *http.Request) {
	// Require admin role for NOAA reports
	if !isAdmin(r) {
		http.Error(w, "Forbidden: NOAA reports require admin access", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	yearStr := r.URL.Query().Get("year")
	forceStr := r.URL.Query().Get("force")
	force := forceStr == "1" || forceStr == "true"
	year := time.Now().Year()
	if yearStr != "" {
		if v, err := strconv.Atoi(yearStr); err == nil {
			year = v
		}
	}
	content, err := GetOrGenerateYearly(db, NOAAYearlyParams{Year: year}, force)
	if err != nil {
		log.Println("NOAA yearly error:", err)
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(content))
}

// -------------------- /api/statistics --------------------

func handleStatistics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dur := getRangeDuration(r)
	since := time.Now().Add(-dur).Unix()

	// Get local midnight for "today" calculations
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	midnightUnix := midnight.Unix()

	// Single query to fetch all necessary data
	rows, err := db.Query(`
		SELECT dateTime, rain, rainRate, lightning_strike_count, lightning_distance,
		       outTemp, dewpoint, outHumidity, barometer,
		       heatindex, windchill, windSpeed, windGust, windDir,
		       inTemp, inHumidity
		FROM archive
		WHERE dateTime >= ?
		ORDER BY dateTime ASC
	`, since)
	if err != nil {
		log.Println("DB query error (statistics):", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Accumulators for range
	var rainRangeTotal float64
	var rainMidnightTotal float64
	var strikeRangeTotal int
	var strikeMidnightTotal int

	// Lightning distance (closest/minimum)
	var lightningDistMid float64 = 999   // closest today (min value)
	var lightningDistRange float64 = 999 // closest in range (min value)

	// Temperature metrics
	var tHiMid, fHiMid, dHiMid, hHiMid, bHiMid float64 = -999, -999, -999, -999, -999
	var tLoMid, fLoMid, dLoMid, hLoMid, bLoMid float64 = 999, 999, 999, 999, 999
	var tHiRange, fHiRange, dHiRange, hHiRange, bHiRange float64 = -999, -999, -999, -999, -999
	var tLoRange, fLoRange, dLoRange, hLoRange, bLoRange float64 = 999, 999, 999, 999, 999

	// Inside temp/humidity
	var inTHiMid, inHHiMid float64 = -999, -999
	var inTLoMid, inHLoMid float64 = 999, 999
	var inTHiRange, inHHiRange float64 = -999, -999
	var inTLoRange, inHLoRange float64 = 999, 999

	// Wind metrics
	var windSumRange, windSumMid float64
	var windCountRange, windCountMid int
	var windSqSumRange, windSqSumMid float64
	var vecUxRange, vecUyRange, vecUxMid, vecUyMid float64
	var gustMaxRange, gustMaxMid float64
	var maxWindDirRange, maxWindDirMid sql.NullFloat64

	// Rain rate max
	var rrMid, rrRange float64

	for rows.Next() {
		var epochSec int64
		var rain, rainRate sql.NullFloat64
		var strikes sql.NullInt64
		var lightningDist sql.NullFloat64
		var outTemp, dewpoint, outHumidity, barometer sql.NullFloat64
		var heatindex, windchill sql.NullFloat64
		var windSpeed, windGust, windDir sql.NullFloat64
		var inTemp, inHumidity sql.NullFloat64

		if err := rows.Scan(&epochSec, &rain, &rainRate, &strikes, &lightningDist,
			&outTemp, &dewpoint, &outHumidity, &barometer,
			&heatindex, &windchill, &windSpeed, &windGust, &windDir,
			&inTemp, &inHumidity); err != nil {
			log.Println("DB scan error (statistics):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		isMidnight := epochSec >= midnightUnix

		// Rain accumulation
		if rain.Valid {
			rainRangeTotal += rain.Float64
			if isMidnight {
				rainMidnightTotal += rain.Float64
			}
		}

		// Lightning strikes
		if strikes.Valid {
			strikeRangeTotal += int(strikes.Int64)
			if isMidnight {
				strikeMidnightTotal += int(strikes.Int64)
			}
		}

		// Lightning distance (track minimum/closest)
		if lightningDist.Valid && lightningDist.Float64 > 0 {
			if lightningDist.Float64 < lightningDistRange {
				lightningDistRange = lightningDist.Float64
			}
			if isMidnight && lightningDist.Float64 < lightningDistMid {
				lightningDistMid = lightningDist.Float64
			}
		}

		// Rain rate max
		if rainRate.Valid {
			if rainRate.Float64 > rrRange {
				rrRange = rainRate.Float64
			}
			if isMidnight && rainRate.Float64 > rrMid {
				rrMid = rainRate.Float64
			}
		}

		// Temperature hi/lo
		if outTemp.Valid {
			if outTemp.Float64 > tHiRange {
				tHiRange = outTemp.Float64
			}
			if outTemp.Float64 < tLoRange {
				tLoRange = outTemp.Float64
			}
			if isMidnight {
				if outTemp.Float64 > tHiMid {
					tHiMid = outTemp.Float64
				}
				if outTemp.Float64 < tLoMid {
					tLoMid = outTemp.Float64
				}
			}
		}

		// Feels-like (prefer heatindex, fallback to windchill, then outTemp)
		var feels float64
		if heatindex.Valid {
			feels = heatindex.Float64
		} else if windchill.Valid {
			feels = windchill.Float64
		} else if outTemp.Valid {
			feels = outTemp.Float64
		} else {
			feels = -999
		}

		if feels != -999 {
			if feels > fHiRange {
				fHiRange = feels
			}
			if feels < fLoRange {
				fLoRange = feels
			}
			if isMidnight {
				if feels > fHiMid {
					fHiMid = feels
				}
				if feels < fLoMid {
					fLoMid = feels
				}
			}
		}

		// Dewpoint hi/lo
		if dewpoint.Valid {
			if dewpoint.Float64 > dHiRange {
				dHiRange = dewpoint.Float64
			}
			if dewpoint.Float64 < dLoRange {
				dLoRange = dewpoint.Float64
			}
			if isMidnight {
				if dewpoint.Float64 > dHiMid {
					dHiMid = dewpoint.Float64
				}
				if dewpoint.Float64 < dLoMid {
					dLoMid = dewpoint.Float64
				}
			}
		}

		// Humidity hi/lo
		if outHumidity.Valid {
			if outHumidity.Float64 > hHiRange {
				hHiRange = outHumidity.Float64
			}
			if outHumidity.Float64 < hLoRange {
				hLoRange = outHumidity.Float64
			}
			if isMidnight {
				if outHumidity.Float64 > hHiMid {
					hHiMid = outHumidity.Float64
				}
				if outHumidity.Float64 < hLoMid {
					hLoMid = outHumidity.Float64
				}
			}
		}

		// Barometer hi/lo
		if barometer.Valid {
			if barometer.Float64 > bHiRange {
				bHiRange = barometer.Float64
			}
			if barometer.Float64 < bLoRange {
				bLoRange = barometer.Float64
			}
			if isMidnight {
				if barometer.Float64 > bHiMid {
					bHiMid = barometer.Float64
				}
				if barometer.Float64 < bLoMid {
					bLoMid = barometer.Float64
				}
			}
		}

		// Wind statistics
		if windSpeed.Valid {
			speed := windSpeed.Float64
			windSumRange += speed
			windCountRange++
			windSqSumRange += speed * speed

			if windDir.Valid {
				rad := (windDir.Float64 * math.Pi) / 180.0
				vecUxRange += speed * math.Cos(rad)
				vecUyRange += speed * math.Sin(rad)
			}

			if isMidnight {
				windSumMid += speed
				windCountMid++
				windSqSumMid += speed * speed

				if windDir.Valid {
					rad := (windDir.Float64 * math.Pi) / 180.0
					vecUxMid += speed * math.Cos(rad)
					vecUyMid += speed * math.Sin(rad)
				}
			}
		}

		if windGust.Valid && windGust.Float64 > gustMaxRange {
			gustMaxRange = windGust.Float64
			if windDir.Valid {
				maxWindDirRange = windDir
			}
		}

		if isMidnight && windGust.Valid && windGust.Float64 > gustMaxMid {
			gustMaxMid = windGust.Float64
			if windDir.Valid {
				maxWindDirMid = windDir
			}
		}

		// Inside temp/humidity
		if inTemp.Valid {
			if inTemp.Float64 > inTHiRange {
				inTHiRange = inTemp.Float64
			}
			if inTemp.Float64 < inTLoRange {
				inTLoRange = inTemp.Float64
			}
			if isMidnight {
				if inTemp.Float64 > inTHiMid {
					inTHiMid = inTemp.Float64
				}
				if inTemp.Float64 < inTLoMid {
					inTLoMid = inTemp.Float64
				}
			}
		}

		if inHumidity.Valid {
			if inHumidity.Float64 > inHHiRange {
				inHHiRange = inHumidity.Float64
			}
			if inHumidity.Float64 < inHLoRange {
				inHLoRange = inHumidity.Float64
			}
			if isMidnight {
				if inHumidity.Float64 > inHHiMid {
					inHHiMid = inHumidity.Float64
				}
				if inHumidity.Float64 < inHLoMid {
					inHLoMid = inHumidity.Float64
				}
			}
		}
	}

	if err := rows.Err(); err != nil {
		log.Println("DB rows error (statistics):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
	}

	// Format helper functions
	fmt0 := func(v float64) string { return fmt.Sprintf("%d", int(math.Round(v))) }
	fmt1 := func(v float64) string { return fmt.Sprintf("%.1f", v) }
	fmt2 := func(v float64) string { return fmt.Sprintf("%.2f", v) }
	hiLo := func(hi, lo float64, formatter func(float64) string) string {
		if hi == -999 || lo == 999 {
			return "--"
		}
		return formatter(hi) + " / " + formatter(lo)
	}

	// Wind calculations
	avgMid := "--"
	avgRange := "--"
	if windCountMid > 0 {
		avgMid = fmt0(windSumMid / float64(windCountMid))
	}
	if windCountRange > 0 {
		avgRange = fmt0(windSumRange / float64(windCountRange))
	}

	rmsMid := "--"
	rmsRange := "--"
	if windCountMid > 0 {
		rmsMid = fmt0(math.Sqrt(windSqSumMid / float64(windCountMid)))
	}
	if windCountRange > 0 {
		rmsRange = fmt0(math.Sqrt(windSqSumRange / float64(windCountRange)))
	}

	vecMagMid := "--"
	vecMagRange := "--"
	vecDirMid := "--"
	vecDirRange := "--"

	if windCountMid > 0 && (vecUxMid != 0 || vecUyMid != 0) {
		mag := math.Sqrt(vecUxMid*vecUxMid+vecUyMid*vecUyMid) / float64(windCountMid)
		vecMagMid = fmt0(mag)
		dir := math.Atan2(vecUyMid, vecUxMid) * 180.0 / math.Pi
		if dir < 0 {
			dir += 360
		}
		vecDirMid = fmt0(dir)
	}

	if windCountRange > 0 && (vecUxRange != 0 || vecUyRange != 0) {
		mag := math.Sqrt(vecUxRange*vecUxRange+vecUyRange*vecUyRange) / float64(windCountRange)
		vecMagRange = fmt0(mag)
		dir := math.Atan2(vecUyRange, vecUxRange) * 180.0 / math.Pi
		if dir < 0 {
			dir += 360
		}
		vecDirRange = fmt0(dir)
	}

	maxWindMid := "--"
	maxWindRange := "--"
	if gustMaxMid > 0 {
		maxWindMid = fmt0(gustMaxMid)
		if maxWindDirMid.Valid {
			maxWindMid += " • " + fmt0(maxWindDirMid.Float64)
		}
	}
	if gustMaxRange > 0 {
		maxWindRange = fmt0(gustMaxRange)
		if maxWindDirRange.Valid {
			maxWindRange += " • " + fmt0(maxWindDirRange.Float64)
		}
	}

	// Build response
	stats := StatisticsData{
		RainToday: rainMidnightTotal,
		RainRange: rainRangeTotal,

		StrikesToday: strikeMidnightTotal,
		StrikesRange: strikeRangeTotal,

		TempToday: hiLo(tHiMid, tLoMid, fmt1),
		TempRange: hiLo(tHiRange, tLoRange, fmt1),

		FeelsToday: hiLo(fHiMid, fLoMid, fmt1),
		FeelsRange: hiLo(fHiRange, fLoRange, fmt1),

		DewToday: hiLo(dHiMid, dLoMid, fmt1),
		DewRange: hiLo(dHiRange, dLoRange, fmt1),

		HumidityToday: hiLo(hHiMid, hLoMid, fmt0),
		HumidityRange: hiLo(hHiRange, hLoRange, fmt0),

		BarometerToday: hiLo(bHiMid, bLoMid, fmt2),
		BarometerRange: hiLo(bHiRange, bLoRange, fmt2),

		WindAvgToday: avgMid,
		WindAvgRange: avgRange,

		WindMaxToday: maxWindMid,
		WindMaxRange: maxWindRange,

		WindRmsToday: rmsMid,
		WindRmsRange: rmsRange,

		WindVectorToday: vecMagMid,
		WindVectorRange: vecMagRange,

		WindVectorDirToday: vecDirMid,
		WindVectorDirRange: vecDirRange,

		RainRateToday: fmt2(rrMid),
		RainRateRange: fmt2(rrRange),

		LightningDistToday: func() string {
			if lightningDistMid < 999 {
				return fmt.Sprintf("%.1f", lightningDistMid)
			}
			return "--"
		}(),
		LightningDistRange: func() string {
			if lightningDistRange < 999 {
				return fmt.Sprintf("%.1f", lightningDistRange)
			}
			return "--"
		}(),

		InsideTempToday: hiLo(inTHiMid, inTLoMid, fmt1),
		InsideTempRange: hiLo(inTHiRange, inTLoRange, fmt1),

		InsideHumToday: hiLo(inHHiMid, inHLoMid, fmt0),
		InsideHumRange: hiLo(inHHiRange, inHLoRange, fmt0),
	}

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
}

// -------------------- /api/celestial --------------------

func handleCelestial(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse date parameter (defaults to today in local time)
	dateStr := r.URL.Query().Get("date")

	// Use Rita Ranch, AZ coordinates (from config or hardcoded)
	coords := astroglide.Coordinates{
		Lat: 32.093174,   // Rita Ranch latitude
		Lon: -110.777557, // Rita Ranch longitude (west is negative)
	}

	// Load MST/America/Phoenix timezone
	loc, err := time.LoadLocation("America/Phoenix")
	if err != nil {
		log.Println("Failed to load America/Phoenix timezone:", err)
		http.Error(w, "Timezone error", http.StatusInternalServerError)
		return
	}

	var date time.Time
	if dateStr == "" {
		// Default to today
		now := time.Now().In(loc)
		date = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	} else {
		// Parse YYYY-MM-DD format
		parsed, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			log.Println("Invalid date parameter:", dateStr, err)
			http.Error(w, "Invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		date = parsed
	}

	// Cache key is date + timezone to support different locales if requested
	cacheKey := date.Format("2006-01-02") + "|" + loc.String()
	celestialCache.RLock()
	if ce, ok := celestialCache.m[cacheKey]; ok && time.Now().Before(ce.expiry) {
		// Serve cached response
		_ = json.NewEncoder(w).Encode(ce.data)
		celestialCache.RUnlock()
		return
	}
	celestialCache.RUnlock()

	// Use singleflight to dedupe concurrent compute for the same cacheKey
	result, err, _ := celestialGroup.Do(cacheKey, func() (interface{}, error) {
		// Double-check cache inside singleflight (another goroutine may have populated it)
		celestialCache.RLock()
		if ce, ok := celestialCache.m[cacheKey]; ok && time.Now().Before(ce.expiry) {
			celestialCache.RUnlock()
			return ce.data, nil
		}
		celestialCache.RUnlock()

		return computeCelestialData(coords, date, loc)
	})

	if err != nil {
		log.Println("Error computing celestial data:", err)
		http.Error(w, "Celestial calculation error", http.StatusInternalServerError)
		return
	}

	celestial := result.(CelestialData)

	// Cache the result until the next local midnight for the requested date
	// Compute next midnight explicitly in the request's location to avoid timezone issues.
	nextMidnight := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc).Add(24 * time.Hour)
	expiry := nextMidnight
	celestialCache.Lock()
	celestialCache.m[cacheKey] = &cachedCelestial{data: celestial, expiry: expiry}
	celestialCache.Unlock()

	if err := json.NewEncoder(w).Encode(celestial); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
}

// computeCelestialData performs the actual astroglide calculations
func computeCelestialData(coords astroglide.Coordinates, date time.Time, loc *time.Location) (CelestialData, error) {
	// Compute sun rise/set using astroglide
	sunRS, sunErr := astroglide.RiseSetFor(astroglide.Sun, coords, date)
	if sunErr != nil && sunErr != astroglide.ErrNoRiseNoSet {
		return CelestialData{}, fmt.Errorf("sunrise/sunset calculation: %w", sunErr)
	}

	// Compute moon rise/set using astroglide
	moonRS, moonErr := astroglide.RiseSetFor(astroglide.Moon, coords, date)
	if moonErr != nil && moonErr != astroglide.ErrNoRiseNoSet {
		return CelestialData{}, fmt.Errorf("moonrise/moonset calculation: %w", moonErr)
	}

	// Compute moon phase at current time (or noon on the requested date)
	noonTime := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, loc)
	moonPhase, phaseErr := astroglide.MoonPhaseAt(noonTime)
	// Non-fatal if phase calculation fails

	// Compute twilight times
	civilTwilight, _ := astroglide.TwilightFor(coords, date, astroglide.TwilightCivil)
	nauticalTwilight, _ := astroglide.TwilightFor(coords, date, astroglide.TwilightNautical)
	astronomicalTwilight, _ := astroglide.TwilightFor(coords, date, astroglide.TwilightAstronomical)

	// Compute golden hour and blue hour
	goldenHour, _ := astroglide.GoldenHourFor(coords, date)
	blueHour, _ := astroglide.BlueHourFor(coords, date)

	// Compute daylight hours (duration between sunrise and sunset)
	daylightHours, daylightErr := astroglide.DaylightHours(coords, date)

	// Build response
	celestial := CelestialData{
		Date:     date.Format("2006-01-02"),
		Timezone: loc.String(),
	}

	// Only include rise/set if they exist
	if sunErr == nil || sunErr == astroglide.ErrNoRiseNoSet {
		if !sunRS.Rise.IsZero() {
			celestial.Sunrise = &sunRS.Rise
			celestial.Sunrise24 = sunRS.Rise.In(loc).Format("15:04")
		}
		if !sunRS.Set.IsZero() {
			celestial.Sunset = &sunRS.Set
			celestial.Sunset24 = sunRS.Set.In(loc).Format("15:04")
		}
	}

	// Include daylight hours if calculation succeeded
	if daylightErr == nil {
		celestial.DaylightHours = daylightHours
		// Format as "Xh Ym" for display
		hours := int(daylightHours)
		minutes := int((daylightHours - float64(hours)) * 60)
		celestial.DaylightHoursFormatted = fmt.Sprintf("%dh %dm", hours, minutes)
	}

	if moonErr == nil || moonErr == astroglide.ErrNoRiseNoSet {
		if !moonRS.Rise.IsZero() {
			celestial.Moonrise = &moonRS.Rise
			celestial.Moonrise24 = moonRS.Rise.In(loc).Format("15:04")
		}
		if !moonRS.Set.IsZero() {
			celestial.Moonset = &moonRS.Set
			celestial.Moonset24 = moonRS.Set.In(loc).Format("15:04")
		}
	}

	// Add moon phase if available
	if phaseErr == nil {
		percentage := int(moonPhase.Fraction * 100)
		celestial.MoonPhase = &MoonPhase{
			Fraction:   moonPhase.Fraction,
			Elongation: moonPhase.Elongation,
			Waxing:     moonPhase.Waxing,
			Name:       moonPhase.Name,
			Percentage: percentage,
		}
	}

	// Add civil twilight
	if !civilTwilight.Rise.IsZero() {
		celestial.CivilDawn = &civilTwilight.Rise
		celestial.CivilDawn24 = civilTwilight.Rise.In(loc).Format("15:04")
	}
	if !civilTwilight.Set.IsZero() {
		celestial.CivilDusk = &civilTwilight.Set
		celestial.CivilDusk24 = civilTwilight.Set.In(loc).Format("15:04")
	}

	// Add nautical twilight
	if !nauticalTwilight.Rise.IsZero() {
		celestial.NauticalDawn = &nauticalTwilight.Rise
		celestial.NauticalDawn24 = nauticalTwilight.Rise.In(loc).Format("15:04")
	}
	if !nauticalTwilight.Set.IsZero() {
		celestial.NauticalDusk = &nauticalTwilight.Set
		celestial.NauticalDusk24 = nauticalTwilight.Set.In(loc).Format("15:04")
	}

	// Add astronomical twilight
	if !astronomicalTwilight.Rise.IsZero() {
		celestial.AstronomicalDawn = &astronomicalTwilight.Rise
		celestial.AstronomicalDawn24 = astronomicalTwilight.Rise.In(loc).Format("15:04")
	}
	if !astronomicalTwilight.Set.IsZero() {
		celestial.AstronomicalDusk = &astronomicalTwilight.Set
		celestial.AstronomicalDusk24 = astronomicalTwilight.Set.In(loc).Format("15:04")
	}

	// Add golden hour
	if goldenHour.HasMorning {
		celestial.GoldenHourMorningStart = &goldenHour.Morning.Start
		celestial.GoldenHourMorningEnd = &goldenHour.Morning.End
		celestial.GoldenHourMorningStart24 = goldenHour.Morning.Start.In(loc).Format("15:04")
		celestial.GoldenHourMorningEnd24 = goldenHour.Morning.End.In(loc).Format("15:04")
	}
	if goldenHour.HasEvening {
		celestial.GoldenHourEveningStart = &goldenHour.Evening.Start
		celestial.GoldenHourEveningEnd = &goldenHour.Evening.End
		celestial.GoldenHourEveningStart24 = goldenHour.Evening.Start.In(loc).Format("15:04")
		celestial.GoldenHourEveningEnd24 = goldenHour.Evening.End.In(loc).Format("15:04")
	}

	// Add blue hour
	if blueHour.HasMorning {
		celestial.BlueHourMorningStart = &blueHour.Morning.Start
		celestial.BlueHourMorningEnd = &blueHour.Morning.End
		celestial.BlueHourMorningStart24 = blueHour.Morning.Start.In(loc).Format("15:04")
		celestial.BlueHourMorningEnd24 = blueHour.Morning.End.In(loc).Format("15:04")
	}
	if blueHour.HasEvening {
		celestial.BlueHourEveningStart = &blueHour.Evening.Start
		celestial.BlueHourEveningEnd = &blueHour.Evening.End
		celestial.BlueHourEveningStart24 = blueHour.Evening.Start.In(loc).Format("15:04")
		celestial.BlueHourEveningEnd24 = blueHour.Evening.End.In(loc).Format("15:04")
	}

	return celestial, nil
}

// refreshCelestialCacheDaily runs a background goroutine that refreshes today's and tomorrow's
// celestial cache at 00:05 local time daily to avoid first-request latency.
func refreshCelestialCacheDaily(stop <-chan struct{}) {
	// Load timezone once
	loc, err := time.LoadLocation("America/Phoenix")
	if err != nil {
		log.Println("[Celestial Refresh] Failed to load timezone:", err)
		return
	}

	// Rita Ranch coordinates (same as handler)
	coords := astroglide.Coordinates{
		Lat: 32.093174,
		Lon: -110.777557,
	}

	for {
		// Compute next 00:05 local
		now := time.Now().In(loc)
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), 0, 5, 0, 0, loc)
		if now.After(nextRun) {
			// Already past 00:05 today, schedule for tomorrow
			nextRun = nextRun.Add(24 * time.Hour)
		}

		waitDuration := time.Until(nextRun)
		log.Printf("[Celestial Refresh] Next refresh scheduled at %s (in %s)\n", nextRun.Format("2006-01-02 15:04:05 MST"), waitDuration.Round(time.Second))

		select {
		case <-time.After(waitDuration):
			// Refresh today and tomorrow
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
			tomorrow := today.Add(24 * time.Hour)

			log.Println("[Celestial Refresh] Refreshing cache for today and tomorrow...")

			// Refresh today
			if data, err := computeCelestialData(coords, today, loc); err == nil {
				cacheKey := today.Format("2006-01-02") + "|" + loc.String()
				expiry := today.Add(24 * time.Hour)
				celestialCache.Lock()
				celestialCache.m[cacheKey] = &cachedCelestial{data: data, expiry: expiry}
				celestialCache.Unlock()
				log.Printf("[Celestial Refresh] Cached data for %s\n", today.Format("2006-01-02"))
			} else {
				log.Printf("[Celestial Refresh] Failed to compute today: %v\n", err)
			}

			// Refresh tomorrow
			if data, err := computeCelestialData(coords, tomorrow, loc); err == nil {
				cacheKey := tomorrow.Format("2006-01-02") + "|" + loc.String()
				expiry := tomorrow.Add(24 * time.Hour)
				celestialCache.Lock()
				celestialCache.m[cacheKey] = &cachedCelestial{data: data, expiry: expiry}
				celestialCache.Unlock()
				log.Printf("[Celestial Refresh] Cached data for %s\n", tomorrow.Format("2006-01-02"))
			} else {
				log.Printf("[Celestial Refresh] Failed to compute tomorrow: %v\n", err)
			}

		case <-stop:
			log.Println("[Celestial Refresh] Stopping background refresh")
			return
		}
	}
}

// -------------------- /api/csv/daily --------------------

func handleCSVDaily(w http.ResponseWriter, r *http.Request) {
	// Get date parameters
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")
	dayStr := r.URL.Query().Get("day")
	columnsParam := r.URL.Query().Get("columns")

	if yearStr == "" || monthStr == "" || dayStr == "" {
		http.Error(w, "Missing required parameters: year, month, day", http.StatusBadRequest)
		return
	}

	// Parse date
	dateStr := fmt.Sprintf("%s-%s-%s", yearStr, monthStr, dayStr)
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Get local timezone for midnight calculations
	loc, err := time.LoadLocation("America/Phoenix")
	if err != nil {
		loc = time.Local
	}

	// Calculate start and end times (midnight to midnight in local time)
	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	startUnix := startOfDay.Unix()
	endUnix := endOfDay.Unix()

	// Parse columns parameter (default to all columns if not specified)
	requestedColumns := []string{
		"dateTime", "outTemp", "dewpoint", "outHumidity", "barometer",
		"heatindex", "windchill", "windSpeed", "windGust", "windDir",
		"rainRate", "rain", "lightning_strike_count", "lightning_distance",
		"inTemp", "inHumidity",
	}

	if columnsParam != "" {
		requestedColumns = strings.Split(columnsParam, ",")
		// Validate that dateTime is always first
		if len(requestedColumns) == 0 || requestedColumns[0] != "dateTime" {
			http.Error(w, "Timestamp (dateTime) must be the first column", http.StatusBadRequest)
			return
		}
	}

	// Build SQL query dynamically
	selectClause := strings.Join(requestedColumns, ", ")
	query := fmt.Sprintf("SELECT %s FROM archive WHERE dateTime >= ? AND dateTime < ? ORDER BY dateTime ASC", selectClause)

	rows, err := db.Query(query, startUnix, endUnix)
	if err != nil {
		log.Println("CSV query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Set headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"weather_%s.csv\"", dateStr))

	// Column labels mapping
	columnLabels := map[string]string{
		"dateTime":               "Timestamp",
		"outTemp":                "Temperature_F",
		"dewpoint":               "Dewpoint_F",
		"outHumidity":            "Humidity_Pct",
		"barometer":              "Barometer_inHg",
		"heatindex":              "HeatIndex_F",
		"windchill":              "WindChill_F",
		"windSpeed":              "WindSpeed_mph",
		"windGust":               "WindGust_mph",
		"windDir":                "WindDirection_deg",
		"rainRate":               "RainRate_in_hr",
		"rain":                   "Rain_in",
		"lightning_strike_count": "LightningStrikes",
		"lightning_distance":     "LightningDistance_mi",
		"inTemp":                 "InsideTemp_F",
		"inHumidity":             "InsideHumidity_Pct",
	}

	// Write CSV header
	headerParts := make([]string, len(requestedColumns))
	for i, col := range requestedColumns {
		if label, ok := columnLabels[col]; ok {
			headerParts[i] = label
		} else {
			headerParts[i] = col
		}
	}
	fmt.Fprintf(w, "%s\n", strings.Join(headerParts, ","))

	// Helper to format nullable float
	formatFloat := func(nf sql.NullFloat64) string {
		if nf.Valid {
			return fmt.Sprintf("%.2f", nf.Float64)
		}
		return ""
	}

	formatInt := func(ni sql.NullInt64) string {
		if ni.Valid {
			return fmt.Sprintf("%d", ni.Int64)
		}
		return ""
	}

	// Write data rows
	rowCount := 0
	for rows.Next() {
		// Prepare scan destinations based on requested columns
		scanDest := make([]interface{}, len(requestedColumns))
		values := make([]string, len(requestedColumns))

		for i, col := range requestedColumns {
			switch col {
			case "dateTime":
				var epochSec int64
				scanDest[i] = &epochSec
			case "lightning_strike_count":
				var val sql.NullInt64
				scanDest[i] = &val
			default:
				var val sql.NullFloat64
				scanDest[i] = &val
			}
		}

		if err := rows.Scan(scanDest...); err != nil {
			log.Println("CSV scan error:", err)
			continue
		}

		// Format values for CSV output
		for i, col := range requestedColumns {
			switch col {
			case "dateTime":
				epochSec := *scanDest[i].(*int64)
				ts := time.Unix(epochSec, 0).In(loc)
				values[i] = ts.Format("2006-01-02 15:04:05")
			case "lightning_strike_count":
				values[i] = formatInt(*scanDest[i].(*sql.NullInt64))
			default:
				values[i] = formatFloat(*scanDest[i].(*sql.NullFloat64))
			}
		}

		fmt.Fprintf(w, "%s\n", strings.Join(values, ","))
		rowCount++
	}

	if err := rows.Err(); err != nil {
		log.Println("CSV rows error:", err)
	}

	log.Printf("[CSV] Generated daily CSV for %s: %d rows, %d columns\n", dateStr, rowCount, len(requestedColumns))
}

func handleCSVRange(w http.ResponseWriter, r *http.Request) {
	// Get date range parameters
	startYearStr := r.URL.Query().Get("startYear")
	startMonthStr := r.URL.Query().Get("startMonth")
	startDayStr := r.URL.Query().Get("startDay")
	endYearStr := r.URL.Query().Get("endYear")
	endMonthStr := r.URL.Query().Get("endMonth")
	endDayStr := r.URL.Query().Get("endDay")
	columnsParam := r.URL.Query().Get("columns")

	if startYearStr == "" || startMonthStr == "" || startDayStr == "" ||
		endYearStr == "" || endMonthStr == "" || endDayStr == "" {
		http.Error(w, "Missing required parameters: startYear, startMonth, startDay, endYear, endMonth, endDay", http.StatusBadRequest)
		return
	}

	// Parse start date
	startDateStr := fmt.Sprintf("%s-%s-%s", startYearStr, startMonthStr, startDayStr)
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}

	// Parse end date
	endDateStr := fmt.Sprintf("%s-%s-%s", endYearStr, endMonthStr, endDayStr)
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	// Validate date range
	if startDate.After(endDate) {
		http.Error(w, "Start date must be before or equal to end date", http.StatusBadRequest)
		return
	}

	// Get local timezone for midnight calculations
	loc, err := time.LoadLocation("America/Phoenix")
	if err != nil {
		loc = time.Local
	}

	// Calculate start and end times (midnight to midnight in local time)
	startOfRange := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
	endOfRange := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, loc).Add(time.Second)

	startUnix := startOfRange.Unix()
	endUnix := endOfRange.Unix()

	// Parse columns parameter (default to all columns if not specified)
	requestedColumns := []string{
		"dateTime", "outTemp", "dewpoint", "outHumidity", "barometer",
		"heatindex", "windchill", "windSpeed", "windGust", "windDir",
		"rainRate", "rain", "lightning_strike_count", "lightning_distance",
		"inTemp", "inHumidity",
	}

	if columnsParam != "" {
		requestedColumns = strings.Split(columnsParam, ",")
		// Validate that dateTime is always first
		if len(requestedColumns) == 0 || requestedColumns[0] != "dateTime" {
			http.Error(w, "Timestamp (dateTime) must be the first column", http.StatusBadRequest)
			return
		}
	}

	// Build SQL query dynamically
	selectClause := strings.Join(requestedColumns, ", ")
	query := fmt.Sprintf("SELECT %s FROM archive WHERE dateTime >= ? AND dateTime < ? ORDER BY dateTime ASC", selectClause)

	rows, err := db.Query(query, startUnix, endUnix)
	if err != nil {
		log.Println("CSV range query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Set headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"weather_%s_to_%s.csv\"", startDateStr, endDateStr))

	// Column labels mapping
	columnLabels := map[string]string{
		"dateTime":               "Timestamp",
		"outTemp":                "Temperature_F",
		"dewpoint":               "Dewpoint_F",
		"outHumidity":            "Humidity_Pct",
		"barometer":              "Barometer_inHg",
		"heatindex":              "HeatIndex_F",
		"windchill":              "WindChill_F",
		"windSpeed":              "WindSpeed_mph",
		"windGust":               "WindGust_mph",
		"windDir":                "WindDirection_deg",
		"rainRate":               "RainRate_in_hr",
		"rain":                   "Rain_in",
		"lightning_strike_count": "LightningStrikes",
		"lightning_distance":     "LightningDistance_mi",
		"inTemp":                 "InsideTemp_F",
		"inHumidity":             "InsideHumidity_Pct",
	}

	// Write CSV header
	headerParts := make([]string, len(requestedColumns))
	for i, col := range requestedColumns {
		if label, ok := columnLabels[col]; ok {
			headerParts[i] = label
		} else {
			headerParts[i] = col
		}
	}
	fmt.Fprintf(w, "%s\n", strings.Join(headerParts, ","))

	// Helper to format nullable float
	formatFloat := func(nf sql.NullFloat64) string {
		if nf.Valid {
			return fmt.Sprintf("%.2f", nf.Float64)
		}
		return ""
	}

	formatInt := func(ni sql.NullInt64) string {
		if ni.Valid {
			return fmt.Sprintf("%d", ni.Int64)
		}
		return ""
	}

	// Write data rows
	rowCount := 0
	for rows.Next() {
		// Prepare scan destinations based on requested columns
		scanDest := make([]interface{}, len(requestedColumns))
		values := make([]string, len(requestedColumns))

		for i, col := range requestedColumns {
			switch col {
			case "dateTime":
				var epochSec int64
				scanDest[i] = &epochSec
			case "lightning_strike_count":
				var val sql.NullInt64
				scanDest[i] = &val
			default:
				var val sql.NullFloat64
				scanDest[i] = &val
			}
		}

		if err := rows.Scan(scanDest...); err != nil {
			log.Println("CSV scan error:", err)
			continue
		}

		// Format values for CSV output
		for i, col := range requestedColumns {
			switch col {
			case "dateTime":
				epochSec := *scanDest[i].(*int64)
				ts := time.Unix(epochSec, 0).In(loc)
				values[i] = ts.Format("2006-01-02 15:04:05")
			case "lightning_strike_count":
				values[i] = formatInt(*scanDest[i].(*sql.NullInt64))
			default:
				values[i] = formatFloat(*scanDest[i].(*sql.NullFloat64))
			}
		}

		fmt.Fprintf(w, "%s\n", strings.Join(values, ","))
		rowCount++
	}

	if err := rows.Err(); err != nil {
		log.Println("CSV rows error:", err)
	}

	log.Printf("[CSV] Generated range CSV from %s to %s: %d rows, %d columns\n", startDateStr, endDateStr, rowCount, len(requestedColumns))
}
