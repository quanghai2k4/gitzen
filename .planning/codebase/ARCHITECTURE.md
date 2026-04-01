# Architecture

**Analysis Date:** 2026-04-01

## Pattern Overview

**Overall:** Component-based TUI with MVC pattern using Bubble Tea framework

**Key Characteristics:**
- Clean separation between UI components and business logic
- Event-driven architecture with centralized state management
- Layered design with clear dependency boundaries
- Command-response pattern for asynchronous git operations

## Layers

**Main Entry Layer:**
- Purpose: Application bootstrap and argument parsing
- Location: `cmd/gitzen`
- Contains: Main function, version handling, uninstall logic
- Depends on: internal/app package
- Used by: OS (executable entry point)

**Application Layer:**
- Purpose: Core application orchestration and model management
- Location: `internal/app`
- Contains: Main model, command handlers, key bindings, application runner
- Depends on: internal/components, internal/git, internal/ui, internal/logger
- Used by: cmd/gitzen main

**Components Layer:**
- Purpose: UI component implementations and rendering logic
- Location: `internal/components`
- Contains: Individual pane components, modal, diff viewers, base pane abstraction
- Depends on: internal/ui, internal/git (for data structures)
- Used by: internal/app

**Git Integration Layer:**
- Purpose: Git command execution and output parsing
- Location: `internal/git`
- Contains: Git runner, parsers for status/log/hunks, repository detection
- Depends on: internal/limits, standard library
- Used by: internal/app, internal/components

**UI Foundation Layer:**
- Purpose: Layout calculations, styling, and theme definitions
- Location: `internal/ui`
- Contains: Layout engine, theme/styles, keymap definitions
- Depends on: Charm libraries (lipgloss)
- Used by: internal/app, internal/components

**Support Services:**
- Purpose: Cross-cutting concerns and utilities
- Location: `internal/logger`, `internal/limits`, `internal/tui`
- Contains: Logging infrastructure, configuration limits, diff coloring
- Depends on: Standard library only
- Used by: Various layers as needed

## Data Flow

**Application Startup Flow:**

1. `cmd/gitzen/main.go` parses arguments and calls `app.Run()`
2. `app.Run()` initializes logger and validates git environment
3. `app.NewModel()` creates main model with all components
4. Bubble Tea program starts with initial data loading commands
5. Components receive data and render initial UI

**User Interaction Flow:**

1. User input captured by Bubble Tea framework
2. Main model `Update()` method receives input messages
3. Key handling delegates to appropriate component or executes git commands
4. Git commands executed via `internal/git.Runner`
5. Results parsed and sent as messages back to components
6. Components update their state and trigger re-renders

**State Management:**
- Centralized in main model (`internal/app/model.go`)
- Components hold display state, main model holds application state
- Data flows down to components, events flow up to main model

## Key Abstractions

**BasePane:**
- Purpose: Common functionality for all UI panes
- Examples: `internal/components/pane.go`
- Pattern: Template pattern with interface-based polymorphism

**Git Runner:**
- Purpose: Unified interface for git command execution
- Examples: `internal/git/git.go`
- Pattern: Command pattern with timeout and error handling

**Component Interface:**
- Purpose: Standardized pane behavior and rendering
- Examples: All files in `internal/components/`
- Pattern: Interface segregation with common base implementation

**Layout System:**
- Purpose: Responsive terminal layout management
- Examples: `internal/ui/layout.go`
- Pattern: Strategy pattern with accordion-style expansion

## Entry Points

**Main Entry Point:**
- Location: `cmd/gitzen/main.go`
- Triggers: Command-line execution
- Responsibilities: Argument parsing, application launch, cleanup

**Application Runner:**
- Location: `internal/app/run.go`
- Triggers: Called from main entry point
- Responsibilities: Environment setup, git validation, TUI program launch

**Model Initialization:**
- Location: `internal/app/model.go` (NewModel, Init)
- Triggers: Called from application runner
- Responsibilities: Component creation, initial data loading

## Error Handling

**Strategy:** Layered error handling with user feedback

**Patterns:**
- Git command errors: Captured and displayed in modal dialogs
- System errors: Logged and shown as status messages
- Validation errors: Prevented at input level with immediate feedback
- Timeouts: Configured per operation type with graceful degradation

## Cross-Cutting Concerns

**Logging:** Structured logging via `internal/logger` with configurable output
**Validation:** Git repository detection and command validation at startup
**Authentication:** Relies on system git configuration and credentials

---

*Architecture analysis: 2026-04-01*
