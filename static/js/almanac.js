// Almanac Modal Handler (sun/moon + day-of-year climate records)
(function() {
    const modal = document.getElementById('almanacModal');
    const btnOpen = document.getElementById('almanacBtn');
    const btnClose = document.getElementById('almanacModalClose');
    const btnLoad = document.getElementById('almanacLoad');
    const btnToday = document.getElementById('almanacToday');
    const dateInput = document.getElementById('almanacDate');
    const bodyEl = document.getElementById('almanacBody');

    const MOON_EMOJI = {
        'New Moon': '🌑',
        'Waxing Crescent': '🌒',
        'First Quarter': '🌓',
        'Waxing Gibbous': '🌔',
        'Full Moon': '🌕',
        'Waning Gibbous': '🌖',
        'Last Quarter': '🌗',
        'Waning Crescent': '🌘',
    };

    function todayStr() {
        const d = new Date();
        const y = d.getFullYear();
        const m = String(d.getMonth() + 1).padStart(2, '0');
        const day = String(d.getDate()).padStart(2, '0');
        return `${y}-${m}-${day}`;
    }

    // Open modal
    btnOpen.addEventListener('click', () => {
        modal.style.display = 'block';
        if (!dateInput.value) dateInput.value = todayStr();
        loadAlmanac(dateInput.value);
    });

    // Close modal
    btnClose.addEventListener('click', () => { modal.style.display = 'none'; });
    window.addEventListener('click', (e) => {
        if (e.target === modal) modal.style.display = 'none';
    });

    btnLoad.addEventListener('click', () => {
        if (!dateInput.value) return;
        loadAlmanac(dateInput.value);
    });

    btnToday.addEventListener('click', () => {
        dateInput.value = todayStr();
        loadAlmanac(dateInput.value);
    });

    async function loadAlmanac(dateStr) {
        bodyEl.innerHTML = '<p style="color: var(--text-muted);">Loading...</p>';
        btnLoad.disabled = true;
        try {
            const resp = await fetch(`/api/almanac?date=${encodeURIComponent(dateStr)}`);
            if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
            const data = await resp.json();
            renderAlmanac(data, dateStr);
        } catch (err) {
            bodyEl.innerHTML = `<p style="color: var(--text-muted);">Error loading almanac: ${escapeHtml(err.message)}</p>`;
            console.error('[Almanac] load failed:', err);
        } finally {
            btnLoad.disabled = false;
        }
    }

    function escapeHtml(s) {
        const div = document.createElement('div');
        div.textContent = String(s);
        return div.innerHTML;
    }

    function fmtTemp(v) {
        return (typeof v === 'number' && isFinite(v)) ? v.toFixed(1) + ' °F' : '--';
    }

    function fmtRain(v) {
        return (typeof v === 'number' && isFinite(v)) ? v.toFixed(2) + ' in' : '--';
    }

    function row(label, value) {
        return `<div class="almanac-row"><span>${label}</span><span>${value || '--'}</span></div>`;
    }

    function renderAlmanac(data, dateStr) {
        const c = data.celestial || {};
        const moon = c.moonPhase || {};
        const moonEmoji = MOON_EMOJI[moon.name] || '🌙';

        let html = '<div class="almanac-grid">';

        html += `<div class="almanac-card">
            <h3>☀️ Sun</h3>
            ${row('Sunrise', c.sunrise24)}
            ${row('Sunset', c.sunset24)}
            ${row('Daylight', c.daylightHoursFormatted)}
        </div>`;

        html += `<div class="almanac-card">
            <h3>${moonEmoji} Moon</h3>
            ${row('Phase', moon.name)}
            ${row('Illumination', typeof moon.percentage === 'number' ? moon.percentage.toFixed(0) + '%' : '--')}
            ${row('Moonrise', c.moonrise24)}
            ${row('Moonset', c.moonset24)}
        </div>`;

        html += `<div class="almanac-card">
            <h3>🌆 Twilight</h3>
            ${row('Civil dawn / dusk', (c.civilDawn24 || '--') + ' / ' + (c.civilDusk24 || '--'))}
            ${row('Nautical dawn / dusk', (c.nauticalDawn24 || '--') + ' / ' + (c.nauticalDusk24 || '--'))}
            ${row('Astronomical dawn / dusk', (c.astronomicalDawn24 || '--') + ' / ' + (c.astronomicalDusk24 || '--'))}
        </div>`;

        html += `<div class="almanac-card">
            <h3>📸 Golden / Blue Hour</h3>
            ${row('Morning golden hour', (c.goldenHourMorningStart24 || '--') + ' – ' + (c.goldenHourMorningEnd24 || '--'))}
            ${row('Evening golden hour', (c.goldenHourEveningStart24 || '--') + ' – ' + (c.goldenHourEveningEnd24 || '--'))}
            ${row('Evening blue hour', (c.blueHourEveningStart24 || '--') + ' – ' + (c.blueHourEveningEnd24 || '--'))}
        </div>`;

        if (data.yearsOfData > 0) {
            html += `<div class="almanac-card">
                <h3>📊 Records for this date</h3>
                ${row('Record high', fmtTemp(data.recordHighTemp) + (data.recordHighYear ? ` (${data.recordHighYear})` : ''))}
                ${row('Record low', fmtTemp(data.recordLowTemp) + (data.recordLowYear ? ` (${data.recordLowYear})` : ''))}
                ${row('Average high', fmtTemp(data.avgHighTemp))}
                ${row('Average low', fmtTemp(data.avgLowTemp))}
                ${row('Years of station data', data.yearsOfData)}
            </div>`;
        }

        html += '</div>'; // .almanac-grid

        if (Array.isArray(data.history) && data.history.length > 0) {
            html += '<table class="almanac-history-table"><thead><tr>' +
                '<th>Year</th><th>High</th><th>Low</th><th>Rain</th>' +
                '</tr></thead><tbody>';
            data.history.forEach(h => {
                html += `<tr><td>${h.year}</td><td>${fmtTemp(h.highTemp)}</td><td>${fmtTemp(h.lowTemp)}</td><td>${fmtRain(h.rainTotal)}</td></tr>`;
            });
            html += '</tbody></table>';
        } else {
            html += '<p style="color: var(--text-muted); font-size: 0.85rem;">No station history available for this day yet.</p>';
        }

        bodyEl.innerHTML = html;
    }
})();
