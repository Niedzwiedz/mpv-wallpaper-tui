package infra

import (
	"fmt"
	"os/exec"
	"strings"

	"mpv-wallpaper-tui/domain"
)

// ChafaPreviewer renders wallpaper previews by piping an extracted frame
// through chafa, which produces Unicode/ANSI art suited to the terminal.
type ChafaPreviewer struct{}

func NewChafaPreviewer() *ChafaPreviewer { return &ChafaPreviewer{} }

func chafaSize(cols, rows int) string {
	return fmt.Sprintf("%dx%d", cols, rows)
}

func chafaRender(path, size string) (string, error) {
	out, err := exec.Command(
		"chafa",
		"--format", "symbols", // force ANSI text — never kitty/sixels/iterm
		"--size", size,
		"--stretch", // fill the requested area exactly
		path,
	).Output()
	if err != nil {
		return "", fmt.Errorf("chafa: %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func (p *ChafaPreviewer) Render(w domain.Wallpaper, cols, rows int) (string, error) {
	tmp, err := extractFrameToFile(w.Path)
	if err != nil {
		return "", fmt.Errorf("extract frame from %q: %w", w.Path, err)
	}
	return chafaRender(tmp, chafaSize(cols, rows))
}

func (p *ChafaPreviewer) RenderFrames(w domain.Wallpaper, cols, rows int) ([]string, error) {
	paths, err := extractFramesToFiles(w.Path, 20)
	if err != nil {
		return nil, err
	}
	size := chafaSize(cols, rows)
	return renderEachFrame(paths, func(fp string) (string, error) {
		return chafaRender(fp, size)
	})
}

// NewAutoPreviewer returns a ChafaPreviewer when chafa is on PATH,
// falling back to HalfBlockPreviewer otherwise.
func NewAutoPreviewer() domain.Previewer {
	if _, err := exec.LookPath("chafa"); err == nil {
		return NewChafaPreviewer()
	}
	return NewHalfBlockPreviewer()
}
