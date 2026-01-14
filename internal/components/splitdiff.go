package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"gitzen/internal/tui"
	"gitzen/internal/ui"
)

// SplitPane indicates which pane is focused in split mode
type SplitPane int

const (
	SplitPaneUnstaged SplitPane = iota
	SplitPaneStaged
)

// SplitDiffView shows unstaged and staged diffs side by side (vertically)
type SplitDiffView struct {
	width  int
	height int

	// Two viewports for split mode
	unstagedVP viewport.Model
	stagedVP   viewport.Model

	// Content
	unstagedDiff string
	stagedDiff   string
	filePath     string

	// Which pane is focused
	focusedPane SplitPane

	// Styling
	diffStyler tui.DiffStyler
	styles     ui.Styles
}

// NewSplitDiffView creates a new split diff view
func NewSplitDiffView(styles ui.Styles) *SplitDiffView {
	return &SplitDiffView{
		unstagedVP:  viewport.New(0, 0),
		stagedVP:    viewport.New(0, 0),
		diffStyler:  tui.DefaultDiffStyler(),
		styles:      styles,
		focusedPane: SplitPaneUnstaged,
	}
}

// SetSize updates dimensions
func (s *SplitDiffView) SetSize(width, height int) {
	s.width = width
	s.height = height

	// Split height between two panes (minus borders)
	// Each pane has 2 lines for border (top + bottom)
	innerW := max(1, width-2)
	halfH := height / 2
	innerH := max(1, halfH-2)

	s.unstagedVP.Width = innerW
	s.unstagedVP.Height = innerH
	s.stagedVP.Width = innerW
	s.stagedVP.Height = innerH
}

// SetDiffs sets both unstaged and staged diffs
func (s *SplitDiffView) SetDiffs(unstaged, staged, filePath string) {
	s.unstagedDiff = unstaged
	s.stagedDiff = staged
	s.filePath = filePath

	// Apply colorization
	if strings.TrimSpace(unstaged) == "" {
		s.unstagedVP.SetContent("(no unstaged changes)")
	} else {
		s.unstagedVP.SetContent(s.diffStyler.Colorize(unstaged))
	}

	if strings.TrimSpace(staged) == "" {
		s.stagedVP.SetContent("(no staged changes)")
	} else {
		s.stagedVP.SetContent(s.diffStyler.Colorize(staged))
	}

	s.unstagedVP.GotoTop()
	s.stagedVP.GotoTop()
}

// Clear clears the view
func (s *SplitDiffView) Clear() {
	s.unstagedDiff = ""
	s.stagedDiff = ""
	s.filePath = ""
	s.unstagedVP.SetContent("")
	s.stagedVP.SetContent("")
	s.focusedPane = SplitPaneUnstaged
}

// FocusedPane returns which pane is focused
func (s *SplitDiffView) FocusedPane() SplitPane {
	return s.focusedPane
}

// ToggleFocus switches focus between unstaged and staged
func (s *SplitDiffView) ToggleFocus() {
	if s.focusedPane == SplitPaneUnstaged {
		s.focusedPane = SplitPaneStaged
	} else {
		s.focusedPane = SplitPaneUnstaged
	}
}

// SetFocusPane sets focus to a specific pane
func (s *SplitDiffView) SetFocusPane(pane SplitPane) {
	s.focusedPane = pane
}

// ScrollUp scrolls the focused pane up
func (s *SplitDiffView) ScrollUp(lines int) {
	if s.focusedPane == SplitPaneUnstaged {
		s.unstagedVP.SetYOffset(max(0, s.unstagedVP.YOffset-lines))
	} else {
		s.stagedVP.SetYOffset(max(0, s.stagedVP.YOffset-lines))
	}
}

