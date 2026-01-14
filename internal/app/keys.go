package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"gitzen/internal/components"
	"gitzen/internal/ui"
)

func (m model) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys
	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab":
		// In split mode (Main from Files), tab toggles between panes
		if m.focus == ui.PaneMain && m.mainViewSource == ui.PaneFiles {
			m.splitDiffView.ToggleFocus()
			return m, nil
		}
		// If in Main/CmdLog, go back to Files (lazygit style)
		if m.focus == ui.PaneMain || m.focus == ui.PaneCmdLog {
			m.focus = ui.PaneFiles
			m.mainViewSource = 0
		} else {
			m.focus = m.nextFocusablePane()
		}
		m.layout = ui.CalculateLayout(m.layout.Width, m.layout.Height, m.focus)
		m.resizeComponents()
		m.refreshAllPanes()
		return m, m.loadDiffForCurrentPane()
	case "shift+tab":
		// In split mode, shift+tab also toggles
		if m.focus == ui.PaneMain && m.mainViewSource == ui.PaneFiles {
			m.splitDiffView.ToggleFocus()
			return m, nil
		}
		// If in Main/CmdLog, go back to Files (lazygit style)
		if m.focus == ui.PaneMain || m.focus == ui.PaneCmdLog {
			m.focus = ui.PaneFiles
			m.mainViewSource = 0
		} else {
			m.focus = m.prevFocusablePane()
		}
		m.layout = ui.CalculateLayout(m.layout.Width, m.layout.Height, m.focus)
		m.resizeComponents()
		m.refreshAllPanes()
		return m, m.loadDiffForCurrentPane()
	case "esc":
		// If in Main/CmdLog, go back to previous sidebar pane
		if m.focus == ui.PaneMain || m.focus == ui.PaneCmdLog {
			m.focus = ui.PaneFiles
			m.mainViewSource = 0 // Reset
			m.layout = ui.CalculateLayout(m.layout.Width, m.layout.Height, m.focus)
			m.resizeComponents()
			m.refreshAllPanes()
			return m, nil
		}
		m.modal.Close()
		return m, nil

	// Commit keys (from Files pane)
	case "c":
		if m.focus == ui.PaneFiles && m.filesPane.HasStaged() {
			m.modal.OpenCommit(false)
		}
		return m, nil
	case "A":
		if m.focus == ui.PaneFiles && m.filesPane.HasStaged() {
			m.modal.OpenCommit(true)
		}
		return m, nil

	// Global git operations
	case "p":
		if m.focus == ui.PaneStash {
			// 'p' in stash pane = pop
			return m.handleStashKeys(key)
		}
		return m, pullCmd(m.git)
	case "P":
		return m, pushCmd(m.git)
	case "f":
		return m, fetchCmd(m.git)

	// Jump keys (sidebar panes only, lazygit style)
	case "1":
		m.focus = ui.PaneFiles
		m.layout = ui.CalculateLayout(m.layout.Width, m.layout.Height, m.focus)
		m.resizeComponents()
		m.refreshAllPanes()
		return m, m.loadDiffForCurrentPane()
	case "2":
		m.focus = ui.PaneBranches
		m.layout = ui.CalculateLayout(m.layout.Width, m.layout.Height, m.focus)
		m.resizeComponents()
		m.refreshAllPanes()
		return m, m.loadBranchDiff()
	case "3":
		m.focus = ui.PaneCommits
		m.layout = ui.CalculateLayout(m.layout.Width, m.layout.Height, m.focus)
		m.resizeComponents()
		m.refreshAllPanes()
		return m, m.loadCommitDiff()
	case "4":
		m.focus = ui.PaneStash
		m.layout = ui.CalculateLayout(m.layout.Width, m.layout.Height, m.focus)
		m.resizeComponents()
		m.refreshAllPanes()
		return m, m.loadStashDiff()
	}

	// Pane-specific keys
	switch m.focus {
	case ui.PaneFiles:
		return m.handleFilesKeys(key)
	case ui.PaneBranches:
		return m.handleBranchesKeys(key)
	case ui.PaneCommits:
		return m.handleCommitsKeys(key)
	case ui.PaneStash:
		return m.handleStashKeys(key)
	case ui.PaneCmdLog:
		return m.handleCmdLogKeys(key)
	case ui.PaneMain:
		return m.handleMainKeys(key)
	}

	return m, nil
}

