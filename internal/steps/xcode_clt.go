package steps

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type XcodeCLT struct{}

func (XcodeCLT) Name() string { return "Xcode Command Line Tools" }

func (XcodeCLT) Check(ctx context.Context) (bool, error) {
	err := exec.CommandContext(ctx, "xcode-select", "-p").Run()
	return err == nil, nil
}

func (XcodeCLT) Run(ctx context.Context, out chan<- string) error {
	out <- "Opening the macOS \"Install Command Line Developer Tools\" dialog."
	out <- ""
	out <- "  >> ACTION REQUIRED in a macOS dialog <<"
	out <- ""
	out <- "  Click \"Install\", then \"Agree\". The dialog may open BEHIND this"
	out <- "  window — check the Dock or press Cmd+Tab if you don't see it."
	out <- ""
	out <- "The download takes 5-10 minutes; we continue automatically when it's done."

	// xcode-select --install opens the GUI installer and exits. It returns
	// non-zero if the tools are already installed; we ignore that.
	_ = exec.CommandContext(ctx, "xcode-select", "--install").Run()

	deadline := time.Now().Add(20 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		if exec.CommandContext(ctx, "xcode-select", "-p").Run() == nil {
			out <- "Xcode CLT ready."
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("Xcode CLT installation did not finish within 20 minutes — if the dialog was closed, re-run iqdev to reopen it")
		}
	}
}
