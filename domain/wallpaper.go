package domain

// Wallpaper is a video file that can be applied as a desktop wallpaper.
// It is the value passed to Player and Previewer.
type Wallpaper struct {
	Path string
	Name string
}

// Node is a single entry in the wallpaper file tree.
// It is either a directory (IsDir=true) or a video file (IsDir=false).
type Node struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*Node
}

// Wallpaper returns the Wallpaper value for a file node.
func (n *Node) Wallpaper() Wallpaper {
	return Wallpaper{Name: n.Name, Path: n.Path}
}

// AllMonitorsID is the sentinel monitor ID that targets every display output.
const AllMonitorsID = "ALL"

// Monitor represents a display output.
type Monitor struct {
	ID         string // passed to mpvpaper, e.g. "ALL", "eDP-1", "DP-3"
	Resolution string // display only, e.g. "3840×2160"; empty for ALL
}

// Label returns the display string shown in the monitor list.
func (m Monitor) Label() string {
	if m.Resolution == "" {
		return m.ID
	}
	return m.ID + "  " + m.Resolution
}

// Repository loads the wallpaper tree from a source.
type Repository interface {
	Tree() ([]*Node, error)
}

// MonitorRepository detects available display outputs.
type MonitorRepository interface {
	List() []Monitor
}

// Player applies a wallpaper to a specific monitor.
type Player interface {
	Apply(w Wallpaper, monitor Monitor) error
}

// Previewer renders a terminal image preview for a wallpaper.
type Previewer interface {
	Render(w Wallpaper, cols, rows int) (string, error)
}
