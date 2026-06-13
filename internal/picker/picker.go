// Package picker provides a reusable BubbleTea multi-select list.
//
// It is used in two places:
//   - As a sub-screen in the main iqdev TUI, between the welcome and the
//     install pipeline, so a fresh-Mac user picks optional team apps before
//     anything starts running.
//   - Standalone, as the entire TUI for the `iqdev pick` subcommand, so a
//     dev can re-pick later without re-running the whole pipeline.
//
// Picker is intentionally minimal — no fuzzy search, no scrolling
// indicators. The team-apps menu is short (a dozen items) so a flat list
// fits in a terminal. If the menu grows we'll add viewport scrolling.
package picker

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iqthink/setup/internal/ui"
)

// Option is one row in the picker.
type Option struct {
	Key   string // stable identifier — for iqdev, the brew package name
	Label string // human-readable name shown in the list
	Hint  string // optional one-liner shown beside the label
}

// Model is a BubbleTea model. Use New to construct; embed in a parent
// model or run directly via tea.NewProgram.
type Model struct {
	title    string
	options  []Option
	selected map[string]bool
	cursor   int

	confirmed bool
	cancelled bool
}

// New builds a Model. initiallySelected may be nil; matching Option.Key
// values are pre-checked.
func New(title string, options []Option, initiallySelected []string) Model {
	sel := map[string]bool{}
	for _, k := range initiallySelected {
		sel[k] = true
	}
	return Model{title: title, options: options, selected: sel}
}

func (m Model) Init() tea.Cmd { return nil }

// Confirmed reports whether the user pressed enter to accept the current
// selection. Cancelled reports whether the user pressed esc or q to bail
// without saving.
func (m Model) Confirmed() bool { return m.confirmed }
func (m Model) Cancelled() bool { return m.cancelled }

// Selected returns the Keys of every checked option, in declaration order.
func (m Model) Selected() []string {
	var out []string
	for _, o := range m.options {
		if m.selected[o.Key] {
			out = append(out, o.Key)
		}
	}
	return out
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "ctrl+c", "esc", "q":
		m.cancelled = true
		return m, tea.Quit
	case "enter":
		m.confirmed = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.options)-1 {
			m.cursor++
		}
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		m.cursor = len(m.options) - 1
	case " ", "x":
		if len(m.options) > 0 {
			k := m.options[m.cursor].Key
			m.selected[k] = !m.selected[k]
		}
	case "a":
		for _, o := range m.options {
			m.selected[o.Key] = true
		}
	case "n":
		for _, o := range m.options {
			m.selected[o.Key] = false
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString("\n  ")
	b.WriteString(ui.Title.Render(m.title))
	b.WriteString("\n\n")
	for i, o := range m.options {
		b.WriteString(m.row(i, o))
		b.WriteString("\n")
	}
	b.WriteString("\n  ")
	b.WriteString(ui.Hint.Render(
		fmt.Sprintf("%d selected · ↑/↓ move · space toggle · a all · n none · enter confirm · esc cancel",
			len(m.Selected())),
	))
	b.WriteString("\n")
	return b.String()
}

func (m Model) row(i int, o Option) string {
	cursor := "  "
	if i == m.cursor {
		cursor = ui.StepRunning.Render("▸ ")
	}
	box := ui.GlyphPending
	style := ui.StepPending
	if m.selected[o.Key] {
		box = ui.GlyphDone
		style = ui.StepDone
	}
	line := fmt.Sprintf("%s%s  %s", cursor, style.Render(box), o.Label)
	if o.Hint != "" {
		line += "  " + ui.Hint.Render("— "+o.Hint)
	}
	return "  " + line
}
