package ui

import (
	"mpv-wallpaper-tui/domain"

	tea "github.com/charmbracelet/bubbletea"
)

const listPanelW = 38 // total rendered width of the left panel including border

// Model is the root Bubble Tea model.
// Dependencies are injected via New; the model itself owns only UI state.
type Model struct {
	wallpapers []domain.Wallpaper
	cursor     int
	previews   map[string]string // path → rendered half-block string
	loading    map[string]bool   // paths currently being rendered

	width  int
	height int

	player    domain.Player
	previewer domain.Previewer
}

// New constructs a Model with injected infrastructure dependencies.
func New(wallpapers []domain.Wallpaper, player domain.Player, previewer domain.Previewer) *Model {
	return &Model{
		wallpapers: wallpapers,
		previews:   make(map[string]string),
		loading:    make(map[string]bool),
		player:     player,
		previewer:  previewer,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

// current returns a pointer to the wallpaper under the cursor, or nil.
func (m *Model) current() *domain.Wallpaper {
	if m.cursor < 0 || m.cursor >= len(m.wallpapers) {
		return nil
	}
	return &m.wallpapers[m.cursor]
}

// previewDims returns the usable (cols, rows) for the image preview area.
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
