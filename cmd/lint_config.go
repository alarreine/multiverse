// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/alarreine/multiverse/internal/config"
	"github.com/alarreine/multiverse/internal/lint"
	"github.com/alarreine/multiverse/internal/state"
	"github.com/spf13/cobra"
)

var lintStrict bool

var lintConfigCmd = &cobra.Command{
	Use:   "lint-config [file]",
	Short: "Check a config file for problems",
	Long: `Check every universe the way 'use' would, without touching your environment.

Errors mean a universe cannot be used at all: a cycle in extends, a cycle in
interpolation, a variable name bash cannot take, or a variable under the
reserved MULTIVERSE_ prefix. Warnings mean it works but probably does not do
what you meant: a value referring to a name nothing defines expands to the
empty string in silence, which is how a typo hides.

With no argument it checks the config it would otherwise use; pass a path to
check a candidate before installing it. It exits non-zero when there are
errors, so it works from a pre-commit hook or from CI, and --strict makes
warnings count too.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLintConfig,
}

func init() {
	rootCmd.AddCommand(lintConfigCmd)
	lintConfigCmd.Flags().BoolVar(&lintStrict, "strict", false, "Treat warnings as errors")
}

func runLintConfig(cmd *cobra.Command, args []string) error {
	path := cfgFile
	if len(args) == 1 {
		path = args[0]
	}
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	// Lint against the environment as it was before any active universe, for the
	// same reason `status --verbose` does: otherwise what is already applied
	// would mask the references that are really undefined.
	prev, _, err := state.FromEnv(os.Getenv)
	if err != nil {
		return err
	}

	findings := lint.Config(cfg, pristineLookup(prev))
	var errors, warnings int
	for _, f := range findings {
		if f.Fatal {
			errors++
		} else {
			warnings++
		}
	}

	if !quiet {
		report(cmd.OutOrStdout(), cfg, findings, errors, warnings)
	}
	switch {
	case errors > 0:
		return fmt.Errorf("%s has %s", cfg.Path(), plural(errors, "error"))
	case warnings > 0 && lintStrict:
		return fmt.Errorf("%s has %s and --strict is set", cfg.Path(), plural(warnings, "warning"))
	}
	return nil
}

// report prints one finding per line, each carrying its universe, so the output
// survives being piped through grep.
func report(w io.Writer, cfg *config.Config, findings []lint.Finding, errors, warnings int) {
	fmt.Fprintf(w, "config: %s\n", cfg.Path())
	if len(findings) > 0 {
		fmt.Fprintln(w)
	}
	for _, f := range findings {
		severity := "warning"
		if f.Fatal {
			severity = "error"
		}
		fmt.Fprintf(w, "%s: %s: %s\n", f.Universe, severity, f.Message)
	}

	fmt.Fprintln(w)
	universes := plural(len(cfg.Environments), "universe")
	if len(findings) == 0 {
		fmt.Fprintf(w, "%s, no problems found\n", universes)
		return
	}
	fmt.Fprintf(w, "%s, %s, %s\n", universes, plural(errors, "error"), plural(warnings, "warning"))
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}
