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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Environment struct {
	Name string            `yaml:"name"`
	Envs map[string]string `yaml:"envs"`
}

type Config struct {
	Global       map[string]string `yaml:"global"`
	Environments []Environment     `yaml:"environments"`
}

var cfgFile string
var (
	envCfg         Config
	loggingEnabled bool
)

var rootCmd = &cobra.Command{
	Use:   "multiverse",
	Short: "Manage environment configurations effortlessly",
	Long: `Welcome to Multiverse, the command-line tool that's your portal to effortlessly 
navigating through the vast landscapes of environment configurations. Think of it as your 
trusty time machine, capable of zipping you between parallel worlds of development, testing, and production. 

But beware, traveler! Just like in any time-travel adventure, there are rules: changes you make are 
like altering timelines – they affect the future (your child processes) but can't rewrite the past (your parent shell). 

If you choose the path of persistence with the --persist flag, don't forget to inscribe 
the magical incantation 'source $HOME/.envrc' in your sacred .bashrc scrolls. This ensures your 
environment settings resonate across the echoes of every new terminal portal you open.
Ready your gear, set your coordinates, and enjoy the multiverse hopping!

Remember, with great power comes great responsibility - use your newfound abilities wisely!`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.multiverse.yaml)")
	rootCmd.PersistentFlags().StringP("env", "e", "", "Set the environment of this envName")
	rootCmd.PersistentFlags().BoolVarP(&loggingEnabled, "quiet", "q", false, "Mute the chatter, enjoy the quiet cosmos")

	viper.BindPFlag("env", rootCmd.PersistentFlags().Lookup("env"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".multiverse" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".multiverse")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logPrintf("#using config file: %s\n", viper.ConfigFileUsed())
	} else {
		panic("error reading config file: %s")
	}

	if err := viper.Unmarshal(&envCfg); err != nil {
		panic("error reading config file: %s")
	}
}

func getConfigEnvs() (map[string]string, error) {
	envVars := make(map[string]string)
	found := false

	if !viper.GetBool("omit-global") {
		envVars = envCfg.Global
	}

	for _, universe := range envCfg.Environments {
		if universe.Name == viper.GetString("env") {
			for key, value := range universe.Envs {
				envVars[key] = value
			}
			found = true
			break
		}
	}

	if !found {
		return envVars, fmt.Errorf("#environment %s not found in config file", viper.GetString("env"))
	}
	return envVars, nil
}

func logPrintf(format string, a ...any) {
	if !loggingEnabled {
		fmt.Printf(format, a...)
	}
}
