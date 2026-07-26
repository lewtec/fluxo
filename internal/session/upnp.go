package session

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/rain/torrent"
	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway2"
)

type UPNPManager struct {
	eventBus *EventBus
	subID    string
	mappings map[string]int // torrent ID -> port
	mu       sync.Mutex
	stopCh   chan struct{}

	clientsMu sync.RWMutex
	clients   []*internetgateway2.WANIPConnection1

	// Discovery synchronization
	discoveryOnce sync.Once
	discoveryDone chan struct{}
}

func NewUPNPManager(eb *EventBus) *UPNPManager {
	return &UPNPManager{
		eventBus:      eb,
		mappings:      make(map[string]int),
		stopCh:        make(chan struct{}),
		discoveryDone: make(chan struct{}),
	}
}

func (m *UPNPManager) Start() {
	id, ch := m.eventBus.Subscribe()
	m.subID = id

	// Start initial discovery asynchronously
	go m.ensureDiscovery()

	go func() {
		for {
			select {
			case event, ok := <-ch:
				if !ok {
					return
				}
				m.handleEvent(event)
			case <-m.stopCh:
				return
			}
		}
	}()
}

func (m *UPNPManager) ensureDiscovery() {
	m.discoveryOnce.Do(func() {
		m.discover()
		close(m.discoveryDone)
	})
}

func (m *UPNPManager) discover() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	devs, err := goupnp.DiscoverDevicesCtx(ctx, internetgateway2.URN_WANConnectionDevice_1)
	if err != nil {
		log.Printf("UPnP: Discovery failed: %v", err)
		return
	}

	var newClients []*internetgateway2.WANIPConnection1
	for _, dev := range devs {
		if dev.Root == nil {
			continue
		}

		clients, err := internetgateway2.NewWANIPConnection1ClientsFromRootDevice(dev.Root, dev.Location)
		if err != nil {
			continue
		}
		newClients = append(newClients, clients...)
	}

	m.clientsMu.Lock()
	m.clients = newClients
	m.clientsMu.Unlock()

	if len(newClients) > 0 {
		log.Printf("UPnP: Discovered %d IGD clients", len(newClients))
	} else {
		log.Printf("UPnP: No IGD clients discovered")
	}
}

func (m *UPNPManager) Stop() {
	close(m.stopCh)
	if m.subID != "" {
		m.eventBus.Unsubscribe(m.subID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var wg sync.WaitGroup
	for id, port := range m.mappings {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			_ = m.removeMapping(p)
		}(port)
		delete(m.mappings, id)
	}
	// Wait for all unmapping operations to complete with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All cleaned up
	case <-time.After(5 * time.Second):
		log.Printf("UPnP: Shutdown timed out waiting for unmapping")
	}
}

func (m *UPNPManager) handleEvent(event Event) {
	switch event.Type {
	case EventTorrentStarted:
		if event.Torrent != nil {
			go m.addMappingForTorrent(event.Torrent.ID(), event.Torrent.Port())
		}
	case EventTorrentStopped:
		if event.Torrent != nil {
			go m.removeMappingForTorrent(event.Torrent.ID())
		}
	case EventTorrentRemoved:
		// RemoveTorrent closes the torrent without EventTorrentStopped.
		if event.ID != "" {
			go m.removeMappingForTorrent(event.ID)
		}
	}
}

// torrentWantsPortMap reports whether a torrent status should have an external
// port mapping (anything actively using the listen port).
func torrentWantsPortMap(st torrent.Status) bool {
	return st != torrent.Stopped && st != torrent.Stopping
}

// MapExisting maps listen ports for torrents already running (e.g. after
// ResumeOnStartup). EventTorrentStarted is not published for those.
func (m *UPNPManager) MapExisting(torrents []*torrent.Torrent) {
	for _, t := range torrents {
		if t == nil || !torrentWantsPortMap(t.Stats().Status) {
			continue
		}
		id, port := t.ID(), t.Port()
		go m.addMappingForTorrent(id, port)
	}
}

func (m *UPNPManager) addMappingForTorrent(id string, port int) {
	if port <= 0 {
		return
	}

	m.mu.Lock()
	if _, exists := m.mappings[id]; exists {
		m.mu.Unlock()
		return
	}
	m.mappings[id] = port
	m.mu.Unlock()

	err := m.addMapping(port)
	if err != nil {
		log.Printf("UPnP: Failed to map port %d for torrent %s: %v", port, id, err)
		// Clean up on failure
		m.mu.Lock()
		delete(m.mappings, id)
		m.mu.Unlock()
	} else {
		log.Printf("UPnP: Mapped port %d for torrent %s", port, id)
	}
}

func (m *UPNPManager) removeMappingForTorrent(id string) {
	m.mu.Lock()
	port, exists := m.mappings[id]
	if !exists {
		m.mu.Unlock()
		return
	}
	delete(m.mappings, id)
	m.mu.Unlock()

	err := m.removeMapping(port)
	if err != nil {
		log.Printf("UPnP: Failed to unmap port %d for torrent %s: %v", port, id, err)
	} else {
		log.Printf("UPnP: Unmapped port %d for torrent %s", port, id)
	}
}

func (m *UPNPManager) getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no suitable local IP found")
}

func getLocalIPForClient(client *internetgateway2.WANIPConnection1) (string, error) {
	host := client.ServiceClient.Location.Host
	if strings.Contains(host, ":") {
		h, _, err := net.SplitHostPort(host)
		if err == nil {
			host = h
		}
	}

	conn, err := net.Dial("udp", host+":80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}


// addMapping adds UPnP port mappings for peer traffic.
// BitTorrent peer handshakes use TCP on the torrent listen port; UDP is also
// mapped so µTP (and same-port DHT) keep working when the gateway allows it.
// Success requires TCP; UDP failures are logged and ignored.
func (m *UPNPManager) addMapping(port int) error {
	// Wait for initial discovery to complete
	select {
	case <-m.discoveryDone:
	case <-time.After(15 * time.Second): // Wait longer than discovery timeout
		return fmt.Errorf("timeout waiting for UPnP discovery")
	}

	m.clientsMu.RLock()
	clients := m.clients
	m.clientsMu.RUnlock()

	if len(clients) == 0 {
		return fmt.Errorf("no UPnP clients available")
	}

	var errs error
	for _, client := range clients {
		localIP, err := getLocalIPForClient(client)
		if err != nil {
			localIP, err = m.getLocalIP()
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}
		}

		if err := client.AddPortMapping("", uint16(port), "TCP", uint16(port), localIP, true, "fluxo torrent", 0); err != nil {
			errs = errors.Join(errs, err)
			log.Printf("UPnP: TCP map %d failed: %v", port, err)
			continue
		}
		if err := client.AddPortMapping("", uint16(port), "UDP", uint16(port), localIP, true, "fluxo torrent", 0); err != nil {
			// TCP already mapped; UDP is best-effort.
			log.Printf("UPnP: UDP map %d failed (TCP ok): %v", port, err)
		}
		return nil
	}

	if errs != nil {
		return fmt.Errorf("mapping failed on all devices: %w", errs)
	}
	return fmt.Errorf("mapping failed on all devices")
}

// removeMapping removes TCP and UDP UPnP mappings for the port.
func (m *UPNPManager) removeMapping(port int) error {
	m.clientsMu.RLock()
	clients := m.clients
	m.clientsMu.RUnlock()

	for _, client := range clients {
		_ = client.DeletePortMapping("", uint16(port), "TCP")
		_ = client.DeletePortMapping("", uint16(port), "UDP")
	}
	return nil
}
