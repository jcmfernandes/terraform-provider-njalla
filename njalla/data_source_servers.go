package njalla

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServers() *schema.Resource {
	computed := func(t schema.ValueType) *schema.Schema {
		return &schema.Schema{Type: t, Computed: true}
	}

	return &schema.Resource{
		ReadContext: dataSourceServersRead,

		Schema: map[string]*schema.Schema{
			"servers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Servers in the account.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":           computed(schema.TypeString),
						"name":         computed(schema.TypeString),
						"type":         computed(schema.TypeString),
						"os":           computed(schema.TypeString),
						"status":       computed(schema.TypeString),
						"os_state":     computed(schema.TypeString),
						"expiry":       computed(schema.TypeString),
						"autorenew":    computed(schema.TypeBool),
						"ssh_key":      computed(schema.TypeString),
						"reverse_name": computed(schema.TypeString),
						"ips": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceServersRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	servers, err := config.Client.ListServers(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	result := make([]map[string]any, 0, len(servers))
	for _, server := range servers {
		result = append(result, map[string]any{
			"id":           server.ID,
			"name":         server.Name,
			"type":         server.Type,
			"os":           server.Os,
			"status":       server.Status,
			"os_state":     server.OsState,
			"expiry":       server.Expiry,
			"autorenew":    server.Autorenew,
			"ssh_key":      server.SSHKey,
			"reverse_name": server.ReverseName,
			"ips":          server.Ips,
		})
	}

	d.SetId("servers")
	if err := d.Set("servers", result); err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}
