# rclone with TelDrive v2

A custom build of [rclone](https://rclone.org/) with **TelDrive v2 backend support**.

TelDrive provides a file-storage layer backed by Telegram, while rclone provides the familiar command-line interface for copying, syncing, mounting, and managing files.

> **Status:** Working / development build
> **Architecture tested:** Linux ARM64 (Raspberry Pi)
> **Go version tested:** Go 1.25
> **rclone base:** v1.76.0-DEV

## Features

* TelDrive v2 backend
* TelDrive API-key authentication
* Telegram-backed file storage through TelDrive
* Directory listing
* File uploads/downloads
* rclone commands such as `ls`, `lsd`, `copy`, `sync`, etc.
* Configurable upload chunk size
* Configurable upload concurrency
* Optional TelDrive file encryption
* Optional BLAKE3 tree hashing
* Public link support
* Threaded upload streams

## Requirements

* Go 1.25 or newer
* Linux/macOS/Windows supported by Go, depending on the target platform
* A running TelDrive v2 server
* A TelDrive API key
* A TelDrive storage channel configured on the server

For Raspberry Pi, a 64-bit OS is recommended.

## Build

Clone the repository:

```bash
git clone https://github.com/YOUR_USERNAME/YOUR_REPOSITORY.git
cd YOUR_REPOSITORY
```

Build rclone:

```bash
go build -o rclone .
```

Check the resulting binary:

```bash
./rclone version
```

You should see something similar to:

```text
rclone v1.76.0-DEV
- os/type: linux
- os/arch: arm64
- go/version: go1.25.x
```

### Install system-wide

On Linux:

```bash
sudo install -m 0755 rclone /usr/local/bin/rclone-teldrive
```

Then:

```bash
rclone-teldrive version
```

Using the name `rclone-teldrive` keeps this build separate from a normal system installation of official rclone.

## TelDrive configuration

Create a TelDrive remote using:

```bash
rclone-teldrive config
```

Choose:

```text
n) New remote
```

Set the backend to:

```text
teldrive
```

The important settings are:

```text
api_key
api_host
channel_id
```

Example configuration:

```ini
[teldrive]
type = teldrive
api_key = YOUR_TELDRIVE_API_KEY
api_host = https://your-teldrive-server.example.com
channel_id = YOUR_CHANNEL_ID
```

**Never commit your API key to Git.**

The rclone configuration is normally stored at:

```text
~/.config/rclone/rclone.conf
```

## Test the connection

List directories:

```bash
rclone-teldrive lsd teldrive:
```

List files:

```bash
rclone-teldrive ls teldrive:
```

Get information about the backend:

```bash
rclone-teldrive help backend teldrive
```

## Example operations

Copy a local file to TelDrive:

```bash
rclone-teldrive copy ./file.iso teldrive:
```

Copy a directory:

```bash
rclone-teldrive copy ./my-folder teldrive:my-folder
```

List a remote directory:

```bash
rclone-teldrive ls teldrive:my-folder
```

Sync a local directory:

```bash
rclone-teldrive sync ./my-folder teldrive:my-folder
```

> Be careful with `sync`: files at the destination can be deleted to make the destination match the source.

## Performance settings

The TelDrive backend supports configurable upload parameters.

### Chunk size

Default:

```text
512 MiB
```

Example:

```bash
rclone-teldrive copy ./large-file teldrive: \
  --teldrive-chunk-size 512M
```

Supported range:

```text
64 MiB – 2000 MiB
```

### Upload concurrency

Default:

```text
4
```

Example:

```bash
rclone-teldrive copy ./large-file teldrive: \
  --teldrive-upload-concurrency 4
```

Higher values can increase upload speed but also increase memory usage and may increase API/Telegram rate limiting.

## Optional encryption

TelDrive native encryption can be enabled with:

```text
--teldrive-encrypt-files
```

For example:

```bash
rclone-teldrive copy ./file.bin teldrive: \
  --teldrive-encrypt-files
```

Encryption behavior depends on the TelDrive server configuration.

## BLAKE3 integrity checking

BLAKE3 tree hashing is enabled by default.

It can be disabled with:

```text
--teldrive-hash-enabled=false
```

The backend uses 16 MiB blocks for the tree hash.

## Raspberry Pi

This project has been tested on:

```text
Raspberry Pi
64-bit Raspberry Pi OS
Linux ARM64 / AArch64
Go 1.25
```

Example:

```bash
go build -o rclone .
sudo install -m 0755 rclone /usr/local/bin/rclone-teldrive
```

Then:

```bash
/usr/local/bin/rclone-teldrive version
```

## Troubleshooting

### `authentication is required`

Verify that the API key is correct.

The TelDrive API uses the:

```text
X-Api-Key
```

HTTP header.

You can test authentication directly:

```bash
curl -i \
  -H "X-Api-Key: YOUR_TELDRIVE_API_KEY" \
  https://your-teldrive-server.example.com/api/v1/me
```

A successful response should return information about the authenticated TelDrive user.

### `didn't find section in config file`

Make sure the remote name matches your configuration.

For example, if your configuration contains:

```ini
[teldrive]
type = teldrive
```

use:

```bash
rclone-teldrive ls teldrive:
```

If your remote is named:

```ini
[teldrive-v2]
type = teldrive
```

use:

```bash
rclone-teldrive ls teldrive-v2:
```

### `no overview data found for "teldrive"`

The repository includes:

```text
docs/data/backends/teldrive.yaml
```

This file provides the backend overview metadata required by this rclone build.

If you are working from an older checkout, make sure this file is present before rebuilding.

## Development

Run the TelDrive backend tests:

```bash
go test ./backend/teldrive
```

Run the complete test suite:

```bash
go test ./...
```

Build:

```bash
go build -o rclone .
```

## Disclaimer

This is a custom/development build of rclone with TelDrive v2 support.

It is **not an official rclone release** unless explicitly stated otherwise.

Use production deployments at your own risk and test important data operations before relying on them for backups or synchronization.

## License

This project is based on rclone.

See the repository's existing license files and upstream rclone licensing information for applicable terms.

## Credits

* [rclone](https://rclone.org/) — command-line cloud storage tool
* TelDrive — Telegram-backed storage server

If you find a problem with the TelDrive backend, please open an issue with:

* rclone version
* operating system
* architecture
* Go version
* command being executed
* relevant `-vv` output

**Never include API keys, passwords, tokens, or other credentials in an issue.**
