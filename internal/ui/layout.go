package ui

// Layout constants matching lazygit defaults
const (
	StatusPaneFixedHeight   = 3
	StashPaneMinHeight      = 3
	CmdLogMinHeight         = 3
	CmdLogExpandedHeight    = 10
	AccordionModeEnabled    = true
	ExpandedSidePanelWeight = 2
	NormalSidePanelWeight   = 1
	SidePanelWidthRatio     = 0.3333 // ~1/3 của screen width
	MinSidebarWidth         = 30
	MaxSidebarWidth         = 60
	MinTerminalWidth        = 60
	MinTerminalHeight       = 15
	InfoBarHeight           = 1
)

// PaneID định danh các pane
type PaneID int

const (
	PaneStatus PaneID = iota
	PaneFiles
	PaneBranches
	PaneCommits
	PaneStash
	PaneMain
	PaneCmdLog
)

// String trả về tên pane cho keymap lookup
func (p PaneID) String() string {
	switch p {
	case PaneStatus:
		return "status"
	case PaneFiles:
		return "files"
	case PaneBranches:
		return "branches"
	case PaneCommits:
		return "commits"
	case PaneStash:
		return "stash"
	case PaneMain:
		return "main"
	case PaneCmdLog:
		return "cmdlog"
	default:
		return ""
	}
}

// Title trả về title hiển thị cho pane
func (p PaneID) Title() string {
	switch p {
	case PaneStatus:
		return "Status"
	case PaneFiles:
		return "Files"
	case PaneBranches:
		return "Branches"
	case PaneCommits:
		return "Commits"
	case PaneStash:
		return "Stash"
	case PaneMain:
		return "Main"
	case PaneCmdLog:
		return "Command Log"
	default:
		return ""
	}
}

// SidebarPanes - các pane ở sidebar (trái)
var SidebarPanes = []PaneID{PaneStatus, PaneFiles, PaneBranches, PaneCommits, PaneStash}

// FocusableSidebarPanes - các pane có thể focus (không bao gồm Status)
var FocusableSidebarPanes = []PaneID{PaneFiles, PaneBranches, PaneCommits, PaneStash}

// Layout chứa các kích thước đã tính toán
type Layout struct {
	// Terminal size
	Width  int
	Height int

	// Sidebar (left column)
	SidebarWidth  int
	StatusHeight  int
	FilesHeight   int
	BranchHeight  int
	CommitsHeight int
	StashHeight   int

	// Right column
	MainWidth    int
	MainHeight   int
	CmdLogHeight int

	// Info bar
	InfoBarHeight int
}

// CalculateLayout tính toán layout dựa trên terminal size và focused pane
func CalculateLayout(width, height int, focusedPane PaneID) Layout {
	l := Layout{
		Width:         width,
		Height:        height,
		InfoBarHeight: InfoBarHeight,
	}

	totalH := height - l.InfoBarHeight

	// Sidebar width: ~1/3, clamped between min/max
	l.SidebarWidth = int(float64(width) * SidePanelWidthRatio)
	if l.SidebarWidth < MinSidebarWidth {
		l.SidebarWidth = MinSidebarWidth
	}
	if l.SidebarWidth > MaxSidebarWidth {
		l.SidebarWidth = MaxSidebarWidth
	}
	l.MainWidth = width - l.SidebarWidth

	// Right side: Main + CmdLog
	if focusedPane == PaneCmdLog {
		l.CmdLogHeight = CmdLogExpandedHeight
	} else {
		l.CmdLogHeight = CmdLogMinHeight
	}
	l.MainHeight = totalH - l.CmdLogHeight

	// Sidebar vertical layout
	l.StatusHeight = StatusPaneFixedHeight

	// Stash: fixed unless focused
	if focusedPane == PaneStash {
		l.StashHeight = 0 // will be calculated with weight
	} else {
		l.StashHeight = StashPaneMinHeight
	}

	remainH := totalH - l.StatusHeight - l.StashHeight

	// Flexible panes với accordion mode
	flexPanes := []PaneID{PaneFiles, PaneBranches, PaneCommits}
	if focusedPane == PaneStash {
		flexPanes = append(flexPanes, PaneStash)
	}

	weights := make([]int, len(flexPanes))
	totalWeight := 0
	for i, p := range flexPanes {
		if AccordionModeEnabled && focusedPane == p {
			weights[i] = ExpandedSidePanelWeight
		} else {
			weights[i] = NormalSidePanelWeight
		}
		totalWeight += weights[i]
	}

	if totalWeight == 0 {
		totalWeight = 1
	}
	unitH := remainH / totalWeight

	// Assign heights
	idx := 0
	l.FilesHeight = unitH * weights[idx]
	idx++
	l.BranchHeight = unitH * weights[idx]
	idx++
	l.CommitsHeight = unitH * weights[idx]
	idx++

	if focusedPane == PaneStash && idx < len(weights) {
		l.StashHeight = unitH * weights[idx]
	}

	// Adjust last flexible pane to fill remaining space
	usedH := l.StatusHeight + l.FilesHeight + l.BranchHeight + l.CommitsHeight + l.StashHeight
	if usedH < totalH {
		l.CommitsHeight += totalH - usedH
	} else if usedH > totalH {
		l.CommitsHeight -= usedH - totalH
	}

	return l
}

// ContentWidth trả về width của content bên trong pane (trừ border)
func (l Layout) ContentWidth(pane PaneID) int {
	switch pane {
	case PaneStatus, PaneFiles, PaneBranches, PaneCommits, PaneStash:
		return l.SidebarWidth - 2
	case PaneMain, PaneCmdLog:
		return l.MainWidth - 2
	default:
		return 0
	}
}

// ContentHeight trả về height của content bên trong pane (trừ border)
func (l Layout) ContentHeight(pane PaneID) int {
	h := l.PaneHeight(pane) - 2
	if h < 1 {
		return 1
	}
	return h
}

// PaneHeight trả về total height của pane (bao gồm border)
func (l Layout) PaneHeight(pane PaneID) int {
	switch pane {
	case PaneStatus:
		return l.StatusHeight
	case PaneFiles:
		return l.FilesHeight
	case PaneBranches:
		return l.BranchHeight
	case PaneCommits:
		return l.CommitsHeight
	case PaneStash:
		return l.StashHeight
	case PaneMain:
		return l.MainHeight
	case PaneCmdLog:
		return l.CmdLogHeight
	default:
		return 0
	}
}

// PaneWidth trả về total width của pane (bao gồm border)
func (l Layout) PaneWidth(pane PaneID) int {
	switch pane {
	case PaneStatus, PaneFiles, PaneBranches, PaneCommits, PaneStash:
		return l.SidebarWidth
	case PaneMain, PaneCmdLog:
		return l.MainWidth
	default:
		return 0
	}
}
