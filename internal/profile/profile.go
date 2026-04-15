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

package profile

import (
	"fmt"
	"math"

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
)

func CalculatePPI(widthPx, heightPx, widthMM, heightMM int) (float64, error) {
	if widthMM <= 0 || heightMM <= 0 {
		return 0, fmt.Errorf("physical display size must be positive")
	}

	diagonalPx := math.Hypot(float64(widthPx), float64(heightPx))
	diagonalIn := math.Hypot(float64(widthMM)/25.4, float64(heightMM)/25.4)
	if diagonalIn == 0 {
		return 0, fmt.Errorf("physical display diagonal must not be zero")
	}

	return diagonalPx / diagonalIn, nil
}

func SelectDisplay(displays []model.DisplayInfo, preferred string) (model.DisplayInfo, error) {
	if len(displays) == 0 {
		return model.DisplayInfo{}, fmt.Errorf("no displays available")
	}

	if preferred != "" {
		for _, display := range displays {
			if display.Name == preferred {
				return display, nil
			}
		}
		return model.DisplayInfo{}, fmt.Errorf("preferred display %q was not found", preferred)
	}

	for _, display := range displays {
		if display.IsPrimary {
			return display, nil
		}
	}

	return displays[0], nil
}

func SelectProfile(profiles []model.Profile, ppi float64) (model.Profile, error) {
	for _, profile := range profiles {
		if ppi <= profile.MaxPPI {
			return profile, nil
		}
	}

	return model.Profile{}, fmt.Errorf("no matching profile for PPI %.2f", ppi)
}

func NormalizeDisplays(cfg model.DisplayConfig, displays []model.DisplayInfo) ([]model.DisplayInfo, error) {
	normalized := make([]model.DisplayInfo, 0, len(displays))
	for _, display := range displays {
		current := display

		if override, ok := cfg.DiagonalOverrides[current.Name]; ok {
			adjusted, err := ApplyDiagonalOverride(current, override)
			if err != nil {
				return nil, err
			}
			current = adjusted
		} else if current.WidthMM <= 0 || current.HeightMM <= 0 || current.PPI <= 0 {
			adjusted, err := ApplyAssumedPPI(current, 100)
			if err != nil {
				return nil, err
			}
			current = adjusted
		}

		if current.WidthMM <= 0 || current.HeightMM <= 0 || current.PPI <= 0 {
			continue
		}
		normalized = append(normalized, current)
	}
	return normalized, nil
}

func IsSuspicious(cfg model.DisplayConfig, display model.DisplayInfo) bool {
	if display.WidthMM <= 0 || display.HeightMM <= 0 {
		return true
	}
	if display.WidthMM == display.WidthPx && display.HeightMM == display.HeightPx {
		return true
	}
	if display.PPI < cfg.MinReasonablePPI || display.PPI > cfg.MaxReasonablePPI {
		return true
	}
	return false
}

func ApplyDiagonalOverride(display model.DisplayInfo, diagonalInches float64) (model.DisplayInfo, error) {
	if diagonalInches <= 0 {
		return model.DisplayInfo{}, fmt.Errorf("diagonal override must be positive")
	}

	diagonalPx := math.Hypot(float64(display.WidthPx), float64(display.HeightPx))
	if diagonalPx == 0 {
		return model.DisplayInfo{}, fmt.Errorf("display resolution diagonal must not be zero")
	}

	widthMM := int(math.Round((float64(display.WidthPx) / diagonalPx) * diagonalInches * 25.4))
	heightMM := int(math.Round((float64(display.HeightPx) / diagonalPx) * diagonalInches * 25.4))
	ppi, err := CalculatePPI(display.WidthPx, display.HeightPx, widthMM, heightMM)
	if err != nil {
		return model.DisplayInfo{}, err
	}

	display.WidthMM = widthMM
	display.HeightMM = heightMM
	display.PPI = ppi
	display.PhysicalSizeSource = "override"
	display.OverrideDiagonalInches = diagonalInches
	display.AssumedPPI = 0
	return display, nil
}

func ApplyAssumedPPI(display model.DisplayInfo, assumedPPI float64) (model.DisplayInfo, error) {
	if assumedPPI <= 0 {
		return model.DisplayInfo{}, fmt.Errorf("assumed PPI must be positive")
	}

	diagonalPx := math.Hypot(float64(display.WidthPx), float64(display.HeightPx))
	if diagonalPx == 0 {
		return model.DisplayInfo{}, fmt.Errorf("display resolution diagonal must not be zero")
	}

	diagonalInches := diagonalPx / assumedPPI
	adjusted, err := ApplyDiagonalOverride(display, diagonalInches)
	if err != nil {
		return model.DisplayInfo{}, err
	}
	adjusted.PhysicalSizeSource = "assumed-ppi"
	adjusted.AssumedPPI = assumedPPI
	adjusted.OverrideDiagonalInches = 0
	return adjusted, nil
}
