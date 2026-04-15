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

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
)

type recordingRunner struct {
	lastCommand []string
}

func (r *recordingRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	r.lastCommand = append([]string{name}, args...)
	return nil, nil
}

func testConfig() model.Config {
	return model.Config{
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

func testResolvedSettings() model.ResolvedSettings {
	return model.ResolvedSettings{
		Fonts: model.FontFamilies{
			UI:        "Cantarell",
			Document:  "Cantarell",
			Monospace: "Monospace",
			Titlebar:  "Cantarell Bold",
		},
		Profile: model.Profile{
			TextScaling:       1.25,
			UIFontSize:        14,
			DocumentFontSize:  14,
			MonospaceFontSize: 13,
			TitlebarFontSize:  15,
			KittyFontSize:     12,
		},
	}
}

func testKittyAction() model.Action {
	return model.Action{
		Target:      "kitty",
		Description: "Set kitty font size to 13",
		Command:     []string{"kitten", "@", "set-font-size", "--all", "13"},
	}
}
