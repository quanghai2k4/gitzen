package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"gitzen/internal/ui"
)

// ModalType định nghĩa loại modal
type ModalType int

const (
	ModalNone ModalType = iota
	ModalCommit
	ModalCreateBranch
	ModalConfirm
	ModalError
)

// Modal component cho các dialog
type Modal struct {
	modalType ModalType
	styles    ui.Styles

	// Text input (cho commit, create branch)
	input textinput.Model

	// Confirm dialog
	confirmTitle   string
	confirmAction  func() tea.Cmd
	confirmYesText string

	// Error dialog
	errorMsg string

	// Commit modal
	amendMode bool
}

// NewModal tạo Modal mới
func NewModal(styles ui.Styles) *Modal {
	input := textinput.New()
	input.CharLimit = 200
	input.Prompt = ""

	return &Modal{
		styles: styles,
		input:  input,
	}
}

// IsOpen kiểm tra modal có đang mở không
func (m *Modal) IsOpen() bool {
	return m.modalType != ModalNone
}

// Type returns current modal type
func (m *Modal) Type() ModalType {
	return m.modalType
}

// --- Open Modals ---

// OpenCommit mở commit modal
func (m *Modal) OpenCommit(amend bool) {
	m.modalType = ModalCommit
	m.amendMode = amend
	m.input.Reset()
	if amend {
		m.input.Placeholder = "Leave empty to keep old message"
	} else {
		m.input.Placeholder = "Enter commit message"
	}
	m.input.Focus()
}

// OpenCreateBranch mở create branch modal
func (m *Modal) OpenCreateBranch() {
	m.modalType = ModalCreateBranch
	m.input.Reset()
	m.input.Placeholder = "Enter branch name"
	m.input.Focus()
}

// OpenConfirm mở confirm dialog
func (m *Modal) OpenConfirm(title string, action func() tea.Cmd) {
	m.modalType = ModalConfirm
	m.confirmTitle = title
	m.confirmAction = action
}

// OpenError mở error dialog
func (m *Modal) OpenError(msg string) {
	m.modalType = ModalError
	m.errorMsg = msg
}

// Close đóng modal
func (m *Modal) Close() {
	m.modalType = ModalNone
	m.input.Blur()
}

// --- Input Access ---

// InputValue returns current input value
func (m *Modal) InputValue() string {
	return m.input.Value()
}

// IsAmendMode returns true if commit modal is in amend mode
func (m *Modal) IsAmendMode() bool {
	return m.amendMode
}

// ConfirmAction returns the confirm action function
func (m *Modal) ConfirmAction() func() tea.Cmd {
	return m.confirmAction
}

// --- Update & View ---

// Update xử lý input cho modal
func (m *Modal) Update(msg tea.Msg) tea.Cmd {
	if m.modalType == ModalCommit || m.modalType == ModalCreateBranch {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return cmd
	}
	return nil
}

// View renders modal content
func (m *Modal) View() string {
	switch m.modalType {
	case ModalCommit:
		return m.renderCommitModal()
	case ModalCreateBranch:
		return m.renderCreateBranchModal()
	case ModalConfirm:
		return m.renderConfirmModal()
	case ModalError:
		return m.renderErrorModal()
	default:
		return ""
	}
}

// renderBox renders lazygit-style box with title on border
func renderBox(title, content string, width int, borderColor lipgloss.Color, titleColor lipgloss.Color) string {
	// Lazygit style: title embedded in top border
	// ╭─ Title ──────────────────╮
	// │ content                  │
	// ╰──────────────────────────╯

	innerWidth := width - 2 // subtract left and right border

	// Build top border with title
	titleStr := ""
	if title != "" {
		titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(true)
		titleStr = " " + titleStyle.Render(title) + " "
	}
	titleLen := ansi.StringWidth(titleStr)

	dashesNeeded := innerWidth - titleLen - 1 // -1 for the dash before title
	if dashesNeeded < 0 {
		dashesNeeded = 0
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	topBorder := borderStyle.Render("╭─") + titleStr + borderStyle.Render(strings.Repeat("─", dashesNeeded)+"╮")

	// Process content lines
	contentLines := strings.Split(content, "\n")
	var bodyLines []string
	for _, line := range contentLines {
		lineWidth := ansi.StringWidth(line)
		padding := innerWidth - lineWidth
		if padding < 0 {
			// Truncate if too long
			line = ansi.Truncate(line, innerWidth, "…")
			padding = 0
		}
		bodyLines = append(bodyLines, borderStyle.Render("│")+line+strings.Repeat(" ", padding)+borderStyle.Render("│"))
	}

	// Bottom border
	bottomBorder := borderStyle.Render("╰" + strings.Repeat("─", innerWidth) + "╯")

	return topBorder + "\n" + strings.Join(bodyLines, "\n") + "\n" + bottomBorder
}

func (m *Modal) renderCommitModal() string {
	width := 60
	innerWidth := width - 2

	title := "Commit"
	if m.amendMode {
		title = "Amend Commit"
	}

	// Input line with visual prompt
	inputLine := m.input.View()

	// Pad input to full width
	inputWidth := ansi.StringWidth(inputLine)
	if inputWidth < innerWidth {
		inputLine = inputLine + strings.Repeat(" ", innerWidth-inputWidth)
	}

	// Footer with keybindings
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
		"enter: confirm • esc: cancel",
	)
	footerWidth := ansi.StringWidth(footer)
	if footerWidth < innerWidth {
		footer = footer + strings.Repeat(" ", innerWidth-footerWidth)
	}

	content := inputLine + "\n" + footer

	return renderBox(title, content, width, lipgloss.Color("2"), lipgloss.Color("2"))
}

