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
