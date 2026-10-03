package njalla

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/Sighery/gonjalla"
)

func resourceRecordSSHFP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRecordSSHFPCreate,
		ReadContext:   resourceRecordSSHFPRead,
		UpdateContext: resourceRecordSSHFPUpdate,
		DeleteContext: resourceRecordSSHFPDelete,

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
				ValidateFunc: validation.IntInSlice(gonjalla.ValidTTL),
			},
			"ssh_algorithm": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "SSH key algorithm: 1 RSA, 2 DSA, 3 ECDSA, 4 Ed25519, 5 XMSS.",
				ValidateFunc: validation.IntBetween(1, 5),
			},
			"ssh_type": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Fingerprint type: 1 SHA-1, 2 SHA-256.",
				ValidateFunc: validation.IntBetween(1, 2),
			},
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Hex-encoded fingerprint for the record.",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceRecordSSHFPImport,
		},
	}
}

func resourceRecordSSHFPCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	if err := createRecord(config.Token, "SSHFP", d); err != nil {
		return diag.FromErr(err)
	}

	return resourceRecordSSHFPRead(ctx, d, m)
}

func resourceRecordSSHFPRead(
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

	setRecord("SSHFP", d, record)
	return diags
}

func resourceRecordSSHFPUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	if err := updateRecord(config.Token, "SSHFP", d); err != nil {
		return diag.FromErr(err)
	}

	return resourceRecordSSHFPRead(ctx, d, m)
}

func resourceRecordSSHFPDelete(
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

func resourceRecordSSHFPImport(
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
	setRecord("SSHFP", d, record)

	return []*schema.ResourceData{d}, nil
}
