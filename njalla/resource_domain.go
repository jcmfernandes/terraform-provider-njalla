package njalla

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jcmfernandes/terraform-provider-njalla/internal/client"
)

// domainInfo is a `list-domains`/`get-domain` entry.
type domainInfo struct {
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	Expiry         string   `json:"expiry"`
	Locked         bool     `json:"locked"`
	Mailforwarding bool     `json:"mailforwarding"`
	DNSSEC         bool     `json:"dnssec"`
	MaxNameservers int      `json:"max_nameservers"`
	Nameservers    []string `json:"nameservers"`
}

// Njalla has no delete-domain, so destroying only forgets the domain.
// `years` only applies to the registration; renewals aren't managed.
func resourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainCreate,
		ReadContext:   resourceDomainRead,
		UpdateContext: resourceDomainUpdate,
		DeleteContext: resourceDomainDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Domain name to register.",
			},
			"years": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				Description:  "Years to register the domain for.",
				ValidateFunc: validation.IntAtLeast(1),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Id() != ""
				},
			},
			"mailforwarding": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether email forwarding is enabled.",
			},
			"dnssec": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether DNSSEC is enabled.",
			},
			"lock": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the domain is locked against transfers.",
			},
			"contacts": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Custom WHOIS contact IDs.",
			},
			"nameservers": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Custom nameservers. Empty uses Njalla's.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the domain.",
			},
			"expiry": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expiry date of the domain.",
			},
			"max_nameservers": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum number of custom nameservers.",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceDomainCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	name := d.Get("name").(string)

	params := map[string]any{
		"domain": name,
		"years":  d.Get("years").(int),
	}

	data, err := config.Client.Request(ctx, "register-domain", params)
	if err != nil {
		return diag.FromErr(err)
	}

	var task struct {
		Task string `json:"task"`
	}
	if err := json.Unmarshal(data, &task); err != nil {
		return diag.FromErr(err)
	}

	// Once Njalla has accepted the registration, problems are only warnings:
	// an error would taint the domain, and replacing it registers it again.
	var diags diag.Diagnostics

	err = waitForDomain(
		ctx, config.Client, name, task.Task, d.Timeout(schema.TimeoutCreate),
	)
	var failed *registrationError
	if errors.As(err, &failed) {
		return diag.FromErr(err)
	}

	d.SetId(name)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Domain registration not confirmed",
			Detail:   err.Error(),
		})
	}

	// Only send the settings present in the configuration.
	edit := map[string]any{}
	raw := d.GetRawConfig()
	for _, key := range []string{"mailforwarding", "dnssec", "lock"} {
		if !raw.GetAttr(key).IsNull() {
			edit[key] = d.Get(key)
		}
	}
	if v, ok := d.GetOk("contacts"); ok {
		edit["contacts"] = v
	}
	if v, ok := d.GetOk("nameservers"); ok {
		edit["nameservers"] = v
	}

	if err := editDomain(ctx, config.Client, name, edit); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Domain settings not applied",
			Detail: fmt.Sprintf(
				"%s. The next apply will retry them.", err.Error(),
			),
		})
	}

	return append(diags, resourceDomainRead(ctx, d, m)...)
}

// registrationError means Njalla reported the registration as failed.
type registrationError struct {
	domain string
	reason any
}

func (e *registrationError) Error() string {
	return fmt.Sprintf("Registering domain %s failed: %v", e.domain, e.reason)
}

func resourceDomainRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	domains, err := listDomains(ctx, config.Client)
	if err != nil {
		return diag.FromErr(err)
	}

	owned := false
	for _, domain := range domains {
		if domain.Name == d.Id() {
			owned = true
			break
		}
	}

	if !owned {
		d.SetId("")
		return diags
	}

	domain, err := getDomain(ctx, config.Client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", domain.Name)
	d.Set("mailforwarding", domain.Mailforwarding)
	d.Set("dnssec", domain.DNSSEC)
	d.Set("lock", domain.Locked)
	d.Set("nameservers", domain.Nameservers)
	d.Set("status", domain.Status)
	d.Set("expiry", domain.Expiry)
	d.Set("max_nameservers", domain.MaxNameservers)

	return diags
}

func resourceDomainUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	edit := map[string]any{}
	for _, key := range []string{
		"mailforwarding", "dnssec", "lock", "contacts", "nameservers",
	} {
		if d.HasChange(key) {
			edit[key] = d.Get(key)
		}
	}

	if err := editDomain(ctx, config.Client, d.Id(), edit); err != nil {
		return diag.FromErr(err)
	}

	return resourceDomainRead(ctx, d, m)
}

func resourceDomainDelete(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	return diag.Diagnostics{
		{
			Severity: diag.Warning,
			Summary:  "Domain removed from state only",
			Detail: fmt.Sprintf(
				"Njalla can't delete domains. %s stays registered to your "+
					"account until it expires.",
				d.Id(),
			),
		},
	}
}

// waitForDomain polls until the registration task finishes. check-task
// only documents `status` as an object, so completion is detected by the
// domain turning up as active.
func waitForDomain(
	ctx context.Context, c *client.Client, name string, task string,
	timeout time.Duration,
) error {
	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		data, err := c.Request(
			ctx, "check-task", map[string]any{"id": task},
		)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		var response struct {
			Status map[string]any `json:"status"`
		}
		if err := json.Unmarshal(data, &response); err != nil {
			return retry.NonRetryableError(err)
		}

		if msg, ok := response.Status["error"]; ok && msg != nil {
			return retry.NonRetryableError(
				&registrationError{domain: name, reason: msg},
			)
		}

		domains, err := listDomains(ctx, c)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		for _, domain := range domains {
			if domain.Name == name && domain.Status == "active" {
				return nil
			}
		}

		return retry.RetryableError(fmt.Errorf(
			"Domain %s not active yet, task status: %v",
			name, response.Status,
		))
	})
}

func editDomain(
	ctx context.Context, c *client.Client, name string, edit map[string]any,
) error {
	if len(edit) == 0 {
		return nil
	}

	edit["domain"] = name
	_, err := c.Request(ctx, "edit-domain", edit)
	return err
}

func listDomains(ctx context.Context, c *client.Client) ([]domainInfo, error) {
	data, err := c.Request(ctx, "list-domains", map[string]any{})
	if err != nil {
		return nil, err
	}

	var response struct {
		Domains []domainInfo `json:"domains"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return response.Domains, nil
}

func getDomain(
	ctx context.Context, c *client.Client, name string,
) (domainInfo, error) {
	data, err := c.Request(
		ctx, "get-domain", map[string]any{"domain": name},
	)
	if err != nil {
		return domainInfo{}, err
	}

	var response domainInfo
	if err := json.Unmarshal(data, &response); err != nil {
		return domainInfo{}, err
	}

	return response, nil
}
