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
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
		DryRun:                 true,
		PreferredDisplay:       "",
		DisplayBackendPriority: []string{"gnome-wayland", "xrandr"},
		TargetNames:            []string{"gnome", "kitty"},
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
			{Name: "ppi-110", MaxPPI: 110, TextScaling: 1.00, UIFontSize: 11, DocumentFontSize: 11, MonospaceFontSize: 10, TitlebarFontSize: 11},
			{Name: "ppi-140", MaxPPI: 140, TextScaling: 1.00, UIFontSize: 12, DocumentFontSize: 12, MonospaceFontSize: 11, TitlebarFontSize: 12},
			{Name: "ppi-180", MaxPPI: 180, TextScaling: 1.00, UIFontSize: 13, DocumentFontSize: 13, MonospaceFontSize: 12, TitlebarFontSize: 13},
			{Name: "ppi-240", MaxPPI: 240, TextScaling: 1.00, UIFontSize: 14, DocumentFontSize: 14, MonospaceFontSize: 13, TitlebarFontSize: 14},
			{Name: "ppi-9999", MaxPPI: 9999, TextScaling: 1.60, UIFontSize: 16, DocumentFontSize: 16, MonospaceFontSize: 14, TitlebarFontSize: 16},
		},
		Targets: model.TargetConfigs{
			GNOME: model.GNOMETargetConfig{
				Enabled: true,
				Mode:    "full_fonts",
			},
			Kitty: model.KittyTargetConfig{
				Enabled:       false,
				Strategy:      "remote",
				Socket:        "",
				All:           true,
				FontSizeField: "monospace_font_size",
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
			return cfg, nil
		}
		return model.Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	if err := parse(string(data), &cfg); err != nil {
		return model.Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	return cfg, validate(cfg)
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
	scanner := bufio.NewScanner(strings.NewReader(input))
	section := ""
	var currentProfile *model.Profile
	seenProfiles := false

	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(stripComment(scanner.Text()))
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[[") && strings.HasSuffix(line, "]]") {
			arraySection := strings.TrimSpace(line[2 : len(line)-2])
			if arraySection != "profiles" {
				return fmt.Errorf("line %d: unsupported array section %q", lineNumber, arraySection)
			}

			if !seenProfiles {
				cfg.Profiles = nil
				seenProfiles = true
			}

			cfg.Profiles = append(cfg.Profiles, model.Profile{})
			currentProfile = &cfg.Profiles[len(cfg.Profiles)-1]
			section = "profiles"
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			currentProfile = nil
			continue
		}

		key, value, err := splitAssignment(line)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNumber, err)
		}

		if err := assign(section, key, value, cfg, currentProfile); err != nil {
			return fmt.Errorf("line %d: %w", lineNumber, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func assign(section, key, value string, cfg *model.Config, currentProfile *model.Profile) error {
	switch section {
	case "":
		return assignRoot(key, value, cfg)
	case "display":
		return assignDisplay(key, value, cfg)
	case "display.diagonal_overrides":
		number, err := parseFloat(value)
		if err != nil {
			return err
		}
		cfg.Display.DiagonalOverrides[trimQuoted(key)] = number
		return nil
	case "fonts":
		return assignFonts(key, value, cfg)
	case "profiles":
		if currentProfile == nil {
			return fmt.Errorf("profile value specified before [[profiles]]")
		}
		return assignProfile(key, value, currentProfile)
	case "target.gnome":
		return assignGNOME(key, value, cfg)
	case "target.kitty":
		return assignKitty(key, value, cfg)
	default:
		return fmt.Errorf("unsupported section %q", section)
	}
}

func assignRoot(key, value string, cfg *model.Config) error {
	switch key {
	case "dry_run":
		parsed, err := parseBool(value)
		if err != nil {
			return err
		}
		cfg.DryRun = parsed
	case "preferred_display":
		parsed, err := parseString(value)
		if err != nil {
			return err
		}
		cfg.PreferredDisplay = parsed
	case "display_backend_priority":
		parsed, err := parseStringArray(value)
		if err != nil {
			return err
		}
		cfg.DisplayBackendPriority = parsed
	case "target_names":
		parsed, err := parseStringArray(value)
		if err != nil {
			return err
		}
		cfg.TargetNames = parsed
	default:
		return fmt.Errorf("unsupported root key %q", key)
	}
	return nil
}

func assignDisplay(key, value string, cfg *model.Config) error {
	switch key {
	case "min_reasonable_ppi":
		parsed, err := parseFloat(value)
		if err != nil {
			return err
		}
		cfg.Display.MinReasonablePPI = parsed
	case "max_reasonable_ppi":
		parsed, err := parseFloat(value)
		if err != nil {
			return err
		}
		cfg.Display.MaxReasonablePPI = parsed
	default:
		return fmt.Errorf("unsupported display key %q", key)
	}
	return nil
}

func assignFonts(key, value string, cfg *model.Config) error {
	parsed, err := parseString(value)
	if err != nil {
		return err
	}

	switch key {
	case "ui_family":
		cfg.Fonts.UI = parsed
	case "document_family":
		cfg.Fonts.Document = parsed
	case "monospace_family":
		cfg.Fonts.Monospace = parsed
	case "titlebar_family":
		cfg.Fonts.Titlebar = parsed
	default:
		return fmt.Errorf("unsupported fonts key %q", key)
	}
	return nil
}

func assignProfile(key, value string, profile *model.Profile) error {
	switch key {
	case "name":
		parsed, err := parseString(value)
		if err != nil {
			return err
		}
		profile.Name = parsed
	case "max_ppi":
		parsed, err := parseFloat(value)
		if err != nil {
			return err
		}
		profile.MaxPPI = parsed
	case "text_scaling":
		parsed, err := parseFloat(value)
		if err != nil {
			return err
		}
		profile.TextScaling = parsed
	case "ui_font_size":
		parsed, err := parseInt(value)
		if err != nil {
			return err
		}
		profile.UIFontSize = parsed
	case "document_font_size":
		parsed, err := parseInt(value)
		if err != nil {
			return err
		}
		profile.DocumentFontSize = parsed
	case "monospace_font_size":
		parsed, err := parseInt(value)
		if err != nil {
			return err
		}
		profile.MonospaceFontSize = parsed
	case "titlebar_font_size":
		parsed, err := parseInt(value)
		if err != nil {
			return err
		}
		profile.TitlebarFontSize = parsed
	default:
		return fmt.Errorf("unsupported profile key %q", key)
	}
	return nil
}

func assignGNOME(key, value string, cfg *model.Config) error {
	switch key {
	case "enabled":
		parsed, err := parseBool(value)
		if err != nil {
			return err
		}
		cfg.Targets.GNOME.Enabled = parsed
	case "mode":
		parsed, err := parseString(value)
		if err != nil {
			return err
		}
		cfg.Targets.GNOME.Mode = parsed
	default:
		return fmt.Errorf("unsupported target.gnome key %q", key)
	}
	return nil
}

func assignKitty(key, value string, cfg *model.Config) error {
	switch key {
	case "enabled":
		parsed, err := parseBool(value)
		if err != nil {
			return err
		}
		cfg.Targets.Kitty.Enabled = parsed
	case "strategy":
		parsed, err := parseString(value)
		if err != nil {
			return err
		}
		cfg.Targets.Kitty.Strategy = parsed
	case "socket":
		parsed, err := parseString(value)
		if err != nil {
			return err
		}
		cfg.Targets.Kitty.Socket = parsed
	case "all":
		parsed, err := parseBool(value)
		if err != nil {
			return err
		}
		cfg.Targets.Kitty.All = parsed
	case "font_size_field":
		parsed, err := parseString(value)
		if err != nil {
			return err
		}
		cfg.Targets.Kitty.FontSizeField = parsed
	default:
		return fmt.Errorf("unsupported target.kitty key %q", key)
	}
	return nil
}

func stripComment(line string) string {
	var builder strings.Builder
	inString := false
	escaped := false

	for _, r := range line {
		switch {
		case escaped:
			builder.WriteRune(r)
			escaped = false
		case r == '\\' && inString:
			builder.WriteRune(r)
			escaped = true
		case r == '"':
			builder.WriteRune(r)
			inString = !inString
		case r == '#' && !inString:
			return builder.String()
		default:
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func splitAssignment(line string) (string, string, error) {
	inString := false
	for index, r := range line {
		switch r {
		case '"':
			inString = !inString
		case '=':
			if inString {
				continue
			}
			key := strings.TrimSpace(line[:index])
			value := strings.TrimSpace(line[index+1:])
			if key == "" || value == "" {
				return "", "", fmt.Errorf("invalid assignment")
			}
			return trimQuoted(key), value, nil
		}
	}

	return "", "", fmt.Errorf("missing key/value separator")
}

func parseString(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return "", fmt.Errorf("expected string literal, got %q", value)
	}

	parsed, err := strconv.Unquote(value)
	if err != nil {
		return "", err
	}
	return parsed, nil
}

func parseBool(value string) (bool, error) {
	return strconv.ParseBool(strings.TrimSpace(value))
}

func parseInt(value string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(value))
}

func parseFloat(value string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(value), 64)
}

func parseStringArray(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if len(value) < 2 || value[0] != '[' || value[len(value)-1] != ']' {
		return nil, fmt.Errorf("expected string array, got %q", value)
	}

	content := strings.TrimSpace(value[1 : len(value)-1])
	if content == "" {
		return []string{}, nil
	}

	items := splitArrayItems(content)
	values := make([]string, 0, len(items))
	for _, item := range items {
		parsed, err := parseString(strings.TrimSpace(item))
		if err != nil {
			return nil, err
		}
		values = append(values, parsed)
	}
	return values, nil
}

func splitArrayItems(content string) []string {
	var items []string
	var builder strings.Builder
	inString := false
	escaped := false

	for _, r := range content {
		switch {
		case escaped:
			builder.WriteRune(r)
			escaped = false
		case r == '\\' && inString:
			builder.WriteRune(r)
			escaped = true
		case r == '"':
			builder.WriteRune(r)
			inString = !inString
		case r == ',' && !inString:
			items = append(items, builder.String())
			builder.Reset()
		default:
			builder.WriteRune(r)
		}
	}

	items = append(items, builder.String())
	return items
}

func trimQuoted(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, `"`)
	value = strings.TrimSuffix(value, `"`)
	return value
}
