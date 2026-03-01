package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"mpv-wallpaper-tui/infra"
	"mpv-wallpaper-tui/ui"
)

func main() {
	repo := infra.NewFSRepository(wallpaperDir())
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

func wallpaperDir() string {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".config", "mpv_wallpapers")
	}
	return filepath.Join(cfg, "mpv_wallpapers")
}
