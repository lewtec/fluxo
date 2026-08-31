package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestLoadKeepsDefaultPaths(t *testing.T) {
	cmd := &cobra.Command{Use: "fluxo"}
	AddFlags(cmd)
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cmd)
	if err != nil {
		t.Fatal(err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		// DefaultConfig falls back to $HOME env or "."; mirror getHomeDir for assert
		home = os.Getenv("HOME")
		if home == "" {
			home = "."
		}
	}
	// getHomeDir prefers $HOME over UserHomeDir — match that
	if h := os.Getenv("HOME"); h != "" {
		home = h
	}

	wantDB := filepath.Join(home, ".fluxo", "session.db")
	wantData := filepath.Join(home, ".fluxo", "downloads")

	if cfg.Torrent.Database != wantDB {
		t.Errorf("Database = %q, want %q", cfg.Torrent.Database, wantDB)
	}
	if cfg.Torrent.DataDir != wantData {
		t.Errorf("DataDir = %q, want %q", cfg.Torrent.DataDir, wantData)
	}
}

func TestLoadExplicitDatabaseOverride(t *testing.T) {
	cmd := &cobra.Command{Use: "fluxo"}
	AddFlags(cmd)
	if err := cmd.ParseFlags([]string{"--database", "/tmp/custom.db", "--data-dir", "/tmp/dl"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Torrent.Database != "/tmp/custom.db" {
		t.Errorf("Database = %q, want /tmp/custom.db", cfg.Torrent.Database)
	}
	if cfg.Torrent.DataDir != "/tmp/dl" {
		t.Errorf("DataDir = %q, want /tmp/dl", cfg.Torrent.DataDir)
	}
}

func TestLoadMissingExplicitConfigErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-fluxo.yaml")
	cmd := &cobra.Command{Use: "fluxo"}
	AddFlags(cmd)
	if err := cmd.ParseFlags([]string{"--config", missing}); err != nil {
		t.Fatal(err)
	}
	_, err := Load(cmd)
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing --config file")
	}
}

func TestLoadInvalidConfigErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fluxo.yaml")
	// Unclosed quote → YAML parse failure (not "file not found")
	if err := os.WriteFile(path, []byte("api-port: \"8080\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "fluxo"}
	AddFlags(cmd)
	if err := cmd.ParseFlags([]string{"--config", path}); err != nil {
		t.Fatal(err)
	}
	_, err := Load(cmd)
	if err == nil {
		t.Fatal("Load() error = nil, want error for invalid config YAML")
	}
}

func TestLoadValidExplicitConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fluxo.yaml")
	if err := os.WriteFile(path, []byte("api-port: 9090\napi-host: 0.0.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "fluxo"}
	AddFlags(cmd)
	if err := cmd.ParseFlags([]string{"--config", path}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIPort != 9090 {
		t.Errorf("APIPort = %d, want 9090", cfg.APIPort)
	}
	if cfg.APIHost != "0.0.0.0" {
		t.Errorf("APIHost = %q, want 0.0.0.0", cfg.APIHost)
	}
}

func TestToRainConfigMapsDebug(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ToRainConfig().Debug {
		t.Fatal("default Debug should be false on rain config")
	}
	cfg.Debug = true
	if !cfg.ToRainConfig().Debug {
		t.Fatal("Debug=true must set rain Config.Debug so session enables debug logging")
	}
}

func TestLoadDebugFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "fluxo"}
	AddFlags(cmd)
	if err := cmd.ParseFlags([]string{"--debug"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Debug {
		t.Fatal("Load with --debug: Config.Debug = false, want true")
	}
	if !cfg.ToRainConfig().Debug {
		t.Fatal("Load with --debug: rain Config.Debug = false, want true")
	}
}
