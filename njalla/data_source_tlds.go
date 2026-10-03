package njalla

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Sighery/gonjalla"
)

// get-tlds returns an object keyed by TLD; it's exposed as a list sorted by
// name, since maps can't hold objects.
func dataSourceTLDs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTLDsRead,

		Schema: map[string]*schema.Schema{
			"tlds": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "TLDs Njalla can register.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"price": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_year": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"dnssec": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTLDsRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	data, err := gonjalla.Request(config.Token, "get-tlds", map[string]any{})
	if err != nil {
		return diag.FromErr(err)
	}

	var tlds map[string]struct {
		Price   int  `json:"price"`
		MaxYear int  `json:"max_year"`
		DNSSEC  bool `json:"dnssec"`
	}
	if err := json.Unmarshal(data, &tlds); err != nil {
		return diag.FromErr(err)
	}

	names := make([]string, 0, len(tlds))
	for name := range tlds {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]map[string]any, 0, len(names))
	for _, name := range names {
		result = append(result, map[string]any{
			"name":     name,
			"price":    tlds[name].Price,
			"max_year": tlds[name].MaxYear,
			"dnssec":   tlds[name].DNSSEC,
		})
	}

	d.SetId("tlds")
	if err := d.Set("tlds", result); err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}
