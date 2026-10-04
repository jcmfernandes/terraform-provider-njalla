// Package client talks to Njalla's JSON-RPC API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultEndpoint is Njalla's only API endpoint.
const DefaultEndpoint = "https://njal.la/api/1/"

// DefaultTimeout bounds a single API call.
const DefaultTimeout = 60 * time.Second

// DefaultMaxRetries and DefaultRetryWait bound retries of busy responses.
// The wait doubles after each retry: 1s, 2s, 4s, 8s, 16s.
const (
	DefaultMaxRetries = 5
	DefaultRetryWait  = time.Second
)

// Client sends requests to Njalla's API. Endpoint and HTTPClient can be
// overridden, e.g. to point at an httptest server.
type Client struct {
	Token      string
	Endpoint   string
	HTTPClient *http.Client
	MaxRetries int
	RetryWait  time.Duration
}

// New returns a Client for the given API token.
func New(token string) *Client {
	return &Client{
		Token:      token,
		Endpoint:   DefaultEndpoint,
		HTTPClient: &http.Client{Timeout: DefaultTimeout},
		MaxRetries: DefaultMaxRetries,
		RetryWait:  DefaultRetryWait,
	}
}

// Error is an `error` object returned by the API.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("Njalla API error %d: %s", e.Code, e.Message)
}

// Request calls the given API method and returns its raw `result`.
func (c *Client) Request(
	ctx context.Context, method string, params map[string]any,
) ([]byte, error) {
	body, err := json.Marshal(map[string]any{
		"method": method,
		"params": params,
	})
	if err != nil {
		return nil, err
	}

	// Njalla answers 429 or 503 when it's busy, e.g. when tofu refreshes
	// many records at once.
	var resp *http.Response
	var data []byte
	wait := c.RetryWait
	for attempt := 0; ; attempt++ {
		resp, data, err = c.post(ctx, body)
		if err != nil {
			return nil, err
		}
		busy := resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode == http.StatusServiceUnavailable
		if !busy || attempt == c.MaxRetries {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
		wait *= 2
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(data, &response); err != nil {
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Njalla API returned HTTP %s", resp.Status)
		}
		return nil, err
	}

	if result, ok := response["result"]; ok {
		return result, nil
	}

	if raw, ok := response["error"]; ok {
		var apiErr *Error
		if json.Unmarshal(raw, &apiErr) == nil && apiErr != nil {
			return nil, apiErr
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Njalla API returned HTTP %s", resp.Status)
	}

	return nil, fmt.Errorf("Missing result %s", data)
}

// post sends one request and reads the whole response body.
func (c *Client) post(
	ctx context.Context, body []byte,
) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.Endpoint, bytes.NewReader(body),
	)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Njalla "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp, data, nil
}
