# iqdev setup

Global Mac bootstrap for iqthink.

A single TUI (BubbleTea) that gets a fresh Mac ready for iqthink work in
minutes — Rails dev stack plus the team-baseline apps every iqthink dev uses.
(Homebrew is required first.)

## What it installs

**Rails dev stack** (always)
Xcode CLT · mise · gum · gh · 1Password CLI · hivemind · Stripe CLI · libpq ·
libyaml · openssl@3 · gmp · rust · vips · ImageMagick · redis · OrbStack ·
Node LTS (via mise) · Yarn (Corepack) · pnpm · shell PATH + mise activate.

**iqthink team apps** (always)
1Password GUI · Basecamp (app + CLI) · Zoom · Claude · Conductor · Ghostty ·
TablePro · Bruno · CleanShot · Dia browser.

**Optional team apps** (you pick on first run, saved to `~/.iqdev-team-profile`)
Raycast · ChatGPT · Figma · Tailscale · gcloud CLI · Zed · GitHub Desktop ·
Screen Studio · Microsoft Office · Dropbox · HEY · WhatsApp.

The optional picker only appears on the first run on each Mac. After that,
iqdev reads the profile silently — run `iqdev pick` whenever you want to add
or remove options.

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
missing, and runs it. On a fresh Mac, the TUI will pause for you to pick
optional team apps before any install starts.

When the pipeline finishes, iqdev prints your next steps:

1. **Enable the 1Password CLI** (one-time, easy to miss): open the
   1Password app, sign in, then Settings → Developer → turn on
   *"Integrate with 1Password CLI"*. Verify with `op signin`.
2. Authenticate the GitHub CLI: `gh auth login`.
3. In a **new** terminal (so mise activates), go to your Rails app and run
   `bin/setup`. No repo yet? `gh repo clone iqthink/<app>` first.

Some apps ask for macOS permissions the first time you open them (CleanShot
→ screen recording, Zoom → mic/camera, Tailscale → VPN); approve them in
System Settings → Privacy & Security when prompted.

## Commands

```
iqdev              Install whatever is missing (idempotent). On a fresh
                   Mac, opens the optional-apps picker once.
iqdev pick         Re-open the optional team-apps picker. Saves to
                   ~/.iqdev-team-profile. Does not run the install
                   pipeline — re-run `iqdev` afterward to apply.
iqdev update       Update the binary to the latest release.
iqdev --version    Print the version.
iqdev --help       Show help.
```

## Development

```bash
mise install                                              # installs Go
mise exec -- go run ./cmd/iqdev                           # run the TUI live
mise exec -- go build -o /tmp/iqdev ./cmd/iqdev           # build a binary
mise exec -- go test ./...                                # all tests
mise exec -- go vet ./...
```

Releases are published via GitHub Actions whenever a `v*` tag is pushed.
Artifacts are produced with GoReleaser for `darwin/arm64` and `darwin/amd64`.
