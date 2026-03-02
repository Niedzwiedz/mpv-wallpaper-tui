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

	gridMode       bool
	gridCursor     int
	gridScroll     int
	gridWallpapers []domain.Wallpaper

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
	if m.gridMode {
		if m.gridCursor >= len(m.gridWallpapers) {
			return nil
		}
		w := m.gridWallpapers[m.gridCursor]
		return &w
	}
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

// ── Grid layout helpers ───────────────────────────────────────────────────────

func (m *Model) gridCols() int {
	switch {
	case m.availW() >= 120:
		return 4
	case m.availW() >= 90:
		return 3
	case m.availW() >= 60:
		return 2
	default:
		return 1
	}
}

// cellDims returns the inner (content) dimensions of each grid cell.
// cols is the character width; rows is the preview height (excludes the name line).
// The /5 ratio keeps cells compact so multiple rows fit on screen at once.
func (m *Model) cellDims() (cols, rows int) {
	numCols := m.gridCols()
	cols = m.availW()/numCols - 1 // subtract 1-char right margin gap between cells
	if cols < 10 {
		cols = 10
	}
	rows = cols / 5
	if rows < 3 {
		rows = 3
	}
	return
}

// visibleGridRows returns how many full rows of cells fit on screen.
func (m *Model) visibleGridRows() int {
	_, cellRows := m.cellDims()
	cellH := cellRows + 1 + 2 // preview + name line + border
	v := (m.availH() - 1) / cellH
	if v < 1 {
		return 1
	}
	return v
}

func (m *Model) clampGridScroll() {
	numCols := m.gridCols()
	curRow := m.gridCursor / numCols
	visRows := m.visibleGridRows()
	if curRow < m.gridScroll {
		m.gridScroll = curRow
	}
	if curRow >= m.gridScroll+visRows {
		m.gridScroll = curRow - visRows + 1
	}
	if m.gridScroll < 0 {
		m.gridScroll = 0
	}
}

func collectWallpapers(nodes []*domain.Node) []domain.Wallpaper {
	var out []domain.Wallpaper
	for _, n := range nodes {
		if n.IsDir {
			out = append(out, collectWallpapers(n.Children)...)
		} else {
			out = append(out, n.Wallpaper())
		}
	}
	return out
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
