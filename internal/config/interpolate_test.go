// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package config

import (
	"strings"
	"testing"
)

func lookupFrom(m map[string]string) Lookup {
	return func(name string) (string, bool) {
		v, ok := m[name]
		return v, ok
	}
}

func TestInterpolate(t *testing.T) {
	env := map[string]string{"PATH": "/usr/bin", "REGION": "eu"}

	tests := map[string]struct {
		vars map[string]string
		key  string
		want string
	}{
		"literal value is untouched": {
			vars: map[string]string{"A": "plain"}, key: "A", want: "plain",
		},
		"$VAR from the environment": {
			vars: map[string]string{"A": "$PATH"}, key: "A", want: "/usr/bin",
		},
		"${VAR} from the environment": {
			vars: map[string]string{"A": "https://${REGION}.api.com"}, key: "A", want: "https://eu.api.com",
		},
		"reference to a sibling key": {
			vars: map[string]string{"HOME_DIR": "/opt/tools", "BIN": "${HOME_DIR}/bin"},
			key:  "BIN", want: "/opt/tools/bin",
		},
		"sibling wins over the environment": {
			vars: map[string]string{"REGION": "us", "URL": "$REGION"}, key: "URL", want: "us",
		},
		"prepending to PATH": {
			vars: map[string]string{"TOOLS": "/opt/t", "PATH": "${TOOLS}/bin:$PATH"},
			key:  "PATH", want: "/opt/t/bin:/usr/bin",
		},
		"transitive chain": {
			vars: map[string]string{"A": "$B", "B": "$C", "C": "deep"}, key: "A", want: "deep",
		},
		"unknown name becomes empty": {
			vars: map[string]string{"A": "[$NOPE]"}, key: "A", want: "[]",
		},
		"double dollar is a literal dollar": {
			vars: map[string]string{"A": "cost: $$5"}, key: "A", want: "cost: $5",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Interpolate(tc.vars, lookupFrom(env))
			if err != nil {
				t.Fatalf("Interpolate() error: %v", err)
			}
			if got[tc.key] != tc.want {
				t.Errorf("%s = %q, want %q", tc.key, got[tc.key], tc.want)
			}
		})
	}
}

// A self-reference is the PATH idiom and must not be mistaken for a cycle: the
// name is defined by the universe, but the reference resolves to the shell.
func TestInterpolateSelfReferenceIsNotACycle(t *testing.T) {
	got, err := Interpolate(
		map[string]string{"PATH": "/extra:$PATH"},
		lookupFrom(map[string]string{"PATH": "/usr/bin"}),
	)
	if err != nil {
		t.Fatalf("Interpolate() error: %v", err)
	}
	if want := "/extra:/usr/bin"; got["PATH"] != want {
		t.Errorf("PATH = %q, want %q", got["PATH"], want)
	}
}

func TestInterpolateCycle(t *testing.T) {
	_, err := Interpolate(map[string]string{"A": "$B", "B": "$A"}, nil)
	if err == nil {
		t.Fatal("expected a cycle error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error = %q, want it to mention a cycle", err)
	}
}

// Map iteration order is random, so a value that depends on a sibling must give
// the same answer every run.
func TestInterpolateIsDeterministic(t *testing.T) {
	vars := map[string]string{"A": "$B", "B": "$C", "C": "$D", "D": "end", "E": "$A", "F": "$E"}
	for i := 0; i < 50; i++ {
		got, err := Interpolate(vars, nil)
		if err != nil {
			t.Fatalf("Interpolate() error: %v", err)
		}
		if got["F"] != "end" {
			t.Fatalf("run %d: F = %q, want %q", i, got["F"], "end")
		}
	}
}
