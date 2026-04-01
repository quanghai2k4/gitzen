#!/bin/bash
# Simple test to verify file watcher detects git operations in real time

echo "🧪 Testing GitZen file watcher with real git operations"
echo "======================================================"

# Enable debug logging
export GITZEN_DEBUG=1
export GITZEN_LOG="/tmp/gitzen_watcher_test.log"

echo "📋 Starting GitZen in background with debug logging enabled..."
echo "   Log file: $GITZEN_LOG"

# Start GitZen in background (will exit after 10 seconds)
timeout 10s ./bin/gitzen &
GITZEN_PID=$!

# Give GitZen time to start up
sleep 2

echo ""
echo "🔄 Performing git operations to test real-time detection..."

# Test 1: Create and stage a file
echo "1. Creating new test file..."
echo 'package main

import "fmt"

func main() {
    fmt.Println("Real-time watcher test")
}' > realtime_test.go
sleep 1

echo "2. Staging the file..."
git add realtime_test.go
sleep 1

echo "3. Committing the file..."
git commit -m "test: real-time watcher test file" >/dev/null 2>&1
sleep 1

echo "4. Creating new branch..."
git checkout -b test/realtime-watcher >/dev/null 2>&1
sleep 1

echo "5. Modifying the file..."
echo '    fmt.Println("Modified for real-time test")' >> realtime_test.go
sleep 1

echo "6. Staging modified file..."
git add realtime_test.go
sleep 1

echo "7. Committing changes..."
git commit -m "test: modify file for real-time test" >/dev/null 2>&1
sleep 1

echo ""
echo "⏰ Waiting for GitZen to process events..."
wait $GITZEN_PID 2>/dev/null || true

echo ""
echo "📋 Checking debug logs for file watcher activity..."
if [[ -f "$GITZEN_LOG" ]]; then
    echo "   Found log file with $(wc -l < "$GITZEN_LOG") lines"
    echo ""
    echo "🔍 File watcher events detected:"
    grep -i "file watcher" "$GITZEN_LOG" || echo "   No file watcher events found in logs"
    echo ""
    echo "🔄 Git refresh events:"
    grep -i "refresh\|status" "$GITZEN_LOG" | head -10 || echo "   No refresh events found in logs"
else
    echo "   ⚠️  No log file found at $GITZEN_LOG"
fi

echo ""
echo "🧹 Cleaning up..."
git checkout master >/dev/null 2>&1
git branch -D test/realtime-watcher >/dev/null 2>&1 
git reset --hard HEAD~1 >/dev/null 2>&1
rm -f realtime_test.go "$GITZEN_LOG"

echo "✅ Real-time file watcher test completed!"