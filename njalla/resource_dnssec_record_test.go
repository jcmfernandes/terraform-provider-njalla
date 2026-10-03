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

func TestAccDNSSECRecord_Create(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckDNSSEC(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckDNSSECRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDNSSECRecordCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSECRecordExists(
						"njalla_dnssec_record.test_create",
					),
					resource.TestCheckResourceAttr(
						"njalla_dnssec_record.test_create", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_dnssec_record.test_create", "algorithm", "13",
					),
					resource.TestCheckResourceAttr(
						"njalla_dnssec_record.test_create", "digest_type", "2",
					),
					resource.TestCheckResourceAttr(
						"njalla_dnssec_record.test_create", "key_tag", "11111",
					),
					resource.TestCheckResourceAttr(
						"njalla_dnssec_record.test_create",
						"digest",
						"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					),
				),
			},
		},
	})
}

func TestAccDNSSECRecord_Import(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckDNSSEC(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckDNSSECRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDNSSECRecordImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSECRecordExists(
						"njalla_dnssec_record.test_import",
					),
				),
			},
			{
				ResourceName:        "njalla_dnssec_record.test_import",
				ImportStateIdPrefix: fmt.Sprintf("%s:", domain),
				ImportState:         true,
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccDNSSECRecord_DigestAndPublicKey(t *testing.T) {
	expectedErr := regexp.MustCompile(
		"only one of `digest,public_key` can be specified",
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckDNSSECRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDNSSECRecordDigestAndPublicKey(),
				ExpectError: expectedErr,
			},
		},
	})
}

func TestNewDNSSECID(t *testing.T) {
	before := []dnssecRecord{{ID: "1"}, {ID: "2"}}

	id, err := newDNSSECID(before, append(before, dnssecRecord{ID: "3"}))
	if err != nil {
		t.Fatalf("%q", err)
	}
	if id != "3" {
		t.Fatalf("Got ID %s, expected 3", id)
	}

	if _, err := newDNSSECID(before, before); err == nil {
		t.Fatal("Unexpected success with no new record")
	}

	after := append(before, dnssecRecord{ID: "3"}, dnssecRecord{ID: "4"})
	if _, err := newDNSSECID(before, after); err == nil {
		t.Fatal("Unexpected success with two new records")
	}
}

func testAccCheckDNSSECRecordDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_dnssec_record" {
			continue
		}

		records, err := listDNSSEC(context.Background(), config.Client, domain)
		if err != nil {
			return fmt.Errorf(
				"Error fetching the DNSSEC records for domain %s: %s",
				domain, err,
			)
		}

		for _, record := range records {
			if record.ID == rs.Primary.ID {
				return fmt.Errorf(
					"DNSSEC record %s still exists in domain %s",
					rs.Primary.ID, domain,
				)
			}
		}
	}

	return nil
}

func testAccCheckDNSSECRecordExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No DNSSEC record ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
		records, err := listDNSSEC(context.Background(), config.Client, domain)
		if err != nil {
			return fmt.Errorf(
				"Error fetching the DNSSEC records for domain %s: %s",
				domain, err,
			)
		}

		for _, record := range records {
			if record.ID == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf(
			"DNSSEC record %s doesn't exist for domain %s",
			rs.Primary.ID, domain,
		)
	}
}

func testAccCheckDNSSECRecordCreate() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_dnssec_record test_create {
  domain = %q
  algorithm = 13
  digest_type = 2
  key_tag = 11111
  digest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
`, domain)
}

func testAccCheckDNSSECRecordImport() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_dnssec_record test_import {
  domain = %q
  algorithm = 13
  digest_type = 2
  key_tag = 22222
  digest = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
}
`, domain)
}

func testAccCheckDNSSECRecordDigestAndPublicKey() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_dnssec_record test_digest_and_public_key {
  domain = %q
  algorithm = 13
  digest_type = 2
  key_tag = 33333
  digest = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
  public_key = "testacc3-dnssec-public-key"
}
`, domain)
}
