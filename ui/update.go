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
			m.grid.clampScroll(m.availW(), m.availH())
			return m, m.loadGridVisibleCmd()
		}
		m.list.clampScroll(m.availH())
		return m, m.loadPreviewCmd(m.list.cursor)

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
		if m.list.moveDown(m.availH()) {
			m.frameIdx = 0
			return tea.Batch(m.loadPreviewCmd(m.list.cursor), m.startTick())
		}

	case "up", "k":
		if m.list.moveUp(m.availH()) {
			m.frameIdx = 0
			return tea.Batch(m.loadPreviewCmd(m.list.cursor), m.startTick())
		}

	case "l", "right":
		m.list.open(m.availH())
		m.frameIdx = 0
		return tea.Batch(m.loadPreviewCmd(m.list.cursor), m.startTick())

	case "h", "left":
		m.list.close(m.availH())
		m.frameIdx = 0
		return tea.Batch(m.loadPreviewCmd(m.list.cursor), m.startTick())

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
		m.grid.populate(m.list.roots)
		m.grid.reset(m.availW(), m.availH())
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
	pendingG := m.grid.pendingG
	m.grid.pendingG = false

	switch key {
	case "ctrl+c", "q":
		return tea.Quit

	case "v":
		m.gridMode = false
		m.clearCaches()
		m.frameIdx = 0
		return tea.Batch(m.loadPreviewCmd(m.list.cursor), m.startTick())

	case "right", "l":
		if m.grid.moveRight(m.availW(), m.availH()) {
			m.frameIdx = 0
			return tea.Batch(m.loadGridVisibleCmd(), m.startTick())
		}

	case "left", "h":
		if m.grid.moveLeft(m.availW(), m.availH()) {
			m.frameIdx = 0
			return tea.Batch(m.loadGridVisibleCmd(), m.startTick())
		}

	case "down", "j":
		if m.grid.moveDown(m.availW(), m.availH()) {
			m.frameIdx = 0
			return tea.Batch(m.loadGridVisibleCmd(), m.startTick())
		}

	case "up", "k":
		if m.grid.moveUp(m.availW(), m.availH()) {
			m.frameIdx = 0
			return tea.Batch(m.loadGridVisibleCmd(), m.startTick())
		}

	case "g":
		if pendingG {
			m.grid.goFirst(m.availW(), m.availH())
			m.frameIdx = 0
			return tea.Batch(m.loadGridVisibleCmd(), m.startTick())
		}
		m.grid.pendingG = true

	case "G":
		m.grid.goLast(m.availW(), m.availH())
		m.frameIdx = 0
		return tea.Batch(m.loadGridVisibleCmd(), m.startTick())

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

func (m *Model) clearCaches() {
	m.previews = make(map[string]string)
	m.frames = make(map[string][]string)
	m.loading = make(map[string]bool)
}

func (m *Model) loadPreviewCmd(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.list.flat) || m.width == 0 {
		return nil
	}
	e := m.list.flat[idx]
	if e.node.IsDir {
		return nil
	}
	cols, rows := m.previewDims()
	return m.loadWallpaperCmd(e.node.Wallpaper(), true, cols, rows)
}

// loadWallpaperCmd returns a command that renders w at the given dimensions.
// When animated is true and the previewer supports animation, frames are loaded
// even if a static preview is already cached.
func (m *Model) loadWallpaperCmd(w domain.Wallpaper, animated bool, cols, rows int) tea.Cmd {
	_, hasStatic := m.previews[w.Path]
	_, hasFrames := m.frames[w.Path]

	if hasFrames {
		return nil // already fully loaded
	}
	if !animated && hasStatic {
		return nil // static-only load already done
	}
	if m.loading[w.Path] {
		return nil // load already in flight
	}
	m.loading[w.Path] = true
	previewer := m.previewer

	if animated {
		if ap, ok := previewer.(domain.AnimatedPreviewer); ok {
			var cmds []tea.Cmd
			if !hasStatic {
				cmds = append(cmds, func() tea.Msg {
					content, err := previewer.Render(w, cols, rows)
					if err != nil {
						return nil
					}
					return previewReady{path: w.Path, content: content}
				})
			}
			cmds = append(cmds, func() tea.Msg {
				frames, err := ap.RenderFrames(w, cols, rows)
				if err != nil || len(frames) == 0 {
					return nil
				}
				return framesReady{path: w.Path, frames: frames}
			})
			return tea.Batch(cmds...)
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
	numCols := m.grid.cols(m.availW())
	cellCols, cellRows := m.grid.cellDims(m.availW())
	first := m.grid.scroll * numCols
	last := min((m.grid.scroll+m.grid.visibleRows(m.availW(), m.availH()))*numCols, len(m.grid.wallpapers))

	cmds := make([]tea.Cmd, 0, last-first)
	for i := first; i < last; i++ {
		cmd := m.loadWallpaperCmd(m.grid.wallpapers[i], i == m.grid.cursor, cellCols, cellRows)
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
