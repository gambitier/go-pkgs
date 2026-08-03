package logging

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/gambitier/go-pkgs/logging/correlation"
)

type logrusFields = logrus.Fields

// Logger provides immutable, scoped, structured logging methods.
type Logger interface {
	WithFields(fields Fields) Logger
	WithCorrelationID(correlationID string) Logger
	WithContext(ctx context.Context) Logger
	AddHook(hook logrus.Hook)

	Debug(message string, fields Fields)
	Info(message string, fields Fields)
	Warn(message string, fields Fields)
	Error(message string, err error, fields Fields)
	Fatal(message string, fields Fields)
}

type logrusLogger struct {
	entry *logrus.Entry
}

// NewDefault creates a logger with default settings (JSON format, info level).
// For bootstrap scenarios where no config is available yet.
func NewDefault(serviceName string) Logger {
	base := logrus.New()
	base.SetLevel(logrus.InfoLevel)
	base.SetFormatter(&logrus.JSONFormatter{})
	return &logrusLogger{
		entry: logrus.NewEntry(base).WithField("service", serviceName),
	}
}

// New creates an independent logger instance.
func New(cfg Config) (Logger, error) {
	base := logrus.New()
	if cfg.Output != nil {
		base.SetOutput(cfg.Output)
	} else if output, err := cfg.BuildOutput(); err != nil {
		return nil, fmt.Errorf("build log output: %w", err)
	} else if output != nil {
		base.SetOutput(output)
	}

	level := strings.TrimSpace(strings.ToLower(cfg.Level))
	if level == "" {
		level = "info"
	}
	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}
	base.SetLevel(parsedLevel)

	format := strings.TrimSpace(strings.ToLower(cfg.Format))
	switch format {
	case "", "json":
		base.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		base.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	default:
		return nil, fmt.Errorf("unsupported log format %q", cfg.Format)
	}

	fields := cloneFields(cfg.BaseFields)
	if cfg.ServiceName != "" {
		if fields == nil {
			fields = make(logrus.Fields, 1)
		}
		fields["service"] = cfg.ServiceName
	}

	return &logrusLogger{
		entry: logrus.NewEntry(base).WithFields(fields),
	}, nil
}

func (l *logrusLogger) WithFields(fields Fields) Logger {
	return &logrusLogger{entry: l.entry.WithFields(cloneFields(fields))}
}

func (l *logrusLogger) WithCorrelationID(correlationID string) Logger {
	if strings.TrimSpace(correlationID) == "" {
		return l
	}
	return &logrusLogger{entry: l.entry.WithField("correlation_id", correlationID)}
}

func (l *logrusLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}
	scoped := l.entry.WithContext(ctx)
	if correlationID := correlation.GetCorrelationID(ctx); correlationID != "" {
		scoped = scoped.WithField("correlation_id", correlationID)
	}
	if traceFields := TraceFields(ctx); len(traceFields) > 0 {
		scoped = scoped.WithFields(cloneFields(traceFields))
	}
	return &logrusLogger{entry: scoped}
}

func (l *logrusLogger) AddHook(hook logrus.Hook) {
	if hook == nil {
		return
	}
	l.entry.Logger.AddHook(hook)
}

func (l *logrusLogger) Debug(message string, fields Fields) {
	l.entry.WithFields(cloneFields(fields)).Debug(message)
}

func (l *logrusLogger) Info(message string, fields Fields) {
	l.entry.WithFields(cloneFields(fields)).Info(message)
}

func (l *logrusLogger) Warn(message string, fields Fields) {
	l.entry.WithFields(cloneFields(fields)).Warn(message)
}

func (l *logrusLogger) Error(message string, err error, fields Fields) {
	merged := cloneFields(fields)
	if err != nil {
		merged = enrichErrorLogFields(merged, err)
	}
	l.entry.WithFields(merged).Error(message)
}

// enrichErrorLogFields adds a basic "error" field. Domain-specific enrichment
// lives in go-pkgs/errors.LogFields; apps adapt via platform glue into Fields.
func enrichErrorLogFields(fields logrusFields, err error) logrusFields {
	if err == nil {
		return fields
	}
	if fields == nil {
		fields = make(logrusFields, 1)
	}
	if _, exists := fields["error"]; !exists {
		fields["error"] = err.Error()
	}
	return fields
}

func (l *logrusLogger) Fatal(message string, fields Fields) {
	l.entry.WithFields(cloneFields(fields)).Fatal(message)
}
