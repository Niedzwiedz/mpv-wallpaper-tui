package ui

import (
	"mpv-wallpaper-tui/domain"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, m.loadPreviewCmd(m.cursor)

	case previewReady:
		delete(m.loading, msg.path)
		m.previews[msg.path] = msg.content
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "down", "j":
			if m.cursor < len(m.wallpapers)-1 {
				m.cursor++
				return m, m.loadPreviewCmd(m.cursor)
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				return m, m.loadPreviewCmd(m.cursor)
			}
		case "enter", " ":
			if w := m.current(); w != nil {
				return m, m.applyCmd(*w)
			}
		}
	}
	return m, nil
}

// loadPreviewCmd starts an async preview render for wallpaper at idx.
// It is a no-op when the preview is already cached or being loaded.
func (m *Model) loadPreviewCmd(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.wallpapers) || m.width == 0 {
		return nil
	}
	w := m.wallpapers[idx]
	if _, cached := m.previews[w.Path]; cached {
		return nil
	}
	if m.loading[w.Path] {
		return nil
	}
	m.loading[w.Path] = true

	cols, rows := m.previewDims()
	previewer := m.previewer // capture for goroutine

	return func() tea.Msg {
		content, err := previewer.Render(w, cols, rows)
		if err != nil {
			content = dimStyle.Render("  preview unavailable: " + err.Error())
		}
		return previewReady{path: w.Path, content: content}
	}
}

// applyCmd fires off the player in a background goroutine.
func (m *Model) applyCmd(w domain.Wallpaper) tea.Cmd {
	player := m.player // capture for goroutine
	return func() tea.Msg {
		player.Apply(w) //nolint:errcheck
		return nil
	}
}
