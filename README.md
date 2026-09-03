# atty

TUI window switcher for Alacritty on macOS. List, search, and raise Alacritty windows from the command line.

## Features

- **List Alacritty windows**: View all open terminal tabs/named panes
- **Keyboard navigation**: Use `j`/`k` or `↑`/`↓` to navigate
- **Filter windows**: Type `/` to enable filtering mode, type patterns separated by `/` for combined results
- **Raise & switch**: Press `enter` to raise the selected window; the window `atty` was launched from is closed automatically
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

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j` `↓` | Move cursor down |
| `k` `↑` | Move cursor up |
| `/` | Enable filter mode |
| `space` | Combine filter terms (same as `/`) |
| `enter` | Raise selected Alacritty window |
| `esc` | Exit filter mode / cancel command |
| `backspace` | Delete character in filter |
| `q` `Ctrl+C` | Quit |

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

![demo](demo.gif)

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
