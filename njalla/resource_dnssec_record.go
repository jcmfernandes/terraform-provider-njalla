package njalla

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Sighery/terraform-provider-njalla/internal/client"
)

// dnssecRecord is a `list-dnssec` entry.
type dnssecRecord struct {
	ID         string `json:"id"`
	Algorithm  int    `json:"algorithm"`
	Digest     string `json:"digest"`
	DigestType int    `json:"digest_type"`
	KeyTag     int    `json:"key_tag"`
	PublicKey  string `json:"public_key"`
}

// There's no edit-dnssec, so every argument forces a new record. add-dnssec
// returns nothing, so the new record's ID is found by listing before and
// after adding it.
func resourceDNSSECRecord() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDNSSECRecordCreate,
		ReadContext:   resourceDNSSECRecordRead,
		DeleteContext: resourceDNSSECRecordDelete,

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Specifies the domain this record will be applied to.",
			},
			"algorithm": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "DNSSEC algorithm number.",
			},
			"digest": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "Digest of the DNSKEY.",
				ExactlyOneOf: []string{"digest", "public_key"},
				RequiredWith: []string{"digest", "digest_type", "key_tag"},
			},
			"digest_type": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Description:  "Digest type number.",
				RequiredWith: []string{"digest", "digest_type", "key_tag"},
			},
			"key_tag": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Description:  "Key tag of the DNSKEY.",
				RequiredWith: []string{"digest", "digest_type", "key_tag"},
			},
			"public_key": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "Public key of the DNSKEY.",
				ExactlyOneOf: []string{"digest", "public_key"},
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceDNSSECRecordImport,
		},
	}
}

func resourceDNSSECRecordCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	params := map[string]any{
		"domain":    domain,
		"algorithm": d.Get("algorithm").(int),
	}
	if v, ok := d.GetOk("public_key"); ok {
		params["public_key"] = v
	} else {
		params["digest"] = d.Get("digest").(string)
		params["digest_type"] = d.Get("digest_type").(int)
		params["key_tag"] = d.Get("key_tag").(int)
	}

	before, err := listDNSSEC(ctx, config.Client, domain)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = config.Client.Request(ctx, "add-dnssec", params)
	if err != nil {
		return diag.FromErr(err)
	}

	after, err := listDNSSEC(ctx, config.Client, domain)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := newDNSSECID(before, after)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return resourceDNSSECRecordRead(ctx, d, m)
}

func resourceDNSSECRecordRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	var diags diag.Diagnostics

	records, err := listDNSSEC(ctx, config.Client, domain)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, record := range records {
		if d.Id() == record.ID {
			setDNSSECRecord(d, record)

			return diags
		}
	}

	d.SetId("")
	return diags
}

func resourceDNSSECRecordDelete(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	params := map[string]any{
		"domain": d.Get("domain").(string),
		"id":     d.Id(),
	}

	_, err := config.Client.Request(ctx, "remove-dnssec", params)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}

func resourceDNSSECRecordImport(
	ctx context.Context, d *schema.ResourceData, m any,
) ([]*schema.ResourceData, error) {
	domain, id, err := parseImportID(d.Id())
	if err != nil {
		return nil, err
	}

	config := m.(*Config)

	records, err := listDNSSEC(ctx, config.Client, domain)
	if err != nil {
		return nil, fmt.Errorf(
			"Reading DNSSEC records for domain %s failed: %s",
			domain, err.Error(),
		)
	}

	for _, record := range records {
		if id == record.ID {
			d.SetId(id)
			d.Set("domain", domain)
			setDNSSECRecord(d, record)

			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, fmt.Errorf(
		"Couldn't find DNSSEC record %s for domain %s", id, domain,
	)
}

// setDNSSECRecord only sets the fields of the form the record was added
// with, either digest or public key.
func setDNSSECRecord(d *schema.ResourceData, record dnssecRecord) {
	d.Set("algorithm", record.Algorithm)
	if record.PublicKey != "" {
		d.Set("public_key", record.PublicKey)
		return
	}
	d.Set("digest", record.Digest)
	d.Set("digest_type", record.DigestType)
	d.Set("key_tag", record.KeyTag)
}

// newDNSSECID returns the ID of the only record in `after` that isn't in
// `before`.
func newDNSSECID(before []dnssecRecord, after []dnssecRecord) (string, error) {
	known := map[string]bool{}
	for _, record := range before {
		known[record.ID] = true
	}

	var added []string
	for _, record := range after {
		if !known[record.ID] {
			added = append(added, record.ID)
		}
	}

	if len(added) != 1 {
		return "", fmt.Errorf(
			"Expected 1 new DNSSEC record after adding, found %d", len(added),
		)
	}

	return added[0], nil
}

func listDNSSEC(
	ctx context.Context, c *client.Client, domain string,
) ([]dnssecRecord, error) {
	data, err := c.Request(
		ctx, "list-dnssec", map[string]any{"domain": domain},
	)
	if err != nil {
		return nil, err
	}

	var response struct {
		DNSSEC []dnssecRecord `json:"dnssec"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return response.DNSSEC, nil
}
