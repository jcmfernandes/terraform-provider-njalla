package njalla

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/Sighery/terraform-provider-njalla/internal/client"
)

func resourceRecordNS() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRecordNSCreate,
		ReadContext:   resourceRecordNSRead,
		UpdateContext: resourceRecordNSUpdate,
		DeleteContext: resourceRecordNSDelete,

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Specifies the domain this record will be applied to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
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
				Description: "Content for the record.",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceRecordNSImport,
		},
	}
}

func resourceRecordNSCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	record := client.Record{
		Type:    "NS",
		Name:    d.Get("name").(string),
		Content: d.Get("content").(string),
		TTL:     d.Get("ttl").(int),
	}

	saved, err := config.Client.AddRecord(ctx, domain, record)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(saved.ID)

	return resourceRecordNSRead(ctx, d, m)

}

func resourceRecordNSRead(
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
	d.Set("content", record.Content)

	return diags
}

func resourceRecordNSUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	updateRecord := client.Record{
		ID:      d.Id(),
		Name:    d.Get("name").(string),
		Type:    "NS",
		Content: d.Get("content").(string),
		TTL:     d.Get("ttl").(int),
	}

	err := config.Client.EditRecord(ctx, domain, updateRecord)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRecordNSRead(ctx, d, m)
}

func resourceRecordNSDelete(
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

func resourceRecordNSImport(
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
	d.Set("content", record.Content)

	return []*schema.ResourceData{d}, nil
}
