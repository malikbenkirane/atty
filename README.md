# atty

TUI window switcher for Alacritty on macOS. List, search, and raise Alacritty windows from the command line.

![demo](demo.gif)

## Features

- **List Alacritty windows**: View all open terminal tabs/named panes
- **Keyboard navigation**: Use `j`/`k` or `↑`/`↓` to navigate
- **Filter windows**: Type patterns separated by `/` for combined results
- **Raise & switch**: Press `enter` to raise the selected window; the window `atty` was launched from is closed automatically (disable with `-no-close`)
- **Recency ordering**: Window selection history is stored in a local SQLite cache, and windows you used recently are listed first
- **Clean interface**: Minimal TUI built with bubbletea

## Installation

Clone the repository:

```bash
git clone https://github.com/malikbenkirane/atty.git
cd atty
```

Install:

```bash
go install
```

## CLI Flags

| Flag | Effect |
|------|--------|
| `-no-close` | Keep the window `atty` was launched from open after raising the selected window |
| `-info` | Print the SQLite cache location and exit without launching the TUI |

## Keyboard Shortcuts

`atty` starts directly in filter mode, so you can type right away.

| Key | Action |
|-----|--------|
| `j` `↓` `ctrl+n` | Move cursor down |
| `k` `↑` `ctrl+p` | Move cursor up |
| `/` | Resume filter mode (after cancelling) |
| `space` | Combine filter terms (same as `/`) |
| `enter` | Raise selected Alacritty window |
| `esc` | Quit; exit filter mode / cancel filter when filtering |
| `backspace` | Delete character in filter |
| `q` `ctrl+c` | Quit (in filter mode, cancel filter) |

`ctrl+n` moves down and `ctrl+p` moves up in all modes.

In filter mode, `esc` cancels the filter and switches to navigation mode; `ctrl+c` in filter mode also cancels the filter.

`space` combines filter terms only while filtering; in navigation mode it has no effect.

## Usage Example

```bash
# Run the TUI
atty

# You'll see your Alacritty windows listed
# > bash
# > tmux: vim
# > zsh
# > npm dev

# Type to filter and search terms
# filter: "bash" and ":vim"

# Press enter to select and type enter to bring that window to the foreground
```

## Configuration

Set automatic window titles based on directory using shell hooks:

### Bash

Add to `~/.bashrc` or `~/.bash_profile`:

```bash
dl() {
  # Your existing dl function here
}
dl -title $PWD
```

### Zsh

Add to `~/.zshrc`:

```zsh
autoload -U add-zsh-hook

title_hook() {
  dl -title $PWD
}

add-zsh-hook precmd title_hook
```

Both shells support hooks (`precmd` in zsh and bash traps) to set titles automatically. We use zsh on macOS as it's the default shell recommended by Apple.

### Alacritty

Add this to your Alacritty configuration file (~/.config/alacritty/alacritty.toml):

```toml
[general]
ipc_socket = true

[keyboard]
bindings = [
  { key = "/", mods = "Command", command = {
    program = "/Applications/Alacritty.app/Contents/MacOS/alacritty",
    args = ["msg", "create-window", "-e", "/path/to/atty"] }
  },
]
```

**Notes:**
- Enable `ipc_socket = true` in `[general]` - required for the `msg` command to work
- Use an absolute path for the `atty` program (Alacritty doesn't expand tildes)

## How It Works

On macOS, `atty` uses AppleScript via `osascript` to:
1. Retrieve the list of Alacritty window names
2. Raise a window using the `AXRaise` accessibility action
3. Close the window it was launched from via its close button

The TUI is built with Bubbletea, providing a smooth interactive experience.

`atty` tags its own Alacritty window with a random title, so it can hide itself from the list. When you raise another window, `atty` closes the window it was launched from (via its close button), so you land directly in the selected window.

## Requirements

- macOS
- [Alacritty terminal emulator](https://alacritty.org/)
- Go 1.26.2+

## License

See the LICENSE file for details.
