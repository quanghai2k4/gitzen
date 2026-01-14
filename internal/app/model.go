package app

import (
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitzen/internal/components"
	"gitzen/internal/git"
	"gitzen/internal/ui"
)

// model là orchestrator chính, sử dụng components để render và xử lý
type model struct {
	repoRoot string
	repoName string
	git      git.Runner

	// Current focus
	focus ui.PaneID

	// Components
	statusPane   *components.StatusPane
	filesPane    *components.FilesPane
	branchesPane *components.BranchesPane
	commitsPane  *components.CommitsPane
	stashPane    *components.StashPane
	diffView     *components.DiffView
	cmdLogPane   *components.CmdLogPane
	modal        *components.Modal

	// UI
	styles ui.Styles
	layout ui.Layout

	// Messages
	statusMsg  string
	lastGitCmd string
}

func NewModel(repoRoot string) tea.Model {
	styles := ui.DefaultStyles

	m := model{
		repoRoot: repoRoot,
		repoName: filepath.Base(repoRoot),
		git:      git.New(repoRoot),
		focus:    ui.PaneFiles,
		styles:   styles,

		// Initialize components
		statusPane:   components.NewStatusPane(styles),
		filesPane:    components.NewFilesPane(styles),
		branchesPane: components.NewBranchesPane(styles),
		commitsPane:  components.NewCommitsPane(styles),
		stashPane:    components.NewStashPane(styles),
		diffView:     components.NewDiffView(styles),
		cmdLogPane:   components.NewCmdLogPane(styles),
		modal:        components.NewModal(styles),
	}

	// Set initial repo info
	m.statusPane.SetData(m.repoName, "")

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		loadStatusCmd(m.git),
		loadCommitsCmd(m.git),
		loadBranchCmd(m.git),
		loadBranchesCmd(m.git),
		loadStashCmd(m.git),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout = ui.CalculateLayout(msg.Width, msg.Height, m.focus)
		m.resizeComponents()
		m.refreshAllPanes()
		return m, nil

	case statusLoadedMsg:
		m.filesPane.SetData(msg.Status.Staged, msg.Status.Unstaged)
		return m, m.loadDiffForCurrentPane()

	case commitsLoadedMsg:
		m.commitsPane.SetData(msg.Commits)
		return m, nil

	case branchLoadedMsg:
		m.statusPane.SetData(m.repoName, msg.Branch)
		return m, nil

	case branchesLoadedMsg:
		m.branchesPane.SetData(msg.Branches)
		return m, nil

	case stashLoadedMsg:
		m.stashPane.SetData(msg.Entries)
		return m, nil

	case cmdLogMsg:
		m.cmdLogPane.AddEntry(string(msg))
		return m, nil

	case diffLoadedMsg:
		m.diffView.SetDiff(msg.Diff)
		return m, nil

	case gitCmdMsg:
		m.lastGitCmd = string(msg)
		return m, nil

	case errMsg:
		m.modal.OpenError(string(msg))
		return m, nil

	case statusToastMsg:
		m.statusMsg = string(msg)
		m.cmdLogPane.AddEntry(string(msg))
		return m, tea.Batch(
			loadStatusCmd(m.git),
			loadCommitsCmd(m.git),
			loadBranchCmd(m.git),
			loadBranchesCmd(m.git),
			loadStashCmd(m.git),
		)

	case gitResultMsg:
		// Log the git command
		m.cmdLogPane.AddEntry(msg.Cmd)
		m.lastGitCmd = msg.Cmd

		if msg.Err != nil {
			m.modal.OpenError(msg.Err.Error())
			return m, nil
		}

		// Show result as status toast
		m.statusMsg = msg.Result
		return m, tea.Batch(
			loadStatusCmd(m.git),
			loadCommitsCmd(m.git),
			loadBranchCmd(m.git),
			loadBranchesCmd(m.git),
			loadStashCmd(m.git),
		)
	}

	// Handle modal input first
	if m.modal.IsOpen() {
		return m.handleModalInput(msg)
	}

	// Handle key input
	if key, ok := msg.(tea.KeyMsg); ok {
		return m.handleKeys(key)
	}

	return m, nil
}

