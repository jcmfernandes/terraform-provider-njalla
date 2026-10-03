# njalla_domain Data Source

One of the domains in your Njalla account. Reading a domain that isn't in the
account fails.

## Example Usage

```hcl
data njalla_domain example {
  name = "example.com"
}
```

## Argument Reference

* `name` - (Required) Domain name.

## Attributes Reference

* `id` - The domain name.
* `status` - Status of the domain.
* `expiry` - Expiry date of the domain.
* `mailforwarding` - Whether email forwarding is enabled.
* `dnssec` - Whether DNSSEC is enabled.
* `lock` - Whether the domain is locked against transfers.
* `nameservers` - List of custom nameservers.
* `max_nameservers` - Maximum number of custom nameservers.
