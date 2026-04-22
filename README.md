# auto-font-by-ppi

Automatically adjust GNOME font settings based on the target display PPI.

`gnome-auto-font-by-ppi.sh` detects the active display, calculates its pixels per inch (PPI), selects a profile, and updates GNOME text scaling and font sizes.

## Features

- Supports Xorg through `xrandr`
- Supports GNOME Wayland through Mutter DisplayConfig D-Bus
- Selects the primary display automatically, or a specific display by name
- Applies either text scaling only or full GNOME font settings
- Handles broken EDID physical size values with manual monitor size overrides
- Applies manual diagonal overrides before using detected physical size values on the Go CLI
- On the Go CLI, supplements missing GNOME Wayland millimeter data from `xrandr` when available
- On the Go CLI, falls back to an assumed `PPI=100` when no usable physical size data is available
- On the Go CLI, supports a dedicated `kitty_font_size` per profile
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

## Go CLI Status

The Go rewrite currently provides an experimental `auto-font-by-ppi` CLI alongside the Bash script.

Example usage:

```bash
go run ./cmd/auto-font-by-ppi --dry-run
go run ./cmd/auto-font-by-ppi --display eDP-1
go run ./cmd/auto-font-by-ppi --dry-run --display-diagonal HDMI-1=31.5
```

The Go CLI reads TOML config from:

```text
~/.config/auto-font-by-ppi/config.toml
```

If that file does not exist, the Go CLI creates a default sample config there automatically and then loads it.

By default, the Go CLI applies settings immediately. Use `--dry-run` or `dry_run = true` when you want a preview only.

See [examples/config.toml](./examples/config.toml) for the current config format.

`target_names` is the single source of truth for which targets run and in what order. The `target.*` sections contain per-target settings only. By default, kitty is opt-in, so add `"kitty"` to `target_names` when you want it applied.

For kitty, each profile can define `kitty_font_size`. It accepts integers or decimals such as `12.5`. The default `target.kitty.font_size_field` is `kitty_font_size`, and if that field is not set in an older config, the Go CLI falls back to `monospace_font_size`.

## Manual Monitor Size Override

Some monitors report invalid physical size values through EDID. A common failure mode is that the reported size in millimeters matches the pixel resolution, which produces obviously wrong PPI values.

When that happens, provide the diagonal size manually.

On the Go CLI, a configured manual diagonal override always takes precedence for the matching display. Without an override, the Go CLI first uses GNOME Wayland detection, then supplements missing millimeter values from `xrandr` when available, and finally falls back to an assumed `PPI=100` if no usable physical size data is available.

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

The Bash script recalculates the physical size and PPI from the diagonal inches when suspicious metrics are detected and a manual diagonal override exists for that display.

On the Go CLI, the current precedence is:

- Use `display.diagonal_overrides` or `--display-diagonal` when an override exists for the display.
- Otherwise, use GNOME Wayland detection first when that backend is selected.
- If GNOME Wayland detection does not provide usable millimeter values, supplement them from `xrandr` when available.
- If no usable physical size data is available after detection, assume `PPI=100` and derive the physical size from the detected resolution.

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

The current Go CLI supports:

```text
--dry-run
--display NAME
--display-diagonal NAME=INCHES
--config PATH
--verbose
--help
```

## Profiles

The script currently uses these PPI profiles:

```text
110  -> scaling 1.00, UI 11, document 11, monospace 10, titlebar 11, kitty 10
140  -> scaling 1.00, UI 12, document 12, monospace 11, titlebar 12, kitty 11
180  -> scaling 1.00, UI 13, document 13, monospace 12, titlebar 13, kitty 12
240  -> scaling 1.00, UI 14, document 14, monospace 13, titlebar 14, kitty 13
9999 -> scaling 1.60, UI 16, document 16, monospace 14, titlebar 16, kitty 14
```

## License

MIT License. See the source files for the full header text.
