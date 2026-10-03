# njalla_dnssec_record Resource

Njalla DNSSEC (DS) record published to the registry for a given domain. Give
either `digest`, `digest_type` and `key_tag`, or `public_key`.

## Example Usage

```hcl
resource njalla_dnssec_record example {
  domain = "example.com"
  algorithm = 13
  digest_type = 2
  key_tag = 12345
  digest = "1f2a...e9"
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain this record will be applied to.
* `algorithm` - (Required) DNSSEC algorithm number.
* `digest` - (Optional) Digest of the DNSKEY. Conflicts with `public_key`.
* `digest_type` - (Optional) Digest type number. Required with `digest`.
* `key_tag` - (Optional) Key tag of the DNSKEY. Required with `digest`.
* `public_key` - (Optional) Public key of the DNSKEY. Conflicts with `digest`.

~> **Note** Njalla can't edit DNSSEC records, so changing any argument forces
a new record.

## Attributes Reference

* `id` - Njalla ID for this record.

## Import

```sh
$ terraform import njalla_dnssec_record.example example.com:12345
```
