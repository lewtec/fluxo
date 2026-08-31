package session

import (
	"context"
	"time"

	"github.com/cenkalti/rain/torrent"
)

// Watcher watches for torrent changes and publishes events
type Watcher struct {
	manager  *Manager
	interval time.Duration
	cache    map[string]*torrent.Stats
}

// NewWatcher creates a new watcher
func NewWatcher(manager *Manager, interval time.Duration) *Watcher {
	return &Watcher{
		manager:  manager,
		interval: interval,
		cache:    make(map[string]*torrent.Stats),
	}
}

// Start starts the watcher loop
func (w *Watcher) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Initial check
	w.check()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.check()
		}
	}
}

func (w *Watcher) check() {
	// Get current torrents
	torrents := w.manager.GetTorrents()
	currentIDs := make(map[string]bool)

	// Check for updates
	for _, t := range torrents {
		id := t.ID()
		currentIDs[id] = true

		stats := t.Stats()

		// Check if changed
		if cached, ok := w.cache[id]; ok {
			if statsChanged(cached, &stats) {
				w.manager.eventBus.Publish(Event{
					Type:    EventTorrentUpdated,
					Torrent: t,
				})
				w.cache[id] = &stats
			}
		} else {
			// New torrent detected (not added via AddTorrent)
			w.cache[id] = &stats
		}
	}

	// Check for removed torrents
	for id := range w.cache {
		if !currentIDs[id] {
			delete(w.cache, id)
		}
	}

	// Publish stats update
	stats := w.manager.GetStats()
	w.manager.eventBus.Publish(Event{
		Type:  EventStatsUpdated,
		Stats: &stats,
	})
}

// errorChanged reports whether two errors differ by presence or message
// (interface != only compares identity).
func errorChanged(a, b error) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil || b == nil {
		return true
	}
	return a.Error() != b.Error()
}

// statsChanged compares two torrent stats for changes
func statsChanged(old, new *torrent.Stats) bool {
	// Compare fields that drive the UI. Name and Bytes.Total change when
	// magnet metadata arrives; Bytes.Uploaded / Pieces.Have change while seeding
	// even if instantaneous speed is zero between ticks.
	return old.Status != new.Status ||
		old.Name != new.Name ||
		old.Bytes.Completed != new.Bytes.Completed ||
		old.Bytes.Total != new.Bytes.Total ||
		old.Bytes.Uploaded != new.Bytes.Uploaded ||
		old.Pieces.Have != new.Pieces.Have ||
		old.Speed.Download != new.Speed.Download ||
		old.Speed.Upload != new.Speed.Upload ||
		old.Peers.Total != new.Peers.Total ||
		errorChanged(old.Error, new.Error)
}
