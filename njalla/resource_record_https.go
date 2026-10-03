package njalla

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/Sighery/gonjalla"
)

func resourceRecordHTTPS() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRecordHTTPSCreate,
		ReadContext:   resourceRecordHTTPSRead,
		UpdateContext: resourceRecordHTTPSUpdate,
		DeleteContext: resourceRecordHTTPSDelete,

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Specifies the domain this record will be applied to.",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: func() (any, error) {
					return "@", nil
				},
				Description: "Name for the record.",
			},
			"priority": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Priority for the record.",
				ValidateFunc: validation.IntInSlice(gonjalla.ValidPriority),
			},
			"target": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Target name for the record.",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceRecordHTTPSImport,
		},
	}
}

func resourceRecordHTTPSCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	if err := createRecord(config.Token, "HTTPS", d); err != nil {
		return diag.FromErr(err)
	}

	return resourceRecordHTTPSRead(ctx, d, m)
}

func resourceRecordHTTPSRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	var diags diag.Diagnostics

	record, err := findRecord(config.Token, domain, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if record == nil {
		d.SetId("")
		return diags
	}

	setRecord("HTTPS", d, record)
	return diags
}

func resourceRecordHTTPSUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	if err := updateRecord(config.Token, "HTTPS", d); err != nil {
		return diag.FromErr(err)
	}

	return resourceRecordHTTPSRead(ctx, d, m)
}

func resourceRecordHTTPSDelete(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	err := gonjalla.RemoveRecord(config.Token, domain, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}

func resourceRecordHTTPSImport(
	ctx context.Context, d *schema.ResourceData, m any,
) ([]*schema.ResourceData, error) {
	domain, id, err := parseImportID(d.Id())
	if err != nil {
		return nil, err
	}

	config := m.(*Config)

	record, err := findRecord(config.Token, domain, id)
	if err != nil {
		return nil, fmt.Errorf(
			"Reading records for domain %s failed: %s", domain, err.Error(),
		)
	}

	if record == nil {
		return nil, fmt.Errorf(
			"Couldn't find record %s for domain %s", id, domain,
		)
	}

	d.SetId(id)
	d.Set("domain", domain)
	setRecord("HTTPS", d, record)

	return []*schema.ResourceData{d}, nil
}
