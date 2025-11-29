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
)

// -------------------- helpers --------------------

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	// Provide client-side polling interval (ms) to the template
	data := struct {
		ClientPollMs int
		AssetVersion string
		LocationName string
	}{
		ClientPollMs: appConfig.Server.ClientPollSeconds * 1000,
		AssetVersion: time.Now().Format("20060102T150405"),
		LocationName: appConfig.Location.Name,
	}

	if err := tmplIndex.Execute(w, data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(PingResponse{Message: "pong"})
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
		SELECT dateTime, heatindex, windchill
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
		var heatF, chillF float64

		if err := rows.Scan(&epochSec, &heatF, &chillF); err != nil {
			log.Println("DB scan error (feelslike):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		ts := time.Unix(epochSec, 0)

		readings = append(readings, FeelsLikeReading{
			Timestamp: ts,
			HeatIndex: heatF,
			WindChill: chillF,
		})
	}
	if err := rows.Err(); err != nil {
		log.Println("DB rows error (feelslike):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(readings); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
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

	if err := json.NewEncoder(w).Encode(readings); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
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
        SELECT dateTime, lightning_strike_count
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

		if err := rows.Scan(&epochSec, &strikes); err != nil {
			log.Println("DB scan error (lightning):", err)
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return
		}

		ts := time.Unix(epochSec, 0)

		readings = append(readings, LightningReading{
			Timestamp: ts,
			Strikes:   strikes,
		})
	}

	if err := rows.Err(); err != nil {
		log.Println("DB rows error (lightning):", err)
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return
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
		SELECT dateTime, rain, rainRate, lightning_strike_count,
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
		var outTemp, dewpoint, outHumidity, barometer sql.NullFloat64
		var heatindex, windchill sql.NullFloat64
		var windSpeed, windGust, windDir sql.NullFloat64
		var inTemp, inHumidity sql.NullFloat64

		if err := rows.Scan(&epochSec, &rain, &rainRate, &strikes,
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

		TempToday: hiLo(tHiMid, tLoMid, fmt0),
		TempRange: hiLo(tHiRange, tLoRange, fmt0),

		FeelsToday: hiLo(fHiMid, fLoMid, fmt0),
		FeelsRange: hiLo(fHiRange, fLoRange, fmt0),

		WindchillToday: func() string {
			if fLoMid == 999 {
				return "--"
			}
			return fmt0(fLoMid)
		}(),
		WindchillRange: func() string {
			if fLoRange == 999 {
				return "--"
			}
			return fmt0(fLoRange)
		}(),

		DewToday: hiLo(dHiMid, dLoMid, fmt0),
		DewRange: hiLo(dHiRange, dLoRange, fmt0),

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

		LightningDistToday: "--",
		LightningDistRange: "--",

		InsideTempToday: hiLo(inTHiMid, inTLoMid, fmt0),
		InsideTempRange: hiLo(inTHiRange, inTLoRange, fmt0),

		InsideHumToday: hiLo(inHHiMid, inHLoMid, fmt0),
		InsideHumRange: hiLo(inHHiRange, inHLoRange, fmt0),
	}

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
}