func (m model) View() string {
	if m.layout.Width == 0 || m.layout.Height == 0 {
		return ""
	}

	if m.layout.Width < ui.MinTerminalWidth || m.layout.Height < ui.MinTerminalHeight {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Render("Terminal too small (min 60x15)")
	}

	// === Sidebar (left) ===
	statusBox := m.statusPane.RenderBox(m.focus == ui.PaneStatus, m.styles)
	filesBox := m.filesPane.RenderBox(m.focus == ui.PaneFiles, m.styles)
	branchesBox := m.branchesPane.RenderBox(m.focus == ui.PaneBranches, m.styles)
	commitsBox := m.commitsPane.RenderBox(m.focus == ui.PaneCommits, m.styles)
	stashBox := m.stashPane.RenderBox(m.focus == ui.PaneStash, m.styles)

	sidebar := lipgloss.JoinVertical(lipgloss.Left,
		statusBox, filesBox, branchesBox, commitsBox, stashBox)

	// === Right side: Main + CmdLog ===
	mainBox := m.diffView.RenderBox(m.focus == ui.PaneMain, m.styles)
	cmdLogBox := m.cmdLogPane.RenderBox(m.focus == ui.PaneCmdLog, m.styles)
	rightSide := lipgloss.JoinVertical(lipgloss.Left, mainBox, cmdLogBox)

	// Join sidebar and right side
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, rightSide)

	// === Info bar ===
	infoBar := m.renderInfoBar()

	out := body + "\n" + infoBar

	// Modal overlay
	if m.modal.IsOpen() {
		out = components.OverlayCentered(out, m.modal.View(), m.layout.Width)
	}

	return out
}

// resizeComponents cập nhật kích thước cho tất cả components
func (m *model) resizeComponents() {
	m.statusPane.SetSize(m.layout.SidebarWidth, m.layout.StatusHeight)
	m.filesPane.SetSize(m.layout.SidebarWidth, m.layout.FilesHeight)
	m.branchesPane.SetSize(m.layout.SidebarWidth, m.layout.BranchHeight)
	m.commitsPane.SetSize(m.layout.SidebarWidth, m.layout.CommitsHeight)
	m.stashPane.SetSize(m.layout.SidebarWidth, m.layout.StashHeight)
	m.diffView.SetSize(m.layout.MainWidth, m.layout.MainHeight)
	m.cmdLogPane.SetSize(m.layout.MainWidth, m.layout.CmdLogHeight)
}

// refreshAllPanes cập nhật focus state và re-render tất cả panes
func (m *model) refreshAllPanes() {
	// Update focus states
	m.filesPane.SetFocus(m.focus == ui.PaneFiles)
	m.branchesPane.SetFocus(m.focus == ui.PaneBranches)
	m.commitsPane.SetFocus(m.focus == ui.PaneCommits)
	m.stashPane.SetFocus(m.focus == ui.PaneStash)
	m.diffView.SetFocus(m.focus == ui.PaneMain)
	m.cmdLogPane.SetFocus(m.focus == ui.PaneCmdLog)

	// Refresh content
	m.filesPane.Refresh()
	m.branchesPane.Refresh()
	m.commitsPane.Refresh()
	m.stashPane.Refresh()
	m.statusPane.Refresh()
	m.cmdLogPane.Refresh()
}

// renderInfoBar renders the info bar at bottom
func (m model) renderInfoBar() string {
	optStyle := m.styles.OptionsStyle
	infoStyle := m.styles.InfoStyle
	dimStyle := m.styles.DimStyle

	// Keybindings based on focus
	var opts string
	switch m.focus {
	case ui.PaneFiles:
		opts = "space: stage | a: all | c: commit | A: amend | d: discard"
	case ui.PaneBranches:
		opts = "space: checkout | n: new | d: delete | D: force delete"
	case ui.PaneCommits:
		opts = "enter: view | r: undo (staged) | R: undo (unstaged)"
	case ui.PaneStash:
		opts = "space: apply | p: pop | d: drop"
	case ui.PaneCmdLog:
		opts = "j/k: scroll | g/G: top/bottom"
	case ui.PaneMain:
		opts = "j/k: scroll | d/u: page | g/G: top/bottom"
	default:
		opts = "tab: switch | p: pull | P: push | f: fetch | q: quit"
	}

	left := optStyle.Render(opts)

	// Right side
	var right string
	if m.statusMsg != "" {
		right = infoStyle.Render(m.statusMsg)
	} else if m.lastGitCmd != "" {
		right = dimStyle.Render(m.lastGitCmd)
	} else {
		right = infoStyle.Render("gitzen")
	}

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	space := m.layout.Width - leftW - rightW
	if space < 1 {
		space = 1
	}

	return left + strings.Repeat(" ", space) + right
}

// loadDiffForCurrentPane loads diff based on current focus and selection
func (m model) loadDiffForCurrentPane() tea.Cmd {
	switch m.focus {
	case ui.PaneFiles:
		item, staged, found := m.filesPane.SelectedItem()
		if !found {
			return func() tea.Msg { return diffLoadedMsg{Diff: "(no file selected)"} }
		}
		return loadDiffCmd(m.git, item.Path, staged)

	case ui.PaneCommits:
		commit, found := m.commitsPane.SelectedCommit()
		if !found {
			return nil
		}
		return loadShowCommitCmd(m.git, commit.Hash)

	case ui.PaneBranches:
		branch, found := m.branchesPane.SelectedBranch()
		if !found {
			return nil
		}
		return loadBranchDiffCmd(m.git, branch.Name)

	case ui.PaneStash:
		entry, found := m.stashPane.SelectedEntry()
		if !found {
			return nil
		}
		return loadStashDiffCmd(m.git, entry.Ref)

	default:
		return nil
	}
}
