package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/cenkalti/rain/torrent"
)

// Config holds all configuration for Fluxo
type Config struct {
	// API settings
	APIPort  int    `mapstructure:"api-port"`
	APIHost  string `mapstructure:"api-host"`
	Debug    bool   `mapstructure:"debug"`
	DevMode  bool   `mapstructure:"dev-mode"`
	DevProxy string `mapstructure:"dev-proxy"`

	// Watcher settings
	WatchInterval time.Duration `mapstructure:"watch-interval"`

	// Rain torrent config (embedded)
	Torrent TorrentConfig `mapstructure:",squash"`
}

// TorrentConfig wraps Rain's config with our custom defaults
type TorrentConfig struct {
	Database                 string        `mapstructure:"database"`
	DataDir                  string        `mapstructure:"data-dir"`
	DataDirIncludesTorrentID bool          `mapstructure:"data-dir-includes-torrent-id"`
	PortBegin                uint16        `mapstructure:"port-begin"`
	PortEnd                  uint16        `mapstructure:"port-end"`
	DHTEnabled               bool          `mapstructure:"dht-enabled"`
	DHTPort                  uint16        `mapstructure:"dht-port"`
	DHTHost                  string        `mapstructure:"dht-host"`
	PEXEnabled               bool          `mapstructure:"pex-enabled"`
	RPCEnabled               bool          `mapstructure:"rpc-enabled"`
	RPCHost                  string        `mapstructure:"rpc-host"`
	RPCPort                  uint16        `mapstructure:"rpc-port"`
	MaxOpenFiles             uint          `mapstructure:"max-open-files"`
	PieceCacheSize           int64         `mapstructure:"piece-cache-size"`
	ParallelMetadata         int           `mapstructure:"parallel-metadata"`
	BlocklistEnabled         bool          `mapstructure:"blocklist-enabled"`
	BlocklistURL             string        `mapstructure:"blocklist-url"`
	ResumeWriteInterval      time.Duration `mapstructure:"resume-write-interval"`
	HealthCheckInterval      time.Duration `mapstructure:"health-check-interval"`
	HealthCheckTimeout       time.Duration `mapstructure:"health-check-timeout"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir := getHomeDir()

	return &Config{
		APIPort:       8080,
		APIHost:       "127.0.0.1",
		Debug:         false,
		DevMode:       false,
		DevProxy:      "http://localhost:5173",
		WatchInterval: 1 * time.Second,
		Torrent: TorrentConfig{
			Database:                 filepath.Join(homeDir, ".fluxo", "session.db"),
			DataDir:                  filepath.Join(homeDir, ".fluxo", "downloads"),
			DataDirIncludesTorrentID: false,
			PortBegin:                50000,
			PortEnd:                  60000,
			DHTEnabled:               true,
			DHTPort:                  7246,
			DHTHost:                  "0.0.0.0",
			PEXEnabled:               true,
			RPCEnabled:               false,
			RPCHost:                  "127.0.0.1",
			RPCPort:                  7245,
			MaxOpenFiles:             1024,
			PieceCacheSize:           256 * 1024 * 1024,
			ParallelMetadata:         2,
			BlocklistEnabled:         false,
			BlocklistURL:             "",
			ResumeWriteInterval:      30 * time.Second,
			HealthCheckInterval:      30 * time.Second,
			HealthCheckTimeout:       5 * time.Minute,
		},
	}
}

// ToRainConfig converts Fluxo config into Rain's torrent.Config, including
// top-level flags that are not part of TorrentConfig (currently Debug).
func (c *Config) ToRainConfig() *torrent.Config {
	config := c.Torrent.ToRainConfig()
	// --debug was registered and unmarshaled but never applied; Rain enables
	// its debug logger only when Config.Debug is true (session.go).
	config.Debug = c.Debug
	return config
}

// ToRainConfig converts TorrentConfig to Rain's torrent.Config
func (c *TorrentConfig) ToRainConfig() *torrent.Config {
	config := torrent.DefaultConfig
	config.Database = c.Database

	config.DataDir = c.DataDir
	config.DataDirIncludesTorrentID = c.DataDirIncludesTorrentID
	config.PortBegin = c.PortBegin
	config.PortEnd = c.PortEnd
	config.DHTEnabled = c.DHTEnabled
	config.DHTPort = c.DHTPort
	config.DHTHost = c.DHTHost
	config.PEXEnabled = c.PEXEnabled
	config.RPCEnabled = c.RPCEnabled
	config.RPCHost = c.RPCHost
	config.RPCPort = int(c.RPCPort)
	config.MaxOpenFiles = uint64(c.MaxOpenFiles)
	config.ReadCacheSize = c.PieceCacheSize
	config.ParallelMetadataDownloads = c.ParallelMetadata
	// Single --blocklist-enabled maps onto all three Rain blocklist gates.
	config.BlocklistEnabledForTrackers = c.BlocklistEnabled
	config.BlocklistEnabledForOutgoingConnections = c.BlocklistEnabled
	config.BlocklistEnabledForIncomingConnections = c.BlocklistEnabled
	config.BlocklistURL = c.BlocklistURL
	config.ResumeWriteInterval = c.ResumeWriteInterval
	config.HealthCheckInterval = c.HealthCheckInterval
	config.HealthCheckTimeout = c.HealthCheckTimeout
	return &config
}

func getHomeDir() string {
	// Try $HOME first
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	// Fallback to current directory
	return "."
}
