package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitzen/internal/ui"
)

// Pane interface cho tất cả pane components
type Pane interface {
	// Identity
	ID() ui.PaneID

	// Dimensions
	SetSize(width, height int)

	// Rendering
	View() string
	RenderBox(focused bool, styles ui.Styles) string

	// Navigation (cho list-based panes)
	CursorUp()
	CursorDown()
	CursorTop()
	CursorBottom()
	SelectedIndex() int
	ItemCount() int

	// Focus
	SetFocus(focused bool)
	IsFocused() bool

	// Scrolling (cho viewport-based panes)
	ScrollUp(lines int)
	ScrollDown(lines int)
	PageUp()
	PageDown()
	GotoTop()
	GotoBottom()
}

// BasePane cung cấp implementation chung cho tất cả panes
type BasePane struct {
	id       ui.PaneID
	width    int
	height   int
	focused  bool
	viewport viewport.Model
	cursor   int
	items    int // total items for cursor bounds
}

// NewBasePane tạo một BasePane mới
func NewBasePane(id ui.PaneID) BasePane {
	return BasePane{
		id:       id,
		viewport: viewport.New(0, 0),
	}
}

// ID returns pane ID
func (p *BasePane) ID() ui.PaneID {
	return p.id
}

// SetSize cập nhật kích thước pane
func (p *BasePane) SetSize(width, height int) {
	p.width = width
	p.height = height
	// Content area = total - 2 for border
	p.viewport.Width = max(1, width-2)
	p.viewport.Height = max(1, height-2)
}

// Width returns current width
func (p *BasePane) Width() int {
	return p.width
}

// Height returns current height
func (p *BasePane) Height() int {
	return p.height
}

// ContentWidth returns viewport width
func (p *BasePane) ContentWidth() int {
	return p.viewport.Width
}

// ContentHeight returns viewport height
func (p *BasePane) ContentHeight() int {
	return p.viewport.Height
}

// SetFocus sets focus state
func (p *BasePane) SetFocus(focused bool) {
	p.focused = focused
}

// IsFocused returns focus state
func (p *BasePane) IsFocused() bool {
	return p.focused
}

// SetContent đặt nội dung cho viewport
func (p *BasePane) SetContent(content string) {
	p.viewport.SetContent(content)
}

// ViewportView returns viewport's view
func (p *BasePane) ViewportView() string {
	return p.viewport.View()
}

// --- Cursor Navigation ---

// SetItemCount cập nhật số lượng items (để clamp cursor)
func (p *BasePane) SetItemCount(count int) {
	p.items = count
	if p.cursor >= count {
		p.cursor = max(0, count-1)
	}
}

// CursorUp di chuyển cursor lên
func (p *BasePane) CursorUp() {
	if p.cursor > 0 {
		p.cursor--
		p.ensureCursorVisible()
	}
}

// CursorDown di chuyển cursor xuống
func (p *BasePane) CursorDown() {
	if p.cursor < p.items-1 {
		p.cursor++
		p.ensureCursorVisible()
	}
}

// CursorTop di chuyển đến đầu
func (p *BasePane) CursorTop() {
	p.cursor = 0
	p.viewport.GotoTop()
}

// CursorBottom di chuyển đến cuối
func (p *BasePane) CursorBottom() {
	if p.items > 0 {
		p.cursor = p.items - 1
	}
	p.viewport.GotoBottom()
}

// SelectedIndex returns current cursor position
func (p *BasePane) SelectedIndex() int {
	return p.cursor
}

// SetCursor sets cursor position directly
func (p *BasePane) SetCursor(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx >= p.items && p.items > 0 {
		idx = p.items - 1
	}
	p.cursor = idx
}

// ItemCount returns total items
func (p *BasePane) ItemCount() int {
	return p.items
}

// ensureCursorVisible cuộn viewport để cursor luôn hiển thị
func (p *BasePane) ensureCursorVisible() {
	// Nếu cursor nằm ngoài viewport, cuộn để hiển thị
	viewStart := p.viewport.YOffset
	viewEnd := viewStart + p.viewport.Height

	if p.cursor < viewStart {
		p.viewport.SetYOffset(p.cursor)
	} else if p.cursor >= viewEnd {
		p.viewport.SetYOffset(p.cursor - p.viewport.Height + 1)
	}
}

