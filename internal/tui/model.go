package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/iqthink/setup/internal/picker"
	"github.com/iqthink/setup/internal/profile"
	"github.com/iqthink/setup/internal/steps"
	"github.com/iqthink/setup/internal/ui"
)

type state int

const (
	stateWelcome state = iota
	stateConfig // picker for optional team apps
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

	pick picker.Model

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

	// Pre-load the picker with the team-apps optional menu so the user can
	// dive straight in without us building it on a state transition. If a
	// profile already exists, the welcome→picker step is skipped, but the
	// picker is still wired so `iqdev pick` (which uses the same model) is
	// only one screen away.
	opts := make([]picker.Option, 0, len(steps.TeamAppsOptionalCandidates))
	for _, c := range steps.TeamAppsOptionalCandidates {
		opts = append(opts, picker.Option{Key: c.Pkg, Label: c.Label, Hint: c.Hint})
	}
	initial, _ := profile.Load()

	return Model{
		steps:    all,
		statuses: make([]stepStatus, len(all)),
		state:    stateWelcome,
		spin:     sp,
		vp:       vp,
		pick:     picker.New("Optional iqthink team apps", opts, initial),
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
