# Architecture Research

**Domain:** TUI Git Client Auto Fetch Integration
**Researched:** 2026-04-01
**Confidence:** HIGH

## Integration with Existing Architecture

### Current GitZen Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      TUI Layer (Bubble Tea)                  │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │ Status  │  │ Files   │  │Branches │  │Commits  │        │
│  │ View    │  │ View    │  │ View    │  │ View    │        │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │
│       │            │            │            │              │
├───────┴────────────┴────────────┴────────────┴──────────────┤
│                   Event Dispatcher                           │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐    │
│  │                  Git Operations                      │    │
│  │           (internal/git Runner)                      │    │
│  └─────────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│                    Command Layer                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                   │
│  │ Git CLI  │  │ Config   │  │  File    │                   │
│  │ Commands │  │ Manager  │  │ System   │                   │
│  └──────────┘  └──────────┘  └──────────┘                   │
└─────────────────────────────────────────────────────────────┘
```

### Auto Fetch Integration Points

```
┌─────────────────────────────────────────────────────────────┐
│                      TUI Layer                               │
│  ┌─────────┐  ┌─────────┐  ┌──────────┐  ┌──────────┐      │
│  │ Status  │  │ Files   │  │ Branches │  │ Fetch    │      │
│  │ (+ ind) │  │ View    │  │(+ ind)   │  │ Status   │      │
│  └────┬────┘  └────┬────┘  └────┬─────┘  └────┬─────┘      │
│       │            │            │             │             │
├───────┼────────────┼────────────┼─────────────┼─────────────┤
│       │            │            │             │             │
│  ┌────▼────────────▼────────────▼─────────────▼─────────┐   │
│  │               Event Dispatcher                       │   │
│  │              (+ auto fetch msgs)                     │   │
│  └──┬───────────────────────────────────────────────┬───┘   │
│     │                                               │       │
├─────▼───────────────────────────────────────────────▼───────┤
│  ┌─────────────────┐                   ┌─────────────────┐  │
│  │ Git Operations  │                   │ Auto Fetch      │  │
│  │ (existing)      │◄──────────────────┤ Manager         │  │
│  └─────────────────┘                   └─────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                   │
│  │ Git CLI  │  │ Config   │  │ Timer    │                   │
│  │ Commands │  │ Manager  │  │ Service  │                   │
│  └──────────┘  └──────────┘  └──────────┘                   │
└─────────────────────────────────────────────────────────────┘
```

## Component Integration

### New Components for Auto Fetch

| Component | Responsibility | Integration Point |
|-----------|---------------|-------------------|
| AutoFetchManager | Manages fetch scheduling and execution | Event dispatcher - receives timer messages |
| FetchTimer | 30-minute interval timer using tea.Every() | Bubble Tea command layer |
| FetchStatusIndicator | Shows fetch status in UI | Status/Branches views |
| SafetyChecker | Validates working directory is clean | Git runner - reuses existing patterns |

### Modified Existing Components

| Component | Modifications | Purpose |
|-----------|--------------|---------|
| Main Model | Add auto fetch state fields | Store fetch status, timer state |
| Update() Method | Handle auto fetch messages | Process timer ticks, fetch results |
| Status View | Add fetch status indicator | Visual feedback to user |
| Config System | Add auto fetch settings | Enable/disable, interval configuration |

## Background Timer Patterns

### Bubble Tea Timer Architecture

**Pattern:** Use `tea.Every()` for periodic tasks that need precise intervals
**Alternative:** Use `tea.Tick()` for one-time delayed execution

```go
type AutoFetchMsg struct {
    Type string // "timer_tick", "fetch_start", "fetch_complete", "fetch_error"
    Data interface{}
}

// In Init() - start the timer
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        // Existing init commands...
        startAutoFetchTimer(), // New auto fetch timer
    )
}

// Timer command using tea.Every
func startAutoFetchTimer() tea.Cmd {
    return tea.Every(30*time.Minute, func(t time.Time) tea.Msg {
        return AutoFetchMsg{Type: "timer_tick", Data: t}
    })
}
```

### Integration with Existing Async Git Pattern

GitZen already uses async command execution. Auto fetch extends this pattern:

```go
type FetchCommand struct {
    Branches []string // main + current branch
    Safe     bool     // only if working directory clean
}

// Reuse existing Git Runner pattern
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case AutoFetchMsg:
        switch msg.Type {
        case "timer_tick":
            if m.autoFetchEnabled && m.isWorkingDirClean() {
                return m, m.gitRunner.Fetch(FetchCommand{
                    Branches: []string{"main", m.currentBranch},
                    Safe:     true,
                })
            }
        case "fetch_complete":
            m.fetchStatus = "success"
            m.lastFetchTime = time.Now()
            // Update UI indicators
        }
    }
    return m, nil
}
```

## State Management Integration

### Centralized State Pattern (Current)

GitZen uses centralized state in the main model. Auto fetch extends this:

```go
type Model struct {
    // Existing state...
    views        map[string]View
    gitRunner    git.Runner
    
    // Auto fetch state additions
    autoFetchEnabled bool
    fetchTimer      time.Time
    lastFetchTime   time.Time
    fetchStatus     string // "idle", "fetching", "success", "error"
    newCommits      []CommitInfo // commits found since last fetch
}
```

### Event Flow for Auto Fetch

```
Timer Tick
    ↓
Safety Check (clean working dir)
    ↓
Fetch Command → Git Runner → Git CLI
    ↓                ↓           ↓
