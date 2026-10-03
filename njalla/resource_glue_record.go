package njalla

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/Sighery/terraform-provider-njalla/internal/client"
)

// glueRecord is a `list-glue` entry.
type glueRecord struct {
	Name     string `json:"name"`
	Address4 string `json:"address4"`
	Address6 string `json:"address6"`
}

// Glue records have no ID; they're identified by domain and name, so the
// resource ID is `domain:name`.
func resourceGlueRecord() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGlueRecordCreate,
		ReadContext:   resourceGlueRecordRead,
		UpdateContext: resourceGlueRecordUpdate,
		DeleteContext: resourceGlueRecordDelete,

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Specifies the domain this glue record belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Subdomain of the nameserver.",
			},
			"address4": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "IPv4 address of the nameserver.",
				ValidateFunc: validation.IsIPv4Address,
				AtLeastOneOf: []string{"address4", "address6"},
			},
			"address6": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "IPv6 address of the nameserver.",
				ValidateFunc: validation.IsIPv6Address,
				AtLeastOneOf: []string{"address4", "address6"},
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceGlueRecordImport,
		},
	}
}

func resourceGlueRecordCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	_, err := config.Client.Request(ctx, "add-glue", glueParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf(
		"%s:%s", d.Get("domain").(string), d.Get("name").(string),
	))

	return resourceGlueRecordRead(ctx, d, m)
}

func resourceGlueRecordRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)
	name := d.Get("name").(string)

	var diags diag.Diagnostics

	glue, err := listGlue(ctx, config.Client, domain)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, record := range glue {
		if record.Name == name {
			d.Set("address4", record.Address4)
			d.Set("address6", record.Address6)

			return diags
		}
	}

	d.SetId("")
	return diags
}

func resourceGlueRecordUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	_, err := config.Client.Request(ctx, "edit-glue", glueParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceGlueRecordRead(ctx, d, m)
}

func resourceGlueRecordDelete(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	params := map[string]any{
		"domain": d.Get("domain").(string),
		"name":   d.Get("name").(string),
	}

	_, err := config.Client.Request(ctx, "remove-glue", params)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}

func resourceGlueRecordImport(
	ctx context.Context, d *schema.ResourceData, m any,
) ([]*schema.ResourceData, error) {
	domain, name, err := parseImportID(d.Id())
	if err != nil {
		return nil, err
	}

	config := m.(*Config)

	glue, err := listGlue(ctx, config.Client, domain)
	if err != nil {
		return nil, fmt.Errorf(
			"Reading glue records for domain %s failed: %s",
			domain, err.Error(),
		)
	}

	for _, record := range glue {
		if record.Name == name {
			d.Set("domain", domain)
			d.Set("name", record.Name)
			d.Set("address4", record.Address4)
			d.Set("address6", record.Address6)

			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, fmt.Errorf(
		"Couldn't find glue record %s for domain %s", name, domain,
	)
}

// glueParams always sends both addresses, an unset one as an empty string.
func glueParams(d *schema.ResourceData) map[string]any {
	return map[string]any{
		"domain":   d.Get("domain").(string),
		"name":     d.Get("name").(string),
		"address4": d.Get("address4").(string),
		"address6": d.Get("address6").(string),
	}
}

func listGlue(
	ctx context.Context, c *client.Client, domain string,
) ([]glueRecord, error) {
	data, err := c.Request(
		ctx, "list-glue", map[string]any{"domain": domain},
	)
	if err != nil {
		return nil, err
	}

	var response struct {
		Glue []glueRecord `json:"glue"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return response.Glue, nil
}
