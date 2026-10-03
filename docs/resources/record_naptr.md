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
* `content` - (Required) Content for the record. Value must follow
  [RFC 2915][]'s syntax from section 2.

~> **Note** Changing the `domain` attribute forces the existing resource to be
deleted from the previous domain, and created into the new domain.

## Attributes Reference

* `id` - Njalla ID for this record.

[RFC 2915]: https://tools.ietf.org/html/rfc2915
