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
		m.previews = make(map[string]string)
		m.loading = make(map[string]bool)
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
			if m.cursor < len(m.flat)-1 {
				m.cursor++
				return m, m.loadPreviewCmd(m.cursor)
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				return m, m.loadPreviewCmd(m.cursor)
			}

		case "l", "right":
			m.handleOpen()
			return m, m.loadPreviewCmd(m.cursor)

		case "h", "left":
			m.handleClose()
			return m, m.loadPreviewCmd(m.cursor)

		case "enter", " ":
			if w := m.currentWallpaper(); w != nil {
				return m, m.applyCmd(*w)
			}
		}
	}
	return m, nil
}

// handleOpen expands a collapsed directory, or steps into its first child
// if it is already expanded.
func (m *Model) handleOpen() {
	e := m.current()
	if e == nil || !e.node.IsDir {
		return
	}
	if !m.expanded[e.node.Path] {
		m.expanded[e.node.Path] = true
		m.rebuildFlat()
	} else if m.cursor+1 < len(m.flat) && m.flat[m.cursor+1].depth > e.depth {
		m.cursor++
	}
}

// handleClose collapses an expanded directory, or moves the cursor up to the
// parent directory (collapsing it) when on a collapsed dir or a file.
func (m *Model) handleClose() {
	e := m.current()
	if e == nil {
		return
	}
	if e.node.IsDir && m.expanded[e.node.Path] {
		delete(m.expanded, e.node.Path)
		m.rebuildFlat()
		return
	}
	if e.depth > 0 {
		for i := m.cursor - 1; i >= 0; i-- {
			if m.flat[i].depth == e.depth-1 {
				m.cursor = i
				if m.flat[i].node.IsDir {
					delete(m.expanded, m.flat[i].node.Path)
					m.rebuildFlat()
				}
				break
			}
		}
	}
}

// loadPreviewCmd starts an async preview render for the entry at idx.
// No-op for directory entries, already-cached previews, or in-flight renders.
func (m *Model) loadPreviewCmd(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.flat) || m.width == 0 {
		return nil
	}
	e := m.flat[idx]
	if e.node.IsDir {
		return nil
	}
	w := e.node.Wallpaper()
	if _, cached := m.previews[w.Path]; cached {
		return nil
	}
	if m.loading[w.Path] {
		return nil
	}
	m.loading[w.Path] = true

	cols, rows := m.previewDims()
	previewer := m.previewer

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
	player := m.player
	return func() tea.Msg {
		player.Apply(w) //nolint:errcheck
		return nil
	}
}
