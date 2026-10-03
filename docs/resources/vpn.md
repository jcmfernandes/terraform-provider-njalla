# njalla_vpn Resource

Njalla VPN client.

!> **Warning** Creating this resource buys a VPN, which charges your Njalla
wallet.

## Example Usage

```hcl
resource njalla_vpn example {
  name = "laptop"
  autorenew = true
  backend = "wireguard"
  publickey = "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg="
}
```

## Argument Reference

* `name` - (Required) Name of the VPN client.
* `autorenew` - (Optional) Whether to renew the VPN automatically. Default is
  `false`.
* `backend` - (Optional) VPN backend, `wireguard` or `openvpn`. Njalla's
  default if unset.
* `publickey` - (Optional) WireGuard public key of the client. Njalla
  generates one if unset.

-> **Note** Manual renewals aren't managed. Use the Njalla web interface or
API.

## Attributes Reference

* `id` - Njalla ID for this VPN.
* `expiry` - Expiry date of the VPN.

## Import

```sh
$ terraform import njalla_vpn.example <id>
```
