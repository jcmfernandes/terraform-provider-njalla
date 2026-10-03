package njalla

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAPIToken_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAPITokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAPITokenCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPITokenExists("njalla_api_token.test_create"),
					resource.TestCheckResourceAttrSet(
						"njalla_api_token.test_create", "key",
					),
					resource.TestCheckResourceAttr(
						"njalla_api_token.test_create",
						"comment",
						"testacc1-token-create",
					),
					resource.TestCheckResourceAttr(
						"njalla_api_token.test_create", "from.#", "1",
					),
					resource.TestCheckTypeSetElemAttr(
						"njalla_api_token.test_create", "from.*", "192.0.2.0/24",
					),
				),
			},
		},
	})
}

func TestAccAPIToken_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAPITokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAPITokenUpdatePre(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPITokenExists("njalla_api_token.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_api_token.test_update",
						"comment",
						"testacc2-token-update1",
					),
					resource.TestCheckResourceAttr(
						"njalla_api_token.test_update", "allowed_types.#", "1",
					),
				),
			},
			{
				Config: testAccCheckAPITokenUpdatePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPITokenExists("njalla_api_token.test_update"),
					resource.TestCheckResourceAttr(
						"njalla_api_token.test_update",
						"comment",
						"testacc2-token-update2",
					),
					resource.TestCheckResourceAttr(
						"njalla_api_token.test_update", "allowed_types.#", "2",
					),
					resource.TestCheckResourceAttr(
						"njalla_api_token.test_update", "from.#", "0",
					),
				),
			},
		},
	})
}

func TestAccAPIToken_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAPITokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAPITokenImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPITokenExists("njalla_api_token.test_import"),
				),
			},
			{
				ResourceName: "njalla_api_token.test_import",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["njalla_api_token.test_import"]
					return rs.Primary.Attributes["key"], nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestNewTokenKey(t *testing.T) {
	before := []apiToken{{Key: "a"}, {Key: "b"}}

	key, err := newTokenKey(before, append(before, apiToken{Key: "c"}))
	if err != nil {
		t.Fatalf("%q", err)
	}
	if key != "c" {
		t.Fatalf("Got key %s, expected c", key)
	}

	if _, err := newTokenKey(before, before); err == nil {
		t.Fatal("Unexpected success with no new token")
	}

	after := append(before, apiToken{Key: "c"}, apiToken{Key: "d"})
	if _, err := newTokenKey(before, after); err == nil {
		t.Fatal("Unexpected success with two new tokens")
	}
}

func testAccCheckAPITokenDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "njalla_api_token" {
			continue
		}

		tokens, err := listTokens(context.Background(), config.Client)
		if err != nil {
			return fmt.Errorf("Error fetching the API tokens: %s", err)
		}

		for _, token := range tokens {
			if token.Key == rs.Primary.Attributes["key"] {
				return fmt.Errorf("API token %s still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAPITokenExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No API token ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		tokens, err := listTokens(context.Background(), config.Client)
		if err != nil {
			return fmt.Errorf("Error fetching the API tokens: %s", err)
		}

		for _, token := range tokens {
			if token.Key == rs.Primary.Attributes["key"] {
				return nil
			}
		}

		return fmt.Errorf("API token %s doesn't exist", rs.Primary.ID)
	}
}

func testAccCheckAPITokenCreate() string {
	return `
resource njalla_api_token test_create {
  comment = "testacc1-token-create"
  from = ["192.0.2.0/24"]
}
`
}

func testAccCheckAPITokenUpdatePre() string {
	return `
resource njalla_api_token test_update {
  comment = "testacc2-token-update1"
  from = ["192.0.2.0/24"]
  allowed_types = ["TXT"]
}
`
}

func testAccCheckAPITokenUpdatePost() string {
	return `
resource njalla_api_token test_update {
  comment = "testacc2-token-update2"
  allowed_types = ["TXT", "A"]
}
`
}

func testAccCheckAPITokenImport() string {
	return `
resource njalla_api_token test_import {
  comment = "testacc3-token-import"
}
`
}
