package njalla

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRecordDS_Create(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordDSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordDSCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordDSExists("njalla_record_ds.test_create"),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_create", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_create", "name", "testacc1-ds-create-name",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_create", "ttl", "10800",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_create", "content", "11111 13 2 aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					),
				),
			},
		},
	})
}

func TestAccRecordDS_Update(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordDSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordDSUpdatePre(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordDSExists("njalla_record_ds.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_update", "name", "testacc2-ds-update-name1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_update", "ttl", "10800",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_update", "content", "22222 13 2 aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					),
				),
			},
			{
				Config: testAccCheckRecordDSUpdatePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordDSExists("njalla_record_ds.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_update", "name", "testacc2-ds-update-name2",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_update", "ttl", "3600",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_ds.test_update", "content", "33333 13 2 bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
					),
				),
			},
		},
	})
}

func TestAccRecordDS_Import(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordDSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordDSImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordDSExists("njalla_record_ds.test_import"),
				),
			},
			{
				ResourceName:        "njalla_record_ds.test_import",
				ImportStateIdPrefix: fmt.Sprintf("%s:", domain),
				ImportState:         true,
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccRecordDS_InvalidTTL(t *testing.T) {
	expectedErr := regexp.MustCompile("expected ttl to be one of .+, got 999")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordDSDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckRecordDSInvalidTTL(),
				ExpectError: expectedErr,
			},
		},
	})
}

func testAccCheckRecordDSDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_record_ds" {
			continue
		}

		record, err := findRecord(
			context.Background(), config.Client, domain, rs.Primary.ID,
		)
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

func testAccCheckRecordDSExists(resource string) resource.TestCheckFunc {
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
		record, err := findRecord(
			context.Background(), config.Client, domain, rs.Primary.ID,
		)
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

		if record.Type != "DS" {
			return fmt.Errorf(
				"Record %s has type %s, expected DS",
				rs.Primary.ID, record.Type,
			)
		}

		return nil
	}
}

func testAccCheckRecordDSCreate() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_ds test_create {
  domain = %q
  name = "testacc1-ds-create-name"
  ttl = 10800
  content = "11111 13 2 aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
`, domain)
}

func testAccCheckRecordDSUpdatePre() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_ds test_update {
  domain = %q
  name = "testacc2-ds-update-name1"
  ttl = 10800
  content = "22222 13 2 aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
`, domain)
}

func testAccCheckRecordDSUpdatePost() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_ds test_update {
  domain = %q
  name = "testacc2-ds-update-name2"
  ttl = 3600
  content = "33333 13 2 bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
}
`, domain)
}

func testAccCheckRecordDSImport() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_ds test_import {
  domain = %q
  name = "testacc3-ds-import-name"
  ttl = 10800
  content = "44444 13 2 aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
`, domain)
}

func testAccCheckRecordDSInvalidTTL() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_ds test_invalid_ttl {
  domain = %q
  name = "testacc4-ds-invalidttl-name"
  ttl = 999
  content = "55555 13 2 aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
`, domain)
}
