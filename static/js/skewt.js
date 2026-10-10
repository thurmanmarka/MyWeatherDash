// Local Skew-T Sounding Modal Handler
(function() {
    const modal = document.getElementById('skewtModal');
    const btnOpen = document.getElementById('skewtBtn');
    const btnClose = document.getElementById('skewtModalClose');
    const btnRefresh = document.getElementById('skewtRefresh');
    const statusEl = document.getElementById('skewtStatus');
    const footerEl = document.getElementById('skewtFooter');
    const canvas = document.getElementById('skewtCanvas');

    let cachedData = null;
    let loading = false;
    let resizeTimer = null;

    // Open modal
    btnOpen.addEventListener('click', () => {
        modal.style.display = 'block';
        if (cachedData) {
            drawSkewT(cachedData);
        }
        loadSounding(false);
    });

    // Close modal
    btnClose.addEventListener('click', () => { modal.style.display = 'none'; });
    window.addEventListener('click', (e) => {
        if (e.target === modal) modal.style.display = 'none';
    });

    btnRefresh.addEventListener('click', () => loadSounding(true));

    window.addEventListener('resize', () => {
        if (modal.style.display !== 'block' || !cachedData) return;
        clearTimeout(resizeTimer);
        resizeTimer = setTimeout(() => drawSkewT(cachedData), 150);
    });

    async function loadSounding(force) {
        if (loading) return;
        loading = true;
        statusEl.textContent = force ? 'Refreshing...' : (cachedData ? 'Refreshing...' : 'Loading...');
        btnRefresh.disabled = true;

        try {
            const url = force ? '/api/sounding?force=1' : '/api/sounding';
            const resp = await fetch(url);
            if (!resp.ok) {
                throw new Error(`HTTP ${resp.status}`);
            }
            const data = await resp.json();
            cachedData = data;

            const genTime = new Date(data.generatedAt);
            const timeLabel = isNaN(genTime.getTime()) ? data.generatedAt : genTime.toLocaleString();
            statusEl.textContent = `Profile valid: ${timeLabel}`;
            footerEl.textContent = `${data.source}. Temperature in °C, wind in knots, pressure levels in hPa. Refreshes hourly.`;

            drawSkewT(data);
        } catch (err) {
            statusEl.textContent = `Error loading sounding: ${err.message}`;
            console.error('[Skew-T] load failed:', err);
        } finally {
            loading = false;
            btnRefresh.disabled = false;
        }
    }

    function themeColor(varName, fallback) {
        const v = getComputedStyle(document.documentElement).getPropertyValue(varName).trim();
        return v || fallback;
    }

    // Total rightward shift (in degrees C, applied in "effective temperature" space —
    // see T2X below) from the bottom of the profile to the top. This is what gives a
    // skew-T its signature ~45 degree isotherms; it's a fixed convention, not derived
    // from the data.
    const SKEW_DEG_PER_UNIT_HEIGHT = 100;

    function drawSkewT(data) {
        if (!canvas.isConnected) return;
        const rect = canvas.getBoundingClientRect();
        if (rect.width === 0 || rect.height === 0) return;

        const dpr = window.devicePixelRatio || 1;
        canvas.width = Math.round(rect.width * dpr);
        canvas.height = Math.round(rect.height * dpr);

        const ctx = canvas.getContext('2d');
        ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

        const W = rect.width, H = rect.height;

        const textMain = themeColor('--text-main', '#1f2933');
        const textMuted = themeColor('--text-muted', '#6b7280');
        const panelBorder = themeColor('--panel-border', '#d2d7e0');
        const pageBg = themeColor('--page-bg', '#f5f7fb');

        ctx.fillStyle = pageBg;
        ctx.fillRect(0, 0, W, H);

        const padLeft = 60, padRight = 20, padTop = 24, padBottom = 36, windColW = 55;
        // Reserved, currently-blank strip to the right of the wind barb column for
        // future data panels (indices, hodograph, etc.) — keeps the skew-T diagram
        // itself from stretching to fill the whole canvas width.
        const reservedRight = 220;
        const plotLeft = padLeft;
        const plotWidth = Math.max(100, W - padLeft - padRight - windColW - reservedRight);
        const plotTop = padTop;
        const plotHeight = Math.max(100, H - padTop - padBottom);

        const levels = (data.levels || []).slice().sort((a, b) => b.pressureHPa - a.pressureHPa);
        if (levels.length < 2) {
            ctx.fillStyle = textMuted;
            ctx.textAlign = 'center';
            ctx.fillText('No sounding data available', W / 2, H / 2);
            return;
        }

        const Pbottom = Math.max(1050, Math.ceil(levels[0].pressureHPa / 50) * 50);
        const Ptop = Math.min(100, Math.floor(levels[levels.length - 1].pressureHPa / 50) * 50);

        function P2Y(p) {
            // frac=0 at the surface (Pbottom), frac=1 at the top of the profile (Ptop) —
            // surface belongs at the BOTTOM of the canvas, so invert here.
            const frac = (Math.log(Pbottom) - Math.log(p)) / (Math.log(Pbottom) - Math.log(Ptop));
            return plotTop + plotHeight - frac * plotHeight;
        }
        function heightFrac(y) {
            return (plotTop + plotHeight - y) / plotHeight; // 0 at surface, 1 at top of profile
        }
        // "Effective" temperature = raw temp plus the skew shift for its height. The
        // x-axis (Tmin/Tmax) is fit to the actual plotted data in this effective space,
        // so the trace always fills the available width instead of leaving the background
        // grid padded out to accommodate temperatures that, once skewed, never get near
        // the edges.
        function effectiveT(tempC, y) {
            return tempC + SKEW_DEG_PER_UNIT_HEIGHT * heightFrac(y);
        }

        let Tmin = -40, Tmax = 40;
        {
            let lo = Infinity, hi = -Infinity;
            levels.forEach(l => {
                const y = P2Y(l.pressureHPa);
                if (isFinite(l.tempC)) {
                    const et = effectiveT(l.tempC, y);
                    lo = Math.min(lo, et); hi = Math.max(hi, et);
                }
                if (isFinite(l.dewpointC)) {
                    const et = effectiveT(l.dewpointC, y);
                    lo = Math.min(lo, et); hi = Math.max(hi, et);
                }
            });
            if (isFinite(lo) && isFinite(hi)) {
                Tmin = Math.floor((lo - 5) / 10) * 10;
                Tmax = Math.ceil((hi + 5) / 10) * 10;
            }
        }

        function T2X(tempC, y) {
            const et = effectiveT(tempC, y);
            return plotLeft + ((et - Tmin) / (Tmax - Tmin)) * plotWidth;
        }

        ctx.save();
        ctx.beginPath();
        ctx.rect(plotLeft, plotTop, plotWidth, plotHeight);
        ctx.clip();

        // Isobars
        const isobarLevels = [1000, 900, 850, 800, 700, 600, 500, 400, 300, 250, 200, 150, 100];
        ctx.strokeStyle = panelBorder;
        ctx.lineWidth = 1;
        isobarLevels.forEach(p => {
            if (p > Pbottom || p < Ptop) return;
            const y = P2Y(p);
            ctx.beginPath();
            ctx.moveTo(plotLeft, y);
            ctx.lineTo(plotLeft + plotWidth, y);
            ctx.stroke();
        });

        // Isotherms (skewed lines, every 10C). Raw temperatures below Tmin can still be
        // visible near the top of the plot (where the skew shift is largest), so extend
        // the loop's lower bound by the full skew range.
        ctx.strokeStyle = 'rgba(100, 140, 210, 0.45)';
        const isothermLo = Math.floor((Tmin - SKEW_DEG_PER_UNIT_HEIGHT) / 10) * 10;
        for (let t = isothermLo; t <= Tmax; t += 10) {
            const yTop = plotTop, yBot = plotTop + plotHeight;
            ctx.beginPath();
            ctx.moveTo(T2X(t, yBot), yBot);
            ctx.lineTo(T2X(t, yTop), yTop);
            ctx.stroke();
        }

        // Dry adiabats (constant potential temperature), every 10C
        ctx.strokeStyle = 'rgba(205, 130, 50, 0.45)';
        for (let theta = -40; theta <= 120; theta += 10) {
            ctx.beginPath();
            let first = true;
            for (let p = Pbottom; p >= Ptop; p -= 10) {
                const thetaK = theta + 273.15;
                const tK = thetaK * Math.pow(p / 1000, 0.286);
                const tC = tK - 273.15;
                const y = P2Y(p);
                const x = T2X(tC, y);
                if (first) { ctx.moveTo(x, y); first = false; } else { ctx.lineTo(x, y); }
            }
            ctx.stroke();
        }

        // Saturation mixing ratio lines (g/kg) — only meaningful in the lower/moister
        // part of the profile, so (as on standard skew-T charts) stop drawing them
        // above ~400 hPa rather than cluttering the dry upper troposphere.
        const mixingRatioTop = Math.max(Ptop, 400);
        ctx.setLineDash([2, 3]);
        ctx.strokeStyle = 'rgba(60, 160, 90, 0.6)';
        [1, 2, 4, 7, 10, 16, 24, 32].forEach(w => {
            ctx.beginPath();
            let first = true;
            for (let p = Pbottom; p >= mixingRatioTop; p -= 25) {
                const e = (w * p) / (622 + w);
                if (e <= 0) continue;
                const a = 17.625, b = 243.04;
                const lnE = Math.log(e / 6.112);
                const tC = (b * lnE) / (a - lnE);
                const y = P2Y(p);
                const x = T2X(tC, y);
                if (first) { ctx.moveTo(x, y); first = false; } else { ctx.lineTo(x, y); }
            }
            ctx.stroke();
        });
        ctx.setLineDash([]);

        // Temperature / dewpoint traces
        function drawTrace(key, color) {
            ctx.strokeStyle = color;
            ctx.lineWidth = 2.5;
            ctx.beginPath();
            let first = true;
            levels.forEach(l => {
                const v = l[key];
                if (!isFinite(v)) return;
                const y = P2Y(l.pressureHPa);
                const x = T2X(v, y);
                if (first) { ctx.moveTo(x, y); first = false; } else { ctx.lineTo(x, y); }
            });
            ctx.stroke();
        }
        drawTrace('tempC', '#e53e3e');
        drawTrace('dewpointC', '#2b8a3e');

        ctx.restore();

        // Plot border
        ctx.strokeStyle = textMuted;
        ctx.lineWidth = 1;
        ctx.strokeRect(plotLeft, plotTop, plotWidth, plotHeight);

        // Isobar labels
        ctx.fillStyle = textMuted;
        ctx.font = '11px system-ui, sans-serif';
        ctx.textAlign = 'right';
        ctx.textBaseline = 'middle';
        isobarLevels.forEach(p => {
            if (p > Pbottom || p < Ptop) return;
            ctx.fillText(p + ' hPa', plotLeft - 8, P2Y(p));
        });

        // Isotherm labels along the bottom axis
        ctx.textAlign = 'center';
        ctx.textBaseline = 'top';
        for (let t = Tmin; t <= Tmax; t += 10) {
            const y = plotTop + plotHeight;
            const x = T2X(t, y);
            if (x >= plotLeft - 5 && x <= plotLeft + plotWidth + 5) {
                ctx.fillText(t + '°C', x, y + 6);
            }
        }

        // Legend
        ctx.textAlign = 'left';
        ctx.textBaseline = 'alphabetic';
        ctx.font = '12px system-ui, sans-serif';
        ctx.fillStyle = '#e53e3e';
        ctx.fillText('— Temperature', plotLeft + 8, plotTop + 14);
        ctx.fillStyle = '#2b8a3e';
        ctx.fillText('— Dewpoint', plotLeft + 125, plotTop + 14);
        ctx.fillStyle = textMuted;
        ctx.fillText('Wind (kt)', plotLeft + plotWidth + windColW / 2 - 24, plotTop - 8);

        // Wind barb column
        const windColX = plotLeft + plotWidth + windColW / 2;
        ctx.strokeStyle = panelBorder;
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(plotLeft + plotWidth + 10, plotTop);
        ctx.lineTo(plotLeft + plotWidth + 10, plotTop + plotHeight);
        ctx.stroke();

        let lastY = null;
        levels.forEach(l => {
            const y = P2Y(l.pressureHPa);
            if (lastY !== null && Math.abs(y - lastY) < 20) return;
            lastY = y;
            drawWindBarb(ctx, windColX, y, l.windSpeedKt, l.windDirDeg, textMain);
        });

        // Parcel indices panel, in the reserved strip to the right of the wind barbs
        const panelLeft = plotLeft + plotWidth + windColW + 16;
        const panelWidth = (W - padRight) - panelLeft;
        if (data.parcel && panelWidth > 70) {
            drawParcelPanel(ctx, data.parcel, panelLeft, plotTop, panelWidth, textMain, textMuted, panelBorder);
        }
    }

    function metersToFeet(m) {
        return m * 3.28084;
    }

    function drawParcelPanel(ctx, parcel, x, yTop, width, textMain, textMuted, panelBorder) {
        let y = yTop;
        const lineH = 17;

        function heading(text) {
            ctx.font = 'bold 12px system-ui, sans-serif';
            ctx.fillStyle = textMain;
            ctx.textAlign = 'left';
            ctx.textBaseline = 'top';
            ctx.fillText(text, x, y);
            y += lineH + 2;
        }
        function row(label, value) {
            ctx.font = '11px system-ui, sans-serif';
            ctx.fillStyle = textMuted;
            ctx.textAlign = 'left';
            ctx.textBaseline = 'top';
            ctx.fillText(label, x, y);
            ctx.fillStyle = textMain;
            ctx.textAlign = 'right';
            ctx.fillText(value, x + width, y);
            y += lineH;
        }
        function divider() {
            y += 4;
            ctx.strokeStyle = panelBorder;
            ctx.lineWidth = 1;
            ctx.beginPath();
            ctx.moveTo(x, y);
            ctx.lineTo(x + width, y);
            ctx.stroke();
            y += 8;
        }

        heading('Surface Parcel');
        row('CAPE', Math.round(parcel.capeJkg || 0) + ' J/kg');
        row('CIN', Math.round(parcel.cinJkg || 0) + ' J/kg');

        divider();
        row('LCL', Math.round(parcel.lclPressureHPa) + ' hPa');
        row('', Math.round(metersToFeet(parcel.lclHeightM)).toLocaleString() + ' ft');
        if (parcel.lfcPressureHPa) {
            row('LFC', Math.round(parcel.lfcPressureHPa) + ' hPa');
            row('', Math.round(metersToFeet(parcel.lfcHeightM)).toLocaleString() + ' ft');
        } else {
            row('LFC', 'none');
        }
        if (parcel.elPressureHPa) {
            row('EL', Math.round(parcel.elPressureHPa) + ' hPa');
            row('', Math.round(metersToFeet(parcel.elHeightM)).toLocaleString() + ' ft');
        } else {
            row('EL', 'none');
        }

        divider();
        row('PWAT', (parcel.precipitableWaterIn || 0).toFixed(2) + ' in');

        divider();
        heading('Lapse Rates');
        row('Sfc-3km', (parcel.lapseRate0to3kmCPerKm || 0).toFixed(1) + ' °C/km');
        row('3-6km', (parcel.lapseRate3to6kmCPerKm || 0).toFixed(1) + ' °C/km');
        row('700-500mb', (parcel.lapseRate700to500CPerKm || 0).toFixed(1) + ' °C/km');
    }

    function drawWindBarb(ctx, x, y, speedKt, dirDeg, color) {
        ctx.strokeStyle = color;
        ctx.fillStyle = color;
        ctx.lineWidth = 1.3;

        ctx.beginPath();
        ctx.arc(x, y, 2, 0, Math.PI * 2);
        ctx.stroke();

        if (!isFinite(speedKt) || !isFinite(dirDeg) || speedKt < 2.5) {
            return;
        }

        const rad = (dirDeg * Math.PI) / 180;
        // Unit vector pointing toward the direction the wind is blowing FROM
        const dx = Math.sin(rad), dy = -Math.cos(rad);
        const px = -dy, py = dx; // perpendicular, for barb feathers

        const shaftLen = 26;
        const ex = x + dx * shaftLen;
        const ey = y + dy * shaftLen;

        ctx.beginPath();
        ctx.moveTo(x, y);
        ctx.lineTo(ex, ey);
        ctx.stroke();

        let remaining = Math.round(speedKt / 5) * 5;
        let pos = 1.0;
        const step = 0.18;

        while (remaining >= 50 && pos > 0) {
            const baseX = x + dx * shaftLen * pos;
            const baseY = y + dy * shaftLen * pos;
            const innerX = x + dx * shaftLen * (pos - step);
            const innerY = y + dy * shaftLen * (pos - step);
            const tipX = baseX + px * 9;
            const tipY = baseY + py * 9;
            ctx.beginPath();
            ctx.moveTo(baseX, baseY);
            ctx.lineTo(tipX, tipY);
            ctx.lineTo(innerX, innerY);
            ctx.closePath();
            ctx.fill();
            pos -= step;
            remaining -= 50;
        }
        while (remaining >= 10 && pos > 0) {
            const baseX = x + dx * shaftLen * pos;
            const baseY = y + dy * shaftLen * pos;
            const tipX = baseX + px * 9;
            const tipY = baseY + py * 9;
            ctx.beginPath();
            ctx.moveTo(baseX, baseY);
            ctx.lineTo(tipX, tipY);
            ctx.stroke();
            pos -= step;
            remaining -= 10;
        }
        if (remaining >= 5 && pos > 0) {
            const baseX = x + dx * shaftLen * pos;
            const baseY = y + dy * shaftLen * pos;
            const tipX = baseX + px * 5;
            const tipY = baseY + py * 5;
            ctx.beginPath();
            ctx.moveTo(baseX, baseY);
            ctx.lineTo(tipX, tipY);
            ctx.stroke();
        }
    }
})();
