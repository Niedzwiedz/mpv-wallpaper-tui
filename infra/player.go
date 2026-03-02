package infra

import (
	"os/exec"
	"path/filepath"
	"syscall"

	"mpv-wallpaper-tui/domain"
)

// MpvPlayer applies wallpapers via mpvpaper, tracking one PID and one wallpaper
// path per monitor in a state file so only the relevant process is replaced on
// each apply and the setup can be restored on login.
type MpvPlayer struct {
	state persistedState
}

func NewMpvPlayer() *MpvPlayer { return &MpvPlayer{state: readState()} }

func (p *MpvPlayer) Apply(w domain.Wallpaper, monitor domain.Monitor) error {
	if monitor.ID == domain.AllMonitorsID {
		// Replacing ALL: kill every tracked process then clear the state.
		for _, pid := range p.state.PIDs {
			killIfMpvpaper(pid)
		}
		p.state = emptyState()
	} else {
		// Replacing a specific monitor: kill the ALL process if one is running
		// (it covers every output, so it would conflict), then kill the
		// existing process for this monitor only.
		if pid, ok := p.state.PIDs[domain.AllMonitorsID]; ok {
			killIfMpvpaper(pid)
			delete(p.state.PIDs, domain.AllMonitorsID)
			delete(p.state.Wallpapers, domain.AllMonitorsID)
		}
		if pid, ok := p.state.PIDs[monitor.ID]; ok {
			killIfMpvpaper(pid)
			delete(p.state.PIDs, monitor.ID)
		}
	}

	abs, err := filepath.Abs(w.Path)
	if err != nil {
		return err
	}
	cmd := exec.Command("mpvpaper", "-l", "bottom", "-f", "-vs", "-o", "no-audio loop", monitor.ID, abs)
	// New session so mpvpaper outlives the TUI process.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}

	p.state.PIDs[monitor.ID] = cmd.Process.Pid
	p.state.Wallpapers[monitor.ID] = abs
	writeState(p.state) //nolint:errcheck
	return nil
}

// Restore replays the saved monitor→wallpaper mapping without launching the TUI.
// Intended for use at login via a systemd user service.
func (p *MpvPlayer) Restore() {
	for monitorID, wallpaperPath := range p.state.Wallpapers {
		w := domain.Wallpaper{Path: wallpaperPath, Name: filepath.Base(wallpaperPath)}
		m := domain.Monitor{ID: monitorID}
		p.Apply(w, m) //nolint:errcheck
	}
}
