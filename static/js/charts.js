// --- Statistics Panel: Rain & Lightning Totals ---
(function(){
    async function fetchJSON(url) {
        const res = await fetch(url);
        if (!res.ok) throw new Error(`HTTP ${res.status} for ${url}`);
        return res.json();
    }

    async function computeStatistics(range) {
        try {
            const q = `?range=${encodeURIComponent(range)}`;
            const stats = await fetchJSON(`/api/statistics${q}`);

            // Update all Today/Range values directly from backend response
            const el = (id) => document.getElementById(id);
            
            // Rain
            if (el('stats-rain-today')) el('stats-rain-today').textContent = stats.rainToday.toFixed(2);
            if (el('stats-rain-range')) el('stats-rain-range').textContent = stats.rainRange.toFixed(2);
            
            // Lightning
            if (el('stats-strike-today')) el('stats-strike-today').textContent = String(stats.strikesToday);
            if (el('stats-strike-range')) el('stats-strike-range').textContent = String(stats.strikesRange);
            
            // Temperature - split high/low
            if (el('stats-temp-hi-today')) {
                const [hiToday, loToday] = stats.tempToday.split(' / ');
                el('stats-temp-hi-today').textContent = hiToday || '--';
                if (el('stats-temp-lo-today')) el('stats-temp-lo-today').textContent = loToday || '--';
            }
            if (el('stats-temp-hi-range')) {
                const [hiRange, loRange] = stats.tempRange.split(' / ');
                el('stats-temp-hi-range').textContent = hiRange || '--';
                if (el('stats-temp-lo-range')) el('stats-temp-lo-range').textContent = loRange || '--';
            }
            
            // Feels Like (Heat Index) - split high/low
            if (el('stats-feels-hi-today')) {
                const [hiToday, loToday] = stats.feelsToday.split(' / ');
                el('stats-feels-hi-today').textContent = hiToday || '--';
                if (el('stats-feels-lo-today')) el('stats-feels-lo-today').textContent = loToday || '--';
            }
            if (el('stats-feels-hi-range')) {
                const [hiRange, loRange] = stats.feelsRange.split(' / ');
                el('stats-feels-hi-range').textContent = hiRange || '--';
                if (el('stats-feels-lo-range')) el('stats-feels-lo-range').textContent = loRange || '--';
            }
            
            // Windchill
            if (el('stats-windchill-today')) el('stats-windchill-today').textContent = stats.windchillToday;
            if (el('stats-windchill-range')) el('stats-windchill-range').textContent = stats.windchillRange;
            
            // Dewpoint - split high/low
            if (el('stats-dew-hi-today')) {
                const [hiToday, loToday] = stats.dewToday.split(' / ');
                el('stats-dew-hi-today').textContent = hiToday || '--';
                if (el('stats-dew-lo-today')) el('stats-dew-lo-today').textContent = loToday || '--';
            }
            if (el('stats-dew-hi-range')) {
                const [hiRange, loRange] = stats.dewRange.split(' / ');
                el('stats-dew-hi-range').textContent = hiRange || '--';
                if (el('stats-dew-lo-range')) el('stats-dew-lo-range').textContent = loRange || '--';
            }
            
            // Humidity
            if (el('stats-humidity-today')) el('stats-humidity-today').textContent = stats.humidityToday;
            if (el('stats-humidity-range')) el('stats-humidity-range').textContent = stats.humidityRange;
            
            // Barometer - split high/low
            if (el('stats-barometer-hi-today')) {
                const [hiToday, loToday] = stats.barometerToday.split(' / ');
                el('stats-barometer-hi-today').textContent = hiToday || '--';
                if (el('stats-barometer-lo-today')) el('stats-barometer-lo-today').textContent = loToday || '--';
            }
            if (el('stats-barometer-hi-range')) {
                const [hiRange, loRange] = stats.barometerRange.split(' / ');
                el('stats-barometer-hi-range').textContent = hiRange || '--';
                if (el('stats-barometer-lo-range')) el('stats-barometer-lo-range').textContent = loRange || '--';
            }
            
            // Wind Average
            if (el('stats-wind-avg-today')) el('stats-wind-avg-today').textContent = stats.windAvgToday;
            if (el('stats-wind-avg-range')) el('stats-wind-avg-range').textContent = stats.windAvgRange;
            
            // Wind Max - split speed and direction
            if (el('stats-wind-max-today')) {
                const [speedToday, dirToday] = stats.windMaxToday.split(' • ');
                el('stats-wind-max-today').textContent = speedToday || '--';
                if (el('stats-wind-max-dir-today')) el('stats-wind-max-dir-today').textContent = dirToday || '--';
            }
            if (el('stats-wind-max-range')) {
                const [speedRange, dirRange] = stats.windMaxRange.split(' • ');
                el('stats-wind-max-range').textContent = speedRange || '--';
                if (el('stats-wind-max-dir-range')) el('stats-wind-max-dir-range').textContent = dirRange || '--';
            }
            
            // Wind RMS
            if (el('stats-wind-rms-today')) el('stats-wind-rms-today').textContent = stats.windRmsToday;
            if (el('stats-wind-rms-range')) el('stats-wind-rms-range').textContent = stats.windRmsRange;
            
            // Wind Vector Speed
            if (el('stats-wind-vector-today')) el('stats-wind-vector-today').textContent = stats.windVectorToday;
            if (el('stats-wind-vector-range')) el('stats-wind-vector-range').textContent = stats.windVectorRange;
            
            // Wind Vector Direction
            if (el('stats-wind-vector-dir-today')) el('stats-wind-vector-dir-today').textContent = stats.windVectorDirToday;
            if (el('stats-wind-vector-dir-range')) el('stats-wind-vector-dir-range').textContent = stats.windVectorDirRange;
            
            // Rain Rate
            if (el('stats-rain-rate-today')) el('stats-rain-rate-today').textContent = stats.rainRateToday;
            if (el('stats-rain-rate-range')) el('stats-rain-rate-range').textContent = stats.rainRateRange;
            
            // Lightning Distance
            if (el('stats-lightning-distance-today')) el('stats-lightning-distance-today').textContent = stats.lightningDistToday;
            if (el('stats-lightning-distance-range')) el('stats-lightning-distance-range').textContent = stats.lightningDistRange;
            
            // Inside Temperature - split high/low
            if (el('stats-inside-temp-hi-today')) {
                const [hiToday, loToday] = stats.insideTempToday.split(' / ');
                el('stats-inside-temp-hi-today').textContent = hiToday || '--';
                if (el('stats-inside-temp-lo-today')) el('stats-inside-temp-lo-today').textContent = loToday || '--';
            }
            if (el('stats-inside-temp-hi-range')) {
                const [hiRange, loRange] = stats.insideTempRange.split(' / ');
                el('stats-inside-temp-hi-range').textContent = hiRange || '--';
                if (el('stats-inside-temp-lo-range')) el('stats-inside-temp-lo-range').textContent = loRange || '--';
            }
            
            // Inside Humidity - split high/low
            if (el('stats-inside-hum-hi-today')) {
                const [hiToday, loToday] = stats.insideHumToday.split(' / ');
                el('stats-inside-hum-hi-today').textContent = hiToday || '--';
                if (el('stats-inside-hum-lo-today')) el('stats-inside-hum-lo-today').textContent = loToday || '--';
            }
            if (el('stats-inside-hum-hi-range')) {
                const [hiRange, loRange] = stats.insideHumRange.split(' / ');
                el('stats-inside-hum-hi-range').textContent = hiRange || '--';
                if (el('stats-inside-hum-lo-range')) el('stats-inside-hum-lo-range').textContent = loRange || '--';
            }

            // Dynamically set units column labels
            const units = {
                'stats-rain-unit': 'in',
                'stats-strike-unit': '',
                'stats-temp-unit': '°F',
                'stats-feels-unit': '°F',
                'stats-windchill-unit': '°F',
                'stats-dew-unit': '°F',
                'stats-humidity-unit': '%',
                'stats-barometer-unit': 'inHg',
                'stats-wind-max-unit': 'mph',
                'stats-wind-avg-unit': 'mph',
                'stats-wind-rms-unit': 'mph',
                'stats-wind-vector-unit': 'mph',
                'stats-wind-vector-dir-unit': '°',
                'stats-rain-rate-unit': 'in/h',
                'stats-lightning-distance-unit': 'mi',
                'stats-inside-temp-unit': '°F',
                'stats-inside-hum-unit': '%'
            };
            for (const [id, text] of Object.entries(units)) {
                const unitEl = el(id);
                if (unitEl) unitEl.textContent = text;
            }

            // Update Range column header to current mode (Day/Week/Month)
            const rangeLabelEl = el('stats-range-label');
            if (rangeLabelEl) {
                rangeLabelEl.textContent = (range === 'day') ? 'Day' : (range === 'week') ? 'Week' : 'Month';
            }
        } catch (e) {
            console.warn('[Statistics] failed to compute', e);
        }
    }

    // Hook into existing lifecycle if available
    if (typeof window !== 'undefined') {
        const originalLoadAllCharts = window.loadAllCharts;
        if (typeof originalLoadAllCharts === 'function') {
            window.loadAllCharts = async function() {
                await originalLoadAllCharts();
                const range = window.currentRange || 'day';
                computeStatistics(range);
            };
        }

        const originalSetRange = window.setRange;
        if (typeof originalSetRange === 'function') {
            window.setRange = function(range) {
                originalSetRange(range);
                computeStatistics(range);
            };
        }

        // Initial compute on load if currentRange exists
        document.addEventListener('DOMContentLoaded', () => {
            const range = window.currentRange || 'day';
            computeStatistics(range);
        });
    }
})();
// ---------------------------------------------------------------------
// Build banner for quick cache/version verification
// ---------------------------------------------------------------------
(function(){
    const stamp = '2025-11-28T09:55Z';
    console.log('[Charts] build', stamp);
})();

