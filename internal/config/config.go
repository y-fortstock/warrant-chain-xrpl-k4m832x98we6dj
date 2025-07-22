package config

import (
	"log/slog"

	"encoding/json"

	"github.com/spf13/viper"
	"github.com/ucarion/redact"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/logger"
)

// Config содержит параметры конфигурации приложения.
type Config struct {
	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"log"`
}

// LoadConfig загружает конфигурацию из Viper в структуру Config.
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LogLevel возвращает slog.Level, соответствующий строковому уровню логирования в конфиге.
func (c *Config) LogLevel() slog.Level {
	switch c.Log.Level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// LogFormat returns the log format ("logfmt" or "json"). Defaults to "logfmt" if not set or invalid.
func (c *Config) LogFormat() string {
	switch c.Log.Format {
	case "json":
		return "json"
	case "logfmt":
		return "logfmt"
	default:
		return "logfmt"
	}
}

// LoggerConfig returns a logger.LoggerConfig constructed from the config values.
func (c *Config) LoggerConfig() logger.LoggerConfig {
	return logger.LoggerConfig{
		Level:  c.LogLevel(),
		Format: c.LogFormat(),
	}
}

// RedactedConfigLog returns a string representation of the config with sensitive fields redacted.
// Uses github.com/ucarion/redact for redaction. Add sensitive field names to the slice as needed.
func (c *Config) RedactedConfigLog() string {
	// List of sensitive fields to redact (add as needed, e.g. "api_key", "password")
	sensitiveFields := [][]string{
		// Example: {"Database", "Password"},
	}
	cfgCopy := *c
	for _, path := range sensitiveFields {
		redact.Redact(path, &cfgCopy)
	}
	b, err := json.Marshal(cfgCopy)
	if err != nil {
		return "<failed to marshal config>"
	}
	return string(b)
}