// ScrollDown scrolls the focused pane down
func (s *SplitDiffView) ScrollDown(lines int) {
	if s.focusedPane == SplitPaneUnstaged {
		s.unstagedVP.SetYOffset(s.unstagedVP.YOffset + lines)
	} else {
		s.stagedVP.SetYOffset(s.stagedVP.YOffset + lines)
	}
}

// PageUp scrolls the focused pane up by a page
func (s *SplitDiffView) PageUp() {
	if s.focusedPane == SplitPaneUnstaged {
		s.unstagedVP.ViewUp()
	} else {
		s.stagedVP.ViewUp()
	}
}

// PageDown scrolls the focused pane down by a page
func (s *SplitDiffView) PageDown() {
	if s.focusedPane == SplitPaneUnstaged {
		s.unstagedVP.ViewDown()
	} else {
		s.stagedVP.ViewDown()
	}
}

// GotoTop scrolls the focused pane to top
func (s *SplitDiffView) GotoTop() {
	if s.focusedPane == SplitPaneUnstaged {
		s.unstagedVP.GotoTop()
	} else {
		s.stagedVP.GotoTop()
	}
}

// GotoBottom scrolls the focused pane to bottom
func (s *SplitDiffView) GotoBottom() {
	if s.focusedPane == SplitPaneUnstaged {
		s.unstagedVP.GotoBottom()
	} else {
		s.stagedVP.GotoBottom()
	}
}

// HasContent returns true if there's any diff content
func (s *SplitDiffView) HasContent() bool {
	return s.unstagedDiff != "" || s.stagedDiff != ""
}

// FilePath returns the current file path
func (s *SplitDiffView) FilePath() string {
	return s.filePath
}

// View renders the split diff view
func (s *SplitDiffView) View() string {
	if s.height <= 0 || s.width <= 0 {
		return ""
	}

	halfH := s.height / 2
	bottomH := s.height - halfH

	// Render both panes
	unstagedBox := s.renderPane("Unstaged Changes", s.unstagedVP.View(), halfH, s.focusedPane == SplitPaneUnstaged)
	stagedBox := s.renderPane("Staged Changes", s.stagedVP.View(), bottomH, s.focusedPane == SplitPaneStaged)

	return lipgloss.JoinVertical(lipgloss.Left, unstagedBox, stagedBox)
}

// renderPane renders a single pane with border
func (s *SplitDiffView) renderPane(title, content string, height int, focused bool) string {
	if height <= 0 {
		return ""
	}

	borderStyle := s.styles.InactiveBorderStyle
	titleStyle := s.styles.InactiveTitleStyle
	if focused {
		borderStyle = s.styles.ActiveBorderStyle
		titleStyle = s.styles.ActiveTitleStyle
	}

	innerW := s.width - 2
	innerH := height - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 0 {
		innerH = 0
	}

	// Process content lines
	lines := strings.Split(content, "\n")
	if len(lines) > innerH {
		lines = lines[:innerH]
	}
	for len(lines) < innerH {
		lines = append(lines, "")
	}
	for i, line := range lines {
		if lipgloss.Width(line) > innerW {
			lines[i] = TruncateString(line, innerW)
		}
	}

	// Border characters (rounded)
	topLeft := "╭"
	topRight := "╮"
	hLine := "─"
	vLine := "│"
	botLeft := "╰"
	botRight := "╯"

	// Top line with title
	titleRendered := titleStyle.Render(" " + title + " ")
	titleLen := lipgloss.Width(titleRendered)
	remainingWidth := innerW - titleLen
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	topLine := borderStyle.Render(topLeft) + titleRendered + borderStyle.Render(strings.Repeat(hLine, remainingWidth)+topRight)

	// Content lines
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

	// Bottom line
	bottomLine := borderStyle.Render(botLeft + strings.Repeat(hLine, innerW) + botRight)

	if len(contentLines) == 0 {
		return topLine + "\n" + bottomLine
	}

	return topLine + "\n" + strings.Join(contentLines, "\n") + "\n" + bottomLine
}
