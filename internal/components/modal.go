package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	input.Prompt = "> "

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
		m.input.Placeholder = "Commit message"
	}
	m.input.Focus()
}

// OpenCreateBranch mở create branch modal
func (m *Modal) OpenCreateBranch() {
	m.modalType = ModalCreateBranch
	m.input.Reset()
	m.input.Placeholder = "Branch name"
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

// View renders modal content (không có overlay)
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

func (m *Modal) renderCommitModal() string {
	title := "Commit Message"
	if m.amendMode {
		title = "Amend Commit (empty = keep old message)"
	}

	return m.styles.ModalStyle.Width(60).Render(
		title + "\n\n" + m.input.View() + "\n\n[ENTER] confirm  [ESC] cancel",
	)
}

func (m *Modal) renderCreateBranchModal() string {
	return m.styles.ModalStyle.Width(50).Render(
		"New Branch\n\n" + m.input.View() + "\n\n[ENTER] create  [ESC] cancel",
	)
}

func (m *Modal) renderConfirmModal() string {
	return m.styles.WarningModalStyle.Width(50).Render(
		m.confirmTitle + "\n\n[y] yes  [n/ESC] no",
	)
}

func (m *Modal) renderErrorModal() string {
	msg := strings.TrimSpace(m.errorMsg)
	if msg == "" {
		msg = "Unknown error"
	}

	return m.styles.ErrorModalStyle.Width(50).Render(
		"Error\n\n" + msg + "\n\n[ESC] close",
	)
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
