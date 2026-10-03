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

// testAccPreCheckPaid skips tests that spend money unless
// NJALLA_TESTACC_PAID is set.
func testAccPreCheckPaid(t *testing.T) {
	if v := os.Getenv("NJALLA_TESTACC_PAID"); v == "" {
		t.Skip("NJALLA_TESTACC_PAID must be set for tests that cost money")
	}
}

// testAccPreCheckDNSSEC skips tests that publish DNSSEC records for the test
// domain unless NJALLA_TESTACC_DNSSEC is set. A made-up DS record left
// behind by a failed run breaks resolution of the domain.
func testAccPreCheckDNSSEC(t *testing.T) {
	if v := os.Getenv("NJALLA_TESTACC_DNSSEC"); v == "" {
		t.Skip("NJALLA_TESTACC_DNSSEC must be set for DNSSEC record tests")
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
