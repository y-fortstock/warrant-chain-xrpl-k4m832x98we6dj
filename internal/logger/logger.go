package logger

import (
	"log/slog"
	"os"

	"gitlab.com/warrant1/warrant/chain-xrpl/internal/config"
)

// NewLogger returns a new slog.Logger instance that writes to standard output.
// Accepts a config.LogConfig parameter to set the log level and format ("logfmt" or "json").
func NewLogger(cfg config.LogConfig) *slog.Logger {
	var handler slog.Handler
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	return slog.New(handler)
}
