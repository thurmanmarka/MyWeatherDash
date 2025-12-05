// ---------------------------------------------------------------------
// Global chart instances
// ---------------------------------------------------------------------
let weatherChart = null;
let barometerChart = null;
let feelsLikeChart = null;
let humidityChart = null;
let windChart = null;
let windDirChart = null;
let windVectorChart = null;
let rainAmountChart = null;
let rainRateChart = null;
let lightningChart = null;
let insideTempChart = null;
let insideHumidityChart = null;

// Function to refresh all charts (useful when theme changes)
function refreshAllCharts() {
    const charts = [
        weatherChart, barometerChart, feelsLikeChart, humidityChart,
        windChart, windDirChart, windVectorChart, rainAmountChart,
        rainRateChart, lightningChart, insideTempChart, insideHumidityChart
    ];
    charts.forEach(chart => {
        if (chart) {
            chart.update('none'); // 'none' = no animation
        }
    });
}

// Make refreshAllCharts available globally for theme toggle
window.refreshAllCharts = refreshAllCharts;

// Celestial data cache (from /api/celestial)
let celestialData = null;

// Latest samples for current-conditions panel
let latestWeather = null;
let latestBarometer = null;
let barometerTrend = 'steady';      // 'rapid-rise', 'slow-rise', 'steady', 'slow-fall', 'rapid-fall'
let barometerLevel = 'normal';      // 'high', 'normal', 'low'
let barometerForecast = '';         // Human-readable forecast based on logic
let latestFeelsLike = null;
let latestHumidity = null;
let latestInsideTemp = null;
let latestInsideHumidity = null;
let latestWind = null;
let latestRainRate = null;
let latestRainToday = null;
let lightningToday = null;

// Active Alerts
let rainRecentlyActive = false;
let lightningRecentlyActive = false;
let windStrong = false;

// Feels-like alert thresholds
const FEELS_EXTREME_HEAT = 95; // °F heat index threshold for alerting
const FEELS_EXTREME_COLD = 32; // °F wind chill threshold for alerting

// Day / Week / Month selector
let currentRange = 'day';

// ---------------------------------------------------------------------
// Celestial data loader
// ---------------------------------------------------------------------
async function loadCelestial() {
    try {
        const res = await fetch('/api/celestial');
        if (!res.ok) throw new Error('HTTP ' + res.status);
        
        const data = await res.json();
        celestialData = data;
        
        console.log('[Celestial] Loaded data:', data);
    } catch (err) {
        console.error('Error loading celestial data:', err);
        celestialData = null;
    }
}

// ---------------------------------------------------------------------
// Master time grid (driven by /api/weather)
// ---------------------------------------------------------------------
let masterTimes = [];     // array<Date>
let masterTimesMs = [];   // array<number>

function setMasterTimesFromRows(rows) {
    masterTimes   = rows.map(r => new Date(r.timestamp));
    masterTimesMs = masterTimes.map(t => t.getTime());
}

function hasMasterTimes() {
    return Array.isArray(masterTimes) && masterTimes.length > 0;
}

// Align data rows to master time grid, inserting nulls as needed
function alignToMasterTimes(rows, valueSelector) {
    if (!hasMasterTimes()) return [];

    const map = new Map();
    for (const r of rows) {
        const ms = new Date(r.timestamp).getTime();
        map.set(ms, valueSelector(r));
    }

    return masterTimesMs.map(ms => map.has(ms) ? map.get(ms) : null);
}

// ---------------------------------------------------------------------
// Helpers: time formatting + tick generator
// ---------------------------------------------------------------------
function formatTime24(date) {
    const h = String(date.getHours()).padStart(2, '0');
    const m = String(date.getMinutes()).padStart(2, '0');
    return `${h}:${m}`;
}

function formatDateTime24(date) {
    const mm = String(date.getMonth() + 1).padStart(2, '0');
    const dd = String(date.getDate()).padStart(2, '0');
    const h  = String(date.getHours()).padStart(2, '0');
    const m  = String(date.getMinutes()).padStart(2, '0');
    return `${mm}/${dd} ${h}:${m}`;
}

