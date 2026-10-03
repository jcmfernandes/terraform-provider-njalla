package njalla

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jcmfernandes/terraform-provider-njalla/internal/client"
)

func resourceRecordDS() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRecordDSCreate,
		ReadContext:   resourceRecordDSRead,
		UpdateContext: resourceRecordDSUpdate,
		DeleteContext: resourceRecordDSDelete,

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
			"ttl": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "TTL for the record.",
				ValidateFunc: validation.IntInSlice(client.ValidTTL),
			},
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "DS record data: key tag, algorithm, digest type and digest.",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceRecordDSImport,
		},
	}
}

func resourceRecordDSCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	if err := createRecord(ctx, config.Client, "DS", d); err != nil {
		return diag.FromErr(err)
	}

	return resourceRecordDSRead(ctx, d, m)
}

func resourceRecordDSRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	var diags diag.Diagnostics

	record, err := findRecord(ctx, config.Client, domain, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if record == nil {
		d.SetId("")
		return diags
	}

	setRecord("DS", d, record)
	return diags
}

func resourceRecordDSUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	if err := updateRecord(ctx, config.Client, "DS", d); err != nil {
		return diag.FromErr(err)
	}

	return resourceRecordDSRead(ctx, d, m)
}

func resourceRecordDSDelete(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	err := config.Client.RemoveRecord(ctx, domain, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}

func resourceRecordDSImport(
	ctx context.Context, d *schema.ResourceData, m any,
) ([]*schema.ResourceData, error) {
	domain, id, err := parseImportID(d.Id())
	if err != nil {
		return nil, err
	}

	config := m.(*Config)

	record, err := findRecord(ctx, config.Client, domain, id)
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
	if err := checkRecordType(record, "DS"); err != nil {
		return nil, err
	}

	d.SetId(id)
	d.Set("domain", domain)
	setRecord("DS", d, record)

	return []*schema.ResourceData{d}, nil
}
