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

// flatEntry is one visible row in the wallpaper tree.
type flatEntry struct {
	node  *domain.Node
	depth int
}

// Model is the root Bubble Tea model.
// Dependencies are injected via New; the model owns only UI state.
type Model struct {
	roots    []*domain.Node
	expanded map[string]bool
	flat     []flatEntry
	cursor   int
	scroll   int // first visible index in the wallpaper list

	monitors           []domain.Monitor
	monitorCursor      int
	savedMonitorCursor int
	modalOpen          bool
	modalContentW      int // pre-computed modal width

	previews map[string]string
	frames   map[string][]string
	loading  map[string]bool
	frameIdx  int
	tickGen   int
	animating bool

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
	modalW := 36
	for _, mon := range monitors {
		if l := len([]rune(mon.Label())) + 6; l > modalW {
			modalW = l
		}
	}
	m := &Model{
		roots:         roots,
		expanded:      make(map[string]bool),
		monitors:      monitors,
		previews:      make(map[string]string),
		frames:        make(map[string][]string),
		loading:       make(map[string]bool),
		player:        player,
		previewer:     previewer,
		modalContentW: modalW,
		animating:     true,
	}
	m.flat = buildFlat(roots, m.expanded, 0)
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) current() *flatEntry {
	if m.cursor < 0 || m.cursor >= len(m.flat) {
		return nil
	}
	return &m.flat[m.cursor]
}

func (m *Model) currentWallpaper() *domain.Wallpaper {
	e := m.current()
	if e == nil || e.node.IsDir {
		return nil
	}
	w := e.node.Wallpaper()
	return &w
}

func (m *Model) selectedMonitor() domain.Monitor {
	if m.monitorCursor < 0 || m.monitorCursor >= len(m.monitors) {
		return domain.Monitor{ID: domain.AllMonitorsID}
	}
	return m.monitors[m.monitorCursor]
}

func (m *Model) availW() int { return m.width - marginH*2 }
func (m *Model) availH() int { return m.height - marginV*2 }

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

// wallpaperListH returns how many rows the wallpaper list can display.
func (m *Model) wallpaperListH() int {
	h := m.availH() - 4 - 2 // panel innerH minus list header (title + blank)
	if h < 1 {
		h = 1
	}
	return h
}

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

func (m *Model) rebuildFlat() {
	m.flat = buildFlat(m.roots, m.expanded, 0)
	if m.cursor >= len(m.flat) {
		m.cursor = max(0, len(m.flat)-1)
	}
	m.clampScroll()
}

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
