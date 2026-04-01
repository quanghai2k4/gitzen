# Phase 3: UI Integration & Visual Feedback - Research

**Researched:** 2026-04-01
**Domain:** TUI visual feedback and status indicators for background operations
**Confidence:** HIGH

## Summary

GitZen's auto-fetch functionality is implemented and working, but users need visual feedback to understand when fetch operations occur, their status, and outcomes. This research investigates TUI visual feedback patterns, status indicators, and non-intrusive notification systems for Bubble Tea applications.

The existing architecture already includes message passing for background operations, a command log system, modal error handling, and a flexible layout engine. The key insight is extending these patterns with dedicated status indicators, progress feedback, and toast-style notifications that integrate seamlessly with the lazygit-inspired interface.

**Primary recommendation:** Implement a multi-layer visual feedback system combining status bar indicators, progress animations, and subtle notifications without disrupting the core workflow.

## User Constraints (from CONTEXT.md)

No CONTEXT.md found - proceeding with full research scope.

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| UI-01 | GitZen displays progress indicators when fetch operations are in progress | Status bar integration, Bubble Tea progress component patterns |
| UI-02 | GitZen shows success/failure notifications after fetch operations complete | Toast notifications, temporary status messages, command log integration |
| UI-03 | GitZen provides non-intrusive status updates that don't disrupt user workflow | Peripheral visual feedback, non-modal notifications |
| UI-04 | GitZen notifies users when new commits are available after successful fetch | Commit count indicators, branch status updates, visual diff indicators |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/charmbracelet/bubbles/spinner | v0.21.0 | Progress animations | Official Bubble Tea component for loading states |
| github.com/charmbracelet/bubbles/progress | v0.21.0 | Progress bars | Official component with animation support |
| github.com/charmbracelet/lipgloss | v1.1.0 | Styling and theming | Already used, provides consistent visual styling |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Bubble Tea timers | Built-in | Auto-dismiss notifications | For temporary toast messages |
| Bubble Tea Tick | Built-in | Animation frames | For spinner animations during fetch |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Built-in progress | Custom animations | Standard components are more maintainable |
| Toast overlays | Status bar only | Toast provides better visibility for important events |

**Installation:**
Already included in current dependencies.

**Version verification:** Current versions confirmed from go.mod:
- github.com/charmbracelet/bubbles v0.21.0
- github.com/charmbracelet/lipgloss v1.1.0

## Architecture Patterns

### Recommended Visual Feedback Layers

```
┌─ Status Bar ─────────────────────────────────┐
│ repo → branch [🔄 Fetching...] │ Last: 2m ago │
├─ Main Interface ────────────────────────────────┤
│ ┌─ Files ─┐ ┌─ Main View ──────────┐         │
│ │         │ │                      │         │
│ │         │ │                      │         │
│ └─────────┘ └──────────────────────┘         │
├─ Toast Notifications (temporary) ──────────────┤
│ ✅ Fetch completed: 3 new commits on main     │
├─ Command Log ───────────────────────────────────┤
│ auto fetch: completed successfully...          │
└─────────────────────────────────────────────────┘
```

### Pattern 1: Status Bar Integration
**What:** Extend existing StatusPane with fetch status indicators
**When to use:** Always visible, non-intrusive progress indication
**Example:**
```go
// Source: Internal architecture analysis
type StatusPane struct {
    BasePane
    repoName      string
    branchName    string
    fetchStatus   FetchStatus
    lastFetchTime time.Time
    styles        ui.Styles
}

type FetchStatus int
const (
    FetchIdle FetchStatus = iota
    FetchInProgress
    FetchSuccess
    FetchError
)
```

### Pattern 2: Toast Notification System
**What:** Temporary overlay notifications for important events
**When to use:** Fetch completion, errors, new commit availability
**Example:**
```go
// Source: Bubble Tea patterns research
type ToastNotification struct {
    message     string
    level       ToastLevel
    duration    time.Duration
    startTime   time.Time
    visible     bool
}
```

### Pattern 3: Progress Indication
**What:** Animated spinner or progress indicator during active fetches
**When to use:** Long-running fetch operations
**Example:**
```go
// Source: Bubbles progress component
import "github.com/charmbracelet/bubbles/spinner"

spinner := spinner.New()
spinner.Spinner = spinner.Dot
spinner.Style = styles.InfoStyle
```

