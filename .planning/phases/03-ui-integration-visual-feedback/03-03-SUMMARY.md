---
phase: 03-ui-integration-visual-feedback
plan: 03
subsystem: ui-components
tags: [visual-feedback, commit-indicators, branches-pane, status-bar]
dependencies:
  requires: [03-01, 03-02]
  provides: [commit-count-display, new-commit-notifications]
  affects: [branches-pane, status-bar, fetch-completion]
tech_stack:
  added: []
  patterns: [commit-counting, ui-integration, async-updates]
key_files:
  created: []
  modified: [internal/git/git.go, internal/components/branches.go, internal/components/status.go, internal/app/model.go, internal/app/cmds.go]
decisions:
  - Load commit counts automatically after successful fetch operations for immediate user feedback
  - Use existing git rev-list command pattern for commit counting to maintain consistency
  - Display commit indicators in both branches pane (+N/-N format) and status bar (total summary)
  - Clear new commit indicators when user views commits to prevent stale information
metrics:
  duration: 25m
  tasks_completed: 3
  files_modified: 5
  commits_made: 3
  completed_date: "2026-04-01T15:15:00Z"
---

# Phase 3 Plan 3: New Commit Indicators Summary

**Enhanced GitZen with commit count indicators that notify users when new commits become available after fetch operations**

## Objective Achievement

✅ **UI-04 COMPLETE**: Users are notified when new commits become available after fetch

Successfully implemented new commit indicators across branches pane and status bar with automatic refresh after fetch operations, providing immediate visual feedback when new commits are available.

## Implementation Summary

### Task 1: Add branch commit counting to git Runner (commit d34202a)
- **Added CommitCount struct**: `Ahead int, Behind int` for tracking commit differences
- **Added BranchCommitCounts type**: `map[string]CommitCount` for managing multiple branches  
- **Implemented GetBranchCommitCounts()**: Uses `git rev-list --count --left-right` pattern
- **Added helper methods**: `ParseCommitCountOutput()` and `GetSingleBranchCount()`
- **Applied error handling**: Graceful fallback to empty counts, no UI failures

### Task 2: Enhance branches pane with commit indicators (commit 1fd736c) 
- **Extended BranchesPane struct**: Added `commitCounts git.BranchCommitCounts` field
- **Added SetCommitCounts() method**: Triggers content refresh when counts update
- **Updated renderBranchWithCommits()**: Shows "+N -N" format badges after branch names
- **Integrated theme styling**: Uses InfoStyle (green) for ahead, WarningStyle (yellow) for behind
- **Applied display logic**: Only shows counts > 0 to reduce visual noise

### Task 3: Wire commit count updates into fetch completion (commit 216460e)
- **Added commitCountsLoadedMsg**: Message type for async commit count delivery
- **Implemented loadCommitCountsCmd()**: Loads counts for main/master/current branches
- **Added message handler**: Updates both branches pane and status bar with counts  
- **Integrated with fetch results**: Triggers count loading after successful startup/auto fetch
- **Extended StatusPane**: Added `SetNewCommitsAvailable()` and "[N new]" indicator

## Technical Details

### Commit Counting Algorithm
```go
// Uses existing GitZen git command patterns
func (r Runner) GetBranchCommitCounts(branches []string) (BranchCommitCounts, error) {
    // Executes: git rev-list --count --left-right origin/branch...branch
    // Returns: map[branch]CommitCount{Ahead: N, Behind: M}
}
```

### UI Integration Pattern
```go
// Follows GitZen async message flow
startupFetchResultMsg -> loadCommitCountsCmd() -> commitCountsLoadedMsg -> UI update
autoFetchResultMsg -> loadCommitCountsCmd() -> commitCountsLoadedMsg -> UI update
```

### Visual Design
- **Branches pane**: `feature-branch +3 -1` (green +, yellow -)
- **Status bar**: `gitzen → main [2 new]` (info style)
- **Lifecycle**: Counts refresh after fetch, clear on commit view

## Deviations from Plan

None - plan executed exactly as written. All tasks completed successfully with proper error handling, performance considerations, and GitZen integration patterns.

## Performance Impact

- **Minimal overhead**: Commit counting only triggered after successful fetch (not continuous)
- **Efficient targeting**: Focuses on main/master/current branches (matches auto fetch scope)
- **Graceful degradation**: Git command failures don't break UI, just show no indicators
- **Batch operations**: Uses tea.Batch to combine with existing fetch result commands

## Integration Quality

- **Theme consistency**: Reuses existing GitZen color system (InfoStyle, WarningStyle)  
- **Component patterns**: Follows BasePane refresh patterns and viewport rendering
- **Message flow**: Integrates seamlessly with existing async command architecture
- **Error boundaries**: Maintains GitZen's philosophy of never failing UI on git errors

## Files Modified

| File | Purpose | Changes |
| ---- | ------- | ------- |
| `internal/git/git.go` | Commit counting backend | +CommitCount types, +GetBranchCommitCounts method |
| `internal/components/branches.go` | Branch indicators display | +commitCounts field, +SetCommitCounts method |
| `internal/components/status.go` | Status bar summary | +newCommitsCount field, +SetNewCommitsAvailable method |
| `internal/app/model.go` | Message handling | +commitCountsLoadedMsg handler, fetch integration |
| `internal/app/cmds.go` | Async commands | +loadCommitCountsCmd function, message types |

## Success Metrics

✅ **Functionality**: All commit counting features work as specified  
✅ **Performance**: No UI blocking, efficient git command usage  
✅ **Integration**: Seamlessly fits GitZen architecture and patterns  
✅ **User Experience**: Clear, non-intrusive visual feedback  
✅ **Quality**: Proper error handling, Vietnamese comments, theme consistency

## Verification Results

1. ✅ Branch pane shows "+N" indicators for commits ahead of remote
2. ✅ Branch pane shows "-N" indicators for commits behind remote  
3. ✅ Status bar displays total new commits summary when available
4. ✅ Commit counts refresh automatically after successful fetch operations
5. ✅ Commit indicators use consistent GitZen styling without UI clutter

## Self-Check: PASSED

**Created files exist:** N/A - no new files created  
**Modified files exist:**
- ✅ FOUND: internal/git/git.go
- ✅ FOUND: internal/components/branches.go  
- ✅ FOUND: internal/components/status.go
- ✅ FOUND: internal/app/model.go
- ✅ FOUND: internal/app/cmds.go

**Commits exist:**
- ✅ FOUND: d34202a (Task 1: git Runner commit counting)
- ✅ FOUND: 1fd736c (Task 2: branches pane indicators) 
- ✅ FOUND: 216460e (Task 3: fetch completion integration)

All deliverables verified successfully. UI-04 requirement complete.