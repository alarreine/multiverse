// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/alarreine/multiverse/internal/config"
	"github.com/alarreine/multiverse/internal/shell"
	"github.com/alarreine/multiverse/internal/state"
	"github.com/spf13/cobra"
)

// reservedPrefix guards the bookkeeping variables: letting a universe set them
// would corrupt the record of what to undo.
const reservedPrefix = "MULTIVERSE_"

var (
	exportShell string
	omitGlobal  bool
	envAlias    string
)

var exportCmd = &cobra.Command{
	Use:   "export <use|off> [universe]",
	Short: "Emit the shell code for a universe change",
	Long: `Print the bash that applies or reverts a universe, without running it.

This is what the shell integration evaluates, and it is also the way to use
multiverse without installing the integration at all:

    eval "$(multiverse export --shell bash use prod)"

Only stdout carries shell code; everything else goes to stderr, and a failed run
prints nothing at all, so a broken config can never be evaluated.`,
	Hidden: true,
	RunE:   runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVar(&exportShell, "shell", "bash", "shell dialect to emit")
	exportCmd.Flags().BoolVarP(&omitGlobal, "omit-global", "o", false, "Omit the global environment settings")
	exportCmd.Flags().StringVarP(&envAlias, "env", "e", "", "Universe to enter (alias for the positional argument)")
}

func runExport(cmd *cobra.Command, args []string) error {
	if exportShell != "bash" {
		return fmt.Errorf("unsupported shell %q: only bash is supported", exportShell)
	}
	if len(args) == 0 {
		return fmt.Errorf("export needs an action: use or off")
	}

	prev, active, err := state.FromEnv(os.Getenv)
	if err != nil {
		return err
	}

	sc := &shell.Script{}
	revert(sc, prev)

	var summary string
	switch action := args[0]; action {
	case "off":
		if !active {
			warnf("universe: none (nothing to leave)")
			return nil
		}
		sc.Unset(state.EnvState, state.EnvUniverse)
		summary = fmt.Sprintf("universe: none (left %s)", prev.Universe)
	case "use":
		name, err := universeName(args[1:])
		if err != nil {
			return err
		}
		count, err := enter(sc, name, prev)
		if err != nil {
			return err
		}
		summary = fmt.Sprintf("universe: %s (%d vars)", name, count)
	default:
		return fmt.Errorf("unknown action %q: expected use or off", action)
	}

	// Report success only once the whole script is known to be sound: nothing
	// may reach stdout, or claim to have happened, after an error.
	if err := sc.Err(); err != nil {
		return err
	}
	fmt.Fprint(os.Stdout, sc.String())
	warnf("%s", summary)
	return nil
}

// revert undoes the active universe: variables that existed before go back to
// what they were, and the ones multiverse invented are removed.
func revert(sc *shell.Script, prev state.State) {
	for _, key := range prev.Keys() {
		if v := prev.Vars[key]; v.Had {
			sc.Export(key, v.Prev)
		} else {
			sc.Unset(key)
		}
	}
}

// enter appends the universe to the script and reports how many variables it
// sets.
func enter(sc *shell.Script, name string, prev state.State) (int, error) {
	cfg, err := loadConfig()
	if err != nil {
		return 0, err
	}

	raw, err := cfg.Resolve(name, omitGlobal)
	if err != nil {
		return 0, err
	}
	for _, key := range config.SortedKeys(raw) {
		if strings.HasPrefix(key, reservedPrefix) {
			return 0, fmt.Errorf("universe %q sets %s, but names starting with %s are reserved", name, key, reservedPrefix)
		}
	}

	// Interpolate against the environment as it was before the active universe,
	// not against the live one. That is what makes `use prod` twice a no-op
	// instead of growing PATH: "${TOOLS}/bin:$PATH" a segment at a time.
	pristine := pristineLookup(prev)
	envs, err := config.Interpolate(raw, pristine)
	if err != nil {
		return 0, err
	}

	next := state.State{Universe: name, Vars: make(map[string]state.Var, len(envs))}
	for _, key := range config.SortedKeys(envs) {
		before, had := pristine(key)
		next.Vars[key] = state.Var{Had: had, Prev: before}
		sc.Export(key, envs[key])
	}

	blob, err := state.Encode(next)
	if err != nil {
		return 0, err
	}
	sc.Export(state.EnvState, blob)
	sc.Export(state.EnvUniverse, name)

	return len(envs), nil
}

// pristineLookup reports the environment as it was before the active universe
// was applied.
func pristineLookup(prev state.State) config.Lookup {
	return func(name string) (string, bool) {
		if v, ok := prev.Vars[name]; ok {
			if !v.Had {
				return "", false
			}
			return v.Prev, true
		}
		return os.LookupEnv(name)
	}
}

// universeName takes the universe from the positional argument, falling back to
// the older --env flag.
func universeName(args []string) (string, error) {
	switch {
	case len(args) > 1:
		return "", fmt.Errorf("expected one universe, got %d: %s", len(args), strings.Join(args, " "))
	case len(args) == 1:
		return args[0], nil
	case envAlias != "":
		return envAlias, nil
	default:
		return "", fmt.Errorf("which universe? try: multiverse list")
	}
}
