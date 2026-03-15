package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLogger_DisabledWhenNoPath kiểm tra khi logger bị vô hiệu hoá thì IsEnabled() == false
// và các phương thức ghi log không gây panic
func TestLogger_DisabledWhenNoPath(t *testing.T) {
	l := &Logger{enabled: false}

	if l.IsEnabled() {
		t.Errorf("expected logger to be disabled")
	}

	// Gọi các phương thức không được gây panic
	l.Debug("debug %s", "test")
	l.Info("info %s", "test")
	l.Warn("warn %s", "test")
	l.Error("error %s", "test")
}

// TestLogger_Writer_Disabled kiểm tra Writer() trả về io.Discard khi bị disabled
func TestLogger_Writer_Disabled(t *testing.T) {
	l := &Logger{enabled: false}
	w := l.Writer()
	if w == nil {
		t.Error("Writer() should not return nil")
	}
	if w != io.Discard {
		t.Error("disabled logger should return io.Discard")
	}
	// Ghi vào Discard không lỗi
	_, err := w.Write([]byte("test"))
	if err != nil {
		t.Errorf("Discard writer should not return error: %v", err)
	}
}

// TestLogger_WriteToFile kiểm tra logger ghi nội dung đúng ra file
func TestLogger_WriteToFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("cannot create log file: %v", err)
	}
	defer f.Close()

	l := newLoggerFromFile(f)

	l.Info("hello %s", "world")
	l.Error("something went wrong: %v", "err detail")

	// Flush & đọc lại file
	_ = f.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("cannot read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "[INFO]") {
		t.Errorf("expected [INFO] in log, got: %s", logContent)
	}
	if !strings.Contains(logContent, "hello world") {
		t.Errorf("expected 'hello world' in log, got: %s", logContent)
	}
	if !strings.Contains(logContent, "[ERROR]") {
		t.Errorf("expected [ERROR] in log, got: %s", logContent)
	}
	if !strings.Contains(logContent, "something went wrong") {
		t.Errorf("expected error message in log, got: %s", logContent)
	}
}

// TestLogger_AllLevels kiểm tra tất cả các cấp độ log đều ghi đúng level label
func TestLogger_AllLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "levels.log")

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("cannot create log file: %v", err)
	}
	defer f.Close()

	l := newLoggerFromFile(f)

	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warn message")
	l.Error("error message")
	_ = f.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("cannot read log file: %v", err)
	}

	logContent := string(content)

	levels := []string{"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]"}
	for _, level := range levels {
		if !strings.Contains(logContent, level) {
			t.Errorf("expected %s in log output, got:\n%s", level, logContent)
		}
	}
}

// TestLogger_IsEnabled kiểm tra IsEnabled() đúng với trạng thái logger
func TestLogger_IsEnabled(t *testing.T) {
	disabled := &Logger{enabled: false}
	if disabled.IsEnabled() {
		t.Error("expected disabled logger to return IsEnabled()=false")
	}

	tmpDir := t.TempDir()
	f, _ := os.CreateTemp(tmpDir, "logger*.log")
	defer f.Close()

	enabled := newLoggerFromFile(f)
	if !enabled.IsEnabled() {
		t.Error("expected enabled logger to return IsEnabled()=true")
	}
}

// TestLogger_Writer_Enabled kiểm tra Writer() trả về file writer khi enabled
func TestLogger_Writer_Enabled(t *testing.T) {
	tmpDir := t.TempDir()
	f, err := os.CreateTemp(tmpDir, "logger*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	l := newLoggerFromFile(f)
	w := l.Writer()

	if w == nil {
		t.Error("expected non-nil Writer() when enabled")
	}
	if w == io.Discard {
		t.Error("enabled logger should not return io.Discard")
	}
}

// TestLogger_TimestampFormat kiểm tra log entry có chứa timestamp hợp lệ
func TestLogger_TimestampFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "ts.log")

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("cannot create log file: %v", err)
	}
	defer f.Close()

	l := newLoggerFromFile(f)
	l.Info("timestamp test")
	_ = f.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("cannot read log file: %v", err)
	}

	// Timestamp format: "2006-01-02 15:04:05.000"
	// Kiểm tra có dạng: [20xx-...
	logContent := string(content)
	if !strings.Contains(logContent, "[20") {
		t.Errorf("expected timestamp in log entry, got: %s", logContent)
	}
}
