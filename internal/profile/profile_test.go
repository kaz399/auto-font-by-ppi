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
	"math"
	"testing"

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
)

func TestNormalizeDisplaysAppliesDiagonalOverride(t *testing.T) {
	t.Parallel()

	cfg := model.DisplayConfig{
		MinReasonablePPI: 50,
		MaxReasonablePPI: 400,
		DiagonalOverrides: map[string]float64{
			"eDP-1": 14.0,
		},
	}

	displays := []model.DisplayInfo{
		{
			Name:      "eDP-1",
			IsPrimary: true,
			WidthPx:   2880,
			HeightPx:  1800,
			WidthMM:   2880,
			HeightMM:  1800,
			PPI:       25.40,
		},
	}

	normalized, err := NormalizeDisplays(cfg, displays)
	if err != nil {
		t.Fatalf("NormalizeDisplays returned error: %v", err)
	}

	display := normalized[0]
	if display.WidthMM == displays[0].WidthMM || display.HeightMM == displays[0].HeightMM {
		t.Fatalf("expected diagonal override to change physical size, got %dx%d", display.WidthMM, display.HeightMM)
	}
	if IsSuspicious(cfg, display) {
		t.Fatalf("expected normalized display to be within a reasonable range: %+v", display)
	}
	if diff := math.Abs(display.PPI - 242.31); diff > 0.5 {
		t.Fatalf("unexpected PPI after override: got %.2f want about 242.31", display.PPI)
	}
	if display.PhysicalSizeSource != "override" {
		t.Fatalf("unexpected physical size source: got %q want %q", display.PhysicalSizeSource, "override")
	}
	if diff := math.Abs(display.OverrideDiagonalInches - 14.0); diff > 0.01 {
		t.Fatalf("unexpected override diagonal: got %.2f want 14.00", display.OverrideDiagonalInches)
	}
}

func TestNormalizeDisplaysAppliesDiagonalOverrideEvenWhenDetectedMetricsAreUsable(t *testing.T) {
	t.Parallel()

	cfg := model.DisplayConfig{
		MinReasonablePPI: 50,
		MaxReasonablePPI: 400,
		DiagonalOverrides: map[string]float64{
			"HDMI-1": 27.0,
		},
	}

	displays := []model.DisplayInfo{
		{
			Name:      "HDMI-1",
			IsPrimary: true,
			WidthPx:   3840,
			HeightPx:  2160,
			WidthMM:   600,
			HeightMM:  340,
			PPI:       129.82,
		},
	}

	normalized, err := NormalizeDisplays(cfg, displays)
	if err != nil {
		t.Fatalf("NormalizeDisplays returned error: %v", err)
	}

	if got := len(normalized); got != 1 {
		t.Fatalf("unexpected normalized display count: got %d want 1", got)
	}
	if normalized[0].WidthMM == displays[0].WidthMM || normalized[0].HeightMM == displays[0].HeightMM {
		t.Fatalf("expected override to replace detected physical size: got %+v", normalized[0])
	}
	if normalized[0].PhysicalSizeSource != "override" {
		t.Fatalf("unexpected physical size source: got %q want %q", normalized[0].PhysicalSizeSource, "override")
	}
}

func TestNormalizeDisplaysAppliesDiagonalOverrideToDisplayWithoutPhysicalSize(t *testing.T) {
	t.Parallel()

	cfg := model.DisplayConfig{
		MinReasonablePPI: 50,
		MaxReasonablePPI: 400,
		DiagonalOverrides: map[string]float64{
			"HDMI-1": 31.5,
		},
	}

	displays := []model.DisplayInfo{
		{
			Name:      "HDMI-1",
			IsPrimary: true,
			WidthPx:   3840,
			HeightPx:  2160,
			WidthMM:   0,
			HeightMM:  0,
			PPI:       0,
		},
	}

	normalized, err := NormalizeDisplays(cfg, displays)
	if err != nil {
		t.Fatalf("NormalizeDisplays returned error: %v", err)
	}

	if got := len(normalized); got != 1 {
		t.Fatalf("unexpected normalized display count: got %d want 1", got)
	}
	if normalized[0].WidthMM <= 0 || normalized[0].HeightMM <= 0 || normalized[0].PPI <= 0 {
		t.Fatalf("expected manual override to restore usable physical size: %+v", normalized[0])
	}
	if normalized[0].PhysicalSizeSource != "override" {
		t.Fatalf("unexpected physical size source: got %q want %q", normalized[0].PhysicalSizeSource, "override")
	}
}

