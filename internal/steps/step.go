package steps

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"
)

type Step interface {
	Name() string
	Check(ctx context.Context) (done bool, err error)
	Run(ctx context.Context, out chan<- string) error
}

// All returns the canonical step pipeline.
//
// Ordering matters: brew packages must be available before mise/node/orbstack
// configure on top of them. Team apps (iqthink baseline) run after the core
// Rails dev packages, before the language-specific steps that the team apps
// don't depend on.
func All() []Step {
	return []Step{
		XcodeCLT{},
		BrewPackages{},
		TeamAppsMust{},
		TeamAppsOptional{},
		MiseConfig{},
		NodeLTS{},
		NodePMs{},
		OrbStack{},
		ShellConfig{},
	}
}

// tapsIn returns the distinct "<user>/<tap>" prefixes of every tap-qualified
// entry in pkgs (e.g. "stripe/stripe-cli/stripe" → "stripe/stripe-cli"),
// preserving declaration order. Shared by BrewPackages and TeamApps so a
// fresh Mac doesn't trip on Homebrew's untrusted-tap warning.
func tapsIn(pkgs []string) []string {
	seen := map[string]bool{}
	var taps []string
	for _, p := range pkgs {
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

// runCmd executes name+args, streaming merged stdout+stderr line-by-line into out.
func runCmd(ctx context.Context, out chan<- string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		_ = pw.Close()
		return err
	}

	waitErr := make(chan error, 1)
	go func() {
		err := cmd.Wait()
		_ = pw.Close()
		waitErr <- err
	}()

	scanner := bufio.NewScanner(pr)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			_ = cmd.Process.Kill()
			<-waitErr
			return ctx.Err()
		case out <- scanner.Text():
		}
	}
	return <-waitErr
}
