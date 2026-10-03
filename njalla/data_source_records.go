package njalla

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	records, err := config.Client.ListRecords(ctx, domain)
	if err != nil {
		return diag.FromErr(err)
	}

	result := make([]map[string]any, 0, len(records))
	for _, r := range records {
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

	var diags diag.Diagnostics
	return diags
}
