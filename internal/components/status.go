package components

import (
	"gitzen/internal/ui"
)

// StatusPane hiển thị repo info (non-focusable)
type StatusPane struct {
	BasePane

	repoName   string
	branchName string
	styles     ui.Styles
}

// NewStatusPane tạo StatusPane mới
func NewStatusPane(styles ui.Styles) *StatusPane {
	return &StatusPane{
		BasePane: NewBasePane(ui.PaneStatus),
		styles:   styles,
	}
}

// SetData cập nhật thông tin repo
func (p *StatusPane) SetData(repoName, branchName string) {
	p.repoName = repoName
	p.branchName = branchName
	p.refreshContent()
}

// RepoName returns repository name
func (p *StatusPane) RepoName() string {
	return p.repoName
}

// BranchName returns current branch name
func (p *StatusPane) BranchName() string {
	return p.branchName
}

// View returns rendered content
func (p *StatusPane) View() string {
	return p.ViewportView()
}

// RenderBox renders pane with border
func (p *StatusPane) RenderBox(focused bool, styles ui.Styles) string {
	// Status pane không bao giờ focused nhưng vẫn dùng interface nhất quán
	return p.BasePane.RenderBox(p.ID().Title(), p.View(), false, styles)
}

// refreshContent cập nhật nội dung
func (p *StatusPane) refreshContent() {
	branch := p.branchName
	if branch == "" {
		branch = "master"
	}

	repoStyle := p.styles.BranchHeadStyle.Bold(true)
	branchStyle := p.styles.BranchLocalStyle

	content := repoStyle.Render(p.repoName) + " → " + branchStyle.Render(branch)
	p.SetContent(content)
}

// Refresh re-renders content
func (p *StatusPane) Refresh() {
	p.refreshContent()
}
