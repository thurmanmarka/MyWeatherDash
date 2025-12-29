# MyWeatherDash: AI Agent Instructions

## Project Overview
MyWeatherDash is a real-time weather dashboard combining a Go backend API with a dynamic Chart.js frontend. It reads historical and live weather data from WeeWX's MariaDB archive and presents it through synchronized time-indexed charts with current conditions display.

**Key Architecture**: WeeWX Database → Go API (JSON) → HTML/Chart.js Frontend

## Big Picture Architecture

### Backend (Go)
- **Entry point**: `main.go` — loads config, initializes MariaDB connection, registers HTTP routes
- **Data handlers**: `handlers.go` — 11 `/api/*` endpoints, each queries the WeeWX `archive` table by timestamp range
- **Types**: `types.go` — JSON-serializable reading structs (e.g., `WeatherReading`, `WindReading`)
- **Config**: `config.go` — YAML-based database credentials loader using DSN builder

**Critical pattern**: All handlers follow identical flow:
1. Extract `range` query param (defaults to `day`, supports `week`/`month`)
2. Query `archive` table with `WHERE dateTime >= ?` (unix epoch comparison)
3. Map rows to `time.Time` structs, JSON encode response

**Database schema assumption**: WeeWX archive table with columns: `dateTime`, `outTemp`, `dewpoint`, `barometer`, `windSpeed`, `windGust`, `windDir`, `rainRate`, `rain`, `heatindex`, `windchill`, `outHumidity`, `inTemp`, `inHumidity`, `lightning_strike_count`

### Frontend (JavaScript)
- **core.js** (591 lines): Global chart instances, master time grid system, data alignment, time formatting
- **charts.js** (1310 lines): Individual chart loaders and renderers using Chart.js
- **index.html**: Sidebar for current conditions + responsive grid of chart panels

**Critical pattern - Master Time Grid**: 
- `/api/weather` response defines `masterTimes[]` array (shared across all charts)
- All other endpoints align their data to this grid using `alignToMasterTimes()` helper
- This ensures perfect synchronization of x-axes across 12 different charts
- Null values inserted where sparse datasets don't have readings at master times

## Essential Integration Points

### Data Flow: Range Selection
1. User clicks "Day"/"Week"/"Month" button → updates `currentRange` global
2. Calls `loadAllCharts()` which sequentially loads all 11 API endpoints
3. Each endpoint receives `?range=day|week|month` as query param
4. Go handler converts to duration: `day=24h`, `week=7*24h`, `month=30*24h`
5. Time comparison: `WHERE dateTime >= NOW() - duration` (unix timestamps)

### Data Flow: Chart Synchronization
1. `loadWeather()` fetches `/api/weather` first, establishes `masterTimes` grid
2. `updateCurrentConditions()` pulls latest values from all loaded data
3. Time ticks generated via `makeTimeTickOptions(times)` — adaptive density (4-hour marks for day view)
4. Day/night background shading plugin activates only for day range

### Sparse Data Handling
- Rain, lightning, inside sensors may have NULL/missing readings
- `sql.NullFloat64` used in `handleWind()` to distinguish NULL from zero
- Frontend inserts null values in aligned arrays, Chart.js renders breaks in lines

## Critical Developer Workflows

### Build & Run
```powershell
# Development (watch mode not built-in)
go run .

# Production build
go build -o weatherdash
./weatherdash
```
Server listens on `http://localhost:8080` — static files served from `./static/`, template from `./templates/index.html`

### Database Connection
1. **No setup script provided** — assumes WeeWX MariaDB already running
2. Copy `config.example.yaml` → `config.yaml`
3. Edit credentials (user, password, host, port, name="weewx")
4. Connection verified via `db.Ping()` on startup; fails fatally if unreachable

### Configuration
- YAML-based, only DB credentials configurable (no server port, static paths, chart options in config)
- `config.yaml` gitignored to protect credentials
- `buildDSN()` formats MySQL connection string with optional params (e.g., `parseTime=false`)

## Project-Specific Patterns & Conventions

### Handler Pattern (all 11 endpoints identical structure)
```go
func handleWeather(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    dur := getRangeDuration(r)
    since := time.Now().Add(-dur).Unix()
    rows, err := db.Query("SELECT ... FROM archive WHERE dateTime >= ? ORDER BY dateTime ASC", since)
    // scan rows → struct → JSON encode
}
```
**When adding new endpoints**: Follow this exact pattern, query archive table with timestamp filter, return time-sorted ascending array.

### Frontend Module Organization
- **Global state**: Chart instances, latest readings, active alerts (rain/lightning/wind)
- **No event system**: Direct function calls (`loadAllCharts()`, `updateCurrentConditions()`)
- **Data arrays**: Always parallel (same indices must align) — `masterTimes[i]`, `temps[i]`, `dews[i]`

### Units & Conventions
- **Temperature**: Fahrenheit (°F)
- **Pressure**: inHg or mbar (not normalized in code; preserve WeeWX values)
- **Wind**: mph, direction 0-359 degrees (0°=N, 90°=E)
- **Rain**: inches
- **Lightning**: strike count (integer)
- **Timestamps**: JSON serialized as RFC3339 (Go `time.Time` marshaling), frontend converts to Date objects

### Styling
- CSS variables: `--accent` (#2563eb blue), `--panel-bg`, `--text-main`, `--text-muted`
- Responsive: flex layout, sidebar→main-grid stacks on <900px
- Icons: Minimal inline SVG (wind, lightning, rain droplets)
- Current conditions: `.cc-row` grid format with label/value pairs

## Common Pitfalls & Patterns to Avoid

1. **Timestamp misalignment**: Always use `masterTimes` established by `/api/weather`; don't mix local dates
2. **NULL handling**: Use `sql.NullFloat64` for optional columns (wind direction, inside sensors); check `.Valid` before use
3. **Chart destruction**: Always call `chart.destroy()` before reassigning to prevent memory leaks
4. **Range query consistency**: All 11 handlers must honor same duration logic; maintain parity in `getRangeDuration()`
5. **JSON encoding errors**: Handlers encode directly to `http.ResponseWriter` via `json.NewEncoder()` — errors after write headers fail silently; prefer encoding to buffer first for complex responses
6. **Database pooling**: Default `sql.Open()` pool reused across all handlers; no explicit connection management needed

## Integration with WeeWX Archive
- **Archive schema**: Fixed by WeeWX; all queries assume standard column names and unix epoch timestamps
- **Data freshness**: Depends on WeeWX update interval (typically 5-min for observation records)
- **Historical queries**: No aggregation/downsampling — returns raw intervals (e.g., 288 daily readings if 5-min archive)
- **Timezone**: Timestamps stored as UTC unix epochs in MariaDB; frontend converts via JavaScript Date objects (local timezone)

## File Reference Guide
- **Backend logic**: `main.go`, `handlers.go` (core API)
- **Frontend logic**: `static/js/core.js` (state, grid alignment), `static/js/charts.js` (rendering)
- **UI layout**: `templates/index.html` (structure, styling, canvas elements)
- **Types**: `types.go` (JSON contracts), `config.go` (YAML loading)
