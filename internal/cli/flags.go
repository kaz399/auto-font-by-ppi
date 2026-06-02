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
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
)

var ErrHelpRequested = errors.New("help requested")

type Options struct {
	ConfigPath               string
	DryRunOverride           *bool
	PreferredDisplay         string
	Auto                     bool
	CurrentDisplayOverride   *model.DisplayOverride
	DisplayDiagonalOverrides map[string]float64
	Verbose                  bool
}

func Parse(args []string) (Options, error) {
	var options Options
	var dryRun bool
	var help bool

	flagSet := flag.NewFlagSet("auto-font-by-ppi", flag.ContinueOnError)
	flagSet.SetOutput(new(strings.Builder))
	flagSet.StringVar(&options.ConfigPath, "config", "", "load configuration from the specified path")
	flagSet.BoolVar(&dryRun, "dry-run", false, "show what would change without applying settings")
	flagSet.StringVar(&options.PreferredDisplay, "display", "", "select a display by connector name")
	flagSet.BoolVar(&options.Auto, "auto", false, "ignore preferred display config and auto-detect instead")
	flagSet.Var(currentDisplayOverrideFlag{target: &options.CurrentDisplayOverride, mode: "inch"}, "inch", "treat the selected display as having the specified diagonal size in inches")
	flagSet.Var(currentDisplayOverrideFlag{target: &options.CurrentDisplayOverride, mode: "ppi"}, "ppi", "treat the selected display as having the specified PPI")
	flagSet.Var(displayDiagonalOverridesFlag{target: &options.DisplayDiagonalOverrides}, "display-diagonal", "override a display diagonal in NAME=INCHES form")
	flagSet.BoolVar(&options.Verbose, "verbose", false, "print more diagnostic information")
	flagSet.BoolVar(&help, "help", false, "show help")
	flagSet.BoolVar(&help, "h", false, "show help")

	if err := flagSet.Parse(args); err != nil {
		return Options{}, err
	}

	if help {
		return Options{}, ErrHelpRequested
	}

	if options.Auto && options.PreferredDisplay != "" {
		return Options{}, fmt.Errorf("--auto and --display cannot be used together")
	}

	if dryRun {
		value := true
		options.DryRunOverride = &value
	}

	return options, nil
}

func Apply(cfg *model.Config, options Options) {
	if options.DryRunOverride != nil {
		cfg.DryRun = *options.DryRunOverride
	}

	if options.Auto {
		cfg.PreferredDisplay = ""
	} else if options.PreferredDisplay != "" {
		cfg.PreferredDisplay = options.PreferredDisplay
	}
	if options.CurrentDisplayOverride != nil {
		cfg.CurrentDisplayOverride = &model.DisplayOverride{
			Mode:  options.CurrentDisplayOverride.Mode,
			Value: options.CurrentDisplayOverride.Value,
		}
	}

	if len(options.DisplayDiagonalOverrides) > 0 {
		if cfg.Display.DiagonalOverrides == nil {
			cfg.Display.DiagonalOverrides = make(map[string]float64, len(options.DisplayDiagonalOverrides))
		}
		for name, diagonal := range options.DisplayDiagonalOverrides {
			cfg.Display.DiagonalOverrides[name] = diagonal
		}
	}
}

func Usage(defaultConfigPath string) string {
	return fmt.Sprintf(`Usage:
  auto-font-by-ppi [options]

Purpose:
  Detect the target display, calculate its PPI, select a profile,
  and update the configured targets.

Options:
  --config PATH
      Load configuration from PATH.
      Default: %s

  --dry-run
      Show what would be changed, but do not apply settings.

  --display NAME
      Use the specified display name instead of auto-detecting.

  --auto
      Ignore the preferred display in the config file and auto-detect instead.
      This option cannot be specified together with --display.

  --inch INCHES
      Treat the selected display as having the specified diagonal size.
      This overrides config-based display size settings.

  --ppi VALUE
      Treat the selected display as having the specified PPI.
      This overrides config-based display size settings.

  --display-diagonal NAME=INCHES
      Override a display's physical diagonal size.
      This option may be specified multiple times.

  --verbose
      Print more diagnostic information.

  --help, -h
      Show this help message and exit.
`, defaultConfigPath)
}

type displayDiagonalOverridesFlag struct {
	target *map[string]float64
}

func (f displayDiagonalOverridesFlag) String() string {
	if f.target == nil || len(*f.target) == 0 {
		return ""
	}

	parts := make([]string, 0, len(*f.target))
	for name, diagonal := range *f.target {
		parts = append(parts, fmt.Sprintf("%s=%g", name, diagonal))
	}
	return strings.Join(parts, ",")
}

func (f displayDiagonalOverridesFlag) Set(value string) error {
	name, diagonal, err := parseDisplayDiagonalOverride(value)
	if err != nil {
		return err
	}

	if *f.target == nil {
		*f.target = make(map[string]float64)
	}
	(*f.target)[name] = diagonal
	return nil
}

func parseDisplayDiagonalOverride(value string) (string, float64, error) {
	name, diagonalText, ok := strings.Cut(value, "=")
	name = strings.TrimSpace(name)
	diagonalText = strings.TrimSpace(diagonalText)
	if !ok || name == "" || diagonalText == "" {
		return "", 0, fmt.Errorf("invalid display diagonal override %q", value)
	}

	diagonal, err := strconv.ParseFloat(diagonalText, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid diagonal value in override %q", value)
	}
	if diagonal <= 0 {
		return "", 0, fmt.Errorf("invalid diagonal value in override %q", value)
	}

	return name, diagonal, nil
}

type currentDisplayOverrideFlag struct {
	target **model.DisplayOverride
	mode   string
}

func (f currentDisplayOverrideFlag) String() string {
	if f.target == nil || *f.target == nil || (*f.target).Mode != f.mode {
		return ""
	}

	return strconv.FormatFloat((*f.target).Value, 'f', -1, 64)
}

func (f currentDisplayOverrideFlag) Set(value string) error {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || parsed <= 0 {
		return fmt.Errorf("invalid %s value %q", f.mode, value)
	}

	*f.target = &model.DisplayOverride{
		Mode:  f.mode,
		Value: parsed,
	}
	return nil
}
