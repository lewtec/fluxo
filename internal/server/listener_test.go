package server

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestHTTPListenerStopBeforeStart(t *testing.T) {
	l := &HTTPListener{}
	if err := l.Stop(t.Context()); err != nil {
		t.Fatalf("Stop before Start: %v", err)
	}
}

func TestHTTPListenerStopShutsDownPublishedServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	l := &HTTPListener{}
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	l.mu.Lock()
	l.server = srv
	l.mu.Unlock()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	// Ensure Serve has accepted the listener before Shutdown.
	time.Sleep(20 * time.Millisecond)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	if err := l.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("Serve: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Serve to return after Stop")
	}
}

// Exercises concurrent publication of l.server with Stop under the race detector.
func TestHTTPListenerServerFieldConcurrent(t *testing.T) {
	l := &HTTPListener{}
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			srv := &http.Server{}
			l.mu.Lock()
			l.server = srv
			l.mu.Unlock()
		}()
		go func() {
			defer wg.Done()
			// Shutdown on a never-started server is a no-op success path.
			_ = l.Stop(t.Context())
		}()
	}
	wg.Wait()
}
