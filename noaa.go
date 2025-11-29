package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"
)

// NOAAMonthly and NOAAYearly summaries are rendered to text files under static/noaa/

type NOAAMonthlyParams struct {
	Year  int
	Month int // 1-12
}

type NOAAYearlyParams struct {
	Year int
}

// RenderMonthlyNOAA generates text content using archive table aggregates
func RenderMonthlyNOAA(db *sql.DB, p NOAAMonthlyParams) (string, error) {
	// Use MST7MDT timezone (Arizona local time, UTC-7)
	loc, err := time.LoadLocation("America/Phoenix")
	if err != nil {
		// Fallback to system local time if timezone not available
		loc = time.Now().Location()
	}
	start := time.Date(p.Year, time.Month(p.Month), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, 0)
	startUnix := start.Unix()
	endUnix := end.Unix()

	// Daily aggregates (temp high/low, mean, rain totals, wind avg & max, dominant dir)
	rows, err := db.Query(`
		SELECT dateTime, outTemp, dewpoint, outHumidity, barometer,
		       windSpeed, windGust, windDir, rain, rainRate
		FROM archive
		WHERE dateTime >= ? AND dateTime < ?
		ORDER BY dateTime ASC
	`, startUnix, endUnix)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Buckets per day
	type dayAgg struct {
		count                    int
		maxTemp, minTemp         float64
		maxTempTime, minTempTime string
		meanTempSum              float64
		rainTotal                float64
		windAvgSum               float64
		windCount                int
		windMax                  float64
		windMaxTime              string
		windDirSinSum            float64
		windDirCosSum            float64
		windDirCount             int
	}
	perDay := map[int]*dayAgg{}

	for rows.Next() {
		var epoch int64
		var outTemp, dew, hum, bar, windSpeed, windGust, windDir, rain, rainRate sql.NullFloat64
		if err := rows.Scan(&epoch, &outTemp, &dew, &hum, &bar, &windSpeed, &windGust, &windDir, &rain, &rainRate); err != nil {
			return "", err
		}
		d := time.Unix(epoch, 0).In(loc).Day()
		ts := time.Unix(epoch, 0).In(loc)
		timeStr := ts.Format("15:04")
		agg := perDay[d]
		if agg == nil {
			agg = &dayAgg{minTemp: 9999}
			perDay[d] = agg
		}
		if outTemp.Valid {
			agg.count++
			agg.meanTempSum += outTemp.Float64
			if outTemp.Float64 > agg.maxTemp {
				agg.maxTemp = outTemp.Float64
				agg.maxTempTime = timeStr
			}
			if outTemp.Float64 < agg.minTemp {
				agg.minTemp = outTemp.Float64
				agg.minTempTime = timeStr
			}
		}
		if rain.Valid {
			agg.rainTotal += rain.Float64
		}
		if windSpeed.Valid {
			agg.windAvgSum += windSpeed.Float64
			agg.windCount++
		}
		if windGust.Valid && windGust.Float64 > agg.windMax {
			agg.windMax = windGust.Float64
			agg.windMaxTime = timeStr
		}
		if windDir.Valid {
			// Convert to radians and track vector components
			radians := windDir.Float64 * math.Pi / 180.0
			agg.windDirSinSum += math.Sin(radians)
			agg.windDirCosSum += math.Cos(radians)
			agg.windDirCount++
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	// Check if any data exists for this month
	hasData := false
	for _, agg := range perDay {
		if agg != nil && agg.count > 0 {
			hasData = true
			break
		}
	}
	if !hasData {
		return "", fmt.Errorf("no data available for %s", start.Format("January 2006"))
	}

	// Compose text using provided monthly template
	// NOTE: This is a simplified renderer; values match units used in dashboard.
	monthName := start.Format("Jan 2006")
	header := fmt.Sprintf("MONTHLY CLIMATOLOGICAL SUMMARY for %s\n\n\nNAME: %s                  \nELEV: %.0f feet    LAT: %.2f N    LONG: %.2f W\n\n\n                   TEMPERATURE (F), RAIN (in), WIND SPEED (mph)\n\n                                         HEAT   COOL         AVG\n      MEAN                               DEG    DEG          WIND                   DOM\nDAY   TEMP   HIGH   TIME    LOW   TIME   DAYS   DAYS   RAIN  SPEED   HIGH   TIME    DIR\n---------------------------------------------------------------------------------------\n",
		monthName,
		appConfig.Location.Name,
		appConfig.Location.Altitude,
		appConfig.Location.Latitude,
		math.Abs(appConfig.Location.Longitude))

	lines := ""
	daysInMonth := int(end.Add(-24 * time.Hour).Day())

	// Track monthly totals for summary row
	var monthMeanSum, monthHighSum, monthLowSum, monthRainSum, monthWindAvgSum, monthWindMaxSum float64
	var monthDaysWithData, monthWindDays int
	var monthWindDirSinSum, monthWindDirCosSum float64
	var monthWindDirCount int

	for d := 1; d <= daysInMonth; d++ {
		agg := perDay[d]
		if agg == nil || agg.count == 0 {
			lines += fmt.Sprintf("%3d   --     --     --     --     --      --     --     --      --    --     --      --\n", d)
			continue
		}
		mean := agg.meanTempSum / float64(agg.count)
		avgWind := 0.0
		if agg.windCount > 0 {
			avgWind = agg.windAvgSum / float64(agg.windCount)
			monthWindAvgSum += avgWind
			monthWindDays++
		}
		domDir := 0
		if agg.windDirCount > 0 {
			// Vector average for wind direction
			avgSin := agg.windDirSinSum / float64(agg.windDirCount)
			avgCos := agg.windDirCosSum / float64(agg.windDirCount)
			domDirRad := math.Atan2(avgSin, avgCos)
			domDirDeg := domDirRad * 180.0 / math.Pi
			if domDirDeg < 0 {
				domDirDeg += 360
			}
			domDir = int(domDirDeg + 0.5)
			// Accumulate for monthly summary
			monthWindDirSinSum += avgSin
			monthWindDirCosSum += avgCos
			monthWindDirCount++
		}

		// Accumulate for summary
		monthMeanSum += mean
		monthHighSum += agg.maxTemp
		monthLowSum += agg.minTemp
		monthRainSum += agg.rainTotal
		monthWindMaxSum += agg.windMax
		monthDaysWithData++

		lines += fmt.Sprintf("%3d   %4.1f  %5.1f  %5s  %5.1f  %5s    --     --   %4.2f    %4.1f   %4.1f  %5s  %5d\n",
			d, mean, agg.maxTemp, agg.maxTempTime, agg.minTemp, agg.minTempTime, agg.rainTotal, avgWind, agg.windMax, agg.windMaxTime, domDir)
	}

	// Calculate summary row means
	summaryMean := monthMeanSum / float64(monthDaysWithData)
	summaryHigh := monthHighSum / float64(monthDaysWithData)
	summaryLow := monthLowSum / float64(monthDaysWithData)
	summaryWindAvg := 0.0
	if monthWindDays > 0 {
		summaryWindAvg = monthWindAvgSum / float64(monthWindDays)
	}
	summaryWindMax := monthWindMaxSum / float64(monthDaysWithData)
	summaryWindDir := "--"
	if monthWindDirCount > 0 {
		avgSin := monthWindDirSinSum / float64(monthWindDirCount)
		avgCos := monthWindDirCosSum / float64(monthWindDirCount)
		domDirRad := math.Atan2(avgSin, avgCos)
		domDirDeg := domDirRad * 180.0 / math.Pi
		if domDirDeg < 0 {
			domDirDeg += 360
		}
		summaryWindDir = fmt.Sprintf("%3d", int(domDirDeg+0.5))
	}

	footer := "---------------------------------------------------------------------------------------\n" +
		fmt.Sprintf("      %4.1f  %5.1f         %5.1f           --     --   %4.2f    %4.1f   %4.1f           %3s\n",
			summaryMean, summaryHigh, summaryLow, monthRainSum, summaryWindAvg, summaryWindMax, summaryWindDir)

	return header + lines + footer, nil
}

// RenderYearlyNOAA generates simplified yearly summary
func RenderYearlyNOAA(db *sql.DB, p NOAAYearlyParams) (string, error) {
	// Use MST7MDT timezone (Arizona local time, UTC-7)
	loc, err := time.LoadLocation("America/Phoenix")
	if err != nil {
		// Fallback to system local time if timezone not available
		loc = time.Now().Location()
	}
	start := time.Date(p.Year, 1, 1, 0, 0, 0, 0, loc)
	end := time.Date(p.Year+1, 1, 1, 0, 0, 0, 0, loc)
	rows, err := db.Query(`
		SELECT dateTime, outTemp, rain, windSpeed, windGust, windDir
		FROM archive
		WHERE dateTime >= ? AND dateTime < ?
		ORDER BY dateTime ASC
	`, start.Unix(), end.Unix())
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Monthly buckets
	type monAgg struct {
		maxTemp, minTemp float64
		rainTotal        float64
		windAvgSum       float64
		windCount        int
		windMax          float64
		domDirSum        float64
		domDirCount      int
	}
	months := make([]monAgg, 13)
	for i := range months {
		months[i].minTemp = 9999
	}

	for rows.Next() {
		var epoch int64
		var outTemp, rain, windSpeed, windGust, windDir sql.NullFloat64
		if err := rows.Scan(&epoch, &outTemp, &rain, &windSpeed, &windGust, &windDir); err != nil {
			return "", err
		}
		m := int(time.Unix(epoch, 0).In(loc).Month())
		agg := &months[m]
		if outTemp.Valid {
			if outTemp.Float64 > agg.maxTemp {
				agg.maxTemp = outTemp.Float64
			}
			if outTemp.Float64 < agg.minTemp {
				agg.minTemp = outTemp.Float64
			}
		}
		if rain.Valid {
			agg.rainTotal += rain.Float64
		}
		if windSpeed.Valid {
			agg.windAvgSum += windSpeed.Float64
			agg.windCount++
		}
		if windGust.Valid && windGust.Float64 > agg.windMax {
			agg.windMax = windGust.Float64
		}
		if windDir.Valid {
			agg.domDirSum += windDir.Float64
			agg.domDirCount++
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	// Check if any data exists for this year
	hasData := false
	for m := 1; m <= 12; m++ {
		if months[m].windCount > 0 {
			hasData = true
			break
		}
	}
	if !hasData {
		return "", fmt.Errorf("no data available for year %d", p.Year)
	}

	header := fmt.Sprintf("CLIMATOLOGICAL SUMMARY for year %d\n\n\nNAME: %s                  \nELEV: %.0f feet    LAT: %.2f N    LONG: %.2f W\n\n\n                                       TEMPERATURE (F)\n\n                              HEAT    COOL                              MAX    MAX    MIN    MIN\n          MEAN   MEAN         DEG     DEG                                >=     <=     <=     <=\n YR  MO   MAX    MIN    MEAN  DAYS    DAYS      HI  DAY     LOW  DAY     90     32     32      0\n------------------------------------------------------------------------------------------------\n",
		p.Year,
		appConfig.Location.Name,
		appConfig.Location.Altitude,
		appConfig.Location.Latitude,
		math.Abs(appConfig.Location.Longitude))
	lines := ""

	// Track yearly totals for summary row
	var yearMaxSum, yearMinSum, yearMeanSum, yearRainSum, yearWindMaxSum float64
	var yearMonthsWithData int

	for m := 1; m <= 12; m++ {
		agg := months[m]
		if agg.windCount == 0 {
			lines += fmt.Sprintf("%4d %02d\n", p.Year, m)
			continue
		}
		meanMax := agg.maxTemp
		meanMin := agg.minTemp
		mean := (meanMax + meanMin) / 2.0

		// Accumulate for summary
		yearMaxSum += meanMax
		yearMinSum += meanMin
		yearMeanSum += mean
		yearRainSum += agg.rainTotal
		yearWindMaxSum += agg.windMax
		yearMonthsWithData++

		lines += fmt.Sprintf("%4d %02d  %5.1f  %5.1f   %5.1f    0.0    0.0   %6.1f   --    %6.1f   --      --      --      --      --\n",
			p.Year, m, meanMax, meanMin, mean, agg.rainTotal, agg.windMax)
	}

	// Calculate summary row means
	summaryMax := yearMaxSum / float64(yearMonthsWithData)
	summaryMin := yearMinSum / float64(yearMonthsWithData)
	summaryMean := yearMeanSum / float64(yearMonthsWithData)
	summaryWindMax := yearWindMaxSum / float64(yearMonthsWithData)

	footer := "------------------------------------------------------------------------------------------------\n" +
		fmt.Sprintf("          %5.1f  %5.1f   %5.1f    0.0    0.0   %6.1f   --    %6.1f   --      --      --      --      --\n\n\n                  PRECIPITATION (in)\n\n                  MAX         ---DAYS OF RAIN---\n                  OBS.               OVER\n YR  MO  TOTAL    DAY  DATE   0.01   0.10   1.00\n------------------------------------------------\n          %0.2f     --     --     --      --      --\n\n\n           WIND SPEED (mph)\n\n                                DOM\n YR  MO    AVG     HI   DATE    DIR\n-----------------------------------\n          --     %4.1f     --     --\n",
			summaryMax, summaryMin, summaryMean, yearRainSum, summaryWindMax, yearRainSum, summaryWindMax)
	return header + lines + footer, nil
}

// SaveTextFile ensures directory and writes content; returns path
func SaveTextFile(relPath string, content string) (string, error) {
	abs := filepath.Join("static", relPath)
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(abs, []byte(content), 0644); err != nil {
		return "", err
	}
	return abs, nil
}

// GetOrGenerateMonthly returns file content; generates if missing or if force=true
func GetOrGenerateMonthly(db *sql.DB, p NOAAMonthlyParams, force bool) (string, error) {
	filename := fmt.Sprintf("noaa/NOAA-%04d-%02d.txt", p.Year, p.Month)
	abs := filepath.Join("static", filename)

	// If force=true, delete cached file
	if force {
		os.Remove(abs)
	}

	if b, err := os.ReadFile(abs); err == nil {
		return string(b), nil
	}
	content, err := RenderMonthlyNOAA(db, p)
	if err != nil {
		return "", err
	}
	if _, err := SaveTextFile(filename, content); err != nil {
		log.Println("Save monthly NOAA failed:", err)
	}
	return content, nil
}

// GetOrGenerateYearly returns file content; generates if missing or if force=true
func GetOrGenerateYearly(db *sql.DB, p NOAAYearlyParams, force bool) (string, error) {
	filename := fmt.Sprintf("noaa/NOAA-%04d.txt", p.Year)
	abs := filepath.Join("static", filename)

	// If force=true, delete cached file
	if force {
		os.Remove(abs)
	}

	if b, err := os.ReadFile(abs); err == nil {
		return string(b), nil
	}
	content, err := RenderYearlyNOAA(db, p)
	if err != nil {
		return "", err
	}
	if _, err := SaveTextFile(filename, content); err != nil {
		log.Println("Save yearly NOAA failed:", err)
	}
	return content, nil
}
