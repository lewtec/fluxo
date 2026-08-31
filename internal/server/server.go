package server

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lucasew/fluxo/internal/config"
	"github.com/lucasew/fluxo/internal/session"
)

// ErrAlreadyRunning is returned when Start is called on a running server.
var ErrAlreadyRunning = errors.New("server already running")

// Server manages the Fluxo server
type Server struct {
	config   *config.Config
	manager  *session.Manager
	watcher  *session.Watcher
	listener *HTTPListener

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
}

// New creates a new server
func New(cfg *config.Config) (*Server, error) {
	// Create session manager (includes top-level --debug → rain Config.Debug)
	rainCfg := cfg.ToRainConfig()
	manager, err := session.New(*rainCfg)
	if err != nil {
		return nil, fmt.Errorf("creating session manager: %w", err)
	}

	// Create watcher
	watcher := session.NewWatcher(manager, cfg.WatchInterval)

	// Create HTTP listener
	listener := NewHTTPListener(cfg, manager)

	return &Server{
		config:   cfg,
		manager:  manager,
		watcher:  watcher,
		listener: listener,
	}, nil
}

// Start starts the server
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return ErrAlreadyRunning
	}
	s.running = true
	s.mu.Unlock()

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.manager.Start(ctx)

	// Start watcher in background
	go s.watcher.Start(ctx)

	// Start HTTP listener (blocks)
	return s.listener.Start(ctx)
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// Create timeout context
	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Stop in order
	var errs []error

	// 1. Stop listeners
	if err := s.listener.Stop(ctx); err != nil {
		errs = append(errs, fmt.Errorf("stopping listener: %w", err))
	}

	// 2. Cancel watcher
	if s.cancel != nil {
		s.cancel()
	}

	// 3. Close session (saves state)
	if err := s.manager.Close(); err != nil {
		errs = append(errs, fmt.Errorf("closing session: %w", err))
	}

	s.running = false

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %w", errors.Join(errs...))
	}

	return nil
}
