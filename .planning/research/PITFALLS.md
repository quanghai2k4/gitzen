# Domain Pitfalls

**Domain:** TUI Git clients with background operations (auto fetch)
**Researched:** 2026-04-01
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: UI Blocking During Background Operations

**What goes wrong:**
Auto fetch operations block the TUI event loop, causing the entire interface to freeze for several seconds. Users lose responsiveness and the app feels broken.

**Why it happens:**
Developers implement background operations synchronously in the main event loop or fail to properly separate Git command execution from UI rendering. Bubble Tea's Update() method blocks until Git commands complete.

**How to avoid:**
- Use Bubble Tea's Command pattern with `tea.ExecProcess` for truly async Git operations
- Implement command-response pattern: dispatch async command in Update(), handle response in subsequent Update() call
- Never call Git operations directly in Update() or View() methods
- Add loading states and progress indicators during operations

**Warning signs:**
- UI becomes unresponsive during fetch operations
- Users report "app freezing" for 2-5 seconds
- Cannot interrupt or cancel ongoing operations
- Keyboard input queues up and floods after operations complete

**Phase to address:**
Phase 1 (Background Operations Foundation) - Must establish async patterns before adding periodic fetch

---

### Pitfall 2: Race Conditions Between User Actions and Background Fetch

**What goes wrong:**
Auto fetch conflicts with user Git operations, causing corruption, merge conflicts, or leaving repository in inconsistent state. User loses work or gets cryptic error messages.

**Why it happens:**
No coordination between background fetch timer and user-initiated Git operations. Multiple Git commands running simultaneously can corrupt repository state or conflict with each other.

**How to avoid:**
- Implement operation queue with mutex to serialize Git commands
- Cancel or delay background fetch when user operations are active
- Check for dirty working directory before auto fetch
- Add "operation in progress" state to prevent concurrent Git commands
- Use file locking or Git's built-in locking mechanisms

**Warning signs:**
- "Another git process seems to be running" errors
- Repository enters detached HEAD state unexpectedly
- Staged changes disappear during background operations
- Git operations fail with locking errors

**Phase to address:**
Phase 1 (Background Operations Foundation) - Critical for any background Git operations

---

### Pitfall 3: Background Operations Triggering Unwanted UI Updates

**What goes wrong:**
Auto fetch causes UI to refresh/rebuild unexpectedly, interrupting user workflow. Current selection jumps, modal dialogs close, or user loses their place in the interface.

**Why it happens:**
Background fetch success triggers full UI refresh without preserving user context. Bubble Tea model updates cause View to rebuild from scratch, losing focus state and selection.

**How to avoid:**
- Preserve UI state (current selection, scroll position, active pane) during model updates
- Use targeted updates instead of full refresh when background operations complete
- Implement UI state preservation pattern in model
- Add "quiet" update mode for background operations
- Only refresh relevant panes (status, branches) not entire UI

**Warning signs:**
- Current selection jumps to top after auto fetch
- Modal dialogs or forms close unexpectedly
- User complaints about "interface jumping around"
- Loss of scroll position after background operations

**Phase to address:**
Phase 2 (UI State Management) - After establishing async patterns

---

### Pitfall 4: No Visual Feedback for Background Operations

**What goes wrong:**
Users don't know when auto fetch is running, completed, or failed. They manually trigger fetch operations that are already running, or don't realize their view is outdated.

**Why it happens:**
Developers focus on making background operations "invisible" but provide no status indicators. Users need to know system state even for automatic operations.

**How to avoid:**
- Add subtle status indicators (spinner, progress bar, timestamp)
- Show last fetch time and status in status bar
- Provide notifications for fetch completion/failure
- Add visual indication when new commits are available
- Use different colors/icons to show fetch states

**Warning signs:**
- Users manually fetch immediately after auto fetch
- Complaints about not knowing if data is current
- Users don't notice new commits until manual refresh
- No way to tell if auto fetch is enabled/working

**Phase to address:**
Phase 2 (Visual Indicators) - Essential for user trust in auto features

---

### Pitfall 5: Background Timer Not Properly Cancelled on Exit

**What goes wrong:**
Auto fetch timer continues running after application exit, potentially causing resource leaks or preventing clean shutdown. In extreme cases, orphaned processes continue fetching.

**Why it happens:**
Bubble Tea program cleanup doesn't properly cancel background timers or goroutines. Timer cleanup logic is not properly integrated with application lifecycle.

**How to avoid:**
- Use context.Context with cancellation for all background operations
- Implement proper cleanup in Bubble Tea's program teardown
- Cancel all timers and goroutines in cleanup handlers
- Use `tea.Quit` command to trigger proper shutdown sequence
- Test application exit scenarios thoroughly

**Warning signs:**
- Application hangs on exit with Ctrl+C
- Background processes visible after app closes
- Resource usage continues after application exit
- Multiple fetch processes running simultaneously

**Phase to address:**
Phase 1 (Background Operations Foundation) - Critical for proper resource management

---

### Pitfall 6: Auto Fetch Interfering with Authentication

**What goes wrong:**
Background fetch fails silently when credentials expire, or triggers authentication prompts that break TUI interface. SSH key passphrases or 2FA prompts cannot be handled in background.

**Why it happens:**
Authentication state changes between user and background operations. TUI applications cannot easily handle interactive authentication prompts during background operations.

**How to avoid:**
- Detect authentication failures and disable auto fetch
- Use credential caching (SSH agent, Git credential helper)
- Fail gracefully on auth errors with user notification
- Provide easy way to re-enable after fixing credentials
- Test with various authentication methods (SSH, HTTPS, tokens)

