# GitZen

A TUI Git client inspired by [lazygit](https://github.com/jesseduffield/lazygit), written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Release](https://img.shields.io/github/v/release/quanghai2k4/gitzen?include_prereleases)

## Features

- Interactive TUI with multiple panes (Status, Files, Branches, Commits, Stash)
- Git operations: stage, unstage, commit, push, pull, fetch
- Diff viewer with syntax highlighting
- Branch management (create, checkout, delete, merge)
- Commit history with split diff view
- Reflog support
- Stash management
- Modal dialogs for commit messages

## Installation

### From Release

Download the latest release for your platform from [Releases](https://github.com/quanghai2k4/gitzen/releases).

```bash
# Linux/macOS
tar -xzf gitzen_*_linux_amd64.tar.gz
sudo mv gitzen /usr/local/bin/

# Windows
# Extract gitzen_*_windows_amd64.zip and add to PATH
```

### From Source

```bash
# Clone repository
git clone https://github.com/quanghai2k4/gitzen.git
cd gitzen

# Build
go build -o gitzen ./cmd/gitzen

# Or using Make
make build
```

### Using Go Install

```bash
go install github.com/quanghai2k4/gitzen/cmd/gitzen@latest
```

## Usage

```bash
# Run in current directory
gitzen

# Run with specific repository
gitzen --repo /path/to/repo
```

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `h` / `l` | Switch between panes |
| `j` / `k` | Move up/down in list |
| `Tab` | Next pane |
| `Shift+Tab` | Previous pane |
| `q` | Quit |

### Git Operations

| Key | Action |
|-----|--------|
| `Space` | Stage/Unstage file |
| `a` | Stage all files |
| `c` | Commit |
| `p` | Push |
| `P` | Pull |
| `f` | Fetch |
| `Enter` | View diff / Checkout branch |

### Branch Operations

| Key | Action |
|-----|--------|
| `n` | New branch |
| `d` | Delete branch |
| `m` | Merge branch |
| `Enter` | Checkout branch |

### Stash Operations

| Key | Action |
|-----|--------|
| `s` | Stash changes |
| `Space` | Apply stash |
| `d` | Drop stash |

## Development

### Requirements

- Go 1.24+
- Git

### Build Commands

```bash
# Enter dev shell (with Nix)
nix develop

# Build
make build

# Build for all platforms
make build-all

# Run tests
make test

# Lint
make lint

# Clean
make clean
```

### Project Structure

```
cmd/gitzen/main.go       # Entry point, CLI flags
internal/
  app/
    run.go               # tea.Program setup
    model.go             # Bubble Tea model, UI rendering
    keys.go              # Keyboard handling
    cmds.go              # tea.Cmd factories for async git ops
  git/
    git.go               # Git Runner (exec-based)
    parse_log.go         # Parse git log output
    parse_status.go      # Parse git status --porcelain
  tui/
    diffcolor.go         # Diff syntax highlighting
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [lazygit](https://github.com/jesseduffield/lazygit) - Inspiration for the UI/UX
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
