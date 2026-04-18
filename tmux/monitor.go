package tmux

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Status int

const (
	StatusIdle    Status = iota
	StatusWorking        // content is actively changing
	StatusDone           // was working, now stopped
	StatusDead           // session no longer exists
)

type State struct {
	Status     Status
	FinishedAt *time.Time
}

type Monitor struct {
	mu          sync.RWMutex
	states      map[string]*mstate
}

type mstate struct {
	status      Status
	finishedAt  *time.Time
	lastContent string
	lastChanged time.Time
}

func NewMonitor() *Monitor {
	return &Monitor{states: make(map[string]*mstate)}
}

// Poll updates the state for all given session names.
func (m *Monitor) Poll(sessions []string) {
	alive := make(map[string]bool, len(sessions))
	for _, name := range sessions {
		alive[name] = true
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Mark removed sessions as dead
	for name, s := range m.states {
		if !alive[name] {
			s.status = StatusDead
		}
	}

	for _, name := range sessions {
		content := capturePane(name)
		s, ok := m.states[name]
		if !ok {
			s = &mstate{lastChanged: time.Now(), lastContent: content}
			m.states[name] = s
		}

		if content != s.lastContent {
			s.lastContent = content
			s.lastChanged = time.Now()
		}

		wasWorking := s.status == StatusWorking
		isChanging := time.Since(s.lastChanged) < 2*time.Second

		switch {
		case isChanging:
			s.status = StatusWorking
		case wasWorking:
			s.status = StatusDone
			now := time.Now()
			s.finishedAt = &now
		case s.status != StatusDone:
			s.status = StatusIdle
		}
	}
}

func (m *Monitor) Get(name string) State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.states[name]
	if !ok {
		return State{Status: StatusIdle}
	}
	return State{Status: s.status, FinishedAt: s.finishedAt}
}

func capturePane(session string) string {
	out, err := exec.Command("tmux", "capture-pane", "-t", session, "-p").Output()
	if err != nil {
		return ""
	}
	return string(out)
}

// CapturePaneOutput returns the last n lines of a session's active pane as plain text.
func CapturePaneOutput(session string, lines int) string {
	out, err := exec.Command("tmux", "capture-pane", "-t", session, "-p",
		"-S", fmt.Sprintf("-%d", lines)).Output()
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(out), "\n")
}
