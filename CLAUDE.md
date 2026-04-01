<!-- GSD:project-start source:PROJECT.md -->
## Project

**GitZen**

GitZen is a TUI (Terminal User Interface) Git client built with Go and Bubble Tea, inspired by lazygit. It provides an interactive terminal interface for Git operations with multiple panes for status, files, branches, commits, and stash management. Users can perform common Git operations like staging, committing, pushing, pulling, and branch management through keyboard shortcuts in a clean, organized interface.

**Core Value:** Users can perform Git operations faster and more intuitively through a visual terminal interface without memorizing complex Git commands.

### Constraints

- **Tech stack**: Go 1.24+, Bubble Tea framework, existing Git command execution pattern — maintain consistency with current architecture
- **Performance**: Background fetching must not block UI interactions — use existing async command pattern
- **Safety**: Only fetch when working directory is clean — prevent disrupting user's work
- **Compatibility**: Must work across Linux, macOS, Windows — leverage existing cross-platform support
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.24.0 - Main application language
- Shell script - Installation scripts (`install.sh`, `install.ps1`)
- YAML - CI/CD configuration and GoReleaser config
- Makefile - Build automation
## Runtime
- Go 1.24.0 runtime
- Cross-platform: Linux, macOS, Windows (amd64/arm64)
- Go Modules (go.mod/go.sum)
- Lockfile: `go.sum` present
## Frameworks
- Bubble Tea v1.3.10 - TUI framework for terminal user interfaces
- Lipgloss v1.1.0 - Terminal styling and layout
- Bubbles v0.21.0 - Pre-built UI components (viewport, textinput)
- Go built-in testing framework
- Race detector enabled in CI
- GoReleaser v2 - Automated release pipeline
- Make - Build automation (`Makefile`)
- Nix flake - Development environment (`flake.nix`)
## Key Dependencies
- `github.com/charmbracelet/bubbletea` v1.3.10 - Core TUI framework
- `github.com/charmbracelet/lipgloss` v1.1.0 - Terminal styling engine
- `github.com/charmbracelet/bubbles` v0.21.0 - UI components
- `github.com/charmbracelet/x/ansi` v0.10.1 - ANSI color support
- `github.com/atotto/clipboard` v0.1.4 - System clipboard integration
- `golang.org/x/sys` v0.36.0 - System-level operations
- `github.com/muesli/termenv` v0.16.0 - Terminal environment detection
## Configuration
- `GITZEN_LOG` - Log file path for debugging
- `GITZEN_DEBUG` - Enable debug logging
- `CGO_ENABLED=0` - Static linking (no C dependencies)
- `go.mod` - Module dependencies
- `.goreleaser.yml` - Release configuration
- `Makefile` - Build targets and automation
- `flake.nix` - Nix development shell
## Platform Requirements
- Go 1.24.0+
- Git (for repository operations)
- Optional: Nix package manager for dev environment
- No runtime dependencies (statically linked binary)
- Git repository for operation
- Terminal with ANSI color support
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- Package main: `main.go` (entry point)
- Implementation files: `lowercase.go` (e.g., `git.go`, `model.go`)
- Test files: `*_test.go` (e.g., `parse_hunk_test.go`)
- Interface pattern: `{domain}.go` (descriptive, not abbreviated)
- PascalCase for exported functions: `New()`, `DetectRepoRoot()`, `StatusPorcelainZ()`
- camelCase for private functions: `parseHunk()`, `atoi()`, `reverseHunk()`
- Constructor pattern: `New{Type}()` (e.g., `NewStatusPane()`, `NewLogger()`)
- camelCase for local variables: `repoRoot`, `branchName`, `logPath`
- PascalCase for exported fields: `RepoRoot`, `OldStart`, `NewLines`
- Private fields with receiver prefix: `m.focus`, `p.repoName`, `l.enabled`
- PascalCase for exported types: `Runner`, `Branch`, `Logger`, `StatusPane`
- Descriptive names: `Theme`, `Styles`, `Layout` (not abbreviated)
- Interface naming: No "I" prefix, descriptive nouns
## Code Style
- Standard `go fmt` with `goimports` integration
- Makefile target: `make fmt` runs both `go fmt` and `goimports -w .`
- Built-in Go tools: `go vet` via `make vet`
- Combined linting: `make lint` runs both `fmt` and `vet`
- No external linters like golangci-lint configured
## Import Organization
- `tea` for `github.com/charmbracelet/bubbletea`
- No other aliases used consistently
## Error Handling
- Package-level error variables: `var ErrGitNotFound = errors.New("git not found")`
- Error wrapping with fmt: `fmt.Errorf("cannot create log directory: %w", err)`
- Early return pattern: Check error immediately after operation
- Contextual error messages: Include operation context in error text
## Logging
- Singleton pattern: `logger.Get().Info("message")`
- Levels: `Debug()`, `Info()`, `Warn()`, `Error()`
- Format strings: `l.Info("hello %s", "world")`
- Thread-safe with mutex protection
## Comments
- Package-level documentation: `// Package logger cung cấp cơ chế ghi log...`
- Exported functions: Vietnamese comments explaining purpose
- Complex logic: Inline comments for business logic
- Type definitions: Comments on struct fields for clarity
- Not applicable (Go codebase)
- Go doc comments follow standard conventions
## Function Design
- Use structs for complex configuration: `app.Options{RepoPath: *repoFlag}`
- Receiver methods preferred: `(r Runner) StatusPorcelainZ()`
- Context as first parameter: `func runWithContext(ctx context.Context, ...)`
- Error as last return value: `(string, error)` or `error`
- Named return values for complex functions: Not commonly used
- Nil checks before using returned values
## Module Design
- Constructor functions: `NewStatusPane()`, `New()` for main types
- Interface methods: All public methods are PascalCase
- Package constants: `DefaultCmdTimeout`, `ErrNotARepository`
- Not applicable (Go doesn't use barrel exports)
- Each package focuses on single responsibility
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## Pattern Overview
- Clean separation between UI components and business logic
- Event-driven architecture with centralized state management
- Layered design with clear dependency boundaries
- Command-response pattern for asynchronous git operations
## Layers
- Purpose: Application bootstrap and argument parsing
- Location: `cmd/gitzen`
- Contains: Main function, version handling, uninstall logic
- Depends on: internal/app package
- Used by: OS (executable entry point)
- Purpose: Core application orchestration and model management
- Location: `internal/app`
- Contains: Main model, command handlers, key bindings, application runner
- Depends on: internal/components, internal/git, internal/ui, internal/logger
- Used by: cmd/gitzen main
- Purpose: UI component implementations and rendering logic
- Location: `internal/components`
- Contains: Individual pane components, modal, diff viewers, base pane abstraction
- Depends on: internal/ui, internal/git (for data structures)
- Used by: internal/app
- Purpose: Git command execution and output parsing
- Location: `internal/git`
- Contains: Git runner, parsers for status/log/hunks, repository detection
- Depends on: internal/limits, standard library
- Used by: internal/app, internal/components
- Purpose: Layout calculations, styling, and theme definitions
- Location: `internal/ui`
- Contains: Layout engine, theme/styles, keymap definitions
- Depends on: Charm libraries (lipgloss)
- Used by: internal/app, internal/components
- Purpose: Cross-cutting concerns and utilities
- Location: `internal/logger`, `internal/limits`, `internal/tui`
- Contains: Logging infrastructure, configuration limits, diff coloring
- Depends on: Standard library only
- Used by: Various layers as needed
## Data Flow
- Centralized in main model (`internal/app/model.go`)
- Components hold display state, main model holds application state
- Data flows down to components, events flow up to main model
## Key Abstractions
- Purpose: Common functionality for all UI panes
- Examples: `internal/components/pane.go`
- Pattern: Template pattern with interface-based polymorphism
- Purpose: Unified interface for git command execution
- Examples: `internal/git/git.go`
- Pattern: Command pattern with timeout and error handling
- Purpose: Standardized pane behavior and rendering
- Examples: All files in `internal/components/`
- Pattern: Interface segregation with common base implementation
- Purpose: Responsive terminal layout management
- Examples: `internal/ui/layout.go`
- Pattern: Strategy pattern with accordion-style expansion
## Entry Points
- Location: `cmd/gitzen/main.go`
- Triggers: Command-line execution
- Responsibilities: Argument parsing, application launch, cleanup
- Location: `internal/app/run.go`
- Triggers: Called from main entry point
- Responsibilities: Environment setup, git validation, TUI program launch
- Location: `internal/app/model.go` (NewModel, Init)
- Triggers: Called from application runner
- Responsibilities: Component creation, initial data loading
## Error Handling
- Git command errors: Captured and displayed in modal dialogs
- System errors: Logged and shown as status messages
- Validation errors: Prevented at input level with immediate feedback
- Timeouts: Configured per operation type with graceful degradation
## Cross-Cutting Concerns
<!-- GSD:architecture-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd:quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd:debug` for investigation and bug fixing
- `/gsd:execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd:profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
