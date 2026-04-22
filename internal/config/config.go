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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
)

func DefaultConfigPath() string {
	if base := os.Getenv("XDG_CONFIG_HOME"); base != "" {
		return filepath.Join(base, "auto-font-by-ppi", "config.toml")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "config.toml"
	}

	return filepath.Join(home, ".config", "auto-font-by-ppi", "config.toml")
}

func DefaultConfig() model.Config {
	return model.Config{
		DryRun:                 false,
		PreferredDisplay:       "",
		DisplayBackendPriority: []string{"gnome-wayland", "xrandr"},
		TargetNames:            []string{"gnome"},
		Display: model.DisplayConfig{
			MinReasonablePPI:  50,
			MaxReasonablePPI:  400,
			DiagonalOverrides: map[string]float64{},
		},
		Fonts: model.FontFamilies{
			UI:        "Cantarell",
			Document:  "Cantarell",
			Monospace: "Monospace",
			Titlebar:  "Cantarell Bold",
		},
		Profiles: []model.Profile{
			{Name: "ppi-110", MaxPPI: 110, TextScaling: 1.00, UIFontSize: 11, DocumentFontSize: 11, MonospaceFontSize: 10, TitlebarFontSize: 11, KittyFontSize: 10},
			{Name: "ppi-140", MaxPPI: 140, TextScaling: 1.00, UIFontSize: 12, DocumentFontSize: 12, MonospaceFontSize: 11, TitlebarFontSize: 12, KittyFontSize: 11},
			{Name: "ppi-180", MaxPPI: 180, TextScaling: 1.00, UIFontSize: 13, DocumentFontSize: 13, MonospaceFontSize: 12, TitlebarFontSize: 13, KittyFontSize: 12},
			{Name: "ppi-240", MaxPPI: 240, TextScaling: 1.00, UIFontSize: 14, DocumentFontSize: 14, MonospaceFontSize: 13, TitlebarFontSize: 14, KittyFontSize: 13},
			{Name: "ppi-9999", MaxPPI: 9999, TextScaling: 1.60, UIFontSize: 16, DocumentFontSize: 16, MonospaceFontSize: 14, TitlebarFontSize: 16, KittyFontSize: 14},
		},
		Targets: model.TargetConfigs{
			GNOME: model.GNOMETargetConfig{
				Mode: "full_fonts",
			},
			Kitty: model.KittyTargetConfig{
				Strategy:      "remote",
				Socket:        "",
				All:           true,
				FontSizeField: "kitty_font_size",
			},
		},
	}
}

func Load(path string) (model.Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := ensureConfigFile(path); err != nil {
				return model.Config{}, err
			}
			data, err = os.ReadFile(path)
			if err != nil {
				return model.Config{}, fmt.Errorf("read config %q: %w", path, err)
			}
		} else {
			return model.Config{}, fmt.Errorf("read config %q: %w", path, err)
		}
	}

	if err := parse(string(data), &cfg); err != nil {
		return model.Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	return cfg, validate(cfg)
}

func ensureConfigFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config directory %q: %w", dir, err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return fmt.Errorf("create config %q: %w", path, err)
	}
	defer file.Close()

	if _, err := file.WriteString(sampleConfigText()); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}

	return nil
}

