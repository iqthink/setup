# iqdev setup

Global Mac bootstrap for running any iqthink Rails app.

A single TUI (BubbleTea) that installs everything your Mac needs to clone a
repo and have it running in minutes: Xcode CLT, mise, gum, gh, 1Password CLI,
hivemind, Stripe CLI, libpq, libyaml, openssl@3, gmp, rust, vips, ImageMagick,
redis, and OrbStack. (Homebrew is required first.)

## Prerequisite: Homebrew

iqdev does **not** install Homebrew, because the brew installer needs an
interactive sudo prompt that doesn't work under `curl | bash`. Install it
yourself first (in a real terminal, so it can ask for your password):

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/iqthink/setup/main/scripts/install.sh | bash
```

This downloads the binary to `~/.local/bin/iqdev`, adds it to your PATH if
missing, and runs it. When it finishes, go to your Rails app and run
`bin/setup`.

## Commands

```
iqdev              Install whatever is missing (idempotent)
iqdev update       Update the binary to the latest release
iqdev --version    Print the version
iqdev --help       Show help
```

## Development

```bash
mise install       # installs Go
go run ./cmd/iqdev # runs the TUI without compiling
go build ./...     # builds
```

Releases are published via GitHub Actions whenever a `v*` tag is pushed.
Artifacts are produced with GoReleaser for `darwin/arm64` and `darwin/amd64`.
