# TelDrive v2 backend notes

This branch updates the TelDrive rclone backend for the v2 `/api/v1` contract.
It also preserves path prefixes in `api_host`, so a Caddy endpoint such as
`https://example.com/tgd` works when Caddy strips `/tgd` before proxying.

Example remote:

```ini
[teldrive]
type = teldrive
api_host = https://example.com/tgd
api_key = tdk_...
chunk_size = 64M
upload_concurrency = 4
encrypt_files = true
hash_enabled = true
```

Build with Go 1.25 or newer:

```sh
go build -o rclone .
```
