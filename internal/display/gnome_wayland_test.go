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

func TestParseGNOMEWaylandDisplayConfig(t *testing.T) {
	t.Parallel()

	raw := `([('eDP-1', 'Built-in display', [(0, 0, 1)], [(2880, 1800, 60.0)], [(302, 189)], 'primary'),
('eDP-1', 'Built-in display', [(0, 0, 1)], [(2880, 1800, 60.0)], [(302, 189)], 'primary'),
('HDMI-1', 'External display', [(2880, 0, 1)], [(3840, 2160, 60.0)], [(698, 392)], 'presentation')], @a{})`

	displays, err := parseGNOMEWaylandDisplayConfig(raw)
	if err != nil {
		t.Fatalf("parseGNOMEWaylandDisplayConfig returned error: %v", err)
	}

	if got := len(displays); got != 2 {
		t.Fatalf("unexpected display count: got %d want 2", got)
	}

	if displays[0].Name != "eDP-1" {
		t.Fatalf("unexpected first display name: got %q want %q", displays[0].Name, "eDP-1")
	}
	if !displays[0].IsPrimary {
		t.Fatalf("expected eDP-1 to be marked primary")
	}
	if diff := math.Abs(displays[0].PPI - 242.31); diff > 1.0 {
		t.Fatalf("unexpected eDP-1 PPI: got %.2f want about 242.31", displays[0].PPI)
	}

	if displays[1].Name != "HDMI-1" {
		t.Fatalf("unexpected second display name: got %q want %q", displays[1].Name, "HDMI-1")
	}
	if displays[1].IsPrimary {
		t.Fatalf("expected HDMI-1 not to be marked primary")
	}
}
