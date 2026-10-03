# njalla_tlds Data Source

TLDs Njalla can register.

## Example Usage

```hcl
data njalla_tlds all {}
```

## Attributes Reference

* `tlds` - List of TLDs sorted by name, each with:
  * `name` - The TLD, e.g. `com`.
  * `price` - Yearly price.
  * `max_year` - Maximum years it can be registered for.
  * `dnssec` - Whether it supports DNSSEC.
