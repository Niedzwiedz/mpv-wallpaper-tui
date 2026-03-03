package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const appName = "mpv-wallpaper-tui"

// Config holds all application settings.
type Config struct {
	WallpapersPath string `toml:"wallpapers_path"`
	Animation      bool   `toml:"animation"`
	DefaultView    string `toml:"default_view"`
	Colors         Colors `toml:"colors"`
}

type Colors struct {
	Accent string `toml:"accent"`
	Muted  string `toml:"muted"`
}

// Load reads the config file, writing defaults if it does not yet exist.
// It also ensures the wallpapers directory exists (creating it if needed).
func Load() (*Config, error) {
	dir, err := appConfigDir()
	if err != nil {
		return nil, fmt.Errorf("locate config dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	cfgPath := filepath.Join(dir, "config.toml")
	cfg := defaults(dir)

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err := writeDefaults(cfgPath, cfg); err != nil {
			return nil, fmt.Errorf("write default config: %w", err)
		}
	} else {
		if _, err := toml.DecodeFile(cfgPath, cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", cfgPath, err)
		}
		cfg.WallpapersPath = expandHome(cfg.WallpapersPath)
	}

	if err := os.MkdirAll(cfg.WallpapersPath, 0o755); err != nil {
		return nil, fmt.Errorf("create wallpapers dir %q: %w", cfg.WallpapersPath, err)
	}

	return cfg, nil
}

func appConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}

func defaults(appDir string) *Config {
	return &Config{
		WallpapersPath: filepath.Join(appDir, "wallpapers"),
		Animation: true,
		DefaultView: "list",
		Colors: Colors{},
	}
}

// writeDefaults creates the config file with default values and inline comments.
func writeDefaults(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f,
		"# Path to the directory containing wallpaper video files.\n"+
			"wallpapers_path = %q\n\n"+
			"# Enable preview animation on startup.\n"+
			"animation = %t\n\n"+
			"# Which view to open on launch: \"list\" or \"grid\".\n"+
			"default_view = %q\n\n"+
			"# Colour overrides — ANSI index (e.g. \"2\") or hex (e.g. \"#ffa07a\").\n"+
			"# Leave empty to follow your terminal's ANSI palette.\n"+
			"[colors]\n"+
			"accent = \"\"\n"+
			"muted  = \"\"\n",
		cfg.WallpapersPath,
		cfg.Animation,
		cfg.DefaultView,
	)
	return err
}

// expandHome replaces a leading ~/ with the user's home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}