// Shared tick config for time-based charts
function makeTimeTickOptions(times) {
    const timesForTicks  = times.slice();
    const labelsForTicks = new Array(timesForTicks.length).fill('');

    if (currentRange === 'day' && timesForTicks.length) {
        const FOUR_HOURS = 4 * 60 * 60 * 1000;

        const first = timesForTicks[0];
        const last  = timesForTicks[timesForTicks.length - 1];

        // Anchor to *local midnight* of the first day shown
        const midnight   = new Date(first.getFullYear(), first.getMonth(), first.getDate());
        const midnightMs = midnight.getTime();

        const startMs = first.getTime();
        const endMs   = last.getTime();

        const offsetFromMidnight = startMs - midnightMs;

        // First 4-hour boundary at or AFTER the first sample
        let boundary = midnightMs + Math.ceil(offsetFromMidnight / FOUR_HOURS) * FOUR_HOURS;

        // March forward in 4-hour steps until we pass the end of the window
        while (boundary <= endMs + FOUR_HOURS) {
            // Find first index whose time >= this boundary
            let idx = -1;
            for (let i = 0; i < timesForTicks.length; i++) {
                if (timesForTicks[i].getTime() >= boundary) {
                    idx = i;
                    break;
                }
            }
            if (idx === -1) break;

            labelsForTicks[idx] = formatTime24(new Date(boundary));
            boundary += FOUR_HOURS;
        }
    } else if (currentRange === 'week' && timesForTicks.length) {
        // Place labels at local midnight for each day in the range
        const first = timesForTicks[0];
        const last  = timesForTicks[timesForTicks.length - 1];

        // Start from the midnight of the first day
        let cursor = new Date(first.getFullYear(), first.getMonth(), first.getDate());
        const endMs = last.getTime();

        while (cursor.getTime() <= endMs) {
            const boundaryMs = cursor.getTime();
            // Find first index whose time >= this boundary
            let idx = -1;
            for (let i = 0; i < timesForTicks.length; i++) {
                if (timesForTicks[i].getTime() >= boundaryMs) {
                    idx = i;
                    break;
                }
            }
            if (idx === -1) break;

            // Label with two-digit day (DD)
            const label = `${String(cursor.getDate()).padStart(2,'0')}`;
            labelsForTicks[idx] = label;

            // advance by one day
            cursor = new Date(cursor.getFullYear(), cursor.getMonth(), cursor.getDate() + 1);
        }
    } else if (currentRange === 'month' && timesForTicks.length) {
        // Label local midnight for each day in the month range; show DD only.
        const first = timesForTicks[0];
        const last  = timesForTicks[timesForTicks.length - 1];
        let cursor = new Date(first.getFullYear(), first.getMonth(), first.getDate());
        const endMs = last.getTime();

        while (cursor.getTime() <= endMs) {
            const boundaryMs = cursor.getTime();
            let idx = -1;
            for (let i = 0; i < timesForTicks.length; i++) {
                if (timesForTicks[i].getTime() >= boundaryMs) { idx = i; break; }
            }
            if (idx === -1) break;
            const dayNum = cursor.getDate();
            // Show only every other day label (odd day numbers) to reduce crowding.
            if (dayNum % 2 === 1) {
                labelsForTicks[idx] = String(dayNum).padStart(2,'0');
            }
            cursor = new Date(cursor.getFullYear(), cursor.getMonth(), cursor.getDate() + 1);
        }
    }

    return {
        autoSkip: false,
        maxRotation: 0,
        callback(value, index) {
            const t = timesForTicks[index];
            if (!t) return '';

            if (currentRange === 'day') {
                return labelsForTicks[index] || '';
            } else if (currentRange === 'week') {
                // show only the computed midnight labels, otherwise empty
                return labelsForTicks[index] || '';
            } else if (currentRange === 'month') {
                return labelsForTicks[index] || '';
            } else {
                return formatDateTime24(t);
            }
        }
    };
}

// ---------------------------------------------------------------------
// Global Chart.js defaults
// ---------------------------------------------------------------------
if (window.Chart && Chart.defaults) {
    Chart.defaults.font.family = 'system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif';
    Chart.defaults.font.size = 11;
    Chart.defaults.color = '#4b5563';
    Chart.defaults.elements.line.borderWidth = 2;
    Chart.defaults.elements.line.tension = 0.25;
    Chart.defaults.elements.point.radius = 0;
    Chart.defaults.elements.point.hoverRadius = 4;
    Chart.defaults.elements.point.hitRadius = 6;
    Chart.defaults.plugins.legend.labels.boxWidth = 10;
    Chart.defaults.aspectRatio = 1.6;
}



