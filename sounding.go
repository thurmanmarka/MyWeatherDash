package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Standard pressure levels (hPa) requested from Open-Meteo, surface to ~16.5km.
var soundingPressureLevels = []int{
	1000, 975, 950, 925, 900, 850, 800, 700, 600, 500, 400, 300, 250, 200, 150, 100,
}

var soundingHTTPClient = &http.Client{Timeout: 10 * time.Second}

var soundingGroup singleflight.Group

var soundingCache = struct {
	sync.RWMutex
	data   *SoundingData
	expiry time.Time
}{}

// -------------------- /api/sounding --------------------

func handleSounding(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	force := r.URL.Query().Get("force") == "1"

	if !force {
		soundingCache.RLock()
		if soundingCache.data != nil && time.Now().Before(soundingCache.expiry) {
			data := soundingCache.data
			soundingCache.RUnlock()
			_ = json.NewEncoder(w).Encode(data)
			return
		}
		soundingCache.RUnlock()
	}

	result, err, _ := soundingGroup.Do("sounding", func() (interface{}, error) {
		if !force {
			soundingCache.RLock()
			if soundingCache.data != nil && time.Now().Before(soundingCache.expiry) {
				d := soundingCache.data
				soundingCache.RUnlock()
				return d, nil
			}
			soundingCache.RUnlock()
		}
		return fetchSounding(appConfig.Location.Latitude, appConfig.Location.Longitude, appConfig.Location.Altitude)
	})

	if err != nil {
		log.Println("Error fetching sounding data:", err)
		http.Error(w, "Failed to fetch sounding data", http.StatusBadGateway)
		return
	}

	sd := result.(*SoundingData)

	soundingCache.Lock()
	soundingCache.data = sd
	soundingCache.expiry = time.Now().Add(30 * time.Minute)
	soundingCache.Unlock()

	if err := json.NewEncoder(w).Encode(sd); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
}

// fetchSounding pulls a model-derived vertical profile from Open-Meteo's
// pressure-level forecast API for the given location. This is not a real
// balloon launch (RAOB) — it's the nearest grid-point model sounding,
// refreshed hourly, which is what lets this work for any lat/lon rather
// than only near a handful of fixed upper-air stations.
func fetchSounding(lat, lon, altitudeFt float64) (*SoundingData, error) {
	vars := []string{"temperature_2m", "dew_point_2m", "surface_pressure", "wind_speed_10m", "wind_direction_10m"}
	for _, lvl := range soundingPressureLevels {
		vars = append(vars,
			fmt.Sprintf("temperature_%dhPa", lvl),
			fmt.Sprintf("relative_humidity_%dhPa", lvl),
			fmt.Sprintf("wind_speed_%dhPa", lvl),
			fmt.Sprintf("wind_direction_%dhPa", lvl),
			fmt.Sprintf("geopotential_height_%dhPa", lvl),
		)
	}

	q := url.Values{}
	q.Set("latitude", strconv.FormatFloat(lat, 'f', 4, 64))
	q.Set("longitude", strconv.FormatFloat(lon, 'f', 4, 64))
	q.Set("hourly", strings.Join(vars, ","))
	q.Set("wind_speed_unit", "kn")
	q.Set("timezone", "auto")
	q.Set("forecast_days", "1")

	reqURL := "https://api.open-meteo.com/v1/forecast?" + q.Encode()

	resp, err := soundingHTTPClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("open-meteo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("open-meteo status %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		Hourly map[string]json.RawMessage `json:"hourly"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding open-meteo response: %w", err)
	}

	var times []string
	if err := json.Unmarshal(raw.Hourly["time"], &times); err != nil {
		return nil, fmt.Errorf("decoding time array: %w", err)
	}
	if len(times) == 0 {
		return nil, fmt.Errorf("no hourly data returned")
	}

	getSeries := func(key string) []float64 {
		msg, ok := raw.Hourly[key]
		if !ok {
			return nil
		}
		var vals []*float64
		if err := json.Unmarshal(msg, &vals); err != nil {
			return nil
		}
		out := make([]float64, len(vals))
		for i, v := range vals {
			if v != nil {
				out[i] = *v
			}
		}
		return out
	}

	// Pick the hourly slot closest to right now.
	idx := 0
	now := time.Now()
	bestDiff := math.MaxFloat64
	for i, ts := range times {
		t, err := time.Parse("2006-01-02T15:04", ts)
		if err != nil {
			continue
		}
		diff := math.Abs(now.Sub(t).Seconds())
		if diff < bestDiff {
			bestDiff = diff
			idx = i
		}
	}

	surfaceTemp := getSeries("temperature_2m")
	surfaceDew := getSeries("dew_point_2m")
	surfacePres := getSeries("surface_pressure")
	surfaceWindSpd := getSeries("wind_speed_10m")
	surfaceWindDir := getSeries("wind_direction_10m")

	sd := &SoundingData{
		GeneratedAt: times[idx],
		Source:      "Open-Meteo model analysis (nearest grid point, not a balloon launch)",
	}
	if idx < len(surfacePres) {
		sd.SurfacePressureHPa = surfacePres[idx]
	}

	if idx < len(surfaceTemp) && idx < len(surfaceDew) && sd.SurfacePressureHPa > 0 {
		sd.Levels = append(sd.Levels, SoundingLevel{
			PressureHPa: sd.SurfacePressureHPa,
			HeightM:     altitudeFt / 3.28084,
			TempC:       surfaceTemp[idx],
			DewpointC:   surfaceDew[idx],
			WindSpeedKt: valueAt(surfaceWindSpd, idx),
			WindDirDeg:  valueAt(surfaceWindDir, idx),
		})
	}

	for _, lvl := range soundingPressureLevels {
		// Skip levels at or below the surface.
		if sd.SurfacePressureHPa > 0 && float64(lvl) >= sd.SurfacePressureHPa {
			continue
		}

		temp := getSeries(fmt.Sprintf("temperature_%dhPa", lvl))
		rh := getSeries(fmt.Sprintf("relative_humidity_%dhPa", lvl))
		if idx >= len(temp) || idx >= len(rh) {
			continue
		}
		wspd := getSeries(fmt.Sprintf("wind_speed_%dhPa", lvl))
		wdir := getSeries(fmt.Sprintf("wind_direction_%dhPa", lvl))
		height := getSeries(fmt.Sprintf("geopotential_height_%dhPa", lvl))

		t := temp[idx]
		sd.Levels = append(sd.Levels, SoundingLevel{
			PressureHPa: float64(lvl),
			HeightM:     valueAt(height, idx),
			TempC:       t,
			DewpointC:   dewpointFromRH(t, rh[idx]),
			WindSpeedKt: valueAt(wspd, idx),
			WindDirDeg:  valueAt(wdir, idx),
		})
	}

	if len(sd.Levels) < 2 {
		return nil, fmt.Errorf("insufficient sounding levels returned for this location")
	}

	sd.Parcel = computeParcelIndices(sd.Levels)

	return sd, nil
}

func valueAt(s []float64, i int) float64 {
	if i >= 0 && i < len(s) {
		return s[i]
	}
	return 0
}

// dewpointFromRH converts temperature (°C) + relative humidity (%) to dewpoint (°C)
// via the Magnus-Tetens approximation. Open-Meteo only exposes relative humidity
// (not dewpoint) at pressure levels.
func dewpointFromRH(tempC, rh float64) float64 {
	if rh <= 0.1 {
		rh = 0.1
	}
	const a = 17.625
	const b = 243.04
	gamma := math.Log(rh/100.0) + (a*tempC)/(b+tempC)
	return (b * gamma) / (a - gamma)
}
