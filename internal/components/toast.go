package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"gitzen/internal/ui"
)

// ToastLevel định nghĩa mức độ quan trọng của toast
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ToastNotification đại diện cho một thông báo toast
type ToastNotification struct {
	ID        int
	Message   string
	Level     ToastLevel
	Duration  time.Duration
	StartTime time.Time
	Visible   bool
}

// ToastManager quản lý tất cả toast notifications
type ToastManager struct {
	toasts    []ToastNotification
	styles    ui.Styles
	maxToasts int
	nextID    int
}

// NewToastManager tạo ToastManager mới
func NewToastManager(styles ui.Styles) *ToastManager {
	return &ToastManager{
		toasts:    make([]ToastNotification, 0),
		styles:    styles,
		maxToasts: 3, // Giới hạn 3 toasts để tránh lấp màn hình
		nextID:    1,
	}
}

// AddToastNotification thêm toast notification trực tiếp
func (tm *ToastManager) AddToastNotification(toast ToastNotification) {
	toast.ID = tm.nextID
	tm.nextID++
	
	// Thêm toast mới
	tm.toasts = append(tm.toasts, toast)
	
	// Giới hạn số lượng toasts
	if len(tm.toasts) > tm.maxToasts {
		// Xóa toast cũ nhất
		tm.toasts = tm.toasts[1:]
	}
}

// RemoveToast xóa toast theo ID
func (tm *ToastManager) RemoveToast(id int) {
	for i, toast := range tm.toasts {
		if toast.ID == id {
			tm.toasts = append(tm.toasts[:i], tm.toasts[i+1:]...)
			break
		}
	}
}

// View renders tất cả active toasts
func (tm *ToastManager) View(screenWidth, screenHeight int) string {
	if len(tm.toasts) == 0 {
		return ""
	}

	// Lọc và xóa toasts đã hết hạn
	tm.removeExpired()
	
	if len(tm.toasts) == 0 {
		return ""
	}

	// Render các toasts từ dưới lên trên (mới nhất ở dưới)
	var renderedToasts []string
	for i := len(tm.toasts) - 1; i >= 0; i-- {
		toast := tm.toasts[i]
		if toast.Visible {
			rendered := tm.renderToast(toast)
			renderedToasts = append(renderedToasts, rendered)
		}
	}

	if len(renderedToasts) == 0 {
		return ""
	}

	// Tính toán vị trí bottom-right
	content := strings.Join(renderedToasts, "\n")
	return tm.positionToasts(content, screenWidth, screenHeight)
}

// removeExpired xóa các toasts đã hết hạn
func (tm *ToastManager) removeExpired() {
	now := time.Now()
	filtered := make([]ToastNotification, 0)
	
	for _, toast := range tm.toasts {
		if now.Sub(toast.StartTime) < toast.Duration {
			filtered = append(filtered, toast)
		}
	}
	
	tm.toasts = filtered
}

// renderToast render một toast notification
func (tm *ToastManager) renderToast(toast ToastNotification) string {
	width := 40
	
	// Chọn icon và border color theo level
	var icon string
	var borderColor lipgloss.Color
	
	switch toast.Level {
	case ToastInfo:
		icon = "ℹ"
		borderColor = lipgloss.Color("4") // blue
	case ToastSuccess:
		icon = "✅"
		borderColor = lipgloss.Color("2") // green
	case ToastWarning:
		icon = "⚠"
		borderColor = lipgloss.Color("3") // yellow
	case ToastError:
		icon = "❌"
		borderColor = lipgloss.Color("1") // red
	}
	
	// Format message với icon
	message := fmt.Sprintf("%s %s", icon, toast.Message)
	
	// Wrap text nếu cần
	innerWidth := width - 2
	lines := wrapText(message, innerWidth)
	
	// Pad các dòng
	var paddedLines []string
	for _, line := range lines {
		lineWidth := ansi.StringWidth(line)
		padding := innerWidth - lineWidth
		if padding < 0 {
			padding = 0
		}
		paddedLines = append(paddedLines, line+strings.Repeat(" ", padding))
	}
	
	content := strings.Join(paddedLines, "\n")
	
	// Sử dụng renderBox pattern giống modal
	return renderBox("", content, width, borderColor, borderColor)
}

// positionToasts đặt toasts ở bottom-right
func (tm *ToastManager) positionToasts(toastContent string, screenWidth, screenHeight int) string {
	toastLines := strings.Split(toastContent, "\n")
	toastHeight := len(toastLines)
	toastWidth := 0
	
	// Tìm width lớn nhất
	for _, line := range toastLines {
		if w := ansi.StringWidth(line); w > toastWidth {
			toastWidth = w
		}
	}
	
	// Tính vị trí bottom-right (2 chars từ mép, trên info bar)
	startX := screenWidth - toastWidth - 2
	startY := screenHeight - toastHeight - 1 - 1 // -1 cho info bar, -1 cho spacing
	
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	
	// Tạo empty base để overlay
	baseLines := make([]string, screenHeight)
	for i := range baseLines {
		baseLines[i] = strings.Repeat(" ", screenWidth)
	}
	
	// Overlay toast lên base
	for i, toastLine := range toastLines {
		targetY := startY + i
		if targetY < len(baseLines) && targetY >= 0 {
			baseLine := baseLines[targetY]
			
			// Tính toán vị trí chèn
			if startX < len(baseLine) {
				endX := startX + ansi.StringWidth(toastLine)
				if endX > len(baseLine) {
					endX = len(baseLine)
				}
				
				// Chèn toast line vào base line
				newLine := baseLine[:startX] + toastLine
				if endX < len(baseLine) {
					newLine += baseLine[endX:]
				}
				baseLines[targetY] = newLine
			}
		}
	}
	
	return strings.Join(baseLines, "\n")
}