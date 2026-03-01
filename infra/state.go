package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// pidFilePath returns the path to the PID state file.
// Uses $XDG_STATE_HOME or falls back to ~/.local/state.
func pidFilePath() (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(base, "mpv-wallpaper-tui", "pids.json"), nil
}

// readPIDs loads the monitor→pid map from disk.
// Returns an empty map on any error (missing file, bad JSON, etc.).
func readPIDs() map[string]int {
	path, err := pidFilePath()
	if err != nil {
		return make(map[string]int)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return make(map[string]int)
	}
	var pids map[string]int
	if err := json.Unmarshal(data, &pids); err != nil {
		return make(map[string]int)
	}
	return pids
}

// writePIDs persists the monitor→pid map to disk.
func writePIDs(pids map[string]int) error {
	path, err := pidFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(pids, "", "  ")
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
