package components

import (
	"fmt"
	"strings"

	"gitzen/internal/git"
	"gitzen/internal/ui"
)

// BranchesPane hiển thị danh sách branches
type BranchesPane struct {
	BasePane

	branches     []git.Branch
	commitCounts git.BranchCommitCounts
	styles       ui.Styles
}

// NewBranchesPane tạo BranchesPane mới
func NewBranchesPane(styles ui.Styles) *BranchesPane {
	return &BranchesPane{
		BasePane: NewBasePane(ui.PaneBranches),
		styles:   styles,
	}
}

// SetData cập nhật danh sách branches
func (p *BranchesPane) SetData(branches []git.Branch) {
	p.branches = branches
	p.SetItemCount(len(branches))
	p.refreshContent()
}

// SetCommitCounts cập nhật commit counts cho các branches
func (p *BranchesPane) SetCommitCounts(counts git.BranchCommitCounts) {
	p.commitCounts = counts
	p.refreshContent()
}

// Branches returns branches list
func (p *BranchesPane) Branches() []git.Branch {
	return p.branches
}

// SelectedBranch trả về branch đang được chọn
func (p *BranchesPane) SelectedBranch() (git.Branch, bool) {
	idx := p.SelectedIndex()
	if idx < len(p.branches) {
		return p.branches[idx], true
	}
	return git.Branch{}, false
}

// CurrentBranch trả về branch hiện tại (IsCurrent = true)
func (p *BranchesPane) CurrentBranch() (git.Branch, bool) {
	for _, b := range p.branches {
		if b.IsCurrent {
			return b, true
		}
	}
	return git.Branch{}, false
}

// View returns rendered content
func (p *BranchesPane) View() string {
	return p.ViewportView()
}

// RenderBox renders pane with border
func (p *BranchesPane) RenderBox(focused bool, styles ui.Styles) string {
	return p.BasePane.RenderBox(p.ID().Title(), p.View(), focused, styles)
}

// refreshContent cập nhật nội dung
func (p *BranchesPane) refreshContent() {
	if len(p.branches) == 0 {
		p.SetContent(p.styles.DimStyle.Render("(no branches)"))
		return
	}

	var lines []string
	for i, b := range p.branches {
		selected := p.IsFocused() && i == p.SelectedIndex()

		prefix := "  "
		branchStyle := p.styles.BranchLocalStyle
		if b.IsCurrent {
			prefix = "* "
			branchStyle = p.styles.BranchHeadStyle
		}
		if b.IsRemote {
			branchStyle = p.styles.BranchRemoteStyle
		}

		line := prefix + b.Name

		// Thêm commit count indicators nếu có
		if p.commitCounts != nil {
			if count, exists := p.commitCounts[b.Name]; exists {
				var indicators []string

				if count.Ahead > 0 {
					aheadIndicator := p.styles.InfoStyle.Render("+" + fmt.Sprintf("%d", count.Ahead))
					indicators = append(indicators, aheadIndicator)
				}

				if count.Behind > 0 {
					behindIndicator := p.styles.WarningStyle.Render("-" + fmt.Sprintf("%d", count.Behind))
					indicators = append(indicators, behindIndicator)
				}

				if len(indicators) > 0 {
					line += " " + strings.Join(indicators, " ")
				}
			}
		}

		if selected {
			line = p.styles.SelectedStyle.Render(line)
		} else {
			line = branchStyle.Render(line)
		}

		lines = append(lines, line)
	}

	p.SetContent(strings.Join(lines, "\n"))
}

// Refresh re-renders content
func (p *BranchesPane) Refresh() {
	p.refreshContent()
}
