# pin.sh

Shell function for pinning &amp; viewing files.

## Quickstart

The install script downloads the `pin.sh` file defining the shell function to the `~/.local/share/pin/` directory, and appends a line sourcing this file to your `~/.bashrc` or `~/.zshrc`.

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/wrnkt/pin.sh/refs/heads/main/install.sh)
source ~/.bashrc
```

### Usage
```sh
pin -p ~/documents/foo      # pin a file
pin                         # view pinned file w/ pager (defaults to cat or less)
pin -c                      # clear pinned

# You can also use digits 1-9 to pin additional files. e.g.

pin -p3 ~/documents/bar     # pin to slot 3
pin -3                       # view pin
pin -c3                      # clear slot

pin --list                  # show all pins
pin -h                      # show help

# set the pager
pin --pager bat
pin --clear-pager

# complete uninstall; deletes data and removes all shell functions
pin --uninstall
```

All data is stored in `~/.local/share/pin/pins.data`