// ---------------------------------------------------------------------
// Plugins: wind vector + day/night shading
// ---------------------------------------------------------------------
const windVectorPlugin = {
    id: 'windVector',
    afterDatasetsDraw(chart, args, pluginOptions) {
        const speeds = pluginOptions.speeds || [];
        const dirs   = pluginOptions.directions || [];
        if (!speeds.length || !dirs.length) return;

        const { ctx, scales } = chart;
        const yScale = scales.y;
        if (!yScale) return;

        const meta = chart.getDatasetMeta(0);
        if (!meta || !meta.data || !meta.data.length) return;

        const baseY = yScale.getPixelForValue(0);

        ctx.save();
        ctx.lineWidth = 1.2;
        ctx.strokeStyle = 'rgba(37, 99, 235, 0.9)';

        meta.data.forEach((elem, i) => {
            const speed = speeds[i];
            const dir   = dirs[i];

            if (speed == null || isNaN(speed) || dir == null || isNaN(dir)) return;

            const x = elem.x; // true x-position from Chart.js

            const maxSpeed = pluginOptions.maxSpeed || Math.max(...speeds);
            if (!isFinite(maxSpeed) || maxSpeed <= 0) return;

            const lengthData = speed;
            const lengthPx = Math.abs(yScale.getPixelForValue(lengthData) - baseY);

            // WeeWX coordinate system with -90° rotation:
            // Wind direction is "from", so add 180° to get "to" direction
            // Then subtract 90° for coordinate rotation
            const dirTo = (dir + 180 - 90) % 360;
            
            // Convert to rotated coordinates
            // After -90° rotation:
            // North (0°) → points up (+Y)
            // East (90°) → points right (+X)
            // South (180°) → points down (-Y)
            // West (270°) → points left (-X)
            const angleRad = dirTo * Math.PI / 180;
            const dx = Math.sin(angleRad) * lengthPx;
            const dy = -Math.cos(angleRad) * lengthPx;

            const x0 = x;
            const y0 = baseY;
            const x1 = x0 + dx;
            const y1 = y0 + dy;

            // Shaft
            ctx.beginPath();
            ctx.moveTo(x0, y0);
            ctx.lineTo(x1, y1);
            ctx.stroke();

            // Arrow head (barbs at the tip)
            const headSize = 5;
            const headAngle = Math.PI * 0.2; // 36 degrees
            const angleHead1 = angleRad + Math.PI - headAngle;
            const angleHead2 = angleRad + Math.PI + headAngle;

            ctx.beginPath();
            ctx.moveTo(x1, y1);
            ctx.lineTo(
                x1 + Math.sin(angleHead1) * headSize,
                y1 - Math.cos(angleHead1) * headSize
            );
            ctx.lineTo(
                x1 + Math.sin(angleHead2) * headSize,
                y1 - Math.cos(angleHead2) * headSize
            );
            ctx.closePath();
            ctx.fillStyle = 'rgba(37, 99, 235, 0.9)';
            ctx.fill();
        });

        ctx.restore();
    }
};

const STATION_LAT = 32.093174;
const STATION_LON = -110.777557;
const STATION_TZ_OFFSET_HOURS = -7;

// Replaced with backend API call - this is now a fallback/stub
function computeSunTimes(dateLocal) {
    // If we have celestial data from the API, use it
    if (celestialData && celestialData.sunrise && celestialData.sunset) {
        return {
            sunrise: new Date(celestialData.sunrise),
            sunset: new Date(celestialData.sunset)
        };
    }
    
    // Fallback: assume 6am sunrise, 6pm sunset if API data unavailable
    const year = dateLocal.getFullYear();
    const month = dateLocal.getMonth();
    const day = dateLocal.getDate();
    
    return {
        sunrise: new Date(year, month, day, 6, 0, 0),
        sunset: new Date(year, month, day, 18, 0, 0)
    };
}

const dayNightBackgroundPlugin = {
    id: 'dayNightBackground',
    beforeDraw(chart, args, pluginOptions) {
        if (!pluginOptions || !pluginOptions.enabled) return;

        const { ctx, chartArea, scales } = chart;
        const xScale = scales.x || scales['x'];
        if (!xScale || !chartArea) return;

        const timesMs = pluginOptions.times;
        if (!timesMs || timesMs.length < 2) return;

        // Detect current theme and use appropriate shading
        const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark';
        const fillStyle = isDarkMode 
            ? 'rgba(0, 0, 0, 0.3)'  // Darker overlay for dark mode
            : 'rgba(148, 163, 184, 0.16)';  // Light gray for light mode

        const top    = chartArea.top;
        const height = chartArea.bottom - chartArea.top;

        // Use celestial data for sunrise/sunset times
        let sunriseTime, sunsetTime;
        if (celestialData && celestialData.sunrise && celestialData.sunset) {
            sunriseTime = new Date(celestialData.sunrise);
            sunsetTime = new Date(celestialData.sunset);
        } else {
            // Fallback: 6am sunrise, 6pm sunset
            const now = new Date();
            sunriseTime = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 6, 0, 0);
            sunsetTime = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 18, 0, 0);
        }

        ctx.save();
        ctx.fillStyle = fillStyle;

        // Get the chart's dataset to know how many data points we have
        let datasetLength = 0;
        
        // For regular charts, use data.labels
        if (chart.data.labels && chart.data.labels.length > 0) {
            datasetLength = chart.data.labels.length;
        }
        // For scatter plots with category x-axis, check the scale configuration
        else if (chart.config.type === 'scatter' && xScale.type === 'category') {
            // Category scale stores labels in the scale itself after initialization
            const scaleLabels = xScale.getLabels ? xScale.getLabels() : [];
            datasetLength = scaleLabels.length;
        }
        
        if (datasetLength === 0) {
            ctx.restore();
            return;
        }

        for (let i = 0; i < datasetLength - 1; i++) {
            // Use the provided times array to check if this period is night
            if (i >= timesMs.length - 1) break;
            
            const tMid = (timesMs[i] + timesMs[i + 1]) / 2;
            const currentDate = new Date(tMid);
            
            // Get hour of day for comparison
            const hour = currentDate.getHours();
            const minute = currentDate.getMinutes();
            const timeInMinutes = hour * 60 + minute;
            
            const sunriseMinutes = sunriseTime.getHours() * 60 + sunriseTime.getMinutes();
            const sunsetMinutes = sunsetTime.getHours() * 60 + sunsetTime.getMinutes();
            
            // Check if this time is before sunrise or after sunset
            const isNight = timeInMinutes < sunriseMinutes || timeInMinutes >= sunsetMinutes;
            if (!isNight) continue;

            // Use data indices for pixel calculation
            const xStart = xScale.getPixelForValue(i);
            const xEnd   = xScale.getPixelForValue(i + 1);

            if (!Number.isFinite(xStart) || !Number.isFinite(xEnd)) continue;

            ctx.fillRect(xStart, top, xEnd - xStart, height);
        }

        ctx.restore();
    }
};

