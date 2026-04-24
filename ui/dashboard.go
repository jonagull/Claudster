package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"claudster/metrics"
	"claudster/skills"
	"claudster/store"
	"claudster/tmux"
)

const cardContentW = 22
const cardContentH = 3

var overviewLogoBig = []string{
	` ██████╗██╗      █████╗ ██╗   ██╗██████╗ ███████╗████████╗███████╗██████╗ `,
	`██╔════╝██║     ██╔══██╗██║   ██║██╔══██╗██╔════╝╚══██╔══╝██╔════╝██╔══██╗`,
	`██║     ██║     ███████║██║   ██║██║  ██║███████╗   ██║   █████╗  ██████╔╝ `,
	`██║     ██║     ██╔══██║██║   ██║██║  ██║╚════██║   ██║   ██╔══╝  ██╔══██╗ `,
	`╚██████╗███████╗██║  ██║╚██████╔╝██████╔╝███████║   ██║   ███████╗██║  ██║ `,
	` ╚═════╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═╝`,
}

// activity bar frames — scrolls right when sessions are working
var activityFrames = []string{
	"▰▱▱▱▱▱▱▱▱▱",
	"▰▰▱▱▱▱▱▱▱▱",
	"▰▰▰▱▱▱▱▱▱▱",
	"▰▰▰▰▱▱▱▱▱▱",
	"▰▰▰▰▰▱▱▱▱▱",
	"▰▰▰▰▰▰▱▱▱▱",
	"▰▰▰▰▰▰▰▱▱▱",
	"▰▰▰▰▰▰▰▰▱▱",
	"▰▰▰▰▰▰▰▰▰▱",
	"▰▰▰▰▰▰▰▰▰▰",
}

func renderRightPanel(m Model) string {
	innerH := m.height - 3

	onOverview := m.cursor >= 0 && m.cursor < len(m.rows) && m.rows[m.cursor].typ == rowTypeOverview

	var content string
	if onOverview {
		rendered := renderOverview(m, m.dashW, innerH)
		if m.dashScrollOffset > 0 {
			rendered = skipLines(rendered, m.dashScrollOffset)
		}
		content = clipLines(rendered, innerH)
	} else {
		m.dashScrollOffset = 0 // reset when leaving overview
		header := lipgloss.NewStyle().Padding(0, 1).Render(InactivePanelTitle.Render("preview"))
		preview := clipLines(renderPreviewSection(m, m.dashW, innerH-1), innerH-1)
		content = lipgloss.JoinVertical(lipgloss.Left, header, preview)
	}

	// Hard-clip before handing to the border so it can never grow the layout.
	content = clipLines(content, innerH)
	// Pre-pad each line to the exact content width using ANSI-aware measurement
	// so lipgloss doesn't need to normalise widths itself. Width normalisation
	// runs ANSI sequences through truncation logic that can strip colours from
	// captured tmux pane output.
	content = padLinesToWidth(strings.TrimRight(content, "\n"), m.dashW)
	return InactiveBorder.
		Height(innerH).
		Render(content)
}

