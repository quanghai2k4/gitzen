package app

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"gitzen/internal/git"
	"gitzen/internal/tui"
)

// Pane IDs - matching lazygit exactly: Status, Files, Branches, Commits, Stash
type pane int

const (
	paneStatus pane = iota
	paneFiles
	paneBranches
	paneCommits
	paneStash
	paneMain // Diff view (right panel)
)

// Colors matching lazygit default theme
var (
	// Border colors
	activeBorderColor   = lipgloss.Color("2") // Green (ANSI 2)
	inactiveBorderColor = lipgloss.Color("7") // Default/white

	// Selection colors (like lazygit: blue background)
	selectedBgColor         = lipgloss.Color("4")  // Blue
	selectedFgColor         = lipgloss.Color("15") // Bright white
	inactiveSelectedFgColor = lipgloss.Color("15") // Bold white when inactive

	// File status colors
	colorUntracked = lipgloss.Color("1") // Red for untracked
	colorModified  = lipgloss.Color("3") // Yellow
	colorStaged    = lipgloss.Color("2") // Green
	colorDeleted   = lipgloss.Color("1") // Red
	colorRenamed   = lipgloss.Color("6") // Cyan
	colorConflict  = lipgloss.Color("5") // Magenta

	// Commit colors
	colorHash   = lipgloss.Color("3") // Yellow for hash
	colorAuthor = lipgloss.Color("6") // Cyan for author
	colorDate   = lipgloss.Color("4") // Blue for date

	// Branch colors
	colorBranchLocal  = lipgloss.Color("6") // Cyan
	colorBranchRemote = lipgloss.Color("5") // Magenta
	colorBranchHead   = lipgloss.Color("2") // Green (current branch)

	// Info bar colors
	optionsColor = lipgloss.Color("4") // Blue for keybindings
	infoColor    = lipgloss.Color("2") // Green for info

	dimColor = lipgloss.Color("8") // Dim gray
)

// Layout settings matching lazygit
const (
	statusPaneFixedHeight   = 3 // Status always 3 lines
	stashPaneMinHeight      = 3 // Stash min 3 lines when not focused
	accordionModeEnabled    = true
	expandedSidePanelWeight = 2
	normalSidePanelWeight   = 1
	sidePanelWidthRatio     = 0.3333 // ~1/3 of screen
)

type model struct {
	repoRoot string
	repoName string
	git      git.Runner

	focus pane

	// Viewports for scrollable content
	filesVP    viewport.Model
	branchesVP viewport.Model
	commitsVP  viewport.Model
	stashVP    viewport.Model
	mainVP     viewport.Model

	// Data
	unstagedItems []git.FileItem
	stagedItems   []git.FileItem
	commitItems   []git.CommitItem
	branchName    string
	branches      []git.Branch
	stashItems    []git.StashEntry

	// Cursors
	filesCursor    int
	branchesCursor int
	commitsCursor  int
	stashCursor    int

	// Dimensions
	width  int
	height int

	// Calculated layout sizes
	sidebarW  int
	mainW     int
	statusH   int
	filesH    int
	branchesH int
	commitsH  int
	stashH    int
	mainH     int
	infoBarH  int

	// Messages
	statusMsg  string
	errorMsg   string
	lastGitCmd string

	// Commit modal
	commitMode bool
	amendMode  bool
	commitIn   textinput.Model

	// Create branch modal
	createBranchMode bool
	branchIn         textinput.Model

	// Confirm dialog
	confirmMode    bool
	confirmTitle   string
	confirmAction  func() tea.Cmd
	confirmYesText string

	diffStyler tui.DiffStyler
}

