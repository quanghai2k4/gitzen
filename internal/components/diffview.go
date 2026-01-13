package components

import (
	"gitzen/internal/tui"
	"gitzen/internal/ui"
)

// DiffView hiển thị diff content (scrollable)
type DiffView struct {
	BasePane

	content    string
	diffStyler tui.DiffStyler
	styles     ui.Styles
}

// NewDiffView tạo DiffView mới
func NewDiffView(styles ui.Styles) *DiffView {
	return &DiffView{
		BasePane:   NewBasePane(ui.PaneMain),
		diffStyler: tui.DefaultDiffStyler(),
		styles:     styles,
	}
}

// SetDiff cập nhật nội dung diff với syntax highlighting
func (p *DiffView) SetDiff(diff string) {
	p.content = diff
	p.SetContent(p.diffStyler.Colorize(diff))
	p.GotoTop()
}

// Clear xóa nội dung
func (p *DiffView) Clear() {
	p.content = ""
	p.SetContent("")
}

// HasContent kiểm tra có nội dung không
func (p *DiffView) HasContent() bool {
	return p.content != ""
}

// View returns rendered content
func (p *DiffView) View() string {
	return p.ViewportView()
}

// RenderBox renders pane with border
func (p *DiffView) RenderBox(focused bool, styles ui.Styles) string {
	return p.BasePane.RenderBox(p.ID().Title(), p.View(), focused, styles)
}

// Refresh re-renders content
func (p *DiffView) Refresh() {
	// DiffView không cần refresh vì content đã được colorize khi set
}
