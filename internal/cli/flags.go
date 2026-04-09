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
	"strings"

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/model"
)

var ErrHelpRequested = errors.New("help requested")

type Options struct {
	ConfigPath       string
	DryRunOverride   *bool
	PreferredDisplay string
	Verbose          bool
}

func Parse(args []string) (Options, error) {
	var options Options
	var dryRun bool
	var apply bool
	var help bool

	flagSet := flag.NewFlagSet("auto-font-by-ppi", flag.ContinueOnError)
	flagSet.SetOutput(new(strings.Builder))
	flagSet.StringVar(&options.ConfigPath, "config", "", "load configuration from the specified path")
	flagSet.BoolVar(&dryRun, "dry-run", false, "show what would change without applying settings")
	flagSet.BoolVar(&apply, "apply", false, "apply settings")
	flagSet.StringVar(&options.PreferredDisplay, "display", "", "select a display by connector name")
	flagSet.BoolVar(&options.Verbose, "verbose", false, "print more diagnostic information")
	flagSet.BoolVar(&help, "help", false, "show help")
	flagSet.BoolVar(&help, "h", false, "show help")

	if err := flagSet.Parse(args); err != nil {
		return Options{}, err
	}

	if help {
		return Options{}, ErrHelpRequested
	}

	if dryRun && apply {
		return Options{}, fmt.Errorf("--dry-run and --apply cannot be used together")
	}

	if dryRun {
		value := true
		options.DryRunOverride = &value
	}
	if apply {
		value := false
		options.DryRunOverride = &value
	}

	return options, nil
}

func Apply(cfg *model.Config, options Options) {
	if options.DryRunOverride != nil {
		cfg.DryRun = *options.DryRunOverride
	}

	if options.PreferredDisplay != "" {
		cfg.PreferredDisplay = options.PreferredDisplay
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

  --apply
      Actually apply settings.

  --display NAME
      Use the specified display name instead of auto-detecting.

  --verbose
      Print more diagnostic information.

  --help, -h
      Show this help message and exit.
`, defaultConfigPath)
}
