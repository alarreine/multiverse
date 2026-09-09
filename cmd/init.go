// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"

	"github.com/alarreine/multiverse/internal/shell"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init bash",
	Short: "Print the shell integration to install",
	Long: `Print the bash function that lets 'use' and 'off' act on your current shell.

Install it by adding this to your ~/.bashrc:

    eval "$(multiverse init bash)"

The generated function hard-codes the absolute path of this binary, so it keeps
working whether or not multiverse is on your PATH. Run 'init' again if you move
the binary.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != "bash" {
			return fmt.Errorf("unsupported shell %q: only bash is supported", args[0])
		}
		bin, err := os.Executable()
		if err != nil {
			return fmt.Errorf("locating the multiverse binary: %w", err)
		}
		fmt.Fprint(cmd.OutOrStdout(), shell.InitScript(bin))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