Chart.register(windVectorPlugin, dayNightBackgroundPlugin);

// ---------------------------------------------------------------------
// Feels-like selection helper
// ---------------------------------------------------------------------
// ---------------------------------------------------------------------
// Range controls
// ---------------------------------------------------------------------
function setRange(range) {
    currentRange = range;

    document.querySelectorAll('.range-button').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.range === range);
    });

    loadAll();
}

// ---------------------------------------------------------------------
// Current Conditions sidebar
// ---------------------------------------------------------------------
function updateCurrentConditions() {
    const baroEl            = document.getElementById('cc-barometer');
    const outTempEl         = document.getElementById('cc-out-temp');
    const outDewEl          = document.getElementById('cc-out-dew');
    const feelsLikeEl       = document.getElementById('cc-feelslike');
    const outHumEl          = document.getElementById('cc-out-hum');
    const windEl            = document.getElementById('cc-wind');
    const windRowEl         = document.getElementById('cc-wind-row');
    const windIconEl        = document.getElementById('cc-wind-icon');
    const rainTodayEl       = document.getElementById('cc-rain-today');
    const rainTodayRowEl    = document.getElementById('cc-rain-today-row');
    const rainRateEl        = document.getElementById('cc-rain-rate');
    const rainRateRowEl     = document.getElementById('cc-rain-rate-row');
    const rainIconEl        = document.getElementById('cc-rain-icon');
    const lightningTodayEl  = document.getElementById('cc-lightning-today');
    const lightningRowEl    = document.getElementById('cc-lightning-row');
    const lightningIconEl   = document.getElementById('cc-lightning-icon');
    // Inside
    const inTempEl          = document.getElementById('cc-in-temp');
    const inHumEl           = document.getElementById('cc-in-hum');

    // Barometer
    if (latestBarometer && typeof latestBarometer.pressure === 'number') {
        baroEl.textContent = latestBarometer.pressure.toFixed(3) + ' inHg';
        // Update barometer icon trend class and forecast
        const baroIconEl = document.getElementById('cc-barometer-icon');
        const baroForecastEl = document.getElementById('cc-barometer-forecast');
        if (baroIconEl) {
            baroIconEl.className = 'cc-barometer-icon ' + barometerTrend;
        }
        if (baroForecastEl) {
            baroForecastEl.textContent = barometerForecast;
        }
    } else {
        baroEl.textContent = '--';
        barometerTrend = 'steady';
        barometerLevel = 'normal';
        barometerForecast = '--';
        const baroIconEl = document.getElementById('cc-barometer-icon');
        const baroForecastEl = document.getElementById('cc-barometer-forecast');
        if (baroIconEl) {
            baroIconEl.className = 'cc-barometer-icon steady';
        }
        if (baroForecastEl) {
            baroForecastEl.textContent = '--';
        }
    }

    // Outside temperature / dew point
    if (latestWeather) {
        if (typeof latestWeather.temperature === 'number') {
            outTempEl.textContent = latestWeather.temperature.toFixed(1) + ' °F';
        } else {
            outTempEl.textContent = '--';
        }
        if (typeof latestWeather.dewpoint === 'number') {
            outDewEl.textContent  = latestWeather.dewpoint.toFixed(1) + ' °F';
        } else {
            outDewEl.textContent  = '--';
        }
    } else {
        outTempEl.textContent = '--';
        outDewEl.textContent  = '--';
    }

    // Feels-like (heat index / wind chill / air temp) + icon + extreme alerts
    const feelsRowEl = document.getElementById('cc-feels-row');
    const feelsIconEl = document.getElementById('cc-feels-icon');

    if (latestFeelsLike && latestFeelsLike.activeValue != null) {
        // Backend now provides activeValue, activeSource, activeLabel
        const activeValue = latestFeelsLike.activeValue;
        const sourceLabel = latestFeelsLike.activeLabel || 'Air Temp';
        const sourceKey = latestFeelsLike.activeSource || 'air';
        
        feelsLikeEl.textContent = activeValue.toFixed(1) + ' °F (' + sourceLabel + ')';

        // Swap icon based on active source
        if (feelsIconEl) {
            let svg = '';
            if (sourceKey === 'heat') {
                // Sun / heat icon
                svg = '<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><circle cx="12" cy="12" r="3.25" stroke="#f59e0b" stroke-width="1.5" fill="none"/><path stroke="#f59e0b" stroke-width="1.5" d="M12 2v1.5M12 20.5V22M4.22 4.22l1.06 1.06M18.72 18.72l1.06 1.06M1 12h1.5M21.5 12H23M4.22 19.78l1.06-1.06M18.72 5.28l1.06-1.06"/></svg>';
            } else if (sourceKey === 'chill') {
                // Snowflake / cold icon
                svg = '<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><g stroke="#06b6d4" stroke-width="1.5" fill="none"><path d="M12 2v20" /><path d="M4 8l16 8" /><path d="M20 8L4 16"/></g></svg>';
            } else {
                // Thermometer (air temp)
                svg = '<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke="#9ca3af" stroke-width="1.5" fill="none" d="M14 14.5V6a2 2 0 10-4 0v8.5a3 3 0 104 0z" /><path stroke="#9ca3af" stroke-width="1.5" d="M12 2v2" /></svg>';
            }
            feelsIconEl.innerHTML = svg;
        }

        // Determine if value is extreme and should alert/pulse
        let extreme = false;
        if (sourceKey === 'heat' && activeValue >= FEELS_EXTREME_HEAT) extreme = true;
        if (sourceKey === 'chill' && activeValue <= FEELS_EXTREME_COLD) extreme = true;

        if (feelsRowEl) {
            if (extreme) {
                feelsRowEl.classList.add('cc-alert');
                console.log('[UI] Feels-like row: cc-alert class ADDED (extreme)');
            } else {
                feelsRowEl.classList.remove('cc-alert');
                console.log('[UI] Feels-like row: cc-alert class REMOVED');
            }
        }

        if (feelsIconEl) {
            if (extreme) {
                feelsIconEl.classList.add('cc-pulse');
                console.log('[UI] Feels icon: cc-pulse class ADDED (extreme)');
            } else {
                feelsIconEl.classList.remove('cc-pulse');
                console.log('[UI] Feels icon: cc-pulse class REMOVED');
            }
        }

    } else if (latestWeather && typeof latestWeather.temperature === 'number') {
        feelsLikeEl.textContent = latestWeather.temperature.toFixed(1) + ' °F (Air Temp)';
        if (feelsRowEl) feelsRowEl.classList.remove('cc-alert');
        if (feelsIconEl) feelsIconEl.classList.remove('cc-pulse');
    } else {
        feelsLikeEl.textContent = '--';
    }

    // Outside humidity
    if (latestHumidity && typeof latestHumidity.humidity === 'number') {
        outHumEl.textContent = latestHumidity.humidity.toFixed(1) + ' %';
    } else {
        outHumEl.textContent = '--';
    }

    // Wind
    if (latestWind && latestWind.speed != null) {
        const spd = latestWind.speed.toFixed(1);
        if (latestWind.speed === 0 || latestWind.direction == null) {
            windEl.textContent = `${spd} mph  -- (--)°`;
        } else {
            const deg = latestWind.direction.toFixed(0);
            // Backend now provides compass direction
            const dir = latestWind.compass || '--';
            windEl.textContent = `${spd} mph  ${dir} (${deg}°)`;
        }
    } else {
        windEl.textContent = `-- mph -- (--)°`;
    }

    // Wind alert styling
if (windRowEl) {
        if (windStrong) {
            windRowEl.classList.add('cc-alert');
            console.log('[UI] Wind alert row: cc-alert class ADDED');
        } else {
            windRowEl.classList.remove('cc-alert');
            console.log('[UI] Wind alert row: cc-alert class REMOVED');
        }
    }

    if (windIconEl) {
        if (windStrong) {
            windIconEl.classList.add('cc-pulse');
            console.log('[UI] Wind icon: cc-pulse class ADDED (SVG should now be amber & pulsing)');
        } else {
            windIconEl.classList.remove('cc-pulse');
            console.log('[UI] Wind icon: cc-pulse class REMOVED');
        }
    }

    // Inside temperature
    if (latestInsideTemp && typeof latestInsideTemp.inside_temp_f === 'number') {
        inTempEl.textContent = latestInsideTemp.inside_temp_f.toFixed(1) + ' °F';
    } else {
        inTempEl.textContent = '--';
    }

    // Inside humidity
    if (latestInsideHumidity && typeof latestInsideHumidity.inside_humidity === 'number') {
        inHumEl.textContent = latestInsideHumidity.inside_humidity.toFixed(1) + ' %';
    } else {
        inHumEl.textContent = '--';
    }

    // Rain Today + Rain Rate
    if (rainTodayEl) {
        if (typeof rainToday === 'number') {
            rainTodayEl.textContent = rainToday.toFixed(2) + ' in';
        } else {
            rainTodayEl.textContent = '--';
        }
    }

    if (rainRateEl) {
        if (typeof latestRainRate === 'number') {
            rainRateEl.textContent = latestRainRate.toFixed(2) + ' in/hr';
        } else {
            rainRateEl.textContent = '0.00 in/hr';
        }
    }

    // Highlight rows when there has been any rain today
    if (rainTodayRowEl) {
        if (typeof rainToday === 'number' && rainToday > 0) {
            rainTodayRowEl.classList.add('cc-alert');
        } else {
            rainTodayRowEl.classList.remove('cc-alert');
        }
    }

    if (rainRateRowEl) {
        if (rainRecentlyActive && typeof latestRainRate === 'number' && latestRainRate > 0) {
            rainRateRowEl.classList.add('cc-alert');
            console.log('[UI] Rain rate alert row: cc-alert class ADDED');
        } else {
            rainRateRowEl.classList.remove('cc-alert');
            console.log('[UI] Rain rate alert row: cc-alert class REMOVED');
        }
    }

    // Pulse icon when there has been rain in the last 10 minutes
    if (rainIconEl) {
        if (rainRecentlyActive && typeof latestRainRate === 'number' && latestRainRate > 0) {
            rainIconEl.classList.add('cc-pulse');
            console.log('[UI] Rain icon: cc-pulse class ADDED');
        } else {
            rainIconEl.classList.remove('cc-pulse');
            console.log('[UI] Rain icon: cc-pulse class REMOVED');
        }
    }

    // Lightning strikes today
    if (lightningTodayEl) {
        if (typeof lightningToday === 'number') {
            lightningTodayEl.textContent = lightningToday.toFixed(0);
        } else {
            lightningTodayEl.textContent = '--';
        }
    }

    // Color + pulse logic
    if (lightningRowEl) {
        // Alert color if there has been *any* lightning today
        if (typeof lightningToday === 'number' && lightningToday > 0) {
            lightningRowEl.classList.add('cc-alert');
            console.log('[UI] Lightning alert row: cc-alert class ADDED');
        } else {
            lightningRowEl.classList.remove('cc-alert');
            console.log('[UI] Lightning alert row: cc-alert class REMOVED');
        }
    }

    if (lightningIconEl) {
        // Pulse only if we've had strikes in the last 10 minutes
        if (lightningRecentlyActive && typeof lightningToday === 'number' && lightningToday > 0) {
            lightningIconEl.classList.add('cc-pulse');
            console.log('[UI] Lightning icon: cc-pulse class ADDED');
        } else {
            lightningIconEl.classList.remove('cc-pulse');
            console.log('[UI] Lightning icon: cc-pulse class REMOVED');
        }
    }

    // Celestial data (sunrise/sunset, moon phase, etc.)
    updateCelestialDisplay();
}

