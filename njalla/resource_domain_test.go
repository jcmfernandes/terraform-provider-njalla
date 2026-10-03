package njalla

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Registering a domain costs money and can't be undone: the domain stays in
// the account until it expires.
func TestAccDomain_Register(t *testing.T) {
	name := fmt.Sprintf("testacc-%s.com", acctest.RandString(12))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckPaid(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDomainRegister(name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists("njalla_domain.test_register"),
					resource.TestCheckResourceAttr(
						"njalla_domain.test_register", "name", name,
					),
					resource.TestCheckResourceAttr(
						"njalla_domain.test_register", "status", "active",
					),
					resource.TestCheckResourceAttr(
						"njalla_domain.test_register",
						"mailforwarding",
						"false",
					),
					resource.TestCheckResourceAttrSet(
						"njalla_domain.test_register", "expiry",
					),
				),
			},
			{
				Config: testAccCheckDomainRegister(name, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"njalla_domain.test_register",
						"mailforwarding",
						"true",
					),
				),
			},
			{
				ResourceName:            "njalla_domain.test_register",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"years"},
			},
		},
	})
}

// Importing the test domain doesn't change it, and as the import isn't
// persisted, nothing gets destroyed afterwards.
func TestAccDomain_Import(t *testing.T) {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        testAccCheckDomainImport(),
				ResourceName:  "njalla_domain.test_import",
				ImportState:   true,
				ImportStateId: domain,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("Expected 1 state, got %d", len(states))
					}
					if states[0].Attributes["name"] != domain {
						return fmt.Errorf(
							"Imported name %s, expected %s",
							states[0].Attributes["name"], domain,
						)
					}
					return nil
				},
			},
		},
	})
}

func testAccCheckDomainExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No domain ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		domains, err := listDomains(context.Background(), config.Client)
		if err != nil {
			return fmt.Errorf("Error fetching the domains: %s", err)
		}

		for _, domain := range domains {
			if domain.Name == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("Domain %s doesn't exist", rs.Primary.ID)
	}
}

func testAccCheckDomainRegister(name string, mailforwarding bool) string {
	return fmt.Sprintf(`
resource njalla_domain test_register {
  name = %q
  mailforwarding = %t
}
`, name, mailforwarding)
}

func testAccCheckDomainImport() string {
	domain := os.Getenv("NJALLA_TESTACC_DOMAIN")
	return fmt.Sprintf(`
resource njalla_domain test_import {
  name = %q
}
`, domain)
}
