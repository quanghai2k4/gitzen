# GitZen Icon System

GitZen now uses beautiful Unicode icons throughout the interface to provide a more visually appealing and intuitive user experience.

## Icon Categories

### File Status Icons

**Staged Files (Ready for commit):**
- `●` Modified (solid circle - ready to commit)
- `✚` Added (plus sign - new file added)
- `✖` Deleted (X mark - file removed)
- `⇄` Renamed (exchange arrows - file moved/renamed)

**Unstaged Files (Work in progress):**
- `◐` Modified (half circle - work in progress)
- `⊗` Deleted (circled X - pending deletion)
- `◯` Untracked (empty circle - not tracked by git)

### Status Bar Icons

**Fetch Operations:**
- `⟳` Fetching (clockwise arrow - operation in progress)
- `✓` Success (checkmark - operation completed)
- `⚠` Error (warning triangle - operation failed)

### Toast Notifications

- `✓` Success (checkmark)
- `✗` Error (X mark)
- `ℹ` Info (information symbol)
- `⚠` Warning (warning triangle)

### Branch Indicators

**Branch Types:**
- `◈` Current branch (diamond - active branch)
- `⧫` Local branch (solid diamond)
- `◇` Remote branch (hollow diamond)

**Commit Counts:**
- `↑` Ahead (up arrow + count)
- `↓` Behind (down arrow + count)

## Technical Details

### Icon Selection Criteria

All icons were chosen based on:
1. **Terminal Compatibility** - Work in most terminal emulators
2. **Visual Clarity** - Easy to distinguish at a glance
3. **Semantic Meaning** - Icons relate to their function
4. **Unicode Standard** - Proper Unicode code points for consistency

### Fallback System

The system includes `AlternativeIcons` for terminals with limited Unicode support:
- Uses more ASCII-compatible characters
- Maintains functionality while ensuring compatibility
- Automatically falls back when needed

### Usage in Components

Icons are integrated through the `ui.Styles` struct:
```go
// Example usage in components
icon := p.styles.Icons.GetFileStatusIcon(fileStatus, isStaged)
```

### Customization

Icons can be customized by:
1. Modifying `DefaultIcons` in `internal/ui/icons.go`
2. Creating custom icon sets
3. Runtime icon theme switching (future feature)

## Benefits

1. **Improved Readability** - Visual icons are faster to scan than text
2. **Modern Interface** - Brings GitZen up to modern TUI standards  
3. **Better UX** - Intuitive symbols reduce cognitive load
4. **Professional Look** - Polished appearance improves user confidence

The icon system maintains GitZen's performance and terminal compatibility while significantly enhancing the visual experience.