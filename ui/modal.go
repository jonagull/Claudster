package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"claudster/skills"
	"claudster/tmux"
)

func renderModal(m Model) string {
	if m.modal.mode == modalSessionPicker {
		return renderSessionPicker(m)
	}
	if m.modal.mode == modalContextMenu {
		return renderContextMenu(m)
	}
	if m.modal.mode == modalConfirmDelete {
		return renderConfirmDelete(m)
	}
	if m.modal.mode == modalConfirmSkillDelete {
		return renderConfirmSkillDelete(m)
	}
	if m.modal.mode == modalHelp {
		return renderHelp(m)
	}
	if m.modal.mode == modalSkillsInfo {
		return renderSkillsInfo(m)
	}
	if m.modal.mode == modalScratchAppend {
		return renderScratchAppend(m)
	}
	if m.modal.step == 1 {
		return renderDangerousConfirm(m)
	}

	var title, fieldLabel, hint string

	switch m.modal.mode {
	case modalNewProject:
		title = "New Project"
		fieldLabel = "Group:"
		hint = "tab to autocomplete  ·  template opens in " + resolveEditor(m.config.UI.Editor)

	case modalNewSession:
		title = fmt.Sprintf("New Session — %s", m.modal.targetProject)
		fieldLabel = "Session name:"
		hint = "starts in " + primaryRepoHint(m)
		if m.dangerousMode {
			hint += "  " + ErrorStyle.Render("⚠ --dangerously-skip-permissions")
		}

	case modalResumeSession:
		title = fmt.Sprintf("Resume Session — %s", m.modal.targetProject)
		fieldLabel = "Session name:"
		hint = "opens claude --resume picker in " + primaryRepoHint(m)
		if m.dangerousMode {
			hint += "  " + ErrorStyle.Render("⚠ --dangerously-skip-permissions")
		}

	case modalNewSkill:
		title = "New Skill"
		fieldLabel = "Skill name:"
		scopeLabel := "Global"
		if m.modal.targetSkillScope != skills.GlobalDir() {
			scopeLabel = filepath.Base(m.modal.targetSkillScope)
		}
		hint = fmt.Sprintf("scope: %s  ·  creates <scope>/<name>/SKILL.md", scopeLabel)

	case modalNewEditorSession:
		if m.modal.targetKind == "lazygit" {
			title = fmt.Sprintf("New Lazygit Session — %s", m.modal.targetProject)
			hint = "opens lazygit in " + primaryRepoHint(m)
			if m.config.UI.GitClient == "github-desktop" {
				hint = "note: GitHub Desktop doesn't run in tmux — use G to open it instead"
			}
		} else {
			title = fmt.Sprintf("New Editor Session — %s", m.modal.targetProject)
			hint = "opens " + resolveEditor(m.config.UI.Editor) + " in " + primaryRepoHint(m)
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

func renderHelp(m Model) string {
	type binding struct{ key, desc string }
	sections := []struct {
		title    string
		bindings []binding
	}{
		{"Navigate", []binding{
			{"j / ↓", "move down"},
			{"k / ↑", "move up"},
			{"ctrl+d / ctrl+u", "jump 5 rows down / up"},
			{"ctrl+p", "global session picker"},
			{"/", "search — jump to match as you type"},
			{"o", "go to overview"},
			{"space", "expand / collapse project"},
		}},
		{"Sessions", []binding{
			{"enter", "attach or start session"},
			{"m", "context menu (attach / restart / delete)"},
			{"n", "new Claude session"},
			{"r", "resume Claude session (picker)"},
			{"c", "scroll output (vim copy mode)"},
			{"T", "new terminal session (persistent)"},
			{"V", "new editor session (persistent)"},
			{"d", "delete session (confirm required)"},
			{"P", "restart Claude session"},
		}},
		{"Quick open", []binding{
			{"v", "open repo in editor"},
			{"t", "open repo in terminal"},
			{"G", "open repo in lazygit"},
			{"s", "quick-add to scratch"},
			{"S", "open full scratch file"},
		}},
		{"Project / config", []binding{
			{"N", "new project"},
			{"e", "edit config file"},
		}},
		{"Skills", []binding{
			{"a", "new skill (in current scope)"},
			{"enter / v", "edit skill in editor"},
			{"d", "delete skill"},
			{"i", "what are skills? (reference)"},
		}},
		{"UI", []binding{
			{"[ / ]", "resize sidebar"},
			{"p", "toggle --dangerously-skip-permissions"},
			{"?", "this help page"},
			{"q / ctrl+q", "quit"},
		}},
	}

	var col1, col2 []string
	col1 = append(col1, OverlayTitle.Render("Keybindings"), "")
	col2 = append(col2, "", "")

	for _, sec := range sections {
		col1 = append(col1, PreviewKey.Render(sec.title))
		col2 = append(col2, "")
		for _, b := range sec.bindings {
			col1 = append(col1, HelpKey.Render("  "+b.key))
			col2 = append(col2, HelpDesc.Render(b.desc))
		}
		col1 = append(col1, "")
		col2 = append(col2, "")
	}

	// Pad both columns to equal length
	for len(col1) < len(col2) {
		col1 = append(col1, "")
	}
	for len(col2) < len(col1) {
		col2 = append(col2, "")
	}

	keyW := 24
	var rows []string
	for i := range col1 {
		keyCell := lipgloss.NewStyle().Width(keyW).Render(col1[i])
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, keyCell, col2[i]))
	}

	rows = append(rows, HelpDesc.Render("esc  close"))

	body := strings.Join(rows, "\n")
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		OverlayStyle.Render(body),
	)
}

