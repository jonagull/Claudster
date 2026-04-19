package metrics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Stats struct {
	InputTokens  int64
	OutputTokens int64
	CacheRead    int64
	CacheWrite   int64
	Messages     int64
	Convos       int64
	CollectedAt  time.Time
}

// EstimatedCost returns a rough USD estimate using Claude Sonnet 4.6 pricing.
func (s Stats) EstimatedCost() float64 {
	return float64(s.InputTokens)/1e6*3.00 +
		float64(s.OutputTokens)/1e6*15.00 +
		float64(s.CacheRead)/1e6*0.30 +
		float64(s.CacheWrite)/1e6*3.75
}

// FmtTokens formats a token count as a short human-readable string.
func FmtTokens(n int64) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(n)/1e9)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1e6)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1e3)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// FmtCost formats a USD cost value.
func FmtCost(usd float64) string {
	switch {
	case usd >= 1000:
		return fmt.Sprintf("$%.0f", usd)
	case usd >= 1:
		return fmt.Sprintf("$%.2f", usd)
	default:
		return fmt.Sprintf("$%.4f", usd)
	}
}

// jsonlEntry is the minimal structure needed to extract token usage.
type jsonlEntry struct {
	Type    string `json:"type"` // "user", "file-history-snapshot", etc.; absent for assistant
	Message *struct {
		Role  string `json:"role"`
		Usage *struct {
			Input       int64 `json:"input_tokens"`
			Output      int64 `json:"output_tokens"`
			CacheRead   int64 `json:"cache_read_input_tokens"`
			CacheCreate int64 `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// LatestConvoID returns the UUID of the most recently modified conversation
// for the given project directory (e.g. "~/code/myproject"). Returns "" if
// nothing is found so the caller can fall back to `claude --resume`.
func LatestConvoID(projectPath string) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(projectPath, "~/") {
		projectPath = filepath.Join(home, projectPath[2:])
	}
	// Claude encodes the project path as the directory name by replacing / with -
	encoded := strings.ReplaceAll(projectPath, "/", "-")
	dir := filepath.Join(home, ".claude", "projects", encoded)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var latestMod time.Time
	var latestID string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".jsonl" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestMod) {
			latestMod = info.ModTime()
			latestID = strings.TrimSuffix(e.Name(), ".jsonl")
		}
	}
	return latestID
}

// Collect scans all Claude Code project directories and aggregates token usage.
func Collect() Stats {
	home, _ := os.UserHomeDir()
	projectsDir := filepath.Join(home, ".claude", "projects")

	var s Stats
	s.CollectedAt = time.Now()

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return s
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(projectsDir, entry.Name())
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if filepath.Ext(f.Name()) != ".jsonl" {
				continue
			}
			s.Convos++
			parseFile(filepath.Join(dir, f.Name()), &s)
		}
	}

	return s
}

func parseFile(path string, s *Stats) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	const maxLine = 10 * 1024 * 1024 // 10 MB — tool results can be large
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, maxLine), maxLine)

	for sc.Scan() {
		var e jsonlEntry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			continue
		}
		if e.Type != "assistant" {
			continue
		}
		if e.Message == nil || e.Message.Usage == nil {
			continue
		}
		s.Messages++
		s.InputTokens += e.Message.Usage.Input
		s.OutputTokens += e.Message.Usage.Output
		s.CacheRead += e.Message.Usage.CacheRead
		s.CacheWrite += e.Message.Usage.CacheCreate
	}
}
