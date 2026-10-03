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

func TestAccGlueRecord_Create(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckGlueRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckGlueRecordCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueRecordExists(
						"njalla_glue_record.test_create",
					),
					resource.TestCheckResourceAttr(
						"njalla_glue_record.test_create", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_glue_record.test_create",
						"name",
						"testacc1-glue-create-name",
					),
					resource.TestCheckResourceAttr(
						"njalla_glue_record.test_create",
						"address4",
						"1.1.1.1",
					),
					resource.TestCheckResourceAttr(
						"njalla_glue_record.test_create",
						"address6",
						"2001:db8::1",
					),
				),
			},
		},
	})
}

func TestAccGlueRecord_Update(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckGlueRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckGlueRecordUpdatePre(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueRecordExists(
						"njalla_glue_record.test_update",
					),
					resource.TestCheckResourceAttr(
						"njalla_glue_record.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_glue_record.test_update",
						"address4",
						"1.1.1.2",
					),
					resource.TestCheckResourceAttr(
						"njalla_glue_record.test_update", "address6", "",
					),
				),
			},
			{
				Config: testAccCheckGlueRecordUpdatePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueRecordExists(
						"njalla_glue_record.test_update",
					),
					resource.TestCheckResourceAttr(
						"njalla_glue_record.test_update",
						"address4",
						"1.1.1.3",
					),
					resource.TestCheckResourceAttr(
						"njalla_glue_record.test_update",
						"address6",
						"2001:db8::3",
					),
				),
			},
		},
	})
}

func TestAccGlueRecord_Import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckGlueRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckGlueRecordImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueRecordExists(
						"njalla_glue_record.test_import",
					),
				),
			},
			{
				ResourceName:      "njalla_glue_record.test_import",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGlueRecord_NoAddress(t *testing.T) {
	expectedErr := regexp.MustCompile(
		"one of `address4,address6` must be specified",
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckGlueRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckGlueRecordNoAddress(),
				ExpectError: expectedErr,
			},
		},
	})
}

func testAccCheckGlueRecordDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_glue_record" {
			continue
		}

		domain := rs.Primary.Attributes["domain"]
		glue, err := listGlue(context.Background(), config.Client, domain)
		if err != nil {
			return fmt.Errorf(
				"Error fetching the glue records for domain %s: %s",
				domain, err,
			)
		}

		for _, record := range glue {
			if record.Name == rs.Primary.Attributes["name"] {
				return fmt.Errorf(
					"Glue record %s still exists", rs.Primary.ID,
				)
			}
		}
	}

	return nil
}

func testAccCheckGlueRecordExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No glue record ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		domain := rs.Primary.Attributes["domain"]
		glue, err := listGlue(context.Background(), config.Client, domain)
		if err != nil {
			return fmt.Errorf(
				"Error fetching the glue records for domain %s: %s",
				domain, err,
			)
		}

		for _, record := range glue {
			if record.Name == rs.Primary.Attributes["name"] {
				return nil
			}
		}

		return fmt.Errorf("Glue record %s doesn't exist", rs.Primary.ID)
	}
}

func testAccCheckGlueRecordCreate() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_glue_record test_create {
  domain = %q
  name = "testacc1-glue-create-name"
  address4 = "1.1.1.1"
  address6 = "2001:db8::1"
}
`, domain)
}

func testAccCheckGlueRecordUpdatePre() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_glue_record test_update {
  domain = %q
  name = "testacc2-glue-update-name"
  address4 = "1.1.1.2"
}
`, domain)
}

func testAccCheckGlueRecordUpdatePost() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_glue_record test_update {
  domain = %q
  name = "testacc2-glue-update-name"
  address4 = "1.1.1.3"
  address6 = "2001:db8::3"
}
`, domain)
}

func testAccCheckGlueRecordImport() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_glue_record test_import {
  domain = %q
  name = "testacc3-glue-import-name"
  address4 = "1.1.1.4"
}
`, domain)
}

func testAccCheckGlueRecordNoAddress() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_glue_record test_no_address {
  domain = %q
  name = "testacc4-glue-noaddress-name"
}
`, domain)
}
