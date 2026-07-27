package session

import (
	"context"
	"fmt"

	"github.com/cenkalti/rain/torrent"
)

// Manager wraps Rain's session and provides event bus
type Manager struct {
	session     *torrent.Session
	eventBus    *EventBus
	upnpManager *UPNPManager
}

// New creates a new session manager. Call Start to begin background services
// (UPnP) with a request-scoped context from the caller.
func New(cfg torrent.Config) (*Manager, error) {
	session, err := torrent.NewSession(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	eb := NewEventBus()
	return &Manager{
		session:     session,
		eventBus:    eb,
		upnpManager: NewUPNPManager(eb),
	}, nil
}

// Start begins background services owned by the manager.
func (m *Manager) Start(ctx context.Context) {
	m.upnpManager.Start(ctx)
}

// Session returns the underlying Rain session
func (m *Manager) Session() *torrent.Session {
	return m.session
}

// EventBus returns the event bus
func (m *Manager) EventBus() *EventBus {
	return m.eventBus
}

// torrentByID looks up a torrent or returns ErrTorrentNotFound.
func (m *Manager) torrentByID(id string) (*torrent.Torrent, error) {
	t := m.session.GetTorrent(id)
	if t == nil {
		return nil, fmt.Errorf("%w: %s", ErrTorrentNotFound, id)
	}
	return t, nil
}

// AddTorrent adds a new torrent
func (m *Manager) AddTorrent(uri string, opts *torrent.AddTorrentOptions) (*torrent.Torrent, error) {
	if uri == "" {
		return nil, fmt.Errorf("%w: empty URI", ErrInvalidURI)
	}

	t, err := m.session.AddURI(uri, opts)
	if err != nil {
		// Wrap URI parsing errors
		return nil, fmt.Errorf("%w: %w", ErrInvalidURI, err)
	}

	// Publish event
	m.eventBus.Publish(Event{
		Type:    EventTorrentAdded,
		Torrent: t,
	})

	return t, nil
}

// RemoveTorrent removes a torrent
func (m *Manager) RemoveTorrent(id string) error {
	if _, err := m.torrentByID(id); err != nil {
		return err
	}

	if err := m.session.RemoveTorrent(id); err != nil {
		return fmt.Errorf("removing torrent: %w", err)
	}

	// Publish event
	m.eventBus.Publish(Event{
		Type: EventTorrentRemoved,
		ID:   id,
	})

	return nil
}

// GetTorrent returns a torrent by ID
func (m *Manager) GetTorrent(id string) (*torrent.Torrent, error) {
	return m.torrentByID(id)
}

// GetTorrents returns all torrents
func (m *Manager) GetTorrents() []*torrent.Torrent {
	return m.session.ListTorrents()
}

// GetStats returns session statistics
func (m *Manager) GetStats() torrent.SessionStats {
	return m.session.Stats()
}

// Close closes the session and event bus
func (m *Manager) Close() error {
	m.upnpManager.Stop()
	m.eventBus.Close()
	return m.session.Close()
}

// applyTorrent runs op on the torrent and publishes eventType on success.
func (m *Manager) applyTorrent(id string, op func(*torrent.Torrent) error, eventType EventType) error {
	t, err := m.torrentByID(id)
	if err == nil {
		err = op(t)
	}
	if err != nil {
		return err
	}
	m.eventBus.Publish(Event{
		Type:    eventType,
		Torrent: t,
	})
	return nil
}

// forEachTorrent runs op on every torrent and publishes eventType on success.
func (m *Manager) forEachTorrent(op func(*torrent.Torrent) error, eventType EventType) {
	for _, t := range m.session.ListTorrents() {
		if err := op(t); err == nil {
			m.eventBus.Publish(Event{
				Type:    eventType,
				Torrent: t,
			})
		}
	}
}

// StartTorrent starts a torrent
func (m *Manager) StartTorrent(id string) error {
	return m.applyTorrent(id, (*torrent.Torrent).Start, EventTorrentStarted)
}

// StopTorrent stops a torrent
func (m *Manager) StopTorrent(id string) error {
	return m.applyTorrent(id, (*torrent.Torrent).Stop, EventTorrentStopped)
}

// withTorrent looks up id and runs op on it.
func (m *Manager) withTorrent(id string, op func(*torrent.Torrent) error) error {
	t, err := m.torrentByID(id)
	if err == nil {
		return op(t)
	}
	return err
}

// VerifyTorrent verifies a torrent's data
func (m *Manager) VerifyTorrent(id string) error {
	return m.withTorrent(id, (*torrent.Torrent).Verify)
}

// AnnounceTorrent forces an announce to trackers. Missing IDs are a no-op.
func (m *Manager) AnnounceTorrent(id string) {
	t, err := m.torrentByID(id)
	if err != nil {
		return
	}
	t.Announce()
}

// AddTracker adds a tracker to a torrent
func (m *Manager) AddTracker(id, url string) error {
	return m.withTorrent(id, func(t *torrent.Torrent) error {
		return t.AddTracker(url)
	})
}

// AddPeer adds a peer to a torrent
func (m *Manager) AddPeer(id, addr string) error {
	return m.withTorrent(id, func(t *torrent.Torrent) error {
		return t.AddPeer(addr)
	})
}

// GetPeers returns peers for a torrent
func (m *Manager) GetPeers(id string) ([]torrent.Peer, error) {
	var peers []torrent.Peer
	err := m.withTorrent(id, func(t *torrent.Torrent) error {
		peers = t.Peers()
		return nil
	})
	return peers, err
}

// StartAll starts all torrents
func (m *Manager) StartAll() {
	m.forEachTorrent((*torrent.Torrent).Start, EventTorrentStarted)
}

// StopAll stops all torrents
func (m *Manager) StopAll() {
	m.forEachTorrent((*torrent.Torrent).Stop, EventTorrentStopped)
}
