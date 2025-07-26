package config

import (
	"encoding/json"

	"github.com/spf13/viper"
	"github.com/ucarion/redact"
)

// LogConfig holds configuration for logging. Used by logger implementations.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// NetworkConfig holds configuration for XRPL network connection.
type NetworkConfig struct {
	URL     string `mapstructure:"url"`
	Timeout int64  `mapstructure:"timeout"`
}

// Config содержит параметры конфигурации приложения.
type Config struct {
	Log     LogConfig     `mapstructure:"log"`
	Network NetworkConfig `mapstructure:"network"`
	Server  struct {
		Listen string `mapstructure:"listen"`
	} `mapstructure:"server"`
}

// LoadConfig загружает конфигурацию из Viper в структуру Config.
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoggerConfig returns a LogConfig constructed from the config values.
func (c *Config) LoggerConfig() LogConfig {
	return c.Log
}

// NetworkConfig returns a NetworkConfig constructed from the config values.
func (c *Config) NetworkConfig() NetworkConfig {
	return c.Network
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
