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
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply environment configurations",
	Long: `This command applies specific environment configurations based on your settings. 

It can set global and environment-specific variables, either temporarily for the current session 
or persistently by writing them to a '.envrc' file in your home directory. When using the --persist flag 
to save configurations persistently, remember to add 'source $HOME/.envrc' to your .bashrc file to ensure 
these settings are loaded in each new shell session. Use the --omit-global flag to exclude global variables. 

This tool is particularly useful for managing different development or deployment environments, allowing you 
to switch contexts easily and safely. It simplifies the process of managing and switching between various 
configured environments by automating the setup of environment variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		apply()
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.PersistentFlags().BoolP("omit-global", "o", false, "Omit global environment setting")
	applyCmd.PersistentFlags().BoolP("persist", "p", false, "Persist the environment on .envrc")

	viper.BindPFlag("persist", applyCmd.PersistentFlags().Lookup("persist"))
	viper.BindPFlag("omit-global", applyCmd.PersistentFlags().Lookup("omit-global"))

}

func apply() {

	envVars, err := getConfigEnvs()
	if err != nil {
		panic(err)
	}

	if viper.GetBool("persist") {
		err := persistEnvironment(envVars)
		if err != nil {
			panic(err)
		}
		printExports(envVars)

	} else {
		printExports(envVars)
	}

}

func printExports(envVars map[string]string) {

	for key, value := range envVars {
		logPrintf("export %s='%s'\n", strings.ToUpper(key), value)
	}
}

func persistEnvironment(envVars map[string]string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home dir %s", err)
	}

	envrcPath := filepath.Join(homeDir, ".envrc")

	file, err := os.Create(envrcPath)
	if err != nil {
		return fmt.Errorf("error opening file .envrc  %s", err)
	}
	defer file.Close()

	for key, value := range envVars {
		file.WriteString(fmt.Sprintf("export %s='%s'\n", strings.ToUpper(key), value))
	}

	logPrintf(`Environment variables set in '.envrc'. 
Beam them into your current universe with 'source ~/.bashrc' 
or by launching a new terminal session. 
It's like a quick warp jump to your ideal coding environment. Ready for the journey?`)
	return nil
}
