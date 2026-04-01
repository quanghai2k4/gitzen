# Coding Conventions

**Analysis Date:** 2026-04-01

## Naming Patterns

**Files:**
- Package main: `main.go` (entry point)
- Implementation files: `lowercase.go` (e.g., `git.go`, `model.go`)
- Test files: `*_test.go` (e.g., `parse_hunk_test.go`)
- Interface pattern: `{domain}.go` (descriptive, not abbreviated)

**Functions:**
- PascalCase for exported functions: `New()`, `DetectRepoRoot()`, `StatusPorcelainZ()`
- camelCase for private functions: `parseHunk()`, `atoi()`, `reverseHunk()`
- Constructor pattern: `New{Type}()` (e.g., `NewStatusPane()`, `NewLogger()`)

**Variables:**
- camelCase for local variables: `repoRoot`, `branchName`, `logPath`
- PascalCase for exported fields: `RepoRoot`, `OldStart`, `NewLines`
- Private fields with receiver prefix: `m.focus`, `p.repoName`, `l.enabled`

**Types:**
- PascalCase for exported types: `Runner`, `Branch`, `Logger`, `StatusPane`
- Descriptive names: `Theme`, `Styles`, `Layout` (not abbreviated)
- Interface naming: No "I" prefix, descriptive nouns

## Code Style

**Formatting:**
- Standard `go fmt` with `goimports` integration
- Makefile target: `make fmt` runs both `go fmt` and `goimports -w .`

**Linting:**
- Built-in Go tools: `go vet` via `make vet`
- Combined linting: `make lint` runs both `fmt` and `vet`
- No external linters like golangci-lint configured

## Import Organization

**Order:**
1. Standard library imports (e.g., `"fmt"`, `"strings"`, `"context"`)
2. Third-party imports (e.g., `tea "github.com/charmbracelet/bubbletea"`)
3. Local imports (e.g., `"gitzen/internal/git"`, `"gitzen/internal/ui"`)

**Path Aliases:**
- `tea` for `github.com/charmbracelet/bubbletea`
- No other aliases used consistently

## Error Handling

**Patterns:**
- Package-level error variables: `var ErrGitNotFound = errors.New("git not found")`
- Error wrapping with fmt: `fmt.Errorf("cannot create log directory: %w", err)`
- Early return pattern: Check error immediately after operation
- Contextual error messages: Include operation context in error text

## Logging

**Framework:** Custom logger in `internal/logger/logger.go`

**Patterns:**
- Singleton pattern: `logger.Get().Info("message")`
- Levels: `Debug()`, `Info()`, `Warn()`, `Error()`
- Format strings: `l.Info("hello %s", "world")`
- Thread-safe with mutex protection

## Comments

**When to Comment:**
- Package-level documentation: `// Package logger cung cấp cơ chế ghi log...`
- Exported functions: Vietnamese comments explaining purpose
- Complex logic: Inline comments for business logic
- Type definitions: Comments on struct fields for clarity

**JSDoc/TSDoc:**
- Not applicable (Go codebase)
- Go doc comments follow standard conventions

## Function Design

**Size:** Functions generally under 50 lines, largest around 100 lines

**Parameters:** 
- Use structs for complex configuration: `app.Options{RepoPath: *repoFlag}`
- Receiver methods preferred: `(r Runner) StatusPorcelainZ()`
- Context as first parameter: `func runWithContext(ctx context.Context, ...)`

**Return Values:** 
- Error as last return value: `(string, error)` or `error`
- Named return values for complex functions: Not commonly used
- Nil checks before using returned values

## Module Design

**Exports:** 
- Constructor functions: `NewStatusPane()`, `New()` for main types
- Interface methods: All public methods are PascalCase
- Package constants: `DefaultCmdTimeout`, `ErrNotARepository`

**Barrel Files:** 
- Not applicable (Go doesn't use barrel exports)
- Each package focuses on single responsibility

---

*Convention analysis: 2026-04-01*