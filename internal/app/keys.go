package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys
	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab":
		// Cycle through: Files -> Branches -> Commits -> Stash -> Main -> CmdLog -> Files
		m.focus = (m.focus + 1) % 7
		if m.focus == paneStatus {
			m.focus = paneFiles // skip status pane
		}
		m.resize()
		m.refreshAllViews()
		return m, m.loadDiffForCurrentPane()
	case "shift+tab":
		m.focus = (m.focus + 6) % 7
		if m.focus == paneStatus {
			m.focus = paneCmdLog
		}
		m.resize()
		m.refreshAllViews()
		return m, m.loadDiffForCurrentPane()
	case "esc":
		m.errorMsg = ""
		m.confirmMode = false
		return m, nil
	case "c":
		if m.focus == paneFiles {
			m.commitMode = true
			m.amendMode = false
			m.commitIn.SetValue("")
			m.commitIn.Focus()
		}
		return m, nil
	case "A": // Amend commit
		if m.focus == paneFiles && len(m.stagedItems) > 0 {
			m.commitMode = true
			m.amendMode = true
			m.commitIn.SetValue("")
			m.commitIn.Placeholder = "New message (empty = keep old)"
			m.commitIn.Focus()
		}
		return m, nil

	// Global git operations
	case "p": // Pull
		return m, pullCmd(m.git)
	case "P": // Push
		return m, pushCmd(m.git)
	case "f": // Fetch
		return m, fetchCmd(m.git)

	// Jump keys (like lazygit: 1-5 for side panels, 0 for main)
	case "1":
		// Status pane - not focusable
		return m, nil
	case "2":
		m.focus = paneFiles
		m.resize()
		m.refreshAllViews()
		return m, m.loadDiffForCurrentPane()
	case "3":
		m.focus = paneBranches
		m.resize()
		m.refreshAllViews()
		return m, nil
	case "4":
		m.focus = paneCommits
		m.resize()
		m.refreshAllViews()
		return m, m.loadDiffForCurrentPane()
	case "5":
		m.focus = paneStash
		m.resize()
		m.refreshAllViews()
		return m, m.loadDiffForCurrentPane()
	case "6":
		m.focus = paneCmdLog
		m.resize()
		m.refreshAllViews()
		return m, nil
	case "0":
		m.focus = paneMain
		m.resize()
		m.refreshAllViews()
		return m, nil
	}

	// Pane-specific keys
	switch m.focus {
	case paneFiles:
		return m.handleFilesKeys(key)
	case paneBranches:
		return m.handleBranchesKeys(key)
	case paneCommits:
		return m.handleCommitsKeys(key)
	case paneStash:
		return m.handleStashKeys(key)
	case paneCmdLog:
		return m.handleCmdLogKeys(key)
	case paneMain:
		return m.handleMainKeys(key)
	}

	return m, nil
}

func (m model) handleFilesKeys(key string) (tea.Model, tea.Cmd) {
	totalFiles := len(m.stagedItems) + len(m.unstagedItems)

	switch key {
	case "j", "down":
		if m.filesCursor < totalFiles-1 {
			m.filesCursor++
			m.refreshAllViews()
			return m, m.loadDiffForCurrentPane()
		}
	case "k", "up":
		if m.filesCursor > 0 {
			m.filesCursor--
			m.refreshAllViews()
			return m, m.loadDiffForCurrentPane()
		}
	case "g":
		m.filesCursor = 0
		m.refreshAllViews()
		return m, m.loadDiffForCurrentPane()
	case "G":
		if totalFiles > 0 {
			m.filesCursor = totalFiles - 1
			m.refreshAllViews()
			return m, m.loadDiffForCurrentPane()
		}
	case " ": // space to toggle stage/unstage
		return m, m.toggleStageCmd()
	case "a":
		// Stage all
		return m, stageAllCmd(m)
	case "d": // Discard changes
		return m.discardSelectedFile()
	}
	return m, nil
}

