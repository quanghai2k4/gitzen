package components

import (
	"strings"

	"gitzen/internal/ui"
)

// CmdLogPane hiển thị lịch sử git commands đã thực thi
type CmdLogPane struct {
	BasePane

	entries []string
	maxSize int // giới hạn số entries
	styles  ui.Styles
}

// NewCmdLogPane tạo CmdLogPane mới
func NewCmdLogPane(styles ui.Styles) *CmdLogPane {
	return &CmdLogPane{
		BasePane: NewBasePane(ui.PaneCmdLog),
		maxSize:  100,
		styles:   styles,
	}
}

// AddEntry thêm một command vào log
func (p *CmdLogPane) AddEntry(entry string) {
	p.entries = append(p.entries, entry)
	if len(p.entries) > p.maxSize {
		p.entries = p.entries[1:]
	}
	p.refreshContent()
	p.GotoBottom()
}

// Entries returns all entries
func (p *CmdLogPane) Entries() []string {
	return p.entries
}

// Clear xóa tất cả entries
func (p *CmdLogPane) Clear() {
	p.entries = nil
	p.refreshContent()
}

// View returns rendered content
func (p *CmdLogPane) View() string {
	return p.ViewportView()
}

// RenderBox renders pane with border
func (p *CmdLogPane) RenderBox(focused bool, styles ui.Styles) string {
	return p.BasePane.RenderBox(p.ID().Title(), p.View(), focused, styles)
}

// refreshContent cập nhật nội dung
func (p *CmdLogPane) refreshContent() {
	if len(p.entries) == 0 {
		p.SetContent(p.styles.DimStyle.Render("(no commands executed)"))
		return
	}

	var lines []string
	cmdStyle := p.styles.BranchLocalStyle
	for _, cmd := range p.entries {
		lines = append(lines, cmdStyle.Render(cmd))
	}

	p.SetContent(strings.Join(lines, "\n"))
}

// Refresh re-renders content
func (p *CmdLogPane) Refresh() {
	p.refreshContent()
}
