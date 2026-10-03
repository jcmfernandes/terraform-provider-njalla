package njalla

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServerImages() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServerImagesRead,

		Schema: map[string]*schema.Schema{
			"images": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Server images available for new servers.",
			},
		},
	}
}

func dataSourceServerImagesRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	data, err := config.Client.Request(
		ctx, "list-server-images", map[string]any{},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	var response struct {
		Images []string `json:"images"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("server_images")
	d.Set("images", response.Images)

	var diags diag.Diagnostics
	return diags
}
