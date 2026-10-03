package njalla

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Sighery/terraform-provider-njalla/internal/client"
)

// apiToken is a `list-tokens` entry.
type apiToken struct {
	Key             string   `json:"key"`
	Comment         string   `json:"comment"`
	From            []string `json:"from"`
	AllowedDomains  []string `json:"allowed_domains"`
	AllowedServers  []string `json:"allowed_servers"`
	AllowedMethods  []string `json:"allowed_methods"`
	AllowedPrefixes []string `json:"allowed_prefixes"`
	AllowedTypes    []string `json:"allowed_types"`
}

// apiTokenLists are the list arguments shared by add-token and edit-token.
var apiTokenLists = []string{
	"from",
	"allowed_domains",
	"allowed_servers",
	"allowed_methods",
	"allowed_prefixes",
	"allowed_types",
}

// add-token returns nothing, so the new key is found by listing before and
// after adding it. The key is the token's only identifier, but it's a
// secret, so the resource ID is its SHA-256 hash.
func resourceAPIToken() *schema.Resource {
	stringSet := func(description string, computed bool) *schema.Schema {
		return &schema.Schema{
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    computed,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: description,
		}
	}

	return &schema.Resource{
		CreateContext: resourceAPITokenCreate,
		ReadContext:   resourceAPITokenRead,
		UpdateContext: resourceAPITokenUpdate,
		DeleteContext: resourceAPITokenDelete,

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Comment for the token.",
			},
			"from": stringSet(
				"IPv4 or IPv6 addresses or networks allowed to use the token.",
				false,
			),
			"allowed_domains": stringSet(
				"Domains the token is restricted to.", false,
			),
			"allowed_servers": stringSet(
				"Server IDs the token is restricted to.", false,
			),
			// `acme` fills in the following three, hence Computed.
			"allowed_methods": stringSet(
				"API methods the token is restricted to.", true,
			),
			"allowed_prefixes": stringSet(
				"DNS record name prefixes the token is restricted to.", true,
			),
			"allowed_types": stringSet(
				"DNS record types the token is restricted to.", true,
			),
			"acme": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
				Description: "Add the methods, prefixes and types needed for " +
					"the ACME DNS challenge.",
			},
			"key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The API token.",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceAPITokenImport,
		},
	}
}

func resourceAPITokenCreate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	params := map[string]any{}
	if v, ok := d.GetOk("comment"); ok {
		params["comment"] = v
	}
	for _, key := range apiTokenLists {
		if v, ok := d.GetOk(key); ok {
			params[key] = v.(*schema.Set).List()
		}
	}
	if d.Get("acme").(bool) {
		params["acme"] = true
	}

	before, err := listTokens(ctx, config.Client)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = config.Client.Request(ctx, "add-token", params)
	if err != nil {
		return diag.FromErr(err)
	}

	after, err := listTokens(ctx, config.Client)
	if err != nil {
		return diag.FromErr(err)
	}

	key, err := newTokenKey(before, after)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(apiTokenID(key))
	d.Set("key", key)

	return resourceAPITokenRead(ctx, d, m)
}

func resourceAPITokenRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	tokens, err := listTokens(ctx, config.Client)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, token := range tokens {
		if token.Key == d.Get("key").(string) {
			setAPIToken(d, token)

			return diags
		}
	}

	d.SetId("")
	return diags
}

func resourceAPITokenUpdate(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	// Send every argument, so removed restrictions get cleared.
	params := map[string]any{
		"key":     d.Get("key").(string),
		"comment": d.Get("comment").(string),
	}
	for _, key := range apiTokenLists {
		params[key] = d.Get(key).(*schema.Set).List()
	}

	_, err := config.Client.Request(ctx, "edit-token", params)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceAPITokenRead(ctx, d, m)
}

func resourceAPITokenDelete(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	config := m.(*Config)

	params := map[string]any{
		"key": d.Get("key").(string),
	}

	_, err := config.Client.Request(ctx, "remove-token", params)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}

// resourceAPITokenImport takes the token key as the import ID.
func resourceAPITokenImport(
	ctx context.Context, d *schema.ResourceData, m any,
) ([]*schema.ResourceData, error) {
	key := d.Id()

	config := m.(*Config)

	tokens, err := listTokens(ctx, config.Client)
	if err != nil {
		return nil, fmt.Errorf("Reading API tokens failed: %s", err.Error())
	}

	for _, token := range tokens {
		if token.Key == key {
			d.SetId(apiTokenID(key))
			d.Set("key", key)
			// list-tokens can't tell whether `acme` was used.
			d.Set("acme", false)
			setAPIToken(d, token)

			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, fmt.Errorf("Couldn't find the given API token")
}

func setAPIToken(d *schema.ResourceData, token apiToken) {
	d.Set("comment", token.Comment)
	d.Set("from", token.From)
	d.Set("allowed_domains", token.AllowedDomains)
	d.Set("allowed_servers", token.AllowedServers)
	d.Set("allowed_methods", token.AllowedMethods)
	d.Set("allowed_prefixes", token.AllowedPrefixes)
	d.Set("allowed_types", token.AllowedTypes)
}

func apiTokenID(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// newTokenKey returns the key of the only token in `after` that isn't in
// `before`.
func newTokenKey(before []apiToken, after []apiToken) (string, error) {
	known := map[string]bool{}
	for _, token := range before {
		known[token.Key] = true
	}

	var added []string
	for _, token := range after {
		if !known[token.Key] {
			added = append(added, token.Key)
		}
	}

	if len(added) != 1 {
		return "", fmt.Errorf(
			"Expected 1 new API token after adding, found %d", len(added),
		)
	}

	return added[0], nil
}

func listTokens(ctx context.Context, c *client.Client) ([]apiToken, error) {
	data, err := c.Request(ctx, "list-tokens", map[string]any{})
	if err != nil {
		return nil, err
	}

	var response struct {
		Tokens []apiToken `json:"tokens"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return response.Tokens, nil
}
