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
  digest = "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain this record will be applied to.
* `algorithm` - (Required) DNSSEC algorithm number.
* `digest` - (Optional) Digest of the DNSKEY.
* `digest_type` - (Optional) Digest type number.
* `key_tag` - (Optional) Key tag of the DNSKEY.
* `public_key` - (Optional) Public key of the DNSKEY.

Exactly one of `digest` and `public_key` must be set. `digest`, `digest_type`
and `key_tag` must be set together.

~> **Note** Njalla can't edit DNSSEC records, so changing any argument forces
a new record.

## Attributes Reference

* `id` - Njalla ID for this record.

## Import

DNSSEC records are imported by `domain:id`, where `id` is the record's Njalla
ID from the API's `list-dnssec` method:

```sh
$ tofu import njalla_dnssec_record.example example.com:12345
```