// ---------------------------------------------------------------------
// Weather (temp + dew point) — DEFINES MASTER TIME GRID
// ---------------------------------------------------------------------
async function loadWeather() {
    const statusEl = document.getElementById('status');

    // Clear stale values and master timeline while we reload
    latestWeather = null;
    masterTimes = [];
    masterTimesMs = [];
    updateCurrentConditions();
    statusEl.textContent = 'Loading data (' + currentRange + ')...';

    try {
        const res = await fetch('/api/weather?range=' + encodeURIComponent(currentRange));
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const data = await res.json();
        if (!Array.isArray(data) || data.length === 0) {
            statusEl.textContent = 'No weather data for selected range.';
            latestWeather = null;
            masterTimes = [];
            masterTimesMs = [];
            updateCurrentConditions();
            return;
        }

        // Master timeline for all charts
        setMasterTimesFromRows(data);

        latestWeather = data[data.length - 1];

        const temps = alignToMasterTimes(data, r => r.temperature);
        const dews  = alignToMasterTimes(data, r => r.dewpoint);
        const labels = masterTimes.map(() => '');

        renderWeatherChart(labels, temps, dews, masterTimes);
        statusEl.textContent = 'Loaded ' + data.length + ' weather records (' + currentRange + ').';
        updateCurrentConditions();
    } catch (err) {
        console.error(err);
        statusEl.textContent = 'Error loading weather: ' + err.message;
        latestWeather = null;
        masterTimes = [];
        masterTimesMs = [];
        updateCurrentConditions();
    }
}

