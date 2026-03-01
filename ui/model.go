package ui

import (
	"mpv-wallpaper-tui/domain"

	tea "github.com/charmbracelet/bubbletea"
)

const listPanelW = 38 // total rendered width of the left panel including border

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

	cursor int

	previews map[string]string // path → rendered preview string
	loading  map[string]bool   // paths currently being rendered

	width  int
	height int

	player    domain.Player
	previewer domain.Previewer
}

// New constructs a Model with injected infrastructure dependencies.
func New(roots []*domain.Node, player domain.Player, previewer domain.Previewer) *Model {
	m := &Model{
		roots:     roots,
		expanded:  make(map[string]bool),
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

// previewDims returns usable (cols, rows) for the image preview area.
func (m *Model) previewDims() (cols, rows int) {
	cols = m.width - listPanelW - 3
	rows = m.height - 6
	if cols < 20 {
		cols = 60
	}
	if rows < 5 {
		rows = 20
	}
	return
}

// rebuildFlat recomputes the visible flat list from the tree and expanded state.
// The cursor is clamped to stay within bounds after a rebuild.
func (m *Model) rebuildFlat() {
	m.flat = buildFlat(m.roots, m.expanded, 0)
	if m.cursor >= len(m.flat) {
		m.cursor = max(0, len(m.flat)-1)
	}
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
