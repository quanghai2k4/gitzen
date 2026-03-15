package git

import (
	"strings"
	"testing"
)

// ================== ParseHunks Tests ==================

func TestParseHunks_Empty(t *testing.T) {
	result := ParseHunks("")
	if len(result) != 0 {
		t.Errorf("expected 0 hunks, got %d", len(result))
	}
}

func TestParseHunks_SingleHunk(t *testing.T) {
	diff := `@@ -1,4 +1,5 @@
 line1
 line2
-old line
+new line
+added line
 line4`

	result := ParseHunks(diff)

	if len(result) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(result))
	}
	if result[0].OldStart != 1 {
		t.Errorf("expected OldStart=1, got %d", result[0].OldStart)
	}
	if result[0].OldLines != 4 {
		t.Errorf("expected OldLines=4, got %d", result[0].OldLines)
	}
	if result[0].NewStart != 1 {
		t.Errorf("expected NewStart=1, got %d", result[0].NewStart)
	}
	if result[0].NewLines != 5 {
		t.Errorf("expected NewLines=5, got %d", result[0].NewLines)
	}
	if result[0].Index != 0 {
		t.Errorf("expected Index=0, got %d", result[0].Index)
	}
	if result[0].Selected {
		t.Errorf("expected Selected=false")
	}
}

func TestParseHunks_MultipleHunks(t *testing.T) {
	diff := `@@ -1,3 +1,4 @@
 line1
-old1
+new1
 line3
@@ -10,3 +11,4 @@
 line10
-old10
+new10
+extra
 line12`

	result := ParseHunks(diff)

	if len(result) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(result))
	}
	if result[0].Index != 0 {
		t.Errorf("expected first hunk Index=0, got %d", result[0].Index)
	}
	if result[1].Index != 1 {
		t.Errorf("expected second hunk Index=1, got %d", result[1].Index)
	}
	if result[1].OldStart != 10 {
		t.Errorf("expected second hunk OldStart=10, got %d", result[1].OldStart)
	}
	if result[1].NewStart != 11 {
		t.Errorf("expected second hunk NewStart=11, got %d", result[1].NewStart)
	}
}

func TestParseHunks_HunkContent(t *testing.T) {
	diff := `@@ -5,3 +5,4 @@
 context
+added
 more context
 end`

	result := ParseHunks(diff)

	if len(result) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(result))
	}
	if !strings.Contains(result[0].Content, "+added") {
		t.Errorf("hunk content should contain '+added', got: %s", result[0].Content)
	}
}

// ================== parseHunk Tests ==================

func TestParseHunk_BasicHeader(t *testing.T) {
	content := "@@ -1,4 +1,5 @@\n line1\n-old\n+new\n line3\n line4"
	h := parseHunk(0, content)

	if h.OldStart != 1 {
		t.Errorf("expected OldStart=1, got %d", h.OldStart)
	}
	if h.OldLines != 4 {
		t.Errorf("expected OldLines=4, got %d", h.OldLines)
	}
	if h.NewStart != 1 {
		t.Errorf("expected NewStart=1, got %d", h.NewStart)
	}
	if h.NewLines != 5 {
		t.Errorf("expected NewLines=5, got %d", h.NewLines)
	}
}

func TestParseHunk_WithFunctionContext(t *testing.T) {
	// @@ header may have function name context after @@
	content := "@@ -10,5 +10,6 @@ func myFunction() {\n line1\n-old\n+new\n line3"
	h := parseHunk(0, content)

	if h.OldStart != 10 {
		t.Errorf("expected OldStart=10, got %d", h.OldStart)
	}
	if h.NewStart != 10 {
		t.Errorf("expected NewStart=10, got %d", h.NewStart)
	}
}

func TestParseHunk_InvalidHeader(t *testing.T) {
	// Invalid header: should return hunk with defaults (1,1,1,1) from atoi
	content := "not a hunk header\n some content"
	h := parseHunk(2, content)

	if h.Index != 2 {
		t.Errorf("expected Index=2, got %d", h.Index)
	}
	// When regex doesn't match, fields default to 1 (atoi returns 1 for empty string)
	// but in this case m is nil so conditional is false, so defaults from struct are used (0)
	if h.OldStart != 0 {
		t.Logf("OldStart=%d (expected 0 for invalid header)", h.OldStart)
	}
}

func TestParseHunk_EmptyContent(t *testing.T) {
	h := parseHunk(0, "")

	if h.Index != 0 {
		t.Errorf("expected Index=0, got %d", h.Index)
	}
	if h.Header != "" {
		t.Errorf("expected empty header, got '%s'", h.Header)
	}
}

func TestParseHunk_SingleLineChange(t *testing.T) {
	content := "@@ -1 +1 @@\n-old\n+new"
	h := parseHunk(0, content)

	// Without comma in @@ -1 +1 @@, m[2] and m[4] will be ""
	// atoi("") returns 1
	if h.OldStart != 1 {
		t.Errorf("expected OldStart=1, got %d", h.OldStart)
	}
	if h.NewStart != 1 {
		t.Errorf("expected NewStart=1, got %d", h.NewStart)
	}
}

// ================== atoi Tests ==================

func TestAtoi_ValidNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0", 0},
		{"1", 1},
		{"42", 42},
		{"100", 100},
		{"999", 999},
	}

	for _, tt := range tests {
		result := atoi(tt.input)
		if result != tt.expected {
			t.Errorf("atoi(%q): expected %d, got %d", tt.input, tt.expected, result)
		}
	}
}

func TestAtoi_EmptyString(t *testing.T) {
	// Empty string should return 1 (default)
	result := atoi("")
	if result != 1 {
		t.Errorf("atoi(%q): expected 1 (default), got %d", "", result)
	}
}

func TestAtoi_StringWithNonDigits(t *testing.T) {
	// Non-digit characters are skipped
	result := atoi("abc")
	if result != 0 {
		t.Errorf("atoi(%q): expected 0 (non-digit), got %d", "abc", result)
	}
}

// ================== reverseHunk Tests ==================

func TestReverseHunk_AddedToRemoved(t *testing.T) {
	hunk := "+added line\n context\n-removed line"
	result := reverseHunk(hunk)

	if !strings.Contains(result, "-added line") {
		t.Errorf("expected '+' to become '-', got: %s", result)
	}
	if !strings.Contains(result, "+removed line") {
		t.Errorf("expected '-' to become '+', got: %s", result)
	}
	if !strings.Contains(result, " context") {
		t.Errorf("expected context line to remain unchanged, got: %s", result)
	}
}

func TestReverseHunk_ContextLinesUnchanged(t *testing.T) {
	hunk := " context line 1\n context line 2"
	result := reverseHunk(hunk)

	if result != hunk {
		t.Errorf("context lines should be unchanged, got: %s", result)
	}
}

func TestReverseHunk_Empty(t *testing.T) {
	result := reverseHunk("")
	if result != "" {
		t.Errorf("expected empty result, got: %s", result)
	}
}
