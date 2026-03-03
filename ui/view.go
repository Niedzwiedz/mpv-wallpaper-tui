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
	if m.gridMode {
		return m.gridView()
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
	renderer := m.previewer.Name()
	anim := "on"

	var left string
	var bar string

	if !m.animating {
		anim = "off"
	}

	right := " RENDER: " + renderer + "  "
	if m.gridMode {
		left = 	"  h/j/k/l  navigate    ↵/space  apply" +
				"    m: monitor(" + mon.Label() + ")" +
				"    a: anim(" + anim + ")" +
				"    tab: list" +
				"    q  quit"
	} else {
		left = 	"  ↑/k ↓/j  navigate    l/→ open    h/← close    ↵/space  apply" +
				"    m: monitor(" + mon.Label() + ")" +
				"    a: anim(" + anim + ")" +
				"    tab: grid" +
				"    q  quit"
	}
	gap := m.availW() - lipgloss.Width(left) - lipgloss.Width(right)

	if gap < 1 {
		bar = left
	} else {
		bar = left + strings.Repeat(" ", gap) + right
	}

	return helpStyle.Render(bar)
}

// ── List panel ────────────────────────────────────────────────────────────────

func (m *Model) listPanel(innerH int) string {
	contentW := listPanelW - 4 // border (2) + manual padding (2)
	listH := m.list.wallpaperListH(m.availH())

	lines := []string{
		titleStyle.Width(contentW).Render("Wallpapers"),
		"",
	}

	end := min(m.list.scroll+listH, len(m.list.flat))
	for i := m.list.scroll; i < end; i++ {
		label := truncate(m.list.entryLabel(m.list.flat[i]), contentW-3)
		if i == m.list.cursor {
			lines = append(lines, selectedStyle.Width(contentW).Render(label))
		} else {
			lines = append(lines, itemStyle.Width(contentW).Render(label))
		}
	}
	// Pad remaining rows so the panel height stays stable.
	for i := end - m.list.scroll; i < listH; i++ {
		lines = append(lines, "")
	}

	return panelStyle.
		Width(listPanelW - 2).
		Height(innerH).
		Render(strings.Join(lines, "\n"))
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
	e := m.list.current()
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

// ── Grid view ─────────────────────────────────────────────────────────────────

func (m *Model) gridView() string {
	numCols := m.grid.cols(m.availW())
	cellCols, cellRows := m.grid.cellDims(m.availW())

	var rows []string
	lastRow := m.grid.scroll + m.grid.visibleRows(m.availW(), m.availH())
	for row := m.grid.scroll; row < lastRow; row++ {
		first := row * numCols
		if first >= len(m.grid.wallpapers) {
			break
		}
		last := min(first+numCols, len(m.grid.wallpapers))
		var cells []string
		for i := first; i < last; i++ {
			cells = append(cells, m.renderCell(i, cellCols, cellRows))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	grid := strings.Join(rows, "\n")
	return lipgloss.NewStyle().
		Margin(marginV, marginH).
		Render(grid + "\n" + m.helpBar())
}

func (m *Model) renderCell(idx, cellCols, cellRows int) string {
	w := m.grid.wallpapers[idx]
	isFocused := idx == m.grid.cursor

	var preview string
	if frames, ok := m.frames[w.Path]; ok {
		if isFocused && m.animating && len(frames) > 1 {
			preview = frames[m.frameIdx%len(frames)]
		} else {
			preview = frames[0]
		}
	} else if rendered, ok := m.previews[w.Path]; ok {
		preview = rendered
	} else {
		preview = dimStyle.Render("loading…")
	}

	nameFg := muted
	if isFocused {
		nameFg = accent
	}
	name := lipgloss.NewStyle().
		Foreground(nameFg).
		Bold(isFocused).
		Width(cellCols).
		Render(truncate(w.Name, cellCols))

	return lipgloss.NewStyle().
		Width(cellCols).
		Height(cellRows + 1).
		MarginRight(1).
		Render(name + "\n" + preview)
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
