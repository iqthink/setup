package steps

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/iqthink/setup/internal/brewenv"
)

type OrbStack struct{}

func (OrbStack) Name() string { return "OrbStack (Docker)" }

func (OrbStack) Check(ctx context.Context) (bool, error) {
	if _, err := exec.LookPath("docker"); err != nil {
		return false, nil
	}
	return exec.CommandContext(ctx, "docker", "info").Run() == nil, nil
}

func (OrbStack) Run(ctx context.Context, out chan<- string) error {
	if !brewenv.CaskInstalled("orbstack") {
		if err := runCmd(ctx, out, brewenv.BrewPath(), "install", "--cask", "orbstack"); err != nil {
			return err
		}
	}
	out <- "Launching OrbStack..."
	_ = exec.CommandContext(ctx, "open", "/Applications/OrbStack.app").Run()
	out <- ""
	out <- "  >> ACTION REQUIRED in the OrbStack window <<"
	out <- ""
	out <- "  1. If macOS asks to install Rosetta, click \"Install\""
	out <- "     (required on Apple Silicon to run x86 containers)."
	out <- "  2. In OrbStack's onboarding, choose \"Docker\""
	out <- "     as the engine — that's what we wait for below."
	out <- "  3. Approve any system permissions OrbStack requests."
	out <- ""
	out <- "Waiting for Docker to respond — we'll continue automatically once it's ready."

	deadline := time.Now().Add(5 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	nudge := time.NewTicker(45 * time.Second)
	defer nudge.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-nudge.C:
			out <- "Still waiting for Docker — finish the OrbStack onboarding (Rosetta + Docker engine)."
			continue
		case <-ticker.C:
		}
		if exec.CommandContext(ctx, "docker", "info").Run() == nil {
			out <- "Docker ready."
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("Docker did not respond within 5 minutes — finish the OrbStack onboarding (install Rosetta, choose Docker engine) and re-run iqdev")
		}
	}
}
