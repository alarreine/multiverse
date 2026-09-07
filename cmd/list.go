/*
Copyright © 2024 Agustin LARREINEGABE

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/alarreine/multiverse/internal/state"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Tour the Multiverse: See All Available Realms",
	Long: `Embark on a galactic tour with the list command! It's like having a telescope that peers into the multiverse,
revealing all the mystical environments hidden within your config file.
Each environment is a unique realm, waiting for your adventurous spirit.
The realm you currently inhabit, if any, is marked with a *.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()

		if len(cfg.Environments) == 0 {
			fmt.Fprintf(out, "No universes declared in %s\n", cfg.Path())
			return nil
		}

		active := os.Getenv(state.EnvUniverse)
		fmt.Fprintf(out, "Available universes (%s):\n", cfg.Path())
		for _, env := range cfg.Environments {
			marker := " "
			if env.Name == active {
				marker = "*"
			}
			line := fmt.Sprintf(" %s %s", marker, env.Name)
			if env.Extends != "" {
				line += fmt.Sprintf("  (extends %s)", env.Extends)
			}
			fmt.Fprintln(out, line)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
