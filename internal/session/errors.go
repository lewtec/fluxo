package session

// sessionError is a stable session-level sentinel. Prefer these (or
// fmt.Errorf %w wrapping them) over bare errors.New so callers can errors.Is.
type sessionError string

func (e sessionError) Error() string { return string(e) }

// Session error table. Dynamic detail is attached with fmt.Errorf %w.
const (
	ErrTorrentNotFound      sessionError = "torrent not found"
	ErrInvalidURI           sessionError = "invalid torrent URI"
	ErrNoLocalIP            sessionError = "no suitable local IP found"
	ErrUPNPDiscoveryTimeout sessionError = "timeout waiting for UPnP discovery"
	ErrNoUPNPClients        sessionError = "no UPnP clients available"
	ErrUPNPMappingFailed    sessionError = "mapping failed on all devices"
)
