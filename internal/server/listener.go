package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/lucasew/fluxo/internal/config"
	"github.com/lucasew/fluxo/internal/graphql"
	"github.com/lucasew/fluxo/internal/session"
	"github.com/lucasew/fluxo/web"
)

// HTTPListener implements the HTTP/WebSocket listener
type HTTPListener struct {
	config  *config.Config
	manager *session.Manager

	// mu guards server. Start publishes the *http.Server under lock before
	// ListenAndServe; Stop reads it under lock so a concurrent graceful
	// shutdown cannot race with the write (detected by go test -race).
	mu     sync.Mutex
	server *http.Server

	// devProxy is built once in Start when DevMode is on. Rebuilding
	// httputil.ReverseProxy on every request re-parses the URL and
	// reallocates transport state for no benefit (DevProxy is static).
	devProxy *httputil.ReverseProxy
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
	// Create GraphQL resolver
	resolver := graphql.NewResolver(l.manager)

	// Create GraphQL schema
	schema := l.createSchema(resolver)

	// Create router
	mux := http.NewServeMux()

	// GraphQL endpoint
	mux.Handle("/graphql", schema)

	// GraphiQL playground
	mux.Handle("/graphiql", playground.Handler("GraphQL Playground", "/graphql"))

	// Static files (React app)
	if l.config.DevMode {
		target, err := url.Parse(l.config.DevProxy)
		if err != nil {
			return fmt.Errorf("invalid dev proxy URL %q: %w", l.config.DevProxy, err)
		}
		l.devProxy = httputil.NewSingleHostReverseProxy(target)
		mux.HandleFunc("/", l.proxyToVite)
	} else {
		// Serve embedded files in production with SPA catch-all
		webFS, err := web.WebDist()
		if err != nil {
			return fmt.Errorf("accessing embedded files: %w", err)
		}

		fileServer := http.FileServer(http.FS(webFS))

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Check if file exists in FS
			f, err := webFS.Open(strings.TrimPrefix(r.URL.Path, "/"))
			if err == nil {
				defer f.Close()
				stat, err := f.Stat()
				if err == nil && !stat.IsDir() {
					fileServer.ServeHTTP(w, r)
					return
				}
			} else if !strings.HasPrefix(r.URL.Path, "/graphql") && !strings.HasPrefix(r.URL.Path, "/api") {
				// Fallback to index.html for SPA routes, unless it's an API call
				// (Though checking prefix here is a safeguard, the mux router prioritizes exact/prefix matches if registered.
				// But since we use mux.HandleFunc("/", ...), it matches everything not matched by others.)

				// Serve index.html
				index, err := webFS.Open("index.html")
				if err != nil {
					http.Error(w, "index.html not found", http.StatusInternalServerError)
					return
				}
				defer index.Close()

				// Get stat for ModTime
				stat, err := index.Stat()
				if err != nil {
					http.Error(w, "index.html stat failed", http.StatusInternalServerError)
					return
				}

				http.ServeContent(w, r, "index.html", stat.ModTime(), index.(io.ReadSeeker))
				return
			}

			// Default 404 for other cases (like /api/unknown) handled by fileServer usually, but here we might want explicit logic.
			// Actually fileServer handles directory listings or 404s.
			// But since we intercepted above, we only fallback to index.html for non-API routes.
			fileServer.ServeHTTP(w, r)
		})
	}

	// Create server.
	// GraphQL subscriptions use long-lived WebSockets with keepalive pings.
	// A WriteTimeout would abort those connections mid-stream; leave it zero.
	// Prefer ReadHeaderTimeout over ReadTimeout so slowloris protection does
	// not apply a read deadline to the whole hijacked WebSocket lifetime.
	// IdleTimeout reaps keep-alive HTTP connections that go quiet; hijacked
	// WebSocket connections are not subject to it after the upgrade.
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

func (l *HTTPListener) createSchema(resolver *graphql.Resolver) *handler.Server {
	execSchema := graphql.NewExecutableSchema(graphql.Config{
		Resolvers: resolver,
	})
	// Build with handler.New: NewDefaultServer pre-registers a WebSocket transport,
	// and gqlgen picks the first matching transport — so a later AddTransport(Websocket)
	// never ran (keepalive/CheckOrigin were dead config).
	srv := handler.New(execSchema)

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 30 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Match existing CORS middleware (allow all); tighten later if needed.
				return true
			},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})

	return srv
}

func (l *HTTPListener) withMiddleware(handler http.Handler) http.Handler {
	// CORS middleware
	cors := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

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

func (l *HTTPListener) proxyToVite(w http.ResponseWriter, r *http.Request) {
	// Start always installs devProxy before ListenAndServe registers handlers.
	l.devProxy.ServeHTTP(w, r)
}
