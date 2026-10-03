package njalla

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jcmfernandes/terraform-provider-njalla/internal/client"
)

func resourceRecordTLSA() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRecordTLSACreate,
		ReadContext:   resourceRecordTLSARead,
		UpdateContext: resourceRecordTLSAUpdate,
		DeleteContext: resourceRecordTLSADelete,

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
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Content for the record.",
				ValidateFunc: validateTLSAContent,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceRecordTLSAImport,
		},
	}
}

func resourceRecordTLSACreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	record := client.Record{
		Type:    "TLSA",
		Name:    d.Get("name").(string),
		Content: d.Get("content").(string),
		TTL:     d.Get("ttl").(int),
	}

	saved, err := config.Client.AddRecord(ctx, domain, record)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(saved.ID)

	return resourceRecordTLSARead(ctx, d, m)

}

func resourceRecordTLSARead(
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

func resourceRecordTLSAUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	updateRecord := client.Record{
		ID:      d.Id(),
		Name:    d.Get("name").(string),
		Type:    "TLSA",
		Content: d.Get("content").(string),
		TTL:     d.Get("ttl").(int),
	}

	err := config.Client.EditRecord(ctx, domain, updateRecord)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRecordTLSARead(ctx, d, m)
}

func resourceRecordTLSADelete(
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

func resourceRecordTLSAImport(
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
	if err := checkRecordType(record, "TLSA"); err != nil {
		return nil, err
	}

	d.SetId(id)
	d.Set("domain", domain)
	d.Set("name", record.Name)
	d.Set("ttl", record.TTL)
	d.Set("content", record.Content)

	return []*schema.ResourceData{d}, nil
}

// validateTLSAContent will be the `ValidateFunc` used to check a given
// content for a TLSA DNS record matches the specification. If you're up for
// some heavy reading, check RFC 6698 points 2 and 7:
// https://tools.ietf.org/html/rfc6698
func validateTLSAContent(
	val any, key string,
) (warns []string, errs []error) {
	v := val.(string)
	values := strings.Split(v, " ")

	rfc := "Check RFC 6698 sections 2 and 7"

	if len(values) != 4 {
		msg := fmt.Errorf(
			"expected 4 arguments, got: %d. %s",
			len(values), rfc,
		)
		errs = append(errs, msg)
		return
	}

	certificateUsage, err := strconv.Atoi(values[0])
	if err != nil {
		msg := fmt.Errorf(
			"expected Certificate Usage field to be int, got: %s. %s",
			values[0], rfc,
		)
		errs = append(errs, msg)
		return
	}

	if certificateUsage < 0 || certificateUsage > 255 {
		msg := fmt.Errorf(
			"expected Certificate Usage field to be between 0 and 255 "+
				"(inclusive), got: %d. %s",
			certificateUsage, rfc,
		)
		errs = append(errs, msg)
		return
	}

	selector, err := strconv.Atoi(values[1])
	if err != nil {
		msg := fmt.Errorf(
			"expected Selector field to be int, got: %s. %s",
			values[1], rfc,
		)
		errs = append(errs, msg)
		return
	}

	if selector < 0 || selector > 255 {
		msg := fmt.Errorf(
			"expected Selector field to be between 0 and 255 (inclusive), "+
				"got: %d. %s",
			selector, rfc,
		)
		errs = append(errs, msg)
		return
	}

	matchingType, err := strconv.Atoi(values[2])
	if err != nil {
		msg := fmt.Errorf(
			"expected Matching Type field to be int, got: %s. %s",
			values[2], rfc,
		)
		errs = append(errs, msg)
		return
	}

	if matchingType < 0 || matchingType > 255 {
		msg := fmt.Errorf(
			"expected Matching Type field to be between 0 and 255 "+
				"(inclusive), got: %d. %s",
			matchingType, rfc,
		)
		errs = append(errs, msg)
		return
	}

	return
}
