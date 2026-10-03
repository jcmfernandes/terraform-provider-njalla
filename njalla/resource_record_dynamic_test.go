package njalla

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/Sighery/gonjalla"
)

func TestAccRecordDynamic_Create(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordDynamicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordDynamicCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordDynamicExists(
						"njalla_record_dynamic.test_create",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_dynamic.test_create", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_dynamic.test_create",
						"name",
						"testacc1-dynamic-create-name",
					),
					resource.TestCheckResourceAttrSet(
						"njalla_record_dynamic.test_create", "key",
					),
				),
			},
		},
	})
}

func TestAccRecordDynamic_Import(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordDynamicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordDynamicImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordDynamicExists(
						"njalla_record_dynamic.test_import",
					),
				),
			},
			{
				ResourceName:        "njalla_record_dynamic.test_import",
				ImportStateIdPrefix: fmt.Sprintf("%s:", domain),
				ImportState:         true,
				ImportStateVerify:   true,
			},
		},
	})
}

func testAccCheckRecordDynamicDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_record_dynamic" {
			continue
		}

		records, err := gonjalla.ListRecords(config.Token, domain)
		if err != nil {
			return fmt.Errorf(
				"Error fetching the records data for domain %s: %s",
				domain, err,
			)
		}

		for _, record := range records {
			if record.ID == rs.Primary.ID {
				return fmt.Errorf(
					"Record %s still exists in domain %s",
					rs.Primary.ID, domain,
				)
			}
		}
	}

	return nil
}

func testAccCheckRecordDynamicExists(resource string) resource.TestCheckFunc {
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
		records, err := gonjalla.ListRecords(config.Token, domain)
		if err != nil {
			return fmt.Errorf(
				"Error fetching the records data for domain %s: %s",
				domain, err,
			)
		}

		for _, record := range records {
			if record.ID == rs.Primary.ID {
				if record.Type != "Dynamic" {
					return fmt.Errorf(
						"Record %s has type %s, expected Dynamic",
						rs.Primary.ID, record.Type,
					)
				}
				return nil
			}
		}

		return fmt.Errorf(
			"Record %s doesn't exist for domain %s", rs.Primary.ID, domain,
		)
	}
}

func testAccCheckRecordDynamicCreate() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_dynamic test_create {
  domain = %q
  name = "testacc1-dynamic-create-name"
}
`, domain)
}

func testAccCheckRecordDynamicImport() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_dynamic test_import {
  domain = %q
  name = "testacc2-dynamic-import-name"
}
`, domain)
}