// --- Viewport Scrolling ---

// ScrollUp cuộn viewport lên
func (p *BasePane) ScrollUp(lines int) {
	p.viewport.SetYOffset(max(0, p.viewport.YOffset-lines))
}

// ScrollDown cuộn viewport xuống
func (p *BasePane) ScrollDown(lines int) {
	p.viewport.SetYOffset(p.viewport.YOffset + lines)
}

// PageUp cuộn lên một trang
func (p *BasePane) PageUp() {
	p.viewport.ViewUp()
}

// PageDown cuộn xuống một trang
func (p *BasePane) PageDown() {
	p.viewport.ViewDown()
}

// GotoTop cuộn đến đầu
func (p *BasePane) GotoTop() {
	p.viewport.GotoTop()
}

// GotoBottom cuộn đến cuối
func (p *BasePane) GotoBottom() {
	p.viewport.GotoBottom()
}

// --- Box Rendering ---

// RenderBox vẽ pane với border và title (lazygit style)
func (p *BasePane) RenderBox(title, content string, focused bool, styles ui.Styles) string {
	if p.height <= 0 {
		return ""
	}

	borderStyle := styles.InactiveBorderStyle
	titleStyle := styles.InactiveTitleStyle
	if focused {
		borderStyle = styles.ActiveBorderStyle
		titleStyle = styles.ActiveTitleStyle
	}

	innerW := p.width - 2
	innerH := p.height - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 0 {
		innerH = 0
	}

	// Truncate/pad content lines
	lines := strings.Split(content, "\n")
	if len(lines) > innerH {
		lines = lines[:innerH]
	}
	for len(lines) < innerH {
		lines = append(lines, "")
	}
	for i, line := range lines {
		if lipgloss.Width(line) > innerW {
			lines[i] = TruncateString(line, innerW)
		}
	}

	// Border characters (rounded)
	topLeft := "╭"
	topRight := "╮"
	hLine := "─"
	vLine := "│"
	botLeft := "╰"
	botRight := "╯"

	// Top line: ╭─ Title ─────╮
	titleRendered := titleStyle.Render(" " + title + " ")
	titleLen := lipgloss.Width(titleRendered)
	remainingWidth := innerW - titleLen
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	topLine := borderStyle.Render(topLeft) + titleRendered + borderStyle.Render(strings.Repeat(hLine, remainingWidth)+topRight)

	// Content lines
	var contentLines []string
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		padding := innerW - lineWidth
		if padding < 0 {
			padding = 0
		}
		paddedLine := line + strings.Repeat(" ", padding)
		contentLines = append(contentLines, borderStyle.Render(vLine)+paddedLine+borderStyle.Render(vLine))
	}

	// Bottom line
	bottomLine := borderStyle.Render(botLeft + strings.Repeat(hLine, innerW) + botRight)

	if len(contentLines) == 0 {
		return topLine + "\n" + bottomLine
	}

	return topLine + "\n" + strings.Join(contentLines, "\n") + "\n" + bottomLine
}

// TruncateString cắt string với ellipsis nếu quá dài
func TruncateString(s string, maxW int) string {
	if lipgloss.Width(s) <= maxW {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > maxW-1 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

// max helper
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// --- Action Result ---

// ActionResult từ HandleKey của component
type ActionResult struct {
	Cmd     tea.Cmd // Command để chạy (có thể nil)
	Handled bool    // Đã xử lý key chưa
}

// NoAction - không xử lý key này
func NoAction() ActionResult {
	return ActionResult{Handled: false}
}

// Handled - đã xử lý, không có command
func Handled() ActionResult {
	return ActionResult{Handled: true}
}

// WithCmd - đã xử lý với command
func WithCmd(cmd tea.Cmd) ActionResult {
	return ActionResult{Cmd: cmd, Handled: true}
}