function renderWeatherChart(labels, temps, dews, times) {
    const ctx = document.getElementById('weatherChart').getContext('2d');
    if (weatherChart) weatherChart.destroy();

    weatherChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels,
            datasets: [
                {
                    label: 'Temperature (°F)',
                    data: temps,
                    borderWidth: 2.5,
                    borderColor: 'rgba(220, 38, 38, 1)',
                    pointRadius: 0,
                    yAxisID: 'y'
                },
                {
                    label: 'Dew Point (°F)',
                    data: dews,
                    borderWidth: 2,
                    borderColor: 'rgba(34, 211, 238, 0.95)',
                    backgroundColor: 'rgba(34, 211, 238, 0.04)',
                    pointRadius: 0,
                    borderDash: [4, 4],
                    yAxisID: 'y'
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'index', intersect: false },
            plugins: {
                legend: { display: true, labels: { usePointStyle: true } },
                tooltip: { enabled: true, displayColors: false },
                dayNightBackground: {
                    enabled: true,
                    times: times.map(t => t.getTime())
                }
            },
            scales: {
                y: {
                    type: 'linear',
                    position: 'left',
                    title: { display: true, text: 'Temperature / Dew Point (°F)' },
                    ticks: { maxTicksLimit: 7 },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' }
                },
                x: {
                    ticks: makeTimeTickOptions(times),
                    grid: { display: false }
                }
            }
        }
    });
}

// ---------------------------------------------------------------------
// Barometer (aligned to master timeline)
// ---------------------------------------------------------------------
async function loadBarometer() {
    // Clear stale barometer on reload
    latestBarometer = null;
    updateCurrentConditions();

    try {
        const res = await fetch('/api/barometer?range=' + encodeURIComponent(currentRange));
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const data = await res.json();
        if (!Array.isArray(data) || data.length === 0 || !hasMasterTimes()) {
            console.warn('No barometer data or no master timeline.');
            latestBarometer = null;
            updateCurrentConditions();
            return;
        }

        latestBarometer = data[data.length - 1];

        // Backend now computes trend, level, and forecast
        barometerTrend = latestBarometer.trend || 'steady';
        barometerLevel = latestBarometer.level || 'normal';
        barometerForecast = latestBarometer.forecast || 'conditions unchanged';
        console.log(`[Barometer] Level: ${barometerLevel} | Trend: ${barometerTrend} | Forecast: ${barometerForecast}`);

        const pressures = alignToMasterTimes(data, r => r.pressure);
        const labels    = masterTimes.map(() => '');

        renderBarometerChart(labels, pressures, masterTimes);
        updateCurrentConditions();
    } catch (err) {
        console.error('Error loading barometer:', err);
        latestBarometer = null;
        updateCurrentConditions();
    }
}

function renderBarometerChart(labels, pressures, times) {
    const ctx = document.getElementById('barometerChart').getContext('2d');
    if (barometerChart) barometerChart.destroy();

    barometerChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels,
            datasets: [
                {
                    label: 'Barometer (inHg)',
                    data: pressures,
                    borderWidth: 2.2,
                    borderColor: 'rgba(16, 185, 129, 0.95)',
                    backgroundColor: 'rgba(16, 185, 129, 0.04)',
                    pointRadius: 0
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'index', intersect: false },
            plugins: {
                legend: { display: true, labels: { usePointStyle: true } },
                tooltip: { enabled: true, displayColors: false },
                dayNightBackground: {
                    enabled: true,
                    times: times.map(t => t.getTime())
                }
            },
            scales: {
                y: {
                    title: { display: true, text: 'Barometric Pressure (inHg)' },
                    ticks: { maxTicksLimit: 6 },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' }
                },
                x: {
                    ticks: makeTimeTickOptions(times),
                    grid: { display: false }
                }
            }
        }
    });
}

// ---------------------------------------------------------------------
// Feels Like (aligned to master timeline)
// ---------------------------------------------------------------------
async function loadFeelsLike() {
    // Clear stale feels-like
    latestFeelsLike = null;
    updateCurrentConditions();

    try {
        const res = await fetch('/api/feelslike?range=' + encodeURIComponent(currentRange));
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const data = await res.json();
        if (!Array.isArray(data) || data.length === 0 || !hasMasterTimes()) {
            console.warn('No feels-like data or no master timeline.');
            latestFeelsLike = null;
            updateCurrentConditions();
            return;
        }

        latestFeelsLike = data[data.length - 1];

        const heatVals  = alignToMasterTimes(data, r => r.heatIndex);
        const chillVals = alignToMasterTimes(data, r => r.windChill);

        // Prefer backend-provided active source (heat/chill/air). Fallback to 'air'.
        let activeSource = (latestFeelsLike && latestFeelsLike.activeSource) ? latestFeelsLike.activeSource : 'air';

        const labels = masterTimes.map(() => '');
        renderFeelsLikeChart(labels, heatVals, chillVals, activeSource, masterTimes);
        updateCurrentConditions();
    } catch (err) {
        console.error('Error loading feels-like:', err);
        latestFeelsLike = null;
        updateCurrentConditions();
    }
}

function renderFeelsLikeChart(labels, heatVals, chillVals, activeSource, times) {
    const ctx = document.getElementById('feelsLikeChart').getContext('2d');
    if (feelsLikeChart) feelsLikeChart.destroy();

    const heatActive  = activeSource === 'heat';
    const chillActive = activeSource === 'chill';

    feelsLikeChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels,
            datasets: [
                {
                    label: 'Heat Index (°F)',
                    data: heatVals,
                    yAxisID: 'y',
                    borderWidth: heatActive ? 3 : 1.5,
                    borderColor: heatActive ? 'rgba(239, 68, 68, 1)' : 'rgba(239, 68, 68, 0.3)',
                    backgroundColor: heatActive ? 'rgba(239, 68, 68, 0.10)' : 'rgba(239, 68, 68, 0.02)',
                    pointRadius: heatActive ? 2 : 0,
                    pointHitRadius: 6
                },
                {
                    label: 'Wind Chill (°F)',
                    data: chillVals,
                    yAxisID: 'y',
                    borderWidth: chillActive ? 3 : 1.5,
                    borderColor: chillActive ? 'rgba(6, 182, 212, 1)' : 'rgba(6, 182, 212, 0.35)',
                    backgroundColor: chillActive ? 'rgba(6, 182, 212, 0.12)' : 'rgba(6, 182, 212, 0.02)',
                    borderDash: [5, 5],
                    pointRadius: chillActive ? 2 : 0,
                    pointHitRadius: 6
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'index', intersect: false },
            plugins: {
                legend: { display: true, labels: { usePointStyle: true } },
                tooltip: { enabled: true, displayColors: false },
                dayNightBackground: {
                    enabled: true,
                    times: times.map(t => t.getTime())
                }
            },
            scales: {
                y: {
                    type: 'linear',
                    position: 'left',
                    title: { display: true, text: 'Feels Like (°F)' },
                    ticks: { maxTicksLimit: 7 },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' }
                },
                x: {
                    ticks: makeTimeTickOptions(times),
                    grid: { display: false }
                }
            }
        }
    });
}

