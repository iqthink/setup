package tui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/iqthink/setup/internal/picker"
	"github.com/iqthink/setup/internal/profile"
	"github.com/iqthink/setup/internal/steps"
)

type pipeKind int

const (
	kindStarted pipeKind = iota
	kindSkipped
	kindLine
	kindDone
	kindFailed
)

type pipeMsg struct {
	kind pipeKind
	idx  int
	line string
	err  error
	tail []string
}

type pipeFinishedMsg struct{}

// startPipeline transitions from welcome (or picker) into the running
// state and kicks off the install goroutine. Extracted so both the
// "profile exists → skip picker" path and the "picker confirmed" path
// share one entry point.
func (m Model) startPipeline() (tea.Model, tea.Cmd) {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.ch = make(chan pipeMsg, 64)
	m.state = stateRunning
	go runPipeline(ctx, m.steps, m.ch)
	return m, tea.Batch(m.spin.Tick, m.waitForPipe())
}

func runPipeline(ctx context.Context, all []steps.Step, ch chan<- pipeMsg) {
	defer close(ch)
	for i, st := range all {
		done, err := st.Check(ctx)
		if err != nil {
			// A Check error is unusual; record it and move on rather than
			// aborting the whole run for it.
			ch <- pipeMsg{kind: kindFailed, idx: i, err: err}
			continue
		}
		if done {
			ch <- pipeMsg{kind: kindSkipped, idx: i}
			continue
		}
		ch <- pipeMsg{kind: kindStarted, idx: i}

		out := make(chan string, 64)
		runErr := make(chan error, 1)
		go func() {
			runErr <- st.Run(ctx, out)
			close(out)
		}()
		// Keep the last lines of output so a failure summary can show what
		// actually went wrong, not just the exit status.
		var tail []string
		for line := range out {
			ch <- pipeMsg{kind: kindLine, idx: i, line: line}
			tail = append(tail, line)
			if len(tail) > 20 {
				tail = tail[len(tail)-20:]
			}
		}
		if e := <-runErr; e != nil {
			// User-initiated cancellation stops everything; anything else is
			// recorded and the pipeline continues to the next step.
			if ctx.Err() != nil {
				ch <- pipeMsg{kind: kindFailed, idx: i, err: e}
				return
			}
			ch <- pipeMsg{kind: kindFailed, idx: i, err: e, tail: append([]string(nil), tail...)}
			continue
		}
		ch <- pipeMsg{kind: kindDone, idx: i}
	}
}

func (m Model) waitForPipe() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-m.ch
		if !ok {
			return pipeFinishedMsg{}
		}
		return msg
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		w := m.width - 4
		if w < 20 {
			w = 20
		}
		h := m.height - len(m.steps) - 10
		if h < 6 {
			h = 6
		}
		m.vp.Width = w
		m.vp.Height = h
		return m, nil

	case tea.KeyMsg:
		// In the picker state, hand every key to the picker submodel.
		// Quit/cancel inside the picker is handled by reading Cancelled
		// after Update — we never let ctrl+c bypass the picker mid-state
		// because the picker itself respects it.
		if m.state == stateConfig {
			updated, cmd := m.pick.Update(msg)
			m.pick = updated.(picker.Model)
			if m.pick.Confirmed() {
				if err := profile.Save(m.pick.Selected()); err != nil {
					// Profile-save failure shouldn't sink the run; the picks
					// stay in-memory for this pipeline via the loaded profile
					// reload in TeamAppsOptional, which will re-read disk and
					// install nothing if the save failed silently. We surface
					// the install gap by writing to the running log later.
					m.logLines = append(m.logLines, "warning: could not save ~/.iqdev-team-profile: "+err.Error())
				}
				return m.startPipeline()
			}
			if m.pick.Cancelled() {
				// User opted out of the picker: keep whatever was on disk
				// (or nothing) and run the pipeline.
				return m.startPipeline()
			}
			return m, cmd
		}
		switch msg.String() {
		case "ctrl+c", "q":
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		case "enter", " ":
			switch m.state {
			case stateWelcome:
				// First-run UX: show the picker so the user can select
				// optional team apps before the pipeline starts. On
				// subsequent runs (profile exists) skip straight to running
				// — no nagging.
				if !profile.Exists() {
					m.state = stateConfig
					return m, nil
				}
				return m.startPipeline()
			case stateDone, stateFailed:
				return m, tea.Quit
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case pipeMsg:
		switch msg.kind {
		case kindStarted:
			m.statuses[msg.idx] = statusRunning
			m.current = msg.idx
			m.stepStart = time.Now()
			m.logLines = nil
			m.vp.SetContent("")
		case kindSkipped:
			m.statuses[msg.idx] = statusSkipped
		case kindLine:
			m.logLines = append(m.logLines, msg.line)
			if len(m.logLines) > 200 {
				m.logLines = m.logLines[len(m.logLines)-200:]
			}
			m.vp.SetContent(strings.Join(m.logLines, "\n"))
			m.vp.GotoBottom()
		case kindDone:
			m.statuses[msg.idx] = statusDone
		case kindFailed:
			m.statuses[msg.idx] = statusFailed
			m.failures = append(m.failures, stepFailure{
				idx:  msg.idx,
				err:  msg.err,
				tail: msg.tail,
			})
			// Keep going — the summary at the end lists every failure.
		}
		return m, m.waitForPipe()

	case pipeFinishedMsg:
		if len(m.failures) > 0 {
			m.state = stateFailed
		} else {
			m.state = stateDone
		}
		return m, nil
	}

	return m, nil
}
