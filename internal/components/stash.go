package components

import (
	"strings"

	"gitzen/internal/git"
	"gitzen/internal/ui"
)

// StashPane hiển thị danh sách stash entries
type StashPane struct {
	BasePane

	entries []git.StashEntry
	styles  ui.Styles
}

// NewStashPane tạo StashPane mới
func NewStashPane(styles ui.Styles) *StashPane {
	return &StashPane{
		BasePane: NewBasePane(ui.PaneStash),
		styles:   styles,
	}
}

// SetData cập nhật danh sách stash entries
func (p *StashPane) SetData(entries []git.StashEntry) {
	p.entries = entries
	p.SetItemCount(len(entries))
	p.refreshContent()
}

// Entries returns stash entries
func (p *StashPane) Entries() []git.StashEntry {
	return p.entries
}

// SelectedEntry trả về entry đang được chọn
func (p *StashPane) SelectedEntry() (git.StashEntry, bool) {
	idx := p.SelectedIndex()
	if idx < len(p.entries) {
		return p.entries[idx], true
	}
	return git.StashEntry{}, false
}

// HasItems kiểm tra có stash entries không
func (p *StashPane) HasItems() bool {
	return len(p.entries) > 0
}

// View returns rendered content
func (p *StashPane) View() string {
	return p.ViewportView()
}

// RenderBox renders pane with border
func (p *StashPane) RenderBox(focused bool, styles ui.Styles) string {
	return p.BasePane.RenderBox(p.ID().Title(), p.View(), focused, styles)
}

// refreshContent cập nhật nội dung
func (p *StashPane) refreshContent() {
	if len(p.entries) == 0 {
		p.SetContent(p.styles.DimStyle.Render("(no stash entries)"))
		return
	}

	var lines []string
	for i, s := range p.entries {
		display := s.Ref + ": " + s.Message
		selected := p.IsFocused() && i == p.SelectedIndex()

		if selected {
			lines = append(lines, p.styles.SelectedStyle.Render(display))
		} else {
			lines = append(lines, display)
		}
	}

	p.SetContent(strings.Join(lines, "\n"))
}

// Refresh re-renders content
func (p *StashPane) Refresh() {
	p.refreshContent()
}
