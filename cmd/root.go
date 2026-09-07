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

	"github.com/alarreine/multiverse/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "multiverse",
	Short: "Manage environment configurations effortlessly",
	Long: `Welcome to Multiverse, the command-line tool that's your portal to effortlessly
navigating through the vast landscapes of environment configurations. Think of it as your
trusty time machine, capable of zipping you between parallel worlds of development, testing,
and production.

  multiverse use prod     enter a universe in this very shell
  multiverse off          leave it, putting back what was there before
  multiverse status       where in the multiverse am I?
  multiverse list         which universes exist?

'use' and 'off' reach into your current shell, and no program can do that to its
parent on its own. The trick is a small bash function that evaluates what this
binary prints. Install it once:

    eval "$(multiverse init bash)"      # in your ~/.bashrc

Ready your gear, set your coordinates, and enjoy the multiverse hopping!`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "multiverse: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default: $MULTIVERSE_CONFIG, ~/.config/multiverse/config.yaml, ~/.multiverse.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false,
		"Mute the chatter, enjoy the quiet cosmos")
}

// warnf reports progress to stderr. Nothing but evaluable bash may go to
// stdout: the shell integration feeds stdout straight to eval.
func warnf(format string, a ...any) {
	if quiet {
		return
	}
	fmt.Fprintf(os.Stderr, format+"\n", a...)
}

// loadConfig reads the config file. It is called from the commands that need
// one, so that `init` and `--help` still work on a machine with no config yet.
func loadConfig() (*config.Config, error) {
	return config.Load(cfgFile)
}