// ---------------------------------------------------------------------
// Humidity (aligned to master timeline)
// ---------------------------------------------------------------------
async function loadHumidity() {
    latestHumidity = null;
    updateCurrentConditions();

    try {
        const res = await fetch('/api/humidity?range=' + encodeURIComponent(currentRange));
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const data = await res.json();
        if (!Array.isArray(data) || data.length === 0 || !hasMasterTimes()) {
            console.warn('No humidity data or no master timeline.');
            latestHumidity = null;
            updateCurrentConditions();
            return;
        }

        latestHumidity = data[data.length - 1];

        const hums   = alignToMasterTimes(data, r => r.humidity);
        const labels = masterTimes.map(() => '');

        renderHumidityChart(labels, hums, masterTimes);
        updateCurrentConditions();
    } catch (err) {
        console.error('Error loading humidity:', err);
        latestHumidity = null;
        updateCurrentConditions();
    }
}

function renderHumidityChart(labels, hums, times) {
    const ctx = document.getElementById('humidityChart').getContext('2d');
    if (humidityChart) humidityChart.destroy();

    humidityChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels,
            datasets: [
                {
                    label: 'Humidity (%)',
                    data: hums,
                    borderWidth: 2.5,
                    borderColor: 'rgba(14, 165, 233, 0.95)',
                    backgroundColor: 'rgba(14, 165, 233, 0.08)',
                    tension: 0.25,
                    pointRadius: 0
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'index', intersect: false },
            plugins: {
                legend: { display: true, labels: { usePointStyle: true } },
                tooltip: { enabled: true, displayColors: false },
                dayNightBackground: {
                    enabled: true,
                    times: times.map(t => t.getTime())
                }
            },
            scales: {
                y: {
                    min: 0,
                    max: 100,
                    title: { display: true, text: 'Relative Humidity (%)' },
                    ticks: {
                        maxTicksLimit: 6,
                        callback: v => v + '%'
                    },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' }
                },
                x: {
                    ticks: makeTimeTickOptions(times),
                    grid: { display: false }
                }
            }
        }
    });
}

// WIND SECTION
async function loadWind() {
    // Clear out old state while loading
    latestWind = null;
    windStrong = false;
    updateCurrentConditions();

    try {
        const res = await fetch('/api/wind?range=' + encodeURIComponent(currentRange));
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const data = await res.json();
        if (!Array.isArray(data) || data.length === 0 || !hasMasterTimes()) {
            console.warn('No wind data or no master timeline.');
            latestWind = null;
            windStrong = false;
            updateCurrentConditions();
            return;
        }

        const speedsAligned = alignToMasterTimes(data, r => r.speed);
        const gustsAligned  = alignToMasterTimes(data, r => r.gust);
        const dirsAligned   = alignToMasterTimes(data, r => r.direction);

        // Use the *raw* data to set latestWind (last record with a valid speed)
        let last = null;
        for (let i = data.length - 1; i >= 0; i--) {
            const r = data[i];
            if (r && typeof r.speed === 'number') {
                last = r;
                break;
            }
        }

        if (last) {
            latestWind = {
                speed: last.speed,
                gust:  typeof last.gust === 'number' ? last.gust : null,
                direction: last.direction,
                compass: last.compass || '--',
                strong: last.strong || false
            };
        } else {
            latestWind = {
                speed: 0,
                gust:  null,
                direction: null,
                compass: '--',
                strong: false
            };
        }

        // Backend now computes strong flag
        windStrong = latestWind.strong;
        console.log('[Wind] Speed:', latestWind.speed.toFixed(1), 'mph | Gust:', (latestWind.gust || 0).toFixed(1), 'mph | Strong:', windStrong);

        // Update CC panel with new latestWind + windStrong
        updateCurrentConditions();

        const labels = masterTimes.map(() => '');
        renderWindChart(labels, speedsAligned, gustsAligned, masterTimes);
        renderWindVectorChart(speedsAligned, dirsAligned, masterTimes);

    } catch (err) {
        console.error('Error loading wind:', err);
        latestWind = null;
        windStrong = false;
        updateCurrentConditions();
    }
}

function renderWindChart(labels, speeds, gusts, times) {
    const ctx = document.getElementById('windChart').getContext('2d');
    if (windChart) windChart.destroy();

    windChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels,
            datasets: [
                {
                    label: 'Wind Speed (mph)',
                    data: speeds,
                    yAxisID: 'y',
                    borderWidth: 1,
                    borderColor: 'rgba(96, 165, 250, 0.95)',
                    backgroundColor: 'rgba(96, 165, 250, 0.05)',
                    pointRadius: 0
                },
                {
                    label: 'Wind Gust (mph)',
                    data: gusts,
                    yAxisID: 'y',
                    borderWidth: 1,
                    borderColor: 'rgba(37, 99, 235, 0.9)',
                    backgroundColor: 'rgba(37, 99, 235, 0.05)',
                    borderDash: [4, 4],
                    pointRadius: 0
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'index', intersect: false },
            plugins: {
                legend: { display: true, labels: { usePointStyle: true } },
                tooltip: { enabled: true, displayColors: false },
                dayNightBackground: {
                    enabled: true,
                    times: times.map(t => t.getTime())
                }
            },
            scales: {
                y: {
                    type: 'linear',
                    position: 'left',
                    title: { display: true, text: 'Wind (mph)' },
                    ticks: { maxTicksLimit: 7 },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' }
                },
                x: {
                    ticks: makeTimeTickOptions(times),
                    grid: { display: false }
                }
            }
        }
    });
}

