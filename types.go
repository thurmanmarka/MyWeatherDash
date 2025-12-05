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

type CelestialData struct {
	Date          string     `json:"date"`                    // YYYY-MM-DD
	Timezone      string     `json:"timezone"`                // e.g. "America/Phoenix"
	Sunrise       *time.Time `json:"sunrise,omitempty"`       // Local time
	Sunset        *time.Time `json:"sunset,omitempty"`        // Local time
	Sunrise24     string     `json:"sunrise24,omitempty"`     // Local time formatted HH:MM (24h)
	Sunset24      string     `json:"sunset24,omitempty"`      // Local time formatted HH:MM (24h)
	DaylightHours float64    `json:"daylightHours,omitempty"` // Hours of daylight (sunrise to sunset)
	Moonrise      *time.Time `json:"moonrise,omitempty"`      // Local time
	Moonset       *time.Time `json:"moonset,omitempty"`       // Local time
	Moonrise24    string     `json:"moonrise24,omitempty"`
	Moonset24     string     `json:"moonset24,omitempty"`
	MoonPhase     *MoonPhase `json:"moonPhase,omitempty"` // Current phase

	// Twilight times (civil dawn/dusk at -6°)
	CivilDawn   *time.Time `json:"civilDawn,omitempty"`
	CivilDusk   *time.Time `json:"civilDusk,omitempty"`
	CivilDawn24 string     `json:"civilDawn24,omitempty"`
	CivilDusk24 string     `json:"civilDusk24,omitempty"`

	// Nautical twilight (-12°)
	NauticalDawn   *time.Time `json:"nauticalDawn,omitempty"`
	NauticalDusk   *time.Time `json:"nauticalDusk,omitempty"`
	NauticalDawn24 string     `json:"nauticalDawn24,omitempty"`
	NauticalDusk24 string     `json:"nauticalDusk24,omitempty"`

	// Astronomical twilight (-18°)
	AstronomicalDawn   *time.Time `json:"astronomicalDawn,omitempty"`
	AstronomicalDusk   *time.Time `json:"astronomicalDusk,omitempty"`
	AstronomicalDawn24 string     `json:"astronomicalDawn24,omitempty"`
	AstronomicalDusk24 string     `json:"astronomicalDusk24,omitempty"`

	// Golden hour windows (Sun altitude -4° to +6°)
	GoldenHourMorningStart   *time.Time `json:"goldenHourMorningStart,omitempty"`
	GoldenHourMorningEnd     *time.Time `json:"goldenHourMorningEnd,omitempty"`
	GoldenHourEveningStart   *time.Time `json:"goldenHourEveningStart,omitempty"`
	GoldenHourEveningEnd     *time.Time `json:"goldenHourEveningEnd,omitempty"`
	GoldenHourMorningStart24 string     `json:"goldenHourMorningStart24,omitempty"`
	GoldenHourMorningEnd24   string     `json:"goldenHourMorningEnd24,omitempty"`
	GoldenHourEveningStart24 string     `json:"goldenHourEveningStart24,omitempty"`
	GoldenHourEveningEnd24   string     `json:"goldenHourEveningEnd24,omitempty"`

	// Blue hour windows (Sun altitude -6° to -4°)
	BlueHourMorningStart   *time.Time `json:"blueHourMorningStart,omitempty"`
	BlueHourMorningEnd     *time.Time `json:"blueHourMorningEnd,omitempty"`
	BlueHourEveningStart   *time.Time `json:"blueHourEveningStart,omitempty"`
	BlueHourEveningEnd     *time.Time `json:"blueHourEveningEnd,omitempty"`
	BlueHourMorningStart24 string     `json:"blueHourMorningStart24,omitempty"`
	BlueHourMorningEnd24   string     `json:"blueHourMorningEnd24,omitempty"`
	BlueHourEveningStart24 string     `json:"blueHourEveningStart24,omitempty"`
	BlueHourEveningEnd24   string     `json:"blueHourEveningEnd24,omitempty"`
}

type MoonPhase struct {
	Fraction   float64 `json:"fraction"`   // 0.0 = new, 1.0 = full
	Elongation float64 `json:"elongation"` // degrees
	Waxing     bool    `json:"waxing"`     // true if waxing
	Name       string  `json:"name"`       // e.g. "Waxing Gibbous"
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
