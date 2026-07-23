package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/iqthink/setup/internal/steps"
	"github.com/iqthink/setup/internal/ui"
)

func (m Model) View() string {
	switch m.state {
	case stateWelcome:
		return m.welcomeView()
	case stateConfig:
		return m.pick.View()
	case stateRunning:
		return m.runningView()
	case stateDone:
		return m.doneView()
	case stateFailed:
		return m.failedView()
	}
	return ""
}

func (m Model) welcomeView() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(ui.Title.Render(ui.Logo))
	b.WriteString("\n\n")
	b.WriteString(ui.Hint.Render("  Global Mac bootstrap for iqthink Rails apps."))
	b.WriteString("\n\n")
	b.WriteString("  We'll install (skipping anything already in place):\n")
	for _, s := range m.steps {
		b.WriteString(fmt.Sprintf("    %s %s\n", ui.StepPending.Render(ui.GlyphPending), s.Name()))
	}
	b.WriteString("\n")
	b.WriteString(ui.Note.Render("  Heads-up: a couple of steps open macOS windows (Xcode tools dialog,"))
	b.WriteString("\n")
	b.WriteString(ui.Note.Render("  OrbStack) and wait for a click — watch for >> ACTION REQUIRED <<."))
	b.WriteString("\n\n")
	b.WriteString(ui.Hint.Render("  Press Enter to start · Ctrl+C to quit"))
	b.WriteString("\n")
	return b.String()
}

func (m Model) runningView() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(ui.Title.Render("iqdev"))
	b.WriteString("\n\n")
	for i, s := range m.steps {
		b.WriteString("  ")
		b.WriteString(m.stepLine(i, s))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if m.vp.Width > 0 && m.statuses[m.current] == statusRunning {
		b.WriteString("  ")
		b.WriteString(ui.LogBox.Width(m.vp.Width).Render(m.vp.View()))
		b.WriteString("\n")
		b.WriteString(ui.Note.Render("  Large downloads (e.g. OrbStack ~1 GB) run silently — the timer above means it's still working."))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(ui.Hint.Render("  Ctrl+C to cancel"))
	b.WriteString("\n")
	return b.String()
}

// elapsed renders a compact "1m23s" / "45s" duration.
func elapsed(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d / time.Minute)
	s := int((d % time.Minute) / time.Second)
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func (m Model) stepLine(i int, s steps.Step) string {
	glyph := ui.GlyphPending
	style := ui.StepPending
	switch m.statuses[i] {
	case statusRunning:
		glyph = m.spin.View()
		style = ui.StepRunning
		if !m.stepStart.IsZero() {
			return fmt.Sprintf("%s  %s %s", style.Render(glyph), s.Name(),
				ui.Note.Render("("+elapsed(time.Since(m.stepStart))+")"))
		}
	case statusDone:
		glyph = ui.GlyphDone
		style = ui.StepDone
	case statusSkipped:
		glyph = ui.GlyphSkipped
		style = ui.StepSkipped
	case statusFailed:
		glyph = ui.GlyphFailed
		style = ui.StepFailed
	}
	return fmt.Sprintf("%s  %s", style.Render(glyph), s.Name())
}

// NextStepsBody is the single source of truth for the post-install
// checklist. It renders twice: inside the TUI's done box, and printed to
// the plain terminal by main.go after the alt-screen closes (so it stays
// in the scrollback). Edit it here and both stay in sync.
//
// Step 1 is deliberately spelled out sub-step by sub-step: enabling
// "Integrate with 1Password CLI" is a GUI-only toggle that `op signin`
// gives no useful hint about, and it is the single most-missed action on
// a fresh Mac.
func NextStepsBody() string {
	num := func(n string) string { return ui.StepNumber.Render(n + ".") }
	label := func(s string) string { return ui.StepLabel.Render(s) }
	cmd := func(s string) string { return ui.Command.Render(s) }
	note := func(s string) string { return ui.Note.Render(s) }

	lines := []string{
		ui.StepHeader.Render("Next steps:"),
		"",
		"  " + num("1") + " " + label("Enable the 1Password CLI") + "  " + note("(one-time — easy to miss!)"),
		"       " + label("a) Open the 1Password app and sign in."),
		"       " + label("b) In 1Password: Settings → Developer →"),
		"          " + label("turn ON \"Integrate with 1Password CLI\"."),
		"       " + label("c) Back in this terminal, verify:") + "  " + cmd("op signin"),
		"       " + note("Skipping a) or b) makes `op signin` fail or hang."),
		"",
		"  " + num("2") + " " + label("Authenticate the GitHub CLI:"),
		"       " + cmd("gh auth login"),
		"",
		"  " + num("3") + " " + label("Open a NEW terminal window (activates mise),"),
		"     " + label("then in your Rails app run:"),
		"       " + cmd("bin/setup"),
		"       " + note("No repo yet? First: gh repo clone iqthink/<app>"),
		"",
		"  " + num("4") + " " + label("Open your other apps and set them up (sign in,"),
		"     " + label("grant access). They installed but stay closed:"),
	}
	for _, l := range wrapJoin(steps.AppsToOpen(), " · ", 52) {
		lines = append(lines, "       "+label(l))
	}
	lines = append(lines,
		"       "+note("Find them in Applications — or Spotlight (⌘-Space)."),
		"",
		note("Some apps ask for macOS permissions on first launch"),
		note("(CleanShot → screen recording, Zoom → mic/camera,"),
		note("Tailscale → VPN). Click Allow / approve them in"),
		note("System Settings → Privacy & Security when asked."),
	)
	return strings.Join(lines, "\n")
}

// wrapJoin joins items with sep, wrapping to a new string whenever the current
// line would exceed width. Used to lay out the app list across a few lines
// inside the fixed-width done box.
func wrapJoin(items []string, sep string, width int) []string {
	var lines []string
	cur := ""
	for _, it := range items {
		switch {
		case cur == "":
			cur = it
		case len(cur)+len(sep)+len(it) > width:
			lines = append(lines, cur)
			cur = it
		default:
			cur += sep + it
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

func (m Model) doneView() string {
	body := ui.DoneHeader.Render(ui.GlyphDone+"  Done") + "\n\n" + NextStepsBody()
	return "\n" + ui.BoxDone.Render(body) + "\n\n" +
		ui.Hint.Render("  Press Enter to exit.") + "\n"
}

func (m Model) failedView() string {
	failures := m.Failures()

	var lines []string
	header := fmt.Sprintf("%s  Finished with %d issue(s)", ui.GlyphFailed, len(failures))
	lines = append(lines, header, "")
	lines = append(lines, ui.Note.Render("Everything else installed. The rest can be retried — iqdev is idempotent."), "")

	for _, f := range failures {
		lines = append(lines, ui.StepFailed.Render(ui.GlyphFailed+"  "+f.Step))
		if f.Err != nil {
			lines = append(lines, "     "+f.Err.Error())
		}
		if len(f.Tail) > 0 {
			tail := f.Tail
			if len(tail) > 6 {
				tail = tail[len(tail)-6:]
			}
			for _, t := range tail {
				lines = append(lines, ui.LogLine.Render("     "+t))
			}
		}
		lines = append(lines, "")
	}
	lines = append(lines, "Re-run `iqdev` to retry.")

	return "\n" + ui.BoxFail.Render(strings.Join(lines, "\n")) + "\n\n" +
		ui.Hint.Render("  Press Enter to exit.") + "\n"
}
