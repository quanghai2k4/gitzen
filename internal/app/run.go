package app

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"gitzen/internal/git"
	"gitzen/internal/logger"
)

type Options struct {
	RepoPath string
	Version  string
	Commit   string
	LogFile  string // đường dẫn file log, rỗng = vô hiệu hoá logging
}

func Run(opts Options) int {
	// Khởi tạo logger. Nếu LogFile rỗng và biến môi trường GITZEN_LOG được đặt,
	// sử dụng giá trị đó. Nếu không thì log vào ~/.gitzen/gitzen.log khi debug mode.
	logPath := opts.LogFile
	if logPath == "" {
		logPath = os.Getenv("GITZEN_LOG")
	}
	if logPath == "" && os.Getenv("GITZEN_DEBUG") != "" {
		home, err := os.UserHomeDir()
		if err == nil {
			logPath = filepath.Join(home, ".gitzen", "gitzen.log")
		}
	}

	if err := logger.Init(logPath); err != nil {
		// Không block ứng dụng nếu không khởi tạo được logger
		fmt.Fprintf(os.Stderr, "warning: logger init failed: %v\n", err)
	}
	defer logger.Close()

	log := logger.Get()
	log.Info("gitzen %s (%s) starting", opts.Version, opts.Commit)

	if err := git.LookPath(); err != nil {
		log.Error("git not found in PATH")
		fmt.Fprintln(os.Stderr, "gitzen: git not found in PATH")
		return 3
	}

	repoRoot, err := git.DetectRepoRoot(opts.RepoPath)
	if err != nil {
		log.Error("not a git repository: %v", err)
		fmt.Fprintln(os.Stderr, "gitzen: Not a git repository. Run inside a repo or pass --repo <path>.")
		return 2
	}

	log.Info("repo root detected: %s", repoRoot)

	m := NewModel(repoRoot)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Error("tea program error: %v", err)
		fmt.Fprintln(os.Stderr, "gitzen:", err)
		return 1
	}

	log.Info("gitzen exited cleanly")
	return 0
}
