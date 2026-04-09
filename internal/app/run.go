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
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/display"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/execx"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/profile"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/target"
)

type Runner struct {
	DisplayBackends []display.Backend
	Targets         []target.Adapter
	CommandRunner   execx.Runner
	Output          io.Writer
	Verbose         bool
}

func (r Runner) Run(ctx context.Context, cfg model.Config) error {
	displays, backendName, err := r.detectDisplays(ctx, cfg)
	if err != nil {
		return err
	}

	normalized, err := profile.NormalizeDisplays(cfg.Display, displays)
	if err != nil {
		return err
	}

	selected, err := profile.SelectDisplay(normalized, cfg.PreferredDisplay)
	if err != nil {
		return err
	}

	selectedProfile, err := profile.SelectProfile(cfg.Profiles, selected.PPI)
	if err != nil {
		return err
	}

	resolved := model.ResolvedSettings{
		Display:             selected,
		Profile:             selectedProfile,
		Fonts:               cfg.Fonts,
		SelectedTargetNames: cfg.TargetNames,
		SourceBackend:       backendName,
	}

	plan, err := r.buildPlan(cfg, resolved)
	if err != nil {
		return err
	}

	r.printPlan(plan)
	if plan.DryRun {
		return nil
	}

	return r.applyPlan(ctx, plan)
}

func (r Runner) detectDisplays(ctx context.Context, cfg model.Config) ([]model.DisplayInfo, string, error) {
	backendsByName := make(map[string]display.Backend, len(r.DisplayBackends))
	for _, backend := range r.DisplayBackends {
		backendsByName[backend.Name()] = backend
	}

	for _, backendName := range cfg.DisplayBackendPriority {
		backend, ok := backendsByName[backendName]
		if !ok {
			continue
		}

		displays, err := backend.Detect(ctx)
		if err == nil && len(displays) > 0 {
			return displays, backendName, nil
		}

		if r.Verbose && err != nil {
			fmt.Fprintf(r.Output, "backend %s failed: %v\n", backendName, err)
		}
	}

	return nil, "", fmt.Errorf("could not determine display information")
}

func (r Runner) buildPlan(cfg model.Config, resolved model.ResolvedSettings) (model.Plan, error) {
	adapterByName := make(map[string]target.Adapter, len(r.Targets))
	for _, adapter := range r.Targets {
		adapterByName[adapter.Name()] = adapter
	}

	plan := model.Plan{
		DryRun:   cfg.DryRun,
		Resolved: resolved,
	}

	for _, targetName := range cfg.TargetNames {
		adapter, ok := adapterByName[targetName]
		if !ok {
			return model.Plan{}, fmt.Errorf("unknown target %q", targetName)
		}
		if !adapter.Enabled(cfg) {
			continue
		}

		actions, err := adapter.BuildActions(cfg, resolved)
		if err != nil {
			return model.Plan{}, err
		}
		plan.Actions = append(plan.Actions, actions...)
	}

	return plan, nil
}

func (r Runner) applyPlan(ctx context.Context, plan model.Plan) error {
	adapterByName := make(map[string]target.Adapter, len(r.Targets))
	for _, adapter := range r.Targets {
		adapterByName[adapter.Name()] = adapter
	}

	for _, action := range plan.Actions {
		adapter, ok := adapterByName[action.Target]
		if !ok {
			return fmt.Errorf("no adapter registered for target %q", action.Target)
		}

		if err := adapter.Apply(ctx, action, r.CommandRunner); err != nil {
			return err
		}
	}

	return nil
}

func (r Runner) printPlan(plan model.Plan) {
	displayInfo := plan.Resolved.Display
	selectedProfile := plan.Resolved.Profile

	fmt.Fprintf(r.Output, "Source backend: %s\n", plan.Resolved.SourceBackend)
	fmt.Fprintf(
		r.Output,
		"Selected display: %s (%dx%d px, %dx%d mm, PPI=%.2f)\n",
		displayInfo.Name,
		displayInfo.WidthPx,
		displayInfo.HeightPx,
		displayInfo.WidthMM,
		displayInfo.HeightMM,
		displayInfo.PPI,
	)
	fmt.Fprintf(
		r.Output,
		"Selected profile: %s (text_scaling=%.2f, ui=%d, document=%d, monospace=%d, titlebar=%d)\n",
		selectedProfile.Name,
		selectedProfile.TextScaling,
		selectedProfile.UIFontSize,
		selectedProfile.DocumentFontSize,
		selectedProfile.MonospaceFontSize,
		selectedProfile.TitlebarFontSize,
	)

	if plan.DryRun {
		fmt.Fprintln(r.Output, "Mode: dry-run")
	} else {
		fmt.Fprintln(r.Output, "Mode: apply")
	}

	if len(plan.Actions) == 0 {
		fmt.Fprintln(r.Output, "No actions to run.")
		return
	}

	fmt.Fprintln(r.Output, "Actions:")
	for _, action := range plan.Actions {
		fmt.Fprintf(r.Output, "  - [%s] %s\n", action.Target, action.Description)
		fmt.Fprintf(r.Output, "    %s\n", shellJoin(action.Command))
	}
}

func shellJoin(parts []string) string {
	quoted := make([]string, len(parts))
	for index, part := range parts {
		if part == "" {
			quoted[index] = `""`
			continue
		}
		if strings.ContainsAny(part, " \t\n\"'") {
			quoted[index] = fmt.Sprintf("%q", part)
			continue
		}
		quoted[index] = part
	}
	return strings.Join(quoted, " ")
}
