package njalla

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	data, err := config.Client.Request(
		ctx, "list-server-types", map[string]any{},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	var response struct {
		Types []string `json:"types"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("server_types")
	d.Set("types", response.Types)

	var diags diag.Diagnostics
	return diags
}
