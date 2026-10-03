package njalla

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomains() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainsRead,

		Schema: map[string]*schema.Schema{
			"domains": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Domains in the account.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expiry": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDomainsRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domains, err := listDomains(ctx, config.Client)
	if err != nil {
		return diag.FromErr(err)
	}

	result := make([]map[string]any, 0, len(domains))
	for _, domain := range domains {
		result = append(result, map[string]any{
			"name":   domain.Name,
			"status": domain.Status,
			"expiry": domain.Expiry,
		})
	}

	d.SetId("domains")
	if err := d.Set("domains", result); err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}
