package njalla

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// VPNs cost money, so a single test covers create, update and import.
func TestAccVPN_Lifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckPaid(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckVPNDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVPN("testacc1-vpn-name1", "wireguard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNExists("njalla_vpn.test"),
					resource.TestCheckResourceAttr(
						"njalla_vpn.test", "name", "testacc1-vpn-name1",
					),
					resource.TestCheckResourceAttr(
						"njalla_vpn.test", "backend", "wireguard",
					),
					resource.TestCheckResourceAttr(
						"njalla_vpn.test", "autorenew", "false",
					),
				),
			},
			{
				Config: testAccCheckVPN("testacc1-vpn-name2", "openvpn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNExists("njalla_vpn.test"),
					resource.TestCheckResourceAttr(
						"njalla_vpn.test", "name", "testacc1-vpn-name2",
					),
					resource.TestCheckResourceAttr(
						"njalla_vpn.test", "backend", "openvpn",
					),
				),
			},
			{
				ResourceName:      "njalla_vpn.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPNDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_vpn" {
			continue
		}

		vpns, err := listVPNs(context.Background(), config.Client)
		if err != nil {
			return fmt.Errorf("Error fetching the VPNs: %s", err)
		}

		for _, vpn := range vpns {
			if vpn.ID == rs.Primary.ID {
				return fmt.Errorf("VPN %s still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckVPNExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPN ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		vpns, err := listVPNs(context.Background(), config.Client)
		if err != nil {
			return fmt.Errorf("Error fetching the VPNs: %s", err)
		}

		for _, vpn := range vpns {
			if vpn.ID == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("VPN %s doesn't exist", rs.Primary.ID)
	}
}

func testAccCheckVPN(name string, backend string) string {
	return fmt.Sprintf(`
resource njalla_vpn test {
  name = %q
  backend = %q
}
`, name, backend)
}
