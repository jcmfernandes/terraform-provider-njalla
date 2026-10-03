package njalla

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jcmfernandes/terraform-provider-njalla/internal/client"
)

// fakeListAPI serves a list method and an add method that appends to it.
// Listing is slow, so concurrent creates overlap.
func fakeListAPI(
	t *testing.T, listMethod string, addMethod string, listKey string,
	newEntry func(n int) map[string]any,
) *Config {
	t.Helper()

	var mu sync.Mutex
	entries := []map[string]any{}

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Method string `json:"method"`
			}
			json.NewDecoder(r.Body).Decode(&body)

			switch body.Method {
			case listMethod:
				mu.Lock()
				snapshot := append([]map[string]any{}, entries...)
				mu.Unlock()
				time.Sleep(20 * time.Millisecond)
				json.NewEncoder(w).Encode(map[string]any{
					"result": map[string]any{listKey: snapshot},
				})
			case addMethod:
				mu.Lock()
				entries = append(entries, newEntry(len(entries)))
				mu.Unlock()
				w.Write([]byte(`{"result": {}}`))
			default:
				t.Errorf("unexpected method %s", body.Method)
			}
		},
	))
	t.Cleanup(server.Close)

	c := client.New("secret")
	c.Endpoint = server.URL
	return &Config{Client: c}
}

// createConcurrently runs two creates of the same resource at once, as
// tofu does for independent resources.
func createConcurrently(
	t *testing.T, r *schema.Resource, raw map[string]any, config *Config,
) {
	t.Helper()

	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d := schema.TestResourceDataRaw(t, r.Schema, raw)
			if diags := r.CreateContext(context.Background(), d, config); diags.HasError() {
				t.Errorf("Create failed: %v", diags)
			}
		}()
	}
	wg.Wait()
}

func TestAPITokenConcurrentCreate(t *testing.T) {
	config := fakeListAPI(t, "list-tokens", "add-token", "tokens",
		func(n int) map[string]any {
			return map[string]any{"key": fmt.Sprintf("key-%d", n)}
		},
	)
	createConcurrently(t, resourceAPIToken(), map[string]any{}, config)
}

func TestDNSSECRecordConcurrentCreate(t *testing.T) {
	config := fakeListAPI(t, "list-dnssec", "add-dnssec", "dnssec",
		func(n int) map[string]any {
			return map[string]any{"id": fmt.Sprint(n), "algorithm": 13}
		},
	)
	createConcurrently(t, resourceDNSSECRecord(), map[string]any{
		"domain":     "a.b",
		"algorithm":  13,
		"public_key": "key",
	}, config)
}
