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

package model

type Config struct {
	DryRun                 bool
	PreferredDisplay       string
	DisplayBackendPriority []string
	TargetNames            []string
	CurrentDisplayOverride *DisplayOverride

	Display  DisplayConfig
	Fonts    FontFamilies
	Profiles []Profile
	Targets  TargetConfigs
}

type DisplayConfig struct {
	MinReasonablePPI  float64
	MaxReasonablePPI  float64
	DiagonalOverrides map[string]float64
}

type DisplayOverride struct {
	Mode  string
	Value float64
}

type FontFamilies struct {
	UI        string
	Document  string
	Monospace string
	Titlebar  string
}

type Profile struct {
	Name              string
	MaxPPI            float64
	TextScaling       float64
	UIFontSize        int
	DocumentFontSize  int
	MonospaceFontSize int
	TitlebarFontSize  int
	KittyFontSize     float64
}

type TargetConfigs struct {
	GNOME GNOMETargetConfig
	Kitty KittyTargetConfig
}

type GNOMETargetConfig struct {
	Mode string
}

type KittyTargetConfig struct {
	Strategy      string
	Socket        string
	All           bool
	FontSizeField string
}

type DisplayInfo struct {
	Name                   string
	IsPrimary              bool
	WidthPx                int
	HeightPx               int
	WidthMM                int
	HeightMM               int
	PPI                    float64
	PhysicalSizeSource     string
	OverrideDiagonalInches float64
	AssumedPPI             float64
}

type ResolvedSettings struct {
	Display             DisplayInfo
	Profile             Profile
	Fonts               FontFamilies
	SelectedTargetNames []string
	SourceBackend       string
}

type Action struct {
	Target      string
	Description string
	Command     []string
}

type Plan struct {
	DryRun   bool
	Resolved ResolvedSettings
	Actions  []Action
}
