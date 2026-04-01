# Codebase Structure

**Analysis Date:** 2026-04-01

## Directory Layout

```
gitzen/
├── cmd/                    # Entry points and command-line interface
│   └── gitzen/            # Main executable
├── internal/              # Internal application code
│   ├── app/               # Application orchestration layer
│   ├── components/        # UI component implementations
│   ├── git/               # Git command execution and parsing
│   ├── limits/            # Configuration and limits
│   ├── logger/            # Logging infrastructure
│   ├── tui/               # Terminal UI utilities
│   └── ui/                # UI foundation (layout, themes, styles)
├── .github/               # GitHub workflows and configuration
├── .planning/             # Project planning and documentation
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile              # Build automation
├── README.md             # Project documentation
├── flake.nix             # Nix package definition
└── install.sh/.ps1       # Installation scripts
```

## Directory Purposes

**cmd/gitzen/:**
- Purpose: Main application entry point
- Contains: main.go with CLI argument parsing and application bootstrap
- Key files: `main.go`

**internal/app/:**
- Purpose: Core application logic and orchestration
- Contains: Main model, command handlers, key bindings, application runner
- Key files: `model.go` (main orchestrator), `run.go` (application launcher), `cmds.go` (command definitions), `keys.go` (key handling)

**internal/components/:**
- Purpose: Reusable UI components and pane implementations
- Contains: Individual pane components with common base functionality
- Key files: `pane.go` (base abstraction), `files.go`, `branches.go`, `commits.go`, `stash.go`, `status.go`, `diffview.go`, `splitdiff.go`, `hunkview.go`, `cmdlog.go`, `modal.go`

**internal/git/:**
- Purpose: Git command execution and output parsing
- Contains: Git runner, command wrappers, parsers for various git outputs
- Key files: `git.go` (main runner), `parse_status.go`, `parse_log.go`, `parse_hunk.go`

**internal/ui/:**
- Purpose: UI foundation layer with layout and styling
- Contains: Layout calculations, theme definitions, style constants
- Key files: `layout.go` (layout engine), `theme.go` (styling), `keymap.go` (key definitions)

**internal/logger/:**
- Purpose: Structured logging infrastructure
- Contains: Logger initialization and management
- Key files: `logger.go`

**internal/limits/:**
- Purpose: Configuration constants and system limits
- Contains: Timeout values, size limits, performance constraints
- Key files: `limits.go`

**internal/tui/:**
- Purpose: Terminal UI utilities and helpers
- Contains: Diff syntax highlighting, color schemes
- Key files: `diffcolor.go`

## Key File Locations

**Entry Points:**
- `cmd/gitzen/main.go`: Application entry point and CLI handling

**Configuration:**
- `go.mod`: Go module dependencies and version
- `Makefile`: Build targets and automation
- `.goreleaser.yml`: Release configuration

**Core Logic:**
- `internal/app/model.go`: Main application state and orchestration
- `internal/app/run.go`: Application lifecycle management
- `internal/git/git.go`: Git command interface

**Testing:**
- `internal/git/parse_*_test.go`: Unit tests for git parsers
- `internal/logger/logger_test.go`: Logger functionality tests

## Naming Conventions

**Files:**
- Go source files: snake_case (e.g., `parse_log.go`, `split_diff.go`)
- Test files: `*_test.go` suffix
- Configuration: descriptive names (e.g., `Makefile`, `go.mod`)

**Directories:**
- Single word, lowercase where possible
- Descriptive of contained functionality

## Where to Add New Code

**New UI Component:**
- Primary code: `internal/components/[component_name].go`
- Integration: Update `internal/app/model.go` to include component
- Layout: Update `internal/ui/layout.go` if new pane type needed

**New Git Operation:**
- Implementation: Add method to `internal/git/git.go`
- Parser: Create `internal/git/parse_[operation].go` if needed
- Integration: Add command and message types to `internal/app/cmds.go`

**New Feature:**
- Primary code: Appropriate layer based on responsibility
- Tests: Co-located `*_test.go` files
- Integration: Update `internal/app/model.go` for orchestration

**Utilities:**
- Shared helpers: Appropriate internal package based on domain
- UI utilities: `internal/ui/` or `internal/tui/`
- Git utilities: `internal/git/`

## Special Directories

**.github/:**
- Purpose: GitHub Actions workflows and repository configuration
- Generated: No
- Committed: Yes

**.planning/:**
- Purpose: Project planning documents and codebase analysis
- Generated: Partially (codebase mapping documents)
- Committed: Yes

**vendor/ (if present):**
- Purpose: Vendored Go dependencies
- Generated: Yes (by `go mod vendor`)
- Committed: Depends on project policy

## Module Organization

**Package Structure:**
- Clean separation between layers using Go's internal package visibility
- Each package has a single, well-defined responsibility
- Dependencies flow in one direction (no circular imports)

**Import Patterns:**
- Standard library imports first
- Third-party dependencies second
- Internal packages last
- Clear separation between import groups

**Interface Design:**
- Interfaces defined where used, not where implemented
- Small, focused interfaces following Go idioms
- Common behavior abstracted through base types (e.g., `BasePane`)

---

*Structure analysis: 2026-04-01*