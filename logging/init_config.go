package logging

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// InitConfig reads logging config from viper and applies sane defaults.
// The caller can still override fields after this returns.
func InitConfig(serviceName string) (Config, error) {
	cfg := Config{
		ServiceName: strings.TrimSpace(serviceName),
		Level:       "info",
		Format:      "json",
	}

	if subv := viper.Sub("logging"); subv != nil {
		if err := subv.Unmarshal(&cfg); err != nil {
			return Config{}, fmt.Errorf("unmarshal logging config: %w", err)
		}
	}

	if strings.TrimSpace(cfg.ServiceName) == "" {
		cfg.ServiceName = strings.TrimSpace(serviceName)
	}
	if strings.TrimSpace(cfg.Level) == "" {
		cfg.Level = "info"
	}
	if strings.TrimSpace(cfg.Format) == "" {
		cfg.Format = "json"
	}

	return cfg, nil
}
