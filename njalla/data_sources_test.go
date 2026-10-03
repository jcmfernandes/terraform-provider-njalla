package njalla

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceDomains(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data njalla_domains all {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.njalla_domains.all", "domains.0.name",
					),
				),
			},
		},
	})
}

func TestAccDataSourceDomain(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data njalla_domain test {
  name = %q
}
`, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.njalla_domain.test", "id", domain,
					),
					resource.TestCheckResourceAttrSet(
						"data.njalla_domain.test", "status",
					),
				),
			},
		},
	})
}

func TestAccDataSourceRecords(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordTXTDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource njalla_record_txt test {
  domain = %q
  name = "testacc1-records-datasource-name"
  ttl = 10800
  content = "testacc1-records-datasource-content"
}

data njalla_records test {
  domain = njalla_record_txt.test.domain
  depends_on = [njalla_record_txt.test]
}
`, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.njalla_records.test",
						"records.*",
						map[string]string{
							"name":    "testacc1-records-datasource-name",
							"type":    "TXT",
							"content": "testacc1-records-datasource-content",
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceTLDs(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data njalla_tlds all {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.njalla_tlds.all",
						"tlds.*",
						map[string]string{"name": "com"},
					),
				),
			},
		},
	})
}

func TestAccDataSourceServerTypesAndImages(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data njalla_server_types all {}
data njalla_server_images all {}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.njalla_server_types.all", "types.0",
					),
					resource.TestCheckResourceAttrSet(
						"data.njalla_server_images.all", "images.0",
					),
				),
			},
		},
	})
}

// The account may have no servers or VPNs, so these only check the calls
// succeed.
func TestAccDataSourceServersAndVPNs(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data njalla_servers all {}
data njalla_vpns all {}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.njalla_servers.all", "servers.#",
					),
					resource.TestCheckResourceAttrSet(
						"data.njalla_vpns.all", "vpns.#",
					),
				),
			},
		},
	})
}
