// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"

	"github.com/alarreine/multiverse/internal/config"
	"github.com/alarreine/multiverse/internal/state"
	"github.com/spf13/cobra"
)

var statusVerbose bool

var statusCmd = &cobra.Command{
	Use:   "status [universe]",
	Short: "Report which universe is active",
	Long: `Report which universe you are currently in, if any.

Add --verbose to also compare your real environment against what the universe
says it should be, variable by variable:

  ✓   aligned with the cosmic order (the value matches)
  ✗   venturing into unknown territories (the variable is not set)
  !=  parallel dimensions detected (set, but to something else)

With no argument the comparison uses the active universe; pass a name to check
against a different one.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVarP(&statusVerbose, "verbose", "v", false, "Compare the shell against the universe, variable by variable")
	statusCmd.Flags().BoolVarP(&omitGlobal, "omit-global", "o", false, "Omit the global environment settings")
}

func runStatus(cmd *cobra.Command, args []string) error {
	prev, active, err := state.FromEnv(os.Getenv)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()

	if active {
		fmt.Fprintf(out, "universe: %s\n", prev.Universe)
		fmt.Fprintf(out, "vars:     %d\n", len(prev.Vars))
	} else {
		fmt.Fprintln(out, "universe: none")
	}
	if !statusVerbose {
		return nil
	}

	target := prev.Universe
	if len(args) == 1 {
		target = args[0]
	}
	if target == "" {
		return fmt.Errorf("no universe is active; name one to compare against: multiverse status -v <universe>")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	raw, err := cfg.Resolve(target, omitGlobal)
	if err != nil {
		return err
	}
	// Compare against the values a clean shell would get, so an already applied
	// universe does not look different from itself.
	envs, err := config.Interpolate(raw, pristineLookup(prev))
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "config:   %s\n", cfg.Path())
	fmt.Fprintf(out, "\nagainst %s:\n", target)

	width := 0
	keys := config.SortedKeys(envs)
	for _, key := range keys {
		if len(key) > width {
			width = len(key)
		}
	}
	for _, key := range keys {
		actual, exists := os.LookupEnv(key)
		mark := "✓"
		switch {
		case !exists:
			mark = "✗"
		case actual != envs[key]:
			mark = "!="
		}
		fmt.Fprintf(out, "  %-*s  %s\n", width, key, mark)
	}
	return nil
}
