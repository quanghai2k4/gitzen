package app

import (
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
		if err := r.Add(path); err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("staged " + path)
	}
}

// Unstage a specific file
func unstageFileCmd(r git.Runner, path string) tea.Cmd {
	return func() tea.Msg {
		if err := r.RestoreStaged(path); err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("unstaged " + path)
	}
}

func stageAllCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		if err := r.Add("."); err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("staged all")
	}
}

func commitCmd(r git.Runner, message string) tea.Cmd {
	return func() tea.Msg {
		_, err := r.Commit(message)
		if err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("committed")
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
		var err error
		if isUntracked {
			err = r.DiscardUntracked(path)
		} else {
			err = r.DiscardFile(path)
		}
		if err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("discarded " + path)
	}
}

// Pull from remote
func pullCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		out, err := r.Pull()
		if err != nil {
			return errMsg(err.Error())
		}
		if strings.Contains(out, "Already up to date") {
			return statusToastMsg("Already up to date")
		}
		return statusToastMsg("Pulled successfully")
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
				return errMsg(err.Error())
			}
			branch = strings.TrimSpace(branch)
			remote, err := r.GetRemote()
			if err != nil {
				return errMsg("No remote configured")
			}
			_, err = r.PushSetUpstream(remote, branch)
			if err != nil {
				return errMsg(err.Error())
			}
			return statusToastMsg("Pushed (set upstream " + remote + "/" + branch + ")")
		}
		_, err := r.Push()
		if err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("Pushed successfully")
	}
}

// Checkout branch
func checkoutBranchCmd(r git.Runner, branch string) tea.Cmd {
	return func() tea.Msg {
		_, err := r.CheckoutBranch(branch)
		if err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("Switched to " + branch)
	}
}

// Create new branch
func createBranchCmd(r git.Runner, name string) tea.Cmd {
	return func() tea.Msg {
		_, err := r.CreateBranch(name)
		if err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("Created branch " + name)
	}
}

// ========== MEDIUM PRIORITY COMMANDS ==========

// Fetch from remote
func fetchCmd(r git.Runner) tea.Cmd {
	return func() tea.Msg {
		_, err := r.Fetch()
		if err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("Fetched all remotes")
	}
}

// Delete branch
func deleteBranchCmd(r git.Runner, name string, force bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if force {
			err = r.DeleteBranchForce(name)
		} else {
			err = r.DeleteBranch(name)
		}
		if err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("Deleted branch " + name)
	}
}

// Stash apply
func stashApplyCmd(r git.Runner, ref string) tea.Cmd {
	return func() tea.Msg {
		if err := r.StashApply(ref); err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("Applied " + ref)
	}
}

// Stash pop
func stashPopCmd(r git.Runner, ref string) tea.Cmd {
	return func() tea.Msg {
		if err := r.StashPop(ref); err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("Popped " + ref)
	}
}

// Stash drop
func stashDropCmd(r git.Runner, ref string) tea.Cmd {
	return func() tea.Msg {
		if err := r.StashDrop(ref); err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("Dropped " + ref)
	}
}

// Amend commit
func commitAmendCmd(r git.Runner, message string) tea.Cmd {
	return func() tea.Msg {
		_, err := r.CommitAmend(message)
		if err != nil {
			return errMsg(err.Error())
		}
		if message == "" {
			return statusToastMsg("Amended commit (kept message)")
		}
		return statusToastMsg("Amended commit")
	}
}

// Reset soft (undo commit, keep staged)
func resetSoftCmd(r git.Runner, n int) tea.Cmd {
	return func() tea.Msg {
		if err := r.ResetSoftHead(n); err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("Reset soft HEAD~" + string(rune('0'+n)))
	}
}

// Reset mixed (undo commit, keep unstaged)
func resetMixedCmd(r git.Runner, n int) tea.Cmd {
	return func() tea.Msg {
		if err := r.ResetMixedHead(n); err != nil {
			return errMsg(err.Error())
		}
		return statusToastMsg("Reset mixed HEAD~" + string(rune('0'+n)))
	}
}
