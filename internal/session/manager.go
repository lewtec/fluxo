package session

import (
	"fmt"

	"github.com/cenkalti/rain/torrent"
)

// Manager wraps Rain's session and provides event bus
type Manager struct {
	session     *torrent.Session
	eventBus    *EventBus
	upnpManager *UPNPManager
}

// New creates a new session manager
func New(cfg torrent.Config) (*Manager, error) {
	session, err := torrent.NewSession(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	eb := NewEventBus()
	upnp := NewUPNPManager(eb)
	upnp.Start()
	// Rain resumes previously-started torrents inside NewSession without
	// going through StartTorrent, so publish no EventTorrentStarted for them.
	upnp.MapExisting(session.ListTorrents())

	return &Manager{
		session:     session,
		eventBus:    eb,
		upnpManager: upnp,
	}, nil
}

// Session returns the underlying Rain session
func (m *Manager) Session() *torrent.Session {
	return m.session
}

// EventBus returns the event bus
func (m *Manager) EventBus() *EventBus {
	return m.eventBus
}

// AddTorrent adds a new torrent
func (m *Manager) AddTorrent(uri string, opts *torrent.AddTorrentOptions) (*torrent.Torrent, error) {
	if uri == "" {
		return nil, fmt.Errorf("%w: empty URI", ErrInvalidURI)
	}

	// Rain auto-starts when Stopped is false (default). That path never calls
	// StartTorrent, so we must emit EventTorrentStarted for UPnP (and peers).
	autoStarted := opts == nil || !opts.Stopped

	t, err := m.session.AddURI(uri, opts)
	if err != nil {
		// Wrap URI parsing errors
		return nil, fmt.Errorf("%w: %v", ErrInvalidURI, err)
	}

	// Publish event
	m.eventBus.Publish(Event{
		Type:    EventTorrentAdded,
		Torrent: t,
	})
	if autoStarted {
		m.eventBus.Publish(Event{
			Type:    EventTorrentStarted,
			Torrent: t,
		})
	}

	return t, nil
}

// RemoveTorrent removes a torrent
func (m *Manager) RemoveTorrent(id string) error {
	t := m.session.GetTorrent(id)
	if t == nil {
		return fmt.Errorf("%w: %s", ErrTorrentNotFound, id)
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
	t := m.session.GetTorrent(id)
	if t == nil {
		return nil, fmt.Errorf("%w: %s", ErrTorrentNotFound, id)
	}
	return t, nil
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

// StartTorrent starts a torrent
func (m *Manager) StartTorrent(id string) error {
	t := m.session.GetTorrent(id)
	if t == nil {
		return fmt.Errorf("%w: %s", ErrTorrentNotFound, id)
	}
	err := t.Start()
	if err == nil {
		m.eventBus.Publish(Event{
			Type:    EventTorrentStarted,
			Torrent: t,
		})
	}
	return err
}

// StopTorrent stops a torrent
func (m *Manager) StopTorrent(id string) error {
	t := m.session.GetTorrent(id)
	if t == nil {
		return fmt.Errorf("%w: %s", ErrTorrentNotFound, id)
	}
	err := t.Stop()
	if err == nil {
		m.eventBus.Publish(Event{
			Type:    EventTorrentStopped,
			Torrent: t,
		})
	}
	return err
}

// VerifyTorrent verifies a torrent's data
func (m *Manager) VerifyTorrent(id string) error {
	t := m.session.GetTorrent(id)
	if t == nil {
		return fmt.Errorf("%w: %s", ErrTorrentNotFound, id)
	}
	return t.Verify()
}

// AnnounceTorrent forces an announce to trackers
func (m *Manager) AnnounceTorrent(id string) {
	t := m.session.GetTorrent(id)
	if t == nil {
		return
	}
	t.Announce()
}

// AddTracker adds a tracker to a torrent
func (m *Manager) AddTracker(id, url string) error {
	t := m.session.GetTorrent(id)
	if t == nil {
		return fmt.Errorf("%w: %s", ErrTorrentNotFound, id)
	}
	return t.AddTracker(url)
}

// AddPeer adds a peer to a torrent
func (m *Manager) AddPeer(id, addr string) error {
	t := m.session.GetTorrent(id)
	if t == nil {
		return fmt.Errorf("%w: %s", ErrTorrentNotFound, id)
	}
	return t.AddPeer(addr)
}

// GetPeers returns peers for a torrent
func (m *Manager) GetPeers(id string) ([]torrent.Peer, error) {
	t := m.session.GetTorrent(id)
	if t == nil {
		return nil, fmt.Errorf("%w: %s", ErrTorrentNotFound, id)
	}
	return t.Peers(), nil
}

// StartAll starts all torrents
func (m *Manager) StartAll() {
	for _, t := range m.session.ListTorrents() {
		err := t.Start()
		if err == nil {
			m.eventBus.Publish(Event{
				Type:    EventTorrentStarted,
				Torrent: t,
			})
		}
	}
}

// StopAll stops all torrents
func (m *Manager) StopAll() {
	for _, t := range m.session.ListTorrents() {
		err := t.Stop()
		if err == nil {
			m.eventBus.Publish(Event{
				Type:    EventTorrentStopped,
				Torrent: t,
			})
		}
	}
}
