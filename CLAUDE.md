# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`iqdev` is a Go BubbleTea TUI distributed as a single Mac binary. It is the
**global** half of iqthink Mac onboarding: anything that goes on every iqthink
dev's machine. Each app's own `bin/setup` still handles the **project-specific**
half (Postgres in Docker, master.key, gems, JS, db, seeds).

Scope is three concentric layers:

1. **Rails dev stack** (always installed) — Homebrew packages, Xcode CLT,
   mise, OrbStack, shell rc. Identical across all 20+ Rails apps.
2. **Team-apps baseline** (always installed) — apps every iqthink dev uses
   for work, regardless of which app they're touching: 1Password GUI,
   Basecamp + CLI, Zoom, Claude, Conductor, Ghostty, TablePro, Bruno,
   CleanShot, Dia browser. See `internal/steps/team_apps.go`
   (`TeamAppsMustPackages`).
3. **Optional team apps** (per-Mac, opted in via picker) — apps the team
   uses but not every dev needs on every machine (Raycast, Figma, Zed,
   gcloud, etc.). Selection persists to `~/.iqdev-team-profile`. See
   `TeamAppsOptionalCandidates` in the same file.

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

`steps.All()` returns the canonical ordered pipeline:
Xcode CLT → BrewPackages → TeamAppsMust → TeamAppsOptional → mise → Node LTS
→ Node PMs → OrbStack → shell config. The team-apps steps land after the
Rails brew packages because they share Homebrew (and `brew update` only
needs to run once), but before the language-specific steps that the team
apps don't depend on. `runCmd(ctx, out, name, args...)` in `step.go` is
the shared helper: it merges stdout+stderr through an `io.Pipe` and
bufio-scans lines into the channel. Use it from any new step instead of
rolling another exec wrapper.

The brew packages step keeps a single `Packages` list in `brew_packages.go`.
It runs `brew update` first (stale installs trip on cask DSL features),
calls `brew trust` for any 3rd-party tap referenced in a tap-qualified
name (e.g. `stripe/stripe-cli/stripe` ⇒ trust `stripe/stripe-cli`) — newer
Homebrew refuses to install from untrusted taps — then issues one
`brew install <missing packages>`. Brew auto-resolves casks vs. formulae,
so we don't pass `--cask` even for casks like `1password-cli` (this matches
the 1Password docs, which show `brew install 1password-cli` plain). The
existence check accepts either formula OR cask presence (`brew list
--versions` and `brew list --cask --versions`) so we don't loop on whichever
side brew lands on. The tap-extraction helper `tapsIn([]string)` lives in
`step.go` and is reused by every step that handles tap-qualified packages
(`BrewPackages`, `TeamAppsMust`, `TeamAppsOptional`).

### Team-apps steps

`TeamAppsMust` is a thin clone of `BrewPackages` over `TeamAppsMustPackages`:
trust → install missing. Edits to the team baseline happen here.

`TeamAppsOptional` reads `~/.iqdev-team-profile` (via the `profile`
package) and installs whatever package names it finds. Empty or missing
profile = step is "done" (skipped in the TUI), not "no-op work": the
user has been through the picker and chose nothing, or hasn't been
through it yet and the welcome→picker→running flow handles them.

The candidate list shown by the picker is `TeamAppsOptionalCandidates`
(a slice of `{Pkg, Label, Hint}`) in the same file — add a new optional
app by appending here.

### Picker state

The TUI now has four runtime states: `welcome → config → running → done/failed`,
where `config` is the optional-apps picker. The transition rule:

- First run (no profile on disk): welcome → enter → config (picker) →
  enter|esc → running.
- Subsequent runs (profile present): welcome → enter → running, no picker.
- `iqdev pick` subcommand: standalone picker, no welcome/running, just
  load profile → mutate → save profile → exit.

The picker model lives in `internal/picker/picker.go` and is reused by the
main TUI's `config` state and by `cmd/iqdev/pick.go`. It is a pure model
(no I/O) — the saving happens in `tui/update.go` (main flow) and
`cmd/iqdev/pick.go` (subcommand). This way the picker is unit-testable
without a real filesystem.

Profile persistence lives in `internal/profile/profile.go`. It is a
newline-delimited file (one package name per line, `#` comments
ignored) at `~/.iqdev-team-profile`. Format choice: human-readable so a
dev can `cat` or hand-edit it during debugging, atomic-rename on save
to survive a half-written file.

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
