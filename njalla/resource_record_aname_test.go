package njalla

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRecordANAME_Create(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordANAMEDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordANAMECreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordANAMEExists("njalla_record_aname.test_create"),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_create", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_create", "name", "testacc1-aname-create-name",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_create", "ttl", "10800",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_create", "content", "testacc1.example.com",
					),
				),
			},
		},
	})
}

func TestAccRecordANAME_Update(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordANAMEDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordANAMEUpdatePre(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordANAMEExists("njalla_record_aname.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_update", "name", "testacc2-aname-update-name1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_update", "ttl", "10800",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_update", "content", "testacc2.example.com",
					),
				),
			},
			{
				Config: testAccCheckRecordANAMEUpdatePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordANAMEExists("njalla_record_aname.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_update", "name", "testacc2-aname-update-name2",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_update", "ttl", "3600",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_aname.test_update", "content", "testacc3.example.com",
					),
				),
			},
		},
	})
}

func TestAccRecordANAME_Import(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordANAMEDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordANAMEImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordANAMEExists("njalla_record_aname.test_import"),
				),
			},
			{
				ResourceName:        "njalla_record_aname.test_import",
				ImportStateIdPrefix: fmt.Sprintf("%s:", domain),
				ImportState:         true,
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccRecordANAME_InvalidTTL(t *testing.T) {
	expectedErr := regexp.MustCompile("expected ttl to be one of .+, got 999")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordANAMEDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckRecordANAMEInvalidTTL(),
				ExpectError: expectedErr,
			},
		},
	})
}

func testAccCheckRecordANAMEDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_record_aname" {
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

func testAccCheckRecordANAMEExists(resource string) resource.TestCheckFunc {
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

		if record.Type != "ANAME" {
			return fmt.Errorf(
				"Record %s has type %s, expected ANAME",
				rs.Primary.ID, record.Type,
			)
		}

		return nil
	}
}

func testAccCheckRecordANAMECreate() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_aname test_create {
  domain = %q
  name = "testacc1-aname-create-name"
  ttl = 10800
  content = "testacc1.example.com"
}
`, domain)
}

func testAccCheckRecordANAMEUpdatePre() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_aname test_update {
  domain = %q
  name = "testacc2-aname-update-name1"
  ttl = 10800
  content = "testacc2.example.com"
}
`, domain)
}

func testAccCheckRecordANAMEUpdatePost() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_aname test_update {
  domain = %q
  name = "testacc2-aname-update-name2"
  ttl = 3600
  content = "testacc3.example.com"
}
`, domain)
}

func testAccCheckRecordANAMEImport() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_aname test_import {
  domain = %q
  name = "testacc3-aname-import-name"
  ttl = 10800
  content = "testacc4.example.com"
}
`, domain)
}

func testAccCheckRecordANAMEInvalidTTL() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_aname test_invalid_ttl {
  domain = %q
  name = "testacc4-aname-invalidttl-name"
  ttl = 999
  content = "testacc5.example.com"
}
`, domain)
}
