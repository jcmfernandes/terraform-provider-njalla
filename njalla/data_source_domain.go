package njalla

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomain() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Domain name.",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiry": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mailforwarding": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"dnssec": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"lock": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"nameservers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"max_nameservers": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceDomainRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain, err := getDomain(ctx, config.Client, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(domain.Name)
	d.Set("status", domain.Status)
	d.Set("expiry", domain.Expiry)
	d.Set("mailforwarding", domain.Mailforwarding)
	d.Set("dnssec", domain.DNSSEC)
	d.Set("lock", domain.Locked)
	d.Set("nameservers", domain.Nameservers)
	d.Set("max_nameservers", domain.MaxNameservers)

	var diags diag.Diagnostics
	return diags
}
