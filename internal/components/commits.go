package components

import (
	"strings"

	"gitzen/internal/git"
	"gitzen/internal/ui"
)

// CommitsPane hiển thị danh sách commits
type CommitsPane struct {
	BasePane

	commits []git.CommitItem
	styles  ui.Styles
}

// NewCommitsPane tạo CommitsPane mới
func NewCommitsPane(styles ui.Styles) *CommitsPane {
	return &CommitsPane{
		BasePane: NewBasePane(ui.PaneCommits),
		styles:   styles,
	}
}

// SetData cập nhật danh sách commits
func (p *CommitsPane) SetData(commits []git.CommitItem) {
	p.commits = commits
	p.SetItemCount(len(commits))
	p.refreshContent()
}

// Commits returns commits list
func (p *CommitsPane) Commits() []git.CommitItem {
	return p.commits
}

// SelectedCommit trả về commit đang được chọn
func (p *CommitsPane) SelectedCommit() (git.CommitItem, bool) {
	idx := p.SelectedIndex()
	if idx < len(p.commits) {
		return p.commits[idx], true
	}
	return git.CommitItem{}, false
}

// View returns rendered content
func (p *CommitsPane) View() string {
	return p.ViewportView()
}

// RenderBox renders pane with border
func (p *CommitsPane) RenderBox(focused bool, styles ui.Styles) string {
	return p.BasePane.RenderBox(p.ID().Title(), p.View(), focused, styles)
}

// refreshContent cập nhật nội dung
func (p *CommitsPane) refreshContent() {
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

// Refresh re-renders content
func (p *CommitsPane) Refresh() {
	p.refreshContent()
}
