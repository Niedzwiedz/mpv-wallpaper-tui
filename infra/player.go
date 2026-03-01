package infra

import (
	"os/exec"
	"path/filepath"
	"syscall"

	"mpv-wallpaper-tui/domain"
)

// MpvPlayer applies wallpapers via mpvpaper, tracking one PID per monitor
// in a state file so only the relevant process is replaced on each apply.
type MpvPlayer struct {
	pids map[string]int
}

func NewMpvPlayer() *MpvPlayer { return &MpvPlayer{pids: readPIDs()} }

func (p *MpvPlayer) Apply(w domain.Wallpaper, monitor domain.Monitor) error {
	if monitor.ID == domain.AllMonitorsID {
		// Replacing ALL: kill every tracked process then clear the map.
		for _, pid := range p.pids {
			killIfMpvpaper(pid)
		}
		clear(p.pids)
	} else {
		// Replacing a specific monitor: kill the ALL process if one is running
		// (it covers every output, so it would conflict), then kill the
		// existing process for this monitor only.
		if pid, ok := p.pids[domain.AllMonitorsID]; ok {
			killIfMpvpaper(pid)
			delete(p.pids, domain.AllMonitorsID)
		}
		if pid, ok := p.pids[monitor.ID]; ok {
			killIfMpvpaper(pid)
			delete(p.pids, monitor.ID)
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

	p.pids[monitor.ID] = cmd.Process.Pid
	writePIDs(p.pids) //nolint:errcheck
	return nil
}