func sampleConfigText() string {
	cfg := DefaultConfig()

	var builder strings.Builder
	builder.WriteString("# Generated default configuration for auto-font-by-ppi.\n")
	builder.WriteString("# Edit this file as needed.\n\n")

	fmt.Fprintf(&builder, "dry_run = %t\n", cfg.DryRun)
	fmt.Fprintf(&builder, "preferred_display = %q\n", cfg.PreferredDisplay)
	fmt.Fprintf(&builder, "display_backend_priority = %s\n", formatStringArray(cfg.DisplayBackendPriority))
	fmt.Fprintf(&builder, "target_names = %s\n\n", formatStringArray(cfg.TargetNames))

	builder.WriteString("[display]\n")
	fmt.Fprintf(&builder, "min_reasonable_ppi = %s\n", formatFloat(cfg.Display.MinReasonablePPI))
	fmt.Fprintf(&builder, "max_reasonable_ppi = %s\n\n", formatFloat(cfg.Display.MaxReasonablePPI))

	builder.WriteString("# Example diagonal overrides:\n")
	builder.WriteString("# [display.diagonal_overrides]\n")
	builder.WriteString("# \"HDMI-1\" = 31.5\n")
	builder.WriteString("# \"eDP-1\" = 14.0\n\n")

	builder.WriteString("# Add \"kitty\" to target_names to enable kitty updates.\n")
	builder.WriteString("# Example: target_names = [\"gnome\", \"kitty\"]\n\n")

	builder.WriteString("[fonts]\n")
	fmt.Fprintf(&builder, "ui_family = %q\n", cfg.Fonts.UI)
	fmt.Fprintf(&builder, "document_family = %q\n", cfg.Fonts.Document)
	fmt.Fprintf(&builder, "monospace_family = %q\n", cfg.Fonts.Monospace)
	fmt.Fprintf(&builder, "titlebar_family = %q\n\n", cfg.Fonts.Titlebar)

	builder.WriteString("# kitty_font_size accepts integers or decimals such as 12.5.\n\n")

	for _, profile := range cfg.Profiles {
		builder.WriteString("[[profiles]]\n")
		fmt.Fprintf(&builder, "name = %q\n", profile.Name)
		fmt.Fprintf(&builder, "max_ppi = %s\n", formatFloat(profile.MaxPPI))
		fmt.Fprintf(&builder, "text_scaling = %.2f\n", profile.TextScaling)
		fmt.Fprintf(&builder, "ui_font_size = %d\n", profile.UIFontSize)
		fmt.Fprintf(&builder, "document_font_size = %d\n", profile.DocumentFontSize)
		fmt.Fprintf(&builder, "monospace_font_size = %d\n", profile.MonospaceFontSize)
		fmt.Fprintf(&builder, "titlebar_font_size = %d\n", profile.TitlebarFontSize)
		fmt.Fprintf(&builder, "kitty_font_size = %s\n\n", formatFloat(profile.KittyFontSize))
	}

	builder.WriteString("[target.gnome]\n")
	fmt.Fprintf(&builder, "mode = %q\n\n", cfg.Targets.GNOME.Mode)

	builder.WriteString("[target.kitty]\n")
	fmt.Fprintf(&builder, "strategy = %q\n", cfg.Targets.Kitty.Strategy)
	fmt.Fprintf(&builder, "socket = %q\n", cfg.Targets.Kitty.Socket)
	fmt.Fprintf(&builder, "all = %t\n", cfg.Targets.Kitty.All)
	fmt.Fprintf(&builder, "font_size_field = %q\n", cfg.Targets.Kitty.FontSizeField)

	return builder.String()
}

func formatStringArray(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, strconv.Quote(value))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func validate(cfg model.Config) error {
	if len(cfg.DisplayBackendPriority) == 0 {
		return fmt.Errorf("display_backend_priority must not be empty")
	}
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("profiles must not be empty")
	}
	if cfg.Display.MinReasonablePPI <= 0 {
		return fmt.Errorf("display.min_reasonable_ppi must be positive")
	}
	if cfg.Display.MaxReasonablePPI < cfg.Display.MinReasonablePPI {
		return fmt.Errorf("display.max_reasonable_ppi must be greater than or equal to display.min_reasonable_ppi")
	}
	return nil
}

func parse(input string, cfg *model.Config) error {
	var decoded fileConfig

	meta, err := toml.Decode(input, &decoded)
	if err != nil {
		return err
	}

	if err := rejectUndecoded(meta); err != nil {
		return err
	}

	mergeConfig(cfg, decoded)
	return nil
}

type fileConfig struct {
	DryRun                 *bool               `toml:"dry_run"`
	PreferredDisplay       *string             `toml:"preferred_display"`
	DisplayBackendPriority *[]string           `toml:"display_backend_priority"`
	TargetNames            *[]string           `toml:"target_names"`
	Display                *displayFileConfig  `toml:"display"`
	Fonts                  *fontsFileConfig    `toml:"fonts"`
	Profiles               []profileFileConfig `toml:"profiles"`
	Target                 *targetFileConfig   `toml:"target"`
}

type displayFileConfig struct {
	MinReasonablePPI  *float64           `toml:"min_reasonable_ppi"`
	MaxReasonablePPI  *float64           `toml:"max_reasonable_ppi"`
	DiagonalOverrides map[string]float64 `toml:"diagonal_overrides"`
}

type fontsFileConfig struct {
	UI        *string `toml:"ui_family"`
	Document  *string `toml:"document_family"`
	Monospace *string `toml:"monospace_family"`
	Titlebar  *string `toml:"titlebar_family"`
}

type profileFileConfig struct {
	Name              string  `toml:"name"`
	MaxPPI            float64 `toml:"max_ppi"`
	TextScaling       float64 `toml:"text_scaling"`
	UIFontSize        int     `toml:"ui_font_size"`
	DocumentFontSize  int     `toml:"document_font_size"`
	MonospaceFontSize int     `toml:"monospace_font_size"`
	TitlebarFontSize  int     `toml:"titlebar_font_size"`
	KittyFontSize     float64 `toml:"kitty_font_size"`
}

