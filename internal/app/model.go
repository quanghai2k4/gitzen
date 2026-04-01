package app

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"gitzen/internal/background"
	"gitzen/internal/components"
	"gitzen/internal/config"
	"gitzen/internal/git"
	"gitzen/internal/ui"
)

// model là orchestrator chính, sử dụng components để render và xử lý
type model struct {
	repoRoot string
	repoName string
	git      git.Runner

	// Background operations
	backgroundManager *background.Manager
	backgroundCancel  context.CancelFunc

	// Current focus
	focus ui.PaneID

	// Components
	statusPane    *components.StatusPane
	filesPane     *components.FilesPane
	branchesPane  *components.BranchesPane
	commitsPane   *components.CommitsPane
	stashPane     *components.StashPane
	diffView      *components.DiffView
	splitDiffView *components.SplitDiffView
	hunkView      *components.HunkView
	cmdLogPane    *components.CmdLogPane
	modal         *components.Modal
	toastManager  *components.ToastManager

	// Track which pane we entered Main from (for split mode)
	mainViewSource ui.PaneID

	// Track hunk view mode
	inHunkView bool

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
		repoRoot:   repoRoot,
		repoName:   filepath.Base(repoRoot),
		git:        git.New(repoRoot),
		focus:      ui.PaneFiles,
		styles:     styles,
		inHunkView: false,

		// Initialize background operations
		backgroundManager: background.New(git.New(repoRoot)),

		// Initialize components
		statusPane:    components.NewStatusPane(styles),
		filesPane:     components.NewFilesPane(styles),
		branchesPane:  components.NewBranchesPane(styles),
		commitsPane:   components.NewCommitsPane(styles),
		stashPane:     components.NewStashPane(styles),
		diffView:      components.NewDiffView(styles),
		splitDiffView: components.NewSplitDiffView(styles),
		hunkView:      components.NewHunkView(styles),
		cmdLogPane:    components.NewCmdLogPane(styles),
		modal:         components.NewModal(styles),
		toastManager:  components.NewToastManager(styles),
	}

	// Set initial repo info
	m.statusPane.SetData(m.repoName, "")

	// Initialize file watcher với configuration
	repoConfig, err := config.LoadRepoConfig(repoRoot)
	if err != nil {
		m.cmdLogPane.AddEntry("warning: failed to load config, using defaults: " + err.Error())
		repoConfig = config.NewDefaultConfig()
	}

	if err := m.backgroundManager.InitFileWatcher(repoRoot, repoConfig.FileWatch.Enabled); err != nil {
		// Log warning but don't fail - file watching is not critical
		m.cmdLogPane.AddEntry("warning: failed to initialize file watcher: " + err.Error())
	}

	return m
}

