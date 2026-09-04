package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/lucasew/fluxo/internal/config"
	"github.com/lucasew/fluxo/internal/server"
)

// Set via goreleaser ldflags: -X main.version={{ .Version }}
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fluxo",
		Version: version,
		Short:   "BitTorrent client with a web UI",
		RunE: runServer,
	}

	// Add flags
	config.AddFlags(cmd)

	return cmd
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cmd)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Create server
	srv, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}

	ctx := cmd.Context()

	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Wait for signal (ctx cancelled) or server error
	select {
	case <-ctx.Done():
		log.Println("Received shutdown signal, stopping gracefully...")

		// Independent of the cancelled signal ctx so Shutdown can finish.
		shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Stop(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}

		log.Println("Server stopped")
		return nil

	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	}
}
