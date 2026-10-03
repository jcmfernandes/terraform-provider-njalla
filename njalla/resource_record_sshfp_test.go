package njalla

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRecordSSHFP_Create(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSSHFPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordSSHFPCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSSHFPExists("njalla_record_sshfp.test_create"),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_create", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_create", "name", "testacc1-sshfp-create-name",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_create", "ttl", "10800",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_create", "ssh_algorithm", "4",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_create", "ssh_type", "2",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_create", "content", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					),
				),
			},
		},
	})
}

func TestAccRecordSSHFP_Update(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSSHFPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordSSHFPUpdatePre(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSSHFPExists("njalla_record_sshfp.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "name", "testacc2-sshfp-update-name1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "ttl", "10800",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "ssh_algorithm", "4",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "ssh_type", "2",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "content", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					),
				),
			},
			{
				Config: testAccCheckRecordSSHFPUpdatePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSSHFPExists("njalla_record_sshfp.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "name", "testacc2-sshfp-update-name2",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "ttl", "3600",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "ssh_algorithm", "1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "ssh_type", "1",
					),
					resource.TestCheckResourceAttr(
						"njalla_record_sshfp.test_update", "content", "cccccccccccccccccccccccccccccccccccccccc",
					),
				),
			},
		},
	})
}

func TestAccRecordSSHFP_Import(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSSHFPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRecordSSHFPImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecordSSHFPExists("njalla_record_sshfp.test_import"),
				),
			},
			{
				ResourceName:        "njalla_record_sshfp.test_import",
				ImportStateIdPrefix: fmt.Sprintf("%s:", domain),
				ImportState:         true,
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccRecordSSHFP_InvalidSSHAlgorithm(t *testing.T) {
	expectedErr := regexp.MustCompile("expected ssh_algorithm to be in the range \\(1 - 5\\), got 6")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSSHFPDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckRecordSSHFPInvalidSSHAlgorithm(),
				ExpectError: expectedErr,
			},
		},
	})
}

func TestAccRecordSSHFP_InvalidSSHType(t *testing.T) {
	expectedErr := regexp.MustCompile("expected ssh_type to be in the range \\(1 - 2\\), got 3")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRecordSSHFPDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckRecordSSHFPInvalidSSHType(),
				ExpectError: expectedErr,
			},
		},
	})
}

func testAccCheckRecordSSHFPDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_record_sshfp" {
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

func testAccCheckRecordSSHFPExists(resource string) resource.TestCheckFunc {
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

		if record.Type != "SSHFP" {
			return fmt.Errorf(
				"Record %s has type %s, expected SSHFP",
				rs.Primary.ID, record.Type,
			)
		}

		return nil
	}
}

func testAccCheckRecordSSHFPCreate() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_sshfp test_create {
  domain = %q
  name = "testacc1-sshfp-create-name"
  ttl = 10800
  ssh_algorithm = 4
  ssh_type = 2
  content = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
`, domain)
}

func testAccCheckRecordSSHFPUpdatePre() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_sshfp test_update {
  domain = %q
  name = "testacc2-sshfp-update-name1"
  ttl = 10800
  ssh_algorithm = 4
  ssh_type = 2
  content = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
`, domain)
}

func testAccCheckRecordSSHFPUpdatePost() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_sshfp test_update {
  domain = %q
  name = "testacc2-sshfp-update-name2"
  ttl = 3600
  ssh_algorithm = 1
  ssh_type = 1
  content = "cccccccccccccccccccccccccccccccccccccccc"
}
`, domain)
}

func testAccCheckRecordSSHFPImport() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_sshfp test_import {
  domain = %q
  name = "testacc3-sshfp-import-name"
  ttl = 10800
  ssh_algorithm = 4
  ssh_type = 2
  content = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
}
`, domain)
}

func testAccCheckRecordSSHFPInvalidSSHAlgorithm() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_sshfp test_invalid_sshalgorithm {
  domain = %q
  name = "testacc4-sshfp-invalidsshalgorithm-name"
  ttl = 10800
  ssh_algorithm = 6
  ssh_type = 2
  content = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
`, domain)
}

func testAccCheckRecordSSHFPInvalidSSHType() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_record_sshfp test_invalid_sshtype {
  domain = %q
  name = "testacc5-sshfp-invalidsshtype-name"
  ttl = 10800
  ssh_algorithm = 4
  ssh_type = 3
  content = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
`, domain)
}
