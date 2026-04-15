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

package target

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/execx"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
)

type GNOMEAdapter struct{}

func (GNOMEAdapter) Name() string {
	return "gnome"
}

func (GNOMEAdapter) BuildActions(cfg model.Config, resolved model.ResolvedSettings) ([]model.Action, error) {
	gnomeConfig := cfg.Targets.GNOME
	scaling := strconv.FormatFloat(resolved.Profile.TextScaling, 'f', 2, 64)

	actions := []model.Action{
		{
			Target:      "gnome",
			Description: fmt.Sprintf("Set GNOME text scaling factor to %s", scaling),
			Command: []string{
				"gsettings", "set", "org.gnome.desktop.interface", "text-scaling-factor", scaling,
			},
		},
	}

	switch gnomeConfig.Mode {
	case "scaling_only":
		return actions, nil
	case "full_fonts":
		actions = append(actions,
			model.Action{
				Target:      "gnome",
				Description: fmt.Sprintf("Set GNOME interface font to %s %d", resolved.Fonts.UI, resolved.Profile.UIFontSize),
				Command: []string{
					"gsettings", "set", "org.gnome.desktop.interface", "font-name",
					fmt.Sprintf("%s %d", resolved.Fonts.UI, resolved.Profile.UIFontSize),
				},
			},
			model.Action{
				Target:      "gnome",
				Description: fmt.Sprintf("Set GNOME document font to %s %d", resolved.Fonts.Document, resolved.Profile.DocumentFontSize),
				Command: []string{
					"gsettings", "set", "org.gnome.desktop.interface", "document-font-name",
					fmt.Sprintf("%s %d", resolved.Fonts.Document, resolved.Profile.DocumentFontSize),
				},
			},
			model.Action{
				Target:      "gnome",
				Description: fmt.Sprintf("Set GNOME monospace font to %s %d", resolved.Fonts.Monospace, resolved.Profile.MonospaceFontSize),
				Command: []string{
					"gsettings", "set", "org.gnome.desktop.interface", "monospace-font-name",
					fmt.Sprintf("%s %d", resolved.Fonts.Monospace, resolved.Profile.MonospaceFontSize),
				},
			},
			model.Action{
				Target:      "gnome",
				Description: fmt.Sprintf("Set GNOME titlebar font to %s %d", resolved.Fonts.Titlebar, resolved.Profile.TitlebarFontSize),
				Command: []string{
					"gsettings", "set", "org.gnome.desktop.wm.preferences", "titlebar-font",
					fmt.Sprintf("%s %d", resolved.Fonts.Titlebar, resolved.Profile.TitlebarFontSize),
				},
			},
		)
		return actions, nil
	default:
		return nil, fmt.Errorf("unsupported GNOME mode %q", gnomeConfig.Mode)
	}
}

func (GNOMEAdapter) Apply(ctx context.Context, action model.Action, runner execx.Runner) error {
	return applyCommandAction(ctx, action, runner)
}
