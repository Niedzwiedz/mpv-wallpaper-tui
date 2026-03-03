package ui

import "mpv-wallpaper-tui/domain"

// gridModel owns the navigation state for the grid view.
type gridModel struct {
	wallpapers []domain.Wallpaper
	cursor     int
	scroll     int
	pendingG   bool
}

func (g *gridModel) currentWallpaper() *domain.Wallpaper {
	if g.cursor >= len(g.wallpapers) {
		return nil
	}
	w := g.wallpapers[g.cursor]
	return &w
}

func (g *gridModel) cols(availW int) int {
	switch {
	case availW >= 120:
		return 4
	case availW >= 90:
		return 3
	case availW >= 60:
		return 2
	default:
		return 1
	}
}

func (g *gridModel) cellDims(availW int) (cols, rows int) {
	numCols := g.cols(availW)
	cols = max(10, availW/numCols-1)
	rows = max(3, cols/5)
	return
}

func (g *gridModel) visibleRows(availW, availH int) int {
	_, cellRows := g.cellDims(availW)
	return max(1, (availH-1)/(cellRows+1+2))
}

func (g *gridModel) clampScroll(availW, availH int) {
	numCols := g.cols(availW)
	curRow := g.cursor / numCols
	visRows := g.visibleRows(availW, availH)
	if curRow < g.scroll {
		g.scroll = curRow
	}
	if curRow >= g.scroll+visRows {
		g.scroll = curRow - visRows + 1
	}
	g.scroll = max(0, g.scroll)
}

func (g *gridModel) moveRight(availW, availH int) bool {
	n := len(g.wallpapers)
	cols := g.cols(availW)
	if g.cursor < n-1 && (g.cursor+1)%cols != 0 {
		g.cursor++
		g.clampScroll(availW, availH)
		return true
	}
	return false
}

func (g *gridModel) moveLeft(availW, availH int) bool {
	cols := g.cols(availW)
	if g.cursor > 0 && g.cursor%cols != 0 {
		g.cursor--
		g.clampScroll(availW, availH)
		return true
	}
	return false
}

func (g *gridModel) moveDown(availW, availH int) bool {
	cols := g.cols(availW)
	if g.cursor+cols < len(g.wallpapers) {
		g.cursor += cols
		g.clampScroll(availW, availH)
		return true
	}
	return false
}

func (g *gridModel) moveUp(availW, availH int) bool {
	cols := g.cols(availW)
	if g.cursor >= cols {
		g.cursor -= cols
		g.clampScroll(availW, availH)
		return true
	}
	return false
}

func (g *gridModel) goFirst(availW, availH int) {
	g.cursor = 0
	g.clampScroll(availW, availH)
}

func (g *gridModel) goLast(availW, availH int) {
	if n := len(g.wallpapers); n > 0 {
		g.cursor = n - 1
		g.clampScroll(availW, availH)
	}
}

// populate fills wallpapers from the tree on first entry into grid mode.
func (g *gridModel) populate(roots []*domain.Node) {
	if g.wallpapers == nil {
		g.wallpapers = collectWallpapers(roots)
	}
}

// reset returns cursor and scroll to the top (called each time grid mode is entered).
func (g *gridModel) reset(availW, availH int) {
	g.cursor = 0
	g.scroll = 0
	g.pendingG = false
	g.clampScroll(availW, availH)
}
