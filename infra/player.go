package infra

import (
	"os/exec"
	"path/filepath"
	"syscall"

	"mpv-wallpaper-tui/domain"
)

// MpvPlayer applies wallpapers via mpvpaper, tracking one PID per monitor
// in a state file so only the relevant process is replaced on each apply.
type MpvPlayer struct{}

func NewMpvPlayer() *MpvPlayer { return &MpvPlayer{} }

func (p *MpvPlayer) Apply(w domain.Wallpaper, monitor domain.Monitor) error {
	pids := readPIDs()

	if monitor.ID == "ALL" {
		// Replacing ALL: kill every tracked process.
		for _, pid := range pids {
			killIfMpvpaper(pid)
		}
		pids = make(map[string]int)
	} else {
		// Replacing a specific monitor: kill the ALL process if one is running
		// (it covers every output, so it would conflict), then kill the
		// existing process for this monitor only.
		if pid, ok := pids["ALL"]; ok {
			killIfMpvpaper(pid)
			delete(pids, "ALL")
		}
		if pid, ok := pids[monitor.ID]; ok {
			killIfMpvpaper(pid)
			delete(pids, monitor.ID)
		}
	}

	abs, err := filepath.Abs(w.Path)
	if err != nil {
		return err
	}
	cmd := exec.Command("mpvpaper", "-f", "-vs", "-o", "no-audio loop", monitor.ID, abs)
	// New session so mpvpaper outlives the TUI process.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}

	pids[monitor.ID] = cmd.Process.Pid
	writePIDs(pids) //nolint:errcheck
	return nil
}