### Anti-Patterns to Avoid
- **Modal interruptions:** Don't use modal dialogs for routine fetch updates
- **Aggressive notifications:** Don't show toasts for every background operation
- **Status spam:** Don't update status indicators more than once per second

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Progress animations | Custom spinner logic | bubbles/spinner component | Handles animation timing, various styles, built-in patterns |
| Progress bars | Manual bar rendering | bubbles/progress component | Smooth animations, color blending, percentage display |
| Color blending | RGB interpolation | lipgloss.Color and gradients | Handles terminal color limitations, consistent theming |
| Toast positioning | Overlay math | lipgloss positioning helpers | Terminal-safe positioning, responsive layout |

**Key insight:** Bubble Tea ecosystem provides battle-tested components for all common UI feedback patterns. Custom implementations risk timing issues and terminal compatibility problems.

## Common Pitfalls

### Pitfall 1: Update Frequency Overload
**What goes wrong:** Status indicators update too frequently, causing visual noise and performance issues
**Why it happens:** Direct coupling between git operations and UI updates
**How to avoid:** Debounce status updates, limit to meaningful state changes only
**Warning signs:** Screen flickering, high CPU usage, status bar "jumping"

### Pitfall 2: Modal Fetch Dialogs
**What goes wrong:** Using modal dialogs for fetch progress blocks user workflow
**Why it happens:** Treating background operations like foreground user actions
**How to avoid:** Use peripheral indicators (status bar, subtle animations)
**Warning signs:** User complaints about interruptions, workflow disruption

### Pitfall 3: Invisible Error States
**What goes wrong:** Network errors or fetch failures go unnoticed by users
**Why it happens:** Background operations fail silently with no user feedback
**How to avoid:** Always surface errors through multiple channels (toast + command log)
**Warning signs:** Users unaware of stale repository state, confusion about sync status

### Pitfall 4: Animation Performance Impact
**What goes wrong:** Complex animations impact TUI responsiveness
**Why it happens:** Animation loops consume terminal I/O resources
**How to avoid:** Use efficient Bubble Tea tick patterns, limit concurrent animations
**Warning signs:** Input lag, delayed key responses, reduced scrolling performance

## Code Examples

Verified patterns from official sources:

### Status Bar with Fetch Indicator
```go
// Source: GitZen internal architecture + Bubble Tea patterns
func (p *StatusPane) refreshContent() {
    branch := p.branchName
    if branch == "" {
        branch = "master"
    }

    repoStyle := p.styles.BranchHeadStyle.Bold(true)
    branchStyle := p.styles.BranchLocalStyle

    var statusIndicator string
    switch p.fetchStatus {
    case FetchInProgress:
        statusIndicator = p.styles.InfoStyle.Render(" 🔄")
    case FetchSuccess:
        statusIndicator = p.styles.InfoStyle.Render(" ✅")
    case FetchError:
        statusIndicator = p.styles.ErrorStyle.Render(" ❌")
    default:
        statusIndicator = ""
    }

    content := repoStyle.Render(p.repoName) + " → " + 
              branchStyle.Render(branch) + statusIndicator
    p.SetContent(content)
}
```

### Toast Notification Component
```go
// Source: Bubble Tea message patterns + Modal analysis
type ToastManager struct {
    toasts   []ToastNotification
    styles   ui.Styles
    maxToasts int
}

func (tm *ToastManager) AddToast(message string, level ToastLevel, duration time.Duration) tea.Cmd {
    toast := ToastNotification{
        message:   message,
        level:     level,
        duration:  duration,
        startTime: time.Now(),
        visible:   true,
    }
    
    tm.toasts = append(tm.toasts, toast)
    if len(tm.toasts) > tm.maxToasts {
        tm.toasts = tm.toasts[1:] // Remove oldest
    }
    
    return tea.Tick(duration, func(t time.Time) tea.Msg {
        return toastExpiredMsg{startTime: toast.startTime}
    })
}
```

