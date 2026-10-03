# njalla_record_https Resource

Njalla `HTTPS` DNS record for a given domain. Njalla doesn't take TTL or
SvcParams for this type.

## Example Usage

```hcl
resource njalla_record_https example-https {
  domain = "example.com"
  name = "example-name"
  priority = 1
  target = "."
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain this record will be applied to.
* `name` - (Optional) Name for the record. Default is `@`.
* `priority` - (Required) Priority for the record. Value must be one of
  `0`, `1`, `5`, `10`, `20`, `30`, `40`, `50` or `60`.
* `target` - (Required) Target name for the record.

~> **Note** Changing the `domain` attribute forces the existing resource to be
deleted from the previous domain, and created into the new domain.

## Attributes Reference

* `id` - Njalla ID for this record.

## Import

Records are imported by `domain:id`, where `id` is the record's Njalla ID
(see the `njalla_records` data source):

```sh
$ tofu import njalla_record_https.example-https example.com:12345
```
