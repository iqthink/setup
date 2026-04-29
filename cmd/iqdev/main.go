package main

import (
	"fmt"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iqthink/setup/internal/brewenv"
	"github.com/iqthink/setup/internal/tui"
	"github.com/iqthink/setup/internal/ui"
	"github.com/iqthink/setup/internal/update"
)

// version is injected at build time via -ldflags "-X main.version=...".
var version = "dev"

const helpText = `iqdev — Global Mac bootstrap for iqthink Rails apps

USAGE:
  iqdev              Install whatever is missing. Idempotent: if everything
                     is already in place, it finishes in seconds.
  iqdev update       Download the latest release and replace the binary.
  iqdev --version    Print the version.
  iqdev --help       Show this help.

PREREQUISITE:
  Homebrew. iqdev does not install it because it needs a real terminal for
  the sudo prompt. Install it first:

    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

WHAT IT INSTALLS:
  · Xcode Command Line Tools (opens the native dialog)
  · Homebrew packages: gum, mise, gh, 1password-cli, hivemind, stripe-cli,
    libpq, libyaml, openssl@3, gmp, rust, vips, imagemagick, redis
  · OrbStack (Docker for macOS)
  · Shell config (~/.zshrc or ~/.bashrc): PATH and mise activate

AFTER iqdev:
  Go to your Rails app and run 'bin/setup'. Then 'bin/dev'.

REPO: https://github.com/iqthink/setup
`

const homebrewMissingMsg = `iqdev requires Homebrew, but it is not installed.

Install it first (in this terminal, so it can ask for your password):

  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

Then re-run iqdev.
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
		fmt.Fprintf(os.Stderr, "\n✗ Failed: %s\n", m.FailedStep())
		if m.FailErr() != nil {
			fmt.Fprintf(os.Stderr, "  %s\n", m.FailErr())
		}
		fmt.Fprintln(os.Stderr, "Re-run `iqdev` to retry.")
		os.Exit(1)
	}

	printNextSteps()
}

func printNextSteps() {
	step := func(num, label, cmd string, notes ...string) {
		fmt.Println("  " + ui.StepNumber.Render(num+".") + " " + ui.StepLabel.Render(label))
		fmt.Println("       " + ui.Command.Render(cmd))
		for _, n := range notes {
			fmt.Println("     " + ui.Note.Render(n))
		}
	}

	fmt.Println()
	fmt.Println(ui.DoneHeader.Render("✓ Done."))
	fmt.Println()
	fmt.Println(ui.StepHeader.Render("Next steps:"))
	fmt.Println()
	step("1", "Sign in to 1Password CLI:", "op signin",
		"(First time? Open 1Password → Settings → Developer →",
		" enable \"Integrate with 1Password CLI\".)")
	fmt.Println()
	step("2", "Authenticate the GitHub CLI:", "gh auth login")
	fmt.Println()
	step("3", "In your Rails app, run:", "bin/setup")
	fmt.Println()
	fmt.Println(ui.Note.Render("  First time? Close and reopen your terminal so mise activates."))
}
