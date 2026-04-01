package background

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gitzen/internal/logger"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// FileWatchEvent represents a file system change
type FileWatchEvent struct {
	Type FileEventType
	Path string
	Time time.Time
}

type FileEventType int

const (
	FileCreated FileEventType = iota
	FileModified
	FileDeleted
	FileRenamed
)

// FileWatcher manages file system watching for git repositories
type FileWatcher struct {
	mu        sync.Mutex
	watcher   *fsnotify.Watcher
	repoRoot  string
	enabled   bool
	eventChan chan FileWatchEvent
	done      chan struct{}

	// Debouncing
	debounceTimer *time.Timer
	pendingEvents map[string]FileEventType
	debounceDelay time.Duration
}

// NewFileWatcher creates a new file watcher instance
func NewFileWatcher(repoRoot string) (*FileWatcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:       fsWatcher,
		repoRoot:      repoRoot,
		eventChan:     make(chan FileWatchEvent, 100),
		done:          make(chan struct{}),
		pendingEvents: make(map[string]FileEventType),
		debounceDelay: 200 * time.Millisecond, // Reduced from 300ms for faster response
		enabled:       true, // Default enabled
	}, nil
}

// Start begins watching the repository
func (fw *FileWatcher) Start(ctx context.Context) tea.Cmd {
	if !fw.enabled {
		return nil
	}

	// Add repository root to watcher
	if err := fw.addWatchPaths(); err != nil {
		logger.Get().Warn("file watcher: failed to add paths: %v", err)
		return nil
	}

	// Start the event processing goroutine
	go fw.processEvents(ctx)

	// Return command to listen for our events
	return fw.listenForEvents()
}

// addWatchPaths adds relevant paths to the watcher
func (fw *FileWatcher) addWatchPaths() error {
	// Watch repository root
	if err := fw.watcher.Add(fw.repoRoot); err != nil {
		return err
	}

	// Add critical git files and directories for detecting git operations
	gitPaths := []string{
		filepath.Join(fw.repoRoot, ".git"),
		filepath.Join(fw.repoRoot, ".git", "HEAD"),
		filepath.Join(fw.repoRoot, ".git", "index"),
		filepath.Join(fw.repoRoot, ".git", "refs"),
		filepath.Join(fw.repoRoot, ".git", "refs", "heads"),
		filepath.Join(fw.repoRoot, ".git", "refs", "remotes"),
		filepath.Join(fw.repoRoot, ".git", "ORIG_HEAD"), // Tracks checkout operations
		filepath.Join(fw.repoRoot, ".git", "FETCH_HEAD"), // Tracks fetch operations
	}

	for _, gitPath := range gitPaths {
		if _, err := os.Stat(gitPath); err == nil {
			if err := fw.watcher.Add(gitPath); err != nil {
				logger.Get().Warn("file watcher: failed to watch %s: %v", gitPath, err)
			}
		}
	}

	// Add subdirectories (but be selective about .git subdirs)
	return filepath.Walk(fw.repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errored paths
		}

		if !info.IsDir() {
			return nil
		}

		// Handle .git directory specially - only watch key subdirs
		if strings.Contains(path, "/.git/") {
			// Allow watching key git subdirectories
			if strings.Contains(path, "/refs/") || strings.Contains(path, "/objects/") {
				return fw.watcher.Add(path)
			}
			// Skip other .git subdirectories to avoid noise
			return filepath.SkipDir
		}

		// Skip the .git directory itself - we handle it above
		if strings.HasSuffix(path, "/.git") {
			return nil // Already handled above
		}

		// Skip ignored directories
		if fw.shouldIgnoreDir(path) {
			return filepath.SkipDir
		}

		return fw.watcher.Add(path)
	})
}

// processEvents handles raw fsnotify events and debounces them
func (fw *FileWatcher) processEvents(ctx context.Context) {
	defer close(fw.eventChan)

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleRawEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			logger.Get().Warn("file watcher error: %v", err)
		}
	}
}

