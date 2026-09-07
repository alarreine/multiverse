// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package config

import (
	"strings"
	"testing"
)

func testConfig() *Config {
	return &Config{
		Global: map[string]string{"SHARED": "global", "ONLY_GLOBAL": "g"},
		Environments: []Environment{
			{Name: "base", Envs: map[string]string{"SHARED": "base", "ONLY_BASE": "b"}},
			{Name: "prod", Extends: "base", Envs: map[string]string{"SHARED": "prod"}},
			{Name: "sandbox", Envs: map[string]string{"SHARED": "sandbox"}},
		},
	}
}

func TestResolvePrecedence(t *testing.T) {
	tests := map[string]struct {
		universe   string
		omitGlobal bool
		want       map[string]string
	}{
		"universe beats global": {
			universe: "sandbox",
			want:     map[string]string{"SHARED": "sandbox", "ONLY_GLOBAL": "g"},
		},
		"universe beats parent beats global": {
			universe: "prod",
			want:     map[string]string{"SHARED": "prod", "ONLY_BASE": "b", "ONLY_GLOBAL": "g"},
		},
		"omit-global drops the global block": {
			universe:   "prod",
			omitGlobal: true,
			want:       map[string]string{"SHARED": "prod", "ONLY_BASE": "b"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := testConfig().Resolve(tc.universe, tc.omitGlobal)
			if err != nil {
				t.Fatalf("Resolve() error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("Resolve() = %v, want %v", got, tc.want)
			}
			for k, want := range tc.want {
				if got[k] != want {
					t.Errorf("%s = %q, want %q", k, got[k], want)
				}
			}
		})
	}
}

func TestResolveDoesNotMutateGlobal(t *testing.T) {
	cfg := testConfig()
	if _, err := cfg.Resolve("prod", false); err != nil {
		t.Fatal(err)
	}
	if got := cfg.Global["SHARED"]; got != "global" {
		t.Errorf("the global block was overwritten in place: SHARED = %q, want %q", got, "global")
	}
	if _, leaked := cfg.Global["ONLY_BASE"]; leaked {
		t.Error("universe variables leaked into the global block")
	}
}

func TestResolveUnknownUniverse(t *testing.T) {
	_, err := testConfig().Resolve("nowhere", false)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "available: base, prod, sandbox") {
		t.Errorf("error = %q, want it to list the available universes", err)
	}
}

func TestResolveExtendsCycle(t *testing.T) {
	cfg := &Config{Environments: []Environment{
		{Name: "a", Extends: "b"},
		{Name: "b", Extends: "a"},
	}}
	_, err := cfg.Resolve("a", false)
	if err == nil {
		t.Fatal("expected a cycle error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error = %q, want it to mention a cycle", err)
	}
}