func (m model) handleFilesKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		m.filesPane.CursorDown()
		m.filesPane.Refresh()
		return m, m.loadDiffForCurrentPane()
	case "k", "up":
		m.filesPane.CursorUp()
		m.filesPane.Refresh()
		return m, m.loadDiffForCurrentPane()
	case "g":
		m.filesPane.CursorTop()
		m.filesPane.Refresh()
		return m, m.loadDiffForCurrentPane()
	case "G":
		m.filesPane.CursorBottom()
		m.filesPane.Refresh()
		return m, m.loadDiffForCurrentPane()
	case "enter": // Focus main view (lazygit style) with split diff
		m.focus = ui.PaneMain
		m.mainViewSource = ui.PaneFiles
		m.layout = ui.CalculateLayout(m.layout.Width, m.layout.Height, m.focus)
		m.resizeComponents()
		m.refreshAllPanes()
		// Load split diff for selected file
		item, _, found := m.filesPane.SelectedItem()
		if found {
			return m, loadSplitDiffCmd(m.git, item.Path)
		}
		return m, nil
	case " ": // space to toggle stage/unstage
		return m, m.toggleStageCmd()
	case "a":
		return m, stageAllCmd(m.git)
	case "d":
		return m.discardSelectedFile()
	}
	return m, nil
}

func (m model) handleBranchesKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		m.branchesPane.CursorDown()
		m.branchesPane.Refresh()
		return m, m.loadBranchDiff()
	case "k", "up":
		m.branchesPane.CursorUp()
		m.branchesPane.Refresh()
		return m, m.loadBranchDiff()
	case "g":
		m.branchesPane.CursorTop()
		m.branchesPane.Refresh()
		return m, m.loadBranchDiff()
	case "G":
		m.branchesPane.CursorBottom()
		m.branchesPane.Refresh()
		return m, m.loadBranchDiff()
	case "enter", " ":
		branch, found := m.branchesPane.SelectedBranch()
		if found && !branch.IsCurrent {
			return m, checkoutBranchCmd(m.git, branch.Name)
		}
	case "n":
		m.modal.OpenCreateBranch()
		return m, nil
	case "d":
		branch, found := m.branchesPane.SelectedBranch()
		if found {
			if branch.IsCurrent {
				m.modal.OpenError("Cannot delete current branch")
				return m, nil
			}
			m.modal.OpenConfirm("Delete branch "+branch.Name+"?", func() tea.Cmd {
				return deleteBranchCmd(m.git, branch.Name, false)
			})
		}
		return m, nil
	case "D":
		branch, found := m.branchesPane.SelectedBranch()
		if found {
			if branch.IsCurrent {
				m.modal.OpenError("Cannot delete current branch")
				return m, nil
			}
			m.modal.OpenConfirm("Force delete branch "+branch.Name+"?", func() tea.Cmd {
				return deleteBranchCmd(m.git, branch.Name, true)
			})
		}
		return m, nil
	}
	return m, nil
}

