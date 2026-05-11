package steps

import (
	"context"
	"os/exec"
)

// NodePMs enables yarn via Corepack and installs pnpm into the mise-managed
// Node LTS. All commands run through `mise exec --` so they hit the same
// node/npm/corepack that NodeLTS just installed — at this point in the
// pipeline the user's shell hasn't been reloaded yet, so mise shims aren't
// on PATH for the iqdev process.
type NodePMs struct{}

func (NodePMs) Name() string { return "Yarn (Corepack) and pnpm" }

func (NodePMs) Check(ctx context.Context) (bool, error) {
	if _, err := exec.LookPath("mise"); err != nil {
		return false, nil
	}
	if exec.CommandContext(ctx, "mise", "exec", "--", "which", "yarn").Run() != nil {
		return false, nil
	}
	if exec.CommandContext(ctx, "mise", "exec", "--", "which", "pnpm").Run() != nil {
		return false, nil
	}
	return true, nil
}

func (NodePMs) Run(ctx context.Context, out chan<- string) error {
	out <- "Enabling yarn via Corepack"
	if err := runCmd(ctx, out, "mise", "exec", "--", "corepack", "enable", "yarn"); err != nil {
		return err
	}
	out <- "Installing pnpm globally via npm"
	return runCmd(ctx, out, "mise", "exec", "--", "npm", "install", "-g", "pnpm")
}
