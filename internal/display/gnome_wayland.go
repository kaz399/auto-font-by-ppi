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

package display

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/execx"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/profile"
)

var (
	gnomeWaylandConnectorPattern   = regexp.MustCompile(`'((?:HDMI|DP|eDP|DVI|VGA|Virtual|DisplayPort)[^']*)'`)
	gnomeWaylandModePattern        = regexp.MustCompile(`[,(\[]\s*(\d{3,5})\s*,\s*(\d{3,5})\s*,\s*(?:[0-9]+(?:\.[0-9]+)?)`)
	gnomeWaylandMillimeterPattern  = regexp.MustCompile(`[,(\[]\s*(\d{2,5})\s*,\s*(\d{2,5})\s*[)\]]`)
	gnomeWaylandPrimaryHintPattern = regexp.MustCompile(`(?i)primary`)
)

type GNOMEWaylandBackend struct {
	Runner execx.Runner
}

func (GNOMEWaylandBackend) Name() string {
	return "gnome-wayland"
}

func (b GNOMEWaylandBackend) Detect(ctx context.Context) ([]model.DisplayInfo, error) {
	output, err := b.Runner.Run(
		ctx,
		"gdbus",
		"call",
		"--session",
		"--dest", "org.gnome.Mutter.DisplayConfig",
		"--object-path", "/org/gnome/Mutter/DisplayConfig",
		"--method", "org.gnome.Mutter.DisplayConfig.GetCurrentState",
	)
	if err != nil {
		return nil, err
	}

	displays, err := parseGNOMEWaylandDisplayConfig(string(output))
	if err != nil {
		return nil, err
	}
	if len(displays) == 0 {
		return nil, fmt.Errorf("gdbus did not report a connected display with usable size information")
	}

	return displays, nil
}

func parseGNOMEWaylandDisplayConfig(raw string) ([]model.DisplayInfo, error) {
	matches := gnomeWaylandConnectorPattern.FindAllStringSubmatchIndex(raw, -1)
	displays := make([]model.DisplayInfo, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))

	for _, match := range matches {
		name := raw[match[2]:match[3]]
		tailStart := match[1]
		tailEnd := tailStart + 2500
		if tailEnd > len(raw) {
			tailEnd = len(raw)
		}
		tail := raw[tailStart:tailEnd]

		mode := gnomeWaylandModePattern.FindStringSubmatch(tail)
		millimeters := gnomeWaylandMillimeterPattern.FindStringSubmatch(tail)
		if mode == nil || millimeters == nil {
			continue
		}

		widthPx, err := strconv.Atoi(mode[1])
		if err != nil {
			return nil, fmt.Errorf("parse width pixels: %w", err)
		}
		heightPx, err := strconv.Atoi(mode[2])
		if err != nil {
			return nil, fmt.Errorf("parse height pixels: %w", err)
		}
		widthMM, err := strconv.Atoi(millimeters[1])
		if err != nil {
			return nil, fmt.Errorf("parse width millimeters: %w", err)
		}
		heightMM, err := strconv.Atoi(millimeters[2])
		if err != nil {
			return nil, fmt.Errorf("parse height millimeters: %w", err)
		}
		if widthMM <= 0 || heightMM <= 0 {
			continue
		}

		ppi, err := profile.CalculatePPI(widthPx, heightPx, widthMM, heightMM)
		if err != nil {
			return nil, err
		}

		primaryHint := tail
		if len(primaryHint) > 500 {
			primaryHint = primaryHint[:500]
		}
		isPrimary := gnomeWaylandPrimaryHintPattern.MatchString(primaryHint)

		key := strings.Join([]string{
			name,
			strconv.FormatBool(isPrimary),
			strconv.Itoa(widthPx),
			strconv.Itoa(heightPx),
			strconv.Itoa(widthMM),
			strconv.Itoa(heightMM),
		}, "\t")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		displays = append(displays, model.DisplayInfo{
			Name:      name,
			IsPrimary: isPrimary,
			WidthPx:   widthPx,
			HeightPx:  heightPx,
			WidthMM:   widthMM,
			HeightMM:  heightMM,
			PPI:       ppi,
		})
	}

	return displays, nil
}
