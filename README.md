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

`wenet package` reads `edge.toml` from the current directory, applies `.edgeignore`
only, and writes a deployment archive under `.wenet/`. `.gitignore` is not used
for packaging rules.

`wenet push` ensures the project exists in WENet, then creates or overwrites the
package identified by `(project, tag)`.

`wenet deploy` performs the same package push, then creates a rollout using the
targeting and runtime settings from `edge.toml`.

```bash
# Local archive upload: build locally, upload file to WENet.
wenet push
wenet deploy

# Existing archive upload: send a prebuilt .tar.gz.
wenet push --package-file dist/app.tar.gz
wenet deploy --package-file dist/app.tar.gz

# Git repo import: WENet clones and packs the repo server-side.
wenet push --source-url https://github.com/org/repo --source-ref main
wenet deploy --source-url https://github.com/org/repo --source-ref main
```

For private repos, pass a PAT/deploy token with `--source-token` or
`SOURCE_TOKEN`. Source URLs must be HTTPS clone URLs; SSH URLs are not supported.

CI env vars:

```bash
SOURCE_URL=https://github.com/org/repo
SOURCE_REF=main
SOURCE_TOKEN=...
PACKAGE_FILE=dist/app.tar.gz
```

## edge.toml

```toml
project = "my-web-server"
tag = "1.2.0"
all = true
download_base_dir = "/tmp"
cleanup = true

[scripts]
linux = "deploy.sh"
darwin = "deploy.sh"
windows = "deploy.ps1"
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
