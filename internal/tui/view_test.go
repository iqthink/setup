package tui

import (
	"strings"
	"testing"
)

func TestWelcomeViewRenders(t *testing.T) {
	m := New()
	out := m.View()
	if !strings.Contains(out, "Press Enter") {
		t.Fatalf("welcome view missing prompt; got:\n%s", out)
	}
	if !strings.Contains(out, "Xcode Command Line Tools") {
		t.Fatalf("welcome view missing first step name; got:\n%s", out)
	}
}

// The 1Password CLI toggle is a GUI-only step that `op signin` gives no
// useful hint about; this guards against it ever dropping out of the
// post-install guidance (both the done screen and main.go's post-exit
// print render NextStepsBody).
func TestNextStepsCover1PasswordIntegration(t *testing.T) {
	out := NextStepsBody()
	for _, want := range []string{
		"Integrate with 1Password CLI",
		"op signin",
		"gh auth login",
		"bin/setup",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("next steps missing %q; got:\n%s", want, out)
		}
	}
}
