package ui

// Icons định nghĩa tất cả các icon Unicode sử dụng trong GitZen
// Tất cả icon được chọn để tương thích tốt với terminal và có tính thẩm mỹ cao
type Icons struct {
	// File Status Icons - Staged files (sẵn sàng commit)
	StagedModified string // ● - solid circle (đã stage, sẵn sàng)
	StagedAdded    string // ✚ - plus sign (thêm mới)
	StagedDeleted  string // ✖ - X mark (xóa file)
	StagedRenamed  string // ⇄ - exchange arrows (đổi tên/di chuyển)

	// File Status Icons - Unstaged files (chưa stage)
	UnstagedModified string // ◐ - half circle (đang làm dở)
	UnstagedDeleted  string // ⊗ - circled X (chuẩn bị xóa)
	Untracked        string // ◯ - empty circle (chưa track)

	// Status Bar Icons - Fetch operations
	FetchInProgress string // ⟳ - clockwise arrow (đang xoay/loading)
	FetchSuccess    string // ✓ - checkmark (thành công)
	FetchError      string // ⚠ - warning triangle (lỗi)
	SyncAvailable   string // ↑↓ - up/down arrows (có thể đồng bộ)

	// Toast Notification Icons
	ToastSuccess string // ✓ - checkmark (thành công)
	ToastError   string // ✗ - X mark (lỗi)
	ToastInfo    string // ℹ - info symbol (thông tin)
	ToastWarning string // ⚠ - warning triangle (cảnh báo)

	// Branch Indicators
	BranchCurrent string // ◈ - diamond (branch hiện tại)
	BranchLocal   string // ⧫ - solid diamond (branch local)
	BranchRemote  string // ◇ - hollow diamond (branch remote)
	AheadCommits  string // ↑ - up arrow (commits ahead)
	BehindCommits string // ↓ - down arrow (commits behind)

	// Navigation & UI Icons
	ExpandedFolder    string // ▼ - down triangle (folder mở)
	CollapsedFolder   string // ▶ - right triangle (folder đóng)
	FileIcon          string // ◦ - small circle (file thông thường)
	SelectedIndicator string // ▸ - right arrow (item được chọn)
}

// DefaultIcons - bộ icon mặc định với Unicode đẹp và tương thích cao
var DefaultIcons = Icons{
	// File Status - Staged (Green family)
	StagedModified: "●", // U+25CF - Black Circle
	StagedAdded:    "✚", // U+271A - Heavy Greek Cross
	StagedDeleted:  "✖", // U+2716 - Heavy Multiplication X
	StagedRenamed:  "⇄", // U+21C4 - Rightwards Arrow Over Leftwards Arrow

	// File Status - Unstaged (Work in progress family)
	UnstagedModified: "◐", // U+25D0 - Circle With Left Half Black
	UnstagedDeleted:  "⊗", // U+2297 - Circled Times
	Untracked:        "◯", // U+25EF - Large Circle

	// Status Bar - Operations
	FetchInProgress: "⟳", // U+27F3 - Clockwise Gapped Circle Arrow
	FetchSuccess:    "✓", // U+2713 - Check Mark
	FetchError:      "⚠", // U+26A0 - Warning Sign
	SyncAvailable:   "↕", // U+2195 - Up Down Arrow

	// Toast Notifications
	ToastSuccess: "✓", // U+2713 - Check Mark
	ToastError:   "✗", // U+2717 - Ballot X
	ToastInfo:    "ℹ", // U+2139 - Information Source
	ToastWarning: "⚠", // U+26A0 - Warning Sign

	// Branch Indicators
	BranchCurrent: "◈", // U+25C8 - White Diamond Containing Black Small Diamond
	BranchLocal:   "⧫", // U+29EB - Black Lozenge
	BranchRemote:  "◇", // U+25C7 - White Diamond
	AheadCommits:  "↑", // U+2191 - Upwards Arrow
	BehindCommits: "↓", // U+2193 - Downwards Arrow

	// Navigation & UI
	ExpandedFolder:    "▼", // U+25BC - Black Down-Pointing Triangle
	CollapsedFolder:   "▶", // U+25B6 - Black Right-Pointing Triangle
	FileIcon:          "◦", // U+25E6 - White Bullet
	SelectedIndicator: "▸", // U+25B8 - Black Right-Pointing Small Triangle
}

