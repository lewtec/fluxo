package session

import (
	"testing"

	"github.com/cenkalti/rain/torrent"
)

func TestTorrentWantsPortMap(t *testing.T) {
	tests := []struct {
		st   torrent.Status
		want bool
	}{
		{torrent.Stopped, false},
		{torrent.Stopping, false},
		{torrent.Downloading, true},
		{torrent.Seeding, true},
		{torrent.DownloadingMetadata, true},
		{torrent.Allocating, true},
		{torrent.Verifying, true},
	}
	for _, tt := range tests {
		if got := torrentWantsPortMap(tt.st); got != tt.want {
			t.Errorf("torrentWantsPortMap(%v) = %v, want %v", tt.st, got, tt.want)
		}
	}
}

func TestRemoveMappingForTorrentClearsEntry(t *testing.T) {
	m := NewUPNPManager(NewEventBus())
	m.mappings["abc"] = 50001
	m.removeMappingForTorrent("abc")
	if _, ok := m.mappings["abc"]; ok {
		t.Fatal("expected mapping for abc to be removed")
	}
}

func TestMapExistingEmptyNoPanic(t *testing.T) {
	m := NewUPNPManager(NewEventBus())
	m.MapExisting(nil)
	m.MapExisting([]*torrent.Torrent{})
}
