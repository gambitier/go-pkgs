package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

func (cfg Config) BuildOutput() (io.Writer, error) {
	if len(cfg.Sinks) == 0 {
		return os.Stdout, nil
	}
	if err := validateSinks(cfg.Sinks); err != nil {
		return nil, err
	}

	writers := make([]io.Writer, 0, len(cfg.Sinks))
	for idx, sink := range cfg.Sinks {
		if !sink.Enabled {
			continue
		}
		writer, err := sink.buildWriter(idx)
		if err != nil {
			return nil, err
		}
		writers = append(writers, writer)
	}

	if len(writers) == 0 {
		return nil, fmt.Errorf("logging.sinks must have at least one enabled sink")
	}
	if len(writers) == 1 {
		return writers[0], nil
	}
	return io.MultiWriter(writers...), nil
}

func validateSinks(sinks []SinkConfig) error {
	if len(sinks) == 0 {
		return nil
	}

	enabled := 0
	for idx, sink := range sinks {
		if !sink.Enabled {
			continue
		}
		enabled++
		if err := sink.validate(idx); err != nil {
			return err
		}
	}
	if enabled == 0 {
		return fmt.Errorf("logging.sinks must have at least one enabled sink")
	}
	return nil
}

func (sink SinkConfig) validate(idx int) error {
	sinkType := strings.ToLower(strings.TrimSpace(sink.Type))
	validators := map[string]func(int, SinkConfig) error{
		"console": validateConsoleSink,
		"file":    validateFileSink,
	}
	validateFn, ok := validators[sinkType]
	if !ok {
		return fmt.Errorf("logging.sinks[%d].type must be one of: console, file", idx)
	}
	return validateFn(idx, sink)
}

func validateConsoleSink(_ int, _ SinkConfig) error {
	return nil
}

func validateFileSink(idx int, sink SinkConfig) error {
	if strings.TrimSpace(sink.File.Path) == "" {
		return fmt.Errorf("logging.sinks[%d].file.path is required for file sink", idx)
	}
	return nil
}

func (sink SinkConfig) buildWriter(idx int) (io.Writer, error) {
	switch strings.ToLower(strings.TrimSpace(sink.Type)) {
	case "console":
		return os.Stdout, nil
	case "file":
		path := strings.TrimSpace(sink.File.Path)
		if path == "" {
			return nil, fmt.Errorf("logging.sinks[%d].file.path is required for file sink", idx)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create log directory for sink %d: %w", idx, err)
		}
		return &lumberjack.Logger{
			Filename:   path,
			MaxSize:    withDefault(sink.File.MaxSizeMB, 100),
			MaxBackups: withDefault(sink.File.MaxBackups, 5),
			MaxAge:     withDefault(sink.File.MaxAgeDays, 30),
			Compress:   sink.File.Compress,
		}, nil
	default:
		return nil, fmt.Errorf("logging.sinks[%d].type must be one of: console, file", idx)
	}
}

func withDefault(value int, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
