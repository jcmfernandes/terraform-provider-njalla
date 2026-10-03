package njalla

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Sighery/terraform-provider-njalla/internal/client"
)

// A record of another type with an unexpected shape must not stop Read from
// finding the record it's looking for.
func TestRecordReadSkipsMalformedRecords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"result": {"records": [
				{"id": "1", "type": "SRV", "weight": "heavy"},
				{"id": "2", "name": "n", "content": "c", "ttl": 3600, "prio": 10}
			]}}`))
		},
	))
	t.Cleanup(server.Close)

	c := client.New("secret")
	c.Endpoint = server.URL
	config := &Config{Client: c}

	resources := []string{
		"njalla_record_a", "njalla_record_aaaa", "njalla_record_caa",
		"njalla_record_cname", "njalla_record_mx", "njalla_record_naptr",
		"njalla_record_ns", "njalla_record_ptr", "njalla_record_tlsa",
		"njalla_record_txt", "njalla_record_dynamic",
	}

	for _, name := range resources {
		r := Provider().ResourcesMap[name]

		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(
				t, r.Schema, map[string]any{"domain": "a.b"},
			)
			d.SetId("2")

			if diags := r.ReadContext(context.Background(), d, config); diags.HasError() {
				t.Fatalf("Read failed: %v", diags)
			}
			if d.Id() != "2" {
				t.Errorf("record not found, id = %q", d.Id())
			}
		})
	}
}
