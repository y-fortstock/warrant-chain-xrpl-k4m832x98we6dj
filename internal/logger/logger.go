package logger

import (
	"log/slog"
	"os"
)

type LoggerConfig struct {
	Level  slog.Level
	Format string
}

// NewLogger returns a new slog.Logger instance that writes to standard output.
// Accepts a LoggerConfig parameter to set the log level and format ("logfmt" or "json").
func NewLogger(cfg LoggerConfig) *slog.Logger {
	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Level})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Level})
	}
	return slog.New(handler)
}