func renderSkillsInfo(m Model) string {
	h := func(text string) string { return PreviewKey.Render(text) }
	b := func(text string) string { return HelpDesc.Render(text) }
	c := func(text string) string { return PreviewComment.Render(text) }
	key := func(k, desc string) string {
		return HelpKey.Render(k) + HelpSep.Render("  ") + HelpDesc.Render(desc)
	}
	code := func(text string) string {
		return lipgloss.NewStyle().Foreground(ColorSecondary).Render(text)
	}
	globalDir := skills.GlobalDir()

	lines := []string{
		OverlayTitle.Render("What are Skills?"),
		"",
		b("Skills teach Claude how to do specific things — they are Markdown files that"),
		b("Claude reads before responding. Think of them as reusable instructions or"),
		b("playbooks that live in your filesystem, not inside any conversation."),
		"",
		h("Where skills live"),
		code("  ~/.claude/skills/<name>/SKILL.md") + c("  ← global (all projects)"),
		code("  <repo>/.claude/skills/<name>/SKILL.md") + c("  ← project-specific"),
		b("  Current global dir: " + globalDir),
		"",
		h("SKILL.md format"),
		code("  ---"),
		code("  name: git-commit"),
		code("  description: Write conventional commit messages."),
		code("  ---"),
		"",
		code("  # Instructions for Claude"),
		code("  When writing commit messages, use the format: type(scope): summary"),
		"",
		b("  The description field tells Claude when to load this skill automatically."),
		b("  The body is the actual instructions — plain Markdown, any length."),
		"",
		h("How Claude uses them"),
		b("  · Claude sees all skill descriptions at session start (low cost)."),
		b("  · It loads a skill's full content when it's relevant to the task."),
		b("  · You can also invoke one manually: type /skill-name in Claude."),
		b("  · Loaded skills persist for the whole session, even after compaction."),
		"",
		h("Managing skills here"),
		"  " + key("a", "new skill — creates <scope>/<name>/SKILL.md and opens it"),
		"  " + key("enter / v", "edit an existing skill in your editor"),
		"  " + key("d", "delete a skill (confirm required)"),
		"  " + key("i", "this page"),
		"",
		HelpDesc.Render("esc  close"),
	}

	body := strings.Join(lines, "\n")
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

func renderConfirmSkillDelete(m Model) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		OverlayTitle.Render("Delete Skill"),
		"",
		HelpDesc.Render("Are you sure you want to delete:"),
		"",
		lipgloss.NewStyle().Foreground(ColorText).Bold(true).PaddingLeft(2).Render(m.modal.targetSkillName),
		MutedItem.PaddingLeft(2).Render(m.modal.targetSkillDir),
		"",
		HelpDesc.Render("This will permanently remove the skill directory."),
		"",
		ErrorStyle.Render("y")+HelpSep.Render("  confirm    ")+HelpKey.Render("esc")+HelpSep.Render("  cancel"),
	)
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		OverlayStyle.Render(body),
	)
}

func renderDangerousConfirm(m Model) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		OverlayTitle.Render("New Session — "+m.modal.targetProject),
		"",
		PreviewKey.Render("Session name:"),
		NormalItem.PaddingLeft(2).Render(m.modal.pendingName),
		"",
		PreviewKey.Render("Run with --dangerously-skip-permissions?"),
		HelpDesc.Render("Skips permission prompts. Only use if you trust the codebase."),
		"",
		HelpKey.Render("y")+" "+HelpDesc.Render("yes    ")+HelpKey.Render("n / enter")+" "+HelpDesc.Render("no    ")+HelpKey.Render("esc")+" "+HelpDesc.Render("cancel"),
	)
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		OverlayStyle.Render(body),
	)
}

