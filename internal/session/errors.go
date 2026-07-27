package session

import "errors"

// Common errors that can be checked with errors.Is.
var (
	ErrTorrentNotFound      = errors.New("torrent not found")
	ErrInvalidURI           = errors.New("invalid torrent URI")
	ErrNoLocalIP            = errors.New("no suitable local IP found")
	ErrUPNPDiscoveryTimeout = errors.New("timeout waiting for UPnP discovery")
	ErrNoUPNPClients        = errors.New("no UPnP clients available")
	ErrUPNPMappingFailed    = errors.New("mapping failed on all devices")
)
