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

type KittyAdapter struct{}

func (KittyAdapter) Name() string {
	return "kitty"
}

func (KittyAdapter) BuildActions(cfg model.Config, resolved model.ResolvedSettings) ([]model.Action, error) {
	kittyConfig := cfg.Targets.Kitty
	if kittyConfig.Strategy != "remote" {
		return nil, fmt.Errorf("unsupported kitty strategy %q", kittyConfig.Strategy)
	}

	size, err := kittyFontSize(kittyConfig.FontSizeField, resolved)
	if err != nil {
		return nil, err
	}

	command := []string{"kitten", "@"}
	if kittyConfig.Socket != "" {
		command = append(command, "--to", kittyConfig.Socket)
	}
	command = append(command, "set-font-size")
	if kittyConfig.All {
		command = append(command, "--all")
	}
	command = append(command, strconv.Itoa(size))

	return []model.Action{
		{
			Target:      "kitty",
			Description: fmt.Sprintf("Set kitty font size to %d", size),
			Command:     command,
		},
	}, nil
}

func (KittyAdapter) Apply(ctx context.Context, action model.Action, runner execx.Runner) error {
	return applyCommandAction(ctx, action, runner)
}

func kittyFontSize(field string, resolved model.ResolvedSettings) (int, error) {
	switch field {
	case "":
		return kittyFontSize("kitty_font_size", resolved)
	case "kitty_font_size":
		if resolved.Profile.KittyFontSize > 0 {
			return resolved.Profile.KittyFontSize, nil
		}
		return resolved.Profile.MonospaceFontSize, nil
	case "monospace_font_size":
		return resolved.Profile.MonospaceFontSize, nil
	case "ui_font_size":
		return resolved.Profile.UIFontSize, nil
	case "document_font_size":
		return resolved.Profile.DocumentFontSize, nil
	default:
		return 0, fmt.Errorf("unsupported kitty font_size_field %q", field)
	}
}