func renderOverview(m Model, w, h int) string {
	// Tally session states (tool sessions excluded from Claude counts)
	var nWorking, nDone, nIdle, nStopped int
	for _, g := range m.config.Groups {
		for _, p := range g.Projects {
			for _, s := range p.Sessions {
				if s.IsToolSession() {
					continue
				}
				if !m.monitor.Exists(s.Name) {
					nStopped++
					continue
				}
				switch m.monitor.Get(s.Name).Status {
				case tmux.StatusWorking:
					nWorking++
				case tmux.StatusDone:
					nDone++
				default:
					nIdle++
				}
			}
		}
	}

	var lines []string

	// Logo
	if w >= 74 {
		for _, l := range overviewLogoBig {
			lines = append(lines, lipgloss.NewStyle().Foreground(ColorPrimary).Render(l))
		}
	} else {
		lines = append(lines,
			lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Padding(0, 1).Render("claudster"),
			lipgloss.NewStyle().Foreground(ColorMuted).Padding(0, 1).Render("claude session manager"),
		)
	}
	lines = append(lines, "")

	// Claude usage stats
	lines = append(lines, renderMetricsBar(m, w))
	lines = append(lines, "")

	// Animated activity bar when sessions are working
	if nWorking > 0 {
		bar := activityFrames[m.spinFrame%len(activityFrames)]
		lines = append(lines,
			lipgloss.NewStyle().Padding(0, 1).Render(
				WorkingBadge.Render(spinner[m.spinFrame])+"  "+
					WorkingBadge.Render(bar)+"  "+
					WorkingBadge.Bold(false).Render(fmt.Sprintf("%d session(s) active", nWorking)),
			),
		)
		lines = append(lines, "")
	}

	// Stats row
	var stats []string
	if nWorking > 0 {
		stats = append(stats, WorkingBadge.Render(fmt.Sprintf("● %d working", nWorking)))
	}
	if nDone > 0 {
		stats = append(stats, DoneBadge.Render(fmt.Sprintf("✓ %d done", nDone)))
	}
	if nIdle > 0 {
		stats = append(stats, MutedItem.Render(fmt.Sprintf("─ %d idle", nIdle)))
	}
	if nStopped > 0 {
		stats = append(stats, MutedItem.Render(fmt.Sprintf("○ %d stopped", nStopped)))
	}
	if len(stats) > 0 {
		lines = append(lines,
			lipgloss.NewStyle().Padding(0, 1).Render(strings.Join(stats, "   ")),
		)
		lines = append(lines, "")
	}

	// Cards
	cardsH := h - len(lines)
	if cardsH < 1 {
		cardsH = 1
	}
	lines = append(lines, renderCardGrid(m, w, cardsH))

	return strings.Join(lines, "\n")
}

func renderMetricsBar(m Model, w int) string {
	s := m.claudeStats
	if s.Messages == 0 {
		return lipgloss.NewStyle().Padding(0, 1).Render(MutedItem.Render("collecting usage stats…"))
	}

	sep := MutedItem.Render("  ·  ")

	parts := []string{
		PreviewKey.Render("out") + "  " + PreviewValue.Render(metrics.FmtTokens(s.OutputTokens)),
		PreviewKey.Render("ctx") + "  " + PreviewValue.Render(metrics.FmtTokens(s.CacheRead+s.CacheWrite)) +
			MutedItem.Render(" cached"),
		PreviewKey.Render("api cost") + "  " + MutedItem.Render("~"+metrics.FmtCost(s.EstimatedCost())) +
			MutedItem.Render(" (if not on Max)"),
		PreviewKey.Render("calls") + "  " + PreviewValue.Render(fmt.Sprintf("%d", s.Messages)),
	}

	line := strings.Join(parts, sep)
	return lipgloss.NewStyle().Padding(0, 1).Render(line)
}

func renderCardGrid(m Model, w, h int) string {
	selectedName := ""
	if m.cursor >= 0 && m.cursor < len(m.rows) && m.rows[m.cursor].typ == rowTypeSession {
		selectedName = m.rows[m.cursor].label
	}

	colW := cardContentW + 4
	cols := w / colW
	if cols < 1 {
		cols = 1
	}

	var sections []string
	anyCards := false

	for _, g := range m.config.Groups {
		var cards []cardData
		for _, p := range g.Projects {
			for _, s := range p.Sessions {
				if s.IsToolSession() {
					continue
				}
				anyCards = true
				state := m.monitor.Get(s.Name)
				running := m.monitor.Exists(s.Name)
				cards = append(cards, cardData{
					session:  s,
					project:  p,
					group:    g,
					state:    state,
					running:  running,
					selected: s.Name == selectedName,
				})
			}
		}
		if len(cards) == 0 {
			continue
		}

		// Group header with a rule extending to panel width
		nameW := lipgloss.Width(g.Name)
		ruleW := w - nameW - 2 // 1 space + name + 1 space
		if ruleW < 0 {
			ruleW = 0
		}
		groupHeader := lipgloss.NewStyle().Foreground(ColorSubtle).Bold(true).Render(g.Name) +
			lipgloss.NewStyle().Foreground(ColorDimBorder).Render(" " + strings.Repeat("─", ruleW))

		var rows []string
		rows = append(rows, groupHeader)
		for i := 0; i < len(cards); i += cols {
			end := i + cols
			if end > len(cards) {
				end = len(cards)
			}
			var rendered []string
			for _, c := range cards[i:end] {
				rendered = append(rendered, renderCard(m, c))
			}
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rendered...))
		}
		sections = append(sections, strings.Join(rows, "\n"))
	}

	if !anyCards {
		return MutedItem.Padding(1, 2).
			Render("No sessions yet — select a project and press n.")
	}

	return strings.Join(sections, "\n\n")
}

