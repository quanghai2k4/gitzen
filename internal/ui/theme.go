package ui

import "github.com/charmbracelet/lipgloss"

// Theme định nghĩa tất cả colors/styles cho UI
type Theme struct {
	// Border
	ActiveBorder   lipgloss.Color
	InactiveBorder lipgloss.Color

	// Selection (item đang được chọn)
	SelectedBg         lipgloss.Color
	SelectedFg         lipgloss.Color
	InactiveSelectedFg lipgloss.Color

	// File status colors
	Staged    lipgloss.Color
	Modified  lipgloss.Color
	Untracked lipgloss.Color
	Deleted   lipgloss.Color
	Renamed   lipgloss.Color
	Conflict  lipgloss.Color

	// Git objects
	Hash   lipgloss.Color
	Author lipgloss.Color
	Date   lipgloss.Color

	// Branch colors
	BranchLocal  lipgloss.Color
	BranchRemote lipgloss.Color
	BranchHead   lipgloss.Color // current branch

	// UI elements
	Dim     lipgloss.Color
	Info    lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Options lipgloss.Color // keybindings in info bar
}

// DefaultTheme - lazygit-style colors
var DefaultTheme = Theme{
	// Border: green khi focus, trắng khi không
	ActiveBorder:   lipgloss.Color("2"),
	InactiveBorder: lipgloss.Color("7"),

	// Selection: blue background
	SelectedBg:         lipgloss.Color("4"),
	SelectedFg:         lipgloss.Color("15"),
	InactiveSelectedFg: lipgloss.Color("15"),

	// File status
	Staged:    lipgloss.Color("2"), // green
	Modified:  lipgloss.Color("3"), // yellow
	Untracked: lipgloss.Color("1"), // red
	Deleted:   lipgloss.Color("1"), // red
	Renamed:   lipgloss.Color("6"), // cyan
	Conflict:  lipgloss.Color("5"), // magenta

	// Git objects
	Hash:   lipgloss.Color("3"), // yellow
	Author: lipgloss.Color("6"), // cyan
	Date:   lipgloss.Color("4"), // blue

	// Branch
	BranchLocal:  lipgloss.Color("6"), // cyan
	BranchRemote: lipgloss.Color("5"), // magenta
	BranchHead:   lipgloss.Color("2"), // green

	// UI
	Dim:     lipgloss.Color("8"),
	Info:    lipgloss.Color("2"),
	Warning: lipgloss.Color("3"),
	Error:   lipgloss.Color("1"),
	Options: lipgloss.Color("4"),
}

// Styles là pre-built styles từ Theme
type Styles struct {
	// Pane borders
	ActiveBorderStyle   lipgloss.Style
	InactiveBorderStyle lipgloss.Style

	// Pane titles
	ActiveTitleStyle   lipgloss.Style
	InactiveTitleStyle lipgloss.Style

	// Selection
	SelectedStyle         lipgloss.Style
	InactiveSelectedStyle lipgloss.Style

	// File status
	StagedStyle    lipgloss.Style
	ModifiedStyle  lipgloss.Style
	UntrackedStyle lipgloss.Style
	DeletedStyle   lipgloss.Style
	RenamedStyle   lipgloss.Style
	ConflictStyle  lipgloss.Style

	// Git objects
	HashStyle   lipgloss.Style
	AuthorStyle lipgloss.Style
	DateStyle   lipgloss.Style

	// Branch
	BranchLocalStyle  lipgloss.Style
	BranchRemoteStyle lipgloss.Style
	BranchHeadStyle   lipgloss.Style

	// UI
	DimStyle     lipgloss.Style
	InfoStyle    lipgloss.Style
	WarningStyle lipgloss.Style
	ErrorStyle   lipgloss.Style
	OptionsStyle lipgloss.Style

	// Modal
	ModalStyle        lipgloss.Style
	ErrorModalStyle   lipgloss.Style
	WarningModalStyle lipgloss.Style
}

// NewStyles tạo Styles từ Theme
func NewStyles(t Theme) Styles {
	return Styles{
		// Border styles
		ActiveBorderStyle:   lipgloss.NewStyle().Foreground(t.ActiveBorder),
		InactiveBorderStyle: lipgloss.NewStyle().Foreground(t.InactiveBorder),

		// Title styles
		ActiveTitleStyle:   lipgloss.NewStyle().Foreground(t.ActiveBorder).Bold(true),
		InactiveTitleStyle: lipgloss.NewStyle().Foreground(t.InactiveBorder),

		// Selection
		SelectedStyle: lipgloss.NewStyle().
			Background(t.SelectedBg).
			Foreground(t.SelectedFg),
		InactiveSelectedStyle: lipgloss.NewStyle().
			Foreground(t.InactiveSelectedFg).
			Bold(true),

		// File status
		StagedStyle:    lipgloss.NewStyle().Foreground(t.Staged),
		ModifiedStyle:  lipgloss.NewStyle().Foreground(t.Modified),
		UntrackedStyle: lipgloss.NewStyle().Foreground(t.Untracked),
		DeletedStyle:   lipgloss.NewStyle().Foreground(t.Deleted),
		RenamedStyle:   lipgloss.NewStyle().Foreground(t.Renamed),
		ConflictStyle:  lipgloss.NewStyle().Foreground(t.Conflict),

		// Git objects
		HashStyle:   lipgloss.NewStyle().Foreground(t.Hash),
		AuthorStyle: lipgloss.NewStyle().Foreground(t.Author),
		DateStyle:   lipgloss.NewStyle().Foreground(t.Date),

		// Branch
		BranchLocalStyle:  lipgloss.NewStyle().Foreground(t.BranchLocal),
		BranchRemoteStyle: lipgloss.NewStyle().Foreground(t.BranchRemote),
		BranchHeadStyle:   lipgloss.NewStyle().Foreground(t.BranchHead),

		// UI
		DimStyle:     lipgloss.NewStyle().Foreground(t.Dim),
		InfoStyle:    lipgloss.NewStyle().Foreground(t.Info),
		WarningStyle: lipgloss.NewStyle().Foreground(t.Warning),
		ErrorStyle:   lipgloss.NewStyle().Foreground(t.Error),
		OptionsStyle: lipgloss.NewStyle().Foreground(t.Options),

		// Modal borders
		ModalStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.ActiveBorder).
			Padding(1, 2),
		ErrorModalStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Error).
			Padding(1, 2),
		WarningModalStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Warning).
			Padding(1, 2),
	}
}

// DefaultStyles = Styles từ DefaultTheme
var DefaultStyles = NewStyles(DefaultTheme)