func (m model) Init() tea.Cmd {
	// Create background context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	m.backgroundCancel = cancel

	commands := []tea.Cmd{
		loadStatusCmd(m.git),
		loadCommitsCmd(m.git),
		loadReflogCmd(m.git),
		loadBranchCmd(m.git),
		loadBranchesCmd(m.git),
		loadStashCmd(m.git),
		m.backgroundManager.Start(ctx),
		m.backgroundManager.StartFileWatcher(ctx), // Start file watching
	}

	// Add startup fetch if in valid repository
	if m.repoRoot != "" {
		commands = append(commands, startupFetchCmd())
	}

	return tea.Batch(commands...)
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

	case reflogLoadedMsg:
		m.commitsPane.SetReflogData(msg.Entries)
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

	case backgroundTickMsg:
		// Background timer tick - execute auto fetch and continue timer loop
		return m, tea.Batch(
			updateFetchStatusCmd(components.FetchInProgress),
			m.executeAutoFetchCmd(),
			backgroundTickCmd(),
		)

	case cmdLogMsg:
		m.cmdLogPane.AddEntry(string(msg))
		return m, nil

	case diffLoadedMsg:
		m.diffView.SetDiffWithContext(msg.Diff, components.DiffContext(msg.Context), msg.Subtitle)
		return m, nil

	case splitDiffLoadedMsg:
		m.splitDiffView.SetDiffs(msg.Unstaged, msg.Staged, msg.FilePath)
		return m, nil

	case hunksLoadedMsg:
		m.hunkView.SetHunks(msg.Hunks, msg.Path, msg.Staged)
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
			loadReflogCmd(m.git),
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
			loadReflogCmd(m.git),
			loadBranchCmd(m.git),
			loadBranchesCmd(m.git),
			loadStashCmd(m.git),
		)

	case startupFetchMsg:
		// Handle startup fetch trigger
		return m, tea.Batch(
			updateFetchStatusCmd(components.FetchInProgress),
			handleStartupFetch(m),
		)

	case startupFetchResultMsg:
		// Log startup fetch result, update status, and show toast
		var cmd tea.Cmd
		if msg.Success {
			if msg.Skipped {
				m.cmdLogPane.AddEntry("startup fetch: " + msg.Message)
				cmd = updateFetchStatusCmd(components.FetchIdle) // Keep idle for skipped
			} else {
				m.cmdLogPane.AddEntry("startup fetch: " + msg.Message)
				m.statusMsg = "startup fetch completed"
				cmd = tea.Batch(
					updateFetchStatusCmd(components.FetchSuccess),
					clearFetchStatusCmd(),
					addToastCmd("Startup fetch completed", components.ToastSuccess, 3*time.Second),
					loadCommitCountsCmd(m.git), // Load commit counts after successful fetch
				)
			}
		} else {
			m.cmdLogPane.AddEntry("startup fetch failed: " + msg.Message)
			cmd = tea.Batch(
				updateFetchStatusCmd(components.FetchError),
				clearFetchStatusCmd(),
				addToastCmd("Startup fetch failed: "+msg.Message, components.ToastError, 5*time.Second),
			)
		}
		return m, cmd

	case autoFetchResultMsg:
		// Log auto fetch result, update status, and show toast
		var cmd tea.Cmd
		if msg.Success {
			if !msg.Skipped {
				m.cmdLogPane.AddEntry("auto fetch: " + msg.Message)
				cmd = tea.Batch(
					updateFetchStatusCmd(components.FetchSuccess),
					clearFetchStatusCmd(),
					addToastCmd("Auto fetch completed", components.ToastSuccess, 3*time.Second),
					loadCommitCountsCmd(m.git), // Load commit counts after successful fetch
				)
			} else {
				// Keep current status for skipped operations
				cmd = updateFetchStatusCmd(components.FetchIdle)
			}
		} else {
			// Show error toast for background fetch failures
			cmd = tea.Batch(
				updateFetchStatusCmd(components.FetchError),
				clearFetchStatusCmd(),
				addToastCmd("Auto fetch failed", components.ToastError, 5*time.Second),
			)
		}
		return m, cmd

	case fetchStatusUpdateMsg:
		// Update fetch status in status pane
		m.statusPane.SetFetchStatus(msg.Status)
		if msg.Status == components.FetchSuccess {
			m.statusPane.SetLastFetchTime(msg.Timestamp)
		}
		return m, nil

	case fetchStatusClearMsg:
		// Clear fetch status back to idle (only if not currently in progress)
		if m.statusPane.GetFetchStatus() != components.FetchInProgress {
			m.statusPane.SetFetchStatus(components.FetchIdle)
		}
		return m, nil

	case toastAddMsg:
		// Add new toast notification
		m.toastManager.AddToastNotification(msg.toast)
		return m, nil

	case toastExpiredMsg:
		// Remove expired toast
		m.toastManager.RemoveToast(msg.id)
		return m, nil

	case commitCountsLoadedMsg:
		// Update branches pane with commit counts
		m.branchesPane.SetCommitCounts(msg.Counts)

		// Update status bar with new commits summary
		totalAhead := 0
		for _, count := range msg.Counts {
			totalAhead += count.Ahead
		}
		if totalAhead > 0 {
			m.statusPane.SetNewCommitsAvailable(totalAhead)
		}
		return m, nil

	case background.FileWatchEventMsg:
		// File system change detected, refresh git status
		ctx, cancel := context.WithCancel(context.Background())
		m.backgroundCancel = cancel

		return m, tea.Batch(
			fileWatchRefreshCmd(m.git),
			m.backgroundManager.StartFileWatcher(ctx), // Continue watching
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
	var mainBox string
	if m.inHunkView {
		mainBox = m.hunkView.RenderBox(true, m.styles)
	} else if m.focus == ui.PaneMain && m.mainViewSource == ui.PaneFiles {
		// Split view for Files: Unstaged + Staged
		mainBox = m.renderSplitMainBox()
	} else {
		mainBox = m.diffView.RenderBox(m.focus == ui.PaneMain, m.styles)
	}
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

	// Toast overlay - render toasts over everything except modals
	toastContent := m.toastManager.View(m.layout.Width, m.layout.Height)
	if toastContent != "" {
		out = m.renderWithToasts(out, toastContent)
	}

	return out
}

