# GitZen Updater

The updater package provides automatic update functionality for GitZen, allowing users to update to the latest version from GitHub releases.

## Features

- ✅ Check for latest releases from GitHub API
- ✅ Cross-platform support (Linux, macOS, Windows)
- ✅ Secure checksum verification
- ✅ Automatic binary replacement with backup
- ✅ Dry-run mode for testing
- ✅ Rate limit handling
- ✅ Network error handling
- ✅ Platform detection and asset matching

## Usage

### Command Line

```bash
# Check and install updates
gitzen --update

# Dry run (show what would be updated)
gitzen --update-dry-run
```

### Programmatic

```go
package main

import "gitzen/internal/updater"

func main() {
    u := updater.NewUpdater("1.0.0")
    
    // Check for updates
    release, err := u.CheckForUpdate()
    if err != nil {
        // handle error
    }
    
    if release != nil {
        // Update available
        options := updater.UpdateOptions{
            DryRun:  false,
            Force:   false,
            Backup:  true,
            Verbose: true,
        }
        
        result, err := u.Update(options)
        // handle result
    }
}
```

## Architecture

### Components

1. **Updater** - Main coordinator for the update process
2. **GitHubClient** - Handles GitHub API interactions
3. **PlatformDetector** - Detects current platform and matches assets

### Process Flow

1. **Check** - Query GitHub API for latest release
2. **Compare** - Compare current version with latest
3. **Download** - Download appropriate binary for platform
4. **Verify** - Verify checksum if available
5. **Backup** - Create backup of current binary
6. **Replace** - Replace current binary with new version

## Security

- Always downloads over HTTPS
- Verifies SHA256 checksums when available
- Creates backups before replacement
- Never executes downloaded code during update

## Error Handling

The updater handles various error conditions gracefully:

- Network connectivity issues
- GitHub API rate limiting
- Invalid/missing checksums
- Permission errors during replacement
- Unsupported platform detection

## Platform Support

Automatically detects and supports:

- Linux (amd64, arm64, 386)
- macOS/Darwin (amd64, arm64) 
- Windows (amd64, 386)

## Configuration

Environment variables (optional):

- `GITZEN_LOG` - Log file path for debugging
- `GITZEN_DEBUG` - Enable debug logging

## Testing

```bash
# Run unit tests
go test ./internal/updater/

# Run tests with coverage
go test -cover ./internal/updater/
```

## Implementation Details

### Asset Matching

The platform detector uses naming patterns to match release assets:

- `{binary}_{os}_{arch}` (e.g., `gitzen_linux_amd64`)
- `{binary}-{os}-{arch}` (e.g., `gitzen-linux-amd64`)
- Windows assets must have `.exe` extension

### Checksum Verification

Looks for checksum files in releases:
- Files containing "checksum" or "sha256" in name
- Supports standard format: `{hash}  {filename}`

### Binary Replacement

- Creates backup with `.backup.{version}` suffix
- Handles Windows executable replacement specially
- Preserves file permissions where possible