function renderWindVectorChart(speeds, dirs, times) {
    const ctx = document.getElementById('windVectorChart').getContext('2d');
    if (windVectorChart) windVectorChart.destroy();

    // Keep same array structure as other charts (don't filter nulls)
    // This ensures day/night shading stays synchronized
    const avgVectors = speeds.map((speed, i) => {
        const dir = dirs[i];
        if (speed != null && !isNaN(speed) && dir != null && !isNaN(dir)) {
            return { speed, direction: dir };
        }
        return null;
    });
    const avgTimes = times;

    const maxAbsSpeed = avgVectors.filter(v => v).length ? Math.max(...avgVectors.filter(v => v).map(v => v.speed)) : 1;
    let displayMax = Math.ceil(maxAbsSpeed / 5) * 5;
    displayMax = Math.max(displayMax, 5);

    const dummyData  = avgTimes.map((_, i) => ({ x: i, y: 0 }));
    const axisLabels = avgTimes.map(() => '');

    windVectorChart = new Chart(ctx, {
        type: 'scatter',
        data: {
            labels: axisLabels,
            datasets: [
                {
                    label: 'Wind Vectors',
                    data: dummyData,
                    pointRadius: 0,
                    showLine: false
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'nearest', intersect: false },
            scales: {
                x: {
                    type: 'category',
                    labels: axisLabels,
                    ticks: makeTimeTickOptions(avgTimes),
                    grid: { display: false }
                },
                y: {
                    min: -displayMax,
                    max:  displayMax,
                    title: { display: true, text: 'Wind Vector (mph)' },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' },
                    ticks: {
                        stepSize: 5,
                        callback: v => v.toString()
                    }
                }
            },
            plugins: {
                legend: { display: false },
                tooltip: {
                    enabled: true,
                    callbacks: {
                        label(ctx) {
                            const i = ctx.dataIndex;
                            const vec = avgVectors[i];
                            if (!vec) return '';
                            return `Speed: ${vec.speed.toFixed(1)} mph, Dir: ${vec.direction.toFixed(0)}°`;
                        }
                    }
                },
                dayNightBackground: {
                    enabled: true,
                    times: avgTimes.map(t => t.getTime())
                },
                windVector: {
                    speeds: avgVectors.map(v => v ? v.speed : null),
                    directions: avgVectors.map(v => v ? v.direction : null),
                    maxSpeed: displayMax
                }
            }
        }
    });
}

// ---------------------------------------------------------------------
// Wind Direction scatter (aligned)
// ---------------------------------------------------------------------
async function loadWindDirection() {
    try {
        const res = await fetch('/api/wind?range=' + encodeURIComponent(currentRange));
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const data = await res.json();
        if (!Array.isArray(data) || data.length === 0 || !hasMasterTimes()) {
            console.warn('No wind direction data or no master timeline.');
            return;
        }

        const speedsAligned = alignToMasterTimes(data, r => r.speed);
        const dirsAligned   = alignToMasterTimes(data, r => r.direction);

        const labels = masterTimes.map(() => '');
        renderWindDirectionChart(labels, dirsAligned, speedsAligned, masterTimes);
    } catch (err) {
        console.error('Error loading wind direction:', err);
    }
}

function renderWindDirectionChart(labels, dirs, speeds, times) {
    const ctx = document.getElementById('windDirChart').getContext('2d');
    if (windDirChart) windDirChart.destroy();

    const scatterData = labels.map((lbl, i) => ({
        x: lbl,
        y: dirs[i],
        s: speeds[i]
    }));

    windDirChart = new Chart(ctx, {
        type: 'scatter',
        data: {
            datasets: [
                {
                    label: 'Wind Direction (°)',
                    data: scatterData,
                    borderWidth: 0,
                    backgroundColor: ctx => {
                        const speed = ctx.raw.s;
                        if (speed == null) return 'rgba(148, 163, 184, 0.4)';
                        if (speed < 3)  return 'rgba(148, 163, 184, 0.55)';
                        if (speed < 8)  return 'rgba(59, 130, 246, 0.85)';
                        if (speed < 15) return 'rgba(234, 179, 8, 0.9)';
                        if (speed < 25) return 'rgba(249, 115, 22, 0.9)';
                        return 'rgba(220, 38, 38, 0.95)';
                    },
                    pointRadius: ctx => {
                        const speed = ctx.raw.s;
                        return speed < 2 ? 2 : 3;
                    },
                    pointHoverRadius: 6
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'nearest', intersect: false },
            scales: {
                x: {
                    type: 'category',
                    labels,
                    ticks: makeTimeTickOptions(times),
                    grid: { display: false }
                },
                y: {
                    min: 0,
                    max: 360,
                    title: { display: true, text: 'Direction (°)' },
                    ticks: {
                        callback: v => v + '°',
                        stepSize: 45
                    },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' }
                }
            },
            plugins: {
                legend: { display: false },
                tooltip: {
                    callbacks: {
                        label(ctx) {
                            const d = ctx.raw.y;
                            const s = ctx.raw.s;
                            return `Dir: ${d}°, Speed: ${s} mph`;
                        }
                    }
                },
                dayNightBackground: {
                    enabled: true,
                    times: times.map(t => t.getTime())
                }
            }
        }
    });
}

// ---------------------------------------------------------------------
// Rain (aligned)
// ---------------------------------------------------------------------
async function loadRain() {
    try {
        const res = await fetch('/api/rain?range=' + encodeURIComponent(currentRange));
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const data = await res.json();

        // Compute Rain Today and current rate from data
        if (Array.isArray(data) && data.length > 0) {
            const now = new Date();
            const y = now.getFullYear();
            const m = now.getMonth();
            const d = now.getDate();

            let totalToday = 0;
            let lastRate = null;

            data.forEach((r, idx) => {
                if (!r) return;

                const t = new Date(r.timestamp);

                // Today total
                if (
                    t.getFullYear() === y &&
                    t.getMonth() === m &&
                    t.getDate() === d &&
                    typeof r.amount === 'number'
                ) {
                    totalToday += r.amount;
                }

                // Track latest rate (using last non-null record)
                if (typeof r.rate === 'number' && !Number.isNaN(r.rate)) {
                    lastRate = r.rate;
                }
            });

            // Backend now provides recentlyActive flag on latest reading
            const latestReading = data[data.length - 1];
            rainRecentlyActive = latestReading?.recentlyActive || false;

            rainToday = totalToday;
            latestRainRate = (typeof lastRate === 'number') ? lastRate : null;
            console.log('[Rain] Today total:', totalToday.toFixed(2), 'in | Recent rate:', (latestRainRate || 0).toFixed(3), 'in/hr | Recently active:', rainRecentlyActive);
        } else {
            rainToday = null;
            latestRainRate = null;
            rainRecentlyActive = false;
            console.log('[Rain] No rain data');
        }

        updateCurrentConditions();

        if (!Array.isArray(data) || data.length === 0 || !hasMasterTimes()) {
            console.warn('No rain data or no master timeline.');
            return;
        }

        const ratesAligned   = alignToMasterTimes(data, r => r.rate);
        const amountsAligned = alignToMasterTimes(data, r => r.amount);

        const labels = masterTimes.map(() => '');
        renderRainAmountChart(labels, amountsAligned, masterTimes);
        renderRainRateChart(labels, ratesAligned, masterTimes);
    } catch (err) {
        console.error('Error loading rain:', err);
        rainToday = null;
        latestRainRate = null;
        rainRecentlyActive = false;
        updateCurrentConditions();
    }
}

// Rain Amount (single-axis line chart)
function renderRainAmountChart(labels, amounts, times) {
    const ctx = document.getElementById('rainAmountChart').getContext('2d');
    if (rainAmountChart) rainAmountChart.destroy();

    rainAmountChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels,
            datasets: [
                {
                    label: 'Rain Amount',
                    data: amounts,
                    borderWidth: 2,
                    borderColor: 'rgba(6, 182, 212, 0.9)',
                    backgroundColor: 'rgba(6, 182, 212, 0.5)',
                    tension: 0.25,
                    pointRadius: 0
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'index', intersect: false },
            plugins: {
                legend: { display: true, labels: { usePointStyle: true } },
                tooltip: { enabled: true, displayColors: true },
                dayNightBackground: {
                    enabled: true,
                    times: times.map(t => t.getTime())
                }
            },
            scales: {
                y: {
                    type: 'linear',
                    position: 'left',
                    min: 0,
                    suggestedMax: undefined,
                    title: { display: true, text: 'Rain Amount' },
                    ticks: { maxTicksLimit: 6 },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' }
                },
                x: {
                    ticks: makeTimeTickOptions(times),
                    grid: { display: false }
                }
            }
        }
    });
}