func NewModel(repoRoot string) tea.Model {
	m := model{
		repoRoot:   repoRoot,
		repoName:   filepath.Base(repoRoot),
		git:        git.New(repoRoot),
		focus:      paneFiles,
		diffStyler: tui.DefaultDiffStyler(),
		infoBarH:   1,
	}

	m.filesVP = viewport.New(0, 0)
	m.branchesVP = viewport.New(0, 0)
	m.commitsVP = viewport.New(0, 0)
	m.stashVP = viewport.New(0, 0)
	m.mainVP = viewport.New(0, 0)

	m.commitIn = textinput.New()
	m.commitIn.Placeholder = "Commit message"
	m.commitIn.CharLimit = 200
	m.commitIn.Prompt = "> "

	m.branchIn = textinput.New()
	m.branchIn.Placeholder = "Branch name"
	m.branchIn.CharLimit = 100
	m.branchIn.Prompt = "> "

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
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		m.refreshAllViews()
		return m, nil
	case statusLoadedMsg:
		m.unstagedItems = msg.Status.Unstaged
		m.stagedItems = msg.Status.Staged
		if m.filesCursor >= len(m.unstagedItems)+len(m.stagedItems) {
			m.filesCursor = max(0, len(m.unstagedItems)+len(m.stagedItems)-1)
		}
		m.refreshAllViews()
		return m, m.loadDiffForCurrentPane()
	case commitsLoadedMsg:
		m.commitItems = msg.Commits
		if m.commitsCursor >= len(m.commitItems) {
			m.commitsCursor = max(0, len(m.commitItems)-1)
		}
		m.refreshAllViews()
		return m, nil
	case branchLoadedMsg:
		m.branchName = msg.Branch
		return m, nil
	case branchesLoadedMsg:
		m.branches = msg.Branches
		if m.branchesCursor >= len(m.branches) {
			m.branchesCursor = max(0, len(m.branches)-1)
		}
		m.refreshAllViews()
		return m, nil
	case stashLoadedMsg:
		m.stashItems = msg.Entries
		if m.stashCursor >= len(m.stashItems) {
			m.stashCursor = max(0, len(m.stashItems)-1)
		}
		m.refreshAllViews()
		return m, nil
	case diffLoadedMsg:
		m.mainVP.SetContent(m.diffStyler.Colorize(msg.Diff))
		m.mainVP.GotoTop()
		return m, nil
	case gitCmdMsg:
		m.lastGitCmd = string(msg)
		return m, nil
	case errMsg:
		m.errorMsg = string(msg)
		return m, nil
	case statusToastMsg:
		m.statusMsg = string(msg)
		// Reload all data after any git operation
		return m, tea.Batch(
			loadStatusCmd(m.git),
			loadCommitsCmd(m.git),
			loadBranchCmd(m.git),
			loadBranchesCmd(m.git),
			loadStashCmd(m.git),
		)
	}

	if m.commitMode {
		return m.updateCommitMode(msg)
	}

	if m.createBranchMode {
		return m.updateCreateBranchMode(msg)
	}

	if m.confirmMode {
		return m.updateConfirmMode(msg)
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		return m.handleKeys(key)
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	if m.width < 60 || m.height < 15 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("Terminal too small (min 60x15)")
	}

	// === Sidebar (left) - 5 panes like lazygit ===
	statusBox := m.renderPane("Status", m.renderStatusContent(), paneStatus, m.sidebarW, m.statusH)
	filesBox := m.renderPane("Files", m.filesVP.View(), paneFiles, m.sidebarW, m.filesH)
	branchesBox := m.renderPane("Branches", m.branchesVP.View(), paneBranches, m.sidebarW, m.branchesH)
	commitsBox := m.renderPane("Commits", m.commitsVP.View(), paneCommits, m.sidebarW, m.commitsH)
	stashBox := m.renderPane("Stash", m.stashVP.View(), paneStash, m.sidebarW, m.stashH)

	sidebar := lipgloss.JoinVertical(lipgloss.Left, statusBox, filesBox, branchesBox, commitsBox, stashBox)

	// === Main panel (right) ===
	mainBox := m.renderPane("Main", m.mainVP.View(), paneMain, m.mainW, m.mainH)

	// Join sidebar and main
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainBox)

	// === Info bar (bottom) ===
	infoBar := m.renderInfoBar()

	out := body + "\n" + infoBar

	// Modals overlay (centered)
	if m.errorMsg != "" {
		out = m.overlayModalCentered(out, m.renderErrorModal())
	}
	if m.commitMode {
		out = m.overlayModalCentered(out, m.renderCommitModal())
	}
	if m.createBranchMode {
		out = m.overlayModalCentered(out, m.renderCreateBranchModal())
	}
	if m.confirmMode {
		out = m.overlayModalCentered(out, m.renderConfirmModal())
	}

	return out
}

