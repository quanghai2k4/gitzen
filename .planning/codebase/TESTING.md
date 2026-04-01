# Testing Patterns

**Analysis Date:** 2026-04-01

## Test Framework

**Runner:**
- Go built-in testing framework
- Config: Standard Go test configuration (no external config file)

**Assertion Library:**
- Built-in Go testing with `t.Errorf()`, `t.Fatalf()`

**Run Commands:**
```bash
go test ./...              # Run all tests
go test -v ./...           # Verbose output
make test                  # Run with race detection and coverage
make test-coverage         # Generate HTML coverage report
```

## Test File Organization

**Location:**
- Co-located with source files in same package

**Naming:**
- `*_test.go` pattern (e.g., `parse_hunk_test.go`, `logger_test.go`)

**Structure:**
```
internal/
├── git/
│   ├── git.go
│   ├── parse_hunk.go
│   ├── parse_hunk_test.go
│   ├── parse_log_test.go
│   └── parse_status_test.go
└── logger/
    ├── logger.go
    └── logger_test.go
```

## Test Structure

**Suite Organization:**
```go
func TestFunctionName_Scenario(t *testing.T) {
    // Test implementation
}

// Multiple scenarios for same function
func TestParseHunks_Empty(t *testing.T) { ... }
func TestParseHunks_SingleHunk(t *testing.T) { ... }
func TestParseHunks_MultipleHunks(t *testing.T) { ... }
```

**Patterns:**
- Test name format: `Test{FunctionName}_{Scenario}`
- Setup: Direct variable assignment, no complex setup
- Assertions: `t.Errorf()` for failures, `t.Fatalf()` for fatal errors
- Cleanup: `defer` statements for resource cleanup

## Mocking

**Framework:** No external mocking framework

**Patterns:**
```go
// Manual test helpers
func newLoggerFromFile(f *os.File) *Logger {
    return &Logger{
        file:    f,
        logger:  log.New(f, "", 0),
        enabled: true,
    }
}

// Temporary directories for file tests
tmpDir := t.TempDir()
logPath := filepath.Join(tmpDir, "test.log")
```

**What to Mock:**
- File system operations using `t.TempDir()`
- External dependencies through constructor injection

**What NOT to Mock:**
- Standard library functions
- Simple data transformations
- Pure functions without side effects

## Fixtures and Factories

**Test Data:**
```go
// Inline test data with descriptive variables
diff := `@@ -1,4 +1,5 @@
 line1
 line2
-old line
+new line
+added line
 line4`

// Table-driven tests for multiple inputs
tests := []struct {
    input    string
    expected int
}{
    {"0", 0},
    {"1", 1},
    {"42", 42},
}
```

**Location:**
- Test data defined inline within test functions
- No separate fixture files

## Coverage

**Requirements:** No enforced coverage threshold

**View Coverage:**
```bash
make test-coverage         # Generates coverage.out and coverage.html
go tool cover -html=coverage.out -o coverage.html
```

## Test Types

**Unit Tests:**
- Pure function testing: `TestAtoi_ValidNumber()`, `TestReverseHunk_AddedToRemoved()`
- Isolated component testing: `TestLogger_WriteToFile()`
- Error case testing: `TestParseHunk_InvalidHeader()`

**Integration Tests:**
- File system operations: `TestLogger_WriteToFile()`
- Git command parsing: `TestParseHunks_MultipleHunks()`
- Component interaction: Not extensively used

**E2E Tests:**
- Not implemented
- TUI testing would be complex with bubbletea

## Common Patterns

**Async Testing:**
```go
// Not commonly needed - most operations are synchronous
// No special async testing patterns observed
```

**Error Testing:**
```go
func TestLogger_DisabledWhenNoPath(t *testing.T) {
    l := &Logger{enabled: false}
    
    if l.IsEnabled() {
        t.Errorf("expected logger to be disabled")
    }
    
    // Verify operations don't panic
    l.Debug("debug %s", "test")
    l.Info("info %s", "test")
}
```

**File System Testing:**
```go
func TestLogger_WriteToFile(t *testing.T) {
    tmpDir := t.TempDir()
    logPath := filepath.Join(tmpDir, "test.log")
    
    f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        t.Fatalf("cannot create log file: %v", err)
    }
    defer f.Close()
    
    // Test implementation
}
```

**Table-Driven Tests:**
```go
func TestAtoi_ValidNumber(t *testing.T) {
    tests := []struct {
        input    string
        expected int
    }{
        {"0", 0},
        {"1", 1},
        {"42", 42},
    }
    
    for _, tt := range tests {
        result := atoi(tt.input)
        if result != tt.expected {
            t.Errorf("atoi(%q): expected %d, got %d", tt.input, tt.expected, result)
        }
    }
}
```

---

*Testing analysis: 2026-04-01*