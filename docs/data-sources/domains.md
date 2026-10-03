# njalla_domains Data Source

Domains in your Njalla account.

## Example Usage

```hcl
data njalla_domains all {}
```

## Attributes Reference

* `id` - Always `domains`.
* `domains` - List of domains, each with:
  * `name` - Domain name.
  * `status` - Status of the domain.
  * `expiry` - Expiry date of the domain.
