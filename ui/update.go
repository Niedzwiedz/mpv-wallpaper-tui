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
		m.clampScroll()
		return m, m.loadPreviewCmd(m.cursor)

	case previewReady:
		delete(m.loading, msg.path)
		m.previews[msg.path] = msg.content
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			if m.focus == sectionWallpapers {
				m.focus = sectionMonitors
			} else {
				m.focus = sectionWallpapers
			}

		case "down", "j":
			return m, m.moveDown()

		case "up", "k":
			return m, m.moveUp()

		case "l", "right":
			if m.focus == sectionWallpapers {
				m.handleOpen()
				return m, m.loadPreviewCmd(m.cursor)
			}

		case "h", "left":
			if m.focus == sectionWallpapers {
				m.handleClose()
				return m, m.loadPreviewCmd(m.cursor)
			}

		case "enter", " ":
			if m.focus == sectionWallpapers {
				if w := m.currentWallpaper(); w != nil {
					return m, m.applyCmd(*w)
				}
			}
		}
	}
	return m, nil
}

func (m *Model) moveDown() tea.Cmd {
	switch m.focus {
	case sectionWallpapers:
		if m.cursor < len(m.flat)-1 {
			m.cursor++
			m.clampScroll()
			return m.loadPreviewCmd(m.cursor)
		}
	case sectionMonitors:
		if m.monitorCursor < len(m.monitors)-1 {
			m.monitorCursor++
		}
	}
	return nil
}

func (m *Model) moveUp() tea.Cmd {
	switch m.focus {
	case sectionWallpapers:
		if m.cursor > 0 {
			m.cursor--
			m.clampScroll()
			return m.loadPreviewCmd(m.cursor)
		}
	case sectionMonitors:
		if m.monitorCursor > 0 {
			m.monitorCursor--
		}
	}
	return nil
}

// handleOpen expands a collapsed directory, or steps into its first child.
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
		m.clampScroll()
	}
}

// handleClose collapses an expanded directory, or moves to the parent.
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
				m.clampScroll()
				break
			}
		}
	}
}

// loadPreviewCmd starts an async preview render for the entry at idx.
// No-op for directories, cached previews, or in-flight renders.
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

// applyCmd applies the wallpaper to the selected monitor.
func (m *Model) applyCmd(w domain.Wallpaper) tea.Cmd {
	player := m.player
	monitor := m.selectedMonitor()
	return func() tea.Msg {
		player.Apply(w, monitor) //nolint:errcheck
		return nil
	}
}
