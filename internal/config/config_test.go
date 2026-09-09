// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const minimal = "environments:\n  - name: prod\n    envs:\n      A: '1'\n"

func TestCandidatesPrecedence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("MULTIVERSE_CONFIG", "/from/env.yaml")

	got := Candidates("")
	want := []string{
		"/from/env.yaml",
		filepath.Join(home, ".config", "multiverse", "config.yaml"),
		filepath.Join(home, ".multiverse.yaml"),
	}
	if len(got) != len(want) {
		t.Fatalf("Candidates() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("candidate %d = %q, want %q", i, got[i], want[i])
		}
	}

	if only := Candidates("/explicit.yaml"); len(only) != 1 || only[0] != "/explicit.yaml" {
		t.Errorf("--config should win outright, got %v", only)
	}
}

func TestLoadFallsBackToLegacyPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("MULTIVERSE_CONFIG", "")

	legacy := filepath.Join(home, ".multiverse.yaml")
	write(t, legacy, minimal)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Path() != legacy {
		t.Errorf("Path() = %q, want %q", cfg.Path(), legacy)
	}

	// The XDG location, once it exists, takes precedence over the legacy one.
	xdg := filepath.Join(home, ".config", "multiverse", "config.yaml")
	write(t, xdg, minimal)
	cfg, err = Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Path() != xdg {
		t.Errorf("Path() = %q, want %q", cfg.Path(), xdg)
	}
}

func TestLoadNotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("MULTIVERSE_CONFIG", "")

	_, err := Load("")
	var nf *NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("Load() error = %v, want *NotFoundError", err)
	}
	if len(nf.Searched) == 0 {
		t.Error("NotFoundError should list where it looked")
	}
}

func TestParseRejectsBadConfigs(t *testing.T) {
	tests := map[string]struct{ body, want string }{
		"unknown field": {"enviroments:\n  - name: prod\n", "field enviroments"},
		"no name":       {"environments:\n  - envs:\n      A: '1'\n", "no name"},
		"duplicate":     {"environments:\n  - name: prod\n  - name: prod\n", "more than once"},
		"bad extends":   {"environments:\n  - name: prod\n    extends: ghost\n", "does not exist"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := parse([]byte(tc.body), "test.yaml")
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}
