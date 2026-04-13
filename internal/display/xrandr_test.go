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
	"math"
	"testing"
)

func TestParseXRandrQuery(t *testing.T) {
	t.Parallel()

	raw := `Screen 0: minimum 320 x 200, current 5760 x 2160, maximum 16384 x 16384
eDP-1 connected primary 2880x1800+0+360 (normal left inverted right x axis y axis) 302mm x 189mm
HDMI-1 connected 3840x2160+2880+0 (normal left inverted right x axis y axis) 698mm x 392mm
DP-1 disconnected (normal left inverted right x axis y axis)`

	displays, err := parseXRandrQuery(raw)
	if err != nil {
		t.Fatalf("parseXRandrQuery returned error: %v", err)
	}

	if got := len(displays); got != 2 {
		t.Fatalf("unexpected display count: got %d want 2", got)
	}
	if displays[0].Name != "eDP-1" || !displays[0].IsPrimary {
		t.Fatalf("unexpected first display: %+v", displays[0])
	}
	if diff := math.Abs(displays[0].PPI - 242.31); diff > 1.0 {
		t.Fatalf("unexpected eDP-1 PPI: got %.2f want about 242.31", displays[0].PPI)
	}
	if displays[1].Name != "HDMI-1" || displays[1].IsPrimary {
		t.Fatalf("unexpected second display: %+v", displays[1])
	}
}

func TestParseXRandrQueryKeepsDisplayWithoutPhysicalSize(t *testing.T) {
	t.Parallel()

	raw := `HDMI-1 connected primary 3840x2160+0+0 (normal left inverted right x axis y axis)
DP-1 connected 1920x1080+3840+0 (normal left inverted right x axis y axis) 0mm x 0mm`

	displays, err := parseXRandrQuery(raw)
	if err != nil {
		t.Fatalf("parseXRandrQuery returned error: %v", err)
	}

	if got := len(displays); got != 2 {
		t.Fatalf("unexpected display count: got %d want 2", got)
	}
	if displays[0].WidthMM != 0 || displays[0].HeightMM != 0 || displays[0].PPI != 0 {
		t.Fatalf("expected HDMI-1 to keep missing physical size as zero values: %+v", displays[0])
	}
	if displays[1].WidthMM != 0 || displays[1].HeightMM != 0 || displays[1].PPI != 0 {
		t.Fatalf("expected DP-1 to keep invalid physical size as zero values: %+v", displays[1])
	}
}
