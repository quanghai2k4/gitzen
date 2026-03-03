package git

import (
	"regexp"
	"strings"
)

type Hunk struct {
	Index    int
	Header   string
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Content  string
	Selected bool
}

var (
	hunkHeaderRE = regexp.MustCompile(`^@@ -(\d+),?(\d*) \+(\d+),?(\d*) @@(.*)`)
)

func ParseHunks(diff string) []Hunk {
	var hunks []Hunk
	lines := strings.Split(diff, "\n")

	var currentLines []string
	hunkIndex := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			if len(currentLines) > 0 {
				hunks = append(hunks, parseHunk(hunkIndex, strings.Join(currentLines, "\n")))
				hunkIndex++
			}
			currentLines = []string{line}
		} else if len(currentLines) > 0 || strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") {
			if len(currentLines) > 0 {
				currentLines = append(currentLines, line)
			}
		}
	}

	if len(currentLines) > 0 {
		hunks = append(hunks, parseHunk(hunkIndex, strings.Join(currentLines, "\n")))
	}

	return hunks
}

func parseHunk(index int, content string) Hunk {
	lines := strings.Split(content, "\n")
	var header string
	if len(lines) > 0 {
		header = lines[0]
	}

	m := hunkHeaderRE.FindStringSubmatch(header)
	h := Hunk{
		Index:    index,
		Header:   header,
		Content:  content,
		Selected: false,
	}

	if len(m) >= 5 {
		oldStart := 1
		oldLines := 1
		newStart := 1
		newLines := 1

		if m[1] != "" {
			oldStart = atoi(m[1])
		}
		if m[2] != "" {
			oldLines = atoi(m[2])
		}
		if m[3] != "" {
			newStart = atoi(m[3])
		}
		if m[4] != "" {
			newLines = atoi(m[4])
		}

		h.OldStart = oldStart
		h.OldLines = oldLines
		h.NewStart = newStart
		h.NewLines = newLines

		if len(m) > 5 {
			h.Header = "@@" + m[0][2:] + m[5] + " @@"
		}
	}

	return h
}

func atoi(s string) int {
	if s == "" {
		return 1
	}
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
