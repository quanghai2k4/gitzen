package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"gitzen/internal/git"
)

type statusLoadedMsg struct{ Status git.Status }

type commitsLoadedMsg struct{ Commits []git.CommitItem }

type branchLoadedMsg struct{ Branch string }

type branchesLoadedMsg struct{ Branches []git.Branch }

type stashLoadedMsg struct{ Entries []git.StashEntry }

type diffLoadedMsg struct{ Diff string }

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

func loadDiffCmd(r git.Runner, path string, staged bool) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(path) == "" {
			return diffLoadedMsg{Diff: ""}
		}
		out, err := r.DiffFile(path, staged)
		if err != nil {
			return errMsg(err.Error())
		}
		if strings.TrimSpace(out) == "" {
			out = "(no diff)"
		}
		return diffLoadedMsg{Diff: out}
	}
}

func loadShowCommitCmd(r git.Runner, hash string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(hash) == "" {
			return diffLoadedMsg{Diff: ""}
		}
		out, err := r.ShowCommit(hash)
		if err != nil {
			return errMsg(err.Error())
		}
		return diffLoadedMsg{Diff: out}
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
			return diffLoadedMsg{Diff: ""}
		}
		out, err := r.ShowStash(ref)
		if err != nil {
			return errMsg(err.Error())
		}
		return diffLoadedMsg{Diff: out}
	}
}

func loadBranchDiffCmd(r git.Runner, branch string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(branch) == "" {
			return diffLoadedMsg{Diff: ""}
		}
		out, err := r.DiffBranch(branch)
		if err != nil {
			// May fail for current branch, show empty
			return diffLoadedMsg{Diff: "(no diff from current branch)"}
		}
		if strings.TrimSpace(out) == "" {
			out = "(no diff from current branch)"
		}
		return diffLoadedMsg{Diff: out}
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
