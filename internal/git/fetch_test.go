package git

import (
	"strings"
	"testing"
)

func TestFetchBranches(t *testing.T) {
	// Tạo runner với repository root giả định
	runner := New("/tmp/test-repo")

	// Test 1: FetchBranches với empty branch list không làm gì
	err := runner.FetchBranches("origin", []string{})
	if err != nil {
		t.Errorf("FetchBranches với empty list should not error, got: %v", err)
	}

	// Test 2: FetchBranches với branch list hợp lệ
	// Note: Test này sẽ fail trong môi trường không có git repo thực
	// nhưng nó kiểm tra interface và error handling
	err = runner.FetchBranches("origin", []string{"main", "feature"})
	if err == nil {
		t.Error("FetchBranches should error without real git repo")
	}
	if !strings.Contains(err.Error(), "git fetch") {
		t.Errorf("Expected error to mention 'git fetch', got: %v", err)
	}
}

func TestGetDefaultBranch(t *testing.T) {
	runner := New("/tmp/test-repo")

	// Test: GetDefaultBranch trả về error khi không có repo
	branch, err := runner.GetDefaultBranch("origin")
	if err == nil {
		t.Error("GetDefaultBranch should error without real git repo")
	}
	// Fallback behavior: trả về "main" khi có lỗi
	if branch != "main" {
		t.Errorf("Expected fallback to 'main', got: %q", branch)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	runner := New("/tmp/test-repo")

	// Test: GetCurrentBranch trả về error khi không có repo
	branch, err := runner.GetCurrentBranch()
	if err == nil {
		t.Error("GetCurrentBranch should error without real git repo")
	}
	// Fallback behavior: trả về "HEAD" khi có lỗi
	if branch != "HEAD" {
		t.Errorf("Expected fallback to 'HEAD', got: %q", branch)
	}
}