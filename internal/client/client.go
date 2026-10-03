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

// Client sends requests to Njalla's API. Endpoint and HTTPClient can be
// overridden, e.g. to point at an httptest server.
type Client struct {
	Token      string
	Endpoint   string
	HTTPClient *http.Client
}

// New returns a Client for the given API token.
func New(token string) *Client {
	return &Client{
		Token:      token,
		Endpoint:   DefaultEndpoint,
		HTTPClient: &http.Client{Timeout: DefaultTimeout},
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

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.Endpoint, bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Njalla "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