func renderPreviewSection(m Model, w, h int) string {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return MutedItem.Padding(1, 2).Render("select a session to see details")
	}

	row := m.rows[m.cursor]

	switch row.typ {
	case rowTypeSession:
		return renderSessionPreview(m, row, w, h)
	case rowTypeProject:
		return renderProjectPreview(m, row, w, h)
	case rowTypeSkill:
		return renderSkillPreview(row, w, h)
	case rowTypeSkillScope:
		return renderSkillScopePreview(row, w, h)
	case rowTypeSkillsHeader:
		return renderSkillsHeaderPreview(w, h)
	default:
		return MutedItem.Padding(1, 2).Render("select a session to see details")
	}
}

func renderSessionPreview(m Model, row sidebarRow, w, h int) string {
	if !m.monitor.Exists(row.label) {
		return lipgloss.NewStyle().Padding(1, 2).Render(
			MutedItem.Render("session stopped — press enter to start"),
		)
	}

	pane := tmux.CapturePaneOutput(row.label, h)
	if pane == "" {
		return lipgloss.NewStyle().Padding(1, 2).Render(MutedItem.Render("no output yet"))
	}

	// Strip trailing spaces from each line (tmux pads to pane width).
	// Don't byte-truncate — that breaks ANSI escape sequences.
	// Take the last h lines so we always show the most recent output.
	rawLines := strings.Split(pane, "\n")
	for i, line := range rawLines {
		rawLines[i] = strings.TrimRight(line, " ")
	}
	if len(rawLines) > h {
		rawLines = rawLines[len(rawLines)-h:]
	}
	return strings.Join(rawLines, "\n")
}

func renderProjectPreview(m Model, row sidebarRow, w, h int) string {
	proj := m.config.Groups[row.groupIdx].Projects[row.projectIdx]

	kv := func(k, v string) string {
		return PreviewKey.Render(k+":") + " " + PreviewValue.Render(v)
	}

	var lines []string
	lines = append(lines, kv("project", proj.Name))
	lines = append(lines, kv("group", m.config.Groups[row.groupIdx].Name))

	if len(proj.Repos) > 0 {
		lines = append(lines, PreviewKey.Render("repos:"))
		for i, r := range proj.Repos {
			prefix := "  "
			if i == 0 {
				prefix = "  * "
			}
			lines = append(lines, PreviewValue.Render(prefix+r))
		}
	}

	lines = append(lines, "")
	if len(proj.Sessions) == 0 {
		lines = append(lines, MutedItem.Render("  press n to start a session"))
	} else {
		lines = append(lines, PreviewKey.Render("sessions:"))
		for _, s := range proj.Sessions {
			running := m.monitor.Exists(s.Name)
			state := m.monitor.Get(s.Name)
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
			lines = append(lines, "  "+icon+" "+NormalItem.Render(s.Name))
		}
	}

	return lipgloss.NewStyle().Padding(0, 2).Render(strings.Join(lines, "\n"))
}

