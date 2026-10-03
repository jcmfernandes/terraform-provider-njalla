# njalla_server Resource

Njalla server.

!> **Warning** Creating this resource buys a server, which charges your Njalla
wallet. Destroying it stops the server and deletes all its data.

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
* `months` - (Optional) Months to pay for upfront, between 1 and 12. Default is
  `1`. Only used when buying; changes afterwards are ignored.
* `autorenew` - (Optional) Whether to renew the server automatically. Default
  is `false`.
* `reverse_name` - (Optional) Reverse DNS name of the server's addresses.

~> **Note** Changing `os` replaces the server: the old one is removed with all
its data, and a new one is bought.

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

```sh
$ terraform import njalla_server.example <id>
```
