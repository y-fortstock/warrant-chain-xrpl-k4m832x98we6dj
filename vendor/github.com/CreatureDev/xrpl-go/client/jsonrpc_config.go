package client

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

var ErrEmptyUrl = errors.New("empty port and IP provided")

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type JsonRpcConfig struct {
	HTTPClient HTTPClient
	Url        string
	Faucet     string
	Headers    map[string][]string
}

type JsonRpcConfigOpt func(c *JsonRpcConfig)

func WithHttpClient(cl HTTPClient) JsonRpcConfigOpt {
	return func(c *JsonRpcConfig) {
		c.HTTPClient = cl
	}
}

func WithFaucet(faucet string) JsonRpcConfigOpt {
	return func(c *JsonRpcConfig) {
		c.Faucet = faucet
	}
}

func NewJsonRpcConfig(url string, opts ...JsonRpcConfigOpt) (*JsonRpcConfig, error) {

	// validate a url has been passed in
	if len(url) == 0 {
		return nil, ErrEmptyUrl
	}
	// add slash if doesn't already end with one
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}

	cfg := &JsonRpcConfig{
		HTTPClient: &http.Client{Timeout: time.Duration(1) * time.Second}, // default timeout value - allow custom timme out?
		Url:        url,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Faucet: defaultFaucet(url),
	}

	for _, opt := range opts {
		opt(cfg)
	}
	return cfg, nil
}

func defaultFaucet(url string) string {
	if strings.Contains(url, "altnet") {
		return "https://faucet.altnet.rippletest.net/accounts"
	}
	if strings.Contains(url, "devnet") {
		return "https://faucet.devnet.rippletest.net/accounts"
	}
	return ""
}
