package ui

import (
	"time"

	"mpv-wallpaper-tui/domain"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.clearCaches()
		if m.gridMode {
			m.clampGridScroll()
			return m, m.loadGridVisibleCmd()
		}
		m.clampScroll()
		return m, m.loadPreviewCmd(m.cursor)

	case previewReady:
		delete(m.loading, msg.path)
		m.previews[msg.path] = msg.content
		return m, nil

	case framesReady:
		delete(m.loading, msg.path)
		m.frames[msg.path] = msg.frames
		if w := m.currentWallpaper(); w != nil && w.Path == msg.path {
			return m, m.startTick()
		}
		return m, nil

	case animTick:
		if msg.gen != m.tickGen || !m.animating {
			return m, nil // stale tick or animation disabled
		}
		if w := m.currentWallpaper(); w != nil {
			if frames, ok := m.frames[w.Path]; ok && len(frames) > 1 {
				m.frameIdx = (m.frameIdx + 1) % len(frames)
				return m, m.tick()
			}
		}
		return m, nil

	case tea.KeyMsg:
		if m.modalOpen {
			return m, m.handleModalKey(msg.String())
		}
		if m.gridMode {
			return m, m.handleGridKey(msg.String())
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
			m.frameIdx = 0
			return tea.Batch(m.loadPreviewCmd(m.cursor), m.startTick())
		}

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.clampScroll()
			m.frameIdx = 0
			return tea.Batch(m.loadPreviewCmd(m.cursor), m.startTick())
		}

	case "l", "right":
		m.handleOpen()
		m.frameIdx = 0
		return tea.Batch(m.loadPreviewCmd(m.cursor), m.startTick())

	case "h", "left":
		m.handleClose()
		m.frameIdx = 0
		return tea.Batch(m.loadPreviewCmd(m.cursor), m.startTick())

	case "m":
		m.savedMonitorCursor = m.monitorCursor
		m.modalOpen = true

	case "a":
		m.animating = !m.animating
		if m.animating {
			return m.startTick()
		}

	case "v":
		m.gridMode = true
		m.clearCaches()
		if m.gridWallpapers == nil {
			m.gridWallpapers = collectWallpapers(m.roots)
		}
		m.gridCursor = 0
		m.gridScroll = 0
		m.frameIdx = 0
		return tea.Batch(m.loadGridVisibleCmd(), m.startTick())

	case "enter", " ":
		if w := m.currentWallpaper(); w != nil {
			return m.applyCmd(*w)
		}
	}
	return nil
}

func (m *Model) handleGridKey(key string) tea.Cmd {
	n := len(m.gridWallpapers)
	cols := m.gridCols()

	switch key {
	case "ctrl+c", "q":
		return tea.Quit

	case "v", "esc":
		m.gridMode = false
		m.clearCaches()
		m.frameIdx = 0
		return tea.Batch(m.loadPreviewCmd(m.cursor), m.startTick())

	case "right", "l":
		if m.gridCursor < n-1 && (m.gridCursor+1)%cols != 0 {
			m.gridCursor++
			m.frameIdx = 0
			m.clampGridScroll()
			return tea.Batch(m.loadGridVisibleCmd(), m.startTick())
		}

	case "left", "h":
		if m.gridCursor > 0 && m.gridCursor%cols != 0 {
			m.gridCursor--
			m.frameIdx = 0
			m.clampGridScroll()
			return tea.Batch(m.loadGridVisibleCmd(), m.startTick())
		}

	case "down", "j":
		if m.gridCursor+cols < n {
			m.gridCursor += cols
			m.frameIdx = 0
			m.clampGridScroll()
			return tea.Batch(m.loadGridVisibleCmd(), m.startTick())
		}

	case "up", "k":
		if m.gridCursor >= cols {
			m.gridCursor -= cols
			m.frameIdx = 0
			m.clampGridScroll()
			return tea.Batch(m.loadGridVisibleCmd(), m.startTick())
		}

	case "m":
		m.savedMonitorCursor = m.monitorCursor
		m.modalOpen = true

	case "a":
		m.animating = !m.animating
		if m.animating {
			return m.startTick()
		}

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

func (m *Model) clearCaches() {
	m.previews = make(map[string]string)
	m.frames = make(map[string][]string)
	m.loading = make(map[string]bool)
}

func (m *Model) loadPreviewCmd(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.flat) || m.width == 0 {
		return nil
	}
	e := m.flat[idx]
	if e.node.IsDir {
		return nil
	}
	cols, rows := m.previewDims()
	return m.loadWallpaperCmd(e.node.Wallpaper(), true, cols, rows)
}

// loadWallpaperCmd returns a command that renders w at the given dimensions.
// When animated is true and the previewer supports animation, two concurrent
// commands are returned: a fast static render and a slower full-frame render.
func (m *Model) loadWallpaperCmd(w domain.Wallpaper, animated bool, cols, rows int) tea.Cmd {
	if _, cached := m.previews[w.Path]; cached {
		return nil
	}
	if _, cached := m.frames[w.Path]; cached {
		return nil
	}
	if m.loading[w.Path] {
		return nil
	}
	m.loading[w.Path] = true
	previewer := m.previewer

	if animated {
		if ap, ok := previewer.(domain.AnimatedPreviewer); ok {
			staticCmd := func() tea.Msg {
				content, err := previewer.Render(w, cols, rows)
				if err != nil {
					return nil
				}
				return previewReady{path: w.Path, content: content}
			}
			framesCmd := func() tea.Msg {
				frames, err := ap.RenderFrames(w, cols, rows)
				if err != nil || len(frames) == 0 {
					return nil
				}
				return framesReady{path: w.Path, frames: frames}
			}
			return tea.Batch(staticCmd, framesCmd)
		}
	}

	return func() tea.Msg {
		content, err := previewer.Render(w, cols, rows)
		if err != nil {
			content = dimStyle.Render("  preview unavailable: " + err.Error())
		}
		return previewReady{path: w.Path, content: content}
	}
}

// loadGridVisibleCmd loads previews for all currently visible grid cells.
// The focused cell gets animated loading; others get static only.
func (m *Model) loadGridVisibleCmd() tea.Cmd {
	if m.width == 0 {
		return nil
	}
	numCols := m.gridCols()
	cellCols, cellRows := m.cellDims()
	first := m.gridScroll * numCols
	last := min((m.gridScroll+m.visibleGridRows())*numCols, len(m.gridWallpapers))

	cmds := make([]tea.Cmd, 0, last-first)
	for i := first; i < last; i++ {
		cmd := m.loadWallpaperCmd(m.gridWallpapers[i], i == m.gridCursor, cellCols, cellRows)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// startTick begins a new animation tick chain for the current wallpaper.
// Returns nil when animation is disabled or the current entry has ≤1 frame.
func (m *Model) startTick() tea.Cmd {
	if !m.animating {
		return nil
	}
	w := m.currentWallpaper()
	if w == nil {
		return nil
	}
	if frames, ok := m.frames[w.Path]; !ok || len(frames) <= 1 {
		return nil
	}
	m.tickGen++
	return m.tick()
}

func (m *Model) tick() tea.Cmd {
	gen := m.tickGen
	return tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg {
		return animTick{gen: gen}
	})
}

func (m *Model) applyCmd(w domain.Wallpaper) tea.Cmd {
	player := m.player
	monitor := m.selectedMonitor()
	return func() tea.Msg {
		player.Apply(w, monitor) //nolint:errcheck
		return nil
	}
}