// Rain Rate (separate bar chart)
function renderRainRateChart(labels, rates, times) {
    const ctx = document.getElementById('rainRateChart').getContext('2d');
    if (rainRateChart) rainRateChart.destroy();

    rainRateChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels,
            datasets: [
                {
                    label: 'Rain Rate',
                    data: rates,
                    borderWidth: 0,
                    backgroundColor: ctx => {
                        
                        const v = ctx.raw ?? ctx.parsed.y;
                        if (v == null || isNaN(v)) {
                            return 'rgba(0, 0, 0, 0)'; // truly missing
                        }
                        if (v <= 0) {
                            return 'rgba(148, 163, 184, 0.25)'; // light grey for 0 rate
                        }

                        // Adjust thresholds to your units (these assume in/hr)
                        if (v < 0.02)  return 'rgba(148, 163, 184, 0.6)'; // very light
                        if (v < 0.1)   return 'rgba(96, 165, 250, 0.8)';  // light
                        if (v < 0.3)   return 'rgba(37, 99, 235, 0.9)';   // moderate
                        if (v < 0.7)   return 'rgba(234, 179, 8, 0.9)';   // heavy
                        return 'rgba(220, 38, 38, 0.95)';                 // very heavy
                    }
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'index', intersect: false },
            plugins: {
                legend: { display: true, labels: { usePointStyle: true } },
                tooltip: { enabled: true, displayColors: true },
                dayNightBackground: {
                    enabled: true,
                    times: times.map(t => t.getTime())
                }
            },
                scales: {
                y: {
                    type: 'linear',
                    position: 'left',
                    title: { display: true, text: 'Rain Rate' },
                    ticks: { maxTicksLimit: 6 },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' }
                },
                x: {
                    ticks: makeTimeTickOptions(times),
                    grid: { display: false }
                }
            }
        }
    });
}
// ---------------------------------------------------------------------
// Lightning (stays on its own time scale: hourly/daily totals)
// ---------------------------------------------------------------------
async function loadLightning() {
    try {
        const res = await fetch('/api/lightning?range=' + encodeURIComponent(currentRange));
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const data = await res.json();

        // Compute lightningToday from data
        if (Array.isArray(data) && data.length > 0) {
            const now = new Date();
            const y = now.getFullYear();
            const m = now.getMonth();
            const d = now.getDate();

            let totalToday = 0;

            data.forEach(r => {
                if (!r || typeof r.strikes !== 'number') return;

                const t = new Date(r.timestamp);

                if (
                    t.getFullYear() === y &&
                    t.getMonth() === m &&
                    t.getDate() === d
                ) {
                    totalToday += r.strikes;
                }
            });

            // Backend now provides recentlyActive flag on latest reading
            const latestReading = data[data.length - 1];
            lightningRecentlyActive = latestReading?.recentlyActive || false;

            lightningToday = totalToday;
            console.log('[Lightning] Today total:', totalToday.toFixed(0), 'strikes | Recently active (10 min):', lightningRecentlyActive);
        } else {
            lightningToday = null;
            lightningRecentlyActive = false;
            console.log('[Lightning] No lightning data');
        }

        updateCurrentConditions();

        if (!Array.isArray(data) || data.length === 0) {
            console.warn('No lightning data for selected range.');
            if (lightningChart) {
                lightningChart.destroy();
                lightningChart = null;
            }
            return;
        }
        const times   = data.map(r => new Date(r.timestamp));
        const strikes = data.map(r => r.strikes || 0);

        if (currentRange === 'day') {
            // Build rolling 24-hour window from (now - 24h) to now, in 1-hour buckets
            const now = new Date();
            const startTime = new Date(now.getTime() - 24 * 60 * 60 * 1000);
            
            // Create 24 hourly buckets
            const hourlyBuckets = [];
            const hourlyTimes = [];
            for (let i = 0; i < 24; i++) {
                const bucketStart = new Date(startTime.getTime() + i * 60 * 60 * 1000);
                hourlyTimes.push(bucketStart);
                hourlyBuckets.push(0);
            }
            
            // Aggregate strikes into buckets
            times.forEach((t, i) => {
                const elapsed = t.getTime() - startTime.getTime();
                const bucketIndex = Math.floor(elapsed / (60 * 60 * 1000));
                if (bucketIndex >= 0 && bucketIndex < 24) {
                    hourlyBuckets[bucketIndex] += strikes[i];
                }
            });
            
            // Label every 2nd hour to reduce crowding
            const labels = hourlyBuckets.map((_, i) => {
                const h = hourlyTimes[i].getHours();
                if (i === 0 || i === 23 || h % 2 === 0) {
                    return h.toString().padStart(2,'0') + ':00';
                }
                return '';
            });
            
            renderLightningChart(labels, hourlyBuckets, 'day', hourlyTimes);
        } else {
            // Daily totals for week/month
            const dailyTotals = new Map();
            times.forEach((t, i) => {
                const key = new Date(t.getFullYear(), t.getMonth(), t.getDate()).getTime();
                dailyTotals.set(key, (dailyTotals.get(key) || 0) + strikes[i]);
            });
            const sortedKeys = Array.from(dailyTotals.keys()).sort((a,b)=>a-b);
            const values = sortedKeys.map(k => dailyTotals.get(k));
            const timesForBars = sortedKeys.map(ms => new Date(ms));
            const labels = sortedKeys.map(ms => {
                const dObj = new Date(ms);
                const dayNum = dObj.getDate();
                if (currentRange === 'week') {
                    // show all day numbers
                    return String(dayNum).padStart(2,'0');
                } else { // month
                    // match month tick strategy: only odd days
                    return dayNum % 2 === 1 ? String(dayNum).padStart(2,'0') : '';
                }
            });
            renderLightningChart(labels, values, currentRange, timesForBars);
        }
    } catch (err) {
        console.error('Error loading lightning:', err);
        lightningToday = null;
        lightningRecentlyActive = false;
        updateCurrentConditions();
    }
}

