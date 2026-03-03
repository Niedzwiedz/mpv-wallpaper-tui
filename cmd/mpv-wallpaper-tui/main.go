package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"mpv-wallpaper-tui/internal/config"
	"mpv-wallpaper-tui/internal/infra"
	"mpv-wallpaper-tui/internal/ui"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--restore" {
		infra.NewMpvPlayer().Restore()
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	roots, err := infra.NewFSRepository(cfg.WallpapersPath).Tree()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read wallpapers: %v\n", err)
		os.Exit(1)
	}

	monitors := infra.NewMonitorDetector().List()

	model := ui.New(roots, monitors, infra.NewMpvPlayer(), infra.NewAutoPreviewer())

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
