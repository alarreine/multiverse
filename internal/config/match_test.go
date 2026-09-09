// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"path"
	"strings"
	"testing"
)

func TestMatchName(t *testing.T) {
	const name = "bpm-prod/bonita1/production"

	tests := map[string]struct {
		pattern string
		want    bool
	}{
		// Substring: what the `list | grep foo` pipeline used to do.
		"substring at the start":   {"bpm", true},
		"substring in the middle":  {"bonita", true},
		"substring at the end":     {"production", true},
		"substring across a slash": {"bonita1/prod", true},
		"substring not present":    {"acme", false},
		"substring is case blind":  {"BONITA", true},
		"empty pattern matches":    {"", true},
		"whole name":               {name, true},
		"longer than the name":     {name + "/extra", false},

		// Glob: anchored at both ends, and * stops at a slash.
		"glob needs every level":    {"*/production", false},
		"glob with every level":     {"*/*/production", true},
		"glob is anchored":          {"bonita*", false},
		"glob from the start":       {"bpm-prod/*", false}, // one * cannot span two levels
		"glob spanning two levels":  {"bpm-prod/*/*", true},
		"glob is case blind":        {"BPM-PROD/*/PRODUCTION", true},
		"question mark":             {"bpm-prod/bonita?/production", true},
		"question mark is one char": {"bpm-prod/bonita??/production", false},
		"character class":           {"bpm-prod/bonita[12]/production", true},
		"character class excluding": {"bpm-prod/bonita[23]/production", false},
	}

	for label, tc := range tests {
		t.Run(label, func(t *testing.T) {
			got, err := MatchName(name, tc.pattern)
			if err != nil {
				t.Fatalf("MatchName(%q, %q) error: %v", name, tc.pattern, err)
			}
			if got != tc.want {
				t.Errorf("MatchName(%q, %q) = %v, want %v", name, tc.pattern, got, tc.want)
			}
		})
	}
}

// A malformed glob must be an error, so that the caller can tell it apart from
// an honest "nothing matched".
func TestMatchNameBadPattern(t *testing.T) {
	for _, pattern := range []string{"bonita[", "[a-", "bpm-prod/[abc"} {
		_, err := MatchName("bpm-prod/bonita1", pattern)
		if !errors.Is(err, path.ErrBadPattern) {
			t.Errorf("MatchName(_, %q) error = %v, want path.ErrBadPattern", pattern, err)
		}
	}
}

// A reversed range is not a bad pattern to path.Match, it simply matches
// nothing. Pinned because it is the kind of stdlib detail worth not guessing at.
func TestReversedRangeMatchesNothing(t *testing.T) {
	got, err := MatchName("bpm-prod/bonita1", "bpm-prod/bonita[3-1]")
	if err != nil {
		t.Fatalf("MatchName() error = %v, want nil", err)
	}
	if got {
		t.Error("MatchName() = true, want false")
	}
}

// A pattern without wildcards can never be a bad pattern, even when it holds
// characters a glob would choke on.
func TestSubstringPatternNeverErrors(t *testing.T) {
	if _, err := MatchName("a]b", "]"); err != nil {
		t.Errorf("a wildcard-free pattern returned an error: %v", err)
	}
}

func testMatchConfig() *Config {
	return &Config{Environments: []Environment{
		{Name: "bpm-prod"},
		{Name: "bpm-prod/bonita1", Extends: "bpm-prod"},
		{Name: "bpm-prod/bonita1/jenkins", Extends: "bpm-prod/bonita1"},
		{Name: "bpm-prod/bonita1/production", Extends: "bpm-prod/bonita1"},
		{Name: "bpm-dev/acme5/jenkins", Extends: "bpm-dev/acme5"},
	}}
}

func TestMatch(t *testing.T) {
	tests := map[string]struct {
		pattern string
		want    []string
	}{
		"empty pattern returns everything": {"", []string{
			"bpm-prod", "bpm-prod/bonita1", "bpm-prod/bonita1/jenkins",
			"bpm-prod/bonita1/production", "bpm-dev/acme5/jenkins",
		}},
		"substring": {"bonita1", []string{
			"bpm-prod/bonita1", "bpm-prod/bonita1/jenkins", "bpm-prod/bonita1/production",
		}},
		"glob across accounts": {"*/*/jenkins", []string{
			"bpm-prod/bonita1/jenkins", "bpm-dev/acme5/jenkins",
		}},
		"nothing matches": {"zzz", nil},
	}

	for label, tc := range tests {
		t.Run(label, func(t *testing.T) {
			got, err := testMatchConfig().Match(tc.pattern)
			if err != nil {
				t.Fatalf("Match(%q) error: %v", tc.pattern, err)
			}
			var names []string
			for _, env := range got {
				names = append(names, env.Name)
			}
			if strings.Join(names, ",") != strings.Join(tc.want, ",") {
				t.Errorf("Match(%q) = %v, want %v (declaration order)", tc.pattern, names, tc.want)
			}
		})
	}
}

// No match is not the matcher's error to raise: cmd/list.go decides that.
func TestMatchNoResultsIsNotAnError(t *testing.T) {
	got, err := testMatchConfig().Match("nothing-like-this")
	if err != nil {
		t.Errorf("Match() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("Match() = %v, want empty", got)
	}
}

func TestMatchPropagatesBadPattern(t *testing.T) {
	if _, err := testMatchConfig().Match("bonita["); !errors.Is(err, path.ErrBadPattern) {
		t.Errorf("Match() error = %v, want path.ErrBadPattern", err)
	}
}
