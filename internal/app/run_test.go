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

package app

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/execx"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/target"
)

func TestBuildPlanRespectsTargetOrder(t *testing.T) {
	t.Parallel()

	alphaActions := []model.Action{
		{Target: "alpha", Description: "alpha-1", Command: []string{"alpha", "1"}},
		{Target: "alpha", Description: "alpha-2", Command: []string{"alpha", "2"}},
	}
	betaActions := []model.Action{
		{Target: "beta", Description: "beta-1", Command: []string{"beta", "1"}},
	}

	runner := Runner{
		Targets: adaptersFromValues([]fakeAdapter{
			{name: "alpha", buildActions: alphaActions},
			{name: "beta", buildActions: betaActions},
		}),
	}

	cfg := model.Config{
		DryRun:      true,
		TargetNames: []string{"alpha"},
	}

	plan, err := runner.buildPlan(cfg, testResolvedSettings())
	if err != nil {
		t.Fatalf("buildPlan returned error: %v", err)
	}

	if !plan.DryRun {
		t.Fatal("expected plan to inherit dry-run flag")
	}

	want := alphaActions
	if !reflect.DeepEqual(plan.Actions, want) {
		t.Fatalf("unexpected plan actions: got %v want %v", plan.Actions, want)
	}
}

func TestBuildPlanReturnsErrorForUnknownTarget(t *testing.T) {
	t.Parallel()

	runner := Runner{
		Targets: adaptersFromValues([]fakeAdapter{
			{name: "alpha"},
		}),
	}

	cfg := model.Config{
		TargetNames: []string{"missing"},
	}

	if _, err := runner.buildPlan(cfg, testResolvedSettings()); err == nil {
		t.Fatal("expected unknown target error")
	}
}

func TestBuildPlanPropagatesBuildActionsError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("build failed")
	runner := Runner{
		Targets: adaptersFromValues([]fakeAdapter{
			{name: "alpha", buildErr: wantErr},
		}),
	}

	cfg := model.Config{
		TargetNames: []string{"alpha"},
	}

	_, err := runner.buildPlan(cfg, testResolvedSettings())
	if !errors.Is(err, wantErr) {
		t.Fatalf("unexpected buildPlan error: got %v want %v", err, wantErr)
	}
}

func TestApplyPlanRunsActionsInOrder(t *testing.T) {
	t.Parallel()

	alpha := &fakeAdapter{name: "alpha"}
	beta := &fakeAdapter{name: "beta"}
	execRunner := &fakeExecRunner{}

	runner := Runner{
		Targets:       adaptersFromPointers(alpha, beta),
		CommandRunner: execRunner,
	}

	plan := model.Plan{
		Actions: []model.Action{
			{Target: "beta", Description: "beta-1", Command: []string{"beta", "1"}},
			{Target: "alpha", Description: "alpha-1", Command: []string{"alpha", "1"}},
			{Target: "beta", Description: "beta-2", Command: []string{"beta", "2"}},
		},
	}

	if err := runner.applyPlan(context.Background(), plan); err != nil {
		t.Fatalf("applyPlan returned error: %v", err)
	}

	if !reflect.DeepEqual(alpha.appliedDescriptions, []string{"alpha-1"}) {
		t.Fatalf("unexpected alpha apply order: got %v", alpha.appliedDescriptions)
	}
	if !reflect.DeepEqual(beta.appliedDescriptions, []string{"beta-1", "beta-2"}) {
		t.Fatalf("unexpected beta apply order: got %v", beta.appliedDescriptions)
	}
	if beta.lastRunner != execRunner || alpha.lastRunner != execRunner {
		t.Fatal("expected applyPlan to pass the configured command runner to adapters")
	}
}

func TestApplyPlanReturnsErrorForUnknownActionTarget(t *testing.T) {
	t.Parallel()

	runner := Runner{
		Targets: adaptersFromValues([]fakeAdapter{
			{name: "alpha"},
		}),
		CommandRunner: &fakeExecRunner{},
	}

	plan := model.Plan{
		Actions: []model.Action{
			{Target: "missing", Description: "missing"},
		},
	}

	if err := runner.applyPlan(context.Background(), plan); err == nil {
		t.Fatal("expected unknown target error")
	}
}

