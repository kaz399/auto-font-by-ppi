# Go Rewrite Status

## Purpose

This document summarizes the Go rewrite direction for `auto-font-by-ppi`, the work completed so far, and the remaining tasks.

The current rewrite work is being developed on the `go-rewrite-skeleton` branch. The existing Bash implementation remains in the repository as a reference and fallback during the migration.

## Rewrite Goals

- Replace the monolithic Bash implementation with a maintainable Go CLI.
- Separate display detection, PPI/profile selection, and target application into clear modules.
- Support multiple output targets, not only GNOME.
- Keep dry-run and apply behavior explicit and testable.
- Make it straightforward to add new backends and target adapters.

## Design and Implementation Policy

### Core principles

- Keep the executable as a single CLI binary.
- Use a layered design:
  - `display`: detect connected displays and physical size information.
  - `profile`: normalize metrics, detect suspicious values, apply diagonal overrides, and select a profile.
  - `target`: convert resolved settings into concrete actions for each program.
  - `app`: orchestrate detection, selection, planning, dry-run output, and apply execution.
- Represent execution through a plan:
  - detect and resolve once
  - build actions
  - print the plan
  - apply the plan only when not in dry-run mode
- Prefer target adapters over target-specific branching in the main flow.
- Keep config-driven behavior in `config.toml`.

### Current architecture

```text
cmd/auto-font-by-ppi/
internal/app/
internal/cli/
internal/config/
internal/display/
internal/execx/
internal/model/
internal/profile/
internal/target/
examples/config.toml
```

### Main data flow

1. Parse CLI flags.
2. Load `config.toml`.
3. Detect displays using configured backends.
4. Normalize suspicious display metrics and apply diagonal overrides when available.
5. Select the target display.
6. Select a profile from the detected PPI.
7. Build a plan from enabled targets.
8. Print the plan.
9. Apply actions only when `dry_run = false`.

## Implemented So Far

### Go CLI scaffold

- Added `go.mod`.
- Added a minimal entry point in `cmd/auto-font-by-ppi/main.go`.
- Added internal packages for app flow, config, display backends, profile logic, execution, and target adapters.

### Configuration

- Added a new `examples/config.toml`.
- Added a minimal TOML parser implemented with the Go standard library.
- Added default configuration values in Go.
- Added support for:
  - `dry_run`
  - `preferred_display`
  - `display_backend_priority`
  - `target_names`
  - display sanity thresholds
  - diagonal overrides
  - font families
  - profiles
  - `target.gnome`
  - `target.kitty`

### Display backends

- Implemented `xrandr` backend for X11/Xorg environments.
- Implemented `gnome-wayland` backend using `gdbus` and Mutter `DisplayConfig`.
- Configured backend priority to try `gnome-wayland` before `xrandr` by default.

### Profile logic

- Implemented PPI calculation.
- Implemented suspicious metric detection.
- Implemented diagonal override application.
- Implemented display selection:
  - explicit preferred display
  - primary display
  - first display fallback
- Implemented profile selection by `max_ppi`.

### Target adapters

- Implemented GNOME target adapter:
  - `scaling_only`
  - `full_fonts`
- Implemented kitty target adapter:
  - `remote` strategy only
  - optional `--to` socket
  - configurable font size field

### Planning and execution

- Implemented `buildPlan`.
- Implemented `applyPlan`.
- Implemented dry-run printing of the selected display, selected profile, and generated actions.

## Tests Added

### Config tests

- Parse the example config file.
- Verify that config values override defaults correctly.

### Display tests

- Parse GNOME Wayland `gdbus` output.
- Deduplicate repeated display entries.
- Detect primary display from the parsed output.

### Profile tests

- Apply diagonal overrides to suspicious display metrics.
- Select preferred and primary displays correctly.
- Select the expected profile for a given PPI.

### Target tests

- Verify GNOME action generation for `full_fonts` and `scaling_only`.
- Verify GNOME rejects unsupported modes.
- Verify kitty remote command generation.
- Verify kitty rejects unsupported strategies and unsupported font size fields.
- Verify adapters pass commands through the runner on apply.

### App tests

- Verify `buildPlan` respects target order.
- Verify disabled targets are skipped.
- Verify unknown targets return errors.
- Verify adapter build errors propagate.
- Verify `applyPlan` executes actions in order.
- Verify `applyPlan` stops on the first adapter error.

### Validation status

The current Go code compiles and the current test suite passes with:

```bash
go test ./...
```

In the sandboxed environment used during development, the command was executed with temporary Go cache directories under `/tmp`.

## Current Limitations

- The custom TOML parser is intentionally minimal and only supports the subset currently used by this project.
- Wayland support currently targets GNOME Mutter only.
- Non-GNOME Wayland compositors such as Sway and Hyprland are not implemented yet.
- Kitty support currently changes the font size of running kitty instances through remote control only.
- Persistent config-file editing for kitty is not implemented.
- The existing Bash script is still the canonical end-user implementation until the Go version reaches feature parity.
- User-facing documentation for the Go CLI is not complete yet.

## Remaining Tasks

### High priority

- Reach feature parity with the current Bash script where appropriate.
- Decide whether the Go CLI should replace the existing script name or ship in parallel for a while.
- Replace the minimal TOML parser with a proper TOML library if long-term config growth is expected.
- Add documentation for the Go CLI usage and migration path.

### Display/backend work

- Add support for additional Wayland compositors if needed:
  - Sway via `swaymsg`
  - Hyprland via `hyprctl`
- Add backend-specific test fixtures under `testdata/`.
- Consider more robust parsing for display metadata if Mutter output shape changes.

### Target work

- Add a generic command target for arbitrary CLI-configurable applications.
- Decide how to support persistent font settings for programs such as kitty.
- Add more adapters if needed:
  - terminal emulators
  - editors
  - compositor-specific font settings

### App and CLI work

- Expand CLI flags beyond the current minimal set.
- Add verbose diagnostics around backend fallback decisions.
- Add integration-style tests for the full runner flow.

### Documentation and packaging

- Update `README.md` and `README_ja_JP.md` to describe the Go implementation.
- Add migration notes from the Bash script.
- Add build/install instructions for the Go binary.
- Consider systemd user service examples for automatic execution.

## Recommended Next Steps

1. Add a proper generic command target adapter for arbitrary programs.
2. Decide whether to keep the custom TOML parser or adopt a TOML library now.
3. Add `testdata/` fixtures for `xrandr` and GNOME Wayland parsing.
4. Expand the CLI and documentation so the Go tool can be evaluated by users directly.
5. Only retire the Bash script after the Go implementation is functionally complete enough for daily use.
