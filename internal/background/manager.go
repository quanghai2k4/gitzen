package background

import (
	"context"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"gitzen/internal/git"
)

// backgroundTickMsg thông báo khi background timer được kích hoạt
type backgroundTickMsg time.Time

// Manager quản lý các hoạt động background với timer và serialization
type Manager struct {
	mu        sync.Mutex
	running   bool
	gitRunner git.Runner
}

// New tạo một Manager mới với git.Runner được cung cấp
func New(gitRunner git.Runner) *Manager {
	return &Manager{
		gitRunner: gitRunner,
	}
}

// Start khởi tạo background timer với context để có thể hủy bỏ
func (m *Manager) Start(ctx context.Context) tea.Cmd {
	return m.backgroundTickCmd(ctx)
}

// ExecuteIfSafe thực thi một function nếu working directory sạch và không có operation nào khác đang chạy
func (m *Manager) ExecuteIfSafe(fn func() error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil // Skip if already running
	}

	// Check if working directory is clean before proceeding
	clean, err := m.gitRunner.IsWorkingDirectoryClean()
	if err != nil {
		return err
	}
	if !clean {
		return nil // Skip if working directory has changes
	}

	m.running = true
	defer func() { m.running = false }()

	return fn()
}

// backgroundTickCmd tạo tea.Cmd cho background timer với 30 giây interval
func (m *Manager) backgroundTickCmd(ctx context.Context) tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		select {
		case <-ctx.Done():
			return nil // Context cancelled, stop timer
		default:
			return backgroundTickMsg(t)
		}
	})
}