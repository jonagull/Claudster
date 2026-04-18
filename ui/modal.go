package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

func renderModal(m Model) string {
	var title, fieldLabel, hint string

	switch m.modal.mode {
	case modalNewProject:
		title = "New Project"
		fieldLabel = "Group:"
		hint = "a template opens in " + editorName() + " — fill in name & repos there"

	case modalNewSession:
		title = fmt.Sprintf("New Session — %s", m.modal.targetProject)
		fieldLabel = "Session name:"
		hint = "starts in " + primaryRepoHint(m)
		if m.dangerousMode {
			hint += "  " + ErrorStyle.Render("⚠ --dangerously-skip-permissions")
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		OverlayTitle.Render(title),
		"",
		PreviewKey.Render(fieldLabel),
		HelpDesc.Render(hint),
		InputStyle.Width(46).Render(m.modal.input.View()),
		"",
		HelpDesc.Render("enter  confirm    esc  cancel"),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		OverlayStyle.Render(body),
	)
}

func editorName() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "nvim"
}

func primaryRepoHint(m Model) string {
	for _, g := range m.config.Groups {
		if g.Name == m.modal.targetGroup {
			for _, p := range g.Projects {
				if p.Name == m.modal.targetProject {
					return p.PrimaryRepo()
				}
			}
		}
	}
	return "primary repo"
}
