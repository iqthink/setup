package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/iqthink/setup/internal/ui"
)

// sudoersDropin is the temporary rule that lets the pipeline's installers run
// `sudo` without prompting. It is created after one up-front password check and
// removed when iqdev exits (and defensively at the start of the next run).
//
// sudoers.d filenames must not contain a dot, hence plain "iqdev".
const sudoersDropin = "/etc/sudoers.d/iqdev"

// askpassPwEnv holds the password for the short-lived `sudo -A` helper calls
// that verify the password and write the sudoers rule. It is set ONLY on those
// specific commands' environments — never on this process's env — so it is
// never inherited by brew or any cask installer.
const askpassPwEnv = "IQDEV_SUDO_PW"

// primeSudo obtains admin rights ONCE, up front, while stdin is still a real
// terminal — before the alt-screen TUI grabs the keyboard in raw mode.
//
// Background: several installers (pkg-based casks like Zoom, OrbStack's
// privileged helper) call `sudo`. If that prompt fired inside the TUI, sudo and
// BubbleTea would fight over the same terminal — the symptom was "password
// incorrect" on a correct password, then a permanent hang. steps.runCmd now
// detaches the controlling terminal (setsid) so sudo can never do that; the
// cost is that a detached `sudo` can't prompt at all. So we authorize once here
// and install a temporary NOPASSWD rule that lets those installers proceed
// silently for the duration of the run.
//
// Non-fatal by design: if the user can't or won't authorize, we warn and return.
// The pipeline still runs; steps needing root fail fast (no hang) and are listed
// in the end-of-run summary rather than blocking everything else.
//
// The returned cleanup removes the sudoers rule and the askpass helper; call it
// on exit.
func primeSudo(ctx context.Context) (cleanup func()) {
	fmt.Println()
	fmt.Println(ui.StepHeader.Render("Administrator access"))
	fmt.Println(ui.Note.Render("  A few installers (e.g. Zoom, OrbStack) need your Mac login password —"))
	fmt.Println(ui.Note.Render("  the one you unlock this Mac with. Enter it once here and iqdev handles"))
	fmt.Println(ui.Note.Render("  the rest; nothing will prompt mid-install. Typing is invisible."))
	fmt.Println()

	script, err := writeAskpass()
	if err != nil {
		warn("could not set up sudo helper: " + err.Error())
		warn("continuing without admin rights; apps needing them will be reported.")
		return func() {}
	}

	// Remove a rule left behind by a previous run that was force-killed before
	// its cleanup ran. Best-effort: if there's nothing to remove or we lack
	// rights yet, it's a no-op.
	_ = removeDropin(ctx, "", script)

	cleanup = func() {
		_ = removeDropin(ctx, "", script)
		_ = os.RemoveAll(filepath.Dir(script))
	}

	pw, ok := promptAndVerify(ctx, script)
	if !ok {
		warn("skipping admin access; apps needing it will be reported at the end.")
		return cleanup
	}

	if err := installDropin(ctx, pw, script); err != nil {
		warn("could not enable passwordless install: " + err.Error())
		warn("apps needing admin rights may be skipped and reported.")
	}
	return cleanup
}

// promptAndVerify reads the password (echo off) and verifies it against sudo,
// up to three times, all in the plain terminal so a typo gets a clean retry
// instead of a mangled prompt inside the TUI. Returns the verified password.
func promptAndVerify(ctx context.Context, script string) (string, bool) {
	for attempt := 1; attempt <= 3; attempt++ {
		pw, err := readPassword("  Password: ")
		if err != nil {
			warn("could not read password.")
			return "", false
		}
		// -A: read via our askpass helper. -k: ignore any cached credential so
		// we actually test THIS password. Fails fast on a wrong password.
		if runSudo(ctx, pw, script, "-A", "-k", "-v") == nil {
			return pw, true
		}
		if attempt < 3 {
			fmt.Println(ui.Note.Render("  Sorry, that password was incorrect. Try again."))
		}
	}
	return "", false
}

// installDropin writes and validates the temporary NOPASSWD rule.
func installDropin(ctx context.Context, pw, script string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	content := fmt.Sprintf("%s ALL=(ALL) NOPASSWD: ALL\n", u.Username)

	tmp := filepath.Join(filepath.Dir(script), "sudoers")
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return err
	}
	// Validate syntax before we ever put it in place — a malformed sudoers file
	// must never reach /etc/sudoers.d.
	if out, err := exec.CommandContext(ctx, "visudo", "-cf", tmp).CombinedOutput(); err != nil {
		return fmt.Errorf("sudoers syntax check failed: %s", strings.TrimSpace(string(out)))
	}
	if err := runSudo(ctx, pw, script, "-A", "install", "-m", "0440", "-o", "root", "-g", "wheel", tmp, sudoersDropin); err != nil {
		return err
	}
	// Belt-and-suspenders: confirm the whole sudoers config still parses; if
	// not, pull our file back out immediately.
	if err := runSudo(ctx, pw, script, "-A", "visudo", "-c"); err != nil {
		_ = removeDropin(ctx, pw, script)
		return fmt.Errorf("sudoers validation failed after install")
	}
	return nil
}

// removeDropin deletes the temporary rule. pw may be "" once the NOPASSWD rule
// is active (then -n suffices); we pass -A too so it also works before/without.
func removeDropin(ctx context.Context, pw, script string) error {
	if _, err := os.Stat(sudoersDropin); err != nil {
		// Can't stat as non-root reliably; try removal regardless but don't
		// treat a missing file as an error.
	}
	return runSudo(ctx, pw, script, "-A", "rm", "-f", sudoersDropin)
}

// runSudo runs `sudo <args>` with the askpass helper wired up and the password
// injected only into THIS command's environment (never the parent's, so brew
// children never see it).
func runSudo(ctx context.Context, pw, script string, args ...string) error {
	cmd := exec.CommandContext(ctx, "sudo", args...)
	cmd.Env = append(os.Environ(), "SUDO_ASKPASS="+script, askpassPwEnv+"="+pw)
	// No stdio: askpass supplies the password; we don't want sudo touching a
	// terminal. Discard output — callers only care about success/failure.
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	return cmd.Run()
}

// writeAskpass drops a helper script into a private temp dir. It holds no
// secret — it just echoes whatever password is in the environment when sudo
// execs it (which runSudo sets per-call).
func writeAskpass() (string, error) {
	dir, err := os.MkdirTemp("", "iqdev-sudo")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "askpass.sh")
	body := "#!/bin/sh\nprintf '%s\\n' \"$" + askpassPwEnv + "\"\n"
	if err := os.WriteFile(path, []byte(body), 0o700); err != nil {
		return "", err
	}
	return path, nil
}

// readPassword reads a line from the terminal with echo disabled, restoring the
// terminal afterward. Uses stty so we don't take on a new dependency.
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	stty := func(arg string) {
		c := exec.Command("stty", arg)
		c.Stdin = os.Stdin
		_ = c.Run()
	}
	stty("-echo")
	defer func() {
		stty("echo")
		fmt.Println()
	}()

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func warn(msg string) {
	fmt.Fprintln(os.Stderr, ui.Note.Render("  "+msg))
}
