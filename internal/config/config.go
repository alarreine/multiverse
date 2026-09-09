// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

// Package config reads the multiverse config file and merges the variables of
// a universe. It knows nothing about shells or about the process environment.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Environment is one universe: a named set of environment variables, optionally
// inheriting from another universe.
type Environment struct {
	Name    string            `yaml:"name"`
	Extends string            `yaml:"extends"`
	Envs    map[string]string `yaml:"envs"`
}

// Config is the parsed contents of a multiverse config file.
type Config struct {
	Global       map[string]string `yaml:"global"`
	Environments []Environment     `yaml:"environments"`

	path string
}

// Path reports the file this config was read from.
func (c *Config) Path() string { return c.path }

// Find returns the universe with the given name.
func (c *Config) Find(name string) (Environment, bool) {
	for _, env := range c.Environments {
		if env.Name == name {
			return env, true
		}
	}
	return Environment{}, false
}

// Names lists every universe in declaration order.
func (c *Config) Names() []string {
	names := make([]string, 0, len(c.Environments))
	for _, env := range c.Environments {
		names = append(names, env.Name)
	}
	return names
}

// NotFoundError reports that no config file exists at any candidate location.
type NotFoundError struct{ Searched []string }

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("no config file found, looked in:\n  %s", strings.Join(e.Searched, "\n  "))
}

// Candidates lists the config locations to search, highest precedence first.
// explicit is the --config value and, when set, is the only candidate.
func Candidates(explicit string) []string {
	if explicit != "" {
		return []string{explicit}
	}

	var out []string
	if p := os.Getenv("MULTIVERSE_CONFIG"); p != "" {
		out = append(out, p)
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		out = append(out, filepath.Join(xdg, "multiverse", "config.yaml"))
	} else if home, err := os.UserHomeDir(); err == nil {
		out = append(out, filepath.Join(home, ".config", "multiverse", "config.yaml"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		out = append(out, filepath.Join(home, ".multiverse.yaml"))
	}
	return out
}

// Load reads the first config file that exists among Candidates(explicit).
func Load(explicit string) (*Config, error) {
	candidates := Candidates(explicit)
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		return parse(data, path)
	}
	return nil, &NotFoundError{Searched: candidates}
}

func parse(data []byte, path string) (*Config, error) {
	var cfg Config
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	// Strict: an unknown key is almost always a typo, and silently ignoring it
	// would leave the user hunting for a variable that never gets exported.
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	cfg.path = path
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	seen := make(map[string]bool, len(c.Environments))
	for i, env := range c.Environments {
		if env.Name == "" {
			return fmt.Errorf("environments[%d] has no name", i)
		}
		if seen[env.Name] {
			return fmt.Errorf("universe %q is declared more than once", env.Name)
		}
		seen[env.Name] = true
	}
	for _, env := range c.Environments {
		if env.Extends == "" {
			continue
		}
		if !seen[env.Extends] {
			return fmt.Errorf("universe %q extends %q, which does not exist", env.Name, env.Extends)
		}
	}
	return nil
}

// SortedKeys returns the keys of m in a stable order, so that generated shell
// scripts and diagnostics do not depend on Go's map iteration order.
func SortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
