# Technology Stack

**Project:** GitZen Auto Fetch
**Researched:** 2026-04-01
**Confidence:** HIGH

## Recommended Stack

### Core Framework (Existing)
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.24+ | Runtime, concurrency | Existing codebase requirement, excellent context/goroutine support for background tasks |
| Bubble Tea | 1.3.10+ | TUI framework | Already integrated, provides tea.Cmd pattern for async operations |
| Lip Gloss | 1.1.0+ | TUI styling | Already integrated for consistent UI |

### Background Task Management
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| time.Ticker | std lib | Periodic scheduling | Standard Go approach for 30-minute intervals, integrates with context for cancellation |
| context.Context | std lib | Cancellation/timeout | Already used throughout Git layer, essential for safe background operations |
| sync.Mutex | std lib | State protection | Prevents race conditions between UI updates and background fetch |

### Configuration Management
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| gopkg.in/yaml.v3 | v3.0.1 | Config file parsing | For ~/.config/gitzen/config.yaml settings |
| os.UserConfigDir | std lib | Config directory | Cross-platform config path discovery |
| filepath.Join | std lib | Path construction | Safe config file path building |

### Working Directory Safety
| Component | Version | Purpose | When to Use |
|-----------|---------|---------|-------------|
| git status --porcelain | existing | Clean check | Before every auto fetch to ensure safety |
| Existing git.Runner | current | Git operations | Reuse existing patterns for fetch commands |
| NetworkTimeout | 30s | Fetch timeout | Already defined in limits package |

## Installation

```bash
# Add to go.mod (config parsing only)
go get gopkg.in/yaml.v3@v3.0.1

# All other components are standard library or already integrated
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| time.Ticker | time.After loop | Never - Ticker is more efficient for regular intervals |
| gopkg.in/yaml.v3 | github.com/BurntSushi/toml | If team strongly prefers TOML format |
| Context cancellation | Channel-based cancellation | Never - context is Go standard for cancellation |
| Existing tea.Cmd pattern | Direct goroutines | Never - breaks Bubble Tea architecture |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| cron libraries | Overkill for simple 30min intervals | time.Ticker |
| Complex config libraries | Adds unnecessary deps | Simple YAML parsing |
| Background goroutines without context | Can't be cancelled cleanly | Context-aware patterns |
| Direct UI updates from goroutines | Race conditions with Bubble Tea | tea.Cmd message passing |

## Integration Points with Existing Architecture

### Message Passing Pattern
```go
// New message types to add
type autoFetchStartedMsg struct{}
type autoFetchCompletedMsg struct{ Success bool, Error error }
type configLoadedMsg struct{ Config AutoFetchConfig }
```

### Configuration Structure
```go
type AutoFetchConfig struct {
    Enabled    bool          `yaml:"enabled"`
    Interval   time.Duration `yaml:"interval"`
    OnStartup  bool          `yaml:"on_startup"`
    SafetyMode bool          `yaml:"safety_mode"` // only when working dir clean
}
```

### Background Service Integration
```go
// Integrates with existing app.model via tea.Cmd
func (m model) startAutoFetch() tea.Cmd {
    return func() tea.Msg {
        // Uses existing git.Runner and context patterns
        return autoFetchService.Start(m.git, m.config)
    }
}
```

## Architecture Decisions

**Timer Management:**
- Use `time.NewTicker(config.Interval)` for regularity
- Stop ticker on app exit via context cancellation
- Reset ticker when config changes

**Safety Checks:**
- Call existing `git.StatusPorcelainZ()` before fetch
- Only proceed if no staged/unstaged changes detected
- Use existing `git.Fetch()` method with NetworkTimeout

**State Management:**
- Configuration stored in app.model
- Auto fetch state tracked via boolean flags
- All updates go through Bubble Tea message system

**Error Handling:**
- Network failures: retry with exponential backoff
- Git errors: surface via existing error modal system
- Config errors: fall back to defaults, log warning

## Version Compatibility

| Package | Compatible With | Notes |
|---------|------------------|-------|
| gopkg.in/yaml.v3@v3.0.1 | Go 1.18+ | Requires generics support |
| time.Ticker | Go 1.0+ | Standard library, always compatible |
| context.Context | Go 1.7+ | Well below Go 1.24 requirement |

## Sources

- [Go pkg.go.dev/time](https://pkg.go.dev/time) — Ticker patterns and context integration
- [Go pkg.go.dev/context](https://pkg.go.dev/context) — Background task cancellation patterns
- [Go blog context-and-structs](https://go.dev/blog/context-and-structs) — Context best practices for background services
- GitZen codebase analysis — Existing patterns in internal/git and internal/app

---
*Stack research for: TUI Git Client Auto Fetch Feature*
*Researched: 2026-04-01*