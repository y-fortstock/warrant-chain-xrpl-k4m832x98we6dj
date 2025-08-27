// Package logger provides logging functionality for the XRPL blockchain service.
// It implements structured logging using Go's standard slog package with support
// for different log levels and output formats.
package logger

import (
	"log/slog"
	"os"

	"gitlab.com/warrant1/warrant/chain-xrpl/internal/config"
)

// NewLogger returns a new slog.Logger instance that writes to standard output.
// It configures the logger based on the provided LogConfig settings.
//
// The function supports two output formats:
// - "logfmt": Human-readable text format (default)
// - "json": Structured JSON format for machine processing
//
// Supported log levels:
// - "debug": Most verbose logging, includes all messages
// - "info": Standard logging level, excludes debug messages
// - "warn": Only warning and error messages
// - "error": Only error messages
// - Any other value defaults to "info"
//
// Parameters:
// - cfg: Logging configuration specifying level and format
//
// Returns a configured slog.Logger instance that writes to stdout.
// The logger is ready to use immediately after creation.
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
