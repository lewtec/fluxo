package session

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cenkalti/rain/torrent"
)

// Test fixtures for errorChanged (Err* table entries for the lint catalog).
var (
	ErrBoom = errors.New("boom")
	ErrA    = errors.New("a")
	ErrB    = errors.New("b")
)

// stringErr is a distinct error value with a fixed message (avoids ad-hoc errors.New in cases).
type stringErr string

func (e stringErr) Error() string { return string(e) }

func TestErrorChanged(t *testing.T) {
	tests := []struct {
		name string
		a, b error
		want bool
	}{
		{"both nil", nil, nil, false},
		{"nil to error", nil, ErrBoom, true},
		{"error to nil", ErrBoom, nil, true},
		// Distinct instances with the same text compare equal by Error() string.
		{"same message different instances", stringErr("boom"), stringErr("boom"), false},
		{"different messages", ErrA, ErrB, true},
		{"wrapped same text", fmt.Errorf("%w", ErrBoom), ErrBoom, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errorChanged(tt.a, tt.b); got != tt.want {
				t.Fatalf("errorChanged(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestStatsChangedNameAndMetadata(t *testing.T) {
	base := &torrent.Stats{
		Name:   "magnet-hash",
		Status: torrent.DownloadingMetadata,
	}
	base.Bytes.Total = 0
	base.Bytes.Completed = 0
	base.Bytes.Uploaded = 0
	base.Pieces.Have = 0

	t.Run("name change after metadata", func(t *testing.T) {
		next := *base
		next.Name = "Real Torrent Name"
		if !statsChanged(base, &next) {
			t.Fatal("expected statsChanged when Name updates")
		}
	})

	t.Run("bytes total after metadata", func(t *testing.T) {
		next := *base
		next.Bytes.Total = 1 << 30
		if !statsChanged(base, &next) {
			t.Fatal("expected statsChanged when Bytes.Total updates")
		}
	})

	t.Run("uploaded while speed idle", func(t *testing.T) {
		next := *base
		next.Bytes.Uploaded = 1024
		if !statsChanged(base, &next) {
			t.Fatal("expected statsChanged when Bytes.Uploaded updates")
		}
	})

	t.Run("pieces have", func(t *testing.T) {
		next := *base
		next.Pieces.Have = 3
		if !statsChanged(base, &next) {
			t.Fatal("expected statsChanged when Pieces.Have updates")
		}
	})

	t.Run("unchanged", func(t *testing.T) {
		next := *base
		if statsChanged(base, &next) {
			t.Fatal("expected no change for identical stats")
		}
	})
}
