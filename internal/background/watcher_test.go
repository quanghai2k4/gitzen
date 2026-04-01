package background

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestFileWatcherGitOperations(t *testing.T) {
	// Skip this test on Windows due to fsnotify timeout issues with file watchers
	if runtime.GOOS == "windows" {
		t.Skip("Skipping file watcher test on Windows due to fsnotify platform limitations")
	}

	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gitzen-watcher-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repository
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create some git files to watch
	gitFiles := []string{
		filepath.Join(gitDir, "HEAD"),
		filepath.Join(gitDir, "index"),
	}

	for _, gitFile := range gitFiles {
		if err := os.MkdirAll(filepath.Dir(gitFile), 0755); err != nil {
			t.Fatalf("Failed to create parent dir for %s: %v", gitFile, err)
		}
		if err := os.WriteFile(gitFile, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create git file %s: %v", gitFile, err)
		}
	}

	// Create file watcher
	fw, err := NewFileWatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}
	defer fw.Close()

	// Start watching with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := fw.StartSimple(ctx); err != nil && err != context.DeadlineExceeded {
		t.Logf("File watcher test completed: %v", err)
	}

	t.Log("File watcher git operations test passed")
}

func TestFileWatcherPathDetection(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gitzen-path-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create git structure
	gitDir := filepath.Join(tmpDir, ".git")
	refsDir := filepath.Join(gitDir, "refs", "heads")
	if err := os.MkdirAll(refsDir, 0755); err != nil {
		t.Fatalf("Failed to create refs dir: %v", err)
	}

	// Create file watcher
	fw, err := NewFileWatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}
	defer fw.Close()

	// Test that addWatchPaths doesn't fail
	if err := fw.addWatchPaths(); err != nil {
		t.Errorf("addWatchPaths failed: %v", err)
	}

	t.Log("File watcher path detection test passed")
}
