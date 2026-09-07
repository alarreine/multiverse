# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`multiverse` is a single-binary Go CLI (Cobra) that switches the **current bash session** between
named sets of environment variables, in the spirit of `direnv` but keyed on an explicit universe
name rather than on the working directory.

## Commands

```bash
go build                     # produces ./multiverse (gitignored)
go test ./...                # unit tests live in internal/, cmd/ has none
go test ./internal/config -run TestInterpolate
go vet ./... && gofmt -l .
go run . list --config .example.multiverse.yaml
```

`--config .example.multiverse.yaml` (or `MULTIVERSE_CONFIG=`) is how to exercise the tool in-repo
without a config in `$HOME`.

## The central constraint

A child process cannot modify its parent shell's environment. `use` and `off` therefore do **not**
work by themselves: `multiverse init bash` prints a bash function that the user installs with
`eval "$(multiverse init bash)"`, and that function evaluates what the binary prints. Everything
about the design follows from this.

**stdout carries shell code that will be `eval`'d; nothing else may go there.** All diagnostics and
errors go to stderr via `warnf` (`cmd/root.go`) or by returning an error from a `RunE`. Never add a
`fmt.Println` to a command on the export path, and never `panic` — a stack trace on stdout would be
executed by the user's shell. The whole script is built in a `shell.Script` buffer and written only
after `sc.Err()` comes back nil, so a failed run emits nothing at all.

`cmd/use.go` and `cmd/off.go` are deliberately non-functional stubs: when the binary is reached
directly it means the integration is missing (or was bypassed), and they explain which. The real
work is in the hidden `export` command.

## Architecture

`cmd/` is a thin Cobra layer — flags, wiring, output. All logic lives in `internal/` so it can be
tested without a shell:

- `internal/config` — parsing (strict yaml.v3, unknown keys are errors), file discovery, `Resolve`
  (merges `global` + the `extends` chain + the universe's own `envs`), and `Interpolate`.
- `internal/state` — the `MULTIVERSE_STATE` blob: JSON → gzip → base64, carried in the environment
  and inherited by the binary as a normal child process.
- `internal/shell` — `Quote`, the `Script` builder, and the `init bash` template.

`cmd/export.go` is where it all comes together and is the file to read first.

### Two invariants that are easy to break

**Interpolate against the *reverted* environment, not `os.Environ()`.** `pristineLookup`
(`cmd/export.go`) reports the environment as it was before the active universe was applied. Using
the live environment instead would make `use prod` twice grow `PATH: "${TOOLS}/bin:$PATH"` by a
segment each time, and would make `status --verbose` report an applied universe as different from
itself.

**A self-reference is not a cycle.** In `internal/config/interpolate.go`, `$PATH` inside the value
of `PATH` resolves to `lookup` (the shell), never recursively. Resolution is otherwise recursive and
memoised with cycle detection, because Go map order is random and a single pass over the map would
give different answers on different runs.

### Other things worth knowing

- Variable names are emitted **unquoted** into bash, so `shell.ValidName` gates every one of them —
  it is an injection boundary, not a nicety. Values are safe by construction via `shell.Quote`.
- Names starting with `MULTIVERSE_` are rejected in config: a universe that could set the state blob
  would corrupt the record of what to undo.
- `state.Var.Had` is the difference between restoring a value and unsetting it. Dropping it would
  make `off` on a universe that sets `PATH` leave the shell without one.
- Config loading is lazy, called from the commands that need it, so `init` and `--help` work on a
  machine with no config file.
- Only bash is supported; `--shell` and `init <shell>` reject anything else rather than guessing.
