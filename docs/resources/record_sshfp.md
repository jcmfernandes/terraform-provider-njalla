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
  content = "6c3a...b0"
}
```

## Argument Reference

* `domain` - (Required) Specifies the domain this record will be applied to.
* `name` - (Optional) Name for the record. Default is `@`.
* `ttl` - (Required) TTL for the record. Value must be one of
  `60`, `300`, `900`, `3600`, `10800`, `21600` or `86400`.
* `ssh_algorithm` - (Required) SSH key algorithm: `1` RSA, `2` DSA, `3` ECDSA,
  `4` Ed25519, `5` XMSS.
* `ssh_type` - (Required) Fingerprint type: `1` SHA-1, `2` SHA-256.
* `content` - (Required) Hex-encoded fingerprint for the record.

~> **Note** Changing the `domain` attribute forces the existing resource to be
deleted from the previous domain, and created into the new domain.

## Attributes Reference

* `id` - Njalla ID for this record.
