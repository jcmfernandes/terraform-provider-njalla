# njalla_record_txt Resource

Njalla `TXT` DNS record for a given domain.

## Example Usage

```hcl
resource njalla_record_txt example-txt {
  domain = "example.com"
  name = "example-name"
  ttl = 10800
  content = "example-content"
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain this record will be applied to.
* `name` - (Optional) Name for the record. Default is `@`.
* `ttl` - (Required) TTL for the record. Value must be one of
  `60`, `300`, `900`, `3600`, `10800`, `21600` or `86400`.
* `content` - (Required) Content for the record.

~> **Note** Changing the `domain` attribute forces the existing resource to be
deleted from the previous domain, and created into the new domain.

## Attributes Reference

* `id` - Njalla ID for this record.
