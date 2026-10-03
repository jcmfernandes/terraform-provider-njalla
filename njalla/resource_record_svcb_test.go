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

func TestAccRecordSVCB_Create(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSVCBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordSVCBCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSVCBExists("njalla_record_svcb.test_create"),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_create", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_create", "name", "testacc1-svcb-create-name",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_create", "priority", "1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_create", "target", ".",
					),
				),
			},
		},
	})
}

func TestAccRecordSVCB_Update(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSVCBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordSVCBUpdatePre(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSVCBExists("njalla_record_svcb.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_update", "name", "testacc2-svcb-update-name1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_update", "priority", "1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_update", "target", ".",
					),
				),
			},
			{
				Config: testAccCheckRecordSVCBUpdatePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSVCBExists("njalla_record_svcb.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_update", "name", "testacc2-svcb-update-name2",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_update", "priority", "10",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_svcb.test_update", "target", "testacc.example.com",
					),
				),
			},
		},
	})
}

func TestAccRecordSVCB_Import(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSVCBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordSVCBImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSVCBExists("njalla_record_svcb.test_import"),
				),
			},
			{
				ResourceName:        "njalla_record_svcb.test_import",
				ImportStateIdPrefix: fmt.Sprintf("%s:", domain),
				ImportState:         true,
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccRecordSVCB_InvalidPriority(t *testing.T) {
	expectedErr := regexp.MustCompile("expected priority to be one of .+, got 7")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSVCBDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckRecordSVCBInvalidPriority(),
				ExpectError: expectedErr,
			},
		},
	})
}

func testAccCheckRecordSVCBDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_record_svcb" {
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

func testAccCheckRecordSVCBExists(resource string) resource.TestCheckFunc {
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

		if record.Type != "SVCB" {
			return fmt.Errorf(
				"Record %s has type %s, expected SVCB",
				rs.Primary.ID, record.Type,
			)
		}

		return nil
	}
}

func testAccCheckRecordSVCBCreate() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_svcb test_create {
  domain = %q
  name = "testacc1-svcb-create-name"
  priority = 1
  target = "."
}
`, domain)
}

func testAccCheckRecordSVCBUpdatePre() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_svcb test_update {
  domain = %q
  name = "testacc2-svcb-update-name1"
  priority = 1
  target = "."
}
`, domain)
}

func testAccCheckRecordSVCBUpdatePost() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_svcb test_update {
  domain = %q
  name = "testacc2-svcb-update-name2"
  priority = 10
  target = "testacc.example.com"
}
`, domain)
}

func testAccCheckRecordSVCBImport() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_svcb test_import {
  domain = %q
  name = "testacc3-svcb-import-name"
  priority = 1
  target = "."
}
`, domain)
}

func testAccCheckRecordSVCBInvalidPriority() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_svcb test_invalid_priority {
  domain = %q
  name = "testacc4-svcb-invalidpriority-name"
  priority = 7
  target = "."
}
`, domain)
}
