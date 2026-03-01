package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	if m.width == 0 {
		return "Loading…"
	}
	innerH := m.height - 4
	help := helpStyle.Render("  ↑/k ↓/j  navigate    ↵/space  apply    q  quit")
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
	for i, w := range m.wallpapers {
		label := truncate(w.Name, contentW-3)
		if i == m.cursor {
			lines = append(lines, selectedStyle.Width(contentW).Render(label))
		} else {
			lines = append(lines, itemStyle.Width(contentW).Render(label))
		}
	}

	return panelStyle.
		Width(listPanelW - 2). // lipgloss Width = content; border adds 2
		Height(innerH).
		Render(strings.Join(lines, "\n"))
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
	w := m.current()
	if w == nil {
		return "\n  " + dimStyle.Render("No wallpapers found")
	}
	if rendered, ok := m.previews[w.Path]; ok {
		header := titleStyle.Render(w.Name) + "  " + dimStyle.Render("↵ apply")
		return header + "\n" + rendered
	}
	return "\n  " + dimStyle.Render("Loading preview…")
}

// truncate shortens s to max runes, appending "…" if needed.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
