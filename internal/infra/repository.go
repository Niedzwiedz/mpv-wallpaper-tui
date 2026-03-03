package infra

import (
	"os"
	"path/filepath"
	"strings"

	"mpv-wallpaper-tui/internal/domain"
)

var videoExts = map[string]bool{
	".mp4": true, ".mkv": true, ".webm": true, ".avi": true, ".mov": true,
}

// FSRepository loads the wallpaper tree from a local directory.
type FSRepository struct {
	dir string
}

func NewFSRepository(dir string) *FSRepository {
	return &FSRepository{dir: dir}
}

func (r *FSRepository) Tree() ([]*domain.Node, error) {
	return scanDir(r.dir)
}

// scanDir recursively walks dir, returning nodes for video files and
// sub-directories that contain at least one video descendant.
func scanDir(dir string) ([]*domain.Node, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var nodes []*domain.Node
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if e.IsDir() {
			children, _ := scanDir(path) // skip unreadable sub-dirs silently
			if len(children) > 0 {
				nodes = append(nodes, &domain.Node{
					Name:     e.Name(),
					Path:     path,
					IsDir:    true,
					Children: children,
				})
			}
		} else if videoExts[strings.ToLower(filepath.Ext(e.Name()))] {
			nodes = append(nodes, &domain.Node{
				Name:  strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())),
				Path:  path,
				IsDir: false,
			})
		}
	}
	return nodes, nil
}
