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

func stageAllCmd(m model) tea.Cmd {
	return func() tea.Msg {
		if err := m.git.Add("."); err != nil {
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
