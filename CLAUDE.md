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

Nothing under `cmd/` is unit-tested, so anything touching the shell contract needs checking by hand
in a throwaway bash:

```bash
go build && export MULTIVERSE_CONFIG=$PWD/.example.multiverse.yaml
bash -c 'eval "$(./multiverse init bash)"; before=$PATH
         multiverse use prod; multiverse use prod   # must be idempotent: PATH grows once
         multiverse use sandbox; multiverse off
         [ "$PATH" = "$before" ] && echo restored'
./multiverse export --shell bash use prod 2>/dev/null   # stdout must be pure bash
```

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

`cmd/use.go` holds `useCmd` and `offCmd`, both deliberately non-functional stubs: reaching the binary
directly means the integration is missing, or was bypassed, and they explain which (told apart by the
`MULTIVERSE_SHELL_INTEGRATION` marker the template exports). The real work is in the hidden `export`
command.

**Adding a subcommand that changes the environment means editing two files.** The generated function
dispatches on a hard-coded list — `case "$1" in use|off)` in `internal/shell/bash.go` — and forwards
`"$@"` after `export --shell bash`, so `export`'s first positional argument is the action name. A new
mutating command that is not added to that list silently reaches the plain binary and does nothing at
all, which is the quietest possible failure.

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
- The flag variables `omitGlobal` and `envAlias` are declared once in `cmd/export.go` and bound by
  several commands (`export`, `use`, and `omitGlobal` again by `status`). That is safe only because
  exactly one command runs per process — do not read them outside a `RunE`.
- `universeName` (`cmd/export.go`) takes the universe from the positional argument and falls back to
  the older `-e/--env` flag, which is kept working for muscle memory from the `apply -e prod` days.
