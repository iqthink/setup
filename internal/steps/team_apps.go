package steps

import (
	"context"

	"github.com/iqthink/setup/internal/brewenv"
	"github.com/iqthink/setup/internal/profile"
)

// TeamAppsMustPackages is the always-on iqthink team baseline: apps every
// dev uses for day-to-day work. Installed on every Mac that runs iqdev.
//
// This list scopes iqdev beyond "Rails dev stack" into "iqthink Mac
// baseline" — what makes a brand-new machine usable for our team
// (1Password for auth, Basecamp for PM, Zoom for meetings, Claude for AI,
// Ghostty for terminal, etc.).
//
// Per-developer preferences belong in TeamAppsOptionalCandidates and are
// installed only when the user opts in via `iqdev pick`.
//
// Cask tokens are deliberate:
//   - "1password" is the GUI; the CLI is already installed plain by
//     BrewPackages as "1password-cli" per 1Password's own docs.
//   - "basecamp/tap/basecamp-cli" is the official Basecamp tap CLI.
//   - "thebrowsercompany-dia" is the namespaced cask for Dia browser.
//   - "tablepro" is the team's preferred DB GUI (we used to default to
//     TablePlus; TablePro is the current pick).
var TeamAppsMustPackages = []string{
	"1password",                 // GUI; 1password-cli ships from BrewPackages
	"basecamp",                  // project management
	"basecamp/tap/basecamp-cli", // Basecamp CLI for scripting + sync
	"zoom",                      // meetings
	"claude",                    // Claude desktop
	"conductor",                 // parallel AI agents
	"ghostty",                   // terminal
	"tablepro",                  // database GUI
	"bruno",                     // local API client
	"cleanshot",                 // screenshots
	"thebrowsercompany-dia",     // Dia browser (AI-first)
}

// TeamAppsMust installs every package in TeamAppsMustPackages.
type TeamAppsMust struct{}

func (TeamAppsMust) Name() string { return "iqthink team apps (baseline)" }

func (TeamAppsMust) Check(ctx context.Context) (bool, error) {
	for _, p := range TeamAppsMustPackages {
		if !installed(p) {
			return false, nil
		}
	}
	return true, nil
}

func (TeamAppsMust) Run(ctx context.Context, out chan<- string) error {
	for _, tap := range tapsIn(TeamAppsMustPackages) {
		if err := runCmd(ctx, out, brewenv.BrewPath(), "trust", tap); err != nil {
			return err
		}
	}
	args := []string{"install"}
	for _, p := range TeamAppsMustPackages {
		if !installed(p) {
			args = append(args, p)
		}
	}
	if len(args) == 1 {
		return nil
	}
	return runCmd(ctx, out, brewenv.BrewPath(), args...)
}

// TeamAppsOptionalCandidate is one entry in the team-apps picker. The Pkg
// is the brew package name; Label is the human-readable name shown in the
// TUI; Hint is the one-line description that helps a dev decide.
type TeamAppsOptionalCandidate struct {
	Pkg   string
	Label string
	Hint  string
}

// TeamAppsOptionalCandidates is the menu shown by `iqdev pick`. Every
// entry is something the team uses but not everyone needs on every Mac.
// Add/remove freely — install order in the brew batch follows this slice.
var TeamAppsOptionalCandidates = []TeamAppsOptionalCandidate{
	{Pkg: "raycast", Label: "Raycast", Hint: "Launcher / clipboard / window mgmt"},
	{Pkg: "chatgpt", Label: "ChatGPT", Hint: "ChatGPT desktop app"},
	{Pkg: "figma", Label: "Figma", Hint: "Design files (view + edit)"},
	{Pkg: "tailscale-app", Label: "Tailscale", Hint: "Internal-services VPN"},
	{Pkg: "gcloud-cli", Label: "gcloud CLI", Hint: "Google Cloud CLI"},
	{Pkg: "zed", Label: "Zed", Hint: "Editor — fast, native"},
	{Pkg: "github", Label: "GitHub Desktop", Hint: "GUI for non-CLI git tasks"},
	{Pkg: "screen-studio", Label: "Screen Studio", Hint: "Polished screen recordings"},
	{Pkg: "microsoft-office", Label: "Microsoft Office", Hint: "Word / Excel / PowerPoint"},
	{Pkg: "dropbox", Label: "Dropbox", Hint: "File sync"},
	{Pkg: "hey-desktop", Label: "HEY", Hint: "Email (Basecamp / HEY)"},
	{Pkg: "whatsapp", Label: "WhatsApp", Hint: "Chat"},
}

// TeamAppsOptional installs whatever the user opted into via the picker.
// It reads ~/.iqdev-team-profile and installs each listed package that
// isn't already present.
type TeamAppsOptional struct{}

func (TeamAppsOptional) Name() string { return "iqthink team apps (optional, picked)" }

func (TeamAppsOptional) Check(ctx context.Context) (bool, error) {
	picks, err := profile.Load()
	if err != nil {
		return false, err
	}
	if len(picks) == 0 {
		// No profile OR profile is empty — nothing to do; treat as done so
		// the step renders as a skip rather than a no-op work item.
		return true, nil
	}
	for _, p := range picks {
		if !installed(p) {
			return false, nil
		}
	}
	return true, nil
}

func (TeamAppsOptional) Run(ctx context.Context, out chan<- string) error {
	picks, err := profile.Load()
	if err != nil {
		return err
	}
	if len(picks) == 0 {
		return nil
	}
	for _, tap := range tapsIn(picks) {
		if err := runCmd(ctx, out, brewenv.BrewPath(), "trust", tap); err != nil {
			return err
		}
	}
	args := []string{"install"}
	for _, p := range picks {
		if !installed(p) {
			args = append(args, p)
		}
	}
	if len(args) == 1 {
		return nil
	}
	return runCmd(ctx, out, brewenv.BrewPath(), args...)
}
