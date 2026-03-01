package infra

import (
	"os"
	"path/filepath"
	"strings"

	"mpv-wallpaper-tui/domain"
)

var videoExts = map[string]bool{
	".mp4": true, ".mkv": true, ".webm": true, ".avi": true, ".mov": true,
}

// FSRepository loads wallpapers from a local directory.
type FSRepository struct {
	dir string
}

func NewFSRepository(dir string) *FSRepository {
	return &FSRepository{dir: dir}
}

func (r *FSRepository) List() ([]domain.Wallpaper, error) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, err
	}
	var out []domain.Wallpaper
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if !videoExts[ext] {
			continue
		}
		out = append(out, domain.Wallpaper{
			Path: filepath.Join(r.dir, e.Name()),
			Name: strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())),
		})
	}
	return out, nil
}
