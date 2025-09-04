package rpc

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/common"
)

var ErrEmptyURL = errors.New("empty port and IP provided")

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Config struct {
	HTTPClient HTTPClient
	URL        string
	Headers    map[string][]string

	// Retry config
	maxRetries int
	retryDelay time.Duration

	// Fee config
	maxFeeXRP  float32
	feeCushion float32

	// Faucet config
	faucetProvider common.FaucetProvider

	timeout time.Duration
}

type ConfigOpt func(c *Config)

func WithHTTPClient(cl HTTPClient) ConfigOpt {
	return func(c *Config) {
		c.HTTPClient = cl
	}
}

func WithMaxFeeXRP(maxFeeXRP float32) ConfigOpt {
	return func(c *Config) {
		c.maxFeeXRP = maxFeeXRP
	}
}

func WithFeeCushion(feeCushion float32) ConfigOpt {
	return func(c *Config) {
		c.feeCushion = feeCushion
	}
}

func WithFaucetProvider(fp common.FaucetProvider) ConfigOpt {
	return func(c *Config) {
		c.faucetProvider = fp
	}
}

func WithTimeout(timeout time.Duration) ConfigOpt {
	return func(c *Config) {
		c.timeout = timeout
		if hc, ok := c.HTTPClient.(*http.Client); ok {
			hc.Timeout = timeout
		}
	}
}

func NewClientConfig(url string, opts ...ConfigOpt) (*Config, error) {

	// validate a url has been passed in
	if len(url) == 0 {
		return nil, ErrEmptyURL
	}
	// add slash if doesn't already end with one
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}

	cfg := &Config{
		HTTPClient: &http.Client{},
		URL:        url,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},

		maxRetries: common.DefaultMaxRetries,
		retryDelay: common.DefaultRetryDelay,

		maxFeeXRP:  common.DefaultMaxFeeXRP,
		feeCushion: common.DefaultFeeCushion,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// Ensure the HTTPClient has the correct timeout if user did not set one
	if hc, ok := cfg.HTTPClient.(*http.Client); ok && cfg.timeout == 0 {
		hc.Timeout = common.DefaultTimeout
	}

	return cfg, nil
}