**Warning signs:**
- Auto fetch silently stops working after some time
- Authentication prompts break TUI display
- Background operations fail without user notification
- Credentials work for manual operations but not auto fetch

**Phase to address:**
Phase 3 (Error Handling & Auth) - After basic functionality works

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Synchronous fetch in UI thread | Simple implementation | UI blocking, poor UX | Never - always use async |
| No operation serialization | Faster initial development | Race conditions, data corruption | Never - Git operations must be serialized |
| Polling instead of smart timing | Simple timer logic | Unnecessary network/CPU usage | Only during MVP prototype |
| Global state for fetch status | Easy access across components | Tight coupling, hard to test | Only in Phase 1, refactor in Phase 2 |
| No error handling for background ops | Cleaner happy path code | Silent failures, user confusion | Never - background operations need robust error handling |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Git SSH | Not handling SSH agent properly | Check SSH_AUTH_SOCK, handle key passphrase scenarios |
| Git HTTPS | Hardcoding credentials | Use Git credential helpers, respect system config |
| Remote repositories | Assuming network is always available | Implement timeout, retry logic, offline mode |
| Git configuration | Ignoring user's Git config | Respect user's fetch settings, merge strategies |
| Multiple remotes | Only fetching 'origin' | Fetch from all configured remotes or user-selected ones |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Fetching entire history | Long initial fetch times | Use shallow fetch or limit depth | Repos with >1000 commits |
| No fetch batching | Multiple network requests | Batch multiple branch fetches | >10 branches to track |
| Full UI refresh on fetch | Screen flicker, poor performance | Targeted updates to changed data | >100 commits in history view |
| No operation caching | Repeated expensive Git operations | Cache Git status, branch lists | Complex repos with many refs |
| Synchronous Git parsing | Blocking on large repos | Stream/lazy parsing of Git output | Repos with >10MB output |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Running Git commands with user input | Command injection attacks | Sanitize all inputs, use Git library APIs |
| Ignoring SSH host key verification | Man-in-the-middle attacks | Respect SSH known_hosts, warn on changes |
| Storing credentials in config files | Credential exposure | Use system credential stores, environment variables |
| Following symbolic links in Git operations | Directory traversal attacks | Validate paths, use Git's built-in protections |
| Auto-fetching from untrusted remotes | Malicious code execution via Git hooks | Validate remote URLs, disable hooks during fetch |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No indication of auto fetch status | Users don't trust the feature | Clear status indicators, last update timestamps |
| Auto fetch interrupting user workflow | Frustration, lost productivity | Respect user context, preserve UI state |
| No way to disable auto fetch quickly | Users feel loss of control | Prominent toggle in settings and status bar |
| Silent background failures | Users work with stale data | Error notifications, fallback to manual mode |
| Auto fetch during conflicts/merges | Confusion, potential data loss | Disable during active Git operations |
| No feedback on what changed | Users miss important updates | Highlight new commits, show fetch summary |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Auto Fetch:** Often missing proper cleanup on exit — verify timers are cancelled and goroutines stopped
- [ ] **Background Operations:** Often missing operation serialization — verify Git commands cannot run concurrently
- [ ] **Error Handling:** Often missing user notification for silent failures — verify errors are displayed appropriately
- [ ] **UI State:** Often missing state preservation during updates — verify selection and position are maintained
- [ ] **Authentication:** Often missing credential refresh handling — verify behavior when auth expires
- [ ] **Network Issues:** Often missing offline/timeout handling — verify graceful degradation without network
- [ ] **Configuration:** Often missing user preference persistence — verify settings survive app restart
- [ ] **Performance:** Often missing operation throttling — verify behavior under high frequency operations

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| UI Blocking | MEDIUM | Refactor to async pattern, add loading states, test responsiveness |
| Race Conditions | HIGH | Implement operation queue, add Git locking, audit all Git operations |
| UI State Loss | LOW | Add state preservation pattern, store selection/position in model |
| Silent Failures | LOW | Add error handling, notification system, user feedback mechanisms |
| Timer Leaks | MEDIUM | Implement context cancellation, test cleanup scenarios, fix lifecycle |
| Auth Issues | MEDIUM | Add auth detection, credential validation, user notification system |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| UI Blocking | Phase 1: Background Operations | UI remains responsive during 30+ second fetch operations |
| Race Conditions | Phase 1: Background Operations | Concurrent user operations don't conflict with auto fetch |
| UI State Loss | Phase 2: UI State Management | Selection and position preserved through fetch updates |
| No Visual Feedback | Phase 2: Visual Indicators | Users can see fetch status, progress, and results |
| Timer Leaks | Phase 1: Background Operations | Clean shutdown with no orphaned processes |
| Auth Interference | Phase 3: Error Handling | Graceful handling of credential expiration scenarios |
| Performance Issues | Phase 4: Optimization | Smooth performance with large repos and frequent fetches |
| Configuration Missing | Phase 3: Configuration | User preferences persist and work correctly |

## Sources

- Bubble Tea examples and documentation (official timer patterns)
- LazyGit codebase analysis (background operation patterns)  
- GitZen existing architecture (command-response pattern)
- TUI application best practices (async operation handling)
- Git CLI documentation (fetch behavior and locking)
- Personal experience with TUI Git clients (common failure modes)
- Community discussions on TUI responsiveness issues
- Git repository corruption scenarios and prevention

---
*Pitfalls research for: TUI Git clients with auto fetch*
*Researched: 2026-04-01*