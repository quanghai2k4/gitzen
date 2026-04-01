package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"gitzen/internal/background"
	"gitzen/internal/components"
	"gitzen/internal/git"
	"gitzen/internal/logger"
)

type statusLoadedMsg struct{ Status git.Status }

type commitsLoadedMsg struct{ Commits []git.CommitItem }

type reflogLoadedMsg struct{ Entries []git.ReflogEntry }

type branchLoadedMsg struct{ Branch string }

type branchesLoadedMsg struct{ Branches []git.Branch }

type stashLoadedMsg struct{ Entries []git.StashEntry }

// backgroundTickMsg thông báo khi background timer được kích hoạt
type backgroundTickMsg time.Time

// startupFetchMsg triggers startup fetch process
type startupFetchMsg struct{}

// startupFetchResultMsg reports startup fetch completion/failure
type startupFetchResultMsg struct {
	Success bool
	Skipped bool
	Message string
}

// autoFetchResultMsg reports background auto fetch results
type autoFetchResultMsg struct {
	Success bool
	Skipped bool
	Message string
}

// fetchStatusUpdateMsg thông báo cập nhật trạng thái fetch
type fetchStatusUpdateMsg struct {
	Status    components.FetchStatus
	Timestamp time.Time
}

// fetchStatusClearMsg thông báo xóa trạng thái fetch (success/error -> idle)
type fetchStatusClearMsg struct{}

// updateFetchStatusCmd tạo command để cập nhật trạng thái fetch
func updateFetchStatusCmd(status components.FetchStatus) tea.Cmd {
	return func() tea.Msg {
		return fetchStatusUpdateMsg{Status: status, Timestamp: time.Now()}
	}
}

// clearFetchStatusCmd tạo command để xóa trạng thái fetch sau 3 giây
func clearFetchStatusCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return fetchStatusClearMsg{}
	})
}

// addToastCmd tạo command để thêm toast notification với auto-expiration
func addToastCmd(message string, level components.ToastLevel, duration time.Duration) tea.Cmd {
	toastID := int(time.Now().UnixNano()) // Unique ID based on timestamp

	return tea.Batch(
		func() tea.Msg {
			return toastAddMsg{
				toast: components.ToastNotification{
					ID:        toastID,
					Message:   message,
					Level:     level,
					Duration:  duration,
					StartTime: time.Now(),
					Visible:   true,
				},
			}
		},
		tea.Tick(duration, func(t time.Time) tea.Msg {
			return toastExpiredMsg{id: toastID}
		}),
	)
}

// toastAddMsg message để thêm toast
type toastAddMsg struct {
	toast components.ToastNotification
}

// toastExpiredMsg message khi toast hết hạn
type toastExpiredMsg struct {
	id int
}

// commitCountsLoadedMsg thông báo khi commit counts được load
type commitCountsLoadedMsg struct {
	Counts git.BranchCommitCounts
}

// fileWatchRefreshCmd triggers a refresh of git status after file changes
func fileWatchRefreshCmd(gitRunner git.Runner) tea.Cmd {
	return tea.Batch(
		loadStatusCmd(gitRunner),
		loadBranchCmd(gitRunner),   // Refresh current branch info
		loadBranchesCmd(gitRunner), // Refresh branch list
	)
}

// fileWatchEventMsg wraps background.FileWatchEvent for app-level message passing
type fileWatchEventMsg background.FileWatchEvent

type diffLoadedMsg struct {
	Diff     string
	Context  int    // DiffContext type
	Subtitle string // file path, commit hash, etc.
}

type hunksLoadedMsg struct {
	Hunks  []git.Hunk
	Path   string
	Staged bool
}

// splitDiffLoadedMsg contains both unstaged and staged diffs for split view
type splitDiffLoadedMsg struct {
	Unstaged string
	Staged   string
	FilePath string
}

