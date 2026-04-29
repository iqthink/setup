# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`iqdev` is a Go BubbleTea TUI distributed as a single Mac binary. It is the
**global** half of the iqthink Rails-app onboarding flow: it installs
everything that is identical across the 20+ apps (Homebrew, Xcode CLT, brew
packages, OrbStack, shell rc). Each app's own `bin/setup` then handles the
**project-specific** half (Postgres in Docker, master.key, gems, JS, db, seeds).

Module: `github.com/iqthink/setup`. Binary: `iqdev`. macOS-only (arm64 + amd64).

## Commands

Go is managed by mise. Always invoke Go through `mise exec --` from the repo root:

```bash
mise install                                              # install Go (first time)
mise exec -- go run ./cmd/iqdev                           # run the TUI live (no build)
mise exec -- go build -o /tmp/iqdev ./cmd/iqdev           # build a local binary
mise exec -- go test ./...                                # run all tests
mise exec -- go test ./internal/tui -run TestWelcomeViewRenders  # one test
mise exec -- go vet ./...
```

Cross-compile for release smoke:

```bash
mise exec -- bash -c 'GOOS=darwin GOARCH=arm64 go build ./cmd/iqdev'
mise exec -- bash -c 'GOOS=darwin GOARCH=amd64 go build ./cmd/iqdev'
```

A release is cut by tagging `vX.Y.Z` and pushing the tag. The
`.github/workflows/release.yml` action runs GoReleaser (config in
`.goreleaser.yaml`) which publishes per-arch binaries plus a per-arch
`SHA256SUMS-darwin-<arch>.txt`. `scripts/install.sh` reads exactly that layout.

## Architecture

### Homebrew is a prerequisite, not a step

`iqdev` does **not** install Homebrew. Earlier versions tried to bootstrap it
before `tea.NewProgram(...)`, but that broke under `curl … | bash` (the brew
installer reads from stdin for sudo, but stdin is the curl pipe — even with
`/dev/tty` reattached at the iqdev level, the inner `bash -c "curl … | bash"`
gets a pipe). So `cmd/iqdev/main.go` checks `brewenv.Installed()`; if missing,
it prints the official Homebrew install command and exits non-zero. Both
`scripts/install.sh` and the iqdev binary do this check. Once Homebrew is
present, `brewenv.AddToPath()` prepends `/opt/homebrew/bin` (or
`/usr/local/bin` on Intel) so the TUI's brew calls find it.

### Step pipeline

Every install step implements the `Step` interface in `internal/steps/step.go`:

```go
Name()  string
Check(ctx) (done bool, err error)   // skip if already installed
Run(ctx, out chan<- string) error   // stream stdout+stderr lines to `out`
```

`steps.All()` returns the canonical ordered pipeline (Xcode CLT → brew packages →
OrbStack → shell config). `runCmd(ctx, out, name, args...)` in `step.go` is the
shared helper: it merges stdout+stderr through an `io.Pipe` and bufio-scans
lines into the channel. Use it from any new step instead of rolling another
exec wrapper.

The brew packages step keeps a single `Packages` list in `brew_packages.go`.
It runs `brew update` first (stale installs trip on cask DSL features), then
issues one `brew install <missing packages>` — brew auto-resolves casks vs.
formulae, so we don't pass `--cask` even for casks like `1password-cli` (this
matches the 1Password docs, which show `brew install 1password-cli` plain).
The existence check accepts either formula OR cask presence (`brew list
--versions` and `brew list --cask --versions`) so we don't loop on whichever
side brew lands on. The Stripe CLI tap is qualified inline
(`stripe/stripe-cli/stripe`); for the existence check the helper takes the
last path segment as the package name.

### TUI is a thin shell over the pipeline

`internal/tui` is the BubbleTea program. The pipeline runs in its own goroutine
(`runPipeline` in `update.go`) and emits `pipeMsg` values (`kindStarted`,
`kindLine`, `kindDone`, `kindSkipped`, `kindFailed`) into a buffered channel.
The `Model.Update` method consumes them via `waitForPipe` — a `tea.Cmd` that
blocks on one channel receive per call. After processing each message, Update
returns `waitForPipe()` again to fetch the next. This is the recommended
BubbleTea pattern for streaming long-running work; do not try to drive UI
updates from inside the step goroutines.

States: `welcome → running → (done | failed)`. The welcome screen lists the
steps (always shown, never short-circuited even if everything will be skipped)
so the user always sees what we're about to do.

`tea.WithAltScreen()` is used, so when the program exits the screen is
restored. The post-exit summary is printed by `main.go` (not the View),
checking `m.Failed()` / `m.FailedStep()` / `m.FailErr()`.

### `iqdev update` is a self-contained Go re-implementation of install.sh

`internal/update/update.go` calls the GitHub API for the latest tag, downloads
`iqdev-darwin-<arch>` and `SHA256SUMS-darwin-<arch>.txt`, verifies the SHA-256,
writes to `<self>.new`, and `os.Rename`s over the running executable. Both
this code and `install.sh` rely on GoReleaser's split checksum layout — keep
them in sync.

### `install.sh` redirects stdin from /dev/tty

When the user runs `curl ... | bash`, the shell's stdin is the curl pipe, so
the TUI cannot read keys. Before `exec`ing `iqdev`, the script re-attaches
stdin from `/dev/tty`. Any further changes to the install flow must preserve
this redirection.

## Conventions

- All user-facing text (TUI, help, install.sh, update messages) is **English**.
  Internal Go errors are also English.
- Don't use `tea.ExecProcess`; the Homebrew bootstrap pattern obviates it. If a
  future step truly needs to suspend the TUI, that's the time to introduce it.
- New steps go in `internal/steps/<name>.go` and register in `steps.All()`.
  They must be idempotent (`Check` returning true must mean the next run
  skips quickly).
- The split between this repo and per-app `bin/setup`: anything common across
  all iqthink Rails apps belongs here; anything that varies per repo (Postgres
  port, master.key vault path, seeds) stays in `bin/setup` of that repo.
