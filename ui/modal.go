package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

func renderModal(m Model) string {
	if m.modal.mode == modalConfirmDelete {
		return renderConfirmDelete(m)
	}

	var title, fieldLabel, hint string

	switch m.modal.mode {
	case modalNewProject:
		title = "New Project"
		fieldLabel = "Group:"
		hint = "tab to autocomplete  ·  template opens in " + editorName()

	case modalNewSession:
		title = fmt.Sprintf("New Session — %s", m.modal.targetProject)
		fieldLabel = "Session name:"
		hint = "starts in " + primaryRepoHint(m)
		if m.dangerousMode {
			hint += "  " + ErrorStyle.Render("⚠ --dangerously-skip-permissions")
		}

	case modalNewEditorSession:
		if m.modal.targetKind == "lazygit" {
			title = fmt.Sprintf("New Lazygit Session — %s", m.modal.targetProject)
			hint = "opens lazygit in " + primaryRepoHint(m)
		} else {
			title = fmt.Sprintf("New Editor Session — %s", m.modal.targetProject)
			hint = "opens " + editorName() + " in " + primaryRepoHint(m)
		}
		fieldLabel = "Session name:"
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

func renderConfirmDelete(m Model) string {
	sessionName := ""
	if m.cursor >= 0 && m.cursor < len(m.rows) {
		sessionName = m.rows[m.cursor].label
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		OverlayTitle.Render("Delete Session"),
		"",
		HelpDesc.Render("Are you sure you want to delete:"),
		"",
		lipgloss.NewStyle().Foreground(ColorText).Bold(true).PaddingLeft(2).Render(sessionName),
		"",
		HelpDesc.Render("This will kill the tmux session and remove it from config."),
		"",
		ErrorStyle.Render("y")+HelpSep.Render("  confirm    ")+HelpKey.Render("esc")+HelpSep.Render("  cancel"),
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
