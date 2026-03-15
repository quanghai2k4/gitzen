// Package logger cung cấp cơ chế ghi log ra file để hỗ trợ debug
// mà không làm ảnh hưởng đến giao diện TUI.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	once     sync.Once
	instance *Logger
)

// Logger là cấu trúc quản lý việc ghi log ra file
type Logger struct {
	mu      sync.Mutex
	file    *os.File
	logger  *log.Logger
	enabled bool
}

// Init khởi tạo logger singleton và ghi log vào file tại logPath.
// Nếu logPath rỗng, logger sẽ bị vô hiệu hoá (no-op).
func Init(logPath string) error {
	var initErr error
	once.Do(func() {
		if logPath == "" {
			instance = &Logger{enabled: false}
			return
		}

		// Đảm bảo thư mục tồn tại
		if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
			initErr = fmt.Errorf("logger: cannot create log directory: %w", err)
			instance = &Logger{enabled: false}
			return
		}

		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			initErr = fmt.Errorf("logger: cannot open log file %s: %w", logPath, err)
			instance = &Logger{enabled: false}
			return
		}

		instance = &Logger{
			file:    f,
			logger:  log.New(f, "", 0),
			enabled: true,
		}
	})
	return initErr
}

// Get trả về logger singleton. Nếu Init chưa được gọi, trả về no-op logger.
func Get() *Logger {
	if instance == nil {
		return &Logger{enabled: false}
	}
	return instance
}

// Close đóng file log. Nên gọi khi ứng dụng thoát.
func Close() {
	if instance != nil && instance.file != nil {
		_ = instance.file.Close()
	}
}

// newLoggerFromFile tạo Logger từ os.File đã mở sẵn (dùng trong testing)
func newLoggerFromFile(f *os.File) *Logger {
	return &Logger{
		file:    f,
		logger:  log.New(f, "", 0),
		enabled: true,
	}
}

// log ghi một message với level và timestamp
func (l *Logger) log(level, format string, args ...any) {
	if !l.enabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	ts := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] [%s] %s", ts, level, msg)
}

// Debug ghi log với level DEBUG
func (l *Logger) Debug(format string, args ...any) {
	l.log("DEBUG", format, args...)
}

// Info ghi log với level INFO
func (l *Logger) Info(format string, args ...any) {
	l.log("INFO", format, args...)
}

// Warn ghi log với level WARN
func (l *Logger) Warn(format string, args ...any) {
	l.log("WARN", format, args...)
}

// Error ghi log với level ERROR
func (l *Logger) Error(format string, args ...any) {
	l.log("ERROR", format, args...)
}

// Writer trả về io.Writer của logger để tích hợp với các thư viện khác.
// Nếu logger bị vô hiệu hoá, trả về io.Discard.
func (l *Logger) Writer() io.Writer {
	if !l.enabled || l.file == nil {
		return io.Discard
	}
	return l.file
}

// IsEnabled kiểm tra xem logger có đang hoạt động không
func (l *Logger) IsEnabled() bool {
	return l.enabled
}
