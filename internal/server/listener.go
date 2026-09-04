package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/lucasew/fluxo/internal/config"
	"github.com/lucasew/fluxo/internal/session"
	"github.com/lucasew/fluxo/internal/webui"
	"github.com/lucasew/fluxo/web"
)

// HTTPListener implements the HTTP listener
type HTTPListener struct {
	config  *config.Config
	manager *session.Manager

	// mu guards server. Start publishes the *http.Server under lock before
	// ListenAndServe; Stop reads it under lock so a concurrent graceful
	// shutdown cannot race with the write (detected by go test -race).
	mu     sync.Mutex
	server *http.Server
}

// NewHTTPListener creates a new HTTP listener
func NewHTTPListener(cfg *config.Config, manager *session.Manager) *HTTPListener {
	return &HTTPListener{
		config:  cfg,
		manager: manager,
	}
}

// Start starts the HTTP server
func (l *HTTPListener) Start(ctx context.Context) error {
	staticFS, err := web.Static()
	if err != nil {
		return fmt.Errorf("accessing embedded files: %w", err)
	}

	mux := http.NewServeMux()
	ui := webui.New(l.manager, staticFS, l.config.DevMode)
	ui.Register(mux)

	// SSE streams stay open. A WriteTimeout would abort them mid-stream;
	// leave it zero. Prefer ReadHeaderTimeout over ReadTimeout so slowloris
	// protection does not apply a read deadline to the whole SSE lifetime.
	// IdleTimeout reaps keep-alive HTTP connections that go quiet; SSE
	// writes (including heartbeats) reset that timer.
	addr := fmt.Sprintf("%s:%d", l.config.APIHost, l.config.APIPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           l.withMiddleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	l.mu.Lock()
	l.server = srv
	l.mu.Unlock()

	log.Printf("Starting HTTP server on %s", addr)

	// Serve via local reference so we do not re-read l.server without the lock.
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

// Stop stops the HTTP server
func (l *HTTPListener) Stop(ctx context.Context) error {
	l.mu.Lock()
	srv := l.server
	l.mu.Unlock()
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

func (l *HTTPListener) withMiddleware(handler http.Handler) http.Handler {
	// CORS middleware
	cors := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("X-Content-Type-Options", "nosniff")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Logging middleware
	logging := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
		})
	}

	// Recovery middleware
	recovery := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("panic: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}

	return recovery(logging(cors(handler)))
}
