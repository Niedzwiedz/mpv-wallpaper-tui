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
	innerH := m.availH() - 4
	help := helpStyle.Render("  ↑/k ↓/j  navigate    l/→ open    h/← close    tab  switch section    ↵/space  apply    q  quit")
	panels := lipgloss.JoinHorizontal(lipgloss.Top,
		m.listPanel(innerH),
		m.previewPanel(innerH),
	)
	return lipgloss.NewStyle().
		Margin(marginV, marginH).
		Render(panels + "\n" + help)
}

func (m *Model) listPanel(innerH int) string {
	contentW := listPanelW - 4 // border (2) + manual padding (2)
	listH := m.wallpaperListH()

	var lines []string

	// ── Wallpaper section ─────────────────────────────────────────────────────
	wallpaperFocused := m.focus == sectionWallpapers
	lines = append(lines,
		titleStyle.Width(contentW).Render("Wallpapers"),
		"",
	)

	end := min(m.scroll+listH, len(m.flat))
	for i := m.scroll; i < end; i++ {
		e := m.flat[i]
		label := truncate(m.entryLabel(e), contentW-3)
		switch {
		case i == m.cursor && wallpaperFocused:
			lines = append(lines, selectedStyle.Width(contentW).Render(label))
		case i == m.cursor:
			lines = append(lines, activeStyle.Width(contentW).Render(label))
		default:
			lines = append(lines, itemStyle.Width(contentW).Render(label))
		}
	}
	// Pad to fill the wallpaper section height so the separator stays anchored.
	for i := end - m.scroll; i < listH; i++ {
		lines = append(lines, "")
	}

	// ── Separator ─────────────────────────────────────────────────────────────
	lines = append(lines, dimStyle.Render(strings.Repeat("─", contentW)))

	// ── Monitor section ───────────────────────────────────────────────────────
	monitorFocused := m.focus == sectionMonitors
	lines = append(lines,
		titleStyle.Width(contentW).Render("Monitor"),
		"",
	)

	for i, mon := range m.monitors {
		label := truncate(mon.Label(), contentW-3)
		switch {
		case i == m.monitorCursor && monitorFocused:
			lines = append(lines, selectedStyle.Width(contentW).Render(label))
		case i == m.monitorCursor:
			lines = append(lines, activeStyle.Width(contentW).Render(label))
		default:
			lines = append(lines, itemStyle.Width(contentW).Render(label))
		}
	}

	return panelStyle.
		Width(listPanelW - 2).
		Height(innerH).
		Render(strings.Join(lines, "\n"))
}

// entryLabel builds the display string for a flat entry.
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
	w := m.availW() - listPanelW - 3
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
		mon := m.selectedMonitor()
		header := titleStyle.Render(w.Name) +
			"  " + dimStyle.Render("↵ apply  →  "+mon.Label())
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