func (m model) handleBranchesKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		if m.branchesCursor < len(m.branches)-1 {
			m.branchesCursor++
			m.refreshAllViews()
			return m, m.loadBranchDiff()
		}
	case "k", "up":
		if m.branchesCursor > 0 {
			m.branchesCursor--
			m.refreshAllViews()
			return m, m.loadBranchDiff()
		}
	case "g":
		m.branchesCursor = 0
		m.refreshAllViews()
		return m, m.loadBranchDiff()
	case "G":
		if len(m.branches) > 0 {
			m.branchesCursor = len(m.branches) - 1
			m.refreshAllViews()
			return m, m.loadBranchDiff()
		}
	case "enter", " ": // Checkout branch
		if m.branchesCursor < len(m.branches) {
			branch := m.branches[m.branchesCursor]
			if branch.IsCurrent {
				return m, nil // Already on this branch
			}
			return m, checkoutBranchCmd(m.git, branch.Name)
		}
	case "n": // New branch
		m.createBranchMode = true
		m.branchIn.SetValue("")
		m.branchIn.Focus()
		return m, nil
	case "d": // Delete branch
		if m.branchesCursor < len(m.branches) {
			branch := m.branches[m.branchesCursor]
			if branch.IsCurrent {
				m.errorMsg = "Cannot delete current branch"
				return m, nil
			}
			m.confirmMode = true
			m.confirmTitle = "Delete branch " + branch.Name + "?"
			m.confirmYesText = "y"
			m.confirmAction = func() tea.Cmd {
				return deleteBranchCmd(m.git, branch.Name, false)
			}
			return m, nil
		}
	case "D": // Force delete branch
		if m.branchesCursor < len(m.branches) {
			branch := m.branches[m.branchesCursor]
			if branch.IsCurrent {
				m.errorMsg = "Cannot delete current branch"
				return m, nil
			}
			m.confirmMode = true
			m.confirmTitle = "Force delete branch " + branch.Name + "?"
			m.confirmYesText = "y"
			m.confirmAction = func() tea.Cmd {
				return deleteBranchCmd(m.git, branch.Name, true)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) handleCommitsKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		if m.commitsCursor < len(m.commitItems)-1 {
			m.commitsCursor++
			m.refreshAllViews()
			return m, m.loadCommitDiff()
		}
	case "k", "up":
		if m.commitsCursor > 0 {
			m.commitsCursor--
			m.refreshAllViews()
			return m, m.loadCommitDiff()
		}
	case "g":
		m.commitsCursor = 0
		m.refreshAllViews()
		return m, m.loadCommitDiff()
	case "G":
		if len(m.commitItems) > 0 {
			m.commitsCursor = len(m.commitItems) - 1
			m.refreshAllViews()
			return m, m.loadCommitDiff()
		}
	case "enter":
		return m, m.loadCommitDiff()
	case "r": // Reset soft (undo commit, keep staged)
		if m.commitsCursor == 0 && len(m.commitItems) > 0 {
			m.confirmMode = true
			m.confirmTitle = "Undo last commit (keep staged)?"
			m.confirmYesText = "y"
			m.confirmAction = func() tea.Cmd {
				return resetSoftCmd(m.git, 1)
			}
			return m, nil
		}
	case "R": // Reset mixed (undo commit, keep unstaged)
		if m.commitsCursor == 0 && len(m.commitItems) > 0 {
			m.confirmMode = true
			m.confirmTitle = "Undo last commit (keep unstaged)?"
			m.confirmYesText = "y"
			m.confirmAction = func() tea.Cmd {
				return resetMixedCmd(m.git, 1)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) handleStashKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		if m.stashCursor < len(m.stashItems)-1 {
			m.stashCursor++
			m.refreshAllViews()
			return m, m.loadStashDiffForCursor()
		}
	case "k", "up":
		if m.stashCursor > 0 {
			m.stashCursor--
			m.refreshAllViews()
			return m, m.loadStashDiffForCursor()
		}
	case "g":
		m.stashCursor = 0
		m.refreshAllViews()
		return m, m.loadStashDiffForCursor()
	case "G":
		if len(m.stashItems) > 0 {
			m.stashCursor = len(m.stashItems) - 1
			m.refreshAllViews()
			return m, m.loadStashDiffForCursor()
		}
	case "enter":
		return m, m.loadStashDiffForCursor()
	case " ": // Stash apply
		if m.stashCursor < len(m.stashItems) {
			ref := m.stashItems[m.stashCursor].Ref
			return m, stashApplyCmd(m.git, ref)
		}
	case "p": // Stash pop (like lazygit uses 'g' for pop)
		if m.stashCursor < len(m.stashItems) {
			ref := m.stashItems[m.stashCursor].Ref
			return m, stashPopCmd(m.git, ref)
		}
	case "d": // Stash drop (with confirm)
		if m.stashCursor < len(m.stashItems) {
			ref := m.stashItems[m.stashCursor].Ref
			m.confirmMode = true
			m.confirmTitle = "Drop " + ref + "?"
			m.confirmYesText = "y"
			m.confirmAction = func() tea.Cmd {
				return stashDropCmd(m.git, ref)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) handleCmdLogKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		m.cmdLogVP.LineDown(1)
	case "k", "up":
		m.cmdLogVP.LineUp(1)
	case "g":
		m.cmdLogVP.GotoTop()
	case "G":
		m.cmdLogVP.GotoBottom()
	}
	return m, nil
}

func (m model) handleMainKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		m.mainVP.LineDown(1)
	case "k", "up":
		m.mainVP.LineUp(1)
	case "d":
		m.mainVP.HalfViewDown()
	case "u":
		m.mainVP.HalfViewUp()
	case "g":
		m.mainVP.GotoTop()
	case "G":
		m.mainVP.GotoBottom()
	}
	return m, nil
}

func (m model) updateCommitMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.commitMode = false
			m.amendMode = false
			m.commitIn.Blur()
			m.commitIn.Placeholder = "Commit message"
			return m, nil
		case "enter":
			msgVal := strings.TrimSpace(m.commitIn.Value())

			if m.amendMode {
				// Amend mode: empty message = keep old message
				m.commitMode = false
				m.amendMode = false
				m.commitIn.Blur()
				m.commitIn.Placeholder = "Commit message"
				return m, commitAmendCmd(m.git, msgVal)
			}

			// Normal commit mode
			if msgVal == "" {
				m.errorMsg = "Commit message is empty"
				return m, nil
			}
			m.commitMode = false
			m.commitIn.Blur()
			return m, commitCmd(m.git, msgVal)
		}
	}

	var cmd tea.Cmd
	m.commitIn, cmd = m.commitIn.Update(msg)
	return m, cmd
}

