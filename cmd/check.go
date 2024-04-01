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
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Verify environment variable settings",
	Long: `Embark on a verification voyage with the check command, your trusty tool for navigating the intricate 
multiverse of environment variables. It's like having a cosmic map that shows if you're in sync with 
the universe (your configuration file)! 

For each variable, it conjures a mystical symbol: 
a harmonious check mark (✓) if you're aligned with the cosmic order (variable matches), 
a perplexing 'X' (✗) if you've ventured into unknown territories (variable not set), and 
a mysterious '!=' if you find yourself in a parallel dimension (variable set but different). 

Use this command to ensure you're not astray in the nebulous realms of your development or deployment 
galaxies, keeping your interstellar journey smooth and predictable.`,
	Run: func(cmd *cobra.Command, args []string) {
		check()
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func check() {

	envVars, err := getConfigEnvs()
	if err != nil {
		panic(err)
	}

	for key, expectedValue := range envVars {
		key = strings.ToUpper(key)
		if value, exists := os.LookupEnv(key); !exists {
			logPrintf("%s: ✗\n", key)
		} else if value != expectedValue {
			logPrintf("%s: !=\n", key)
		} else {
			logPrintf("%s: ✓\n", key)
		}
	}

}
