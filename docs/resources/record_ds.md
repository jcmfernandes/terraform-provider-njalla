# njalla_record_ds Resource

Njalla `DS` DNS record for a given domain.

## Example Usage

```hcl
resource njalla_record_ds example-ds {
  domain = "example.com"
  name = "sub"
  ttl = 10800
  content = "12345 13 2 1f2a...e9"
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain this record will be applied to.
* `name` - (Optional) Name for the record. Default is `@`.
* `ttl` - (Required) TTL for the record. Value must be one of
  [gonjalla's `ValidTTL`][gonjalla variable ValidTTL].
* `content` - (Required) DS record data: key tag, algorithm, digest type and
  digest, space separated.

~> **Note** Changing the `domain` attribute forces the existing resource to be
deleted from the previous domain, and created into the new domain.

## Attributes Reference

* `id` - Njalla ID for this record.

[gonjalla variable ValidTTL]: https://pkg.go.dev/github.com/Sighery/gonjalla?tab=doc#pkg-variables
