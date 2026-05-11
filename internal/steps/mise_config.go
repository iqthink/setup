package steps

import (
	"context"
	"os/exec"
	"strings"
)

// MiseConfig sets `ruby.compile=false` so mise downloads precompiled Ruby
// binaries instead of compiling from source. Per the mise docs, precompiled
// binaries become the default in 2026.8.0; setting it now keeps installs
// fast for everyone on older mise versions.
type MiseConfig struct{}

func (MiseConfig) Name() string { return "Mise config (precompiled Ruby)" }

func (MiseConfig) Check(ctx context.Context) (bool, error) {
	if _, err := exec.LookPath("mise"); err != nil {
		return false, nil
	}
	out, err := exec.CommandContext(ctx, "mise", "settings", "get", "ruby.compile").Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) == "false", nil
}

func (MiseConfig) Run(ctx context.Context, out chan<- string) error {
	out <- "Setting ruby.compile=false (use precompiled Ruby binaries)"
	return runCmd(ctx, out, "mise", "settings", "ruby.compile=false")
}
