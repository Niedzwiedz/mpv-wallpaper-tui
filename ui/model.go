package ui

import (
	"mpv-wallpaper-tui/domain"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	listPanelW = 38 // total rendered width of the left panel including border
	marginH    = 2  // left AND right margin in terminal columns (each side)
	marginV    = 1  // top AND bottom margin in terminal rows (each side)
)

// section identifies which part of the left panel has keyboard focus.
type section uint8

const (
	sectionWallpapers section = iota
	sectionMonitors
)

// flatEntry is one visible row in the wallpaper tree.
type flatEntry struct {
	node  *domain.Node
	depth int
}

// Model is the root Bubble Tea model.
// Dependencies are injected via New; the model owns only UI state.
type Model struct {
	roots    []*domain.Node
	expanded map[string]bool // set of expanded directory paths
	flat     []flatEntry     // visible depth-first flattened tree
	cursor   int
	scroll   int // first visible index in the wallpaper list

	monitors      []domain.Monitor
	monitorCursor int
	focus         section

	previews map[string]string // path → rendered preview string
	loading  map[string]bool   // paths currently being rendered

	width  int
	height int

	player    domain.Player
	previewer domain.Previewer
}

// New constructs a Model with injected infrastructure dependencies.
func New(
	roots []*domain.Node,
	monitors []domain.Monitor,
	player domain.Player,
	previewer domain.Previewer,
) *Model {
	m := &Model{
		roots:     roots,
		expanded:  make(map[string]bool),
		monitors:  monitors,
		previews:  make(map[string]string),
		loading:   make(map[string]bool),
		player:    player,
		previewer: previewer,
	}
	m.flat = buildFlat(roots, m.expanded, 0)
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

// current returns the flat entry under the cursor, or nil.
func (m *Model) current() *flatEntry {
	if m.cursor < 0 || m.cursor >= len(m.flat) {
		return nil
	}
	return &m.flat[m.cursor]
}

// currentWallpaper returns the Wallpaper for the cursor if it is a file node.
func (m *Model) currentWallpaper() *domain.Wallpaper {
	e := m.current()
	if e == nil || e.node.IsDir {
		return nil
	}
	w := e.node.Wallpaper()
	return &w
}

// selectedMonitor returns the currently selected monitor.
func (m *Model) selectedMonitor() domain.Monitor {
	if m.monitorCursor < 0 || m.monitorCursor >= len(m.monitors) {
		return domain.Monitor{ID: "ALL"}
	}
	return m.monitors[m.monitorCursor]
}

// availW returns the usable terminal width after applying horizontal margins.
func (m *Model) availW() int { return m.width - marginH*2 }

// availH returns the usable terminal height after applying vertical margins.
func (m *Model) availH() int { return m.height - marginV*2 }

// previewDims returns usable (cols, rows) for the image preview area.
func (m *Model) previewDims() (cols, rows int) {
	cols = m.availW() - listPanelW - 3
	rows = m.availH() - 6
	if cols < 20 {
		cols = 60
	}
	if rows < 5 {
		rows = 20
	}
	return
}

// wallpaperListH returns how many rows the wallpaper list can show.
// It subtracts the monitor section and decorations from the panel height.
func (m *Model) wallpaperListH() int {
	innerH := m.availH() - 4
	// header (title + blank) = 2
	// separator = 1
	// monitor section (title + blank + items) = 2 + len(monitors)
	overhead := 2 + 1 + 2 + len(m.monitors)
	h := innerH - overhead
	if h < 1 {
		h = 1
	}
	return h
}

// clampScroll adjusts m.scroll so the cursor stays within the visible window.
func (m *Model) clampScroll() {
	h := m.wallpaperListH()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+h {
		m.scroll = m.cursor - h + 1
	}
	if m.scroll < 0 {
		m.scroll = 0
	}
}

// rebuildFlat recomputes the visible flat list and keeps cursor/scroll valid.
func (m *Model) rebuildFlat() {
	m.flat = buildFlat(m.roots, m.expanded, 0)
	if m.cursor >= len(m.flat) {
		m.cursor = max(0, len(m.flat)-1)
	}
	m.clampScroll()
}

// buildFlat produces a depth-first flat list of visible nodes.
func buildFlat(nodes []*domain.Node, expanded map[string]bool, depth int) []flatEntry {
	var out []flatEntry
	for _, n := range nodes {
		out = append(out, flatEntry{node: n, depth: depth})
		if n.IsDir && expanded[n.Path] {
			out = append(out, buildFlat(n.Children, expanded, depth+1)...)
		}
	}
	return out
}
