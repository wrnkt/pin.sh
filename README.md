# pin.sh

Pin files to numbered slots for quick access.

## Installing

### Go Binary

Requires Go 1.21+. Builds a standalone binary with no shell integration needed.

```sh
git clone https://github.com/wrnkt/pin.sh.git
cd pin.sh
make install-bin
```

This installs the binary to `~/.local/bin/pin`. If that directory isn't in your PATH yet:

```sh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Shell Function

Downloads `pin.sh` and sources it from your shell rc file.

```sh
mkdir -p ~/.local/share/pin

curl -o ~/.local/share/pin/pin.sh \
  https://raw.githubusercontent.com/wrnkt/pin.sh/main/pin.sh

echo 'source ~/.local/share/pin/pin.sh' >> ~/.bashrc
source ~/.bashrc
```

For zsh, replace `~/.bashrc` with `~/.zshrc`.

Alternatively, use the installer script which handles shell detection automatically:

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/wrnkt/pin.sh/refs/heads/main/install.sh)
source ~/.bashrc
```

## Usage

```sh
pin -p ~/documents/foo      # pin a file to slot 0
pin -0                      # open slot 0 in pager

pin -p3 ~/documents/bar     # pin to slot 3
pin -3                      # open slot 3
pin -c3                     # clear slot 3

pin --list                  # show all slots and current pager
pin -c                      # clear all slots
pin -h                      # show help

pin --pager bat             # set preferred pager
pin --pager-clear           # reset to auto-detected pager

pin --uninstall             # remove all data (~/.local/share/pin/)
```

Supported pagers: `bat`, `less`, `more`, `most`, `cat`, `pg` (auto-detects `bat` → `less` → `cat`).

All data is stored in `~/.local/share/pin/pins.data`.
