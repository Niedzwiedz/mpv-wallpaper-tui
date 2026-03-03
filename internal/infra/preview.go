package infra

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
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

// extractFramesToFiles extracts up to maxFrames frames from videoPath at 10 fps
// into a per-video temp directory and returns their paths. The directory is
// reused on subsequent calls for the same video (disk cache).
func extractFramesToFiles(videoPath string, maxFrames int) ([]string, error) {
	dir := filepath.Join(os.TempDir(), "mpvwall_anim_"+filepath.Base(videoPath))

	if paths := collectFramePaths(dir, maxFrames); len(paths) > 0 {
		return paths, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if err := exec.Command(
		"ffmpeg", "-y", "-i", videoPath,
		"-vf", "fps=10",
		"-frames:v", strconv.Itoa(maxFrames),
		"-q:v", "3",
		filepath.Join(dir, "f%04d.jpg"),
	).Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %w", err)
	}
	paths := collectFramePaths(dir, maxFrames)
	if len(paths) == 0 {
		return nil, fmt.Errorf("no frames extracted from %q", videoPath)
	}
	return paths, nil
}

func collectFramePaths(dir string, max int) []string {
	matches, _ := filepath.Glob(filepath.Join(dir, "f????.jpg"))
	if len(matches) > max {
		matches = matches[:max]
	}
	return matches
}

// renderEachFrame calls render for each path concurrently, preserving order
// and silently skipping frames where render returns an error.
func renderEachFrame(paths []string, render func(string) (string, error)) ([]string, error) {
	type result struct {
		idx   int
		frame string
	}
	results := make([]result, 0, len(paths))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i, fp := range paths {
		wg.Add(1)
		go func(i int, fp string) {
			defer wg.Done()
			frame, err := render(fp)
			if err != nil {
				return
			}
			mu.Lock()
			results = append(results, result{i, frame})
			mu.Unlock()
		}(i, fp)
	}
	wg.Wait()
	sort.Slice(results, func(a, b int) bool { return results[a].idx < results[b].idx })
	frames := make([]string, 0, len(results))
	for _, r := range results {
		frames = append(frames, r.frame)
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("no frames rendered")
	}
	return frames, nil
}
