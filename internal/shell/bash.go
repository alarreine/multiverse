// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

// Package shell generates the bash that the shell integration evaluates.
package shell

import (
	"fmt"
	"regexp"
	"strings"
)

// Name of the marker the integration exports, so the binary can tell whether it
// was reached through the shell function or called directly.
const EnvIntegration = "MULTIVERSE_SHELL_INTEGRATION"

var nameRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ValidName reports whether name is a usable environment variable name. It
// matters more than it looks: values are quoted before being emitted, but names
// are not, so an unchecked name is a shell injection straight into eval.
func ValidName(name string) bool { return nameRE.MatchString(name) }

// Quote renders s as a bash single-quoted string. Inside single quotes bash
// interprets nothing at all, so ending the quote is the only escape needed.
func Quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Script accumulates bash statements. Errors are collected instead of returned
// per call so that a caller can build the whole script and check once; nothing
// should reach stdout unless Err returns nil.
type Script struct {
	b   strings.Builder
	err error
}

// Export sets a variable in the calling shell.
func (s *Script) Export(name, value string) {
	if !s.checkName(name) {
		return
	}
	fmt.Fprintf(&s.b, "export %s=%s\n", name, Quote(value))
}

// Unset removes variables from the calling shell.
func (s *Script) Unset(names ...string) {
	for _, name := range names {
		if !s.checkName(name) {
			return
		}
		fmt.Fprintf(&s.b, "unset %s\n", name)
	}
}

func (s *Script) checkName(name string) bool {
	if !ValidName(name) {
		if s.err == nil {
			s.err = fmt.Errorf("%q is not a valid environment variable name", name)
		}
		return false
	}
	return true
}

// Err reports the first invalid name seen, if any.
func (s *Script) Err() error { return s.err }

// Empty reports whether the script would do nothing.
func (s *Script) Empty() bool { return s.b.Len() == 0 }

func (s *Script) String() string { return s.b.String() }

const initTemplate = `# multiverse shell integration for bash.
# Install by adding this line to ~/.bashrc:
#
#     eval "$(multiverse init bash)"
#
# 'use' and 'off' have to run in this shell, so they go through eval; every
# other subcommand is passed straight to the binary.

export %s=bash

multiverse() {
  case "$1" in
    use|off)
      local __mv_out __mv_rc
      __mv_out="$(%s export --shell bash "$@")"
      __mv_rc=$?
      # Only evaluate a successful run: a failed one must leave the shell alone.
      if [ $__mv_rc -ne 0 ]; then
        return $__mv_rc
      fi
      eval "$__mv_out"
      ;;
    *)
      %s "$@"
      ;;
  esac
}
`

// InitScript renders the bash integration. bin is the absolute path of this
// binary, resolved at install time so the function keeps working whether or not
// multiverse is on PATH.
func InitScript(bin string) string {
	quoted := Quote(bin)
	return fmt.Sprintf(initTemplate, EnvIntegration, quoted, quoted)
}
