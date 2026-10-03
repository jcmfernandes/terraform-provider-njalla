package njalla

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jcmfernandes/terraform-provider-njalla/internal/client"
)

func resourceRecordMX() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRecordMXCreate,
		ReadContext:   resourceRecordMXRead,
		UpdateContext: resourceRecordMXUpdate,
		DeleteContext: resourceRecordMXDelete,

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
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Content for the record.",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceRecordMXImport,
		},
	}
}

func resourceRecordMXCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)
	priority := d.Get("priority").(int)

	record := client.Record{
		Type:     "MX",
		Name:     d.Get("name").(string),
		Content:  d.Get("content").(string),
		TTL:      d.Get("ttl").(int),
		Priority: &priority,
	}

	saved, err := config.Client.AddRecord(ctx, domain, record)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(saved.ID))

	return resourceRecordMXRead(ctx, d, m)

}

func resourceRecordMXRead(
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

	d.Set("name", record.Name)
	d.Set("ttl", record.TTL)
	d.Set("priority", recordValue(record, "priority"))
	d.Set("content", record.Content)

	return diags
}

func resourceRecordMXUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)
	priority := d.Get("priority").(int)

	updateRecord := client.Record{
		ID:       d.Id(),
		Name:     d.Get("name").(string),
		Type:     "MX",
		Content:  d.Get("content").(string),
		TTL:      d.Get("ttl").(int),
		Priority: &priority,
	}

	err := config.Client.EditRecord(ctx, domain, updateRecord)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRecordMXRead(ctx, d, m)
}

func resourceRecordMXDelete(
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

func resourceRecordMXImport(
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
		return nil, fmt.Errorf("Couldn't find record %s for domain %s", id, domain)
	}

	d.SetId(id)
	d.Set("domain", domain)
	d.Set("name", record.Name)
	d.Set("ttl", record.TTL)
	d.Set("priority", recordValue(record, "priority"))
	d.Set("content", record.Content)

	return []*schema.ResourceData{d}, nil
}
