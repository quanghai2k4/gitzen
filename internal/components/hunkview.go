package components

import (
	"strings"

	"gitzen/internal/git"
	"gitzen/internal/tui"
	"gitzen/internal/ui"
)

type HunkView struct {
	BasePane

	hunks       []git.Hunk
	diffStyler  tui.DiffStyler
	styles      ui.Styles
	currentPath string
	isStaged    bool
}

func NewHunkView(styles ui.Styles) *HunkView {
	return &HunkView{
		BasePane:   NewBasePane(ui.PaneMain),
		diffStyler: tui.DefaultDiffStyler(),
		styles:     styles,
	}
}

func (p *HunkView) SetHunks(hunks []git.Hunk, path string, staged bool) {
	p.hunks = hunks
	p.currentPath = path
	p.isStaged = staged
	p.SetItemCount(len(hunks))
	p.refreshContent()
}

func (p *HunkView) HunkCount() int {
	return len(p.hunks)
}

func (p *HunkView) SelectedHunk() (git.Hunk, bool) {
	idx := p.SelectedIndex()
	if idx >= 0 && idx < len(p.hunks) {
		return p.hunks[idx], true
	}
	return git.Hunk{}, false
}

func (p *HunkView) CurrentPath() string {
	return p.currentPath
}

func (p *HunkView) IsStaged() bool {
	return p.isStaged
}

func (p *HunkView) Clear() {
	p.hunks = nil
	p.currentPath = ""
	p.isStaged = false
	p.SetItemCount(0)
	p.SetContent("")
}

func (p *HunkView) HasHunks() bool {
	return len(p.hunks) > 0
}

func (p *HunkView) View() string {
	return p.ViewportView()
}

func (p *HunkView) RenderBox(focused bool, styles ui.Styles) string {
	title := "Hunks"
	if p.currentPath != "" {
		title += " - " + p.currentPath
	}
	return p.BasePane.RenderBox(title, p.View(), focused, styles)
}

func (p *HunkView) refreshContent() {
	if len(p.hunks) == 0 {
		p.SetContent(p.styles.DimStyle.Render("(no hunks)"))
		return
	}

	var lines []string
	for i, h := range p.hunks {
		selected := p.IsFocused() && i == p.SelectedIndex()

		hunkLine := p.formatHunkHeader(h, selected)
		lines = append(lines, hunkLine)

		diffLines := strings.Split(p.diffStyler.Colorize(h.Content), "\n")
		for _, line := range diffLines {
			if selected {
				lines = append(lines, p.styles.SelectedStyle.Render(line))
			} else {
				lines = append(lines, line)
			}
		}
	}

	p.SetContent(strings.Join(lines, "\n"))
}

func (p *HunkView) formatHunkHeader(h git.Hunk, selected bool) string {
	var prefix string
	if p.isStaged {
		prefix = "s"
	} else {
		prefix = "y"
	}

	header := prefix + " " + h.Header

	if selected {
		header = p.styles.SelectedStyle.Render("> " + h.Header)
	} else {
		header = p.styles.DimStyle.Render(header)
	}

	return header
}

func (p *HunkView) Refresh() {
	p.refreshContent()
}
