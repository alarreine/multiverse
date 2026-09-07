// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"strings"
)

// Lookup resolves a variable name that a value refers to but the universe does
// not define, normally the surrounding process environment.
type Lookup func(name string) (string, bool)

// Interpolate expands $VAR and ${VAR} in every value. A reference is resolved
// against the other keys of vars first and against lookup second, so a universe
// can compose its own variables. The exception is a variable that refers to
// itself, PATH: "${TOOLS_HOME}/bin:$PATH", which always means the incoming
// value from lookup: that idiom is the whole point of interpolating, and
// resolving it against vars would be a cycle. Use $$ for a literal dollar sign.
// An unknown name expands to the empty string, like a shell would.
//
// Resolution is recursive and memoised rather than a single pass over the map:
// Go map order is random, so expanding A: "$B" in place would give a different
// answer depending on whether B happened to come first.
func Interpolate(vars map[string]string, lookup Lookup) (map[string]string, error) {
	in := &interpolator{
		raw:    vars,
		lookup: lookup,
		done:   make(map[string]string, len(vars)),
		open:   make(map[string]bool),
	}

	out := make(map[string]string, len(vars))
	for _, k := range SortedKeys(vars) { // sorted: a cycle always reports the same key
		v, err := in.resolve(k)
		if err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, nil
}

// OSLookup resolves names against the process environment.
func OSLookup(name string) (string, bool) { return os.LookupEnv(name) }

type interpolator struct {
	raw    map[string]string
	lookup Lookup
	done   map[string]string
	open   map[string]bool
	stack  []string
}

// external resolves a name outside the universe, normally from the shell.
func (in *interpolator) external(name string) string {
	if in.lookup == nil {
		return ""
	}
	v, _ := in.lookup(name)
	return v
}

func (in *interpolator) resolve(key string) (string, error) {
	if v, ok := in.done[key]; ok {
		return v, nil
	}
	if in.open[key] {
		return "", fmt.Errorf("interpolation cycle: %s", strings.Join(append(in.stack, key), " -> "))
	}

	in.open[key] = true
	in.stack = append(in.stack, key)
	defer func() {
		delete(in.open, key)
		in.stack = in.stack[:len(in.stack)-1]
	}()

	var err error
	value := os.Expand(in.raw[key], func(name string) string {
		if err != nil {
			return ""
		}
		if name == "$" { // os.Expand hands us "$" for the $$ escape
			return "$"
		}
		if name == key {
			// PATH: "$PATH:/extra" extends the shell's value; it is not a
			// definition of PATH in terms of itself.
			return in.external(name)
		}
		if _, ok := in.raw[name]; ok {
			var nested string
			nested, err = in.resolve(name)
			return nested
		}
		return in.external(name)
	})
	if err != nil {
		return "", err
	}

	in.done[key] = value
	return value, nil
}
