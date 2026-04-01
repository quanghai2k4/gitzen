# Codebase Concerns

**Analysis Date:** 2026-04-01

## Tech Debt

**Large Key Handler Function:**
- Issue: Single monolithic function handling all keyboard input with 579 lines
- Files: `internal/app/keys.go`
- Impact: Difficult to maintain, test individual key behaviors, and add new shortcuts
- Fix approach: Split into focused handlers by pane type (files, commits, branches, etc.)

**Complex Command Orchestration:**
- Issue: Single file with 492 lines mixing different command types and error handling
- Files: `internal/app/cmds.go`
- Impact: Hard to trace command flow, add new git operations, or modify error handling
- Fix approach: Separate into domain-specific command modules (status, diff, git ops, etc.)

**Modal Component Complexity:**
- Issue: Single component handling multiple modal types with mixed responsibilities
- Files: `internal/components/modal.go` (399 lines)
- Impact: Adding new modal types requires touching existing logic, hard to test individually
- Fix approach: Create separate modal components for each type (CommitModal, ConfirmModal, etc.)

**Git Operations Mixed with UI Logic:**
- Issue: Git command execution tightly coupled with UI state management
- Files: `internal/git/git.go`, `internal/app/cmds.go`
- Impact: Hard to unit test git operations, reuse logic outside TUI
- Fix approach: Extract pure git operations from UI concerns, add proper interfaces

## Known Bugs

**Unused Variable in Hunk Operations:**
- Symptoms: `_ = stagedContent` indicates incomplete implementation
- Files: `internal/git/git.go:214`
- Trigger: During hunk staging/unstaging operations
- Workaround: Operations may not handle staged content correctly

**Missing Error Context:**
- Symptoms: Generic error messages without specific git operation context
- Files: `internal/git/git.go:238-245`
- Trigger: When git commands fail with empty stderr
- Workaround: Users get uninformative "command failed" messages

## Security Considerations

**Command Injection Risk:**
- Risk: Git command arguments not properly escaped
- Files: `internal/git/git.go` (all command functions)
- Current mitigation: Using exec.Command with separate args reduces risk
- Recommendations: Add input validation for file paths and commit messages

**Log File Permissions:**
- Risk: Debug logs may contain sensitive information with world-readable permissions
- Files: `internal/logger/logger.go:45`
- Current mitigation: Creates files with 0644 permissions
- Recommendations: Use more restrictive 0600 permissions for log files

## Performance Bottlenecks

**Synchronous Git Operations:**
- Problem: All git commands block UI thread
- Files: `internal/git/git.go` (all run methods)
- Cause: No async execution or background processing
- Improvement path: Implement command queue with goroutines and proper cancellation

**Large Diff Rendering:**
- Problem: No pagination for large diffs, limited to 5000 lines
- Files: `internal/limits/limits.go:15`, diff rendering components
- Cause: Loading entire diff into memory and viewport
- Improvement path: Implement virtual scrolling or streaming diff display

**Repeated Status Checks:**
- Problem: Git status called frequently without caching
- Files: `internal/app/cmds.go:68-77`
- Cause: Every pane refresh triggers new status check
- Improvement path: Add intelligent caching with file system watchers

## Fragile Areas

**Layout Calculation:**
- Files: `internal/ui/layout.go`, `internal/app/model.go`
- Why fragile: Complex interdependencies between pane sizes and focus state
- Safe modification: Always test with different terminal sizes
- Test coverage: No automated tests for layout logic

**Git Parser Functions:**
- Files: `internal/git/parse_*.go`
- Why fragile: String parsing dependent on git output format
- Safe modification: Add validation for unexpected formats
- Test coverage: Good test coverage exists

**Hunk Manipulation:**
- Files: `internal/git/git.go:175-216`
- Why fragile: Complex diff parsing and reversal logic
- Safe modification: Extensive testing with various diff formats required
- Test coverage: Some tests present but may miss edge cases

## Scaling Limits

**Memory Usage with Large Repos:**
- Current capacity: Limited to 200 commits, 5000 diff lines
- Limit: Memory exhaustion with large repositories
- Scaling path: Implement pagination and lazy loading

**Command Timeout Constraints:**
- Current capacity: 3 seconds for commands, 10 for diffs, 30 for network
- Limit: Fails with slow repositories or network
- Scaling path: Make timeouts configurable or adaptive

## Dependencies at Risk

**Go Version Requirement:**
- Risk: Using Go 1.24.0 which is very recent
- Impact: Limits deployment on older systems
- Migration plan: Consider compatibility with Go 1.21+ for broader support

**Charm Dependencies:**
- Risk: Heavy reliance on Charmbracelet ecosystem
- Impact: Breaking changes in bubbles/bubbletea affect entire UI
- Migration plan: Abstract UI framework behind interfaces

## Missing Critical Features

**Error Recovery:**
- Problem: No graceful handling of git repository corruption or network failures
- Blocks: Users lose work when operations fail unexpectedly

**Configuration System:**
- Problem: No way to customize keybindings, colors, or behavior
- Blocks: Users cannot adapt tool to their workflow

**Undo/Redo Operations:**
- Problem: No way to undo git operations performed through the UI
- Blocks: Users fear making mistakes with destructive operations

## Test Coverage Gaps

**UI Integration Tests:**
- What's not tested: Full user workflows, keyboard navigation
- Files: All components in `internal/components/`
- Risk: UI regressions go unnoticed
- Priority: High

**Error Handling Paths:**
- What's not tested: Git command failures, invalid repository states
- Files: `internal/git/git.go`, `internal/app/cmds.go`
- Risk: Crashes or undefined behavior in error conditions
- Priority: Medium

**Concurrency Safety:**
- What's not tested: Race conditions in logger, simultaneous git operations
- Files: `internal/logger/logger.go`
- Risk: Data races or corrupted logs
- Priority: Medium

---

*Concerns audit: 2026-04-01*
