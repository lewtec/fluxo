package session

import "errors"

// Common errors that can be checked with errors.Is
var (
	// ErrTorrentNotFound is returned when a torrent ID is not found
	ErrTorrentNotFound = errors.New("torrent not found")

	// ErrInvalidURI is returned when a torrent URI is invalid
	ErrInvalidURI = errors.New("invalid torrent URI")
)
