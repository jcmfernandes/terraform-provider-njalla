package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

// newTestClient returns a Client pointed at a server running handler.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	c := New("secret")
	c.Endpoint = server.URL
	return c
}

func TestRequestSuccess(t *testing.T) {
	var gotAuth string
	var gotBody map[string]any
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(`{"jsonrpc": "2.0", "result": {"id": "42"}}`))
	})

	data, err := c.Request(
		context.Background(), "get-domain", map[string]any{"domain": "a.b"},
	)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != `{"id": "42"}` {
		t.Errorf("result = %s", data)
	}
	if gotAuth != "Njalla secret" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	want := map[string]any{
		"method": "get-domain",
		"params": map[string]any{"domain": "a.b"},
	}
	if !reflect.DeepEqual(gotBody, want) {
		t.Errorf("body = %v, want %v", gotBody, want)
	}
}

func TestRequestAPIError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(
			`{"jsonrpc": "2.0", "error": {"code": 403, "message": "Denied"}}`,
		))
	})

	_, err := c.Request(context.Background(), "list-domains", nil)

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if apiErr.Code != 403 || apiErr.Message != "Denied" {
		t.Errorf("err = %+v", apiErr)
	}
}

func TestRequestHTTPError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	})

	_, err := c.Request(context.Background(), "list-domains", nil)
	if err == nil || err.Error() != "Njalla API returned HTTP 502 Bad Gateway" {
		t.Errorf("err = %v", err)
	}
}

func TestRequestMissingResult(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"jsonrpc": "2.0"}`))
	})

	_, err := c.Request(context.Background(), "list-domains", nil)
	if err == nil {
		t.Error("err = nil, want missing result")
	}
}

func TestRequestTimeout(t *testing.T) {
	done := make(chan struct{})
	defer close(done)
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		<-done
	})
	c.HTTPClient.Timeout = 50 * time.Millisecond

	_, err := c.Request(context.Background(), "list-domains", nil)

	var netErr interface{ Timeout() bool }
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Errorf("err = %v, want timeout", err)
	}
}

func TestRequestContextCancelled(t *testing.T) {
	done := make(chan struct{})
	defer close(done)
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		<-done
	})

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	_, err := c.Request(ctx, "list-domains", nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

// AddRecord sends the record's fields, without an id.
func TestAddRecordParams(t *testing.T) {
	var gotBody struct {
		Params map[string]any `json:"params"`
	}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(`{"result": {"id": "7", "type": "MX", "prio": 10}}`))
	})

	prio := 10
	saved, err := c.AddRecord(context.Background(), "a.b", Record{
		Type:     "MX",
		Name:     "@",
		Content:  "mx.a.b",
		TTL:      3600,
		Priority: &prio,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]any{
		"domain":  "a.b",
		"name":    "@",
		"type":    "MX",
		"content": "mx.a.b",
		"ttl":     float64(3600),
		"prio":    float64(10),
	}
	if !reflect.DeepEqual(gotBody.Params, want) {
		t.Errorf("params = %v, want %v", gotBody.Params, want)
	}
	if saved.ID != "7" || saved.Priority == nil || *saved.Priority != 10 {
		t.Errorf("saved = %+v", saved)
	}
}
