package git

import "strings"

type CommitItem struct {
	Hash    string
	Message string
	Raw     string
}

func ParseLogOneline(out string) []CommitItem {
	lines := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n")
	items := make([]CommitItem, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 0 {
			continue
		}
		hash := parts[0]
		msg := ""
		if len(parts) == 2 {
			msg = parts[1]
		}
		items = append(items, CommitItem{Hash: hash, Message: msg, Raw: line})
	}
	return items
}

// ReflogEntry represents a git reflog entry
type ReflogEntry struct {
	Hash    string // short hash
	Ref     string // HEAD@{0}, HEAD@{1}, etc.
	Action  string // commit, checkout, rebase, etc.
	Message string // full message
}

// ParseReflog parses git reflog output
// Format: hash HEAD@{n}: action: message
func ParseReflog(out string) []ReflogEntry {
	lines := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n")
	entries := make([]ReflogEntry, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: abc1234 HEAD@{0}: commit: message
		// or: abc1234 HEAD@{0}: checkout: moving from x to y
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		hash := parts[0]
		rest := parts[1]

		// Parse HEAD@{n}: action: message
		colonIdx := strings.Index(rest, ": ")
		if colonIdx == -1 {
			entries = append(entries, ReflogEntry{
				Hash:    hash,
				Ref:     rest,
				Action:  "",
				Message: "",
			})
			continue
		}

		ref := rest[:colonIdx]
		afterRef := rest[colonIdx+2:]

		// Parse action: message
		action := ""
		message := afterRef
		colonIdx2 := strings.Index(afterRef, ": ")
		if colonIdx2 != -1 {
			action = afterRef[:colonIdx2]
			message = afterRef[colonIdx2+2:]
		}

		entries = append(entries, ReflogEntry{
			Hash:    hash,
			Ref:     ref,
			Action:  action,
			Message: message,
		})
	}
	return entries
}
