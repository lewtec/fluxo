package webui

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/cenkalti/rain/torrent"
	"github.com/lucasew/fluxo/internal/session"
)

const maxTorrentUpload = 10 << 20

var (
	errInvalidForm     = errors.New("invalid form")
	errInvalidFile     = errors.New("invalid torrent file")
	errNoTorrentSource = errors.New("magnet URI or .torrent file is required")
)

// Handler serves the templ UI and SSE event stream.
type Handler struct {
	manager *session.Manager
	static  http.Handler
	devMode bool
}

// New returns a Handler that serves pages from manager and files from staticFS.
func New(manager *session.Manager, staticFS fs.FS, devMode bool) *Handler {
	return &Handler{
		manager: manager,
		static:  http.FileServer(http.FS(staticFS)),
		devMode: devMode,
	}
}

// Register mounts UI routes on mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.home)
	mux.HandleFunc("GET /add", h.addForm)
	mux.HandleFunc("GET /torrents/{id}", h.detail)
	mux.HandleFunc("POST /torrents", h.add)
	mux.HandleFunc("POST /torrents/{id}/start", h.start)
	mux.HandleFunc("POST /torrents/{id}/stop", h.stop)
	mux.HandleFunc("POST /torrents/{id}/remove", h.remove)
	mux.HandleFunc("GET /events", h.events)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.HandlerFunc(h.serveStatic)))
}

func (h *Handler) serveStatic(w http.ResponseWriter, r *http.Request) {
	if h.devMode {
		w.Header().Set("Cache-Control", "no-store")
	}
	h.static.ServeHTTP(w, r)
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, status int, n templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if h.devMode {
		w.Header().Set("Cache-Control", "no-store")
	}
	w.WriteHeader(status)
	if err := n.Render(r.Context(), w); err != nil {
		log.Printf("render: %v", err)
	}
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	torrents := h.manager.GetTorrents()
	h.render(w, r, http.StatusOK, HomePage(speedStats(torrents), torrentViews(torrents)))
}

func (h *Handler) addForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, http.StatusOK, AddTorrentPage(speedStats(h.manager.GetTorrents()), AddPage{}))
}

func (h *Handler) detail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := h.manager.GetTorrent(id)
	if err != nil {
		h.render(w, r, http.StatusNotFound, NotFoundPage(speedStats(h.manager.GetTorrents())))
		return
	}
	h.render(w, r, http.StatusOK, TorrentPage(speedStats(h.manager.GetTorrents()), torrentView(t)))
}

func (h *Handler) add(w http.ResponseWriter, r *http.Request) {
	in, err := parseAddRequest(r)
	if err != nil {
		h.render(w, r, http.StatusBadRequest, AddTorrentPage(speedStats(h.manager.GetTorrents()), AddPage{
			Error: err.Error(),
			URI:   strings.TrimSpace(r.FormValue("uri")),
		}))
		return
	}

	opts := &torrent.AddTorrentOptions{}
	if in.File != nil {
		defer in.File.Close()
		_, err = h.manager.AddTorrentFile(in.File, opts)
	} else {
		_, err = h.manager.AddTorrent(in.URI, opts)
	}
	if err != nil {
		status := http.StatusInternalServerError
		msg := "could not add torrent"
		if errors.Is(err, session.ErrInvalidURI) || errors.Is(err, session.ErrInvalidTorrent) {
			status = http.StatusBadRequest
			msg = err.Error()
		} else {
			log.Printf("add torrent: %v", err)
		}
		h.render(w, r, status, AddTorrentPage(speedStats(h.manager.GetTorrents()), AddPage{
			Error: msg,
			URI:   in.URI,
		}))
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	h.mutate(w, r, h.manager.StartTorrent)
}

func (h *Handler) stop(w http.ResponseWriter, r *http.Request) {
	h.mutate(w, r, h.manager.StopTorrent)
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.manager.RemoveTorrent(id); err != nil {
		h.writeMutateError(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) mutate(w http.ResponseWriter, r *http.Request, op func(string) error) {
	id := r.PathValue("id")
	if err := op(id); err != nil {
		h.writeMutateError(w, r, err)
		return
	}
	http.Redirect(w, r, "/torrents/"+id, http.StatusSeeOther)
}

func (h *Handler) writeMutateError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, session.ErrTorrentNotFound) {
		h.render(w, r, http.StatusNotFound, NotFoundPage(speedStats(h.manager.GetTorrents())))
		return
	}
	log.Printf("torrent action: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

type addInput struct {
	URI  string
	File io.ReadCloser
}

func parseAddRequest(r *http.Request) (addInput, error) {
	if err := r.ParseMultipartForm(maxTorrentUpload); err != nil {
		if err := r.ParseForm(); err != nil {
			return addInput{}, errInvalidForm
		}
	}

	file, _, err := r.FormFile("file")
	if err == nil {
		return addInput{File: file}, nil
	}
	if !errors.Is(err, http.ErrMissingFile) && !errors.Is(err, http.ErrNotMultipart) {
		return addInput{}, errInvalidFile
	}

	uri := strings.TrimSpace(r.FormValue("uri"))
	if uri == "" {
		return addInput{}, errNoTorrentSource
	}
	return addInput{URI: uri}, nil
}
