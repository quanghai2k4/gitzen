package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"gitzen/internal/git"
)

type Options struct {
	RepoPath string
	Version  string
	Commit   string
}

func Run(opts Options) int {
	if err := git.LookPath(); err != nil {
		fmt.Fprintln(os.Stderr, "gitzen: git not found in PATH")
		return 3
	}

	repoRoot, err := git.DetectRepoRoot(opts.RepoPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gitzen: Not a git repository. Run inside a repo or pass --repo <path>.")
		return 2
	}

	m := NewModel(repoRoot)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "gitzen:", err)
		return 1
	}
	return 0
}
