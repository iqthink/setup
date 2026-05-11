package steps

import (
	"context"
	"os/exec"
	"strings"
)

// MiseConfig sets `ruby.compile=false` so mise downloads precompiled Ruby
// binaries instead of compiling from source. Per the mise docs, precompiled
// binaries become the default in 2026.8.0; setting it now keeps installs
// fast for everyone on older mise versions. It also sets `node.verify=false`
// to skip GPG signature verification on Node downloads, which avoids slow or
// flaky keyserver lookups on fresh machines.
type MiseConfig struct{}

func (MiseConfig) Name() string { return "Mise config (precompiled Ruby, skip Node verify)" }

func (MiseConfig) Check(ctx context.Context) (bool, error) {
	if _, err := exec.LookPath("mise"); err != nil {
		return false, nil
	}
	return miseSettingIs(ctx, "ruby.compile", "false") &&
		miseSettingIs(ctx, "node.verify", "false"), nil
}

func (MiseConfig) Run(ctx context.Context, out chan<- string) error {
	out <- "Setting ruby.compile=false (use precompiled Ruby binaries)"
	if err := runCmd(ctx, out, "mise", "settings", "ruby.compile=false"); err != nil {
		return err
	}
	out <- "Setting node.verify=false (skip Node GPG verification)"
	return runCmd(ctx, out, "mise", "settings", "node.verify=false")
}

func miseSettingIs(ctx context.Context, key, want string) bool {
	out, err := exec.CommandContext(ctx, "mise", "settings", "get", key).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == want
}
