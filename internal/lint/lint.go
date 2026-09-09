// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

// Package lint reports problems in a multiverse config without applying it.
package lint

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/alarreine/multiverse/internal/config"
	"github.com/alarreine/multiverse/internal/shell"
	"github.com/alarreine/multiverse/internal/state"
)

// Finding is one problem found in a config.
type Finding struct {
	Universe string
	Message  string
	// Fatal separates "this universe cannot be used at all" from "it works, but
	// it probably does not do what you meant".
	Fatal bool
}

// Config checks every universe the way `use` would, but without touching the
// environment: that the extends chain resolves, that every variable name is
// usable in bash, that none invades multiverse's own namespace, that
// interpolation terminates, and that no value refers to a name that expands to
// nothing. lookup supplies the surrounding environment.
//
// Findings come back in the order the config declares the universes.
func Config(cfg *config.Config, lookup config.Lookup) []Finding {
	var out []Finding
	for _, name := range cfg.Names() {
		out = append(out, universe(cfg, name, lookup)...)
	}
	return out
}

func universe(cfg *config.Config, name string, lookup config.Lookup) []Finding {
	raw, err := cfg.Resolve(name, false)
	if err != nil {
		// A chain that does not resolve makes every other check meaningless.
		return []Finding{{Universe: name, Message: err.Error(), Fatal: true}}
	}

	var out []Finding
	for _, key := range config.SortedKeys(raw) {
		switch {
		case !shell.ValidName(key):
			out = append(out, Finding{
				Universe: name,
				Message: fmt.Sprintf("%q cannot be an environment variable name: bash needs a letter "+
					"or _ followed by letters, digits or _", key),
				Fatal: true,
			})
		case state.IsReserved(key):
			out = append(out, Finding{
				Universe: name,
				Message: fmt.Sprintf("%s is reserved: multiverse keeps its own bookkeeping under %s",
					key, state.EnvPrefix),
				Fatal: true,
			})
		}
	}

	refs := &recorder{base: lookup, seen: map[string]bool{}}
	if _, err := config.Interpolate(raw, refs.lookup); err != nil {
		return append(out, Finding{Universe: name, Message: err.Error(), Fatal: true})
	}
	for _, ref := range refs.missing() {
		msg := fmt.Sprintf("$%s is defined neither by this universe nor by the environment, "+
			"so it expands to nothing", ref)
		if users := referencedBy(raw, ref); len(users) > 0 {
			msg += fmt.Sprintf(" (used by %s)", strings.Join(users, ", "))
		}
		out = append(out, Finding{Universe: name, Message: msg})
	}
	return out
}

// recorder wraps a Lookup to note the names it could not resolve.
// config.Interpolate consults the lookup only for names the universe does not
// define itself, so a failed lookup means exactly "defined nowhere".
type recorder struct {
	base config.Lookup
	seen map[string]bool
}

func (r *recorder) lookup(name string) (string, bool) {
	if r.base != nil {
		if value, ok := r.base(name); ok {
			return value, true
		}
	}
	r.seen[name] = true
	return "", false
}

func (r *recorder) missing() []string {
	names := make([]string, 0, len(r.seen))
	for name := range r.seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// referencedBy lists the keys whose value mentions $name or ${name}. It expands
// with os.Expand rather than matching a regexp so that this parser cannot drift
// from the one config.Interpolate uses.
func referencedBy(raw map[string]string, name string) []string {
	var keys []string
	for _, key := range config.SortedKeys(raw) {
		mentioned := false
		os.Expand(raw[key], func(ref string) string {
			if ref == name {
				mentioned = true
			}
			return ""
		})
		if mentioned {
			keys = append(keys, key)
		}
	}
	return keys
}