// Get moon phase emoji based on fraction (0.0-1.0)
function getMoonPhaseEmoji(fraction) {
    // fraction: 0.0 = new moon, 0.25 = first quarter, 0.5 = full, 0.75 = last quarter, 1.0 = new again
    if (fraction < 0.0625) return '🌑'; // New Moon
    if (fraction < 0.1875) return '🌒'; // Waxing Crescent
    if (fraction < 0.3125) return '🌓'; // First Quarter
    if (fraction < 0.4375) return '🌔'; // Waxing Gibbous
    if (fraction < 0.5625) return '🌕'; // Full Moon
    if (fraction < 0.6875) return '🌖'; // Waning Gibbous
    if (fraction < 0.8125) return '🌗'; // Last Quarter
    if (fraction < 0.9375) return '🌘'; // Waning Crescent
    return '🌑'; // New Moon (wrapping around)
}

function updateCelestialDisplay() {
    if (!celestialData) return;

    const sunriseSunsetEl = document.getElementById('cc-sunrise-sunset');
    const daylightHoursEl = document.getElementById('cc-daylight-hours');
    const moonPhaseEl = document.getElementById('cc-moon-phase');
    const moonriseMoonsetEl = document.getElementById('cc-moonrise-moonset');
    const civilTwilightEl = document.getElementById('cc-civil-twilight');
    const nauticalTwilightEl = document.getElementById('cc-nautical-twilight');
    const astronomicalTwilightEl = document.getElementById('cc-astronomical-twilight');
    const goldenHourMorningEl = document.getElementById('cc-golden-hour-morning');
    const goldenHourEveningEl = document.getElementById('cc-golden-hour-evening');
    const blueHourMorningEl = document.getElementById('cc-blue-hour-morning');
    const blueHourEveningEl = document.getElementById('cc-blue-hour-evening');
    const moonPhaseIconEl = document.getElementById('cc-moon-phase-icon');

    // Format ISO timestamp to 24-hour HH:MM
    const formatTime24FromISO = (timeStr) => {
        if (!timeStr) return '--';
        // If already a short 24h string provided by backend (e.g. "07:08"), return it
        if (typeof timeStr === 'string' && /^\d{1,2}:\d{2}$/.test(timeStr)) {
            // Ensure zero-padded hour
            const parts = timeStr.split(':');
            return parts[0].padStart(2, '0') + ':' + parts[1];
        }
        const d = new Date(timeStr);
        if (isNaN(d.getTime())) return '--';
        return formatTime24(d);
    };

    // Sunrise / Sunset (prefer backend 24-hour fields)
    if (sunriseSunsetEl) {
        const sunrise = celestialData.sunrise24 ? celestialData.sunrise24 : formatTime24FromISO(celestialData.sunrise);
        const sunset  = celestialData.sunset24  ? celestialData.sunset24  : formatTime24FromISO(celestialData.sunset);
        sunriseSunsetEl.textContent = `${sunrise} / ${sunset}`;
    }

    // Daylight Hours (use pre-formatted value from backend)
    if (daylightHoursEl) {
        if (celestialData.daylightHoursFormatted) {
            daylightHoursEl.textContent = celestialData.daylightHoursFormatted;
        } else if (celestialData.daylightHours) {
            // Fallback to client-side formatting if backend doesn't provide formatted value
            const hours = Math.floor(celestialData.daylightHours);
            const minutes = Math.round((celestialData.daylightHours - hours) * 60);
            daylightHoursEl.textContent = `${hours}h ${minutes}m`;
        } else {
            daylightHoursEl.textContent = '--';
        }
    }

    // Moon Phase (use pre-computed percentage from backend)
    if (moonPhaseEl && celestialData.moonPhase) {
        const percentage = celestialData.moonPhase.percentage || Math.round(celestialData.moonPhase.fraction * 100);
        moonPhaseEl.textContent = `${celestialData.moonPhase.name} (${percentage}%)`;
        
        // Update moon icon based on phase
        if (moonPhaseIconEl) {
            const emoji = getMoonPhaseEmoji(celestialData.moonPhase.fraction);
            moonPhaseIconEl.textContent = emoji;
        }
    } else if (moonPhaseEl) {
        moonPhaseEl.textContent = '--';
        if (moonPhaseIconEl) {
            moonPhaseIconEl.textContent = '🌙';
        }
    }

    // Moonrise / Moonset (24-hour)
    if (moonriseMoonsetEl) {
        const moonrise = celestialData.moonrise24 ? celestialData.moonrise24 : formatTime24FromISO(celestialData.moonrise);
        const moonset  = celestialData.moonset24  ? celestialData.moonset24  : formatTime24FromISO(celestialData.moonset);
        moonriseMoonsetEl.textContent = `${moonrise} / ${moonset}`;
    }

    // Civil Twilight (24-hour)
    if (civilTwilightEl) {
        const dawn = celestialData.civilDawn24 ? celestialData.civilDawn24 : formatTime24FromISO(celestialData.civilDawn);
        const dusk = celestialData.civilDusk24 ? celestialData.civilDusk24 : formatTime24FromISO(celestialData.civilDusk);
        civilTwilightEl.textContent = `${dawn} / ${dusk}`;
    }

    // Nautical Twilight (24-hour)
    if (nauticalTwilightEl) {
        const ndawn = celestialData.nauticalDawn24 ? celestialData.nauticalDawn24 : formatTime24FromISO(celestialData.nauticalDawn);
        const ndusk = celestialData.nauticalDusk24 ? celestialData.nauticalDusk24 : formatTime24FromISO(celestialData.nauticalDusk);
        nauticalTwilightEl.textContent = `${ndawn} / ${ndusk}`;
    }

    // Astronomical Twilight (24-hour)
    if (astronomicalTwilightEl) {
        const adawn = celestialData.astronomicalDawn24 ? celestialData.astronomicalDawn24 : formatTime24FromISO(celestialData.astronomicalDawn);
        const adusk = celestialData.astronomicalDusk24 ? celestialData.astronomicalDusk24 : formatTime24FromISO(celestialData.astronomicalDusk);
        astronomicalTwilightEl.textContent = `${adawn} / ${adusk}`;
    }

    // Golden Hour - split into morning/evening rows
    if (goldenHourMorningEl) {
        let mornText = '--';
        if (celestialData.goldenHourMorningStart24 && celestialData.goldenHourMorningEnd24) {
            mornText = `${celestialData.goldenHourMorningStart24}-${celestialData.goldenHourMorningEnd24}`;
        } else if (celestialData.goldenHourMorningStart && celestialData.goldenHourMorningEnd) {
            const start = formatTime24FromISO(celestialData.goldenHourMorningStart);
            const end = formatTime24FromISO(celestialData.goldenHourMorningEnd);
            mornText = `${start}-${end}`;
        }
        goldenHourMorningEl.textContent = mornText;
    }

    if (goldenHourEveningEl) {
        let eveText = '--';
        if (celestialData.goldenHourEveningStart24 && celestialData.goldenHourEveningEnd24) {
            eveText = `${celestialData.goldenHourEveningStart24}-${celestialData.goldenHourEveningEnd24}`;
        } else if (celestialData.goldenHourEveningStart && celestialData.goldenHourEveningEnd) {
            const start = formatTime24FromISO(celestialData.goldenHourEveningStart);
            const end = formatTime24FromISO(celestialData.goldenHourEveningEnd);
            eveText = `${start}-${end}`;
        }
        goldenHourEveningEl.textContent = eveText;
    }

    // Blue Hour - split into morning/evening rows
    if (blueHourMorningEl) {
        let bmText = '--';
        if (celestialData.blueHourMorningStart24 && celestialData.blueHourMorningEnd24) {
            bmText = `${celestialData.blueHourMorningStart24}-${celestialData.blueHourMorningEnd24}`;
        } else if (celestialData.blueHourMorningStart && celestialData.blueHourMorningEnd) {
            const start = formatTime24FromISO(celestialData.blueHourMorningStart);
            const end = formatTime24FromISO(celestialData.blueHourMorningEnd);
            bmText = `${start}-${end}`;
        }
        blueHourMorningEl.textContent = bmText;
    }

    if (blueHourEveningEl) {
        let beText = '--';
        if (celestialData.blueHourEveningStart24 && celestialData.blueHourEveningEnd24) {
            beText = `${celestialData.blueHourEveningStart24}-${celestialData.blueHourEveningEnd24}`;
        } else if (celestialData.blueHourEveningStart && celestialData.blueHourEveningEnd) {
            const start = formatTime24FromISO(celestialData.blueHourEveningStart);
            const end = formatTime24FromISO(celestialData.blueHourEveningEnd);
            beText = `${start}-${end}`;
        }
        blueHourEveningEl.textContent = beText;
    }
}

