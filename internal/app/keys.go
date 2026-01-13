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
		// Cycle through: Files -> Branches -> Commits -> Stash -> Main -> Files
		m.focus = (m.focus + 1) % 6
		if m.focus == paneStatus {
			m.focus = paneFiles // skip status pane
		}
		m.resize()
		m.refreshAllViews()
		return m, m.loadDiffForCurrentPane()
	case "shift+tab":
		m.focus = (m.focus + 5) % 6
		if m.focus == paneStatus {
			m.focus = paneMain
		}
		m.resize()
		m.refreshAllViews()
		return m, m.loadDiffForCurrentPane()
	case "esc":
		m.errorMsg = ""
		return m, nil
	case "c":
		if m.focus == paneFiles {
			m.commitMode = true
			m.commitIn.SetValue("")
			m.commitIn.Focus()
		}
		return m, nil

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
	}
	return m, nil
}

func (m model) handleBranchesKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		if m.branchesCursor < len(m.branches)-1 {
			m.branchesCursor++
			m.refreshAllViews()
		}
	case "k", "up":
		if m.branchesCursor > 0 {
			m.branchesCursor--
			m.refreshAllViews()
		}
	case "g":
		m.branchesCursor = 0
		m.refreshAllViews()
	case "G":
		if len(m.branches) > 0 {
			m.branchesCursor = len(m.branches) - 1
			m.refreshAllViews()
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
	}
	return m, nil
}

func (m model) handleStashKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		if m.stashCursor < len(m.stashItems)-1 {
			m.stashCursor++
			m.refreshAllViews()
		}
	case "k", "up":
		if m.stashCursor > 0 {
			m.stashCursor--
			m.refreshAllViews()
		}
	case "g":
		m.stashCursor = 0
		m.refreshAllViews()
	case "G":
		if len(m.stashItems) > 0 {
			m.stashCursor = len(m.stashItems) - 1
			m.refreshAllViews()
		}
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
			m.commitIn.Blur()
			return m, nil
		case "enter":
			msg := strings.TrimSpace(m.commitIn.Value())
			if msg == "" {
				m.errorMsg = "Commit message is empty"
				return m, nil
			}
			m.commitMode = false
			m.commitIn.Blur()
			return m, commitCmd(m.git, msg)
		}
	}

	var cmd tea.Cmd
	m.commitIn, cmd = m.commitIn.Update(msg)
	return m, cmd
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
