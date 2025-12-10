package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	Params   string `yaml:"params"`
}

type ServerConfig struct {
	// HTTP server port
	Port int `yaml:"port"`
	// Server-side SSE poll interval in seconds (how often the broker polls DB)
	SSEPollSeconds int `yaml:"sse_poll_seconds"`
	// Client-side poll interval in seconds (how often the browser's poller calls loadAll())
	ClientPollSeconds int `yaml:"client_poll_seconds"`
}

type LocationConfig struct {
	// Station name
	Name string `yaml:"name"`
	// Latitude in decimal degrees
	Latitude float64 `yaml:"latitude"`
	// Longitude in decimal degrees
	Longitude float64 `yaml:"longitude"`
	// Altitude in feet above sea level
	Altitude float64 `yaml:"altitude"`
}

type AlertsConfig struct {
	// Extreme heat threshold (°F heat index)
	ExtremeHeat float64 `yaml:"extreme_heat"`
	// Extreme cold threshold (°F wind chill)
	ExtremeCold float64 `yaml:"extreme_cold"`
	// Strong wind speed threshold (mph)
	WindSpeed float64 `yaml:"wind_speed"`
	// Strong wind gust threshold (mph)
	WindGust float64 `yaml:"wind_gust"`
}

type AppConfig struct {
	DB       DBConfig       `yaml:"db"`
	Server   ServerConfig   `yaml:"server"`
	Location LocationConfig `yaml:"location"`
	Alerts   AlertsConfig   `yaml:"alerts"`
}

var appConfig AppConfig

func loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, &appConfig); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	// Apply sensible defaults if not set
	if appConfig.Server.SSEPollSeconds <= 0 {
		appConfig.Server.SSEPollSeconds = 60
	}
	if appConfig.Server.ClientPollSeconds <= 0 {
		appConfig.Server.ClientPollSeconds = 60
	}
	if appConfig.Server.Port == 0 {
		appConfig.Server.Port = 8081
	}

	return nil
}

func buildDSN() string {
	db := appConfig.DB
	// e.g. "user:pass@tcp(host:port)/name?params"
	if db.Params != "" {
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
			db.User, db.Password, db.Host, db.Port, db.Name, db.Params)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		db.User, db.Password, db.Host, db.Port, db.Name)
}
