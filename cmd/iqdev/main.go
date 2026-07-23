package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iqthink/setup/internal/brewenv"
	"github.com/iqthink/setup/internal/steps"
	"github.com/iqthink/setup/internal/tui"
	"github.com/iqthink/setup/internal/ui"
	"github.com/iqthink/setup/internal/update"
)

// version is injected at build time via -ldflags "-X main.version=...".
var version = "dev"

const helpText = `iqdev — Global Mac bootstrap for iqthink

USAGE:
  iqdev              Install whatever is missing. Idempotent: if everything
                     is already in place, it finishes in seconds. On a fresh
                     Mac, prompts once for which optional team apps to add.
  iqdev pick         Re-open the optional team-apps picker without running
                     the rest of the pipeline.
  iqdev update       Download the latest release and replace the binary.
  iqdev --version    Print the version.
  iqdev --help       Show this help.

PREREQUISITE:
  Homebrew. iqdev does not install it because it needs a real terminal for
  the sudo prompt. Install it first:

    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

WHAT IT INSTALLS:
  Rails dev stack:
  · Xcode Command Line Tools (opens the native dialog)
  · Homebrew packages: gum, mise, gh, 1password-cli, hivemind, stripe-cli,
    libpq, libyaml, openssl@3, gmp, rust, vips, imagemagick, redis
  · Mise config (precompiled Ruby binaries instead of compiling from source)
  · Node.js LTS (latest, pinned globally via mise)
  · Yarn (via Corepack) and pnpm (global npm install)
  · OrbStack (Docker for macOS)
  · Shell config (~/.zshrc or ~/.bashrc): PATH and mise activate

  iqthink team apps (always installed):
  · 1Password GUI · Basecamp + CLI · Zoom · Claude · Conductor
  · Ghostty · TablePro · Bruno · CleanShot · Dia browser

  Optional team apps (picked on first run, saved to ~/.iqdev-team-profile):
  · Raycast · ChatGPT · Figma · Tailscale · gcloud CLI · Zed · GitHub Desktop
  · Screen Studio · Microsoft Office · Dropbox · HEY · WhatsApp

AFTER iqdev:
  1. Enable the 1Password CLI (one-time, in the GUI): open the 1Password
     app, sign in, then Settings → Developer → turn ON "Integrate with
     1Password CLI". Verify with 'op signin'.
  2. Authenticate the GitHub CLI: 'gh auth login'.
  3. In a NEW terminal (so mise activates), go to your Rails app and run
     'bin/setup'. Then 'bin/dev'.

REPO: https://github.com/iqthink/setup
`

const homebrewMissingMsg = `iqdev requires Homebrew, but it is not installed.

Install it first (copy-paste this in this terminal):

  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

It will ask for your Mac login password (typing is invisible — that's
normal), then ask you to press RETURN. It also installs Apple's Command
Line Tools, so it can take several minutes.

When it says "Installation successful!", run iqdev again. You can ignore
Homebrew's own "Next steps" about PATH — iqdev handles that.
`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help", "help":
			fmt.Print(helpText)
			return
		case "-v", "--version", "version", "--v":
			fmt.Println(version)
			return
		case "update":
			if err := update.Run(version); err != nil {
				fmt.Fprintln(os.Stderr, "✗", err)
				os.Exit(1)
			}
			return
		case "pick":
			if err := runPicker(); err != nil {
				fmt.Fprintln(os.Stderr, "✗", err)
				os.Exit(1)
			}
			return
		default:
			fmt.Fprintf(os.Stderr, "iqdev: unknown subcommand %q\n\n", os.Args[1])
			fmt.Fprint(os.Stderr, helpText)
			os.Exit(2)
		}
	}

	runTUI()
}

func runTUI() {
	if runtime.GOOS != "darwin" {
		fmt.Fprintf(os.Stderr, "iqdev: only macOS is supported for now (you are on %s)\n", runtime.GOOS)
		os.Exit(1)
	}

	if !brewenv.Installed() {
		fmt.Fprint(os.Stderr, homebrewMissingMsg)
		os.Exit(1)
	}
	brewenv.AddToPath()

	// Prompt once for sudo up front (while stdin is still a real terminal) if
	// there is pending work, so installers that need root (Zoom, OrbStack) get
	// the password via $SUDO_ASKPASS instead of an unanswerable prompt inside
	// the alt-screen TUI. Non-fatal: if it can't be obtained we continue and
	// let the summary report whatever needed it. Skipped when nothing pends.
	ctx := context.Background()
	if steps.AnyPending(ctx) {
		defer primeSudo(ctx)()
	}

	p := tea.NewProgram(tui.New(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	m, ok := finalModel.(tui.Model)
	if !ok {
		return
	}
	if m.Failed() {
		failures := m.Failures()
		fmt.Fprintf(os.Stderr, "\n✗ Finished with %d issue(s) — everything else installed:\n\n", len(failures))
		for _, f := range failures {
			fmt.Fprintf(os.Stderr, "  ✗ %s\n", f.Step)
			if f.Err != nil {
				fmt.Fprintf(os.Stderr, "      %s\n", f.Err)
			}
			for _, t := range lastLines(f.Tail, 6) {
				fmt.Fprintf(os.Stderr, "      %s\n", t)
			}
		}
		fmt.Fprintln(os.Stderr, "\nRe-run `iqdev` to retry (it's idempotent).")
		os.Exit(1)
	}

	printNextSteps()
}

func lastLines(s []string, n int) []string {
	if len(s) > n {
		return s[len(s)-n:]
	}
	return s
}

// printNextSteps re-prints the done screen's checklist after the alt-screen
// TUI closes, so it survives in the terminal scrollback. The content is
// tui.NextStepsBody() — the same single source the done screen renders — so
// the two can never drift.
func printNextSteps() {
	fmt.Println()
	fmt.Println(ui.DoneHeader.Render("✓ Done."))
	fmt.Println()
	fmt.Println(tui.NextStepsBody())
	fmt.Println()
}
