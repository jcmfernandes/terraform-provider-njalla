# njalla_glue_record Resource

Njalla glue record for a nameserver under a given domain.

## Example Usage

```hcl
resource njalla_glue_record example {
  domain = "example.com"
  name = "ns1"
  address4 = "192.0.2.1"
  address6 = "2001:db8::1"
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain this glue record belongs to.
* `name` - (Required) Subdomain of the nameserver.
* `address4` - (Optional) IPv4 address of the nameserver. Must be a valid
  IPv4 address.
* `address6` - (Optional) IPv6 address of the nameserver. Must be a valid
  IPv6 address.

At least one of `address4` and `address6` must be set.

~> **Note** Changing `domain` or `name` deletes the glue record and creates a
new one. The addresses are updated in place.

## Attributes Reference

* `id` - `domain:name` of the glue record.

## Import

Glue records are imported by `domain:name`:

```sh
$ tofu import njalla_glue_record.example example.com:ns1
```
