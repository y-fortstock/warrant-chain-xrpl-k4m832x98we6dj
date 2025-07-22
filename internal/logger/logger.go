package logger

import (
	"log/slog"
	"os"
)

// NewLogger returns a new slog.Logger instance that writes to standard output.
// In the future, this can be extended to support configurable log levels and formats.
func NewLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}
