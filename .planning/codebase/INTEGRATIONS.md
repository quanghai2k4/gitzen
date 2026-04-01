# External Integrations

**Analysis Date:** 2026-04-01

## APIs & External Services

**Version Control:**
- Git - Core dependency for all repository operations
  - SDK/Client: `os/exec` with `git` commands
  - Auth: System git credential manager

**GitHub Integration:**
- GitHub API - Release automation via GoReleaser
  - SDK/Client: `goreleaser/goreleaser-action@v6`
  - Auth: `GITHUB_TOKEN` (CI/CD only)

## Data Storage

**Databases:**
- None - Stateless application reading git repository data

**File Storage:**
- Local filesystem only
  - Git repository files
  - Optional debug logs (`GITZEN_LOG` environment variable)

**Caching:**
- None - Real-time git command execution

## Authentication & Identity

**Auth Provider:**
- System Git credentials
  - Implementation: Delegates to system git configuration
  - Uses existing SSH keys, credential managers, or HTTPS tokens

## Monitoring & Observability

**Error Tracking:**
- None - Local debugging only

**Logs:**
- File-based debug logging (`internal/logger`)
  - Enabled via `GITZEN_DEBUG` or `GITZEN_LOG` environment variables
  - Non-blocking for TUI operation

## CI/CD & Deployment

**Hosting:**
- GitHub Releases - Binary distribution
- Homebrew Tap - macOS package manager integration
- Scoop Bucket - Windows package manager integration

**CI Pipeline:**
- GitHub Actions
  - `.github/workflows/ci.yml` - Test and build verification
  - `.github/workflows/release.yml` - Automated releases
  - Codecov integration for coverage reporting

## Environment Configuration

**Required env vars:**
- None (all optional)

**Optional env vars:**
- `GITZEN_LOG` - Path for debug log file
- `GITZEN_DEBUG` - Enable debug logging to temp file

**Secrets location:**
- GitHub repository secrets (CI/CD only)
  - `GITHUB_TOKEN` - Release automation
  - `CODECOV_TOKEN` - Coverage reporting

## Webhooks & Callbacks

**Incoming:**
- None

**Outgoing:**
- None

## System Dependencies

**Runtime:**
- Git executable in PATH
  - Used via `os/exec.Command("git", ...)`
  - Required for all repository operations

**Development:**
- Go toolchain 1.24.0+
- Git (for repository management)
- Optional: Nix (development environment)
- Optional: GoReleaser (for releases)

## Package Distribution

**Automated Distribution:**
- GitHub Releases - Cross-platform binaries
- Homebrew Tap - `quanghai2k4/homebrew-tap`
- Scoop Bucket - `quanghai2k4/scoop-bucket`

**Manual Installation:**
- `install.sh` - Linux/macOS installer script
- `install.ps1` - Windows PowerShell installer
- Direct binary download from releases

---

*Integration audit: 2026-04-01*