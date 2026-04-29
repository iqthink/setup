package ui

import "github.com/charmbracelet/lipgloss"

var (
	Red    = lipgloss.Color("196")
	Green  = lipgloss.Color("46")
	Yellow = lipgloss.Color("214")
	Purple = lipgloss.Color("135")
	Gray   = lipgloss.Color("240")
	White  = lipgloss.Color("255")

	Title = lipgloss.NewStyle().Bold(true).Foreground(Red)
	Hint  = lipgloss.NewStyle().Foreground(Gray)

	StepRunning = lipgloss.NewStyle().Foreground(Yellow)
	StepDone    = lipgloss.NewStyle().Foreground(Green)
	StepSkipped = lipgloss.NewStyle().Foreground(Gray)
	StepFailed  = lipgloss.NewStyle().Foreground(Red)
	StepPending = lipgloss.NewStyle().Foreground(Gray)

	LogLine = lipgloss.NewStyle().Foreground(Gray)

	BoxDone = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(Green).
		Padding(1, 4)

	DoneHeader  = lipgloss.NewStyle().Bold(true).Foreground(Green)
	StepHeader  = lipgloss.NewStyle().Bold(true).Foreground(White)
	StepNumber  = lipgloss.NewStyle().Bold(true).Foreground(Purple)
	StepLabel   = lipgloss.NewStyle().Foreground(White)
	Command     = lipgloss.NewStyle().Bold(true).Foreground(Yellow)
	Note        = lipgloss.NewStyle().Italic(true).Foreground(Gray)

	BoxFail = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(Red).
		Foreground(Red).
		Bold(true).
		Padding(1, 4)

	LogBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Gray).
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
