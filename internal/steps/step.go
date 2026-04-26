package steps

import (
	"bufio"
	"context"
	"io"
	"os/exec"
)

type Step interface {
	Name() string
	Check(ctx context.Context) (done bool, err error)
	Run(ctx context.Context, out chan<- string) error
}

// All returns the canonical step pipeline.
func All() []Step {
	return []Step{
		XcodeCLT{},
		BrewPackages{},
		OrbStack{},
		ShellConfig{},
	}
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
