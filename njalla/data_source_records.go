package njalla

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jcmfernandes/terraform-provider-njalla/internal/client"
)

// Fields a record type doesn't use are left at their zero value.
func dataSourceRecords() *schema.Resource {
	computed := func(t schema.ValueType) *schema.Schema {
		return &schema.Schema{Type: t, Computed: true}
	}

	return &schema.Resource{
		ReadContext: dataSourceRecordsRead,

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Domain to list the records of.",
			},
			"records": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "DNS records of the domain.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":            computed(schema.TypeString),
						"name":          computed(schema.TypeString),
						"type":          computed(schema.TypeString),
						"content":       computed(schema.TypeString),
						"ttl":           computed(schema.TypeInt),
						"priority":      computed(schema.TypeInt),
						"weight":        computed(schema.TypeInt),
						"port":          computed(schema.TypeInt),
						"target":        computed(schema.TypeString),
						"ssh_algorithm": computed(schema.TypeInt),
						"ssh_type":      computed(schema.TypeInt),
					},
				},
			},
		},
	}
}

func dataSourceRecordsRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	data, err := config.Client.Request(
		ctx, "list-records", map[string]any{"domain": domain},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	var response struct {
		Records []json.RawMessage `json:"records"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	// Decode entries one at a time, so a record with an unexpected shape is
	// skipped instead of failing the whole listing.
	result := make([]map[string]any, 0, len(response.Records))
	for _, raw := range response.Records {
		var r client.Record
		if err := json.Unmarshal(raw, &r); err != nil {
			// Only the ID: the record's content may be a DDNS key.
			var ref struct {
				ID any `json:"id"`
			}
			json.Unmarshal(raw, &ref)
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Skipped a record Njalla returned in an unexpected shape",
				Detail:   fmt.Sprintf("Record %v: %s", ref.ID, err),
			})
			continue
		}

		result = append(result, map[string]any{
			"id":            r.ID,
			"name":          r.Name,
			"type":          r.Type,
			"content":       r.Content,
			"ttl":           r.TTL,
			"priority":      recordValue(&r, "priority"),
			"weight":        r.Weight,
			"port":          r.Port,
			"target":        r.Target,
			"ssh_algorithm": r.SSHAlgorithm,
			"ssh_type":      r.SSHType,
		})
	}

	d.SetId(domain)
	if err := d.Set("records", result); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
