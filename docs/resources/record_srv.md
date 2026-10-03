# njalla_record_srv Resource

Njalla `SRV` DNS record for a given domain.

## Example Usage

```hcl
resource njalla_record_srv example-srv {
  domain = "example.com"
  name = "_sip._tcp"
  ttl = 10800
  priority = 10
  weight = 5
  port = 5060
  content = "sip.example.com"
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain this record will be applied to.
* `name` - (Optional) Name for the record. Default is `@`.
* `ttl` - (Required) TTL for the record. Value must be one of
  `60`, `300`, `900`, `3600`, `10800`, `21600` or `86400`.
* `priority` - (Required) Priority for the record. Value must be one of
  `0`, `1`, `5`, `10`, `20`, `30`, `40`, `50` or `60`.
* `weight` - (Required) Weight for the record, between 0 and 65535.
* `port` - (Required) Port of the service, between 0 and 65535.
* `content` - (Required) Target host of the service.

~> **Note** Changing the `domain` attribute forces the existing resource to be
deleted from the previous domain, and created into the new domain.

## Attributes Reference

* `id` - Njalla ID for this record.
