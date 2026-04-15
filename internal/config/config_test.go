/*
Copyright 2026 Yabe Kazuhiro

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseExampleConfig(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "examples", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example config: %v", err)
	}

	cfg := DefaultConfig()
	if err := parse(string(data), &cfg); err != nil {
		t.Fatalf("parse example config: %v", err)
	}

	if err := validate(cfg); err != nil {
		t.Fatalf("validate example config: %v", err)
	}

	if got, want := cfg.DisplayBackendPriority, []string{"gnome-wayland", "xrandr"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("unexpected backend priority: got %v want %v", got, want)
	}
	if got, want := cfg.TargetNames, []string{"gnome"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("unexpected target names: got %v want %v", got, want)
	}
	if got := cfg.Display.DiagonalOverrides["HDMI-1"]; got != 31.5 {
		t.Fatalf("unexpected HDMI-1 diagonal override: got %v want 31.5", got)
	}
	if got := cfg.Targets.Kitty.Strategy; got != "remote" {
		t.Fatalf("unexpected kitty strategy: got %q want %q", got, "remote")
	}
	if got := cfg.Targets.Kitty.FontSizeField; got != "kitty_font_size" {
		t.Fatalf("unexpected kitty font size field: got %q want %q", got, "kitty_font_size")
	}
	if got := len(cfg.Profiles); got != 5 {
		t.Fatalf("unexpected profile count: got %d want 5", got)
	}
	if got := cfg.Profiles[1].KittyFontSize; got != 11 {
		t.Fatalf("unexpected kitty font size: got %d want 11", got)
	}
}

func TestParseOverridesDefaults(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	input := `
dry_run = false
preferred_display = "DP-1"
display_backend_priority = ["xrandr", "gnome-wayland"]
target_names = ["kitty"]

[display]
min_reasonable_ppi = 70
max_reasonable_ppi = 300

[display.diagonal_overrides]
"DP-1" = 27.0

[fonts]
ui_family = "Noto Sans"
document_family = "Noto Sans"
monospace_family = "JetBrains Mono"
titlebar_family = "Noto Sans Bold"

[[profiles]]
name = "custom"
max_ppi = 200
text_scaling = 1.10
ui_font_size = 15
document_font_size = 16
monospace_font_size = 14
titlebar_font_size = 17
kitty_font_size = 18

[target.gnome]
mode = "scaling_only"

[target.kitty]
strategy = "remote"
socket = "unix:/tmp/kitty.sock"
all = false
font_size_field = "ui_font_size"
`

	if err := parse(input, &cfg); err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if got := cfg.DryRun; got {
		t.Fatalf("unexpected dry_run: got %v want false", got)
	}
	if got := cfg.PreferredDisplay; got != "DP-1" {
		t.Fatalf("unexpected preferred display: got %q want %q", got, "DP-1")
	}
	if got := cfg.Display.MinReasonablePPI; got != 70 {
		t.Fatalf("unexpected min reasonable ppi: got %v want 70", got)
	}
	if got := cfg.Fonts.Monospace; got != "JetBrains Mono" {
		t.Fatalf("unexpected monospace family: got %q want %q", got, "JetBrains Mono")
	}
	if got := cfg.Display.DiagonalOverrides["DP-1"]; got != 27.0 {
		t.Fatalf("unexpected DP-1 diagonal override: got %v want 27.0", got)
	}
	if got := len(cfg.Profiles); got != 1 {
		t.Fatalf("unexpected profile count: got %d want 1", got)
	}
	if got := cfg.Profiles[0].Name; got != "custom" {
		t.Fatalf("unexpected profile name: got %q want %q", got, "custom")
	}
	if got := cfg.Targets.Kitty.Socket; got != "unix:/tmp/kitty.sock" {
		t.Fatalf("unexpected kitty socket: got %q want %q", got, "unix:/tmp/kitty.sock")
	}
	if got := cfg.Targets.Kitty.All; got {
		t.Fatalf("unexpected kitty all flag: got %v want false", got)
	}
	if got := cfg.Profiles[0].KittyFontSize; got != 18 {
		t.Fatalf("unexpected kitty font size: got %d want 18", got)
	}
}

func TestParseRejectsUnknownRootKey(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	err := parse("unknown_key = true\n", &cfg)
	if err == nil {
		t.Fatal("expected parse to fail for an unknown root key")
	}
	if !strings.Contains(err.Error(), `unsupported config key "unknown_key"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseRejectsUnknownNestedKey(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	input := `
[target.kitty]
strategy = "remote"
unknown_key = "value"
`

	err := parse(input, &cfg)
	if err == nil {
		t.Fatal("expected parse to fail for an unknown nested key")
	}
	if !strings.Contains(err.Error(), `unsupported config key "target.kitty.unknown_key"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseRejectsLegacyEnabledKeys(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	input := `
[target.gnome]
enabled = false
mode = "scaling_only"

[target.kitty]
enabled = true
strategy = "remote"
`

	err := parse(input, &cfg)
	if err == nil {
		t.Fatal("expected parse to fail for legacy enabled keys")
	}
	if !strings.Contains(err.Error(), `unsupported config key "target.gnome.enabled"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadCreatesSampleConfigWhenMissing(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "auto-font-by-ppi", "config.toml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	want := DefaultConfig()
	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("unexpected loaded config after sample generation: got %+v want %+v", cfg, want)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	if !strings.Contains(string(data), "Generated default configuration for auto-font-by-ppi.") {
		t.Fatalf("generated config did not contain the expected header")
	}
	if !strings.Contains(string(data), "# Add \"kitty\" to target_names to enable kitty updates.") {
		t.Fatalf("generated config did not contain the expected kitty opt-in hint")
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load on generated config returned error: %v", err)
	}
	if !reflect.DeepEqual(reloaded, want) {
		t.Fatalf("unexpected config after reloading generated sample: got %+v want %+v", reloaded, want)
	}
}
