package steps

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"

	"github.com/iqthink/setup/internal/brewenv"
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

// AnyPending reports whether any step still has work to do. main uses this
// to decide whether to prompt once for sudo before the alt-screen TUI starts:
// several installers (pkg-based casks like Zoom, OrbStack's privileged helper)
// shell out to `sudo`, which cannot prompt from inside the TUI. If everything
// is already installed, nothing is pending and we skip the sudo prompt so the
// idempotent re-run still finishes in seconds with no password ask.
// A Check error counts as pending so the TUI runs and surfaces it properly.
func AnyPending(ctx context.Context) bool {
	for _, s := range All() {
		done, err := s.Check(ctx)
		if err != nil || !done {
			return true
		}
	}
	return false
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

// brewInstallEach installs each missing package individually rather than in a
// single `brew install a b c` batch. A batch aborts on the first failing
// package, so one flaky cask would block every package after it; installing
// one at a time means a single failure is isolated, reported, and the rest
// still install. It returns a non-nil error naming the packages that failed
// (nil if all succeeded or were already present), so the caller/pipeline can
// surface a precise summary while continuing to the next step.
//
// --adopt: if a cask's app is already sitting in /Applications (installed
// manually by drag-and-drop), adopt it instead of failing with "there is
// already an App at ...". Harmless for formulae.
func brewInstallEach(ctx context.Context, out chan<- string, pkgs []string) error {
	var failed []string
	for _, p := range pkgs {
		if installed(p) {
			continue
		}
		if err := runCmd(ctx, out, brewenv.BrewPath(), "install", "--adopt", p); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			out <- fmt.Sprintf("  %s %s failed: %v", "✗", p, err)
			failed = append(failed, p)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("%d of %d failed: %s", len(failed), len(pkgs), strings.Join(failed, ", "))
	}
	return nil
}

// runCmd executes name+args, streaming merged stdout+stderr line-by-line into out.
//
// Setsid puts the child in its own session with no controlling terminal. This
// is deliberate: some installers (pkg-based casks, OrbStack's helper) shell out
// to `sudo`. Without a controlling terminal, sudo cannot pop a password prompt
// onto the tty that the BubbleTea TUI is reading in raw mode — which previously
// deadlocked the whole run. Instead sudo falls back to $SUDO_ASKPASS (primed
// before the TUI starts) when the credential isn't cached, or fails fast if no
// askpass is available. Either way it never hangs waiting for keystrokes that
// the TUI is intercepting.
func runCmd(ctx context.Context, out chan<- string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
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
