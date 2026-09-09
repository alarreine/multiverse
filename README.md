# Multiverse: The CLI Time Machine for Environment Configurations

Welcome to **Multiverse**, the command-line tool that's not just a tool, but a portal to effortlessly navigate through the vast landscapes of environment configurations. Think of it as your trusty time machine, zipping you between parallel worlds of development, testing, and production.

```bash
$ multiverse use prod
universe: prod (8 vars)
$ echo $API_URL
https://eu.api.example.com          # applied right here, in this shell
$ multiverse off
universe: none (left prod)
$ echo $API_URL
                                    # and cleanly undone
```

## Inspiration
Multiverse draws creative inspiration from the powerful simplicity of [direnv](https://direnv.net/), a pioneering tool in the realm of environment switching. We owe a cosmic hat tip to [direnv](https://direnv.net/) for lighting the path and showing us the vast potential of environment management. Where direnv follows your *directory*, Multiverse follows your *intent*: universes are named, and you enter one wherever you happen to be standing.

## Features

- 🌌 **Navigate Through Configurations**: Like jumping through different realities, switch between your environments with ease.
- 🪄 **Applied Where You Stand**: `use` and `off` change the shell you are typing in — no `source <(...)`, no subshell, no new terminal.
- ↩️ **Nothing Left Behind**: `off` puts back the values that were there before and unsets the ones the universe invented. Your `PATH` survives.
- 🧬 **Inheritance and Interpolation**: universes can `extend` one another, and values can refer to other values with `$VAR`.
- 🤫 **Whisper Mode**: Activate with `--quiet` to mute the chatter, and enjoy the quiet cosmos while navigating your configurations.

## Installation

Grab your universal translator (aka your terminal) and type the ancient runes:

```bash
git clone git@github.com:alarreine/multiverse.git
cd multiverse
go build
sudo install multiverse /usr/local/bin/     # or put it anywhere on your PATH
```

### Install the shell integration

This step is not optional, and here is why: no program can reach into the shell
that launched it and change its variables. What Multiverse does instead is print
the shell code and let a small bash function evaluate it for you. Add this to
your `~/.bashrc`:

```bash
eval "$(multiverse init bash)"
```

Open a new terminal, and `multiverse use` starts working on the spot. Only bash
is supported today.

If you would rather not install anything, `export` is the same thing with the
seams showing:

```bash
eval "$(multiverse export --shell bash use prod)"
```

## Configuration

Multiverse reads the first file it finds, in this order:

1. `--config <path>`
2. `$MULTIVERSE_CONFIG`
3. `~/.config/multiverse/config.yaml`
4. `~/.multiverse.yaml`

See [`.example.multiverse.yaml`](.example.multiverse.yaml):

```yaml
global:                       # applied on top of every universe
  MY_GLOBAL_ENV1: value1

environments:
  - name: base
    envs:
      TOOLS_HOME: /opt/tools
      PATH: "${TOOLS_HOME}/bin:$PATH"
      REGION: eu

  - name: prod
    extends: base             # inherit base, then override
    envs:
      API_URL: "https://${REGION}.api.example.com"
```

**Interpolation.** Values may use `$VAR` and `${VAR}`. A name is looked up among
the universe's own variables first, then in your shell. The one exception is a
variable that mentions itself — `PATH: "${TOOLS_HOME}/bin:$PATH"` — which always
means the value already in your shell; that is what lets you prepend to `PATH`
without it growing every time you run `use`. Write `$$` for a literal `$`.

**Inheritance.** `extends` pulls in another universe. Precedence, lowest to
highest: `global`, then ancestors from farthest to nearest, then the universe's
own `envs`.

## Usage

### Global Flags
* `--config`: Point at a specific config file.
* `--quiet`, `-q`: Mute the chatter, enjoy the quiet cosmos.

### use
Enter a universe, right here in this shell:

```bash
multiverse use prod
multiverse use prod --omit-global   # skip the global block
```

Running it again is a no-op, and switching straight to another universe cleans up
the first one on the way.

### off
Leave the current universe. Variables that existed before you arrived go back to
their previous values; the ones the universe introduced are unset.

```bash
multiverse off
```

### status
Peer into the cosmic map of your environment variables:

```bash
multiverse status        # which universe am I in?
multiverse status -v     # ...and does my shell actually match it?
```

#### This command conjures mystical symbols:

✓: Aligned with the cosmic order.
✗: Venturing into unknown territories.
!=: Parallel dimensions detected!

### list
Take a galactic tour of all available environments, with the one you currently
inhabit marked by a `*`:

```bash
multiverse list
```

### lint-config
Check every universe the way `use` would, without touching your environment.
Useful once a config grows past the point where you visit every universe.

```bash
multiverse lint-config                    # the config it would otherwise use
multiverse lint-config candidate.yaml     # a file before you install it
multiverse lint-config --strict           # warnings count as errors too
```

Errors mean a universe cannot be used at all — a cycle in `extends` or in
interpolation, a variable name bash cannot take, a variable under the reserved
`MULTIVERSE_` prefix. Warnings mean it works but probably does not do what you
meant: a value referring to a name nothing defines expands to the empty string
in silence, which is how a typo hides.

```console
$ multiverse lint-config
config: /home/you/.config/multiverse/config.yaml

prod: warning: $REGIN is defined neither by this universe nor by the environment, so it expands to nothing (used by API_URL)

12 universes, 0 errors, 1 warning
```

It exits non-zero when there are errors, so it works from a pre-commit hook or
from CI. `--quiet` drops the report and leaves just the exit code.

### init
Print the bash integration described above.

```bash
multiverse init bash
```

## Contributing to the Multiverse
Interested in shaping the cosmos? Pull requests are the wormholes we love! Jump in and let's explore new realms of possibilities together.

```bash
go build && go test ./...
```

## License
Crafted by @alarreine. Licensed under MIT - the license of the cosmic explorers!
