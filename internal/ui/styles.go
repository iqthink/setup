package ui

import "github.com/charmbracelet/lipgloss"

// Adaptive palette: each color has a light-bg variant (darker shade for
// contrast) and a dark-bg variant (brighter shade). lipgloss picks based
// on the terminal's detected background, so the same styles render
// legibly on both themes.
var (
	primary = lipgloss.AdaptiveColor{Light: "232", Dark: "255"}
	dim     = lipgloss.AdaptiveColor{Light: "242", Dark: "240"}
	green   = lipgloss.AdaptiveColor{Light: "28", Dark: "46"}
	yellow  = lipgloss.AdaptiveColor{Light: "130", Dark: "214"}
	purple  = lipgloss.AdaptiveColor{Light: "91", Dark: "135"}
	red     = lipgloss.AdaptiveColor{Light: "124", Dark: "196"}

	Title = lipgloss.NewStyle().Bold(true).Foreground(red)
	Hint  = lipgloss.NewStyle().Foreground(dim)

	StepRunning = lipgloss.NewStyle().Foreground(yellow)
	StepDone    = lipgloss.NewStyle().Foreground(green)
	StepSkipped = lipgloss.NewStyle().Foreground(dim)
	StepFailed  = lipgloss.NewStyle().Foreground(red)
	StepPending = lipgloss.NewStyle().Foreground(dim)

	LogLine = lipgloss.NewStyle().Foreground(dim)

	BoxDone = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(green).
		Padding(1, 4)

	DoneHeader = lipgloss.NewStyle().Bold(true).Foreground(green)
	StepHeader = lipgloss.NewStyle().Bold(true).Foreground(primary)
	StepNumber = lipgloss.NewStyle().Bold(true).Foreground(purple)
	StepLabel  = lipgloss.NewStyle().Foreground(primary)
	Command    = lipgloss.NewStyle().Bold(true).Foreground(yellow)
	Note       = lipgloss.NewStyle().Italic(true).Foreground(dim)

	BoxFail = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(red).
		Foreground(red).
		Bold(true).
		Padding(1, 4)

	LogBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(dim).
		Padding(0, 1)
)

const (
	GlyphPending = "○"
	GlyphRunning = "◐"
	GlyphDone    = "✓"
	GlyphSkipped = "⤳"
	GlyphFailed  = "✗"
)

const Logo = `██╗ ██████╗ ██████╗ ███████╗██╗   ██╗
██║██╔═══██╗██╔══██╗██╔════╝██║   ██║
██║██║   ██║██║  ██║█████╗  ██║   ██║
██║██║▄▄ ██║██║  ██║██╔══╝  ╚██╗ ██╔╝
██║╚██████╔╝██████╔╝███████╗ ╚████╔╝
╚═╝ ╚══▀▀═╝ ╚═════╝ ╚══════╝  ╚═══╝ `
