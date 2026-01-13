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
