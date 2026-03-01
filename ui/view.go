package ui

import (
	"fmt"
	"strings"

	"mpv-wallpaper-tui/domain"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	if m.width == 0 {
		return "Loading…"
	}
	innerH := m.height - 4
	help := helpStyle.Render("  ↑/k ↓/j  navigate    l/→ open    h/← close    ↵/space  apply    q  quit")
	panels := lipgloss.JoinHorizontal(lipgloss.Top,
		m.listPanel(innerH),
		m.previewPanel(innerH),
	)
	return panels + "\n" + help
}

func (m *Model) listPanel(innerH int) string {
	contentW := listPanelW - 4 // border (2) + manual padding (2)

	lines := []string{
		titleStyle.Width(contentW).Render("Wallpapers"),
		"",
	}

	for i, e := range m.flat {
		label := m.entryLabel(e)
		label = truncate(label, contentW-3)
		if i == m.cursor {
			lines = append(lines, selectedStyle.Width(contentW).Render(label))
		} else {
			lines = append(lines, itemStyle.Width(contentW).Render(label))
		}
	}

	return panelStyle.
		Width(listPanelW - 2).
		Height(innerH).
		Render(strings.Join(lines, "\n"))
}

// entryLabel builds the display string for a flat entry, including
// depth indentation and a directory expand/collapse indicator.
func (m *Model) entryLabel(e flatEntry) string {
	indent := strings.Repeat("  ", e.depth)
	if e.node.IsDir {
		indicator := "▸ "
		if m.expanded[e.node.Path] {
			indicator = "▾ "
		}
		return indent + indicator + e.node.Name
	}
	return indent + "  " + e.node.Name
}

func (m *Model) previewPanel(innerH int) string {
	w := m.width - listPanelW - 3
	if w < 10 {
		w = 10
	}
	return panelStyle.
		Width(w).
		Height(innerH).
		Render(m.previewContent())
}

func (m *Model) previewContent() string {
	e := m.current()
	if e == nil {
		return "\n  " + dimStyle.Render("No wallpapers found")
	}
	if e.node.IsDir {
		n := countDescendants(e.node)
		return "\n  " + dimStyle.Render(fmt.Sprintf("folder: %s  (%d video%s)", e.node.Name, n, plural(n)))
	}
	w := e.node.Wallpaper()
	if rendered, ok := m.previews[w.Path]; ok {
		header := titleStyle.Render(w.Name) + "  " + dimStyle.Render("↵ apply")
		return header + "\n" + rendered
	}
	return "\n  " + dimStyle.Render("Loading preview…")
}

// countDescendants returns the total number of video file descendants of n.
func countDescendants(n *domain.Node) int {
	if !n.IsDir {
		return 1
	}
	total := 0
	for _, c := range n.Children {
		total += countDescendants(c)
	}
	return total
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// truncate shortens s to max runes, appending "…" if needed.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
