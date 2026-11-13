// Package profile reads and writes ~/.iqdev-team-profile, the per-Mac
// record of which optional iqthink team apps the user opted into.
//
// File format is one package name per line. Lines starting with `#` and
// blank lines are ignored. Order is preserved so re-pick keeps the user's
// intent stable.
//
// The profile is per-Mac (not synced) because optional choices vary by
// device — a developer might want Figma on the design machine and not on
// the laptop they pair on.
package profile

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const fileName = ".iqdev-team-profile"

// Path returns the absolute path to the profile file.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, fileName), nil
}

// Exists reports whether the profile file is present on disk.
func Exists() bool {
	p, err := Path()
	if err != nil {
		return false
	}
	_, err = os.Stat(p)
	return err == nil
}

// Load reads the profile and returns the selected package names. A missing
// file returns ([], nil) — the caller distinguishes "no file" from "empty
// file" via Exists.
func Load() ([]string, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var picks []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		picks = append(picks, line)
	}
	return picks, sc.Err()
}

// Save writes the selected package names to disk, atomically replacing
// any existing file. Empty picks still writes a file so future runs know
// the user has been through the picker.
func Save(picks []string) error {
	p, err := Path()
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	fmt.Fprintln(w, "# iqdev team-apps profile — auto-managed by `iqdev pick`.")
	fmt.Fprintln(w, "# One Homebrew package name per line. Edit by hand or re-run `iqdev pick`.")
	for _, pick := range picks {
		fmt.Fprintln(w, pick)
	}
	if err := w.Flush(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
