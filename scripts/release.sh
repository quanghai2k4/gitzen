#!/bin/bash
# GitZen Release Helper Script
# Simplifies manual tag creation and release process

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    printf "${1}%s${NC}\n" "${2}"
}

print_color "$BLUE" "🚀 GitZen Release Helper"

# Get current version
current_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
current_version="${current_tag#v}"

print_color "$YELLOW" "Current version: $current_tag"

# Parse version components
major=$(echo "$current_version" | cut -d. -f1)
minor=$(echo "$current_version" | cut -d. -f2)  
patch=$(echo "$current_version" | cut -d. -f3 | cut -d- -f1)

# Calculate next versions
next_patch="v$((major)).$((minor)).$((patch + 1))"
next_minor="v$((major)).$((minor + 1)).0"
next_major="v$((major + 1)).0.0"

# Check if version type provided as argument
if [ "$1" = "patch" ]; then
    new_version="$next_patch"
    choice="1"
elif [ "$1" = "minor" ]; then
    new_version="$next_minor"
    choice="2"
elif [ "$1" = "major" ]; then
    new_version="$next_major"
    choice="3"
else
    # Interactive mode
    echo ""
    print_color "$BLUE" "Available version bumps:"
    echo "1. Patch ($next_patch) - Bug fixes, small improvements"
    echo "2. Minor ($next_minor) - New features, backward compatible"  
    echo "3. Major ($next_major) - Breaking changes"
    echo "4. Custom - Enter your own version"
    echo ""

    # Get user choice
    read -p "Choose version bump (1-4): " choice

    case "$choice" in
        1)
            new_version="$next_patch"
            ;;
        2)
            new_version="$next_minor"
            ;;
        3) 
            new_version="$next_major"
            ;;
        4)
            read -p "Enter custom version (e.g., v1.2.3-beta): " new_version
            if [[ ! "$new_version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
                print_color "$RED" "Invalid version format. Must start with v and follow semver (e.g., v1.2.3)"
                exit 1
            fi
            ;;
        *)
            print_color "$RED" "Invalid choice"
            exit 1
            ;;
    esac
fi

# Confirm version
echo ""
print_color "$YELLOW" "New version: $new_version"
read -p "Proceed with release? (y/N): " confirm

if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
    print_color "$RED" "Release cancelled"
    exit 0
fi

# Check if tag already exists
if git rev-parse "$new_version" >/dev/null 2>&1; then
    print_color "$RED" "Tag $new_version already exists!"
    exit 1
fi

# Check working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    print_color "$RED" "Working directory is not clean. Commit or stash changes first."
    git status --short
    exit 1
fi

# Ensure we're on master/main
current_branch=$(git branch --show-current)
if [[ "$current_branch" != "master" && "$current_branch" != "main" ]]; then
    print_color "$YELLOW" "Warning: Not on master/main branch (currently on: $current_branch)"
    read -p "Continue anyway? (y/N): " branch_confirm
    if [[ "$branch_confirm" != "y" && "$branch_confirm" != "Y" ]]; then
        exit 0
    fi
fi

# Generate changelog
print_color "$BLUE" "Generating changelog..."
if [ "$current_tag" = "v0.0.0" ]; then
    # First release
    changelog=$(git log --oneline --decorate | head -20)
else
    # Get commits since last tag
    changelog=$(git log "${current_tag}..HEAD" --oneline --decorate)
fi

if [ -z "$changelog" ]; then
    print_color "$YELLOW" "No changes since last release"
    changelog="No changes"
fi

echo ""
print_color "$BLUE" "Changelog:"
echo "$changelog"
echo ""

# Create and push tag
print_color "$BLUE" "Creating release tag $new_version..."

git tag -a "$new_version" -m "Release $new_version

Changes since $current_tag:
$changelog"

print_color "$GREEN" "✅ Tag created successfully"

# Push tag
print_color "$BLUE" "Pushing tag to remote..."
git push origin "$new_version"

print_color "$GREEN" "✅ Tag pushed successfully"

# Wait a moment for GitHub to process
sleep 2

# Open release page
release_url="https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]//' | sed 's/.git$//')/releases/tag/$new_version"
print_color "$GREEN" "🎉 Release created!"
print_color "$BLUE" "Release URL: $release_url"

# Try to open in browser (if available)
if command -v open >/dev/null 2>&1; then
    open "$release_url"
elif command -v xdg-open >/dev/null 2>&1; then
    xdg-open "$release_url"
fi

print_color "$GREEN" "Release workflow will automatically build and publish binaries."
print_color "$BLUE" "Check GitHub Actions for progress: https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]//' | sed 's/.git$//')/actions"