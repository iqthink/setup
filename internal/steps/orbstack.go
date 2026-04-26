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
	out <- "Launching OrbStack (it may ask for system permissions)."
	_ = exec.CommandContext(ctx, "open", "/Applications/OrbStack.app").Run()
	out <- "Waiting for Docker to respond..."

	deadline := time.Now().Add(3 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		if exec.CommandContext(ctx, "docker", "info").Run() == nil {
			out <- "Docker ready."
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("Docker did not respond within 3 minutes")
		}
	}
}
