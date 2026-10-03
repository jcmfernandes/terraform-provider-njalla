package njalla

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Sighery/gonjalla"
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

	images, err := gonjalla.ListServerImages(config.Token)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("server_images")
	d.Set("images", images)

	var diags diag.Diagnostics
	return diags
}
