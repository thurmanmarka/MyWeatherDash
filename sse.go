package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// SSEBroker polls the database for new readings and broadcasts them to connected SSE clients.
type SSEBroker struct {
	db        *sql.DB
	mu        sync.Mutex
	clients   map[chan string]struct{}
	lastEpoch int64
}

func NewSSEBroker(db *sql.DB) *SSEBroker {
	return &SSEBroker{
		db:      db,
		clients: make(map[chan string]struct{}),
	}
}

func (b *SSEBroker) AddClient(ch chan string) {
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	log.Printf("[SSE] client connected; total=%d", b.clientCount())
}

func (b *SSEBroker) RemoveClient(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	log.Printf("[SSE] client disconnected; total=%d", b.clientCount())
}

func (b *SSEBroker) clientCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.clients)
}

func (b *SSEBroker) broadcast(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(b.clients)
	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
			// If client is slow, drop this update for them
		}
	}
	log.Printf("[SSE] broadcast sent to %d clients", n)
}

// pollOnce checks the archive for the newest timestamp and, if it's newer than the
// last seen, reads the latest row and broadcasts a JSON payload.
func (b *SSEBroker) pollOnce() error {
	var maxEpoch sql.NullInt64
	err := b.db.QueryRow(`SELECT MAX(dateTime) FROM archive`).Scan(&maxEpoch)
	if err != nil {
		log.Printf("[SSE] pollOnce: error reading MAX(dateTime): %v", err)
		return err
	}
	if !maxEpoch.Valid {
		log.Printf("[SSE] pollOnce: no rows in archive (MAX invalid)")
		return nil
	}
	if maxEpoch.Int64 == b.lastEpoch {
		log.Printf("[SSE] pollOnce: no change (max=%d, last=%d)", maxEpoch.Int64, b.lastEpoch)
		return nil
	}

	log.Printf("[SSE] pollOnce: change detected (max=%d, last=%d) — loading latest row", maxEpoch.Int64, b.lastEpoch)

	// New data available; read the latest full record
	row := b.db.QueryRow(`
        SELECT dateTime, outTemp, dewpoint, barometer, outHumidity, windSpeed, windGust, windDir, rainRate, rain, lightning_strike_count, inTemp, inHumidity
        FROM archive
        WHERE dateTime IS NOT NULL
        ORDER BY dateTime DESC
        LIMIT 1
    `)

	var epoch int64
	var outTemp, dewpoint, baro, outHum, windSpeed, windGust, windDir, rainRate, rainAmt, lightning sql.NullFloat64
	var inTemp, inHum sql.NullFloat64

	if err := row.Scan(&epoch, &outTemp, &dewpoint, &baro, &outHum, &windSpeed, &windGust, &windDir, &rainRate, &rainAmt, &lightning, &inTemp, &inHum); err != nil {
		log.Printf("[SSE] pollOnce: error scanning latest row: %v", err)
		return err
	}

	payload := map[string]interface{}{
		"timestamp": epoch,
	}

	if outTemp.Valid {
		payload["outTemp"] = outTemp.Float64
	}
	if dewpoint.Valid {
		payload["dewpoint"] = dewpoint.Float64
	}
	if baro.Valid {
		payload["barometer"] = baro.Float64
	}
	if outHum.Valid {
		payload["outHumidity"] = outHum.Float64
	}
	if windSpeed.Valid {
		payload["windSpeed"] = windSpeed.Float64
	}
	if windGust.Valid {
		payload["windGust"] = windGust.Float64
	}
	if windDir.Valid {
		payload["windDir"] = windDir.Float64
	}
	if rainRate.Valid {
		payload["rainRate"] = rainRate.Float64
	}
	if rainAmt.Valid {
		payload["rain"] = rainAmt.Float64
	}
	if lightning.Valid {
		payload["lightning_strike_count"] = lightning.Float64
	}
	if inTemp.Valid {
		payload["inTemp"] = inTemp.Float64
	}
	if inHum.Valid {
		payload["inHumidity"] = inHum.Float64
	}

	b.lastEpoch = epoch

	bts, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Broadcast as a single SSE update event
	msg := string(bts)
	b.broadcast(msg)
	log.Printf("[SSE] broadcast: new update epoch=%d payloadSize=%dB", epoch, len(bts))
	return nil
}

// StartPolling runs a background loop that polls every interval and broadcasts when new data appears.
func (b *SSEBroker) StartPolling(interval time.Duration, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("[SSE] poller starting with interval=%s", interval.String())

		// Run an initial poll immediately
		if err := b.pollOnce(); err != nil {
			log.Println("[SSE] Initial poll error:", err)
		}

		for {
			select {
			case <-ticker.C:
				log.Printf("[SSE] poll tick")
				if err := b.pollOnce(); err != nil {
					log.Println("[SSE] Poll error:", err)
				}
			case <-stopCh:
				log.Println("[SSE] Stopping poller")
				return
			}
		}
	}()
}

// HTTP handler for /api/stream
func (b *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	msgCh := make(chan string, 4)
	b.AddClient(msgCh)
	defer b.RemoveClient(msgCh)

	// Send a short comment to establish the stream
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	// Listen until client disconnects
	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			log.Printf("[SSE] client disconnected (http context done)")
			return
		case msg := <-msgCh:
			// SSE event named 'update'
			fmt.Fprintf(w, "event: update\n")
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
			// Note: per-client send logged at broadcast; here we can trace per-delivery if needed.
		}
	}
}
