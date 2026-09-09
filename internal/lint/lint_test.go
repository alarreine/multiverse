// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

package lint

import (
	"strings"
	"testing"

	"github.com/alarreine/multiverse/internal/config"
)

func lookupFrom(env map[string]string) config.Lookup {
	return func(name string) (string, bool) {
		v, ok := env[name]
		return v, ok
	}
}

// split separates findings by severity, keeping the messages.
func split(findings []Finding) (fatal, warn []string) {
	for _, f := range findings {
		line := f.Universe + ": " + f.Message
		if f.Fatal {
			fatal = append(fatal, line)
		} else {
			warn = append(warn, line)
		}
	}
	return fatal, warn
}

func TestConfig(t *testing.T) {
	tests := map[string]struct {
		cfg       *config.Config
		env       map[string]string
		wantFatal []string // substrings, one per expected error
		wantWarn  []string
	}{
		"clean config": {
			cfg: &config.Config{
				Global: map[string]string{"KUBECONFIG": "$HOME/.kube/config-${CLUSTER_NAME}"},
				Environments: []config.Environment{
					{Name: "base", Envs: map[string]string{"CLUSTER_NAME": "c1"}},
					{Name: "leaf", Extends: "base", Envs: map[string]string{"NAMESPACE": "ns"}},
				},
			},
			env: map[string]string{"HOME": "/home/x"},
		},
		"unusable variable name": {
			cfg: &config.Config{Environments: []config.Environment{
				{Name: "u", Envs: map[string]string{"HAS SPACE": "1"}},
			}},
			wantFatal: []string{`"HAS SPACE" cannot be an environment variable name`},
		},
		"reserved namespace": {
			cfg: &config.Config{Environments: []config.Environment{
				{Name: "u", Envs: map[string]string{"MULTIVERSE_STATE": "hijack"}},
			}},
			wantFatal: []string{"MULTIVERSE_STATE is reserved"},
		},
		"extends cycle": {
			cfg: &config.Config{Environments: []config.Environment{
				{Name: "a", Extends: "b"},
				{Name: "b", Extends: "a"},
			}},
			// Both universes are unresolvable, and each reports it.
			wantFatal: []string{"extends cycle", "extends cycle"},
		},
		"interpolation cycle": {
			cfg: &config.Config{Environments: []config.Environment{
				{Name: "u", Envs: map[string]string{"A": "$B", "B": "$A"}},
			}},
			wantFatal: []string{"interpolation cycle"},
		},
		"reference nothing defines": {
			cfg: &config.Config{Environments: []config.Environment{
				{Name: "u", Envs: map[string]string{
					"URL":  "https://${REGION}.example.com",
					"HOST": "$REGION",
				}},
			}},
			wantWarn: []string{"$REGION is defined neither by this universe nor by the environment"},
		},
		"reference satisfied by the environment": {
			cfg: &config.Config{Environments: []config.Environment{
				{Name: "u", Envs: map[string]string{"URL": "https://${REGION}.example.com"}},
			}},
			env: map[string]string{"REGION": "eu"},
		},
		// The real-world case: KUBECONFIG in global refers to a CLUSTER_NAME that
		// only the cluster universes define, so a parent-only universe warns.
		"parent-only universe": {
			cfg: &config.Config{
				Global: map[string]string{"KUBECONFIG": "$HOME/.kube/config-${CLUSTER_NAME}"},
				Environments: []config.Environment{
					{Name: "account", Envs: map[string]string{"AWS_PROFILE": "p"}},
				},
			},
			env:      map[string]string{"HOME": "/home/x"},
			wantWarn: []string{"$CLUSTER_NAME is defined neither"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fatal, warn := split(Config(tc.cfg, lookupFrom(tc.env)))

			for _, group := range []struct {
				kind string
				got  []string
				want []string
			}{{"error", fatal, tc.wantFatal}, {"warning", warn, tc.wantWarn}} {
				if len(group.got) != len(group.want) {
					t.Errorf("got %d %ss, want %d:\n  got:  %v\n  want: %v",
						len(group.got), group.kind, len(group.want), group.got, group.want)
					continue
				}
				for i, want := range group.want {
					if !strings.Contains(group.got[i], want) {
						t.Errorf("%s %d = %q, want it to contain %q", group.kind, i, group.got[i], want)
					}
				}
			}
		})
	}
}

// The warning has to name the keys that use the reference, or it is not
// actionable in a config with dozens of variables.
func TestWarningNamesTheKeysUsingTheReference(t *testing.T) {
	cfg := &config.Config{Environments: []config.Environment{
		{Name: "u", Envs: map[string]string{
			"URL":       "https://${REGION}.example.com",
			"BUCKET":    "data-$REGION",
			"UNRELATED": "plain",
		}},
	}}

	findings := Config(cfg, nil)
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(findings), findings)
	}
	if got := findings[0].Message; !strings.Contains(got, "(used by BUCKET, URL)") {
		t.Errorf("message = %q, want it to name BUCKET and URL in order", got)
	}
}

// A self-reference resolves against the shell, so it is a normal reference and
// must be reported when the shell does not have it either.
func TestSelfReferenceIsReportedOnlyWhenTheShellLacksIt(t *testing.T) {
	cfg := &config.Config{Environments: []config.Environment{
		{Name: "u", Envs: map[string]string{"PATH": "/extra:$PATH"}},
	}}

	if findings := Config(cfg, lookupFrom(map[string]string{"PATH": "/usr/bin"})); len(findings) != 0 {
		t.Errorf("with PATH in the shell, got %+v, want no findings", findings)
	}
	findings := Config(cfg, nil)
	if len(findings) != 1 || findings[0].Fatal {
		t.Fatalf("without PATH in the shell, got %+v, want one warning", findings)
	}
}

// An unresolvable chain has to stop the other checks: reporting a cycle and
// then a pile of consequences of that cycle is noise.
func TestCycleSuppressesOtherFindings(t *testing.T) {
	cfg := &config.Config{Environments: []config.Environment{
		{Name: "a", Extends: "b", Envs: map[string]string{"BAD NAME": "1"}},
		{Name: "b", Extends: "a"},
	}}

	for _, f := range Config(cfg, nil) {
		if strings.Contains(f.Message, "BAD NAME") {
			t.Errorf("reported %q on top of the cycle", f.Message)
		}
	}
}
