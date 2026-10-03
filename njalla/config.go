package njalla

import "github.com/Sighery/terraform-provider-njalla/internal/client"

// Config is the metadata interface provider passed later on to resources
type Config struct {
	Client *client.Client
}
