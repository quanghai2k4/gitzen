package components

import (
	"gitzen/internal/tui"
	"gitzen/internal/ui"
)

// DiffContext represents what's being shown in the diff view
type DiffContext int

const (
	DiffContextNone   DiffContext = iota
	DiffContextFile               // File diff (staged or unstaged)
	DiffContextCommit             // Commit diff
	DiffContextStash              // Stash diff
	DiffContextBranch             // Branch comparison
)

// DiffView hiển thị diff content (scrollable)
type DiffView struct {
	BasePane

	content    string
	diffStyler tui.DiffStyler
	styles     ui.Styles

	// Context info for dynamic title
	context  DiffContext
	title    string // Dynamic title based on context
	subtitle string // File path, commit hash, etc.
}

// NewDiffView tạo DiffView mới
func NewDiffView(styles ui.Styles) *DiffView {
	return &DiffView{
		BasePane:   NewBasePane(ui.PaneMain),
		diffStyler: tui.DefaultDiffStyler(),
		styles:     styles,
		context:    DiffContextNone,
		title:      "Main",
	}
}

// SetContext sets the diff context and updates title
func (p *DiffView) SetContext(ctx DiffContext, subtitle string) {
	p.context = ctx
	p.subtitle = subtitle

	switch ctx {
	case DiffContextFile:
		p.title = "Diff"
	case DiffContextCommit:
		p.title = "Patch"
	case DiffContextStash:
		p.title = "Stash"
	case DiffContextBranch:
		p.title = "Log"
	default:
		p.title = "Main"
	}
}

// SetDiff cập nhật nội dung diff với syntax highlighting
func (p *DiffView) SetDiff(diff string) {
	p.content = diff
	p.SetContent(p.diffStyler.Colorize(diff))
	p.GotoTop()
}

// SetDiffWithContext sets diff content with context info
func (p *DiffView) SetDiffWithContext(diff string, ctx DiffContext, subtitle string) {
	p.SetContext(ctx, subtitle)
	p.SetDiff(diff)
}

// Clear xóa nội dung
func (p *DiffView) Clear() {
	p.content = ""
	p.subtitle = ""
	p.context = DiffContextNone
	p.title = "Main"
	p.SetContent("")
}

// HasContent kiểm tra có nội dung không
func (p *DiffView) HasContent() bool {
	return p.content != ""
}

// Title returns dynamic title
func (p *DiffView) Title() string {
	return p.title
}

// Subtitle returns context subtitle
func (p *DiffView) Subtitle() string {
	return p.subtitle
}

// FullTitle returns title with subtitle (lazygit style)
func (p *DiffView) FullTitle() string {
	if p.subtitle != "" {
		return p.title + " - " + p.subtitle
	}
	return p.title
}

// View returns rendered content
func (p *DiffView) View() string {
	return p.ViewportView()
}

// RenderBox renders pane with border
func (p *DiffView) RenderBox(focused bool, styles ui.Styles) string {
	return p.BasePane.RenderBox(p.FullTitle(), p.View(), focused, styles)
}

// Refresh re-renders content
func (p *DiffView) Refresh() {
	// DiffView không cần refresh vì content đã được colorize khi set
}
