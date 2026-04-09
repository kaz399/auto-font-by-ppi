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

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
)

func TestGNOMEBuildActionsFullFonts(t *testing.T) {
	t.Parallel()

	adapter := GNOMEAdapter{}
	cfg := testConfig()
	cfg.Targets.GNOME.Mode = "full_fonts"

	actions, err := adapter.BuildActions(cfg, testResolvedSettings())
	if err != nil {
		t.Fatalf("BuildActions returned error: %v", err)
	}

	if got := len(actions); got != 5 {
		t.Fatalf("unexpected action count: got %d want 5", got)
	}

	wantFirst := []string{"gsettings", "set", "org.gnome.desktop.interface", "text-scaling-factor", "1.25"}
	if !reflect.DeepEqual(actions[0].Command, wantFirst) {
		t.Fatalf("unexpected first command: got %v want %v", actions[0].Command, wantFirst)
	}

	wantLast := []string{"gsettings", "set", "org.gnome.desktop.wm.preferences", "titlebar-font", "Cantarell Bold 15"}
	if !reflect.DeepEqual(actions[4].Command, wantLast) {
		t.Fatalf("unexpected last command: got %v want %v", actions[4].Command, wantLast)
	}
}

func TestGNOMEBuildActionsScalingOnly(t *testing.T) {
	t.Parallel()

	adapter := GNOMEAdapter{}
	cfg := testConfig()
	cfg.Targets.GNOME.Mode = "scaling_only"

	actions, err := adapter.BuildActions(cfg, testResolvedSettings())
	if err != nil {
		t.Fatalf("BuildActions returned error: %v", err)
	}

	if got := len(actions); got != 1 {
		t.Fatalf("unexpected action count: got %d want 1", got)
	}
}

func TestGNOMEBuildActionsRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()

	adapter := GNOMEAdapter{}
	cfg := testConfig()
	cfg.Targets.GNOME.Mode = "invalid"

	if _, err := adapter.BuildActions(cfg, testResolvedSettings()); err == nil {
		t.Fatal("expected unsupported mode error")
	}
}

func TestGNOMEApplyRunsCommand(t *testing.T) {
	t.Parallel()

	adapter := GNOMEAdapter{}
	runner := &recordingRunner{}
	action := model.Action{
		Target:      "gnome",
		Description: "Set GNOME text scaling factor",
		Command:     []string{"gsettings", "set", "org.gnome.desktop.interface", "text-scaling-factor", "1.25"},
	}

	if err := adapter.Apply(context.Background(), action, runner); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	wantCommand := []string{"gsettings", "set", "org.gnome.desktop.interface", "text-scaling-factor", "1.25"}
	if !reflect.DeepEqual(runner.lastCommand, wantCommand) {
		t.Fatalf("unexpected runner command: got %v want %v", runner.lastCommand, wantCommand)
	}
}
