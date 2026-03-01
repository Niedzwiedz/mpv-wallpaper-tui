package infra

import (
	"os/exec"
	"path/filepath"
	"syscall"

	"mpv-wallpaper-tui/domain"
)

// MpvPlayer applies wallpapers via mpvpaper.
type MpvPlayer struct{}

func NewMpvPlayer() *MpvPlayer { return &MpvPlayer{} }

func (p *MpvPlayer) Apply(w domain.Wallpaper) error {
	// Kill any existing mpvpaper instance before launching a new one.
	exec.Command("pkill", "-f", "mpvpaper").Run() //nolint:errcheck

	abs, err := filepath.Abs(w.Path)
	if err != nil {
		return err
	}
	cmd := exec.Command("mpvpaper", "-f", "-vs", "-o", "no-audio loop", "ALL", abs)
	// Start in a new session so mpvpaper outlives the TUI process.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Start()
}
