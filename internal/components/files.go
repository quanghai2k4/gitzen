package components

import (
	"strings"

	"gitzen/internal/git"
	"gitzen/internal/ui"
)

// FilesPane hiển thị staged và unstaged files
type FilesPane struct {
	BasePane

	stagedItems   []git.FileItem
	unstagedItems []git.FileItem
	styles        ui.Styles
}

// NewFilesPane tạo FilesPane mới
func NewFilesPane(styles ui.Styles) *FilesPane {
	return &FilesPane{
		BasePane: NewBasePane(ui.PaneFiles),
		styles:   styles,
	}
}

// SetData cập nhật dữ liệu files
func (p *FilesPane) SetData(staged, unstaged []git.FileItem) {
	p.stagedItems = staged
	p.unstagedItems = unstaged
	p.SetItemCount(len(staged) + len(unstaged))
	p.refreshContent()
}

// StagedItems returns staged files
func (p *FilesPane) StagedItems() []git.FileItem {
	return p.stagedItems
}

// UnstagedItems returns unstaged files
func (p *FilesPane) UnstagedItems() []git.FileItem {
	return p.unstagedItems
}

// SelectedItem trả về item đang được chọn
func (p *FilesPane) SelectedItem() (git.FileItem, bool, bool) {
	idx := p.SelectedIndex()
	if idx < len(p.stagedItems) {
		return p.stagedItems[idx], true, true // item, staged, found
	}
	unstagedIdx := idx - len(p.stagedItems)
	if unstagedIdx < len(p.unstagedItems) {
		return p.unstagedItems[unstagedIdx], false, true
	}
	return git.FileItem{}, false, false
}

// IsSelectedStaged kiểm tra item đang chọn có phải staged không
func (p *FilesPane) IsSelectedStaged() bool {
	return p.SelectedIndex() < len(p.stagedItems)
}

// HasItems kiểm tra có files nào không
func (p *FilesPane) HasItems() bool {
	return len(p.stagedItems) > 0 || len(p.unstagedItems) > 0
}

// HasStaged kiểm tra có staged files không
func (p *FilesPane) HasStaged() bool {
	return len(p.stagedItems) > 0
}

// View returns rendered content (for viewport)
func (p *FilesPane) View() string {
	return p.ViewportView()
}

// RenderBox renders pane with border
func (p *FilesPane) RenderBox(focused bool, styles ui.Styles) string {
	return p.BasePane.RenderBox(p.ID().Title(), p.View(), focused, styles)
}

// refreshContent cập nhật nội dung viewport
func (p *FilesPane) refreshContent() {
	var lines []string

	// Staged files
	for i, f := range p.stagedItems {
		selected := p.IsFocused() && i == p.SelectedIndex()
		lines = append(lines, p.renderFileItem(f, true, selected))
	}

	// Unstaged files
	for i, f := range p.unstagedItems {
		idx := len(p.stagedItems) + i
		selected := p.IsFocused() && idx == p.SelectedIndex()
		lines = append(lines, p.renderFileItem(f, false, selected))
	}

	if len(lines) == 0 {
		content := p.styles.DimStyle.Render("(no changed files)")
		p.SetContent(content)
		return
	}

	p.SetContent(strings.Join(lines, "\n"))
}

// renderFileItem renders một file item với beautiful icons
func (p *FilesPane) renderFileItem(f git.FileItem, staged bool, selected bool) string {
	// Lấy icon phù hợp từ icon system
	icon := p.styles.Icons.GetFileStatusIcon(f.Status, staged)

	var statusStyle = p.styles.DimStyle

	if staged {
		switch f.Status {
		case "M":
			statusStyle = p.styles.StagedStyle
		case "D":
			statusStyle = p.styles.DeletedStyle
		case "R":
			statusStyle = p.styles.RenamedStyle
		default:
			statusStyle = p.styles.StagedStyle
		}
	} else {
		switch f.Status {
		case "?":
			statusStyle = p.styles.UntrackedStyle
		case "M":
			statusStyle = p.styles.ModifiedStyle
		case "D":
			statusStyle = p.styles.DeletedStyle
		default:
			statusStyle = p.styles.ModifiedStyle
		}
	}

	line := statusStyle.Render(icon) + " " + f.Path

	if selected {
		line = p.styles.SelectedStyle.Render(icon + " " + f.Path)
	}

	return line
}

// Refresh re-renders content (call after cursor move or focus change)
func (p *FilesPane) Refresh() {
	p.refreshContent()
}
