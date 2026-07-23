package tui

import (
	"context"
	"time"

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
	stateConfig        // picker for optional team apps
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
	stepStart time.Time // when the current running step started, for elapsed display

	// failures accumulates every step that failed. The pipeline no longer
	// aborts on the first failure: it records each one, keeps going, and the
	// run ends on a summary screen listing them all.
	failures []stepFailure

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

// stepFailure records a step that failed, with the tail of its output so the
// summary can show what actually went wrong.
type stepFailure struct {
	idx  int
	err  error
	tail []string
}

// Failure is the exported view of a failed step, for the post-exit summary.
type Failure struct {
	Step string
	Err  error
	Tail []string
}

// Failed reports whether the pipeline ended with at least one failed step.
func (m Model) Failed() bool { return len(m.failures) > 0 }

// Failures returns every step that failed, in pipeline order.
func (m Model) Failures() []Failure {
	out := make([]Failure, 0, len(m.failures))
	for _, f := range m.failures {
		name := ""
		if f.idx < len(m.steps) {
			name = m.steps[f.idx].Name()
		}
		out = append(out, Failure{Step: name, Err: f.err, Tail: f.tail})
	}
	return out
}
