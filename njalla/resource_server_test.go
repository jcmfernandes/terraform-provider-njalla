package njalla

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Servers cost money, so a single test covers create, update and import.
func TestAccServer_Lifecycle(t *testing.T) {
	publicKey, _, err := acctest.RandSSHKeyPair("testacc")
	if err != nil {
		t.Fatalf("%q", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckPaid(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServer("testacc1-server-name1", publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("njalla_server.test"),
					resource.TestCheckResourceAttr(
						"njalla_server.test", "name", "testacc1-server-name1",
					),
					resource.TestCheckResourceAttr(
						"njalla_server.test", "ssh_key", publicKey,
					),
					resource.TestCheckResourceAttr(
						"njalla_server.test", "autorenew", "false",
					),
				),
			},
			{
				Config: testAccCheckServer("testacc1-server-name2", publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("njalla_server.test"),
					resource.TestCheckResourceAttr(
						"njalla_server.test", "name", "testacc1-server-name2",
					),
				),
			},
			{
				ResourceName:            "njalla_server.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"months"},
			},
		},
	})
}

func testAccCheckServerDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_server" {
			continue
		}

		exists, err := serverExists(
			context.Background(), config.Client, rs.Primary.ID,
		)
		if err != nil {
			return fmt.Errorf("Error fetching the servers: %s", err)
		}
		if exists {
			return fmt.Errorf("Server %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckServerExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No server ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		exists, err := serverExists(
			context.Background(), config.Client, rs.Primary.ID,
		)
		if err != nil {
			return fmt.Errorf("Error fetching the servers: %s", err)
		}
		if !exists {
			return fmt.Errorf("Server %s doesn't exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServer(name string, publicKey string) string {
	return fmt.Sprintf(`
data njalla_server_types all {}
data njalla_server_images all {}

resource njalla_server test {
  name = %q
  type = data.njalla_server_types.all.types[0]
  os = data.njalla_server_images.all.images[0]
  ssh_key = %q
  months = 1
}
`, name, publicKey)
}
