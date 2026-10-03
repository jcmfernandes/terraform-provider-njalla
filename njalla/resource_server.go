package njalla

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/Sighery/terraform-provider-njalla/internal/client"
)

// Changing `os` replaces the server rather than calling reset-server, so
// the plan shows the data loss. `months` only applies to the purchase;
// renewals are left to `autorenew`.
func resourceServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the server.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Server type, from `list-server-types`.",
			},
			"os": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Server image, from `list-server-images`.",
			},
			"ssh_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Public SSH key installed on the server.",
			},
			"months": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				Description:  "Months to pay for upfront.",
				ValidateFunc: validation.IntBetween(1, 12),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Id() != ""
				},
			},
			"autorenew": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to renew the server automatically.",
			},
			"reverse_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Reverse DNS name of the server's addresses.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the server.",
			},
			"os_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "State of the server's OS installation.",
			},
			"expiry": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expiry date of the server.",
			},
			"ips": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IP addresses of the server.",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceServerCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	params := map[string]any{
		"name":      d.Get("name").(string),
		"type":      d.Get("type").(string),
		"os":        d.Get("os").(string),
		"ssh_key":   d.Get("ssh_key").(string),
		"months":    d.Get("months").(int),
		"autorenew": d.Get("autorenew").(bool),
	}

	data, err := config.Client.Request(ctx, "add-server", params)
	if err != nil {
		return diag.FromErr(err)
	}

	var saved client.Server
	if err := json.Unmarshal(data, &saved); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(saved.ID)

	// add-server doesn't take `reverse_name`. A failure here is a warning:
	// an error would taint the server, and replacing it buys another one.
	var diags diag.Diagnostics
	if v, ok := d.GetOk("reverse_name"); ok {
		edit := map[string]any{"id": saved.ID, "reverse_name": v}
		_, err := config.Client.Request(ctx, "edit-server", edit)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Server reverse name not applied",
				Detail: fmt.Sprintf(
					"%s. The next apply will retry it.", err.Error(),
				),
			})
		}
	}

	return append(diags, resourceServerRead(ctx, d, m)...)
}

func resourceServerRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	exists, err := serverExists(ctx, config.Client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if !exists {
		d.SetId("")
		return diags
	}

	data, err := config.Client.Request(
		ctx, "get-server", map[string]any{"id": d.Id()},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	var server client.Server
	if err := json.Unmarshal(data, &server); err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", server.Name)
	d.Set("type", server.Type)
	d.Set("os", server.Os)
	d.Set("ssh_key", server.SSHKey)
	d.Set("autorenew", server.Autorenew)
	d.Set("reverse_name", server.ReverseName)
	d.Set("status", server.Status)
	d.Set("os_state", server.OsState)
	d.Set("expiry", server.Expiry)
	d.Set("ips", server.Ips)

	return diags
}

func resourceServerUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	// Only send what changed, so an unchanged `type` can't trigger a resize.
	params := map[string]any{"id": d.Id()}
	for _, key := range []string{
		"name", "type", "ssh_key", "reverse_name", "autorenew",
	} {
		if d.HasChange(key) {
			params[key] = d.Get(key)
		}
	}

	_, err := config.Client.Request(ctx, "edit-server", params)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceServerRead(ctx, d, m)
}

func resourceServerDelete(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	data, err := config.Client.Request(
		ctx, "remove-server", map[string]any{"id": d.Id()},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	var removed client.Server
	if err := json.Unmarshal(data, &removed); err != nil {
		return diag.FromErr(err)
	}

	// remove-server returns a task; wait until the server is gone.
	timeout := d.Timeout(schema.TimeoutDelete)
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		exists, err := serverExists(ctx, config.Client, d.Id())
		if err != nil {
			return retry.NonRetryableError(err)
		}
		if exists {
			return retry.RetryableError(
				fmt.Errorf("Server %s still exists", d.Id()),
			)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}

func serverExists(
	ctx context.Context, c *client.Client, id string,
) (bool, error) {
	servers, err := c.ListServers(ctx)
	if err != nil {
		return false, err
	}

	for _, server := range servers {
		if server.ID == id {
			return true, nil
		}
	}

	return false, nil
}
