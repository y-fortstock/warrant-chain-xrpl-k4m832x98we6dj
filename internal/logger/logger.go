package logger

import (
	"log/slog"
	"os"
)

// NewLogger returns a new slog.Logger instance that writes to standard output.
// Accepts a slog.Level parameter to set the log level.
func NewLogger(level slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}