// Diff context constants (match components.DiffContext)
const (
	diffContextNone   = 0
	diffContextFile   = 1
	diffContextCommit = 2
	diffContextStash  = 3
	diffContextBranch = 4
)

type gitCmdMsg string

type errMsg string

type statusToastMsg string

// cmdLogMsg is used to add entries to the command log pane
type cmdLogMsg string

// gitResultMsg contains both the command and result for logging
type gitResultMsg struct {
	Cmd    string
	Result string
	Err    error
}

func loadStatusCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		b, err := r.StatusPorcelainZ()
		if err != nil {
			return errMsg(err.Error())
		}
		st := git.ParseStatusPorcelainV1Z(b)
		return statusLoadedMsg{Status: st}
	}
}

func loadCommitsCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		out, err := r.LogOneline()
		if err != nil {
			// Handle empty repo (no commits yet)
			errStr := err.Error()
			if strings.Contains(errStr, "does not have any commits") ||
				strings.Contains(errStr, "bad revision") ||
				strings.Contains(errStr, "unknown revision") {
				return commitsLoadedMsg{Commits: nil}
			}
			return errMsg(errStr)
		}
		return commitsLoadedMsg{Commits: git.ParseLogOneline(out)}
	}
}

func loadReflogCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		out, err := r.Reflog()
		if err != nil {
			return reflogLoadedMsg{Entries: nil}
		}
		return reflogLoadedMsg{Entries: git.ParseReflog(out)}
	}
}

func loadDiffCmd(r git.Runner, path string, staged bool) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(path) == "" {
			return diffLoadedMsg{Diff: "", Context: diffContextNone}
		}
		out, err := r.DiffFile(path, staged)
		if err != nil {
			return errMsg(err.Error())
		}
		if strings.TrimSpace(out) == "" {
			out = "(no diff)"
		}
		return diffLoadedMsg{Diff: out, Context: diffContextFile, Subtitle: path}
	}
}

func loadHunksCmd(r git.Runner, path string, staged bool) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(path) == "" {
			return hunksLoadedMsg{Hunks: nil, Path: "", Staged: staged}
		}
		out, err := r.DiffFile(path, staged)
		if err != nil {
			return errMsg(err.Error())
		}
		hunks := git.ParseHunks(out)
		return hunksLoadedMsg{Hunks: hunks, Path: path, Staged: staged}
	}
}

func loadShowCommitCmd(r git.Runner, hash string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(hash) == "" {
			return diffLoadedMsg{Diff: "", Context: diffContextNone}
		}
		out, err := r.ShowCommit(hash)
		if err != nil {
			return errMsg(err.Error())
		}
		return diffLoadedMsg{Diff: out, Context: diffContextCommit, Subtitle: hash}
	}
}

// loadSplitDiffCmd loads both unstaged and staged diffs for a file
func loadSplitDiffCmd(r git.Runner, path string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(path) == "" {
			return splitDiffLoadedMsg{}
		}

		// Get unstaged diff
		unstaged, _ := r.DiffFile(path, false)

		// Get staged diff
		staged, _ := r.DiffFile(path, true)

		return splitDiffLoadedMsg{
			Unstaged: unstaged,
			Staged:   staged,
			FilePath: path,
		}
	}
}

// Stage a specific file
func stageFileCmd(r git.Runner, path string) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git add -- %s", path)
		if err := r.Add(path); err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "staged " + path}
	}
}

// Unstage a specific file
func unstageFileCmd(r git.Runner, path string) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git restore --staged -- %s", path)
		if err := r.RestoreStaged(path); err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "unstaged " + path}
	}
}

func stageAllCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		cmd := "git add ."
		if err := r.Add("."); err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "staged all"}
	}
}

func commitCmd(r git.Runner, message string) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git commit -m %q", message)
		_, err := r.Commit(message)
		if err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "committed"}
	}
}

func loadBranchCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		out, err := r.CurrentBranch()
		if err != nil {
			return branchLoadedMsg{Branch: ""}
		}
		return branchLoadedMsg{Branch: strings.TrimSpace(out)}
	}
}

func loadBranchesCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		branches, err := r.ListBranches()
		if err != nil {
			return branchesLoadedMsg{Branches: nil}
		}
		return branchesLoadedMsg{Branches: branches}
	}
}

func loadStashCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		entries, err := r.ListStash()
		if err != nil {
			return stashLoadedMsg{Entries: nil}
		}
		return stashLoadedMsg{Entries: entries}
	}
}

func loadStashDiffCmd(r git.Runner, ref string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(ref) == "" {
			return diffLoadedMsg{Diff: "", Context: diffContextNone}
		}
		out, err := r.ShowStash(ref)
		if err != nil {
			return errMsg(err.Error())
		}
		return diffLoadedMsg{Diff: out, Context: diffContextStash, Subtitle: ref}
	}
}

func loadBranchDiffCmd(r git.Runner, branch string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(branch) == "" {
			return diffLoadedMsg{Diff: "", Context: diffContextNone}
		}
		out, err := r.DiffBranch(branch)
		if err != nil {
			// May fail for current branch, show empty
			return diffLoadedMsg{Diff: "(no diff from current branch)", Context: diffContextBranch, Subtitle: branch}
		}
		if strings.TrimSpace(out) == "" {
			out = "(no diff from current branch)"
		}
		return diffLoadedMsg{Diff: out, Context: diffContextBranch, Subtitle: branch}
	}
}

// ========== HIGH PRIORITY COMMANDS ==========

// Discard changes in a file
func discardFileCmd(r git.Runner, path string, isUntracked bool) tea.Cmd {
	return func() tea.Msg {
		var cmd string
		var err error
		if isUntracked {
			cmd = fmt.Sprintf("git clean -f -- %s", path)
			err = r.DiscardUntracked(path)
		} else {
			cmd = fmt.Sprintf("git checkout -- %s", path)
			err = r.DiscardFile(path)
		}
		if err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "discarded " + path}
	}
}

// Pull from remote
func pullCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		cmd := "git pull"
		out, err := r.Pull()
		if err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		if strings.Contains(out, "Already up to date") {
			return gitResultMsg{Cmd: cmd, Result: "Already up to date"}
		}
		return gitResultMsg{Cmd: cmd, Result: "Pulled successfully"}
	}
}

// Push to remote
func pushCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		// Check if upstream exists
		if !r.HasUpstream() {
			// Get current branch and remote
			branch, err := r.CurrentBranch()
			if err != nil {
				return gitResultMsg{Cmd: "git push", Err: err}
			}
			branch = strings.TrimSpace(branch)
			remote, err := r.GetRemote()
			if err != nil {
				return gitResultMsg{Cmd: "git push", Err: fmt.Errorf("No remote configured")}
			}
			cmd := fmt.Sprintf("git push -u %s %s", remote, branch)
			_, err = r.PushSetUpstream(remote, branch)
			if err != nil {
				return gitResultMsg{Cmd: cmd, Err: err}
			}
			return gitResultMsg{Cmd: cmd, Result: "Pushed (set upstream " + remote + "/" + branch + ")"}
		}
		cmd := "git push"
		_, err := r.Push()
		if err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "Pushed successfully"}
	}
}

// Checkout branch
func checkoutBranchCmd(r git.Runner, branch string) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git checkout %s", branch)
		_, err := r.CheckoutBranch(branch)
		if err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "Switched to " + branch}
	}
}

// Create new branch
func createBranchCmd(r git.Runner, name string) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git checkout -b %s", name)
		_, err := r.CreateBranch(name)
		if err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "Created branch " + name}
	}
}

// ========== MEDIUM PRIORITY COMMANDS ==========

// Fetch from remote
func fetchCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		cmd := "git fetch --all --prune"
		_, err := r.Fetch()
		if err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "Fetched all remotes"}
	}
}