func TestApplyPlanStopsOnAdapterError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("apply failed")
	alpha := &fakeAdapter{name: "alpha", applyErr: wantErr}
	beta := &fakeAdapter{name: "beta"}

	runner := Runner{
		Targets:       adaptersFromPointers(alpha, beta),
		CommandRunner: &fakeExecRunner{},
	}

	plan := model.Plan{
		Actions: []model.Action{
			{Target: "alpha", Description: "alpha-1"},
			{Target: "beta", Description: "beta-1"},
		},
	}

	err := runner.applyPlan(context.Background(), plan)
	if !errors.Is(err, wantErr) {
		t.Fatalf("unexpected applyPlan error: got %v want %v", err, wantErr)
	}
	if !reflect.DeepEqual(alpha.appliedDescriptions, []string{"alpha-1"}) {
		t.Fatalf("unexpected alpha apply log: got %v", alpha.appliedDescriptions)
	}
	if len(beta.appliedDescriptions) != 0 {
		t.Fatalf("expected beta adapter not to run after error, got %v", beta.appliedDescriptions)
	}
}

func TestPrintPlanShowsDiagonalOverrideNote(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	runner := Runner{
		Output: &output,
	}

	runner.printPlan(model.Plan{
		DryRun: true,
		Resolved: model.ResolvedSettings{
			SourceBackend: "gnome-wayland",
			Display: model.DisplayInfo{
				Name:                   "HDMI-1",
				WidthPx:                3072,
				HeightPx:               1728,
				WidthMM:                698,
				HeightMM:               393,
				PPI:                    112.11,
				PhysicalSizeSource:     "override",
				OverrideDiagonalInches: 31.5,
			},
			Profile: model.Profile{
				Name:              "ppi-140",
				TextScaling:       1.0,
				UIFontSize:        12,
				DocumentFontSize:  12,
				MonospaceFontSize: 11,
				TitlebarFontSize:  12,
			},
		},
	})

	if !strings.Contains(output.String(), "Selected display: HDMI-1 (3072x1728 px, 698x393 mm, PPI=112.11, diagonal override=31.50 in)") {
		t.Fatalf("expected selected display output to mention diagonal override, got:\n%s", output.String())
	}
}

func TestPrintPlanShowsAssumedPPINote(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	runner := Runner{
		Output: &output,
	}

	runner.printPlan(model.Plan{
		DryRun: true,
		Resolved: model.ResolvedSettings{
			SourceBackend: "gnome-wayland",
			Display: model.DisplayInfo{
				Name:               "HDMI-1",
				WidthPx:            3072,
				HeightPx:           1728,
				WidthMM:            780,
				HeightMM:           439,
				PPI:                100.0,
				PhysicalSizeSource: "assumed-ppi",
				AssumedPPI:         100.0,
			},
			Profile: model.Profile{
				Name:              "ppi-110",
				TextScaling:       1.0,
				UIFontSize:        11,
				DocumentFontSize:  11,
				MonospaceFontSize: 10,
				TitlebarFontSize:  11,
			},
		},
	})

	if !strings.Contains(output.String(), "Selected display: HDMI-1 (3072x1728 px, 780x439 mm, PPI=100.00, assumed PPI=100.00)") {
		t.Fatalf("expected selected display output to mention assumed PPI, got:\n%s", output.String())
	}
}

type fakeAdapter struct {
	name                string
	buildActions        []model.Action
	buildErr            error
	applyErr            error
	appliedDescriptions []string
	lastRunner          execx.Runner
}

func (a *fakeAdapter) Name() string {
	return a.name
}

func (a *fakeAdapter) BuildActions(_ model.Config, _ model.ResolvedSettings) ([]model.Action, error) {
	if a.buildErr != nil {
		return nil, a.buildErr
	}
	return append([]model.Action(nil), a.buildActions...), nil
}

func (a *fakeAdapter) Apply(_ context.Context, action model.Action, runner execx.Runner) error {
	a.appliedDescriptions = append(a.appliedDescriptions, action.Description)
	a.lastRunner = runner
	if a.applyErr != nil {
		return a.applyErr
	}
	return nil
}

type fakeExecRunner struct{}

func (*fakeExecRunner) Run(_ context.Context, _ string, _ ...string) ([]byte, error) {
	return nil, nil
}

func testResolvedSettings() model.ResolvedSettings {
	return model.ResolvedSettings{
		Profile: model.Profile{
			Name: "default",
		},
	}
}

func adaptersFromValues(fixtures []fakeAdapter) []target.Adapter {
	adapters := make([]target.Adapter, 0, len(fixtures))
	for index := range fixtures {
		adapter := fixtures[index]
		adapters = append(adapters, &adapter)
	}
	return adapters
}

func adaptersFromPointers(fixtures ...*fakeAdapter) []target.Adapter {
	adapters := make([]target.Adapter, 0, len(fixtures))
	for _, adapter := range fixtures {
		adapters = append(adapters, adapter)
	}
	return adapters
}
