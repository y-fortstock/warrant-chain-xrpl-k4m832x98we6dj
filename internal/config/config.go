package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

// Config содержит параметры конфигурации приложения.
type Config struct {
	Log struct {
		Level string `mapstructure:"level"`
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
