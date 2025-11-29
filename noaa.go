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
		heatDegDays              float64
		coolDegDays              float64
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
	var monthHeatDegDays, monthCoolDegDays float64

	for d := 1; d <= daysInMonth; d++ {
		agg := perDay[d]
		if agg == nil || agg.count == 0 {
			lines += fmt.Sprintf("%3d     --     --     --     --     --    --     --     --      --     --     --     --\n", d)
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

		// Calculate degree days (base 65°F)
		dailyAvgTemp := (agg.maxTemp + agg.minTemp) / 2.0
		if dailyAvgTemp < 65.0 {
			agg.heatDegDays = 65.0 - dailyAvgTemp
		}
		if dailyAvgTemp > 65.0 {
			agg.coolDegDays = dailyAvgTemp - 65.0
		}

		// Accumulate for summary
		monthMeanSum += mean
		monthHighSum += agg.maxTemp
		monthLowSum += agg.minTemp
		monthRainSum += agg.rainTotal
		monthWindMaxSum += agg.windMax
		monthDaysWithData++
		monthHeatDegDays += agg.heatDegDays
		monthCoolDegDays += agg.coolDegDays

		lines += fmt.Sprintf("%3d   %4.1f  %5.1f  %5s  %5.1f  %5s  %4.0f   %4.0f   %4.2f    %4.1f   %4.1f  %5s  %5d\n",
			d, mean, agg.maxTemp, agg.maxTempTime, agg.minTemp, agg.minTempTime, agg.heatDegDays, agg.coolDegDays, agg.rainTotal, avgWind, agg.windMax, agg.windMaxTime, domDir)
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
		fmt.Sprintf("      %4.1f  %5.1f         %5.1f         %4.0f   %4.0f   %4.2f    %4.1f   %4.1f           %3s\n",
			summaryMean, summaryHigh, summaryLow, monthHeatDegDays, monthCoolDegDays, monthRainSum, summaryWindAvg, summaryWindMax, summaryWindDir)

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

	// First pass: aggregate by day to get daily highs and lows
	type dayAgg struct {
		maxTemp, minTemp float64
		count            int
		rainTotal        float64
		windAvgSum       float64
		windCount        int
		windMax          float64
		windDirSinSum    float64
		windDirCosSum    float64
		windDirCount     int
	}
	// Map key: "YYYYMMDD" (year, month, day combined)
	perDay := map[string]*dayAgg{}

	for rows.Next() {
		var epoch int64
		var outTemp, rain, windSpeed, windGust, windDir sql.NullFloat64
		if err := rows.Scan(&epoch, &outTemp, &rain, &windSpeed, &windGust, &windDir); err != nil {
			return "", err
		}
		ts := time.Unix(epoch, 0).In(loc)
		dayKey := fmt.Sprintf("%04d%02d%02d", ts.Year(), ts.Month(), ts.Day())

		agg := perDay[dayKey]
		if agg == nil {
			agg = &dayAgg{minTemp: 9999}
			perDay[dayKey] = agg
		}

		if outTemp.Valid {
			agg.count++
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
			// Convert to radians and track vector components
			radians := windDir.Float64 * math.Pi / 180.0
			agg.windDirSinSum += math.Sin(radians)
			agg.windDirCosSum += math.Cos(radians)
			agg.windDirCount++
		}
	}

	// Second pass: aggregate daily values into monthly buckets
	type monAgg struct {
		maxTempSum    float64 // Sum of daily highs
		minTempSum    float64 // Sum of daily lows
		daysWithData  int     // Number of days with temp data
		rainTotal     float64
		windMax       float64
		windMaxDay    int     // Day of max wind gust
		windAvgSum    float64 // Sum of daily average wind speeds
		windDays      int     // Days with wind data
		windDirSinSum float64 // Vector sum for direction
		windDirCosSum float64 // Vector sum for direction
		windDirCount  int     // Count for direction averaging
		hiTemp        float64 // Highest temperature in month
		hiTempDay     int     // Day of highest temperature
		lowTemp       float64 // Lowest temperature in month
		lowTempDay    int     // Day of lowest temperature
		daysMaxGE90   int     // Days with max temp >= 90°F
		daysMaxLE32   int     // Days with max temp <= 32°F
		daysMinLE32   int     // Days with min temp <= 32°F
		daysMinLE0    int     // Days with min temp <= 0°F
		maxDailyRain  float64 // Highest daily rainfall
		maxRainDay    int     // Day of highest rainfall
		rainDaysGE01  int     // Days with >= 0.01 inches
		rainDaysGE10  int     // Days with >= 0.10 inches
		rainDaysGE100 int     // Days with >= 1.00 inches
		heatDegDays   float64 // Heating degree days (base 65°F)
		coolDegDays   float64 // Cooling degree days (base 65°F)
	}
	months := make([]monAgg, 13)
	for i := range months {
		months[i].lowTemp = 9999 // Initialize to high value
	}

	for dayKey, dayData := range perDay {
		if dayData.count == 0 {
			continue
		}
		// Extract month and day from dayKey (format: YYYYMMDD)
		m := int(dayKey[4]-'0')*10 + int(dayKey[5]-'0')
		d := int(dayKey[6]-'0')*10 + int(dayKey[7]-'0')

		agg := &months[m]
		agg.maxTempSum += dayData.maxTemp
		agg.minTempSum += dayData.minTemp
		agg.daysWithData++
		agg.rainTotal += dayData.rainTotal
		if dayData.windMax > agg.windMax {
			agg.windMax = dayData.windMax
			agg.windMaxDay = d
		}
		// Track average wind speed
		if dayData.windCount > 0 {
			dailyAvgWind := dayData.windAvgSum / float64(dayData.windCount)
			agg.windAvgSum += dailyAvgWind
			agg.windDays++
		}
		// Track wind direction vectors
		if dayData.windDirCount > 0 {
			avgSin := dayData.windDirSinSum / float64(dayData.windDirCount)
			avgCos := dayData.windDirCosSum / float64(dayData.windDirCount)
			agg.windDirSinSum += avgSin
			agg.windDirCosSum += avgCos
			agg.windDirCount++
		}
		// Track highest temperature for the month
		if dayData.maxTemp > agg.hiTemp {
			agg.hiTemp = dayData.maxTemp
			agg.hiTempDay = d
		}
		// Track lowest temperature for the month
		if dayData.minTemp < agg.lowTemp {
			agg.lowTemp = dayData.minTemp
			agg.lowTempDay = d
		}
		// Track max daily rainfall
		if dayData.rainTotal > agg.maxDailyRain {
			agg.maxDailyRain = dayData.rainTotal
			agg.maxRainDay = d
		}
		// Count rain days by threshold
		if dayData.rainTotal >= 0.01 {
			agg.rainDaysGE01++
		}
		if dayData.rainTotal >= 0.10 {
			agg.rainDaysGE10++
		}
		if dayData.rainTotal >= 1.00 {
			agg.rainDaysGE100++
		}
		// Count temperature threshold days
		if dayData.maxTemp >= 90.0 {
			agg.daysMaxGE90++
		}
		if dayData.maxTemp <= 32.0 {
			agg.daysMaxLE32++
		}
		if dayData.minTemp <= 32.0 {
			agg.daysMinLE32++
		}
		if dayData.minTemp <= 0.0 {
			agg.daysMinLE0++
		}
		// Calculate degree days (base 65°F)
		dailyAvgTemp := (dayData.maxTemp + dayData.minTemp) / 2.0
		if dailyAvgTemp < 65.0 {
			agg.heatDegDays += 65.0 - dailyAvgTemp
		}
		if dailyAvgTemp > 65.0 {
			agg.coolDegDays += dailyAvgTemp - 65.0
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	// Check if any data exists for this year
	hasData := false
	for m := 1; m <= 12; m++ {
		if months[m].daysWithData > 0 {
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
	var yearDaysMaxGE90, yearDaysMaxLE32, yearDaysMinLE32, yearDaysMinLE0 int
	var yearHiTemp float64
	var yearLowTemp float64 = 9999
	var yearMaxDailyRain float64
	var yearRainDaysGE01, yearRainDaysGE10, yearRainDaysGE100 int
	var yearWindAvgSum float64
	var yearWindMonths int
	var yearWindDirSinSum, yearWindDirCosSum float64
	var yearWindDirCount int
	var yearHeatDegDays, yearCoolDegDays float64

	for m := 1; m <= 12; m++ {
		agg := months[m]
		if agg.daysWithData == 0 {
			lines += fmt.Sprintf("%4d %02d     --     --      --   --     --      --   --      --   --      --     --     --     --\n", p.Year, m)
			continue
		}
		// MEAN MAX = average of all daily high temperatures
		meanMax := agg.maxTempSum / float64(agg.daysWithData)
		// MEAN MIN = average of all daily low temperatures
		meanMin := agg.minTempSum / float64(agg.daysWithData)
		// MEAN = average of MEAN MAX and MEAN MIN
		mean := (meanMax + meanMin) / 2.0

		// Accumulate for summary
		yearMaxSum += meanMax
		yearMinSum += meanMin
		yearMeanSum += mean
		yearRainSum += agg.rainTotal
		yearWindMaxSum += agg.windMax
		yearMonthsWithData++
		yearDaysMaxGE90 += agg.daysMaxGE90
		yearDaysMaxLE32 += agg.daysMaxLE32
		yearDaysMinLE32 += agg.daysMinLE32
		yearDaysMinLE0 += agg.daysMinLE0
		yearRainDaysGE01 += agg.rainDaysGE01
		yearRainDaysGE10 += agg.rainDaysGE10
		yearRainDaysGE100 += agg.rainDaysGE100
		yearHeatDegDays += agg.heatDegDays
		yearCoolDegDays += agg.coolDegDays
		if agg.windDays > 0 {
			monthlyAvgWind := agg.windAvgSum / float64(agg.windDays)
			yearWindAvgSum += monthlyAvgWind
			yearWindMonths++
		}
		if agg.windDirCount > 0 {
			avgSin := agg.windDirSinSum / float64(agg.windDirCount)
			avgCos := agg.windDirCosSum / float64(agg.windDirCount)
			yearWindDirSinSum += avgSin
			yearWindDirCosSum += avgCos
			yearWindDirCount++
		}
		if agg.hiTemp > yearHiTemp {
			yearHiTemp = agg.hiTemp
		}
		if agg.lowTemp < yearLowTemp {
			yearLowTemp = agg.lowTemp
		}
		if agg.maxDailyRain > yearMaxDailyRain {
			yearMaxDailyRain = agg.maxDailyRain
		}

		lines += fmt.Sprintf("%4d %02d  %5.1f  %5.1f   %5.1f  %3.0f   %3.0f   %5.1f  %3d   %5.1f  %3d     %3d    %3d    %3d    %3d\n",
			p.Year, m, meanMax, meanMin, mean, agg.heatDegDays, agg.coolDegDays, agg.hiTemp, agg.hiTempDay, agg.lowTemp, agg.lowTempDay,
			agg.daysMaxGE90, agg.daysMaxLE32, agg.daysMinLE32, agg.daysMinLE0)
	}

	// Calculate summary row means
	summaryMax := yearMaxSum / float64(yearMonthsWithData)
	summaryMin := yearMinSum / float64(yearMonthsWithData)
	summaryMean := yearMeanSum / float64(yearMonthsWithData)
	summaryWindMax := yearWindMaxSum / float64(yearMonthsWithData)

	footer := "------------------------------------------------------------------------------------------------\n" +
		fmt.Sprintf("          %5.1f  %5.1f   %5.1f  %3.0f   %3.0f   %5.1f  --    %5.1f  --      %2d     %2d     %2d     %2d\n\n\n                  PRECIPITATION (in)\n\n                  MAX         ---DAYS OF RAIN---\n                  OBS.               OVER\n YR  MO  TOTAL    DAY  DATE   0.01   0.10   1.00\n------------------------------------------------\n",
			summaryMax, summaryMin, summaryMean, yearHeatDegDays, yearCoolDegDays, yearHiTemp, yearLowTemp,
			yearDaysMaxGE90, yearDaysMaxLE32, yearDaysMinLE32, yearDaysMinLE0)

	// Add monthly precipitation rows
	for m := 1; m <= 12; m++ {
		agg := months[m]
		if agg.daysWithData == 0 {
			footer += fmt.Sprintf("%4d %02d    --      --   --     --     --    --\n", p.Year, m)
			continue
		}
		footer += fmt.Sprintf("%4d %02d %5.2f   %5.2f   %2d     %2d     %2d    %2d\n",
			p.Year, m, agg.rainTotal, agg.maxDailyRain, agg.maxRainDay,
			agg.rainDaysGE01, agg.rainDaysGE10, agg.rainDaysGE100)
	}

	footer += "------------------------------------------------\n" +
		fmt.Sprintf("        %5.2f   %5.2f          %2d     %2d    %2d\n\n\n           WIND SPEED (mph)\n\n                                DOM\n YR  MO    AVG     HI   DATE    DIR\n-----------------------------------\n",
			yearRainSum, yearMaxDailyRain,
			yearRainDaysGE01, yearRainDaysGE10, yearRainDaysGE100)

	// Add monthly wind rows
	for m := 1; m <= 12; m++ {
		agg := months[m]
		if agg.daysWithData == 0 {
			footer += fmt.Sprintf("%4d %02d     --     --     --    --\n", p.Year, m)
			continue
		}
		monthlyAvgWind := 0.0
		if agg.windDays > 0 {
			monthlyAvgWind = agg.windAvgSum / float64(agg.windDays)
		}
		domDir := "--"
		if agg.windDirCount > 0 {
			avgSin := agg.windDirSinSum / float64(agg.windDirCount)
			avgCos := agg.windDirCosSum / float64(agg.windDirCount)
			domDirRad := math.Atan2(avgSin, avgCos)
			domDirDeg := domDirRad * 180.0 / math.Pi
			if domDirDeg < 0 {
				domDirDeg += 360
			}
			domDir = fmt.Sprintf("%3d", int(domDirDeg+0.5))
		}
		footer += fmt.Sprintf("%4d %02d  %5.1f  %5.1f     %2d   %3s\n",
			p.Year, m, monthlyAvgWind, agg.windMax, agg.windMaxDay, domDir)
	}

	// Calculate yearly wind summary
	yearlyAvgWind := 0.0
	if yearWindMonths > 0 {
		yearlyAvgWind = yearWindAvgSum / float64(yearWindMonths)
	}
	yearlyDomDir := "--"
	if yearWindDirCount > 0 {
		avgSin := yearWindDirSinSum / float64(yearWindDirCount)
		avgCos := yearWindDirCosSum / float64(yearWindDirCount)
		domDirRad := math.Atan2(avgSin, avgCos)
		domDirDeg := domDirRad * 180.0 / math.Pi
		if domDirDeg < 0 {
			domDirDeg += 360
		}
		yearlyDomDir = fmt.Sprintf("%3d", int(domDirDeg+0.5))
	}

	footer += "-----------------------------------\n" +
		fmt.Sprintf("         %5.1f  %5.1f          %3s\n",
			yearlyAvgWind, summaryWindMax, yearlyDomDir)
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
