package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// clearStatusMsg được gửi để xóa temporary status message
type clearStatusMsg struct{}

// showNotificationMsg để hiển thị temporary notification
type showNotificationMsg struct {
	message  string
	duration time.Duration
}

// newCommitDetectedMsg được gửi khi có new commits được detect
type newCommitDetectedMsg struct {
	branch     string
	newHash    string
	commitCount int
}

// commitCheckCompleteMsg được gửi khi commit check hoàn thành (thành công hoặc thất bại)
type commitCheckCompleteMsg struct {
	branch string
	hash   string
	err    error
}

// handleStartupFetchResult xử lý kết quả từ startup fetch với enhanced visual feedback
func (m model) handleStartupFetchResult(msg startupFetchResultMsg) (model, tea.Cmd) {
	// Update fetch status dựa trên success/failure
	if msg.Success {
		m.fetchStatus = FetchSuccess
		m.lastFetchTime = time.Now()
		
		if msg.Skipped {
			m.cmdLogPane.AddEntry("startup fetch: " + msg.Message)
		} else {
			m.cmdLogPane.AddEntry("startup fetch: " + msg.Message)
			
			// Check for new commits after successful fetch
			newCommitCmd := m.checkForNewCommits()
			
			// Show success notification for 3 seconds
			m.statusMsg = "startup fetch completed"
			return m, tea.Batch(
				newCommitCmd,
				showTemporaryNotification(3 * time.Second),
			)
		}
	} else {
		m.fetchStatus = FetchFailed
		m.cmdLogPane.AddEntry("startup fetch failed: " + msg.Message)
		// Show failure notification for 3 seconds
		m.statusMsg = "startup fetch failed"
		return m, showTemporaryNotification(3 * time.Second)
	}
	
	return m, nil
}

// handleAutoFetchResult xử lý kết quả từ background auto fetch với enhanced visual feedback
func (m model) handleAutoFetchResult(msg autoFetchResultMsg) (model, tea.Cmd) {
	// Update fetch status dựa trên success/failure
	if msg.Success {
		m.fetchStatus = FetchSuccess
		m.lastFetchTime = time.Now()
		
		if !msg.Skipped {
			m.cmdLogPane.AddEntry("auto fetch: " + msg.Message)
			
			// Check for new commits after successful fetch
			newCommitCmd := m.checkForNewCommits()
			
			// Show success notification for 3 seconds
			m.statusMsg = "auto fetch completed"
			return m, tea.Batch(
				newCommitCmd,
				showTemporaryNotification(3 * time.Second),
			)
		}
	} else {
		m.fetchStatus = FetchFailed
		// Auto fetch failures được log nhưng không show UI notification để tránh noise
		// Chỉ show trong cmdLog
		return m, nil
	}
	
	return m, nil
}

// handleFetchStart xử lý khi fetch operation bắt đầu
func (m model) handleFetchStart() (model, tea.Cmd) {
	m.fetchStatus = FetchInProgress
	// Start spinner animation
	return m, m.fetchSpinner.Tick
}

// showTemporaryNotification tạo tea.Cmd để auto-clear status message sau duration
func showTemporaryNotification(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// handleClearStatus xử lý việc xóa temporary status message
func (m model) handleClearStatus() (model, tea.Cmd) {
	m.statusMsg = ""
	return m, nil
}

// checkForNewCommits tạo tea.Cmd để check for new commits trên current branch
func (m model) checkForNewCommits() tea.Cmd {
	return func() tea.Msg {
		// Get current branch using existing method
		currentBranch, err := m.git.CurrentBranch()
		if err != nil {
			return commitCheckCompleteMsg{err: err}
		}
		currentBranch = strings.TrimSpace(currentBranch)
		
		// Get current hash cho branch bằng cách sử dụng rev-parse
		currentHash, err := m.getCurrentBranchHash()
		if err != nil {
			return commitCheckCompleteMsg{branch: currentBranch, err: err}
		}
		
		// Check if we have a known hash for this branch
		if lastHash, exists := m.lastKnownHashes[currentBranch]; exists {
			if lastHash != currentHash {
				// We have new commits! 
				return newCommitDetectedMsg{
					branch:      currentBranch,
					newHash:     currentHash,
					commitCount: 1, // Simple approach: assume 1 new commit
				}
			}
		}
		
		// No new commits, just update hash
		return commitCheckCompleteMsg{
			branch: currentBranch,
			hash:   currentHash,
		}
	}
}

// getCurrentBranchHash lấy hash của current commit sử dụng existing git runner
func (m model) getCurrentBranchHash() (string, error) {
	// Use LogOneline để get current commit hash
	logOutput, err := m.git.LogOneline()
	if err != nil {
		return "", err
	}
	
	// Parse first line để get hash
	lines := strings.Split(logOutput, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no commits found")
	}
	
	firstLine := strings.TrimSpace(lines[0])
	if firstLine == "" {
		return "", fmt.Errorf("empty log output")
	}
	
	// Extract hash (first part before space)
	parts := strings.Split(firstLine, " ")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid log format")
	}
	
	return parts[0], nil
}

// handleNewCommitDetected xử lý khi có new commits được detected
func (m model) handleNewCommitDetected(msg newCommitDetectedMsg) (model, tea.Cmd) {
	// Update lastKnownHashes với new hash
	m.lastKnownHashes[msg.branch] = msg.newHash
	
	// Show notification về new commits
	var commitText string
	if msg.commitCount == 1 {
		commitText = "1 new commit"
	} else {
		commitText = fmt.Sprintf("%d new commits", msg.commitCount)
	}
	
	notification := fmt.Sprintf("%s available on %s", commitText, msg.branch)
	m.statusMsg = notification
	
	return m, showTemporaryNotification(5 * time.Second) // Show longer for important info
}

// handleCommitCheckComplete xử lý khi commit check hoàn thành (no new commits)
func (m model) handleCommitCheckComplete(msg commitCheckCompleteMsg) (model, tea.Cmd) {
	if msg.err != nil {
		// Silently handle error - commit checking không nên disrupt UI
		return m, nil
	}
	
	// Update lastKnownHashes with current hash
	if msg.branch != "" && msg.hash != "" {
		m.lastKnownHashes[msg.branch] = msg.hash
	}
	
	return m, nil
}