package domain

// Wallpaper is the core domain entity.
type Wallpaper struct {
	Path string // absolute or relative path to the video file
	Name string // display name (filename without extension)
}

// Repository loads wallpapers from a source.
type Repository interface {
	List() ([]Wallpaper, error)
}

// Player applies a wallpaper to the desktop.
type Player interface {
	Apply(Wallpaper) error
}

// Previewer renders a terminal image preview for a wallpaper.
type Previewer interface {
	Render(w Wallpaper, cols, rows int) (string, error)
}
