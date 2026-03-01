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

// Repository loads the wallpaper tree from a source.
type Repository interface {
	Tree() ([]*Node, error)
}

// Player applies a wallpaper to the desktop.
type Player interface {
	Apply(Wallpaper) error
}

// Previewer renders a terminal image preview for a wallpaper.
type Previewer interface {
	Render(w Wallpaper, cols, rows int) (string, error)
}
