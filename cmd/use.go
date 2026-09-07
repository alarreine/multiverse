// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"

	"github.com/alarreine/multiverse/internal/shell"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <universe>",
	Short: "Enter a universe in the current shell",
	Long: `Apply a universe's variables to the shell you are typing in.

This command is handled by the bash function that 'multiverse init bash'
installs; the binary alone cannot change the environment of the shell that
launched it. If you see this text as an error, the integration is missing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := universeName(args); err != nil {
			return err
		}
		return errNoIntegration("use")
	},
}

var offCmd = &cobra.Command{
	Use:   "off",
	Short: "Leave the current universe",
	Long: `Undo the active universe: variables that existed beforehand go back to their
previous values, and the ones the universe introduced are unset.

Like 'use', this is handled by the bash function from 'multiverse init bash'.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errNoIntegration("off")
	},
}

func init() {
	rootCmd.AddCommand(useCmd, offCmd)
	useCmd.Flags().BoolVarP(&omitGlobal, "omit-global", "o", false, "Omit the global environment settings")
	useCmd.Flags().StringVarP(&envAlias, "env", "e", "", "Universe to enter (alias for the positional argument)")
}

// errNoIntegration explains why reaching this binary directly cannot work. The
// two cases deserve different advice, so tell them apart by the marker the
// integration exports.
func errNoIntegration(action string) error {
	if os.Getenv(shell.EnvIntegration) != "" {
		return fmt.Errorf(`%q was run as a program, bypassing the multiverse shell function.
  Drop the 'command' prefix, or evaluate it yourself:

    eval "$(multiverse export --shell bash %s ...)"`, action, action)
	}
	return fmt.Errorf(`%q has to run inside your shell, and the bash integration is not installed.
  Add this to your ~/.bashrc, then open a new terminal:

    eval "$(multiverse init bash)"

  For a one-off, without installing anything:

    eval "$(multiverse export --shell bash %s ...)"`, action, action)
}