func (m *Modal) renderCreateBranchModal() string {
	width := 50
	innerWidth := width - 2

	inputLine := m.input.View()
	inputWidth := ansi.StringWidth(inputLine)
	if inputWidth < innerWidth {
		inputLine = inputLine + strings.Repeat(" ", innerWidth-inputWidth)
	}

	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
		"enter: create • esc: cancel",
	)
	footerWidth := ansi.StringWidth(footer)
	if footerWidth < innerWidth {
		footer = footer + strings.Repeat(" ", innerWidth-footerWidth)
	}

	content := inputLine + "\n" + footer

	return renderBox("New Branch", content, width, lipgloss.Color("2"), lipgloss.Color("2"))
}

func (m *Modal) renderConfirmModal() string {
	width := 50
	innerWidth := width - 2

	// Message
	msg := m.confirmTitle
	msgWidth := ansi.StringWidth(msg)
	if msgWidth < innerWidth {
		msg = msg + strings.Repeat(" ", innerWidth-msgWidth)
	}

	// Footer
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
		"y: yes • n/esc: no",
	)
	footerWidth := ansi.StringWidth(footer)
	if footerWidth < innerWidth {
		footer = footer + strings.Repeat(" ", innerWidth-footerWidth)
	}

	content := msg + "\n" + footer

	return renderBox("Confirm", content, width, lipgloss.Color("3"), lipgloss.Color("3"))
}

func (m *Modal) renderErrorModal() string {
	width := 50
	innerWidth := width - 2

	msg := strings.TrimSpace(m.errorMsg)
	if msg == "" {
		msg = "Unknown error"
	}

	// Wrap long messages
	lines := wrapText(msg, innerWidth)
	var paddedLines []string
	for _, line := range lines {
		lineWidth := ansi.StringWidth(line)
		if lineWidth < innerWidth {
			line = line + strings.Repeat(" ", innerWidth-lineWidth)
		}
		paddedLines = append(paddedLines, line)
	}

	// Footer
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
		"esc: close",
	)
	footerWidth := ansi.StringWidth(footer)
	if footerWidth < innerWidth {
		footer = footer + strings.Repeat(" ", innerWidth-footerWidth)
	}

	content := strings.Join(paddedLines, "\n") + "\n" + footer

	return renderBox("Error", content, width, lipgloss.Color("1"), lipgloss.Color("1"))
}

// wrapText wraps text to specified width
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if ansi.StringWidth(currentLine+" "+word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// --- Overlay Helper ---

// OverlayCentered overlay modal lên base content (centered)
func OverlayCentered(base, modal string, screenWidth int) string {
	modalLines := strings.Split(modal, "\n")
	modalH := len(modalLines)
	modalW := 0
	for _, line := range modalLines {
		if w := ansi.StringWidth(line); w > modalW {
			modalW = w
		}
	}

	baseLines := strings.Split(base, "\n")

	startY := (len(baseLines) - modalH) / 2
	startX := (screenWidth - modalW) / 2

	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	for i, modalLine := range modalLines {
		targetY := startY + i
		if targetY < len(baseLines) {
			baseLine := baseLines[targetY]
			baseWidth := ansi.StringWidth(baseLine)
			modalWidth := ansi.StringWidth(modalLine)

			var newLine string

			if startX > 0 {
				if baseWidth >= startX {
					newLine = ansi.Truncate(baseLine, startX, "")
				} else {
					newLine = baseLine + strings.Repeat(" ", startX-baseWidth)
				}
			}

			newLine += modalLine

			endX := startX + modalWidth
			if baseWidth > endX {
				rightPart := ansi.Cut(baseLine, endX, baseWidth)
				newLine += rightPart
			}

			baseLines[targetY] = newLine
		}
	}

	return strings.Join(baseLines, "\n")
}