// renderWithToasts overlays toast notifications on base content
func (m model) renderWithToasts(baseContent, toastContent string) string {
	if toastContent == "" {
		return baseContent
	}

	toastLines := strings.Split(toastContent, "\n")
	toastHeight := len(toastLines)
	toastWidth := 0

	// Tìm width lớn nhất của toasts
	for _, line := range toastLines {
		if w := ansi.StringWidth(line); w > toastWidth {
			toastWidth = w
		}
	}

	baseLines := strings.Split(baseContent, "\n")

	// Tính vị trí bottom-right (2 chars từ mép, trên info bar)
	startX := m.layout.Width - toastWidth - 2
	startY := len(baseLines) - toastHeight - 2 // -1 cho info bar, -1 cho spacing

	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	// Overlay toast lên base content
	for i, toastLine := range toastLines {
		targetY := startY + i
		if targetY < len(baseLines) && targetY >= 0 {
			baseLine := baseLines[targetY]
			baseWidth := ansi.StringWidth(baseLine)
			toastLineWidth := ansi.StringWidth(toastLine)

			var newLine string

			// Add left part of base line
			if startX > 0 {
				if baseWidth >= startX {
					newLine = ansi.Truncate(baseLine, startX, "")
				} else {
					newLine = baseLine + strings.Repeat(" ", startX-baseWidth)
				}
			}

			// Add toast line
			newLine += toastLine

			// Add right part of base line if it exists
			endX := startX + toastLineWidth
			if baseWidth > endX {
				rightPart := ansi.Cut(baseLine, endX, baseWidth)
				newLine += rightPart
			}

			baseLines[targetY] = newLine
		}
	}

	return strings.Join(baseLines, "\n")
}

// resizeComponents cập nhật kích thước cho tất cả components
func (m *model) resizeComponents() {
	m.statusPane.SetSize(m.layout.SidebarWidth, m.layout.StatusHeight)
	m.filesPane.SetSize(m.layout.SidebarWidth, m.layout.FilesHeight)
	m.branchesPane.SetSize(m.layout.SidebarWidth, m.layout.BranchHeight)
	m.commitsPane.SetSize(m.layout.SidebarWidth, m.layout.CommitsHeight)
	m.stashPane.SetSize(m.layout.SidebarWidth, m.layout.StashHeight)
	m.diffView.SetSize(m.layout.MainWidth, m.layout.MainHeight)
	m.splitDiffView.SetSize(m.layout.MainWidth, m.layout.MainHeight)
	m.hunkView.SetSize(m.layout.MainWidth, m.layout.MainHeight)
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
	m.hunkView.SetFocus(m.inHunkView)
	m.cmdLogPane.SetFocus(m.focus == ui.PaneCmdLog)

	// Refresh content
	m.filesPane.Refresh()
	m.branchesPane.Refresh()
	m.commitsPane.Refresh()
	m.stashPane.Refresh()
	m.statusPane.Refresh()
	m.hunkView.Refresh()
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
		opts = "[/]: commits/reflog | enter: view | r/R: undo"
	case ui.PaneStash:
		opts = "space: apply | p: pop | d: drop"
	case ui.PaneCmdLog:
		opts = "j/k: scroll | g/G: top/bottom"
	case ui.PaneMain:
		if m.inHunkView {
			opts = "space: stage/unstage | j/k: navigate | esc: exit"
		} else if m.mainViewSource == ui.PaneFiles {
			opts = "tab: switch pane | j/k: scroll | d/u: page | g/G: top/bottom"
		} else {
			opts = "j/k: scroll | d/u: page | g/G: top/bottom"
		}
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

// renderSplitMainBox renders the split diff view for Files pane
func (m model) renderSplitMainBox() string {
	return m.splitDiffView.View()
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

// loadHunksForCurrentFile loads hunks for the selected file
func (m model) loadHunksForCurrentFile() tea.Cmd {
	item, staged, found := m.filesPane.SelectedItem()
	if !found {
		return nil
	}
	return loadHunksCmd(m.git, item.Path, staged)
}

// executeAutoFetchCmd executes background auto fetch using the background manager
func (m model) executeAutoFetchCmd() tea.Cmd {
	return m.backgroundManager.ExecuteAutoFetch(m.repoRoot)
}
