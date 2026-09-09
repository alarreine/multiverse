// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package state

import (
	"strings"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	want := State{
		Universe: "prod",
		Vars: map[string]Var{
			"PATH":      {Had: true, Prev: "/usr/bin"},
			"API_URL":   {Had: false},
			"EMPTY":     {Had: true, Prev: ""},
			"AWKWARD":   {Had: true, Prev: "it's \"quoted\"\nand multiline\ttoo"},
			"UNICODE":   {Had: true, Prev: "días ✓ 世界"},
			"EQUALSIGN": {Had: true, Prev: "a=b=c"},
		},
	}

	blob, err := Encode(want)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	// The blob travels through an environment variable, so it must survive as a
	// single shell word with nothing a shell would interpret.
	if strings.ContainsAny(blob, " \t\n'\"$`\\") {
		t.Errorf("Encode() produced shell-unsafe characters: %q", blob)
	}

	got, err := Decode(blob)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	if got.Universe != want.Universe {
		t.Errorf("Universe = %q, want %q", got.Universe, want.Universe)
	}
	for name, wantVar := range want.Vars {
		if got.Vars[name] != wantVar {
			t.Errorf("Vars[%s] = %+v, want %+v", name, got.Vars[name], wantVar)
		}
	}
}

// Had distinguishes restoring a value from unsetting it, so an empty previous
// value must not come back as "never existed".
func TestEmptyValueSurvivesAsHad(t *testing.T) {
	blob, err := Encode(State{Universe: "u", Vars: map[string]Var{"E": {Had: true, Prev: ""}}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(blob)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Vars["E"].Had {
		t.Error("Had was lost for a variable whose previous value was empty")
	}
}

func TestDecodeGarbage(t *testing.T) {
	for _, blob := range []string{"not base64 !!", "aGVsbG8", ""} {
		if _, err := Decode(blob); err == nil {
			t.Errorf("Decode(%q) succeeded, want an error", blob)
		}
	}
}

func TestFromEnv(t *testing.T) {
	blob, err := Encode(State{Universe: "prod", Vars: map[string]Var{"A": {Had: false}}})
	if err != nil {
		t.Fatal(err)
	}
	env := map[string]string{EnvState: blob}
	getenv := func(k string) string { return env[k] }

	got, active, err := FromEnv(getenv)
	if err != nil || !active || got.Universe != "prod" {
		t.Fatalf("FromEnv() = (%+v, %v, %v), want the prod state", got, active, err)
	}

	env[EnvState] = ""
	if _, active, err := FromEnv(getenv); active || err != nil {
		t.Errorf("an unset blob should mean no universe, got (active=%v, err=%v)", active, err)
	}

	env[EnvState] = "corrupt!!"
	_, _, err = FromEnv(getenv)
	if err == nil {
		t.Fatal("a corrupt blob should be an error, not a silent reset")
	}
	if !strings.Contains(err.Error(), "unset "+EnvState) {
		t.Errorf("error = %q, want it to say how to recover", err)
	}
}

func TestKeysAreSorted(t *testing.T) {
	s := State{Vars: map[string]Var{"C": {}, "A": {}, "B": {}}}
	got := strings.Join(s.Keys(), ",")
	if got != "A,B,C" {
		t.Errorf("Keys() = %q, want %q", got, "A,B,C")
	}
}

func TestIsReserved(t *testing.T) {
	// The two variables this package carries must be covered by the prefix, or
	// a universe could overwrite the record of what to undo.
	for _, name := range []string{EnvState, EnvUniverse, EnvPrefix + "ANYTHING", "MULTIVERSE_SHELL_INTEGRATION"} {
		if !IsReserved(name) {
			t.Errorf("IsReserved(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"PATH", "AWS_PROFILE", "MULTIVERSE", "MY_MULTIVERSE_VAR", ""} {
		if IsReserved(name) {
			t.Errorf("IsReserved(%q) = true, want false", name)
		}
	}
}