// ProjectPreviewSessionY returns the screen Y of the first session name in the
// project preview, and how many sessions follow, so click hit-testing can work.
// Layout: right border(1) + header(1) + padding(1) + project(1) + group(1) + repos(N) + blank(1) + "sessions:"(1) = fixed offset
func projectPreviewSessionOffset(proj store.Project) int {
	// 1 border + 1 header + 1 padding-top (Padding(0,2) adds 0 vertical) +
	// project + group + blank-before-repos +
	// repos block + blank + "sessions:" label
	offset := 1 + 1 + 1 // border + header + 1-based click offset
	offset += 1         // "project:" line
	offset += 1         // "group:" line
	if len(proj.Repos) > 0 {
		offset += 1 + len(proj.Repos) // "repos:" + each repo
	}
	offset += 1 // blank line
	offset += 1 // "sessions:" label
	return offset
}

func renderSkillPreview(row sidebarRow, w, h int) string {
	kv := func(k, v string) string {
		return PreviewKey.Render(k+":") + " " + PreviewValue.Render(v)
	}

	scopeLabel := "Global"
	if row.skillScope != skills.GlobalDir() {
		scopeLabel = row.skillScope
	}

	content := skills.ReadContent(row.skillPath)

	var lines []string
	lines = append(lines, kv("skill", row.label))
	lines = append(lines, kv("scope", scopeLabel))
	lines = append(lines, kv("file", row.skillPath))
	lines = append(lines, "")
	lines = append(lines, PreviewKey.Render("content:"))
	lines = append(lines, "")

	if content == "" {
		lines = append(lines, MutedItem.Render("  (empty)"))
	} else {
		for _, l := range strings.Split(strings.TrimRight(content, "\n"), "\n") {
			lines = append(lines, PreviewComment.Render("  "+l))
		}
	}

	lines = append(lines, "")
	lines = append(lines, MutedItem.Render("  enter/v  edit    d  delete"))

	return lipgloss.NewStyle().Padding(0, 2).Render(strings.Join(lines, "\n"))
}

func renderSkillScopePreview(row sidebarRow, w, h int) string {
	kv := func(k, v string) string {
		return PreviewKey.Render(k+":") + " " + PreviewValue.Render(v)
	}
	scopeDisplay := "Global  (~/.claude/skills/)"
	if row.skillScope != skills.GlobalDir() {
		scopeDisplay = row.label + "  (" + row.skillScope + ")"
	}
	var lines []string
	lines = append(lines, kv("scope", scopeDisplay))
	lines = append(lines, "")
	lines = append(lines, MutedItem.Render("  a  add new skill in this scope"))
	return lipgloss.NewStyle().Padding(0, 2).Render(strings.Join(lines, "\n"))
}

func renderSkillsHeaderPreview(w, h int) string {
	var lines []string
	lines = append(lines, PreviewKey.Render("Skills"))
	lines = append(lines, "")
	lines = append(lines, MutedItem.Render("  Skills are Markdown files that teach Claude how to do specific things."))
	lines = append(lines, MutedItem.Render("  They live in your filesystem and load automatically when relevant."))
	lines = append(lines, "")
	lines = append(lines, MutedItem.Render("  Global skills  →  ~/.claude/skills/<name>/SKILL.md"))
	lines = append(lines, MutedItem.Render("  Project skills →  <repo>/.claude/skills/<name>/SKILL.md"))
	lines = append(lines, "")
	lines = append(lines, HelpKey.Render("  i")+HelpDesc.Render("  full reference page    ")+HelpKey.Render("a")+HelpDesc.Render("  new skill"))
	return lipgloss.NewStyle().Padding(0, 2).Render(strings.Join(lines, "\n"))
}

type cardData struct {
	session  store.Session
	project  store.Project
	group    store.Group
	state    tmux.State
	running  bool
	selected bool
}

