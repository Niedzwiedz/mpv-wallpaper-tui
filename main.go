package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"mpv-wallpaper-tui/config"
	"mpv-wallpaper-tui/infra"
	"mpv-wallpaper-tui/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	repo := infra.NewFSRepository(cfg.WallpapersPath)
	wallpapers, err := repo.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read wallpapers: %v\n", err)
		os.Exit(1)
	}

	model := ui.New(wallpapers, infra.NewMpvPlayer(), infra.NewAutoPreviewer())

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
