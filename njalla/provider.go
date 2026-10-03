package njalla

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Sighery/terraform-provider-njalla/internal/client"
)

// Provider for Njalla resources
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NJALLA_API_TOKEN", nil),
				Description: "Njalla API token",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"njalla_record_txt":     resourceRecordTXT(),
			"njalla_record_a":       resourceRecordA(),
			"njalla_record_aaaa":    resourceRecordAAAA(),
			"njalla_record_mx":      resourceRecordMX(),
			"njalla_record_cname":   resourceRecordCNAME(),
			"njalla_record_caa":     resourceRecordCAA(),
			"njalla_record_ptr":     resourceRecordPTR(),
			"njalla_record_ns":      resourceRecordNS(),
			"njalla_record_tlsa":    resourceRecordTLSA(),
			"njalla_record_naptr":   resourceRecordNAPTR(),
			"njalla_record_dynamic": resourceRecordDynamic(),
			"njalla_record_aname":   resourceRecordANAME(),
			"njalla_record_srv":     resourceRecordSRV(),
			"njalla_record_https":   resourceRecordHTTPS(),
			"njalla_record_svcb":    resourceRecordSVCB(),
			"njalla_record_sshfp":   resourceRecordSSHFP(),
			"njalla_record_ds":      resourceRecordDS(),
			"njalla_domain":         resourceDomain(),
			"njalla_glue_record":    resourceGlueRecord(),
			"njalla_dnssec_record":  resourceDNSSECRecord(),
			"njalla_email_forward":  resourceEmailForward(),
			"njalla_api_token":      resourceAPIToken(),
			"njalla_server":         resourceServer(),
			"njalla_vpn":            resourceVPN(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"njalla_domains":       dataSourceDomains(),
			"njalla_domain":        dataSourceDomain(),
			"njalla_records":       dataSourceRecords(),
			"njalla_server_types":  dataSourceServerTypes(),
			"njalla_server_images": dataSourceServerImages(),
			"njalla_servers":       dataSourceServers(),
			"njalla_tlds":          dataSourceTLDs(),
			"njalla_vpns":          dataSourceVPNs(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v, ok := d.GetOk("api_token"); ok {
		token := v.(string)
		config := Config{
			Client: client.New(token),
		}

		return &config, diags
	}

	// Reaching here means the token wasn't given through the Terraform
	// config NOR environment variable (`DefaultFunc`).
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to setup Njalla provider",
		Detail:   "Missing required API token for provider Njalla",
	})
	return nil, diags
}
