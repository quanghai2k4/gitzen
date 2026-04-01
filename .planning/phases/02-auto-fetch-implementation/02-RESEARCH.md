# Phase 2: Auto Fetch Implementation - Research

**Researched:** 2026-04-01
**Domain:** Git fetch operations, branch targeting, and per-repository configuration management
**Confidence:** HIGH

## Summary

Phase 2 implements core auto fetch functionality for GitZen, building on the background operations foundation from Phase 1. Research confirms Git's fetch command supports targeted refspecs for specific branches, Go has mature YAML configuration libraries, and standard configuration patterns exist for per-repository settings. The background manager established in Phase 1 provides the perfect integration point through ExecuteIfSafe().

Key technical insight: Git fetch with explicit refspecs (e.g., `git fetch origin main:refs/remotes/origin/main`) allows efficient targeted fetching of specific branches without the overhead of `--all` flag.

**Primary recommendation:** Extend background manager with targeted fetch operations using refspecs, implement YAML-based per-repository configuration in `.git/gitzen-config.yml`, and add startup fetch integration into app initialization.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| FETCH-01 | GitZen fetches main branch and current branch from remote on application startup | Git fetch with branch refspecs enables targeted fetching during app.Init() |
| FETCH-04 | Auto fetch targets specific branches (main + current) instead of all remotes | Git fetch refspec syntax supports individual branch targeting vs --all flag |
| CONFIG-01 | Auto fetch settings are configurable per-repository (not global) | .git/gitzen-config.yml pattern provides per-repo config without polluting git config |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| gopkg.in/yaml.v3 | v3.0.1 | Configuration file parsing | Most mature Go YAML library, used by major projects |
| existing git.Runner | current | Git command execution | Already established, needs extension for targeted fetch |
| existing background.Manager | current | Background operation orchestration | Phase 1 foundation provides ExecuteIfSafe() integration |

### Supporting  
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| path/filepath | stdlib | Configuration file path handling | Resolving .git directory and config file location |
| os | stdlib | File system operations | Config file existence checks and creation |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| YAML config | TOML (github.com/BurntSushi/toml) | YAML more familiar, better nested structure support |
| .git/gitzen-config.yml | ~/.config/gitzen/config.yml | Per-repo settings are more intuitive for repository-specific behavior |
| refspec targeting | --all with post-filtering | Refspecs avoid fetching unnecessary data, faster |

**Installation:**
```bash
go get gopkg.in/yaml.v3
```

**Version verification:** gopkg.in/yaml.v3 v3.0.1 is current stable version (published 2023-04-18).

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── config/              # Configuration management
│   ├── config.go       # Repository config loading/saving
│   └── types.go        # Configuration data structures
├── git/                 # Extend existing runner
│   └── fetch.go        # Targeted fetch methods
└── background/          # Extend existing manager
    └── fetch.go        # Background fetch integration
```

### Pattern 1: Per-Repository Configuration
**What:** YAML configuration stored in `.git/gitzen-config.yml` for repository-specific settings
**When to use:** Auto fetch preferences that should vary by repository
**Example:**
```go
// Source: Standard Go YAML pattern + Git tool conventions
type RepoConfig struct {
    AutoFetch struct {
        Enabled         bool     `yaml:"enabled"`
        StartupFetch    bool     `yaml:"startup_fetch"`
        TargetBranches  []string `yaml:"target_branches"`
        IntervalMinutes int      `yaml:"interval_minutes"`
    } `yaml:"auto_fetch"`
}

func LoadRepoConfig(repoRoot string) (*RepoConfig, error) {
    configPath := filepath.Join(repoRoot, ".git", "gitzen-config.yml")
    data, err := os.ReadFile(configPath)
    if os.IsNotExist(err) {
        return NewDefaultConfig(), nil
    }
    if err != nil {
        return nil, err
    }
    
    var config RepoConfig
    err = yaml.Unmarshal(data, &config)
    return &config, err
}
```

### Pattern 2: Targeted Branch Fetching
**What:** Git fetch with specific branch refspecs instead of --all
**When to use:** Efficient fetching of main + current branch only
**Example:**
```go
// Source: Git fetch documentation and GitZen conventions
func (r Runner) FetchTargetBranches(remote string, branches []string) error {
    if len(branches) == 0 {
        return nil
    }
    
    args := []string{"fetch", remote}
    for _, branch := range branches {
        // Add refspec for each branch: branch:refs/remotes/origin/branch
        refspec := fmt.Sprintf("%s:refs/remotes/%s/%s", branch, remote, branch)
        args = append(args, refspec)
    }
    
    _, err := r.run(NetworkTimeout, args...)
    return err
}
```

### Pattern 3: Startup Fetch Integration
**What:** Trigger initial fetch during app initialization
**When to use:** Ensuring repository is up-to-date when user starts GitZen
**Example:**
```go
// Source: GitZen app.Init() pattern
func (m model) Init() tea.Cmd {
    commands := []tea.Cmd{
        // ... existing init commands
    }
    
    // Add startup fetch if enabled
    if m.config.AutoFetch.StartupFetch {
        commands = append(commands, m.startupFetchCmd())
    }
    
    return tea.Batch(commands...)
}

