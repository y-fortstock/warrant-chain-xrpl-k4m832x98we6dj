// Package config provides configuration management for the XRPL blockchain service.
// It handles loading and parsing of configuration files, environment variables,
// and provides structured access to application settings.
package config

import (
	"encoding/json"

	"github.com/spf13/viper"
	"github.com/ucarion/redact"
)

// LogConfig holds configuration for logging. Used by logger implementations.
// It specifies the log level and output format for the application.
type LogConfig struct {
	// Level specifies the minimum log level to output.
	// Valid values: "debug", "info", "warn", "error"
	Level string `mapstructure:"level"`

	// Format specifies the output format for log messages.
	// Valid values: "logfmt" (default), "json"
	Format string `mapstructure:"format"`
}

// NetworkConfig holds configuration for XRPL network connection.
// It specifies the RPC endpoint, timeout settings, and system account credentials.
type NetworkConfig struct {
	// URL specifies the XRPL RPC endpoint URL.
	// Example: "https://s.altnet.rippletest.net:51234"
	URL string `mapstructure:"url"`

	// Timeout specifies the network request timeout in seconds.
	// This applies to all RPC calls to the XRPL network.
	Timeout int64 `mapstructure:"timeout"`

	// System contains configuration for the system account used by the service.
	System struct {
		// Account specifies the system account's XRPL address.
		// This account is used for funding operations and token management.
		Account string `mapstructure:"account"`

		// Secret specifies the system account's private key.
		// This is used for signing transactions on behalf of the system.
		Secret string `mapstructure:"secret"`

		// Public specifies the system account's public key.
		// This is used for transaction validation and verification.
		Public string `mapstructure:"public"`
	} `mapstructure:"system"`
}

// FeatureConfig holds configuration for feature flags.
// It controls which features are enabled or disabled in the application.
type FeatureConfig struct {
	// Loan specifies whether the loan feature is enabled.
	// When true, loan-related functionality will be available.
	Loan bool `mapstructure:"loan"`
}

// Config contains all configuration parameters for the application.
// It aggregates settings from multiple sources and provides a unified interface.
type Config struct {
	// Log contains logging configuration settings.
	Log LogConfig `mapstructure:"log"`

	// Network contains XRPL network connection settings.
	Network NetworkConfig `mapstructure:"network"`

	// Features contains feature flag configuration settings.
	Features FeatureConfig `mapstructure:"features"`

	// Server contains HTTP/gRPC server configuration.
	Server struct {
		// Listen specifies the address and port for the server to listen on.
		// Example: ":8080" or "localhost:9090"
		Listen string `mapstructure:"listen"`
	} `mapstructure:"server"`
}

// LoadConfig loads configuration from Viper into the Config structure.
// It reads from configuration files, environment variables, and command line flags.
//
// Returns a populated Config instance or an error if loading fails.
// The configuration is automatically loaded from:
// - Configuration files (config.yaml, config.json, etc.)
// - Environment variables (prefixed with the application name)
// - Command line flags
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoggerConfig returns a LogConfig constructed from the config values.
// This method provides access to logging configuration in a structured format.
//
// Returns the LogConfig section of the main configuration.
func (c *Config) LoggerConfig() LogConfig {
	return c.Log
}

// NetworkConfig returns a NetworkConfig constructed from the config values.
// This method provides access to network configuration in a structured format.
//
// Returns the NetworkConfig section of the main configuration.
func (c *Config) NetworkConfig() NetworkConfig {
	return c.Network
}

// FeatureConfig returns a FeatureConfig constructed from the config values.
// This method provides access to feature configuration in a structured format.
//
// Returns the FeatureConfig section of the main configuration.
func (c *Config) FeatureConfig() *FeatureConfig {
	return &c.Features
}

// RedactedConfigLog returns a string representation of the config with sensitive fields redacted.
// Uses github.com/ucarion/redact for redaction to prevent logging of sensitive information
// like private keys, passwords, and API tokens.
//
// Sensitive fields are automatically masked in the output string.
// This is useful for logging configuration without exposing security-sensitive data.
//
// Returns a JSON string representation of the configuration with sensitive fields redacted.
// If marshaling fails, returns an error message string.
func (c *Config) RedactedConfigLog() string {
	// List of sensitive fields to redact (add as needed, e.g. "api_key", "password")
	sensitiveFields := [][]string{
		{"Network", "System", "Secret"},
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