func renderSessionPicker(m Model) string {
	entries := m.pickerEntries()
	isFiltering := m.modal.pickerQuery != ""

	prompt := HelpKey.Render("  ⌕  ") + NormalItem.Render(m.modal.pickerQuery) + MutedItem.Render("█")

	const maxVisible = 14
	var resultLines []string
	if len(entries) == 0 {
		resultLines = append(resultLines, MutedItem.PaddingLeft(2).Render("no sessions match"))
	} else {
		start := 0
		if m.modal.pickerCursor >= maxVisible {
			start = m.modal.pickerCursor - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(entries) {
			end = len(entries)
		}

		// Track section transitions so we can insert a divider
		prevRecent := entries[start].recent
		if !isFiltering && start == 0 && len(m.recentSessions) > 0 {
			resultLines = append(resultLines, MutedItem.PaddingLeft(2).Render("recent"))
		}

		for i := start; i < end; i++ {
			e := entries[i]

			// Insert "all sessions" header when we cross from recent → rest
			if !isFiltering && !e.recent && prevRecent {
				resultLines = append(resultLines, "")
				resultLines = append(resultLines, MutedItem.PaddingLeft(2).Render("all sessions"))
			}
			prevRecent = e.recent

			running := m.monitor.Exists(e.sessionName)
			state := m.monitor.Get(e.sessionName)

			var icon string
			if !running {
				icon = MutedItem.Render("○")
			} else if state.Status == tmux.StatusWorking {
				icon = WorkingBadge.Render("●")
			} else if state.Status == tmux.StatusDone {
				icon = DoneBadge.Render("✓")
			} else {
				icon = NormalItem.Render("─")
			}

			meta := MutedItem.Render(" · " + e.projectName + " · " + e.groupName)
			if i == m.modal.pickerCursor {
				resultLines = append(resultLines, lipgloss.NewStyle().PaddingLeft(2).Render(
					icon+"  "+SelectedItem.Render(e.sessionName)+meta,
				))
			} else {
				resultLines = append(resultLines, lipgloss.NewStyle().PaddingLeft(2).Render(
					icon+"  "+NormalItem.Render(e.sessionName)+meta,
				))
			}
		}
	}

	divider := MutedItem.Render(strings.Repeat("─", 48))
	body := lipgloss.JoinVertical(lipgloss.Left,
		prompt,
		divider,
		strings.Join(resultLines, "\n"),
		"",
		HelpDesc.Render("  ↑/↓  navigate    enter  attach    esc  close"),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		OverlayStyle.Render(body),
	)
}

func renderContextMenu(m Model) string {
	name := m.modal.pendingName
	running := m.monitor.Exists(name)
	state := m.monitor.Get(name)

	var statusLine string
	if !running {
		statusLine = MutedItem.Render("○  stopped")
	} else {
		switch state.Status {
		case tmux.StatusWorking:
			statusLine = WorkingBadge.Render("●  working")
		case tmux.StatusDone:
			statusLine = DoneBadge.Render("✓  done")
		default:
			statusLine = MutedItem.Render("─  idle")
		}
	}

	sep := func(key, desc string) string {
		return HelpKey.Render(key) + HelpSep.Render("  ") + HelpDesc.Render(desc)
	}
	errSep := func(key, desc string) string {
		return ErrorStyle.Render(key) + HelpSep.Render("  ") + HelpDesc.Render(desc)
	}

	var opts []string
	opts = append(opts, sep("enter", "attach to session"))
	if running {
		opts = append(opts, sep("c", "scroll output (copy mode)"))
		opts = append(opts, sep("P", "restart session"))
	}
	opts = append(opts, errSep("d", "delete session"))

	body := lipgloss.JoinVertical(lipgloss.Left,
		OverlayTitle.Render(name),
		statusLine,
		"",
		strings.Join(opts, "\n"),
		"",
		HelpDesc.Render("esc  close"),
	)
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		OverlayStyle.Render(body),
	)
}

// resolveEditor returns the editor to use, checking (in order):
// config ui.editor → $EDITOR env → auto-detect from PATH.
func resolveEditor(configured string) string {
	if configured != "" {
		return configured
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	for _, e := range []string{"code", "code-insiders", "nano", "vim", "vi"} {
		if _, err := exec.LookPath(e); err == nil {
			return e
		}
	}
	return "vi"
}

// isVSCode reports whether the editor binary is VS Code.
func isVSCode(editor string) bool {
	base := filepath.Base(editor)
	return base == "code" || base == "code-insiders"
}

// wslPath converts an absolute Linux path to a Windows path when running
// under WSL, so VS Code (a Windows app) can open the file correctly.
// On non-WSL systems it returns the path unchanged.
func wslPath(path string) string {
	out, err := exec.Command("wslpath", "-w", path).Output()
	if err != nil {
		return path
	}
	return strings.TrimSpace(string(out))
}

func renderScratchAppend(m Model) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		OverlayTitle.Render("Scratch — "+m.modal.targetProject),
		"",
		HelpDesc.Render("Quick note (appended as a bullet):"),
		InputStyle.Width(46).Render(m.modal.input.View()),
		"",
		HelpDesc.Render("enter  save    S  open full scratch    esc  cancel"),
	)
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		OverlayStyle.Render(body),
	)
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
