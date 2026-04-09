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
	modePattern = regexp.MustCompile(`\b(\d+)x(\d+)\+\d+\+\d+\b`)
	mmPattern   = regexp.MustCompile(`\b(\d+)mm x (\d+)mm\b`)
)

type XRandrBackend struct {
	Runner execx.Runner
}

func (XRandrBackend) Name() string {
	return "xrandr"
}

func (b XRandrBackend) Detect(ctx context.Context) ([]model.DisplayInfo, error) {
	output, err := b.Runner.Run(ctx, "xrandr", "--query")
	if err != nil {
		return nil, err
	}

	var displays []model.DisplayInfo
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, " connected") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		modeMatch := modePattern.FindStringSubmatch(line)
		sizeMatch := mmPattern.FindStringSubmatch(line)
		if modeMatch == nil || sizeMatch == nil {
			continue
		}

		widthPx, err := strconv.Atoi(modeMatch[1])
		if err != nil {
			return nil, fmt.Errorf("parse width pixels: %w", err)
		}
		heightPx, err := strconv.Atoi(modeMatch[2])
		if err != nil {
			return nil, fmt.Errorf("parse height pixels: %w", err)
		}
		widthMM, err := strconv.Atoi(sizeMatch[1])
		if err != nil {
			return nil, fmt.Errorf("parse width millimeters: %w", err)
		}
		heightMM, err := strconv.Atoi(sizeMatch[2])
		if err != nil {
			return nil, fmt.Errorf("parse height millimeters: %w", err)
		}

		ppi, err := profile.CalculatePPI(widthPx, heightPx, widthMM, heightMM)
		if err != nil {
			return nil, err
		}

		displays = append(displays, model.DisplayInfo{
			Name:      fields[0],
			IsPrimary: strings.Contains(" "+line+" ", " primary "),
			WidthPx:   widthPx,
			HeightPx:  heightPx,
			WidthMM:   widthMM,
			HeightMM:  heightMM,
			PPI:       ppi,
		})
	}

	if len(displays) == 0 {
		return nil, fmt.Errorf("xrandr did not report a connected display with usable size information")
	}

	return displays, nil
}
