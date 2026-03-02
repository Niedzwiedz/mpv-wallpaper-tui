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
	if m.modalOpen {
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.monitorModal(),
		)
	}
	return m.mainView()
}

func (m *Model) mainView() string {
	innerH := m.availH() - 4
	panels := lipgloss.JoinHorizontal(lipgloss.Top,
		m.listPanel(innerH),
		m.previewPanel(innerH),
	)
	return lipgloss.NewStyle().
		Margin(marginV, marginH).
		Render(panels + "\n" + m.helpBar())
}

func (m *Model) helpBar() string {
	mon := m.selectedMonitor()
	anim := "on"
	if !m.animating {
		anim = "off"
	}
	return helpStyle.Render(
		"  ↑/k ↓/j  navigate    l/→ open    h/← close    ↵/space  apply" +
			"    m: monitor(" + mon.Label() + ")" +
			"    a: anim(" + anim + ")" +
			"    q  quit",
	)
}

// ── List panel ────────────────────────────────────────────────────────────────

func (m *Model) listPanel(innerH int) string {
	contentW := listPanelW - 4 // border (2) + manual padding (2)
	listH := m.wallpaperListH()

	lines := []string{
		titleStyle.Width(contentW).Render("Wallpapers"),
		"",
	}

	end := min(m.scroll+listH, len(m.flat))
	for i := m.scroll; i < end; i++ {
		label := truncate(m.entryLabel(m.flat[i]), contentW-3)
		if i == m.cursor {
			lines = append(lines, selectedStyle.Width(contentW).Render(label))
		} else {
			lines = append(lines, itemStyle.Width(contentW).Render(label))
		}
	}
	// Pad remaining rows so the panel height stays stable.
	for i := end - m.scroll; i < listH; i++ {
		lines = append(lines, "")
	}

	return panelStyle.
		Width(listPanelW - 2).
		Height(innerH).
		Render(strings.Join(lines, "\n"))
}

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

// ── Preview panel ─────────────────────────────────────────────────────────────

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
	mon := m.selectedMonitor()
	header := titleStyle.Render(w.Name) +
		"  " + dimStyle.Render("↵ apply → "+mon.Label())

	if frames, ok := m.frames[w.Path]; ok {
		if m.animating && len(frames) > 1 {
			return header + "\n" + frames[m.frameIdx%len(frames)]
		}
		return header + "\n" + frames[0]
	}
	if rendered, ok := m.previews[w.Path]; ok {
		return header + "\n" + rendered
	}
	return "\n  " + dimStyle.Render("Loading preview…")
}

// ── Monitor modal ─────────────────────────────────────────────────────────────

func (m *Model) monitorModal() string {
	contentW := m.modalContentW

	var lines []string
	lines = append(lines,
		titleStyle.Width(contentW-4).Render("Select Monitor"),
		"",
	)
	for i, mon := range m.monitors {
		label := mon.Label()
		if i == m.monitorCursor {
			lines = append(lines, selectedStyle.Width(contentW-4).Render(label))
		} else {
			lines = append(lines, itemStyle.Width(contentW-4).Render(label))
		}
	}
	lines = append(lines,
		"",
		dimStyle.Render("↑/k ↓/j  navigate    ↵  confirm    esc  cancel"),
	)

	return panelStyle.
		Width(contentW - 2).
		Render(strings.Join(lines, "\n"))
}

// ── Helpers ───────────────────────────────────────────────────────────────────

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

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
