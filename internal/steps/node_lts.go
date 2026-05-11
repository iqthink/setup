package steps

import (
	"context"
	"os/exec"
	"strings"
)

// NodeLTS installs the latest Node.js LTS via mise and pins it globally.
// `mise use --global node@lts` both installs (if missing) and writes the
// global config, so the Run is idempotent — but we still gate on Check
// so re-runs print "skipped" instead of re-resolving LTS over the network.
type NodeLTS struct{}

func (NodeLTS) Name() string { return "Node.js (latest LTS, global)" }

func (NodeLTS) Check(ctx context.Context) (bool, error) {
	if _, err := exec.LookPath("mise"); err != nil {
		return false, nil
	}
	out, err := exec.CommandContext(ctx, "mise", "ls", "--global", "node").Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func (NodeLTS) Run(ctx context.Context, out chan<- string) error {
	out <- "Installing Node.js LTS and pinning it globally"
	return runCmd(ctx, out, "mise", "use", "--global", "node@lts")
}
