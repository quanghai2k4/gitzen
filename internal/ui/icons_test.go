package ui

import "testing"

// TestIconSystemBasicFunctionality tests that the icon system works correctly
func TestIconSystemBasicFunctionality(t *testing.T) {
	icons := DefaultIcons

	// Test file status icons
	tests := []struct {
		status  string
		staged  bool
		wantNot string // what we don't want (old system)
	}{
		{"M", true, "M"},   // staged modified should use ● not "M"
		{"A", true, "A"},   // staged added should use ✚ not "A"
		{"D", true, "D"},   // staged deleted should use ✖ not "D"
		{"R", true, "R"},   // staged renamed should use ⇄ not "R"
		{"M", false, " M"}, // unstaged modified should use ◐ not " M"
		{"D", false, " D"}, // unstaged deleted should use ⊗ not " D"
		{"?", false, "??"}, // untracked should use ◯ not "??"
	}

	for _, tt := range tests {
		got := icons.GetFileStatusIcon(tt.status, tt.staged)
		if got == tt.wantNot {
			t.Errorf("GetFileStatusIcon(%q, %v) = %q, should not be old text %q", tt.status, tt.staged, got, tt.wantNot)
		}
		if got == "" {
			t.Errorf("GetFileStatusIcon(%q, %v) returned empty string", tt.status, tt.staged)
		}
	}
}

// TestBranchIcons tests branch icon functionality
func TestBranchIcons(t *testing.T) {
	icons := DefaultIcons

	// Test branch type icons
	currentIcon := icons.GetBranchIcon(true, false)  // current branch
	localIcon := icons.GetBranchIcon(false, false)   // local branch
	remoteIcon := icons.GetBranchIcon(false, true)   // remote branch

	if currentIcon == "*" || localIcon == "*" || remoteIcon == "*" {
		t.Error("Branch icons should use Unicode symbols, not ASCII asterisk")
	}

	if currentIcon == localIcon {
		t.Error("Current branch and local branch should have different icons")
	}

	if localIcon == remoteIcon {
		t.Error("Local branch and remote branch should have different icons")
	}
}

// TestCommitCountIcons tests commit count indicators
func TestCommitCountIcons(t *testing.T) {
	icons := DefaultIcons

	aheadIcon := icons.GetCommitCountIcon(true)
	behindIcon := icons.GetCommitCountIcon(false)

	if aheadIcon == "+" || behindIcon == "-" {
		t.Error("Commit count icons should use Unicode arrows, not ASCII +/-")
	}

	if aheadIcon == behindIcon {
		t.Error("Ahead and behind icons should be different")
	}
}

// TestToastIcons tests toast notification icons
func TestToastIcons(t *testing.T) {
	icons := DefaultIcons

	successIcon := icons.GetToastIcon("success")
	errorIcon := icons.GetToastIcon("error")
	warningIcon := icons.GetToastIcon("warning")
	infoIcon := icons.GetToastIcon("info")

	// All icons should be different
	allIcons := []string{successIcon, errorIcon, warningIcon, infoIcon}
	for i, icon1 := range allIcons {
		for j, icon2 := range allIcons {
			if i != j && icon1 == icon2 {
				t.Errorf("Toast icons should be unique, found duplicate: %q", icon1)
			}
		}
	}
}

// TestFetchStatusIcons tests fetch status indicators
func TestFetchStatusIcons(t *testing.T) {
	icons := DefaultIcons

	progressIcon := icons.GetFetchStatusIcon("in_progress")
	successIcon := icons.GetFetchStatusIcon("success")
	errorIcon := icons.GetFetchStatusIcon("error")

	// Should not use emoji
	if progressIcon == "🔄" || successIcon == "✅" || errorIcon == "❌" {
		t.Error("Fetch status icons should use Unicode symbols, not emoji")
	}

	// All should be non-empty for valid statuses
	if progressIcon == "" || successIcon == "" || errorIcon == "" {
		t.Error("Valid fetch status should return non-empty icons")
	}
}

// TestAlternativeIcons ensures fallback icons work
func TestAlternativeIcons(t *testing.T) {
	icons := AlternativeIcons

	// Alternative icons should be more ASCII-compatible
	stagedAddedIcon := icons.GetFileStatusIcon("A", true)
	if stagedAddedIcon != "+" {
		t.Errorf("Alternative staged added icon should be ASCII +, got %q", stagedAddedIcon)
	}

	untrackedIcon := icons.GetFileStatusIcon("?", false)
	if untrackedIcon != "?" {
		t.Errorf("Alternative untracked icon should be ASCII ?, got %q", untrackedIcon)
	}
}