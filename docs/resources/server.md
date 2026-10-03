# njalla_server Resource

Njalla server.

!> **Warning** Creating this resource buys a server, which charges your Njalla
wallet. Destroying it removes the server and all its data.

## Example Usage

```hcl
data njalla_server_types all {}
data njalla_server_images all {}

resource njalla_server example {
  name = "example"
  type = data.njalla_server_types.all.types[0]
  os = data.njalla_server_images.all.images[0]
  ssh_key = file("~/.ssh/id_ed25519.pub")
  months = 1
  autorenew = true
}
```

## Argument Reference

* `name` - (Required) Name of the server.
* `type` - (Required) Server type, from the `njalla_server_types` data source.
* `os` - (Required) Server image, from the `njalla_server_images` data source.
* `ssh_key` - (Required) Public SSH key installed on the server.
* `months` - (Optional) Months to pay for upfront, between `1` and `12`.
  Default is `1`. Only used when buying; changes afterwards are ignored.
* `autorenew` - (Optional) Whether to renew the server automatically. Default
  is `false`.
* `reverse_name` - (Optional) Reverse DNS name of the server's addresses.
  Njalla's value is kept if unset.

~> **Note** Changing `os` replaces the server: the old one is removed with all
its data, and a new one is bought, charging the wallet again. Changing
`type` resizes the server in place. Other arguments are updated in place.

~> **Note** `reverse_name` is set after the server is bought. If that fails,
the server is kept in the state with a warning, and the next apply retries
it.

-> **Note** Starting, stopping, restarting, manual renewals and extra traffic
packages aren't managed. Use the Njalla web interface or API.

## Attributes Reference

* `id` - Njalla ID for this server.
* `status` - Status of the server.
* `os_state` - State of the server's OS installation.
* `expiry` - Expiry date of the server.
* `ips` - IP addresses of the server.

## Timeouts

* `delete` - (Default `10m`) How long to wait for the server to be removed.

## Import

Servers are imported by Njalla ID (see the `njalla_servers` data source).
`months` isn't read back.

```sh
$ tofu import njalla_server.example <id>
```
