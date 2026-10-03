package njalla

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProvider *schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	testAccProvider = Provider()
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"njalla": func() (*schema.Provider, error) { return testAccProvider, nil },
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("NJALLA_API_TOKEN"); v == "" {
		t.Fatal("NJALLA_API_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("NJALLA_TESTACC_DOMAIN"); v == "" {
		t.Fatal("NJALLA_TESTACC_DOMAIN must be set for acceptance tests")
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
