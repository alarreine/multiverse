# Multiverse: The CLI Time Machine for Environment Configurations

Welcome to **Multiverse**, the command-line tool that's not just a tool, but a portal to effortlessly navigate through the vast landscapes of environment configurations. Think of it as your trusty time machine, zipping you between parallel worlds of development, testing, and production.

## Inspiration
Multiverse draws creative inspiration from the powerful simplicity of [direnv](https://direnv.net/), a pioneering tool in the realm of environment switching. We owe a cosmic hat tip to [direnv](https://direnv.net/) for lighting the path and showing us the vast potential of environment management. While Multiverse embarks on its own interstellar journey, we acknowledge the groundbreaking work done by [direnv](https://direnv.net/) and hope to complement it in the universe of environment configuration tools.

## Features

- 🌌 **Navigate Through Configurations**: Like jumping through different realities, switch between your environments with ease.
- 🚀 **Temporal Precision**: Apply configurations for your current session, or make them permanent for future explorations.
- 🧙‍♂️ **Magical Incantations**: With `--persist`, you can make your settings echo through time and space, persisting them into the `.envrc` of your home directory.
- 🤫 **Whisper Mode**: Activate with --quiet to mute the chatter, and enjoy the quiet cosmos while navigating your configurations.

## Installation

Grab your universal translator (aka your terminal) and type the ancient runes:

```bash
# Replace with actual installation instructions
git clone git@github.com:alarreine/multiverse.git
cd multiverse
go build
```

After building Multiverse, you have a couple of options to ensure it's always within your arm's reach:

1. Add Multiverse to Your PATH:

Edit your shell's configuration file (like .bashrc or .zshrc) and append the path to the Multiverse binary. Replace /path/to/multiverse with the actual path to the binary:

```bash
export PATH="$PATH:/path/to/multiverse"
```

2. Install Multiverse Globally:

For more universal access, you can install Multiverse in a common directory like /usr/local/bin. This might require superuser access:

```bash
sudo install multiverse /usr/local/bin/
```

Now, with the power of Multiverse at your fingertips, you're ready to navigate the cosmic seas of configuration!

## Usage

### Global Flags
* `--env`, `-e`: Specify the universe (environment) you wish to enter. Ideal for quickly hopping between different settings.
* `--quiet`, `-q`: Mute the chatter, enjoy the quiet cosmos. Silence all log messages for a stealthier journey through your config multiverse.

### Apply SubCommand
Apply the fabric of your environment settings:

```bash
multiverse apply --env=<universe_name>
```

Remember, if you use --persist, don't forget to recite the spell source $HOME/.envrc in your .bashrc scrolls.

### Check SubCommand
Peer into the cosmic map of your environment variables:

```bash
multiverse check
```

### List SubCommand
Take a galactic tour of all available environments:

```bash
multiverse list
```

Get an overview of all the mystical environments within your multiverse.yaml, like stars shining in the night sky.

#### This command conjures mystical symbols:

✓: Aligned with the cosmic order.
✗: Venturing into unknown territories.
!=: Parallel dimensions detected!

## Contributing to the Multiverse
Interested in shaping the cosmos? Pull requests are the wormholes we love! Jump in and let's explore new realms of possibilities together.

## License
Crafted by @alarreine. Licensed under MIT - the license of the cosmic explorers!



