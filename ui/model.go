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

// Model is the root Bubble Tea model.
// Dependencies are injected via New; the model owns only UI state.
type Model struct {
	list     listModel
	grid     gridModel
	gridMode bool

	monitors           []domain.Monitor
	monitorCursor      int
	savedMonitorCursor int
	modalOpen          bool
	modalContentW      int // pre-computed modal width

	// Shared preview cache (used by both list and grid rendering paths).
	previews map[string]string
	frames   map[string][]string
	loading  map[string]bool

	// Animation clock (shared across both views).
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
	return &Model{
		list:          newListModel(roots),
		grid:          gridModel{},
		monitors:      monitors,
		previews:      make(map[string]string),
		frames:        make(map[string][]string),
		loading:       make(map[string]bool),
		player:        player,
		previewer:     previewer,
		modalContentW: modalW,
		animating:     true,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) currentWallpaper() *domain.Wallpaper {
	if m.gridMode {
		return m.grid.currentWallpaper()
	}
	return m.list.currentWallpaper()
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
