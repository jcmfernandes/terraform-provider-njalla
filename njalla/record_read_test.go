package njalla

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jcmfernandes/terraform-provider-njalla/internal/client"
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

// An MX record without `prio` must not crash Read or Import.
func TestRecordMXMissingPriority(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"result": {"records": [
				{"id": "2", "type": "MX", "name": "n", "content": "c", "ttl": 3600}
			]}}`))
		},
	))
	t.Cleanup(server.Close)

	c := client.New("secret")
	c.Endpoint = server.URL
	config := &Config{Client: c}
	r := resourceRecordMX()

	d := schema.TestResourceDataRaw(
		t, r.Schema, map[string]any{"domain": "a.b"},
	)
	d.SetId("2")
	if diags := r.ReadContext(context.Background(), d, config); diags.HasError() {
		t.Fatalf("Read failed: %v", diags)
	}

	d = schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	d.SetId("a.b:2")
	if _, err := r.Importer.StateContext(context.Background(), d, config); err != nil {
		t.Fatalf("Import failed: %v", err)
	}
}

// Importing a record ID into a resource of another record type must fail.
func TestRecordImportRejectsOtherTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"result": {"records": [
				{"id": "2", "type": "OTHER", "name": "n", "content": "c"}
			]}}`))
		},
	))
	t.Cleanup(server.Close)

	c := client.New("secret")
	c.Endpoint = server.URL
	config := &Config{Client: c}

	for name, r := range Provider().ResourcesMap {
		if !strings.HasPrefix(name, "njalla_record_") {
			continue
		}

		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
			d.SetId("a.b:2")
			_, err := r.Importer.StateContext(context.Background(), d, config)
			if err == nil || !strings.Contains(err.Error(), "OTHER") {
				t.Errorf("err = %v, want a type mismatch", err)
			}
		})
	}
}

// One malformed record must not stop the data source listing the others.
func TestRecordsDataSourceSkipsMalformedRecords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"result": {"records": [
				{"id": "1", "type": "SRV", "weight": "heavy"},
				{"id": "2", "type": "A", "name": "n", "content": "1.2.3.4"}
			]}}`))
		},
	))
	t.Cleanup(server.Close)

	c := client.New("secret")
	c.Endpoint = server.URL
	config := &Config{Client: c}
	r := dataSourceRecords()

	d := schema.TestResourceDataRaw(
		t, r.Schema, map[string]any{"domain": "a.b"},
	)
	diags := r.ReadContext(context.Background(), d, config)
	if diags.HasError() {
		t.Fatalf("Read failed: %v", diags)
	}
	if len(diags) != 1 || !strings.Contains(diags[0].Detail, "1") {
		t.Errorf("diags = %v, want one warning naming record 1", diags)
	}
	if n := d.Get("records.#").(int); n != 1 {
		t.Errorf("records = %d, want 1", n)
	}
}
