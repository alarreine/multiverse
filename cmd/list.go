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
	Use:   "list [pattern]",
	Short: "Tour the Multiverse: See All Available Realms",
	Long: `Embark on a galactic tour with the list command! It's like having a telescope that peers into the multiverse,
revealing all the mystical environments hidden within your config file.
Each environment is a unique realm, waiting for your adventurous spirit.
The realm you currently inhabit, if any, is marked with a *.

Give it a pattern to narrow the tour down. Without wildcards it is a
case-insensitive substring of the universe name:

    multiverse list bonita

With *, ? or [ it is a glob over the whole name, where * does not cross a '/',
so you can pick one level of the hierarchy at a time:

    multiverse list 'bpm-prod/*/jenkins'    the jenkins of one account
    multiverse list '*/*/production'        every production namespace

Nothing matching is an error, so this works in a script the way grep does.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var pattern string
		if len(args) == 1 {
			pattern = args[0]
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()

		if len(cfg.Environments) == 0 {
			fmt.Fprintf(out, "No universes declared in %s\n", cfg.Path())
			return nil
		}

		matches, err := cfg.Match(pattern)
		if err != nil {
			// path.ErrBadPattern on its own says only "syntax error in pattern".
			return fmt.Errorf("bad pattern %q: %w", pattern, err)
		}
		if len(matches) == 0 {
			return fmt.Errorf("no universe matches %q", pattern)
		}

		if pattern == "" {
			fmt.Fprintf(out, "Available universes (%s):\n", cfg.Path())
		} else {
			fmt.Fprintf(out, "Universes matching %q (%s):\n", pattern, cfg.Path())
		}

		active := os.Getenv(state.EnvUniverse)
		for _, env := range matches {
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

		// Only when filtering: an unfiltered list has to stay byte-identical to
		// what it has always printed.
		if pattern != "" {
			fmt.Fprintf(out, "\n%d of %s\n", len(matches), plural(len(cfg.Environments), "universe"))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
