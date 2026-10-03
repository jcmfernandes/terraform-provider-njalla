package njalla

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRecordSRV_Create(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSRVDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordSRVCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSRVExists("njalla_record_srv.test_create"),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_create", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_create", "name", "testacc1-srv-create-name",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_create", "ttl", "10800",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_create", "priority", "10",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_create", "weight", "5",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_create", "port", "5060",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_create", "content", "sip1.example.com",
					),
				),
			},
		},
	})
}

func TestAccRecordSRV_Update(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSRVDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordSRVUpdatePre(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSRVExists("njalla_record_srv.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "name", "testacc2-srv-update-name1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "ttl", "10800",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "priority", "10",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "weight", "5",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "port", "5060",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "content", "sip2.example.com",
					),
				),
			},
			{
				Config: testAccCheckRecordSRVUpdatePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSRVExists("njalla_record_srv.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "name", "testacc2-srv-update-name2",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "ttl", "3600",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "priority", "20",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "weight", "10",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "port", "5061",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_srv.test_update", "content", "sip3.example.com",
					),
				),
			},
		},
	})
}

func TestAccRecordSRV_Import(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSRVDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordSRVImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSRVExists("njalla_record_srv.test_import"),
				),
			},
			{
				ResourceName:        "njalla_record_srv.test_import",
				ImportStateIdPrefix: fmt.Sprintf("%s:", domain),
				ImportState:         true,
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccRecordSRV_InvalidPriority(t *testing.T) {
	expectedErr := regexp.MustCompile("expected priority to be one of .+, got 7")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSRVDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckRecordSRVInvalidPriority(),
				ExpectError: expectedErr,
			},
		},
	})
}

func TestAccRecordSRV_InvalidPort(t *testing.T) {
	expectedErr := regexp.MustCompile(`expected "port" to be a valid port number or 0, got: 70000`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSRVDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckRecordSRVInvalidPort(),
				ExpectError: expectedErr,
			},
		},
	})
}

func testAccCheckRecordSRVDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_record_srv" {
			continue
		}

		record, err := findRecord(config.Token, domain, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf(
				"Error fetching the records data for domain %s: %s",
				domain, err,
			)
		}

		if record != nil {
			return fmt.Errorf(
				"Record %s still exists in domain %s",
				rs.Primary.ID, domain,
			)
		}
	}

	return nil
}

func testAccCheckRecordSRVExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No record ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
		record, err := findRecord(config.Token, domain, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf(
				"Error fetching the records data for domain %s: %s",
				domain, err,
			)
		}

		if record == nil {
			return fmt.Errorf(
				"Record %s doesn't exist for domain %s", rs.Primary.ID, domain,
			)
		}

		if record.Type != "SRV" {
			return fmt.Errorf(
				"Record %s has type %s, expected SRV",
				rs.Primary.ID, record.Type,
			)
		}

		return nil
	}
}

func testAccCheckRecordSRVCreate() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_srv test_create {
  domain = %q
  name = "testacc1-srv-create-name"
  ttl = 10800
  priority = 10
  weight = 5
  port = 5060
  content = "sip1.example.com"
}
`, domain)
}

func testAccCheckRecordSRVUpdatePre() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_srv test_update {
  domain = %q
  name = "testacc2-srv-update-name1"
  ttl = 10800
  priority = 10
  weight = 5
  port = 5060
  content = "sip2.example.com"
}
`, domain)
}

func testAccCheckRecordSRVUpdatePost() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_srv test_update {
  domain = %q
  name = "testacc2-srv-update-name2"
  ttl = 3600
  priority = 20
  weight = 10
  port = 5061
  content = "sip3.example.com"
}
`, domain)
}

func testAccCheckRecordSRVImport() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_srv test_import {
  domain = %q
  name = "testacc3-srv-import-name"
  ttl = 10800
  priority = 10
  weight = 5
  port = 5060
  content = "sip4.example.com"
}
`, domain)
}

func testAccCheckRecordSRVInvalidPriority() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_srv test_invalid_priority {
  domain = %q
  name = "testacc4-srv-invalidpriority-name"
  ttl = 10800
  priority = 7
  weight = 5
  port = 5060
  content = "sip5.example.com"
}
`, domain)
}

func testAccCheckRecordSRVInvalidPort() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_srv test_invalid_port {
  domain = %q
  name = "testacc5-srv-invalidport-name"
  ttl = 10800
  priority = 10
  weight = 5
  port = 70000
  content = "sip6.example.com"
}
`, domain)
}
