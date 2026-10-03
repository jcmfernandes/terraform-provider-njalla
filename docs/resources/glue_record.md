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
* `address4` - (Optional) IPv4 address of the nameserver.
* `address6` - (Optional) IPv6 address of the nameserver.

At least one of `address4` and `address6` must be set.

~> **Note** Changing `domain` or `name` forces a new glue record.

## Attributes Reference

* `id` - `domain:name` of the glue record.

## Import

```sh
$ terraform import njalla_glue_record.example example.com:ns1
```
