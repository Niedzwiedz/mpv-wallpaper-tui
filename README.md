# mpv-wallpaper-tui

A terminal UI for browsing and applying animated video wallpapers via [mpvpaper](https://github.com/GhostNaN/mpvpaper).

![screenshot](showcase/screenshot.png)

## Dependencies

### Runtime

| Tool | Purpose |
|------|---------|
| [mpvpaper](https://github.com/GhostNaN/mpvpaper) | Sets animated video wallpapers |
| [ffmpeg](https://ffmpeg.org) | Extracts the first frame for preview |
| [chafa](https://hpjansson.org/chafa/) *(optional)* | Higher-quality terminal image preview |

If `chafa` is on your `$PATH` the app uses it for previews; otherwise it falls back to a built-in half-block Unicode renderer that requires no extra tools.

**Arch Linux:**
```bash
sudo pacman -S mpv ffmpeg chafa
yay -S mpvpaper   # AUR
```

### Build

- [Go](https://go.dev) 1.22 or later

## Build & Install

Clone and install to `~/.local/bin` in one step:

```bash
git clone <repo-url>
cd mpv-wallpaper-tui
make install
```

`make install` builds the binary and places it at `~/.local/bin/mpv-wallpaper-tui`.
`~/.local/bin` is the XDG-standard per-user binary directory and is on `$PATH` by default on most modern Linux distributions.

Other targets:

```bash
make build      # build only, outputs ./mpv-wallpaper-tui
make uninstall  # remove from ~/.local/bin
```

### Autostart on login

To restore your last wallpaper automatically on login:

```bash
make install-autostart
```

This detects the best method for your system:

| Condition | Method used |
|-----------|-------------|
| systemd user session running (`/run/user/UID/systemd` exists) | systemd user service (`~/.config/systemd/user/mpv-wallpaper.service`) |
| No systemd user session | XDG autostart entry (`~/.config/autostart/mpv-wallpaper.desktop`) |

Both methods run `mpv-wallpaper-tui --restore` once at session start to reapply the last selected wallpaper.

To remove:

```bash
make uninstall-autostart
```

You can also install each method explicitly:

```bash
make install-service    # systemd only
make uninstall-service

# XDG autostart works on any DE that implements the XDG autostart spec
# (GNOME, KDE, XFCE, and Wayland compositors paired with dex or similar)
install -Dm644 mpv-wallpaper.desktop ~/.config/autostart/mpv-wallpaper.desktop
```

## Preview animation

![demo](showcase/animation.gif)

## Usage

```bash
mpv-wallpaper-tui
```

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `↵` / `space` | Apply selected wallpaper |
| `q` / `Ctrl+C` | Quit |

Applying a wallpaper kills any running `mpvpaper` instance and starts a new one.
The wallpaper keeps playing after you quit the TUI.

## Configuration

On first launch the app creates its config directory automatically.

### Directory layout

```
~/.config/mpv-wallpaper-tui/
├── config.toml      # application config
└── wallpapers/      # default wallpaper directory
```

### config.toml

```toml
# Path to the directory containing wallpaper video files.
wallpapers_path = "/home/user/.config/mpv-wallpaper-tui/wallpapers"
```

| Option | Default | Description |
|--------|---------|-------------|
| `wallpapers_path` | `~/.config/mpv-wallpaper-tui/wallpapers` | Directory scanned for video files |

`wallpapers_path` supports `~/` expansion. Supported video formats: `.mp4`, `.mkv`, `.webm`, `.avi`, `.mov`.

The `wallpapers/` directory is created automatically if it does not exist. Drop your video files there and relaunch.
