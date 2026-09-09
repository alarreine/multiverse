// Copyright © 2024 Agustin LARREINEGABE
// SPDX-License-Identifier: MIT

// Package state encodes what multiverse did to the shell, so that a later
// `off` or `use` can undo it exactly.
package state

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// EnvPrefix is the namespace multiverse keeps for its own bookkeeping.
const EnvPrefix = "MULTIVERSE_"

// Names of the two variables the shell integration carries.
const (
	EnvState    = EnvPrefix + "STATE"    // opaque blob, this package's business
	EnvUniverse = EnvPrefix + "UNIVERSE" // human readable, for status and prompts
)

// IsReserved reports whether a name belongs to multiverse's own namespace. A
// universe that could set one of these would corrupt the record of what to
// undo, so config values under the prefix are refused.
func IsReserved(name string) bool { return strings.HasPrefix(name, EnvPrefix) }

// Var records what a variable looked like before multiverse touched it. Had is
// false when the variable did not exist, which is the difference between
// restoring a value and unsetting it: a universe that sets PATH must not leave
// the shell without one.
type Var struct {
	Had  bool   `json:"had"`
	Prev string `json:"prev,omitempty"`
}

// State is the set of variables the active universe overwrote.
type State struct {
	Universe string         `json:"universe"`
	Vars     map[string]Var `json:"vars"`
}

// Keys lists the recorded variables in a stable order.
func (s State) Keys() []string {
	keys := make([]string, 0, len(s.Vars))
	for k := range s.Vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Encode packs the state for transport in an environment variable. Values can
// contain quotes, newlines and '=', so the JSON is gzipped and base64-encoded
// rather than stored in any readable form.
func Encode(s State) (string, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("encoding state: %w", err)
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(raw); err != nil {
		return "", fmt.Errorf("compressing state: %w", err)
	}
	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("compressing state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf.Bytes()), nil
}

// Decode unpacks a blob produced by Encode.
func Decode(blob string) (State, error) {
	packed, err := base64.RawURLEncoding.DecodeString(blob)
	if err != nil {
		return State{}, fmt.Errorf("decoding %s: %w", EnvState, err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(packed))
	if err != nil {
		return State{}, fmt.Errorf("decoding %s: %w", EnvState, err)
	}
	defer zr.Close()

	raw, err := io.ReadAll(zr)
	if err != nil {
		return State{}, fmt.Errorf("decoding %s: %w", EnvState, err)
	}

	var s State
	if err := json.Unmarshal(raw, &s); err != nil {
		return State{}, fmt.Errorf("decoding %s: %w", EnvState, err)
	}
	return s, nil
}

// FromEnv reads the active state. The second result is false when no universe
// is applied; an unreadable blob is an error rather than a silent reset,
// because forgetting it would strand the previous universe's variables.
func FromEnv(getenv func(string) string) (State, bool, error) {
	blob := getenv(EnvState)
	if blob == "" {
		return State{}, false, nil
	}
	s, err := Decode(blob)
	if err != nil {
		return State{}, false, fmt.Errorf("%w\n  recover with: unset %s %s", err, EnvState, EnvUniverse)
	}
	return s, true, nil
}
