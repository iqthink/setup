package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ShellConfig struct{}

func (ShellConfig) Name() string { return "Shell config (PATH + mise)" }

func (s ShellConfig) Check(ctx context.Context) (bool, error) {
	rc, _, err := s.rcFile()
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(rc)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	text := string(data)
	return strings.Contains(text, "mise activate") &&
		strings.Contains(text, `$HOME/.local/bin`), nil
}

func (s ShellConfig) Run(ctx context.Context, out chan<- string) error {
	rc, shellName, err := s.rcFile()
	if err != nil {
		return err
	}
	out <- fmt.Sprintf("Editing %s", rc)

	existing, _ := os.ReadFile(rc)
	text := string(existing)

	f, err := os.OpenFile(rc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if !strings.Contains(text, `$HOME/.local/bin`) {
		fmt.Fprintln(f)
		fmt.Fprintln(f, "# iqdev: ensure ~/.local/bin is on PATH")
		fmt.Fprintln(f, `export PATH="$HOME/.local/bin:$PATH"`)
		out <- "Added PATH export for ~/.local/bin"
	}
	if !strings.Contains(text, "mise activate") {
		fmt.Fprintln(f)
		fmt.Fprintln(f, "# Activate mise (Ruby/Node version manager)")
		fmt.Fprintf(f, "eval \"$(mise activate %s)\"\n", shellName)
		out <- fmt.Sprintf("Added mise activation for %s", shellName)
	}
	return nil
}

func (ShellConfig) rcFile() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	sh := os.Getenv("SHELL")
	switch {
	case strings.HasSuffix(sh, "/zsh"):
		return filepath.Join(home, ".zshrc"), "zsh", nil
	case strings.HasSuffix(sh, "/bash"):
		return filepath.Join(home, ".bashrc"), "bash", nil
	default:
		return filepath.Join(home, ".zshrc"), "zsh", nil
	}
}
