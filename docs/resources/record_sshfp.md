# njalla_record_sshfp Resource

Njalla `SSHFP` DNS record for a given domain.

## Example Usage

```hcl
resource njalla_record_sshfp example-sshfp {
  domain = "example.com"
  name = "host"
  ttl = 10800
  ssh_algorithm = 4
  ssh_type = 2
  content = "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain this record will be applied to.
* `name` - (Optional) Name for the record. Default is `@`.
* `ttl` - (Required) TTL for the record. Value must be one of
  `60`, `300`, `900`, `3600`, `10800`, `21600` or `86400`.
* `ssh_algorithm` - (Required) SSH key algorithm. Value must be one of `1`
  (RSA), `2` (DSA), `3` (ECDSA), `4` (Ed25519) or `5` (XMSS).
* `ssh_type` - (Required) Fingerprint type. Value must be `1` (SHA-1) or `2`
  (SHA-256).
* `content` - (Required) Hex-encoded fingerprint for the record.

~> **Note** Changing the `domain` attribute forces the existing resource to be
deleted from the previous domain, and created into the new domain.

## Attributes Reference

* `id` - Njalla ID for this record.

## Import

Records are imported by `domain:id`, where `id` is the record's Njalla ID
(see the `njalla_records` data source):

```sh
$ tofu import njalla_record_sshfp.example-sshfp example.com:12345
```