func (m *model) resize() {
	m.infoBarH = 1
	totalH := m.height - m.infoBarH

	// Sidebar width: ~1/3 like lazygit
	m.sidebarW = int(float64(m.width) * sidePanelWidthRatio)
	if m.sidebarW < 30 {
		m.sidebarW = 30
	}
	if m.sidebarW > 60 {
		m.sidebarW = 60
	}
	m.mainW = m.width - m.sidebarW
	m.mainH = totalH

	// Sidebar vertical layout like lazygit:
	// Status: fixed 3
	// Files, Branches, Commits: flexible with accordion
	// Stash: fixed 3 (unless focused)
	m.statusH = statusPaneFixedHeight

	// Stash height
	if m.focus == paneStash {
		m.stashH = 0 // Will be calculated with weight
	} else {
		m.stashH = stashPaneMinHeight
	}

	remainH := totalH - m.statusH - m.stashH

	// Calculate flexible panes with accordion
	flexPanes := []pane{paneFiles, paneBranches, paneCommits}
	if m.focus == paneStash {
		flexPanes = append(flexPanes, paneStash)
	}

	weights := make([]int, len(flexPanes))
	totalWeight := 0
	for i, p := range flexPanes {
		if accordionModeEnabled && m.focus == p {
			weights[i] = expandedSidePanelWeight
		} else {
			weights[i] = normalSidePanelWeight
		}
		totalWeight += weights[i]
	}

	if totalWeight == 0 {
		totalWeight = 1
	}
	unitH := remainH / totalWeight

	// Assign heights
	idx := 0
	m.filesH = unitH * weights[idx]
	idx++
	m.branchesH = unitH * weights[idx]
	idx++
	m.commitsH = unitH * weights[idx]
	idx++

	if m.focus == paneStash {
		m.stashH = remainH - m.filesH - m.branchesH - m.commitsH
	}

	// Adjust last pane to fill remaining space
	usedH := m.statusH + m.filesH + m.branchesH + m.commitsH + m.stashH
	if usedH < totalH {
		m.commitsH += totalH - usedH
	} else if usedH > totalH {
		m.commitsH -= usedH - totalH
	}

	// Update viewport sizes (content area = pane size - 2 for border)
	m.filesVP.Width = m.sidebarW - 2
	m.filesVP.Height = max(1, m.filesH-2)
	m.branchesVP.Width = m.sidebarW - 2
	m.branchesVP.Height = max(1, m.branchesH-2)
	m.commitsVP.Width = m.sidebarW - 2
	m.commitsVP.Height = max(1, m.commitsH-2)
	m.stashVP.Width = m.sidebarW - 2
	m.stashVP.Height = max(1, m.stashH-2)
	m.mainVP.Width = m.mainW - 2
	m.mainVP.Height = max(1, m.mainH-2)
}

func (m *model) refreshAllViews() {
	m.filesVP.SetContent(m.renderFilesContent())
	m.branchesVP.SetContent(m.renderBranchesContent())
	m.commitsVP.SetContent(m.renderCommitsContent())
	m.stashVP.SetContent(m.renderStashContent())
}

// renderPane renders a pane with title in border exactly like lazygit
func (m model) renderPane(title, content string, p pane, w, h int) string {
	if h <= 0 {
		return ""
	}

	focused := m.focus == p

	// Border color: green when focused, default otherwise
	borderColor := inactiveBorderColor
	titleStyle := lipgloss.NewStyle().Foreground(inactiveBorderColor)
	if focused {
		borderColor = activeBorderColor
		titleStyle = lipgloss.NewStyle().Foreground(activeBorderColor).Bold(true)
	}

	// Content dimensions
	innerW := w - 2
	innerH := h - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 0 {
		innerH = 0
	}

	// Truncate/pad content
	lines := strings.Split(content, "\n")
	if len(lines) > innerH {
		lines = lines[:innerH]
	}
	for len(lines) < innerH {
		lines = append(lines, "")
	}
	for i, line := range lines {
		if lipgloss.Width(line) > innerW {
			lines[i] = truncateString(line, innerW)
		}
	}

	// Border characters (rounded like lazygit default)
	topLeft := "╭"
	topRight := "╮"
	hLine := "─"
	vLine := "│"
	botLeft := "╰"
	botRight := "╯"

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Build top line: ╭─ Title ─────────────╮
	titleRendered := titleStyle.Render(" " + title + " ")
	titleLen := lipgloss.Width(titleRendered)
	remainingWidth := innerW - titleLen
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	topLine := borderStyle.Render(topLeft) + titleRendered + borderStyle.Render(strings.Repeat(hLine, remainingWidth)+topRight)

	// Build content lines
	var contentLines []string
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		padding := innerW - lineWidth
		if padding < 0 {
			padding = 0
		}
		paddedLine := line + strings.Repeat(" ", padding)
		contentLines = append(contentLines, borderStyle.Render(vLine)+paddedLine+borderStyle.Render(vLine))
	}

	// Build bottom line
	bottomLine := borderStyle.Render(botLeft + strings.Repeat(hLine, innerW) + botRight)

	if len(contentLines) == 0 {
		return topLine + "\n" + bottomLine
	}

	return topLine + "\n" + strings.Join(contentLines, "\n") + "\n" + bottomLine
}

