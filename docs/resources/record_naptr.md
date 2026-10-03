# njalla_record_naptr Resource

Njalla `NAPTR` DNS record for a given domain.

## Example Usage

```hcl
resource njalla_record_naptr example-naptr {
  domain = "example.com"
  name = "@"
  ttl = 10800
  content = "100 10 \"S\" \"SIP+D2U\" \"!^.*$!sip:customer-service@example.com!\" _sip._udp.example.com."
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain this record will be applied to.
* `name` - (Optional) Name for the record. Default is `@`.
* `ttl` - (Required) TTL for the record. Value must be one of
  `60`, `300`, `900`, `3600`, `10800`, `21600` or `86400`.
* `content` - (Required) Content for the record, as at least six fields
  separated by single spaces: order, preference, flags, service, regexp and
  replacement ([RFC 2915][] section 2). Order and preference must be
  integers.

~> **Note** Changing the `domain` attribute forces the existing resource to be
deleted from the previous domain, and created into the new domain.

## Attributes Reference

* `id` - Njalla ID for this record.

## Import

Records are imported by `domain:id`, where `id` is the record's Njalla ID
(see the `njalla_records` data source):

```sh
$ tofu import njalla_record_naptr.example-naptr example.com:12345
```

[RFC 2915]: https://tools.ietf.org/html/rfc2915
