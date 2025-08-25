package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/CreatureDev/xrpl-go/model/client/faucet"
)

type Faucet interface {
	FundAccount(*faucet.FundAccountRequest) (*faucet.FundAccountResponse, XRPLResponse, error)
}

type faucetImpl struct {
	client Client
}

func (f *faucetImpl) doRequest(ctx context.Context, client *http.Client, url string, body []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	httpReq.Header.Add("Content-Type", "application/json")

	httpReq = httpReq.WithContext(ctx)
	response, err := client.Do(httpReq)
	if err != nil || response == nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	return response, nil
}

func (f *faucetImpl) doWithRetry(ctx context.Context, client *http.Client, url string, body []byte, maxRetries int) (*http.Response, error) {
	backoff := time.Second
	for i := 0; i < maxRetries; i++ {
		resp, err := f.doRequest(ctx, client, url, body)
		if err != nil {
			return nil, fmt.Errorf("request: %w", err)
		}

		if resp.StatusCode != http.StatusServiceUnavailable {
			return resp, nil
		}

		// Close body to avoid leaking resources
		resp.Body.Close()

		if i < maxRetries-1 {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	return nil, fmt.Errorf("server overloaded, rate limit exceeded after %d retries", maxRetries)
}

func (f *faucetImpl) FundAccount(req *faucet.FundAccountRequest) (*faucet.FundAccountResponse, XRPLResponse, error) {
	if req.UserAgent == "" {
		req.UserAgent = "xrpl.go"
	}
	url := f.client.Faucet()

	httpClient := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal request: %w", err)
	}

	response, err := f.doWithRetry(ctx, httpClient, url, body, 3)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response: %w", err)
	}

	var ret faucet.FundAccountResponse
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return nil, nil, fmt.Errorf("fund unmarshal: %w", err)
	}

	return &ret, nil, nil
}