func (m model) handleCommitsKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		m.commitsPane.CursorDown()
		m.commitsPane.Refresh()
		return m, m.loadCommitDiff()
	case "k", "up":
		m.commitsPane.CursorUp()
		m.commitsPane.Refresh()
		return m, m.loadCommitDiff()
	case "g":
		m.commitsPane.CursorTop()
		m.commitsPane.Refresh()
		return m, m.loadCommitDiff()
	case "G":
		m.commitsPane.CursorBottom()
		m.commitsPane.Refresh()
		return m, m.loadCommitDiff()
	case "enter": // Focus main view to see full diff
		m.focus = ui.PaneMain
		m.mainViewSource = ui.PaneCommits
		m.layout = ui.CalculateLayout(m.layout.Width, m.layout.Height, m.focus)
		m.resizeComponents()
		m.refreshAllPanes()
		return m, m.loadCommitDiff()
	case "[", "]": // Toggle between Commits and Reflog (lazygit style)
		m.commitsPane.ToggleMode()
		m.commitsPane.Refresh()
		return m, m.loadCommitDiff()
	case "r": // Reset soft
		if m.commitsPane.SelectedIndex() == 0 && m.commitsPane.ItemCount() > 0 {
			m.modal.OpenConfirm("Undo last commit (keep staged)?", func() tea.Cmd {
				return resetSoftCmd(m.git, 1)
			})
		}
		return m, nil
	case "R": // Reset mixed
		if m.commitsPane.SelectedIndex() == 0 && m.commitsPane.ItemCount() > 0 {
			m.modal.OpenConfirm("Undo last commit (keep unstaged)?", func() tea.Cmd {
				return resetMixedCmd(m.git, 1)
			})
		}
		return m, nil
	}
	return m, nil
}

func (m model) handleStashKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		m.stashPane.CursorDown()
		m.stashPane.Refresh()
		return m, m.loadStashDiff()
	case "k", "up":
		m.stashPane.CursorUp()
		m.stashPane.Refresh()
		return m, m.loadStashDiff()
	case "g":
		m.stashPane.CursorTop()
		m.stashPane.Refresh()
		return m, m.loadStashDiff()
	case "G":
		m.stashPane.CursorBottom()
		m.stashPane.Refresh()
		return m, m.loadStashDiff()
	case "enter": // Focus main view to see full stash diff
		m.focus = ui.PaneMain
		m.mainViewSource = ui.PaneStash
		m.layout = ui.CalculateLayout(m.layout.Width, m.layout.Height, m.focus)
		m.resizeComponents()
		m.refreshAllPanes()
		return m, m.loadStashDiff()
	case " ": // Stash apply
		entry, found := m.stashPane.SelectedEntry()
		if found {
			return m, stashApplyCmd(m.git, entry.Ref)
		}
	case "p": // Stash pop
		entry, found := m.stashPane.SelectedEntry()
		if found {
			return m, stashPopCmd(m.git, entry.Ref)
		}
	case "d": // Stash drop
		entry, found := m.stashPane.SelectedEntry()
		if found {
			m.modal.OpenConfirm("Drop "+entry.Ref+"?", func() tea.Cmd {
				return stashDropCmd(m.git, entry.Ref)
			})
		}
		return m, nil
	}
	return m, nil
}

func (m model) handleCmdLogKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		m.cmdLogPane.ScrollDown(1)
	case "k", "up":
		m.cmdLogPane.ScrollUp(1)
	case "g":
		m.cmdLogPane.GotoTop()
	case "G":
		m.cmdLogPane.GotoBottom()
	}
	return m, nil
}

func (m model) handleMainKeys(key string) (tea.Model, tea.Cmd) {
	// In split mode (from Files pane), handle split pane navigation
	if m.mainViewSource == ui.PaneFiles {
		switch key {
		case "tab":
			m.splitDiffView.ToggleFocus()
			return m, nil
		case "j", "down":
			m.splitDiffView.ScrollDown(1)
		case "k", "up":
			m.splitDiffView.ScrollUp(1)
		case "d":
			m.splitDiffView.PageDown()
		case "u":
			m.splitDiffView.PageUp()
		case "g":
			m.splitDiffView.GotoTop()
		case "G":
			m.splitDiffView.GotoBottom()
		}
		return m, nil
	}

	// Normal single diff view
	switch key {
	case "j", "down":
		m.diffView.ScrollDown(1)
	case "k", "up":
		m.diffView.ScrollUp(1)
	case "d":
		m.diffView.PageDown()
	case "u":
		m.diffView.PageUp()
	case "g":
		m.diffView.GotoTop()
	case "G":
		m.diffView.GotoBottom()
	}
	return m, nil
}

