package git

import (
	"testing"
)

// ================== ParseLogOneline Tests ==================

func TestParseLogOneline_Empty(t *testing.T) {
	result := ParseLogOneline("")
	if len(result) != 0 {
		t.Errorf("expected 0 commits, got %d", len(result))
	}
}

func TestParseLogOneline_SingleCommit(t *testing.T) {
	out := "abc1234 feat: add new feature"
	result := ParseLogOneline(out)

	if len(result) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(result))
	}
	if result[0].Hash != "abc1234" {
		t.Errorf("expected hash 'abc1234', got '%s'", result[0].Hash)
	}
	if result[0].Message != "feat: add new feature" {
		t.Errorf("expected message 'feat: add new feature', got '%s'", result[0].Message)
	}
	if result[0].Raw != out {
		t.Errorf("expected raw '%s', got '%s'", out, result[0].Raw)
	}
}

func TestParseLogOneline_MultipleCommits(t *testing.T) {
	out := "abc1234 feat: add feature\ndef5678 fix: fix bug\nghi9012 chore: update deps"
	result := ParseLogOneline(out)

	if len(result) != 3 {
		t.Fatalf("expected 3 commits, got %d", len(result))
	}
	if result[0].Hash != "abc1234" {
		t.Errorf("expected first hash 'abc1234', got '%s'", result[0].Hash)
	}
	if result[1].Hash != "def5678" {
		t.Errorf("expected second hash 'def5678', got '%s'", result[1].Hash)
	}
	if result[2].Hash != "ghi9012" {
		t.Errorf("expected third hash 'ghi9012', got '%s'", result[2].Hash)
	}
}

func TestParseLogOneline_WithDecoration(t *testing.T) {
	// git log --oneline --decorate includes refs in parens
	out := "abc1234 (HEAD -> main, origin/main) Initial commit"
	result := ParseLogOneline(out)

	if len(result) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(result))
	}
	if result[0].Hash != "abc1234" {
		t.Errorf("expected hash 'abc1234', got '%s'", result[0].Hash)
	}
	if result[0].Message != "(HEAD -> main, origin/main) Initial commit" {
		t.Errorf("unexpected message: '%s'", result[0].Message)
	}
}

func TestParseLogOneline_CommitWithNoMessage(t *testing.T) {
	// A line with only hash and no message
	out := "abc1234"
	result := ParseLogOneline(out)

	if len(result) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(result))
	}
	if result[0].Hash != "abc1234" {
		t.Errorf("expected hash 'abc1234', got '%s'", result[0].Hash)
	}
	if result[0].Message != "" {
		t.Errorf("expected empty message, got '%s'", result[0].Message)
	}
}

func TestParseLogOneline_WindowsLineEndings(t *testing.T) {
	// Windows CRLF line endings should be handled
	out := "abc1234 feat: add feature\r\ndef5678 fix: fix bug\r\n"
	result := ParseLogOneline(out)

	if len(result) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(result))
	}
	if result[0].Hash != "abc1234" {
		t.Errorf("expected hash 'abc1234', got '%s'", result[0].Hash)
	}
}

func TestParseLogOneline_SkipsEmptyLines(t *testing.T) {
	out := "abc1234 feat: add feature\n\ndef5678 fix: fix bug\n"
	result := ParseLogOneline(out)

	if len(result) != 2 {
		t.Fatalf("expected 2 commits (skipping blank line), got %d", len(result))
	}
}

// ================== ParseReflog Tests ==================

func TestParseReflog_Empty(t *testing.T) {
	result := ParseReflog("")
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}

func TestParseReflog_CommitEntry(t *testing.T) {
	out := "abc1234 HEAD@{0}: commit: feat: add feature"
	result := ParseReflog(out)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0].Hash != "abc1234" {
		t.Errorf("expected hash 'abc1234', got '%s'", result[0].Hash)
	}
	if result[0].Ref != "HEAD@{0}" {
		t.Errorf("expected ref 'HEAD@{0}', got '%s'", result[0].Ref)
	}
	if result[0].Action != "commit" {
		t.Errorf("expected action 'commit', got '%s'", result[0].Action)
	}
	if result[0].Message != "feat: add feature" {
		t.Errorf("expected message 'feat: add feature', got '%s'", result[0].Message)
	}
}

func TestParseReflog_CheckoutEntry(t *testing.T) {
	out := "abc1234 HEAD@{1}: checkout: moving from feature to main"
	result := ParseReflog(out)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0].Action != "checkout" {
		t.Errorf("expected action 'checkout', got '%s'", result[0].Action)
	}
	if result[0].Message != "moving from feature to main" {
		t.Errorf("expected message 'moving from feature to main', got '%s'", result[0].Message)
	}
}

func TestParseReflog_MultipleEntries(t *testing.T) {
	out := "abc1234 HEAD@{0}: commit: feat: add feature\ndef5678 HEAD@{1}: checkout: moving from main to feature\nghi9012 HEAD@{2}: commit: initial commit"
	result := ParseReflog(out)

	if len(result) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result))
	}
	if result[0].Hash != "abc1234" {
		t.Errorf("expected first hash 'abc1234', got '%s'", result[0].Hash)
	}
	if result[1].Hash != "def5678" {
		t.Errorf("expected second hash 'def5678', got '%s'", result[1].Hash)
	}
}

func TestParseReflog_EntryWithNoColon(t *testing.T) {
	// Entry without ": " separator
	out := "abc1234 HEAD@{0}"
	result := ParseReflog(out)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0].Ref != "HEAD@{0}" {
		t.Errorf("expected ref 'HEAD@{0}', got '%s'", result[0].Ref)
	}
	if result[0].Action != "" {
		t.Errorf("expected empty action, got '%s'", result[0].Action)
	}
	if result[0].Message != "" {
		t.Errorf("expected empty message, got '%s'", result[0].Message)
	}
}

func TestParseReflog_SkipsEmptyLines(t *testing.T) {
	out := "abc1234 HEAD@{0}: commit: message\n\ndef5678 HEAD@{1}: commit: other\n"
	result := ParseReflog(out)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
}

func TestParseReflog_WindowsLineEndings(t *testing.T) {
	out := "abc1234 HEAD@{0}: commit: feat\r\ndef5678 HEAD@{1}: commit: fix\r\n"
	result := ParseReflog(out)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries (CRLF), got %d", len(result))
	}
}

func TestParseReflog_RebaseEntry(t *testing.T) {
	out := "abc1234 HEAD@{0}: rebase (finish): returning to refs/heads/main"
	result := ParseReflog(out)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0].Action != "rebase (finish)" {
		t.Errorf("expected action 'rebase (finish)', got '%s'", result[0].Action)
	}
}
