package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iqthink/setup/internal/picker"
	"github.com/iqthink/setup/internal/profile"
	"github.com/iqthink/setup/internal/steps"
	"github.com/iqthink/setup/internal/ui"
)

// runPicker drives a standalone TUI for `iqdev pick`. It reads the
// current profile, lets the user toggle optional team apps, and writes
// back the new selection. No install pipeline runs — re-pick is purely
// about updating the profile; the next plain `iqdev` will install or
// remove based on what changed.
func runPicker() error {
	opts := make([]picker.Option, 0, len(steps.TeamAppsOptionalCandidates))
	for _, c := range steps.TeamAppsOptionalCandidates {
		opts = append(opts, picker.Option{Key: c.Pkg, Label: c.Label, Hint: c.Hint})
	}
	current, _ := profile.Load()

	model := picker.New("Optional iqthink team apps", opts, current)
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	final, ok := finalModel.(picker.Model)
	if !ok {
		return nil
	}
	if final.Cancelled() {
		fmt.Println(ui.Hint.Render("  Picker cancelled — profile unchanged."))
		return nil
	}
	picks := final.Selected()
	if err := profile.Save(picks); err != nil {
		return fmt.Errorf("save profile: %w", err)
	}
	path, _ := profile.Path()
	fmt.Printf("  %s Saved %d pick(s) to %s\n",
		ui.DoneHeader.Render("✓"), len(picks), path)
	fmt.Println(ui.Note.Render("  Re-run `iqdev` to install (or remove) according to the new profile."))
	return nil
}