// handleModalInput xử lý input khi modal đang mở
func (m model) handleModalInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch m.modal.Type() {
		case components.ModalCommit:
			switch key.String() {
			case "esc":
				m.modal.Close()
				return m, nil
			case "enter":
				msgVal := strings.TrimSpace(m.modal.InputValue())
				isAmend := m.modal.IsAmendMode()
				m.modal.Close()

				if isAmend {
					return m, commitAmendCmd(m.git, msgVal)
				}
				if msgVal == "" {
					m.modal.OpenError("Commit message is empty")
					return m, nil
				}
				return m, commitCmd(m.git, msgVal)
			}

		case components.ModalCreateBranch:
			switch key.String() {
			case "esc":
				m.modal.Close()
				return m, nil
			case "enter":
				name := strings.TrimSpace(m.modal.InputValue())
				m.modal.Close()
				if name == "" {
					m.modal.OpenError("Branch name is empty")
					return m, nil
				}
				return m, createBranchCmd(m.git, name)
			}

		case components.ModalConfirm:
			switch key.String() {
			case "esc", "n", "N":
				m.modal.Close()
				return m, nil
			case "y", "Y", "enter":
				action := m.modal.ConfirmAction()
				m.modal.Close()
				if action != nil {
					return m, action()
				}
				return m, nil
			}

		case components.ModalError:
			if key.String() == "esc" || key.String() == "enter" {
				m.modal.Close()
				return m, nil
			}
		}
	}

	// Forward to modal for text input
	cmd := m.modal.Update(msg)
	return m, cmd
}

// --- Helper methods ---

func (m model) nextFocusablePane() ui.PaneID {
	// Lazygit style: only sidebar panes (Files, Branches, Commits, Stash)
	order := []ui.PaneID{ui.PaneFiles, ui.PaneBranches, ui.PaneCommits, ui.PaneStash}
	for i, p := range order {
		if p == m.focus {
			return order[(i+1)%len(order)]
		}
	}
	return ui.PaneFiles
}

func (m model) prevFocusablePane() ui.PaneID {
	// Lazygit style: only sidebar panes (Files, Branches, Commits, Stash)
	order := []ui.PaneID{ui.PaneFiles, ui.PaneBranches, ui.PaneCommits, ui.PaneStash}
	for i, p := range order {
		if p == m.focus {
			return order[(i+len(order)-1)%len(order)]
		}
	}
	return ui.PaneFiles
}

func (m model) toggleStageCmd() tea.Cmd {
	item, isStaged, found := m.filesPane.SelectedItem()
	if !found {
		return nil
	}
	if isStaged {
		return unstageFileCmd(m.git, item.Path)
	}
	return stageFileCmd(m.git, item.Path)
}

func (m model) loadCommitDiff() tea.Cmd {
	// Works for both commits and reflog modes
	hash, found := m.commitsPane.SelectedHash()
	if !found {
		return nil
	}
	return loadShowCommitCmd(m.git, hash)
}

func (m model) loadBranchDiff() tea.Cmd {
	branch, found := m.branchesPane.SelectedBranch()
	if !found {
		return nil
	}
	return loadBranchDiffCmd(m.git, branch.Name)
}

func (m model) loadStashDiff() tea.Cmd {
	entry, found := m.stashPane.SelectedEntry()
	if !found {
		return nil
	}
	return loadStashDiffCmd(m.git, entry.Ref)
}

func (m model) discardSelectedFile() (tea.Model, tea.Cmd) {
	if m.filesPane.IsSelectedStaged() {
		m.modal.OpenError("Cannot discard staged file. Unstage first (space)")
		return m, nil
	}

	item, _, found := m.filesPane.SelectedItem()
	if !found {
		return m, nil
	}

	isUntracked := item.Status == "?"
	m.modal.OpenConfirm("Discard changes to "+item.Path+"?", func() tea.Cmd {
		return discardFileCmd(m.git, item.Path, isUntracked)
	})

	return m, nil
}
