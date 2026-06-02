# auto-font-by-ppi

A Go-based CLI utility to automatically adjust GNOME font settings and kitty terminal font sizes based on the target display's PPI (Pixels Per Inch).

The tool detects the active display, calculates its PPI, selects a matching font profile, and updates GNOME's `text-scaling-factor` and various font sizes, along with kitty terminal's font size.

## Features

- **Multi-backend detection**:
  - Uses `xrandr` for Xorg sessions.
  - Uses Mutter DisplayConfig over D-Bus for GNOME Wayland sessions.
- **Robust display selection**:
  - Automatically selects the primary display, or allows manual naming (`--display`).
  - Supports the `--auto` flag to ignore the preferred display from the config file and detect the best active display dynamically.
- **Flexible target application**:
  - Applies text scaling only or full GNOME font settings.
  - Dynamically updates active kitty terminal instances (kitty target).
- **Abnormal metrics correction & fallback**:
  - Supports manual diagonal overrides (`--display-diagonal` or config) for monitors reporting bogus EDID physical dimensions.
  - Manual overrides always take precedence over detected physical sizes.
  - Supplements missing GNOME Wayland millimeter values using `xrandr` when available.
  - Falls back to an assumed `PPI=100` if no usable physical size data can be found.
- **Per-profile customization**:
  - Supports a dedicated `kitty_font_size` for each PPI profile.

## Requirements

- `gsettings` (required for applying GNOME desktop settings)
- Xorg sessions: `xrandr`
- GNOME Wayland sessions: `gdbus` (used automatically along with `xrandr` for fallback completion)

## Installation

Build the binary from the root of the repository:

```bash
# Build the binary
go build -o auto-font-by-ppi ./cmd/auto-font-by-ppi
```

## Usage

Basic commands:

```bash
# Run with dry-run to preview changes
./auto-font-by-ppi --dry-run

# Apply settings (Apply mode is default)
./auto-font-by-ppi

# Select a specific display by name
./auto-font-by-ppi --display eDP-1

# Ignore preferred display in config and auto-detect instead
./auto-font-by-ppi --auto
```

Temporary overrides (applied to the selected display after detection):

```bash
# Assume the selected display is 27 inches
./auto-font-by-ppi --inch 27

# Assume the selected display has 110 PPI
./auto-font-by-ppi --ppi 110
```

Overriding physical diagonal dimensions for specific displays:

```bash
# Set manual diagonal override for HDMI-1
./auto-font-by-ppi --display-diagonal HDMI-1=31.5

# Specify multiple overrides
./auto-font-by-ppi --display-diagonal HDMI-1=31.5 --display-diagonal eDP-1=14.0
```

## Config File

By default, the utility reads settings from:

```text
~/.config/auto-font-by-ppi/config.toml
```

If the file does not exist, a default configuration with sample settings is automatically generated at that path on the first run.

To specify a custom configuration file path:

```bash
# Load custom configuration path
./auto-font-by-ppi --config ./custom-config.toml
```

See [examples/config.toml](./examples/config.toml) for detailed formatting and options.

### Key Configuration Fields
* `dry_run`: Set to `true` to always preview changes without applying them.
* `preferred_display`: Name of the display to prioritize (e.g., `"HDMI-1"`). This can be temporarily ignored using the `--auto` flag.
* `target_names`: Array of active targets to apply (e.g., `["gnome", "kitty"]`). Note that kitty is opt-in and disabled by default.
* `display.diagonal_overrides`: Map of manual diagonal dimensions in inches per connector name.

## Detection and Fallback Behavior

Display physical dimensions are considered suspicious and ignored when:

- Millimeter width or height is zero or negative.
- Millimeter dimensions exactly match the pixel resolution.
- The calculated PPI falls outside the configured reasonable range (default: 50 to 400).

When suspicious metrics are detected, the utility corrects values using the following precedence:

1. Uses `display.diagonal_overrides` or `--display-diagonal` if an override exists.
2. Uses GNOME Wayland detection first when that backend is active.
3. If GNOME Wayland detection does not provide usable millimeter values, supplements them from `xrandr`.
4. If no usable physical size data is available after detection, assumes `PPI=100` and derives dimensions from the resolution.

## Options

```text
--dry-run
    Show what would be changed, but do not apply settings.
--display NAME
    Select a specific display by name instead of auto-detecting.
--auto
    Ignore the preferred display in the config file and auto-detect instead. This option cannot be specified together with --display.
--inch INCHES
    Treat the selected display as having the specified diagonal size in inches. This overrides config-based display size settings.
--ppi VALUE
    Treat the selected display as having the specified PPI. This overrides config-based display size settings.
--display-diagonal NAME=INCHES
    Override a display's physical diagonal size.
--config PATH
    Load configuration from the specified TOML file.
--verbose
    Print more diagnostic information.
--help, -h
    Show the help message and exit.
```

## Profiles

Font settings are selected based on the calculated PPI from the following profiles:

```text
110  -> scaling 1.00, UI 11, document 11, monospace 10, titlebar 11, kitty 10
140  -> scaling 1.00, UI 12, document 12, monospace 11, titlebar 12, kitty 11
180  -> scaling 1.00, UI 13, document 13, monospace 12, titlebar 13, kitty 12
240  -> scaling 1.00, UI 14, document 14, monospace 13, titlebar 14, kitty 13
9999 -> scaling 1.60, UI 16, document 16, monospace 14, titlebar 16, kitty 14
```

## License

MIT License. See the source files for the full header text.
