# Agent Instructions for GitZen

GitZen is a TUI Git client (lazygit-inspired) written in Go using the Bubble Tea framework.

## Build Commands

```bash
nix develop                              # Enter dev shell (Go, gopls, gotools)
go build -o gitzen ./cmd/gitzen          # Build binary
go run ./cmd/gitzen                      # Run without building
./gitzen --repo /path/to/repo            # Run with specific repo
```

## Test Commands

```bash
go test ./...                                    # Run all tests
go test -v ./...                                 # Verbose output
go test -v -run TestFunctionName ./path/to/pkg  # Single test by name
go test -v ./internal/git/...                    # Tests in specific package
go test -race ./...                              # With race detector
go test -cover ./...                             # With coverage
```

## Lint/Format Commands

```bash
go fmt ./...        # Format all files
goimports -w .      # Fix imports
go vet ./...        # Static analysis
go mod tidy         # Check module tidiness
```

## Project Structure

```
cmd/gitzen/main.go       # Entry point, CLI flags, exit codes
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

## Code Style Guidelines

### Import Organization

```go
import (
    "strings"                              // stdlib
    tea "github.com/charmbracelet/bubbletea"  // third-party
    "gitzen/internal/git"                  // internal
)
```

### Naming Conventions

| Type | Convention | Examples |
|------|------------|----------|
| Exported types | PascalCase | `Runner`, `FileItem` |
| Unexported types | camelCase | `model`, `pane` |
| Message types | camelCase + Msg | `statusLoadedMsg`, `errMsg` |
| Command functions | camelCase + Cmd | `loadStatusCmd()` |
| Error sentinels | ErrPascalCase | `ErrGitNotFound` |

### Error Handling

```go
// Sentinel errors at package level
var ErrGitNotFound = errors.New("git not found")

// Early returns; in Bubble Tea commands, return errMsg(err.Error())
if err != nil {
    return nil, err
}
```

### Bubble Tea Patterns

```go
// Commands are factory functions returning tea.Cmd
func loadStatusCmd(r git.Runner) tea.Cmd {
    return func() tea.Msg {
        st, _ := r.StatusPorcelainZ()
        return statusLoadedMsg{Status: st}
    }
}

// Use tea.Batch() for parallel execution
return tea.Batch(loadStatusCmd(m.git), loadCommitsCmd(m.git))

// Handle messages with type switch
switch msg := msg.(type) {
case tea.WindowSizeMsg:
    m.resize()
case statusLoadedMsg:
    m.status = msg.Status
}
```

### Type Definitions

```go
// Enum-like constants with iota
type pane int
const (
    paneStatus pane = iota
    paneFiles
    paneBranches
)

// Message types as simple structs or type aliases
type statusLoadedMsg struct{ Status git.Status }
type errMsg string
```

### Git Operations

```go
// All git commands use exec with context timeout
ctx, cancel := context.WithTimeout(context.Background(), DefaultCmdTimeout)
defer cancel()
cmd := exec.CommandContext(ctx, "git", args...)
cmd.Dir = repoRoot
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | TUI/runtime error |
| 2 | Not a git repository |
| 3 | Git not found in PATH |

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework (Elm architecture)
- `github.com/charmbracelet/bubbles` - UI components (viewport, textinput)
- `github.com/charmbracelet/lipgloss` - Terminal styling/layout

## Testing Guidelines

- Place test files alongside source: `git_test.go` next to `git.go`
- Use table-driven tests for parsing functions
- Test function naming: `TestFunctionName(t *testing.T)`

## Common Tasks

### Adding a new git command

1. Add method to `git.Runner` in `internal/git/git.go`
2. Add message type in `internal/app/cmds.go`
3. Add command factory function in `internal/app/cmds.go`
4. Handle message in `model.Update()` in `internal/app/model.go`

### Adding a new keybinding

1. Add key handling in appropriate handler in `internal/app/keys.go`
2. Update info bar hints in `renderInfoBar()` in `internal/app/model.go`
