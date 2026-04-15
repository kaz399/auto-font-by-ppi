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
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/execx"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/profile"
)

var (
	gnomeWaylandConnectorPattern      = regexp.MustCompile(`'((?:HDMI|DP|eDP|DVI|VGA|Virtual|DisplayPort)[^']*)'`)
	gnomeWaylandModePattern           = regexp.MustCompile(`[,(\[]\s*(\d{3,5})\s*,\s*(\d{3,5})\s*,\s*(?:[0-9]+(?:\.[0-9]+)?)`)
	gnomeWaylandMillimeterPattern     = regexp.MustCompile(`[,(\[]\s*(\d{2,5})\s*,\s*(\d{2,5})\s*(?:,\s*)?[)\]]`)
	gnomeWaylandPrimaryHintPattern    = regexp.MustCompile(`(?i)primary`)
	gnomeWaylandLogicalMonitorPattern = regexp.MustCompile(`\(\s*-?\d+\s*,\s*-?\d+\s*,\s*([0-9]+(?:\.[0-9]+)?)\s*,\s*uint32\s+\d+\s*,\s*(true|false)\s*,\s*\[\('((?:HDMI|DP|eDP|DVI|VGA|Virtual|DisplayPort)[^']*)'`)
)

type logicalMonitorInfo struct {
	scale     float64
	isPrimary bool
}

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
		return nil, fmt.Errorf("gdbus did not report a connected display")
	}

	xrandrOutput, err := b.Runner.Run(ctx, "xrandr", "--query")
	if err == nil {
		xrandrDisplays, parseErr := parseXRandrQuery(string(xrandrOutput))
		if parseErr == nil {
			displays = mergeGNOMEWaylandWithXRandr(displays, xrandrDisplays)
		}
	}

	return displays, nil
}

func parseGNOMEWaylandDisplayConfig(raw string) ([]model.DisplayInfo, error) {
	logicalMonitors, err := parseGNOMEWaylandLogicalMonitors(raw)
	if err != nil {
		return nil, err
	}

	matches := gnomeWaylandConnectorPattern.FindAllStringSubmatchIndex(raw, -1)
	displays := make([]model.DisplayInfo, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))

	for index, match := range matches {
		name := raw[match[2]:match[3]]
		logical, hasLogical := logicalMonitors[name]
		if len(logicalMonitors) > 0 && !hasLogical {
			continue
		}

		tailStart := match[1]
		tailEnd := len(raw)
		if index+1 < len(matches) {
			tailEnd = matches[index+1][0]
		}
		tail := raw[tailStart:tailEnd]

		mode := gnomeWaylandModePattern.FindStringSubmatch(tail)
		if mode == nil {
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
		if hasLogical && logical.scale > 0 {
			widthPx = int(math.Round(float64(widthPx) / logical.scale))
			heightPx = int(math.Round(float64(heightPx) / logical.scale))
		}

		widthMM := 0
		heightMM := 0
		ppi := 0.0
		millimeters := gnomeWaylandMillimeterPattern.FindStringSubmatch(tail)
		if millimeters != nil {
			widthMM, err = strconv.Atoi(millimeters[1])
			if err != nil {
				return nil, fmt.Errorf("parse width millimeters: %w", err)
			}
			heightMM, err = strconv.Atoi(millimeters[2])
			if err != nil {
				return nil, fmt.Errorf("parse height millimeters: %w", err)
			}
			if widthMM > 0 && heightMM > 0 {
				ppi, err = profile.CalculatePPI(widthPx, heightPx, widthMM, heightMM)
				if err != nil {
					return nil, err
				}
			}
		}

		primaryHint := tail
		if len(primaryHint) > 500 {
			primaryHint = primaryHint[:500]
		}
		isPrimary := gnomeWaylandPrimaryHintPattern.MatchString(primaryHint)
		if hasLogical {
			isPrimary = logical.isPrimary
		}

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

func parseGNOMEWaylandLogicalMonitors(raw string) (map[string]logicalMonitorInfo, error) {
	matches := gnomeWaylandLogicalMonitorPattern.FindAllStringSubmatch(raw, -1)
	logicalMonitors := make(map[string]logicalMonitorInfo, len(matches))

	for _, match := range matches {
		scale, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return nil, fmt.Errorf("parse logical monitor scale: %w", err)
		}
		logicalMonitors[match[3]] = logicalMonitorInfo{
			scale:     scale,
			isPrimary: match[2] == "true",
		}
	}

	return logicalMonitors, nil
}

func mergeGNOMEWaylandWithXRandr(gnomeDisplays, xrandrDisplays []model.DisplayInfo) []model.DisplayInfo {
	xrandrByName := make(map[string]model.DisplayInfo, len(xrandrDisplays))
	for _, display := range xrandrDisplays {
		xrandrByName[display.Name] = display
	}

	merged := make([]model.DisplayInfo, 0, len(gnomeDisplays))
	for _, display := range gnomeDisplays {
		if supplemented, ok := xrandrByName[display.Name]; ok &&
			!hasUsablePhysicalMetrics(display) &&
			hasUsablePhysicalMetrics(supplemented) {
			display.WidthMM = supplemented.WidthMM
			display.HeightMM = supplemented.HeightMM
			display.PPI = supplemented.PPI
		}
		merged = append(merged, display)
	}

	return merged
}

func hasUsablePhysicalMetrics(display model.DisplayInfo) bool {
	if display.WidthMM <= 0 || display.HeightMM <= 0 || display.PPI <= 0 {
		return false
	}
	if display.WidthMM == display.WidthPx && display.HeightMM == display.HeightPx {
		return false
	}
	return true
}
