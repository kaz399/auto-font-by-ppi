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

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/app"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/cli"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/config"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/display"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/execx"
	"github.com/kazuhiro-yabe/auto-font-by-ppi/internal/target"
)

func main() {
	options, err := cli.Parse(os.Args[1:])
	if err != nil {
		if errors.Is(err, cli.ErrHelpRequested) {
			fmt.Fprint(os.Stdout, cli.Usage(config.DefaultConfigPath()))
			return
		}
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load(options.ConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	cli.Apply(&cfg, options)

	commandRunner := execx.CommandRunner{}
	runner := app.Runner{
		DisplayBackends: []display.Backend{
			display.GNOMEWaylandBackend{Runner: commandRunner},
			display.XRandrBackend{Runner: commandRunner},
		},
		Targets: []target.Adapter{
			target.GNOMEAdapter{},
			target.KittyAdapter{},
		},
		CommandRunner: commandRunner,
		Output:        os.Stdout,
		Verbose:       options.Verbose,
	}

	if err := runner.Run(context.Background(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
