package background

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	tea "github.com/charmbracelet/bubbletea"
	"gitzen/internal/logger"
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
	mu            sync.Mutex
	watcher       *fsnotify.Watcher
	repoRoot      string
	enabled       bool
	eventChan     chan FileWatchEvent
	done          chan struct{}

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
		debounceDelay: 300 * time.Millisecond,
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

	// Add subdirectories (excluding .git)
	return filepath.Walk(fw.repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errored paths
		}

		if !info.IsDir() {
			return nil
		}

		// Skip .git directory and its subdirectories
		if strings.Contains(path, "/.git/") || strings.HasSuffix(path, "/.git") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
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

	// Skip .git directory events and temporary files
	if strings.Contains(event.Name, "/.git/") || strings.HasSuffix(event.Name, "~") || strings.Contains(event.Name, ".tmp") {
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
		select {
		case fw.eventChan <- FileWatchEvent{
			Type: FileModified, // Simplified: just trigger refresh
			Time: time.Now(),
		}:
		default:
			// Channel full, skip this event
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