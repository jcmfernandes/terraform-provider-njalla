package njalla

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRecordHTTPS_Create(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordHTTPSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordHTTPSCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordHTTPSExists("njalla_record_https.test_create"),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_create", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_create", "name", "testacc1-https-create-name",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_create", "priority", "1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_create", "target", ".",
					),
				),
			},
		},
	})
}

func TestAccRecordHTTPS_Update(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordHTTPSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordHTTPSUpdatePre(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordHTTPSExists("njalla_record_https.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_update", "name", "testacc2-https-update-name1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_update", "priority", "1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_update", "target", ".",
					),
				),
			},
			{
				Config: testAccCheckRecordHTTPSUpdatePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordHTTPSExists("njalla_record_https.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_update", "name", "testacc2-https-update-name2",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_update", "priority", "10",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_https.test_update", "target", "testacc.example.com",
					),
				),
			},
		},
	})
}

func TestAccRecordHTTPS_Import(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordHTTPSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordHTTPSImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordHTTPSExists("njalla_record_https.test_import"),
				),
			},
			{
				ResourceName:        "njalla_record_https.test_import",
				ImportStateIdPrefix: fmt.Sprintf("%s:", domain),
				ImportState:         true,
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccRecordHTTPS_InvalidPriority(t *testing.T) {
	expectedErr := regexp.MustCompile("expected priority to be one of .+, got 7")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordHTTPSDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckRecordHTTPSInvalidPriority(),
				ExpectError: expectedErr,
			},
		},
	})
}

func testAccCheckRecordHTTPSDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_record_https" {
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

func testAccCheckRecordHTTPSExists(resource string) resource.TestCheckFunc {
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

		if record.Type != "HTTPS" {
			return fmt.Errorf(
				"Record %s has type %s, expected HTTPS",
				rs.Primary.ID, record.Type,
			)
		}

		return nil
	}
}

func testAccCheckRecordHTTPSCreate() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_https test_create {
  domain = %q
  name = "testacc1-https-create-name"
  priority = 1
  target = "."
}
`, domain)
}

func testAccCheckRecordHTTPSUpdatePre() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_https test_update {
  domain = %q
  name = "testacc2-https-update-name1"
  priority = 1
  target = "."
}
`, domain)
}

func testAccCheckRecordHTTPSUpdatePost() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_https test_update {
  domain = %q
  name = "testacc2-https-update-name2"
  priority = 10
  target = "testacc.example.com"
}
`, domain)
}

func testAccCheckRecordHTTPSImport() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_https test_import {
  domain = %q
  name = "testacc3-https-import-name"
  priority = 1
  target = "."
}
`, domain)
}

func testAccCheckRecordHTTPSInvalidPriority() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_https test_invalid_priority {
  domain = %q
  name = "testacc4-https-invalidpriority-name"
  priority = 7
  target = "."
}
`, domain)
}
