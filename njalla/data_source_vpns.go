package njalla

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVPNs() *schema.Resource {
	computed := func(t schema.ValueType) *schema.Schema {
		return &schema.Schema{Type: t, Computed: true}
	}

	return &schema.Resource{
		ReadContext: dataSourceVPNsRead,

		Schema: map[string]*schema.Schema{
			"vpns": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "VPN clients in the account.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":        computed(schema.TypeString),
						"name":      computed(schema.TypeString),
						"autorenew": computed(schema.TypeBool),
						"backend":   computed(schema.TypeString),
						"publickey": computed(schema.TypeString),
						"expiry":    computed(schema.TypeString),
					},
				},
			},
		},
	}
}

func dataSourceVPNsRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	vpns, err := listVPNs(ctx, config.Client)
	if err != nil {
		return diag.FromErr(err)
	}

	result := make([]map[string]any, 0, len(vpns))
	for _, vpn := range vpns {
		result = append(result, map[string]any{
			"id":        vpn.ID,
			"name":      vpn.Name,
			"autorenew": vpn.Autorenew,
			"backend":   vpn.Backend,
			"publickey": vpn.PublicKey,
			"expiry":    vpn.Expiry,
		})
	}

	d.SetId("vpns")
	if err := d.Set("vpns", result); err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}
