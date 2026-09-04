package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestLoadKeepsDefaultPaths(t *testing.T) {
	unsetFluxoEnv(t)
	cfg := loadWithArgs(t)

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
	unsetFluxoEnv(t)
	cfg := loadWithArgs(t, "--database", "/tmp/custom.db", "--data-dir", "/tmp/dl")
	if cfg.Torrent.Database != "/tmp/custom.db" {
		t.Errorf("Database = %q, want /tmp/custom.db", cfg.Torrent.Database)
	}
	if cfg.Torrent.DataDir != "/tmp/dl" {
		t.Errorf("DataDir = %q, want /tmp/dl", cfg.Torrent.DataDir)
	}
}

func TestLoadMissingExplicitConfigErrors(t *testing.T) {
	unsetFluxoEnv(t)
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
	unsetFluxoEnv(t)
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
	unsetFluxoEnv(t)
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

func TestLoadContainerDefaults(t *testing.T) {
	unsetFluxoEnv(t)
	t.Setenv("FLUXO_CONTAINER", "1")

	cfg := loadWithArgs(t)
	if cfg.APIHost != containerAPIHost {
		t.Errorf("APIHost = %q, want %q", cfg.APIHost, containerAPIHost)
	}
	if cfg.APIPort != 8080 {
		t.Errorf("APIPort = %d, want 8080", cfg.APIPort)
	}
	if cfg.Torrent.DataDir != containerDataDir {
		t.Errorf("DataDir = %q, want %q", cfg.Torrent.DataDir, containerDataDir)
	}
	if cfg.Torrent.Database != containerDatabase {
		t.Errorf("Database = %q, want %q", cfg.Torrent.Database, containerDatabase)
	}
}

func TestLoadContainerFalseyKeepsHostDefaults(t *testing.T) {
	unsetFluxoEnv(t)
	t.Setenv("FLUXO_CONTAINER", "false")

	cfg := loadWithArgs(t)
	home := getHomeDir()
	wantDB := filepath.Join(home, ".fluxo", "session.db")
	if cfg.APIHost != "127.0.0.1" {
		t.Errorf("APIHost = %q, want 127.0.0.1", cfg.APIHost)
	}
	if cfg.Torrent.Database != wantDB {
		t.Errorf("Database = %q, want %q", cfg.Torrent.Database, wantDB)
	}
}

func TestLoadContainerSpecificEnvOverridesDefaults(t *testing.T) {
	unsetFluxoEnv(t)
	t.Setenv("FLUXO_CONTAINER", "1")
	t.Setenv("FLUXO_API_HOST", "127.0.0.1")
	t.Setenv("FLUXO_DATA_DIR", "/mnt/dl")
	t.Setenv("FLUXO_DATABASE", "/mnt/session.db")
	t.Setenv("FLUXO_API_PORT", "9090")

	cfg := loadWithArgs(t)
	if cfg.APIHost != "127.0.0.1" {
		t.Errorf("APIHost = %q, want 127.0.0.1", cfg.APIHost)
	}
	if cfg.APIPort != 9090 {
		t.Errorf("APIPort = %d, want 9090", cfg.APIPort)
	}
	if cfg.Torrent.DataDir != "/mnt/dl" {
		t.Errorf("DataDir = %q, want /mnt/dl", cfg.Torrent.DataDir)
	}
	if cfg.Torrent.Database != "/mnt/session.db" {
		t.Errorf("Database = %q, want /mnt/session.db", cfg.Torrent.Database)
	}
}

func TestLoadContainerFlagsOverrideEnv(t *testing.T) {
	unsetFluxoEnv(t)
	t.Setenv("FLUXO_CONTAINER", "1")
	t.Setenv("FLUXO_DATA_DIR", "/mnt/dl")

	cfg := loadWithArgs(t, "--data-dir", "/tmp/flag-dl", "--api-host", "10.0.0.1")
	if cfg.APIHost != "10.0.0.1" {
		t.Errorf("APIHost = %q, want 10.0.0.1", cfg.APIHost)
	}
	if cfg.Torrent.DataDir != "/tmp/flag-dl" {
		t.Errorf("DataDir = %q, want /tmp/flag-dl", cfg.Torrent.DataDir)
	}
	if cfg.Torrent.Database != containerDatabase {
		t.Errorf("Database = %q, want %q", cfg.Torrent.Database, containerDatabase)
	}
}

func loadWithArgs(t *testing.T, args ...string) *Config {
	t.Helper()
	cmd := &cobra.Command{Use: "fluxo"}
	AddFlags(cmd)
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cmd)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func unsetFluxoEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"FLUXO_CONTAINER",
		"FLUXO_DATA_DIR",
		"FLUXO_DATABASE",
		"FLUXO_API_HOST",
		"FLUXO_API_PORT",
	}
	prev := make(map[string]string, len(keys))
	had := make(map[string]bool, len(keys))
	for _, k := range keys {
		v, ok := os.LookupEnv(k)
		had[k] = ok
		prev[k] = v
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("unset %s: %v", k, err)
		}
	}
	t.Cleanup(func() {
		for _, k := range keys {
			var err error
			if had[k] {
				err = os.Setenv(k, prev[k])
			} else {
				err = os.Unsetenv(k)
			}
			if err != nil {
				t.Errorf("restore %s: %v", k, err)
			}
		}
	})
}
