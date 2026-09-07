// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"strings"
)

// Resolve merges the global block, the extends chain and the universe's own
// envs into a single map, without interpolating. Precedence, lowest to highest:
// global, then each ancestor from the farthest to the nearest, then the
// universe itself.
func (c *Config) Resolve(name string, omitGlobal bool) (map[string]string, error) {
	chain, err := c.chain(name)
	if err != nil {
		return nil, err
	}

	out := make(map[string]string)
	if !omitGlobal {
		for k, v := range c.Global {
			out[k] = v
		}
	}
	for i := len(chain) - 1; i >= 0; i-- {
		for k, v := range chain[i].Envs {
			out[k] = v
		}
	}
	return out, nil
}

// chain returns the universe followed by its ancestors, nearest first.
func (c *Config) chain(name string) ([]Environment, error) {
	var (
		chain []Environment
		path  []string
		seen  = make(map[string]bool)
	)
	for cur := name; cur != ""; {
		if seen[cur] {
			return nil, fmt.Errorf("extends cycle: %s", strings.Join(append(path, cur), " -> "))
		}
		env, ok := c.Find(cur)
		if !ok {
			return nil, c.unknownUniverse(cur)
		}
		seen[cur] = true
		path = append(path, cur)
		chain = append(chain, env)
		cur = env.Extends
	}
	return chain, nil
}

func (c *Config) unknownUniverse(name string) error {
	available := c.Names()
	if len(available) == 0 {
		return fmt.Errorf("unknown universe %q: %s declares none", name, c.path)
	}
	return fmt.Errorf("unknown universe %q (available: %s)", name, strings.Join(available, ", "))
}