func renderCard(m Model, c cardData) string {
	name := c.session.Name
	if len(name) > cardContentW {
		name = name[:cardContentW-1] + "…"
	}

	subtitle := MutedItem.Render(c.project.Name + " · " + c.group.Name)

	var statusLine string
	if !c.running {
		statusLine = MutedItem.Render("○  stopped")
	} else if c.session.IsToolSession() {
		switch c.session.Kind {
		case "lazygit":
			statusLine = MutedItem.Render("⎇  lazygit")
		case "terminal":
			statusLine = MutedItem.Render("$  terminal")
		default:
			statusLine = MutedItem.Render("✎  editor")
		}
	} else {
		switch c.state.Status {
		case tmux.StatusWorking:
			statusLine = WorkingBadge.Render(spinner[m.spinFrame] + "  working")
		case tmux.StatusDone:
			ts := ""
			if c.state.FinishedAt != nil {
				ts = "  " + relTime(*c.state.FinishedAt)
			}
			statusLine = DoneBadge.Render("✓  done") + TimestampStyle.Render(ts)
		default:
			statusLine = MutedItem.Render("─  idle")
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(ColorText).Bold(c.selected).Render(name),
		subtitle,
		"",
		statusLine,
	)

	var border lipgloss.Style
	if c.selected {
		border = CardSelected
	} else if !c.running {
		border = CardStopped
	} else {
		switch c.state.Status {
		case tmux.StatusWorking:
			border = CardWorking
		case tmux.StatusDone:
			border = CardDone
		default:
			border = CardIdle
		}
	}

	return border.Width(cardContentW).Height(cardContentH).Render(content)
}

// renderHoverTooltip renders a compact floating preview for the hovered card.
func renderHoverTooltip(m Model) string {
	name := m.hoverSession
	if name == "" {
		return ""
	}

	running := m.monitor.Exists(name)
	state := m.monitor.Get(name)

	var header string
	if !running {
		header = MutedItem.Render("○  ") + SelectedItem.Render(name)
	} else {
		switch state.Status {
		case tmux.StatusWorking:
			frame := m.spinFrame % len(spinner)
			color := workingPalette[m.spinFrame%len(workingPalette)]
			icon := lipgloss.NewStyle().Foreground(color).Bold(true).Render(spinner[frame])
			header = icon + "  " + SelectedItem.Render(name)
		case tmux.StatusDone:
			header = DoneBadge.Render("✓  ") + SelectedItem.Render(name)
		default:
			header = MutedItem.Render("─  ") + SelectedItem.Render(name)
		}
	}

	const contentW = 76
	var bodyLines []string
	bodyLines = append(bodyLines, header)
	bodyLines = append(bodyLines, MutedItem.Render(strings.Repeat("─", contentW)))

	if !running {
		bodyLines = append(bodyLines, MutedItem.Render("  session stopped"))
	} else if m.hoverPreview == "" {
		bodyLines = append(bodyLines, MutedItem.Render("  loading…"))
	} else {
		rawLines := strings.Split(m.hoverPreview, "\n")
		// Trim trailing spaces (tmux pads to pane width)
		for i, l := range rawLines {
			rawLines[i] = strings.TrimRight(l, " ")
		}
		// Take last 18 lines
		if len(rawLines) > 18 {
			rawLines = rawLines[len(rawLines)-18:]
		}
		for _, l := range rawLines {
			// Hard-truncate visible width so tooltip stays within bounds
			if lipgloss.Width(l) > contentW {
				runes := []rune(l)
				if len(runes) > contentW {
					l = string(runes[:contentW])
				}
			}
			bodyLines = append(bodyLines, PreviewComment.Render(l))
		}
	}

	body := strings.Join(bodyLines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 1).
		Width(contentW).
		Render(body)
}

// padLinesToWidth pads each line to exactly w visible characters using
// lipgloss.Width for ANSI-aware measurement, so that the outer container
// never needs to truncate lines (which can corrupt embedded ANSI sequences).
// Lines already at or above w are left as-is.
func padLinesToWidth(content string, w int) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		vis := lipgloss.Width(line)
		if vis < w {
			lines[i] = line + strings.Repeat(" ", w-vis)
		}
	}
	return strings.Join(lines, "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// clipLines truncates s to at most n lines (newline-separated).
func skipLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if n >= len(lines) {
		return ""
	}
	return strings.Join(lines[n:], "\n")
}

func clipLines(s string, n int) string {
	if n <= 0 {
		return ""
	}
	lines := strings.SplitN(s, "\n", n+1)
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}
