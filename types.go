package main

import "time"

type PingResponse struct {
	Message string `json:"message"`
}

type BarometerReading struct {
	Timestamp time.Time `json:"timestamp"`
	Pressure  float64   `json:"pressure"` // inHg or mbar
	// Computed fields (only on latest reading)
	Trend    string `json:"trend,omitempty"`    // rapid-rise, slow-rise, steady, slow-fall, rapid-fall
	Level    string `json:"level,omitempty"`    // high, normal, low
	Forecast string `json:"forecast,omitempty"` // weather forecast based on trend+level
}

type WeatherReading struct {
	Timestamp   time.Time `json:"timestamp"`
	Temperature float64   `json:"temperature"`
	Dewpoint    float64   `json:"dewpoint"`
}

type FeelsLikeReading struct {
	Timestamp time.Time `json:"timestamp"`
	HeatIndex float64   `json:"heatIndex"`
	WindChill float64   `json:"windChill"`
	// Computed fields (only on latest reading)
	ActiveValue  float64 `json:"activeValue,omitempty"`  // The chosen feels-like value
	ActiveSource string  `json:"activeSource,omitempty"` // "heat", "chill", or "air"
	ActiveLabel  string  `json:"activeLabel,omitempty"`  // "Heat Index", "Wind Chill", or "Air Temp"
}

type HumidityReading struct {
	Timestamp time.Time `json:"timestamp"`
	Humidity  float64   `json:"humidity"`
}

type WindReading struct {
	Timestamp time.Time `json:"timestamp"`
	Speed     float64   `json:"speed"`     // windSpeed
	Gust      float64   `json:"gust"`      // windGust
	Direction *float64  `json:"direction"` // windDir (degrees), nil for calm/unknown
	// Computed fields (only on latest reading)
	Compass string `json:"compass,omitempty"` // N, NNE, NE, etc.
	Strong  bool   `json:"strong,omitempty"`  // true if speed >= 20 mph or gust >= 25 mph
}

type RainReading struct {
	Timestamp time.Time `json:"timestamp"`
	Rate      float64   `json:"rate"`   // rainRate
	Amount    float64   `json:"amount"` // rain (interval or total)
	// Computed fields (only on latest reading)
	RecentlyActive bool `json:"recentlyActive,omitempty"` // true if rain in last 10 minutes
}

type LightningReading struct {
	Timestamp time.Time `json:"timestamp"`
	Strikes   float64   `json:"strikes"`
	// Computed fields (only on latest reading)
	RecentlyActive bool `json:"recentlyActive,omitempty"` // true if lightning in last 10 minutes
}

type InsideTemperature struct {
	Timestamp   time.Time `json:"timestamp"`
	InsideTempF float64   `json:"inside_temp_f"`
}

type InsideHumidityReading struct {
	Timestamp      time.Time `json:"timestamp"`
	InsideHumidity float64   `json:"inside_humidity"`
}

// StatisticsData holds aggregated metrics for today and the selected range
type StatisticsData struct {
	// Rain
	RainToday float64 `json:"rainToday"`
	RainRange float64 `json:"rainRange"`

	// Lightning
	StrikesToday int `json:"strikesToday"`
	StrikesRange int `json:"strikesRange"`

	// Temperature (hi/lo format: "75 / 62")
	TempToday string `json:"tempToday"`
	TempRange string `json:"tempRange"`

	// Feels Like (hi/lo format: "78 / 60")
	FeelsToday string `json:"feelsToday"`
	FeelsRange string `json:"feelsRange"`

	// Windchill (low value only)
	WindchillToday string `json:"windchillToday"`
	WindchillRange string `json:"windchillRange"`

	// Dewpoint (hi/lo format)
	DewToday string `json:"dewToday"`
	DewRange string `json:"dewRange"`

	// Humidity (hi/lo format)
	HumidityToday string `json:"humidityToday"`
	HumidityRange string `json:"humidityRange"`

	// Barometer (hi/lo format: "30.15 / 29.85")
	BarometerToday string `json:"barometerToday"`
	BarometerRange string `json:"barometerRange"`

	// Wind Average
	WindAvgToday string `json:"windAvgToday"`
	WindAvgRange string `json:"windAvgRange"`

	// Wind Max (with direction: "35 • 270")
	WindMaxToday string `json:"windMaxToday"`
	WindMaxRange string `json:"windMaxRange"`

	// Wind RMS
	WindRmsToday string `json:"windRmsToday"`
	WindRmsRange string `json:"windRmsRange"`

	// Wind Vector Speed
	WindVectorToday string `json:"windVectorToday"`
	WindVectorRange string `json:"windVectorRange"`

	// Wind Vector Direction
	WindVectorDirToday string `json:"windVectorDirToday"`
	WindVectorDirRange string `json:"windVectorDirRange"`

	// Rain Rate (max)
	RainRateToday string `json:"rainRateToday"`
	RainRateRange string `json:"rainRateRange"`

	// Lightning Distance (placeholder)
	LightningDistToday string `json:"lightningDistToday"`
	LightningDistRange string `json:"lightningDistRange"`

	// Inside Temperature (hi/lo format)
	InsideTempToday string `json:"insideTempToday"`
	InsideTempRange string `json:"insideTempRange"`

	// Inside Humidity (hi/lo format)
	InsideHumToday string `json:"insideHumToday"`
	InsideHumRange string `json:"insideHumRange"`
}
