package logging

import "io"

// Config defines logger setup for a single service instance.
type Config struct {
	ServiceName               string       `mapstructure:"service_name"`
	Level                     string       `mapstructure:"level" validate:"omitempty,oneof=debug info warn error"`
	Format                    string       `mapstructure:"format" validate:"omitempty,oneof=json text"`
	Sinks                     []SinkConfig `mapstructure:"sinks"`
	ErrorFrameIncludePrefixes []string     `mapstructure:"error_frame_include_prefixes"`
	ErrorFrameStripPrefixes   []string     `mapstructure:"error_frame_strip_prefixes"`
	ErrorFrameMaxDepth        int          `mapstructure:"error_frame_max_depth"`
	BaseFields                Fields       `mapstructure:"-"`
	Output                    io.Writer    `mapstructure:"-"`
}

type SinkConfig struct {
	Type    string         `mapstructure:"type"`
	Enabled bool           `mapstructure:"enabled"`
	File    FileSinkConfig `mapstructure:"file"`
}

type FileSinkConfig struct {
	Path       string `mapstructure:"path"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
}