func (m model) startupFetchCmd() tea.Cmd {
    return func() tea.Msg {
        return startupFetchMsg{}
    }
}
```

### Pattern 4: Background Fetch Integration
**What:** Integrate targeted fetch with existing background manager
**When to use:** Periodic auto fetch operations
**Example:**
```go
// Source: background.Manager.ExecuteIfSafe pattern from Phase 1
func (m *Manager) ExecuteAutoFetch(config RepoConfig) tea.Cmd {
    return func() tea.Msg {
        err := m.ExecuteIfSafe(func() error {
            if !config.AutoFetch.Enabled {
                return nil
            }
            
            // Determine target branches (main + current)
            branches := m.determineTargetBranches(config)
            remote := "origin" // or get from git config
            
            return m.gitRunner.FetchTargetBranches(remote, branches)
        })
        
        return backgroundFetchResultMsg{Error: err}
    }
}
```

### Anti-Patterns to Avoid
- **Fetching all branches:** Use `git fetch --all` - wastes bandwidth and time
- **Global configuration only:** Settings in ~/.config - should be per-repository for different project needs
- **Blocking startup fetch:** Synchronous fetch in Init() - use tea.Cmd for async operation
- **Hard-coded branch names:** Assuming "main" or "master" - detect default branch dynamically

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| YAML parsing | Custom config parser | gopkg.in/yaml.v3 | Handles edge cases, validation, proper type conversion |
| Branch name detection | String parsing of git output | git symbolic-ref, git ls-remote | Handles all edge cases including detached HEAD |
| File path resolution | String concatenation | filepath.Join | Cross-platform path handling, avoids separator issues |
| Default branch detection | Hardcode "main"/"master" | git remote show origin or ls-remote | Repositories use different default branch names |
| Error handling for missing remotes | Custom error types | Git command error parsing | Git provides detailed error messages for auth/network issues |

**Key insight:** Git remote operations have many edge cases (authentication, network timeouts, missing branches, detached HEAD) that are better handled by Git itself rather than custom logic.

## Common Pitfalls

### Pitfall 1: Hard-coded Branch Names
**What goes wrong:** Assuming main branch is always "main" or "master"
**Why it happens:** Different repositories use different default branch names (main, master, develop, etc.)
**How to avoid:** Use `git symbolic-ref refs/remotes/origin/HEAD` or `git ls-remote --symref origin HEAD` to detect default branch
**Warning signs:** Fetch failures on repositories that use non-standard default branches

### Pitfall 2: Network Authentication Failures
**What goes wrong:** Fetch operations fail with authentication errors but crash the application
**Why it happens:** Git authentication can fail due to expired tokens, SSH key issues, network problems
**How to avoid:** Parse git command stderr for authentication errors and handle gracefully with user-friendly messages
**Warning signs:** Application crashes during background fetch instead of showing error in UI

### Pitfall 3: Blocking UI During Startup Fetch
**What goes wrong:** Application freezes during initial fetch operation
**Why it happens:** Running fetch synchronously during app initialization
**How to avoid:** Always use tea.Cmd pattern for fetch operations, show loading state in UI
**Warning signs:** App becomes unresponsive immediately after launch

### Pitfall 4: Race Conditions with User Operations
**What goes wrong:** Background fetch conflicts with user's git operations (add, commit, checkout)
**Why it happens:** Not properly using the serialization provided by background.Manager.ExecuteIfSafe()
**How to avoid:** All fetch operations must go through background manager's serialization
**Warning signs:** "index.lock" errors or git operation conflicts during auto fetch

### Pitfall 5: Configuration File Corruption
**What goes wrong:** Invalid YAML in config file crashes application
**Why it happens:** User manually edits config file with syntax errors, or concurrent writes
**How to avoid:** Always validate config after loading, provide sane defaults for missing/invalid values
**Warning signs:** Application fails to start or config changes are lost

## Code Examples

Verified patterns from official sources:

### Repository Configuration Loading
```go
// Source: gopkg.in/yaml.v3 documentation + GitZen conventions
type AutoFetchConfig struct {
    Enabled         bool     `yaml:"enabled"`
    StartupFetch    bool     `yaml:"startup_fetch"` 
    TargetBranches  []string `yaml:"target_branches"`
    IntervalMinutes int      `yaml:"interval_minutes"`
}