function renderLightningChart(labels, values, range, times) {
    const ctx = document.getElementById('lightningChart').getContext('2d');
    if (lightningChart) lightningChart.destroy();

    const maxValRaw = Math.max(...values, 0);
    let yMax;

    if (maxValRaw <= 0) {
        yMax = 1;
    } else if (range === 'day') {
        yMax = maxValRaw <= 10 ? 10 : Math.ceil(maxValRaw / 5) * 5;
    } else {
        yMax = Math.ceil(maxValRaw / 5) * 5;
    }

    lightningChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels,
            datasets: [
                {
                    label: range === 'day'
                        ? 'Lightning Strikes (hourly total)'
                        : 'Lightning Strikes (daily total)',
                    data: values,
                    borderWidth: 0,
                    backgroundColor: ctx => {
                        const v = ctx.raw ?? ctx.parsed.y;
                        if (v == null || isNaN(v)) return 'rgba(0,0,0,0)';
                        if (v === 0) return 'rgba(148, 163, 184, 0.2)';
                        return 'rgba(251, 191, 36, 0.85)';
                    },
                    borderRadius: 3,
                    barPercentage: range === 'day' ? 0.9 : 0.6,
                    categoryPercentage: range === 'day' ? 0.9 : 0.6
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'index', intersect: false },
            plugins: {
                legend: { display: true, labels: { usePointStyle: true } },
                tooltip: {
                    enabled: true,
                    callbacks: {
                        label(ctx) {
                            const v = ctx.parsed.y;
                            return `Strikes: ${v}`;
                        }
                    }
                },
                dayNightBackground: {
                    enabled: range === 'day',
                    times: times.map(t => t.getTime())
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    max: yMax,
                    title: {
                        display: true,
                        text: range === 'day'
                            ? 'Strikes per Hour'
                            : 'Strikes per Day'
                    },
                    ticks: { maxTicksLimit: 6, precision: 0 },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' }
                },
                x: {
                    type: 'category',
                    ticks: { maxTicksLimit: range === 'day' ? 8 : 7 },
                    grid: { display: false }
                }
            }
        }
    });
}// ---------------------------------------------------------------------
// Inside Temperature (aligned to master timeline)
// ---------------------------------------------------------------------
async function loadInsideTemp() {
    latestInsideTemp = null;
    updateCurrentConditions();

    try {
        const res = await fetch('/api/insideTemp?range=' + encodeURIComponent(currentRange));
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const data = await res.json();
        if (!Array.isArray(data) || data.length === 0 || !hasMasterTimes()) {
            console.warn('No inside temperature data or no master timeline.');
            latestInsideTemp = null;
            updateCurrentConditions();
            return;
        }

        latestInsideTemp = data[data.length - 1];

        const tempsAligned = alignToMasterTimes(data, r => r.inside_temp_f);
        const labels       = masterTimes.map(() => '');

        renderInsideTempChart(labels, tempsAligned, masterTimes);
        updateCurrentConditions();
    } catch (err) {
        console.error('Error loading inside temperature:', err);
        latestInsideTemp = null;
        updateCurrentConditions();
    }
}

