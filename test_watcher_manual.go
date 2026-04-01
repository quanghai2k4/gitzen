package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"gitzen/internal/background"
	"gitzen/internal/git"
)

func main() {
	// Get current working directory (should be git repo)
	repoRoot, err := git.DetectRepoRoot(".")
	if err != nil {
		fmt.Printf("Error: Not in a git repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🧪 Testing file watcher in repository: %s\n", repoRoot)
	fmt.Println("=====================================")

	// Create file watcher
	fw, err := background.NewFileWatcher(repoRoot)
	if err != nil {
		fmt.Printf("Error creating file watcher: %v\n", err)
		os.Exit(1)
	}
	defer fw.Close()

	fmt.Println("📋 File watcher created successfully")

	// Test addWatchPaths directly
	fmt.Println("🔧 Setting up watch paths...")
	// This will test the path setup without starting the event loop

	// Start watching for 8 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	fmt.Println("📋 Starting file watcher for 8 seconds...")
	fmt.Println("💡 Perform git operations to test detection:")
	fmt.Println("   - touch test.txt && git add test.txt && git commit -m 'test'")
	fmt.Println("   - git checkout -b test-branch")
	fmt.Println("   - git checkout master")
	fmt.Println("")

	start := time.Now()
	// Run the file watcher
	if err := fw.StartSimple(ctx); err != nil && err != context.DeadlineExceeded {
		fmt.Printf("File watcher error: %v\n", err)
	}

	fmt.Printf("\n✅ File watcher test completed after %v!\n", time.Since(start))
}
