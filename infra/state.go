package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// persistedState holds everything that must survive across TUI restarts.
type persistedState struct {
	PIDs       map[string]int    `json:"pids"`
	Wallpapers map[string]string `json:"wallpapers"`
}

func emptyState() persistedState {
	return persistedState{
		PIDs:       make(map[string]int),
		Wallpapers: make(map[string]string),
	}
}

// stateFilePath returns the path to the unified state file.
// Uses $XDG_STATE_HOME or falls back to ~/.local/state.
func stateFilePath() (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(base, "mpv-wallpaper-tui", "state.json"), nil
}

// readState loads the persisted state from disk.
// Returns an empty state on any error (missing file, bad JSON, etc.).
func readState() persistedState {
	path, err := stateFilePath()
	if err != nil {
		return emptyState()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return emptyState()
	}
	var s persistedState
	if err := json.Unmarshal(data, &s); err != nil {
		return emptyState()
	}
	if s.PIDs == nil {
		s.PIDs = make(map[string]int)
	}
	if s.Wallpapers == nil {
		s.Wallpapers = make(map[string]string)
	}
	return s
}

// writeState persists the state to disk.
func writeState(s persistedState) error {
	path, err := stateFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// killIfMpvpaper kills pid only if /proc/<pid>/cmdline confirms it is mpvpaper.
// Silently ignores errors (stale PID, process already gone, etc.).
func killIfMpvpaper(pid int) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return // process is gone
	}
	// cmdline entries are NUL-separated; check the executable name
	if !strings.Contains(string(data), "mpvpaper") {
		return // PID was reused by a different process
	}
	if proc, err := os.FindProcess(pid); err == nil {
		proc.Kill() //nolint:errcheck
	}
}