// AlternativeIcons - bộ icon thay thế cho các terminal không hỗ trợ đầy đủ Unicode
var AlternativeIcons = Icons{
	// File Status - Staged (sử dụng ASCII mở rộng)
	StagedModified: "●", // Giữ nguyên vì tương thích cao
	StagedAdded:    "+", // Fallback to ASCII plus
	StagedDeleted:  "×", // U+00D7 - Multiplication Sign (tương thích cao)
	StagedRenamed:  "→", // U+2192 - Rightwards Arrow

	// File Status - Unstaged
	UnstagedModified: "○", // U+25CB - White Circle
	UnstagedDeleted:  "×", // U+00D7 - Multiplication Sign
	Untracked:        "?", // ASCII question mark

	// Status Bar
	FetchInProgress: "~", // ASCII tilde for spinning
	FetchSuccess:    "✓", // Giữ nguyên vì tương thích cao
	FetchError:      "!", // ASCII exclamation
	SyncAvailable:   "^", // ASCII caret

	// Toast Notifications
	ToastSuccess: "✓", // Giữ nguyên
	ToastError:   "×", // U+00D7
	ToastInfo:    "i", // ASCII i
	ToastWarning: "!", // ASCII exclamation

	// Branch Indicators
	BranchCurrent: "*", // ASCII asterisk
	BranchLocal:   "•", // U+2022 - Bullet
	BranchRemote:  "°", // U+00B0 - Degree Sign
	AheadCommits:  "+", // ASCII plus
	BehindCommits: "-", // ASCII minus

	// Navigation & UI
	ExpandedFolder:    "v", // ASCII v
	CollapsedFolder:   ">", // ASCII greater than
	FileIcon:          "-", // ASCII minus
	SelectedIndicator: ">", // ASCII greater than
}

// GetFileStatusIcon trả về icon phù hợp cho file status
func (icons Icons) GetFileStatusIcon(status string, staged bool) string {
	if staged {
		switch status {
		case "M":
			return icons.StagedModified
		case "A":
			return icons.StagedAdded
		case "D":
			return icons.StagedDeleted
		case "R":
			return icons.StagedRenamed
		default:
			return icons.StagedAdded // Mặc định cho staged files
		}
	} else {
		switch status {
		case "M":
			return icons.UnstagedModified
		case "D":
			return icons.UnstagedDeleted
		case "?":
			return icons.Untracked
		default:
			return icons.UnstagedModified
		}
	}
}

// GetBranchIcon trả về icon cho branch dựa trên type
func (icons Icons) GetBranchIcon(isCurrent, isRemote bool) string {
	if isCurrent {
		return icons.BranchCurrent
	}
	if isRemote {
		return icons.BranchRemote
	}
	return icons.BranchLocal
}

// GetCommitCountIcon trả về icon cho commit count (ahead/behind)
func (icons Icons) GetCommitCountIcon(isAhead bool) string {
	if isAhead {
		return icons.AheadCommits
	}
	return icons.BehindCommits
}

// GetToastIcon trả về icon cho toast notification dựa trên string level
func (icons Icons) GetToastIcon(level string) string {
	switch level {
	case "success":
		return icons.ToastSuccess
	case "error":
		return icons.ToastError
	case "warning":
		return icons.ToastWarning
	case "info":
		return icons.ToastInfo
	default:
		return icons.ToastInfo
	}
}

// GetFetchStatusIcon trả về icon cho fetch status dựa trên string status
func (icons Icons) GetFetchStatusIcon(status string) string {
	switch status {
	case "in_progress":
		return icons.FetchInProgress
	case "success":
		return icons.FetchSuccess
	case "error":
		return icons.FetchError
	default:
		return ""
	}
}
