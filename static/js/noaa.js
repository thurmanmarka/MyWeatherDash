// NOAA Reports Modal Handler
(function() {
    let currentContent = '';
    let currentFilename = '';

    const modal = document.getElementById('noaaModal');
    const btnOpen = document.getElementById('noaaBtn');
    const btnClose = document.querySelector('.noaa-modal-close');
    const typeSelect = document.getElementById('noaaType');
    const yearSelect = document.getElementById('noaaYear');
    const monthSelect = document.getElementById('noaaMonth');
    const btnLoad = document.getElementById('noaaLoad');
    const btnDownload = document.getElementById('noaaDownload');
    const btnRecompile = document.getElementById('noaaRecompile');
    const contentEl = document.getElementById('noaaContent');

    // Populate year dropdown (current year - 5 to current year)
    const currentYear = new Date().getFullYear();
    for (let y = currentYear; y >= currentYear - 5; y--) {
        const opt = document.createElement('option');
        opt.value = y;
        opt.textContent = y;
        if (y === currentYear) opt.selected = true;
        yearSelect.appendChild(opt);
    }

    // Set current month as default
    const currentMonth = new Date().getMonth() + 1;
    monthSelect.value = currentMonth;

    // Show/hide month selector based on type
    typeSelect.addEventListener('change', () => {
        monthSelect.style.display = typeSelect.value === 'monthly' ? 'inline-block' : 'none';
    });

    // Open modal
    btnOpen.addEventListener('click', () => {
        modal.style.display = 'block';
    });

    // Close modal
    btnClose.addEventListener('click', () => {
        modal.style.display = 'none';
    });

    // Close on outside click
    window.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.style.display = 'none';
        }
    });

    // Load report
    btnLoad.addEventListener('click', async () => {
        const type = typeSelect.value;
        const year = yearSelect.value;
        const month = monthSelect.value;

        contentEl.textContent = 'Loading...';
        btnLoad.disabled = true;

        try {
            let url = `/api/noaa/${type}?year=${year}`;
            if (type === 'monthly') {
                url += `&month=${month}`;
                currentFilename = `NOAA-${year}-${month.padStart(2, '0')}.txt`;
            } else {
                currentFilename = `NOAA-${year}.txt`;
            }

            const response = await fetch(url);
            if (!response.ok) {
                if (response.status === 500) {
                    contentEl.textContent = 'No data available for this period.\n\nThe weather station may not have been collecting data during the requested time range.';
                } else {
                    throw new Error(`HTTP ${response.status}`);
                }
                currentContent = '';
                return;
            }
            
            currentContent = await response.text();
            contentEl.textContent = currentContent;
        } catch (err) {
            contentEl.textContent = `Error loading report: ${err.message}`;
            currentContent = '';
            console.error('[NOAA] Load failed:', err);
        } finally {
            btnLoad.disabled = false;
        }
    });

    // Download report
    btnDownload.addEventListener('click', () => {
        if (!currentContent) {
            alert('Please load a report first');
            return;
        }

        const blob = new Blob([currentContent], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = currentFilename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    });

    // Force recompile report (deletes cached file and regenerates)
    btnRecompile.addEventListener('click', async () => {
        const type = typeSelect.value;
        const year = yearSelect.value;
        const month = monthSelect.value;

        if (!confirm('Force recompile? This will delete the cached report and regenerate it.')) {
            return;
        }

        contentEl.textContent = 'Recompiling...';
        btnRecompile.disabled = true;

        try {
            let url = `/api/noaa/${type}?year=${year}&force=1`;
            if (type === 'monthly') {
                url += `&month=${month}`;
                currentFilename = `NOAA-${year}-${month.padStart(2, '0')}.txt`;
            } else {
                currentFilename = `NOAA-${year}.txt`;
            }

            const response = await fetch(url);
            if (!response.ok) {
                if (response.status === 500) {
                    contentEl.textContent = 'No data available for this period.\n\nThe weather station may not have been collecting data during the requested time range.';
                } else {
                    throw new Error(`HTTP ${response.status}`);
                }
                currentContent = '';
                return;
            }
            
            currentContent = await response.text();
            contentEl.textContent = currentContent;
        } catch (err) {
            contentEl.textContent = `Error recompiling report: ${err.message}`;
            currentContent = '';
            console.error('[NOAA] Recompile failed:', err);
        } finally {
            btnRecompile.disabled = false;
        }
    });

    // Initialize month visibility
    monthSelect.style.display = typeSelect.value === 'monthly' ? 'inline-block' : 'none';
})();
