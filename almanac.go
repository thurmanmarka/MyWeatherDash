package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/thurmanmarka/astroglide"
)

// -------------------- /api/almanac --------------------

func handleAlmanac(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dateStr := r.URL.Query().Get("date")

	coords := astroglide.Coordinates{
		Lat: appConfig.Location.Latitude,
		Lon: appConfig.Location.Longitude,
	}

	loc, err := time.LoadLocation("America/Phoenix")
	if err != nil {
		log.Println("Failed to load America/Phoenix timezone:", err)
		http.Error(w, "Timezone error", http.StatusInternalServerError)
		return
	}

	var date time.Time
	if dateStr == "" {
		now := time.Now().In(loc)
		date = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	} else {
		parsed, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			log.Println("Invalid date parameter:", dateStr, err)
			http.Error(w, "Invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		date = parsed
	}

	celestial, err := computeCelestialData(coords, date, loc)
	if err != nil {
		log.Println("Error computing celestial data for almanac:", err)
		http.Error(w, "Celestial calculation error", http.StatusInternalServerError)
		return
	}

	almanac := AlmanacData{Celestial: celestial}

	history, err := computeDayRecords(int(date.Month()), date.Day())
	if err != nil {
		// Non-fatal: still return sun/moon data even if the DB query fails
		log.Println("Error computing day records for almanac:", err)
	} else {
		almanac.History = history
		almanac.YearsOfData = len(history)

		var hiSum, loSum float64
		recordHigh, recordLow := -999.0, 999.0
		recordHighYear, recordLowYear := 0, 0
		for _, h := range history {
			hiSum += h.HighTemp
			loSum += h.LowTemp
			if h.HighTemp > recordHigh {
				recordHigh = h.HighTemp
				recordHighYear = h.Year
			}
			if h.LowTemp < recordLow {
				recordLow = h.LowTemp
				recordLowYear = h.Year
			}
		}
		if len(history) > 0 {
			almanac.RecordHighTemp = recordHigh
			almanac.RecordHighYear = recordHighYear
			almanac.RecordLowTemp = recordLow
			almanac.RecordLowYear = recordLowYear
			almanac.AvgHighTemp = hiSum / float64(len(history))
			almanac.AvgLowTemp = loSum / float64(len(history))
		}
	}

	if err := json.NewEncoder(w).Encode(almanac); err != nil {
		http.Error(w, "JSON error", http.StatusInternalServerError)
	}
}

// computeDayRecords queries the archive table for every past occurrence of the
// given month/day (across all years of station history), returning per-year
// high/low temperature and total rainfall, newest year first.
func computeDayRecords(month, day int) ([]DayRecord, error) {
	rows, err := db.Query(`
		SELECT YEAR(FROM_UNIXTIME(dateTime)) AS yr,
		       MAX(outTemp) AS hiTemp,
		       MIN(outTemp) AS loTemp,
		       COALESCE(SUM(rain), 0) AS rainTotal
		FROM archive
		WHERE MONTH(FROM_UNIXTIME(dateTime)) = ? AND DAY(FROM_UNIXTIME(dateTime)) = ?
		GROUP BY yr
		ORDER BY yr DESC
	`, month, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []DayRecord
	for rows.Next() {
		var rec DayRecord
		if err := rows.Scan(&rec.Year, &rec.HighTemp, &rec.LowTemp, &rec.RainTotal); err != nil {
			return nil, err
		}
		history = append(history, rec)
	}
	return history, rows.Err()
}