### Fetch Progress Integration
```go
// Source: Bubble Tea command patterns + background manager
func (m model) handleFetchStart() (model, tea.Cmd) {
    m.statusPane.SetFetchStatus(FetchInProgress)
    
    return m, tea.Batch(
        m.backgroundManager.ExecuteAutoFetch(m.repoRoot),
        m.spinner.Tick, // Start spinner animation
    )
}

func (m model) handleFetchComplete(msg autoFetchResultMsg) (model, tea.Cmd) {
    var cmds []tea.Cmd
    
    if msg.Success && !msg.Skipped {
        m.statusPane.SetFetchStatus(FetchSuccess)
        m.statusPane.SetLastFetchTime(time.Now())
        
        // Show success toast
        cmds = append(cmds, m.toastManager.AddToast(
            "Fetch completed: " + msg.Message,
            ToastSuccess,
            3*time.Second,
        ))
        
        // Refresh data views
        cmds = append(cmds, 
            loadCommitsCmd(m.git),
            loadBranchesCmd(m.git),
        )
    } else if !msg.Success {
        m.statusPane.SetFetchStatus(FetchError)
        
        // Show error toast
        cmds = append(cmds, m.toastManager.AddToast(
            "Fetch failed: " + msg.Message,
            ToastError,
            5*time.Second,
        ))
    }
    
    return m, tea.Batch(cmds...)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Silent background operations | Multi-layer visual feedback | 2024+ TUI apps | Users always know system state |
| Modal progress dialogs | Peripheral progress indicators | Bubble Tea v2 patterns | Non-disruptive workflow |
| Log-only notifications | Toast + log combination | Modern TUI design | Critical events get attention |

**Deprecated/outdated:**
- Modal progress bars: Replaced by in-context indicators
- Text-only status: Enhanced with unicode icons and colors
- Single notification channel: Multi-channel approach (status bar + toast + log)

## Open Questions

1. **Toast Positioning Strategy**
   - What we know: Can overlay using lipgloss positioning
   - What's unclear: Optimal placement relative to focused pane
   - Recommendation: Position in bottom-right, avoid blocking active content

2. **New Commit Count Display**
   - What we know: Need to indicate new commits after fetch
   - What's unclear: Where to show counts (branches pane vs status bar)
   - Recommendation: Both - badge on branch items, summary in status bar

## Environment Availability

All dependencies are already installed and available:

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Bubble Tea | Core framework | ✓ | v1.3.10 | — |
| Bubbles components | Progress/spinner | ✓ | v0.21.0 | — |
| Lipgloss styling | Visual feedback | ✓ | v1.1.0 | — |

**Missing dependencies with no fallback:**
None - all required components available

**Missing dependencies with fallback:**
None - visual feedback is additive to existing functionality

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go built-in testing + Manual verification |
| Config file | none — see Wave 0 |
| Quick run command | `go test ./internal/components/...` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| UI-01 | Progress indicators during fetch | unit | `go test ./internal/components -run TestStatusPane` | ❌ Wave 0 |
| UI-02 | Success/failure notifications | unit | `go test ./internal/components -run TestToastManager` | ❌ Wave 0 |
| UI-03 | Non-intrusive status updates | integration | Manual verification with real fetch | ❌ Wave 0 |
| UI-04 | New commit notifications | unit | `go test ./internal/components -run TestCommitIndicators` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/components/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green + manual UI verification

### Wave 0 Gaps
- [ ] `internal/components/status_test.go` — covers UI-01 status indicators
- [ ] `internal/components/toast_test.go` — covers UI-02 notifications
- [ ] `internal/components/fetch_feedback_test.go` — covers UI integration patterns
- [ ] Test fixtures for mock fetch operations

## Sources

### Primary (HIGH confidence)
- GitZen internal architecture - existing components and patterns analysis
- Bubble Tea official documentation - component integration patterns
- Bubbles component library - progress, spinner, and animation patterns

### Secondary (MEDIUM confidence)
- Lazygit UI patterns - inspiration for Git TUI visual feedback
- Terminal UI best practices - non-intrusive notification patterns

### Tertiary (LOW confidence)
None - all findings verified against official sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using existing dependencies and official components
- Architecture: HIGH - Extends proven patterns from current codebase
- Pitfalls: HIGH - Based on TUI development best practices and Bubble Tea patterns

**Research date:** 2026-04-01
**Valid until:** 2026-05-01 (stable UI patterns, framework mature)