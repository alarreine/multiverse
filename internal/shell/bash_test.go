// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package shell

import (
	"strings"
	"testing"
)

func TestQuote(t *testing.T) {
	tests := map[string]string{
		"":            `''`,
		"plain":       `'plain'`,
		"with space":  `'with space'`,
		"it's":        `'it'\''s'`,
		"$(rm -rf /)": `'$(rm -rf /)'`,
		"a\nb":        "'a\nb'",
		"back`tick":   "'back`tick'",
	}
	for in, want := range tests {
		if got := Quote(in); got != want {
			t.Errorf("Quote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidName(t *testing.T) {
	valid := []string{"PATH", "_x", "A1", "lower_case", "MULTIVERSE_STATE"}
	for _, name := range valid {
		if !ValidName(name) {
			t.Errorf("ValidName(%q) = false, want true", name)
		}
	}
	// Names are emitted unquoted, so anything a shell could act on must be
	// rejected before it reaches eval.
	invalid := []string{"", "1LEADING", "with space", "A=B", "A;rm -rf /", "A-B", "$A", "A\nB"}
	for _, name := range invalid {
		if ValidName(name) {
			t.Errorf("ValidName(%q) = true, want false", name)
		}
	}
}

func TestScript(t *testing.T) {
	var s Script
	s.Export("FOO", "bar")
	s.Export("QUOTED", "it's")
	s.Unset("GONE", "ALSO_GONE")

	want := "export FOO='bar'\n" + `export QUOTED='it'\''s'` + "\nunset GONE\nunset ALSO_GONE\n"
	if got := s.String(); got != want {
		t.Errorf("Script.String() =\n%q\nwant\n%q", got, want)
	}
	if s.Err() != nil {
		t.Errorf("Err() = %v, want nil", s.Err())
	}
}

func TestScriptRejectsInjectedName(t *testing.T) {
	var s Script
	s.Export("OK", "1")
	s.Export("EVIL; rm -rf /", "1")

	if s.Err() == nil {
		t.Fatal("Err() = nil, want an error for an unusable variable name")
	}
	if strings.Contains(s.String(), "rm -rf") {
		t.Errorf("the rejected name still reached the script: %q", s.String())
	}
}

func TestScriptEmpty(t *testing.T) {
	var s Script
	if !s.Empty() {
		t.Error("a fresh Script should be empty")
	}
	s.Unset("A")
	if s.Empty() {
		t.Error("Empty() = true after writing a statement")
	}
}

func TestInitScript(t *testing.T) {
	got := InitScript("/opt/my tools/multiverse")

	// The binary path is interpolated into bash, so a space in it must not split
	// the command.
	if !strings.Contains(got, `'/opt/my tools/multiverse' export --shell bash "$@"`) {
		t.Errorf("InitScript() does not quote the binary path:\n%s", got)
	}
	for _, want := range []string{
		"multiverse() {",
		"case \"$1\" in",
		"use|off)",
		"export " + EnvIntegration + "=bash",
		"if [ $__mv_rc -ne 0 ]; then",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("InitScript() is missing %q:\n%s", want, got)
		}
	}
	// eval must come after the exit status check, or a failed run gets evaluated.
	if strings.Index(got, "__mv_rc -ne 0") > strings.Index(got, `eval "$__mv_out"`) {
		t.Error("InitScript() evaluates the output before checking the exit status")
	}
}
