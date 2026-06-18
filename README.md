# wenet-cli

Command-line client for WENet deployments.

## Install

Linux/macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/wenet-ec/wenet-cli/main/install.sh | sh
```

Windows PowerShell:

```powershell
iwr https://raw.githubusercontent.com/wenet-ec/wenet-cli/main/install.ps1 -UseB | iex
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/wenet-ec/wenet-cli/main/install.sh | WENET_CLI_VERSION=v0.1.0 sh
```

```powershell
$env:WENET_CLI_VERSION="v0.1.0"; iwr https://raw.githubusercontent.com/wenet-ec/wenet-cli/main/install.ps1 -UseB | iex
```

Manual downloads are published on GitHub Releases for Linux, macOS, and Windows on `amd64` and `arm64`.

## Commands

```bash
wenet login <token>
wenet logout
wenet package
wenet push
wenet deploy
```

`wenet package` reads `edge.toml` from the current directory, applies `.gitignore`
and `.edgeignore`, and writes a deployment archive under `.wenet/`.

`wenet push` and `wenet deploy` are command slots for the public API integration.

## edge.toml

```toml
project = "my-web-server"
tag = "1.2.0"

[scripts]
linux = "deploy.sh"
darwin = "deploy.sh"
windows = "deploy.ps1"

all = true
download_base_dir = "/tmp"
cleanup = true
```

At least one script key is required. Valid keys are `linux`, `darwin`, and
`windows`. Every declared script must exist in the package archive.

## Release

Releases are built by GoReleaser when a `v*` tag is pushed:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow builds Linux, macOS, and Windows binaries for `amd64` and
`arm64`, uploads archives to GitHub Releases, and publishes SHA-256 checksums.
