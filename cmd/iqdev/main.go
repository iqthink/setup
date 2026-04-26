package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iqthink/setup/internal/brewenv"
	"github.com/iqthink/setup/internal/tui"
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

WHAT IT INSTALLS:
  · Homebrew (if missing; installed before the TUI; will ask for your Mac password)
  · Xcode Command Line Tools (opens the native dialog)
  · Homebrew packages: gum, mise, gh, 1password-cli, hivemind, stripe-cli,
    libpq, libyaml, openssl@3, gmp, rust, vips, imagemagick, redis
  · OrbStack (Docker for macOS)
  · Shell config (~/.zshrc or ~/.bashrc): PATH and mise activate

AFTER iqdev:
  Go to your Rails app and run 'bin/setup'. Then 'bin/dev'.

REPO: https://github.com/iqthink/setup
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
		if err := bootstrapHomebrew(); err != nil {
			fmt.Fprintln(os.Stderr, "✗ Homebrew installation failed:", err)
			os.Exit(1)
		}
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

	fmt.Println()
	fmt.Println("✓ Done. Now go to your Rails app and run: bin/setup")
	fmt.Println("  (First time? Close and reopen your terminal so mise activates.)")
}

func bootstrapHomebrew() error {
	fmt.Println()
	fmt.Println("──────────────────────────────────────────────────────────────────")
	fmt.Println("  Homebrew is not installed. Let's install it first.")
	fmt.Println("  It will ask for your Mac password.")
	fmt.Println("──────────────────────────────────────────────────────────────────")
	fmt.Println()

	cmd := exec.Command("/bin/bash", "-c",
		`curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh | bash`)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}