UI Update ← Result Parse ← Git Output
```

## Architectural Patterns

### Pattern 1: Background Timer with Safety Gates

**What:** Periodic background operations that check preconditions before executing
**When to use:** Any background task that could interfere with user work
**Trade-offs:** More complex state management vs user safety

**Implementation:**
```go
func (m Model) shouldAutoFetch() bool {
    return m.autoFetchEnabled && 
           m.isWorkingDirClean() && 
           time.Since(m.lastFetchTime) >= 30*time.Minute &&
           !m.isUserActivelyEditing()
}
```

### Pattern 2: Async Git Operations with UI Feedback

**What:** Non-blocking Git commands with progressive UI updates
**When to use:** Any Git operation that could take time (fetch, clone, etc.)
**Trade-offs:** Complexity vs responsiveness

**Implementation:**
```go
// Command returns immediately, UI stays responsive
func (r Runner) FetchAsync(cmd FetchCommand) tea.Cmd {
    return func() tea.Msg {
        // Background execution
        result, err := r.runWithTimeout("fetch", "origin", cmd.Branches...)
        return AutoFetchMsg{Type: "fetch_complete", Data: FetchResult{result, err}}
    }
}
```

### Pattern 3: Composable UI Indicators

**What:** Status indicators that can be embedded in multiple views
**When to use:** Cross-cutting concerns like fetch status, dirty state
**Trade-offs:** Consistency vs component coupling

**Implementation:**
```go
func (m Model) renderFetchIndicator() string {
    switch m.fetchStatus {
    case "fetching":
        return "⟳ Fetching..."
    case "success":
        return fmt.Sprintf("✓ Fetched %s", m.lastFetchTime.Format("15:04"))
    case "error":
        return "✗ Fetch failed"
    default:
        return ""
    }
}
```

## Configuration Integration

### Extending Existing Config System

GitZen likely has a config system. Auto fetch adds:

```go
type Config struct {
    // Existing config...
    
    // Auto fetch configuration
    AutoFetch struct {
        Enabled         bool          `yaml:"enabled" default:"false"`
        Interval        time.Duration `yaml:"interval" default:"30m"`
        SafeOnly        bool          `yaml:"safe_only" default:"true"`
        FetchOnStartup  bool          `yaml:"fetch_on_startup" default:"true"`
        Branches        []string      `yaml:"branches" default:"[\"main\"]"`
    } `yaml:"auto_fetch"`
}
```

## Error Handling and Safety

### Safety Mechanisms

1. **Working Directory Check:** Only fetch if `git status --porcelain` is empty
2. **Network Timeout:** Use existing `NetworkTimeout` from GitZen config
3. **Graceful Degradation:** Failed fetches don't break UI
4. **User Override:** Manual operations pause auto fetch temporarily

### Error Recovery Patterns

```go
func (m Model) handleFetchError(err error) (Model, tea.Cmd) {
    m.fetchStatus = "error"
    m.lastError = err
    
    // Exponential backoff for retry
    nextAttempt := time.Now().Add(time.Duration(m.fetchFailures) * 5 * time.Minute)
    return m, tea.Tick(nextAttempt.Sub(time.Now()), func(t time.Time) tea.Msg {
        return AutoFetchMsg{Type: "retry_fetch"}
    })
}
```

## Implementation Phases

### Phase 1: Core Timer Infrastructure
- Add timer using `tea.Every()`
- Basic auto fetch state management
- Safety checks (clean working directory)

### Phase 2: Git Integration  
- Extend existing Git Runner with fetch commands
- Handle fetch results and errors
- Basic UI status indicator

### Phase 3: Enhanced Features
- Configuration system integration
- Multiple branch support (main + current)
- Startup fetch option

### Phase 4: Polish & Visual Indicators
- Enhanced UI feedback
- New commits notification
- Error handling improvements

## Anti-Patterns to Avoid

### Anti-Pattern 1: Blocking UI with Git Operations

**What people do:** Call Git commands synchronously in Update()
**Why it's wrong:** Freezes the entire TUI during network operations
**Do this instead:** Use GitZen's existing async command pattern with tea.Cmd

### Anti-Pattern 2: Timer State in Multiple Places

**What people do:** Create separate timers in each component
**Why it's wrong:** Leads to race conditions and inconsistent state
**Do this instead:** Centralized timer state in main model, propagated to views

### Anti-Pattern 3: Ignoring Safety Checks

**What people do:** Fetch regardless of working directory state
**Why it's wrong:** Can conflict with user's uncommitted work
**Do this instead:** Always check working directory cleanliness first

## Integration Timeline

| Week | Focus | Deliverables |
|------|-------|-------------|
| 1 | Timer Infrastructure | Basic timer, safety checks, config |
| 2 | Git Integration | Fetch commands, result handling |
| 3 | UI Integration | Status indicators, error display |
| 4 | Testing & Polish | Edge cases, error recovery, UX |

## Sources

- Bubble Tea Documentation: https://pkg.go.dev/charm.land/bubbletea/v2
- Bubble Tea Timer Examples: https://github.com/charmbracelet/bubbletea/tree/main/examples/timer
- Bubble Tea Realtime Examples: https://github.com/charmbracelet/bubbletea/tree/main/examples/realtime
- GitZen Existing Architecture: internal/git/git.go patterns
- Lazygit Architecture Reference: https://github.com/jesseduffield/lazygit

---
*Architecture research for: TUI Git Client Auto Fetch Integration*
*Researched: 2026-04-01*