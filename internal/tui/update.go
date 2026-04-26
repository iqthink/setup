package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

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
}

type pipeFinishedMsg struct{}

func runPipeline(ctx context.Context, all []steps.Step, ch chan<- pipeMsg) {
	defer close(ch)
	for i, st := range all {
		done, err := st.Check(ctx)
		if err != nil {
			ch <- pipeMsg{kind: kindFailed, idx: i, err: err}
			return
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
		for line := range out {
			ch <- pipeMsg{kind: kindLine, idx: i, line: line}
		}
		if e := <-runErr; e != nil {
			ch <- pipeMsg{kind: kindFailed, idx: i, err: e}
			return
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
		switch msg.String() {
		case "ctrl+c", "q":
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		case "enter", " ":
			switch m.state {
			case stateWelcome:
				ctx, cancel := context.WithCancel(context.Background())
				m.cancel = cancel
				m.ch = make(chan pipeMsg, 64)
				m.state = stateRunning
				go runPipeline(ctx, m.steps, m.ch)
				return m, tea.Batch(m.spin.Tick, m.waitForPipe())
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
			m.failedIdx = msg.idx
			m.failErr = msg.err
			m.state = stateFailed
			return m, nil
		}
		return m, m.waitForPipe()

	case pipeFinishedMsg:
		if m.state != stateFailed {
			m.state = stateDone
		}
		return m, nil
	}

	return m, nil
}
