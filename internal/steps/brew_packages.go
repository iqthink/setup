package steps

import (
	"context"
	"strings"

	"github.com/iqthink/setup/internal/brewenv"
)

// Packages installed by `brew install`. Tap-qualified names are allowed
// (e.g. "stripe/stripe-cli/stripe"); for the existence check we use the
// final segment as the formula/cask name.
//
// 1password-cli per the official docs is installed plain (`brew install
// 1password-cli`); brew auto-resolves it. The check tries both formula
// and cask state so we don't loop on whichever side brew lands on.
var Packages = []string{
	"gum",
	"mise",
	"gh",
	"1password-cli",
	"hivemind",
	"stripe/stripe-cli/stripe",
	"libpq",
	"libyaml",
	"openssl@3",
	"gmp",
	"rust",
	"vips",
	"imagemagick",
	"redis",
}

type BrewPackages struct{}

func (BrewPackages) Name() string { return "Homebrew packages" }

func (BrewPackages) Check(ctx context.Context) (bool, error) {
	for _, p := range Packages {
		if !installed(p) {
			return false, nil
		}
	}
	return true, nil
}

func (BrewPackages) Run(ctx context.Context, out chan<- string) error {
	out <- "Updating Homebrew..."
	if err := runCmd(ctx, out, brewenv.BrewPath(), "update"); err != nil {
		return err
	}
	// Newer Homebrew refuses to install from a 3rd-party tap until that tap
	// has been explicitly trusted, so we trust each tap referenced by a
	// tap-qualified package in Packages (e.g. "stripe/stripe-cli/stripe").
	// `brew trust` on an already-trusted tap is a no-op.
	for _, tap := range tapsInPackages() {
		if err := runCmd(ctx, out, brewenv.BrewPath(), "trust", tap); err != nil {
			return err
		}
	}
	args := []string{"install"}
	for _, p := range Packages {
		if !installed(p) {
			args = append(args, p)
		}
	}
	if len(args) == 1 {
		return nil
	}
	return runCmd(ctx, out, brewenv.BrewPath(), args...)
}

// tapsInPackages returns the distinct "<user>/<tap>" prefixes of every
// tap-qualified entry in Packages, preserving declaration order.
func tapsInPackages() []string {
	seen := map[string]bool{}
	var taps []string
	for _, p := range Packages {
		parts := strings.Split(p, "/")
		if len(parts) < 3 {
			continue
		}
		tap := parts[0] + "/" + parts[1]
		if seen[tap] {
			continue
		}
		seen[tap] = true
		taps = append(taps, tap)
	}
	return taps
}

func installed(pkg string) bool {
	name := pkg
	if i := strings.LastIndex(pkg, "/"); i >= 0 {
		name = pkg[i+1:]
	}
	return brewenv.PackageInstalled(name) || brewenv.CaskInstalled(name)
}
