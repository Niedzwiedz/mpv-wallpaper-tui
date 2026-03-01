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
		if m.modalOpen {
			return m, m.handleModalKey(msg.String())
		}
		return m, m.handleKey(msg.String())
	}
	return m, nil
}

func (m *Model) handleKey(key string) tea.Cmd {
	switch key {
	case "ctrl+c", "q":
		return tea.Quit

	case "down", "j":
		if m.cursor < len(m.flat)-1 {
			m.cursor++
			m.clampScroll()
			return m.loadPreviewCmd(m.cursor)
		}

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.clampScroll()
			return m.loadPreviewCmd(m.cursor)
		}

	case "l", "right":
		m.handleOpen()
		return m.loadPreviewCmd(m.cursor)

	case "h", "left":
		m.handleClose()
		return m.loadPreviewCmd(m.cursor)

	case "m":
		m.savedMonitorCursor = m.monitorCursor
		m.modalOpen = true

	case "enter", " ":
		if w := m.currentWallpaper(); w != nil {
			return m.applyCmd(*w)
		}
	}
	return nil
}

func (m *Model) handleModalKey(key string) tea.Cmd {
	switch key {
	case "ctrl+c", "q", "esc", "m":
		m.monitorCursor = m.savedMonitorCursor
		m.modalOpen = false
	case "down", "j":
		if m.monitorCursor < len(m.monitors)-1 {
			m.monitorCursor++
		}
	case "up", "k":
		if m.monitorCursor > 0 {
			m.monitorCursor--
		}
	case "enter", " ":
		m.modalOpen = false
	}
	return nil
}

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

func (m *Model) applyCmd(w domain.Wallpaper) tea.Cmd {
	player := m.player
	monitor := m.selectedMonitor()
	return func() tea.Msg {
		player.Apply(w, monitor) //nolint:errcheck
		return nil
	}
}
