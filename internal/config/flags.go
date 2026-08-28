package config

import (
	"github.com/spf13/cobra"
)

// AddFlags adds all configuration flags to the command.
// Flag defaults match DefaultConfig so BindPFlags + Unmarshal cannot wipe them
// with empty strings (notably --database / --data-dir).
func AddFlags(cmd *cobra.Command) {
	defaults := DefaultConfig()
	flags := cmd.Flags()

	// Config file
	flags.String("config", "", "config file path")

	// API settings
	flags.Int("api-port", defaults.APIPort, "API server port")
	flags.String("api-host", defaults.APIHost, "API server host")
	flags.Bool("debug", defaults.Debug, "enable debug logging")
	flags.Bool("dev-mode", defaults.DevMode, "development mode (proxy to Vite)")
	flags.String("dev-proxy", defaults.DevProxy, "Vite dev server URL")

	// Watcher settings
	flags.Duration("watch-interval", defaults.WatchInterval, "torrent watch interval")

	// Rain/Torrent settings
	flags.String("database", defaults.Torrent.Database, "path to session database")
	flags.String("data-dir", defaults.Torrent.DataDir, "path to downloads directory")
	flags.Bool(
		"data-dir-includes-torrent-id",
		defaults.Torrent.DataDirIncludesTorrentID,
		"include torrent ID in data directory path",
	)
	flags.Uint16("port-begin", defaults.Torrent.PortBegin, "beginning of port range for incoming connections")
	flags.Uint16("port-end", defaults.Torrent.PortEnd, "end of port range for incoming connections")
	flags.Bool("dht-enabled", defaults.Torrent.DHTEnabled, "enable DHT")
	flags.Uint16("dht-port", defaults.Torrent.DHTPort, "DHT UDP port")
	flags.String("dht-host", defaults.Torrent.DHTHost, "DHT host to bind to")
	flags.Bool("pex-enabled", defaults.Torrent.PEXEnabled, "enable Peer Exchange (PEX)")
	flags.Bool("rpc-enabled", defaults.Torrent.RPCEnabled, "enable Rain RPC server")
	flags.String("rpc-host", defaults.Torrent.RPCHost, "RPC server host")
	flags.Uint16("rpc-port", defaults.Torrent.RPCPort, "RPC server port")
	flags.Uint("max-open-files", defaults.Torrent.MaxOpenFiles, "maximum number of open files")
	flags.Int64("piece-cache-size", defaults.Torrent.PieceCacheSize, "piece cache size in bytes")
	flags.Int("parallel-metadata", defaults.Torrent.ParallelMetadata, "number of parallel metadata downloads")
	flags.Bool("blocklist-enabled", defaults.Torrent.BlocklistEnabled, "enable peer blocklist")
	flags.String("blocklist-url", defaults.Torrent.BlocklistURL, "URL to download blocklist from")
	flags.Duration("resume-write-interval", defaults.Torrent.ResumeWriteInterval, "interval to write resume data")
	flags.Duration("health-check-interval", defaults.Torrent.HealthCheckInterval, "interval to check torrent health")
	flags.Duration(
		"health-check-timeout",
		defaults.Torrent.HealthCheckTimeout,
		"timeout before crashing on unresponsive torrent",
	)
}
