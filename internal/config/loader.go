package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Load loads configuration from flags, env vars, and config file
func Load(cmd *cobra.Command) (*Config, error) {
	v := viper.New()

	explicitConfig := cmd.Flag("config").Value.String()

	// Set config file search paths
	if explicitConfig != "" {
		v.SetConfigFile(explicitConfig)
	} else {
		v.SetConfigName("fluxo")
		v.SetConfigType("yaml")

		// Add search paths. Use getHomeDir (same as DefaultConfig) so the
		// optional ~/.fluxo config path matches default database/data-dir roots
		// when $HOME and os.UserHomeDir diverge.
		v.AddConfigPath(".")
		v.AddConfigPath(filepath.Join(getHomeDir(), ".fluxo"))
		v.AddConfigPath("/etc/fluxo")
	}

	// Environment variables
	v.SetEnvPrefix("FLUXO")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// Read config file: missing optional search-path file is fine; parse errors
	// and an explicit --config path that cannot be read must fail.
	if err := v.ReadInConfig(); err != nil {
		if explicitConfig != "" {
			return nil, fmt.Errorf("reading config %q: %w", explicitConfig, err)
		}
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	// Bind command flags
	if err := v.BindPFlags(cmd.Flags()); err != nil {
		return nil, fmt.Errorf("binding flags: %w", err)
	}

	// Unmarshal to Config struct
	cfg := DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return cfg, nil
}