type targetFileConfig struct {
	GNOME *gnomeTargetFileConfig `toml:"gnome"`
	Kitty *kittyTargetFileConfig `toml:"kitty"`
}

type gnomeTargetFileConfig struct {
	Mode *string `toml:"mode"`
}

type kittyTargetFileConfig struct {
	Strategy      *string `toml:"strategy"`
	Socket        *string `toml:"socket"`
	All           *bool   `toml:"all"`
	FontSizeField *string `toml:"font_size_field"`
}

func rejectUndecoded(meta toml.MetaData) error {
	undecoded := meta.Undecoded()
	if len(undecoded) == 0 {
		return nil
	}

	return fmt.Errorf("unsupported config key %q", undecoded[0].String())
}

func mergeConfig(cfg *model.Config, decoded fileConfig) {
	if decoded.DryRun != nil {
		cfg.DryRun = *decoded.DryRun
	}
	if decoded.PreferredDisplay != nil {
		cfg.PreferredDisplay = *decoded.PreferredDisplay
	}
	if decoded.DisplayBackendPriority != nil {
		cfg.DisplayBackendPriority = append([]string(nil), (*decoded.DisplayBackendPriority)...)
	}
	if decoded.TargetNames != nil {
		cfg.TargetNames = append([]string(nil), (*decoded.TargetNames)...)
	}

	if decoded.Display != nil {
		mergeDisplayConfig(&cfg.Display, *decoded.Display)
	}
	if decoded.Fonts != nil {
		mergeFontsConfig(&cfg.Fonts, *decoded.Fonts)
	}
	if decoded.Profiles != nil {
		cfg.Profiles = make([]model.Profile, 0, len(decoded.Profiles))
		for _, profile := range decoded.Profiles {
			cfg.Profiles = append(cfg.Profiles, model.Profile{
				Name:              profile.Name,
				MaxPPI:            profile.MaxPPI,
				TextScaling:       profile.TextScaling,
				UIFontSize:        profile.UIFontSize,
				DocumentFontSize:  profile.DocumentFontSize,
				MonospaceFontSize: profile.MonospaceFontSize,
				TitlebarFontSize:  profile.TitlebarFontSize,
				KittyFontSize:     profile.KittyFontSize,
			})
		}
	}

	if decoded.Target != nil {
		mergeTargetConfig(&cfg.Targets, *decoded.Target)
	}
}

func mergeDisplayConfig(cfg *model.DisplayConfig, decoded displayFileConfig) {
	if decoded.MinReasonablePPI != nil {
		cfg.MinReasonablePPI = *decoded.MinReasonablePPI
	}
	if decoded.MaxReasonablePPI != nil {
		cfg.MaxReasonablePPI = *decoded.MaxReasonablePPI
	}
	if decoded.DiagonalOverrides != nil {
		cfg.DiagonalOverrides = make(map[string]float64, len(decoded.DiagonalOverrides))
		for name, diagonal := range decoded.DiagonalOverrides {
			cfg.DiagonalOverrides[name] = diagonal
		}
	}
}

func mergeFontsConfig(cfg *model.FontFamilies, decoded fontsFileConfig) {
	if decoded.UI != nil {
		cfg.UI = *decoded.UI
	}
	if decoded.Document != nil {
		cfg.Document = *decoded.Document
	}
	if decoded.Monospace != nil {
		cfg.Monospace = *decoded.Monospace
	}
	if decoded.Titlebar != nil {
		cfg.Titlebar = *decoded.Titlebar
	}
}

func mergeTargetConfig(cfg *model.TargetConfigs, decoded targetFileConfig) {
	if decoded.GNOME != nil {
		if decoded.GNOME.Mode != nil {
			cfg.GNOME.Mode = *decoded.GNOME.Mode
		}
	}

	if decoded.Kitty != nil {
		if decoded.Kitty.Strategy != nil {
			cfg.Kitty.Strategy = *decoded.Kitty.Strategy
		}
		if decoded.Kitty.Socket != nil {
			cfg.Kitty.Socket = *decoded.Kitty.Socket
		}
		if decoded.Kitty.All != nil {
			cfg.Kitty.All = *decoded.Kitty.All
		}
		if decoded.Kitty.FontSizeField != nil {
			cfg.Kitty.FontSizeField = *decoded.Kitty.FontSizeField
		}
	}
}
