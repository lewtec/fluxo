package webui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/lucasew/fluxo/internal/session"
)

const sseHeartbeat = 15 * time.Second

func writeSSE(w io.Writer, event, data string) error {
	if event != "" {
		if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
			return err
		}
	}
	for line := range strings.SplitSeq(data, "\n") {
		if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\n")
	return err
}

func renderComponent(ctx context.Context, n templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := n.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (h *Handler) events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ctx := r.Context()
	events := subscribeFilter(ctx, h.manager.EventBus(), func(ev session.Event) (session.Event, bool) {
		switch ev.Type {
		case session.EventTorrentAdded, session.EventTorrentRemoved,
			session.EventTorrentUpdated, session.EventTorrentStarted,
			session.EventTorrentStopped, session.EventStatsUpdated:
			return ev, true
		default:
			return session.Event{}, false
		}
	})

	ticker := time.NewTicker(sseHeartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case ev, ok := <-events:
			if !ok {
				return
			}
			if err := h.pushEvent(ctx, w, ev); err != nil {
				log.Printf("sse: %v", err)
				return
			}
			flusher.Flush()
		}
	}
}

func (h *Handler) pushEvent(ctx context.Context, w io.Writer, ev session.Event) error {
	switch ev.Type {
	case session.EventTorrentRemoved:
		if err := writeSSE(w, "removed", ev.ID); err != nil {
			return err
		}
		return h.pushListAndStats(ctx, w)
	case session.EventStatsUpdated:
		return h.pushStats(ctx, w)
	default:
		if ev.Torrent != nil {
			html, err := renderComponent(ctx, TorrentDetail(torrentView(ev.Torrent)))
			if err != nil {
				return err
			}
			if err := writeSSE(w, "detail", html); err != nil {
				return err
			}
		}
		return h.pushListAndStats(ctx, w)
	}
}

func (h *Handler) pushListAndStats(ctx context.Context, w io.Writer) error {
	if err := h.pushList(ctx, w); err != nil {
		return err
	}
	return h.pushStats(ctx, w)
}

func (h *Handler) pushList(ctx context.Context, w io.Writer) error {
	html, err := renderComponent(ctx, TorrentList(torrentViews(h.manager.GetTorrents())))
	if err != nil {
		return err
	}
	return writeSSE(w, "list", html)
}

func (h *Handler) pushStats(ctx context.Context, w io.Writer) error {
	html, err := renderComponent(ctx, HeaderStats(speedStats(h.manager.GetTorrents())))
	if err != nil {
		return err
	}
	return writeSSE(w, "stats", html)
}
