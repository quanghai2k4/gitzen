package git

import (
	"testing"
)

func TestParseStatusPorcelainV1Z_Empty(t *testing.T) {
	result := ParseStatusPorcelainV1Z([]byte{})
	if len(result.Staged) != 0 {
		t.Errorf("expected 0 staged, got %d", len(result.Staged))
	}
	if len(result.Unstaged) != 0 {
		t.Errorf("expected 0 unstaged, got %d", len(result.Unstaged))
	}
}

func TestParseStatusPorcelainV1Z_StagedFile(t *testing.T) {
	// M staged, space unstaged => only staged
	data := []byte("M  file.go\x00")
	result := ParseStatusPorcelainV1Z(data)

	if len(result.Staged) != 1 {
		t.Fatalf("expected 1 staged, got %d", len(result.Staged))
	}
	if result.Staged[0].Path != "file.go" {
		t.Errorf("expected path 'file.go', got '%s'", result.Staged[0].Path)
	}
	if result.Staged[0].Status != "M" {
		t.Errorf("expected status 'M', got '%s'", result.Staged[0].Status)
	}
	if !result.Staged[0].Staged {
		t.Errorf("expected Staged=true")
	}
	if len(result.Unstaged) != 0 {
		t.Errorf("expected 0 unstaged, got %d", len(result.Unstaged))
	}
}

func TestParseStatusPorcelainV1Z_UnstagedFile(t *testing.T) {
	// space staged, M unstaged => only unstaged
	data := []byte(" M file.go\x00")
	result := ParseStatusPorcelainV1Z(data)

	if len(result.Unstaged) != 1 {
		t.Fatalf("expected 1 unstaged, got %d", len(result.Unstaged))
	}
	if result.Unstaged[0].Path != "file.go" {
		t.Errorf("expected path 'file.go', got '%s'", result.Unstaged[0].Path)
	}
	if result.Unstaged[0].Status != "M" {
		t.Errorf("expected status 'M', got '%s'", result.Unstaged[0].Status)
	}
	if result.Unstaged[0].Staged {
		t.Errorf("expected Staged=false")
	}
	if len(result.Staged) != 0 {
		t.Errorf("expected 0 staged, got %d", len(result.Staged))
	}
}

func TestParseStatusPorcelainV1Z_BothStagedAndUnstaged(t *testing.T) {
	// MM => both staged (M) and unstaged (M)
	data := []byte("MM file.go\x00")
	result := ParseStatusPorcelainV1Z(data)

	if len(result.Staged) != 1 {
		t.Fatalf("expected 1 staged, got %d", len(result.Staged))
	}
	if len(result.Unstaged) != 1 {
		t.Fatalf("expected 1 unstaged, got %d", len(result.Unstaged))
	}
	if result.Staged[0].Path != "file.go" {
		t.Errorf("staged: expected path 'file.go', got '%s'", result.Staged[0].Path)
	}
	if result.Unstaged[0].Path != "file.go" {
		t.Errorf("unstaged: expected path 'file.go', got '%s'", result.Unstaged[0].Path)
	}
}

func TestParseStatusPorcelainV1Z_NewFile(t *testing.T) {
	// A staged, space unstaged => new file staged
	data := []byte("A  new_file.go\x00")
	result := ParseStatusPorcelainV1Z(data)

	if len(result.Staged) != 1 {
		t.Fatalf("expected 1 staged, got %d", len(result.Staged))
	}
	if result.Staged[0].Status != "A" {
		t.Errorf("expected status 'A', got '%s'", result.Staged[0].Status)
	}
	if result.Staged[0].Path != "new_file.go" {
		t.Errorf("expected path 'new_file.go', got '%s'", result.Staged[0].Path)
	}
}

func TestParseStatusPorcelainV1Z_UntrackedFile(t *testing.T) {
	// ?? => both x and y are '?'
	data := []byte("?? untracked.go\x00")
	result := ParseStatusPorcelainV1Z(data)

	// '?' != ' ' so both staged and unstaged lists will get this file
	if len(result.Staged) != 1 {
		t.Fatalf("expected 1 in staged (due to ?), got %d", len(result.Staged))
	}
	if len(result.Unstaged) != 1 {
		t.Fatalf("expected 1 in unstaged (due to ?), got %d", len(result.Unstaged))
	}
	if result.Staged[0].Status != "?" {
		t.Errorf("expected status '?', got '%s'", result.Staged[0].Status)
	}
}

func TestParseStatusPorcelainV1Z_DeletedFile(t *testing.T) {
	// D staged
	data := []byte("D  deleted.go\x00")
	result := ParseStatusPorcelainV1Z(data)

	if len(result.Staged) != 1 {
		t.Fatalf("expected 1 staged, got %d", len(result.Staged))
	}
	if result.Staged[0].Status != "D" {
		t.Errorf("expected status 'D', got '%s'", result.Staged[0].Status)
	}
}

func TestParseStatusPorcelainV1Z_MultipleFiles(t *testing.T) {
	// Multiple entries separated by null byte
	data := []byte("M  file1.go\x00 M file2.go\x00A  file3.go\x00")
	result := ParseStatusPorcelainV1Z(data)

	if len(result.Staged) != 2 {
		t.Fatalf("expected 2 staged, got %d", len(result.Staged))
	}
	if len(result.Unstaged) != 1 {
		t.Fatalf("expected 1 unstaged, got %d", len(result.Unstaged))
	}
}

func TestParseStatusPorcelainV1Z_ShortEntry(t *testing.T) {
	// Entry less than 3 chars should be ignored
	data := []byte("M\x00")
	result := ParseStatusPorcelainV1Z(data)

	if len(result.Staged) != 0 {
		t.Errorf("expected 0 staged (short entry), got %d", len(result.Staged))
	}
	if len(result.Unstaged) != 0 {
		t.Errorf("expected 0 unstaged (short entry), got %d", len(result.Unstaged))
	}
}

func TestParseStatusPorcelainV1Z_EmptyPath(t *testing.T) {
	// Only the status characters, no actual path
	data := []byte("MM \x00")
	result := ParseStatusPorcelainV1Z(data)

	// Path would be whitespace-only which trims to "", should be skipped
	if len(result.Staged) != 0 {
		t.Errorf("expected 0 staged (empty path), got %d", len(result.Staged))
	}
	if len(result.Unstaged) != 0 {
		t.Errorf("expected 0 unstaged (empty path), got %d", len(result.Unstaged))
	}
}
