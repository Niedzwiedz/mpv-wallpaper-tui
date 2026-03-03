package infra

import (
	"fmt"
	"image"
	"os"
	"strings"

	"mpv-wallpaper-tui/domain"
	xdraw "golang.org/x/image/draw"
)

// HalfBlockPreviewer renders wallpaper previews using ▀ half-block glyphs
// with 24-bit ANSI colour codes. No external tools beyond ffmpeg required.
type HalfBlockPreviewer struct{}

func NewHalfBlockPreviewer() *HalfBlockPreviewer { return &HalfBlockPreviewer{} }

func (p *HalfBlockPreviewer) Name() string { return "builtin" }

func (p *HalfBlockPreviewer) Render(w domain.Wallpaper, cols, rows int) (string, error) {
	tmp, err := extractFrameToFile(w.Path)
	if err != nil {
		return "", fmt.Errorf("extract frame from %q: %w", w.Path, err)
	}
	img, err := decodeImageFromFile(tmp)
	if err != nil {
		return "", err
	}
	return renderHalfBlocks(img, cols, rows), nil
}

func (p *HalfBlockPreviewer) RenderFrames(w domain.Wallpaper, cols, rows int) ([]string, error) {
	paths, err := extractFramesToFiles(w.Path, 20)
	if err != nil {
		return nil, err
	}
	return renderEachFrame(paths, func(fp string) (string, error) {
		img, err := decodeImageFromFile(fp)
		if err != nil {
			return "", err
		}
		return renderHalfBlocks(img, cols, rows), nil
	})
}

func decodeImageFromFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %q: %w", path, err)
	}
	return img, nil
}

// renderHalfBlocks renders src into cols×rows terminal character cells.
//
// The ▀ (upper-half block) glyph is used for every cell:
//   - foreground colour = top pixel row
//   - background colour = bottom pixel row
//
// This yields an effective vertical resolution of rows×2 pixels.
// Requires a true-colour (24-bit) terminal.
func renderHalfBlocks(src image.Image, cols, rows int) string {
	dst := image.NewRGBA(image.Rect(0, 0, cols, rows*2))
	xdraw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)

	var sb strings.Builder
	for row := range rows {
		for col := range cols {
			top := dst.RGBAAt(col, row*2)
			bot := dst.RGBAAt(col, row*2+1)
			fmt.Fprintf(&sb,
				"\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm▀",
				top.R, top.G, top.B,
				bot.R, bot.G, bot.B,
			)
		}
		sb.WriteString("\033[0m") // reset colours at end of each row
		if row < rows-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}
