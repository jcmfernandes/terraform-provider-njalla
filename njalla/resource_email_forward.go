package njalla

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Sighery/gonjalla"
)

// emailForward is a `list-forwards` entry.
type emailForward struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Forwards have neither ID nor edit call, so every argument forces a new
// forward and the resource ID is `domain:from:to`.
func resourceEmailForward() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEmailForwardCreate,
		ReadContext:   resourceEmailForwardRead,
		DeleteContext: resourceEmailForwardDelete,

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Specifies the domain mail is forwarded from.",
			},
			"from": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Address mail is forwarded from.",
			},
			"to": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Address mail is forwarded to.",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceEmailForwardImport,
		},
	}
}

func resourceEmailForwardCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)
	from := d.Get("from").(string)
	to := d.Get("to").(string)

	params := map[string]any{
		"domain": domain,
		"from":   from,
		"to":     to,
	}

	_, err := gonjalla.Request(config.Token, "add-forward", params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", domain, from, to))

	return resourceEmailForwardRead(ctx, d, m)
}

func resourceEmailForwardRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	domain := d.Get("domain").(string)

	var diags diag.Diagnostics

	forwards, err := listForwards(config.Token, domain)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, forward := range forwards {
		if forward.From == d.Get("from").(string) &&
			forward.To == d.Get("to").(string) {
			return diags
		}
	}

	d.SetId("")
	return diags
}

func resourceEmailForwardDelete(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	params := map[string]any{
		"domain": d.Get("domain").(string),
		"from":   d.Get("from").(string),
		"to":     d.Get("to").(string),
	}

	_, err := gonjalla.Request(config.Token, "remove-forward", params)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}

func resourceEmailForwardImport(
	ctx context.Context, d *schema.ResourceData, m any,
) ([]*schema.ResourceData, error) {
	domain, from, to, err := parseEmailForwardID(d.Id())
	if err != nil {
		return nil, err
	}

	config := m.(*Config)

	forwards, err := listForwards(config.Token, domain)
	if err != nil {
		return nil, fmt.Errorf(
			"Reading email forwards for domain %s failed: %s",
			domain, err.Error(),
		)
	}

	for _, forward := range forwards {
		if forward.From == from && forward.To == to {
			d.Set("domain", domain)
			d.Set("from", from)
			d.Set("to", to)

			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, fmt.Errorf(
		"Couldn't find email forward from %s to %s for domain %s",
		from, to, domain,
	)
}

// parseEmailForwardID parses an ID with the format `domain:from:to`.
func parseEmailForwardID(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		msg := fmt.Errorf(
			"unexpected format of ID (%s), expected domain:from:to", id,
		)
		return "", "", "", msg
	}

	return parts[0], parts[1], parts[2], nil
}

func listForwards(token string, domain string) ([]emailForward, error) {
	data, err := gonjalla.Request(
		token, "list-forwards", map[string]any{"domain": domain},
	)
	if err != nil {
		return nil, err
	}

	var response struct {
		Forwards []emailForward `json:"forwards"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return response.Forwards, nil
}
