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
	"reflect"
	"testing"
)

func TestKittyBuildActionsRemoteWithSocket(t *testing.T) {
	t.Parallel()

	adapter := KittyAdapter{}
	cfg := testConfig()
	cfg.Targets.Kitty.Enabled = true
	cfg.Targets.Kitty.Strategy = "remote"
	cfg.Targets.Kitty.Socket = "unix:/tmp/kitty.sock"
	cfg.Targets.Kitty.All = true
	cfg.Targets.Kitty.FontSizeField = "monospace_font_size"

	actions, err := adapter.BuildActions(cfg, testResolvedSettings())
	if err != nil {
		t.Fatalf("BuildActions returned error: %v", err)
	}

	want := []string{"kitten", "@", "--to", "unix:/tmp/kitty.sock", "set-font-size", "--all", "13"}
	if got := len(actions); got != 1 {
		t.Fatalf("unexpected action count: got %d want 1", got)
	}
	if !reflect.DeepEqual(actions[0].Command, want) {
		t.Fatalf("unexpected kitty command: got %v want %v", actions[0].Command, want)
	}
}

func TestKittyBuildActionsUsesConfiguredFontField(t *testing.T) {
	t.Parallel()

	adapter := KittyAdapter{}
	cfg := testConfig()
	cfg.Targets.Kitty.Enabled = true
	cfg.Targets.Kitty.Strategy = "remote"
	cfg.Targets.Kitty.All = false
	cfg.Targets.Kitty.FontSizeField = "ui_font_size"

	actions, err := adapter.BuildActions(cfg, testResolvedSettings())
	if err != nil {
		t.Fatalf("BuildActions returned error: %v", err)
	}

	want := []string{"kitten", "@", "set-font-size", "14"}
	if !reflect.DeepEqual(actions[0].Command, want) {
		t.Fatalf("unexpected kitty command: got %v want %v", actions[0].Command, want)
	}
}

func TestKittyBuildActionsRejectsUnsupportedStrategy(t *testing.T) {
	t.Parallel()

	adapter := KittyAdapter{}
	cfg := testConfig()
	cfg.Targets.Kitty.Enabled = true
	cfg.Targets.Kitty.Strategy = "config-file"

	if _, err := adapter.BuildActions(cfg, testResolvedSettings()); err == nil {
		t.Fatal("expected unsupported strategy error")
	}
}

func TestKittyBuildActionsRejectsUnsupportedFontField(t *testing.T) {
	t.Parallel()

	adapter := KittyAdapter{}
	cfg := testConfig()
	cfg.Targets.Kitty.Enabled = true
	cfg.Targets.Kitty.Strategy = "remote"
	cfg.Targets.Kitty.FontSizeField = "titlebar_font_size"

	if _, err := adapter.BuildActions(cfg, testResolvedSettings()); err == nil {
		t.Fatal("expected unsupported font field error")
	}
}

func TestKittyApplyRunsCommand(t *testing.T) {
	t.Parallel()

	adapter := KittyAdapter{}
	runner := &recordingRunner{}

	if err := adapter.Apply(context.Background(), testKittyAction(), runner); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	want := []string{"kitten", "@", "set-font-size", "--all", "13"}
	if !reflect.DeepEqual(runner.lastCommand, want) {
		t.Fatalf("unexpected runner command: got %v want %v", runner.lastCommand, want)
	}
}
