# GitZen

## What This Is

GitZen is a TUI (Terminal User Interface) Git client built with Go and Bubble Tea, inspired by lazygit. It provides an interactive terminal interface for Git operations with multiple panes for status, files, branches, commits, and stash management. Users can perform common Git operations like staging, committing, pushing, pulling, and branch management through keyboard shortcuts in a clean, organized interface.

## Core Value

Users can perform Git operations faster and more intuitively through a visual terminal interface without memorizing complex Git commands.

## Requirements

### Validated

- ✓ Interactive TUI with multiple panes (Status, Files, Branches, Commits, Stash) — existing
- ✓ Basic Git operations (stage, unstage, commit, push, pull, fetch) — existing  
- ✓ Diff viewer with syntax highlighting — existing
- ✓ Branch management (create, checkout, delete, merge) — existing
- ✓ Commit history with diff view — existing
- ✓ Reflog support — existing
- ✓ Stash management — existing
- ✓ Modal dialogs for commit messages — existing
- ✓ Keyboard shortcuts for navigation and operations — existing
- ✓ Cross-platform support (Linux, macOS, Windows) — existing
- ✓ Installation scripts and release automation — existing

### Active

- [ ] Auto fetch from remote (main branch and current branch)
- [ ] Periodic fetch timer (30-minute intervals)
- [ ] Fetch on application startup
- [ ] Safe fetch (only when working directory is clean)
- [ ] Visual indicators and notifications for new commits
- [ ] Configuration option to enable/disable auto fetch
- [ ] Background fetch without disrupting user workflow

### Out of Scope

- Auto merge functionality — too risky, user should handle conflicts manually
- Real-time collaboration features — beyond scope of local Git client
- GUI version — focus is on terminal interface
- Plugin system — keep core functionality simple

## Context

GitZen already has a solid foundation with a component-based TUI architecture using Bubble Tea. The existing Git integration layer (`internal/git`) provides command execution and output parsing. The app uses an event-driven architecture with centralized state management, making it suitable for adding background operations like periodic fetching.

The current architecture has:
- Clean separation between UI components and business logic
- Asynchronous Git operations using command-response pattern  
- Existing Git runner for command execution
- Modal system for user interactions

The auto fetch feature will integrate with the existing Git layer and use the same patterns for background operations.

## Constraints

- **Tech stack**: Go 1.24+, Bubble Tea framework, existing Git command execution pattern — maintain consistency with current architecture
- **Performance**: Background fetching must not block UI interactions — use existing async command pattern
- **Safety**: Only fetch when working directory is clean — prevent disrupting user's work
- **Compatibility**: Must work across Linux, macOS, Windows — leverage existing cross-platform support

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Fetch-only strategy (no auto merge) | Prevents conflicts and data loss, user maintains control | — Pending |
| 30-minute interval with startup fetch | Balance between staying updated and performance | — Pending |
| Clean working directory requirement | Safety first - never disrupt uncommitted work | — Pending |
| Main + current branch fetching | Most relevant branches for user workflow | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-01 after initialization*