# njalla_records Data Source

DNS records of a given domain.

## Example Usage

```hcl
data njalla_records example {
  domain = "example.com"
}
```

## Argument Reference

* `domain` - (Required) Domain to list the records of.

## Attributes Reference

* `id` - The domain name.
* `records` - List of records, each with `id`, `name`, `type`, `content`,
  `ttl`, `priority`, `weight`, `port`, `target`, `ssh_algorithm` and
  `ssh_type`. Fields a record type doesn't use are empty or `0`.

~> **Note** A record Njalla returns in an unexpected shape is left out of
`records`, with a warning naming its ID.