function getMoonPhaseSVG(fraction) {
    // Return different moon phase SVGs based on illumination
    if (fraction < 0.05) {
        // New Moon (dark circle)
        return `<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="8" stroke="currentColor" stroke-width="1.5" fill="#1e293b"/></svg>`;
    } else if (fraction < 0.25) {
        // Waxing Crescent
        return `<svg viewBox="0 0 24 24"><path stroke="currentColor" stroke-width="1.5" fill="none" d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>`;
    } else if (fraction < 0.35) {
        // First Quarter
        return `<svg viewBox="0 0 24 24"><path d="M12 4 A8 8 0 0 1 12 20 L12 4" stroke="currentColor" stroke-width="1.5" fill="none"/><circle cx="12" cy="12" r="8" stroke="currentColor" stroke-width="1.5" fill="none"/></svg>`;
    } else if (fraction < 0.65) {
        // Waxing Gibbous / Full
        return `<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="8" stroke="currentColor" stroke-width="1.5" fill="currentColor"/></svg>`;
    } else if (fraction < 0.75) {
        // Last Quarter
        return `<svg viewBox="0 0 24 24"><path d="M12 4 A8 8 0 0 0 12 20 L12 4" stroke="currentColor" stroke-width="1.5" fill="none"/><circle cx="12" cy="12" r="8" stroke="currentColor" stroke-width="1.5" fill="none"/></svg>`;
    } else {
        // Waning Crescent
        return `<svg viewBox="0 0 24 24"><path stroke="currentColor" stroke-width="1.5" fill="none" d="M3 11.21A9 9 0 0 0 12.79 21 7 7 0 0 1 3 11.21z"/></svg>`;
    }
}

// Celestial details toggle
document.addEventListener('DOMContentLoaded', () => {
    const toggleBtn = document.getElementById('cc-celestial-toggle');
    const detailRows = document.querySelectorAll('.cc-celestial-details');
    
    if (toggleBtn) {
        toggleBtn.addEventListener('click', () => {
            const isHidden = detailRows[0]?.style.display === 'none';
            detailRows.forEach(row => {
                row.style.display = isHidden ? 'flex' : 'none';
            });
            toggleBtn.textContent = isHidden ? 'Hide Details ▲' : 'Show Details ▼';
        });
    }
});

