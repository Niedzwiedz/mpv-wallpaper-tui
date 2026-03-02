package ui

import (
	"strings"

	"mpv-wallpaper-tui/domain"
)

// flatEntry is one visible row in the wallpaper tree.
type flatEntry struct {
	node  *domain.Node
	depth int
}

// listModel owns the tree-navigation state for the list view.
type listModel struct {
	roots    []*domain.Node
	expanded map[string]bool
	flat     []flatEntry
	cursor   int
	scroll   int
}

func newListModel(roots []*domain.Node) listModel {
	l := listModel{
		roots:    roots,
		expanded: make(map[string]bool),
	}
	l.flat = buildFlat(roots, l.expanded, 0)
	return l
}

func (l *listModel) current() *flatEntry {
	if l.cursor < 0 || l.cursor >= len(l.flat) {
		return nil
	}
	return &l.flat[l.cursor]
}

func (l *listModel) currentWallpaper() *domain.Wallpaper {
	e := l.current()
	if e == nil || e.node.IsDir {
		return nil
	}
	w := e.node.Wallpaper()
	return &w
}

func (l *listModel) wallpaperListH(availH int) int {
	h := availH - 4 - 2
	if h < 1 {
		h = 1
	}
	return h
}

func (l *listModel) clampScroll(availH int) {
	h := l.wallpaperListH(availH)
	if l.cursor < l.scroll {
		l.scroll = l.cursor
	}
	if l.cursor >= l.scroll+h {
		l.scroll = l.cursor - h + 1
	}
	if l.scroll < 0 {
		l.scroll = 0
	}
}

func (l *listModel) rebuildFlat(availH int) {
	l.flat = buildFlat(l.roots, l.expanded, 0)
	if l.cursor >= len(l.flat) {
		l.cursor = max(0, len(l.flat)-1)
	}
	l.clampScroll(availH)
}

func (l *listModel) moveDown(availH int) bool {
	if l.cursor < len(l.flat)-1 {
		l.cursor++
		l.clampScroll(availH)
		return true
	}
	return false
}

func (l *listModel) moveUp(availH int) bool {
	if l.cursor > 0 {
		l.cursor--
		l.clampScroll(availH)
		return true
	}
	return false
}

func (l *listModel) open(availH int) {
	e := l.current()
	if e == nil || !e.node.IsDir {
		return
	}
	if !l.expanded[e.node.Path] {
		l.expanded[e.node.Path] = true
		l.rebuildFlat(availH)
	} else if l.cursor+1 < len(l.flat) && l.flat[l.cursor+1].depth > e.depth {
		l.cursor++
		l.clampScroll(availH)
	}
}

func (l *listModel) close(availH int) {
	e := l.current()
	if e == nil {
		return
	}
	if e.node.IsDir && l.expanded[e.node.Path] {
		delete(l.expanded, e.node.Path)
		l.rebuildFlat(availH)
		return
	}
	if e.depth > 0 {
		for i := l.cursor - 1; i >= 0; i-- {
			if l.flat[i].depth == e.depth-1 {
				l.cursor = i
				if l.flat[i].node.IsDir {
					delete(l.expanded, l.flat[i].node.Path)
					l.rebuildFlat(availH)
				}
				l.clampScroll(availH)
				break
			}
		}
	}
}

func (l *listModel) entryLabel(e flatEntry) string {
	indent := strings.Repeat("  ", e.depth)
	if e.node.IsDir {
		indicator := "▸ "
		if l.expanded[e.node.Path] {
			indicator = "▾ "
		}
		return indent + indicator + e.node.Name
	}
	return indent + "  " + e.node.Name
}

// ── Package-level tree helpers ─────────────────────────────────────────────────

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