// Delete branch
func deleteBranchCmd(r git.Runner, name string, force bool) tea.Cmd {
	return func() tea.Msg {
		var cmd string
		var err error
		if force {
			cmd = fmt.Sprintf("git branch -D %s", name)
			err = r.DeleteBranchForce(name)
		} else {
			cmd = fmt.Sprintf("git branch -d %s", name)
			err = r.DeleteBranch(name)
		}
		if err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "Deleted branch " + name}
	}
}

// Stash apply
func stashApplyCmd(r git.Runner, ref string) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git stash apply %s", ref)
		if err := r.StashApply(ref); err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "Applied " + ref}
	}
}

// Stash pop
func stashPopCmd(r git.Runner, ref string) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git stash pop %s", ref)
		if err := r.StashPop(ref); err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "Popped " + ref}
	}
}

// Stash drop
func stashDropCmd(r git.Runner, ref string) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git stash drop %s", ref)
		if err := r.StashDrop(ref); err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "Dropped " + ref}
	}
}

// Amend commit
func commitAmendCmd(r git.Runner, message string) tea.Cmd {
	return func() tea.Msg {
		var cmd string
		if message == "" {
			cmd = "git commit --amend --no-edit"
		} else {
			cmd = fmt.Sprintf("git commit --amend -m %q", message)
		}
		_, err := r.CommitAmend(message)
		if err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		if message == "" {
			return gitResultMsg{Cmd: cmd, Result: "Amended commit (kept message)"}
		}
		return gitResultMsg{Cmd: cmd, Result: "Amended commit"}
	}
}

// Reset soft (undo commit, keep staged)
func resetSoftCmd(r git.Runner, n int) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git reset --soft HEAD~%d", n)
		if err := r.ResetSoftHead(n); err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: fmt.Sprintf("Reset soft HEAD~%d", n)}
	}
}

// Reset mixed (undo commit, keep unstaged)
func resetMixedCmd(r git.Runner, n int) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git reset --mixed HEAD~%d", n)
		if err := r.ResetMixedHead(n); err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: fmt.Sprintf("Reset mixed HEAD~%d", n)}
	}
}

// StageHunkCmd stages a single hunk
func stageHunkCmd(r git.Runner, path string, hunkContent string) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git stage hunk in %s", path)
		if err := r.StageHunk(path, hunkContent); err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "staged hunk in " + path}
	}
}

// UnstageHunkCmd unstages a single hunk
func unstageHunkCmd(r git.Runner, path string, hunkContent string) tea.Cmd {
	return func() tea.Msg {
		cmd := fmt.Sprintf("git unstage hunk in %s", path)
		if err := r.UnstageHunk(path, hunkContent); err != nil {
			return gitResultMsg{Cmd: cmd, Err: err}
		}
		return gitResultMsg{Cmd: cmd, Result: "unstaged hunk in " + path}
	}
}

// backgroundTickCmd tạo tea.Cmd cho background timer với 30 giây interval
func backgroundTickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return backgroundTickMsg(t)
	})
}

// loadCommitCountsCmd tạo command để load commit counts cho branches
func loadCommitCountsCmd(gitRunner git.Runner) tea.Cmd {
	return func() tea.Msg {
		// Get current branches from git - start with common defaults
		branches := []string{"main", "master"}

		// Add current branch if it's not HEAD and not already included
		if current, err := gitRunner.CurrentBranch(); err == nil {
			current = strings.TrimSpace(current)
			if current != "HEAD" && current != "main" && current != "master" {
				branches = append(branches, current)
			}
		}

		// Get commit counts for these branches
		counts, err := gitRunner.GetBranchCommitCounts(branches)
		if err != nil {
			// Don't fail UI on git errors, just log warning
			logger.Get().Warn("failed to load commit counts: %v", err)
			return commitCountsLoadedMsg{Counts: make(git.BranchCommitCounts)}
		}

		return commitCountsLoadedMsg{Counts: counts}
	}
}
