package webui

import (
	"time"

	"github.com/cenkalti/rain/torrent"
)

// SpeedStats is the header download/upload totals.
type SpeedStats struct {
	Download int
	Upload   int
}

// FileRow is one file in the torrent detail table.
type FileRow struct {
	Path     string
	Size     string
	Progress float64
}

// TrackerRow is one tracker in the torrent detail table.
type TrackerRow struct {
	URL    string
	Status string
}

// TorrentView is the data a list card or detail page needs.
type TorrentView struct {
	ID            string
	Name          string
	Status        string
	StatusClass   string
	BytesDone     string
	BytesTotal    string
	DownloadSpeed string
	UploadSpeed   string
	ETA           string
	Progress      float64
	ProgressClass string
	InfoHash      string
	AddedAt       string
	PeersTotal    int
	PeersIn       int
	PeersOut      int
	Files         []FileRow
	Trackers      []TrackerRow
	Stopped       bool
}

// AddPage is the add-torrent form state.
type AddPage struct {
	Error string
	URI   string
}

func mapStatus(status torrent.Status) string {
	switch status {
	case torrent.Stopped:
		return "Stopped"
	case torrent.DownloadingMetadata:
		return "Downloading metadata"
	case torrent.Allocating:
		return "Allocating"
	case torrent.Verifying:
		return "Verifying"
	case torrent.Downloading:
		return "Downloading"
	case torrent.Seeding:
		return "Seeding"
	case torrent.Stopping:
		return "Stopping"
	default:
		return "Stopped"
	}
}

func statusClass(status torrent.Status, hasErr bool) string {
	if hasErr {
		return "badge-error"
	}
	switch status {
	case torrent.Seeding:
		return "badge-success"
	case torrent.Downloading, torrent.DownloadingMetadata:
		return "badge-info"
	case torrent.Stopped, torrent.Stopping:
		return "badge-neutral"
	default:
		return "badge-ghost"
	}
}

func mapTrackerStatus(status torrent.TrackerStatus, hasErr bool) string {
	if hasErr {
		return "Error"
	}
	switch status {
	case torrent.NotContactedYet:
		return "Idle"
	case torrent.Contacting:
		return "Announcing"
	case torrent.Working:
		return "Waiting"
	case torrent.NotWorking:
		return "Error"
	default:
		return "Idle"
	}
}

func mapTorrentFiles(t *torrent.Torrent, stats torrent.Stats) []FileRow {
	fileStats, err := t.FileStats()
	if err != nil {
		raw, ferr := t.Files()
		if ferr != nil {
			return nil
		}
		fileStats = make([]torrent.FileStats, len(raw))
		fullyDone := stats.Bytes.Total > 0 && stats.Bytes.Completed >= stats.Bytes.Total
		for i, f := range raw {
			bc := int64(0)
			if fullyDone {
				bc = f.Length()
			}
			fileStats[i] = torrent.FileStats{File: f, BytesCompleted: bc}
		}
	}
	rows := make([]FileRow, len(fileStats))
	for i, f := range fileStats {
		rows[i] = FileRow{
			Path:     f.Path(),
			Size:     formatBytes(f.Length()),
			Progress: progressPercent(f.BytesCompleted, f.Length()),
		}
	}
	return rows
}

func mapTrackers(trackers []torrent.Tracker) []TrackerRow {
	rows := make([]TrackerRow, len(trackers))
	for i, tr := range trackers {
		rows[i] = TrackerRow{
			URL:    tr.URL,
			Status: mapTrackerStatus(tr.Status, tr.Error != nil),
		}
	}
	return rows
}

func torrentView(t *torrent.Torrent) TorrentView {
	stats := t.Stats()

	var eta *int
	if stats.ETA != nil {
		sec := int(stats.ETA.Seconds())
		eta = &sec
	}

	progress := progressPercent(stats.Bytes.Completed, stats.Bytes.Total)
	progressClass := "progress-primary"
	if stats.Status == torrent.Seeding || progress >= 100 {
		progressClass = "progress-success"
	}

	return TorrentView{
		ID:            t.ID(),
		Name:          stats.Name,
		Status:        mapStatus(stats.Status),
		StatusClass:   statusClass(stats.Status, stats.Error != nil),
		BytesDone:     formatBytes(stats.Bytes.Completed),
		BytesTotal:    formatBytes(stats.Bytes.Total),
		DownloadSpeed: formatSpeed(stats.Speed.Download),
		UploadSpeed:   formatSpeed(stats.Speed.Upload),
		ETA:           formatTime(eta),
		Progress:      progress,
		ProgressClass: progressClass,
		InfoHash:      t.InfoHash().String(),
		AddedAt:       t.AddedAt().Local().Format(time.DateTime),
		PeersTotal:    stats.Peers.Total,
		PeersIn:       stats.Peers.Incoming,
		PeersOut:      stats.Peers.Outgoing,
		Files:         mapTorrentFiles(t, stats),
		Trackers:      mapTrackers(t.Trackers()),
		Stopped:       stats.Status == torrent.Stopped,
	}
}

func torrentViews(torrents []*torrent.Torrent) []TorrentView {
	out := make([]TorrentView, len(torrents))
	for i, t := range torrents {
		out[i] = torrentView(t)
	}
	return out
}

func speedStats(torrents []*torrent.Torrent) SpeedStats {
	var s SpeedStats
	for _, t := range torrents {
		st := t.Stats()
		s.Download += st.Speed.Download
		s.Upload += st.Speed.Upload
	}
	return s
}
