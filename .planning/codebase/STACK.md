# Technology Stack

**Analysis Date:** 2026-04-01

## Languages

**Primary:**
- Go 1.24.0 - Main application language

**Secondary:**
- Shell script - Installation scripts (`install.sh`, `install.ps1`)
- YAML - CI/CD configuration and GoReleaser config
- Makefile - Build automation

## Runtime

**Environment:**
- Go 1.24.0 runtime
- Cross-platform: Linux, macOS, Windows (amd64/arm64)

**Package Manager:**
- Go Modules (go.mod/go.sum)
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- Bubble Tea v1.3.10 - TUI framework for terminal user interfaces
- Lipgloss v1.1.0 - Terminal styling and layout
- Bubbles v0.21.0 - Pre-built UI components (viewport, textinput)

**Testing:**
- Go built-in testing framework
- Race detector enabled in CI

**Build/Dev:**
- GoReleaser v2 - Automated release pipeline
- Make - Build automation (`Makefile`)
- Nix flake - Development environment (`flake.nix`)

## Key Dependencies

**Critical:**
- `github.com/charmbracelet/bubbletea` v1.3.10 - Core TUI framework
- `github.com/charmbracelet/lipgloss` v1.1.0 - Terminal styling engine
- `github.com/charmbracelet/bubbles` v0.21.0 - UI components
- `github.com/charmbracelet/x/ansi` v0.10.1 - ANSI color support

**Infrastructure:**
- `github.com/atotto/clipboard` v0.1.4 - System clipboard integration
- `golang.org/x/sys` v0.36.0 - System-level operations
- `github.com/muesli/termenv` v0.16.0 - Terminal environment detection

## Configuration

**Environment:**
- `GITZEN_LOG` - Log file path for debugging
- `GITZEN_DEBUG` - Enable debug logging
- `CGO_ENABLED=0` - Static linking (no C dependencies)

**Build:**
- `go.mod` - Module dependencies
- `.goreleaser.yml` - Release configuration
- `Makefile` - Build targets and automation
- `flake.nix` - Nix development shell

## Platform Requirements

**Development:**
- Go 1.24.0+
- Git (for repository operations)
- Optional: Nix package manager for dev environment

**Production:**
- No runtime dependencies (statically linked binary)
- Git repository for operation
- Terminal with ANSI color support

---

*Stack analysis: 2026-04-01*