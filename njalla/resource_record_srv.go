package njalla

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/Sighery/terraform-provider-njalla/internal/client"
)

func resourceRecordSRV() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRecordSRVCreate,
		ReadContext:   resourceRecordSRVRead,
		UpdateContext: resourceRecordSRVUpdate,
		DeleteContext: resourceRecordSRVDelete,

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
			"priority": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Priority for the record.",
				ValidateFunc: validation.IntInSlice(client.ValidPriority),
			},
			"weight": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Weight for the record.",
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Port of the service.",
				ValidateFunc: validation.IsPortNumberOrZero,
			},
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Target host of the service.",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceRecordSRVImport,
		},
	}
}

func resourceRecordSRVCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	if err := createRecord(ctx, config.Client, "SRV", d); err != nil {
		return diag.FromErr(err)
	}

	return resourceRecordSRVRead(ctx, d, m)
}

func resourceRecordSRVRead(
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

	setRecord("SRV", d, record)
	return diags
}

func resourceRecordSRVUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	if err := updateRecord(ctx, config.Client, "SRV", d); err != nil {
		return diag.FromErr(err)
	}

	return resourceRecordSRVRead(ctx, d, m)
}

func resourceRecordSRVDelete(
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

func resourceRecordSRVImport(
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

	d.SetId(id)
	d.Set("domain", domain)
	setRecord("SRV", d, record)

	return []*schema.ResourceData{d}, nil
}
