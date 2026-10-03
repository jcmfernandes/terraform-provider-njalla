package njalla

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccEmailForward_Create(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckEmailForwardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEmailForwardCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailForwardExists(
						"njalla_email_forward.test_create",
					),
					resource.TestCheckResourceAttr(
						"njalla_email_forward.test_create", "domain", domain,
					),
					resource.TestCheckResourceAttr(
						"njalla_email_forward.test_create",
						"from",
						"testacc1-forward-create",
					),
					resource.TestCheckResourceAttr(
						"njalla_email_forward.test_create",
						"to",
						"testacc1@example.com",
					),
				),
			},
		},
	})
}

func TestAccEmailForward_Import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckEmailForwardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEmailForwardImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailForwardExists(
						"njalla_email_forward.test_import",
					),
				),
			},
			{
				ResourceName:      "njalla_email_forward.test_import",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestParseEmailForwardIDExpected(t *testing.T) {
	domain, from, to, err := parseEmailForwardID(
		"testing.com:me:me@example.com",
	)
	if err != nil {
		t.Fatalf("%q", err)
	}

	if domain != "testing.com" || from != "me" || to != "me@example.com" {
		t.Fatalf("Unexpected parse result %s, %s, %s", domain, from, to)
	}
}

func TestParseEmailForwardIDMissingPart(t *testing.T) {
	for _, id := range []string{"testing.com:me", "testing.com::x", "a:b:c:d"} {
		if _, _, _, err := parseEmailForwardID(id); err == nil {
			t.Fatalf("Unexpected success for %s", id)
		}
	}
}

func testAccCheckEmailForwardDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_email_forward" {
			continue
		}

		exists, err := testAccEmailForwardExists(config, rs)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Email forward %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckEmailForwardExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No email forward ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		exists, err := testAccEmailForwardExists(config, rs)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Email forward %s doesn't exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEmailForwardExists(
	config *Config, rs *terraform.ResourceState,
) (bool, error) {
	domain := rs.Primary.Attributes["domain"]
	forwards, err := listForwards(context.Background(), config.Client, domain)
	if err != nil {
		return false, fmt.Errorf(
			"Error fetching the email forwards for domain %s: %s",
			domain, err,
		)
	}

	for _, forward := range forwards {
		if forward.From == rs.Primary.Attributes["from"] &&
			forward.To == rs.Primary.Attributes["to"] {
			return true, nil
		}
	}

	return false, nil
}

func testAccCheckEmailForwardCreate() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_email_forward test_create {
  domain = %q
  from = "testacc1-forward-create"
  to = "testacc1@example.com"
}
`, domain)
}

func testAccCheckEmailForwardImport() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_email_forward test_import {
  domain = %q
  from = "testacc2-forward-import"
  to = "testacc2@example.com"
}
`, domain)
}
