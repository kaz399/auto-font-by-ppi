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

package cli

import (
	"strings"
	"testing"

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
)

func TestParseAcceptsDisplayDiagonalOverrides(t *testing.T) {
	t.Parallel()

	options, err := Parse([]string{
		"--display-diagonal", "HDMI-1=31.5",
		"--display-diagonal", "eDP-1=14.0",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if got := len(options.DisplayDiagonalOverrides); got != 2 {
		t.Fatalf("unexpected override count: got %d want 2", got)
	}
	if got := options.DisplayDiagonalOverrides["HDMI-1"]; got != 31.5 {
		t.Fatalf("unexpected HDMI-1 override: got %v want 31.5", got)
	}
	if got := options.DisplayDiagonalOverrides["eDP-1"]; got != 14.0 {
		t.Fatalf("unexpected eDP-1 override: got %v want 14.0", got)
	}
}

func TestParseRejectsInvalidDisplayDiagonalOverride(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"HDMI-1", "=31.5", "HDMI-1=0", "HDMI-1=-1", "HDMI-1=abc"} {
		_, err := Parse([]string{"--display-diagonal", value})
		if err == nil {
			t.Fatalf("expected Parse to fail for %q", value)
		}
	}
}

func TestParseRejectsRemovedApplyFlag(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"--apply"})
	if err == nil {
		t.Fatal("expected Parse to fail for removed --apply flag")
	}
	if !strings.Contains(err.Error(), "flag provided but not defined: -apply") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseAcceptsCurrentDisplayOverrides(t *testing.T) {
	t.Parallel()

	options, err := Parse([]string{"--inch", "27", "--ppi", "110", "--inch", "31.5"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if options.CurrentDisplayOverride == nil {
		t.Fatal("expected current display override to be set")
	}
	if got := options.CurrentDisplayOverride.Mode; got != "inch" {
		t.Fatalf("unexpected override mode: got %q want %q", got, "inch")
	}
	if got := options.CurrentDisplayOverride.Value; got != 31.5 {
		t.Fatalf("unexpected override value: got %v want 31.5", got)
	}
}

func TestParseRejectsInvalidCurrentDisplayOverrides(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		args []string
		want string
	}{
		{args: []string{"--inch", "0"}, want: "invalid inch value"},
		{args: []string{"--inch", "abc"}, want: "invalid inch value"},
		{args: []string{"--ppi", "-1"}, want: "invalid ppi value"},
		{args: []string{"--ppi", "abc"}, want: "invalid ppi value"},
	} {
		_, err := Parse(testCase.args)
		if err == nil {
			t.Fatalf("expected Parse to fail for %v", testCase.args)
		}
		if !strings.Contains(err.Error(), testCase.want) {
			t.Fatalf("unexpected error for %v: %v", testCase.args, err)
		}
	}
}

func TestApplyMergesDisplayDiagonalOverrides(t *testing.T) {
	t.Parallel()

	cfg := model.Config{
		Display: model.DisplayConfig{
			DiagonalOverrides: map[string]float64{
				"HDMI-1": 30.0,
			},
		},
	}
	options := Options{
		CurrentDisplayOverride: &model.DisplayOverride{
			Mode:  "ppi",
			Value: 123,
		},
		DisplayDiagonalOverrides: map[string]float64{
			"HDMI-1": 31.5,
			"eDP-1":  14.0,
		},
	}

	Apply(&cfg, options)

	if got := cfg.Display.DiagonalOverrides["HDMI-1"]; got != 31.5 {
		t.Fatalf("unexpected HDMI-1 override after Apply: got %v want 31.5", got)
	}
	if got := cfg.Display.DiagonalOverrides["eDP-1"]; got != 14.0 {
		t.Fatalf("unexpected eDP-1 override after Apply: got %v want 14.0", got)
	}
	if cfg.CurrentDisplayOverride == nil {
		t.Fatal("expected Apply to set current display override")
	}
	if got := cfg.CurrentDisplayOverride.Mode; got != "ppi" {
		t.Fatalf("unexpected current display override mode: got %q want %q", got, "ppi")
	}
	if got := cfg.CurrentDisplayOverride.Value; got != 123 {
		t.Fatalf("unexpected current display override value: got %v want 123", got)
	}
}

func TestParseAcceptsAutoFlag(t *testing.T) {
	t.Parallel()

	options, err := Parse([]string{"--auto"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if !options.Auto {
		t.Fatal("expected Auto to be true")
	}
}

func TestParseRejectsAutoAndDisplayTogether(t *testing.T) {
	t.Parallel()

	_, err := Parse([]string{"--auto", "--display", "HDMI-1"})
	if err == nil {
		t.Fatal("expected Parse to fail when --auto and --display are used together")
	}
	if !strings.Contains(err.Error(), "--auto and --display cannot be used together") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyClearsPreferredDisplayWhenAutoIsTrue(t *testing.T) {
	t.Parallel()

	cfg := model.Config{
		PreferredDisplay: "HDMI-1",
	}
	options := Options{
		Auto: true,
	}

	Apply(&cfg, options)

	if cfg.PreferredDisplay != "" {
		t.Fatalf("expected PreferredDisplay to be cleared, got %q", cfg.PreferredDisplay)
	}
}

