package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	tmplIndex *template.Template
	tmplKiosk *template.Template
	db        *sql.DB
)

func main() {
	var err error

	// Load config
	if err := loadConfig("config.yaml"); err != nil {
		log.Fatal("Error loading config:", err)
	}

	// Load HTML template
	tmplIndex, err = template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal("Error loading template:", err)
	}

	// Load kiosk template
	tmplKiosk, err = template.ParseFiles("templates/kiosk.html")
	if err != nil {
		log.Fatal("Error loading kiosk template:", err)
	}

	// Build DSN from config
	dsn := buildDSN()
	db, err = sql.Open("mysql", dsn)

	if err != nil {
		log.Fatal("Error opening DB:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Error pinging DB:", err)
	}

	// Routes
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handleWeatherDash)
	http.HandleFunc("/weather", handleWeatherDash)
	http.HandleFunc("/kiosk", handleKiosk)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/api/ping", handlePing)
	http.HandleFunc("/api/weather", handleWeather)
	http.HandleFunc("/api/barometer", handleBarometer)
	http.HandleFunc("/api/feelslike", handleFeelsLike)
	http.HandleFunc("/api/humidity", handleHumidity)
	http.HandleFunc("/api/wind", handleWind)
	http.HandleFunc("/api/rain", handleRain)
	http.HandleFunc("/api/lightning", handleLightning)
	http.HandleFunc("/api/insideTemp", handleInsideTemp)
	http.HandleFunc("/api/insideHumidity", handleInsideHumidity)
	http.HandleFunc("/api/celestial", handleCelestial)
	http.HandleFunc("/api/noaa/monthly", handleNOAAMonthly)
	http.HandleFunc("/api/noaa/yearly", handleNOAAYearly)
	http.HandleFunc("/api/statistics", handleStatistics)
	http.HandleFunc("/api/csv/daily", handleCSVDaily)
	http.HandleFunc("/api/csv/range", handleCSVRange)
	// Server-Sent Events stream (push updates)
	broker := NewSSEBroker(db)
	stopSSE := make(chan struct{})
	// Poll DB every configured seconds for new rows and broadcast
	broker.StartPolling(time.Duration(appConfig.Server.SSEPollSeconds)*time.Second, stopSSE)
	http.Handle("/api/stream", broker)

	// Start background celestial cache refresh (runs at 00:05 local daily)
	stopCelestialRefresh := make(chan struct{})
	go refreshCelestialCacheDaily(stopCelestialRefresh)

	addr := fmt.Sprintf(":%d", appConfig.Server.Port)
	log.Println("Server listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		close(stopSSE)
		close(stopCelestialRefresh)
		log.Fatal(err)
	}
}
