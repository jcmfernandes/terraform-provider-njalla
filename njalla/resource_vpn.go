package njalla

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jcmfernandes/terraform-provider-njalla/internal/client"
)

// vpn is a `list-vpns`/`get-vpn` entry.
type vpn struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Autorenew bool   `json:"autorenew"`
	Backend   string `json:"backend"`
	PublicKey string `json:"publickey"`
	Expiry    string `json:"expiry"`
}

// add-vpn only takes `name` and `autorenew`; `backend` and `publickey` are
// set through edit-vpn afterwards. Renewals are left to `autorenew`.
func resourceVPN() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPNCreate,
		ReadContext:   resourceVPNRead,
		UpdateContext: resourceVPNUpdate,
		DeleteContext: resourceVPNDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VPN client.",
			},
			"autorenew": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to renew the VPN automatically.",
			},
			"backend": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice(
					[]string{"wireguard", "openvpn"}, false,
				),
				Description: "VPN backend, `wireguard` or `openvpn`.",
			},
			"publickey": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "WireGuard public key of the client.",
			},
			"expiry": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expiry date of the VPN.",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceVPNCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	params := map[string]any{
		"name":      d.Get("name").(string),
		"autorenew": d.Get("autorenew").(bool),
	}

	data, err := config.Client.Request(ctx, "add-vpn", params)
	if err != nil {
		return diag.FromErr(err)
	}

	var saved vpn
	if err := json.Unmarshal(data, &saved); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(saved.ID)

	edit := map[string]any{}
	for _, key := range []string{"backend", "publickey"} {
		if v, ok := d.GetOk(key); ok {
			edit[key] = v
		}
	}

	// A failure here is a warning: an error would taint the VPN, and
	// replacing it buys another one.
	var diags diag.Diagnostics
	if len(edit) > 0 {
		edit["id"] = saved.ID
		_, err := config.Client.Request(ctx, "edit-vpn", edit)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "VPN settings not applied",
				Detail: fmt.Sprintf(
					"%s. The next apply will retry them.", err.Error(),
				),
			})
		}
	}

	return append(diags, resourceVPNRead(ctx, d, m)...)
}

func resourceVPNRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	vpns, err := listVPNs(ctx, config.Client)
	if err != nil {
		return diag.FromErr(err)
	}

	exists := false
	for _, vpn := range vpns {
		if vpn.ID == d.Id() {
			exists = true
			break
		}
	}

	if !exists {
		d.SetId("")
		return diags
	}

	data, err := config.Client.Request(
		ctx, "get-vpn", map[string]any{"id": d.Id()},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	var current vpn
	if err := json.Unmarshal(data, &current); err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", current.Name)
	d.Set("autorenew", current.Autorenew)
	d.Set("backend", current.Backend)
	d.Set("publickey", current.PublicKey)
	d.Set("expiry", current.Expiry)

	return diags
}

func resourceVPNUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	params := map[string]any{"id": d.Id()}
	for _, key := range []string{"name", "autorenew", "backend", "publickey"} {
		if d.HasChange(key) {
			params[key] = d.Get(key)
		}
	}

	_, err := config.Client.Request(ctx, "edit-vpn", params)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVPNRead(ctx, d, m)
}

func resourceVPNDelete(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	_, err := config.Client.Request(
		ctx, "remove-vpn", map[string]any{"id": d.Id()},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}

func listVPNs(ctx context.Context, c *client.Client) ([]vpn, error) {
	data, err := c.Request(ctx, "list-vpns", map[string]any{})
	if err != nil {
		return nil, err
	}

	var response struct {
		VPNs []vpn `json:"vpns"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return response.VPNs, nil
}
