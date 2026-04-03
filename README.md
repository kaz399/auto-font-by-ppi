# auto-font-by-ppi

Automatically adjust GNOME font settings based on the target display PPI.

`gnome-auto-font-by-ppi.sh` detects the active display, calculates its pixels per inch (PPI), selects a profile, and updates GNOME text scaling and font sizes.

## Features

- Supports Xorg through `xrandr`
- Supports GNOME Wayland through Mutter DisplayConfig D-Bus
- Selects the primary display automatically, or a specific display by name
- Applies either text scaling only or full GNOME font settings
- Handles broken EDID physical size values with manual monitor size overrides
- Supports command-line options, environment variables, and a config file

## Requirements

- Bash
- `gsettings`
- `xrandr` for Xorg sessions
- `gdbus` and `python3` for GNOME Wayland sessions

## Usage

```bash
./gnome-auto-font-by-ppi.sh --dry-run
./gnome-auto-font-by-ppi.sh --apply
./gnome-auto-font-by-ppi.sh --apply --display eDP-1
./gnome-auto-font-by-ppi.sh --apply --mode scaling_only
```

## Manual Monitor Size Override

Some monitors report invalid physical size values through EDID. A common failure mode is that the reported size in millimeters matches the pixel resolution, which produces obviously wrong PPI values.

When that happens, provide the diagonal size manually.

### Command line

```bash
./gnome-auto-font-by-ppi.sh --dry-run --display-diagonal HDMI-1=31.5
./gnome-auto-font-by-ppi.sh --dry-run --display-diagonal HDMI-1=31.5 --display-diagonal eDP-1=14.0
```

### Environment variable

```bash
MONITOR_DIAGONAL_OVERRIDES="HDMI-1=31.5,eDP-1=14.0" ./gnome-auto-font-by-ppi.sh --dry-run
```

## Config File

By default, the script loads:

```text
~/.config/gnome-auto-font-by-ppi.conf
```

You can also specify a custom file:

```bash
./gnome-auto-font-by-ppi.sh --config ./gnome-auto-font-by-ppi.conf.example --dry-run
```

Example config:

```bash
PREFERRED_DISPLAY="HDMI-1"
APPLY_MODE="full_fonts"
MONITOR_DIAGONAL_OVERRIDES="HDMI-1=31.5,eDP-1=14.0"
MIN_REASONABLE_PPI=50
MAX_REASONABLE_PPI=400
```

See [gnome-auto-font-by-ppi.conf.example](./gnome-auto-font-by-ppi.conf.example).

## Detection and Fallback Behavior

The script treats display metrics as suspicious when:

- Width or height in millimeters is zero or negative
- Width and height in millimeters exactly match the pixel resolution
- Calculated PPI is outside the configured reasonable range

If suspicious metrics are detected and a manual diagonal override exists for that display, the script recalculates the physical size and PPI from the diagonal inches.

## Options

```text
--dry-run
--apply
--display NAME
--display-diagonal NAME=INCHES
--config PATH
--mode scaling_only|full_fonts
--verbose
--help
```

## Profiles

The script currently uses these PPI profiles:

```text
110  -> scaling 1.00, UI 11, document 11, monospace 10, titlebar 11
140  -> scaling 1.10, UI 12, document 12, monospace 11, titlebar 12
180  -> scaling 1.25, UI 13, document 13, monospace 12, titlebar 13
240  -> scaling 1.40, UI 14, document 14, monospace 13, titlebar 14
9999 -> scaling 1.60, UI 16, document 16, monospace 14, titlebar 16
```

## License

MIT License. See the source files for the full header text.
