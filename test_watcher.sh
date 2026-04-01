#!/bin/bash
# Test script to verify file watcher detects git operations

set -e

echo "🧪 Testing GitZen file watcher improvements"
echo "==========================================="

# Switch back to master để có clean state
echo "📋 Switching to master branch..."
git checkout master >/dev/null 2>&1

# Create test environment
TEST_FILE="test_watcher_detection.go"
TEST_BRANCH="test/watcher-detection"

echo ""
echo "🔍 Testing file watcher detection for various git operations:"

echo "1. Creating new file..."
cat > "$TEST_FILE" << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Testing file watcher detection")
}
EOF
echo "   ✅ Created $TEST_FILE"

echo "2. Testing git add (staging)..."
git add "$TEST_FILE"
echo "   ✅ Staged $TEST_FILE"

echo "3. Testing git commit..."
git commit -m "test: add file for watcher detection testing" >/dev/null 2>&1
echo "   ✅ Committed $TEST_FILE"

echo "4. Testing branch creation..."
git checkout -b "$TEST_BRANCH" >/dev/null 2>&1
echo "   ✅ Created and switched to branch $TEST_BRANCH"

echo "5. Testing file modification..."
echo "" >> "$TEST_FILE"
echo 'fmt.Println("Modified for watcher test")' >> "$TEST_FILE"
echo "   ✅ Modified $TEST_FILE"

echo "6. Testing git add of modified file..."
git add "$TEST_FILE"
echo "   ✅ Staged modified $TEST_FILE"

echo "7. Testing commit of modified file..."
git commit -m "test: modify file for watcher detection" >/dev/null 2>&1
echo "   ✅ Committed modified $TEST_FILE"

echo "8. Testing branch switch..."
git checkout master >/dev/null 2>&1
echo "   ✅ Switched back to master"

echo ""
echo "🧹 Cleaning up test artifacts..."
git branch -D "$TEST_BRANCH" >/dev/null 2>&1
git reset --hard HEAD~1 >/dev/null 2>&1
rm -f "$TEST_FILE"
echo "   ✅ Cleaned up test branch and file"

echo ""
echo "✅ File watcher test scenario completed!"
echo ""
echo "💡 To test with GitZen:"
echo "   1. Run: GITZEN_DEBUG=1 ./bin/gitzen"
echo "   2. Perform git operations in another terminal"  
echo "   3. Watch for real-time updates in GitZen UI"
echo "   4. Check logs for file watcher debug messages"