package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type DiffStyler struct {
	Added   lipgloss.Style
	Removed lipgloss.Style
	Hunk    lipgloss.Style
	Header  lipgloss.Style
	Dim     lipgloss.Style
}

func DefaultDiffStyler() DiffStyler {
	return DiffStyler{
		Added:   lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		Removed: lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Hunk:    lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
		Header:  lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Bold(true),
		Dim:     lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
	}
}

func (s DiffStyler) Colorize(diff string) string {
	if diff == "" {
		return ""
	}

	lines := strings.Split(strings.ReplaceAll(diff, "\r\n", "\n"), "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			lines[i] = s.Added.Render(line)
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			lines[i] = s.Removed.Render(line)
		case strings.HasPrefix(line, "@@"):
			lines[i] = s.Hunk.Render(line)
		case strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "index ") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
			lines[i] = s.Header.Render(line)
		case strings.HasPrefix(line, "\\ No newline at end of file"):
			lines[i] = s.Dim.Render(line)
		}
	}
	return strings.Join(lines, "\n")
}