func (m model) renderStatusContent() string {
	branch := m.branchName
	if branch == "" {
		branch = "master"
	}

	repoStyle := lipgloss.NewStyle().Foreground(colorBranchHead).Bold(true)
	branchStyle := lipgloss.NewStyle().Foreground(colorBranchLocal)

	return repoStyle.Render(m.repoName) + " → " + branchStyle.Render(branch)
}

func (m model) renderFilesContent() string {
	// Combine staged and unstaged like lazygit Files pane
	var lines []string

	// Staged files (green checkmark style)
	for i, f := range m.stagedItems {
		idx := i
		line := m.renderFileItem(f, true, idx == m.filesCursor && m.focus == paneFiles)
		lines = append(lines, line)
	}

	// Unstaged files
	for i, f := range m.unstagedItems {
		idx := len(m.stagedItems) + i
		line := m.renderFileItem(f, false, idx == m.filesCursor && m.focus == paneFiles)
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return lipgloss.NewStyle().Foreground(dimColor).Render("(no changed files)")
	}

	return strings.Join(lines, "\n")
}

func (m model) renderFileItem(f git.FileItem, staged bool, selected bool) string {
	// Status indicator like lazygit
	var statusChar string
	var statusColor lipgloss.Color

	if staged {
		statusChar = "A" // or M for modified
		statusColor = colorStaged
		if f.Status == "M" {
			statusChar = "M"
		} else if f.Status == "D" {
			statusChar = "D"
			statusColor = colorDeleted
		} else if f.Status == "R" {
			statusChar = "R"
			statusColor = colorRenamed
		}
	} else {
		switch f.Status {
		case "?":
			statusChar = "??"
			statusColor = colorUntracked
		case "M":
			statusChar = " M"
			statusColor = colorModified
		case "D":
			statusChar = " D"
			statusColor = colorDeleted
		default:
			statusChar = f.Status
			statusColor = dimColor
		}
	}

	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	line := statusStyle.Render(statusChar) + " " + f.Path

	if selected {
		// Blue background for selected line like lazygit
		line = lipgloss.NewStyle().
			Background(selectedBgColor).
			Foreground(selectedFgColor).
			Render(statusChar + " " + f.Path)
	}

	return line
}

