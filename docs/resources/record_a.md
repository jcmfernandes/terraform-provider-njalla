# njalla_record_a Resource

Njalla `A` DNS record for a given domain.

## Example Usage

```hcl
resource njalla_record_a example-a {
  domain = "example.com"
  name = "example-name"
  ttl = 10800
  content = "138.201.81.199"
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain this record will be applied to.
* `name` - (Optional) Name for the record. Default is `@`.
* `ttl` - (Required) TTL for the record. Value must be one of
  `60`, `300`, `900`, `3600`, `10800`, `21600` or `86400`.
* `content` - (Required) IPv4 address for the record. Must be a valid IPv4
  address.

~> **Note** Changing the `domain` attribute forces the existing resource to be
deleted from the previous domain, and created into the new domain.

## Attributes Reference

* `id` - Njalla ID for this record.

## Import

Records are imported by `domain:id`, where `id` is the record's Njalla ID
(see the `njalla_records` data source):

```sh
$ tofu import njalla_record_a.example-a example.com:12345
```
