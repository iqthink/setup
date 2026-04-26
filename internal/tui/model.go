package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/iqthink/setup/internal/steps"
	"github.com/iqthink/setup/internal/ui"
)

type state int

const (
	stateWelcome state = iota
	stateRunning
	stateDone
	stateFailed
)

type stepStatus int

const (
	statusPending stepStatus = iota
	statusRunning
	statusDone
	statusSkipped
	statusFailed
)

type Model struct {
	steps    []steps.Step
	statuses []stepStatus
	state    state

	current   int
	failedIdx int
	failErr   error

	logLines []string

	spin spinner.Model
	vp   viewport.Model

	width  int
	height int

	ch     chan pipeMsg
	cancel context.CancelFunc
}

func New() Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = ui.StepRunning

	vp := viewport.New(0, 0)

	all := steps.All()
	return Model{
		steps:    all,
		statuses: make([]stepStatus, len(all)),
		state:    stateWelcome,
		spin:     sp,
		vp:       vp,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spin.Tick
}

// Failed reports whether the pipeline ended in a failed state.
func (m Model) Failed() bool { return m.state == stateFailed }

// FailedStep returns the name of the step that failed, or empty.
func (m Model) FailedStep() string {
	if m.state != stateFailed || m.failedIdx >= len(m.steps) {
		return ""
	}
	return m.steps[m.failedIdx].Name()
}

// FailErr returns the error from the failed step, if any.
func (m Model) FailErr() error { return m.failErr }
