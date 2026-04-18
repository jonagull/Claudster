# Claudster

lazygit/lazydocker — but for Claude Code sessions.

A single terminal UI that gives you a live overview of every Claude agent you're running across all your projects. See what's working, what's idle, what just finished. Jump in and out.

---

## Data Model

Three levels:

```
Group
  └── Project   (defines what you're working on — owns repos)
        └── Session  (a Claude process — ephemeral, lives in tmux)
```

- **Group** — organizational bucket (e.g. "work", "personal")
- **Project** — a body of work spanning one or more repos. The first repo listed is the master/primary repo. Projects are stable and defined in the config.
- **Session** — a tmux process running `claude` in the project's primary repo. Multiple sessions can run under the same project in parallel. Sessions are stored in the config so stopped/dead ones are still visible and can be renamed or cleared.

---

## Config File: `~/.claudster.yaml`

```yaml
groups:
  - name: work
    projects:
      - name: platform
        repos:
          - ~/code/frontend   # first = master repo (where sessions start)
          - ~/code/backend
          - ~/code/api
        sessions:
          - name: auth-refactor
          - name: ui-cleanup

      - name: api
        repos:
          - ~/code/api
        sessions:
          - name: api-session

  - name: personal
    projects:
      - name: dotfiles
        repos:
          - ~/.dotfiles
        sessions:
          - name: dotfiles
```

The config is the source of truth. Sessions are shown even when stopped.

---

## Sidebar Tree

```
work
  ▼ platform
      ⟳ auth-refactor      working
      ─ ui-cleanup          idle
      ○ old-session         stopped
  ▶ api                     (collapsed)

personal
  ▼ dotfiles
      ─ dotfiles            idle
```

- Projects are expandable/collapsible with `space`
- Sessions under a project show live status + timestamps
- Stopped sessions (not in tmux) show as `○`
- Pressing `enter` on a session attaches to it
- Pressing `n` when a project is focused creates a new session under it

---

## Session Status Indicators

| Symbol | Meaning                              |
|--------|--------------------------------------|
| `⟳`    | Actively working (content changing)  |
| `─`    | Idle, waiting for input              |
| `✓`    | Just finished — shows "Xs ago"       |
| `○`    | Stopped (no tmux session)            |
| `✗`    | Exited unexpectedly                  |

---

## Creating a Project (`N`)

1. Group name (existing or new)
2. Project name
3. Repos — comma-separated, first is master

## Creating a Session (`n` when on a project)

1. Session name
Spawns `claude` in the project's primary (first) repo.

---

## Keybindings

| Key     | Action                                       |
|---------|----------------------------------------------|
| `n`     | New session (under focused project)          |
| `N`     | New project                                  |
| `d`     | Delete selected session (kill + remove)      |
| `space` | Expand/collapse project                      |
| `j/k`   | Navigate (skips group headers)               |
| `enter` | Attach to session                            |
| `e`     | Edit `~/.claudster.yaml` in $EDITOR          |
| `pgup/dn` | Scroll preview                             |
| `q`     | Quit                                         |

---

## Preview Panel

When a **session** is selected:
- Session name, project, group
- Repos listed (master highlighted)
- Live status + finish timestamp
- Last ~100 lines of pane output (scrollable)

When a **project** is selected:
- Project name + group
- All repos
- Count of running / stopped sessions

---

## Notifications

When a session finishes:
- System notification via `osascript`
- In-TUI toast overlay

---

## Persistence

`~/.claudster.yaml` is always the source of truth. Adding a session via `n` writes it to the file immediately. Pressing `e` opens it directly in `$EDITOR`.

---

## File Structure

```
claudster/
├── main.go
├── tmux/
│   ├── session.go       # create, kill, list tmux sessions
│   └── monitor.go       # poll pane output, detect activity, record finish time
├── store/
│   └── store.go         # read/write ~/.claudster.yaml
├── ui/
│   ├── model.go         # Bubble Tea model + update loop
│   ├── sidebar.go       # tree rendering (groups, projects, sessions)
│   ├── preview.go       # preview panel
│   ├── modal.go         # new session / new project forms
│   └── styles.go        # lipgloss styling
└── go.mod
```