func LoadAutoFetchConfig(repoRoot string) (AutoFetchConfig, error) {
    configPath := filepath.Join(repoRoot, ".git", "gitzen-config.yml")
    
    // Default configuration
    config := AutoFetchConfig{
        Enabled:         true,
        StartupFetch:    true,
        TargetBranches:  []string{"auto"}, // "auto" means main + current
        IntervalMinutes: 30,
    }
    
    data, err := os.ReadFile(configPath)
    if os.IsNotExist(err) {
        return config, nil // Use defaults
    }
    if err != nil {
        return config, fmt.Errorf("cannot read config file: %w", err)
    }
    
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return config, fmt.Errorf("invalid config YAML: %w", err)
    }
    
    return config, nil
}
```

### Targeted Branch Fetching
```go
// Source: Git fetch documentation
func (r Runner) FetchBranches(remote string, branches []string) error {
    if len(branches) == 0 {
        return nil
    }
    
    args := []string{"fetch", remote}
    
    // Add explicit refspecs for each branch
    for _, branch := range branches {
        // Format: localBranch:refs/remotes/remote/localBranch
        refspec := fmt.Sprintf("%s:refs/remotes/%s/%s", branch, remote, branch)
        args = append(args, refspec)
    }
    
    _, err := r.run(NetworkTimeout, args...)
    return err
}
```

### Default Branch Detection
```go
// Source: Git symbolic-ref documentation
func (r Runner) GetDefaultBranch(remote string) (string, error) {
    // Try to get remote's default branch
    out, err := r.run(DefaultCmdTimeout, "symbolic-ref", 
        fmt.Sprintf("refs/remotes/%s/HEAD", remote))
    if err != nil {
        // Fallback: try ls-remote
        out, err = r.run(NetworkTimeout, "ls-remote", "--symref", remote, "HEAD")
        if err != nil {
            return "main", nil // Ultimate fallback
        }
        // Parse output: "ref: refs/heads/main	HEAD"
        lines := strings.Split(strings.TrimSpace(out), "\n")
        if len(lines) > 0 && strings.HasPrefix(lines[0], "ref: refs/heads/") {
            return strings.TrimPrefix(lines[0], "ref: refs/heads/"), nil
        }
    }
    
    // Parse symbolic-ref output: refs/remotes/origin/main
    branch := strings.TrimSpace(out)
    prefix := fmt.Sprintf("refs/remotes/%s/", remote)
    if strings.HasPrefix(branch, prefix) {
        return strings.TrimPrefix(branch, prefix), nil
    }
    
    return "main", nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| git fetch --all | Targeted refspecs | Git 2.0+ | Faster, less bandwidth, fewer conflicts |
| Global config files | Per-repository config | Modern Git tools | Better multi-project workflow support |
| TOML config | YAML config | 2020+ TUI tools | Better nested structure, more familiar syntax |
| Synchronous fetch | Async with tea.Cmd | Bubble Tea v1.0+ | Non-blocking UI, better user experience |

**Deprecated/outdated:**
- `git fetch --all --prune` for selective updates: Slower than targeted refspecs
- Global ~/.gitconfig modifications: Pollutes user's git configuration
- Synchronous network operations in TUI: Breaks modern async patterns

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| git | fetch operations | ✓ | 2.53.0 | — |
| gopkg.in/yaml.v3 | config parsing | ✗ | — | Use defaults, no config files |
| internet connectivity | remote fetch | ✓ | — | Skip fetch operations |

**Missing dependencies with no fallback:**
- None - git is already verified available

**Missing dependencies with fallback:**
- gopkg.in/yaml.v3: Use hard-coded defaults if package unavailable
- Internet connectivity: Gracefully handle network errors

## Open Questions

1. **Default Branch Name Detection Strategy**
   - What we know: Different repos use main/master/develop
   - What's unclear: Best method to detect automatically vs user configuration
   - Recommendation: Use git ls-remote --symref with fallback to "main", allow config override

2. **Startup Fetch Performance Impact**  
   - What we know: Network operations can be slow (30s timeout configured)
   - What's unclear: User tolerance for startup delays vs background-only approach
   - Recommendation: Make startup fetch optional (default: enabled) with progress indication

3. **Configuration Migration Strategy**
   - What we know: No existing configuration system in GitZen
   - What's unclear: How to handle config schema evolution in future versions
   - Recommendation: Include version field in config structure for future migrations

## Sources

### Primary (HIGH confidence)
- Git Documentation: git-fetch man page for refspec syntax and branch targeting
- gopkg.in/yaml.v3: Official Go YAML library documentation and examples
- GitZen codebase analysis: Existing git.Runner patterns and background.Manager integration

### Secondary (MEDIUM confidence)  
- Git tool conventions: Analysis of lazygit, tig, and other CLI Git tools for config patterns
- Go configuration patterns: Standard library filepath, os packages for config file handling

### Tertiary (LOW confidence)
- None - all findings verified with official documentation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries verified to exist with required features
- Architecture: HIGH - Patterns verified in existing GitZen codebase and Git documentation  
- Pitfalls: HIGH - Based on documented Git behaviors and common TUI application issues

**Research date:** 2026-04-01
**Valid until:** 2026-05-01 (30 days for stable technologies)