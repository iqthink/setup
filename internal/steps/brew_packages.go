package steps

import (
	"context"
	"strings"

	"github.com/iqthink/setup/internal/brewenv"
)

// Packages installed by `brew install`. Tap-qualified names are allowed
// (e.g. "stripe/stripe-cli/stripe"); for the existence check we use the
// final segment as the formula name.
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
		name := p
		if i := strings.LastIndex(p, "/"); i >= 0 {
			name = p[i+1:]
		}
		if !brewenv.PackageInstalled(name) {
			return false, nil
		}
	}
	return true, nil
}

func (BrewPackages) Run(ctx context.Context, out chan<- string) error {
	args := append([]string{"install"}, Packages...)
	return runCmd(ctx, out, brewenv.BrewPath(), args...)
}