func TestNormalizeDisplaysFallsBackToAssumedPPIWhenPhysicalSizeIsMissing(t *testing.T) {
	t.Parallel()

	cfg := model.DisplayConfig{
		MinReasonablePPI: 50,
		MaxReasonablePPI: 400,
	}

	displays := []model.DisplayInfo{
		{
			Name:      "HDMI-1",
			IsPrimary: true,
			WidthPx:   3840,
			HeightPx:  2160,
			WidthMM:   0,
			HeightMM:  0,
			PPI:       0,
		},
	}

	normalized, err := NormalizeDisplays(cfg, displays)
	if err != nil {
		t.Fatalf("NormalizeDisplays returned error: %v", err)
	}

	if got := len(normalized); got != 1 {
		t.Fatalf("unexpected normalized display count: got %d want 1", got)
	}
	if diff := math.Abs(normalized[0].PPI - 100.0); diff > 1.0 {
		t.Fatalf("expected fallback PPI around 100, got %.2f", normalized[0].PPI)
	}
	if normalized[0].PhysicalSizeSource != "assumed-ppi" {
		t.Fatalf("unexpected physical size source: got %q want %q", normalized[0].PhysicalSizeSource, "assumed-ppi")
	}
}

func TestNormalizeDisplaysFallsBackToAssumedPPIWhenMetricsAreSuspicious(t *testing.T) {
	t.Parallel()

	cfg := model.DisplayConfig{
		MinReasonablePPI: 50,
		MaxReasonablePPI: 400,
	}

	displays := []model.DisplayInfo{
		{
			Name:      "eDP-1",
			IsPrimary: true,
			WidthPx:   2880,
			HeightPx:  1800,
			WidthMM:   2880,
			HeightMM:  1800,
			PPI:       25.4,
		},
		{
			Name:      "HDMI-1",
			IsPrimary: false,
			WidthPx:   3840,
			HeightPx:  2160,
			WidthMM:   600,
			HeightMM:  340,
			PPI:       999,
		},
	}

	normalized, err := NormalizeDisplays(cfg, displays)
	if err != nil {
		t.Fatalf("NormalizeDisplays returned error: %v", err)
	}

	if got := len(normalized); got != 2 {
		t.Fatalf("unexpected normalized display count: got %d want 2", got)
	}

	if normalized[0].PhysicalSizeSource != "assumed-ppi" {
		t.Fatalf("unexpected first display physical size source: got %q want %q", normalized[0].PhysicalSizeSource, "assumed-ppi")
	}
	if diff := math.Abs(normalized[0].PPI - 100.0); diff > 1.0 {
		t.Fatalf("expected first display fallback PPI around 100, got %.2f", normalized[0].PPI)
	}

	if normalized[1].PhysicalSizeSource != "assumed-ppi" {
		t.Fatalf("unexpected second display physical size source: got %q want %q", normalized[1].PhysicalSizeSource, "assumed-ppi")
	}
	if diff := math.Abs(normalized[1].PPI - 100.0); diff > 1.0 {
		t.Fatalf("expected second display fallback PPI around 100, got %.2f", normalized[1].PPI)
	}
}

func TestSelectDisplay(t *testing.T) {
	t.Parallel()

	displays := []model.DisplayInfo{
		{Name: "HDMI-1"},
		{Name: "eDP-1", IsPrimary: true},
	}

	preferred, err := SelectDisplay(displays, "HDMI-1")
	if err != nil {
		t.Fatalf("SelectDisplay with preferred display returned error: %v", err)
	}
	if preferred.Name != "HDMI-1" {
		t.Fatalf("unexpected preferred display: got %q want %q", preferred.Name, "HDMI-1")
	}

	primary, err := SelectDisplay(displays, "")
	if err != nil {
		t.Fatalf("SelectDisplay with primary fallback returned error: %v", err)
	}
	if primary.Name != "eDP-1" {
		t.Fatalf("unexpected primary display: got %q want %q", primary.Name, "eDP-1")
	}
}

func TestSelectProfile(t *testing.T) {
	t.Parallel()

	profiles := []model.Profile{
		{Name: "low", MaxPPI: 110},
		{Name: "high", MaxPPI: 220},
	}

	selected, err := SelectProfile(profiles, 150)
	if err != nil {
		t.Fatalf("SelectProfile returned error: %v", err)
	}
	if selected.Name != "high" {
		t.Fatalf("unexpected profile: got %q want %q", selected.Name, "high")
	}
}