func (m model) updateCreateBranchMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.createBranchMode = false
			m.branchIn.Blur()
			return m, nil
		case "enter":
			name := strings.TrimSpace(m.branchIn.Value())
			if name == "" {
				m.errorMsg = "Branch name is empty"
				return m, nil
			}
			m.createBranchMode = false
			m.branchIn.Blur()
			return m, createBranchCmd(m.git, name)
		}
	}

	var cmd tea.Cmd
	m.branchIn, cmd = m.branchIn.Update(msg)
	return m, cmd
}

func (m model) updateConfirmMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "n", "N":
			m.confirmMode = false
			m.confirmAction = nil
			return m, nil
		case "y", "Y", "enter":
			m.confirmMode = false
			action := m.confirmAction
			m.confirmAction = nil
			if action != nil {
				return m, action()
			}
			return m, nil
		}
	}
	return m, nil
}

// Helper methods for loading diff based on current selection

func (m model) loadDiffForCurrentPane() tea.Cmd {
	switch m.focus {
	case paneFiles:
		path, staged := m.selectedFilePath()
		if path == "" {
			return func() tea.Msg { return diffLoadedMsg{Diff: "(no file selected)"} }
		}
		return loadDiffCmd(m.git, path, staged)
	case paneCommits:
		return m.loadCommitDiff()
	case paneBranches:
		return m.loadBranchDiff()
	case paneStash:
		return m.loadStashDiffForCursor()
	default:
		return nil
	}
}

func (m model) selectedFilePath() (string, bool) {
	// Files pane shows staged first, then unstaged
	if m.filesCursor < len(m.stagedItems) {
		return m.stagedItems[m.filesCursor].Path, true
	}
	unstagedIdx := m.filesCursor - len(m.stagedItems)
	if unstagedIdx < len(m.unstagedItems) {
		return m.unstagedItems[unstagedIdx].Path, false
	}
	return "", false
}

func (m model) loadCommitDiff() tea.Cmd {
	if m.commitsCursor < len(m.commitItems) {
		hash := m.commitItems[m.commitsCursor].Hash
		return loadShowCommitCmd(m.git, hash)
	}
	return nil
}

func (m model) toggleStageCmd() tea.Cmd {
	// If cursor is on staged file, unstage it; otherwise stage it
	if m.filesCursor < len(m.stagedItems) {
		// Unstage
		path := m.stagedItems[m.filesCursor].Path
		return unstageFileCmd(m.git, path)
	}
	// Stage
	unstagedIdx := m.filesCursor - len(m.stagedItems)
	if unstagedIdx < len(m.unstagedItems) {
		path := m.unstagedItems[unstagedIdx].Path
		return stageFileCmd(m.git, path)
	}
	return nil
}

func (m model) loadBranchDiff() tea.Cmd {
	if m.branchesCursor < len(m.branches) {
		branch := m.branches[m.branchesCursor].Name
		return loadBranchDiffCmd(m.git, branch)
	}
	return nil
}

func (m model) loadStashDiffForCursor() tea.Cmd {
	if m.stashCursor < len(m.stashItems) {
		ref := m.stashItems[m.stashCursor].Ref
		return loadStashDiffCmd(m.git, ref)
	}
	return nil
}

// discardSelectedFile handles discarding changes for the selected file
func (m model) discardSelectedFile() (tea.Model, tea.Cmd) {
	// Can only discard unstaged files
	if m.filesCursor < len(m.stagedItems) {
		// File is staged, need to unstage first
		m.errorMsg = "Cannot discard staged file. Unstage first (space)"
		return m, nil
	}

	unstagedIdx := m.filesCursor - len(m.stagedItems)
	if unstagedIdx >= len(m.unstagedItems) {
		return m, nil
	}

	file := m.unstagedItems[unstagedIdx]
	isUntracked := file.Status == "?"

	// Setup confirm dialog
	m.confirmMode = true
	m.confirmTitle = "Discard changes to " + file.Path + "?"
	m.confirmYesText = "y"
	m.confirmAction = func() tea.Cmd {
		return discardFileCmd(m.git, file.Path, isUntracked)
	}

	return m, nil
}

// selectedFileInfo returns path, isStaged, isUntracked for selected file
func (m model) selectedFileInfo() (path string, staged bool, untracked bool) {
	if m.filesCursor < len(m.stagedItems) {
		f := m.stagedItems[m.filesCursor]
		return f.Path, true, false
	}
	unstagedIdx := m.filesCursor - len(m.stagedItems)
	if unstagedIdx < len(m.unstagedItems) {
		f := m.unstagedItems[unstagedIdx]
		return f.Path, false, f.Status == "?"
	}
	return "", false, false
}
