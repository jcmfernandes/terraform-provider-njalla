package njalla

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Sighery/gonjalla"
)

func dataSourceServerTypes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServerTypesRead,

		Schema: map[string]*schema.Schema{
			"types": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Server types available for new servers.",
			},
		},
	}
}

func dataSourceServerTypesRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	types, err := gonjalla.ListServerTypes(config.Token)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("server_types")
	d.Set("types", types)

	var diags diag.Diagnostics
	return diags
}
