// Easy Button - CSV Export for Daily Statistics
(function() {
    const modal = document.getElementById('easyModal');
    const btn = document.getElementById('easyBtn');
    const closeBtn = document.getElementById('easyModalClose');
    const downloadBtn = document.getElementById('easyDownload');
    const startYearSelect = document.getElementById('easyStartYear');
    const startMonthSelect = document.getElementById('easyStartMonth');
    const startDaySelect = document.getElementById('easyStartDay');
    const endYearSelect = document.getElementById('easyEndYear');
    const endMonthSelect = document.getElementById('easyEndMonth');
    const endDaySelect = document.getElementById('easyEndDay');
    const columnContainer = document.getElementById('easyColumnContainer');
    const addColumnBtn = document.getElementById('easyAddColumn');
    const removeColumnBtn = document.getElementById('easyRemoveColumn');

    // Available columns for selection
    const availableColumns = [
        { value: 'outTemp', label: 'Temperature' },
        { value: 'dewpoint', label: 'Dewpoint' },
        { value: 'outHumidity', label: 'Humidity' },
        { value: 'barometer', label: 'Barometer' },
        { value: 'heatindex', label: 'Heat Index' },
        { value: 'windchill', label: 'Wind Chill' },
        { value: 'windSpeed', label: 'Wind Speed' },
        { value: 'windGust', label: 'Wind Gust' },
        { value: 'windDir', label: 'Wind Direction' },
        { value: 'rainRate', label: 'Rain Rate' },
        { value: 'rain', label: 'Rain Amount' },
        { value: 'lightning_strike_count', label: 'Lightning Strikes' },
        { value: 'lightning_distance', label: 'Lightning Distance' },
        { value: 'inTemp', label: 'Inside Temperature' },
        { value: 'inHumidity', label: 'Inside Humidity' }
    ];

    // Track column dropdowns
    let columnCount = 0;

    // Create a column dropdown
    function createColumnDropdown() {
        columnCount++;
        const div = document.createElement('div');
        div.className = 'easy-column-row';
        div.setAttribute('data-column-id', columnCount);
        div.style.display = 'flex';
        div.style.alignItems = 'center';
        div.style.gap = '8px';

        const label = document.createElement('span');
        label.style.fontSize = '0.85rem';
        label.style.color = 'var(--text-muted)';
        label.style.minWidth = '80px';
        label.textContent = `Column ${columnCount + 1}:`;

        const select = document.createElement('select');
        select.style.flex = '1';
        select.style.padding = '6px';
        select.style.borderRadius = '4px';
        select.style.border = '1px solid var(--panel-border)';
        select.style.background = 'var(--panel-bg)';
        select.style.color = 'var(--text-main)';
        select.style.cursor = 'pointer';

        // Add placeholder option
        const placeholderOpt = document.createElement('option');
        placeholderOpt.value = '';
        placeholderOpt.textContent = '-- Select Column --';
        select.appendChild(placeholderOpt);

        // Add all available columns
        availableColumns.forEach(col => {
            const opt = document.createElement('option');
            opt.value = col.value;
            opt.textContent = col.label;
            select.appendChild(opt);
        });

        div.appendChild(label);
        div.appendChild(select);
        columnContainer.appendChild(div);

        // Update labels when columns change
        updateColumnLabels();
    }

    // Update column labels (Column 2, Column 3, etc.)
    function updateColumnLabels() {
        const rows = columnContainer.querySelectorAll('.easy-column-row');
        rows.forEach((row, index) => {
            const label = row.querySelector('span');
            label.textContent = `Column ${index + 2}:`;
        });
    }

    // Add column button
    addColumnBtn.addEventListener('click', () => {
        if (columnCount < 15) { // Max 15 additional columns
            createColumnDropdown();
        } else {
            alert('Maximum of 15 data columns reached');
        }
    });

    // Remove column button
    removeColumnBtn.addEventListener('click', () => {
        const rows = columnContainer.querySelectorAll('.easy-column-row');
        if (rows.length > 0) {
            columnContainer.removeChild(rows[rows.length - 1]);
            columnCount--;
            updateColumnLabels();
        }
    });

    // Initialize with default columns (Temperature, Humidity)
    createColumnDropdown();
    createColumnDropdown();
    const defaultRows = columnContainer.querySelectorAll('.easy-column-row select');
    if (defaultRows[0]) defaultRows[0].value = 'outTemp';
    if (defaultRows[1]) defaultRows[1].value = 'outHumidity';

    // Populate year dropdowns (2020 to current year)
    const currentYear = new Date().getFullYear();
    for (let y = currentYear; y >= 2020; y--) {
        const startOpt = document.createElement('option');
        startOpt.value = y;
        startOpt.textContent = y;
        startYearSelect.appendChild(startOpt);
        
        const endOpt = document.createElement('option');
        endOpt.value = y;
        endOpt.textContent = y;
        endYearSelect.appendChild(endOpt);
    }

    // Populate day dropdown based on selected month/year
    function updateStartDays() {
        const year = parseInt(startYearSelect.value);
        const month = parseInt(startMonthSelect.value);
        const daysInMonth = new Date(year, month, 0).getDate();
        
        startDaySelect.innerHTML = '';
        for (let d = 1; d <= daysInMonth; d++) {
            const opt = document.createElement('option');
            opt.value = d;
            opt.textContent = d;
            startDaySelect.appendChild(opt);
        }
    }

    function updateEndDays() {
        const year = parseInt(endYearSelect.value);
        const month = parseInt(endMonthSelect.value);
        const daysInMonth = new Date(year, month, 0).getDate();
        
        endDaySelect.innerHTML = '';
        for (let d = 1; d <= daysInMonth; d++) {
            const opt = document.createElement('option');
            opt.value = d;
            opt.textContent = d;
            endDaySelect.appendChild(opt);
        }
    }

    // Initialize with current date for both start and end
    const now = new Date();
    startYearSelect.value = now.getFullYear();
    startMonthSelect.value = now.getMonth() + 1;
    endYearSelect.value = now.getFullYear();
    endMonthSelect.value = now.getMonth() + 1;
    updateStartDays();
    updateEndDays();
    startDaySelect.value = now.getDate();
    endDaySelect.value = now.getDate();

    // Update days when month/year changes
    startYearSelect.addEventListener('change', updateStartDays);
    startMonthSelect.addEventListener('change', updateStartDays);
    endYearSelect.addEventListener('change', updateEndDays);
    endMonthSelect.addEventListener('change', updateEndDays);

    // Open modal
    btn.addEventListener('click', () => {
        modal.style.display = 'flex';
    });

    // Close modal
    closeBtn.addEventListener('click', () => {
        modal.style.display = 'none';
    });

    // Close on outside click
    window.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.style.display = 'none';
        }
    });

    // Download CSV
    downloadBtn.addEventListener('click', async () => {
        const startYear = startYearSelect.value;
        const startMonth = startMonthSelect.value.padStart(2, '0');
        const startDay = startDaySelect.value.padStart(2, '0');
        const endYear = endYearSelect.value;
        const endMonth = endMonthSelect.value.padStart(2, '0');
        const endDay = endDaySelect.value.padStart(2, '0');
        
        // Validate date range
        const startDate = new Date(`${startYear}-${startMonth}-${startDay}`);
        const endDate = new Date(`${endYear}-${endMonth}-${endDay}`);
        
        if (startDate > endDate) {
            alert('Start date must be before or equal to end date');
            return;
        }
        
        // Collect selected columns
        const selectedColumns = ['dateTime']; // Timestamp always first
        const columnRows = columnContainer.querySelectorAll('.easy-column-row select');
        
        columnRows.forEach(select => {
            const value = select.value;
            if (value && !selectedColumns.includes(value)) {
                selectedColumns.push(value);
            }
        });

        // Validate at least one data column selected
        if (selectedColumns.length === 1) {
            alert('Please select at least one data column');
            return;
        }
        
        // Store the original HTML content
        const originalHTML = downloadBtn.innerHTML;
        
        downloadBtn.disabled = true;
        downloadBtn.textContent = '⏳ Generating...';
        
        try {
            // Build URL with columns parameter and date range
            const columnsParam = selectedColumns.join(',');
            const response = await fetch(`/api/csv/range?startYear=${startYear}&startMonth=${startMonth}&startDay=${startDay}&endYear=${endYear}&endMonth=${endMonth}&endDay=${endDay}&columns=${encodeURIComponent(columnsParam)}`);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            // Get the CSV data as text
            const csvData = await response.text();
            
            // Create a blob and download it
            const blob = new Blob([csvData], { type: 'text/csv' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `weather_${startYear}-${startMonth}-${startDay}_to_${endYear}-${endMonth}-${endDay}.csv`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            
            console.log(`[Easy Button] Downloaded CSV from ${startYear}-${startMonth}-${startDay} to ${endYear}-${endMonth}-${endDay} with columns: ${selectedColumns.join(', ')}`);
        } catch (err) {
            console.error('[Easy Button] Download failed:', err);
            alert('Failed to download CSV: ' + err.message);
        } finally {
            downloadBtn.disabled = false;
            // Restore the original HTML with SVG icon
            downloadBtn.innerHTML = originalHTML;
        }
    });
})();
