# njalla_vpns Data Source

VPN clients in your Njalla account.

## Example Usage

```hcl
data njalla_vpns all {}
```

## Attributes Reference

* `id` - Always `vpns`.
* `vpns` - List of VPN clients, each with `id`, `name`, `autorenew`,
  `backend`, `publickey` and `expiry`.
