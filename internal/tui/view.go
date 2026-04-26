package tui

import (
	"fmt"
	"strings"

	"github.com/iqthink/setup/internal/steps"
	"github.com/iqthink/setup/internal/ui"
)

func (m Model) View() string {
	switch m.state {
	case stateWelcome:
		return m.welcomeView()
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
	}
	b.WriteString("\n")
	b.WriteString(ui.Hint.Render("  Ctrl+C to cancel"))
	b.WriteString("\n")
	return b.String()
}

func (m Model) stepLine(i int, s steps.Step) string {
	glyph := ui.GlyphPending
	style := ui.StepPending
	switch m.statuses[i] {
	case statusRunning:
		glyph = m.spin.View()
		style = ui.StepRunning
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

func (m Model) doneView() string {
	body := strings.Join([]string{
		ui.GlyphDone + "  Done",
		"",
		"Now go to your Rails app and run:",
		"",
		"  bin/setup",
		"",
		"If this is your first time, close and reopen",
		"your terminal so mise activates.",
	}, "\n")
	return "\n" + ui.BoxDone.Render(body) + "\n\n" +
		ui.Hint.Render("  Press Enter to exit.") + "\n"
}

func (m Model) failedView() string {
	name := m.steps[m.failedIdx].Name()
	errMsg := ""
	if m.failErr != nil {
		errMsg = m.failErr.Error()
	}
	logTail := ""
	if len(m.logLines) > 0 {
		tail := m.logLines
		if len(tail) > 12 {
			tail = tail[len(tail)-12:]
		}
		logTail = "\n\n" + ui.LogLine.Render(strings.Join(tail, "\n"))
	}
	body := strings.Join([]string{
		ui.GlyphFailed + "  Failed: " + name,
		"",
		errMsg,
		"",
		"Re-run `iqdev` to retry (it's idempotent).",
	}, "\n") + logTail
	return "\n" + ui.BoxFail.Render(body) + "\n\n" +
		ui.Hint.Render("  Press Enter to exit.") + "\n"
}