// handleRawEvent processes a single fsnotify event with debouncing
func (fw *FileWatcher) handleRawEvent(event fsnotify.Event) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Skip temporary files but allow important git files
	if strings.HasSuffix(event.Name, "~") || strings.Contains(event.Name, ".tmp") {
		return
	}

	// Allow important git files that indicate repository changes
	isImportantGitFile := strings.Contains(event.Name, "/.git/HEAD") ||
		strings.Contains(event.Name, "/.git/index") ||
		strings.Contains(event.Name, "/.git/refs/") ||
		strings.Contains(event.Name, "/.git/ORIG_HEAD") || // Track checkout operations
		strings.Contains(event.Name, "/.git/FETCH_HEAD") || // Track fetch operations
		(strings.Contains(event.Name, "/.git/") && !strings.Contains(event.Name, "/.git/objects/") && !strings.Contains(event.Name, "/.git/logs/"))

	// Skip most .git directory events except important ones
	if strings.Contains(event.Name, "/.git/") && !isImportantGitFile {
		return
	}

	// Convert fsnotify event to our event type
	var eventType FileEventType
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		eventType = FileCreated
	case event.Op&fsnotify.Write == fsnotify.Write:
		eventType = FileModified
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		eventType = FileDeleted
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		eventType = FileRenamed
	default:
		return // Skip chmod and other events
	}

	// Add to pending events for debouncing
	fw.pendingEvents[event.Name] = eventType
	
	// Enhanced debug logging for external git operations
	if strings.Contains(event.Name, "/.git/HEAD") {
		logger.Get().Debug("file watcher: HEAD file changed (likely branch switch) - %v: %s", eventType, event.Name)
	} else if strings.Contains(event.Name, "/.git/ORIG_HEAD") {
		logger.Get().Debug("file watcher: ORIG_HEAD changed (checkout operation) - %v: %s", eventType, event.Name)
	} else if strings.Contains(event.Name, "/.git/index") {
		logger.Get().Debug("file watcher: index changed (staging operation) - %v: %s", eventType, event.Name)
	} else {
		logger.Get().Debug("file watcher: detected %v event for %s", eventType, event.Name)
	}

	// Reset debounce timer
	if fw.debounceTimer != nil {
		fw.debounceTimer.Stop()
	}

	fw.debounceTimer = time.AfterFunc(fw.debounceDelay, func() {
		fw.flushPendingEvents()
	})
}

// flushPendingEvents sends accumulated events
func (fw *FileWatcher) flushPendingEvents() {
	fw.mu.Lock()
	events := make(map[string]FileEventType)
	for path, eventType := range fw.pendingEvents {
		events[path] = eventType
	}
	fw.pendingEvents = make(map[string]FileEventType)
	fw.mu.Unlock()

	// Send consolidated event (we only care that something changed)
	if len(events) > 0 {
		logger.Get().Debug("file watcher: flushing %d events, triggering git status refresh", len(events))
		for path, eventType := range events {
			logger.Get().Debug("  - %v: %s", eventType, path)
		}

		select {
		case fw.eventChan <- FileWatchEvent{
			Type: FileModified, // Simplified: just trigger refresh
			Time: time.Now(),
		}:
			logger.Get().Debug("file watcher: successfully sent refresh event")
		default:
			logger.Get().Warn("file watcher: event channel full, skipping refresh")
		}
	}
}

// listenForEvents returns a tea.Cmd that listens for file events
func (fw *FileWatcher) listenForEvents() tea.Cmd {
	return func() tea.Msg {
		select {
		case event, ok := <-fw.eventChan:
			if !ok {
				return nil
			}
			return FileWatchEventMsg(event)
		}
	}
}

// FileWatchEventMsg wraps FileWatchEvent for Bubble Tea message passing
type FileWatchEventMsg FileWatchEvent

// SetEnabled enables/disables file watching
func (fw *FileWatcher) SetEnabled(enabled bool) {
	fw.mu.Lock()
	fw.enabled = enabled
	fw.mu.Unlock()
}

// Close cleans up the watcher
func (fw *FileWatcher) Close() error {
	close(fw.done)
	return fw.watcher.Close()
}

// StartSimple starts file watching for testing (without Bubble Tea)
func (fw *FileWatcher) StartSimple(ctx context.Context) error {
	if !fw.enabled {
		return fmt.Errorf("file watcher is disabled")
	}

	// Add repository root to watcher
	if err := fw.addWatchPaths(); err != nil {
		return fmt.Errorf("failed to add watch paths: %w", err)
	}

	fmt.Println("Started monitoring file system events...")

	// Start event processing
	go fw.processEvents(ctx)

	// Listen for events and print them
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-fw.eventChan:
			if !ok {
				return nil
			}
			fmt.Printf("File change detected: %+v\n", event)
		}
	}
}

// shouldIgnoreDir checks if a directory should be ignored
func (fw *FileWatcher) shouldIgnoreDir(path string) bool {
	dirname := filepath.Base(path)
	ignoredDirs := []string{
		"node_modules", "vendor", ".next", "dist", "build",
		".cache", ".tmp", "__pycache__", ".pytest_cache",
		".opencode", // GitZen specific
	}

	for _, ignored := range ignoredDirs {
		if dirname == ignored {
			return true
		}
	}
	return false
}