function renderInsideTempChart(labels, temps, times) {
    const ctx = document.getElementById('insideTempChart').getContext('2d');
    if (insideTempChart) insideTempChart.destroy();

    insideTempChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels,
            datasets: [
                {
                    label: 'Inside Temperature (°F)',
                    data: temps,
                    borderWidth: 2.2,
                    borderColor: 'rgba(248, 113, 113, 1)',
                    backgroundColor: 'rgba(248, 113, 113, 0.06)',
                    pointRadius: 0,
                    tension: 0.25
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'index', intersect: false },
            plugins: {
                legend: { display: true, labels: { usePointStyle: true } },
                tooltip: { enabled: true, displayColors: false },
                dayNightBackground: {
                    enabled: true,
                    times: times.map(t => t.getTime())
                }
            },
            scales: {
                y: {
                    type: 'linear',
                    position: 'left',
                    title: { display: true, text: 'Inside Temperature (°F)' },
                    ticks: { maxTicksLimit: 7 },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' }
                },
                x: {
                    ticks: makeTimeTickOptions(times),
                    grid: { display: false }
                }
            }
        }
    });
}

// ---------------------------------------------------------------------
// Inside Humidity (aligned to master timeline)
// ---------------------------------------------------------------------
async function loadInsideHumidity() {
    latestInsideHumidity = null;
    updateCurrentConditions();

    try {
        const res = await fetch('/api/insideHumidity?range=' + encodeURIComponent(currentRange));
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const data = await res.json();
        if (!Array.isArray(data) || data.length === 0 || !hasMasterTimes()) {
            console.warn('No inside humidity data or no master timeline.');
            latestInsideHumidity = null;
            updateCurrentConditions();
            return;
        }

        latestInsideHumidity = data[data.length - 1];

        const inHums = alignToMasterTimes(data, r => r.inside_humidity);
        const labels = masterTimes.map(() => '');

        renderInsideHumidityChart(labels, inHums, masterTimes);
        updateCurrentConditions();
    } catch (err) {
        console.error('Error loading insideHumidity:', err);
        latestInsideHumidity = null;
        updateCurrentConditions();
    }
}

function renderInsideHumidityChart(labels, inHums, times) {
    const ctx = document.getElementById('insideHumidityChart').getContext('2d');
    if (insideHumidityChart) insideHumidityChart.destroy();

    insideHumidityChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels,
            datasets: [
                {
                    label: 'Inside Humidity (%)',
                    data: inHums,
                    borderWidth: 2.5,
                    borderColor: 'rgba(34, 211, 238, 0.95)',
                    backgroundColor: 'rgba(34, 211, 238, 0.08)',
                    tension: 0.25,
                    pointRadius: 0
                }
            ]
        },
        options: {
            responsive: true,
            interaction: { mode: 'index', intersect: false },
            plugins: {
                legend: { display: true, labels: { usePointStyle: true } },
                tooltip: { enabled: true, displayColors: false },
                dayNightBackground: {
                    enabled: true,
                    times: times.map(t => t.getTime())
                }
            },
            scales: {
                y: {
                    min: 0,
                    max: 100,
                    title: { display: true, text: 'Relative Humidity (%)' },
                    ticks: {
                        maxTicksLimit: 6,
                        callback: v => v + '%'
                    },
                    grid: { color: 'rgba(148, 163, 184, 0.25)' }
                },
                x: {
                    ticks: makeTimeTickOptions(times),
                    grid: { display: false }
                }
            }
        }
    });
}

// ---------------------------------------------------------------------
// Load everything (weather first to establish master timeline)
// ---------------------------------------------------------------------
// Flag to prevent overlapping loadAll() runs when polling
let isLoadAllRunning = false;
function loadAll() {
    return loadWeather()
        .then(() => {
            if (!hasMasterTimes()) return;

            return Promise.all([
                loadCelestial(),
                loadBarometer(),
                loadFeelsLike(),
                loadHumidity(),
                loadWind(),
                loadWindDirection(),
                loadRain(),
                loadLightning(),
                loadInsideTemp(),
                loadInsideHumidity()
            ]);
        })
        .catch(err => {
            console.error('Error in loadAll:', err);
        });
}

// Set up once DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // Range buttons
    document.querySelectorAll('.range-button').forEach(btn => {
        btn.addEventListener('click', () => setRange(btn.dataset.range));
    });

    // Optional reload button, if it exists
    const reloadBtn = document.getElementById('reloadBtn');
    if (reloadBtn) {
        reloadBtn.addEventListener('click', loadAll);
    }

    // Initial state
    setRange('day');  // this will call loadAll() via setRange

    // Client polling is disabled — only SSE (server push) triggers refreshes
    console.log('[Init] Client polling disabled — waiting for SSE updates from server');

    // Server-Sent Events: receive push updates from the server and trigger an immediate refresh
    if (window.EventSource) {
        try {
            const es = new EventSource('/api/stream');
            es.addEventListener('open', () => console.log('[SSE] connected to /api/stream'));
            es.addEventListener('error', (e) => console.warn('[SSE] error', e));

            // Keep track of the last epoch we acted on to avoid duplicate reloads.
            let lastSSEEpoch = null;

            es.addEventListener('update', async (ev) => {
                try {
                    const payload = JSON.parse(ev.data || '{}');
                    const ts = Number(payload.timestamp || payload.dateTime || 0);
                    if (!ts) {
                        console.log('[SSE] update received (no timestamp) — triggering safe refresh');
                    } else {
                        if (lastSSEEpoch && ts <= lastSSEEpoch) {
                            console.log('[SSE] update timestamp not newer (', ts, '<=', lastSSEEpoch, '); ignoring');
                            return;
                        }
                        lastSSEEpoch = ts;
                        console.log('[SSE] update received, new timestamp:', ts);
                    }
                } catch (e) {
                    console.warn('[SSE] failed to parse update payload, triggering refresh', e);
                }

                // Trigger a safe refresh; the poller also exists so this is best-effort.
                if (isLoadAllRunning) {
                    console.log('[SSE] loadAll already running; skipping SSE-triggered refresh');
                    return;
                }
                try {
                    isLoadAllRunning = true;
                    await loadAll();
                } catch (err) {
                    console.error('[SSE] Error during SSE-triggered loadAll():', err);
                } finally {
                    isLoadAllRunning = false;
                }
            });
        } catch (err) {
            console.warn('[SSE] failed to create EventSource:', err);
        }
    } else {
        console.warn('[SSE] EventSource not supported by this browser');
    }
});
