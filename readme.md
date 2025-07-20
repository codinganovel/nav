# 🧭 nav
### Universal File Navigator
or instead use :
	cd long/nested/path/that/you/have/to/type/perfectly

A simple, fast terminal-based file navigator that opens directories and files in new terminal windows. Navigate your filesystem with keyboard shortcuts and launch new terminal sessions instantly.

## 🚀 Quick Start

```bash
# Navigate current directory
nav

# Navigate specific directory  
nav /path/to/directory

# Show help
nav --help
```

## 📦 Installation

```bash
go build -o nav
mv nav /usr/local/bin/  # Optional: add to PATH
```

## ⌨️ Keybindings

| Key | Action |
|-----|--------|
| `↑`/`↓` | Navigate up/down through items |
| `Enter` | Open directory / Open file's parent directory in terminal |
| `o` | Open selected item in new terminal window |
| `/` | Search (type to filter, `Esc` to exit) |
| `q` | Quit |

## 🎯 Smart Terminal Detection

`nav` automatically detects your terminal with this priority:

1. **`$TERMINAL` environment variable** (highest priority)
2. **`$TERM_PROGRAM` detection** (iTerm2, Ghostty, Wezterm, etc.)
3. **OS defaults** (Terminal.app, gnome-terminal, cmd)

### Examples:
```bash
# Use your preferred terminal
export TERMINAL="open -a Ghostty"
export TERMINAL="wezterm start --cwd"
export TERMINAL="alacritty --working-directory"
```

## ✨ Features

- **Fast & Responsive**: Instant startup, smooth navigation
- **Tree-Style Display**: Clean visual hierarchy with `├──` and `└──`
- **Hidden Files**: Shows all files including `.hidden` files
- **Real-Time Search**: Filter files as you type with `/`
- **Cross-Platform**: macOS, Linux, Windows support
- **Smart Sorting**: Directories first, then files (alphabetical)
- **Error Handling**: User-friendly messages for permission and access issues
- **Smart Truncation**: Intelligently truncates long filenames while preserving extensions

## 🖥️ Interface

```
/Users/sam/Documents/coding/nav

├── ../
├── main.go
├── navigator.go
├── navigator_test.go
├── go.mod
└── README.md

[5 items] • ↑↓ navigate • Enter open • o open in terminal • q quit • / search
```

## 🔧 How It Works

- **Directories**: Navigate into them with `Enter`, or open in new terminal with `o`
- **Files**: `Enter` opens the file's parent directory in a new terminal
- **Search**: Press `/` to filter items, `Esc` to clear search
- **Terminal Spawning**: Non-blocking - nav keeps running after opening terminals
- **Error Recovery**: Automatically handles permission issues and path problems

## 📋 Requirements

- Go 1.21+
- Works in most terminal emulators
- Supports tcell-compatible terminals

## 🎨 Design Philosophy

`nav` follows the Unix philosophy: do one thing well. It's a focused tool for quick filesystem navigation and terminal launching, without complex features that would slow it down or complicate its use.

## 📄 License

under ☕️, check out [the-coffee-license](https://github.com/codinganovel/The-Coffee-License)

I've included both licenses with the repo, do what you know is right. The licensing works by assuming your operating under good faith.
