package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xdraw "golang.org/x/image/draw"
)

const listPanelW = 38 // total rendered width of left panel (content + border)

var wallpaperDir = func() string {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".config", "mpv_wallpapers")
	}
	return filepath.Join(cfg, "mpv_wallpapers")
}()

var (
	orange = lipgloss.Color("208")
	muted  = lipgloss.Color("240")
	black  = lipgloss.Color("0")

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(muted)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(orange)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Background(orange).
			Foreground(black).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(muted)

	dimStyle = lipgloss.NewStyle().
			Foreground(muted)
)

// ── Messages ──────────────────────────────────────────────────────────────────

type previewMsg struct {
	path    string
	content string
}

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	wallpapers []string
	cursor     int
	previews   map[string]string
	loading    map[string]bool
	width      int
	height     int
}

func initial() model {
	return model{
		wallpapers: listWallpapers(),
		previews:   make(map[string]string),
		loading:    make(map[string]bool),
	}
}

func listWallpapers() []string {
	entries, err := os.ReadDir(wallpaperDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".mp4", ".mkv", ".webm", ".avi", ".mov":
			out = append(out, e.Name())
		}
	}
	return out
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd { return nil }

// ── Update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, m.loadCmd(m.cursor)

	case previewMsg:
		delete(m.loading, msg.path)
		m.previews[msg.path] = msg.content
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "down", "j":
			if m.cursor < len(m.wallpapers)-1 {
				m.cursor++
				return m, m.loadCmd(m.cursor)
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				return m, m.loadCmd(m.cursor)
			}
		case "enter", " ":
			if len(m.wallpapers) > 0 {
				return m, cmdApplyWallpaper(m.videoPath(m.cursor))
			}
		}
	}
	return m, nil
}

func (m model) videoPath(idx int) string {
	if idx < 0 || idx >= len(m.wallpapers) {
		return ""
	}
	return filepath.Join(wallpaperDir, m.wallpapers[idx])
}

// previewDims returns usable (cols, rows) for the image preview area.
func (m model) previewDims() (int, int) {
	w := m.width - listPanelW - 3 // right panel content (subtract border)
	h := m.height - 6             // subtract border + header line + help bar
	if w < 20 {
		w = 60
	}
	if h < 5 {
		h = 20
	}
	return w, h
}

func (m model) loadCmd(idx int) tea.Cmd {
	path := m.videoPath(idx)
	if path == "" || m.loading[path] {
		return nil
	}
	if _, ok := m.previews[path]; ok {
		return nil
	}
	if m.width == 0 {
		return nil
	}
	m.loading[path] = true
	cols, rows := m.previewDims()
	return cmdGeneratePreview(path, cols, rows)
}

// ── Preview generation ────────────────────────────────────────────────────────

func cmdGeneratePreview(videoPath string, cols, rows int) tea.Cmd {
	return func() tea.Msg {
		// Use a per-file temp path so parallel loads don't collide.
		tmpJPG := "/tmp/mpvwall_" + filepath.Base(videoPath) + ".jpg"

		err := exec.Command(
			"ffmpeg", "-y", "-i", videoPath,
			"-vframes", "1", "-q:v", "2", tmpJPG,
		).Run()
		if err != nil {
			return previewMsg{path: videoPath, content: dimStyle.Render("  preview unavailable")}
		}

		f, err := os.Open(tmpJPG)
		if err != nil {
			return previewMsg{path: videoPath, content: dimStyle.Render("  could not open frame")}
		}
		defer f.Close()

		src, _, err := image.Decode(f)
		if err != nil {
			return previewMsg{path: videoPath, content: dimStyle.Render("  decode error")}
		}

		return previewMsg{path: videoPath, content: renderHalfBlocks(src, cols, rows)}
	}
}

// renderHalfBlocks renders src into cols×rows terminal character cells.
// Each cell shows two pixel rows using the ▀ half-block glyph:
//   - foreground color → top pixel
//   - background color → bottom pixel
//
// Requires a true-color (24-bit) terminal.
func renderHalfBlocks(src image.Image, cols, rows int) string {
	dst := image.NewRGBA(image.Rect(0, 0, cols, rows*2))
	xdraw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)

	var sb strings.Builder
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			top := dst.RGBAAt(col, row*2)
			bot := dst.RGBAAt(col, row*2+1)
			fmt.Fprintf(&sb,
				"\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm▀",
				top.R, top.G, top.B,
				bot.R, bot.G, bot.B,
			)
		}
		sb.WriteString("\033[0m") // reset at end of each row
		if row < rows-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// ── Apply wallpaper ───────────────────────────────────────────────────────────

func cmdApplyWallpaper(path string) tea.Cmd {
	return func() tea.Msg {
		// Kill any running mpvpaper instance first.
		exec.Command("pkill", "-f", "mpvpaper").Run()

		abs, _ := filepath.Abs(path)
		cmd := exec.Command(
			"mpvpaper", "-f", "-vs",
			"-o", "no-audio loop",
			"ALL", abs,
		)
		// Detach from our process group so mpvpaper survives TUI exit.
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		cmd.Start() //nolint:errcheck
		return nil
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.width == 0 {
		return "Loading…"
	}

	innerH := m.height - 4 // total height minus help bar + margins

	// ── Left panel: wallpaper list ────────────────────────────────────────────
	contentW := listPanelW - 4 // border (2) + padding we apply manually (2)

	var leftLines []string
	leftLines = append(leftLines, titleStyle.Width(contentW).Render("Wallpapers"))
	leftLines = append(leftLines, "")

	for i, name := range m.wallpapers {
		display := strings.TrimSuffix(name, filepath.Ext(name))
		if runes := []rune(display); len(runes) > contentW-3 {
			display = string(runes[:contentW-6]) + "…"
		}
		if i == m.cursor {
			leftLines = append(leftLines, selectedStyle.Width(contentW).Render(display))
		} else {
			leftLines = append(leftLines, itemStyle.Width(contentW).Render(display))
		}
	}

	leftPanel := panelStyle.
		Width(listPanelW-2). // lipgloss Width = content area; total = +2 for border
		Height(innerH).
		Render(strings.Join(leftLines, "\n"))

	// ── Right panel: image preview ────────────────────────────────────────────
	rightContentW := m.width - listPanelW - 3
	if rightContentW < 10 {
		rightContentW = 10
	}

	var rightContent string
	if path := m.videoPath(m.cursor); path != "" {
		if content, ok := m.previews[path]; ok {
			name := strings.TrimSuffix(m.wallpapers[m.cursor], filepath.Ext(m.wallpapers[m.cursor]))
			header := titleStyle.Render(name) + "  " + dimStyle.Render("↵ apply")
			rightContent = header + "\n" + content
		} else {
			rightContent = "\n  " + dimStyle.Render("Loading preview…")
		}
	} else {
		rightContent = "\n  " + dimStyle.Render("No wallpapers found in "+wallpaperDir)
	}

	rightPanel := panelStyle.
		Width(rightContentW).
		Height(innerH).
		Render(rightContent)

	// ── Help bar ──────────────────────────────────────────────────────────────
	help := helpStyle.Render("  ↑/k ↓/j  navigate    ↵/space  apply wallpaper    q  quit")

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel) + "\n" + help
}

// ── Entry point ───────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initial(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
