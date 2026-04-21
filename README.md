# Claudster

A terminal UI for managing multiple Claude sessions across projects.

<img width="2553" height="1520" alt="CleanShot 2026-04-20 at 10 26 54" src="https://github.com/user-attachments/assets/2ab3e880-aead-4748-8c18-f4b1bf963fda" />

## Prerequisites

- [tmux](https://github.com/tmux/tmux/wiki/Installing) — claudster runs everything inside tmux sessions
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-code) — `claude` must be on your PATH

## Installation

Download the binary for your platform from the [releases page](../../releases) and put it on your PATH.

```bash
# macOS (Apple Silicon)
curl -L https://github.com/jonathangulliksen/claudster/releases/latest/download/claudster-darwin-arm64 -o claudster

# macOS (Intel)
curl -L https://github.com/jonathangulliksen/claudster/releases/latest/download/claudster-darwin-amd64 -o claudster

# Linux / WSL
curl -L https://github.com/jonathangulliksen/claudster/releases/latest/download/claudster-linux-amd64 -o claudster

chmod +x claudster
sudo mv claudster /usr/local/bin/
```

## Starting claudster

Claudster must run inside tmux. If you haven't used tmux before, think of it as a terminal multiplexer — it lets claudster run in the background while you switch between Claude sessions.

```bash
# Start tmux and launch claudster in one go
tmux new-session -s claudster -d 'claudster' && tmux attach -t claudster

# Or if you're already inside a tmux session
claudster
```

Once claudster is running, it stays in its own tmux window. You switch between claudster and your Claude sessions using your normal tmux keybinds (default: `Ctrl+b` then the window number).

## Setup

Claudster reads its config from `~/.claudster.yaml`. Create it:

```bash
cp claudster.example.yaml ~/.claudster.yaml
```

Then edit it to point at your repos. A project can have multiple repos — all of them get passed to Claude so it knows about the full codebase:

```yaml
groups:
  - name: work
    projects:
      - name: my-app
        repos:
          - ~/code/my-app        # primary repo — sessions start here
          - ~/code/my-app-api    # also passed to Claude via --add-dir
        sessions:
          - name: feature-x
```

Sessions listed under a project are tracked by claudster. They show up in the sidebar with live status (working / done / idle).

## Editor

Claudster auto-detects your editor from `$EDITOR`, then tries `code`, `nano`, `vim`, `vi` in that order. You can also pin it explicitly in the config:

```yaml
ui:
  editor: code   # or nvim, nano, vim, etc.
```

**VS Code users (including WSL):** `code` is fully supported. Claudster opens config files with `--wait` so it pauses while you edit, and opens repo folders in a new VS Code window. The one exception is `V` (persistent editor session inside tmux) — that doesn't work with VS Code, use `v` to open the folder instead.

## Keybindings

| Key | Action |
|-----|--------|
| `enter` | Attach to session (starts it if not running) |
| `n` | New Claude session |
| `r` | Resume a previous Claude session (interactive picker) |
| `d` | Delete session |
| `P` | Restart session |
| `v` | Open repo in editor |
| `V` | New persistent editor session (terminal editors only) |
| `t` | Open repo in terminal popup |
| `G` | Open repo in lazygit |
| `N` | New project |
| `e` | Edit config file |
| `p` | Toggle `--dangerously-skip-permissions` (global) |
| `[ / ]` | Resize sidebar |
| `?` | Full keybinding help |
| `q` | Quit |

When creating a new session (`n` or `r`) you'll be asked whether to enable `--dangerously-skip-permissions` before the session starts.

### Notifications

When a Claude session finishes, a toast appears in the corner. If you're in another tmux session at the time, a notification also flashes in your tmux status bar.

- Press `1`–`9` in claudster to jump to that session
- Press `opt+1`–`opt+9` (Mac) or `alt+1`–`alt+9` (Linux/WSL) from anywhere in tmux to jump directly without going back to claudster first

> **Mac note:** For `opt+N` to work in tmux, your terminal needs Option key configured as Meta/Esc+. In iTerm2: Preferences → Profiles → Keys → Left Option Key → set to `Esc+`.

## Building from source

Requires Go 1.21+.

```bash
make build        # current platform
make release      # all platforms → dist/
make install      # build + install to /usr/local/bin
```
