package components

import (
	"fmt"
	"time"
	"gitzen/internal/ui"
)

// FetchStatus định nghĩa trạng thái của fetch operation
type FetchStatus int

const (
	FetchIdle FetchStatus = iota
	FetchInProgress  
	FetchSuccess
	FetchError
)

// StatusPane hiển thị repo info (non-focusable)
type StatusPane struct {
	BasePane

	repoName        string
	branchName      string
	fetchStatus     FetchStatus
	lastFetchTime   time.Time
	newCommitsCount int
	styles          ui.Styles
}

// NewStatusPane tạo StatusPane mới
func NewStatusPane(styles ui.Styles) *StatusPane {
	return &StatusPane{
		BasePane:     NewBasePane(ui.PaneStatus),
		fetchStatus:  FetchIdle,
		lastFetchTime: time.Now(),
		styles:       styles,
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

// SetFetchStatus cập nhật trạng thái fetch
func (p *StatusPane) SetFetchStatus(status FetchStatus) {
	p.fetchStatus = status
	p.refreshContent()
}

// SetLastFetchTime cập nhật thời gian fetch cuối cùng
func (p *StatusPane) SetLastFetchTime(t time.Time) {
	p.lastFetchTime = t
	p.refreshContent()
}

// GetFetchStatus trả về trạng thái fetch hiện tại
func (p *StatusPane) GetFetchStatus() FetchStatus {
	return p.fetchStatus
}

// SetNewCommitsAvailable cập nhật số lượng commit mới có sẵn
func (p *StatusPane) SetNewCommitsAvailable(count int) {
	p.newCommitsCount = count
	p.refreshContent()
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
	
	// Add fetch status indicator
	switch p.fetchStatus {
	case FetchInProgress:
		fetchIndicator := p.styles.FetchingStyle.Render(" [🔄 Fetching...]")
		content += fetchIndicator
	case FetchSuccess:
		fetchIndicator := p.styles.FetchSuccessStyle.Render(" [✅]")
		content += fetchIndicator
	case FetchError:
		fetchIndicator := p.styles.FetchErrorStyle.Render(" [❌]")
		content += fetchIndicator
	case FetchIdle:
		if !p.lastFetchTime.IsZero() {
			elapsed := time.Since(p.lastFetchTime)
			var timeStr string
			if elapsed < time.Minute {
				timeStr = "now"
			} else if elapsed < time.Hour {
				minutes := int(elapsed.Minutes())
				timeStr = fmt.Sprintf("%dm ago", minutes)
			} else if elapsed < 24*time.Hour {
				hours := int(elapsed.Hours())
				timeStr = fmt.Sprintf("%dh ago", hours)
			} else {
				days := int(elapsed.Hours() / 24)
				timeStr = fmt.Sprintf("%dd ago", days)
			}
			fetchIndicator := p.styles.DimStyle.Render(" [Last: " + timeStr + "]")
			content += fetchIndicator
		}
	}
	
	// Add new commits indicator if available
	if p.newCommitsCount > 0 {
		newCommitsIndicator := p.styles.InfoStyle.Render(fmt.Sprintf(" [%d new]", p.newCommitsCount))
		content += newCommitsIndicator
	}
	
	p.SetContent(content)
}

// Refresh re-renders content
func (p *StatusPane) Refresh() {
	p.refreshContent()
}
