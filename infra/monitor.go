package infra

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"mpv-wallpaper-tui/domain"
)

// MonitorDetector discovers available display outputs.
// It tries detection methods in order: hyprctl → wlr-randr → xrandr.
// The special "ALL" entry is always first in the returned list.
type MonitorDetector struct{}

func NewMonitorDetector() *MonitorDetector { return &MonitorDetector{} }

func (d *MonitorDetector) List() []domain.Monitor {
	all := domain.Monitor{ID: "ALL"}

	for _, fn := range []func() []domain.Monitor{
		fromHyprctl,
		fromWlrRandr,
		fromXrandr,
	} {
		if monitors := fn(); len(monitors) > 0 {
			return append([]domain.Monitor{all}, monitors...)
		}
	}
	return []domain.Monitor{all}
}

// fromHyprctl detects monitors via hyprctl on Hyprland.
func fromHyprctl() []domain.Monitor {
	out, err := exec.Command("hyprctl", "monitors", "-j").Output()
	if err != nil {
		return nil
	}
	var raw []struct {
		Name   string `json:"name"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil
	}
	monitors := make([]domain.Monitor, 0, len(raw))
	for _, r := range raw {
		monitors = append(monitors, domain.Monitor{
			ID:         r.Name,
			Resolution: fmt.Sprintf("%d×%d", r.Width, r.Height),
		})
	}
	return monitors
}

// fromWlrRandr detects monitors via wlr-randr on wlr-based compositors.
func fromWlrRandr() []domain.Monitor {
	out, err := exec.Command("wlr-randr").Output()
	if err != nil {
		return nil
	}
	var monitors []domain.Monitor
	var current domain.Monitor
	resRe := regexp.MustCompile(`(\d+)x(\d+) px.*\(current\)`)
	for _, line := range strings.Split(string(out), "\n") {
		if line != "" && line[0] != ' ' && line[0] != '\t' {
			if current.ID != "" {
				monitors = append(monitors, current)
			}
			current = domain.Monitor{ID: strings.Fields(line)[0]}
		}
		if m := resRe.FindStringSubmatch(line); m != nil {
			current.Resolution = m[1] + "×" + m[2]
		}
	}
	if current.ID != "" {
		monitors = append(monitors, current)
	}
	return monitors
}

// fromXrandr detects connected monitors via xrandr on X11.
func fromXrandr() []domain.Monitor {
	out, err := exec.Command("xrandr", "--query").Output()
	if err != nil {
		return nil
	}
	connectedRe := regexp.MustCompile(`^(\S+) connected.*?(\d+)x(\d+)\+`)
	var monitors []domain.Monitor
	for _, line := range strings.Split(string(out), "\n") {
		if m := connectedRe.FindStringSubmatch(line); m != nil {
			monitors = append(monitors, domain.Monitor{
				ID:         m[1],
				Resolution: m[2] + "×" + m[3],
			})
		}
	}
	return monitors
}
