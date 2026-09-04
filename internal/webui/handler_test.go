package webui

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cenkalti/rain/torrent"
)

func TestParseAddRequestRequiresSource(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/torrents", strings.NewReader("uri="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_, err := parseAddRequest(req)
	if !errors.Is(err, errNoTorrentSource) {
		t.Fatalf("err = %v, want errNoTorrentSource", err)
	}
}

func TestParseAddRequestMagnet(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/torrents", strings.NewReader("uri=magnet%3A%3Fxt%3Durn%3Abtih%3Aabc"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	in, err := parseAddRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if in.URI != "magnet:?xt=urn:btih:abc" {
		t.Errorf("URI = %q", in.URI)
	}
	if in.File != nil {
		t.Error("expected no file")
	}
}

func TestParseAddRequestFile(t *testing.T) {
	t.Parallel()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", "test.torrent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("not-a-real-torrent")); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/torrents", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	in, err := parseAddRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if in.File == nil {
		t.Fatal("expected file")
	}
	defer in.File.Close()
}

func TestMapStatusAndClass(t *testing.T) {
	t.Parallel()
	if got := mapStatus(torrent.Downloading); got != "Downloading" {
		t.Errorf("mapStatus(Downloading) = %q, want Downloading", got)
	}
	if got := mapStatus(torrent.Seeding); got != "Seeding" {
		t.Errorf("mapStatus(Seeding) = %q, want Seeding", got)
	}
	if got := statusClass(torrent.Seeding, false); got != "text-success" {
		t.Errorf("statusClass(Seeding) = %q, want text-success", got)
	}
	if got := statusClass(torrent.Downloading, true); got != "text-error" {
		t.Errorf("statusClass(err) = %q, want text-error", got)
	}
	if got := statusCode(torrent.Seeding); got != "SEED" {
		t.Errorf("statusCode(Seeding) = %q, want SEED", got)
	}
	if got := mapTrackerStatus(torrent.Working, false); got != "Waiting" {
		t.Errorf("mapTrackerStatus(Working) = %q, want Waiting", got)
	}
	if got := mapTrackerStatus(torrent.Contacting, true); got != "Error" {
		t.Errorf("mapTrackerStatus(err) = %q, want Error", got)
	}
}
