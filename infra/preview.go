package infra

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"mpv-wallpaper-tui/domain"
	xdraw "golang.org/x/image/draw"
)

// extractFrameToFile extracts the first frame of videoPath into a JPEG temp
// file and returns its path. The file is reused across calls for the same video.
func extractFrameToFile(videoPath string) (string, error) {
	tmp := filepath.Join(os.TempDir(), "mpvwall_"+filepath.Base(videoPath)+".jpg")
	if err := exec.Command(
		"ffmpeg", "-y", "-i", videoPath,
		"-vframes", "1", "-q:v", "2", tmp,
	).Run(); err != nil {
		return "", fmt.Errorf("ffmpeg: %w", err)
	}
	return tmp, nil
}

// ── HalfBlockPreviewer ────────────────────────────────────────────────────────

// HalfBlockPreviewer renders wallpaper previews using ▀ half-block glyphs
// with 24-bit ANSI colour codes. No external tools beyond ffmpeg required.
type HalfBlockPreviewer struct{}

func NewHalfBlockPreviewer() *HalfBlockPreviewer { return &HalfBlockPreviewer{} }

func (p *HalfBlockPreviewer) Render(w domain.Wallpaper, cols, rows int) (string, error) {
	tmp, err := extractFrameToFile(w.Path)
	if err != nil {
		return "", fmt.Errorf("extract frame from %q: %w", w.Path, err)
	}
	f, err := os.Open(tmp)
	if err != nil {
		return "", err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("decode frame: %w", err)
	}
	return renderHalfBlocks(img, cols, rows), nil
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
		sb.WriteString("\033[0m") // reset colours at end of each row
		if row < rows-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}
