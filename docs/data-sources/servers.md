# njalla_servers Data Source

Servers in your Njalla account.

## Example Usage

```hcl
data njalla_servers all {}
```

## Attributes Reference

* `id` - Always `servers`.
* `servers` - List of servers, each with `id`, `name`, `type`, `os`,
  `status`, `os_state`, `expiry`, `autorenew`, `ssh_key`, `reverse_name` and
  `ips`.
