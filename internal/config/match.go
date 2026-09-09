// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package config

import (
	"path"
	"strings"
)

// MatchName reports whether name matches pattern.
//
// A pattern without wildcards is a case-insensitive substring, which is what a
// `list | grep foo` used to do. A pattern containing *, ? or [ is a glob over
// the whole name, matched with path.Match: * does not cross a '/', which is
// what makes globbing useful over path-shaped universe names
// ("bpm-prod/*/jenkins"). Globs are lowercased too, so both forms agree about
// case. An unparseable glob comes back as path.ErrBadPattern.
func MatchName(name, pattern string) (bool, error) {
	name, pattern = strings.ToLower(name), strings.ToLower(pattern)
	if !strings.ContainsAny(pattern, "*?[") {
		return strings.Contains(name, pattern), nil
	}
	return path.Match(pattern, name)
}

// Match returns the universes whose name matches pattern, in declaration order.
// An empty pattern matches every universe. No match is an empty slice and no
// error: whether that is a failure is the caller's call, unlike a malformed
// pattern, which always is one.
func (c *Config) Match(pattern string) ([]Environment, error) {
	if pattern == "" {
		return c.Environments, nil
	}

	var out []Environment
	for _, env := range c.Environments {
		ok, err := MatchName(env.Name, pattern)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, env)
		}
	}
	return out, nil
}