func (m model) renderBranchesContent() string {
	if len(m.branches) == 0 {
		// Fall back to just showing current branch
		branch := m.branchName
		if branch == "" {
			branch = "master"
		}
		branchStyle := lipgloss.NewStyle().Foreground(colorBranchHead)
		selected := m.focus == paneBranches && m.branchesCursor == 0
		line := branchStyle.Render("* " + branch)
		if selected {
			line = lipgloss.NewStyle().
				Background(selectedBgColor).
				Foreground(selectedFgColor).
				Render("* " + branch)
		}
		return line
	}

	var lines []string
	for i, b := range m.branches {
		selected := m.focus == paneBranches && i == m.branchesCursor

		prefix := "  "
		color := colorBranchLocal
		if b.IsCurrent {
			prefix = "* "
			color = colorBranchHead
		}
		if b.IsRemote {
			color = colorBranchRemote
		}

		line := prefix + b.Name
		if selected {
			line = lipgloss.NewStyle().
				Background(selectedBgColor).
				Foreground(selectedFgColor).
				Render(line)
		} else {
			line = lipgloss.NewStyle().Foreground(color).Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m model) renderCommitsContent() string {
	if len(m.commitItems) == 0 {
		return lipgloss.NewStyle().Foreground(dimColor).Render("(no commits)")
	}

	var lines []string
	for i, c := range m.commitItems {
		selected := m.focus == paneCommits && i == m.commitsCursor

		// Format like lazygit: hash author date message
		hashStyle := lipgloss.NewStyle().Foreground(colorHash)
		msgStyle := lipgloss.NewStyle()

		line := hashStyle.Render(c.Hash) + " " + msgStyle.Render(c.Message)

		if selected {
			line = lipgloss.NewStyle().
				Background(selectedBgColor).
				Foreground(selectedFgColor).
				Render(c.Hash + " " + c.Message)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m model) renderStashContent() string {
	if len(m.stashItems) == 0 {
		return lipgloss.NewStyle().Foreground(dimColor).Render("(no stash entries)")
	}

	var lines []string
	for i, s := range m.stashItems {
		// Format: stash@{0}: message
		display := s.Ref + ": " + s.Message
		selected := m.focus == paneStash && i == m.stashCursor
		if selected {
			lines = append(lines, lipgloss.NewStyle().
				Background(selectedBgColor).
				Foreground(selectedFgColor).
				Render(display))
		} else {
			lines = append(lines, display)
		}
	}

	return strings.Join(lines, "\n")
}

func (m model) renderInfoBar() string {
	// Left: context-sensitive keybindings (blue)
	// Right: info/version (green)
	optStyle := lipgloss.NewStyle().Foreground(optionsColor)
	infoStyle := lipgloss.NewStyle().Foreground(infoColor)
	dimStyle := lipgloss.NewStyle().Foreground(dimColor)

	// Keybindings based on focus
	var opts string
	switch m.focus {
	case paneFiles:
		opts = "space: stage | a: all | c: commit | A: amend | d: discard"
	case paneBranches:
		opts = "space: checkout | n: new | d: delete | D: force delete"
	case paneCommits:
		opts = "enter: view | r: undo (staged) | R: undo (unstaged)"
	case paneStash:
		opts = "space: apply | p: pop | d: drop"
	case paneMain:
		opts = "j/k: scroll | d/u: page | g/G: top/bottom"
	default:
		opts = "tab: switch | p: pull | P: push | f: fetch | q: quit"
	}

	left := optStyle.Render(opts)

	// Right side: status or version
	var right string
	if m.statusMsg != "" {
		right = infoStyle.Render(m.statusMsg)
	} else if m.lastGitCmd != "" {
		right = dimStyle.Render("$ " + m.lastGitCmd)
	} else {
		right = infoStyle.Render("gitzen")
	}

	// Calculate spacing
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	space := m.width - leftW - rightW
	if space < 1 {
		space = 1
	}

	return left + strings.Repeat(" ", space) + right
}

func (m model) renderErrorModal() string {
	msg := strings.TrimSpace(m.errorMsg)
	if msg == "" {
		msg = "Unknown error"
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).
		Padding(1, 2).
		Width(50)

	return box.Render("Error\n\n" + msg + "\n\n[ESC] close")
}

func (m model) renderCommitModal() string {
	title := "Commit Message"
	if m.amendMode {
		title = "Amend Commit (empty = keep old message)"
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(activeBorderColor).
		Padding(1, 2).
		Width(60)

	return box.Render(title + "\n\n" + m.commitIn.View() + "\n\n[ENTER] confirm  [ESC] cancel")
}

func (m model) renderCreateBranchModal() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(activeBorderColor).
		Padding(1, 2).
		Width(50)

	return box.Render("New Branch\n\n" + m.branchIn.View() + "\n\n[ENTER] create  [ESC] cancel")
}

func (m model) renderConfirmModal() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("3")). // Yellow for warning
		Padding(1, 2).
		Width(50)

	return box.Render(m.confirmTitle + "\n\n[y] yes  [n/ESC] no")
}

func (m model) overlayModalCentered(base, modal string) string {
	// Calculate modal dimensions
	modalLines := strings.Split(modal, "\n")
	modalH := len(modalLines)
	modalW := 0
	for _, line := range modalLines {
		if w := ansi.StringWidth(line); w > modalW {
			modalW = w
		}
	}

	// Base dimensions
	baseLines := strings.Split(base, "\n")

	// Calculate starting position (centered)
	startY := (len(baseLines) - modalH) / 2
	startX := (m.width - modalW) / 2

	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	// Overlay modal onto base using ANSI-aware truncation
	for i, modalLine := range modalLines {
		targetY := startY + i
		if targetY < len(baseLines) {
			baseLine := baseLines[targetY]

			// Get visual widths
			baseWidth := ansi.StringWidth(baseLine)
			modalWidth := ansi.StringWidth(modalLine)

			// Build the new line:
			// [left part of base] + [modal] + [right part of base]
			var newLine string

			// Left part: truncate base to startX width
			if startX > 0 {
				if baseWidth >= startX {
					newLine = ansi.Truncate(baseLine, startX, "")
				} else {
					// Pad with spaces if base is shorter than startX
					newLine = baseLine + strings.Repeat(" ", startX-baseWidth)
				}
			}

			// Add modal content
			newLine += modalLine

			// Right part: cut from base after modal ends
			endX := startX + modalWidth
			if baseWidth > endX {
				// ansi.Cut(s, left, right) returns substring from left to right
				rightPart := ansi.Cut(baseLine, endX, baseWidth)
				newLine += rightPart
			}

			baseLines[targetY] = newLine
		}
	}

	return strings.Join(baseLines, "\n")
}

func truncateString(s string, maxW int) string {
	if lipgloss.Width(s) <= maxW {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > maxW-1 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
