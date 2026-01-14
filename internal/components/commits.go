package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"gitzen/internal/git"
	"gitzen/internal/ui"
)

// CommitsMode defines the display mode
type CommitsMode int

const (
	ModeCommits CommitsMode = iota
	ModeReflog
)

// CommitsPane hiển thị danh sách commits hoặc reflog
type CommitsPane struct {
	BasePane

	mode    CommitsMode
	commits []git.CommitItem
	reflog  []git.ReflogEntry
	styles  ui.Styles
}

// NewCommitsPane tạo CommitsPane mới
func NewCommitsPane(styles ui.Styles) *CommitsPane {
	return &CommitsPane{
		BasePane: NewBasePane(ui.PaneCommits),
		mode:     ModeCommits,
		styles:   styles,
	}
}

// Mode returns current display mode
func (p *CommitsPane) Mode() CommitsMode {
	return p.mode
}

// SetMode sets display mode
func (p *CommitsPane) SetMode(mode CommitsMode) {
	p.mode = mode
	p.CursorTop()
	p.refreshContent()
}

// ToggleMode switches between commits and reflog
func (p *CommitsPane) ToggleMode() {
	if p.mode == ModeCommits {
		p.mode = ModeReflog
	} else {
		p.mode = ModeCommits
	}
	p.CursorTop()
	p.refreshContent()
}

// SetData cập nhật danh sách commits
func (p *CommitsPane) SetData(commits []git.CommitItem) {
	p.commits = commits
	if p.mode == ModeCommits {
		p.SetItemCount(len(commits))
	}
	p.refreshContent()
}

// SetReflogData cập nhật danh sách reflog
func (p *CommitsPane) SetReflogData(entries []git.ReflogEntry) {
	p.reflog = entries
	if p.mode == ModeReflog {
		p.SetItemCount(len(entries))
	}
	p.refreshContent()
}

// Commits returns commits list
func (p *CommitsPane) Commits() []git.CommitItem {
	return p.commits
}

// SelectedCommit trả về commit đang được chọn
func (p *CommitsPane) SelectedCommit() (git.CommitItem, bool) {
	if p.mode != ModeCommits {
		return git.CommitItem{}, false
	}
	idx := p.SelectedIndex()
	if idx < len(p.commits) {
		return p.commits[idx], true
	}
	return git.CommitItem{}, false
}

// SelectedReflog trả về reflog entry đang được chọn
func (p *CommitsPane) SelectedReflog() (git.ReflogEntry, bool) {
	if p.mode != ModeReflog {
		return git.ReflogEntry{}, false
	}
	idx := p.SelectedIndex()
	if idx < len(p.reflog) {
		return p.reflog[idx], true
	}
	return git.ReflogEntry{}, false
}

// SelectedHash returns hash of selected item (works for both modes)
func (p *CommitsPane) SelectedHash() (string, bool) {
	if p.mode == ModeCommits {
		if c, ok := p.SelectedCommit(); ok {
			return c.Hash, true
		}
	} else {
		if r, ok := p.SelectedReflog(); ok {
			return r.Hash, true
		}
	}
	return "", false
}

// View returns rendered content
func (p *CommitsPane) View() string {
	return p.ViewportView()
}

// Title returns pane title with tab indicator (lazygit style)
func (p *CommitsPane) Title() string {
	// Return simple title - tabs will be styled in RenderBox
	if p.mode == ModeReflog {
		return "Commits | Reflog"
	}
	return "Commits | Reflog"
}

// ActiveTab returns which tab is active for styling
func (p *CommitsPane) ActiveTab() string {
	if p.mode == ModeReflog {
		return "Reflog"
	}
	return "Commits"
}

// RenderBox renders pane with border and tabbed title
func (p *CommitsPane) RenderBox(focused bool, styles ui.Styles) string {
	// Build tabbed title with active highlight
	var title string
	if p.mode == ModeCommits {
		activeStyle := lipgloss.NewStyle().Bold(true).Underline(true)
		title = activeStyle.Render("Commits") + " | Reflog"
	} else {
		activeStyle := lipgloss.NewStyle().Bold(true).Underline(true)
		title = "Commits | " + activeStyle.Render("Reflog")
	}
	return p.BasePane.RenderBox(title, p.View(), focused, styles)
}

// refreshContent cập nhật nội dung
func (p *CommitsPane) refreshContent() {
	if p.mode == ModeCommits {
		p.refreshCommits()
	} else {
		p.refreshReflog()
	}
}

func (p *CommitsPane) refreshCommits() {
	p.SetItemCount(len(p.commits))

	if len(p.commits) == 0 {
		p.SetContent(p.styles.DimStyle.Render("(no commits)"))
		return
	}

	var lines []string
	for i, c := range p.commits {
		selected := p.IsFocused() && i == p.SelectedIndex()

		// Format: hash message
		hashPart := c.Hash
		msgPart := c.Message

		if selected {
			line := p.styles.SelectedStyle.Render(hashPart + " " + msgPart)
			lines = append(lines, line)
		} else {
			line := p.styles.HashStyle.Render(hashPart) + " " + msgPart
			lines = append(lines, line)
		}
	}

	p.SetContent(strings.Join(lines, "\n"))
}

func (p *CommitsPane) refreshReflog() {
	p.SetItemCount(len(p.reflog))

	if len(p.reflog) == 0 {
		p.SetContent(p.styles.DimStyle.Render("(no reflog)"))
		return
	}

	var lines []string
	for i, r := range p.reflog {
		selected := p.IsFocused() && i == p.SelectedIndex()

		// Format: hash action: message
		hashPart := r.Hash
		actionPart := r.Action
		msgPart := r.Message

		// Truncate message if needed
		if len(msgPart) > 50 {
			msgPart = msgPart[:47] + "..."
		}

		var line string
		if selected {
			if actionPart != "" {
				line = p.styles.SelectedStyle.Render(hashPart + " " + actionPart + ": " + msgPart)
			} else {
				line = p.styles.SelectedStyle.Render(hashPart + " " + msgPart)
			}
		} else {
			if actionPart != "" {
				line = p.styles.HashStyle.Render(hashPart) + " " +
					p.styles.BranchLocalStyle.Render(actionPart) + ": " + msgPart
			} else {
				line = p.styles.HashStyle.Render(hashPart) + " " + msgPart
			}
		}
		lines = append(lines, line)
	}

	p.SetContent(strings.Join(lines, "\n"))
}

// Refresh re-renders content
func (p *CommitsPane) Refresh() {
	p.refreshContent()
}
